package service

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// mockOrderRepo implements repository.OrderRepository for testing.
type mockOrderRepo struct {
	orders     map[uuid.UUID]*domain.Order
	orderLines map[uuid.UUID][]*domain.OrderLine
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{
		orders:     make(map[uuid.UUID]*domain.Order),
		orderLines: make(map[uuid.UUID][]*domain.OrderLine),
	}
}

// ── Order ──────────────────────────────────────────────────

func (m *mockOrderRepo) CreateOrder(ctx context.Context, o *domain.Order) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	m.orders[o.ID] = o
	m.orderLines[o.ID] = []*domain.OrderLine{}
	return nil
}

func (m *mockOrderRepo) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	o, ok := m.orders[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("order", id.String())
	}
	return o, nil
}

func (m *mockOrderRepo) GetOrderByNo(ctx context.Context, orderNo string) (*domain.Order, error) {
	for _, o := range m.orders {
		if o.OrderNo == orderNo {
			return o, nil
		}
	}
	return nil, pkgerrors.NewNotFound("order", orderNo)
}

func (m *mockOrderRepo) ListOrders(ctx context.Context, filter repository.OrderFilter) ([]*domain.Order, error) {
	var result []*domain.Order
	for _, o := range m.orders {
		if filter.WarehouseID != uuid.Nil && o.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.OrderType != "" && o.OrderType != filter.OrderType {
			continue
		}
		if filter.Status != "" && o.Status != filter.Status {
			continue
		}
		result = append(result, o)
	}
	return result, nil
}

func (m *mockOrderRepo) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	o, ok := m.orders[id]
	if !ok {
		return pkgerrors.NewNotFound("order", id.String())
	}
	o.Status = status
	return nil
}

func (m *mockOrderRepo) CountOrders(ctx context.Context, filter repository.OrderFilter) (int, error) {
	count := 0
	for _, o := range m.orders {
		if filter.WarehouseID != uuid.Nil && o.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.OrderType != "" && o.OrderType != filter.OrderType {
			continue
		}
		if filter.Status != "" && o.Status != filter.Status {
			continue
		}
		count++
	}
	return count, nil
}

// ── OrderLine ───────────────────────────────────────────────

func (m *mockOrderRepo) CreateOrderLine(ctx context.Context, line *domain.OrderLine) error {
	if line.ID == uuid.Nil {
		line.ID = uuid.New()
	}
	m.orderLines[line.OrderID] = append(m.orderLines[line.OrderID], line)
	return nil
}

func (m *mockOrderRepo) GetOrderLines(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderLine, error) {
	lines, ok := m.orderLines[orderID]
	if !ok {
		return nil, nil
	}
	return lines, nil
}

func (m *mockOrderRepo) UpdateOrderLineStatus(ctx context.Context, id uuid.UUID, status domain.OrderLineStatus) error {
	for _, lines := range m.orderLines {
		for _, l := range lines {
			if l.ID == id {
				l.Status = status
				return nil
			}
		}
	}
	return pkgerrors.NewNotFound("order line", id.String())
}

func (m *mockOrderRepo) UpdateOrderLineFulfilledQty(ctx context.Context, id uuid.UUID, qty float64) error {
	for _, lines := range m.orderLines {
		for _, l := range lines {
			if l.ID == id {
				l.FulfilledQty = qty
				return nil
			}
		}
	}
	return pkgerrors.NewNotFound("order line", id.String())
}

// ── ASN (not used by OrderService tests) ───────────────────

func (m *mockOrderRepo) CreateASN(ctx context.Context, asn *domain.ASN) error { return nil }
func (m *mockOrderRepo) GetASN(ctx context.Context, id uuid.UUID) (*domain.ASN, error) {
	return nil, nil
}
func (m *mockOrderRepo) GetASNByNo(ctx context.Context, asnNo string) (*domain.ASN, error) {
	return nil, nil
}
func (m *mockOrderRepo) UpdateASNStatus(ctx context.Context, id uuid.UUID, status domain.ASNStatus) error {
	return nil
}

