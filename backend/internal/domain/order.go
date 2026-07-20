package domain

import (
	"time"

	"github.com/google/uuid"
)

// Order represents a warehouse order (inbound or outbound).
type Order struct {
	ID            uuid.UUID    `json:"id"`
	OrderNo       string       `json:"order_no"`       // Business order number, e.g. "IN-20260720-001"
	OrderType     OrderType    `json:"order_type"`
	WarehouseID   uuid.UUID    `json:"warehouse_id"`
	Status        OrderStatus  `json:"status"`
	Priority      OrderPriority `json:"priority"`
	ExternalRef   string       `json:"external_ref"`   // Reference from ERP/MES, e.g. PO number, SO number
	ExternalType  string       `json:"external_type"`  // "purchase_order", "sales_order", "work_order", "return_order"
	Lines         []OrderLine  `json:"lines"`          // Order line items
	Notes         string       `json:"notes"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	CompletedAt   *time.Time   `json:"completed_at,omitempty"`
	CreatedBy     string       `json:"created_by"`
}

// OrderType classifies inbound vs outbound orders.
type OrderType string

const (
	OrderTypeInbound  OrderType = "inbound"  // Receiving into warehouse
	OrderTypeOutbound OrderType = "outbound" // Shipping out of warehouse
	OrderTypeTransfer OrderType = "transfer" // Internal transfer between warehouses
	OrderTypeReturn   OrderType = "return"   // Customer return to warehouse
)

// OrderStatus represents the lifecycle of an order.
type OrderStatus string

const (
	OrderStatusDraft      OrderStatus = "draft"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing" // Tasks being executed
	OrderStatusPartial    OrderStatus = "partial"    // Partially completed
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// OrderPriority indicates the urgency of order processing.
type OrderPriority string

const (
	OrderPriorityLow      OrderPriority = "low"
	OrderPriorityNormal   OrderPriority = "normal"
	OrderPriorityHigh     OrderPriority = "high"
	OrderPriorityUrgent   OrderPriority = "urgent"
)

// OrderLine represents a single line item within an order.
type OrderLine struct {
	ID          uuid.UUID `json:"id"`
	OrderID     uuid.UUID `json:"order_id"`
	LineNo      int       `json:"line_no"`
	SKUID       uuid.UUID `json:"sku_id"`
	OrderedQty  float64   `json:"ordered_qty"`
	FulfilledQty float64  `json:"fulfilled_qty"`
	UOM         string    `json:"uom"`
	BatchNo     string    `json:"batch_no,omitempty"` // Preferred batch (for FEFO/FIFO allocation)
	Status      OrderLineStatus `json:"status"`
	Notes       string    `json:"notes,omitempty"`
}

// OrderLineStatus represents the fulfillment state of an order line.
type OrderLineStatus string

const (
	OrderLineStatusPending    OrderLineStatus = "pending"
	OrderLineStatusAllocated  OrderLineStatus = "allocated"
	OrderLineStatusPartial    OrderLineStatus = "partial"
	OrderLineStatusFulfilled  OrderLineStatus = "fulfilled"
	OrderLineStatusCancelled  OrderLineStatus = "cancelled"
)

// ASN (Advanced Shipping Notice) – pre-notification of inbound delivery.
type ASN struct {
	ID          uuid.UUID  `json:"id"`
	ASNNo       string     `json:"asn_no"`        // e.g. "ASN-20260720-001"
	WarehouseID uuid.UUID  `json:"warehouse_id"`
	OrderID     uuid.UUID  `json:"order_id,omitempty"` // Linked inbound order
	Carrier     string     `json:"carrier"`
	TrackingNo  string     `json:"tracking_no"`
	ExpectedAt  time.Time  `json:"expected_at"`
	ArrivedAt   *time.Time `json:"arrived_at,omitempty"`
	Lines       []ASNLine  `json:"lines"`
	Status      ASNStatus  `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ASNLine represents a line in an Advanced Shipping Notice.
type ASNLine struct {
	ID          uuid.UUID `json:"id"`
	ASNID       uuid.UUID `json:"asn_id"`
	SKUID       uuid.UUID `json:"sku_id"`
	ExpectedQty float64   `json:"expected_qty"`
	ReceivedQty float64   `json:"received_qty"`
	BatchNo     string    `json:"batch_no,omitempty"`
	Status      ASNLineStatus `json:"status"`
}

// ASNStatus represents the receipt state of an ASN.
type ASNStatus string

const (
	ASNStatusPending     ASNStatus = "pending"
	ASNStatusArrived     ASNStatus = "arrived"
	ASNStatusReceiving   ASNStatus = "receiving"
	ASNStatusPartial     ASNStatus = "partial"
	ASNStatusReceived    ASNStatus = "received"
)

// ASNLineStatus represents the receipt state of an ASN line.
type ASNLineStatus string

const (
	ASNLineStatusPending   ASNLineStatus = "pending"
	ASNLineStatusPartial   ASNLineStatus = "partial"
	ASNLineStatusReceived  ASNLineStatus = "received"
)

// ── State Machine Methods ────────────────────────────────────────────────────

// IsTerminal returns true if the order is in a terminal (immutable) state.
func (o *Order) IsTerminal() bool {
	return o.Status == OrderStatusCancelled || o.Status == OrderStatusCompleted
}

// CanTransitionTo checks whether the order can transition from its current
// status to the target status. This is the authoritative order state machine.
//
// Valid transitions:
//
//	draft      → confirmed, cancelled
//	confirmed  → processing, cancelled
//	processing → completed, partial, cancelled
//	partial    → completed, cancelled
//	cancelled  → (terminal)
//	completed  → (terminal)
func (o *Order) CanTransitionTo(target OrderStatus) bool {
	if o.Status == target {
		return false
	}
	if o.IsTerminal() {
		return false
	}
	// Any non-terminal status can be cancelled.
	if target == OrderStatusCancelled {
		return true
	}

	switch o.Status {
	case OrderStatusDraft:
		return target == OrderStatusConfirmed
	case OrderStatusConfirmed:
		return target == OrderStatusProcessing
	case OrderStatusProcessing:
		return target == OrderStatusCompleted || target == OrderStatusPartial
	case OrderStatusPartial:
		return target == OrderStatusCompleted
	default:
		return false
	}
}

// IsTerminal returns true if the order line is in a terminal state.
func (ol *OrderLine) IsTerminal() bool {
	return ol.Status == OrderLineStatusFulfilled || ol.Status == OrderLineStatusCancelled
}

// CanTransitionTo checks whether the order line can transition state.
//
// Valid transitions:
//
//	pending   → allocated, cancelled
//	allocated → partial, fulfilled, cancelled
//	partial   → fulfilled, cancelled
//	fulfilled → (terminal)
//	cancelled → (terminal)
func (ol *OrderLine) CanTransitionTo(target OrderLineStatus) bool {
	if ol.Status == target {
		return false
	}
	if ol.IsTerminal() {
		return false
	}
	if target == OrderLineStatusCancelled {
		return true
	}

	switch ol.Status {
	case OrderLineStatusPending:
		return target == OrderLineStatusAllocated
	case OrderLineStatusAllocated:
		return target == OrderLineStatusPartial || target == OrderLineStatusFulfilled
	case OrderLineStatusPartial:
		return target == OrderLineStatusFulfilled
	default:
		return false
	}
}

// IsTerminal returns true if the ASN is in a terminal state.
func (a *ASN) IsTerminal() bool {
	return a.Status == ASNStatusReceived
}

// CanTransitionTo checks whether the ASN can transition state.
//
// Valid transitions:
//
//	pending   → arrived
//	arrived   → receiving
//	receiving → partial, received
//	partial   → received
//	received  → (terminal)
func (a *ASN) CanTransitionTo(target ASNStatus) bool {
	if a.Status == target {
		return false
	}
	if a.IsTerminal() {
		return false
	}

	switch a.Status {
	case ASNStatusPending:
		return target == ASNStatusArrived
	case ASNStatusArrived:
		return target == ASNStatusReceiving
	case ASNStatusReceiving:
		return target == ASNStatusPartial || target == ASNStatusReceived
	case ASNStatusPartial:
		return target == ASNStatusReceived
	default:
		return false
	}
}
