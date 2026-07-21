// Package redis provides Redis client bootstrap and connection management
// for the WMS application. It wraps go-redis/v9 with configuration-driven
// initialization and health checking, following the same pattern as the
// PostgreSQL DB bootstrap in internal/repository/postgres/.
package redis

import (
	"context"
	"fmt"
	"log/slog"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ai-wms/ai-wms/backend/pkg/config"
)

// Client wraps a go-redis client with connection management.
// It follows the same pattern as postgres.DB — a thin wrapper that
// provides bootstrap, ping, and graceful close.
type Client struct {
	*goredis.Client
}

// New creates a new Redis client from application configuration.
// It connects to Redis, verifies connectivity with PING, and logs
// the connection details. Returns an error if the connection fails.
func New(ctx context.Context, cfg *config.Config) (*Client, error) {
	opts := &goredis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	client := goredis.NewClient(opts)

	// Verify connectivity — fail fast if Redis is unreachable.
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	slog.Info("Redis connection established",
		slog.String("addr", cfg.RedisAddr()),
		slog.Int("db", cfg.RedisDB),
	)
	return &Client{Client: client}, nil
}

// Close gracefully shuts down the Redis client, waiting for pending
// commands to complete.
func (c *Client) Close() error {
	if err := c.Client.Close(); err != nil {
		return fmt.Errorf("close redis: %w", err)
	}
	slog.Info("Redis connection closed")
	return nil
}

// Ping verifies Redis connectivity. Useful for health checks.
func (c *Client) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}
