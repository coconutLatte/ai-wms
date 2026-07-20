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

// InventoryService orchestrates business logic for inventory records.
type InventoryService struct {
	repo      repository.InventoryRepository
	txManager repository.TxManager
}

// NewInventoryService creates a new InventoryService.
func NewInventoryService(repo repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

// NewInventoryServiceWithTx creates a new InventoryService with transaction support.
// When txManager is provided, the service uses it to wrap multi-step inventory
// operations in atomic database transactions.
func NewInventoryServiceWithTx(repo repository.InventoryRepository, txManager repository.TxManager) *InventoryService {
	return &InventoryService{repo: repo, txManager: txManager}
}

// ── Input Types ──────────────────────────────────────────────────────────────────────────

// AdjustInventoryInput is the input for manually adjusting inventory quantity.
type AdjustInventoryInput struct {
	DeltaQty      float64 `json:"delta_qty"`      // Positive = increase, Negative = decrease
	ReferenceType string  `json:"reference_type"` // e.g. "adjustment", "cycle_count"
	ReferenceID   string  `json:"reference_id,omitempty"`
	Reason        string  `json:"reason,omitempty"` // Human-readable reason
	CreatedBy     string  `json:"created_by"`
}

// Validate checks the input for business rule violations.
func (in *AdjustInventoryInput) Validate() error {
	if in.DeltaQty == 0 {
		return pkgerrors.NewInvalidInput("delta_qty must not be zero")
	}
	if in.ReferenceType == "" {
		return pkgerrors.NewInvalidInput("reference_type is required")
	}
	if in.CreatedBy == "" {
		return pkgerrors.NewInvalidInput("created_by is required")
	}
	return nil
}

// CreateInventoryInput is the input for creating a new inventory record.
// This is typically called by other services (e.g., receiving).
type CreateInventoryInput struct {
	SKUID          uuid.UUID              `json:"sku_id"`
	LocationID     uuid.UUID              `json:"location_id"`
	WarehouseID    uuid.UUID              `json:"warehouse_id"`
	BatchNo        string                 `json:"batch_no,omitempty"`
	Qty            float64                `json:"qty"`
	ReservedQty    float64                `json:"reserved_qty"`
	Status         domain.InventoryStatus `json:"status,omitempty"`
	ProductionDate *string                `json:"production_date,omitempty"` // RFC 3339
	ExpiryDate     *string                `json:"expiry_date,omitempty"`     // RFC 3339
}

// Validate checks the input for business rule violations.
func (in *CreateInventoryInput) Validate() error {
	if in.SKUID == uuid.Nil {
		return pkgerrors.NewInvalidInput("sku_id is required")
	}
	if in.LocationID == uuid.Nil {
		return pkgerrors.NewInvalidInput("location_id is required")
	}
	if in.WarehouseID == uuid.Nil {
		return pkgerrors.NewInvalidInput("warehouse_id is required")
	}
	if in.Qty < 0 {
		return pkgerrors.NewInvalidInput("qty must not be negative")
	}
	return nil
}

// QueryInventoryInput is the input for querying inventory records.
type QueryInventoryInput struct {
	WarehouseID string                 `json:"warehouse_id,omitempty"`
	SKUID       string                 `json:"sku_id,omitempty"`
	LocationID  string                 `json:"location_id,omitempty"`
	BatchNo     string                 `json:"batch_no,omitempty"`
	Status      domain.InventoryStatus `json:"status,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Offset      int                    `json:"offset,omitempty"`
}

// ToFilter converts query input to a repository filter.
func (in *QueryInventoryInput) ToFilter() (repository.InventoryFilter, error) {
	f := repository.InventoryFilter{
		BatchNo: in.BatchNo,
		Status:  in.Status,
		Limit:   in.Limit,
		Offset:  in.Offset,
	}

	if in.WarehouseID != "" {
		id, err := uuid.Parse(in.WarehouseID)
		if err != nil {
			return f, pkgerrors.NewInvalidInput("invalid warehouse_id UUID")
		}
		f.WarehouseID = id
	}
	if in.SKUID != "" {
		id, err := uuid.Parse(in.SKUID)
		if err != nil {
			return f, pkgerrors.NewInvalidInput("invalid sku_id UUID")
		}
		f.SKUID = id
	}
	if in.LocationID != "" {
		id, err := uuid.Parse(in.LocationID)
		if err != nil {
			return f, pkgerrors.NewInvalidInput("invalid location_id UUID")
		}
		f.LocationID = id
	}

	return f, nil
}

// ── Service Methods ──────────────────────────────────────────────────────────────────────

// QueryInventory searches inventory records with optional filters.
func (s *InventoryService) QueryInventory(ctx context.Context, input QueryInventoryInput) ([]*domain.Inventory, int, error) {
	filter, err := input.ToFilter()
	if err != nil {
		return nil, 0, err
	}

	results, err := s.repo.QueryInventory(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("inventory service: query: %w", err)
	}

	total, err := s.repo.CountInventory(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("inventory service: count: %w", err)
	}

	return results, total, nil
}

// GetInventory retrieves a single inventory record by ID.
func (s *InventoryService) GetInventory(ctx context.Context, id uuid.UUID) (*domain.Inventory, error) {
	inv, err := s.repo.GetInventory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("inventory service: get %s: %w", id, err)
	}
	return inv, nil
}

// AdjustInventory adjusts the quantity of an inventory record and records a transaction.
// Business rule: qty (on-hand) must never go below zero.
//
// When a TxManager is configured, the entire read-check-write flow executes within
// a single database transaction. Inside that transaction, GetAndLockInventory
// acquires a row-level lock (SELECT ... FOR UPDATE), preventing concurrent
// adjustments from racing on stale reads.
func (s *InventoryService) AdjustInventory(ctx context.Context, id uuid.UUID, input AdjustInventoryInput) (*domain.Inventory, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Parse reference ID.
	var refID uuid.UUID
	if input.ReferenceID != "" {
		refID, _ = uuid.Parse(input.ReferenceID)
	}

	// The write+audit closure, now also responsible for reading with a lock
	// when inside a transaction.
	doWrites := func(ctx context.Context) error {
		// Read current state with a row-level lock when inside a transaction.
		// Outside a transaction, GetAndLockInventory falls back to a plain
		// SELECT (no lock) — the caller accepts the weaker guarantee.
		inv, err := s.repo.GetAndLockInventory(ctx, id)
		if err != nil {
			return fmt.Errorf("get inventory: %w", err)
		}

		// Check negative qty constraint against locked data.
		newQty := inv.Qty + input.DeltaQty
		if newQty < 0 {
			return pkgerrors.NewInvalidInput(
				fmt.Sprintf("adjustment would result in negative quantity: current=%.2f, delta=%.2f",
					inv.Qty, input.DeltaQty),
			)
		}

		if err := s.repo.UpdateInventoryQty(ctx, id, input.DeltaQty, 0); err != nil {
			return fmt.Errorf("update qty: %w", err)
		}

		tx := &domain.InventoryTransaction{
			InventoryID:   id,
			SKUID:         inv.SKUID,
			LocationID:    inv.LocationID,
			Type:          domain.InventoryTxAdjustment,
			DeltaQty:      input.DeltaQty,
			ResultingQty:  newQty,
			ReferenceType: input.ReferenceType,
			ReferenceID:   refID,
			CreatedBy:     input.CreatedBy,
		}
		if err := s.repo.CreateTransaction(ctx, tx); err != nil {
			return fmt.Errorf("create transaction: %w", err)
		}
		return nil
	}

	if s.txManager != nil {
		if err := s.txManager.WithTx(ctx, doWrites); err != nil {
			return nil, fmt.Errorf("inventory service: adjust: %w", err)
		}
	} else {
		if err := doWrites(ctx); err != nil {
			return nil, fmt.Errorf("inventory service: adjust: %w", err)
		}
	}

	// Re-fetch updated inventory to return fresh state.
	updated, err := s.repo.GetInventory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("inventory service: adjust: re-fetch: %w", err)
	}

	return updated, nil
}

// CreateInventory creates a new inventory record. This is typically called by
// other services (e.g., receiving) rather than directly from the admin API.
func (s *InventoryService) CreateInventory(ctx context.Context, input CreateInventoryInput) (*domain.Inventory, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	inv := &domain.Inventory{
		SKUID:       input.SKUID,
		LocationID:  input.LocationID,
		WarehouseID: input.WarehouseID,
		BatchNo:     input.BatchNo,
		Qty:         input.Qty,
		ReservedQty: input.ReservedQty,
		Status:      input.Status,
	}

	if err := s.repo.CreateInventory(ctx, inv); err != nil {
		return nil, fmt.Errorf("inventory service: create: %w", err)
	}

	return inv, nil
}

// GetTransactions returns the transaction history for an inventory record.
// When limit is 0, returns all transactions (no pagination limit).
func (s *InventoryService) GetTransactions(ctx context.Context, inventoryID uuid.UUID, limit, offset int) ([]*domain.InventoryTransaction, int, error) {
	txs, err := s.repo.ListTransactions(ctx, inventoryID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("inventory service: transactions %s: %w", inventoryID, err)
	}

	total, err := s.repo.CountTransactions(ctx, inventoryID)
	if err != nil {
		return nil, 0, fmt.Errorf("inventory service: count transactions %s: %w", inventoryID, err)
	}

	return txs, total, nil
}

// InventoryRetrievalInput is the input for FEFO / FIFO inventory retrieval queries.
// These queries are designed for picking decisions: they return only available,
// non-zero inventory records sorted by the appropriate strategy key.
type InventoryRetrievalInput struct {
	WarehouseID string `json:"warehouse_id,omitempty"`
	SKUID       string `json:"sku_id,omitempty"`
	Limit       int    `json:"limit,omitempty"`
}

// ToRetrievalFilter converts the input to a repository InventoryRetrievalFilter.
func (in *InventoryRetrievalInput) ToRetrievalFilter() (repository.InventoryRetrievalFilter, error) {
	f := repository.InventoryRetrievalFilter{
		Limit: in.Limit,
	}

	if in.WarehouseID != "" {
		id, err := uuid.Parse(in.WarehouseID)
		if err != nil {
			return f, pkgerrors.NewInvalidInput("invalid warehouse_id UUID")
		}
		f.WarehouseID = id
	}
	if in.SKUID != "" {
		id, err := uuid.Parse(in.SKUID)
		if err != nil {
			return f, pkgerrors.NewInvalidInput("invalid sku_id UUID")
		}
		f.SKUID = id
	}

	return f, nil
}

// GetOldestInventory returns available inventory sorted by received_at ASC
// (oldest first — FIFO strategy). This is suitable for non-perishable goods
// where the first stock received should be picked first.
func (s *InventoryService) GetOldestInventory(ctx context.Context, input InventoryRetrievalInput) ([]*domain.Inventory, error) {
	filter, err := input.ToRetrievalFilter()
	if err != nil {
		return nil, err
	}

	results, err := s.repo.GetOldestInventory(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("inventory service: fifo: %w", err)
	}
	return results, nil
}

// GetExpiringInventory returns available inventory sorted by expiry_date ASC
// NULLS LAST (earliest expiring first — FEFO strategy). This is the preferred
// strategy for perishable goods to minimise waste.
func (s *InventoryService) GetExpiringInventory(ctx context.Context, input InventoryRetrievalInput) ([]*domain.Inventory, error) {
	filter, err := input.ToRetrievalFilter()
	if err != nil {
		return nil, err
	}

	results, err := s.repo.GetExpiringInventory(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("inventory service: fefo: %w", err)
	}
	return results, nil
}
