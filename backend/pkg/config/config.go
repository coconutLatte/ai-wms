// Package config provides application configuration loaded from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	// Server
	AdminPort string
	PDAPort   string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DBMaxConns int32
	DBMinConns int32

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// Application
	LogLevel    string
	Env         string
	JWTSecret   string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		AdminPort:  getEnv("ADMIN_PORT", "8080"),
		PDAPort:    getEnv("PDA_PORT", "8081"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "wms"),
		DBPassword: getEnv("DB_PASSWORD", "wms_dev_2026"),
		DBName:     getEnv("DB_NAME", "wms"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		DBMaxConns: int32(getEnvInt("DB_MAX_CONNS", 20)),
		DBMinConns: int32(getEnvInt("DB_MIN_CONNS", 2)),
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		Env:        getEnv("ENV", "development"),
		JWTSecret:  getEnv("JWT_SECRET", "change-me-in-production-REPLACE-ME"),
		JWTAccessTTL:  time.Duration(getEnvInt("JWT_ACCESS_TTL_MINUTES", 15)) * time.Minute,
		JWTRefreshTTL: time.Duration(getEnvInt("JWT_REFRESH_TTL_HOURS", 168)) * time.Hour, // 7 days
	}
}

// Validate checks that all required configuration fields are set correctly.
// Returns nil if the configuration is valid, or an error describing the first
// problem found. This should be called early during application startup so
// misconfiguration fails fast with a clear message.
func (c *Config) Validate() error {
	if c.AdminPort == "" {
		return fmt.Errorf("ADMIN_PORT is required")
	}
	if c.PDAPort == "" {
		return fmt.Errorf("PDA_PORT is required")
	}
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.DBPort == "" {
		return fmt.Errorf("DB_PORT is required")
	}
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.DBMaxConns < 1 {
		return fmt.Errorf("DB_MAX_CONNS must be >= 1, got %d", c.DBMaxConns)
	}
	if c.DBMinConns < 0 {
		return fmt.Errorf("DB_MIN_CONNS must be >= 0, got %d", c.DBMinConns)
	}
	if c.DBMinConns > c.DBMaxConns {
		return fmt.Errorf("DB_MIN_CONNS (%d) must not exceed DB_MAX_CONNS (%d)", c.DBMinConns, c.DBMaxConns)
	}

	// Validate port numbers are reasonable.
	adminPort, err := strconv.Atoi(c.AdminPort)
	if err != nil || adminPort < 1 || adminPort > 65535 {
		return fmt.Errorf("ADMIN_PORT must be a valid port number (1-65535), got %q", c.AdminPort)
	}
	pdaPort, err := strconv.Atoi(c.PDAPort)
	if err != nil || pdaPort < 1 || pdaPort > 65535 {
		return fmt.Errorf("PDA_PORT must be a valid port number (1-65535), got %q", c.PDAPort)
	}
	dbPort, err := strconv.Atoi(c.DBPort)
	if err != nil || dbPort < 1 || dbPort > 65535 {
		return fmt.Errorf("DB_PORT must be a valid port number (1-65535), got %q", c.DBPort)
	}

	// Validate log level.
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("LOG_LEVEL must be one of [debug, info, warn, error], got %q", c.LogLevel)
	}

	// Validate JWT configuration.
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.JWTSecret == "change-me-in-production-REPLACE-ME" && c.IsProduction() {
		return fmt.Errorf("JWT_SECRET must be changed from the default in production")
	}
	if c.JWTAccessTTL <= 0 {
		return fmt.Errorf("JWT_ACCESS_TTL_MINUTES must be > 0, got %v", c.JWTAccessTTL)
	}
	if c.JWTRefreshTTL <= 0 {
		return fmt.Errorf("JWT_REFRESH_TTL_HOURS must be > 0, got %v", c.JWTRefreshTTL)
	}

	return nil
}

// DSN returns the PostgreSQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

// RedisAddr returns the Redis address.
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// Timeout returns the default operation timeout.
func (c *Config) Timeout() time.Duration {
	if c.IsProduction() {
		return 30 * time.Second
	}
	return 60 * time.Second
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
