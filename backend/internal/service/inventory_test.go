package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
)

// ── Inventory Service Tests ────────────────────────────────────────────────────────────

func TestInventoryService_QueryInventory_ByWarehouse(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	whID := uuid.New()
	// Seed inventory records in mock repo
	_, _ = svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: whID,
		Qty:         100,
	})
	_, _ = svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(), // Different warehouse
		Qty:         50,
	})

	results, _, err := svc.QueryInventory(ctx, QueryInventoryInput{
		WarehouseID: whID.String(),
	})
	if err != nil {
		t.Fatalf("QueryInventory failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestInventoryService_QueryInventory_BySKU(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	skuID := uuid.New()
	_, _ = svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       skuID,
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         200,
	})
	_, _ = svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         100,
	})

	results, _, err := svc.QueryInventory(ctx, QueryInventoryInput{
		SKUID: skuID.String(),
	})
	if err != nil {
		t.Fatalf("QueryInventory failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestInventoryService_QueryInventory_ByStatus(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	_, _ = svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         10,
		Status:      domain.InventoryStatusAvailable,
	})
	_, _ = svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         5,
		Status:      domain.InventoryStatusQuarantine,
	})

	results, _, err := svc.QueryInventory(ctx, QueryInventoryInput{
		Status: domain.InventoryStatusQuarantine,
	})
	if err != nil {
		t.Fatalf("QueryInventory failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 quarantined result, got %d", len(results))
	}
}

func TestInventoryService_QueryInventory_InvalidUUID(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	_, _, err := svc.QueryInventory(ctx, QueryInventoryInput{
		WarehouseID: "not-a-uuid",
	})
	if err == nil {
		t.Fatal("expected error for invalid warehouse_id UUID")
	}
}

func TestInventoryService_GetInventory(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	inv, err := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         50.0,
		BatchNo:     "BATCH-001",
	})
	if err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	got, err := svc.GetInventory(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetInventory failed: %v", err)
	}
	if got.Qty != 50.0 {
		t.Errorf("qty = %f, want 50.0", got.Qty)
	}
	if got.BatchNo != "BATCH-001" {
		t.Errorf("batch_no = %q, want BATCH-001", got.BatchNo)
	}
}

func TestInventoryService_GetInventory_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	_, err := svc.GetInventory(ctx, uuid.New())
	if err == nil {
		t.Fatal("expected error for non-existent inventory")
	}
}

func TestInventoryService_AdjustInventory_Increase(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         100.0,
	})

	updated, err := svc.AdjustInventory(ctx, inv.ID, AdjustInventoryInput{
		DeltaQty:      50.0,
		ReferenceType: "adjustment",
		CreatedBy:     "admin",
		Reason:        "cycle count found extra stock",
	})
	if err != nil {
		t.Fatalf("AdjustInventory failed: %v", err)
	}
	if updated.Qty != 150.0 {
		t.Errorf("qty = %f, want 150.0", updated.Qty)
	}

	// Verify transaction was created.
	txs, err := svc.GetTransactions(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetTransactions failed: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txs))
	}
	if txs[0].DeltaQty != 50.0 {
		t.Errorf("delta_qty = %f, want 50.0", txs[0].DeltaQty)
	}
	if txs[0].Type != domain.InventoryTxAdjustment {
		t.Errorf("type = %q, want adjustment", txs[0].Type)
	}
	if txs[0].CreatedBy != "admin" {
		t.Errorf("created_by = %q, want admin", txs[0].CreatedBy)
	}
}

func TestInventoryService_AdjustInventory_Decrease(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         100.0,
	})

	updated, err := svc.AdjustInventory(ctx, inv.ID, AdjustInventoryInput{
		DeltaQty:      -30.0,
		ReferenceType: "adjustment",
		CreatedBy:     "admin",
		Reason:        "damaged goods removed",
	})
	if err != nil {
		t.Fatalf("AdjustInventory failed: %v", err)
	}
	if updated.Qty != 70.0 {
		t.Errorf("qty = %f, want 70.0", updated.Qty)
	}

	// Verify resulting_qty in transaction.
	txs, _ := svc.GetTransactions(ctx, inv.ID)
	if txs[0].ResultingQty != 70.0 {
		t.Errorf("resulting_qty = %f, want 70.0", txs[0].ResultingQty)
	}
}

func TestInventoryService_AdjustInventory_NegativeQty(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         10.0,
	})

	_, err := svc.AdjustInventory(ctx, inv.ID, AdjustInventoryInput{
		DeltaQty:      -20.0, // Would result in -10.0
		ReferenceType: "adjustment",
		CreatedBy:     "admin",
	})
	if err == nil {
		t.Fatal("expected error for adjustment that would result in negative quantity")
	}
}

