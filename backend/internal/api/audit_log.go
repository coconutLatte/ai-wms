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

// AuditLogHandler handles HTTP requests for audit log resources (Admin only).
type AuditLogHandler struct {
	svc *service.AuditLogService
	log *slog.Logger
}

// NewAuditLogHandler creates a new AuditLogHandler.
func NewAuditLogHandler(svc *service.AuditLogService, log *slog.Logger) *AuditLogHandler {
	return &AuditLogHandler{svc: svc, log: log}
}

// ── Response Types ───────────────────────────────────────────────────────────

// auditLogResponse is the JSON shape returned for audit log endpoints.
type auditLogResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	Action     string `json:"action"`
	Resource   string `json:"resource"`
	ResourceID string `json:"resource_id"`
	Details    string `json:"details"`
	IPAddress  string `json:"ip_address"`
	CreatedAt  string `json:"created_at"`
}

func toAuditLogResponse(l *domain.AuditLog) auditLogResponse {
	return auditLogResponse{
		ID:         l.ID.String(),
		UserID:     l.UserID.String(),
		Username:   l.Username,
		Action:     l.Action,
		Resource:   l.Resource,
		ResourceID: l.ResourceID,
		Details:    l.Details,
		IPAddress:  l.IPAddress,
		CreatedAt:  l.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ── Handlers ─────────────────────────────────────────────────────────────────

// ListAuditLogs handles GET /api/v1/audit-logs
func (h *AuditLogHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	page, pageSize := PaginationParams(r)
	offset := paginationOffset(page, pageSize)

	filter := repository.AuditLogFilter{
		Limit:  pageSize,
		Offset: offset,
	}

	if raw := QueryParam(r, "user_id", ""); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			WriteError(w, r, pkgerrors.NewInvalidInput("invalid user_id UUID"))
			return
		}
		filter.UserID = id
	}
	if raw := QueryParam(r, "action", ""); raw != "" {
		filter.Action = raw
	}
	if raw := QueryParam(r, "resource", ""); raw != "" {
		filter.Resource = raw
	}
	if raw := QueryParam(r, "date_from", ""); raw != "" {
		filter.DateFrom = raw
	}
	if raw := QueryParam(r, "date_to", ""); raw != "" {
		filter.DateTo = raw
	}

	logs, total, err := h.svc.ListAuditLogs(r.Context(), filter)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]auditLogResponse, 0, len(logs))
	for _, l := range logs {
		resp = append(resp, toAuditLogResponse(l))
	}

	WriteJSON(w, http.StatusOK, ListResponse[auditLogResponse]{
		Data:       resp,
		Pagination: NewPaginationMeta(total, page, pageSize),
	})
}
