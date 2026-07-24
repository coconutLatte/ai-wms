package service

import (
	"context"
	"strings"
	"testing"
	"time"

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

func (m *mockOrderRepo) CountOrdersByStatus(ctx context.Context) (map[domain.OrderStatus]int, error) {
	result := make(map[domain.OrderStatus]int)
	for _, o := range m.orders {
		result[o.Status]++
	}
	return result, nil
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

func (m *mockOrderRepo) ListASNs(ctx context.Context, filter repository.ASNFilter) ([]*domain.ASN, error) {
	var result []*domain.ASN
	for _, a := range m.asns {
		if filter.WarehouseID != uuid.Nil && a.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.Status != "" && a.Status != filter.Status {
			continue
		}
		result = append(result, a)
	}
	return result, nil
}

func (m *mockOrderRepo) CountASNs(ctx context.Context, filter repository.ASNFilter) (int, error) {
	count := 0
	for _, a := range m.asns {
		if filter.WarehouseID != uuid.Nil && a.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.Status != "" && a.Status != filter.Status {
			continue
		}
		count++
	}
	return count, nil
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

func (m *mockOrderRepo) GetASNLine(ctx context.Context, id uuid.UUID) (*domain.ASNLine, error) {
	for _, lines := range m.asnLines {
		for _, l := range lines {
			if l.ID == id {
				return l, nil
			}
		}
	}
	return nil, pkgerrors.NewNotFound("asn line", id.String())
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

// mockTaskRepoForOrder implements repository.TaskRepository for OrderService testing.
type mockTaskRepoForOrder struct {
	tasks map[uuid.UUID]*domain.Task
}

func newMockTaskRepoForOrder() *mockTaskRepoForOrder {
	return &mockTaskRepoForOrder{
		tasks: make(map[uuid.UUID]*domain.Task),
	}
}

func (m *mockTaskRepoForOrder) CreateTask(ctx context.Context, t *domain.Task) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	m.tasks[t.ID] = t
	return nil
}
func (m *mockTaskRepoForOrder) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("task", id.String())
	}
	return t, nil
}
func (m *mockTaskRepoForOrder) GetTasksByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.Task, error) {
	var result []*domain.Task
	for _, t := range m.tasks {
		if t.OrderID != nil && *t.OrderID == orderID {
			result = append(result, t)
		}
	}
	return result, nil
}
func (m *mockTaskRepoForOrder) ListTasks(ctx context.Context, filter repository.TaskFilter) ([]*domain.Task, error) {
	return nil, nil
}
func (m *mockTaskRepoForOrder) AssignTask(ctx context.Context, id uuid.UUID, assignedTo string) error {
	return nil
}
func (m *mockTaskRepoForOrder) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status domain.TaskStatus) error {
	t, ok := m.tasks[id]
	if !ok {
		return pkgerrors.NewNotFound("task", id.String())
	}
	t.Status = status
	return nil
}
func (m *mockTaskRepoForOrder) CompleteTask(ctx context.Context, id uuid.UUID, actualQty float64, toLocationID *uuid.UUID) error {
	return nil
}
func (m *mockTaskRepoForOrder) CountTasks(ctx context.Context, filter repository.TaskFilter) (int, error) {
	return 0, nil
}
func (m *mockTaskRepoForOrder) CountTasksByStatus(ctx context.Context) (map[domain.TaskStatus]int, error) {
	return nil, nil
}
func (m *mockTaskRepoForOrder) CreateWave(ctx context.Context, w *domain.Wave) error {
	return nil
}
func (m *mockTaskRepoForOrder) GetWave(ctx context.Context, id uuid.UUID) (*domain.Wave, error) {
	return nil, nil
}
func (m *mockTaskRepoForOrder) ListWaves(ctx context.Context, filter repository.WaveFilter) ([]*domain.Wave, error) {
	return nil, nil
}
func (m *mockTaskRepoForOrder) UpdateWaveStatus(ctx context.Context, id uuid.UUID, status domain.WaveStatus) error {
	return nil
}
func (m *mockTaskRepoForOrder) AddWaveOrders(ctx context.Context, id uuid.UUID, orderIDs []uuid.UUID) error {
	return nil
}
func (m *mockTaskRepoForOrder) RemoveWaveOrders(ctx context.Context, id uuid.UUID, orderIDs []uuid.UUID) error {
	return nil
}
func (m *mockTaskRepoForOrder) CountWaves(ctx context.Context, filter repository.WaveFilter) (int, error) {
	return 0, nil
}

// newMockOrderService creates an OrderService backed by mock repositories.
func newMockOrderService() (*OrderService, *mockOrderRepo, *mockTaskRepoForOrder) {
	orderRepo := newMockOrderRepo()
	taskRepo := newMockTaskRepoForOrder()
	return NewOrderService(orderRepo, taskRepo), orderRepo, taskRepo
}

// newMockOrderServiceWithRepos creates an OrderService with pre-existing repos.
func newMockOrderServiceWithRepos(orderRepo *mockOrderRepo, taskRepo *mockTaskRepoForOrder) *OrderService {
	return NewOrderService(orderRepo, taskRepo)
}

// ── Tests ───────────────────────────────────────────────────────────────────

func TestOrderService_CreateOrder_Inbound(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newMockOrderService()

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
	svc, _, _ := newMockOrderService()

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
	svc, _, _ := newMockOrderService()

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
	svc, _, _ := newMockOrderService()

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
	svc, _, _ := newMockOrderService()

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
	svc, _, _ := newMockOrderService()

	_, err := svc.GetOrder(ctx, uuid.New())
	if err == nil {
		t.Fatal("expected error for unknown order")
	}
}

