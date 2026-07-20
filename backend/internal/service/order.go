// Package service implements business logic orchestration for the WMS domain.
package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// OrderService orchestrates business logic for orders and order lines.
type OrderService struct {
	repo repository.OrderRepository
}

// NewOrderService creates a new OrderService.
func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// ── Input Types ──────────────────────────────────────────────────────────────────────────

// CreateOrderInput is the input for creating a new order.
type CreateOrderInput struct {
	OrderNo      string              `json:"order_no,omitempty"`      // Auto-generated if empty
	OrderType    domain.OrderType    `json:"order_type"`
	WarehouseID  uuid.UUID           `json:"warehouse_id"`
	Priority     domain.OrderPriority `json:"priority,omitempty"`     // Default "normal"
	ExternalRef  string              `json:"external_ref,omitempty"`
	ExternalType string              `json:"external_type,omitempty"`
	Lines        []CreateOrderLineInput `json:"lines"`
	Notes        string              `json:"notes,omitempty"`
	CreatedBy    string              `json:"created_by"`
}

// Validate checks the input for business rule violations.
func (in *CreateOrderInput) Validate() error {
	if !isValidOrderType(in.OrderType) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid order_type: %s", in.OrderType))
	}
	if in.WarehouseID == uuid.Nil {
		return pkgerrors.NewInvalidInput("warehouse_id is required")
	}
	if in.Priority != "" && !isValidOrderPriority(in.Priority) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid priority: %s", in.Priority))
	}
	if len(in.Lines) == 0 {
		return pkgerrors.NewInvalidInput("at least one order line is required")
	}
	if in.CreatedBy == "" {
		return pkgerrors.NewInvalidInput("created_by is required")
	}
	for i, line := range in.Lines {
		if err := line.Validate(); err != nil {
			return fmt.Errorf("line %d: %w", i+1, err)
		}
	}
	return nil
}

// CreateOrderLineInput is the input for a single order line within CreateOrderInput.
type CreateOrderLineInput struct {
	SKUID      uuid.UUID `json:"sku_id"`
	OrderedQty float64   `json:"ordered_qty"`
	UOM        string    `json:"uom,omitempty"`     // Default "EA"
	BatchNo    string    `json:"batch_no,omitempty"` // Preferred batch
	Notes      string    `json:"notes,omitempty"`
}

// Validate checks the line input for business rule violations.
func (in *CreateOrderLineInput) Validate() error {
	if in.SKUID == uuid.Nil {
		return pkgerrors.NewInvalidInput("sku_id is required")
	}
	if in.OrderedQty <= 0 {
		return pkgerrors.NewInvalidInput("ordered_qty must be positive")
	}
	return nil
}

// AddOrderLineInput is the input for adding a line to an existing order.
type AddOrderLineInput struct {
	SKUID      uuid.UUID `json:"sku_id"`
	OrderedQty float64   `json:"ordered_qty"`
	UOM        string    `json:"uom,omitempty"`
	BatchNo    string    `json:"batch_no,omitempty"`
	Notes      string    `json:"notes,omitempty"`
}

// Validate checks the input for business rule violations.
func (in *AddOrderLineInput) Validate() error {
	if in.SKUID == uuid.Nil {
		return pkgerrors.NewInvalidInput("sku_id is required")
	}
	if in.OrderedQty <= 0 {
		return pkgerrors.NewInvalidInput("ordered_qty must be positive")
	}
	return nil
}

// UpdateOrderStatusInput is the input for updating an order's status.
type UpdateOrderStatusInput struct {
	Status domain.OrderStatus `json:"status"`
	Notes  string             `json:"notes,omitempty"`
}

// Validate checks the input for business rule violations.
func (in *UpdateOrderStatusInput) Validate() error {
	if !isValidOrderStatus(in.Status) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid order status: %s", in.Status))
	}
	return nil
}

// ── Service Methods ──────────────────────────────────────────────────────────────────────

