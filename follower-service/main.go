package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"

    "github.com/gorilla/mux"

    "follower-service/handler"
    "follower-service/repository"
)

func main() {
    _, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    neoURI := os.Getenv("NEO4J_URI")
    if neoURI == "" {
        neoURI = "bolt://neo4j:7687"
    }
    neoUser := os.Getenv("NEO4J_USER")
    if neoUser == "" {
        neoUser = "neo4j"
    }
    neoPass := os.Getenv("NEO4J_PASSWORD")
    if neoPass == "" {
        neoPass = "testtest123"
    }

    // Try connecting to Neo4j with retries (Neo4j may take time to become ready)
    var repo *repository.NeoRepository
    var err error
    maxAttempts := 20
    for i := 0; i < maxAttempts; i++ {
        attemptCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
        repo, err = repository.NewNeoRepository(attemptCtx, neoURI, neoUser, neoPass)
        cancel()
        if err == nil {
            log.Println("connected to neo4j")
            break
        }
        // exponential backoff with cap
        backoff := time.Duration(2*(i+1)) * time.Second
        if backoff > 20*time.Second {
            backoff = 20 * time.Second
        }
        log.Printf("neo4j connect attempt %d/%d failed: %v — retrying in %s", i+1, maxAttempts, err, backoff)
        time.Sleep(backoff)
    }
    if err != nil {
        log.Fatal("neo4j connect failed:", err)
    }

    r := mux.NewRouter()
    // health endpoint - use a short timeout so health checks don't hang indefinitely
    r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        if repo == nil {
            http.Error(w, "unhealthy", http.StatusServiceUnavailable)
            return
        }
        ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
        defer cancel()
        // perform a lightweight read to ensure neo is reachable
        if _, err := repo.Following(ctx, "_health_check_"); err != nil {
            // ignore error content, report unhealthy
            http.Error(w, "unhealthy", http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }).Methods("GET")
    handler.RegisterRoutes(r, repo)

    srv := &http.Server{
        Handler:      r,
        Addr:         ":8082",
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
    }

    go func() {
        log.Println("follower-service started on :8082")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop
    log.Println("shutting down follower-service...")
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    srv.Shutdown(shutdownCtx)
    repo.Close()
}