// ── ASNLine (not used by OrderService tests) ────────────────

func (m *mockOrderRepo) CreateASNLine(ctx context.Context, line *domain.ASNLine) error { return nil }
func (m *mockOrderRepo) GetASNLines(ctx context.Context, asnID uuid.UUID) ([]*domain.ASNLine, error) {
	return nil, nil
}
func (m *mockOrderRepo) UpdateASNLineStatus(ctx context.Context, id uuid.UUID, status domain.ASNLineStatus) error {
	return nil
}
func (m *mockOrderRepo) UpdateASNLineReceivedQty(ctx context.Context, id uuid.UUID, qty float64) error {
	return nil
}

// ── Tests ───────────────────────────────────────────────────────────────────

func TestOrderService_CreateOrder_Inbound(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	skuID := uuid.New()
	whID := uuid.New()

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: whID,
		Lines: []CreateOrderLineInput{
			{SKUID: skuID, OrderedQty: 100},
		},
		CreatedBy: "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if order.Status != domain.OrderStatusDraft {
		t.Errorf("status = %q, want %q", order.Status, domain.OrderStatusDraft)
	}
	if order.Priority != domain.OrderPriorityNormal {
		t.Errorf("priority = %q, want %q", order.Priority, domain.OrderPriorityNormal)
	}
	if !strings.HasPrefix(order.OrderNo, "IN-") {
		t.Errorf("order_no should start with IN-: got %q", order.OrderNo)
	}
	if len(order.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(order.Lines))
	}
	if order.Lines[0].SKUID != skuID {
		t.Errorf("line sku_id = %q, want %q", order.Lines[0].SKUID, skuID)
	}
	if order.Lines[0].OrderedQty != 100 {
		t.Errorf("line ordered_qty = %f, want %f", order.Lines[0].OrderedQty, 100.0)
	}
	if order.Lines[0].LineNo != 1 {
		t.Errorf("line line_no = %d, want 1", order.Lines[0].LineNo)
	}
}

func TestOrderService_CreateOrder_OutboundWithPriority(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Priority:    domain.OrderPriorityUrgent,
		Lines: []CreateOrderLineInput{
			{SKUID: uuid.New(), OrderedQty: 50, UOM: "CS"},
			{SKUID: uuid.New(), OrderedQty: 200},
		},
		CreatedBy: "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if order.Priority != domain.OrderPriorityUrgent {
		t.Errorf("priority = %q, want %q", order.Priority, domain.OrderPriorityUrgent)
	}
	if !strings.HasPrefix(order.OrderNo, "OUT-") {
		t.Errorf("order_no should start with OUT-: got %q", order.OrderNo)
	}
	if len(order.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(order.Lines))
	}
	if order.Lines[0].UOM != "CS" {
		t.Errorf("line[0] uom = %q, want CS", order.Lines[0].UOM)
	}
	if order.Lines[1].UOM != "EA" {
		t.Errorf("line[1] uom = %q, want EA (default)", order.Lines[1].UOM)
	}
	if order.Lines[0].LineNo != 1 {
		t.Errorf("line[0] line_no = %d, want 1", order.Lines[0].LineNo)
	}
	if order.Lines[1].LineNo != 2 {
		t.Errorf("line[1] line_no = %d, want 2", order.Lines[1].LineNo)
	}
}

func TestOrderService_CreateOrder_CustomOrderNo(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderNo:     "CUSTOM-ORDER-001",
		OrderType:   domain.OrderTypeTransfer,
		WarehouseID: uuid.New(),
		Lines: []CreateOrderLineInput{
			{SKUID: uuid.New(), OrderedQty: 10},
		},
		CreatedBy: "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if order.OrderNo != "CUSTOM-ORDER-001" {
		t.Errorf("order_no = %q, want %q", order.OrderNo, "CUSTOM-ORDER-001")
	}
}

