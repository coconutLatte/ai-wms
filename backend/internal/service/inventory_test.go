package service

import (
	"context"
	"testing"
	"time"

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
	txs, _, err := svc.GetTransactions(ctx, inv.ID, 0, 0)
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
	txs, _, _ := svc.GetTransactions(ctx, inv.ID, 0, 0)
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

	txs, _, err := svc.GetTransactions(ctx, inv.ID, 0, 0)
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

	txs, _, err := svc.GetTransactions(ctx, inv.ID, 0, 0)
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

	txs, _, err := svc.GetTransactions(ctx, inv.ID, 0, 0)
	if err != nil {
		t.Fatalf("GetTransactions failed: %v", err)
	}
	if len(txs) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(txs))
	}
}

// ── FEFO / FIFO Retrieval Tests ──────────────────────────────────────────────

func TestInventoryService_GetOldestInventory_FIFO(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	skuID := uuid.New()
	now := time.Now()

	// Create inventory records with different received_at times.
	inv1, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 100, Status: domain.InventoryStatusAvailable,
	})
	inv1.ReceivedAt = now.Add(-3 * time.Hour) // Oldest

	inv2, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 200, Status: domain.InventoryStatusAvailable,
	})
	inv2.ReceivedAt = now.Add(-1 * time.Hour)

	inv3, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 50, Status: domain.InventoryStatusQuarantine, // Not available — excluded
	})
	inv3.ReceivedAt = now.Add(-72 * time.Hour)

	results, err := svc.GetOldestInventory(ctx, InventoryRetrievalInput{
		SKUID: skuID.String(),
	})
	if err != nil {
		t.Fatalf("GetOldestInventory failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 available results, got %d", len(results))
	}
	// FIFO: oldest first
	if results[0].ID != inv1.ID {
		t.Errorf("first result should be oldest (inv1), got inv with qty=%f", results[0].Qty)
	}
	if results[1].ID != inv2.ID {
		t.Errorf("second result should be inv2, got inv with qty=%f", results[1].Qty)
	}
}

func TestInventoryService_GetOldestInventory_NoAvailableStock(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	skuID := uuid.New()

	// All inventory is quarantined or has zero qty.
	_, _ = svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 50, Status: domain.InventoryStatusQuarantine,
	})
	_, _ = svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 0, Status: domain.InventoryStatusAvailable,
	})

	results, err := svc.GetOldestInventory(ctx, InventoryRetrievalInput{
		SKUID: skuID.String(),
	})
	if err != nil {
		t.Fatalf("GetOldestInventory failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestInventoryService_GetOldestInventory_Limit(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	skuID := uuid.New()
	now := time.Now()

	for i := 0; i < 5; i++ {
		inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
			SKUID: skuID, LocationID: uuid.New(), WarehouseID: uuid.New(),
			Qty: 10, Status: domain.InventoryStatusAvailable,
		})
		inv.ReceivedAt = now.Add(-time.Duration(5-i) * time.Hour)
	}

	results, err := svc.GetOldestInventory(ctx, InventoryRetrievalInput{
		SKUID: skuID.String(),
		Limit: 3,
	})
	if err != nil {
		t.Fatalf("GetOldestInventory failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results (limit=3), got %d", len(results))
	}
}

func TestInventoryService_GetExpiringInventory_FEFO(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	skuID := uuid.New()

	// Create inventory with different expiry dates.
	expEarly := time.Now().Add(30 * 24 * time.Hour)  // Expires in 30 days
	expLater := time.Now().Add(90 * 24 * time.Hour)  // Expires in 90 days

	inv1, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 100, Status: domain.InventoryStatusAvailable,
	})
	inv1.ExpiryDate = &expEarly // Earliest expiry — should be first

	inv2, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 200, Status: domain.InventoryStatusAvailable,
	})
	inv2.ExpiryDate = &expLater

	inv3, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 50, Status: domain.InventoryStatusAvailable,
	})
	inv3.ExpiryDate = nil // No expiry — should be last

	results, err := svc.GetExpiringInventory(ctx, InventoryRetrievalInput{
		SKUID: skuID.String(),
	})
	if err != nil {
		t.Fatalf("GetExpiringInventory failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	// FEFO: earliest expiring first, nil expiry last
	if results[0].ID != inv1.ID {
		t.Errorf("first result should be earliest expiring (inv1)")
	}
	if results[1].ID != inv2.ID {
		t.Errorf("second result should be later expiring (inv2)")
	}
	if results[2].ID != inv3.ID {
		t.Errorf("third result should be no expiry (inv3)")
	}
}

func TestInventoryService_GetExpiringInventory_NoExpiryDates(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	skuID := uuid.New()

	// All inventory has nil expiry dates — sorting is stable.
	for i := 0; i < 3; i++ {
		_, _ = svc.CreateInventory(ctx, CreateInventoryInput{
			SKUID: skuID, LocationID: uuid.New(), WarehouseID: uuid.New(),
			Qty: float64((i + 1) * 10), Status: domain.InventoryStatusAvailable,
		})
	}

	results, err := svc.GetExpiringInventory(ctx, InventoryRetrievalInput{
		SKUID: skuID.String(),
	})
	if err != nil {
		t.Fatalf("GetExpiringInventory failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results (all no expiry), got %d", len(results))
	}
}

func TestInventoryService_GetExpiringInventory_WarehouseFilter(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	skuID := uuid.New()
	wh1 := uuid.New()
	wh2 := uuid.New()

	_, _ = svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: wh1,
		Qty: 100, Status: domain.InventoryStatusAvailable,
	})
	_, _ = svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: wh2,
		Qty: 200, Status: domain.InventoryStatusAvailable,
	})

	results, err := svc.GetExpiringInventory(ctx, InventoryRetrievalInput{
		SKUID:       skuID.String(),
		WarehouseID: wh1.String(),
	})
	if err != nil {
		t.Fatalf("GetExpiringInventory failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for wh1, got %d", len(results))
	}
	if results[0].WarehouseID != wh1 {
		t.Errorf("warehouse_id = %s, want %s", results[0].WarehouseID, wh1)
	}
}

func TestInventoryService_GetExpiringInventory_InvalidUUID(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	_, err := svc.GetExpiringInventory(ctx, InventoryRetrievalInput{
		WarehouseID: "not-a-uuid",
	})
	if err == nil {
		t.Fatal("expected error for invalid warehouse_id UUID")
	}

	_, err = svc.GetOldestInventory(ctx, InventoryRetrievalInput{
		SKUID: "also-not-a-uuid",
	})
	if err == nil {
		t.Fatal("expected error for invalid sku_id UUID")
	}
}
