package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// setupOrderTestDB creates a test database and cleans up order-related test data.
func setupOrderTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	cfg := testConfig()

	ctx := context.Background()
	db, err := NewDB(ctx, cfg)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
		return nil, nil
	}

	// Clean up previous test data (order matters due to FK constraints)
	_, _ = db.Pool.Exec(ctx, "DELETE FROM order_lines WHERE order_id IN (SELECT id FROM orders WHERE order_no LIKE 'TEST-%')")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM orders WHERE order_no LIKE 'TEST-%'")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM asn_lines WHERE asn_id IN (SELECT id FROM asns WHERE asn_no LIKE 'TEST-%')")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM asns WHERE asn_no LIKE 'TEST-%'")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM inventory_transactions WHERE sku_id IN (SELECT id FROM skus WHERE code LIKE 'TEST-%')")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM inventory WHERE sku_id IN (SELECT id FROM skus WHERE code LIKE 'TEST-%')")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM skus WHERE code LIKE 'TEST-%'")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM locations WHERE code LIKE 'TEST-%'")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM zones WHERE code LIKE 'TEST-%'")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM warehouses WHERE code LIKE 'TEST-%'")

	cleanup := func() {
		_, _ = db.Pool.Exec(ctx, "DELETE FROM order_lines WHERE order_id IN (SELECT id FROM orders WHERE order_no LIKE 'TEST-%')")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM orders WHERE order_no LIKE 'TEST-%'")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM asn_lines WHERE asn_id IN (SELECT id FROM asns WHERE asn_no LIKE 'TEST-%')")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM asns WHERE asn_no LIKE 'TEST-%'")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM inventory_transactions WHERE sku_id IN (SELECT id FROM skus WHERE code LIKE 'TEST-%')")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM inventory WHERE sku_id IN (SELECT id FROM skus WHERE code LIKE 'TEST-%')")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM skus WHERE code LIKE 'TEST-%'")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM locations WHERE code LIKE 'TEST-%'")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM zones WHERE code LIKE 'TEST-%'")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM warehouses WHERE code LIKE 'TEST-%'")
		db.Close()
	}

	return db, cleanup
}

// createTestWarehouse creates a warehouse for order tests.
func createTestWarehouse(t *testing.T, ctx context.Context, repo *WarehouseRepo) *domain.Warehouse {
	t.Helper()

	wh := &domain.Warehouse{
		Code: "TEST-WH-ORD-" + uuid.New().String()[:8],
		Name: "Order Test Warehouse",
	}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}
	return wh
}

// createTestSKU creates a SKU for order line tests.
func createTestSKU(t *testing.T, ctx context.Context, repo *InventoryRepo) *domain.SKU {
	t.Helper()

	sku := &domain.SKU{
		Code: "TEST-SKU-ORD-" + uuid.New().String()[:8],
		Name: "Order Test SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := repo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}
	return sku
}

// ── Order Tests ────────────────────────────────────────────

func TestOrderRepo_CreateAndGetOrder(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)

	o := &domain.Order{
		OrderNo:      "TEST-ORD-001",
		OrderType:    domain.OrderTypeInbound,
		WarehouseID:  wh.ID,
		Priority:     domain.OrderPriorityHigh,
		ExternalRef:  "PO-2026-00123",
		ExternalType: "purchase_order",
		Notes:        "Urgent inbound order",
		CreatedBy:    "test-user",
	}

	err := orderRepo.CreateOrder(ctx, o)
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if o.ID == uuid.Nil {
		t.Error("expected order ID to be set")
	}
	if o.Status != domain.OrderStatusDraft {
		t.Errorf("status = %q, want draft", o.Status)
	}

	got, err := orderRepo.GetOrder(ctx, o.ID)
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}
	if got.OrderNo != o.OrderNo {
		t.Errorf("order_no = %q, want %q", got.OrderNo, o.OrderNo)
	}
	if got.OrderType != domain.OrderTypeInbound {
		t.Errorf("order_type = %q, want inbound", got.OrderType)
	}
	if got.Priority != domain.OrderPriorityHigh {
		t.Errorf("priority = %q, want high", got.Priority)
	}
	if got.ExternalRef != "PO-2026-00123" {
		t.Errorf("external_ref = %q, want PO-2026-00123", got.ExternalRef)
	}
	if got.ExternalType != "purchase_order" {
		t.Errorf("external_type = %q, want purchase_order", got.ExternalType)
	}
	if got.Notes != "Urgent inbound order" {
		t.Errorf("notes = %q, want 'Urgent inbound order'", got.Notes)
	}
	if got.CreatedBy != "test-user" {
		t.Errorf("created_by = %q, want test-user", got.CreatedBy)
	}
	if got.CompletedAt != nil {
		t.Error("expected completed_at to be nil for draft order")
	}
}

