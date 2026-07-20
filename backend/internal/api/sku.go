// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"log/slog"
	"net/http"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/service"
)

// SKUHandler handles HTTP requests for SKU resources.
type SKUHandler struct {
	svc *service.SKUService
	log *slog.Logger
}

// NewSKUHandler creates a new SKUHandler.
func NewSKUHandler(svc *service.SKUService, log *slog.Logger) *SKUHandler {
	return &SKUHandler{svc: svc, log: log}
}

// ── Response Types ─────────────────────────────────────────────────────────────────────

// skuResponse is the JSON shape returned for SKU endpoints.
type skuResponse struct {
	ID          string            `json:"id"`
	Code        string            `json:"code"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Barcode     string            `json:"barcode"`
	UOM         uomResponse       `json:"uom"`
	Attributes  domain.Attributes `json:"attributes,omitempty"`
	Category    string            `json:"category"`
	Status      string            `json:"status"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// uomResponse is the JSON shape for the UOM sub-object.
type uomResponse struct {
	BaseUnit string  `json:"base_unit"`
	PackUnit string  `json:"pack_unit"`
	PackQty  int     `json:"pack_qty"`
	Weight   float64 `json:"weight"`
	Volume   float64 `json:"volume"`
	Length   float64 `json:"length"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
}

func toSKUResponse(s *domain.SKU) skuResponse {
	return skuResponse{
		ID:          s.ID.String(),
		Code:        s.Code,
		Name:        s.Name,
		Description: s.Description,
		Barcode:     s.Barcode,
		UOM: uomResponse{
			BaseUnit: s.UOM.BaseUnit,
			PackUnit: s.UOM.PackUnit,
			PackQty:  s.UOM.PackQty,
			Weight:   s.UOM.Weight,
			Volume:   s.UOM.Volume,
			Length:   s.UOM.Length,
			Width:    s.UOM.Width,
			Height:   s.UOM.Height,
		},
		Attributes: s.Attributes,
		Category:   s.Category,
		Status:     string(s.Status),
		CreatedAt:  s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ── SKU Handlers ───────────────────────────────────────────────────────────────────────

// CreateSKU handles POST /api/v1/skus
func (h *SKUHandler) CreateSKU(w http.ResponseWriter, r *http.Request) {
	var input service.CreateSKUInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	sku, err := h.svc.CreateSKU(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteCreated(w, toSKUResponse(sku))
}

// GetSKU handles GET /api/v1/skus/{id}
func (h *SKUHandler) GetSKU(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	sku, err := h.svc.GetSKU(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toSKUResponse(sku))
}

// ListSKUs handles GET /api/v1/skus
func (h *SKUHandler) ListSKUs(w http.ResponseWriter, r *http.Request) {
	skus, err := h.svc.ListSKUs(r.Context())
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]skuResponse, 0, len(skus))
	for _, s := range skus {
		resp = append(resp, toSKUResponse(s))
	}

	WriteJSON(w, http.StatusOK, resp)
}

// UpdateSKU handles PUT /api/v1/skus/{id}
func (h *SKUHandler) UpdateSKU(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.UpdateSKUInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	sku, err := h.svc.UpdateSKU(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toSKUResponse(sku))
}
