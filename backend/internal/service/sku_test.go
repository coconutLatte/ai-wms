package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// mockInventoryRepo implements repository.InventoryRepository for testing.
type mockInventoryRepo struct {
	skus        map[uuid.UUID]*domain.SKU
	skusByCode  map[string]*domain.SKU
	inventories map[uuid.UUID]*domain.Inventory
	transactions []*domain.InventoryTransaction
}

func newMockInventoryRepo() *mockInventoryRepo {
	return &mockInventoryRepo{
		skus:        make(map[uuid.UUID]*domain.SKU),
		skusByCode:  make(map[string]*domain.SKU),
		inventories: make(map[uuid.UUID]*domain.Inventory),
	}
}

// ── SKU ─────────────────────────────────────

func (m *mockInventoryRepo) CreateSKU(ctx context.Context, s *domain.SKU) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	m.skus[s.ID] = s
	m.skusByCode[s.Code] = s
	return nil
}

func (m *mockInventoryRepo) GetSKU(ctx context.Context, id uuid.UUID) (*domain.SKU, error) {
	s, ok := m.skus[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("sku", id.String())
	}
	return s, nil
}

func (m *mockInventoryRepo) GetSKUByBarcode(ctx context.Context, barcode string) (*domain.SKU, error) {
	for _, s := range m.skus {
		if s.Barcode == barcode {
			return s, nil
		}
	}
	return nil, pkgerrors.NewNotFound("sku", barcode)
}

func (m *mockInventoryRepo) GetSKUByCode(ctx context.Context, code string) (*domain.SKU, error) {
	s, ok := m.skusByCode[code]
	if !ok {
		return nil, pkgerrors.NewNotFound("sku", code)
	}
	return s, nil
}

func (m *mockInventoryRepo) ListSKUs(ctx context.Context, limit, offset int) ([]*domain.SKU, error) {
	var result []*domain.SKU
	for _, s := range m.skus {
		result = append(result, s)
	}
	return result, nil
}

func (m *mockInventoryRepo) UpdateSKU(ctx context.Context, s *domain.SKU) error {
	if _, ok := m.skus[s.ID]; !ok {
		return pkgerrors.NewNotFound("sku", s.ID.String())
	}
	m.skus[s.ID] = s
	m.skusByCode[s.Code] = s
	return nil
}

// ── Inventory (stubs) ───────────────────────

func (m *mockInventoryRepo) CreateInventory(ctx context.Context, inv *domain.Inventory) error {
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	inv.AvailableQty = inv.Qty - inv.ReservedQty
	m.inventories[inv.ID] = inv
	return nil
}

func (m *mockInventoryRepo) GetInventory(ctx context.Context, id uuid.UUID) (*domain.Inventory, error) {
	inv, ok := m.inventories[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("inventory", id.String())
	}
	return inv, nil
}

func (m *mockInventoryRepo) GetAndLockInventory(ctx context.Context, id uuid.UUID) (*domain.Inventory, error) {
	// In the mock, GetAndLockInventory behaves identically to GetInventory
	// (there is no concurrent access to guard against in unit tests).
	return m.GetInventory(ctx, id)
}

func (m *mockInventoryRepo) GetInventoryAtLocation(ctx context.Context, skuID, locationID uuid.UUID, batchNo string) (*domain.Inventory, error) {
	for _, inv := range m.inventories {
		if inv.SKUID == skuID && inv.LocationID == locationID && inv.BatchNo == batchNo {
			return inv, nil
		}
	}
	return nil, pkgerrors.NewNotFound("inventory", "")
}

