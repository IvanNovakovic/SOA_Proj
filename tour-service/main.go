package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"

    "github.com/gorilla/mux"

    "tour-service/auth"
    "tour-service/handler"
    "tour-service/repository"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    mongoURI := os.Getenv("MONGO_URI")
    if mongoURI == "" {
        mongoURI = "mongodb://localhost:27017"
    }
    dbName := os.Getenv("MONGO_DB")
    if dbName == "" {
        dbName = "tours"
    }

    repo, err := repository.NewTourRepository(ctx, mongoURI, dbName)
    if err != nil {
        log.Fatal("failed to connect to mongo:", err)
    }
    defer repo.Close(context.Background())

    r := mux.NewRouter()
    r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }).Methods("GET")

    // create an auth-protected subrouter for protected routes
    authSub := r.PathPrefix("").Subrouter()
    authSub.Use(func(next http.Handler) http.Handler { return auth.JWTAuthMiddleware(next) })

    handler.RegisterRoutes(r, authSub, repo)
    handler.RegisterKeyPointRoutes(r, authSub, repo)
    handler.RegisterReviewRoutes(r, authSub, repo)

    srv := &http.Server{
        Handler:      r,
        Addr:         ":8083",
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
    }

    go func() {
        log.Println("tour-service started on :8083")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop
    log.Println("shutting down tour-service...")
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    srv.Shutdown(shutdownCtx)
}
