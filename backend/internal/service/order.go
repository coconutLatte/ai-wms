// Package service implements business logic orchestration for the WMS domain.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// OrderService orchestrates business logic for orders and order lines.
type OrderService struct {
	repo     repository.OrderRepository
	taskRepo repository.TaskRepository
}

// NewOrderService creates a new OrderService.
func NewOrderService(repo repository.OrderRepository, taskRepo repository.TaskRepository) *OrderService {
	return &OrderService{repo: repo, taskRepo: taskRepo}
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
func (s *OrderService) ListOrders(ctx context.Context, filter repository.OrderFilter) ([]*domain.Order, int, error) {
	orders, err := s.repo.ListOrders(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("order service: list: %w", err)
	}

	total, err := s.repo.CountOrders(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("order service: count: %w", err)
	}

	return orders, total, nil
}

// CountOrdersByStatus returns the number of orders in each status.
// Used by the admin dashboard to show order status distribution.
func (s *OrderService) CountOrdersByStatus(ctx context.Context) (map[domain.OrderStatus]int, error) {
	counts, err := s.repo.CountOrdersByStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("order service: count by status: %w", err)
	}
	return counts, nil
}

// UpdateOrderStatus validates the state transition and updates the order status.
// When status transitions to "confirmed", tasks are auto-generated for the order:
//   - inbound orders  → putaway tasks (one per line)
//   - outbound orders → pick tasks (one per line)
//   - transfer/return → putaway tasks (one per line)
//
// Task generation is idempotent: if tasks already exist for the order, no new tasks are created.
func (s *OrderService) UpdateOrderStatus(ctx context.Context, id uuid.UUID, input UpdateOrderStatusInput) (*domain.Order, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	order, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order service: update status %s: %w", id, err)
	}

	// Validate the state transition.
	if !order.CanTransitionTo(input.Status) {
		return nil, pkgerrors.NewInvalidStatus(string(order.Status), string(input.Status))
	}

	if err := s.repo.UpdateOrderStatus(ctx, id, input.Status); err != nil {
		return nil, fmt.Errorf("order service: update status %s: %w", id, err)
	}

	// Auto-generate tasks when transitioning to "confirmed".
	if input.Status == domain.OrderStatusConfirmed {
		if err := s.generateTasksForOrder(ctx, order); err != nil {
			return nil, fmt.Errorf("order service: generate tasks: %w", err)
		}
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

// UpdateOrderLineStatusInput is the input for updating an order line's status.
type UpdateOrderLineStatusInput struct {
	Status domain.OrderLineStatus `json:"status"`
}

// Validate checks the input for business rule violations.
func (in *UpdateOrderLineStatusInput) Validate() error {
	if !isValidOrderLineStatus(in.Status) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid order line status: %s", in.Status))
	}
	return nil
}

// UpdateOrderLineStatus validates the state transition and updates the order line status.
func (s *OrderService) UpdateOrderLineStatus(ctx context.Context, lineID uuid.UUID, input UpdateOrderLineStatusInput) (*domain.OrderLine, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	line, err := s.repo.GetOrderLine(ctx, lineID)
	if err != nil {
		return nil, fmt.Errorf("order service: update line status %s: %w", lineID, err)
	}

	// Validate the state transition.
	if !line.CanTransitionTo(input.Status) {
		return nil, pkgerrors.NewInvalidStatus(string(line.Status), string(input.Status))
	}

	if err := s.repo.UpdateOrderLineStatus(ctx, lineID, input.Status); err != nil {
		return nil, fmt.Errorf("order service: update line status %s: %w", lineID, err)
	}

	// Re-fetch to get updated state.
	updated, err := s.repo.GetOrderLine(ctx, lineID)
	if err != nil {
		return nil, fmt.Errorf("order service: re-fetch after line status update %s: %w", lineID, err)
	}

	return updated, nil
}

// UpdateASNStatusInput is the input for updating an ASN's status.
type UpdateASNStatusInput struct {
	Status domain.ASNStatus `json:"status"`
}

// Validate checks the input for business rule violations.
func (in *UpdateASNStatusInput) Validate() error {
	if !isValidASNStatus(in.Status) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid ASN status: %s", in.Status))
	}
	return nil
}

// ── ASN Input Types ───────────────────────────────────────────────────────────────────

