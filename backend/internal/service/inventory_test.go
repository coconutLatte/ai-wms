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

// ── Inventory Status Transition Tests ──────────────────────────────────────────

func TestInventoryService_UpdateInventoryStatus_ValidTransitions(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: uuid.New(), LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 100, Status: domain.InventoryStatusAvailable,
	})

	// available → quarantine (quality hold)
	updated, err := svc.UpdateInventoryStatus(ctx, inv.ID, UpdateInventoryStatusInput{
		Status: domain.InventoryStatusQuarantine,
		Reason: "quality inspection flag",
	})
	if err != nil {
		t.Fatalf("available → quarantine failed: %v", err)
	}
	if updated.Status != domain.InventoryStatusQuarantine {
		t.Errorf("status = %q, want %q", updated.Status, domain.InventoryStatusQuarantine)
	}
}

func TestInventoryService_UpdateInventoryStatus_QuarantineRelease(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: uuid.New(), LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 50, Status: domain.InventoryStatusQuarantine,
	})

	// quarantine → available (release from hold)
	updated, err := svc.UpdateInventoryStatus(ctx, inv.ID, UpdateInventoryStatusInput{
		Status: domain.InventoryStatusAvailable,
		Reason: "passed inspection",
	})
	if err != nil {
		t.Fatalf("quarantine → available failed: %v", err)
	}
	if updated.Status != domain.InventoryStatusAvailable {
		t.Errorf("status = %q, want %q", updated.Status, domain.InventoryStatusAvailable)
	}
}

func TestInventoryService_UpdateInventoryStatus_DamageFlow(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: uuid.New(), LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 200, Status: domain.InventoryStatusAvailable,
	})

	// available → damaged
	updated, err := svc.UpdateInventoryStatus(ctx, inv.ID, UpdateInventoryStatusInput{
		Status: domain.InventoryStatusDamaged,
		Reason: "forklift impact damage",
	})
	if err != nil {
		t.Fatalf("available → damaged failed: %v", err)
	}
	if updated.Status != domain.InventoryStatusDamaged {
		t.Errorf("status = %q, want %q", updated.Status, domain.InventoryStatusDamaged)
	}

	// damaged → available (re-graded / repaired)
	updated, err = svc.UpdateInventoryStatus(ctx, inv.ID, UpdateInventoryStatusInput{
		Status: domain.InventoryStatusAvailable,
		Reason: "repaired and re-graded",
	})
	if err != nil {
		t.Fatalf("damaged → available failed: %v", err)
	}
	if updated.Status != domain.InventoryStatusAvailable {
		t.Errorf("status = %q, want %q", updated.Status, domain.InventoryStatusAvailable)
	}
}

func TestInventoryService_UpdateInventoryStatus_ExpireFlow(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: uuid.New(), LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 30, Status: domain.InventoryStatusQuarantine,
	})

	// quarantine → expired
	updated, err := svc.UpdateInventoryStatus(ctx, inv.ID, UpdateInventoryStatusInput{
		Status: domain.InventoryStatusExpired,
		Reason: "past expiry date",
	})
	if err != nil {
		t.Fatalf("quarantine → expired failed: %v", err)
	}
	if updated.Status != domain.InventoryStatusExpired {
		t.Errorf("status = %q, want %q", updated.Status, domain.InventoryStatusExpired)
	}
}

func TestInventoryService_UpdateInventoryStatus_InvalidTransitions(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: uuid.New(), LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 100, Status: domain.InventoryStatusAvailable,
	})

	// available → available (same status)
	_, err := svc.UpdateInventoryStatus(ctx, inv.ID, UpdateInventoryStatusInput{
		Status: domain.InventoryStatusAvailable,
	})
	if err == nil {
		t.Fatal("expected error for available → available (same status)")
	}

	// damaged → quarantine (invalid: damaged can only go to available or expired)
	inv2, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: uuid.New(), LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 50, Status: domain.InventoryStatusDamaged,
	})
	_, err = svc.UpdateInventoryStatus(ctx, inv2.ID, UpdateInventoryStatusInput{
		Status: domain.InventoryStatusQuarantine,
	})
	if err == nil {
		t.Fatal("expected error for damaged → quarantine transition")
	}
}

