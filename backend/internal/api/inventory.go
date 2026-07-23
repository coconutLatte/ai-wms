// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"log/slog"
	"net/http"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/service"
)

// InventoryHandler handles HTTP requests for inventory resources.
type InventoryHandler struct {
	svc *service.InventoryService
	log *slog.Logger
}

// NewInventoryHandler creates a new InventoryHandler.
func NewInventoryHandler(svc *service.InventoryService, log *slog.Logger) *InventoryHandler {
	return &InventoryHandler{svc: svc, log: log}
}

// ── Response Types ─────────────────────────────────────────────────────────────────────

// inventoryResponse is the JSON shape returned for inventory endpoints.
type inventoryResponse struct {
	ID             string  `json:"id"`
	SKUID          string  `json:"sku_id"`
	LocationID     string  `json:"location_id"`
	WarehouseID    string  `json:"warehouse_id"`
	BatchNo        string  `json:"batch_no"`
	Qty            float64 `json:"qty"`
	ReservedQty    float64 `json:"reserved_qty"`
	AvailableQty   float64 `json:"available_qty"`
	Status         string  `json:"status"`
	ProductionDate string  `json:"production_date,omitempty"`
	ExpiryDate     string  `json:"expiry_date,omitempty"`
	ReceivedAt     string  `json:"received_at"`
	UpdatedAt      string  `json:"updated_at"`
}

func toInventoryResponse(inv *domain.Inventory) inventoryResponse {
	r := inventoryResponse{
		ID:           inv.ID.String(),
		SKUID:        inv.SKUID.String(),
		LocationID:   inv.LocationID.String(),
		WarehouseID:  inv.WarehouseID.String(),
		BatchNo:      inv.BatchNo,
		Qty:          inv.Qty,
		ReservedQty:  inv.ReservedQty,
		AvailableQty: inv.AvailableQty,
		Status:       string(inv.Status),
		ReceivedAt:   inv.ReceivedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    inv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if inv.ProductionDate != nil {
		r.ProductionDate = inv.ProductionDate.Format("2006-01-02T15:04:05Z")
	}
	if inv.ExpiryDate != nil {
		r.ExpiryDate = inv.ExpiryDate.Format("2006-01-02T15:04:05Z")
	}
	return r
}

// inventoryTxnResponse is the JSON shape returned for inventory transaction endpoints.
type inventoryTxnResponse struct {
	ID            string  `json:"id"`
	InventoryID   string  `json:"inventory_id"`
	SKUID         string  `json:"sku_id"`
	LocationID    string  `json:"location_id"`
	Type          string  `json:"type"`
	DeltaQty      float64 `json:"delta_qty"`
	ResultingQty  float64 `json:"resulting_qty"`
	ReferenceType string  `json:"reference_type"`
	ReferenceID   string  `json:"reference_id"`
	CreatedAt     string  `json:"created_at"`
	CreatedBy     string  `json:"created_by"`
}

func toInventoryTxnResponse(tx *domain.InventoryTransaction) inventoryTxnResponse {
	r := inventoryTxnResponse{
		ID:            tx.ID.String(),
		InventoryID:   tx.InventoryID.String(),
		SKUID:         tx.SKUID.String(),
		LocationID:    tx.LocationID.String(),
		Type:          string(tx.Type),
		DeltaQty:      tx.DeltaQty,
		ResultingQty:  tx.ResultingQty,
		ReferenceType: tx.ReferenceType,
		ReferenceID:   tx.ReferenceID.String(),
		CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
		CreatedBy:     tx.CreatedBy,
	}
	return r
}

// adjustResponse is the JSON shape returned after an inventory adjustment.
type adjustResponse struct {
	Inventory   inventoryResponse    `json:"inventory"`
	Transaction inventoryTxnResponse `json:"transaction"`
}

// ── Inventory Handlers ─────────────────────────────────────────────────────────────────

// QueryInventory handles GET /api/v1/inventory
func (h *InventoryHandler) QueryInventory(w http.ResponseWriter, r *http.Request) {
	page, pageSize := PaginationParams(r)
	offset := paginationOffset(page, pageSize)

	input := service.QueryInventoryInput{
		WarehouseID: QueryParam(r, "warehouse_id", ""),
		SKUID:       QueryParam(r, "sku_id", ""),
		LocationID:  QueryParam(r, "location_id", ""),
		BatchNo:     QueryParam(r, "batch_no", ""),
		Status:      domain.InventoryStatus(QueryParam(r, "status", "")),
		Limit:       pageSize,
		Offset:      offset,
	}

	results, total, err := h.svc.QueryInventory(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]inventoryResponse, 0, len(results))
	for _, inv := range results {
		resp = append(resp, toInventoryResponse(inv))
	}

	WriteJSON(w, http.StatusOK, ListResponse[inventoryResponse]{
		Data:       resp,
		Pagination: NewPaginationMeta(total, page, pageSize),
	})
}

