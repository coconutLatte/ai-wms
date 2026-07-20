package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// setupInventoryTestDB creates a test database and cleans up inventory/sku related data.
func setupInventoryTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	cfg := testConfig()

	ctx := context.Background()
	db, err := NewDB(ctx, cfg)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
		return nil, nil
	}

	// Clean up previous test data (order matters due to FK constraints)
	db.Pool.Exec(ctx, "DELETE FROM inventory_transactions WHERE sku_id IN (SELECT id FROM skus WHERE code LIKE 'TEST-%')")
	db.Pool.Exec(ctx, "DELETE FROM inventory WHERE sku_id IN (SELECT id FROM skus WHERE code LIKE 'TEST-%')")
	db.Pool.Exec(ctx, "DELETE FROM skus WHERE code LIKE 'TEST-%'")
	db.Pool.Exec(ctx, "DELETE FROM locations WHERE code LIKE 'TEST-%'")
	db.Pool.Exec(ctx, "DELETE FROM zones WHERE code LIKE 'TEST-%'")
	db.Pool.Exec(ctx, "DELETE FROM warehouses WHERE code LIKE 'TEST-%'")

	cleanup := func() {
		db.Pool.Exec(ctx, "DELETE FROM inventory_transactions WHERE sku_id IN (SELECT id FROM skus WHERE code LIKE 'TEST-%')")
		db.Pool.Exec(ctx, "DELETE FROM inventory WHERE sku_id IN (SELECT id FROM skus WHERE code LIKE 'TEST-%')")
		db.Pool.Exec(ctx, "DELETE FROM skus WHERE code LIKE 'TEST-%'")
		db.Pool.Exec(ctx, "DELETE FROM locations WHERE code LIKE 'TEST-%'")
		db.Pool.Exec(ctx, "DELETE FROM zones WHERE code LIKE 'TEST-%'")
		db.Pool.Exec(ctx, "DELETE FROM warehouses WHERE code LIKE 'TEST-%'")
		db.Close()
	}

	return db, cleanup
}

// createTestWarehouseZoneLocation creates a warehouse, zone, and location for inventory tests.
func createTestWarehouseZoneLocation(t *testing.T, ctx context.Context, repo *WarehouseRepo) (wh *domain.Warehouse, loc *domain.Location) {
	t.Helper()

	wh = &domain.Warehouse{
		Code: "TEST-WH-INV-" + uuid.New().String()[:8],
		Name: "Inventory Test Warehouse",
	}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	zone := &domain.Zone{
		WarehouseID: wh.ID,
		Code:        "TEST-ZONE-INV-" + uuid.New().String()[:8],
		Name:        "Inventory Zone",
		ZoneType:    domain.ZoneTypeStorage,
	}
	if err := repo.CreateZone(ctx, zone); err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}

	loc = &domain.Location{
		ZoneID:       zone.ID,
		WarehouseID:  wh.ID,
		Code:         "TEST-LOC-INV-" + uuid.New().String()[:8],
		LocationType: domain.LocationTypeShelf,
	}
	if err := repo.CreateLocation(ctx, loc); err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}

	return wh, loc
}

// ── SKU Tests ──────────────────────────────────────────────

func TestInventoryRepo_CreateAndGetSKU(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewInventoryRepo(db)

	s := &domain.SKU{
		Code:        "TEST-SKU-001",
		Name:        "Test Product",
		Description: "A test product",
		Barcode:     "BC-TEST-001",
		Category:    "Electronics",
		UOM: domain.UOM{
			BaseUnit: "EA",
			PackUnit: "BOX",
			PackQty:  10,
			Weight:   0.5,
			Volume:   0.001,
			Length:   10.0,
			Width:    5.0,
			Height:   2.0,
		},
		Attributes: domain.Attributes{"color": "red", "size": "M"},
	}

	err := repo.CreateSKU(ctx, s)
	if err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}
	if s.ID == uuid.Nil {
		t.Error("expected SKU ID to be set")
	}
	if s.Status != domain.SKUStatusActive {
		t.Errorf("status = %q, want active", s.Status)
	}

	got, err := repo.GetSKU(ctx, s.ID)
	if err != nil {
		t.Fatalf("GetSKU failed: %v", err)
	}
	if got.Code != s.Code {
		t.Errorf("code = %q, want %q", got.Code, s.Code)
	}
	if got.Name != s.Name {
		t.Errorf("name = %q, want %q", got.Name, s.Name)
	}
	if got.Barcode != s.Barcode {
		t.Errorf("barcode = %q, want %q", got.Barcode, s.Barcode)
	}
	if got.UOM.Weight != 0.5 {
		t.Errorf("weight = %f, want 0.5", got.UOM.Weight)
	}
	if got.UOM.PackUnit != "BOX" {
		t.Errorf("pack_unit = %q, want BOX", got.UOM.PackUnit)
	}
	if got.UOM.PackQty != 10 {
		t.Errorf("pack_qty = %d, want 10", got.UOM.PackQty)
	}
	if len(got.Attributes) != 2 || got.Attributes["color"] != "red" {
		t.Errorf("attributes = %v, want {color: red, size: M}", got.Attributes)
	}
}

