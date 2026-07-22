// PDA Service — serves the PDA REST API for warehouse operators.
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
	"github.com/ai-wms/ai-wms/backend/pkg/redis"
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
		slog.String("pda_port", cfg.PDAPort),
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

	// Run database migrations (idempotent — only unapplied migrations execute).
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := db.RunMigrationsFromDir(context.Background(), migrationsDir); err != nil {
		log.Error("Failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize Redis connection (non-fatal if unavailable in development).
	redisClient, err := redis.New(context.Background(), cfg)
	if err != nil {
		if cfg.IsProduction() {
			log.Error("Failed to connect to Redis (required in production)", slog.String("error", err.Error()))
			os.Exit(1)
		}
		log.Warn("Redis not available — continuing without cache (development mode)",
			slog.String("error", err.Error()),
			slog.String("redis_addr", cfg.RedisAddr()),
		)
	} else {
		defer redisClient.Close()
	}

	// Initialize repositories.
	taskRepo := postgres.NewTaskRepo(db)
	inventoryRepo := postgres.NewInventoryRepo(db)
	warehouseRepo := postgres.NewWarehouseRepo(db)
	userRepo := postgres.NewUserRepo(db)
	tokenBLRepo := postgres.NewTokenBlacklistRepo(db)

	// Initialize transaction manager for atomic multi-step operations.
	txManager := postgres.NewTxManager(db)

	// Initialize services.
	taskSvc := service.NewTaskServiceWithTx(taskRepo, inventoryRepo, txManager)
	warehouseSvc := service.NewWarehouseService(warehouseRepo)
	skuSvc := service.NewSKUService(inventoryRepo)
	authSvc := service.NewAuthServiceWithBlacklist(userRepo, tokenBLRepo, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	// Initialize API handlers.
	authHandler := api.NewAuthHandler(authSvc, log.Logger)
	taskHandler := api.NewTaskHandler(taskSvc, log.Logger)
	warehouseHandler := api.NewWarehouseHandler(warehouseSvc, log.Logger)
	skuHandler := api.NewSKUHandler(skuSvc, log.Logger)

	// ── Route Setup ──────────────────────────────────────────────────────────

	mux := http.NewServeMux()

	// Health check (no auth required).
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"pda","version":"0.1.0"}`)
	})

	// Auth routes (no auth required — login/refresh/logout endpoints).
	api.RegisterAuthRoutes(mux, authHandler)

	// Protected routes — wrapped in JWT auth middleware.
	protected := http.NewServeMux()
	api.RegisterTaskRoutes(protected, taskHandler)

	// Barcode lookup routes for the PDA scanner (location barcode → putaway, SKU barcode → inventory).
	api.RegisterWarehouseRoutes(protected, warehouseHandler)
	api.RegisterSKURoutes(protected, skuHandler)

	authMiddleware := middleware.Auth(cfg.JWTSecret)
	mux.Handle("/api/v1/", authMiddleware(protected))

	// ── Global Middleware Stack ───────────────────────────────────────────────

	// RequestID → Recovery → Logger → CORS → mux.
	handler := middleware.RequestID(
		middleware.Recovery(log.Logger)(
			middleware.Logger(log.Logger)(
				middleware.CORS(corsConfig)(mux),
			),
		),
	)

	srv := &http.Server{
		Addr:         ":" + cfg.PDAPort,
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

		log.Info("Shutting down PDA service...",
			slog.String("service", "pda"),
			slog.String("port", cfg.PDAPort))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout())
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("PDA service forced to shutdown", slog.String("error", err.Error()))
		}
		log.Info("PDA service stopped", slog.String("service", "pda"))
	}()

	log.Info("PDA service starting",
		slog.String("service", "pda"),
		slog.String("port", cfg.PDAPort),
		slog.String("health_url", fmt.Sprintf("http://localhost:%s/health", cfg.PDAPort)),
	)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("PDA service failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
