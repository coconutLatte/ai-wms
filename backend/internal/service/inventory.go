// Package service implements business logic orchestration for the WMS domain.
package service

import (
	"context"
	"fmt"
	"time"

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

// QueryTransactionsInput is the input for querying inventory transactions globally
// (across all inventory records, with optional filters).
type QueryTransactionsInput struct {
	SKUID       string                 `json:"sku_id,omitempty"`
	WarehouseID string                 `json:"warehouse_id,omitempty"`
	TxType      domain.InventoryTxType `json:"type,omitempty"`
	DateFrom    string                 `json:"date_from,omitempty"` // RFC 3339
	DateTo      string                 `json:"date_to,omitempty"`   // RFC 3339
	Limit       int                    `json:"limit,omitempty"`
	Offset      int                    `json:"offset,omitempty"`
}

// ToFilter converts the input to a repository filter, parsing UUIDs and dates.
func (in *QueryTransactionsInput) ToFilter() (repository.InventoryTxFilter, error) {
	f := repository.InventoryTxFilter{
		TxType: in.TxType,
		Limit:  in.Limit,
		Offset: in.Offset,
	}
	if in.SKUID != "" {
		id, err := uuid.Parse(in.SKUID)
		if err != nil {
			return f, pkgerrors.NewInvalidInput("invalid sku_id UUID")
		}
		f.SKUID = id
	}
	if in.WarehouseID != "" {
		id, err := uuid.Parse(in.WarehouseID)
		if err != nil {
			return f, pkgerrors.NewInvalidInput("invalid warehouse_id UUID")
		}
		f.WarehouseID = id
	}
	if in.DateFrom != "" {
		t, err := time.Parse(time.RFC3339, in.DateFrom)
		if err != nil {
			return f, pkgerrors.NewInvalidInput("invalid date_from format (use RFC 3339)")
		}
		f.DateFrom = &t
	}
	if in.DateTo != "" {
		t, err := time.Parse(time.RFC3339, in.DateTo)
		if err != nil {
			return f, pkgerrors.NewInvalidInput("invalid date_to format (use RFC 3339)")
		}
		f.DateTo = &t
	}
	return f, nil
}

// QueryTransactions queries inventory transactions globally across all inventory
// records, with optional filters for SKU, warehouse, type, and date range.
func (s *InventoryService) QueryTransactions(ctx context.Context, input QueryTransactionsInput) ([]*domain.InventoryTransaction, int, error) {
	filter, err := input.ToFilter()
	if err != nil {
		return nil, 0, err
	}

	txs, err := s.repo.ListTransactionsGlobal(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("inventory service: query transactions: %w", err)
	}

	total, err := s.repo.CountTransactionsGlobal(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("inventory service: count transactions global: %w", err)
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

// ── Inventory Status Transitions ───────────────────────────────────────────────────────

// UpdateInventoryStatusInput is the input for updating an inventory record's status.
type UpdateInventoryStatusInput struct {
	Status domain.InventoryStatus `json:"status"`
	Reason string                 `json:"reason,omitempty"` // Human-readable reason
}

// Validate checks the input for business rule violations.
func (in *UpdateInventoryStatusInput) Validate() error {
	if !isValidInventoryStatus(in.Status) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid inventory status: %s", in.Status))
	}
	return nil
}

// UpdateInventoryStatus validates the state transition and updates the inventory status.
func (s *InventoryService) UpdateInventoryStatus(ctx context.Context, id uuid.UUID, input UpdateInventoryStatusInput) (*domain.Inventory, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	inv, err := s.repo.GetInventory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("inventory service: update status %s: %w", id, err)
	}

	// Validate the state transition.
	if !inv.CanTransitionTo(input.Status) {
		return nil, pkgerrors.NewInvalidStatus(string(inv.Status), string(input.Status))
	}

	if err := s.repo.UpdateInventoryStatus(ctx, id, input.Status); err != nil {
		return nil, fmt.Errorf("inventory service: update status %s: %w", id, err)
	}

	// Re-fetch to get updated state.
	updated, err := s.repo.GetInventory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("inventory service: re-fetch after status update %s: %w", id, err)
	}

	return updated, nil
}

func isValidInventoryStatus(s domain.InventoryStatus) bool {
	switch s {
	case domain.InventoryStatusAvailable, domain.InventoryStatusQuarantine,
		domain.InventoryStatusDamaged, domain.InventoryStatusExpired:
		return true
	}
	return false
}