func TestInventoryService_UpdateInventoryStatus_TerminalState(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: uuid.New(), LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 10, Status: domain.InventoryStatusAvailable,
	})

	// available → expired (terminal)
	_, err := svc.UpdateInventoryStatus(ctx, inv.ID, UpdateInventoryStatusInput{
		Status: domain.InventoryStatusExpired,
	})
	if err != nil {
		t.Fatalf("available → expired failed: %v", err)
	}

	// Expired is terminal — cannot transition to anything.
	for _, target := range []domain.InventoryStatus{
		domain.InventoryStatusAvailable,
		domain.InventoryStatusQuarantine,
		domain.InventoryStatusDamaged,
	} {
		_, err = svc.UpdateInventoryStatus(ctx, inv.ID, UpdateInventoryStatusInput{
			Status: target,
		})
		if err == nil {
			t.Errorf("expected error for expired → %s transition", target)
		}
	}
}

func TestInventoryService_UpdateInventoryStatus_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	_, err := svc.UpdateInventoryStatus(ctx, uuid.New(), UpdateInventoryStatusInput{
		Status: domain.InventoryStatusAvailable,
	})
	if err == nil {
		t.Fatal("expected error for non-existent inventory")
	}
}

func TestInventoryService_UpdateInventoryStatus_Validation(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: uuid.New(), LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 50, Status: domain.InventoryStatusAvailable,
	})

	// Invalid status value.
	_, err := svc.UpdateInventoryStatus(ctx, inv.ID, UpdateInventoryStatusInput{
		Status: "nonsense",
	})
	if err == nil {
		t.Fatal("expected validation error for invalid status")
	}

	// Empty status.
	_, err = svc.UpdateInventoryStatus(ctx, inv.ID, UpdateInventoryStatusInput{
		Status: "",
	})
	if err == nil {
		t.Fatal("expected validation error for empty status")
	}
}

func TestInventoryService_UpdateInventoryStatus_AllTransitions(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	type transitionTest struct {
		name    string
		current domain.InventoryStatus
		target  domain.InventoryStatus
		valid   bool
	}

	tests := []transitionTest{
		// Available
		{"available → quarantine", domain.InventoryStatusAvailable, domain.InventoryStatusQuarantine, true},
		{"available → damaged", domain.InventoryStatusAvailable, domain.InventoryStatusDamaged, true},
		{"available → expired", domain.InventoryStatusAvailable, domain.InventoryStatusExpired, true},

		// Quarantine
		{"quarantine → available", domain.InventoryStatusQuarantine, domain.InventoryStatusAvailable, true},
		{"quarantine → damaged", domain.InventoryStatusQuarantine, domain.InventoryStatusDamaged, true},
		{"quarantine → expired", domain.InventoryStatusQuarantine, domain.InventoryStatusExpired, true},

		// Damaged
		{"damaged → available", domain.InventoryStatusDamaged, domain.InventoryStatusAvailable, true},
		{"damaged → expired", domain.InventoryStatusDamaged, domain.InventoryStatusExpired, true},
		{"damaged → quarantine", domain.InventoryStatusDamaged, domain.InventoryStatusQuarantine, false},

		// Expired (terminal)
		{"expired → available", domain.InventoryStatusExpired, domain.InventoryStatusAvailable, false},
		{"expired → quarantine", domain.InventoryStatusExpired, domain.InventoryStatusQuarantine, false},
		{"expired → damaged", domain.InventoryStatusExpired, domain.InventoryStatusDamaged, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
				SKUID: uuid.New(), LocationID: uuid.New(), WarehouseID: uuid.New(),
				Qty: 10, Status: tt.current,
			})

			_, err := svc.UpdateInventoryStatus(ctx, inv.ID, UpdateInventoryStatusInput{
				Status: tt.target,
			})
			if tt.valid && err != nil {
				t.Errorf("expected valid transition, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("expected invalid transition, got no error")
			}
		})
	}
}

