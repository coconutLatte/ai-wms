// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"

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
		ID:            inv.ID.String(),
		SKUID:         inv.SKUID.String(),
		LocationID:    inv.LocationID.String(),
		WarehouseID:   inv.WarehouseID.String(),
		BatchNo:       inv.BatchNo,
		Qty:           inv.Qty,
		ReservedQty:   inv.ReservedQty,
		AvailableQty:  inv.AvailableQty,
		Status:        string(inv.Status),
		ReceivedAt:    inv.ReceivedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     inv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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
	ID             string  `json:"id"`
	InventoryID    string  `json:"inventory_id"`
	SKUID          string  `json:"sku_id"`
	LocationID     string  `json:"location_id"`
	Type           string  `json:"type"`
	DeltaQty       float64 `json:"delta_qty"`
	ResultingQty   float64 `json:"resulting_qty"`
	ReferenceType  string  `json:"reference_type"`
	ReferenceID    string  `json:"reference_id"`
	CreatedAt      string  `json:"created_at"`
	CreatedBy      string  `json:"created_by"`
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
	Inventory inventoryResponse  `json:"inventory"`
	Transaction inventoryTxnResponse `json:"transaction"`
}

// ── Helpers ────────────────────────────────────────────────────────────────────────────

// queryParam returns a query string parameter value or the default.
func queryParam(r *http.Request, key, defaultVal string) string {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	return v
}

// queryParamInt returns a query string parameter as an int or the default.
func queryParamInt(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return defaultVal
	}
	return n
}

// ── Inventory Handlers ─────────────────────────────────────────────────────────────────

// QueryInventory handles GET /api/v1/inventory
func (h *InventoryHandler) QueryInventory(w http.ResponseWriter, r *http.Request) {
	input := service.QueryInventoryInput{
		WarehouseID: queryParam(r, "warehouse_id", ""),
		SKUID:       queryParam(r, "sku_id", ""),
		LocationID:  queryParam(r, "location_id", ""),
		BatchNo:     queryParam(r, "batch_no", ""),
		Status:      domain.InventoryStatus(queryParam(r, "status", "")),
		Limit:       queryParamInt(r, "limit", 0),
		Offset:      queryParamInt(r, "offset", 0),
	}

	results, err := h.svc.QueryInventory(r.Context(), input)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := make([]inventoryResponse, 0, len(results))
	for _, inv := range results {
		resp = append(resp, toInventoryResponse(inv))
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetInventory handles GET /api/v1/inventory/{id}
func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	id, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	inv, err := h.svc.GetInventory(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toInventoryResponse(inv))
}

// AdjustInventory handles POST /api/v1/inventory/{id}/adjust
func (h *InventoryHandler) AdjustInventory(w http.ResponseWriter, r *http.Request) {
	id, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input service.AdjustInventoryInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updated, err := h.svc.AdjustInventory(r.Context(), id, input)
	if err != nil {
		if pkgerrors.IsNotFound(err) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Fetch the most recent transaction to include in the response.
	// The service creates exactly one transaction per adjustment, so the most
	// recent one is the adjustment we just made.
	txs, err := h.svc.GetTransactions(r.Context(), id)
	if err != nil || len(txs) == 0 {
		// If we can't get transactions, still return inventory without tx info.
		writeJSON(w, http.StatusOK, map[string]any{
			"inventory": toInventoryResponse(updated),
		})
		return
	}

	writeJSON(w, http.StatusOK, adjustResponse{
		Inventory:   toInventoryResponse(updated),
		Transaction: toInventoryTxnResponse(txs[0]),
	})
}

// GetTransactions handles GET /api/v1/inventory/{id}/transactions
func (h *InventoryHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	id, err := pathUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	txs, err := h.svc.GetTransactions(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	resp := make([]inventoryTxnResponse, 0, len(txs))
	for _, tx := range txs {
		resp = append(resp, toInventoryTxnResponse(tx))
	}

	writeJSON(w, http.StatusOK, resp)
}
