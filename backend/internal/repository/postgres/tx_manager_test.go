package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// ── Transaction Manager Integration Tests ──────────────────────────────────

// createTestSKU creates a SKU for use in transaction tests.
func createTestSKUForTx(t *testing.T, ctx context.Context, repo *InventoryRepo) *domain.SKU {
	t.Helper()
	s := &domain.SKU{
		Code:     fmt.Sprintf("TEST-TX-%s", uuid.New().String()[:8]),
		Name:     "Tx Test SKU",
		Category: "Test",
		UOM: domain.UOM{
			BaseUnit: "EA",
			PackQty:  1,
		},
	}
	if err := repo.CreateSKU(ctx, s); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}
	return s
}

func TestTxManager_WithTx_Commit(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	tm := NewTxManager(db)
	invRepo := NewInventoryRepo(db)
	whRepo := NewWarehouseRepo(db)

	wh, loc := createTestWarehouseZoneLocation(t, ctx, whRepo)
	sku := createTestSKUForTx(t, ctx, invRepo)

	// Create an inventory record inside a transaction.
	var invID uuid.UUID
	err := tm.WithTx(ctx, func(txCtx context.Context) error {
		inv := &domain.Inventory{
			SKUID:       sku.ID,
			LocationID:  loc.ID,
			WarehouseID: wh.ID,
			Qty:         100,
		}
		if err := invRepo.CreateInventory(txCtx, inv); err != nil {
			return err
		}
		invID = inv.ID
		return nil
	})
	if err != nil {
		t.Fatalf("WithTx failed: %v", err)
	}
	if invID == uuid.Nil {
		t.Fatal("expected invID to be set after transaction commit")
	}

	// Verify the record was persisted (commit worked).
	inv, err := invRepo.GetInventory(ctx, invID)
	if err != nil {
		t.Fatalf("GetInventory after commit failed: %v", err)
	}
	if inv.Qty != 100 {
		t.Errorf("qty = %f, want 100", inv.Qty)
	}
}

func TestTxManager_WithTx_RollbackOnError(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	tm := NewTxManager(db)
	invRepo := NewInventoryRepo(db)
	whRepo := NewWarehouseRepo(db)

	wh, loc := createTestWarehouseZoneLocation(t, ctx, whRepo)
	sku := createTestSKUForTx(t, ctx, invRepo)

	testErr := errors.New("simulated failure")

	err := tm.WithTx(ctx, func(txCtx context.Context) error {
		inv := &domain.Inventory{
			SKUID:       sku.ID,
			LocationID:  loc.ID,
			WarehouseID: wh.ID,
			Qty:         200,
		}
		if err := invRepo.CreateInventory(txCtx, inv); err != nil {
			return err
		}
		// Simulate a failure BEFORE the transaction commits.
		return testErr
	})

	// The outer WithTx should return the simulated error.
	if !errors.Is(err, testErr) {
		t.Fatalf("expected %v, got %v", testErr, err)
	}

	// Verify that no record with qty=200 exists (should have been rolled back).
	allInv, err := invRepo.QueryInventory(ctx, repository.InventoryFilter{Limit: 100})
	if err != nil {
		t.Fatalf("QueryInventory failed: %v", err)
	}
	for _, inv := range allInv {
		if inv.Qty == 200 {
			t.Error("found inventory with qty=200; transaction should have rolled back")
		}
	}
}