func TestOrderService_CreateOrder_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	tests := []struct {
		name  string
		input CreateOrderInput
	}{
		{"empty order type", CreateOrderInput{
			OrderType:   "",
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 1}},
			CreatedBy:   "user",
		}},
		{"invalid order type", CreateOrderInput{
			OrderType:   "invalid_type",
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 1}},
			CreatedBy:   "user",
		}},
		{"nil warehouse id", CreateOrderInput{
			OrderType:   domain.OrderTypeInbound,
			WarehouseID: uuid.Nil,
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 1}},
			CreatedBy:   "user",
		}},
		{"no lines", CreateOrderInput{
			OrderType:   domain.OrderTypeInbound,
			WarehouseID: uuid.New(),
			Lines:       nil,
			CreatedBy:   "user",
		}},
		{"empty lines", CreateOrderInput{
			OrderType:   domain.OrderTypeInbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{},
			CreatedBy:   "user",
		}},
		{"missing created_by", CreateOrderInput{
			OrderType:   domain.OrderTypeInbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 1}},
			CreatedBy:   "",
		}},
		{"invalid priority", CreateOrderInput{
			OrderType:   domain.OrderTypeInbound,
			WarehouseID: uuid.New(),
			Priority:    "super-urgent",
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 1}},
			CreatedBy:   "user",
		}},
		{"zero qty line", CreateOrderInput{
			OrderType:   domain.OrderTypeInbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 0}},
			CreatedBy:   "user",
		}},
		{"negative qty line", CreateOrderInput{
			OrderType:   domain.OrderTypeInbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: -5}},
			CreatedBy:   "user",
		}},
		{"nil sku id line", CreateOrderInput{
			OrderType:   domain.OrderTypeInbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.Nil, OrderedQty: 1}},
			CreatedBy:   "user",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateOrder(ctx, tt.input)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestOrderService_GetOrder(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	created, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines: []CreateOrderLineInput{
			{SKUID: uuid.New(), OrderedQty: 50},
			{SKUID: uuid.New(), OrderedQty: 30},
		},
		CreatedBy: "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	got, err := svc.GetOrder(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}
	if got.OrderNo != created.OrderNo {
		t.Errorf("order_no = %q, want %q", got.OrderNo, created.OrderNo)
	}
	if len(got.Lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(got.Lines))
	}
}

func TestOrderService_GetOrder_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	_, err := svc.GetOrder(ctx, uuid.New())
	if err == nil {
		t.Fatal("expected error for unknown order")
	}
}

func TestOrderService_GetOrderByNo(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	created, _ := svc.CreateOrder(ctx, CreateOrderInput{
		OrderNo:     "MY-ORDER-42",
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "testuser",
	})

	got, err := svc.GetOrderByNo(ctx, "MY-ORDER-42")
	if err != nil {
		t.Fatalf("GetOrderByNo failed: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("id = %q, want %q", got.ID, created.ID)
	}
	if len(got.Lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(got.Lines))
	}
}

func TestOrderService_ListOrders(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	svc := NewOrderService(repo)

	wh1 := uuid.New()
	wh2 := uuid.New()

	// Create orders in different warehouses with different types.
	svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: wh1,
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "user",
	})
	svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: wh1,
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 20}},
		CreatedBy:   "user",
	})
	svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: wh2,
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 30}},
		CreatedBy:   "user",
	})

	// All orders.
	all, _, err := svc.ListOrders(ctx, repository.OrderFilter{})
	if err != nil {
		t.Fatalf("ListOrders failed: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 orders, got %d", len(all))
	}

	// Filter by warehouse.
	wh1Orders, _, err := svc.ListOrders(ctx, repository.OrderFilter{WarehouseID: wh1})
	if err != nil {
		t.Fatalf("ListOrders wh1 failed: %v", err)
	}
	if len(wh1Orders) != 2 {
		t.Errorf("expected 2 orders in wh1, got %d", len(wh1Orders))
	}

	// Filter by order_type.
	inbound, _, err := svc.ListOrders(ctx, repository.OrderFilter{OrderType: domain.OrderTypeInbound})
	if err != nil {
		t.Fatalf("ListOrders inbound failed: %v", err)
	}
	if len(inbound) != 2 {
		t.Errorf("expected 2 inbound orders, got %d", len(inbound))
	}
}

