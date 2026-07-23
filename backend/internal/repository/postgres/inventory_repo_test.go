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

// ── FEFO / FIFO Repository Tests ───────────────────────────

func TestInventoryRepo_GetOldestInventory_FIFO(t *testing.T) {
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
		Code: "TEST-SKU-FIFO-001",
		Name: "FIFO SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	// Create inventory records — they will have sequential received_at timestamps
	// because CreateInventory sets ReceivedAt = time.Now().
	inv1 := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		BatchNo: "FIFO-BATCH-1", Qty: 100, Status: domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, inv1); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond) // Ensure distinct ReceivedAt

	inv2 := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		BatchNo: "FIFO-BATCH-2", Qty: 200, Status: domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, inv2); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	inv3 := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		BatchNo: "FIFO-BATCH-3", Qty: 0, Status: domain.InventoryStatusAvailable, // Zero qty — excluded
	}
	if err := invRepo.CreateInventory(ctx, inv3); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	results, err := invRepo.GetOldestInventory(ctx, repository.InventoryRetrievalFilter{
		SKUID: sku.ID,
	})
	if err != nil {
		t.Fatalf("GetOldestInventory failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results (excluding zero-qty), got %d", len(results))
	}
	// FIFO: oldest first
	if results[0].BatchNo != "FIFO-BATCH-1" {
		t.Errorf("first should be FIFO-BATCH-1 (oldest), got %s", results[0].BatchNo)
	}
	if results[1].BatchNo != "FIFO-BATCH-2" {
		t.Errorf("second should be FIFO-BATCH-2, got %s", results[1].BatchNo)
	}
}

func TestInventoryRepo_GetOldestInventory_QuarantineExcluded(t *testing.T) {
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
		Code: "TEST-SKU-FIFO-Q-001",
		Name: "FIFO Quarantine SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	inv := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		BatchNo: "Q-BATCH", Qty: 500, Status: domain.InventoryStatusQuarantine,
	}
	if err := invRepo.CreateInventory(ctx, inv); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	results, err := invRepo.GetOldestInventory(ctx, repository.InventoryRetrievalFilter{
		SKUID: sku.ID,
	})
	if err != nil {
		t.Fatalf("GetOldestInventory failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results (quarantine excluded), got %d", len(results))
	}
}

func TestInventoryRepo_GetExpiringInventory_FEFO(t *testing.T) {
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
		Code: "TEST-SKU-FEFO-001",
		Name: "FEFO SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	expEarly := time.Now().Add(30 * 24 * time.Hour)
	expLater := time.Now().Add(90 * 24 * time.Hour)

	// Earliest expiry — should be first
	inv1 := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		BatchNo: "FEFO-EARLY", Qty: 100, Status: domain.InventoryStatusAvailable,
		ExpiryDate: &expEarly,
	}
	if err := invRepo.CreateInventory(ctx, inv1); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	// Later expiry — should be second
	inv2 := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		BatchNo: "FEFO-LATE", Qty: 200, Status: domain.InventoryStatusAvailable,
		ExpiryDate: &expLater,
	}
	if err := invRepo.CreateInventory(ctx, inv2); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	// No expiry — should be last
	inv3 := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		BatchNo: "FEFO-NONE", Qty: 50, Status: domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, inv3); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	results, err := invRepo.GetExpiringInventory(ctx, repository.InventoryRetrievalFilter{
		SKUID: sku.ID,
	})
	if err != nil {
		t.Fatalf("GetExpiringInventory failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].BatchNo != "FEFO-EARLY" {
		t.Errorf("first should be FEFO-EARLY (earliest expiry), got %s", results[0].BatchNo)
	}
	if results[1].BatchNo != "FEFO-LATE" {
		t.Errorf("second should be FEFO-LATE, got %s", results[1].BatchNo)
	}
	if results[2].BatchNo != "FEFO-NONE" {
		t.Errorf("third should be FEFO-NONE (no expiry, last), got %s", results[2].BatchNo)
	}
}

