package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
)

// setupTestDB creates a test database connection or skips the test if unavailable.
func setupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://wms:wms_dev_2026@localhost:5432/wms?sslmode=disable"
	}

	ctx := context.Background()
	db, err := NewDB(ctx, dsn)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
		return nil, nil
	}

	// Clean up any previous test data
	db.Pool.Exec(ctx, "DELETE FROM locations WHERE code LIKE 'TEST-%'")
	db.Pool.Exec(ctx, "DELETE FROM zones WHERE code LIKE 'TEST-%'")
	db.Pool.Exec(ctx, "DELETE FROM warehouses WHERE code LIKE 'TEST-%'")

	cleanup := func() {
		db.Pool.Exec(ctx, "DELETE FROM locations WHERE code LIKE 'TEST-%'")
		db.Pool.Exec(ctx, "DELETE FROM zones WHERE code LIKE 'TEST-%'")
		db.Pool.Exec(ctx, "DELETE FROM warehouses WHERE code LIKE 'TEST-%'")
		db.Close()
	}

	return db, cleanup
}

func TestWarehouseRepo_CreateAndGetWarehouse(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	w := &domain.Warehouse{
		Code:    "TEST-WH-001",
		Name:    "Test Warehouse",
		Address: "123 Test St",
		Status:  domain.WarehouseStatusActive,
	}

	// Create
	err := repo.CreateWarehouse(ctx, w)
	if err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}
	if w.ID == uuid.Nil {
		t.Error("expected warehouse ID to be set")
	}

	// Get
	got, err := repo.GetWarehouse(ctx, w.ID)
	if err != nil {
		t.Fatalf("GetWarehouse failed: %v", err)
	}
	if got.Code != w.Code {
		t.Errorf("code = %q, want %q", got.Code, w.Code)
	}
	if got.Name != w.Name {
		t.Errorf("name = %q, want %q", got.Name, w.Name)
	}
}

func TestWarehouseRepo_ListWarehouses(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	// Create test warehouses
	for i := 0; i < 3; i++ {
		w := &domain.Warehouse{
			Code: "TEST-WH-LIST-" + uuid.New().String()[:8],
			Name: "List Test Warehouse",
		}
		if err := repo.CreateWarehouse(ctx, w); err != nil {
			t.Fatalf("CreateWarehouse failed: %v", err)
		}
	}

	warehouses, err := repo.ListWarehouses(ctx)
	if err != nil {
		t.Fatalf("ListWarehouses failed: %v", err)
	}
	if len(warehouses) < 3 {
		t.Errorf("expected at least 3 warehouses, got %d", len(warehouses))
	}
}

func TestWarehouseRepo_UpdateWarehouse(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	w := &domain.Warehouse{
		Code: "TEST-WH-UPD-001",
		Name: "Original Name",
	}
	if err := repo.CreateWarehouse(ctx, w); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	w.Name = "Updated Name"
	w.Status = domain.WarehouseStatusInactive

	if err := repo.UpdateWarehouse(ctx, w); err != nil {
		t.Fatalf("UpdateWarehouse failed: %v", err)
	}

	got, err := repo.GetWarehouse(ctx, w.ID)
	if err != nil {
		t.Fatalf("GetWarehouse failed: %v", err)
	}
	if got.Name != "Updated Name" {
		t.Errorf("name = %q, want %q", got.Name, "Updated Name")
	}
	if got.Status != domain.WarehouseStatusInactive {
		t.Errorf("status = %q, want %q", got.Status, domain.WarehouseStatusInactive)
	}
}

func TestWarehouseRepo_CreateAndGetZone(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	// Need a warehouse first
	wh := &domain.Warehouse{Code: "TEST-WH-ZONE", Name: "Zone Test WH"}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	z := &domain.Zone{
		WarehouseID: wh.ID,
		Code:        "TEST-ZONE-001",
		Name:        "Test Zone",
		ZoneType:    domain.ZoneTypeStorage,
	}

	err := repo.CreateZone(ctx, z)
	if err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}
	if z.ID == uuid.Nil {
		t.Error("expected zone ID to be set")
	}

	got, err := repo.GetZone(ctx, z.ID)
	if err != nil {
		t.Fatalf("GetZone failed: %v", err)
	}
	if got.Code != z.Code {
		t.Errorf("code = %q, want %q", got.Code, z.Code)
	}
	if got.ZoneType != domain.ZoneTypeStorage {
		t.Errorf("zone_type = %q, want %q", got.ZoneType, domain.ZoneTypeStorage)
	}
}