func TestOrderService_GetOrderByNo(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newMockOrderService()

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
	svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

	wh1 := uuid.New()
	wh2 := uuid.New()

	// Create orders in different warehouses with different types.
_, _ = svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: wh1,
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "user",
	})
_, _ = svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: wh1,
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 20}},
		CreatedBy:   "user",
	})
_, _ = svc.CreateOrder(ctx, CreateOrderInput{
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
	svc, _, _ := newMockOrderService()

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
	svc, _, _ := newMockOrderService()

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
	svc, _, _ := newMockOrderService()

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
_, _ = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
	_, err = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusCompleted})
	if err == nil {
		t.Fatal("expected error for confirmed → completed transition")
	}
}

func TestOrderService_UpdateOrderStatus_TerminalStates(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newMockOrderService()

	// Create and cancel an order.
	order1, _ := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "user",
	})
_, _ = svc.UpdateOrderStatus(ctx, order1.ID, UpdateOrderStatusInput{Status: domain.OrderStatusCancelled})

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
_, _ = svc.UpdateOrderStatus(ctx, order2.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
_, _ = svc.UpdateOrderStatus(ctx, order2.ID, UpdateOrderStatusInput{Status: domain.OrderStatusProcessing})
_, _ = svc.UpdateOrderStatus(ctx, order2.ID, UpdateOrderStatusInput{Status: domain.OrderStatusCompleted})

	// Completed → something else (should fail).
	_, err = svc.UpdateOrderStatus(ctx, order2.ID, UpdateOrderStatusInput{Status: domain.OrderStatusCancelled})
	if err == nil {
		t.Fatal("expected error for completed → cancelled transition")
	}
}

func TestOrderService_AddOrderLine(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newMockOrderService()

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
	svc, _, _ := newMockOrderService()

	order, _ := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "user",
	})

	// Confirm the order first.
_, _ = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})

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
	svc, _, _ := newMockOrderService()

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
	svc, _, _ := newMockOrderService()

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
	svc, _, _ := newMockOrderService()

	order, _ := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "user",
	})

	// Advance to processing.
