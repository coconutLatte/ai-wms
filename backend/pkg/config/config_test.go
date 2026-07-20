package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	// Clear relevant env vars to test defaults.
	for _, key := range []string{
		"ADMIN_PORT", "PDA_PORT",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"DB_MAX_CONNS", "DB_MIN_CONNS",
		"REDIS_HOST", "REDIS_PORT", "REDIS_DB",
		"LOG_LEVEL", "ENV",
		"JWT_SECRET", "JWT_ACCESS_TTL_MINUTES", "JWT_REFRESH_TTL_HOURS",
	} {
		os.Unsetenv(key)
	}

	cfg := Load()

	if cfg.AdminPort != "8080" {
		t.Errorf("expected AdminPort=8080, got %s", cfg.AdminPort)
	}
	if cfg.PDAPort != "8081" {
		t.Errorf("expected PDAPort=8081, got %s", cfg.PDAPort)
	}
	if cfg.DBHost != "localhost" {
		t.Errorf("expected DBHost=localhost, got %s", cfg.DBHost)
	}
	if cfg.DBMaxConns != 20 {
		t.Errorf("expected DBMaxConns=20, got %d", cfg.DBMaxConns)
	}
	if cfg.DBMinConns != 2 {
		t.Errorf("expected DBMinConns=2, got %d", cfg.DBMinConns)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected LogLevel=info, got %s", cfg.LogLevel)
	}
	if cfg.Env != "development" {
		t.Errorf("expected Env=development, got %s", cfg.Env)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("ADMIN_PORT", "9090")
	os.Setenv("PDA_PORT", "9091")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("DB_MAX_CONNS", "50")
	os.Setenv("DB_MIN_CONNS", "5")
	os.Setenv("REDIS_HOST", "redis.example.com")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("REDIS_DB", "1")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("ENV", "production")
	defer func() {
		for _, key := range []string{
			"ADMIN_PORT", "PDA_PORT",
			"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
			"DB_MAX_CONNS", "DB_MIN_CONNS",
			"REDIS_HOST", "REDIS_PORT", "REDIS_DB",
			"LOG_LEVEL", "ENV",
				"JWT_SECRET", "JWT_ACCESS_TTL_MINUTES", "JWT_REFRESH_TTL_HOURS",
		} {
			os.Unsetenv(key)
		}
	}()

	cfg := Load()

	if cfg.AdminPort != "9090" {
		t.Errorf("expected AdminPort=9090, got %s", cfg.AdminPort)
	}
	if cfg.DBMaxConns != 50 {
		t.Errorf("expected DBMaxConns=50, got %d", cfg.DBMaxConns)
	}
	if cfg.DBMinConns != 5 {
		t.Errorf("expected DBMinConns=5, got %d", cfg.DBMinConns)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel=debug, got %s", cfg.LogLevel)
	}
	if cfg.Env != "production" {
		t.Errorf("expected Env=production, got %s", cfg.Env)
	}
}

func TestDSN(t *testing.T) {
	cfg := &Config{
		DBUser:     "wms",
		DBPassword: "secret",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBName:     "wms",
		DBSSLMode:  "disable",
	}

	dsn := cfg.DSN()
	expected := "postgres://wms:secret@localhost:5432/wms?sslmode=disable"
	if dsn != expected {
		t.Errorf("expected DSN %q, got %q", expected, dsn)
	}
}

func TestRedisAddr(t *testing.T) {
	cfg := &Config{
		RedisHost: "redis.local",
		RedisPort: "6379",
	}

	addr := cfg.RedisAddr()
	if addr != "redis.local:6379" {
		t.Errorf("expected redis.local:6379, got %s", addr)
	}
}

func TestIsDevelopment(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"development", true},
		{"production", false},
		{"staging", false},
	}
	for _, tt := range tests {
		cfg := &Config{Env: tt.env}
		if got := cfg.IsDevelopment(); got != tt.want {
			t.Errorf("IsDevelopment() with env=%s: got %v, want %v", tt.env, got, tt.want)
		}
	}
}

func TestIsProduction(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{"development", false},
		{"production", true},
		{"staging", false},
	}
	for _, tt := range tests {
		cfg := &Config{Env: tt.env}
		if got := cfg.IsProduction(); got != tt.want {
			t.Errorf("IsProduction() with env=%s: got %v, want %v", tt.env, got, tt.want)
		}
	}
}

func TestTimeout(t *testing.T) {
	dev := &Config{Env: "development"}
	prod := &Config{Env: "production"}

	if dev.Timeout().Seconds() != 60 {
		t.Errorf("expected dev timeout = 60s, got %v", dev.Timeout())
	}
	if prod.Timeout().Seconds() != 30 {
		t.Errorf("expected prod timeout = 30s, got %v", prod.Timeout())
	}
}

