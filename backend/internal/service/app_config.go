package service

import (
	"context"
	"fmt"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// AppConfigService orchestrates business logic for system configuration.
type AppConfigService struct {
	repo repository.AppConfigRepository
}

// NewAppConfigService creates a new AppConfigService.
func NewAppConfigService(repo repository.AppConfigRepository) *AppConfigService {
	return &AppConfigService{repo: repo}
}

// GetConfig retrieves the current system configuration.
func (s *AppConfigService) GetConfig(ctx context.Context) (*domain.AppConfigRow, error) {
	row, err := s.repo.GetAppConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("app config service: get: %w", err)
	}
	return row, nil
}

// UpdateConfigInput is the input for updating system configuration.
type UpdateConfigInput struct {
	SiteName           *string `json:"site_name,omitempty"`
	DefaultWarehouseID *string `json:"default_warehouse_id,omitempty"`
	LowStockThreshold  *int    `json:"low_stock_threshold,omitempty"`
	DefaultPageSize    *int    `json:"default_page_size,omitempty"`
	JWTAccessTTL       *int    `json:"jwt_access_ttl,omitempty"`
}

// Validate checks the input for business rule violations.
func (in *UpdateConfigInput) Validate() error {
	if in.SiteName != nil && *in.SiteName == "" {
		return pkgerrors.NewInvalidInput("site_name must not be empty")
	}
	if in.LowStockThreshold != nil && *in.LowStockThreshold < 0 {
		return pkgerrors.NewInvalidInput("low_stock_threshold must be >= 0")
	}
	if in.DefaultPageSize != nil && (*in.DefaultPageSize < 1 || *in.DefaultPageSize > 100) {
		return pkgerrors.NewInvalidInput("default_page_size must be between 1 and 100")
	}
	if in.JWTAccessTTL != nil && *in.JWTAccessTTL < 60 {
		return pkgerrors.NewInvalidInput("jwt_access_ttl must be >= 60 seconds")
	}
	return nil
}

// UpdateConfig applies partial updates to the system configuration.
// Only non-nil fields in the input are applied; nil fields keep their current
// values.
func (s *AppConfigService) UpdateConfig(ctx context.Context, input UpdateConfigInput) (*domain.AppConfigRow, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	current, err := s.repo.GetAppConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("app config service: update: %w", err)
	}

	cfg := current.Config

	if input.SiteName != nil {
		cfg.SiteName = *input.SiteName
	}
	if input.DefaultWarehouseID != nil {
		cfg.DefaultWarehouseID = *input.DefaultWarehouseID
	}
	if input.LowStockThreshold != nil {
		cfg.LowStockThreshold = *input.LowStockThreshold
	}
	if input.DefaultPageSize != nil {
		cfg.DefaultPageSize = *input.DefaultPageSize
	}
	if input.JWTAccessTTL != nil {
		cfg.JWTAccessTTL = *input.JWTAccessTTL
	}

	if err := s.repo.UpdateAppConfig(ctx, cfg); err != nil {
		return nil, fmt.Errorf("app config service: update: %w", err)
	}

	return &domain.AppConfigRow{Config: cfg}, nil
}