_, _ = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
_, _ = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusProcessing})

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
		svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

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
		svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

		order, _ := svc.CreateOrder(ctx, CreateOrderInput{
			OrderType:   domain.OrderTypeOutbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 20}},
			CreatedBy:   "user",
		})
		line := &order.Lines[0]

		// pending → allocated → partial → fulfilled
	_, _ = svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusAllocated})
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
		svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

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
	_, _ = svc.UpdateOrderLineStatus(ctx, line2.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusAllocated})

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
		svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

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
	_, _ = svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusAllocated})
		_, err = svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusPending})
		if err == nil {
			t.Fatal("expected error for allocated → pending (backwards) transition")
		}
	}

	func TestOrderService_UpdateOrderLineStatus_TerminalStates(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

		order, _ := svc.CreateOrder(ctx, CreateOrderInput{
			OrderType:   domain.OrderTypeOutbound,
			WarehouseID: uuid.New(),
			Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
			CreatedBy:   "user",
		})
		line := &order.Lines[0]

		// Fulfilled is terminal.
	_, _ = svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusAllocated})
	_, _ = svc.UpdateOrderLineStatus(ctx, line.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusFulfilled})

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
	_, _ = svc.UpdateOrderLineStatus(ctx, line2.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusCancelled})

		_, err = svc.UpdateOrderLineStatus(ctx, line2.ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusPending})
		if err == nil {
			t.Fatal("expected error for cancelled → pending transition")
		}
	}

	func TestOrderService_UpdateOrderLineStatus_NotFound(t *testing.T) {
		ctx := context.Background()
		svc, _, _ := newMockOrderService()

		_, err := svc.UpdateOrderLineStatus(ctx, uuid.New(), UpdateOrderLineStatusInput{Status: domain.OrderLineStatusAllocated})
		if err == nil {
			t.Fatal("expected error for non-existent order line")
		}
	}

	func TestOrderService_UpdateOrderLineStatus_Validation(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

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
		svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

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
		svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

		asn := setupMockASN(repo)

		// pending → arrived → receiving → partial → received
	_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})
	_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusReceiving})

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
		svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

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
	_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})
	_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusReceiving})

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
		svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

		asn := setupMockASN(repo)

		// Advance to received (terminal).
	_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})
	_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusReceiving})
	_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusReceived})

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
		svc, _, _ := newMockOrderService()

		_, err := svc.UpdateASNStatus(ctx, uuid.New(), UpdateASNStatusInput{Status: domain.ASNStatusArrived})
		if err == nil {
			t.Fatal("expected error for non-existent ASN")
		}
	}

	func TestOrderService_UpdateASNStatus_Validation(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

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
		svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

		asn := setupMockASNWithLines(repo, 2)

		updated, err := svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})
		if err != nil {
			t.Fatalf("pending → arrived failed: %v", err)
		}
		if len(updated.Lines) != 2 {
			t.Errorf("expected 2 ASN lines, got %d", len(updated.Lines))
		}
	}

	// ── CreateASN Tests ──────────────────────────────────────────────────────────────

	func TestOrderService_CreateASN(t *testing.T) {
		ctx := context.Background()
		svc, _, _ := newMockOrderService()

		skuID := uuid.New()
		whID := uuid.New()
		expectedAt := time.Now().Add(48 * time.Hour).Truncate(time.Second)

		asn, err := svc.CreateASN(ctx, CreateASNInput{
			WarehouseID: whID,
			Carrier:     "UPS",
			TrackingNo:  "1Z999AA10123456784",
			ExpectedAt:  expectedAt,
			Lines: []CreateASNLineInput{
				{SKUID: skuID, ExpectedQty: 50},
				{SKUID: uuid.New(), ExpectedQty: 30, BatchNo: "BATCH-001"},
			},
		})
		if err != nil {
			t.Fatalf("CreateASN failed: %v", err)
		}
		if asn.Status != domain.ASNStatusPending {
			t.Errorf("status = %q, want %q", asn.Status, domain.ASNStatusPending)
		}
		if !strings.HasPrefix(asn.ASNNo, "ASN-") {
			t.Errorf("asn_no should start with ASN-: got %q", asn.ASNNo)
		}
		if asn.Carrier != "UPS" {
			t.Errorf("carrier = %q, want UPS", asn.Carrier)
		}
		if asn.TrackingNo != "1Z999AA10123456784" {
			t.Errorf("tracking_no = %q, want 1Z999AA10123456784", asn.TrackingNo)
		}
		if len(asn.Lines) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(asn.Lines))
		}
		if asn.Lines[0].SKUID != skuID {
			t.Errorf("line[0] sku_id = %q, want %q", asn.Lines[0].SKUID, skuID)
		}
		if asn.Lines[0].ExpectedQty != 50 {
			t.Errorf("line[0] expected_qty = %f, want 50", asn.Lines[0].ExpectedQty)
		}
		if asn.Lines[1].BatchNo != "BATCH-001" {
			t.Errorf("line[1] batch_no = %q, want BATCH-001", asn.Lines[1].BatchNo)
		}
	}

	func TestOrderService_CreateASN_CustomNo(t *testing.T) {
		ctx := context.Background()
		svc, _, _ := newMockOrderService()

		asn, err := svc.CreateASN(ctx, CreateASNInput{
			ASNNo:       "ASN-CUSTOM-001",
			WarehouseID: uuid.New(),
			ExpectedAt:  time.Now().Add(24 * time.Hour),
			Lines:       []CreateASNLineInput{{SKUID: uuid.New(), ExpectedQty: 10}},
		})
		if err != nil {
			t.Fatalf("CreateASN failed: %v", err)
		}
		if asn.ASNNo != "ASN-CUSTOM-001" {
			t.Errorf("asn_no = %q, want ASN-CUSTOM-001", asn.ASNNo)
		}
	}

	func TestOrderService_CreateASN_WithOrderLink(t *testing.T) {
		ctx := context.Background()
		svc, _, _ := newMockOrderService()

		orderID := uuid.New()
		asn, err := svc.CreateASN(ctx, CreateASNInput{
			WarehouseID: uuid.New(),
			OrderID:     orderID,
			ExpectedAt:  time.Now().Add(24 * time.Hour),
			Lines:       []CreateASNLineInput{{SKUID: uuid.New(), ExpectedQty: 10}},
		})
		if err != nil {
			t.Fatalf("CreateASN failed: %v", err)
		}
		if asn.OrderID != orderID {
			t.Errorf("order_id = %q, want %q", asn.OrderID, orderID)
		}
	}

	func TestOrderService_CreateASN_ValidationErrors(t *testing.T) {
		ctx := context.Background()
		svc, _, _ := newMockOrderService()

		tests := []struct {
			name  string
			input CreateASNInput
		}{
			{"nil warehouse id", CreateASNInput{
				WarehouseID: uuid.Nil,
				ExpectedAt:  time.Now(),
				Lines:       []CreateASNLineInput{{SKUID: uuid.New(), ExpectedQty: 1}},
			}},
			{"zero expected_at", CreateASNInput{
				WarehouseID: uuid.New(),
				ExpectedAt:  time.Time{},
				Lines:       []CreateASNLineInput{{SKUID: uuid.New(), ExpectedQty: 1}},
			}},
			{"no lines", CreateASNInput{
				WarehouseID: uuid.New(),
				ExpectedAt:  time.Now(),
				Lines:       nil,
			}},
			{"empty lines", CreateASNInput{
				WarehouseID: uuid.New(),
				ExpectedAt:  time.Now(),
				Lines:       []CreateASNLineInput{},
			}},
			{"zero qty line", CreateASNInput{
				WarehouseID: uuid.New(),
				ExpectedAt:  time.Now(),
				Lines:       []CreateASNLineInput{{SKUID: uuid.New(), ExpectedQty: 0}},
			}},
			{"negative qty line", CreateASNInput{
				WarehouseID: uuid.New(),
				ExpectedAt:  time.Now(),
				Lines:       []CreateASNLineInput{{SKUID: uuid.New(), ExpectedQty: -5}},
			}},
			{"nil sku id line", CreateASNInput{
				WarehouseID: uuid.New(),
				ExpectedAt:  time.Now(),
				Lines:       []CreateASNLineInput{{SKUID: uuid.Nil, ExpectedQty: 1}},
			}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := svc.CreateASN(ctx, tt.input)
				if err == nil {
					t.Fatal("expected error")
				}
			})
		}
	}

	// ── GetASN Tests ────────────────────────────────────────────────────────────────

	func TestOrderService_GetASN(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

		asn := setupMockASNWithLines(repo, 3)

		got, err := svc.GetASN(ctx, asn.ID)
		if err != nil {
			t.Fatalf("GetASN failed: %v", err)
		}
		if got.ASNNo != asn.ASNNo {
			t.Errorf("asn_no = %q, want %q", got.ASNNo, asn.ASNNo)
		}
		if len(got.Lines) != 3 {
			t.Errorf("expected 3 lines, got %d", len(got.Lines))
		}
	}

	func TestOrderService_GetASN_NotFound(t *testing.T) {
		ctx := context.Background()
		svc, _, _ := newMockOrderService()

		_, err := svc.GetASN(ctx, uuid.New())
		if err == nil {
			t.Fatal("expected error for unknown ASN")
		}
	}

	// ── ListASNs Tests ──────────────────────────────────────────────────────────────

	func TestOrderService_ListASNs(t *testing.T) {
		ctx := context.Background()
		repo := newMockOrderRepo()
		svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

		wh1 := uuid.New()
		wh2 := uuid.New()

		// Create ASNs in different warehouses with different statuses.
	_ = repo.CreateASN(ctx, &domain.ASN{ASNNo: "ASN-001", WarehouseID: wh1, Status: domain.ASNStatusPending})
	_ = repo.CreateASN(ctx, &domain.ASN{ASNNo: "ASN-002", WarehouseID: wh1, Status: domain.ASNStatusArrived})
	_ = repo.CreateASN(ctx, &domain.ASN{ASNNo: "ASN-003", WarehouseID: wh2, Status: domain.ASNStatusPending})

		// All ASNs.
		all, count, err := svc.ListASNs(ctx, repository.ASNFilter{})
		if err != nil {
			t.Fatalf("ListASNs failed: %v", err)
		}
		if len(all) != 3 {
			t.Errorf("expected 3 asns, got %d", len(all))
		}
		if count != 3 {
			t.Errorf("count = %d, want 3", count)
		}

		// Filter by warehouse.
		wh1ASNs, wh1Count, err := svc.ListASNs(ctx, repository.ASNFilter{WarehouseID: wh1})
		if err != nil {
			t.Fatalf("ListASNs wh1 failed: %v", err)
		}
		if len(wh1ASNs) != 2 {
			t.Errorf("expected 2 asns in wh1, got %d", len(wh1ASNs))
		}
		if wh1Count != 2 {
			t.Errorf("wh1 count = %d, want 2", wh1Count)
		}

		// Filter by status.
		pending, pendingCount, err := svc.ListASNs(ctx, repository.ASNFilter{Status: domain.ASNStatusPending})
		if err != nil {
			t.Fatalf("ListASNs pending failed: %v", err)
		}
		if len(pending) != 2 {
			t.Errorf("expected 2 pending asns, got %d", len(pending))
		}
		if pendingCount != 2 {
			t.Errorf("pending count = %d, want 2", pendingCount)
		}
	}

	// ── Mock setup helpers ────────────────────────────────────────────────────────

	func setupMockASN(repo *mockOrderRepo) *domain.ASN {
		asn := &domain.ASN{
			ASNNo:       "ASN-20260721-001",
			WarehouseID: uuid.New(),
			Status:      domain.ASNStatusPending,
		}
	_ = repo.CreateASN(context.Background(), asn)
		return asn
	}

	func setupMockASNWithLines(repo *mockOrderRepo, lineCount int) *domain.ASN {
		asn := &domain.ASN{
			ASNNo:       "ASN-20260721-002",
			WarehouseID: uuid.New(),
			Status:      domain.ASNStatusPending,
		}
	_ = repo.CreateASN(context.Background(), asn)
		for i := 0; i < lineCount; i++ {
			line := &domain.ASNLine{
				ASNID:       asn.ID,
				SKUID:       uuid.New(),
				ExpectedQty: float64((i + 1) * 10),
				Status:      domain.ASNLineStatusPending,
			}
		_ = repo.CreateASNLine(context.Background(), line)
		}
		return asn
	}

