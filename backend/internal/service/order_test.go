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
	asns       map[uuid.UUID]*domain.ASN
	asnLines   map[uuid.UUID][]*domain.ASNLine
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{
		orders:     make(map[uuid.UUID]*domain.Order),
		orderLines: make(map[uuid.UUID][]*domain.OrderLine),
		asns:       make(map[uuid.UUID]*domain.ASN),
		asnLines:   make(map[uuid.UUID][]*domain.ASNLine),
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

func (m *mockOrderRepo) GetOrderLine(ctx context.Context, id uuid.UUID) (*domain.OrderLine, error) {
	for _, lines := range m.orderLines {
		for _, l := range lines {
			if l.ID == id {
				return l, nil
			}
		}
	}
	return nil, pkgerrors.NewNotFound("order line", id.String())
}

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

// ── ASN ──────────────────────────────────────────────────────

func (m *mockOrderRepo) CreateASN(ctx context.Context, asn *domain.ASN) error {
	if asn.ID == uuid.Nil {
		asn.ID = uuid.New()
	}
	m.asns[asn.ID] = asn
	m.asnLines[asn.ID] = []*domain.ASNLine{}
	return nil
}

func (m *mockOrderRepo) GetASN(ctx context.Context, id uuid.UUID) (*domain.ASN, error) {
	a, ok := m.asns[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("asn", id.String())
	}
	return a, nil
}

func (m *mockOrderRepo) GetASNByNo(ctx context.Context, asnNo string) (*domain.ASN, error) {
	for _, a := range m.asns {
		if a.ASNNo == asnNo {
			return a, nil
		}
	}
	return nil, pkgerrors.NewNotFound("asn", asnNo)
}

func (m *mockOrderRepo) UpdateASNStatus(ctx context.Context, id uuid.UUID, status domain.ASNStatus) error {
	a, ok := m.asns[id]
	if !ok {
		return pkgerrors.NewNotFound("asn", id.String())
	}
	a.Status = status
	return nil
}

// ── ASNLine ──────────────────────────────────────────────────

func (m *mockOrderRepo) CreateASNLine(ctx context.Context, line *domain.ASNLine) error {
	if line.ID == uuid.Nil {
		line.ID = uuid.New()
	}
	m.asnLines[line.ASNID] = append(m.asnLines[line.ASNID], line)
	return nil
}

func (m *mockOrderRepo) GetASNLines(ctx context.Context, asnID uuid.UUID) ([]*domain.ASNLine, error) {
	lines, ok := m.asnLines[asnID]
	if !ok {
		return nil, nil
	}
	return lines, nil
}

func (m *mockOrderRepo) UpdateASNLineStatus(ctx context.Context, id uuid.UUID, status domain.ASNLineStatus) error {
	for _, lines := range m.asnLines {
		for _, l := range lines {
			if l.ID == id {
				l.Status = status
				return nil
			}
		}
	}
	return pkgerrors.NewNotFound("asn line", id.String())
}

func (m *mockOrderRepo) UpdateASNLineReceivedQty(ctx context.Context, id uuid.UUID, qty float64) error {
	for _, lines := range m.asnLines {
		for _, l := range lines {
			if l.ID == id {
				l.ReceivedQty = qty
				return nil
			}
		}
	}
	return pkgerrors.NewNotFound("asn line", id.String())
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

	// ── UpdateOrderLineStatus Tests ────────────────────────────────────────────────

	func TestOrderService_UpdateOrderLineStatus_ValidTransitions(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := NewOrderService(repo)

		order, _ := svc.CreateOrder(ctx, CreateOrderInput{
			OrderType:   domain.OrderTypeOutbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
			CreatedBy:   "user",
		})
		line := &order.Lines[0]

		// pending → allocated
		updated, err := svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusAllocated})
		if err != nil {
			t.Fatalf("pending → allocated failed: %v", err)
		}
		if updated.Status != domain.OrderLineStatusAllocated {
			t.Errorf("status = %q, want %q", updated.Status, domain.OrderLineStatusAllocated)
		}

		// allocated → fulfilled (skipping partial)
		updated, err = svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusFulfilled})
		if err != nil {
			t.Fatalf("allocated → fulfilled failed: %v", err)
		}
		if updated.Status != domain.OrderLineStatusFulfilled {
			t.Errorf("status = %q, want %q", updated.Status, domain.OrderLineStatusFulfilled)
		}
	}

	func TestOrderService_UpdateOrderLineStatus_PartialFlow(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := NewOrderService(repo)

		order, _ := svc.CreateOrder(ctx, CreateOrderInput{
			OrderType:   domain.OrderTypeOutbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 20}},
			CreatedBy:   "user",
		})
		line := &order.Lines[0]

		// pending → allocated → partial → fulfilled
		svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusAllocated})
		updated, err := svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusPartial})
		if err != nil {
			t.Fatalf("allocated → partial failed: %v", err)
		}
		if updated.Status != domain.OrderLineStatusPartial {
			t.Errorf("status = %q, want %q", updated.Status, domain.OrderLineStatusPartial)
		}

		updated, err = svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusFulfilled})
		if err != nil {
			t.Fatalf("partial → fulfilled failed: %v", err)
		}
		if updated.Status != domain.OrderLineStatusFulfilled {
			t.Errorf("status = %q, want %q", updated.Status, domain.OrderLineStatusFulfilled)
		}
	}

	func TestOrderService_UpdateOrderLineStatus_CancelFromAny(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := NewOrderService(repo)

		order, _ := svc.CreateOrder(ctx, CreateOrderInput{
			OrderType:   domain.OrderTypeOutbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
			CreatedBy:   "user",
		})
		line := &order.Lines[0]

		// pending → cancelled
		updated, err := svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusCancelled})
		if err != nil {
			t.Fatalf("pending → cancelled failed: %v", err)
		}
		if updated.Status != domain.OrderLineStatusCancelled {
			t.Errorf("status = %q, want %q", updated.Status, domain.OrderLineStatusCancelled)
		}

		// Allocated line can also be cancelled.
		order2, _ := svc.CreateOrder(ctx, CreateOrderInput{
			OrderType:   domain.OrderTypeInbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 5}},
			CreatedBy:   "user",
		})
		line2 := &order2.Lines[0]
		svc.UpdateOrderLineStatus(ctx, line2.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusAllocated})

		updated, err = svc.UpdateOrderLineStatus(ctx, line2.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusCancelled})
		if err != nil {
			t.Fatalf("allocated → cancelled failed: %v", err)
		}
		if updated.Status != domain.OrderLineStatusCancelled {
			t.Errorf("status = %q, want %q", updated.Status, domain.OrderLineStatusCancelled)
		}
	}

	func TestOrderService_UpdateOrderLineStatus_InvalidTransitions(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := NewOrderService(repo)

		order, _ := svc.CreateOrder(ctx, CreateOrderInput{
			OrderType:   domain.OrderTypeOutbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
			CreatedBy:   "user",
		})
		line := &order.Lines[0]

		// pending → fulfilled (skip allocated)
		_, err := svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusFulfilled})
		if err == nil {
			t.Fatal("expected error for pending → fulfilled transition")
		}

		// Advance to allocated, then try going backwards.
		svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusAllocated})
		_, err = svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusPending})
		if err == nil {
			t.Fatal("expected error for allocated → pending (backwards) transition")
		}
	}

	func TestOrderService_UpdateOrderLineStatus_TerminalStates(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := NewOrderService(repo)

		order, _ := svc.CreateOrder(ctx, CreateOrderInput{
			OrderType:   domain.OrderTypeOutbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
			CreatedBy:   "user",
		})
		line := &order.Lines[0]

		// Fulfilled is terminal.
		svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusAllocated})
		svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusFulfilled})

		_, err := svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusPartial})
		if err == nil {
			t.Fatal("expected error for fulfilled → partial transition")
		}

		// Cancelled is terminal.
		order2, _ := svc.CreateOrder(ctx, CreateOrderInput{
			OrderType:   domain.OrderTypeInbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 5}},
			CreatedBy:   "user",
		})
		line2 := &order2.Lines[0]
		svc.UpdateOrderLineStatus(ctx, line2.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusCancelled})

		_, err = svc.UpdateOrderLineStatus(ctx, line2.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusPending})
		if err == nil {
			t.Fatal("expected error for cancelled → pending transition")
		}
	}

	func TestOrderService_UpdateOrderLineStatus_NotFound(t *testing.T) {
		ctx := context.Background()
		svc := NewOrderService(newMockOrderRepo())

		_, err := svc.UpdateOrderLineStatus(ctx, uuid.New(), UpdateOrderLineStatusInput{Status: domain.OrderLineStatusAllocated})
		if err == nil {
			t.Fatal("expected error for non-existent order line")
		}
	}

	func TestOrderService_UpdateOrderLineStatus_Validation(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := NewOrderService(repo)

		order, _ := svc.CreateOrder(ctx, CreateOrderInput{
			OrderType:   domain.OrderTypeOutbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
			CreatedBy:   "user",
		})
		line := &order.Lines[0]

		// Invalid status value.
		_, err := svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: "invalid_status"})
		if err == nil {
			t.Fatal("expected validation error for invalid status")
		}

		// Empty status.
		_, err = svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: ""})
		if err == nil {
			t.Fatal("expected validation error for empty status")
		}
	}

	// ── UpdateASNStatus Tests ─────────────────────────────────────────────────────

	func TestOrderService_UpdateASNStatus_ValidTransitions(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := NewOrderService(repo)

		asn := setupMockASN(repo)

		// pending → arrived
		updated, err := svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})
		if err != nil {
			t.Fatalf("pending → arrived failed: %v", err)
		}
		if updated.Status != domain.ASNStatusArrived {
			t.Errorf("status = %q, want %q", updated.Status, domain.ASNStatusArrived)
		}

		// arrived → receiving
		updated, err = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusReceiving})
		if err != nil {
			t.Fatalf("arrived → receiving failed: %v", err)
		}
		if updated.Status != domain.ASNStatusReceiving {
			t.Errorf("status = %q, want %q", updated.Status, domain.ASNStatusReceiving)
		}

		// receiving → received
		updated, err = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusReceived})
		if err != nil {
			t.Fatalf("receiving → received failed: %v", err)
		}
		if updated.Status != domain.ASNStatusReceived {
			t.Errorf("status = %q, want %q", updated.Status, domain.ASNStatusReceived)
		}
	}

	func TestOrderService_UpdateASNStatus_PartialFlow(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := NewOrderService(repo)

		asn := setupMockASN(repo)

		// pending → arrived → receiving → partial → received
		svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})
		svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusReceiving})

		updated, err := svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusPartial})
		if err != nil {
			t.Fatalf("receiving → partial failed: %v", err)
		}
		if updated.Status != domain.ASNStatusPartial {
			t.Errorf("status = %q, want %q", updated.Status, domain.ASNStatusPartial)
		}

		updated, err = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusReceived})
		if err != nil {
			t.Fatalf("partial → received failed: %v", err)
		}
		if updated.Status != domain.ASNStatusReceived {
			t.Errorf("status = %q, want %q", updated.Status, domain.ASNStatusReceived)
		}
	}

	func TestOrderService_UpdateASNStatus_InvalidTransitions(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := NewOrderService(repo)

		asn := setupMockASN(repo)

		// pending → receiving (skip arrived)
		_, err := svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusReceiving})
		if err == nil {
			t.Fatal("expected error for pending → receiving transition")
		}

		// pending → received (skip ahead)
		_, err = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusReceived})
		if err == nil {
			t.Fatal("expected error for pending → received transition")
		}

		// Advance to receiving, then try backwards.
		svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})
		svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusReceiving})

		_, err = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusPending})
		if err == nil {
			t.Fatal("expected error for receiving → pending (backwards) transition")
		}

		_, err = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})
		if err == nil {
			t.Fatal("expected error for receiving → arrived (backwards) transition")
		}
	}

	func TestOrderService_UpdateASNStatus_TerminalStates(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := NewOrderService(repo)

		asn := setupMockASN(repo)

		// Advance to received (terminal).
		svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})
		svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusReceiving})
		svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusReceived})

		// Received → anything should fail.
		for _, target := range []domain.ASNStatus{
			domain.ASNStatusPending, domain.ASNStatusArrived,
			domain.ASNStatusReceiving, domain.ASNStatusPartial,
		} {
			_, err := svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: target})
			if err == nil {
				t.Errorf("expected error for received → %s transition", target)
			}
		}
	}

	func TestOrderService_UpdateASNStatus_NotFound(t *testing.T) {
		ctx := context.Background()
		svc := NewOrderService(newMockOrderRepo())

		_, err := svc.UpdateASNStatus(ctx, uuid.New(), UpdateASNStatusInput{Status: domain.ASNStatusArrived})
		if err == nil {
			t.Fatal("expected error for non-existent ASN")
		}
	}

	func TestOrderService_UpdateASNStatus_Validation(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := NewOrderService(repo)

		asn := setupMockASN(repo)

		// Invalid status value.
		_, err := svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: "invalid_status"})
		if err == nil {
			t.Fatal("expected validation error for invalid status")
		}

		// Empty status.
		_, err = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: ""})
		if err == nil {
			t.Fatal("expected validation error for empty status")
		}
	}

	func TestOrderService_UpdateASNStatus_LoadsLines(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := NewOrderService(repo)

		asn := setupMockASNWithLines(repo, 2)

		updated, err := svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})
		if err != nil {
			t.Fatalf("pending → arrived failed: %v", err)
		}
		if len(updated.Lines) != 2 {
			t.Errorf("expected 2 ASN lines, got %d", len(updated.Lines))
		}
	}

	// ── Mock setup helpers ────────────────────────────────────────────────────────

	func setupMockASN(repo *mockOrderRepo) *domain.ASN {
		asn := &domain.ASN{
			ASNNo:       "ASN-20260721-001",
			WarehouseID: uuid.New(),
			Status:      domain.ASNStatusPending,
		}
		repo.CreateASN(context.Background(), asn)
		return asn
	}

	func setupMockASNWithLines(repo *mockOrderRepo, lineCount int) *domain.ASN {
		asn := &domain.ASN{
			ASNNo:       "ASN-20260721-002",
			WarehouseID: uuid.New(),
			Status:      domain.ASNStatusPending,
		}
		repo.CreateASN(context.Background(), asn)
		for i := 0; i < lineCount; i++ {
			line := &domain.ASNLine{
				ASNID:       asn.ID,
				SKUID:       uuid.New(),
				ExpectedQty: float64((i + 1) * 10),
				Status:      domain.ASNLineStatusPending,
			}
			repo.CreateASNLine(context.Background(), line)
		}
		return asn
	}
