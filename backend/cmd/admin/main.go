// Admin Service — serves the admin REST API for warehouse management.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ai-wms/ai-wms/backend/internal/api/middleware"
	"github.com/ai-wms/ai-wms/backend/pkg/logger"
)

func main() {
	port := os.Getenv("ADMIN_PORT")
	if port == "" {
		port = "8080"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	// Initialize structured logger.
	log := logger.New(logLevel)

	// Initialize CORS configuration from environment or defaults.
	corsConfig := middleware.DefaultCORSConfig()
	if origin := os.Getenv("CORS_ALLOWED_ORIGINS"); origin != "" {
		corsConfig.AllowedOrigins = []string{origin}
	}

	// TODO: Initialize database connection (PostgreSQL)
	// TODO: Initialize Redis connection
	// TODO: Initialize repositories
	// TODO: Initialize services
	// TODO: Initialize API router (chi/v5) with middleware
	// TODO: Register API routes

	// Build the middleware chain.
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"admin","version":"0.1.0"}`)
	})

	// Apply middleware stack: RequestID → Recovery → Logger → CORS → handler.
	handler := middleware.RequestID(
		middleware.Recovery(log.Logger)(
			middleware.Logger(log.Logger)(
				middleware.CORS(corsConfig)(mux),
			),
		),
	)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Info("Shutting down admin service...",
			slog.String("service", "admin"),
			slog.String("port", port))

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("Admin service forced to shutdown", slog.String("error", err.Error()))
		}
		log.Info("Admin service stopped", slog.String("service", "admin"))
	}()

	log.Info("Admin service starting",
		slog.String("service", "admin"),
		slog.String("port", port),
		slog.String("health_url", fmt.Sprintf("http://localhost:%s/health", port)),
	)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("Admin service failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