// GetInventory handles GET /api/v1/inventory/{id}
func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	inv, err := h.svc.GetInventory(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toInventoryResponse(inv))
}

// AdjustInventory handles POST /api/v1/inventory/{id}/adjust
func (h *InventoryHandler) AdjustInventory(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.AdjustInventoryInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	updated, err := h.svc.AdjustInventory(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	// Fetch the most recent transaction to include in the response.
	txs, _, err := h.svc.GetTransactions(r.Context(), id, 0, 0)
	if err != nil || len(txs) == 0 {
		WriteJSON(w, http.StatusOK, map[string]any{
			"inventory": toInventoryResponse(updated),
		})
		return
	}

	WriteJSON(w, http.StatusOK, adjustResponse{
		Inventory:   toInventoryResponse(updated),
		Transaction: toInventoryTxnResponse(txs[0]),
	})
}

// GetTransactions handles GET /api/v1/inventory/{id}/transactions
func (h *InventoryHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	page, pageSize := PaginationParams(r)
	offset := paginationOffset(page, pageSize)

	txs, total, err := h.svc.GetTransactions(r.Context(), id, pageSize, offset)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]inventoryTxnResponse, 0, len(txs))
	for _, tx := range txs {
		resp = append(resp, toInventoryTxnResponse(tx))
	}

	WriteJSON(w, http.StatusOK, ListResponse[inventoryTxnResponse]{
		Data:       resp,
		Pagination: NewPaginationMeta(total, page, pageSize),
	})
}

// QueryTransactions handles GET /api/v1/inventory-transactions
// Returns global transaction history with optional filters (SKU, warehouse, type, date range).
func (h *InventoryHandler) QueryTransactions(w http.ResponseWriter, r *http.Request) {
	page, pageSize := PaginationParams(r)
	offset := paginationOffset(page, pageSize)

	input := service.QueryTransactionsInput{
		SKUID:       QueryParam(r, "sku_id", ""),
		WarehouseID: QueryParam(r, "warehouse_id", ""),
		TxType:      domain.InventoryTxType(QueryParam(r, "type", "")),
		DateFrom:    QueryParam(r, "date_from", ""),
		DateTo:      QueryParam(r, "date_to", ""),
		Limit:       pageSize,
		Offset:      offset,
	}

	txs, total, err := h.svc.QueryTransactions(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]inventoryTxnResponse, 0, len(txs))
	for _, tx := range txs {
		resp = append(resp, toInventoryTxnResponse(tx))
	}

	WriteJSON(w, http.StatusOK, ListResponse[inventoryTxnResponse]{
		Data:       resp,
		Pagination: NewPaginationMeta(total, page, pageSize),
	})
}

// GetOldestInventory handles GET /api/v1/inventory/fifo
// Returns available inventory sorted by received_at ASC (oldest first — FIFO).
func (h *InventoryHandler) GetOldestInventory(w http.ResponseWriter, r *http.Request) {
	input := service.InventoryRetrievalInput{
		WarehouseID: QueryParam(r, "warehouse_id", ""),
		SKUID:       QueryParam(r, "sku_id", ""),
		Limit:       QueryParamInt(r, "limit", 0),
	}

	results, err := h.svc.GetOldestInventory(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]inventoryResponse, 0, len(results))
	for _, inv := range results {
		resp = append(resp, toInventoryResponse(inv))
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"data":  resp,
		"count": len(results),
	})
}

