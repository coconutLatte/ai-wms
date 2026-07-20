// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"

	"github.com/ai-wms/ai-wms/backend/internal/service"
)

// OrderHandler handles HTTP requests for order resources.
type OrderHandler struct {
	svc *service.OrderService
	log *slog.Logger
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(svc *service.OrderService, log *slog.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, log: log}
}

// ── Response Types ─────────────────────────────────────────────────────────────────────

// orderResponse is the JSON shape returned for order endpoints.
type orderResponse struct {
	ID           string              `json:"id"`
	OrderNo      string              `json:"order_no"`
	OrderType    string              `json:"order_type"`
	WarehouseID  string              `json:"warehouse_id"`
	Status       string              `json:"status"`
	Priority     string              `json:"priority"`
	ExternalRef  string              `json:"external_ref,omitempty"`
	ExternalType string              `json:"external_type,omitempty"`
	Lines        []orderLineResponse `json:"lines"`
	Notes        string              `json:"notes,omitempty"`
	CreatedAt    string              `json:"created_at"`
	UpdatedAt    string              `json:"updated_at"`
	CompletedAt  string              `json:"completed_at,omitempty"`
	CreatedBy    string              `json:"created_by"`
}

// orderLineResponse is the JSON shape for order line items.
type orderLineResponse struct {
	ID           string  `json:"id"`
	OrderID      string  `json:"order_id"`
	LineNo       int     `json:"line_no"`
	SKUID        string  `json:"sku_id"`
	OrderedQty   float64 `json:"ordered_qty"`
	FulfilledQty float64 `json:"fulfilled_qty"`
	UOM          string  `json:"uom"`
	BatchNo      string  `json:"batch_no,omitempty"`
	Status       string  `json:"status"`
	Notes        string  `json:"notes,omitempty"`
}

func toOrderResponse(o *domain.Order) orderResponse {
	r := orderResponse{
		ID:           o.ID.String(),
		OrderNo:      o.OrderNo,
		OrderType:    string(o.OrderType),
		WarehouseID:  o.WarehouseID.String(),
		Status:       string(o.Status),
		Priority:     string(o.Priority),
		ExternalRef:  o.ExternalRef,
		ExternalType: o.ExternalType,
		Notes:        o.Notes,
		CreatedAt:    o.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    o.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		CreatedBy:    o.CreatedBy,
	}
	if o.CompletedAt != nil {
		r.CompletedAt = o.CompletedAt.Format("2006-01-02T15:04:05Z")
	}
	r.Lines = make([]orderLineResponse, len(o.Lines))
	for i, l := range o.Lines {
		r.Lines[i] = toOrderLineResponse(&l)
	}
	return r
}

func toOrderLineResponse(l *domain.OrderLine) orderLineResponse {
	return orderLineResponse{
		ID:           l.ID.String(),
		OrderID:      l.OrderID.String(),
		LineNo:       l.LineNo,
		SKUID:        l.SKUID.String(),
		OrderedQty:   l.OrderedQty,
		FulfilledQty: l.FulfilledQty,
		UOM:          l.UOM,
		BatchNo:      l.BatchNo,
		Status:       string(l.Status),
		Notes:        l.Notes,
	}
}

// orderSummaryResponse is a lighter response shape for list endpoints (no lines).
type orderSummaryResponse struct {
	ID           string `json:"id"`
	OrderNo      string `json:"order_no"`
	OrderType    string `json:"order_type"`
	WarehouseID  string `json:"warehouse_id"`
	Status       string `json:"status"`
	Priority     string `json:"priority"`
	ExternalRef  string `json:"external_ref,omitempty"`
	ExternalType string `json:"external_type,omitempty"`
	Notes        string `json:"notes,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	CompletedAt  string `json:"completed_at,omitempty"`
	CreatedBy    string `json:"created_by"`
}

func toOrderSummaryResponse(o *domain.Order) orderSummaryResponse {
	r := orderSummaryResponse{
		ID:           o.ID.String(),
		OrderNo:      o.OrderNo,
		OrderType:    string(o.OrderType),
		WarehouseID:  o.WarehouseID.String(),
		Status:       string(o.Status),
		Priority:     string(o.Priority),
		ExternalRef:  o.ExternalRef,
		ExternalType: o.ExternalType,
		Notes:        o.Notes,
		CreatedAt:    o.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    o.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		CreatedBy:    o.CreatedBy,
	}
	if o.CompletedAt != nil {
		r.CompletedAt = o.CompletedAt.Format("2006-01-02T15:04:05Z")
	}
	return r
}

// ── Order Handlers ─────────────────────────────────────────────────────────────────────

// CreateOrder handles POST /api/v1/orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var input service.CreateOrderInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	order, err := h.svc.CreateOrder(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteCreated(w, toOrderResponse(order))
}

// GetOrder handles GET /api/v1/orders/{id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	order, err := h.svc.GetOrder(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toOrderResponse(order))
}

// ListOrders handles GET /api/v1/orders
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	page, pageSize := PaginationParams(r)
	offset := paginationOffset(page, pageSize)

	filter := repository.OrderFilter{
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
	if raw := QueryParam(r, "order_type", ""); raw != "" {
		filter.OrderType = domain.OrderType(raw)
	}
	if raw := QueryParam(r, "status", ""); raw != "" {
		filter.Status = domain.OrderStatus(raw)
	}

	orders, total, err := h.svc.ListOrders(r.Context(), filter)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]orderSummaryResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, toOrderSummaryResponse(o))
	}

	WriteJSON(w, http.StatusOK, ListResponse[orderSummaryResponse]{
		Data:       resp,
		Pagination: NewPaginationMeta(total, page, pageSize),
	})
}

// UpdateOrderStatus handles PUT /api/v1/orders/{id}/status
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.UpdateOrderStatusInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	order, err := h.svc.UpdateOrderStatus(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toOrderResponse(order))
}

// AddOrderLine handles POST /api/v1/orders/{id}/lines
func (h *OrderHandler) AddOrderLine(w http.ResponseWriter, r *http.Request) {
	orderID, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.AddOrderLineInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	line, err := h.svc.AddOrderLine(r.Context(), orderID, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteCreated(w, toOrderLineResponse(line))
}