func TestTxManager_WithTx_AtomicInventoryAdjustment(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	tm := NewTxManager(db)
	invRepo := NewInventoryRepo(db)
	whRepo := NewWarehouseRepo(db)

	wh, loc := createTestWarehouseZoneLocation(t, ctx, whRepo)
	sku := createTestSKUForTx(t, ctx, invRepo)

	// Seed an inventory record outside a transaction.
	inv := &domain.Inventory{
		SKUID:       sku.ID,
		LocationID:  loc.ID,
		WarehouseID: wh.ID,
		Qty:         50,
	}
	if err := invRepo.CreateInventory(ctx, inv); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	// Atomically: update qty + create transaction.
	err := tm.WithTx(ctx, func(txCtx context.Context) error {
		if err := invRepo.UpdateInventoryQty(txCtx, inv.ID, 25, 0); err != nil {
			return err
		}

		txn := &domain.InventoryTransaction{
			InventoryID:   inv.ID,
			SKUID:         inv.SKUID,
			LocationID:    inv.LocationID,
			Type:          domain.InventoryTxAdjustment,
			DeltaQty:      25,
			ResultingQty:  75,
			ReferenceType: "adjustment",
			CreatedBy:     "test",
		}
		return invRepo.CreateTransaction(txCtx, txn)
	})
	if err != nil {
		t.Fatalf("WithTx failed: %v", err)
	}

	// Verify both operations committed.
	updated, err := invRepo.GetInventory(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetInventory failed: %v", err)
	}
	if updated.Qty != 75 {
		t.Errorf("qty = %f, want 75", updated.Qty)
	}

	txs, err := invRepo.ListTransactions(ctx, inv.ID, 0, 0)
	if err != nil {
		t.Fatalf("ListTransactions failed: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txs))
	}
	if txs[0].DeltaQty != 25 {
		t.Errorf("delta_qty = %f, want 25", txs[0].DeltaQty)
	}
}

func TestTxManager_WithTx_AtomicInventoryRollback(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	tm := NewTxManager(db)
	invRepo := NewInventoryRepo(db)
	whRepo := NewWarehouseRepo(db)

	wh, loc := createTestWarehouseZoneLocation(t, ctx, whRepo)
	sku := createTestSKUForTx(t, ctx, invRepo)

	// Seed an inventory record.
	inv := &domain.Inventory{
		SKUID:       sku.ID,
		LocationID:  loc.ID,
		WarehouseID: wh.ID,
		Qty:         100,
	}
	if err := invRepo.CreateInventory(ctx, inv); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	// Attempt atomic operation: qty update succeeds, but transaction creation fails.
	err := tm.WithTx(ctx, func(txCtx context.Context) error {
		if err := invRepo.UpdateInventoryQty(txCtx, inv.ID, -10, 0); err != nil {
			return err
		}
		// Use an invalid column to force a failure.
		_, err := invRepo.exec(txCtx, "INSERT INTO inventory_transactions (invalid_column) VALUES (1)")
		return err
	})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	// Verify the qty update was rolled back.
	inv2, err := invRepo.GetInventory(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetInventory failed: %v", err)
	}
	if inv2.Qty != 100 {
		t.Errorf("qty = %f, want 100 (should have rolled back)", inv2.Qty)
	}

	// Verify no transaction was created.
	txs, _ := invRepo.ListTransactions(ctx, inv.ID, 0, 0)
	if len(txs) != 0 {
		t.Errorf("expected 0 transactions after rollback, got %d", len(txs))
	}
}

func TestTxManager_WithTx_NoTransactionFallback(t *testing.T) {
	// Verify that repo methods work without an active transaction
	// — they should fall back to the pool when no tx is in context.
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	invRepo := NewInventoryRepo(db)
	whRepo := NewWarehouseRepo(db)

	wh, loc := createTestWarehouseZoneLocation(t, ctx, whRepo)
	sku := createTestSKUForTx(t, ctx, invRepo)

	// Without a transaction — should use the pool directly.
	inv := &domain.Inventory{
		SKUID:       sku.ID,
		LocationID:  loc.ID,
		WarehouseID: wh.ID,
		Qty:         300,
	}
	if err := invRepo.CreateInventory(ctx, inv); err != nil {
		t.Fatalf("CreateInventory without tx failed: %v", err)
	}

	got, err := invRepo.GetInventory(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetInventory without tx failed: %v", err)
	}
	if got.Qty != 300 {
		t.Errorf("qty = %f, want 300", got.Qty)
	}
}
