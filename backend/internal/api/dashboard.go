// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"log/slog"
	"net/http"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/service"
)

// DashboardHandler handles HTTP requests for the admin dashboard.
// The dashboard aggregates data from multiple domains to provide
// a high-level overview of warehouse operations.
type DashboardHandler struct {
	warehouseSvc *service.WarehouseService
	skuSvc       *service.SKUService
	inventorySvc *service.InventoryService
	orderSvc     *service.OrderService
	taskSvc      *service.TaskService
	log          *slog.Logger
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(
	warehouseSvc *service.WarehouseService,
	skuSvc *service.SKUService,
	inventorySvc *service.InventoryService,
	orderSvc *service.OrderService,
	taskSvc *service.TaskService,
	log *slog.Logger,
) *DashboardHandler {
	return &DashboardHandler{
		warehouseSvc: warehouseSvc,
		skuSvc:       skuSvc,
		inventorySvc: inventorySvc,
		orderSvc:     orderSvc,
		taskSvc:      taskSvc,
		log:          log,
	}
}

// ── Response Types ─────────────────────────────────────────────────────────────────────

// adminDashboardResponse is the JSON response for the admin dashboard.
type adminDashboardResponse struct {
	WarehouseCount int                     `json:"warehouse_count"`
	SKUCount       int                     `json:"sku_count"`
	InventoryStats *dashboardStatsResponse `json:"inventory_stats"`
	OrderSummary   map[string]int          `json:"order_summary"`
	TaskSummary    map[string]int          `json:"task_summary"`
}

// GetDashboard handles GET /api/v1/dashboard.
// Returns aggregated counts and summaries for the admin landing page.
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Warehouse count
	warehouseCount, err := h.warehouseSvc.CountWarehouses(ctx)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	// SKU count
	skuCount, err := h.skuSvc.CountSKUs(ctx)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	// Inventory stats
	stats, _, _, err := h.inventorySvc.GetDashboardStats(ctx, service.DashboardInput{
		LowStockThreshold: 10.0,
	})
	if err != nil {
		// Inventory might not have data yet; log but don't fail.
		h.log.Warn("dashboard: failed to get inventory stats", slog.String("error", err.Error()))
		stats = nil
	}

	// Order summary by status
	orderSummary, err := h.orderSvc.CountOrdersByStatus(ctx)
	if err != nil {
		h.log.Warn("dashboard: failed to get order summary", slog.String("error", err.Error()))
		orderSummary = nil
	}

	// Task summary by status
	taskSummary, err := h.taskSvc.CountTasksByStatus(ctx)
	if err != nil {
		h.log.Warn("dashboard: failed to get task summary", slog.String("error", err.Error()))
		taskSummary = nil
	}

	// Build response
	resp := adminDashboardResponse{
		WarehouseCount: warehouseCount,
		SKUCount:       skuCount,
		OrderSummary:   convertStatusMap(orderSummary),
		TaskSummary:    convertTaskStatusMap(taskSummary),
	}

	if stats != nil {
		resp.InventoryStats = &dashboardStatsResponse{
			TotalRecords:      stats.TotalRecords,
			TotalQty:          stats.TotalQty,
			TotalReservedQty:  stats.TotalReservedQty,
			TotalAvailableQty: stats.TotalAvailableQty,
			AvailableCount:    stats.AvailableCount,
			QuarantineCount:   stats.QuarantineCount,
			DamagedCount:      stats.DamagedCount,
			ExpiredCount:      stats.ExpiredCount,
			LowStockCount:     stats.LowStockCount,
		}
	}

	WriteJSON(w, http.StatusOK, resp)
}

// convertStatusMap converts domain.OrderStatus keys to string keys for JSON.
func convertStatusMap(m map[domain.OrderStatus]int) map[string]int {
	if m == nil {
		return nil
	}
	result := make(map[string]int, len(m))
	for k, v := range m {
		result[string(k)] = v
	}
	return result
}

// convertTaskStatusMap converts domain.TaskStatus keys to string keys for JSON.
func convertTaskStatusMap(m map[domain.TaskStatus]int) map[string]int {
	if m == nil {
		return nil
	}
	result := make(map[string]int, len(m))
	for k, v := range m {
		result[string(k)] = v
	}
	return result
}