// ── Inventory Reservation ────────────────────────────────────────────────────

// ReserveInventoryInput is the input for reserving inventory for order allocation.
//
// The reservation uses the best-fit strategy:
//   - FEFO (First Expired First Out) for perishable SKUs (those with expiry dates)
//   - FIFO (First In First Out) for non-perishable SKUs
//
// Inventory records are walked in strategy order and reserved from each until the
// requested quantity is fully satisfied. If insufficient available inventory exists,
// the call returns an error and no reservation is made.
//
// Each reserved quantity increments the inventory record's reserved_qty (atomic with
// qty via UpdateInventoryQty) and records an InventoryTransaction of type "reserve"
// with the order line ID as the reference.
type ReserveInventoryInput struct {
	SKUID       uuid.UUID `json:"sku_id"`
	WarehouseID uuid.UUID `json:"warehouse_id"`
	Qty         float64   `json:"qty"`
	OrderLineID uuid.UUID `json:"order_line_id"` // Reference ID for audit trail
}

// Validate checks the input for business rule violations.
func (in *ReserveInventoryInput) Validate() error {
	if in.SKUID == uuid.Nil {
		return pkgerrors.NewInvalidInput("sku_id is required")
	}
	if in.WarehouseID == uuid.Nil {
		return pkgerrors.NewInvalidInput("warehouse_id is required")
	}
	if in.Qty <= 0 {
		return pkgerrors.NewInvalidInput("qty must be positive")
	}
	if in.OrderLineID == uuid.Nil {
		return pkgerrors.NewInvalidInput("order_line_id is required")
	}
	return nil
}

// ReserveInventoryResult holds the outcome of a reservation.
type ReserveInventoryResult struct {
	ReservedInventoryIDs []uuid.UUID `json:"reserved_inventory_ids"` // Affected inventory records
	TotalReserved        float64     `json:"total_reserved"`         // Total quantity reserved
}

// ReserveInventory reserves inventory for order allocation.
//
// Strategy selection:
//  1. Query FEFO-sorted (expiring) inventory. If any records have expiry dates, use FEFO.
//  2. Otherwise, fall back to FIFO (oldest received first).
//
// The entire operation runs within a database transaction when a TxManager is configured,
// ensuring the reservation is atomic.
func (s *InventoryService) ReserveInventory(ctx context.Context, input ReserveInventoryInput) (*ReserveInventoryResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	doWrites := func(ctx context.Context) (*ReserveInventoryResult, error) {
		// Determine strategy: FEFO if any inventory has expiry dates, else FIFO.
		candidates, _, err := s.pickReserveCandidates(ctx, input.SKUID, input.WarehouseID)
		if err != nil {
			return nil, err
		}

		if len(candidates) == 0 {
			return nil, pkgerrors.NewInvalidInput(
				fmt.Sprintf("no available inventory for SKU %s in warehouse %s", input.SKUID, input.WarehouseID),
			)
		}

		// Walk candidates, reserving from each until qty is satisfied.
		remaining := input.Qty
		var reservedIDs []uuid.UUID
		totalReserved := 0.0

		for _, inv := range candidates {
			// Lock the inventory row inside a transaction to prevent race conditions.
			lockedInv, err := s.repo.GetAndLockInventory(ctx, inv.ID)
			if err != nil {
				return nil, fmt.Errorf("lock inventory %s: %w", inv.ID, err)
			}

			available := lockedInv.Available()
			if available <= 0 {
				continue // Already depleted by a concurrent operation.
			}

			reserveQty := available
			if reserveQty > remaining {
				reserveQty = remaining
			}

			if err := s.repo.UpdateInventoryQty(ctx, lockedInv.ID, 0, reserveQty); err != nil {
				return nil, fmt.Errorf("reserve qty from inventory %s: %w", lockedInv.ID, err)
			}

			// Record reserve transaction.
			// NOTE: For reserve/unreserve transactions, DeltaQty stores the
			// reserved/unreserved amount rather than an on-hand change (which is 0).
			// The Type field ("reserve" / "unreserve") disambiguates the meaning.
			tx := &domain.InventoryTransaction{
				InventoryID:   lockedInv.ID,
				SKUID:         lockedInv.SKUID,
				LocationID:    lockedInv.LocationID,
				Type:          domain.InventoryTxReserve,
				DeltaQty:      reserveQty,              // Reserved quantity (not on-hand delta)
				ResultingQty:  lockedInv.Qty,            // Resulting on-hand qty is unchanged
				ReferenceType: "order_line",
				ReferenceID:   input.OrderLineID,
			}
			if err := s.repo.CreateTransaction(ctx, tx); err != nil {
				return nil, fmt.Errorf("create reserve transaction: %w", err)
			}

			reservedIDs = append(reservedIDs, lockedInv.ID)
			totalReserved += reserveQty
			remaining -= reserveQty

			if remaining <= 0 {
				break
			}
		}

		if remaining > 0.005 { // Small epsilon for floating-point tolerance
			return nil, pkgerrors.NewInvalidInput(
				fmt.Sprintf("insufficient available inventory: needed %.2f, only %.2f available for SKU %s",
					input.Qty, input.Qty-remaining, input.SKUID),
			)
		}

		return &ReserveInventoryResult{
			ReservedInventoryIDs: reservedIDs,
			TotalReserved:        totalReserved,
		}, nil
	}

	if s.txManager != nil {
		var result *ReserveInventoryResult
		txErr := s.txManager.WithTx(ctx, func(ctx context.Context) error {
			var innerErr error
			result, innerErr = doWrites(ctx)
			return innerErr
		})
		if txErr != nil {
			return nil, fmt.Errorf("inventory service: reserve: %w", txErr)
		}
		return result, nil
	}

	result, err := doWrites(ctx)
	if err != nil {
		return nil, fmt.Errorf("inventory service: reserve: %w", err)
	}
	return result, nil
}