// ── Order→Task Generation Tests ──────────────────────────────────────────────────

func TestOrderService_ConfirmInbound_GeneratesPutawayTasks(t *testing.T) {
	ctx := context.Background()
	svc, _, taskRepo := newMockOrderService()

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: uuid.New(),
		Lines: []CreateOrderLineInput{
			{SKUID: uuid.New(), OrderedQty: 100, UOM: "CS"},
			{SKUID: uuid.New(), OrderedQty: 50, UOM: "EA"},
		},
		CreatedBy: "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	updated, err := svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
	if err != nil {
		t.Fatalf("draft → confirmed failed: %v", err)
	}
	if updated.Status != domain.OrderStatusConfirmed {
		t.Errorf("status = %q, want %q", updated.Status, domain.OrderStatusConfirmed)
	}

	tasks, _ := taskRepo.GetTasksByOrderID(ctx, order.ID)
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	for i, task := range tasks {
		if task.TaskType != domain.TaskTypePutaway {
			t.Errorf("task[%d] type = %q, want putaway", i, task.TaskType)
		}
		if task.OrderID == nil || *task.OrderID != order.ID {
			t.Errorf("task[%d] order_id not set correctly", i)
		}
		if task.WarehouseID != order.WarehouseID {
			t.Errorf("task[%d] warehouse_id mismatch", i)
		}
		if task.Status != domain.TaskStatusPending {
			t.Errorf("task[%d] status = %q, want pending", i, task.Status)
		}
		if task.ExpectedQty <= 0 {
			t.Errorf("task[%d] expected_qty = %f, want > 0", i, task.ExpectedQty)
		}
		if task.Instructions == "" {
			t.Errorf("task[%d] instructions empty", i)
		}
	}
}

