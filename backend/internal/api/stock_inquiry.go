// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"log/slog"
	"net/http"

	"github.com/ai-wms/ai-wms/backend/internal/service"
)

// StockInquiryHandler handles PDA stock inquiry requests — scan a barcode and get
// current inventory levels at the resolved location or for the resolved SKU.
type StockInquiryHandler struct {
	inventorySvc *service.InventoryService
	warehouseSvc *service.WarehouseService
	skuSvc       *service.SKUService
	log          *slog.Logger
}

// NewStockInquiryHandler creates a new StockInquiryHandler.
func NewStockInquiryHandler(
	inventorySvc *service.InventoryService,
	warehouseSvc *service.WarehouseService,
	skuSvc *service.SKUService,
	log *slog.Logger,
) *StockInquiryHandler {
	return &StockInquiryHandler{
		inventorySvc: inventorySvc,
		warehouseSvc: warehouseSvc,
		skuSvc:       skuSvc,
		log:          log,
	}
}

// ── Response Types ─────────────────────────────────────────────────────────────────────

// stockInquiryLocation holds resolved location info displayed in the stock inquiry result.
type stockInquiryLocation struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	Barcode      string `json:"barcode"`
	LocationType string `json:"location_type"`
	Status       string `json:"status"`
}

// stockInquirySKU holds resolved SKU info displayed in the stock inquiry result.
type stockInquirySKU struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Barcode string `json:"barcode"`
	Status string `json:"status"`
}

// stockInquiryResponse is the top-level JSON shape for GET /api/v1/stock-inquiry.
type stockInquiryResponse struct {
	Barcode    string                  `json:"barcode"`
	EntityType string                  `json:"entity_type"` // "location" or "sku"
	Location   *stockInquiryLocation   `json:"location,omitempty"`
	SKU        *stockInquirySKU        `json:"sku,omitempty"`
	Inventory  []inventoryResponse     `json:"inventory"`
	TotalQty   float64                 `json:"total_qty"`
	TotalReserved float64              `json:"total_reserved"`
	TotalAvailable float64             `json:"total_available"`
}

// ── Handler ─────────────────────────────────────────────────────────────────────────────

// InquireStock handles GET /api/v1/stock-inquiry?barcode=X
// It resolves the barcode to a location or SKU, then queries inventory and returns
// a consolidated result with entity info, inventory list, and aggregated quantities.
func (h *StockInquiryHandler) InquireStock(w http.ResponseWriter, r *http.Request) {
	barcode := QueryParam(r, "barcode", "")
	if barcode == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "barcode query parameter is required",
		})
		return
	}

	ctx := r.Context()
	var resp stockInquiryResponse
	resp.Barcode = barcode

	// Step 1: Try to resolve as a location barcode.
	loc, locErr := h.warehouseSvc.GetLocationByBarcode(ctx, barcode)
	if locErr == nil && loc != nil {
		// Resolved as a location — query inventory at this location.
		resp.EntityType = "location"
		resp.Location = &stockInquiryLocation{
			ID:           loc.ID.String(),
			Code:         loc.Code,
			Barcode:      loc.Barcode,
			LocationType: string(loc.LocationType),
			Status:       string(loc.Status),
		}

		input := service.QueryInventoryInput{
			LocationID: loc.ID.String(),
			Limit:      100,
			Offset:     0,
		}
		inventory, _, err := h.inventorySvc.QueryInventory(ctx, input)
		if err != nil {
			h.log.Warn("stock inquiry: failed to query inventory by location",
				slog.String("barcode", barcode),
				slog.String("location_id", loc.ID.String()),
				slog.String("error", err.Error()),
			)
			WriteError(w, r, err)
			return
		}

		for _, inv := range inventory {
			resp.Inventory = append(resp.Inventory, toInventoryResponse(inv))
			resp.TotalQty += inv.Qty
			resp.TotalReserved += inv.ReservedQty
			resp.TotalAvailable += inv.AvailableQty
		}
		WriteJSON(w, http.StatusOK, resp)
		return
	}

	// Step 2: Try to resolve as a SKU code.
	sku, skuErr := h.skuSvc.GetSKUByCode(ctx, barcode)
	if skuErr == nil && sku != nil {
		// Resolved as a SKU — query inventory for this SKU across all locations.
		resp.EntityType = "sku"
		resp.SKU = &stockInquirySKU{
			ID:     sku.ID.String(),
			Code:   sku.Code,
			Name:   sku.Name,
			Barcode: sku.Barcode,
			Status: string(sku.Status),
		}

		input := service.QueryInventoryInput{
			SKUID:  sku.ID.String(),
			Limit:  100,
			Offset: 0,
		}
		inventory, _, err := h.inventorySvc.QueryInventory(ctx, input)
		if err != nil {
			h.log.Warn("stock inquiry: failed to query inventory by SKU",
				slog.String("barcode", barcode),
				slog.String("sku_id", sku.ID.String()),
				slog.String("error", err.Error()),
			)
			WriteError(w, r, err)
			return
		}

		for _, inv := range inventory {
			resp.Inventory = append(resp.Inventory, toInventoryResponse(inv))
			resp.TotalQty += inv.Qty
			resp.TotalReserved += inv.ReservedQty
			resp.TotalAvailable += inv.AvailableQty
		}
		WriteJSON(w, http.StatusOK, resp)
		return
	}

	// Step 3: Neither location nor SKU matched — return empty result (not an error).
	// The PDA shows a "not found" UI on the frontend.
	h.log.Debug("stock inquiry: barcode not recognized",
		slog.String("barcode", barcode),
		slog.String("loc_error", locErr.Error()),
		slog.String("sku_error", skuErr.Error()),
	)
	WriteJSON(w, http.StatusOK, &stockInquiryResponse{
		Barcode:   barcode,
		Inventory: []inventoryResponse{},
	})
}

// ── Route Registration ───────────────────────────────────────────────────────────────────

// RegisterStockInquiryRoutes registers stock inquiry API routes on the given mux.
// This is a PDA-specific endpoint for barcode-based inventory lookups.
func RegisterStockInquiryRoutes(mux *http.ServeMux, h *StockInquiryHandler) {
	mux.HandleFunc("GET /api/v1/stock-inquiry", h.InquireStock)
}
