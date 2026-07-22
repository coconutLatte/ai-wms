// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"log/slog"
	"net/http"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	"github.com/ai-wms/ai-wms/backend/internal/service"
)

// UserHandler handles HTTP requests for user resources (Admin only).
type UserHandler struct {
	svc *service.UserService
	log *slog.Logger
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc *service.UserService, log *slog.Logger) *UserHandler {
	return &UserHandler{svc: svc, log: log}
}

// ── Response Types ───────────────────────────────────────────────────────────

// userResponse is the JSON shape returned for user endpoints.
// Password hash is never serialized.
type userResponse struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name"`
	RoleIDs     []string `json:"role_ids"`
	Status      string   `json:"status"`
	LastLogin   string   `json:"last_login,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

func toUserResponse(u *domain.User) userResponse {
	roleIDs := make([]string, 0, len(u.RoleIDs))
	for _, id := range u.RoleIDs {
		roleIDs = append(roleIDs, id.String())
	}

	lastLogin := ""
	if u.LastLogin != nil {
		lastLogin = u.LastLogin.Format("2006-01-02T15:04:05Z")
	}

	return userResponse{
		ID:          u.ID.String(),
		Username:    u.Username,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		RoleIDs:     roleIDs,
		Status:      string(u.Status),
		LastLogin:   lastLogin,
		CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ── Handlers ─────────────────────────────────────────────────────────────────

// ListUsers handles GET /api/v1/users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, pageSize := PaginationParams(r)
	offset := paginationOffset(page, pageSize)

	filter := repository.UserFilter{
		Limit:  pageSize,
		Offset: offset,
	}

	if raw := QueryParam(r, "status", ""); raw != "" {
		filter.Status = domain.UserStatus(raw)
	}

	users, total, err := h.svc.ListUsers(r.Context(), filter)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]userResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, toUserResponse(u))
	}

	WriteJSON(w, http.StatusOK, ListResponse[userResponse]{
		Data:       resp,
		Pagination: NewPaginationMeta(total, page, pageSize),
	})
}

// CreateUser handles POST /api/v1/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input service.CreateUserInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	user, err := h.svc.CreateUser(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	h.log.Info("user created",
		slog.String("user_id", user.ID.String()),
		slog.String("username", user.Username),
	)

	WriteCreated(w, toUserResponse(user))
}

// GetUser handles GET /api/v1/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	user, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toUserResponse(user))
}

// UpdateUser handles PUT /api/v1/users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.UpdateUserInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	user, err := h.svc.UpdateUser(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	h.log.Info("user updated",
		slog.String("user_id", user.ID.String()),
		slog.String("username", user.Username),
	)

	WriteJSON(w, http.StatusOK, toUserResponse(user))
}

// UpdateUserStatus handles PUT /api/v1/users/{id}/status
func (h *UserHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.UpdateUserStatusInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	user, err := h.svc.UpdateUserStatus(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	h.log.Info("user status updated",
		slog.String("user_id", user.ID.String()),
		slog.String("username", user.Username),
		slog.String("new_status", string(user.Status)),
	)

	WriteJSON(w, http.StatusOK, toUserResponse(user))
}
