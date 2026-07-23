package domain

import (
	"time"
)

// AppConfig represents the application-level system configuration stored as a
// single-row JSONB document in the database. Settings managed here include the
// site name, default warehouse, inventory thresholds, and pagination defaults.
type AppConfig struct {
	SiteName            string `json:"site_name"`
	DefaultWarehouseID  string `json:"default_warehouse_id"`
	LowStockThreshold   int    `json:"low_stock_threshold"`
	DefaultPageSize     int    `json:"default_page_size"`
	JWTAccessTTL        int    `json:"jwt_access_ttl"`
}

// AppConfigRow is the database row representation with timestamps.
type AppConfigRow struct {
	Config    AppConfig `json:"config"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DefaultAppConfig returns sensible defaults used when no configuration exists
// (seeded by migration 000006).
func DefaultAppConfig() AppConfig {
	return AppConfig{
		SiteName:           "AI-WMS",
		DefaultWarehouseID: "",
		LowStockThreshold:  10,
		DefaultPageSize:    20,
		JWTAccessTTL:       3600,
	}
}
