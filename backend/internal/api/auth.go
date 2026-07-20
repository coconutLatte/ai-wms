// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"log/slog"
	"net/http"

	"github.com/ai-wms/ai-wms/backend/internal/service"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// AuthHandler handles authentication-related HTTP requests (login, refresh).
type AuthHandler struct {
	svc    *service.AuthService
	logger *slog.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc *service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, logger: logger}
}

// Login handles POST /api/v1/auth/login.
// Accepts { "username": "...", "password": "..." }, returns a token pair on success.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	// Basic validation.
	if input.Username == "" || input.Password == "" {
		WriteError(w, r, pkgerrors.NewInvalidInput("username and password are required"))
		return
	}

	pair, _, err := h.svc.Login(r.Context(), input)
	if err != nil {
		h.logger.Warn("login failed",
			slog.String("username", input.Username),
			slog.String("error", err.Error()),
		)
		WriteError(w, r, err)
		return
	}

	h.logger.Info("user logged in",
		slog.String("username", input.Username),
	)

	WriteJSON(w, http.StatusOK, pair)
}

// Refresh handles POST /api/v1/auth/refresh.
// Accepts { "refresh_token": "..." }, returns a new token pair on success.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var input service.RefreshInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	if input.RefreshToken == "" {
		WriteError(w, r, pkgerrors.NewInvalidInput("refresh_token is required"))
		return
	}

	pair, err := h.svc.RefreshToken(r.Context(), input)
	if err != nil {
		h.logger.Warn("token refresh failed",
			slog.String("error", err.Error()),
		)
		WriteError(w, r, err)
		return
	}

	h.logger.Info("token refreshed successfully")

	WriteJSON(w, http.StatusOK, pair)
}

// Me handles GET /api/v1/auth/me.
// Returns the current authenticated user's profile.
// The user must be populated in the context by the auth middleware.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// This endpoint requires the auth middleware, which sets user info in context.
	// For now it returns a simple acknowledgment — the full user profile endpoint
	// will be added when the user service is expanded.
	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "authenticated",
	})
}