func TestInventoryService_UpdateInventoryStatus_PreservesQty(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: uuid.New(), LocationID: uuid.New(), WarehouseID: uuid.New(),
		Qty: 100, ReservedQty: 20, BatchNo: "LOT-42",
		Status: domain.InventoryStatusAvailable,
	})

	// Status change should not affect quantities.
	updated, err := svc.UpdateInventoryStatus(ctx, inv.ID, UpdateInventoryStatusInput{
		Status: domain.InventoryStatusQuarantine,
	})
	if err != nil {
		t.Fatalf("status change failed: %v", err)
	}
	if updated.Qty != 100 {
		t.Errorf("qty = %f, want 100", updated.Qty)
	}
	if updated.ReservedQty != 20 {
		t.Errorf("reserved_qty = %f, want 20", updated.ReservedQty)
	}
	if updated.BatchNo != "LOT-42" {
		t.Errorf("batch_no = %q, want LOT-42", updated.BatchNo)
	}
	if updated.SKUID != inv.SKUID {
		t.Errorf("sku_id changed from %s to %s", inv.SKUID, updated.SKUID)
	}
}

// ── ReserveInventory Tests ─────────────────────────────────────────────────────

func TestInventoryService_ReserveInventory_FIFO(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	skuID := uuid.New()
	whID := uuid.New()
	orderLineID := uuid.New()

	// Create inventory with explicit received_at times to ensure FIFO ordering.
	inv1, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: whID,
		Qty: 50, Status: domain.InventoryStatusAvailable,
	})
	inv1.ReceivedAt = time.Now().Add(-2 * time.Hour) // Older → should be used first

	inv2, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: whID,
		Qty: 100, Status: domain.InventoryStatusAvailable,
	})
	inv2.ReceivedAt = time.Now().Add(-1 * time.Hour) // Newer

	result, err := svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: whID, Qty: 30, OrderLineID: orderLineID,
	})
	if err != nil {
		t.Fatalf("ReserveInventory failed: %v", err)
	}
	if result.TotalReserved != 30.0 {
		t.Errorf("total_reserved = %f, want 30.0", result.TotalReserved)
	}

	got1, _ := svc.GetInventory(ctx, inv1.ID)
	got2, _ := svc.GetInventory(ctx, inv2.ID)
	if got1.ReservedQty != 30.0 {
		t.Errorf("inv1 reserved_qty = %f, want 30.0", got1.ReservedQty)
	}
	if got1.AvailableQty != 20.0 {
		t.Errorf("inv1 available = %f, want 20.0", got1.AvailableQty)
	}
	if got2.ReservedQty != 0 {
		t.Errorf("inv2 reserved_qty = %f, want 0", got2.ReservedQty)
	}

	txs, _, _ := svc.GetTransactions(ctx, inv1.ID, 0, 0)
	if len(txs) != 1 {
		t.Fatalf("expected 1 transaction on inv1, got %d", len(txs))
	}
	if txs[0].Type != domain.InventoryTxReserve {
		t.Errorf("tx type = %q, want %q", txs[0].Type, domain.InventoryTxReserve)
	}
}

func TestInventoryService_ReserveInventory_FEFO(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	skuID := uuid.New()
	whID := uuid.New()
	orderLineID := uuid.New()

	expEarly := time.Now().Add(10 * 24 * time.Hour)
	expLater := time.Now().Add(60 * 24 * time.Hour)

	inv1, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: whID,
		Qty: 100, Status: domain.InventoryStatusAvailable,
	})
	inv1.ExpiryDate = &expLater

	inv2, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: whID,
		Qty: 100, Status: domain.InventoryStatusAvailable,
	})
	inv2.ExpiryDate = &expEarly

	result, err := svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: whID, Qty: 80, OrderLineID: orderLineID,
	})
	if err != nil {
		t.Fatalf("ReserveInventory (FEFO) failed: %v", err)
	}
	if result.TotalReserved != 80.0 {
		t.Errorf("total_reserved = %f, want 80.0", result.TotalReserved)
	}

	got2, _ := svc.GetInventory(ctx, inv2.ID)
	got1, _ := svc.GetInventory(ctx, inv1.ID)
	if got2.ReservedQty != 80.0 {
		t.Errorf("inv2 (earliest expiry) reserved_qty = %f, want 80.0", got2.ReservedQty)
	}
	if got1.ReservedQty != 0 {
		t.Errorf("inv1 (later expiry) reserved_qty = %f, want 0", got1.ReservedQty)
	}
}

