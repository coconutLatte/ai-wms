package api

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	"github.com/ai-wms/ai-wms/backend/internal/service"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// ShipmentHandler handles HTTP requests for shipment resources.
type ShipmentHandler struct {
	svc *service.ShipmentService
	log *slog.Logger
}

// NewShipmentHandler creates a new ShipmentHandler.
func NewShipmentHandler(svc *service.ShipmentService, log *slog.Logger) *ShipmentHandler {
	return &ShipmentHandler{svc: svc, log: log}
}

// ── Response Types ─────────────────────────────────────────────────────────

type shipmentResponse struct {
	ID                string  `json:"id"`
	ShipmentNo        string  `json:"shipment_no"`
	OrderID           string  `json:"order_id"`
	WarehouseID       string  `json:"warehouse_id"`
	Status            string  `json:"status"`
	Carrier           string  `json:"carrier"`
	TrackingNo        string  `json:"tracking_no,omitempty"`
	CarrierService    string  `json:"carrier_service,omitempty"`
	EstimatedDelivery string  `json:"estimated_delivery,omitempty"`
	ActualDelivery    string  `json:"actual_delivery,omitempty"`
	Notes             string  `json:"notes,omitempty"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	ShippedAt         string  `json:"shipped_at,omitempty"`
	DeliveredAt       string  `json:"delivered_at,omitempty"`
}

func toShipmentResponse(s *domain.Shipment) shipmentResponse {
	r := shipmentResponse{
		ID:          s.ID.String(),
		ShipmentNo:  s.ShipmentNo,
		OrderID:     s.OrderID.String(),
		WarehouseID: s.WarehouseID.String(),
		Status:      string(s.Status),
		Carrier:     s.Carrier,
		TrackingNo:  s.TrackingNo,
		CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if s.CarrierService != "" {
		r.CarrierService = s.CarrierService
	}
	if s.EstimatedDelivery != nil {
		r.EstimatedDelivery = s.EstimatedDelivery.Format("2006-01-02T15:04:05Z")
	}
	if s.ActualDelivery != nil {
		r.ActualDelivery = s.ActualDelivery.Format("2006-01-02T15:04:05Z")
	}
	if s.Notes != "" {
		r.Notes = s.Notes
	}
	if s.ShippedAt != nil {
		r.ShippedAt = s.ShippedAt.Format("2006-01-02T15:04:05Z")
	}
	if s.DeliveredAt != nil {
		r.DeliveredAt = s.DeliveredAt.Format("2006-01-02T15:04:05Z")
	}
	return r
}

// ── Handlers ────────────────────────────────────────────────────────────────

// CreateShipment handles POST /api/v1/shipments
func (h *ShipmentHandler) CreateShipment(w http.ResponseWriter, r *http.Request) {
	var input service.CreateShipmentInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	shipment, err := h.svc.CreateShipment(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteCreated(w, toShipmentResponse(shipment))
}

// GetShipment handles GET /api/v1/shipments/{id}
func (h *ShipmentHandler) GetShipment(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	shipment, err := h.svc.GetShipment(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toShipmentResponse(shipment))
}

// ListShipments handles GET /api/v1/shipments
func (h *ShipmentHandler) ListShipments(w http.ResponseWriter, r *http.Request) {
	page, pageSize := PaginationParams(r)
	offset := paginationOffset(page, pageSize)

	filter := repository.ShipmentFilter{
		Limit:  pageSize,
		Offset: offset,
	}

	if raw := QueryParam(r, "warehouse_id", ""); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			WriteError(w, r, pkgerrors.NewInvalidInput("invalid warehouse_id UUID"))
			return
		}
		filter.WarehouseID = id
	}
	if raw := QueryParam(r, "order_id", ""); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			WriteError(w, r, pkgerrors.NewInvalidInput("invalid order_id UUID"))
			return
		}
		filter.OrderID = id
	}
	if raw := QueryParam(r, "status", ""); raw != "" {
		filter.Status = raw
	}
	if raw := QueryParam(r, "carrier", ""); raw != "" {
		filter.Carrier = raw
	}

	shipments, total, err := h.svc.ListShipments(r.Context(), filter)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]shipmentResponse, 0, len(shipments))
	for _, s := range shipments {
		resp = append(resp, toShipmentResponse(s))
	}

	WriteJSON(w, http.StatusOK, ListResponse[shipmentResponse]{
		Data:       resp,
		Pagination: NewPaginationMeta(total, page, pageSize),
	})
}

// UpdateShipmentStatus handles PUT /api/v1/shipments/{id}/status
func (h *ShipmentHandler) UpdateShipmentStatus(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input struct {
		Status string `json:"status"`
	}
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	target := domain.ShipmentStatus(input.Status)
	shipment, err := h.svc.UpdateShipmentStatus(r.Context(), id, target)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toShipmentResponse(shipment))
}

// UpdateTracking handles PUT /api/v1/shipments/{id}/tracking
// Updates carrier and tracking number information.
func (h *ShipmentHandler) UpdateTracking(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.UpdateTrackingInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	shipment, err := h.svc.UpdateTracking(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toShipmentResponse(shipment))
}

// DeliverShipment handles PUT /api/v1/shipments/{id}/deliver
// Marks a shipment as delivered.
func (h *ShipmentHandler) DeliverShipment(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	shipment, err := h.svc.DeliverShipment(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toShipmentResponse(shipment))
}