// UnreserveInventoryInput is the input for releasing inventory previously reserved
// for a specific order line.
type UnreserveInventoryInput struct {
	OrderLineID uuid.UUID `json:"order_line_id"` // Reference ID used during reservation
}

// Validate checks the input for business rule violations.
func (in *UnreserveInventoryInput) Validate() error {
	if in.OrderLineID == uuid.Nil {
		return pkgerrors.NewInvalidInput("order_line_id is required")
	}
	return nil
}

// UnreserveInventory releases inventory that was previously reserved for an order line.
//
// It finds all reserve transactions for the given order line ID, decrements the
// reserved_qty on each affected inventory record by the originally reserved amount
// (stored in the reserve transaction's DeltaQty), and records unreserve transactions.
//
// The entire operation runs within a database transaction when a TxManager is configured.
// Idempotent: if no reserve transactions are found, it returns nil (no-op).
func (s *InventoryService) UnreserveInventory(ctx context.Context, input UnreserveInventoryInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	doWrites := func(ctx context.Context) error {
		// Find all reserve transactions for this order line.
		reserveTxs, err := s.repo.ListTransactionsByReference(ctx, "order_line", input.OrderLineID)
		if err != nil {
			return fmt.Errorf("list reserve transactions: %w", err)
		}

		// Only process reserve-type transactions.
		var relevantTxs []*domain.InventoryTransaction
		for _, tx := range reserveTxs {
			if tx.Type == domain.InventoryTxReserve {
				relevantTxs = append(relevantTxs, tx)
			}
		}

		if len(relevantTxs) == 0 {
			return nil // Nothing to unreserve — idempotent.
		}

		// For each reserve transaction, lock the inventory and release the reserved qty.
		// The reserve transaction's DeltaQty carries the amount that was reserved.
		for _, reserveTx := range relevantTxs {
			lockedInv, err := s.repo.GetAndLockInventory(ctx, reserveTx.InventoryID)
			if err != nil {
				return fmt.Errorf("lock inventory %s: %w", reserveTx.InventoryID, err)
			}

			unreserveQty := reserveTx.DeltaQty
			if unreserveQty > lockedInv.ReservedQty {
				// Clamp to current reserved_qty (defensive: shouldn't happen).
				unreserveQty = lockedInv.ReservedQty
			}
			if unreserveQty <= 0 {
				continue // Already fully unreserved.
			}

			// Decrement reserved_qty by the originally reserved amount.
			if err := s.repo.UpdateInventoryQty(ctx, lockedInv.ID, 0, -unreserveQty); err != nil {
				return fmt.Errorf("unreserve qty from inventory %s: %w", lockedInv.ID, err)
			}

			// Record unreserve transaction.
			utx := &domain.InventoryTransaction{
				InventoryID:   lockedInv.ID,
				SKUID:         lockedInv.SKUID,
				LocationID:    lockedInv.LocationID,
				Type:          domain.InventoryTxUnreserve,
				DeltaQty:      unreserveQty,         // Amount unreserved
				ResultingQty:  lockedInv.Qty,        // On-hand qty unchanged
				ReferenceType: "order_line",
				ReferenceID:   input.OrderLineID,
			}
			if err := s.repo.CreateTransaction(ctx, utx); err != nil {
				return fmt.Errorf("create unreserve transaction: %w", err)
			}
		}

		return nil
	}

	if s.txManager != nil {
		if err := s.txManager.WithTx(ctx, doWrites); err != nil {
			return fmt.Errorf("inventory service: unreserve: %w", err)
		}
		return nil
	}

	if err := doWrites(ctx); err != nil {
		return fmt.Errorf("inventory service: unreserve: %w", err)
	}
	return nil
}