func TestInventoryService_AdjustInventory_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         50.0,
	})

	// Zero delta
	_, err := svc.AdjustInventory(ctx, inv.ID, AdjustInventoryInput{
		DeltaQty:      0,
		ReferenceType: "adjustment",
		CreatedBy:     "admin",
	})
	if err == nil {
		t.Fatal("expected error for zero delta_qty")
	}

	// Missing reference_type
	_, err = svc.AdjustInventory(ctx, inv.ID, AdjustInventoryInput{
		DeltaQty:  10,
		CreatedBy: "admin",
	})
	if err == nil {
		t.Fatal("expected error for missing reference_type")
	}

	// Missing created_by
	_, err = svc.AdjustInventory(ctx, inv.ID, AdjustInventoryInput{
		DeltaQty:      10,
		ReferenceType: "adjustment",
	})
	if err == nil {
		t.Fatal("expected error for missing created_by")
	}
}

func TestInventoryService_AdjustInventory_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	_, err := svc.AdjustInventory(ctx, uuid.New(), AdjustInventoryInput{
		DeltaQty:      10.0,
		ReferenceType: "adjustment",
		CreatedBy:     "admin",
	})
	if err == nil {
		t.Fatal("expected error for non-existent inventory")
	}
}

func TestInventoryService_AdjustInventory_RecordsTransaction(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         100.0,
	})

	// Perform two adjustments.
	_, err := svc.AdjustInventory(ctx, inv.ID, AdjustInventoryInput{
		DeltaQty:      20.0,
		ReferenceType: "adjustment",
		ReferenceID:   uuid.New().String(),
		CreatedBy:     "user-a",
	})
	if err != nil {
		t.Fatalf("first adjustment failed: %v", err)
	}

	_, err = svc.AdjustInventory(ctx, inv.ID, AdjustInventoryInput{
		DeltaQty:      -5.0,
		ReferenceType: "cycle_count",
		CreatedBy:     "user-b",
	})
	if err != nil {
		t.Fatalf("second adjustment failed: %v", err)
	}

	txs, err := svc.GetTransactions(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetTransactions failed: %v", err)
	}
	if len(txs) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(txs))
	}
	if txs[0].DeltaQty != 20.0 {
		t.Errorf("txs[0].delta_qty = %f, want 20.0", txs[0].DeltaQty)
	}
	if txs[0].ResultingQty != 120.0 {
		t.Errorf("txs[0].resulting_qty = %f, want 120.0", txs[0].ResultingQty)
	}
	if txs[0].CreatedBy != "user-a" {
		t.Errorf("txs[0].created_by = %q, want user-a", txs[0].CreatedBy)
	}
	if txs[1].DeltaQty != -5.0 {
		t.Errorf("txs[1].delta_qty = %f, want -5.0", txs[1].DeltaQty)
	}
	if txs[1].ResultingQty != 115.0 {
		t.Errorf("txs[1].resulting_qty = %f, want 115.0", txs[1].ResultingQty)
	}
	if txs[1].CreatedBy != "user-b" {
		t.Errorf("txs[1].created_by = %q, want user-b", txs[1].CreatedBy)
	}
}

func TestInventoryService_CreateInventory(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	inv, err := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         25.0,
		ReservedQty: 5.0,
		BatchNo:     "BATCH-NEW",
	})
	if err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}
	if inv.Qty != 25.0 {
		t.Errorf("qty = %f, want 25.0", inv.Qty)
	}
	if inv.ReservedQty != 5.0 {
		t.Errorf("reserved_qty = %f, want 5.0", inv.ReservedQty)
	}
	if inv.AvailableQty != 20.0 {
		t.Errorf("available_qty = %f, want 20.0", inv.AvailableQty)
	}
}

func TestInventoryService_CreateInventory_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	// Missing SKU
	_, err := svc.CreateInventory(ctx, CreateInventoryInput{
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         10,
	})
	if err == nil {
		t.Fatal("expected error for missing sku_id")
	}

	// Missing Location
	_, err = svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         10,
	})
	if err == nil {
		t.Fatal("expected error for missing location_id")
	}

	// Missing Warehouse
	_, err = svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:      uuid.New(),
		LocationID: uuid.New(),
		Qty:        10,
	})
	if err == nil {
		t.Fatal("expected error for missing warehouse_id")
	}

	// Negative qty
	_, err = svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         -5.0,
	})
	if err == nil {
		t.Fatal("expected error for negative qty")
	}
}

func TestInventoryService_GetTransactions(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         0.0,
	})

	// Create transactions manually via the mock (adjusting from zero).
	_, _ = svc.AdjustInventory(ctx, inv.ID, AdjustInventoryInput{
		DeltaQty:      10.0,
		ReferenceType: "receipt",
		CreatedBy:     "receiver",
	})

	_, _ = svc.AdjustInventory(ctx, inv.ID, AdjustInventoryInput{
		DeltaQty:      -2.0,
		ReferenceType: "adjustment",
		CreatedBy:     "counter",
	})

	txs, err := svc.GetTransactions(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetTransactions failed: %v", err)
	}
	if len(txs) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(txs))
	}
}

func TestInventoryService_GetTransactions_Empty(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID:       uuid.New(),
		LocationID:  uuid.New(),
		WarehouseID: uuid.New(),
		Qty:         50.0,
	})

	txs, err := svc.GetTransactions(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetTransactions failed: %v", err)
	}
	if len(txs) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(txs))
	}
}