func TestInventoryRepo_GetSKUByCode(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewInventoryRepo(db)

	s := &domain.SKU{
		Code: "TEST-SKU-CODE-001",
		Name: "Code Lookup Product",
		UOM:  domain.UOM{BaseUnit: "KG", PackQty: 5},
	}
	if err := repo.CreateSKU(ctx, s); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	got, err := repo.GetSKUByCode(ctx, "TEST-SKU-CODE-001")
	if err != nil {
		t.Fatalf("GetSKUByCode failed: %v", err)
	}
	if got.ID != s.ID {
		t.Errorf("id = %s, want %s", got.ID, s.ID)
	}

	// Not found
	_, err = repo.GetSKUByCode(ctx, "NONEXISTENT")
	if err == nil {
		t.Error("expected error for nonexistent SKU code")
	}
}

func TestInventoryRepo_GetSKUByBarcode(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewInventoryRepo(db)

	s := &domain.SKU{
		Code:    "TEST-SKU-BC-001",
		Name:    "Barcode Lookup Product",
		Barcode: "UNIQUE-BC-SKU-123",
		UOM:     domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := repo.CreateSKU(ctx, s); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	got, err := repo.GetSKUByBarcode(ctx, "UNIQUE-BC-SKU-123")
	if err != nil {
		t.Fatalf("GetSKUByBarcode failed: %v", err)
	}
	if got.ID != s.ID {
		t.Errorf("id = %s, want %s", got.ID, s.ID)
	}

	// SKU without barcode
	s2 := &domain.SKU{
		Code: "TEST-SKU-BC-002",
		Name: "No Barcode Product",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := repo.CreateSKU(ctx, s2); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	_, err = repo.GetSKUByBarcode(ctx, "NONEXISTENT-BC")
	if err == nil {
		t.Error("expected error for nonexistent barcode")
	}
}

func TestInventoryRepo_ListSKUs(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewInventoryRepo(db)

	for range 3 {
		s := &domain.SKU{
			Code: "TEST-SKU-LIST-" + uuid.New().String()[:8],
			Name: "List SKU Product",
			UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
		}
		if err := repo.CreateSKU(ctx, s); err != nil {
			t.Fatalf("CreateSKU failed: %v", err)
		}
	}

	skus, err := repo.ListSKUs(ctx, 0, 0)
	if err != nil {
		t.Fatalf("ListSKUs failed: %v", err)
	}
	if len(skus) < 3 {
		t.Errorf("expected at least 3 SKUs, got %d", len(skus))
	}
}

func TestInventoryRepo_UpdateSKU(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewInventoryRepo(db)

	s := &domain.SKU{
		Code:       "TEST-SKU-UPD-001",
		Name:       "Original SKU Name",
		Barcode:    "BC-UPD-001",
		Attributes: domain.Attributes{"flavor": "original"},
		UOM:        domain.UOM{BaseUnit: "EA", Weight: 1.0},
	}
	if err := repo.CreateSKU(ctx, s); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	s.Name = "Updated SKU Name"
	s.Description = "Updated description"
	s.Category = "Updated Category"
	s.Status = domain.SKUStatusInactive
	s.Attributes["flavor"] = "updated"
	s.Attributes["new_attr"] = "value"
	s.UOM.Weight = 2.5

	if err := repo.UpdateSKU(ctx, s); err != nil {
		t.Fatalf("UpdateSKU failed: %v", err)
	}

	got, err := repo.GetSKU(ctx, s.ID)
	if err != nil {
		t.Fatalf("GetSKU failed: %v", err)
	}
	if got.Name != "Updated SKU Name" {
		t.Errorf("name = %q, want %q", got.Name, "Updated SKU Name")
	}
	if got.Status != domain.SKUStatusInactive {
		t.Errorf("status = %q, want inactive", got.Status)
	}
	if got.UOM.Weight != 2.5 {
		t.Errorf("weight = %f, want 2.5", got.UOM.Weight)
	}
	if got.Attributes["flavor"] != "updated" {
		t.Errorf("attributes[flavor] = %q, want updated", got.Attributes["flavor"])
	}
	if got.Attributes["new_attr"] != "value" {
		t.Errorf("attributes[new_attr] = %q, want value", got.Attributes["new_attr"])
	}
}

func TestInventoryRepo_GetSKU_NotFound(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewInventoryRepo(db)

	_, err := repo.GetSKU(ctx, uuid.New())
	if err == nil {
		t.Error("expected error for nonexistent SKU")
	}
}

// ── Inventory Tests ────────────────────────────────────────

func TestInventoryRepo_CreateAndGetInventory(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)

	wh, loc := createTestWarehouseZoneLocation(t, ctx, whRepo)

	sku := &domain.SKU{
		Code: "TEST-SKU-INV-001",
		Name: "Inventory SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	inv := &domain.Inventory{
		SKUID:      sku.ID,
		LocationID: loc.ID,
		WarehouseID: wh.ID,
		BatchNo:    "BATCH-2026-001",
		Qty:        100.0,
		ReservedQty: 10.0,
		Status:     domain.InventoryStatusAvailable,
	}

	err := invRepo.CreateInventory(ctx, inv)
	if err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}
	if inv.ID == uuid.Nil {
		t.Error("expected inventory ID to be set")
	}

	got, err := invRepo.GetInventory(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetInventory failed: %v", err)
	}
	if got.SKUID != sku.ID {
		t.Errorf("sku_id = %s, want %s", got.SKUID, sku.ID)
	}
	if got.Qty != 100.0 {
		t.Errorf("qty = %f, want 100.0", got.Qty)
	}
	if got.ReservedQty != 10.0 {
		t.Errorf("reserved_qty = %f, want 10.0", got.ReservedQty)
	}
	if got.AvailableQty != 90.0 {
		t.Errorf("available_qty = %f, want 90.0 (computed: qty - reserved_qty)", got.AvailableQty)
	}
	if got.BatchNo != "BATCH-2026-001" {
		t.Errorf("batch_no = %q, want BATCH-2026-001", got.BatchNo)
	}
	if got.Status != domain.InventoryStatusAvailable {
		t.Errorf("status = %q, want available", got.Status)
	}
}

