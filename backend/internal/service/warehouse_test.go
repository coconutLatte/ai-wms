package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// mockWarehouseRepo implements repository.WarehouseRepository for testing.
type mockWarehouseRepo struct {
	warehouses map[uuid.UUID]*domain.Warehouse
	zones      map[uuid.UUID]*domain.Zone
	locations  map[uuid.UUID]*domain.Location
}

func newMockWarehouseRepo() *mockWarehouseRepo {
	return &mockWarehouseRepo{
		warehouses: make(map[uuid.UUID]*domain.Warehouse),
		zones:      make(map[uuid.UUID]*domain.Zone),
		locations:  make(map[uuid.UUID]*domain.Location),
	}
}

// ── Warehouse ─────────────────────────────────────

func (m *mockWarehouseRepo) CreateWarehouse(ctx context.Context, w *domain.Warehouse) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	m.warehouses[w.ID] = w
	return nil
}

func (m *mockWarehouseRepo) GetWarehouse(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	w, ok := m.warehouses[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("warehouse", id.String())
	}
	return w, nil
}

func (m *mockWarehouseRepo) ListWarehouses(ctx context.Context, limit, offset int) ([]*domain.Warehouse, error) {
	var result []*domain.Warehouse
	for _, w := range m.warehouses {
		result = append(result, w)
	}
	return result, nil
}

func (m *mockWarehouseRepo) UpdateWarehouse(ctx context.Context, w *domain.Warehouse) error {
	if _, ok := m.warehouses[w.ID]; !ok {
		return pkgerrors.NewNotFound("warehouse", w.ID.String())
	}
	m.warehouses[w.ID] = w
	return nil
}

func (m *mockWarehouseRepo) CountWarehouses(ctx context.Context) (int, error) {
	return len(m.warehouses), nil
}

// ── Zone ──────────────────────────────────────────

func (m *mockWarehouseRepo) CreateZone(ctx context.Context, z *domain.Zone) error {
	if z.ID == uuid.Nil {
		z.ID = uuid.New()
	}
	m.zones[z.ID] = z
	return nil
}

func (m *mockWarehouseRepo) GetZone(ctx context.Context, id uuid.UUID) (*domain.Zone, error) {
	z, ok := m.zones[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("zone", id.String())
	}
	return z, nil
}

func (m *mockWarehouseRepo) ListZonesByWarehouse(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]*domain.Zone, error) {
	var result []*domain.Zone
	for _, z := range m.zones {
		if z.WarehouseID == warehouseID {
			result = append(result, z)
		}
	}
	return result, nil
}

func (m *mockWarehouseRepo) CountZonesByWarehouse(ctx context.Context, warehouseID uuid.UUID) (int, error) {
	count := 0
	for _, z := range m.zones {
		if z.WarehouseID == warehouseID {
			count++
		}
	}
	return count, nil
}

// ── Location ──────────────────────────────────────

func (m *mockWarehouseRepo) CreateLocation(ctx context.Context, loc *domain.Location) error {
	if loc.ID == uuid.Nil {
		loc.ID = uuid.New()
	}
	m.locations[loc.ID] = loc
	return nil
}

func (m *mockWarehouseRepo) GetLocation(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	loc, ok := m.locations[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("location", id.String())
	}
	return loc, nil
}

func (m *mockWarehouseRepo) GetLocationByBarcode(ctx context.Context, barcode string) (*domain.Location, error) {
	for _, loc := range m.locations {
		if loc.Barcode == barcode {
			return loc, nil
		}
	}
	return nil, pkgerrors.NewNotFound("location", barcode)
}

func (m *mockWarehouseRepo) ListLocationsByZone(ctx context.Context, zoneID uuid.UUID, limit, offset int) ([]*domain.Location, error) {
	var result []*domain.Location
	for _, loc := range m.locations {
		if loc.ZoneID == zoneID {
			result = append(result, loc)
		}
	}
	return result, nil
}

func (m *mockWarehouseRepo) UpdateLocationStatus(ctx context.Context, id uuid.UUID, status domain.LocationStatus) error {
	loc, ok := m.locations[id]
	if !ok {
		return pkgerrors.NewNotFound("location", id.String())
	}
	loc.Status = status
	return nil
}

func (m *mockWarehouseRepo) CountLocationsByZone(ctx context.Context, zoneID uuid.UUID) (int, error) {
	count := 0
	for _, loc := range m.locations {
		if loc.ZoneID == zoneID {
			count++
		}
	}
	return count, nil
}

// ── Tests ───────────────────────────────────────────────────────────────────

func TestWarehouseService_CreateWarehouse(t *testing.T) {
	ctx := context.Background()
	svc := NewWarehouseService(newMockWarehouseRepo())

	w, err := svc.CreateWarehouse(ctx, CreateWarehouseInput{
		Code:    "WH-001",
		Name:    "Test Warehouse",
		Address: "123 Test St",
	})
	if err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}
	if w.Code != "WH-001" {
		t.Errorf("code = %q, want %q", w.Code, "WH-001")
	}
	if w.Status != domain.WarehouseStatusActive {
		t.Errorf("status = %q, want %q", w.Status, domain.WarehouseStatusActive)
	}
}

