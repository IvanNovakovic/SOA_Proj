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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

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
	tourConn          *grpc.ClientConn
	tourClient        pb.TourServiceClient
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

	// Connect to tour service with OpenTelemetry interceptors
	tourConn, err := grpc.Dial("tour-service:50053",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service": "gateway-service",
			"action":  "grpc_client_init",
			"target":  "tour-service",
			"error":   err.Error(),
		}).Error("Failed to connect to tour service")
		stakeholderConn.Close()
		followerConn.Close()
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
		tourConn:          tourConn,
		tourClient:        pb.NewTourServiceClient(tourConn),
	}, nil
}

func (gc *grpcClients) Close() {
	if gc.stakeholderConn != nil {
		gc.stakeholderConn.Close()
	}
	if gc.followerConn != nil {
		gc.followerConn.Close()
	}
	if gc.tourConn != nil {
		gc.tourConn.Close()
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
		"/tour":       "http://tour-service:8083",
		"/tours":      "http://tour-service:8083",
		"/review":     "http://tour-service:8083",
		"/reviews":    "http://tour-service:8083",
		"/kp":         "http://tour-service:8083",
		"/keypoints":  "http://tour-service:8083",
		"/executions": "http://tour-service:8083",

		// stakeholders (users/auth) - register still via HTTP
		"/health":   "http://stakeholders-service:8080",
		"/user":     "http://stakeholders-service:8080",
		"/users":    "http://stakeholders-service:8080",
		"/auth":     "http://stakeholders-service:8080",
		"/register": "http://stakeholders-service:8080",

		// purchase service
		"/cart":   "http://purchase-service:8086",
		"/tokens": "http://purchase-service:8086",
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

			// Check gRPC status code and return appropriate HTTP status
			st, ok := status.FromError(err)
			if ok {
				switch st.Code() {
				case codes.PermissionDenied:
					// User is blocked
					http.Error(w, st.Message(), http.StatusForbidden)
				case codes.Unauthenticated:
					// Invalid credentials
					http.Error(w, st.Message(), http.StatusUnauthorized)
				default:
					http.Error(w, "login failed", http.StatusInternalServerError)
				}
			} else {
				http.Error(w, "login failed", http.StatusUnauthorized)
			}
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

	// gRPC handler for get tour by ID
	handleGetTourByID := func(w http.ResponseWriter, r *http.Request, client pb.TourServiceClient) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract tour ID from path /tours/{id}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 {
			http.Error(w, "tour id required", http.StatusBadRequest)
			return
		}
		tourID := parts[1]

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		resp, err := client.GetTourByID(ctx, &pb.GetTourByIDRequest{
			TourId: tourID,
		})
		if err != nil {
			logger.WithFields(logrus.Fields{
				"service": "gateway-service",
				"action":  "grpc_get_tour_by_id",
				"tour_id": tourID,
				"error":   err.Error(),
			}).Error("gRPC get tour by ID error")

			st, ok := status.FromError(err)
			if ok && st.Code() == codes.NotFound {
				http.Error(w, "tour not found", http.StatusNotFound)
			} else {
				http.Error(w, "failed to get tour", http.StatusInternalServerError)
			}
			return
		}

		// Convert protobuf Tour to JSON with camelCase field names
		tourJSON := map[string]interface{}{
			"id":          resp.Tour.Id,
			"authorId":    resp.Tour.AuthorId,
			"name":        resp.Tour.Name,
			"description": resp.Tour.Description,
			"difficulty":  resp.Tour.Difficulty,
			"tags":        resp.Tour.Tags,
			"status":      resp.Tour.Status,
			"price":       resp.Tour.Price,
			"distance":    resp.Tour.Distance,
			"durations": map[string]interface{}{
				"walking": resp.Tour.Durations.Walking,
				"biking":  resp.Tour.Durations.Biking,
				"driving": resp.Tour.Durations.Driving,
			},
			"createdAt": resp.Tour.CreatedAt,
		}
		if resp.Tour.PublishedAt != "" {
			tourJSON["publishedAt"] = resp.Tour.PublishedAt
		}
		if resp.Tour.ArchivedAt != "" {
			tourJSON["archivedAt"] = resp.Tour.ArchivedAt
		}

		logger.WithFields(logrus.Fields{
			"service":  "gateway-service",
			"action":   "grpc_tour_response",
			"tour_id":  resp.Tour.Id,
			"authorId": resp.Tour.AuthorId,
		}).Info("Sending tour response with authorId")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tourJSON)
	}

	// gRPC handler for get tours by author
	handleGetToursByAuthor := func(w http.ResponseWriter, r *http.Request, client pb.TourServiceClient) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract author ID from path /tours/author/{authorId}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			http.Error(w, "author id required", http.StatusBadRequest)
			return
		}
		authorID := parts[2]

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		resp, err := client.GetToursByAuthor(ctx, &pb.GetToursByAuthorRequest{
			AuthorId: authorID,
		})
		if err != nil {
			logger.WithFields(logrus.Fields{
				"service":   "gateway-service",
				"action":    "grpc_get_tours_by_author",
				"author_id": authorID,
				"error":     err.Error(),
			}).Error("gRPC get tours by author error")
			http.Error(w, "failed to get tours", http.StatusInternalServerError)
			return
		}

		// Convert protobuf Tours to JSON with camelCase field names
		var toursJSON []map[string]interface{}
		for _, tour := range resp.Tours {
			tourJSON := map[string]interface{}{
				"id":          tour.Id,
				"authorId":    tour.AuthorId,
				"name":        tour.Name,
				"description": tour.Description,
				"difficulty":  tour.Difficulty,
				"tags":        tour.Tags,
				"status":      tour.Status,
				"price":       tour.Price,
				"distance":    tour.Distance,
				"durations": map[string]interface{}{
					"walking": tour.Durations.Walking,
					"biking":  tour.Durations.Biking,
					"driving": tour.Durations.Driving,
				},
				"createdAt": tour.CreatedAt,
			}
			if tour.PublishedAt != "" {
				tourJSON["publishedAt"] = tour.PublishedAt
			}
			if tour.ArchivedAt != "" {
				tourJSON["archivedAt"] = tour.ArchivedAt
			}
			toursJSON = append(toursJSON, tourJSON)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(toursJSON)
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

		if r.Method == "GET" && strings.HasPrefix(path, "/tours/") && strings.Contains(path, "/author/") {
			handleGetToursByAuthor(w, r, grpcClients.tourClient)
			return
		}

		if r.Method == "GET" && strings.HasPrefix(path, "/tours/") && len(strings.Split(strings.Trim(path, "/"), "/")) == 2 {
			handleGetTourByID(w, r, grpcClients.tourClient)
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
