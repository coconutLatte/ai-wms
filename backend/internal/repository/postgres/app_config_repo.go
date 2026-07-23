package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
)

// AppConfigRepo implements repository.AppConfigRepository using PostgreSQL.
type AppConfigRepo struct {
	db *DB
}

// NewAppConfigRepo creates a new AppConfigRepo.
func NewAppConfigRepo(db *DB) *AppConfigRepo {
	return &AppConfigRepo{db: db}
}

// GetAppConfig retrieves the current system configuration from the single-row
// app_config table. Returns the default config if the row does not exist yet.
func (r *AppConfigRepo) GetAppConfig(ctx context.Context) (*domain.AppConfigRow, error) {
	const query = `SELECT config, updated_at FROM app_config WHERE id = 1`

	row := &domain.AppConfigRow{
		Config: domain.DefaultAppConfig(),
	}

	var raw []byte
	err := r.db.Pool.QueryRow(ctx, query).Scan(&raw, &row.UpdatedAt)
	if err != nil {
		// If the table is empty (e.g. before seed runs), return defaults.
		return row, nil
	}

	if err := json.Unmarshal(raw, &row.Config); err != nil {
		return nil, fmt.Errorf("get app config: unmarshal: %w", err)
	}

	return row, nil
}

// UpdateAppConfig replaces the system configuration. Uses upsert semantics.
func (r *AppConfigRepo) UpdateAppConfig(ctx context.Context, config domain.AppConfig) error {
	raw, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("update app config: marshal: %w", err)
	}

	now := time.Now()
	const query = `
		INSERT INTO app_config (id, config, updated_at)
		VALUES (1, $1, $2)
		ON CONFLICT (id) DO UPDATE SET config = $1, updated_at = $2`

	if _, err := r.db.Pool.Exec(ctx, query, raw, now); err != nil {
		return fmt.Errorf("update app config: %w", err)
	}
	return nil
}