// CreateOrder validates input and creates a new order with its lines.
func (s *OrderService) CreateOrder(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Generate order number if not provided.
	orderNo := input.OrderNo
	if orderNo == "" {
		orderNo = generateOrderNo(input.OrderType)
	}

	priority := input.Priority
	if priority == "" {
		priority = domain.OrderPriorityNormal
	}

	order := &domain.Order{
		OrderNo:      orderNo,
		OrderType:    input.OrderType,
		WarehouseID:  input.WarehouseID,
		Status:       domain.OrderStatusDraft,
		Priority:     priority,
		ExternalRef:  input.ExternalRef,
		ExternalType: input.ExternalType,
		Notes:        input.Notes,
		CreatedBy:    input.CreatedBy,
	}

	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("order service: create: %w", err)
	}

	// Create order lines.
	for i, lineInput := range input.Lines {
		uom := lineInput.UOM
		if uom == "" {
			uom = "EA"
		}

		line := &domain.OrderLine{
			OrderID:     order.ID,
			LineNo:      i + 1,
			SKUID:       lineInput.SKUID,
			OrderedQty:  lineInput.OrderedQty,
			FulfilledQty: 0,
			UOM:         uom,
			BatchNo:     lineInput.BatchNo,
			Status:      domain.OrderLineStatusPending,
			Notes:       lineInput.Notes,
		}

		if err := s.repo.CreateOrderLine(ctx, line); err != nil {
			return nil, fmt.Errorf("order service: create line %d: %w", i+1, err)
		}
		order.Lines = append(order.Lines, *line)
	}

	return order, nil
}

// GetOrder retrieves an order with its lines populated.
func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	order, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order service: get %s: %w", id, err)
	}

	lines, err := s.repo.GetOrderLines(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order service: get lines %s: %w", id, err)
	}
	order.Lines = make([]domain.OrderLine, len(lines))
	for i, l := range lines {
		order.Lines[i] = *l
	}

	return order, nil
}

// GetOrderByNo retrieves an order by its business order number.
func (s *OrderService) GetOrderByNo(ctx context.Context, orderNo string) (*domain.Order, error) {
	order, err := s.repo.GetOrderByNo(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("order service: get by no %s: %w", orderNo, err)
	}

	lines, err := s.repo.GetOrderLines(ctx, order.ID)
	if err != nil {
		return nil, fmt.Errorf("order service: get lines %s: %w", order.ID, err)
	}
	order.Lines = make([]domain.OrderLine, len(lines))
	for i, l := range lines {
		order.Lines[i] = *l
	}

	return order, nil
}

// ListOrders returns orders matching the specified filter.
func (s *OrderService) ListOrders(ctx context.Context, filter repository.OrderFilter) ([]*domain.Order, error) {
	orders, err := s.repo.ListOrders(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("order service: list: %w", err)
	}
	return orders, nil
}

// UpdateOrderStatus validates the state transition and updates the order status.
func (s *OrderService) UpdateOrderStatus(ctx context.Context, id uuid.UUID, input UpdateOrderStatusInput) (*domain.Order, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	order, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order service: update status %s: %w", id, err)
	}

	// Validate the state transition.
	if !isValidOrderTransition(order.Status, input.Status) {
		return nil, pkgerrors.NewInvalidStatus(string(order.Status), string(input.Status))
	}

	if err := s.repo.UpdateOrderStatus(ctx, id, input.Status); err != nil {
		return nil, fmt.Errorf("order service: update status %s: %w", id, err)
	}

	// Re-fetch to get updated state.
	updated, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order service: re-fetch after status update %s: %w", id, err)
	}

	// Load lines.
	lines, _ := s.repo.GetOrderLines(ctx, id)
	if lines != nil {
		updated.Lines = make([]domain.OrderLine, len(lines))
		for i, l := range lines {
			updated.Lines[i] = *l
		}
	}

	return updated, nil
}

