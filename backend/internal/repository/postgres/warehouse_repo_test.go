package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// setupTestDB creates a test database connection or skips the test if unavailable.
func setupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	cfg := testConfig()

	ctx := context.Background()
	db, err := NewDB(ctx, cfg)
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

	warehouses, err := repo.ListWarehouses(ctx, 0, 0)
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

	zones, err := repo.ListZonesByWarehouse(ctx, wh.ID, 0, 0)
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

// ── Additional WarehouseRepo Tests ────────────────────────────

func TestWarehouseRepo_CountZonesByWarehouse(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	wh := &domain.Warehouse{Code: "TEST-WH-CZW-" + uuid.New().String()[:8], Name: "Count Zones WH"}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	count, err := repo.CountZonesByWarehouse(ctx, wh.ID)
	if err != nil {
		t.Fatalf("CountZonesByWarehouse failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 zones initially, got %d", count)
	}

	for i := 0; i < 3; i++ {
		z := &domain.Zone{
			WarehouseID: wh.ID, Code: "TEST-ZONE-CNT-" + uuid.New().String()[:8],
			Name: "Count Zone", ZoneType: domain.ZoneTypeStorage,
		}
		if err := repo.CreateZone(ctx, z); err != nil {
			t.Fatalf("CreateZone failed: %v", err)
		}
	}

	count, err = repo.CountZonesByWarehouse(ctx, wh.ID)
	if err != nil {
		t.Fatalf("CountZonesByWarehouse failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 zones, got %d", count)
	}
}

func TestWarehouseRepo_ListAllZones(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	wh1 := &domain.Warehouse{Code: "TEST-WH-LAZ1-" + uuid.New().String()[:8], Name: "List All Zones WH1"}
	wh2 := &domain.Warehouse{Code: "TEST-WH-LAZ2-" + uuid.New().String()[:8], Name: "List All Zones WH2"}
	if err := repo.CreateWarehouse(ctx, wh1); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}
	if err := repo.CreateWarehouse(ctx, wh2); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	for i := 0; i < 2; i++ {
		z := &domain.Zone{
			WarehouseID: wh1.ID, Code: "TEST-ZONE-AZ1-" + uuid.New().String()[:8],
			Name: "Zone WH1", ZoneType: domain.ZoneTypeStorage,
		}
		if err := repo.CreateZone(ctx, z); err != nil {
			t.Fatalf("CreateZone failed: %v", err)
		}
	}
	{
		z := &domain.Zone{
			WarehouseID: wh2.ID, Code: "TEST-ZONE-AZ2-" + uuid.New().String()[:8],
			Name: "Zone WH2", ZoneType: domain.ZoneTypePicking,
		}
		if err := repo.CreateZone(ctx, z); err != nil {
			t.Fatalf("CreateZone failed: %v", err)
		}
	}

	all, err := repo.ListAllZones(ctx, repository.ZoneFilter{})
	if err != nil {
		t.Fatalf("ListAllZones failed: %v", err)
	}
	if len(all) < 3 {
		t.Errorf("expected at least 3 zones total, got %d", len(all))
	}

	byWH1, err := repo.ListAllZones(ctx, repository.ZoneFilter{WarehouseID: wh1.ID})
	if err != nil {
		t.Fatalf("ListAllZones by warehouse failed: %v", err)
	}
	if len(byWH1) != 2 {
		t.Errorf("expected 2 zones for wh1, got %d", len(byWH1))
	}

	limited, err := repo.ListAllZones(ctx, repository.ZoneFilter{WarehouseID: wh1.ID, Limit: 1})
	if err != nil {
		t.Fatalf("ListAllZones with limit failed: %v", err)
	}
	if len(limited) != 1 {
		t.Errorf("expected 1 zone with limit, got %d", len(limited))
	}
}

