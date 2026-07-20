// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"net/http"
)

// RegisterWarehouseRoutes registers warehouse, zone, and location API routes on the given mux.
// Uses Go 1.22+ enhanced routing with method patterns and path parameters.
func RegisterWarehouseRoutes(mux *http.ServeMux, h *WarehouseHandler) {
	// Warehouse
	mux.HandleFunc("POST /api/v1/warehouses", h.CreateWarehouse)
	mux.HandleFunc("GET /api/v1/warehouses", h.ListWarehouses)
	mux.HandleFunc("GET /api/v1/warehouses/{id}", h.GetWarehouse)
	mux.HandleFunc("PUT /api/v1/warehouses/{id}", h.UpdateWarehouse)

	// Zone (nested under warehouse)
	mux.HandleFunc("POST /api/v1/warehouses/{id}/zones", h.CreateZone)
	mux.HandleFunc("GET /api/v1/warehouses/{id}/zones", h.ListZones)
	mux.HandleFunc("GET /api/v1/zones/{id}", h.GetZone)

	// Location (nested under zone)
	mux.HandleFunc("POST /api/v1/zones/{id}/locations", h.CreateLocation)
	mux.HandleFunc("GET /api/v1/zones/{id}/locations", h.ListLocations)
	mux.HandleFunc("GET /api/v1/locations/{id}", h.GetLocation)
	mux.HandleFunc("PATCH /api/v1/locations/{id}/status", h.UpdateLocationStatus)
}
