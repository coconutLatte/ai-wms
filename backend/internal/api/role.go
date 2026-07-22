// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"log/slog"
	"net/http"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/service"
)

// RoleHandler handles HTTP requests for role resources (Admin only).
type RoleHandler struct {
	svc *service.RoleService
	log *slog.Logger
}

// NewRoleHandler creates a new RoleHandler.
func NewRoleHandler(svc *service.RoleService, log *slog.Logger) *RoleHandler {
	return &RoleHandler{svc: svc, log: log}
}

// ── Response Types ───────────────────────────────────────────────────────────

// roleResponse is the JSON shape returned for role endpoints.
type roleResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Permissions []domain.Permission `json:"permissions"`
	CreatedAt   string             `json:"created_at"`
}

func toRoleResponse(r *domain.Role) roleResponse {
	perms := r.Permissions
	if perms == nil {
		perms = []domain.Permission{}
	}

	return roleResponse{
		ID:          r.ID.String(),
		Name:        r.Name,
		Description: r.Description,
		Permissions: perms,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ── Handlers ─────────────────────────────────────────────────────────────────

// ListRoles handles GET /api/v1/roles
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, total, err := h.svc.ListRoles(r.Context())
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]roleResponse, 0, len(roles))
	for _, role := range roles {
		resp = append(resp, toRoleResponse(role))
	}

	page, pageSize := PaginationParams(r)
	WriteJSON(w, http.StatusOK, ListResponse[roleResponse]{
		Data:       resp,
		Pagination: NewPaginationMeta(total, page, pageSize),
	})
}

// CreateRole handles POST /api/v1/roles
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var input service.CreateRoleInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	role, err := h.svc.CreateRole(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	h.log.Info("role created",
		slog.String("role_id", role.ID.String()),
		slog.String("name", role.Name),
	)

	WriteCreated(w, toRoleResponse(role))
}

// GetRole handles GET /api/v1/roles/{id}
func (h *RoleHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	role, err := h.svc.GetRole(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toRoleResponse(role))
}

// UpdateRole handles PUT /api/v1/roles/{id}
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.UpdateRoleInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	role, err := h.svc.UpdateRole(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	h.log.Info("role updated",
		slog.String("role_id", role.ID.String()),
		slog.String("name", role.Name),
	)

	WriteJSON(w, http.StatusOK, toRoleResponse(role))
}

// DeleteRole handles DELETE /api/v1/roles/{id}
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	if err := h.svc.DeleteRole(r.Context(), id); err != nil {
		WriteError(w, r, err)
		return
	}

	h.log.Info("role deleted",
		slog.String("role_id", id.String()),
	)

	WriteNoContent(w)
}