func TestInventoryRepo_GetInventoryAtLocation(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)

	wh, loc := createTestWarehouseZoneLocation(t, ctx, whRepo)

	sku := &domain.SKU{
		Code: "TEST-SKU-LOC-001",
		Name: "Location Inventory SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	inv := &domain.Inventory{
		SKUID:       sku.ID,
		LocationID:  loc.ID,
		WarehouseID: wh.ID,
		BatchNo:     "BATCH-LOC-001",
		Qty:         50.0,
		Status:      domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, inv); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	got, err := invRepo.GetInventoryAtLocation(ctx, sku.ID, loc.ID, "BATCH-LOC-001")
	if err != nil {
		t.Fatalf("GetInventoryAtLocation failed: %v", err)
	}
	if got.ID != inv.ID {
		t.Errorf("id = %s, want %s", got.ID, inv.ID)
	}
	if got.Qty != 50.0 {
		t.Errorf("qty = %f, want 50.0", got.Qty)
	}

	// Not found for wrong batch
	_, err = invRepo.GetInventoryAtLocation(ctx, sku.ID, loc.ID, "WRONG-BATCH")
	if err == nil {
		t.Error("expected error for wrong batch")
	}
}

func TestInventoryRepo_QueryInventory(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)

	wh, loc := createTestWarehouseZoneLocation(t, ctx, whRepo)

	sku := &domain.SKU{
		Code: "TEST-SKU-QRY-001",
		Name: "Query SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	// Create inventory records
	for i := range 3 {
		inv := &domain.Inventory{
			SKUID:       sku.ID,
			LocationID:  loc.ID,
			WarehouseID: wh.ID,
			BatchNo:     "BATCH-QRY-00" + string(rune('1'+i)),
			Qty:         float64((i + 1) * 10),
			Status:      domain.InventoryStatusAvailable,
		}
		if err := invRepo.CreateInventory(ctx, inv); err != nil {
			t.Fatalf("CreateInventory failed: %v", err)
		}
	}

	// Query by warehouse
	results, err := invRepo.QueryInventory(ctx, repository.InventoryFilter{
		WarehouseID: wh.ID,
	})
	if err != nil {
		t.Fatalf("QueryInventory failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Query by SKU
	results, err = invRepo.QueryInventory(ctx, repository.InventoryFilter{
		SKUID: sku.ID,
	})
	if err != nil {
		t.Fatalf("QueryInventory by SKU failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Query by batch
	results, err = invRepo.QueryInventory(ctx, repository.InventoryFilter{
		BatchNo: "BATCH-QRY-002",
	})
	if err != nil {
		t.Fatalf("QueryInventory by batch failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	// Query with limit
	results, err = invRepo.QueryInventory(ctx, repository.InventoryFilter{
		WarehouseID: wh.ID,
		Limit:       2,
	})
	if err != nil {
		t.Fatalf("QueryInventory with limit failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results with limit, got %d", len(results))
	}
}

func TestInventoryRepo_UpdateInventoryQty(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)

	wh, loc := createTestWarehouseZoneLocation(t, ctx, whRepo)

	sku := &domain.SKU{
		Code: "TEST-SKU-QTY-001",
		Name: "Qty Update SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	inv := &domain.Inventory{
		SKUID:       sku.ID,
		LocationID:  loc.ID,
		WarehouseID: wh.ID,
		BatchNo:     "BATCH-QTY-001",
		Qty:         100.0,
		ReservedQty: 20.0,
		Status:      domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, inv); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	// Add quantity and reserve more
	err := invRepo.UpdateInventoryQty(ctx, inv.ID, 50.0, 10.0)
	if err != nil {
		t.Fatalf("UpdateInventoryQty failed: %v", err)
	}

	got, err := invRepo.GetInventory(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetInventory failed: %v", err)
	}
	if got.Qty != 150.0 {
		t.Errorf("qty = %f, want 150.0", got.Qty)
	}
	if got.ReservedQty != 30.0 {
		t.Errorf("reserved_qty = %f, want 30.0", got.ReservedQty)
	}
	if got.AvailableQty != 120.0 {
		t.Errorf("available_qty = %f, want 120.0", got.AvailableQty)
	}

	// Deduct quantity and release reservations
	err = invRepo.UpdateInventoryQty(ctx, inv.ID, -100.0, -15.0)
	if err != nil {
		t.Fatalf("UpdateInventoryQty (deduct) failed: %v", err)
	}

	got, err = invRepo.GetInventory(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetInventory failed: %v", err)
	}
	if got.Qty != 50.0 {
		t.Errorf("qty = %f, want 50.0", got.Qty)
	}
	if got.ReservedQty != 15.0 {
		t.Errorf("reserved_qty = %f, want 15.0", got.ReservedQty)
	}
	if got.AvailableQty != 35.0 {
		t.Errorf("available_qty = %f, want 35.0", got.AvailableQty)
	}
}

func TestInventoryRepo_CreateInventory_Defaults(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)

	wh, loc := createTestWarehouseZoneLocation(t, ctx, whRepo)

	sku := &domain.SKU{
		Code: "TEST-SKU-DEF-001",
		Name: "Default Inventory SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	// Minimal inventory with no explicit status
	inv := &domain.Inventory{
		SKUID:       sku.ID,
		LocationID:  loc.ID,
		WarehouseID: wh.ID,
		Qty:         0.0,
	}

	if err := invRepo.CreateInventory(ctx, inv); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}
	if inv.Status != domain.InventoryStatusAvailable {
		t.Errorf("status = %q, want available (default)", inv.Status)
	}

	got, err := invRepo.GetInventory(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetInventory failed: %v", err)
	}
	if got.AvailableQty != 0.0 {
		t.Errorf("available_qty = %f, want 0 when qty and reserved are 0", got.AvailableQty)
	}
}

// ── Inventory Transaction Tests ────────────────────────────

func TestInventoryRepo_CreateAndListTransactions(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)

	wh, loc := createTestWarehouseZoneLocation(t, ctx, whRepo)

	sku := &domain.SKU{
		Code: "TEST-SKU-TX-001",
		Name: "Transaction SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	inv := &domain.Inventory{
		SKUID:       sku.ID,
		LocationID:  loc.ID,
		WarehouseID: wh.ID,
		BatchNo:     "BATCH-TX-001",
		Qty:         0.0,
		Status:      domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, inv); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	// Create receipt transaction
	tx1 := &domain.InventoryTransaction{
		InventoryID:   inv.ID,
		SKUID:         sku.ID,
		LocationID:    loc.ID,
		Type:          domain.InventoryTxReceipt,
		DeltaQty:      100.0,
		ResultingQty:  100.0,
		ReferenceType: "order",
		ReferenceID:   uuid.New(),
		CreatedBy:     "test-user",
	}
	if err := invRepo.CreateTransaction(ctx, tx1); err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}
	if tx1.ID == uuid.Nil {
		t.Error("expected transaction ID to be set")
	}

	// Create pick transaction
	tx2 := &domain.InventoryTransaction{
		InventoryID:   inv.ID,
		SKUID:         sku.ID,
		LocationID:    loc.ID,
		Type:          domain.InventoryTxPick,
		DeltaQty:      -30.0,
		ResultingQty:  70.0,
		ReferenceType: "task",
		ReferenceID:   uuid.New(),
		CreatedBy:     "picker-1",
	}
	if err := invRepo.CreateTransaction(ctx, tx2); err != nil {
		t.Fatalf("CreateTransaction (pick) failed: %v", err)
	}

	// List transactions
	txs, err := invRepo.ListTransactions(ctx, inv.ID, 0, 0)
	if err != nil {
		t.Fatalf("ListTransactions failed: %v", err)
	}
	if len(txs) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(txs))
	}

	// Transactions are ordered by created_at DESC, so tx2 is first
	if txs[1].Type != domain.InventoryTxReceipt {
		t.Errorf("txs[1].type = %q, want receipt", txs[1].Type)
	}
	if txs[1].DeltaQty != 100.0 {
		t.Errorf("txs[1].delta_qty = %f, want 100.0", txs[1].DeltaQty)
	}
	if txs[1].CreatedBy != "test-user" {
		t.Errorf("txs[1].created_by = %q, want test-user", txs[1].CreatedBy)
	}

	if txs[0].Type != domain.InventoryTxPick {
		t.Errorf("txs[0].type = %q, want pick", txs[0].Type)
	}
	if txs[0].DeltaQty != -30.0 {
		t.Errorf("txs[0].delta_qty = %f, want -30.0", txs[0].DeltaQty)
	}
	if txs[0].ReferenceType != "task" {
		t.Errorf("txs[0].reference_type = %q, want task", txs[0].ReferenceType)
	}
}