func TestWarehouseRepo_CountAllZones(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	wh := &domain.Warehouse{Code: "TEST-WH-CAZ-" + uuid.New().String()[:8], Name: "Count All Zones WH"}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	initial, err := repo.CountAllZones(ctx, repository.ZoneFilter{})
	if err != nil {
		t.Fatalf("CountAllZones failed: %v", err)
	}

	for i := 0; i < 2; i++ {
		z := &domain.Zone{
			WarehouseID: wh.ID, Code: "TEST-ZONE-CAC-" + uuid.New().String()[:8],
			Name: "Count All Zone", ZoneType: domain.ZoneTypeStorage,
		}
		if err := repo.CreateZone(ctx, z); err != nil {
			t.Fatalf("CreateZone failed: %v", err)
		}
	}

	count, err := repo.CountAllZones(ctx, repository.ZoneFilter{WarehouseID: wh.ID})
	if err != nil {
		t.Fatalf("CountAllZones by warehouse failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 zones for warehouse, got %d", count)
	}

	total, err := repo.CountAllZones(ctx, repository.ZoneFilter{})
	if err != nil {
		t.Fatalf("CountAllZones failed: %v", err)
	}
	if total < initial+2 {
		t.Errorf("expected at least %d total, got %d", initial+2, total)
	}
}

func TestWarehouseRepo_UpdateZone(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	wh := &domain.Warehouse{Code: "TEST-WH-UZ-" + uuid.New().String()[:8], Name: "Update Zone WH"}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	z := &domain.Zone{
		WarehouseID: wh.ID, Code: "TEST-ZONE-UPD-" + uuid.New().String()[:8],
		Name: "Original Zone Name", ZoneType: domain.ZoneTypeStorage,
	}
	if err := repo.CreateZone(ctx, z); err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}

	z.Name = "Updated Zone Name"
	z.ZoneType = domain.ZoneTypePicking
	z.Status = domain.ZoneStatusInactive

	if err := repo.UpdateZone(ctx, z); err != nil {
		t.Fatalf("UpdateZone failed: %v", err)
	}

	got, err := repo.GetZone(ctx, z.ID)
	if err != nil {
		t.Fatalf("GetZone failed: %v", err)
	}
	if got.Name != "Updated Zone Name" {
		t.Errorf("name = %q, want %q", got.Name, "Updated Zone Name")
	}
	if got.ZoneType != domain.ZoneTypePicking {
		t.Errorf("zone_type = %q, want %q", got.ZoneType, domain.ZoneTypePicking)
	}
	if got.Status != domain.ZoneStatusInactive {
		t.Errorf("status = %q, want %q", got.Status, domain.ZoneStatusInactive)
	}
}

func TestWarehouseRepo_ListLocationsByZone(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	wh := &domain.Warehouse{Code: "TEST-WH-LLZ-" + uuid.New().String()[:8], Name: "List Locs WH"}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	zone := &domain.Zone{
		WarehouseID: wh.ID, Code: "TEST-ZONE-LLZ-" + uuid.New().String()[:8],
		Name: "List Locs Zone", ZoneType: domain.ZoneTypeStorage,
	}
	if err := repo.CreateZone(ctx, zone); err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		loc := &domain.Location{
			ZoneID: zone.ID, WarehouseID: wh.ID,
			Code: "TEST-LOC-LLZ-" + uuid.New().String()[:8],
			LocationType: domain.LocationTypeShelf,
		}
		if err := repo.CreateLocation(ctx, loc); err != nil {
			t.Fatalf("CreateLocation failed: %v", err)
		}
	}

	locs, err := repo.ListLocationsByZone(ctx, zone.ID, 0, 0)
	if err != nil {
		t.Fatalf("ListLocationsByZone failed: %v", err)
	}
	if len(locs) != 3 {
		t.Errorf("expected 3 locations, got %d", len(locs))
	}

	limited, err := repo.ListLocationsByZone(ctx, zone.ID, 1, 0)
	if err != nil {
		t.Fatalf("ListLocationsByZone with limit failed: %v", err)
	}
	if len(limited) != 1 {
		t.Errorf("expected 1 location with limit, got %d", len(limited))
	}
}

