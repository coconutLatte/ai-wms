package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	"github.com/ai-wms/ai-wms/backend/internal/service"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// ── Mock Inventory Repo (for InventoryHandler + SKUHandler tests) ──────────────

type mockInvRepo struct {
	skuByCodeFn    func(ctx context.Context, code string) (*domain.SKU, error)
	getInvFn       func(ctx context.Context, id uuid.UUID) (*domain.Inventory, error)
	updateStatusFn func(ctx context.Context, id uuid.UUID, status domain.InventoryStatus) error
}

func (m *mockInvRepo) CreateSKU(ctx context.Context, s *domain.SKU) error          { return nil }
func (m *mockInvRepo) GetSKU(ctx context.Context, id uuid.UUID) (*domain.SKU, error) {
	return nil, pkgerrors.NewNotFound("sku", id.String())
}
func (m *mockInvRepo) GetSKUByBarcode(ctx context.Context, barcode string) (*domain.SKU, error) {
	return nil, pkgerrors.NewNotFound("sku", barcode)
}
func (m *mockInvRepo) GetSKUByCode(ctx context.Context, code string) (*domain.SKU, error) {
	if m.skuByCodeFn != nil {
		return m.skuByCodeFn(ctx, code)
	}
	return nil, pkgerrors.NewNotFound("sku", code)
}
func (m *mockInvRepo) ListSKUs(ctx context.Context, limit, offset int) ([]*domain.SKU, error) {
	return nil, nil
}
func (m *mockInvRepo) UpdateSKU(ctx context.Context, s *domain.SKU) error { return nil }
func (m *mockInvRepo) CountSKUs(ctx context.Context) (int, error)         { return 0, nil }

func (m *mockInvRepo) CreateInventory(ctx context.Context, inv *domain.Inventory) error { return nil }
func (m *mockInvRepo) GetInventory(ctx context.Context, id uuid.UUID) (*domain.Inventory, error) {
	if m.getInvFn != nil {
		return m.getInvFn(ctx, id)
	}
	return nil, pkgerrors.NewNotFound("inventory", id.String())
}
func (m *mockInvRepo) GetAndLockInventory(ctx context.Context, id uuid.UUID) (*domain.Inventory, error) {
	return m.GetInventory(ctx, id)
}
func (m *mockInvRepo) GetInventoryAtLocation(ctx context.Context, skuID, locationID uuid.UUID, batchNo string) (*domain.Inventory, error) {
	return nil, pkgerrors.NewNotFound("inventory", "")
}
func (m *mockInvRepo) QueryInventory(ctx context.Context, filter repository.InventoryFilter) ([]*domain.Inventory, error) {
	return nil, nil
}
func (m *mockInvRepo) UpdateInventoryQty(ctx context.Context, id uuid.UUID, deltaQty, deltaReserved float64) error {
	return nil
}
func (m *mockInvRepo) UpdateInventoryStatus(ctx context.Context, id uuid.UUID, status domain.InventoryStatus) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}
	return nil
}
func (m *mockInvRepo) CountInventory(ctx context.Context, filter repository.InventoryFilter) (int, error) {
	return 0, nil
}
func (m *mockInvRepo) GetOldestInventory(ctx context.Context, filter repository.InventoryRetrievalFilter) ([]*domain.Inventory, error) {
	return nil, nil
}
func (m *mockInvRepo) GetExpiringInventory(ctx context.Context, filter repository.InventoryRetrievalFilter) ([]*domain.Inventory, error) {
	return nil, nil
}
func (m *mockInvRepo) CreateTransaction(ctx context.Context, tx *domain.InventoryTransaction) error { return nil }
func (m *mockInvRepo) ListTransactions(ctx context.Context, inventoryID uuid.UUID, limit, offset int) ([]*domain.InventoryTransaction, error) {
	return nil, nil
}
func (m *mockInvRepo) CountTransactions(ctx context.Context, inventoryID uuid.UUID) (int, error) { return 0, nil }
func (m *mockInvRepo) GetInventoryDashboardStats(ctx context.Context, warehouseID uuid.UUID, lowStockThreshold float64) (*repository.InventoryDashboardStats, error) {
	return nil, nil
}
func (m *mockInvRepo) GetLowStockInventory(ctx context.Context, threshold float64, warehouseID uuid.UUID, limit int) ([]*domain.Inventory, error) {
	return nil, nil
}
func (m *mockInvRepo) GetInventoryByWarehouse(ctx context.Context) ([]*repository.InventoryByWarehouseRow, error) {
	return nil, nil
}

// ── Mock Order Repo (for OrderHandler tests) ───────────────────────────────────

type mockOrderRepo struct {
	getLineFn       func(ctx context.Context, id uuid.UUID) (*domain.OrderLine, error)
	updateLineFn    func(ctx context.Context, id uuid.UUID, status domain.OrderLineStatus) error
}