func TestOrderService_ConfirmOutbound_GeneratesPickTasks(t *testing.T) {
	ctx := context.Background()
	svc, _, taskRepo := newMockOrderService()

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Priority:    domain.OrderPriorityHigh,
		Lines: []CreateOrderLineInput{
			{SKUID: uuid.New(), OrderedQty: 30},
		},
		CreatedBy: "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	updated, err := svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
	if err != nil {
		t.Fatalf("draft → confirmed failed: %v", err)
	}
	if updated.Status != domain.OrderStatusConfirmed {
		t.Errorf("status = %q, want %q", updated.Status, domain.OrderStatusConfirmed)
	}

	tasks, _ := taskRepo.GetTasksByOrderID(ctx, order.ID)
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.TaskType != domain.TaskTypePick {
		t.Errorf("task type = %q, want pick", task.TaskType)
	}
	if task.Priority != domain.TaskPriorityHigh {
		t.Errorf("task priority = %q, want high", task.Priority)
	}
	if !strings.Contains(task.Instructions, "Pick") {
		t.Errorf("instructions should mention 'Pick': %q", task.Instructions)
	}
}

func TestOrderService_Confirm_TaskFieldsMatchOrder(t *testing.T) {
	ctx := context.Background()
	svc, _, taskRepo := newMockOrderService()

	skuID := uuid.New()
	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Priority:    domain.OrderPriorityUrgent,
		Lines: []CreateOrderLineInput{
			{SKUID: skuID, OrderedQty: 75, UOM: "BX", BatchNo: "BATCH-A"},
		},
		CreatedBy: "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	_, err = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
	if err != nil {
		t.Fatalf("confirm failed: %v", err)
	}

	tasks, _ := taskRepo.GetTasksByOrderID(ctx, order.ID)
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	line := order.Lines[0]
	task := tasks[0]

	if task.WarehouseID != order.WarehouseID {
		t.Errorf("warehouse_id mismatch")
	}
	if task.Priority != domain.TaskPriorityUrgent {
		t.Errorf("priority = %q, want urgent", task.Priority)
	}
	if task.SKUID != skuID {
		t.Errorf("sku_id mismatch")
	}
	if task.ExpectedQty != line.OrderedQty {
		t.Errorf("expected_qty = %f, want %f", task.ExpectedQty, line.OrderedQty)
	}
	if task.UOM != "BX" {
		t.Errorf("uom = %q, want BX", task.UOM)
	}
	if task.BatchNo != "BATCH-A" {
		t.Errorf("batch_no = %q, want BATCH-A", task.BatchNo)
	}
	if task.Status != domain.TaskStatusPending {
		t.Errorf("status = %q, want pending", task.Status)
	}
	if task.OrderID == nil || *task.OrderID != order.ID {
		t.Errorf("order_id not set correctly")
	}
	if task.OrderLineID == nil || *task.OrderLineID != line.ID {
		t.Errorf("order_line_id not set correctly")
	}
}

func TestOrderService_ConfirmTransfer_GeneratesPutawayTasks(t *testing.T) {
	ctx := context.Background()
	svc, _, taskRepo := newMockOrderService()

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeTransfer,
		WarehouseID: uuid.New(),
		Lines: []CreateOrderLineInput{
			{SKUID: uuid.New(), OrderedQty: 200},
		},
		CreatedBy: "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	_, err = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
	if err != nil {
		t.Fatalf("confirm failed: %v", err)
	}

	tasks, _ := taskRepo.GetTasksByOrderID(ctx, order.ID)
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].TaskType != domain.TaskTypePutaway {
		t.Errorf("task type = %q, want putaway", tasks[0].TaskType)
	}
}

func TestOrderService_Confirm_NoTasksForNonConfirmTransition(t *testing.T) {
	ctx := context.Background()
	svc, _, taskRepo := newMockOrderService()

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Cancel the order — no tasks should be generated.
	_, err = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusCancelled})
	if err != nil {
		t.Fatalf("draft → cancelled failed: %v", err)
	}

	tasks, _ := taskRepo.GetTasksByOrderID(ctx, order.ID)
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks for cancelled order, got %d", len(tasks))
	}
}

// ── ReceiveASNLine Tests ─────────────────────────────────────────────────────

func TestOrderService_ReceiveASNLine_FullReceive(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

	asn := setupMockASNWithLines(repo, 1)
	lines, _ := repo.GetASNLines(ctx, asn.ID)
	line := lines[0]
	line.ExpectedQty = 100
	line.ReceivedQty = 0
	line.Status = domain.ASNLineStatusPending

	// Transition ASN to arrived.
_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})

	// Receive the full quantity.
	updated, err := svc.ReceiveASNLine(ctx, asn.ID, line.ID, ReceiveASNLineInput{ReceivedQty: 100})
	if err != nil {
		t.Fatalf("ReceiveASNLine failed: %v", err)
	}

	// ASN should be "received" (arrived → receiving → received).
	if updated.Status != domain.ASNStatusReceived {
		t.Errorf("ASN status = %q, want %q", updated.Status, domain.ASNStatusReceived)
	}
	// Line should be "received".
	if len(updated.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(updated.Lines))
	}
	if updated.Lines[0].Status != domain.ASNLineStatusReceived {
		t.Errorf("line status = %q, want %q", updated.Lines[0].Status, domain.ASNLineStatusReceived)
	}
	if updated.Lines[0].ReceivedQty != 100 {
		t.Errorf("received_qty = %f, want 100", updated.Lines[0].ReceivedQty)
	}
}