func TestInventoryRepo_GetExpiringInventory_WarehouseFilter(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)

	wh1, loc1 := createTestWarehouseZoneLocation(t, ctx, whRepo)
	wh2, loc2 := createTestWarehouseZoneLocation(t, ctx, whRepo)

	sku := &domain.SKU{
		Code: "TEST-SKU-FEFO-WH-001",
		Name: "FEFO WH Filter SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	inv1 := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc1.ID, WarehouseID: wh1.ID,
		BatchNo: "WH1-BATCH", Qty: 100, Status: domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, inv1); err != nil {
		t.Fatalf("CreateInventory wh1 failed: %v", err)
	}

	inv2 := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc2.ID, WarehouseID: wh2.ID,
		BatchNo: "WH2-BATCH", Qty: 200, Status: domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, inv2); err != nil {
		t.Fatalf("CreateInventory wh2 failed: %v", err)
	}

	results, err := invRepo.GetExpiringInventory(ctx, repository.InventoryRetrievalFilter{
		SKUID:       sku.ID,
		WarehouseID: wh1.ID,
	})
	if err != nil {
		t.Fatalf("GetExpiringInventory failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for wh1, got %d", len(results))
	}
	if results[0].WarehouseID != wh1.ID {
		t.Errorf("warehouse_id = %s, want %s", results[0].WarehouseID, wh1.ID)
	}
}

func TestInventoryRepo_GetOldestInventory_Limit(t *testing.T) {
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
		Code: "TEST-SKU-FIFO-LIMIT-001",
		Name: "FIFO Limit SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	for i := range 5 {
		inv := &domain.Inventory{
			SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
			BatchNo: "FIFO-LIMIT-" + string(rune('A'+i)), Qty: float64((i+1)*10),
			Status: domain.InventoryStatusAvailable,
		}
		if err := invRepo.CreateInventory(ctx, inv); err != nil {
			t.Fatalf("CreateInventory %d failed: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	results, err := invRepo.GetOldestInventory(ctx, repository.InventoryRetrievalFilter{
		SKUID: sku.ID,
		Limit: 3,
	})
	if err != nil {
		t.Fatalf("GetOldestInventory with limit failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results (limit=3), got %d", len(results))
	}
}

// ── Additional Inventory Repo Tests ───────────────────────

func TestInventoryRepo_UpdateInventoryStatus(t *testing.T) {
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
		Code: "TEST-SKU-ISTAT-" + uuid.New().String()[:8],
		Name: "Status Update SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	inv := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		Qty: 50.0, Status: domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, inv); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	// Transition to quarantine
	if err := invRepo.UpdateInventoryStatus(ctx, inv.ID, domain.InventoryStatusQuarantine); err != nil {
		t.Fatalf("UpdateInventoryStatus -> quarantine failed: %v", err)
	}
	got, err := invRepo.GetInventory(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetInventory failed: %v", err)
	}
	if got.Status != domain.InventoryStatusQuarantine {
		t.Errorf("status = %q, want quarantine", got.Status)
	}

	// Transition to damaged
	if err := invRepo.UpdateInventoryStatus(ctx, inv.ID, domain.InventoryStatusDamaged); err != nil {
		t.Fatalf("UpdateInventoryStatus -> damaged failed: %v", err)
	}
	got, err = invRepo.GetInventory(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetInventory failed: %v", err)
	}
	if got.Status != domain.InventoryStatusDamaged {
		t.Errorf("status = %q, want damaged", got.Status)
	}

	// Not found
	err = invRepo.UpdateInventoryStatus(ctx, uuid.New(), domain.InventoryStatusExpired)
	if err == nil {
		t.Error("expected error for nonexistent inventory")
	}
}

func TestInventoryRepo_GetAndLockInventory(t *testing.T) {
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
		Code: "TEST-SKU-LOCK-" + uuid.New().String()[:8],
		Name: "Lock SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	inv := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		Qty: 100.0, Status: domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, inv); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	// GetAndLockInventory may be called outside tx for integration testing
	got, err := invRepo.GetAndLockInventory(ctx, inv.ID)
	if err != nil {
		t.Fatalf("GetAndLockInventory failed: %v", err)
	}
	if got.ID != inv.ID {
		t.Errorf("id = %s, want %s", got.ID, inv.ID)
	}
	if got.Qty != 100.0 {
		t.Errorf("qty = %f, want 100.0", got.Qty)
	}

	// Not found with lock
	_, err = invRepo.GetAndLockInventory(ctx, uuid.New())
	if err == nil {
		t.Error("expected error for nonexistent inventory with lock")
	}
}