func (m *mockOrderRepo) CreateOrder(ctx context.Context, o *domain.Order) error { return nil }
func (m *mockOrderRepo) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return nil, pkgerrors.NewNotFound("order", id.String())
}
func (m *mockOrderRepo) GetOrderByNo(ctx context.Context, orderNo string) (*domain.Order, error) {
	return nil, pkgerrors.NewNotFound("order", orderNo)
}
func (m *mockOrderRepo) ListOrders(ctx context.Context, filter repository.OrderFilter) ([]*domain.Order, error) {
	return nil, nil
}
func (m *mockOrderRepo) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	return nil
}
func (m *mockOrderRepo) CountOrders(ctx context.Context, filter repository.OrderFilter) (int, error) { return 0, nil }

func (m *mockOrderRepo) CreateOrderLine(ctx context.Context, line *domain.OrderLine) error { return nil }
func (m *mockOrderRepo) GetOrderLine(ctx context.Context, id uuid.UUID) (*domain.OrderLine, error) {
	if m.getLineFn != nil {
		return m.getLineFn(ctx, id)
	}
	return nil, pkgerrors.NewNotFound("order_line", id.String())
}
func (m *mockOrderRepo) GetOrderLines(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderLine, error) {
	return nil, nil
}
func (m *mockOrderRepo) UpdateOrderLineStatus(ctx context.Context, id uuid.UUID, status domain.OrderLineStatus) error {
	if m.updateLineFn != nil {
		return m.updateLineFn(ctx, id, status)
	}
	return nil
}
func (m *mockOrderRepo) UpdateOrderLineFulfilledQty(ctx context.Context, id uuid.UUID, qty float64) error {
	return nil
}

func (m *mockOrderRepo) CreateASN(ctx context.Context, asn *domain.ASN) error { return nil }
func (m *mockOrderRepo) GetASN(ctx context.Context, id uuid.UUID) (*domain.ASN, error) {
	return nil, pkgerrors.NewNotFound("asn", id.String())
}
func (m *mockOrderRepo) GetASNByNo(ctx context.Context, asnNo string) (*domain.ASN, error) {
	return nil, pkgerrors.NewNotFound("asn", asnNo)
}
func (m *mockOrderRepo) ListASNs(ctx context.Context, filter repository.ASNFilter) ([]*domain.ASN, error) { return nil, nil }
func (m *mockOrderRepo) UpdateASNStatus(ctx context.Context, id uuid.UUID, status domain.ASNStatus) error { return nil }
func (m *mockOrderRepo) CountASNs(ctx context.Context, filter repository.ASNFilter) (int, error)          { return 0, nil }
func (m *mockOrderRepo) CreateASNLine(ctx context.Context, line *domain.ASNLine) error                    { return nil }
func (m *mockOrderRepo) GetASNLines(ctx context.Context, asnID uuid.UUID) ([]*domain.ASNLine, error)      { return nil, nil }
func (m *mockOrderRepo) UpdateASNLineStatus(ctx context.Context, id uuid.UUID, status domain.ASNLineStatus) error {
	return nil
}
func (m *mockOrderRepo) UpdateASNLineReceivedQty(ctx context.Context, id uuid.UUID, qty float64) error { return nil }

// ── Mock Warehouse Repo (for WarehouseHandler tests) ───────────────────────────

type mockWhRepo struct {
	getLocByBarcodeFn func(ctx context.Context, barcode string) (*domain.Location, error)
}

func (m *mockWhRepo) CreateWarehouse(ctx context.Context, w *domain.Warehouse) error  { return nil }
func (m *mockWhRepo) GetWarehouse(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	return nil, pkgerrors.NewNotFound("warehouse", id.String())
}
func (m *mockWhRepo) ListWarehouses(ctx context.Context, limit, offset int) ([]*domain.Warehouse, error) {
	return nil, nil
}
func (m *mockWhRepo) UpdateWarehouse(ctx context.Context, w *domain.Warehouse) error { return nil }
func (m *mockWhRepo) CountWarehouses(ctx context.Context) (int, error)               { return 0, nil }

func (m *mockWhRepo) CreateZone(ctx context.Context, z *domain.Zone) error { return nil }
func (m *mockWhRepo) GetZone(ctx context.Context, id uuid.UUID) (*domain.Zone, error) {
	return nil, pkgerrors.NewNotFound("zone", id.String())
}
func (m *mockWhRepo) ListZonesByWarehouse(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]*domain.Zone, error) {
	return nil, nil
}
func (m *mockWhRepo) CountZonesByWarehouse(ctx context.Context, warehouseID uuid.UUID) (int, error) { return 0, nil }

