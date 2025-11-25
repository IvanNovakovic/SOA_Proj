package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	pb "github.com/IvanNovakovic/SOA_Proj/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var logger = logrus.New()

func initLogger() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
}

func initTracer() func(context.Context) error {
	endpoint := os.Getenv("OTEL_EXPORTER_JAEGER_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:14268/api/traces"
	}

	logger.WithFields(logrus.Fields{
		"service":  "gateway-service",
		"action":   "tracer_init",
		"endpoint": endpoint,
	}).Info("Initializing Jaeger tracer")

	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service": "gateway-service",
			"action":  "tracer_init",
			"error":   err.Error(),
		}).Fatal("Failed to create Jaeger exporter")
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("gateway-service"),
			semconv.ServiceVersionKey.String("1.0.0"),
		)),
	)
	otel.SetTracerProvider(tp)

	logger.WithFields(logrus.Fields{
		"service": "gateway-service",
		"action":  "tracer_init",
	}).Info("Jaeger tracer initialized successfully")

	return tp.Shutdown
}

func newProxy(target string) *httputil.ReverseProxy {
	u, err := url.Parse(target)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service": "gateway-service",
			"action":  "proxy_init",
			"target":  target,
			"error":   err.Error(),
		}).Fatalf("Invalid proxy target")
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	// tweak the director to preserve original host header if needed
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = u.Host
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		// add small header to indicate proxied
		resp.Header.Set("X-Gateway-Proxy", "gateway-service")
		return nil
	}
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		logger.WithFields(logrus.Fields{
			"service": "gateway-service",
			"action":  "proxy_error",
			"path":    req.URL.Path,
			"error":   err.Error(),
		}).Error("Proxy error")
		http.Error(rw, "Bad Gateway", http.StatusBadGateway)
	}
	return proxy
}

// gRPC client connections
type grpcClients struct {
	stakeholderConn   *grpc.ClientConn
	stakeholderClient pb.StakeholderServiceClient
	followerConn      *grpc.ClientConn
	followerClient    pb.FollowerServiceClient
}

func initGRPCClients() (*grpcClients, error) {
	logger.WithFields(logrus.Fields{
		"service": "gateway-service",
		"action":  "grpc_client_init",
	}).Info("Initializing gRPC clients")

	// Connect to stakeholder service with OpenTelemetry interceptors
	stakeholderConn, err := grpc.Dial("stakeholders-service:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service": "gateway-service",
			"action":  "grpc_client_init",
			"target":  "stakeholders-service",
			"error":   err.Error(),
		}).Error("Failed to connect to stakeholders service")
		return nil, err
	}

	// Connect to follower service with OpenTelemetry interceptors
	followerConn, err := grpc.Dial("follower-service:9092",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service": "gateway-service",
			"action":  "grpc_client_init",
			"target":  "follower-service",
			"error":   err.Error(),
		}).Error("Failed to connect to follower service")
		stakeholderConn.Close()
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"service": "gateway-service",
		"action":  "grpc_client_init",
	}).Info("gRPC clients initialized successfully")

	return &grpcClients{
		stakeholderConn:   stakeholderConn,
		stakeholderClient: pb.NewStakeholderServiceClient(stakeholderConn),
		followerConn:      followerConn,
		followerClient:    pb.NewFollowerServiceClient(followerConn),
	}, nil
}

func (gc *grpcClients) Close() {
	if gc.stakeholderConn != nil {
		gc.stakeholderConn.Close()
	}
	if gc.followerConn != nil {
		gc.followerConn.Close()
	}
}

