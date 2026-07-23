package repository

import (
	"context"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
)

// AppConfigRepository defines the data access interface for system configuration.
type AppConfigRepository interface {
	// GetAppConfig retrieves the current system configuration.
	// Returns the default config if no configuration has been saved yet.
	GetAppConfig(ctx context.Context) (*domain.AppConfigRow, error)

	// UpdateAppConfig replaces the system configuration with the provided config.
	UpdateAppConfig(ctx context.Context, config domain.AppConfig) error
}
