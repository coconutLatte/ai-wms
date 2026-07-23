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

// CycleCountHandler handles HTTP requests for cycle count resources (used by both Admin and PDA).
type CycleCountHandler struct {
	svc *service.CycleCountService
	log *slog.Logger
}

// NewCycleCountHandler creates a new CycleCountHandler.
func NewCycleCountHandler(svc *service.CycleCountService, log *slog.Logger) *CycleCountHandler {
	return &CycleCountHandler{svc: svc, log: log}
}

// ── Response Types ─────────────────────────────────────────────────────────────

type cycleCountResponse struct {
	ID           string  `json:"id"`
	CountNo      string  `json:"count_no"`
	WarehouseID  string  `json:"warehouse_id"`
	LocationID   string  `json:"location_id,omitempty"`
	ZoneID       string  `json:"zone_id,omitempty"`
	Status       string  `json:"status"`
	CountedBy    string  `json:"counted_by,omitempty"`
	Notes        string  `json:"notes,omitempty"`
	TotalLines   int     `json:"total_lines"`
	MatchedLines int     `json:"matched_lines"`
	CreatedAt    string  `json:"created_at"`
	StartedAt    string  `json:"started_at,omitempty"`
	CompletedAt  string  `json:"completed_at,omitempty"`
	ApprovedAt   string  `json:"approved_at,omitempty"`
	ApprovedBy   string  `json:"approved_by,omitempty"`
	Lines        []cycleCountLineResponse `json:"lines,omitempty"`
}

type cycleCountLineResponse struct {
	ID          string  `json:"id"`
	SKUID       string  `json:"sku_id"`
	LocationID  string  `json:"location_id"`
	BatchNo     string  `json:"batch_no,omitempty"`
	SystemQty   float64 `json:"system_qty"`
	CountedQty  *float64 `json:"counted_qty,omitempty"`
	Variance    *float64 `json:"variance,omitempty"`
	Status      string  `json:"status"`
	CountedAt   string  `json:"counted_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

func toCycleCountResponse(cc *domain.CycleCount, lines []*domain.CycleCountLine) cycleCountResponse {
	r := cycleCountResponse{
		ID:            cc.ID.String(),
		CountNo:       cc.CountNo,
		WarehouseID:   cc.WarehouseID.String(),
		Status:        string(cc.Status),
		CountedBy:     cc.CountedBy,
		Notes:         cc.Notes,
		TotalLines:    cc.TotalLines,
		MatchedLines:  cc.MatchedLines,
		CreatedAt:     cc.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if cc.LocationID != nil {
		r.LocationID = cc.LocationID.String()
	}
	if cc.ZoneID != nil {
		r.ZoneID = cc.ZoneID.String()
	}
	if cc.StartedAt != nil {
		r.StartedAt = cc.StartedAt.Format("2006-01-02T15:04:05Z")
	}
	if cc.CompletedAt != nil {
		r.CompletedAt = cc.CompletedAt.Format("2006-01-02T15:04:05Z")
	}
	if cc.ApprovedAt != nil {
		r.ApprovedAt = cc.ApprovedAt.Format("2006-01-02T15:04:05Z")
	}
	if cc.ApprovedBy != "" {
		r.ApprovedBy = cc.ApprovedBy
	}
	if lines != nil {
		r.Lines = make([]cycleCountLineResponse, len(lines))
		for i, l := range lines {
			r.Lines[i] = cycleCountLineResponse{
				ID:         l.ID.String(),
				SKUID:      l.SKUID.String(),
				LocationID: l.LocationID.String(),
				BatchNo:    l.BatchNo,
				SystemQty:  l.SystemQty,
				CountedQty: l.CountedQty,
				Variance:   l.Variance,
				Status:     string(l.Status),
				CreatedAt:  l.CreatedAt.Format("2006-01-02T15:04:05Z"),
			}
			if l.CountedAt != nil {
				r.Lines[i].CountedAt = l.CountedAt.Format("2006-01-02T15:04:05Z")
			}
		}
	}
	return r
}

// ── Handlers ────────────────────────────────────────────────────────────────────

// StartCycleCount handles POST /api/v1/cycle-counts
// Creates a new cycle count with lines for each inventory record in the specified location/zone.
func (h *CycleCountHandler) StartCycleCount(w http.ResponseWriter, r *http.Request) {
	var input service.StartCycleCountInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	cc, lines, err := h.svc.StartCycleCount(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteCreated(w, toCycleCountResponse(cc, lines))
}

// GetCycleCount handles GET /api/v1/cycle-counts/{id}
func (h *CycleCountHandler) GetCycleCount(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	cc, lines, err := h.svc.GetCycleCount(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toCycleCountResponse(cc, lines))
}

// ListCycleCounts handles GET /api/v1/cycle-counts
func (h *CycleCountHandler) ListCycleCounts(w http.ResponseWriter, r *http.Request) {
	page, pageSize := PaginationParams(r)
	offset := paginationOffset(page, pageSize)

	filter := repository.CycleCountFilter{
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
	if raw := QueryParam(r, "status", ""); raw != "" {
		filter.Status = raw
	}

	counts, total, err := h.svc.ListCycleCounts(r.Context(), filter)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]cycleCountResponse, 0, len(counts))
	for _, cc := range counts {
		resp = append(resp, toCycleCountResponse(cc, nil))
	}

	WriteJSON(w, http.StatusOK, ListResponse[cycleCountResponse]{
		Data:       resp,
		Pagination: NewPaginationMeta(total, page, pageSize),
	})
}

// SubmitLine handles POST /api/v1/cycle-counts/{id}/lines
// Submits a counted quantity for a single line.
func (h *CycleCountHandler) SubmitLine(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.SubmitLineInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	line, err := h.svc.SubmitLine(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, line)
}

// FinalizeCount handles POST /api/v1/cycle-counts/{id}/finalize
// Completes the counting phase and moves to pending_review.
func (h *CycleCountHandler) FinalizeCount(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.FinalizeCountInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	cc, err := h.svc.FinalizeCount(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toCycleCountResponse(cc, nil))
}

// ApproveCount handles PUT /api/v1/cycle-counts/{id}/approve
// Approves or adjusts the cycle count.
func (h *CycleCountHandler) ApproveCount(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.ApproveCountInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	cc, err := h.svc.ApproveCount(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toCycleCountResponse(cc, nil))
}

// CancelCycleCount handles PUT /api/v1/cycle-counts/{id}/cancel
func (h *CycleCountHandler) CancelCycleCount(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	cc, err := h.svc.CancelCycleCount(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toCycleCountResponse(cc, nil))
}