func TestOrderService_ReceiveASNLine_PartialReceive(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

	asn := setupMockASNWithLines(repo, 2)
	lines, _ := repo.GetASNLines(ctx, asn.ID)
	line0 := lines[0]
	line0.ExpectedQty = 50
	line0.ReceivedQty = 0
	line1 := lines[1]
	line1.ExpectedQty = 30
	line1.ReceivedQty = 0

	// Transition ASN to arrived.
_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})

	// Receive partial quantity on line 0.
	updated, err := svc.ReceiveASNLine(ctx, asn.ID, line0.ID, ReceiveASNLineInput{ReceivedQty: 20})
	if err != nil {
		t.Fatalf("ReceiveASNLine failed: %v", err)
	}

	// ASN should be "receiving" (arrived → receiving, but not all lines done).
	if updated.Status != domain.ASNStatusReceiving {
		t.Errorf("ASN status = %q, want %q", updated.Status, domain.ASNStatusReceiving)
	}

	// Find the updated line 0.
	for _, l := range updated.Lines {
		if l.ID == line0.ID {
			if l.Status != domain.ASNLineStatusPartial {
				t.Errorf("line status = %q, want %q", l.Status, domain.ASNLineStatusPartial)
			}
			if l.ReceivedQty != 20 {
				t.Errorf("received_qty = %f, want 20", l.ReceivedQty)
			}
		}
	}
}

func TestOrderService_ReceiveASNLine_MultiLineAllReceived(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

	asn := setupMockASNWithLines(repo, 2)
	lines, _ := repo.GetASNLines(ctx, asn.ID)
	line0 := lines[0]
	line0.ExpectedQty = 10
	line0.ReceivedQty = 0
	line1 := lines[1]
	line1.ExpectedQty = 5
	line1.ReceivedQty = 0

	// Transition ASN to arrived.
_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})

	// Receive line 0 fully.
_, _ = svc.ReceiveASNLine(ctx, asn.ID, line0.ID, ReceiveASNLineInput{ReceivedQty: 10})

	// Receive line 1 fully.
	updated, err := svc.ReceiveASNLine(ctx, asn.ID, line1.ID, ReceiveASNLineInput{ReceivedQty: 5})
	if err != nil {
		t.Fatalf("ReceiveASNLine line1 failed: %v", err)
	}

	// ASN should be "received" (all lines fully received).
	if updated.Status != domain.ASNStatusReceived {
		t.Errorf("ASN status = %q, want %q", updated.Status, domain.ASNStatusReceived)
	}
}

func TestOrderService_ReceiveASNLine_MultiLinePartialASN(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

	asn := setupMockASNWithLines(repo, 2)
	lines, _ := repo.GetASNLines(ctx, asn.ID)
	line0 := lines[0]
	line0.ExpectedQty = 10
	line0.ReceivedQty = 0

	// Transition ASN to arrived.
_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})

	// Receive line 0 partially.
_, _ = svc.ReceiveASNLine(ctx, asn.ID, line0.ID, ReceiveASNLineInput{ReceivedQty: 5})

	// ASN should be "receiving" (received some but not all lines).
	updated, _ := svc.GetASN(ctx, asn.ID)
	if updated.Status != domain.ASNStatusReceiving {
		t.Errorf("ASN status = %q, want %q", updated.Status, domain.ASNStatusReceiving)
	}
}

func TestOrderService_ReceiveASNLine_ExceedsExpected(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

	asn := setupMockASNWithLines(repo, 1)
	lines, _ := repo.GetASNLines(ctx, asn.ID)
	line := lines[0]
	line.ExpectedQty = 10
	line.ReceivedQty = 0

_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})

	_, err := svc.ReceiveASNLine(ctx, asn.ID, line.ID, ReceiveASNLineInput{ReceivedQty: 15})
	if err == nil {
		t.Fatal("expected error for exceeding expected_qty")
	}
}

func TestOrderService_ReceiveASNLine_AccumulatedExceedsExpected(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

	asn := setupMockASNWithLines(repo, 1)
	lines, _ := repo.GetASNLines(ctx, asn.ID)
	line := lines[0]
	line.ExpectedQty = 10
	line.ReceivedQty = 0

_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})

	// Receive 8 first.
_, _ = svc.ReceiveASNLine(ctx, asn.ID, line.ID, ReceiveASNLineInput{ReceivedQty: 8})

	// Try to receive 5 more (total would be 13 > 10).
	_, err := svc.ReceiveASNLine(ctx, asn.ID, line.ID, ReceiveASNLineInput{ReceivedQty: 5})
	if err == nil {
		t.Fatal("expected error for accumulated qty exceeding expected_qty")
	}
}

func TestOrderService_ReceiveASNLine_ZeroQty(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

	asn := setupMockASNWithLines(repo, 1)
	lines, _ := repo.GetASNLines(ctx, asn.ID)
	line := lines[0]

_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})

	_, err := svc.ReceiveASNLine(ctx, asn.ID, line.ID, ReceiveASNLineInput{ReceivedQty: 0})
	if err == nil {
		t.Fatal("expected error for zero received_qty")
	}
}

func TestOrderService_ReceiveASNLine_NegativeQty(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

	asn := setupMockASNWithLines(repo, 1)
	lines, _ := repo.GetASNLines(ctx, asn.ID)
	line := lines[0]

_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})

	_, err := svc.ReceiveASNLine(ctx, asn.ID, line.ID, ReceiveASNLineInput{ReceivedQty: -1})
	if err == nil {
		t.Fatal("expected error for negative received_qty")
	}
}