func TestWarehouseRepo_ListZonesByWarehouse(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	wh := &domain.Warehouse{Code: "TEST-WH-ZLIST", Name: "Zone List WH"}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		z := &domain.Zone{
			WarehouseID: wh.ID,
			Code:        "TEST-ZONE-LIST-" + uuid.New().String()[:8],
			Name:        "List Zone",
			ZoneType:    domain.ZoneTypePicking,
		}
		if err := repo.CreateZone(ctx, z); err != nil {
			t.Fatalf("CreateZone failed: %v", err)
		}
	}

	zones, err := repo.ListZonesByWarehouse(ctx, wh.ID)
	if err != nil {
		t.Fatalf("ListZonesByWarehouse failed: %v", err)
	}
	if len(zones) != 3 {
		t.Errorf("expected 3 zones, got %d", len(zones))
	}
}

func TestWarehouseRepo_CreateAndGetLocation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	wh := &domain.Warehouse{Code: "TEST-WH-LOC", Name: "Location Test WH"}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	zone := &domain.Zone{
		WarehouseID: wh.ID,
		Code:        "TEST-ZONE-LOC",
		Name:        "Location Zone",
		ZoneType:    domain.ZoneTypeStorage,
	}
	if err := repo.CreateZone(ctx, zone); err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}

	loc := &domain.Location{
		ZoneID:       zone.ID,
		WarehouseID:  wh.ID,
		Code:         "TEST-LOC-001",
		Barcode:      "BARCODE-001",
		LocationType: domain.LocationTypeShelf,
		Capacity: &domain.Capacity{
			MaxWeight: 500.0,
			MaxVolume: 2.5,
			MaxQty:    100,
		},
	}

	err := repo.CreateLocation(ctx, loc)
	if err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}
	if loc.ID == uuid.Nil {
		t.Error("expected location ID to be set")
	}

	got, err := repo.GetLocation(ctx, loc.ID)
	if err != nil {
		t.Fatalf("GetLocation failed: %v", err)
	}
	if got.Code != loc.Code {
		t.Errorf("code = %q, want %q", got.Code, loc.Code)
	}
	if got.Capacity == nil || got.Capacity.MaxWeight != 500.0 {
		t.Error("capacity not correctly retrieved")
	}
}

func TestWarehouseRepo_GetLocationByBarcode(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	wh := &domain.Warehouse{Code: "TEST-WH-BC", Name: "Barcode WH"}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	zone := &domain.Zone{
		WarehouseID: wh.ID, Code: "TEST-ZONE-BC", Name: "Barcode Zone", ZoneType: domain.ZoneTypeStorage,
	}
	if err := repo.CreateZone(ctx, zone); err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}

	loc := &domain.Location{
		ZoneID: zone.ID, WarehouseID: wh.ID,
		Code: "TEST-LOC-BC", Barcode: "UNIQUE-BARCODE-123", LocationType: domain.LocationTypePallet,
	}
	if err := repo.CreateLocation(ctx, loc); err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}

	got, err := repo.GetLocationByBarcode(ctx, "UNIQUE-BARCODE-123")
	if err != nil {
		t.Fatalf("GetLocationByBarcode failed: %v", err)
	}
	if got.ID != loc.ID {
		t.Errorf("id = %q, want %q", got.ID, loc.ID)
	}
}

func TestWarehouseRepo_UpdateLocationStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	wh := &domain.Warehouse{Code: "TEST-WH-LS", Name: "Location Status WH"}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	zone := &domain.Zone{
		WarehouseID: wh.ID, Code: "TEST-ZONE-LS", Name: "Status Zone", ZoneType: domain.ZoneTypeStorage,
	}
	if err := repo.CreateZone(ctx, zone); err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}

	loc := &domain.Location{
		ZoneID: zone.ID, WarehouseID: wh.ID,
		Code: "TEST-LOC-LS", LocationType: domain.LocationTypeShelf,
	}
	if err := repo.CreateLocation(ctx, loc); err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}

	if err := repo.UpdateLocationStatus(ctx, loc.ID, domain.LocationStatusOccupied); err != nil {
		t.Fatalf("UpdateLocationStatus failed: %v", err)
	}

	got, err := repo.GetLocation(ctx, loc.ID)
	if err != nil {
		t.Fatalf("GetLocation failed: %v", err)
	}
	if got.Status != domain.LocationStatusOccupied {
		t.Errorf("status = %q, want %q", got.Status, domain.LocationStatusOccupied)
	}
}