// pickReserveCandidates determines the best retrieval strategy and returns
// candidate inventory records in priority order.
//
// Default: FEFO for perishable goods (those with expiry dates), FIFO otherwise.
// The bool return is true when FEFO was used.
func (s *InventoryService) pickReserveCandidates(ctx context.Context, skuID, warehouseID uuid.UUID) ([]*domain.Inventory, bool, error) {
	filter := repository.InventoryRetrievalFilter{
		WarehouseID: warehouseID,
		SKUID:       skuID,
	}

	// Try FEFO first (expiring inventory).
	fefoCandidates, err := s.repo.GetExpiringInventory(ctx, filter)
	if err != nil {
		return nil, false, fmt.Errorf("get fefo inventory: %w", err)
	}

	// Check if any FEFO candidates actually have expiry dates.
	hasExpiry := false
	for _, inv := range fefoCandidates {
		if inv.ExpiryDate != nil {
			hasExpiry = true
			break
		}
	}

	if hasExpiry {
		return fefoCandidates, true, nil
	}

	// Fall back to FIFO (oldest first).
	fifoCandidates, err := s.repo.GetOldestInventory(ctx, filter)
	if err != nil {
		return nil, false, fmt.Errorf("get fifo inventory: %w", err)
	}

	return fifoCandidates, false, nil
}

// ── Inventory Dashboard ──────────────────────────────────────────────────────

// DashboardInput is the input for the inventory dashboard.
type DashboardInput struct {
	WarehouseID       string  `json:"warehouse_id,omitempty"`
	LowStockThreshold float64 `json:"low_stock_threshold,omitempty"`
}

// Validate sets defaults and validates the input.
func (in *DashboardInput) Validate() {
	if in.LowStockThreshold <= 0 {
		in.LowStockThreshold = 10.0
	}
}

// ToFilter converts to filter params.
func (in *DashboardInput) ToFilter() (warehouseID uuid.UUID, threshold float64, err error) {
	threshold = in.LowStockThreshold
	if threshold <= 0 {
		threshold = 10.0
	}
	if in.WarehouseID != "" {
		id, err := uuid.Parse(in.WarehouseID)
		if err != nil {
			return uuid.Nil, 0, pkgerrors.NewInvalidInput("invalid warehouse_id UUID")
		}
		warehouseID = id
	}
	return warehouseID, threshold, nil
}

// GetDashboardStats returns aggregated inventory dashboard data.
func (s *InventoryService) GetDashboardStats(ctx context.Context, input DashboardInput) (*repository.InventoryDashboardStats, []*domain.Inventory, []*repository.InventoryByWarehouseRow, error) {
	input.Validate()
	warehouseID, threshold, err := input.ToFilter()
	if err != nil {
		return nil, nil, nil, err
	}

	stats, err := s.repo.GetInventoryDashboardStats(ctx, warehouseID, threshold)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("inventory service: dashboard stats: %w", err)
	}

	lowStock, err := s.repo.GetLowStockInventory(ctx, threshold, warehouseID, 20)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("inventory service: low stock: %w", err)
	}

	byWarehouse, err := s.repo.GetInventoryByWarehouse(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("inventory service: by warehouse: %w", err)
	}

	return stats, lowStock, byWarehouse, nil
}
