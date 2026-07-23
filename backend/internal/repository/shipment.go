package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
)

// ShipmentRepository defines the data access interface for shipment operations.
type ShipmentRepository interface {
	// CreateShipment creates a new shipment.
	CreateShipment(ctx context.Context, s *domain.Shipment) error

	// GetShipment retrieves a shipment by ID.
	GetShipment(ctx context.Context, id uuid.UUID) (*domain.Shipment, error)

	// ListShipments returns shipments matching the specified filter.
	ListShipments(ctx context.Context, filter ShipmentFilter) ([]*domain.Shipment, error)

	// CountShipments returns the total count of shipments matching the filter.
	CountShipments(ctx context.Context, filter ShipmentFilter) (int, error)

	// UpdateShipmentStatus transitions a shipment to a new status.
	UpdateShipmentStatus(ctx context.Context, id uuid.UUID, status domain.ShipmentStatus) error

	// UpdateShipmentTracking updates the tracking information for a shipment.
	UpdateShipmentTracking(ctx context.Context, id uuid.UUID, carrier, trackingNo, carrierService string) error

	// DeliverShipment marks a shipment as delivered.
	DeliverShipment(ctx context.Context, id uuid.UUID) error
}

// ShipmentFilter defines filter parameters for listing shipments.
type ShipmentFilter struct {
	WarehouseID uuid.UUID
	OrderID     uuid.UUID
	Status      string
	Carrier     string
	Limit       int
	Offset      int
}