func TestOrderService_ReceiveASNLine_WrongASN(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

	asn1 := setupMockASNWithLines(repo, 1)
	asn2 := setupMockASNWithLines(repo, 1)

	lines, _ := repo.GetASNLines(ctx, asn2.ID)
	line := lines[0]

_, _ = svc.UpdateASNStatus(ctx, asn1.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})

	// Try to receive line from asn2 using asn1's ID.
	_, err := svc.ReceiveASNLine(ctx, asn1.ID, line.ID, ReceiveASNLineInput{ReceivedQty: 5})
	if err == nil {
		t.Fatal("expected error for line not belonging to ASN")
	}
}

func TestOrderService_ReceiveASNLine_LineNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

	asn := setupMockASNWithLines(repo, 1)
_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})

	_, err := svc.ReceiveASNLine(ctx, asn.ID, uuid.New(), ReceiveASNLineInput{ReceivedQty: 5})
	if err == nil {
		t.Fatal("expected error for non-existent line")
	}
}

func TestOrderService_ReceiveASNLine_FromPendingNotArrived(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

	asn := setupMockASNWithLines(repo, 1)
	lines, _ := repo.GetASNLines(ctx, asn.ID)
	line := lines[0]
	line.ExpectedQty = 10
	line.ReceivedQty = 0

	// ASN is still pending — try to receive.
	updated, err := svc.ReceiveASNLine(ctx, asn.ID, line.ID, ReceiveASNLineInput{ReceivedQty: 5})
	if err != nil {
		t.Fatalf("ReceiveASNLine should not fail for pending ASN: %v", err)
	}
	// The ASN stays in pending (CanTransitionTo from pending only allows "arrived").
	if updated.Status != domain.ASNStatusPending {
		t.Errorf("ASN status = %q, want %q (pending ASN cannot auto-transition)", updated.Status, domain.ASNStatusPending)
	}
}

func TestOrderService_ReceiveASNLine_AlreadyReceiving(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

	asn := setupMockASNWithLines(repo, 2)
	lines, _ := repo.GetASNLines(ctx, asn.ID)
	line0 := lines[0]
	line0.ExpectedQty = 20
	line0.ReceivedQty = 0
	line1 := lines[1]
	line1.ExpectedQty = 10
	line1.ReceivedQty = 0

	// Transition to arrived, then start receiving.
_, _ = svc.UpdateASNStatus(ctx, asn.ID, UpdateASNStatusInput{Status: domain.ASNStatusArrived})
_, _ = svc.ReceiveASNLine(ctx, asn.ID, line0.ID, ReceiveASNLineInput{ReceivedQty: 10})

	// Now receive more on the same line. ASN should still be "receiving" (no new transition needed).
	updated, err := svc.ReceiveASNLine(ctx, asn.ID, line0.ID, ReceiveASNLineInput{ReceivedQty: 5})
	if err != nil {
		t.Fatalf("ReceiveASNLine second call failed: %v", err)
	}
	if updated.Status != domain.ASNStatusReceiving {
		t.Errorf("ASN status = %q, want %q", updated.Status, domain.ASNStatusReceiving)
	}
}

// ── CancelOrder Tests ──────────────────────────────────────────────────────────

func TestOrderService_CancelOrder_Draft(t *testing.T) {
	ctx := context.Background()
	svc, _, taskRepo := newMockOrderService()

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Create a task associated with the order.
	lineID := order.Lines[0].ID
_ = taskRepo.CreateTask(ctx, &domain.Task{
		TaskType:    domain.TaskTypePick,
		WarehouseID: order.WarehouseID,
		OrderID:     &order.ID,
		OrderLineID: &lineID,
		Status:      domain.TaskStatusPending,
		SKUID:       order.Lines[0].SKUID,
		ExpectedQty: 10,
	})

	cancelled, err := svc.CancelOrder(ctx, order.ID, CancelOrderInput{Reason: "no longer needed"})
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}
	if cancelled.Status != domain.OrderStatusCancelled {
		t.Errorf("status = %q, want %q", cancelled.Status, domain.OrderStatusCancelled)
	}

	// Verify tasks are cancelled.
	tasks, _ := taskRepo.GetTasksByOrderID(ctx, order.ID)
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Status != domain.TaskStatusCancelled {
		t.Errorf("task status = %q, want %q", tasks[0].Status, domain.TaskStatusCancelled)
	}

	// Verify order lines are cancelled.
	if len(cancelled.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(cancelled.Lines))
	}
	if cancelled.Lines[0].Status != domain.OrderLineStatusCancelled {
		t.Errorf("line status = %q, want %q", cancelled.Lines[0].Status, domain.OrderLineStatusCancelled)
	}
}

func TestOrderService_CancelOrder_Confirmed(t *testing.T) {
	ctx := context.Background()
	svc, _, taskRepo := newMockOrderService()

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeInbound,
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

	// Confirm the order (generates tasks).
_, _ = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})

	cancelled, err := svc.CancelOrder(ctx, order.ID, CancelOrderInput{})
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}
	if cancelled.Status != domain.OrderStatusCancelled {
		t.Errorf("status = %q, want %q", cancelled.Status, domain.OrderStatusCancelled)
	}

	// All generated tasks should be cancelled.
	tasks, _ := taskRepo.GetTasksByOrderID(ctx, order.ID)
	for _, tk := range tasks {
		if tk.Status != domain.TaskStatusCancelled {
			t.Errorf("task %s status = %q, want cancelled", tk.ID, tk.Status)
		}
	}

	// All lines should be cancelled.
	for _, l := range cancelled.Lines {
		if l.Status != domain.OrderLineStatusCancelled {
			t.Errorf("line %s status = %q, want cancelled", l.ID, l.Status)
		}
	}
}