func TestInventoryRepo_CountInventory(t *testing.T) {
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
		Code: "TEST-SKU-CNT-" + uuid.New().String()[:8],
		Name: "Count Inventory SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	// No filter
	initial, err := invRepo.CountInventory(ctx, repository.InventoryFilter{})
	if err != nil {
		t.Fatalf("CountInventory failed: %v", err)
	}

	// Create 3 inventory records
	for i := range 3 {
		inv := &domain.Inventory{
			SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
			Qty: float64((i + 1) * 10), Status: domain.InventoryStatusAvailable,
		}
		if err := invRepo.CreateInventory(ctx, inv); err != nil {
			t.Fatalf("CreateInventory failed: %v", err)
		}
	}

	// Count by warehouse
	count, err := invRepo.CountInventory(ctx, repository.InventoryFilter{WarehouseID: wh.ID})
	if err != nil {
		t.Fatalf("CountInventory by warehouse failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 inventory records, got %d", count)
	}

	// Count by SKU
	count, err = invRepo.CountInventory(ctx, repository.InventoryFilter{SKUID: sku.ID})
	if err != nil {
		t.Fatalf("CountInventory by SKU failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 inventory records for SKU, got %d", count)
	}

	// Total should include initial + 3
	total, err := invRepo.CountInventory(ctx, repository.InventoryFilter{})
	if err != nil {
		t.Fatalf("CountInventory failed: %v", err)
	}
	if total != initial+3 {
		t.Errorf("expected %d total, got %d", initial+3, total)
	}
}

func TestInventoryRepo_CountTransactions(t *testing.T) {
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
		Code: "TEST-SKU-TXCNT-" + uuid.New().String()[:8],
		Name: "TX Count SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	inv := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		Qty: 0.0, Status: domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, inv); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	// Create 3 transactions
	for i := range 3 {
		tx := &domain.InventoryTransaction{
			InventoryID: inv.ID, SKUID: sku.ID, LocationID: loc.ID,
			Type: domain.InventoryTxReceipt, DeltaQty: float64(10 * (i + 1)),
			ResultingQty: float64(10 * (i + 1)), ReferenceType: "test",
			ReferenceID: uuid.New(),
		}
		if err := invRepo.CreateTransaction(ctx, tx); err != nil {
			t.Fatalf("CreateTransaction failed: %v", err)
		}
	}

	count, err := invRepo.CountTransactions(ctx, inv.ID)
	if err != nil {
		t.Fatalf("CountTransactions failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 transactions, got %d", count)
	}

	// Zero for unrelated inventory
	count, err = invRepo.CountTransactions(ctx, uuid.New())
	if err != nil {
		t.Fatalf("CountTransactions for unknown failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 transactions for unknown inventory, got %d", count)
	}
}

func TestInventoryRepo_ListTransactionsByReference(t *testing.T) {
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
		Code: "TEST-SKU-TXREF-" + uuid.New().String()[:8],
		Name: "TX Ref SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	inv := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		Qty: 0.0, Status: domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, inv); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	orderID := uuid.New()
	// Create 2 transactions with the same reference, 1 with different
	for i := range 2 {
		tx := &domain.InventoryTransaction{
			InventoryID: inv.ID, SKUID: sku.ID, LocationID: loc.ID,
			Type: domain.InventoryTxReserve, DeltaQty: -float64(5 * (i + 1)),
			ResultingQty: 95.0 - float64(5*i),
			ReferenceType: "order_line", ReferenceID: orderID,
		}
		if err := invRepo.CreateTransaction(ctx, tx); err != nil {
			t.Fatalf("CreateTransaction failed: %v", err)
		}
	}
	{
		tx := &domain.InventoryTransaction{
			InventoryID: inv.ID, SKUID: sku.ID, LocationID: loc.ID,
			Type: domain.InventoryTxPick, DeltaQty: -10.0,
			ResultingQty: 85.0, ReferenceType: "task",
			ReferenceID: uuid.New(),
		}
		if err := invRepo.CreateTransaction(ctx, tx); err != nil {
			t.Fatalf("CreateTransaction failed: %v", err)
		}
	}

	txs, err := invRepo.ListTransactionsByReference(ctx, "order_line", orderID)
	if err != nil {
		t.Fatalf("ListTransactionsByReference failed: %v", err)
	}
	if len(txs) != 2 {
		t.Errorf("expected 2 transactions for reference, got %d", len(txs))
	}

	txs, err = invRepo.ListTransactionsByReference(ctx, "nonexistent", orderID)
	if err != nil {
		t.Fatalf("ListTransactionsByReference for nonexistent failed: %v", err)
	}
	if len(txs) != 0 {
		t.Errorf("expected 0 transactions for nonexistent reference, got %d", len(txs))
	}
}

