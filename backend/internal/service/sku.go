// Package service implements business logic orchestration for the WMS domain.
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// SKUService orchestrates business logic for SKUs.
type SKUService struct {
	repo repository.InventoryRepository
}

// NewSKUService creates a new SKUService.
func NewSKUService(repo repository.InventoryRepository) *SKUService {
	return &SKUService{repo: repo}
}

// ── Input Types ──────────────────────────────────────────────────────────────────────────

// CreateSKUInput is the input for creating a new SKU.
type CreateSKUInput struct {
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Barcode     string           `json:"barcode,omitempty"`
	UOM         domain.UOM       `json:"uom"`
	Attributes  domain.Attributes `json:"attributes,omitempty"`
	Category    string           `json:"category,omitempty"`
}

// Validate checks the input for business rule violations.
func (in *CreateSKUInput) Validate() error {
	if in.Code == "" {
		return pkgerrors.NewInvalidInput("sku code is required")
	}
	if in.Name == "" {
		return pkgerrors.NewInvalidInput("sku name is required")
	}
	if in.UOM.BaseUnit == "" {
		return pkgerrors.NewInvalidInput("sku base unit is required")
	}
	return nil
}

// UpdateSKUInput is the input for updating an existing SKU.
// All fields are optional — only non-nil fields are applied.
type UpdateSKUInput struct {
	Name        *string            `json:"name,omitempty"`
	Description *string            `json:"description,omitempty"`
	Barcode     *string            `json:"barcode,omitempty"`
	UOM         *domain.UOM        `json:"uom,omitempty"`
	Attributes  *domain.Attributes `json:"attributes,omitempty"`
	Category    *string            `json:"category,omitempty"`
	Status      *domain.SKUStatus  `json:"status,omitempty"`
}

// ── Service Methods ──────────────────────────────────────────────────────────────────────

// CreateSKU validates input and creates a new SKU.
func (s *SKUService) CreateSKU(ctx context.Context, input CreateSKUInput) (*domain.SKU, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Normalize attributes: ensure it is non-nil so JSON marshalling works cleanly.
	attrs := input.Attributes
	if attrs == nil {
		attrs = make(domain.Attributes)
	}

	sku := &domain.SKU{
		Code:        input.Code,
		Name:        input.Name,
		Description: input.Description,
		Barcode:     input.Barcode,
		UOM:         input.UOM,
		Attributes:  cloneAttributes(attrs),
		Category:    input.Category,
		Status:      domain.SKUStatusActive,
	}

	if err := s.repo.CreateSKU(ctx, sku); err != nil {
		return nil, fmt.Errorf("sku service: create: %w", err)
	}

	return sku, nil
}

// GetSKU retrieves a SKU by ID.
func (s *SKUService) GetSKU(ctx context.Context, id uuid.UUID) (*domain.SKU, error) {
	sku, err := s.repo.GetSKU(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("sku service: get %s: %w", id, err)
	}
	return sku, nil
}

// GetSKUByCode retrieves a SKU by its unique code.
func (s *SKUService) GetSKUByCode(ctx context.Context, code string) (*domain.SKU, error) {
	sku, err := s.repo.GetSKUByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("sku service: get by code %s: %w", code, err)
	}
	return sku, nil
}

// ListSKUs returns all SKUs with pagination support.
func (s *SKUService) ListSKUs(ctx context.Context, limit, offset int) ([]*domain.SKU, int, error) {
	skus, err := s.repo.ListSKUs(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("sku service: list: %w", err)
	}
	total, err := s.repo.CountSKUs(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("sku service: count: %w", err)
	}
	return skus, total, nil
}

// CountSKUs returns the total number of SKUs.
func (s *SKUService) CountSKUs(ctx context.Context) (int, error) {
	total, err := s.repo.CountSKUs(ctx)
	if err != nil {
		return 0, fmt.Errorf("sku service: count: %w", err)
	}
	return total, nil
}

// UpdateSKU validates and partially updates an existing SKU.
func (s *SKUService) UpdateSKU(ctx context.Context, id uuid.UUID, input UpdateSKUInput) (*domain.SKU, error) {
	sku, err := s.repo.GetSKU(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("sku service: update %s: %w", id, err)
	}

	if input.Name != nil {
		sku.Name = *input.Name
	}
	if input.Description != nil {
		sku.Description = *input.Description
	}
	if input.Barcode != nil {
		sku.Barcode = *input.Barcode
	}
	if input.UOM != nil {
		sku.UOM = *input.UOM
	}
	if input.Attributes != nil {
		sku.Attributes = cloneAttributes(*input.Attributes)
	}
	if input.Category != nil {
		sku.Category = *input.Category
	}
	if input.Status != nil {
		status := *input.Status
		if !isValidSKUStatus(status) {
			return nil, pkgerrors.NewInvalidInput(fmt.Sprintf("invalid sku status: %s", status))
		}
		sku.Status = status
	}

	if err := s.repo.UpdateSKU(ctx, sku); err != nil {
		return nil, fmt.Errorf("sku service: update %s: %w", id, err)
	}

	return sku, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────────────────

func isValidSKUStatus(s domain.SKUStatus) bool {
	switch s {
	case domain.SKUStatusActive, domain.SKUStatusInactive, domain.SKUStatusDiscontinued:
		return true
	}
	return false
}

// cloneAttributes deep-copies an Attributes map. The maps are always string→string so a
// shallow copy is safe, but a deep copy is defensive against concurrent mutation.
func cloneAttributes(attrs domain.Attributes) domain.Attributes {
	if attrs == nil {
		return make(domain.Attributes)
	}
	clone := make(domain.Attributes, len(attrs))
	for k, v := range attrs {
		clone[k] = v
	}
	return clone
}