func TestOrderService_CancelOrder_Processing(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newMockOrderService()

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 20}},
		CreatedBy:   "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Confirm → Processing.
_, _ = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
_, _ = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusProcessing})

	cancelled, err := svc.CancelOrder(ctx, order.ID, CancelOrderInput{})
	if err != nil {
		t.Fatalf("processing → cancelled failed: %v", err)
	}
	if cancelled.Status != domain.OrderStatusCancelled {
		t.Errorf("status = %q, want %q", cancelled.Status, domain.OrderStatusCancelled)
	}
}

func TestOrderService_CancelOrder_Partial(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newMockOrderService()

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 15}},
		CreatedBy:   "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

_, _ = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
_, _ = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusProcessing})
_, _ = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusPartial})

	cancelled, err := svc.CancelOrder(ctx, order.ID, CancelOrderInput{Reason: "order cancelled during partial fulfillment"})
	if err != nil {
		t.Fatalf("partial → cancelled failed: %v", err)
	}
	if cancelled.Status != domain.OrderStatusCancelled {
		t.Errorf("status = %q, want %q", cancelled.Status, domain.OrderStatusCancelled)
	}
}

func TestOrderService_CancelOrder_TerminalCompleted(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newMockOrderService()

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Complete the order.
_, _ = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusConfirmed})
_, _ = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusProcessing})
_, _ = svc.UpdateOrderStatus(ctx, order.ID, UpdateOrderStatusInput{Status: domain.OrderStatusCompleted})

	_, err = svc.CancelOrder(ctx, order.ID, CancelOrderInput{})
	if err == nil {
		t.Fatal("expected error for cancelling completed order")
	}
}

func TestOrderService_CancelOrder_TerminalAlreadyCancelled(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newMockOrderService()

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Cancel once.
_, _ = svc.CancelOrder(ctx, order.ID, CancelOrderInput{})
	// Cancel again — should fail.
	_, err = svc.CancelOrder(ctx, order.ID, CancelOrderInput{})
	if err == nil {
		t.Fatal("expected error for cancelling already-cancelled order")
	}
}

func TestOrderService_CancelOrder_NotFound(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newMockOrderService()

	_, err := svc.CancelOrder(ctx, uuid.New(), CancelOrderInput{})
	if err == nil {
		t.Fatal("expected error for non-existent order")
	}
}

func TestOrderService_CancelOrder_AlreadyCompletedTaskUnaffected(t *testing.T) {
	ctx := context.Background()
	svc, _, taskRepo := newMockOrderService()

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines:       []CreateOrderLineInput{{SKUID: uuid.New(), OrderedQty: 10}},
		CreatedBy:   "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	lineID := order.Lines[0].ID
	// Create one completed task and one pending task.
	completedTask := &domain.Task{
		TaskType:    domain.TaskTypePick,
		WarehouseID: order.WarehouseID,
		OrderID:     &order.ID,
		OrderLineID: &lineID,
		Status:      domain.TaskStatusCompleted,
		SKUID:       order.Lines[0].SKUID,
		ExpectedQty: 10,
	}
	pendingTask := &domain.Task{
		TaskType:    domain.TaskTypePick,
		WarehouseID: order.WarehouseID,
		OrderID:     &order.ID,
		OrderLineID: &lineID,
		Status:      domain.TaskStatusPending,
		SKUID:       order.Lines[0].SKUID,
		ExpectedQty: 10,
	}
_ = taskRepo.CreateTask(ctx, completedTask)
_ = taskRepo.CreateTask(ctx, pendingTask)

	_, err = svc.CancelOrder(ctx, order.ID, CancelOrderInput{})
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}

	tasks, _ := taskRepo.GetTasksByOrderID(ctx, order.ID)
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	for _, tk := range tasks {
		if tk.ID == completedTask.ID && tk.Status != domain.TaskStatusCompleted {
			t.Errorf("completed task should stay completed, got %q", tk.Status)
		}
		if tk.ID == pendingTask.ID && tk.Status != domain.TaskStatusCancelled {
			t.Errorf("pending task should be cancelled, got %q", tk.Status)
		}
	}
}

func TestOrderService_CancelOrder_AlreadyFulfilledLineUnaffected(t *testing.T) {
	ctx := context.Background()
	repo := newMockOrderRepo()
	svc := newMockOrderServiceWithRepos(repo, newMockTaskRepoForOrder())

	order, err := svc.CreateOrder(ctx, CreateOrderInput{
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: uuid.New(),
		Lines: []CreateOrderLineInput{
			{SKUID: uuid.New(), OrderedQty: 10},
			{SKUID: uuid.New(), OrderedQty: 5},
		},
		CreatedBy: "testuser",
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Mark first line as allocated then fulfilled.
_, _ = svc.UpdateOrderLineStatus(ctx, order.Lines[0].ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusAllocated})
_, _ = svc.UpdateOrderLineStatus(ctx, order.Lines[0].ID, UpdateOrderLineStatusInput{Status: domain.OrderLineStatusFulfilled})

	// Cancel the order.
	cancelled, err := svc.CancelOrder(ctx, order.ID, CancelOrderInput{})
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}

	for _, l := range cancelled.Lines {
		if l.ID == order.Lines[0].ID && l.Status != domain.OrderLineStatusFulfilled {
			t.Errorf("fulfilled line should stay fulfilled, got %q", l.Status)
		}
		if l.ID == order.Lines[1].ID && l.Status != domain.OrderLineStatusCancelled {
			t.Errorf("pending line should be cancelled, got %q", l.Status)
		}
	}
}