func TestOrderService_UpdateOrderStatus_ValidTransitions(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "user",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// draft → confirmed
	updated, err := svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
	if err != nil {
		t.Fatalf("draft → confirmed failed: %v", err)
	}
	if updated.Status != domain.OrderStatusConfirmed {
		t.Errorf("status = %q, want %q", updated.Status, domain.OrderStatusConfirmed)
	}

	// confirmed → processing
	updated, err = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusProcessing})
	if err != nil {
		t.Fatalf("confirmed → processing failed: %v", err)
	}
	if updated.Status != domain.OrderStatusProcessing {
		t.Errorf("status = %q, want %q", updated.Status, domain.OrderStatusProcessing)
	}

	// processing → completed
	updated, err = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusCompleted})
	if err != nil {
		t.Fatalf("processing → completed failed: %v", err)
	}
	if updated.Status != domain.OrderStatusCompleted {
		t.Errorf("status = %q, want %q", updated.Status, domain.OrderStatusCompleted)
	}
}

func TestOrderService_UpdateOrderStatus_CancelFromAny(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	order, _ := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "user",
	})

	// draft → cancelled (valid)
	_, err := svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusCancelled})
	if err != nil {
		t.Fatalf("draft → cancelled failed: %v", err)
	}
}

func TestOrderService_UpdateOrderStatus_InvalidTransitions(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	order, _ := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "user",
	})

	// draft → completed (invalid — must go through confirmed, processing)
	_, err := svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusCompleted})
	if err == nil {
		t.Fatal("expected error for draft → completed transition")
	}

	// draft → processing (invalid — must go through confirmed)
	_, err = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusProcessing})
	if err == nil {
		t.Fatal("expected error for draft → processing transition")
	}

	// Advance to confirmed, then try confirmed → completed (invalid — must go through processing).
	svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
	_, err = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusCompleted})
	if err == nil {
		t.Fatal("expected error for confirmed → completed transition")
	}
}

func TestOrderService_UpdateOrderStatus_TerminalStates(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	// Create and cancel an order.
	order1, _ := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "user",
	})
	svc.UpdateOrderStatus(ctx, order1.ID, UpdateOrderStatusInput{Status: domain.OrderStatusCancelled})

	// Cancelled → something else (should fail).
	_, err := svc.UpdateOrderStatus(ctx, order1.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
	if err == nil {
		t.Fatal("expected error for cancelled → confirmed transition")
	}

	// Create and complete an order.
	order2, _ := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "user",
	})
	svc.UpdateOrderStatus(ctx, order2.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
	svc.UpdateOrderStatus(ctx, order2.ID, UpdateOrderStatusInput{Status: domain.OrderStatusProcessing})
	svc.UpdateOrderStatus(ctx, order2.ID, UpdateOrderStatusInput{Status: domain.OrderStatusCompleted})

	// Completed → something else (should fail).
	_, err = svc.UpdateOrderStatus(ctx, order2.ID, UpdateOrderStatusInput{Status: domain.OrderStatusCancelled})
	if err == nil {
		t.Fatal("expected error for completed → cancelled transition")
	}
}