// AddOrderLine adds a new line to an existing order.
// Only allowed when order is in draft status.
func (s *OrderService) AddOrderLine(ctx context.Context, orderID uuid.UUID, input AddOrderLineInput) (*domain.OrderLine, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order service: add line %s: %w", orderID, err)
	}

	if order.Status != domain.OrderStatusDraft {
		return nil, pkgerrors.NewInvalidInput("can only add lines to draft orders")
	}

	// Determine next line number.
	existingLines, err := s.repo.GetOrderLines(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order service: get existing lines %s: %w", orderID, err)
	}
	nextLineNo := len(existingLines) + 1

	uom := input.UOM
	if uom == "" {
		uom = "EA"
	}

	line := &domain.OrderLine{
		OrderID:     orderID,
		LineNo:      nextLineNo,
		SKUID:       input.SKUID,
		OrderedQty:  input.OrderedQty,
		FulfilledQty: 0,
		UOM:         uom,
		BatchNo:     input.BatchNo,
		Status:      domain.OrderLineStatusPending,
		Notes:       input.Notes,
	}

	if err := s.repo.CreateOrderLine(ctx, line); err != nil {
		return nil, fmt.Errorf("order service: add line: %w", err)
	}

	return line, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────────────────

// generateOrderNo creates a business order number based on type and timestamp.
func generateOrderNo(orderType domain.OrderType) string {
	now := time.Now()
	prefix := "OR"
	switch orderType {
	case domain.OrderTypeInbound:
		prefix = "IN"
	case domain.OrderTypeOutbound:
		prefix = "OUT"
	case domain.OrderTypeTransfer:
		prefix = "TR"
	case domain.OrderTypeReturn:
		prefix = "RET"
	}
	return fmt.Sprintf("%s-%s-%06d", prefix, now.Format("20060102"), now.UnixMilli()%1000000)
}

func isValidOrderType(t domain.OrderType) bool {
	switch t {
	case domain.OrderTypeInbound, domain.OrderTypeOutbound, domain.OrderTypeTransfer, domain.OrderTypeReturn:
		return true
	}
	return false
}

func isValidOrderPriority(p domain.OrderPriority) bool {
	switch p {
	case domain.OrderPriorityLow, domain.OrderPriorityNormal, domain.OrderPriorityHigh, domain.OrderPriorityUrgent:
		return true
	}
	return false
}

func isValidOrderStatus(s domain.OrderStatus) bool {
	switch s {
	case domain.OrderStatusDraft, domain.OrderStatusConfirmed, domain.OrderStatusProcessing,
		domain.OrderStatusPartial, domain.OrderStatusCompleted, domain.OrderStatusCancelled:
		return true
	}
	return false
}

// isValidOrderTransition validates an order status state machine transition.
// Valid transitions:
//
//	draft → confirmed, cancelled
//	confirmed → processing, cancelled
//	processing → completed, partial, cancelled
//	partial → completed, cancelled
func isValidOrderTransition(current, target domain.OrderStatus) bool {
	if current == target {
		return false // No-op
	}

	// Cancelled is terminal — no further transitions allowed.
	if current == domain.OrderStatusCancelled {
		return false
	}
	// Completed is terminal — no further transitions allowed.
	if current == domain.OrderStatusCompleted {
		return false
	}

	// Any non-terminal status CAN be cancelled.
	if target == domain.OrderStatusCancelled {
		return true
	}

	valid := map[domain.OrderStatus][]domain.OrderStatus{
		domain.OrderStatusDraft:     {domain.OrderStatusConfirmed},
		domain.OrderStatusConfirmed: {domain.OrderStatusProcessing},
		domain.OrderStatusProcessing: {domain.OrderStatusCompleted, domain.OrderStatusPartial},
		domain.OrderStatusPartial:   {domain.OrderStatusCompleted},
	}

	allowed, ok := valid[current]
	if !ok {
		return false
	}

	return slices.Contains(allowed, target)
}