func TestInventoryRepo_CreateInventory_WithDates(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)

	wh, loc := createTestWarehouseZoneLocation(t, ctx, whRepo)

	sku := &domain.SKU{
		Code: "TEST-SKU-DATE-001",
		Name: "Date Tracking SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	prodDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	expDate := time.Date(2027, 1, 15, 0, 0, 0, 0, time.UTC)

	inv := &domain.Inventory{
		SKUID:          sku.ID,
		LocationID:     loc.ID,
		WarehouseID:    wh.ID,
		BatchNo:        "BATCH-DATE-001",
		Qty:            200.0,
		ProductionDate: &prodDate,
		ExpiryDate:     &expDate,
	}
	if err := invRepo.CreateInventory(ctx, inv); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	got, err := invRepo.GetInventory(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetInventory failed: %v", err)
	}
	if got.ProductionDate == nil {
		t.Fatal("expected production_date to be set")
	}
	if !got.ProductionDate.Equal(prodDate) {
		t.Errorf("production_date = %v, want %v", got.ProductionDate, prodDate)
	}
	if got.ExpiryDate == nil {
		t.Fatal("expected expiry_date to be set")
	}
	if !got.ExpiryDate.Equal(expDate) {
		t.Errorf("expiry_date = %v, want %v", got.ExpiryDate, expDate)
	}
}
