package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"

	"tour-service/auth"
	"tour-service/handler"
	"tour-service/repository"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
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
		"service":  "tour-service",
		"action":   "tracer_init",
		"endpoint": endpoint,
	}).Info("Initializing Jaeger tracer")

	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service": "tour-service",
			"action":  "tracer_init",
			"error":   err.Error(),
		}).Fatal("Failed to create Jaeger exporter")
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("tour-service"),
			semconv.ServiceVersionKey.String("1.0.0"),
		)),
	)
	otel.SetTracerProvider(tp)

	logger.WithFields(logrus.Fields{
		"service": "tour-service",
		"action":  "tracer_init",
	}).Info("Jaeger tracer initialized successfully")

	return tp.Shutdown
}

func main() {
	initLogger()

	logger.WithFields(logrus.Fields{
		"service": "tour-service",
		"action":  "startup",
	}).Info("Starting tour-service")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize tracer
	cleanup := initTracer()
	defer cleanup(context.Background())

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "tours"
	}

	logger.WithFields(logrus.Fields{
		"service": "tour-service",
		"action":  "db_connect",
		"uri":     mongoURI,
		"db":      dbName,
	}).Info("Connecting to MongoDB")

	repo, err := repository.NewTourRepository(ctx, mongoURI, dbName)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service": "tour-service",
			"action":  "db_connect",
			"error":   err.Error(),
		}).Fatal("Failed to connect to MongoDB")
	}
	defer repo.Close(context.Background())

	logger.WithFields(logrus.Fields{
		"service": "tour-service",
		"action":  "db_connect",
	}).Info("Successfully connected to MongoDB")

	r := mux.NewRouter()

	// Add OpenTelemetry middleware
	r.Use(otelmux.Middleware("tour-service"))

	// Add optional auth middleware to parse JWT if present (for public routes that need user context)
	r.Use(func(next http.Handler) http.Handler { return auth.OptionalAuthMiddleware(next) })

	// Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler()).Methods("GET")

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
		logger.WithFields(logrus.Fields{
			"service": "tour-service",
			"action":  "server_start",
			"address": srv.Addr,
		}).Info("Tour service HTTP server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithFields(logrus.Fields{
				"service": "tour-service",
				"action":  "server_start",
				"error":   err.Error(),
			}).Fatal("Failed to start HTTP server")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	logger.WithFields(logrus.Fields{
		"service": "tour-service",
		"action":  "shutdown",
	}).Info("Shutting down server gracefully")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
	logger.WithFields(logrus.Fields{
		"service": "tour-service",
		"action":  "shutdown",
	}).Info("Server shutdown complete")
}