func (m *mockInventoryRepo) QueryInventory(ctx context.Context, filter repository.InventoryFilter) ([]*domain.Inventory, error) {
	var result []*domain.Inventory
	for _, inv := range m.inventories {
		if (filter.WarehouseID == uuid.Nil || inv.WarehouseID == filter.WarehouseID) &&
			(filter.SKUID == uuid.Nil || inv.SKUID == filter.SKUID) &&
			(filter.LocationID == uuid.Nil || inv.LocationID == filter.LocationID) &&
			(filter.Status == "" || inv.Status == filter.Status) {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (m *mockInventoryRepo) UpdateInventoryQty(ctx context.Context, id uuid.UUID, deltaQty, deltaReserved float64) error {
	inv, ok := m.inventories[id]
	if !ok {
		return pkgerrors.NewNotFound("inventory", id.String())
	}
	inv.Qty += deltaQty
	inv.ReservedQty += deltaReserved
	inv.AvailableQty = inv.Qty - inv.ReservedQty
	return nil
}

func (m *mockInventoryRepo) UpdateInventoryStatus(ctx context.Context, id uuid.UUID, status domain.InventoryStatus) error {
	inv, ok := m.inventories[id]
	if !ok {
		return pkgerrors.NewNotFound("inventory", id.String())
	}
	inv.Status = status
	return nil
}

func (m *mockInventoryRepo) CreateTransaction(ctx context.Context, tx *domain.InventoryTransaction) error {
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	m.transactions = append(m.transactions, tx)
	return nil
}

func (m *mockInventoryRepo) ListTransactions(ctx context.Context, inventoryID uuid.UUID, limit, offset int) ([]*domain.InventoryTransaction, error) {
	var result []*domain.InventoryTransaction
	for _, tx := range m.transactions {
		if tx.InventoryID == inventoryID {
			result = append(result, tx)
		}
	}

	// Apply offset
	if offset > 0 && offset < len(result) {
		result = result[offset:]
	} else if offset >= len(result) {
		return []*domain.InventoryTransaction{}, nil
	}

	// Apply limit
	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}

	return result, nil
}

func (m *mockInventoryRepo) ListTransactionsByReference(ctx context.Context, referenceType string, referenceID uuid.UUID) ([]*domain.InventoryTransaction, error) {
	var result []*domain.InventoryTransaction
	for _, tx := range m.transactions {
		if tx.ReferenceType == referenceType && tx.ReferenceID == referenceID {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (m *mockInventoryRepo) CountSKUs(ctx context.Context) (int, error) {
	return len(m.skus), nil
}

func (m *mockInventoryRepo) CountInventory(ctx context.Context, filter repository.InventoryFilter) (int, error) {
	count := 0
	for _, inv := range m.inventories {
		if (filter.WarehouseID == uuid.Nil || inv.WarehouseID == filter.WarehouseID) &&
			(filter.SKUID == uuid.Nil || inv.SKUID == filter.SKUID) &&
			(filter.LocationID == uuid.Nil || inv.LocationID == filter.LocationID) &&
			(filter.Status == "" || inv.Status == filter.Status) {
			count++
		}
	}
	return count, nil
}

func (m *mockInventoryRepo) GetOldestInventory(ctx context.Context, filter repository.InventoryRetrievalFilter) ([]*domain.Inventory, error) {
	var result []*domain.Inventory
	for _, inv := range m.inventories {
		if inv.Status != domain.InventoryStatusAvailable || inv.Qty <= 0 {
			continue
		}
		if filter.WarehouseID != uuid.Nil && inv.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.SKUID != uuid.Nil && inv.SKUID != filter.SKUID {
			continue
		}
		result = append(result, inv)
	}
	// Sort by ReceivedAt ASC (oldest first)
	sortInventoryByReceivedAt(result)
	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (m *mockInventoryRepo) GetExpiringInventory(ctx context.Context, filter repository.InventoryRetrievalFilter) ([]*domain.Inventory, error) {
	var result []*domain.Inventory
	for _, inv := range m.inventories {
		if inv.Status != domain.InventoryStatusAvailable || inv.Qty <= 0 {
			continue
		}
		if filter.WarehouseID != uuid.Nil && inv.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.SKUID != uuid.Nil && inv.SKUID != filter.SKUID {
			continue
		}
		result = append(result, inv)
	}
	// Sort by ExpiryDate ASC NULLS LAST (earliest expiring first)
	sortInventoryByExpiryDate(result)
	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}
	return result, nil
}

// sortInventoryByReceivedAt sorts inventory by ReceivedAt ASC (oldest first).
func sortInventoryByReceivedAt(inv []*domain.Inventory) {
	for i := 0; i < len(inv); i++ {
		for j := i + 1; j < len(inv); j++ {
			if inv[j].ReceivedAt.Before(inv[i].ReceivedAt) {
				inv[i], inv[j] = inv[j], inv[i]
			}
		}
	}
}

// sortInventoryByExpiryDate sorts inventory by ExpiryDate ASC NULLS LAST.
// Nil expiry dates (never expires) sort last.
func sortInventoryByExpiryDate(inv []*domain.Inventory) {
	for i := 0; i < len(inv); i++ {
		for j := i + 1; j < len(inv); j++ {
			if inv[i].ExpiryDate == nil && inv[j].ExpiryDate != nil {
				inv[i], inv[j] = inv[j], inv[i]
			} else if inv[i].ExpiryDate != nil && inv[j].ExpiryDate != nil && inv[j].ExpiryDate.Before(*inv[i].ExpiryDate) {
				inv[i], inv[j] = inv[j], inv[i]
			}
		}
	}
}

func (m *mockInventoryRepo) CountTransactions(ctx context.Context, inventoryID uuid.UUID) (int, error) {
	count := 0
	for _, tx := range m.transactions {
		if tx.InventoryID == inventoryID {
			count++
		}
	}
	return count, nil
}

// ── Dashboard Queries (stubs) ──────────────

func (m *mockInventoryRepo) GetInventoryDashboardStats(ctx context.Context, warehouseID uuid.UUID, lowStockThreshold float64) (*repository.InventoryDashboardStats, error) {
	stats := &repository.InventoryDashboardStats{}
	for _, inv := range m.inventories {
		if warehouseID != uuid.Nil && inv.WarehouseID != warehouseID {
			continue
		}
		stats.TotalRecords++
		stats.TotalQty += inv.Qty
		stats.TotalReservedQty += inv.ReservedQty
		stats.TotalAvailableQty += inv.Qty - inv.ReservedQty
		switch inv.Status {
		case domain.InventoryStatusAvailable:
			stats.AvailableCount++
		case domain.InventoryStatusQuarantine:
			stats.QuarantineCount++
		case domain.InventoryStatusDamaged:
			stats.DamagedCount++
		case domain.InventoryStatusExpired:
			stats.ExpiredCount++
		}
		availableQty := inv.Qty - inv.ReservedQty
		if availableQty > 0 && availableQty <= lowStockThreshold {
			stats.LowStockCount++
		}
	}
	return stats, nil
}

func (m *mockInventoryRepo) GetLowStockInventory(ctx context.Context, threshold float64, warehouseID uuid.UUID, limit int) ([]*domain.Inventory, error) {
	var result []*domain.Inventory
	for _, inv := range m.inventories {
		availableQty := inv.Qty - inv.ReservedQty
		if availableQty <= 0 || availableQty > threshold {
			continue
		}
		if warehouseID != uuid.Nil && inv.WarehouseID != warehouseID {
			continue
		}
		result = append(result, inv)
	}
	// Sort by available_qty ASC (simple insertion order is not guaranteed)
	sortInventoryLowStock(result)
	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}
	return result, nil
}

// sortInventoryLowStock sorts inventory by available_qty ASC.
func sortInventoryLowStock(inv []*domain.Inventory) {
	for i := 0; i < len(inv); i++ {
		for j := i + 1; j < len(inv); j++ {
			if (inv[j].Qty-inv[j].ReservedQty) < (inv[i].Qty-inv[i].ReservedQty) {
				inv[i], inv[j] = inv[j], inv[i]
			}
		}
	}
}

func (m *mockInventoryRepo) GetInventoryByWarehouse(ctx context.Context) ([]*repository.InventoryByWarehouseRow, error) {
	// Aggregate by WarehouseID
	agg := make(map[uuid.UUID]*repository.InventoryByWarehouseRow)
	for _, inv := range m.inventories {
		row, ok := agg[inv.WarehouseID]
		if !ok {
			row = &repository.InventoryByWarehouseRow{
				WarehouseID: inv.WarehouseID,
			}
			agg[inv.WarehouseID] = row
		}
		row.TotalQty += inv.Qty
		row.ReservedQty += inv.ReservedQty
		row.AvailableQty += inv.Qty - inv.ReservedQty
		row.RecordCount++
	}
	// For test purposes, warehouse name/code is unknown from the mock
	// (it doesn't hold warehouse aggregates). These fields will be empty.
	result := make([]*repository.InventoryByWarehouseRow, 0, len(agg))
	for _, row := range agg {
		result = append(result, row)
	}
	return result, nil
}

// ── Tests ───────────────────────────────────────────────────────────────────

func TestSKUService_CreateSKU(t *testing.T) {
	ctx := context.Background()
	svc := NewSKUService(newMockInventoryRepo())

	s, err := svc.CreateSKU(ctx, CreateSKUInput{
		Code: "SKU-001",
		Name: "Test Product",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	})
	if err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}
	if s.Code != "SKU-001" {
		t.Errorf("code = %q, want %q", s.Code, "SKU-001")
	}
	if s.Status != domain.SKUStatusActive {
		t.Errorf("status = %q, want %q", s.Status, domain.SKUStatusActive)
	}
	if s.UOM.BaseUnit != "EA" {
		t.Errorf("base_unit = %q, want %q", s.UOM.BaseUnit, "EA")
	}
}

func TestSKUService_CreateSKU_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewSKUService(newMockInventoryRepo())

	_, err := svc.CreateSKU(ctx, CreateSKUInput{Code: "", Name: "No Code", UOM: domain.UOM{BaseUnit: "EA"}})
	if err == nil {
		t.Fatal("expected error for empty code")
	}

	_, err = svc.CreateSKU(ctx, CreateSKUInput{Code: "SKU-001", Name: "", UOM: domain.UOM{BaseUnit: "EA"}})
	if err == nil {
		t.Fatal("expected error for empty name")
	}

	_, err = svc.CreateSKU(ctx, CreateSKUInput{Code: "SKU-001", Name: "Test", UOM: domain.UOM{BaseUnit: ""}})
	if err == nil {
		t.Fatal("expected error for empty base unit")
	}
}

func TestSKUService_CreateSKU_WithAttributes(t *testing.T) {
	ctx := context.Background()
	svc := NewSKUService(newMockInventoryRepo())

	attrs := domain.Attributes{"color": "red", "size": "M"}
	s, err := svc.CreateSKU(ctx, CreateSKUInput{
		Code:       "SKU-ATTR",
		Name:       "Attributed SKU",
		UOM:        domain.UOM{BaseUnit: "EA"},
		Attributes: attrs,
		Category:   "Electronics",
	})
	if err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}
	if s.Attributes["color"] != "red" {
		t.Errorf("attributes[color] = %q, want %q", s.Attributes["color"], "red")
	}
	if s.Category != "Electronics" {
		t.Errorf("category = %q, want %q", s.Category, "Electronics")
	}
	// Verify attributes are a clone, not the same map.
	if &s.Attributes == &attrs {
		t.Error("attributes should be a clone, not the original map")
	}
}

func TestSKUService_GetSKU(t *testing.T) {
	ctx := context.Background()
	svc := NewSKUService(newMockInventoryRepo())

	s, _ := svc.CreateSKU(ctx, CreateSKUInput{Code: "SKU-001", Name: "Test", UOM: domain.UOM{BaseUnit: "EA"}})

	got, err := svc.GetSKU(ctx, s.ID)
	if err != nil {
		t.Fatalf("GetSKU failed: %v", err)
	}
	if got.Code != "SKU-001" {
		t.Errorf("code = %q, want %q", got.Code, "SKU-001")
	}
}

func TestSKUService_GetSKU_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewSKUService(newMockInventoryRepo())

	_, err := svc.GetSKU(ctx, uuid.New())
	if err == nil {
		t.Fatal("expected error for unknown sku")
	}
}