func TestInventoryService_ReserveInventory_SpreadAcrossMultiple(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	skuID := uuid.New()
	whID := uuid.New()
	orderLineID := uuid.New()

	for i := 0; i < 3; i++ {
		svc.CreateInventory(ctx, CreateInventoryInput{
			SKUID: skuID, LocationID: uuid.New(), WarehouseID: whID,
			Qty: 30, Status: domain.InventoryStatusAvailable,
		})
	}

	result, err := svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: whID, Qty: 80, OrderLineID: orderLineID,
	})
	if err != nil {
		t.Fatalf("ReserveInventory failed: %v", err)
	}
	if result.TotalReserved != 80.0 {
		t.Errorf("total_reserved = %f, want 80.0", result.TotalReserved)
	}
	if len(result.ReservedInventoryIDs) != 3 {
		t.Errorf("expected 3 inventory IDs reserved, got %d", len(result.ReservedInventoryIDs))
	}
}

func TestInventoryService_ReserveInventory_InsufficientStock(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	skuID := uuid.New()
	whID := uuid.New()

	svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: whID,
		Qty: 10, Status: domain.InventoryStatusAvailable,
	})

	_, err := svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: whID, Qty: 50, OrderLineID: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error for insufficient stock")
	}
}

func TestInventoryService_ReserveInventory_NoAvailableStock(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	skuID := uuid.New()
	whID := uuid.New()

	svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: whID,
		Qty: 100, Status: domain.InventoryStatusQuarantine,
	})

	_, err := svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: whID, Qty: 10, OrderLineID: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error for no available stock")
	}
}

func TestInventoryService_ReserveInventory_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	tests := []struct {
		name  string
		input ReserveInventoryInput
	}{
		{"nil sku_id", ReserveInventoryInput{
			WarehouseID: uuid.New(), Qty: 10, OrderLineID: uuid.New(),
		}},
		{"nil warehouse_id", ReserveInventoryInput{
			SKUID: uuid.New(), Qty: 10, OrderLineID: uuid.New(),
		}},
		{"zero qty", ReserveInventoryInput{
			SKUID: uuid.New(), WarehouseID: uuid.New(), Qty: 0, OrderLineID: uuid.New(),
		}},
		{"negative qty", ReserveInventoryInput{
			SKUID: uuid.New(), WarehouseID: uuid.New(), Qty: -5, OrderLineID: uuid.New(),
		}},
		{"nil order_line_id", ReserveInventoryInput{
			SKUID: uuid.New(), WarehouseID: uuid.New(), Qty: 10,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ReserveInventory(ctx, tt.input)
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestInventoryService_ReserveInventory_AlreadyReserved(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	skuID := uuid.New()
	whID := uuid.New()

	svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: whID,
		Qty: 100, ReservedQty: 80, Status: domain.InventoryStatusAvailable,
	})

	_, err := svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: whID, Qty: 30, OrderLineID: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error: only 20 available (100 on-hand - 80 reserved)")
	}
}

func TestInventoryService_ReserveInventory_PartialWithinLimit(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	skuID := uuid.New()
	whID := uuid.New()

	svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: whID,
		Qty: 100, ReservedQty: 70, Status: domain.InventoryStatusAvailable,
	})

	result, err := svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: whID, Qty: 20, OrderLineID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("ReserveInventory within limit failed: %v", err)
	}
	if result.TotalReserved != 20.0 {
		t.Errorf("total_reserved = %f, want 20.0", result.TotalReserved)
	}
}

