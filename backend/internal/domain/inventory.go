package domain

import (
	"time"

	"github.com/google/uuid"
)

// SKU (Stock Keeping Unit) represents a unique product variant.
type SKU struct {
	ID          uuid.UUID  `json:"id"`
	Code        string     `json:"code"`        // Unique SKU code, e.g. "SKU-12345"
	Name        string     `json:"name"`        // Product name
	Description string     `json:"description"` // Product description
	Barcode     string     `json:"barcode"`     // Primary barcode (UPC/EAN/GS1-128)
	UOM         UOM        `json:"uom"`         // Unit of measure
	Attributes  Attributes `json:"attributes"`  // Flexible attributes (size, weight, color, etc.)
	Category    string     `json:"category"`    // Product category
	Status      SKUStatus  `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// SKUStatus represents the lifecycle state of a SKU.
type SKUStatus string

const (
	SKUStatusActive   SKUStatus = "active"
	SKUStatusInactive SKUStatus = "inactive"
	SKUStatusDiscontinued SKUStatus = "discontinued"
)

// UOM (Unit of Measure) defines how a SKU is measured.
type UOM struct {
	BaseUnit   string  `json:"base_unit"`   // e.g. "EA" (each), "KG", "M"
	PackUnit   string  `json:"pack_unit"`   // e.g. "BOX", "CASE", "PAL"
	PackQty    int     `json:"pack_qty"`    // How many base units per pack
	Weight     float64 `json:"weight"`      // Weight in kg per base unit
	Volume     float64 `json:"volume"`      // Volume in m³ per base unit
	Length     float64 `json:"length"`      // Length in cm
	Width      float64 `json:"width"`       // Width in cm
	Height     float64 `json:"height"`      // Height in cm
}

// Attributes holds flexible key-value product attributes for extensibility.
type Attributes map[string]string

// Inventory represents the quantity of a SKU at a specific location.
type Inventory struct {
	ID          uuid.UUID `json:"id"`
	SKUID       uuid.UUID `json:"sku_id"`
	LocationID  uuid.UUID `json:"location_id"`
	WarehouseID uuid.UUID `json:"warehouse_id"`
	BatchNo     string    `json:"batch_no"`     // Lot/batch number for traceability
	Qty         float64   `json:"qty"`          // On-hand quantity
	ReservedQty float64   `json:"reserved_qty"` // Reserved (allocated but not picked)
	AvailableQty float64  `json:"available_qty"`// Available = Qty - ReservedQty
	Status      InventoryStatus `json:"status"`
	ProductionDate *time.Time `json:"production_date,omitempty"` // Manufacturing date
	ExpiryDate     *time.Time `json:"expiry_date,omitempty"`     // Expiration date (for FEFO)
	ReceivedAt     time.Time  `json:"received_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// InventoryStatus represents the quality/status of inventory.
type InventoryStatus string

const (
	InventoryStatusAvailable  InventoryStatus = "available"
	InventoryStatusQuarantine InventoryStatus = "quarantine" // Quality hold
	InventoryStatusDamaged    InventoryStatus = "damaged"
	InventoryStatusExpired    InventoryStatus = "expired"
)

// InventoryTransaction records every inventory change for audit trail.
type InventoryTransaction struct {
	ID            uuid.UUID              `json:"id"`
	InventoryID   uuid.UUID              `json:"inventory_id"`
	SKUID         uuid.UUID              `json:"sku_id"`
	LocationID    uuid.UUID              `json:"location_id"`
	Type          InventoryTxType        `json:"type"`
	DeltaQty      float64                `json:"delta_qty"`      // Positive = increase, Negative = decrease
	ResultingQty  float64                `json:"resulting_qty"`  // Qty after transaction
	ReferenceType string                 `json:"reference_type"` // "order", "task", "adjustment", "transfer"
	ReferenceID   uuid.UUID              `json:"reference_id"`
	CreatedAt     time.Time              `json:"created_at"`
	CreatedBy     string                 `json:"created_by"`
}

// InventoryTxType classifies the type of inventory movement.
type InventoryTxType string

const (
	InventoryTxReceipt     InventoryTxType = "receipt"     // Goods received
	InventoryTxPutaway     InventoryTxType = "putaway"     // Goods moved to storage
	InventoryTxPick        InventoryTxType = "pick"        // Goods picked for order
	InventoryTxShip        InventoryTxType = "ship"        // Goods shipped out
	InventoryTxTransfer    InventoryTxType = "transfer"    // Location-to-location transfer
	InventoryTxAdjustment  InventoryTxType = "adjustment"  // Manual adjustment (cycle count)
	InventoryTxReturn      InventoryTxType = "return"      // Customer return
)

// ── Business Rule Methods ────────────────────────────────────────────────────

// Available returns the currently available quantity (on-hand minus reserved).
func (inv *Inventory) Available() float64 {
	return inv.Qty - inv.ReservedQty
}

// CanDeduct returns true if the requested quantity can be deducted from available inventory.
// Rule: inventory cannot be negative, and deduction must come from available qty.
func (inv *Inventory) CanDeduct(qty float64) bool {
	if qty <= 0 {
		return false
	}
	return inv.Available() >= qty
}

// CanReserve returns true if the requested quantity can be reserved.
// Rule: reserve only against available inventory, adjusted for already-reserved.
func (inv *Inventory) CanReserve(qty float64) bool {
	if qty <= 0 {
		return false
	}
	return inv.Qty-inv.ReservedQty >= qty
}

// ResultingQty returns what the on-hand quantity would be after applying deltaQty.
// Rules: does NOT perform validation — just arithmetic.
func (inv *Inventory) ResultingQty(deltaQty float64) float64 {
	return inv.Qty + deltaQty
}

// CanAdjustTo checks whether adjusting inventory by deltaQty would violate business rules.
// Rule: resulting qty must be >= 0 (no negative inventory).
func (inv *Inventory) CanAdjustTo(deltaQty float64) bool {
	return inv.ResultingQty(deltaQty) >= 0
}

// IsExpired returns true if the inventory is past its expiry date.
func (inv *Inventory) IsExpired() bool {
	if inv.ExpiryDate == nil {
		return false
	}
	return time.Now().After(*inv.ExpiryDate)
}

// IsExpiredAt returns true if the inventory is past the given reference time.
func (inv *Inventory) IsExpiredAt(ref time.Time) bool {
	if inv.ExpiryDate == nil {
		return false
	}
	return ref.After(*inv.ExpiryDate)
}

// IsOlderThan returns true if this inventory was received before the other.
// Used for FIFO allocation decisions.
func (inv *Inventory) IsOlderThan(other *Inventory) bool {
	return inv.ReceivedAt.Before(other.ReceivedAt)
}

// HasEarlierExpiry returns true if this inventory expires before the other.
// Nil expiry is treated as "never expires" (sorted last). Used for FEFO.
func (inv *Inventory) HasEarlierExpiry(other *Inventory) bool {
	if inv.ExpiryDate == nil {
		return false
	}
	if other.ExpiryDate == nil {
		return true
	}
	return inv.ExpiryDate.Before(*other.ExpiryDate)
}

// ── Inventory Status State Machine ──────────────────────────────────────────

// IsTerminal returns true if the inventory is in a terminal (immutable) status.
func (inv *Inventory) IsTerminal() bool {
	return inv.Status == InventoryStatusExpired
}

// CanTransitionTo checks whether the inventory can transition from its current
// status to the target status. This is the authoritative inventory status state machine.
//
// Valid transitions:
//
//	available  → quarantine, damaged, expired
//	quarantine → available, damaged, expired
//	damaged    → available, expired
//	expired    → (terminal)
func (inv *Inventory) CanTransitionTo(target InventoryStatus) bool {
	if inv.Status == target {
		return false
	}
	if inv.IsTerminal() {
		return false
	}

	switch inv.Status {
	case InventoryStatusAvailable:
		return target == InventoryStatusQuarantine ||
			target == InventoryStatusDamaged ||
			target == InventoryStatusExpired
	case InventoryStatusQuarantine:
		return target == InventoryStatusAvailable ||
			target == InventoryStatusDamaged ||
			target == InventoryStatusExpired
	case InventoryStatusDamaged:
		return target == InventoryStatusAvailable ||
			target == InventoryStatusExpired
	default:
		return false
	}
}

// IsActive returns true if the SKU is available for operations.
func (s *SKU) IsActive() bool {
	return s.Status == SKUStatusActive
}

// IsDiscontinued returns true if the SKU has been discontinued.
func (s *SKU) IsDiscontinued() bool {
	return s.Status == SKUStatusDiscontinued
}

// ── InventoryTransaction Methods ──────────────────────────────────────────────

// IsIncrease returns true if the transaction adds inventory.
func (tx *InventoryTransaction) IsIncrease() bool {
	return tx.DeltaQty > 0
}

// IsDecrease returns true if the transaction removes inventory.
func (tx *InventoryTransaction) IsDecrease() bool {
	return tx.DeltaQty < 0
}