func TestOrderService_AddOrderLine(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	order, _ := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "user",
	})

	skuID := uuid.New()
	line, err := svc.AddOrderLine(ctx, order.ID, AddOrderLineInput{
		SKUID:      skuID,
		OrderedQty: 25,
		UOM:        "BX",
		Notes:      "added later",
	})
	if err != nil {
		t.Fatalf("AddOrderLine failed: %v", err)
	}
	if line.LineNo != 2 {
		t.Errorf("line_no = %d, want 2", line.LineNo)
	}
	if line.SKUID != skuID {
		t.Errorf("sku_id = %q, want %q", line.SKUID, skuID)
	}
	if line.UOM != "BX" {
		t.Errorf("uom = %q, want BX", line.UOM)
	}
	if line.Notes != "added later" {
		t.Errorf("notes = %q, want 'added later'", line.Notes)
	}

	// Verify the order now has 2 lines.
	got, _ := svc.GetOrder(ctx, order.ID)
	if len(got.Lines) != 2 {
		t.Errorf("expected 2 lines after add, got %d", len(got.Lines))
	}
}

func TestOrderService_AddOrderLine_NotDraft(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	order, _ := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "user",
	})

	// Confirm the order first.
	svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})

	_, err := svc.AddOrderLine(ctx, order.ID, AddOrderLineInput{
		SKUID:      uuid.New(),
		OrderedQty: 5,
	})
	if err == nil {
		t.Fatal("expected error when adding line to non-draft order")
	}
}

func TestOrderService_AddOrderLine_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	order, _ := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "user",
	})

	// Zero qty.
	_, err := svc.AddOrderLine(ctx, order.ID, AddOrderLineInput{
		SKUID:      uuid.New(),
		OrderedQty: 0,
	})
	if err == nil {
		t.Fatal("expected error for zero ordered_qty")
	}

	// Nil sku_id.
	_, err = svc.AddOrderLine(ctx, order.ID, AddOrderLineInput{
		SKUID:      uuid.Nil,
		OrderedQty: 5,
	})
	if err == nil {
		t.Fatal("expected error for nil sku_id")
	}
}

func TestOrderService_OrderTypeVariants(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	whID := uuid.New()
	skuID := uuid.New()
	line := CreateOrderLineInput{SKUID: skuID, OrderedQty: 10}

	tests := []struct {
		orderType domain.OrderType
		prefix    string
	}{
		{domain.OrderTypeInbound, "IN-"},
		{domain.OrderTypeOutbound, "OUT-"},
		{domain.OrderTypeTransfer, "TR-"},
		{domain.OrderTypeReturn, "RET-"},
	}

	for _, tt := range tests {
		t.Run(string(tt.orderType), func(t *testing.T) {
			order, err := svc.CreateOrder(ctx, CreateOrderInput{
				OrderType:   tt.orderType,
				WarehouseID: whID,
				Lines:       []CreateOrderLineInput{line},
				CreatedBy:   "user",
			})
			if err != nil {
				t.Fatalf("CreateOrder %s failed: %v", tt.orderType, err)
			}
			if !strings.HasPrefix(order.OrderNo, tt.prefix) {
				t.Errorf("order_no = %q, want prefix %q", order.OrderNo, tt.prefix)
			}
		})
	}
}

func TestOrderService_PartialStatus(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(newMockOrderRepo())

	order, _ := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "user",
	})

	// Advance to processing.
	svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
	svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusProcessing})

	// processing → partial (valid)
	updated, err := svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusPartial})
	if err != nil {
		t.Fatalf("processing → partial failed: %v", err)
	}
	if updated.Status != domain.OrderStatusPartial {
		t.Errorf("status = %q, want %q", updated.Status, domain.OrderStatusPartial)
	}

	// partial → completed (valid)
	updated, err = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusCompleted})
	if err != nil {
		t.Fatalf("partial → completed failed: %v", err)
	}
	if updated.Status != domain.OrderStatusCompleted {
		t.Errorf("status = %q, want %q", updated.Status, domain.OrderStatusCompleted)
	}
}
