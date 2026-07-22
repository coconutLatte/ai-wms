// Package api provides HTTP handlers and route registration for the WMS API.
// Handlers are thin — they parse requests, call services, and return responses.
package api

import (
	"log/slog"
	"net/http"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/service"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
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
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	wh, err := h.svc.CreateWarehouse(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteCreated(w, toWarehouseResponse(wh))
}

// GetWarehouse handles GET /api/v1/warehouses/{id}
func (h *WarehouseHandler) GetWarehouse(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	wh, err := h.svc.GetWarehouse(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toWarehouseResponse(wh))
}

// ListWarehouses handles GET /api/v1/warehouses
func (h *WarehouseHandler) ListWarehouses(w http.ResponseWriter, r *http.Request) {
	page, pageSize := PaginationParams(r)
	offset := paginationOffset(page, pageSize)

	warehouses, total, err := h.svc.ListWarehouses(r.Context(), pageSize, offset)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]warehouseResponse, 0, len(warehouses))
	for _, wh := range warehouses {
		resp = append(resp, toWarehouseResponse(wh))
	}

	WriteJSON(w, http.StatusOK, ListResponse[warehouseResponse]{
		Data:       resp,
		Pagination: NewPaginationMeta(total, page, pageSize),
	})
}

// UpdateWarehouse handles PUT /api/v1/warehouses/{id}
func (h *WarehouseHandler) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.UpdateWarehouseInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	wh, err := h.svc.UpdateWarehouse(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toWarehouseResponse(wh))
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
	warehouseID, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.CreateZoneInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	z, err := h.svc.CreateZone(r.Context(), warehouseID, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteCreated(w, toZoneResponse(z))
}

// GetZone handles GET /api/v1/zones/{id}
func (h *WarehouseHandler) GetZone(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	z, err := h.svc.GetZone(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toZoneResponse(z))
}

// ListZones handles GET /api/v1/warehouses/{id}/zones
func (h *WarehouseHandler) ListZones(w http.ResponseWriter, r *http.Request) {
	warehouseID, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	page, pageSize := PaginationParams(r)
	offset := paginationOffset(page, pageSize)

	zones, total, err := h.svc.ListZones(r.Context(), warehouseID, pageSize, offset)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]zoneResponse, 0, len(zones))
	for _, z := range zones {
		resp = append(resp, toZoneResponse(z))
	}

	WriteJSON(w, http.StatusOK, ListResponse[zoneResponse]{
		Data:       resp,
		Pagination: NewPaginationMeta(total, page, pageSize),
	})
}

// ── Location Handlers ───────────────────────────────────────────────────────────────────

// locationResponse is the JSON shape returned for location endpoints.
type locationResponse struct {
	ID           string           `json:"id"`
	ZoneID       string           `json:"zone_id"`
	WarehouseID  string           `json:"warehouse_id"`
	Code         string           `json:"code"`
	Barcode      string           `json:"barcode"`
	LocationType string           `json:"location_type"`
	Capacity     *domain.Capacity `json:"capacity,omitempty"`
	Status       string           `json:"status"`
	CreatedAt    string           `json:"created_at"`
	UpdatedAt    string           `json:"updated_at"`
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
	zoneID, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.CreateLocationInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	loc, err := h.svc.CreateLocation(r.Context(), zoneID, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteCreated(w, toLocationResponse(loc))
}

// GetLocation handles GET /api/v1/locations/{id}
func (h *WarehouseHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	loc, err := h.svc.GetLocation(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toLocationResponse(loc))
}

// ListLocations handles GET /api/v1/zones/{id}/locations
func (h *WarehouseHandler) ListLocations(w http.ResponseWriter, r *http.Request) {
	zoneID, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	page, pageSize := PaginationParams(r)
	offset := paginationOffset(page, pageSize)

	locs, total, err := h.svc.ListLocations(r.Context(), zoneID, pageSize, offset)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]locationResponse, 0, len(locs))
	for _, loc := range locs {
		resp = append(resp, toLocationResponse(loc))
	}

	WriteJSON(w, http.StatusOK, ListResponse[locationResponse]{
		Data:       resp,
		Pagination: NewPaginationMeta(total, page, pageSize),
	})
}

// UpdateLocationStatus handles PATCH /api/v1/locations/{id}/status
func (h *WarehouseHandler) UpdateLocationStatus(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := ReadJSON(r, &body); err != nil {
		WriteError(w, r, err)
		return
	}

	if err := h.svc.UpdateLocationStatus(r.Context(), id, domain.LocationStatus(body.Status)); err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetLocationByBarcode handles GET /api/v1/locations?barcode=X
func (h *WarehouseHandler) GetLocationByBarcode(w http.ResponseWriter, r *http.Request) {
	barcode := QueryParam(r, "barcode", "")
	if barcode == "" {
		WriteError(w, r, pkgerrors.NewInvalidInput("barcode query parameter is required"))
		return
	}

	loc, err := h.svc.GetLocationByBarcode(r.Context(), barcode)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toLocationResponse(loc))
}
