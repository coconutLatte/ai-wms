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
	if raw := QueryParam(r, "order_no", ""); raw != "" {
		filter.OrderNo = raw
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

// UpdateOrderLineStatus handles PUT /api/v1/orders/{id}/lines/{lineId}/status
func (h *OrderHandler) UpdateOrderLineStatus(w http.ResponseWriter, r *http.Request) {
	lineID, err := PathUUID(r, "lineId")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.UpdateOrderLineStatusInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	line, err := h.svc.UpdateOrderLineStatus(r.Context(), lineID, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toOrderLineResponse(line))
}

// ── ASN Response Types ──────────────────────────────────────────────────────────────

// asnResponse is the JSON shape returned for ASN endpoints.
type asnResponse struct {
	ID          string          `json:"id"`
	ASNNo       string          `json:"asn_no"`
	WarehouseID string          `json:"warehouse_id"`
	OrderID     string          `json:"order_id,omitempty"`
	Carrier     string          `json:"carrier,omitempty"`
	TrackingNo  string          `json:"tracking_no,omitempty"`
	ExpectedAt  string          `json:"expected_at"`
	ArrivedAt   string          `json:"arrived_at,omitempty"`
	Lines       []asnLineResponse `json:"lines"`
	Status      string          `json:"status"`
	CreatedAt   string          `json:"created_at"`
}

// asnLineResponse is the JSON shape for ASN line items.
type asnLineResponse struct {
	ID          string  `json:"id"`
	ASNID       string  `json:"asn_id"`
	SKUID       string  `json:"sku_id"`
	ExpectedQty float64 `json:"expected_qty"`
	ReceivedQty float64 `json:"received_qty"`
	BatchNo     string  `json:"batch_no,omitempty"`
	Status      string  `json:"status"`
}

// asnSummaryResponse is a lighter response shape for list endpoints (no lines).
type asnSummaryResponse struct {
	ID          string `json:"id"`
	ASNNo       string `json:"asn_no"`
	WarehouseID string `json:"warehouse_id"`
	OrderID     string `json:"order_id,omitempty"`
	Carrier     string `json:"carrier,omitempty"`
	TrackingNo  string `json:"tracking_no,omitempty"`
	ExpectedAt  string `json:"expected_at"`
	ArrivedAt   string `json:"arrived_at,omitempty"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

func toASNResponse(a *domain.ASN) asnResponse {
	r := asnResponse{
		ID:          a.ID.String(),
		ASNNo:       a.ASNNo,
		WarehouseID: a.WarehouseID.String(),
		Carrier:     a.Carrier,
		TrackingNo:  a.TrackingNo,
		ExpectedAt:  a.ExpectedAt.Format("2006-01-02T15:04:05Z"),
		Status:      string(a.Status),
		CreatedAt:   a.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if a.OrderID != uuid.Nil {
		r.OrderID = a.OrderID.String()
	}
	if a.ArrivedAt != nil {
		r.ArrivedAt = a.ArrivedAt.Format("2006-01-02T15:04:05Z")
	}
	r.Lines = make([]asnLineResponse, len(a.Lines))
	for i, l := range a.Lines {
		r.Lines[i] = toASNLineResponse(&l)
	}
	return r
}

func toASNLineResponse(l *domain.ASNLine) asnLineResponse {
	return asnLineResponse{
		ID:          l.ID.String(),
		ASNID:       l.ASNID.String(),
		SKUID:       l.SKUID.String(),
		ExpectedQty: l.ExpectedQty,
		ReceivedQty: l.ReceivedQty,
		BatchNo:     l.BatchNo,
		Status:      string(l.Status),
	}
}

func toASNSummaryResponse(a *domain.ASN) asnSummaryResponse {
	r := asnSummaryResponse{
		ID:          a.ID.String(),
		ASNNo:       a.ASNNo,
		WarehouseID: a.WarehouseID.String(),
		Carrier:     a.Carrier,
		TrackingNo:  a.TrackingNo,
		ExpectedAt:  a.ExpectedAt.Format("2006-01-02T15:04:05Z"),
		Status:      string(a.Status),
		CreatedAt:   a.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if a.OrderID != uuid.Nil {
		r.OrderID = a.OrderID.String()
	}
	if a.ArrivedAt != nil {
		r.ArrivedAt = a.ArrivedAt.Format("2006-01-02T15:04:05Z")
	}
	return r
}

// ── ASN Handlers ───────────────────────────────────────────────────────────────────

// CreateASN handles POST /api/v1/asns
func (h *OrderHandler) CreateASN(w http.ResponseWriter, r *http.Request) {
	var input service.CreateASNInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	asn, err := h.svc.CreateASN(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteCreated(w, toASNResponse(asn))
}

// GetASN handles GET /api/v1/asns/{id}
func (h *OrderHandler) GetASN(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	asn, err := h.svc.GetASN(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toASNResponse(asn))
}

// ListASNs handles GET /api/v1/asns
func (h *OrderHandler) ListASNs(w http.ResponseWriter, r *http.Request) {
	page, pageSize := PaginationParams(r)
	offset := paginationOffset(page, pageSize)

	filter := repository.ASNFilter{
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
	if raw := QueryParam(r, "asn_no", ""); raw != "" {
		filter.ASNNo = raw
	}
	if raw := QueryParam(r, "status", ""); raw != "" {
		filter.Status = domain.ASNStatus(raw)
	}

	asns, total, err := h.svc.ListASNs(r.Context(), filter)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]asnSummaryResponse, 0, len(asns))
	for _, a := range asns {
		resp = append(resp, toASNSummaryResponse(a))
	}

	WriteJSON(w, http.StatusOK, ListResponse[asnSummaryResponse]{
		Data:       resp,
		Pagination: NewPaginationMeta(total, page, pageSize),
	})
}

// UpdateASNStatus handles PUT /api/v1/asns/{id}/status
func (h *OrderHandler) UpdateASNStatus(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.UpdateASNStatusInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	asn, err := h.svc.UpdateASNStatus(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toASNResponse(asn))
}

// ReceiveASNLine handles POST /api/v1/asns/{id}/lines/{lineId}/receive
func (h *OrderHandler) ReceiveASNLine(w http.ResponseWriter, r *http.Request) {
	asnID, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	lineID, err := PathUUID(r, "lineId")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.ReceiveASNLineInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	asn, err := h.svc.ReceiveASNLine(r.Context(), asnID, lineID, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toASNResponse(asn))
}