func TestOrderRepo_GetOrderByNo(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)

	o := &domain.Order{
		OrderNo:     "TEST-ORD-BYNO-001",
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateOrder(ctx, o); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	got, err := orderRepo.GetOrderByNo(ctx, "TEST-ORD-BYNO-001")
	if err != nil {
		t.Fatalf("GetOrderByNo failed: %v", err)
	}
	if got.ID != o.ID {
		t.Errorf("id = %s, want %s", got.ID, o.ID)
	}

	// Not found
	_, err = orderRepo.GetOrderByNo(ctx, "NONEXISTENT-ORDER")
	if err == nil {
		t.Error("expected error for nonexistent order number")
	}
}

func TestOrderRepo_GetOrder_NotFound(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	orderRepo := NewOrderRepo(db)

	_, err := orderRepo.GetOrder(ctx, uuid.New())
	if err == nil {
		t.Error("expected error for nonexistent order")
	}
}

func TestOrderRepo_ListOrders(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)

	// Create orders with different types
	types := []domain.OrderType{domain.OrderTypeInbound, domain.OrderTypeOutbound, domain.OrderTypeTransfer}
	for i, ot := range types {
		o := &domain.Order{
			OrderNo:     "TEST-ORD-LIST-00" + string(rune('1'+i)),
			OrderType:   ot,
			WarehouseID: wh.ID,
			Status:      domain.OrderStatusDraft,
		}
		if err := orderRepo.CreateOrder(ctx, o); err != nil {
			t.Fatalf("CreateOrder [%d] failed: %v", i, err)
		}
	}

	// List all
	orders, err := orderRepo.ListOrders(ctx, repository.OrderFilter{})
	if err != nil {
		t.Fatalf("ListOrders failed: %v", err)
	}
	if len(orders) < 3 {
		t.Errorf("expected at least 3 orders, got %d", len(orders))
	}

	// Filter by warehouse
	orders, err = orderRepo.ListOrders(ctx, repository.OrderFilter{
		WarehouseID: wh.ID,
	})
	if err != nil {
		t.Fatalf("ListOrders by warehouse failed: %v", err)
	}
	if len(orders) != 3 {
		t.Errorf("expected 3 orders for warehouse, got %d", len(orders))
	}

	// Filter by type
	orders, err = orderRepo.ListOrders(ctx, repository.OrderFilter{
		OrderType: domain.OrderTypeInbound,
	})
	if err != nil {
		t.Fatalf("ListOrders by type failed: %v", err)
	}
	if len(orders) != 1 {
		t.Errorf("expected 1 inbound order, got %d", len(orders))
	}

	// Filter by status
	orders, err = orderRepo.ListOrders(ctx, repository.OrderFilter{
		Status: domain.OrderStatusDraft,
	})
	if err != nil {
		t.Fatalf("ListOrders by status failed: %v", err)
	}
	if len(orders) != 3 {
		t.Errorf("expected 3 draft orders, got %d", len(orders))
	}

	// Filter with limit
	orders, err = orderRepo.ListOrders(ctx, repository.OrderFilter{
		Limit: 2,
	})
	if err != nil {
		t.Fatalf("ListOrders with limit failed: %v", err)
	}
	if len(orders) != 2 {
		t.Errorf("expected 2 orders with limit, got %d", len(orders))
	}

	// Filter by non-matching warehouse
	orders, err = orderRepo.ListOrders(ctx, repository.OrderFilter{
		WarehouseID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("ListOrders by non-matching warehouse failed: %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("expected 0 orders for unknown warehouse, got %d", len(orders))
	}
}

func TestOrderRepo_UpdateOrderStatus(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)

	o := &domain.Order{
		OrderNo:     "TEST-ORD-STAT-001",
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: wh.ID,
		Status:      domain.OrderStatusDraft,
	}
	if err := orderRepo.CreateOrder(ctx, o); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Transition to confirmed
	if err := orderRepo.UpdateOrderStatus(ctx, o.ID, domain.OrderStatusConfirmed); err != nil {
		t.Fatalf("UpdateOrderStatus -> confirmed failed: %v", err)
	}

	got, err := orderRepo.GetOrder(ctx, o.ID)
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}
	if got.Status != domain.OrderStatusConfirmed {
		t.Errorf("status = %q, want confirmed", got.Status)
	}

	// Transition to completed — should set completed_at
	if err := orderRepo.UpdateOrderStatus(ctx, o.ID, domain.OrderStatusCompleted); err != nil {
		t.Fatalf("UpdateOrderStatus -> completed failed: %v", err)
	}

	got, err = orderRepo.GetOrder(ctx, o.ID)
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}
	if got.Status != domain.OrderStatusCompleted {
		t.Errorf("status = %q, want completed", got.Status)
	}
	if got.CompletedAt == nil {
		t.Error("expected completed_at to be set when transitioning to completed")
	}

	// Not found
	err = orderRepo.UpdateOrderStatus(ctx, uuid.New(), domain.OrderStatusCancelled)
	if err == nil {
		t.Error("expected error for nonexistent order")
	}
}

