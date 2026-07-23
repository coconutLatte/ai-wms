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

// ShipmentService orchestrates business logic for shipment operations.
type ShipmentService struct {
	shipmentRepo repository.ShipmentRepository
	orderRepo    repository.OrderRepository
}

// NewShipmentService creates a new ShipmentService.
func NewShipmentService(
	shipmentRepo repository.ShipmentRepository,
	orderRepo repository.OrderRepository,
) *ShipmentService {
	return &ShipmentService{
		shipmentRepo: shipmentRepo,
		orderRepo:    orderRepo,
	}
}

// ── Input Types ──────────────────────────────────────────────────────────────

// CreateShipmentInput is the input for creating a new shipment.
type CreateShipmentInput struct {
	OrderID           uuid.UUID  `json:"order_id"`
	WarehouseID       uuid.UUID  `json:"warehouse_id"`
	Carrier           string     `json:"carrier"`
	TrackingNo        string     `json:"tracking_no,omitempty"`
	CarrierService    string     `json:"carrier_service,omitempty"`
	EstimatedDelivery *time.Time `json:"estimated_delivery,omitempty"`
	Notes             string     `json:"notes,omitempty"`
}

// Validate checks the input for business rule violations.
func (in *CreateShipmentInput) Validate() error {
	if in.OrderID == uuid.Nil {
		return pkgerrors.NewInvalidInput("order_id is required")
	}
	if in.WarehouseID == uuid.Nil {
		return pkgerrors.NewInvalidInput("warehouse_id is required")
	}
	if in.Carrier == "" {
		return pkgerrors.NewInvalidInput("carrier is required")
	}
	return nil
}

// UpdateTrackingInput is the input for updating tracking information.
type UpdateTrackingInput struct {
	Carrier        string `json:"carrier,omitempty"`
	TrackingNo     string `json:"tracking_no,omitempty"`
	CarrierService string `json:"carrier_service,omitempty"`
}

// ── Service Methods ──────────────────────────────────────────────────────────

// CreateShipment creates a new shipment for a confirmed outbound order.
func (s *ShipmentService) CreateShipment(ctx context.Context, input CreateShipmentInput) (*domain.Shipment, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Verify the order exists and is in a shippable state.
	order, err := s.orderRepo.GetOrder(ctx, input.OrderID)
	if err != nil {
		return nil, fmt.Errorf("shipment service: create: get order: %w", err)
	}
	if order.OrderType != domain.OrderTypeOutbound {
		return nil, pkgerrors.NewInvalidInput(
			fmt.Sprintf("only outbound orders can be shipped (current: %s)", order.OrderType))
	}
	if order.Status == domain.OrderStatusDraft || order.Status == domain.OrderStatusCancelled {
		return nil, pkgerrors.NewInvalidInput(
			fmt.Sprintf("cannot create shipment for order with status %s", order.Status))
	}

	shipment := &domain.Shipment{
		ShipmentNo:        generateShipmentNo(),
		OrderID:           input.OrderID,
		WarehouseID:       input.WarehouseID,
		Status:            domain.ShipmentStatusPending,
		Carrier:           input.Carrier,
		TrackingNo:        input.TrackingNo,
		CarrierService:    input.CarrierService,
		EstimatedDelivery: input.EstimatedDelivery,
		Notes:             input.Notes,
	}

	if err := s.shipmentRepo.CreateShipment(ctx, shipment); err != nil {
		return nil, fmt.Errorf("shipment service: create: %w", err)
	}

	return shipment, nil
}

// GetShipment retrieves a shipment by ID.
func (s *ShipmentService) GetShipment(ctx context.Context, id uuid.UUID) (*domain.Shipment, error) {
	shipment, err := s.shipmentRepo.GetShipment(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("shipment service: get %s: %w", id, err)
	}
	return shipment, nil
}

// ListShipments returns shipments matching the specified filter.
func (s *ShipmentService) ListShipments(ctx context.Context, filter repository.ShipmentFilter) ([]*domain.Shipment, int, error) {
	shipments, err := s.shipmentRepo.ListShipments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("shipment service: list: %w", err)
	}

	total, err := s.shipmentRepo.CountShipments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("shipment service: count: %w", err)
	}

	return shipments, total, nil
}

// UpdateShipmentStatus transitions a shipment to a new status.
func (s *ShipmentService) UpdateShipmentStatus(ctx context.Context, id uuid.UUID, target domain.ShipmentStatus) (*domain.Shipment, error) {
	shipment, err := s.shipmentRepo.GetShipment(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("shipment service: update status: %w", err)
	}

	if !shipment.CanTransitionTo(target) {
		return nil, pkgerrors.NewInvalidStatus(string(shipment.Status), string(target))
	}

	if err := s.shipmentRepo.UpdateShipmentStatus(ctx, id, target); err != nil {
		return nil, fmt.Errorf("shipment service: update status: %w", err)
	}

	shipment.Status = target
	now := time.Now()
	shipment.UpdatedAt = now
	if target == domain.ShipmentStatusInTransit {
		shipment.ShippedAt = &now
	}
	if target == domain.ShipmentStatusDelivered {
		shipment.DeliveredAt = &now
		shipment.ActualDelivery = &now
	}

	return shipment, nil
}

// UpdateTracking updates the carrier and tracking information for a shipment.
func (s *ShipmentService) UpdateTracking(ctx context.Context, id uuid.UUID, input UpdateTrackingInput) (*domain.Shipment, error) {
	shipment, err := s.shipmentRepo.GetShipment(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("shipment service: update tracking: %w", err)
	}

	if shipment.IsTerminal() {
		return nil, pkgerrors.NewInvalidInput("cannot update tracking for a terminal shipment")
	}

	carrier := input.Carrier
	if carrier == "" {
		carrier = shipment.Carrier
	}

	if err := s.shipmentRepo.UpdateShipmentTracking(ctx, id, carrier, input.TrackingNo, input.CarrierService); err != nil {
		return nil, fmt.Errorf("shipment service: update tracking: %w", err)
	}

	shipment.Carrier = carrier
	if input.TrackingNo != "" {
		shipment.TrackingNo = input.TrackingNo
	}
	if input.CarrierService != "" {
		shipment.CarrierService = input.CarrierService
	}
	shipment.UpdatedAt = time.Now()

	return shipment, nil
}

// DeliverShipment marks a shipment as delivered.
func (s *ShipmentService) DeliverShipment(ctx context.Context, id uuid.UUID) (*domain.Shipment, error) {
	shipment, err := s.shipmentRepo.GetShipment(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("shipment service: deliver: %w", err)
	}

	if !shipment.CanTransitionTo(domain.ShipmentStatusDelivered) {
		return nil, pkgerrors.NewInvalidStatus(string(shipment.Status), string(domain.ShipmentStatusDelivered))
	}

	if err := s.shipmentRepo.DeliverShipment(ctx, id); err != nil {
		return nil, fmt.Errorf("shipment service: deliver: %w", err)
	}

	shipment.Status = domain.ShipmentStatusDelivered
	now := time.Now()
	shipment.DeliveredAt = &now
	shipment.ActualDelivery = &now
	shipment.UpdatedAt = now

	return shipment, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func generateShipmentNo() string {
	now := time.Now()
	return fmt.Sprintf("SHP-%s-%06d", now.Format("20060102"), now.UnixMilli()%1000000)
}
