// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"net/http"
)

// RegisterAuthRoutes registers authentication API routes on the given mux.
// Auth routes handle login, token refresh, logout, and current user info.
func RegisterAuthRoutes(mux *http.ServeMux, h *AuthHandler) {
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
	mux.HandleFunc("GET /api/v1/auth/me", h.Me)
}

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

	// Location lookup by barcode
	mux.HandleFunc("GET /api/v1/locations", h.GetLocationByBarcode)
}

// RegisterSKURoutes registers SKU API routes on the given mux.
func RegisterSKURoutes(mux *http.ServeMux, h *SKUHandler) {
	mux.HandleFunc("POST /api/v1/skus", h.CreateSKU)
	mux.HandleFunc("GET /api/v1/skus", h.ListSKUs)
	mux.HandleFunc("GET /api/v1/skus/{id}", h.GetSKU)
	mux.HandleFunc("PUT /api/v1/skus/{id}", h.UpdateSKU)
}

// RegisterInventoryRoutes registers inventory API routes on the given mux.
func RegisterInventoryRoutes(mux *http.ServeMux, h *InventoryHandler) {
	mux.HandleFunc("GET /api/v1/inventory", h.QueryInventory)
	mux.HandleFunc("GET /api/v1/inventory/{id}", h.GetInventory)
	mux.HandleFunc("POST /api/v1/inventory/{id}/adjust", h.AdjustInventory)
	mux.HandleFunc("GET /api/v1/inventory/{id}/transactions", h.GetTransactions)

	// FEFO / FIFO retrieval strategies
	mux.HandleFunc("GET /api/v1/inventory/fifo", h.GetOldestInventory)
	mux.HandleFunc("GET /api/v1/inventory/fefo", h.GetExpiringInventory)
	mux.HandleFunc("GET /api/v1/inventory/dashboard", h.GetDashboard)

	// Inventory status transition
	mux.HandleFunc("PATCH /api/v1/inventory/{id}/status", h.UpdateInventoryStatus)
}

// RegisterOrderRoutes registers order API routes on the given mux.
func RegisterOrderRoutes(mux *http.ServeMux, h *OrderHandler) {
	mux.HandleFunc("POST /api/v1/orders", h.CreateOrder)
	mux.HandleFunc("GET /api/v1/orders", h.ListOrders)
	mux.HandleFunc("GET /api/v1/orders/{id}", h.GetOrder)
	mux.HandleFunc("PUT /api/v1/orders/{id}/status", h.UpdateOrderStatus)
	mux.HandleFunc("POST /api/v1/orders/{id}/lines", h.AddOrderLine)

	// Order line status transition
	mux.HandleFunc("PUT /api/v1/orders/{id}/lines/{lineId}/status", h.UpdateOrderLineStatus)
}

// RegisterASNRoutes registers ASN API routes on the given mux.
func RegisterASNRoutes(mux *http.ServeMux, h *OrderHandler) {
	mux.HandleFunc("POST /api/v1/asns", h.CreateASN)
	mux.HandleFunc("GET /api/v1/asns", h.ListASNs)
	mux.HandleFunc("GET /api/v1/asns/{id}", h.GetASN)
	mux.HandleFunc("PUT /api/v1/asns/{id}/status", h.UpdateASNStatus)
}

// RegisterAuditLogRoutes registers audit log API routes on the given mux (Admin only).
func RegisterAuditLogRoutes(mux *http.ServeMux, h *AuditLogHandler) {
	mux.HandleFunc("GET /api/v1/audit-logs", h.ListAuditLogs)
}

// RegisterUserRoutes registers user API routes on the given mux (Admin only).
func RegisterUserRoutes(mux *http.ServeMux, h *UserHandler) {
	mux.HandleFunc("POST /api/v1/users", h.CreateUser)
	mux.HandleFunc("GET /api/v1/users", h.ListUsers)
	mux.HandleFunc("GET /api/v1/users/{id}", h.GetUser)
	mux.HandleFunc("PUT /api/v1/users/{id}", h.UpdateUser)
	mux.HandleFunc("PUT /api/v1/users/{id}/status", h.UpdateUserStatus)
}

// RegisterRoleRoutes registers role API routes on the given mux (Admin only).
func RegisterRoleRoutes(mux *http.ServeMux, h *RoleHandler) {
	mux.HandleFunc("POST /api/v1/roles", h.CreateRole)
	mux.HandleFunc("GET /api/v1/roles", h.ListRoles)
	mux.HandleFunc("GET /api/v1/roles/{id}", h.GetRole)
	mux.HandleFunc("PUT /api/v1/roles/{id}", h.UpdateRole)
	mux.HandleFunc("DELETE /api/v1/roles/{id}", h.DeleteRole)
}

// RegisterTaskRoutes registers task API routes on the given mux.
// Tasks are accessible from both Admin (management) and PDA (execution).
func RegisterTaskRoutes(mux *http.ServeMux, h *TaskHandler) {
	mux.HandleFunc("POST /api/v1/tasks", h.CreateTask)
	mux.HandleFunc("GET /api/v1/tasks", h.ListTasks)
	mux.HandleFunc("GET /api/v1/tasks/{id}", h.GetTask)
	mux.HandleFunc("POST /api/v1/tasks/{id}/assign", h.AssignTask)
	mux.HandleFunc("PUT /api/v1/tasks/{id}/status", h.UpdateTaskStatus)
	mux.HandleFunc("POST /api/v1/tasks/{id}/complete", h.CompleteTask)
}

// RegisterWaveRoutes registers wave API routes on the given mux.
func RegisterWaveRoutes(mux *http.ServeMux, h *WaveHandler) {
	mux.HandleFunc("POST /api/v1/waves", h.CreateWave)
	mux.HandleFunc("GET /api/v1/waves", h.ListWaves)
	mux.HandleFunc("GET /api/v1/waves/{id}", h.GetWave)
	mux.HandleFunc("PUT /api/v1/waves/{id}/status", h.UpdateWaveStatus)
	mux.HandleFunc("POST /api/v1/waves/{id}/release", h.ReleaseWave)
	mux.HandleFunc("POST /api/v1/waves/{id}/orders", h.AddWaveOrders)
	mux.HandleFunc("DELETE /api/v1/waves/{id}/orders", h.RemoveWaveOrders)
}

// RegisterDashboardRoute registers the admin dashboard API route.
func RegisterDashboardRoute(mux *http.ServeMux, h *DashboardHandler) {
	mux.HandleFunc("GET /api/v1/dashboard", h.GetDashboard)
}