func TestInventoryRepo_GetInventoryDashboardStats(t *testing.T) {
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
		Code: "TEST-SKU-DASH-" + uuid.New().String()[:8],
		Name: "Dashboard SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	available := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		Qty: 200.0, ReservedQty: 50.0, Status: domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, available); err != nil {
		t.Fatalf("CreateInventory available failed: %v", err)
	}

	quarantine := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		Qty: 30.0, Status: domain.InventoryStatusQuarantine,
	}
	if err := invRepo.CreateInventory(ctx, quarantine); err != nil {
		t.Fatalf("CreateInventory quarantine failed: %v", err)
	}

	lowStock := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		Qty: 5.0, Status: domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, lowStock); err != nil {
		t.Fatalf("CreateInventory lowStock failed: %v", err)
	}

	stats, err := invRepo.GetInventoryDashboardStats(ctx, wh.ID, 10.0)
	if err != nil {
		t.Fatalf("GetInventoryDashboardStats failed: %v", err)
	}
	if stats.TotalRecords != 3 {
		t.Errorf("total_records = %d, want 3", stats.TotalRecords)
	}
	if stats.AvailableCount != 2 {
		t.Errorf("available_count = %d, want 2", stats.AvailableCount)
	}
	if stats.QuarantineCount != 1 {
		t.Errorf("quarantine_count = %d, want 1", stats.QuarantineCount)
	}
	if stats.LowStockCount < 1 {
		t.Errorf("low_stock_count = %d, want at least 1", stats.LowStockCount)
	}

	allStats, err := invRepo.GetInventoryDashboardStats(ctx, uuid.Nil, 10.0)
	if err != nil {
		t.Fatalf("GetInventoryDashboardStats (all) failed: %v", err)
	}
	if allStats.TotalRecords < 3 {
		t.Errorf("total_records (all) = %d, want at least 3", allStats.TotalRecords)
	}
}

func TestInventoryRepo_GetLowStockInventory(t *testing.T) {
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
		Code: "TEST-SKU-LOW-" + uuid.New().String()[:8],
		Name: "Low Stock SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	low := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		Qty: 5.0, Status: domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, low); err != nil {
		t.Fatalf("CreateInventory low failed: %v", err)
	}

	high := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		Qty: 500.0, Status: domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, high); err != nil {
		t.Fatalf("CreateInventory high failed: %v", err)
	}

	results, err := invRepo.GetLowStockInventory(ctx, 10.0, wh.ID, 0)
	if err != nil {
		t.Fatalf("GetLowStockInventory failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 low stock record, got %d", len(results))
	}
	if results[0].Qty != 5.0 {
		t.Errorf("qty = %f, want 5.0", results[0].Qty)
	}
}

func TestInventoryRepo_GetInventoryByWarehouse(t *testing.T) {
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
		Code: "TEST-SKU-BYWH-" + uuid.New().String()[:8],
		Name: "By Warehouse SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := invRepo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	inv := &domain.Inventory{
		SKUID: sku.ID, LocationID: loc.ID, WarehouseID: wh.ID,
		Qty: 100.0, ReservedQty: 20.0, Status: domain.InventoryStatusAvailable,
	}
	if err := invRepo.CreateInventory(ctx, inv); err != nil {
		t.Fatalf("CreateInventory failed: %v", err)
	}

	rows, err := invRepo.GetInventoryByWarehouse(ctx)
	if err != nil {
		t.Fatalf("GetInventoryByWarehouse failed: %v", err)
	}

	found := false
	for _, row := range rows {
		if row.WarehouseID == wh.ID {
			found = true
			if row.TotalQty < 100.0 {
				t.Errorf("total_qty = %f, want at least 100.0", row.TotalQty)
			}
			if row.RecordCount < 1 {
				t.Errorf("record_count = %d, want at least 1", row.RecordCount)
			}
			break
		}
	}
	if !found {
		t.Error("expected to find warehouse in inventory by warehouse rows")
	}
}

func TestInventoryRepo_CountSKUs(t *testing.T) {
	db, cleanup := setupInventoryTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewInventoryRepo(db)

	initial, err := repo.CountSKUs(ctx)
	if err != nil {
		t.Fatalf("CountSKUs failed: %v", err)
	}

	s := &domain.SKU{
		Code: "TEST-SKU-CNT-" + uuid.New().String()[:8],
		Name: "Count SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := repo.CreateSKU(ctx, s); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}

	count, err := repo.CountSKUs(ctx)
	if err != nil {
		t.Fatalf("CountSKUs failed: %v", err)
	}
	if count != initial+1 {
		t.Errorf("expected %d SKUs, got %d", initial+1, count)
	}
}