// CreateASNInput is the input for creating a new ASN.
type CreateASNInput struct {
	ASNNo       string               `json:"asn_no,omitempty"` // Auto-generated if empty
	WarehouseID uuid.UUID            `json:"warehouse_id"`
	OrderID     uuid.UUID            `json:"order_id,omitempty"` // Linked inbound order (optional)
	Carrier     string               `json:"carrier,omitempty"`
	TrackingNo  string               `json:"tracking_no,omitempty"`
	ExpectedAt  time.Time            `json:"expected_at"`
	Lines       []CreateASNLineInput `json:"lines"`
}

// Validate checks the input for business rule violations.
func (in *CreateASNInput) Validate() error {
	if in.WarehouseID == uuid.Nil {
		return pkgerrors.NewInvalidInput("warehouse_id is required")
	}
	if in.ExpectedAt.IsZero() {
		return pkgerrors.NewInvalidInput("expected_at is required")
	}
	if len(in.Lines) == 0 {
		return pkgerrors.NewInvalidInput("at least one ASN line is required")
	}
	for i, line := range in.Lines {
		if err := line.Validate(); err != nil {
			return fmt.Errorf("line %d: %w", i+1, err)
		}
	}
	return nil
}

// CreateASNLineInput is the input for a single ASN line within CreateASNInput.
type CreateASNLineInput struct {
	SKUID       uuid.UUID `json:"sku_id"`
	ExpectedQty float64   `json:"expected_qty"`
	BatchNo     string    `json:"batch_no,omitempty"`
}

// Validate checks the line input for business rule violations.
func (in *CreateASNLineInput) Validate() error {
	if in.SKUID == uuid.Nil {
		return pkgerrors.NewInvalidInput("sku_id is required")
	}
	if in.ExpectedQty <= 0 {
		return pkgerrors.NewInvalidInput("expected_qty must be positive")
	}
	return nil
}

// CreateASN validates input and creates a new ASN with its lines.
func (s *OrderService) CreateASN(ctx context.Context, input CreateASNInput) (*domain.ASN, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Generate ASN number if not provided.
	asnNo := input.ASNNo
	if asnNo == "" {
		asnNo = generateASNNo()
	}

	asn := &domain.ASN{
		ASNNo:       asnNo,
		WarehouseID: input.WarehouseID,
		OrderID:     input.OrderID,
		Carrier:     input.Carrier,
		TrackingNo:  input.TrackingNo,
		ExpectedAt:  input.ExpectedAt,
		Status:      domain.ASNStatusPending,
	}

	if err := s.repo.CreateASN(ctx, asn); err != nil {
		return nil, fmt.Errorf("order service: create asn: %w", err)
	}

	// Create ASN lines.
	for i, lineInput := range input.Lines {
		line := &domain.ASNLine{
			ASNID:       asn.ID,
			SKUID:       lineInput.SKUID,
			ExpectedQty: lineInput.ExpectedQty,
			ReceivedQty: 0,
			BatchNo:     lineInput.BatchNo,
			Status:      domain.ASNLineStatusPending,
		}

		if err := s.repo.CreateASNLine(ctx, line); err != nil {
			return nil, fmt.Errorf("order service: create asn line %d: %w", i+1, err)
		}
		asn.Lines = append(asn.Lines, *line)
	}

	return asn, nil
}

// GetASN retrieves an ASN with its lines populated.
func (s *OrderService) GetASN(ctx context.Context, id uuid.UUID) (*domain.ASN, error) {
	asn, err := s.repo.GetASN(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order service: get asn %s: %w", id, err)
	}

	lines, err := s.repo.GetASNLines(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order service: get asn lines %s: %w", id, err)
	}
	asn.Lines = make([]domain.ASNLine, len(lines))
	for i, l := range lines {
		asn.Lines[i] = *l
	}

	return asn, nil
}

// ListASNs returns ASNs matching the specified filter.
func (s *OrderService) ListASNs(ctx context.Context, filter repository.ASNFilter) ([]*domain.ASN, int, error) {
	asns, err := s.repo.ListASNs(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("order service: list asns: %w", err)
	}

	total, err := s.repo.CountASNs(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("order service: count asns: %w", err)
	}

	return asns, total, nil
}