func TestOrderRepo_CreateOrder_Defaults(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)

	// Minimal order with no explicit status or priority
	o := &domain.Order{
		OrderNo:     "TEST-ORD-DEF-001",
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: wh.ID,
	}

	if err := orderRepo.CreateOrder(ctx, o); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if o.Status != domain.OrderStatusDraft {
		t.Errorf("status = %q, want draft (default)", o.Status)
	}
	if o.Priority != domain.OrderPriorityNormal {
		t.Errorf("priority = %q, want normal (default)", o.Priority)
	}

	got, err := orderRepo.GetOrder(ctx, o.ID)
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}
	if got.Status != domain.OrderStatusDraft {
		t.Errorf("status = %q, want draft", got.Status)
	}
}

// ── OrderLine Tests ────────────────────────────────────────

func TestOrderRepo_CreateAndGetOrderLines(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)
	sku := createTestSKU(t, ctx, invRepo)

	o := &domain.Order{
		OrderNo:     "TEST-ORD-OFL-001",
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateOrder(ctx, o); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Create lines
	line1 := &domain.OrderLine{
		OrderID:    o.ID,
		LineNo:     1,
		SKUID:      sku.ID,
		OrderedQty: 100.0,
		UOM:        "EA",
		BatchNo:    "BATCH-A",
	}
	line2 := &domain.OrderLine{
		OrderID:    o.ID,
		LineNo:     2,
		SKUID:      sku.ID,
		OrderedQty: 50.0,
		UOM:        "EA",
		BatchNo:    "BATCH-B",
	}

	if err := orderRepo.CreateOrderLine(ctx, line1); err != nil {
		t.Fatalf("CreateOrderLine [1] failed: %v", err)
	}
	if line1.ID == uuid.Nil {
		t.Error("expected line ID to be set")
	}
	if line1.Status != domain.OrderLineStatusPending {
		t.Errorf("line1 status = %q, want pending", line1.Status)
	}

	if err := orderRepo.CreateOrderLine(ctx, line2); err != nil {
		t.Fatalf("CreateOrderLine [2] failed: %v", err)
	}

	// Get lines
	lines, err := orderRepo.GetOrderLines(ctx, o.ID)
	if err != nil {
		t.Fatalf("GetOrderLines failed: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// Check ordering: sorted by line_no ASC
	if lines[0].LineNo != 1 {
		t.Errorf("lines[0].line_no = %d, want 1", lines[0].LineNo)
	}
	if lines[0].OrderedQty != 100.0 {
		t.Errorf("lines[0].ordered_qty = %f, want 100.0", lines[0].OrderedQty)
	}
	if lines[0].BatchNo != "BATCH-A" {
		t.Errorf("lines[0].batch_no = %q, want BATCH-A", lines[0].BatchNo)
	}
	if lines[1].LineNo != 2 {
		t.Errorf("lines[1].line_no = %d, want 2", lines[1].LineNo)
	}
	if lines[1].OrderedQty != 50.0 {
		t.Errorf("lines[1].ordered_qty = %f, want 50.0", lines[1].OrderedQty)
	}
}

func TestOrderRepo_GetOrderLines_Empty(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)

	o := &domain.Order{
		OrderNo:     "TEST-ORD-EMPTY-001",
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateOrder(ctx, o); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	lines, err := orderRepo.GetOrderLines(ctx, o.ID)
	if err != nil {
		t.Fatalf("GetOrderLines failed: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("expected 0 lines for order with no lines, got %d", len(lines))
	}
}

func TestOrderRepo_UpdateOrderLineStatus(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)
	sku := createTestSKU(t, ctx, invRepo)

	o := &domain.Order{
		OrderNo:     "TEST-ORD-OLS-001",
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateOrder(ctx, o); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	line := &domain.OrderLine{
		OrderID:    o.ID,
		LineNo:     1,
		SKUID:      sku.ID,
		OrderedQty: 100.0,
	}
	if err := orderRepo.CreateOrderLine(ctx, line); err != nil {
		t.Fatalf("CreateOrderLine failed: %v", err)
	}

	// Transition to allocated
	if err := orderRepo.UpdateOrderLineStatus(ctx, line.ID, domain.OrderLineStatusAllocated); err != nil {
		t.Fatalf("UpdateOrderLineStatus failed: %v", err)
	}

	lines, err := orderRepo.GetOrderLines(ctx, o.ID)
	if err != nil {
		t.Fatalf("GetOrderLines failed: %v", err)
	}
	if lines[0].Status != domain.OrderLineStatusAllocated {
		t.Errorf("status = %q, want allocated", lines[0].Status)
	}

	// Not found
	err = orderRepo.UpdateOrderLineStatus(ctx, uuid.New(), domain.OrderLineStatusCancelled)
	if err == nil {
		t.Error("expected error for nonexistent order line")
	}
}

func TestOrderRepo_UpdateOrderLineFulfilledQty(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)
	sku := createTestSKU(t, ctx, invRepo)

	o := &domain.Order{
		OrderNo:     "TEST-ORD-OLQ-001",
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateOrder(ctx, o); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	line := &domain.OrderLine{
		OrderID:      o.ID,
		LineNo:       1,
		SKUID:        sku.ID,
		OrderedQty:   100.0,
		FulfilledQty: 0.0,
	}
	if err := orderRepo.CreateOrderLine(ctx, line); err != nil {
		t.Fatalf("CreateOrderLine failed: %v", err)
	}

	// Update fulfilled qty
	if err := orderRepo.UpdateOrderLineFulfilledQty(ctx, line.ID, 30.0); err != nil {
		t.Fatalf("UpdateOrderLineFulfilledQty failed: %v", err)
	}

	lines, err := orderRepo.GetOrderLines(ctx, o.ID)
	if err != nil {
		t.Fatalf("GetOrderLines failed: %v", err)
	}
	if lines[0].FulfilledQty != 30.0 {
		t.Errorf("fulfilled_qty = %f, want 30.0", lines[0].FulfilledQty)
	}

	// Not found
	err = orderRepo.UpdateOrderLineFulfilledQty(ctx, uuid.New(), 50.0)
	if err == nil {
		t.Error("expected error for nonexistent order line")
	}
}

func TestOrderRepo_CreateOrderLine_Defaults(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)
	sku := createTestSKU(t, ctx, invRepo)

	o := &domain.Order{
		OrderNo:     "TEST-ORD-OLDEF-001",
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateOrder(ctx, o); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Minimal line with no explicit status or UOM
	line := &domain.OrderLine{
		OrderID:    o.ID,
		LineNo:     1,
		SKUID:      sku.ID,
		OrderedQty: 50.0,
	}
	if err := orderRepo.CreateOrderLine(ctx, line); err != nil {
		t.Fatalf("CreateOrderLine failed: %v", err)
	}
	if line.Status != domain.OrderLineStatusPending {
		t.Errorf("status = %q, want pending (default)", line.Status)
	}
	if line.UOM != "EA" {
		t.Errorf("uom = %q, want EA (default)", line.UOM)
	}

	lines, err := orderRepo.GetOrderLines(ctx, o.ID)
	if err != nil {
		t.Fatalf("GetOrderLines failed: %v", err)
	}
	if lines[0].FulfilledQty != 0.0 {
		t.Errorf("fulfilled_qty = %f, want 0.0 (default)", lines[0].FulfilledQty)
	}
}

// ── ASN Tests ─────────────────────────────────────────────

func TestOrderRepo_CreateAndGetASN(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)

	asn := &domain.ASN{
		ASNNo:       "TEST-ASN-001",
		WarehouseID: wh.ID,
		Carrier:     "FedEx Freight",
		TrackingNo:  "TRACK-12345",
	}

	err := orderRepo.CreateASN(ctx, asn)
	if err != nil {
		t.Fatalf("CreateASN failed: %v", err)
	}
	if asn.ID == uuid.Nil {
		t.Error("expected ASN ID to be set")
	}
	if asn.Status != domain.ASNStatusPending {
		t.Errorf("status = %q, want pending", asn.Status)
	}

	got, err := orderRepo.GetASN(ctx, asn.ID)
	if err != nil {
		t.Fatalf("GetASN failed: %v", err)
	}
	if got.ASNNo != asn.ASNNo {
		t.Errorf("asn_no = %q, want %q", got.ASNNo, asn.ASNNo)
	}
	if got.Carrier != "FedEx Freight" {
		t.Errorf("carrier = %q, want FedEx Freight", got.Carrier)
	}
	if got.TrackingNo != "TRACK-12345" {
		t.Errorf("tracking_no = %q, want TRACK-12345", got.TrackingNo)
	}
	if got.ArrivedAt != nil {
		t.Error("expected arrived_at to be nil for pending ASN")
	}
}

func TestOrderRepo_GetASNByNo(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)

	asn := &domain.ASN{
		ASNNo:       "TEST-ASN-BYNO-001",
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateASN(ctx, asn); err != nil {
		t.Fatalf("CreateASN failed: %v", err)
	}

	got, err := orderRepo.GetASNByNo(ctx, "TEST-ASN-BYNO-001")
	if err != nil {
		t.Fatalf("GetASNByNo failed: %v", err)
	}
	if got.ID != asn.ID {
		t.Errorf("id = %s, want %s", got.ID, asn.ID)
	}

	// Not found
	_, err = orderRepo.GetASNByNo(ctx, "NONEXISTENT-ASN")
	if err == nil {
		t.Error("expected error for nonexistent ASN")
	}
}

func TestOrderRepo_UpdateASNStatus(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)

	asn := &domain.ASN{
		ASNNo:       "TEST-ASN-STAT-001",
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateASN(ctx, asn); err != nil {
		t.Fatalf("CreateASN failed: %v", err)
	}

	// Transition to arrived — should set arrived_at
	if err := orderRepo.UpdateASNStatus(ctx, asn.ID, domain.ASNStatusArrived); err != nil {
		t.Fatalf("UpdateASNStatus -> arrived failed: %v", err)
	}

	got, err := orderRepo.GetASN(ctx, asn.ID)
	if err != nil {
		t.Fatalf("GetASN failed: %v", err)
	}
	if got.Status != domain.ASNStatusArrived {
		t.Errorf("status = %q, want arrived", got.Status)
	}
	if got.ArrivedAt == nil {
		t.Error("expected arrived_at to be set when transitioning to arrived")
	}

	// Transition to received
	if err := orderRepo.UpdateASNStatus(ctx, asn.ID, domain.ASNStatusReceived); err != nil {
		t.Fatalf("UpdateASNStatus -> received failed: %v", err)
	}

	got, err = orderRepo.GetASN(ctx, asn.ID)
	if err != nil {
		t.Fatalf("GetASN failed: %v", err)
	}
	if got.Status != domain.ASNStatusReceived {
		t.Errorf("status = %q, want received", got.Status)
	}

	// Not found
	err = orderRepo.UpdateASNStatus(ctx, uuid.New(), domain.ASNStatusArrived)
	if err == nil {
		t.Error("expected error for nonexistent ASN")
	}
}

func TestOrderRepo_CreateASN_WithOrder(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)

	o := &domain.Order{
		OrderNo:     "TEST-ORD-ASN-001",
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateOrder(ctx, o); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	asn := &domain.ASN{
		ASNNo:       "TEST-ASN-LINKED-001",
		WarehouseID: wh.ID,
		OrderID:     o.ID,
		Carrier:     "UPS Ground",
	}
	if err := orderRepo.CreateASN(ctx, asn); err != nil {
		t.Fatalf("CreateASN failed: %v", err)
	}

	got, err := orderRepo.GetASN(ctx, asn.ID)
	if err != nil {
		t.Fatalf("GetASN failed: %v", err)
	}
	if got.OrderID != o.ID {
		t.Errorf("order_id = %s, want %s", got.OrderID, o.ID)
	}
}

// ── ASNLine Tests ─────────────────────────────────────────

func TestOrderRepo_CreateAndGetASNLines(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)
	sku1 := createTestSKU(t, ctx, invRepo)
	sku2 := createTestSKU(t, ctx, invRepo)

	asn := &domain.ASN{
		ASNNo:       "TEST-ASN-LINES-001",
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateASN(ctx, asn); err != nil {
		t.Fatalf("CreateASN failed: %v", err)
	}

	// Create lines
	line1 := &domain.ASNLine{
		ASNID:       asn.ID,
		SKUID:       sku1.ID,
		ExpectedQty: 500.0,
		BatchNo:     "BATCH-2026-01",
	}
	line2 := &domain.ASNLine{
		ASNID:       asn.ID,
		SKUID:       sku2.ID,
		ExpectedQty: 300.0,
		BatchNo:     "BATCH-2026-02",
	}

	if err := orderRepo.CreateASNLine(ctx, line1); err != nil {
		t.Fatalf("CreateASNLine [1] failed: %v", err)
	}
	if line1.ID == uuid.Nil {
		t.Error("expected line ID to be set")
	}
	if line1.Status != domain.ASNLineStatusPending {
		t.Errorf("line1 status = %q, want pending", line1.Status)
	}

	if err := orderRepo.CreateASNLine(ctx, line2); err != nil {
		t.Fatalf("CreateASNLine [2] failed: %v", err)
	}

	// Get lines
	lines, err := orderRepo.GetASNLines(ctx, asn.ID)
	if err != nil {
		t.Fatalf("GetASNLines failed: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// Verify line data (order-independent — UUID ordering is not insertion order)
	foundSKU1 := false
	foundSKU2 := false
	for _, l := range lines {
		if l.SKUID == sku1.ID {
			foundSKU1 = true
			if l.ExpectedQty != 500.0 {
				t.Errorf("line sku1 expected_qty = %f, want 500.0", l.ExpectedQty)
			}
			if l.BatchNo != "BATCH-2026-01" {
				t.Errorf("line sku1 batch_no = %q, want BATCH-2026-01", l.BatchNo)
			}
		}
		if l.SKUID == sku2.ID {
			foundSKU2 = true
			if l.ExpectedQty != 300.0 {
				t.Errorf("line sku2 expected_qty = %f, want 300.0", l.ExpectedQty)
			}
			if l.BatchNo != "BATCH-2026-02" {
				t.Errorf("line sku2 batch_no = %q, want BATCH-2026-02", l.BatchNo)
			}
		}
		if l.ReceivedQty != 0.0 {
			t.Errorf("received_qty = %f, want 0.0 (default)", l.ReceivedQty)
		}
		if l.Status != domain.ASNLineStatusPending {
			t.Errorf("status = %q, want pending", l.Status)
		}
	}
	if !foundSKU1 {
		t.Error("sku1 not found in lines")
	}
	if !foundSKU2 {
		t.Error("sku2 not found in lines")
	}
}

func TestOrderRepo_GetASNLines_Empty(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)

	asn := &domain.ASN{
		ASNNo:       "TEST-ASN-EMPTY-001",
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateASN(ctx, asn); err != nil {
		t.Fatalf("CreateASN failed: %v", err)
	}

	lines, err := orderRepo.GetASNLines(ctx, asn.ID)
	if err != nil {
		t.Fatalf("GetASNLines failed: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("expected 0 lines for ASN with no lines, got %d", len(lines))
	}
}

func TestOrderRepo_UpdateASNLineStatus(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)
	sku := createTestSKU(t, ctx, invRepo)

	asn := &domain.ASN{
		ASNNo:       "TEST-ASN-LSTAT-001",
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateASN(ctx, asn); err != nil {
		t.Fatalf("CreateASN failed: %v", err)
	}

	line := &domain.ASNLine{
		ASNID:       asn.ID,
		SKUID:       sku.ID,
		ExpectedQty: 200.0,
	}
	if err := orderRepo.CreateASNLine(ctx, line); err != nil {
		t.Fatalf("CreateASNLine failed: %v", err)
	}

	// Transition to partial
	if err := orderRepo.UpdateASNLineStatus(ctx, line.ID, domain.ASNLineStatusPartial); err != nil {
		t.Fatalf("UpdateASNLineStatus -> partial failed: %v", err)
	}

	lines, err := orderRepo.GetASNLines(ctx, asn.ID)
	if err != nil {
		t.Fatalf("GetASNLines failed: %v", err)
	}
	if lines[0].Status != domain.ASNLineStatusPartial {
		t.Errorf("status = %q, want partial", lines[0].Status)
	}

	// Transition to received
	if err := orderRepo.UpdateASNLineStatus(ctx, line.ID, domain.ASNLineStatusReceived); err != nil {
		t.Fatalf("UpdateASNLineStatus -> received failed: %v", err)
	}

	lines, err = orderRepo.GetASNLines(ctx, asn.ID)
	if err != nil {
		t.Fatalf("GetASNLines failed: %v", err)
	}
	if lines[0].Status != domain.ASNLineStatusReceived {
		t.Errorf("status = %q, want received", lines[0].Status)
	}

	// Not found
	err = orderRepo.UpdateASNLineStatus(ctx, uuid.New(), domain.ASNLineStatusPartial)
	if err == nil {
		t.Error("expected error for nonexistent ASN line")
	}
}

func TestOrderRepo_UpdateASNLineReceivedQty(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)
	sku := createTestSKU(t, ctx, invRepo)

	asn := &domain.ASN{
		ASNNo:       "TEST-ASN-LQTY-001",
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateASN(ctx, asn); err != nil {
		t.Fatalf("CreateASN failed: %v", err)
	}

	line := &domain.ASNLine{
		ASNID:       asn.ID,
		SKUID:       sku.ID,
		ExpectedQty: 500.0,
	}
	if err := orderRepo.CreateASNLine(ctx, line); err != nil {
		t.Fatalf("CreateASNLine failed: %v", err)
	}

	// Partial receive
	if err := orderRepo.UpdateASNLineReceivedQty(ctx, line.ID, 200.0); err != nil {
		t.Fatalf("UpdateASNLineReceivedQty [200] failed: %v", err)
	}

	lines, err := orderRepo.GetASNLines(ctx, asn.ID)
	if err != nil {
		t.Fatalf("GetASNLines failed: %v", err)
	}
	if lines[0].ReceivedQty != 200.0 {
		t.Errorf("received_qty = %f, want 200.0", lines[0].ReceivedQty)
	}

	// Full receive
	if err := orderRepo.UpdateASNLineReceivedQty(ctx, line.ID, 500.0); err != nil {
		t.Fatalf("UpdateASNLineReceivedQty [500] failed: %v", err)
	}

	lines, err = orderRepo.GetASNLines(ctx, asn.ID)
	if err != nil {
		t.Fatalf("GetASNLines failed: %v", err)
	}
	if lines[0].ReceivedQty != 500.0 {
		t.Errorf("received_qty = %f, want 500.0", lines[0].ReceivedQty)
	}

	// Not found
	err = orderRepo.UpdateASNLineReceivedQty(ctx, uuid.New(), 100.0)
	if err == nil {
		t.Error("expected error for nonexistent ASN line")
	}
}

func TestOrderRepo_CreateASNLine_Defaults(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)
	sku := createTestSKU(t, ctx, invRepo)

	asn := &domain.ASN{
		ASNNo:       "TEST-ASN-LDEF-001",
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateASN(ctx, asn); err != nil {
		t.Fatalf("CreateASN failed: %v", err)
	}

	// Minimal line with no explicit status
	line := &domain.ASNLine{
		ASNID:       asn.ID,
		SKUID:       sku.ID,
		ExpectedQty: 150.0,
	}
	if err := orderRepo.CreateASNLine(ctx, line); err != nil {
		t.Fatalf("CreateASNLine failed: %v", err)
	}
	if line.Status != domain.ASNLineStatusPending {
		t.Errorf("status = %q, want pending (default)", line.Status)
	}
	if line.ReceivedQty != 0.0 {
		t.Errorf("received_qty = %f, want 0.0 (default)", line.ReceivedQty)
	}

	lines, err := orderRepo.GetASNLines(ctx, asn.ID)
	if err != nil {
		t.Fatalf("GetASNLines failed: %v", err)
	}
	if lines[0].ReceivedQty != 0.0 {
		t.Errorf("received_qty = %f, want 0.0", lines[0].ReceivedQty)
	}
}

func TestOrderRepo_CreateASNLine_WithMultipleLines(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)
	sku := createTestSKU(t, ctx, invRepo)

	asn := &domain.ASN{
		ASNNo:       "TEST-ASN-MULTI-001",
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateASN(ctx, asn); err != nil {
		t.Fatalf("CreateASN failed: %v", err)
	}

	// Create 5 lines
	for i := 1; i <= 5; i++ {
		line := &domain.ASNLine{
			ASNID:       asn.ID,
			SKUID:       sku.ID,
			ExpectedQty: float64(i * 100),
		}
		if err := orderRepo.CreateASNLine(ctx, line); err != nil {
			t.Fatalf("CreateASNLine [%d] failed: %v", i, err)
		}
	}

	lines, err := orderRepo.GetASNLines(ctx, asn.ID)
	if err != nil {
		t.Fatalf("GetASNLines failed: %v", err)
	}
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}

	// Verify expected quantities (order-independent)
	qtySet := make(map[float64]bool)
	for _, l := range lines {
		qtySet[l.ExpectedQty] = true
	}
	for i := 1; i <= 5; i++ {
		if !qtySet[float64(i*100)] {
			t.Errorf("expected_qty %d not found in lines", i*100)
		}
	}
}

func TestOrderRepo_UpdateASNLineStatus_TransitionToReceived(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)
	sku := createTestSKU(t, ctx, invRepo)

	asn := &domain.ASN{
		ASNNo:       "TEST-ASN-RCVD-001",
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateASN(ctx, asn); err != nil {
		t.Fatalf("CreateASN failed: %v", err)
	}

	line := &domain.ASNLine{
		ASNID:       asn.ID,
		SKUID:       sku.ID,
		ExpectedQty: 500.0,
	}
	if err := orderRepo.CreateASNLine(ctx, line); err != nil {
		t.Fatalf("CreateASNLine failed: %v", err)
	}

	// Receive full qty and transition to received
	if err := orderRepo.UpdateASNLineReceivedQty(ctx, line.ID, 500.0); err != nil {
		t.Fatalf("UpdateASNLineReceivedQty failed: %v", err)
	}
	if err := orderRepo.UpdateASNLineStatus(ctx, line.ID, domain.ASNLineStatusReceived); err != nil {
		t.Fatalf("UpdateASNLineStatus -> received failed: %v", err)
	}

	lines, err := orderRepo.GetASNLines(ctx, asn.ID)
	if err != nil {
		t.Fatalf("GetASNLines failed: %v", err)
	}
	if lines[0].Status != domain.ASNLineStatusReceived {
		t.Errorf("status = %q, want received", lines[0].Status)
	}
	if lines[0].ReceivedQty != 500.0 {
		t.Errorf("received_qty = %f, want 500.0", lines[0].ReceivedQty)
	}
}

// ── Additional Order Repo Tests ───────────────────────────

func TestOrderRepo_CountOrdersByStatus(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)

	// Create 2 draft, 1 confirmed order
	draft1 := &domain.Order{OrderNo: "TEST-ORD-CNT-001", OrderType: domain.OrderTypeInbound, WarehouseID: wh.ID, Status: domain.OrderStatusDraft}
	draft2 := &domain.Order{OrderNo: "TEST-ORD-CNT-002", OrderType: domain.OrderTypeOutbound, WarehouseID: wh.ID, Status: domain.OrderStatusDraft}
	confirmed := &domain.Order{OrderNo: "TEST-ORD-CNT-003", OrderType: domain.OrderTypeTransfer, WarehouseID: wh.ID, Status: domain.OrderStatusConfirmed}
	if err := orderRepo.CreateOrder(ctx, draft1); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if err := orderRepo.CreateOrder(ctx, draft2); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if err := orderRepo.CreateOrder(ctx, confirmed); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	counts, err := orderRepo.CountOrdersByStatus(ctx)
	if err != nil {
		t.Fatalf("CountOrdersByStatus failed: %v", err)
	}
	// At minimum we expect 2 drafts
	if counts[domain.OrderStatusDraft] < 2 {
		t.Errorf("expected at least 2 drafts, got %d", counts[domain.OrderStatusDraft])
	}
	if counts[domain.OrderStatusConfirmed] < 1 {
		t.Errorf("expected at least 1 confirmed, got %d", counts[domain.OrderStatusConfirmed])
	}
}

func TestOrderRepo_ListASNs(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)

	// Create 3 ASNs
	for i := range 3 {
		asn := &domain.ASN{
			ASNNo:       "TEST-ASN-LIST-00" + string(rune('1'+i)),
			WarehouseID: wh.ID,
			Status:      domain.ASNStatusPending,
		}
		if err := orderRepo.CreateASN(ctx, asn); err != nil {
			t.Fatalf("CreateASN failed: %v", err)
		}
	}
	// Create one arrived ASN
	asn4 := &domain.ASN{
		ASNNo:       "TEST-ASN-LIST-004",
		WarehouseID: wh.ID,
		Status:      domain.ASNStatusArrived,
	}
	if err := orderRepo.CreateASN(ctx, asn4); err != nil {
		t.Fatalf("CreateASN arrived failed: %v", err)
	}

	// List all
	all, err := orderRepo.ListASNs(ctx, repository.ASNFilter{})
	if err != nil {
		t.Fatalf("ListASNs failed: %v", err)
	}
	if len(all) < 4 {
		t.Errorf("expected at least 4 ASNs, got %d", len(all))
	}

	// Filter by warehouse
	byWH, err := orderRepo.ListASNs(ctx, repository.ASNFilter{WarehouseID: wh.ID})
	if err != nil {
		t.Fatalf("ListASNs by warehouse failed: %v", err)
	}
	if len(byWH) != 4 {
		t.Errorf("expected 4 ASNs for warehouse, got %d", len(byWH))
	}

	// Filter by status
	byStatus, err := orderRepo.ListASNs(ctx, repository.ASNFilter{Status: domain.ASNStatusPending})
	if err != nil {
		t.Fatalf("ListASNs by status failed: %v", err)
	}
	if len(byStatus) != 3 {
		t.Errorf("expected 3 pending ASNs, got %d", len(byStatus))
	}

	// Filter by ASNNo
	byNo, err := orderRepo.ListASNs(ctx, repository.ASNFilter{ASNNo: "TEST-ASN-LIST-001"})
	if err != nil {
		t.Fatalf("ListASNs by ASNNo failed: %v", err)
	}
	if len(byNo) != 1 {
		t.Errorf("expected 1 ASN by no, got %d", len(byNo))
	}

	// With limit
	limited, err := orderRepo.ListASNs(ctx, repository.ASNFilter{Limit: 2})
	if err != nil {
		t.Fatalf("ListASNs with limit failed: %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("expected 2 ASNs with limit, got %d", len(limited))
	}
}

func TestOrderRepo_CountASNs(t *testing.T) {
	db, cleanup := setupOrderTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	orderRepo := NewOrderRepo(db)

	wh := createTestWarehouse(t, ctx, whRepo)

	// Create 3 ASNs
	for i := range 3 {
		asn := &domain.ASN{
			ASNNo:       "TEST-ASN-CNT-00" + string(rune('1'+i)),
			WarehouseID: wh.ID,
		}
		if err := orderRepo.CreateASN(ctx, asn); err != nil {
			t.Fatalf("CreateASN failed: %v", err)
		}
	}

	count, err := orderRepo.CountASNs(ctx, repository.ASNFilter{WarehouseID: wh.ID})
	if err != nil {
		t.Fatalf("CountASNs failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 ASNs, got %d", count)
	}

	// Zero for different warehouse
	count, err = orderRepo.CountASNs(ctx, repository.ASNFilter{WarehouseID: uuid.New()})
	if err != nil {
		t.Fatalf("CountASNs for unknown warehouse failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 ASNs for unknown warehouse, got %d", count)
	}
}