// GetExpiringInventory handles GET /api/v1/inventory/fefo
// Returns available inventory sorted by expiry_date ASC NULLS LAST (earliest expiring first — FEFO).
func (h *InventoryHandler) GetExpiringInventory(w http.ResponseWriter, r *http.Request) {
	input := service.InventoryRetrievalInput{
		WarehouseID: QueryParam(r, "warehouse_id", ""),
		SKUID:       QueryParam(r, "sku_id", ""),
		Limit:       QueryParamInt(r, "limit", 0),
	}

	results, err := h.svc.GetExpiringInventory(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]inventoryResponse, 0, len(results))
	for _, inv := range results {
		resp = append(resp, toInventoryResponse(inv))
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"data":  resp,
		"count": len(results),
	})
}

// ── Inventory Status Handler ──────────────────────────────────────────────────────────

// UpdateInventoryStatus handles PATCH /api/v1/inventory/{id}/status
func (h *InventoryHandler) UpdateInventoryStatus(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.UpdateInventoryStatusInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	inv, err := h.svc.UpdateInventoryStatus(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toInventoryResponse(inv))
}

// ── Dashboard Response Types ────────────────────────────────────────────────────────────

// dashboardStatsResponse mirrors repository.InventoryDashboardStats for the JSON API.
type dashboardStatsResponse struct {
	TotalRecords      int     `json:"total_records"`
	TotalQty          float64 `json:"total_qty"`
	TotalReservedQty  float64 `json:"total_reserved_qty"`
	TotalAvailableQty float64 `json:"total_available_qty"`
	AvailableCount    int     `json:"available_count"`
	QuarantineCount   int     `json:"quarantine_count"`
	DamagedCount      int     `json:"damaged_count"`
	ExpiredCount      int     `json:"expired_count"`
	LowStockCount     int     `json:"low_stock_count"`
}

// warehouseBreakdownResponse mirrors repository.InventoryByWarehouseRow.
type warehouseBreakdownResponse struct {
	WarehouseID   string  `json:"warehouse_id"`
	WarehouseName string  `json:"warehouse_name"`
	WarehouseCode string  `json:"warehouse_code"`
	TotalQty      float64 `json:"total_qty"`
	ReservedQty   float64 `json:"reserved_qty"`
	AvailableQty  float64 `json:"available_qty"`
	RecordCount   int     `json:"record_count"`
}

// dashboardResponse is the full JSON shape for GET /api/v1/inventory/dashboard
type dashboardResponse struct {
	Stats       *dashboardStatsResponse      `json:"stats"`
	LowStock    []inventoryResponse          `json:"low_stock"`
	ByWarehouse []warehouseBreakdownResponse `json:"by_warehouse"`
}

// GetDashboard handles GET /api/v1/inventory/dashboard
func (h *InventoryHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	input := service.DashboardInput{
		WarehouseID:       QueryParam(r, "warehouse_id", ""),
		LowStockThreshold: float64(QueryParamInt(r, "low_stock_threshold", 10)),
	}

	stats, lowStock, byWarehouse, err := h.svc.GetDashboardStats(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	// Convert low stock inventory
	lowStockResp := make([]inventoryResponse, 0, len(lowStock))
	for _, inv := range lowStock {
		lowStockResp = append(lowStockResp, toInventoryResponse(inv))
	}

	// Convert warehouse breakdown
	whResp := make([]warehouseBreakdownResponse, 0, len(byWarehouse))
	for _, w := range byWarehouse {
		whResp = append(whResp, warehouseBreakdownResponse{
			WarehouseID:   w.WarehouseID.String(),
			WarehouseName: w.WarehouseName,
			WarehouseCode: w.WarehouseCode,
			TotalQty:      w.TotalQty,
			ReservedQty:   w.ReservedQty,
			AvailableQty:  w.AvailableQty,
			RecordCount:   w.RecordCount,
		})
	}

	WriteJSON(w, http.StatusOK, dashboardResponse{
		Stats: &dashboardStatsResponse{
			TotalRecords:      stats.TotalRecords,
			TotalQty:          stats.TotalQty,
			TotalReservedQty:  stats.TotalReservedQty,
			TotalAvailableQty: stats.TotalAvailableQty,
			AvailableCount:    stats.AvailableCount,
			QuarantineCount:   stats.QuarantineCount,
			DamagedCount:      stats.DamagedCount,
			ExpiredCount:      stats.ExpiredCount,
			LowStockCount:     stats.LowStockCount,
		},
		LowStock:    lowStockResp,
		ByWarehouse: whResp,
	})
}