// UpdateASNStatus validates the state transition and updates the ASN status.
func (s *OrderService) UpdateASNStatus(ctx context.Context, asnID uuid.UUID, input UpdateASNStatusInput) (*domain.ASN, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	asn, err := s.repo.GetASN(ctx, asnID)
	if err != nil {
		return nil, fmt.Errorf("order service: update ASN status %s: %w", asnID, err)
	}

	// Validate the state transition.
	if !asn.CanTransitionTo(input.Status) {
		return nil, pkgerrors.NewInvalidStatus(string(asn.Status), string(input.Status))
	}

	if err := s.repo.UpdateASNStatus(ctx, asnID, input.Status); err != nil {
		return nil, fmt.Errorf("order service: update ASN status %s: %w", asnID, err)
	}

	// Re-fetch to get updated state.
	updated, err := s.repo.GetASN(ctx, asnID)
	if err != nil {
		return nil, fmt.Errorf("order service: re-fetch after ASN status update %s: %w", asnID, err)
	}

	// Load ASN lines.
	lines, _ := s.repo.GetASNLines(ctx, asnID)
	if lines != nil {
		updated.Lines = make([]domain.ASNLine, len(lines))
		for i, l := range lines {
			updated.Lines[i] = *l
		}
	}

	return updated, nil
}

// ── Receive ASN Line ───────────────────────────────────────────────────────

// ReceiveASNLineInput is the input for receiving a quantity on an ASN line.
type ReceiveASNLineInput struct {
	ReceivedQty float64 `json:"received_qty"`
}

// Validate checks the input for business rule violations.
func (in *ReceiveASNLineInput) Validate() error {
	if in.ReceivedQty <= 0 {
		return pkgerrors.NewInvalidInput("received_qty must be positive")
	}
	return nil
}

// ReceiveASNLine records a received quantity on an ASN line.
// Updates the line's received_qty, determines line status (pending/received),
// and updates the parent ASN status accordingly.
func (s *OrderService) ReceiveASNLine(ctx context.Context, asnID, lineID uuid.UUID, input ReceiveASNLineInput) (*domain.ASN, error) {
	// 1. Validate input.
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// 2. Get the ASN line.
	line, err := s.repo.GetASNLine(ctx, lineID)
	if err != nil {
		return nil, fmt.Errorf("order service: receive asn line: %w", err)
	}

	// Verify the line belongs to this ASN.
	if line.ASNID != asnID {
		return nil, pkgerrors.NewInvalidInput("line does not belong to this ASN")
	}

	// 3. Validate: received_qty + existing received_qty <= expected_qty.
	newReceived := line.ReceivedQty + input.ReceivedQty
	if newReceived > line.ExpectedQty {
		return nil, pkgerrors.NewInvalidInput("received_qty exceeds expected_qty")
	}

	// 4. Update received_qty on the line.
	if err := s.repo.UpdateASNLineReceivedQty(ctx, lineID, newReceived); err != nil {
		return nil, fmt.Errorf("order service: update received qty: %w", err)
	}

	// 5. Determine line status.
	var lineStatus domain.ASNLineStatus
	if newReceived >= line.ExpectedQty {
		lineStatus = domain.ASNLineStatusReceived
	} else {
		lineStatus = domain.ASNLineStatusPartial
	}

	// 6. Update line status.
	if err := s.repo.UpdateASNLineStatus(ctx, lineID, lineStatus); err != nil {
		return nil, fmt.Errorf("order service: update line status: %w", err)
	}

	// 7. Get all ASN lines to determine overall ASN status.
	lines, err := s.repo.GetASNLines(ctx, asnID)
	if err != nil {
		return nil, fmt.Errorf("order service: get asn lines: %w", err)
	}

	// Update the just-modified line in the lines slice for status evaluation.
	for i, l := range lines {
		if l.ID == lineID {
			l.ReceivedQty = newReceived
			l.Status = lineStatus
			lines[i] = l
			break
		}
	}

	// 8. Determine ASN status.
	asn, err := s.repo.GetASN(ctx, asnID)
	if err != nil {
		return nil, fmt.Errorf("order service: get asn: %w", err)
	}

	// If ASN is "arrived", first transition to "receiving".
	if asn.Status == domain.ASNStatusArrived && asn.CanTransitionTo(domain.ASNStatusReceiving) {
		if err := s.repo.UpdateASNStatus(ctx, asnID, domain.ASNStatusReceiving); err != nil {
			return nil, fmt.Errorf("order service: update asn status to receiving: %w", err)
		}
		asn.Status = domain.ASNStatusReceiving
	}

	// Check if all lines are fully received.
	allReceived := true
	for _, l := range lines {
		if l.Status != domain.ASNLineStatusReceived {
			allReceived = false
			break
		}
	}

	// 9. If all lines received, transition ASN to "received".
	if allReceived {
		if asn.CanTransitionTo(domain.ASNStatusReceived) {
			if err := s.repo.UpdateASNStatus(ctx, asnID, domain.ASNStatusReceived); err != nil {
				return nil, fmt.Errorf("order service: update asn status to received: %w", err)
			}
		}
	}

	// 10. Return updated ASN with lines loaded.
	updated, err := s.repo.GetASN(ctx, asnID)
	if err != nil {
		return nil, fmt.Errorf("order service: re-fetch asn: %w", err)
	}

	refreshedLines, _ := s.repo.GetASNLines(ctx, asnID)
	if refreshedLines != nil {
		updated.Lines = make([]domain.ASNLine, len(refreshedLines))
		for i, l := range refreshedLines {
			updated.Lines[i] = *l
		}
	}

	return updated, nil
}

