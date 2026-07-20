// Package api provides HTTP handlers and route registration for the WMS API.
// Handlers are thin — they parse requests, call services, and return responses.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"

	"github.com/ai-wms/ai-wms/backend/internal/service"
)

// WarehouseHandler handles HTTP requests for warehouse, zone, and location resources.
type WarehouseHandler struct {
	svc *service.WarehouseService
	log *slog.Logger
}

// NewWarehouseHandler creates a new WarehouseHandler.
func NewWarehouseHandler(svc *service.WarehouseService, log *slog.Logger) *WarehouseHandler {
	return &WarehouseHandler{svc: svc, log: log}
}

// ── Helpers ──────────────────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func pathUUID(r *http.Request, key string) (uuid.UUID, error) {
	raw := r.PathValue(key)
	if raw == "" {
		return uuid.Nil, pkgerrors.NewInvalidInput("missing path parameter: " + key)
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, pkgerrors.NewInvalidInput("invalid UUID format for " + key)
	}
	return id, nil
}

// ── Warehouse Handlers ───────────────────────────────────────────────────────────────────

// warehouseResponse is the JSON shape returned for warehouse endpoints.
type warehouseResponse struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toWarehouseResponse(w *domain.Warehouse) warehouseResponse {
	return warehouseResponse{
		ID:        w.ID.String(),
		Code:      w.Code,
		Name:      w.Name,
		Address:   w.Address,
		Status:    string(w.Status),
		CreatedAt: w.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: w.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// CreateWarehouse handles POST /api/v1/warehouses
func (h *WarehouseHandler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	var input service.CreateWarehouseInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	wh, err := h.svc.CreateWarehouse(r.Context(), input)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toWarehouseResponse(wh))
}

// GetWarehouse handles GET /api/v1/warehouses/{id}
func (h *WarehouseHandler) GetWarehouse(w http.ResponseWriter, r *http.Request) {
	id, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	wh, err := h.svc.GetWarehouse(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toWarehouseResponse(wh))
}

// ListWarehouses handles GET /api/v1/warehouses
func (h *WarehouseHandler) ListWarehouses(w http.ResponseWriter, r *http.Request) {
	warehouses, err := h.svc.ListWarehouses(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]warehouseResponse, 0, len(warehouses))
	for _, wh := range warehouses {
		resp = append(resp, toWarehouseResponse(wh))
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdateWarehouse handles PUT /api/v1/warehouses/{id}
func (h *WarehouseHandler) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	id, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input service.UpdateWarehouseInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	wh, err := h.svc.UpdateWarehouse(r.Context(), id, input)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toWarehouseResponse(wh))
}

// ── Zone Handlers ───────────────────────────────────────────────────────────────────────

// zoneResponse is the JSON shape returned for zone endpoints.
type zoneResponse struct {
	ID          string `json:"id"`
	WarehouseID string `json:"warehouse_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	ZoneType    string `json:"zone_type"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toZoneResponse(z *domain.Zone) zoneResponse {
	return zoneResponse{
		ID:          z.ID.String(),
		WarehouseID: z.WarehouseID.String(),
		Code:        z.Code,
		Name:        z.Name,
		ZoneType:    string(z.ZoneType),
		Status:      string(z.Status),
		CreatedAt:   z.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   z.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// CreateZone handles POST /api/v1/warehouses/{id}/zones
func (h *WarehouseHandler) CreateZone(w http.ResponseWriter, r *http.Request) {
	warehouseID, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input service.CreateZoneInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	z, err := h.svc.CreateZone(r.Context(), warehouseID, input)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toZoneResponse(z))
}

// GetZone handles GET /api/v1/zones/{id}
func (h *WarehouseHandler) GetZone(w http.ResponseWriter, r *http.Request) {
	id, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	z, err := h.svc.GetZone(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toZoneResponse(z))
}

// ListZones handles GET /api/v1/warehouses/{id}/zones
func (h *WarehouseHandler) ListZones(w http.ResponseWriter, r *http.Request) {
	warehouseID, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	zones, err := h.svc.ListZones(r.Context(), warehouseID)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]zoneResponse, 0, len(zones))
	for _, z := range zones {
		resp = append(resp, toZoneResponse(z))
	}

	writeJSON(w, http.StatusOK, resp)
}

// ── Location Handlers ───────────────────────────────────────────────────────────────────

// locationResponse is the JSON shape returned for location endpoints.
type locationResponse struct {
	ID           string                `json:"id"`
	ZoneID       string                `json:"zone_id"`
	WarehouseID  string                `json:"warehouse_id"`
	Code         string                `json:"code"`
	Barcode      string                `json:"barcode"`
	LocationType string                `json:"location_type"`
	Capacity     *domain.Capacity       `json:"capacity,omitempty"`
	Status       string                `json:"status"`
	CreatedAt    string                `json:"created_at"`
	UpdatedAt    string                `json:"updated_at"`
}

func toLocationResponse(loc *domain.Location) locationResponse {
	resp := locationResponse{
		ID:           loc.ID.String(),
		ZoneID:       loc.ZoneID.String(),
		WarehouseID:  loc.WarehouseID.String(),
		Code:         loc.Code,
		Barcode:      loc.Barcode,
		LocationType: string(loc.LocationType),
		Capacity:     loc.Capacity,
		Status:       string(loc.Status),
		CreatedAt:    loc.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    loc.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	return resp
}

// CreateLocation handles POST /api/v1/zones/{id}/locations
func (h *WarehouseHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	zoneID, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input service.CreateLocationInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	loc, err := h.svc.CreateLocation(r.Context(), zoneID, input)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toLocationResponse(loc))
}

// GetLocation handles GET /api/v1/locations/{id}
func (h *WarehouseHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	id, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	loc, err := h.svc.GetLocation(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toLocationResponse(loc))
}

// ListLocations handles GET /api/v1/zones/{id}/locations
func (h *WarehouseHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
	zoneID, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	locs, err := h.svc.ListLocations(r.Context(), zoneID)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]locationResponse, 0, len(locs))
	for _, loc := range locs {
		resp = append(resp, toLocationResponse(loc))
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdateLocationStatus handles PATCH /api/v1/locations/{id}/status
func (h *WarehouseHandler) UpdateLocationStatus(w http.ResponseWriter, r *http.Request) {
	id, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.UpdateLocationStatus(r.Context(), id, domain.LocationStatus(body.Status)); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