func (m *mockWhRepo) CreateLocation(ctx context.Context, l *domain.Location) error  { return nil }
func (m *mockWhRepo) GetLocation(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	return nil, pkgerrors.NewNotFound("location", id.String())
}
func (m *mockWhRepo) GetLocationByBarcode(ctx context.Context, barcode string) (*domain.Location, error) {
	if m.getLocByBarcodeFn != nil {
		return m.getLocByBarcodeFn(ctx, barcode)
	}
	return nil, pkgerrors.NewNotFound("location", barcode)
}
func (m *mockWhRepo) ListLocationsByZone(ctx context.Context, zoneID uuid.UUID, limit, offset int) ([]*domain.Location, error) {
	return nil, nil
}
func (m *mockWhRepo) UpdateLocationStatus(ctx context.Context, id uuid.UUID, status domain.LocationStatus) error { return nil }
func (m *mockWhRepo) CountLocationsByZone(ctx context.Context, zoneID uuid.UUID) (int, error)                    { return 0, nil }

// ── Test Helpers ───────────────────────────────────────────────────────────────

func testTime() time.Time {
	t, _ := time.Parse(time.RFC3339, "2026-07-22T10:00:00Z")
	return t
}

var (
	invID  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	skuID  = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	locID  = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	whID   = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	ordID  = uuid.MustParse("bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	lineID = uuid.MustParse("aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
)

// ── UpdateInventoryStatus Handler Tests ────────────────────────────────────────

func TestUpdateInventoryStatus_Success(t *testing.T) {
	inv := &domain.Inventory{
		ID:          invID,
		SKUID:       skuID,
		LocationID:  locID,
		WarehouseID: whID,
		BatchNo:     "BATCH-001",
		Qty:         100,
		ReservedQty: 20,
		Status:      domain.InventoryStatusAvailable,
		ReceivedAt:  testTime(),
		UpdatedAt:   testTime(),
	}

	var statusUpdated domain.InventoryStatus
	repo := &mockInvRepo{
		getInvFn: func(ctx context.Context, id uuid.UUID) (*domain.Inventory, error) {
			// After the status update, return the updated status.
			if statusUpdated != "" {
				inv.Status = statusUpdated
			}
			return inv, nil
		},
		updateStatusFn: func(ctx context.Context, id uuid.UUID, status domain.InventoryStatus) error {
			statusUpdated = status
			return nil
		},
	}
	handler := &InventoryHandler{svc: service.NewInventoryService(repo)}

	body := `{"status": "quarantine"}`
	r := httptest.NewRequest("PATCH", "/api/v1/inventory/"+invID.String()+"/status", strings.NewReader(body))
	r.SetPathValue("id", invID.String())
	w := httptest.NewRecorder()

	handler.UpdateInventoryStatus(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp inventoryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "quarantine" {
		t.Errorf("status = %q, want quarantine", resp.Status)
	}
}

func TestUpdateInventoryStatus_InvalidUUID(t *testing.T) {
	handler := &InventoryHandler{svc: service.NewInventoryService(&mockInvRepo{})}

	r := httptest.NewRequest("PATCH", "/api/v1/inventory/bad-uuid/status", nil)
	r.SetPathValue("id", "bad-uuid")
	w := httptest.NewRecorder()

	handler.UpdateInventoryStatus(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateInventoryStatus_NotFound(t *testing.T) {
	handler := &InventoryHandler{svc: service.NewInventoryService(&mockInvRepo{})}

	body := `{"status": "quarantine"}`
	r := httptest.NewRequest("PATCH", "/api/v1/inventory/"+invID.String()+"/status", strings.NewReader(body))
	r.SetPathValue("id", invID.String())
	w := httptest.NewRecorder()

	handler.UpdateInventoryStatus(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// ── UpdateOrderLineStatus Handler Tests ────────────────────────────────────────

func TestUpdateOrderLineStatus_Success(t *testing.T) {
	line := &domain.OrderLine{
		ID:      lineID,
		OrderID: ordID,
		SKUID:   skuID,
		LineNo:  1,
		Status:  domain.OrderLineStatusPending,
		UOM:     "EA",
	}

	var statusUpdated domain.OrderLineStatus
	repo := &mockOrderRepo{
		getLineFn: func(ctx context.Context, id uuid.UUID) (*domain.OrderLine, error) {
			if statusUpdated != "" {
				line.Status = statusUpdated
			}
			return line, nil
		},
		updateLineFn: func(ctx context.Context, id uuid.UUID, status domain.OrderLineStatus) error {
			statusUpdated = status
			return nil
		},
	}
	handler := &OrderHandler{svc: service.NewOrderService(repo)}

	body := `{"status": "allocated"}`
	r := httptest.NewRequest("PUT", "/api/v1/orders/"+ordID.String()+"/lines/"+lineID.String()+"/status", strings.NewReader(body))
	r.SetPathValue("id", ordID.String())
	r.SetPathValue("lineId", lineID.String())
	w := httptest.NewRecorder()

	handler.UpdateOrderLineStatus(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp orderLineResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "allocated" {
		t.Errorf("status = %q, want allocated", resp.Status)
	}
}

func TestUpdateOrderLineStatus_InvalidUUID(t *testing.T) {
	handler := &OrderHandler{svc: service.NewOrderService(&mockOrderRepo{})}

	body := `{"status": "allocated"}`
	r := httptest.NewRequest("PUT", "/api/v1/orders/"+ordID.String()+"/lines/bad-uuid/status", strings.NewReader(body))
	r.SetPathValue("id", ordID.String())
	r.SetPathValue("lineId", "bad-uuid")
	w := httptest.NewRecorder()

	handler.UpdateOrderLineStatus(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUpdateOrderLineStatus_NotFound(t *testing.T) {
	handler := &OrderHandler{svc: service.NewOrderService(&mockOrderRepo{})}

	body := `{"status": "allocated"}`
	r := httptest.NewRequest("PUT", "/api/v1/orders/"+ordID.String()+"/lines/"+lineID.String()+"/status", strings.NewReader(body))
	r.SetPathValue("id", ordID.String())
	r.SetPathValue("lineId", lineID.String())
	w := httptest.NewRecorder()

	handler.UpdateOrderLineStatus(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// ── GetLocationByBarcode Handler Tests ─────────────────────────────────────────

func TestGetLocationByBarcode_Success(t *testing.T) {
	loc := &domain.Location{
		ID:           locID,
		ZoneID:       uuid.MustParse("eeeeeee1-eeee-eeee-eeee-eeeeeeeeeeee"),
		WarehouseID:  whID,
		Code:         "A-01-01-01",
		Barcode:      "LOC-BC-001",
		LocationType: domain.LocationTypeShelf,
		Status:       domain.LocationStatusEmpty,
		CreatedAt:    testTime(),
		UpdatedAt:    testTime(),
	}

	repo := &mockWhRepo{
		getLocByBarcodeFn: func(ctx context.Context, barcode string) (*domain.Location, error) {
			return loc, nil
		},
	}
	handler := &WarehouseHandler{svc: service.NewWarehouseService(repo)}

	r := httptest.NewRequest("GET", "/api/v1/locations?barcode=LOC-BC-001", nil)
	w := httptest.NewRecorder()

	handler.GetLocationByBarcode(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp locationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Barcode != "LOC-BC-001" {
		t.Errorf("barcode = %q, want LOC-BC-001", resp.Barcode)
	}
}

func TestGetLocationByBarcode_MissingBarcode(t *testing.T) {
	handler := &WarehouseHandler{svc: service.NewWarehouseService(&mockWhRepo{})}

	r := httptest.NewRequest("GET", "/api/v1/locations", nil)
	w := httptest.NewRecorder()

	handler.GetLocationByBarcode(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetLocationByBarcode_NotFound(t *testing.T) {
	handler := &WarehouseHandler{svc: service.NewWarehouseService(&mockWhRepo{})}

	r := httptest.NewRequest("GET", "/api/v1/locations?barcode=NONEXIST", nil)
	w := httptest.NewRecorder()

	handler.GetLocationByBarcode(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// ── GetSKUByCode Handler Tests ─────────────────────────────────────────────────

func TestGetSKUByCode_Success(t *testing.T) {
	sku := &domain.SKU{
		ID:        skuID,
		Code:      "SKU-12345",
		Name:      "Test SKU",
		Barcode:   "BC-12345",
		Status:    domain.SKUStatusActive,
		CreatedAt: testTime(),
		UpdatedAt: testTime(),
	}

	repo := &mockInvRepo{
		skuByCodeFn: func(ctx context.Context, code string) (*domain.SKU, error) {
			return sku, nil
		},
	}
	handler := &SKUHandler{svc: service.NewSKUService(repo)}

	r := httptest.NewRequest("GET", "/api/v1/skus?code=SKU-12345", nil)
	w := httptest.NewRecorder()

	handler.GetSKUByCode(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp skuResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != "SKU-12345" {
		t.Errorf("code = %q, want SKU-12345", resp.Code)
	}
}

func TestGetSKUByCode_MissingCode(t *testing.T) {
	handler := &SKUHandler{svc: service.NewSKUService(&mockInvRepo{})}

	r := httptest.NewRequest("GET", "/api/v1/skus?code=", nil)
	w := httptest.NewRecorder()

	handler.GetSKUByCode(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetSKUByCode_NotFound(t *testing.T) {
	handler := &SKUHandler{svc: service.NewSKUService(&mockInvRepo{})}

	r := httptest.NewRequest("GET", "/api/v1/skus?code=NONEXIST", nil)
	w := httptest.NewRecorder()

	handler.GetSKUByCode(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