// generateTasksForOrder creates warehouse tasks for an order when it is confirmed.
// Task types:
//   - inbound orders  → putaway (one per line, since receiving is handled by ASN)
//   - outbound orders → pick (one per line)
//   - transfer/return → putaway (one per line)
//
// Deduplication: if tasks already exist for this order, no new tasks are created.
func (s *OrderService) generateTasksForOrder(ctx context.Context, order *domain.Order) error {
	// Deduplication: skip if tasks already exist for this order.
	existing, err := s.taskRepo.GetTasksByOrderID(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("check existing tasks: %w", err)
	}
	if len(existing) > 0 {
		return nil // Tasks already generated — idempotent, skip.
	}

	// Determine task type based on order type.
	taskType := domain.TaskTypePutaway
	switch order.OrderType {
	case domain.OrderTypeOutbound:
		taskType = domain.TaskTypePick
	case domain.OrderTypeInbound, domain.OrderTypeTransfer, domain.OrderTypeReturn:
		taskType = domain.TaskTypePutaway
	}

	// Map order priority to task priority (same enum values).
	taskPriority := domain.TaskPriority(order.Priority)

	// Load order lines if not already loaded.
	if len(order.Lines) == 0 {
		lines, err := s.repo.GetOrderLines(ctx, order.ID)
		if err != nil {
			return fmt.Errorf("load order lines: %w", err)
		}
		for _, l := range lines {
			order.Lines = append(order.Lines, *l)
		}
	}

	// Generate one task per order line.
	for _, line := range order.Lines {
		lineID := line.ID // Capture for pointer.

		instructions := buildTaskInstructions(taskType, order.OrderNo, line.LineNo)

		task := &domain.Task{
			TaskNo:       generateTaskNo(),
			TaskType:     taskType,
			WarehouseID:  order.WarehouseID,
			OrderID:      &order.ID,
			OrderLineID:  &lineID,
			Priority:     taskPriority,
			Status:       domain.TaskStatusPending,
			SKUID:        line.SKUID,
			ExpectedQty:  line.OrderedQty,
			UOM:          line.UOM,
			BatchNo:      line.BatchNo,
			Instructions: instructions,
		}

		if err := s.taskRepo.CreateTask(ctx, task); err != nil {
			return fmt.Errorf("create task for line %d: %w", line.LineNo, err)
		}
	}

	return nil
}

// buildTaskInstructions creates human-readable instructions for a warehouse task.
func buildTaskInstructions(taskType domain.TaskType, orderNo string, lineNo int) string {
	switch taskType {
	case domain.TaskTypePick:
		return fmt.Sprintf("Pick for order %s, line %d. Scan location, then SKU barcode to confirm.", orderNo, lineNo)
	case domain.TaskTypePutaway:
		return fmt.Sprintf("Putaway for order %s, line %d. Scan target location barcode to confirm placement.", orderNo, lineNo)
	default:
		return fmt.Sprintf("Task for order %s, line %d.", orderNo, lineNo)
	}
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

// generateASNNo creates a business ASN number based on timestamp.
func generateASNNo() string {
	now := time.Now()
	return fmt.Sprintf("ASN-%s-%06d", now.Format("20060102"), now.UnixMilli()%1000000)
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
// DEPRECATED: Delegates to Order.CanTransitionTo() — keep for reference.
func isValidOrderTransition(current, target domain.OrderStatus) bool {
	o := &domain.Order{Status: current}
	return o.CanTransitionTo(target)
}

func isValidOrderLineStatus(s domain.OrderLineStatus) bool {
	switch s {
	case domain.OrderLineStatusPending, domain.OrderLineStatusAllocated,
		domain.OrderLineStatusPartial, domain.OrderLineStatusFulfilled,
		domain.OrderLineStatusCancelled:
		return true
	}
	return false
}

func isValidASNStatus(s domain.ASNStatus) bool {
	switch s {
	case domain.ASNStatusPending, domain.ASNStatusArrived,
		domain.ASNStatusReceiving, domain.ASNStatusPartial,
		domain.ASNStatusReceived:
		return true
	}
	return false
}