func TestWarehouseService_CreateWarehouse_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewWarehouseService(newMockWarehouseRepo())

	_, err := svc.CreateWarehouse(ctx, CreateWarehouseInput{Code: "", Name: "No Code"})
	if err == nil {
		t.Fatal("expected error for empty code")
	}

	_, err = svc.CreateWarehouse(ctx, CreateWarehouseInput{Code: "WH-001", Name: ""})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestWarehouseService_GetWarehouse(t *testing.T) {
	ctx := context.Background()
	svc := NewWarehouseService(newMockWarehouseRepo())

	w, _ := svc.CreateWarehouse(ctx, CreateWarehouseInput{Code: "WH-001", Name: "Test"})

	got, err := svc.GetWarehouse(ctx, w.ID)
	if err != nil {
		t.Fatalf("GetWarehouse failed: %v", err)
	}
	if got.Code != "WH-001" {
		t.Errorf("code = %q, want %q", got.Code, "WH-001")
	}
}

func TestWarehouseService_GetWarehouse_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewWarehouseService(newMockWarehouseRepo())

	_, err := svc.GetWarehouse(ctx, uuid.New())
	if err == nil {
		t.Fatal("expected error for unknown warehouse")
	}
}

func TestWarehouseService_ListWarehouses(t *testing.T) {
	ctx := context.Background()
	svc := NewWarehouseService(newMockWarehouseRepo())

	svc.CreateWarehouse(ctx, CreateWarehouseInput{Code: "WH-001", Name: "W1"})
	svc.CreateWarehouse(ctx, CreateWarehouseInput{Code: "WH-002", Name: "W2"})

	list, total, err := svc.ListWarehouses(ctx, 0, 0)
	if err != nil {
		t.Fatalf("ListWarehouses failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 warehouses, got %d", len(list))
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
}

func TestWarehouseService_UpdateWarehouse(t *testing.T) {
	ctx := context.Background()
	svc := NewWarehouseService(newMockWarehouseRepo())

	w, _ := svc.CreateWarehouse(ctx, CreateWarehouseInput{Code: "WH-001", Name: "Original"})

	newName := "Updated"
	newStatus := domain.WarehouseStatusInactive
	updated, err := svc.UpdateWarehouse(ctx, w.ID, UpdateWarehouseInput{
		Name:   &newName,
		Status: &newStatus,
	})
	if err != nil {
		t.Fatalf("UpdateWarehouse failed: %v", err)
	}
	if updated.Name != "Updated" {
		t.Errorf("name = %q, want %q", updated.Name, "Updated")
	}
	if updated.Status != domain.WarehouseStatusInactive {
		t.Errorf("status = %q, want %q", updated.Status, domain.WarehouseStatusInactive)
	}
}

func TestWarehouseService_UpdateWarehouse_InvalidStatus(t *testing.T) {
	ctx := context.Background()
	svc := NewWarehouseService(newMockWarehouseRepo())

	w, _ := svc.CreateWarehouse(ctx, CreateWarehouseInput{Code: "WH-001", Name: "Test"})

	badStatus := domain.WarehouseStatus("nonexistent")
	_, err := svc.UpdateWarehouse(ctx, w.ID, UpdateWarehouseInput{
		Status: &badStatus,
	})
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestWarehouseService_CreateZone(t *testing.T) {
	ctx := context.Background()
	repo := newMockWarehouseRepo()
	svc := NewWarehouseService(repo)

	wh, _ := svc.CreateWarehouse(ctx, CreateWarehouseInput{Code: "WH-001", Name: "Test WH"})

	z, err := svc.CreateZone(ctx, wh.ID, CreateZoneInput{
		Code:     "ZONE-001",
		Name:     "Test Zone",
		ZoneType: domain.ZoneTypeStorage,
	})
	if err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}
	if z.Code != "ZONE-001" {
		t.Errorf("code = %q, want %q", z.Code, "ZONE-001")
	}
	if z.WarehouseID != wh.ID {
		t.Errorf("warehouse_id = %q, want %q", z.WarehouseID, wh.ID)
	}
}

func TestWarehouseService_CreateZone_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewWarehouseService(newMockWarehouseRepo())

	badType := domain.ZoneType("unknown")
	_, err := svc.CreateZone(ctx, uuid.New(), CreateZoneInput{
		Code:     "Z-001",
		Name:     "Test",
		ZoneType: badType,
	})
	if err == nil {
		t.Fatal("expected error for invalid zone type")
	}
}