func TestWarehouseRepo_CountLocationsByZone(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	wh := &domain.Warehouse{Code: "TEST-WH-CLZ-" + uuid.New().String()[:8], Name: "Count Locs WH"}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	zone := &domain.Zone{
		WarehouseID: wh.ID, Code: "TEST-ZONE-CLZ-" + uuid.New().String()[:8],
		Name: "Count Locs Zone", ZoneType: domain.ZoneTypeStorage,
	}
	if err := repo.CreateZone(ctx, zone); err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}

	count, err := repo.CountLocationsByZone(ctx, zone.ID)
	if err != nil {
		t.Fatalf("CountLocationsByZone failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 locations initially, got %d", count)
	}

	for i := 0; i < 2; i++ {
		loc := &domain.Location{
			ZoneID: zone.ID, WarehouseID: wh.ID,
			Code: "TEST-LOC-CLZ-" + uuid.New().String()[:8],
			LocationType: domain.LocationTypeShelf,
		}
		if err := repo.CreateLocation(ctx, loc); err != nil {
			t.Fatalf("CreateLocation failed: %v", err)
		}
	}

	count, err = repo.CountLocationsByZone(ctx, zone.ID)
	if err != nil {
		t.Fatalf("CountLocationsByZone failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 locations, got %d", count)
	}
}

func TestWarehouseRepo_ListAllLocations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	wh1 := &domain.Warehouse{Code: "TEST-WH-LAL1-" + uuid.New().String()[:8], Name: "List All Locs WH1"}
	wh2 := &domain.Warehouse{Code: "TEST-WH-LAL2-" + uuid.New().String()[:8], Name: "List All Locs WH2"}
	if err := repo.CreateWarehouse(ctx, wh1); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}
	if err := repo.CreateWarehouse(ctx, wh2); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	zone1 := &domain.Zone{
		WarehouseID: wh1.ID, Code: "TEST-ZONE-LAL1-" + uuid.New().String()[:8],
		Name: "LAL Zone 1", ZoneType: domain.ZoneTypeStorage,
	}
	zone2 := &domain.Zone{
		WarehouseID: wh2.ID, Code: "TEST-ZONE-LAL2-" + uuid.New().String()[:8],
		Name: "LAL Zone 2", ZoneType: domain.ZoneTypePicking,
	}
	if err := repo.CreateZone(ctx, zone1); err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}
	if err := repo.CreateZone(ctx, zone2); err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}

	for i := 0; i < 2; i++ {
		loc := &domain.Location{
			ZoneID: zone1.ID, WarehouseID: wh1.ID,
			Code: "TEST-LOC-LAL1-" + uuid.New().String()[:8],
			LocationType: domain.LocationTypeShelf,
		}
		if err := repo.CreateLocation(ctx, loc); err != nil {
			t.Fatalf("CreateLocation failed: %v", err)
		}
	}
	{
		loc := &domain.Location{
			ZoneID: zone2.ID, WarehouseID: wh2.ID,
			Code: "TEST-LOC-LAL2-" + uuid.New().String()[:8],
			LocationType: domain.LocationTypePallet,
		}
		if err := repo.CreateLocation(ctx, loc); err != nil {
			t.Fatalf("CreateLocation failed: %v", err)
		}
	}

	all, err := repo.ListAllLocations(ctx, repository.LocationFilter{})
	if err != nil {
		t.Fatalf("ListAllLocations failed: %v", err)
	}
	if len(all) < 3 {
		t.Errorf("expected at least 3 locations total, got %d", len(all))
	}

	byZone, err := repo.ListAllLocations(ctx, repository.LocationFilter{ZoneID: zone1.ID})
	if err != nil {
		t.Fatalf("ListAllLocations by zone failed: %v", err)
	}
	if len(byZone) != 2 {
		t.Errorf("expected 2 locations for zone1, got %d", len(byZone))
	}

	byWH, err := repo.ListAllLocations(ctx, repository.LocationFilter{WarehouseID: wh2.ID})
	if err != nil {
		t.Fatalf("ListAllLocations by warehouse failed: %v", err)
	}
	if len(byWH) != 1 {
		t.Errorf("expected 1 location for wh2, got %d", len(byWH))
	}

	limited, err := repo.ListAllLocations(ctx, repository.LocationFilter{ZoneID: zone1.ID, Limit: 1})
	if err != nil {
		t.Fatalf("ListAllLocations with limit failed: %v", err)
	}
	if len(limited) != 1 {
		t.Errorf("expected 1 location with limit, got %d", len(limited))
	}
}

