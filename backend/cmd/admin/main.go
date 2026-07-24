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
	warehouseRepo := postgres.NewWarehouseRepo(db)
	inventoryRepo := postgres.NewInventoryRepo(db)
	orderRepo := postgres.NewOrderRepo(db)
	taskRepo := postgres.NewTaskRepo(db)
	userRepo := postgres.NewUserRepo(db)
	tokenBLRepo := postgres.NewTokenBlacklistRepo(db)
	cycleCountRepo := postgres.NewCycleCountRepo(db)
	shipmentRepo := postgres.NewShipmentRepo(db)
	appConfigRepo := postgres.NewAppConfigRepo(db)

	// Initialize transaction manager for atomic multi-step operations.
	txManager := postgres.NewTxManager(db)

	// Initialize services.
	warehouseSvc := service.NewWarehouseService(warehouseRepo)
	skuSvc := service.NewSKUService(inventoryRepo)
	inventorySvc := service.NewInventoryServiceWithTx(inventoryRepo, txManager)
	orderSvc := service.NewOrderServiceWithTx(orderRepo, taskRepo, inventoryRepo, txManager)
	taskSvc := service.NewTaskServiceWithTx(taskRepo, inventoryRepo, txManager)
	waveSvc := service.NewWaveService(taskRepo)
	authSvc := service.NewAuthServiceWithBlacklist(userRepo, tokenBLRepo, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	auditLogSvc := service.NewAuditLogService(userRepo)
	userSvc := service.NewUserService(userRepo)
	roleSvc := service.NewRoleService(userRepo)
	cycleCountSvc := service.NewCycleCountService(cycleCountRepo, inventoryRepo, warehouseRepo, txManager)
	shipmentSvc := service.NewShipmentService(shipmentRepo, orderRepo)
	appConfigSvc := service.NewAppConfigService(appConfigRepo)

	// Initialize API handlers.
	warehouseHandler := api.NewWarehouseHandler(warehouseSvc, log.Logger)
	skuHandler := api.NewSKUHandler(skuSvc, log.Logger)
	inventoryHandler := api.NewInventoryHandler(inventorySvc, log.Logger)
	orderHandler := api.NewOrderHandler(orderSvc, log.Logger)
	taskHandler := api.NewTaskHandler(taskSvc, log.Logger)
	waveHandler := api.NewWaveHandler(waveSvc, log.Logger)
	authHandler := api.NewAuthHandler(authSvc, log.Logger)
	auditLogHandler := api.NewAuditLogHandler(auditLogSvc, log.Logger)
	userHandler := api.NewUserHandler(userSvc, log.Logger)
	roleHandler := api.NewRoleHandler(roleSvc, log.Logger)
	dashboardHandler := api.NewDashboardHandler(warehouseSvc, skuSvc, inventorySvc, orderSvc, taskSvc, log.Logger)
	cycleCountHandler := api.NewCycleCountHandler(cycleCountSvc, log.Logger)
	shipmentHandler := api.NewShipmentHandler(shipmentSvc, log.Logger)
	appConfigHandler := api.NewAppConfigHandler(appConfigSvc, log.Logger)

	// ── Route Setup ──────────────────────────────────────────────────────────

	mux := http.NewServeMux()

	// Health check (no auth required).
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"admin","version":"0.1.0"}`)
	})

	// Readiness probe — pings PostgreSQL and Redis, returns per-service status.
	healthHandler := api.NewHealthHandler(db, redisClient, log.Logger)
	mux.HandleFunc("GET /ready", healthHandler.Ready)

	// Swagger API docs (no auth required).
	api.RegisterSwaggerRoutes(mux)

	// Auth routes (no auth required — these are the login/refresh endpoints).
	api.RegisterAuthRoutes(mux, authHandler)

	// Protected routes — wrapped in JWT auth middleware.
	// Go 1.22+ ServeMux: exact method+path patterns take priority over prefix patterns.
	// The auth routes above are exact matches and bypass the middleware.
	// Everything else under /api/v1/ goes through the auth middleware.
	protected := http.NewServeMux()
	api.RegisterWarehouseRoutes(protected, warehouseHandler)
	api.RegisterSKURoutes(protected, skuHandler)
	api.RegisterInventoryRoutes(protected, inventoryHandler)
	api.RegisterOrderRoutes(protected, orderHandler)
	api.RegisterASNRoutes(protected, orderHandler)
	api.RegisterTaskRoutes(protected, taskHandler)
	api.RegisterWaveRoutes(protected, waveHandler)
	api.RegisterDashboardRoute(protected, dashboardHandler)
	api.RegisterShipmentRoutes(protected, shipmentHandler)

	// Admin-only routes (require auth + admin role).
	adminOnly := http.NewServeMux()
	api.RegisterAuditLogRoutes(adminOnly, auditLogHandler)
	api.RegisterUserRoutes(adminOnly, userHandler)
	api.RegisterRoleRoutes(adminOnly, roleHandler)
	api.RegisterCycleCountRoutes(adminOnly, cycleCountHandler)
	api.RegisterAppConfigRoutes(adminOnly, appConfigHandler)

	authMiddleware := middleware.Auth(cfg.JWTSecret)
	adminHandler := authMiddleware(middleware.RequireRole("admin")(adminOnly))

	// Register admin-only paths first (exact prefix patterns take priority).
	mux.Handle("/api/v1/audit-logs", adminHandler)
	mux.Handle("/api/v1/audit-logs/", adminHandler)
	mux.Handle("/api/v1/users", adminHandler)
	mux.Handle("/api/v1/users/", adminHandler)
	mux.Handle("/api/v1/roles", adminHandler)
	mux.Handle("/api/v1/roles/", adminHandler)
	mux.Handle("/api/v1/settings", adminHandler)

	// All other /api/v1/ routes require auth but not admin role.
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
