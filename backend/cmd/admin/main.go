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

	"github.com/ai-wms/ai-wms/backend/internal/api"
	"github.com/ai-wms/ai-wms/backend/internal/api/middleware"
	"github.com/ai-wms/ai-wms/backend/internal/repository/postgres"
	"github.com/ai-wms/ai-wms/backend/internal/service"
	"github.com/ai-wms/ai-wms/backend/pkg/config"
	"github.com/ai-wms/ai-wms/backend/pkg/logger"
)

func main() {
	// Load configuration from environment variables (single source of truth).
	cfg := config.Load()

	// Validate configuration early — fail fast with a clear message.
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger from config.
	log := logger.New(cfg.LogLevel)
	log.Info("Configuration loaded",
		slog.String("env", cfg.Env),
		slog.String("log_level", cfg.LogLevel),
		slog.String("admin_port", cfg.AdminPort),
	)

	// Initialize CORS configuration.
	corsConfig := middleware.DefaultCORSConfig()
	if origin := os.Getenv("CORS_ALLOWED_ORIGINS"); origin != "" {
		corsConfig.AllowedOrigins = []string{origin}
	}

	// Initialize database connection.
	db, err := postgres.NewDB(context.Background(), cfg)
	if err != nil {
		log.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	// Initialize repositories.
	warehouseRepo := postgres.NewWarehouseRepo(db)
	inventoryRepo := postgres.NewInventoryRepo(db)
	orderRepo := postgres.NewOrderRepo(db)
	taskRepo := postgres.NewTaskRepo(db)

	// Initialize services.
	warehouseSvc := service.NewWarehouseService(warehouseRepo)
	skuSvc := service.NewSKUService(inventoryRepo)
	inventorySvc := service.NewInventoryService(inventoryRepo)
	orderSvc := service.NewOrderService(orderRepo)
	taskSvc := service.NewTaskService(taskRepo)

	// Initialize API handlers.
	warehouseHandler := api.NewWarehouseHandler(warehouseSvc, log.Logger)
	skuHandler := api.NewSKUHandler(skuSvc, log.Logger)
	inventoryHandler := api.NewInventoryHandler(inventorySvc, log.Logger)
	orderHandler := api.NewOrderHandler(orderSvc, log.Logger)
	taskHandler := api.NewTaskHandler(taskSvc, log.Logger)

	// Initialize API router with Go 1.22+ enhanced routing.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"admin","version":"0.1.0"}`)
	})
	api.RegisterWarehouseRoutes(mux, warehouseHandler)
	api.RegisterSKURoutes(mux, skuHandler)
	api.RegisterInventoryRoutes(mux, inventoryHandler)
	api.RegisterOrderRoutes(mux, orderHandler)
	api.RegisterTaskRoutes(mux, taskHandler)

	// Apply middleware stack: RequestID → Recovery → Logger → CORS → handler.
	handler := middleware.RequestID(
		middleware.Recovery(log.Logger)(
			middleware.Logger(log.Logger)(
				middleware.CORS(corsConfig)(mux),
			),
		),
	)

	srv := &http.Server{
		Addr:         ":" + cfg.AdminPort,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown.
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Info("Shutting down admin service...",
			slog.String("service", "admin"),
			slog.String("port", cfg.AdminPort))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout())
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("Admin service forced to shutdown", slog.String("error", err.Error()))
		}
		log.Info("Admin service stopped", slog.String("service", "admin"))
	}()

	log.Info("Admin service starting",
		slog.String("service", "admin"),
		slog.String("port", cfg.AdminPort),
		slog.String("health_url", fmt.Sprintf("http://localhost:%s/health", cfg.AdminPort)),
	)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("Admin service failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