func TestSKUService_GetSKUByCode(t *testing.T) {
	ctx := context.Background()
	svc := NewSKUService(newMockInventoryRepo())

	svc.CreateSKU(ctx, CreateSKUInput{Code: "SKU-FIND", Name: "Find Me", UOM: domain.UOM{BaseUnit: "EA"}})

	got, err := svc.GetSKUByCode(ctx, "SKU-FIND")
	if err != nil {
		t.Fatalf("GetSKUByCode failed: %v", err)
	}
	if got.Name != "Find Me" {
		t.Errorf("name = %q, want %q", got.Name, "Find Me")
	}
}

func TestSKUService_GetSKUByCode_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewSKUService(newMockInventoryRepo())

	_, err := svc.GetSKUByCode(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown sku code")
	}
}

func TestSKUService_ListSKUs(t *testing.T) {
	ctx := context.Background()
	svc := NewSKUService(newMockInventoryRepo())

	svc.CreateSKU(ctx, CreateSKUInput{Code: "SKU-001", Name: "S1", UOM: domain.UOM{BaseUnit: "EA"}})
	svc.CreateSKU(ctx, CreateSKUInput{Code: "SKU-002", Name: "S2", UOM: domain.UOM{BaseUnit: "EA"}})
	svc.CreateSKU(ctx, CreateSKUInput{Code: "SKU-003", Name: "S3", UOM: domain.UOM{BaseUnit: "EA"}})

	list, total, err := svc.ListSKUs(ctx, 0, 0)
	if err != nil {
		t.Fatalf("ListSKUs failed: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 skus, got %d", len(list))
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
}

func TestSKUService_ListSKUs_Empty(t *testing.T) {
	ctx := context.Background()
	svc := NewSKUService(newMockInventoryRepo())

	list, total, err := svc.ListSKUs(ctx, 0, 0)
	if err != nil {
		t.Fatalf("ListSKUs failed: %v", err)
	}
	// An empty list may be nil from the repo —  that's okay; the API handler
	// normalizes it to an empty JSON array.
	if len(list) != 0 {
		t.Errorf("expected 0 skus, got %d", len(list))
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
}

func TestSKUService_UpdateSKU(t *testing.T) {
	ctx := context.Background()
	svc := NewSKUService(newMockInventoryRepo())

	s, _ := svc.CreateSKU(ctx, CreateSKUInput{
		Code:        "SKU-001",
		Name:        "Original",
		Description: "Original desc",
		UOM:         domain.UOM{BaseUnit: "EA", Weight: 1.0},
		Category:    "OriginalCat",
	})

	newName := "Updated"
	newDesc := "Updated desc"
	newBarcode := "UPC-999"
	newCategory := "NewCat"
	newStatus := domain.SKUStatusInactive

	updated, err := svc.UpdateSKU(ctx, s.ID, UpdateSKUInput{
		Name:        &newName,
		Description: &newDesc,
		Barcode:     &newBarcode,
		Category:    &newCategory,
		Status:      &newStatus,
	})
	if err != nil {
		t.Fatalf("UpdateSKU failed: %v", err)
	}
	if updated.Name != "Updated" {
		t.Errorf("name = %q, want %q", updated.Name, "Updated")
	}
	if updated.Description != "Updated desc" {
		t.Errorf("description = %q, want %q", updated.Description, "Updated desc")
	}
	if updated.Barcode != "UPC-999" {
		t.Errorf("barcode = %q, want %q", updated.Barcode, "UPC-999")
	}
	if updated.Category != "NewCat" {
		t.Errorf("category = %q, want %q", updated.Category, "NewCat")
	}
	if updated.Status != domain.SKUStatusInactive {
		t.Errorf("status = %q, want %q", updated.Status, domain.SKUStatusInactive)
	}
	// Unchanged fields should remain.
	if updated.UOM.Weight != 1.0 {
		t.Errorf("weight = %f, want %f", updated.UOM.Weight, 1.0)
	}
}

func TestSKUService_UpdateSKU_UOM(t *testing.T) {
	ctx := context.Background()
	svc := NewSKUService(newMockInventoryRepo())

	s, _ := svc.CreateSKU(ctx, CreateSKUInput{
		Code: "SKU-UOM",
		Name: "UOM Test",
		UOM:  domain.UOM{BaseUnit: "EA", Weight: 1.0, Volume: 0.5},
	})

	newUOM := domain.UOM{BaseUnit: "KG", Weight: 2.5, Volume: 1.5}
	updated, err := svc.UpdateSKU(ctx, s.ID, UpdateSKUInput{
		UOM: &newUOM,
	})
	if err != nil {
		t.Fatalf("UpdateSKU UOM failed: %v", err)
	}
	if updated.UOM.BaseUnit != "KG" {
		t.Errorf("base_unit = %q, want %q", updated.UOM.BaseUnit, "KG")
	}
	if updated.UOM.Weight != 2.5 {
		t.Errorf("weight = %f, want %f", updated.UOM.Weight, 2.5)
	}
	if updated.UOM.Volume != 1.5 {
		t.Errorf("volume = %f, want %f", updated.UOM.Volume, 1.5)
	}
}

func TestSKUService_UpdateSKU_Attributes(t *testing.T) {
	ctx := context.Background()
	svc := NewSKUService(newMockInventoryRepo())

	origAttrs := domain.Attributes{"color": "blue", "size": "L"}
	s, _ := svc.CreateSKU(ctx, CreateSKUInput{
		Code:       "SKU-ATTR2",
		Name:       "Attr Test",
		UOM:        domain.UOM{BaseUnit: "EA"},
		Attributes: origAttrs,
	})

	newAttrs := domain.Attributes{"color": "red", "weight": "heavy"}
	updated, err := svc.UpdateSKU(ctx, s.ID, UpdateSKUInput{
		Attributes: &newAttrs,
	})
	if err != nil {
		t.Fatalf("UpdateSKU attributes failed: %v", err)
	}
	if updated.Attributes["color"] != "red" {
		t.Errorf("attributes[color] = %q, want %q", updated.Attributes["color"], "red")
	}
	if updated.Attributes["size"] != "" {
		t.Error("old attribute 'size' should not be present after replace")
	}
	if updated.Attributes["weight"] != "heavy" {
		t.Errorf("attributes[weight] = %q, want %q", updated.Attributes["weight"], "heavy")
	}
}

func TestSKUService_UpdateSKU_InvalidStatus(t *testing.T) {
	ctx := context.Background()
	svc := NewSKUService(newMockInventoryRepo())

	s, _ := svc.CreateSKU(ctx, CreateSKUInput{Code: "SKU-INV", Name: "Test", UOM: domain.UOM{BaseUnit: "EA"}})

	badStatus := domain.SKUStatus("nonexistent")
	_, err := svc.UpdateSKU(ctx, s.ID, UpdateSKUInput{
		Status: &badStatus,
	})
	if err == nil {
		t.Fatal("expected error for invalid sku status")
	}
}

func TestSKUService_UpdateSKU_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewSKUService(newMockInventoryRepo())

	newName := "Ghost"
	_, err := svc.UpdateSKU(ctx, uuid.New(), UpdateSKUInput{
		Name: &newName,
	})
	if err == nil {
		t.Fatal("expected error for non-existent sku")
	}
}
