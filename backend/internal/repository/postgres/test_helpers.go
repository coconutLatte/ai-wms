// Package postgres implements repository interfaces using PostgreSQL with pgx/v5.
package postgres

import (
	"github.com/ai-wms/ai-wms/backend/pkg/config"
)

// testConfig returns a minimal config suitable for integration tests.
// Prefers TEST_DATABASE_URL env var, falling back to local dev defaults.
func testConfig() *config.Config {
	// Use complete config struct — NewDB reads pool sizing from it.
	cfg := &config.Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "wms",
		DBPassword: "wms_dev_2026",
		DBName:     "wms",
		DBSSLMode:  "disable",
		DBMaxConns: 5,
		DBMinConns: 1,
	}

	return cfg
}
