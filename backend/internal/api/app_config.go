package api

import (
	"log/slog"
	"net/http"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/service"
)

// AppConfigHandler handles HTTP requests for system configuration.
type AppConfigHandler struct {
	svc *service.AppConfigService
	log *slog.Logger
}

// NewAppConfigHandler creates a new AppConfigHandler.
func NewAppConfigHandler(svc *service.AppConfigService, log *slog.Logger) *AppConfigHandler {
	return &AppConfigHandler{svc: svc, log: log}
}

// ── Response Types ──────────────────────────────────────────────────────────

type appConfigResponse struct {
	SiteName           string `json:"site_name"`
	DefaultWarehouseID string `json:"default_warehouse_id"`
	LowStockThreshold  int    `json:"low_stock_threshold"`
	DefaultPageSize    int    `json:"default_page_size"`
	JWTAccessTTL       int    `json:"jwt_access_ttl"`
	UpdatedAt          string `json:"updated_at,omitempty"`
}

func toAppConfigResponse(row *domain.AppConfigRow) appConfigResponse {
	return appConfigResponse{
		SiteName:           row.Config.SiteName,
		DefaultWarehouseID: row.Config.DefaultWarehouseID,
		LowStockThreshold:  row.Config.LowStockThreshold,
		DefaultPageSize:    row.Config.DefaultPageSize,
		JWTAccessTTL:       row.Config.JWTAccessTTL,
		UpdatedAt:          row.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ── Handlers ─────────────────────────────────────────────────────────────────

// GetConfig handles GET /api/v1/settings
func (h *AppConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	row, err := h.svc.GetConfig(r.Context())
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toAppConfigResponse(row))
}

// UpdateConfig handles PUT /api/v1/settings
func (h *AppConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var input service.UpdateConfigInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	row, err := h.svc.UpdateConfig(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toAppConfigResponse(row))
}

// RegisterAppConfigRoutes registers system configuration API routes.
// These are admin-only endpoints.
func RegisterAppConfigRoutes(mux *http.ServeMux, h *AppConfigHandler) {
	mux.HandleFunc("GET /api/v1/settings", h.GetConfig)
	mux.HandleFunc("PUT /api/v1/settings", h.UpdateConfig)
}