func TestInventoryService_ReserveInventory_WrongWarehouse(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	skuID := uuid.New()
	wh1 := uuid.New()
	wh2 := uuid.New()

	svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: wh1,
		Qty: 100, Status: domain.InventoryStatusAvailable,
	})

	_, err := svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: wh2, Qty: 10, OrderLineID: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error: no inventory in specified warehouse")
	}
}

// ── UnreserveInventory Tests ───────────────────────────────────────────────────

func TestInventoryService_UnreserveInventory(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	skuID := uuid.New()
	whID := uuid.New()
	orderLineID := uuid.New()

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: whID,
		Qty: 100, Status: domain.InventoryStatusAvailable,
	})

	svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: whID, Qty: 40, OrderLineID: orderLineID,
	})

	got, _ := svc.GetInventory(ctx, inv.ID)
	if got.ReservedQty != 40.0 {
		t.Fatalf("post-reserve reserved_qty = %f, want 40", got.ReservedQty)
	}

	err := svc.UnreserveInventory(ctx, UnreserveInventoryInput{OrderLineID: orderLineID})
	if err != nil {
		t.Fatalf("UnreserveInventory failed: %v", err)
	}

	got, _ = svc.GetInventory(ctx, inv.ID)
	if got.ReservedQty != 0 {
		t.Errorf("post-unreserve reserved_qty = %f, want 0", got.ReservedQty)
	}
	if got.AvailableQty != 100.0 {
		t.Errorf("post-unreserve available = %f, want 100", got.AvailableQty)
	}

	txs, _, _ := svc.GetTransactions(ctx, inv.ID, 0, 0)
	if len(txs) != 2 {
		t.Fatalf("expected 2 transactions (reserve + unreserve), got %d", len(txs))
	}
	// Check that both transaction types are present (order not guaranteed in mock).
	hasReserve := false
	hasUnreserve := false
	for _, tx := range txs {
		if tx.Type == domain.InventoryTxReserve {
			hasReserve = true
		}
		if tx.Type == domain.InventoryTxUnreserve {
			hasUnreserve = true
		}
	}
	if !hasReserve {
		t.Error("expected a reserve transaction")
	}
	if !hasUnreserve {
		t.Error("expected an unreserve transaction")
	}
}

func TestInventoryService_UnreserveInventory_Idempotent(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	err := svc.UnreserveInventory(ctx, UnreserveInventoryInput{OrderLineID: uuid.New()})
	if err != nil {
		t.Fatalf("UnreserveInventory should be idempotent, got error: %v", err)
	}
}

func TestInventoryService_UnreserveInventory_DoubleUnreserve(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	skuID := uuid.New()
	whID := uuid.New()
	orderLineID := uuid.New()

	svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: whID,
		Qty: 50, Status: domain.InventoryStatusAvailable,
	})
	svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: whID, Qty: 30, OrderLineID: orderLineID,
	})

	err := svc.UnreserveInventory(ctx, UnreserveInventoryInput{OrderLineID: orderLineID})
	if err != nil {
		t.Fatalf("first unreserve failed: %v", err)
	}

	err = svc.UnreserveInventory(ctx, UnreserveInventoryInput{OrderLineID: orderLineID})
	if err != nil {
		t.Fatalf("second unreserve (idempotent) failed: %v", err)
	}
}