func TestValidate_Success(t *testing.T) {
	cfg := &Config{
		AdminPort:  "8080",
		PDAPort:    "8081",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "wms",
		DBName:     "wms",
		DBMaxConns: 20,
		DBMinConns: 2,
		LogLevel:   "info",
		JWTSecret:  "test-secret",
		JWTAccessTTL:  15 * time.Minute,
		JWTRefreshTTL: 7 * 24 * time.Hour,
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected validation to pass, got: %v", err)
	}
}

func TestValidate_Errors(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		errText string
	}{
		{
			name:    "missing admin port",
			cfg:     &Config{AdminPort: "", PDAPort: "8081", DBHost: "h", DBPort: "5432", DBUser: "u", DBName: "n", DBMaxConns: 20, DBMinConns: 0, LogLevel: "info", JWTSecret: "s", JWTAccessTTL: time.Minute, JWTRefreshTTL: time.Hour},
			errText: "ADMIN_PORT is required",
		},
		{
			name:    "missing pda port",
			cfg:     &Config{AdminPort: "8080", PDAPort: "", DBHost: "h", DBPort: "5432", DBUser: "u", DBName: "n", DBMaxConns: 20, DBMinConns: 0, LogLevel: "info", JWTSecret: "s", JWTAccessTTL: time.Minute, JWTRefreshTTL: time.Hour},
			errText: "PDA_PORT is required",
		},
		{
			name:    "invalid admin port",
			cfg:     &Config{AdminPort: "99999", PDAPort: "8081", DBHost: "h", DBPort: "5432", DBUser: "u", DBName: "n", DBMaxConns: 20, DBMinConns: 0, LogLevel: "info", JWTSecret: "s", JWTAccessTTL: time.Minute, JWTRefreshTTL: time.Hour},
			errText: "ADMIN_PORT",
		},
		{
			name:    "invalid db port",
			cfg:     &Config{AdminPort: "8080", PDAPort: "8081", DBHost: "h", DBPort: "abc", DBUser: "u", DBName: "n", DBMaxConns: 20, DBMinConns: 0, LogLevel: "info", JWTSecret: "s", JWTAccessTTL: time.Minute, JWTRefreshTTL: time.Hour},
			errText: "DB_PORT",
		},
		{
			name:    "max conns < 1",
			cfg:     &Config{AdminPort: "8080", PDAPort: "8081", DBHost: "h", DBPort: "5432", DBUser: "u", DBName: "n", DBMaxConns: 0, DBMinConns: 0, LogLevel: "info", JWTSecret: "s", JWTAccessTTL: time.Minute, JWTRefreshTTL: time.Hour},
			errText: "DB_MAX_CONNS",
		},
		{
			name:    "min conns > max conns",
			cfg:     &Config{AdminPort: "8080", PDAPort: "8081", DBHost: "h", DBPort: "5432", DBUser: "u", DBName: "n", DBMaxConns: 10, DBMinConns: 20, LogLevel: "info", JWTSecret: "s", JWTAccessTTL: time.Minute, JWTRefreshTTL: time.Hour},
			errText: "DB_MIN_CONNS",
		},
		{
			name:    "min conns negative",
			cfg:     &Config{AdminPort: "8080", PDAPort: "8081", DBHost: "h", DBPort: "5432", DBUser: "u", DBName: "n", DBMaxConns: 10, DBMinConns: -1, LogLevel: "info", JWTSecret: "s", JWTAccessTTL: time.Minute, JWTRefreshTTL: time.Hour},
			errText: "DB_MIN_CONNS",
		},
		{
			name:    "invalid log level",
			cfg:     &Config{AdminPort: "8080", PDAPort: "8081", DBHost: "h", DBPort: "5432", DBUser: "u", DBName: "n", DBMaxConns: 20, DBMinConns: 0, LogLevel: "trace", JWTSecret: "s", JWTAccessTTL: time.Minute, JWTRefreshTTL: time.Hour},
			errText: "LOG_LEVEL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errText) {
				t.Errorf("expected error containing %q, got %q", tt.errText, err.Error())
			}
		})
	}
}

func TestGetEnvInt_Invalid(t *testing.T) {
	os.Setenv("TEST_INT_INVALID", "notanumber")
	defer os.Unsetenv("TEST_INT_INVALID")

	result := getEnvInt("TEST_INT_INVALID", 42)
	if result != 42 {
		t.Errorf("expected fallback to 42, got %d", result)
	}
}

func TestGetEnvInt_Default(t *testing.T) {
	os.Unsetenv("TEST_INT_MISSING")

	result := getEnvInt("TEST_INT_MISSING", 10)
	if result != 10 {
		t.Errorf("expected default 10, got %d", result)
	}
}