func main() {
	initLogger()

	logger.WithFields(logrus.Fields{
		"service": "gateway-service",
		"action":  "startup",
	}).Info("Starting gateway-service")

	// Initialize tracer
	cleanup := initTracer()
	defer cleanup(context.Background())

	// Initialize gRPC clients
	grpcClients, err := initGRPCClients()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service": "gateway-service",
			"action":  "grpc_client_init",
			"error":   err.Error(),
		}).Fatalf("Failed to initialize gRPC clients")
	}
	defer grpcClients.Close()

	// route prefix -> backend service (docker service name + internal port)
	routes := map[string]string{
		// blog service: blogs, comments, likes
		"/blog":     "http://blog-service:8081",
		"/blogs":    "http://blog-service:8081",
		"/comments": "http://blog-service:8081",
		"/likes":    "http://blog-service:8081",

		// follower service (other endpoints still via HTTP)
		"/follow":          "http://follower-service:8082",
		"/following":       "http://follower-service:8082",
		"/recommendations": "http://follower-service:8082",

		// tour service
		"/tour":      "http://tour-service:8083",
		"/tours":     "http://tour-service:8083",
		"/review":    "http://tour-service:8083",
		"/reviews":   "http://tour-service:8083",
		"/kp":        "http://tour-service:8083",
		"/keypoints": "http://tour-service:8083",

		// stakeholders (users/auth) - register still via HTTP
		"/health":   "http://stakeholders-service:8080",
		"/user":     "http://stakeholders-service:8080",
		"/users":    "http://stakeholders-service:8080",
		"/auth":     "http://stakeholders-service:8080",
		"/register": "http://stakeholders-service:8080",
	}

	// build proxies per target to reuse
	proxies := map[string]*httputil.ReverseProxy{}
	for _, target := range routes {
		if _, ok := proxies[target]; !ok {
			proxies[target] = newProxy(target)
		}
	}

	// gRPC handler for login - define before use
	handleLogin := func(w http.ResponseWriter, r *http.Request, client pb.StakeholderServiceClient) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		resp, err := client.Login(ctx, &pb.LoginRequest{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			logger.WithFields(logrus.Fields{
				"service": "gateway-service",
				"action":  "grpc_login",
				"error":   err.Error(),
			}).Error("gRPC login error")
			http.Error(w, "login failed", http.StatusUnauthorized)
			return
		}

		// Return response in format compatible with existing clients
		result := map[string]interface{}{
			"access_token": resp.Token,
			"token_type":   "Bearer",
			"expires_in":   3600,
			"user": map[string]interface{}{
				"id": resp.UserId,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}

	// gRPC handler for get followers
	handleGetFollowers := func(w http.ResponseWriter, r *http.Request, client pb.FollowerServiceClient) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract user ID from path /followers/{id}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 {
			http.Error(w, "user id required", http.StatusBadRequest)
			return
		}
		userID := parts[1]

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		resp, err := client.GetFollowers(ctx, &pb.GetFollowersRequest{
			UserId: userID,
		})
		if err != nil {
			logger.WithFields(logrus.Fields{
				"service": "gateway-service",
				"action":  "grpc_get_followers",
				"user_id": userID,
				"error":   err.Error(),
			}).Error("gRPC get followers error")
			http.Error(w, "failed to get followers", http.StatusInternalServerError)
			return
		}

		// Convert to string array format
		var followers []string
		for _, f := range resp.Followers {
			followers = append(followers, f.UserId)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(followers)
	}

	// Create router
	router := mux.NewRouter()

	// Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Main request handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Handle gRPC-based endpoints
		if path == "/login" || path == "/auth/login" {
			handleLogin(w, r, grpcClients.stakeholderClient)
			return
		}

		if strings.HasPrefix(path, "/followers/") {
			handleGetFollowers(w, r, grpcClients.followerClient)
			return
		}

		// find longest matching prefix for HTTP proxy
		var matchedTarget string
		var matchedPrefix string
		for prefix, target := range routes {
			if strings.HasPrefix(path, prefix) {
				if len(prefix) > len(matchedPrefix) {
					matchedPrefix = prefix
					matchedTarget = target
				}
			}
		}
		if matchedTarget != "" {
			proxies[matchedTarget].ServeHTTP(w, r)
			return
		}

		// no prefix match -> 502 or simple routing by host/path
		logger.WithFields(logrus.Fields{
			"service": "gateway-service",
			"action":  "route_not_found",
			"path":    path,
		}).Warn("Service not found for path")
		http.Error(w, "Service not found", http.StatusBadGateway)
	}

	// Catch-all route with OpenTelemetry tracing
	router.PathPrefix("/").Handler(otelhttp.NewHandler(http.HandlerFunc(handler), "gateway"))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.WithFields(logrus.Fields{
			"service": "gateway-service",
			"action":  "server_start",
			"address": srv.Addr,
		}).Info("Gateway service HTTP server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithFields(logrus.Fields{
				"service": "gateway-service",
				"action":  "server_start",
				"error":   err.Error(),
			}).Fatal("Failed to start HTTP server")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	logger.WithFields(logrus.Fields{
		"service": "gateway-service",
		"action":  "shutdown",
	}).Info("Shutting down server gracefully")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
	logger.WithFields(logrus.Fields{
		"service": "gateway-service",
		"action":  "shutdown",
	}).Info("Server shutdown complete")
}