func TestInventoryService_ReserveUnreserve_RoundTrip(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	skuID := uuid.New()
	whID := uuid.New()
	orderLineID1 := uuid.New()
	orderLineID2 := uuid.New()

	// Create a single inventory record with enough qty for the test.
	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: whID,
		Qty: 100, Status: domain.InventoryStatusAvailable,
	})

	// Reserve 60 for order line 1.
	result1, err := svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: whID, Qty: 60, OrderLineID: orderLineID1,
	})
	if err != nil {
		t.Fatalf("first reserve failed: %v", err)
	}
	if result1.TotalReserved != 60.0 {
		t.Errorf("total_reserved = %f, want 60.0", result1.TotalReserved)
	}

	// After reserve: 40 available, 60 reserved.
	got, _ := svc.GetInventory(ctx, inv.ID)
	if got.AvailableQty != 40.0 {
		t.Errorf("after reserve, available = %f, want 40", got.AvailableQty)
	}
	if got.ReservedQty != 60.0 {
		t.Errorf("after reserve, reserved_qty = %f, want 60", got.ReservedQty)
	}

	// Unreserve order line 1.
	err = svc.UnreserveInventory(ctx, UnreserveInventoryInput{OrderLineID: orderLineID1})
	if err != nil {
		t.Fatalf("unreserve failed: %v", err)
	}

	// After unreserve: back to 100 available, 0 reserved.
	got, _ = svc.GetInventory(ctx, inv.ID)
	if got.AvailableQty != 100.0 {
		t.Errorf("after unreserve, available = %f, want 100", got.AvailableQty)
	}
	if got.ReservedQty != 0 {
		t.Errorf("after unreserve, reserved_qty = %f, want 0", got.ReservedQty)
	}

	// Now reserve 40 for order line 2 → should succeed (all inventory available again).
	result2, err := svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: whID, Qty: 40, OrderLineID: orderLineID2,
	})
	if err != nil {
		t.Fatalf("second reserve (after unreserve) failed: %v", err)
	}
	if result2.TotalReserved != 40.0 {
		t.Errorf("total_reserved = %f, want 40.0", result2.TotalReserved)
	}
}

func TestInventoryService_UnreserveInventory_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewInventoryService(newMockInventoryRepo())

	err := svc.UnreserveInventory(ctx, UnreserveInventoryInput{OrderLineID: uuid.Nil})
	if err == nil {
		t.Fatal("expected validation error for nil order_line_id")
	}
}

func TestInventoryService_ReserveInventory_MultipleOrderLines(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	skuID := uuid.New()
	whID := uuid.New()
	olID1 := uuid.New()
	olID2 := uuid.New()

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: whID,
		Qty: 100, Status: domain.InventoryStatusAvailable,
	})

	svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: whID, Qty: 30, OrderLineID: olID1,
	})
	svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: whID, Qty: 40, OrderLineID: olID2,
	})

	got, _ := svc.GetInventory(ctx, inv.ID)
	if got.ReservedQty != 70.0 {
		t.Errorf("reserved_qty = %f, want 70.0", got.ReservedQty)
	}
	if got.AvailableQty != 30.0 {
		t.Errorf("available = %f, want 30.0", got.AvailableQty)
	}

	err := svc.UnreserveInventory(ctx, UnreserveInventoryInput{OrderLineID: olID1})
	if err != nil {
		t.Fatalf("unreserve ol1 failed: %v", err)
	}

	got, _ = svc.GetInventory(ctx, inv.ID)
	if got.ReservedQty != 40.0 {
		t.Errorf("after unreserve, reserved_qty = %f, want 40.0", got.ReservedQty)
	}
	if got.AvailableQty != 60.0 {
		t.Errorf("after unreserve, available = %f, want 60.0", got.AvailableQty)
	}
}

func TestInventoryService_ReserveInventory_RespectsAvailableQty(t *testing.T) {
	ctx := context.Background()
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)

	skuID := uuid.New()
	whID := uuid.New()

	inv, _ := svc.CreateInventory(ctx, CreateInventoryInput{
		SKUID: skuID, LocationID: uuid.New(), WarehouseID: whID,
		Qty: 100, ReservedQty: 90, Status: domain.InventoryStatusAvailable,
	})

	result, err := svc.ReserveInventory(ctx, ReserveInventoryInput{
		SKUID: skuID, WarehouseID: whID, Qty: 10, OrderLineID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("reserve exact available failed: %v", err)
	}
	if result.TotalReserved != 10.0 {
		t.Errorf("total_reserved = %f, want 10.0", result.TotalReserved)
	}

	got, _ := svc.GetInventory(ctx, inv.ID)
	if got.ReservedQty != 100.0 {
		t.Errorf("reserved_qty = %f, want 100.0", got.ReservedQty)
	}
	if got.AvailableQty != 0 {
		t.Errorf("available = %f, want 0", got.AvailableQty)
	}
}
