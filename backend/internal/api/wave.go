// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"

	"github.com/ai-wms/ai-wms/backend/internal/service"
)

// WaveHandler handles HTTP requests for wave resources.
type WaveHandler struct {
	svc *service.WaveService
	log *slog.Logger
}

// NewWaveHandler creates a new WaveHandler.
func NewWaveHandler(svc *service.WaveService, log *slog.Logger) *WaveHandler {
	return &WaveHandler{svc: svc, log: log}
}

// ── Response Types ─────────────────────────────────────────────────────────────────────

// waveResponse is the JSON shape returned for wave endpoints.
type waveResponse struct {
	ID          string   `json:"id"`
	WaveNo      string   `json:"wave_no"`
	WarehouseID string   `json:"warehouse_id"`
	WaveType    string   `json:"wave_type"`
	Status      string   `json:"status"`
	OrderIDs    []string `json:"order_ids"`
	TaskIDs     []string `json:"task_ids"`
	TotalOrders int      `json:"total_orders"`
	TotalLines  int      `json:"total_lines"`
	TotalQty    float64  `json:"total_qty"`
	CreatedAt   string   `json:"created_at"`
	ReleasedAt  string   `json:"released_at,omitempty"`
	CompletedAt string   `json:"completed_at,omitempty"`
}

func toWaveResponse(w *domain.Wave) waveResponse {
	r := waveResponse{
		ID:          w.ID.String(),
		WaveNo:      w.WaveNo,
		WarehouseID: w.WarehouseID.String(),
		WaveType:    string(w.WaveType),
		Status:      string(w.Status),
		TotalOrders: w.TotalOrders,
		TotalLines:  w.TotalLines,
		TotalQty:    w.TotalQty,
		CreatedAt:   w.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	r.OrderIDs = make([]string, len(w.OrderIDs))
	for i, id := range w.OrderIDs {
		r.OrderIDs[i] = id.String()
	}
	r.TaskIDs = make([]string, len(w.TaskIDs))
	for i, id := range w.TaskIDs {
		r.TaskIDs[i] = id.String()
	}

	if w.ReleasedAt != nil {
		r.ReleasedAt = w.ReleasedAt.Format("2006-01-02T15:04:05Z")
	}
	if w.CompletedAt != nil {
		r.CompletedAt = w.CompletedAt.Format("2006-01-02T15:04:05Z")
	}

	return r
}

// ── Wave Handlers ─────────────────────────────────────────────────────────────────────

// CreateWave handles POST /api/v1/waves
func (h *WaveHandler) CreateWave(w http.ResponseWriter, r *http.Request) {
	var input service.CreateWaveInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	wave, err := h.svc.CreateWave(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteCreated(w, toWaveResponse(wave))
}

// GetWave handles GET /api/v1/waves/{id}
func (h *WaveHandler) GetWave(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	wave, err := h.svc.GetWave(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toWaveResponse(wave))
}

// ListWaves handles GET /api/v1/waves
func (h *WaveHandler) ListWaves(w http.ResponseWriter, r *http.Request) {
	page, pageSize := PaginationParams(r)

	params := service.WaveQueryParams{
		Page:     page,
		PageSize: pageSize,
	}

	if raw := QueryParam(r, "warehouse_id", ""); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			WriteError(w, r, pkgerrors.NewInvalidInput("invalid warehouse_id UUID"))
			return
		}
		params.WarehouseID = id
	}
	if raw := QueryParam(r, "status", ""); raw != "" {
		params.Status = domain.WaveStatus(raw)
	}
	if raw := QueryParam(r, "wave_type", ""); raw != "" {
		params.WaveType = domain.WaveType(raw)
	}

	waves, total, err := h.svc.ListWaves(r.Context(), params)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]waveResponse, 0, len(waves))
	for _, wv := range waves {
		resp = append(resp, toWaveResponse(wv))
	}

	WriteJSON(w, http.StatusOK, ListResponse[waveResponse]{
		Data:       resp,
		Pagination: NewPaginationMeta(total, page, pageSize),
	})
}

// UpdateWaveStatus handles PUT /api/v1/waves/{id}/status
func (h *WaveHandler) UpdateWaveStatus(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.UpdateWaveStatusInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	wave, err := h.svc.UpdateWaveStatus(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toWaveResponse(wave))
}

// ReleaseWave handles POST /api/v1/waves/{id}/release
func (h *WaveHandler) ReleaseWave(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	wave, err := h.svc.ReleaseWave(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toWaveResponse(wave))
}

// AddWaveOrders handles POST /api/v1/waves/{id}/orders
func (h *WaveHandler) AddWaveOrders(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.AddWaveOrdersInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	wave, err := h.svc.AddWaveOrders(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toWaveResponse(wave))
}

// RemoveWaveOrders handles DELETE /api/v1/waves/{id}/orders
func (h *WaveHandler) RemoveWaveOrders(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.RemoveWaveOrdersInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	wave, err := h.svc.RemoveWaveOrders(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toWaveResponse(wave))
}
