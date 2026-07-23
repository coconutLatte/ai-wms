package domain

import (
	"time"

	"github.com/google/uuid"
)

// Shipment represents an outbound shipment dispatched from the warehouse.
// Each shipment is linked to a confirmed outbound order and tracks carrier
// information, tracking details, and delivery status.
type Shipment struct {
	ID                uuid.UUID      `json:"id"`
	ShipmentNo        string         `json:"shipment_no"`        // e.g. "SHP-20260724-001"
	OrderID           uuid.UUID      `json:"order_id"`           // Outbound order being shipped
	WarehouseID       uuid.UUID      `json:"warehouse_id"`
	Status            ShipmentStatus `json:"status"`
	Carrier           string         `json:"carrier"`            // Carrier name, e.g. "FedEx", "UPS"
	TrackingNo        string         `json:"tracking_no,omitempty"`
	CarrierService    string         `json:"carrier_service,omitempty"` // e.g. "Ground", "NextDay"
	EstimatedDelivery *time.Time     `json:"estimated_delivery,omitempty"`
	ActualDelivery    *time.Time     `json:"actual_delivery,omitempty"`
	Notes             string         `json:"notes,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	ShippedAt         *time.Time     `json:"shipped_at,omitempty"`
	DeliveredAt       *time.Time     `json:"delivered_at,omitempty"`
}

// ShipmentStatus represents the lifecycle state of a shipment.
type ShipmentStatus string

const (
	ShipmentStatusPending   ShipmentStatus = "pending"    // Created, awaiting pickup
	ShipmentStatusInTransit ShipmentStatus = "in_transit" // Carrier has picked up
	ShipmentStatusDelivered ShipmentStatus = "delivered"  // Confirmed delivered
	ShipmentStatusCancelled ShipmentStatus = "cancelled"  // Cancelled before/dispatch
)

// ── Shipment State Machine ──────────────────────────────────────────────────

// IsTerminal returns true if the shipment is in a terminal (immutable) state.
func (s *Shipment) IsTerminal() bool {
	return s.Status == ShipmentStatusDelivered || s.Status == ShipmentStatusCancelled
}

// CanTransitionTo checks whether the shipment can transition from its current
// status to the target status.
//
// Valid transitions:
//
//	pending    → in_transit, cancelled
//	in_transit → delivered, cancelled
//	delivered  → (terminal)
//	cancelled  → (terminal)
func (s *Shipment) CanTransitionTo(target ShipmentStatus) bool {
	if s.Status == target {
		return false
	}
	if s.IsTerminal() {
		return false
	}
	// Any non-terminal status can be cancelled.
	if target == ShipmentStatusCancelled {
		return true
	}

	switch s.Status {
	case ShipmentStatusPending:
		return target == ShipmentStatusInTransit
	case ShipmentStatusInTransit:
		return target == ShipmentStatusDelivered
	default:
		return false
	}
}

// CanBeShipped returns true if the shipment can be dispatched to the carrier.
func (s *Shipment) CanBeShipped() bool {
	return s.Status == ShipmentStatusPending
}

// CanBeDelivered returns true if the shipment can be marked as delivered.
func (s *Shipment) CanBeDelivered() bool {
	return s.Status == ShipmentStatusInTransit
}
