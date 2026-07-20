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

	// Redis
	RedisHost string
	RedisPort string
	RedisDB   int

	// Application
	LogLevel string
	Env      string
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
		RedisHost:  getEnv("REDIS_HOST", "localhost"),
		RedisPort:  getEnv("REDIS_PORT", "6379"),
		RedisDB:    getEnvInt("REDIS_DB", 0),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		Env:        getEnv("ENV", "development"),
	}
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