func TestWarehouseRepo_CountAllLocations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	wh := &domain.Warehouse{Code: "TEST-WH-CAL-" + uuid.New().String()[:8], Name: "Count All Locs WH"}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	zone := &domain.Zone{
		WarehouseID: wh.ID, Code: "TEST-ZONE-CAL-" + uuid.New().String()[:8],
		Name: "Count All Locs Zone", ZoneType: domain.ZoneTypeStorage,
	}
	if err := repo.CreateZone(ctx, zone); err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}

	initial, err := repo.CountAllLocations(ctx, repository.LocationFilter{})
	if err != nil {
		t.Fatalf("CountAllLocations failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		loc := &domain.Location{
			ZoneID: zone.ID, WarehouseID: wh.ID,
			Code: "TEST-LOC-CAL-" + uuid.New().String()[:8],
			LocationType: domain.LocationTypeShelf,
		}
		if err := repo.CreateLocation(ctx, loc); err != nil {
			t.Fatalf("CreateLocation failed: %v", err)
		}
	}

	count, err := repo.CountAllLocations(ctx, repository.LocationFilter{WarehouseID: wh.ID})
	if err != nil {
		t.Fatalf("CountAllLocations by warehouse failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 locations for warehouse, got %d", count)
	}

	total, err := repo.CountAllLocations(ctx, repository.LocationFilter{})
	if err != nil {
		t.Fatalf("CountAllLocations failed: %v", err)
	}
	if total < initial+3 {
		t.Errorf("expected at least %d total, got %d", initial+3, total)
	}
}

func TestWarehouseRepo_UpdateLocation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	wh := &domain.Warehouse{Code: "TEST-WH-UL-" + uuid.New().String()[:8], Name: "Update Loc WH"}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	zone := &domain.Zone{
		WarehouseID: wh.ID, Code: "TEST-ZONE-UL-" + uuid.New().String()[:8],
		Name: "Update Loc Zone", ZoneType: domain.ZoneTypeStorage,
	}
	if err := repo.CreateZone(ctx, zone); err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}

	loc := &domain.Location{
		ZoneID: zone.ID, WarehouseID: wh.ID,
		Code: "TEST-LOC-UL-" + uuid.New().String()[:8],
		Barcode: "BC-UL-ORIG", LocationType: domain.LocationTypeShelf,
		Capacity: &domain.Capacity{MaxWeight: 100.0, MaxVolume: 1.0, MaxQty: 10},
	}
	if err := repo.CreateLocation(ctx, loc); err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}

	loc.Code = "TEST-LOC-UL-UPDATED"
	loc.Barcode = "BC-UL-UPDATED"
	loc.LocationType = domain.LocationTypePallet
	loc.Capacity = &domain.Capacity{MaxWeight: 500.0, MaxVolume: 5.0, MaxQty: 50}

	if err := repo.UpdateLocation(ctx, loc); err != nil {
		t.Fatalf("UpdateLocation failed: %v", err)
	}

	got, err := repo.GetLocation(ctx, loc.ID)
	if err != nil {
		t.Fatalf("GetLocation failed: %v", err)
	}
	if got.Code != "TEST-LOC-UL-UPDATED" {
		t.Errorf("code = %q, want %q", got.Code, "TEST-LOC-UL-UPDATED")
	}
	if got.Barcode != "BC-UL-UPDATED" {
		t.Errorf("barcode = %q, want %q", got.Barcode, "BC-UL-UPDATED")
	}
	if got.LocationType != domain.LocationTypePallet {
		t.Errorf("location_type = %q, want pallet", got.LocationType)
	}
	if got.Capacity == nil || got.Capacity.MaxWeight != 500.0 {
		t.Errorf("capacity.MaxWeight = %v, want 500.0", got.Capacity)
	}
	if got.Capacity == nil || got.Capacity.MaxQty != 50 {
		t.Errorf("capacity.MaxQty = %v, want 50", got.Capacity)
	}
}

func TestWarehouseRepo_CountWarehouses(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewWarehouseRepo(db)

	initial, err := repo.CountWarehouses(ctx)
	if err != nil {
		t.Fatalf("CountWarehouses failed: %v", err)
	}

	w := &domain.Warehouse{
		Code: "TEST-WH-CNT-" + uuid.New().String()[:8],
		Name: "Count Warehouse",
	}
	if err := repo.CreateWarehouse(ctx, w); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}

	count, err := repo.CountWarehouses(ctx)
	if err != nil {
		t.Fatalf("CountWarehouses failed: %v", err)
	}
	if count != initial+1 {
		t.Errorf("expected %d warehouses, got %d", initial+1, count)
	}
}