func TestWarehouseService_CreateZone_WarehouseNotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewWarehouseService(newMockWarehouseRepo())

	_, err := svc.CreateZone(ctx, uuid.New(), CreateZoneInput{
		Code:     "Z-001",
		Name:     "Test",
		ZoneType: domain.ZoneTypeStorage,
	})
	if err == nil {
		t.Fatal("expected error for non-existent warehouse")
	}
}

func TestWarehouseService_ListZones(t *testing.T) {
	ctx := context.Background()
	repo := newMockWarehouseRepo()
	svc := NewWarehouseService(repo)

	wh, _ := svc.CreateWarehouse(ctx, CreateWarehouseInput{Code: "WH-001", Name: "Test WH"})

	svc.CreateZone(ctx, wh.ID, CreateZoneInput{Code: "Z-01", Name: "Z1", ZoneType: domain.ZoneTypeStorage})
	svc.CreateZone(ctx, wh.ID, CreateZoneInput{Code: "Z-02", Name: "Z2", ZoneType: domain.ZoneTypePicking})

	zones, total, err := svc.ListZones(ctx, wh.ID, 0, 0)
	if err != nil {
		t.Fatalf("ListZones failed: %v", err)
	}
	if len(zones) != 2 {
		t.Errorf("expected 2 zones, got %d", len(zones))
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
}

func TestWarehouseService_CreateLocation(t *testing.T) {
	ctx := context.Background()
	repo := newMockWarehouseRepo()
	svc := NewWarehouseService(repo)

	wh, _ := svc.CreateWarehouse(ctx, CreateWarehouseInput{Code: "WH-001", Name: "Test WH"})
	zone, _ := svc.CreateZone(ctx, wh.ID, CreateZoneInput{Code: "Z-01", Name: "Z1", ZoneType: domain.ZoneTypeStorage})

	loc, err := svc.CreateLocation(ctx, zone.ID, CreateLocationInput{
		Code:         "A-01-01-01",
		Barcode:      "LOC-BC-001",
		LocationType: domain.LocationTypeShelf,
		Capacity:     &domain.Capacity{MaxWeight: 100, MaxQty: 50},
	})
	if err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}
	if loc.Code != "A-01-01-01" {
		t.Errorf("code = %q, want %q", loc.Code, "A-01-01-01")
	}
	if loc.WarehouseID != wh.ID {
		t.Errorf("warehouse_id = %q, want %q", loc.WarehouseID, wh.ID)
	}
}

func TestWarehouseService_CreateLocation_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewWarehouseService(newMockWarehouseRepo())

	badType := domain.LocationType("unknown")
	_, err := svc.CreateLocation(ctx, uuid.New(), CreateLocationInput{
		Code:         "L-001",
		LocationType: badType,
	})
	if err == nil {
		t.Fatal("expected error for invalid location type")
	}

	_, err = svc.CreateLocation(ctx, uuid.New(), CreateLocationInput{
		Code:         "",
		LocationType: domain.LocationTypeShelf,
	})
	if err == nil {
		t.Fatal("expected error for empty code")
	}
}

func TestWarehouseService_UpdateLocationStatus(t *testing.T) {
	ctx := context.Background()
	repo := newMockWarehouseRepo()
	svc := NewWarehouseService(repo)

	wh, _ := svc.CreateWarehouse(ctx, CreateWarehouseInput{Code: "WH-001", Name: "Test WH"})
	zone, _ := svc.CreateZone(ctx, wh.ID, CreateZoneInput{Code: "Z-01", Name: "Z1", ZoneType: domain.ZoneTypeStorage})
	loc, _ := svc.CreateLocation(ctx, zone.ID, CreateLocationInput{Code: "L-001", LocationType: domain.LocationTypeShelf})

	if err := svc.UpdateLocationStatus(ctx, loc.ID, domain.LocationStatusOccupied); err != nil {
		t.Fatalf("UpdateLocationStatus failed: %v", err)
	}

	got, _ := svc.GetLocation(ctx, loc.ID)
	if got.Status != domain.LocationStatusOccupied {
		t.Errorf("status = %q, want %q", got.Status, domain.LocationStatusOccupied)
	}
}

func TestWarehouseService_UpdateLocationStatus_InvalidStatus(t *testing.T) {
	ctx := context.Background()
	repo := newMockWarehouseRepo()
	svc := NewWarehouseService(repo)

	wh, _ := svc.CreateWarehouse(ctx, CreateWarehouseInput{Code: "WH-001", Name: "Test WH"})
	zone, _ := svc.CreateZone(ctx, wh.ID, CreateZoneInput{Code: "Z-01", Name: "Z1", ZoneType: domain.ZoneTypeStorage})
	loc, _ := svc.CreateLocation(ctx, zone.ID, CreateLocationInput{Code: "L-001", LocationType: domain.LocationTypeShelf})

	err := svc.UpdateLocationStatus(ctx, loc.ID, domain.LocationStatus("invalid"))
	if err == nil {
		t.Fatal("expected error for invalid location status")
	}
}
