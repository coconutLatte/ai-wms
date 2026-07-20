// Package domain defines the core business entities of the WMS system.
// These are pure Go structs with zero external dependencies — the heart of the domain.
// All business rules live here; AI agents can reason about them without understanding infrastructure.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Warehouse represents a physical warehouse.
type Warehouse struct {
	ID        uuid.UUID       `json:"id"`
	Code      string          `json:"code"`    // Unique warehouse code, e.g. "WH-SH-01"
	Name      string          `json:"name"`    // Display name, e.g. "Shanghai Main Warehouse"
	Address   string          `json:"address"` // Physical address
	Status    WarehouseStatus `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// WarehouseStatus represents the operational status of a warehouse.
type WarehouseStatus string

const (
	WarehouseStatusActive   WarehouseStatus = "active"
	WarehouseStatusInactive WarehouseStatus = "inactive"
	WarehouseStatusArchived WarehouseStatus = "archived"
)

// Zone represents a logical area within a warehouse (e.g., "Receiving", "Storage", "Picking", "Shipping").
type Zone struct {
	ID          uuid.UUID  `json:"id"`
	WarehouseID uuid.UUID  `json:"warehouse_id"`
	Code        string     `json:"code"`     // e.g. "ZONE-RCV-01"
	Name        string     `json:"name"`     // e.g. "Receiving Zone A"
	ZoneType    ZoneType   `json:"zone_type"`
	Status      ZoneStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ZoneType classifies the purpose of a zone.
type ZoneType string

const (
	ZoneTypeReceiving ZoneType = "receiving"
	ZoneTypeStorage   ZoneType = "storage"
	ZoneTypePicking   ZoneType = "picking"
	ZoneTypeShipping  ZoneType = "shipping"
	ZoneTypeReturns   ZoneType = "returns"
	ZoneTypeStaging   ZoneType = "staging"
)

// ZoneStatus represents the operational state of a zone.
type ZoneStatus string

const (
	ZoneStatusActive   ZoneStatus = "active"
	ZoneStatusInactive ZoneStatus = "inactive"
	ZoneStatusFull     ZoneStatus = "full"
)

// Location represents a specific storage location (bin/slot) within a zone.
type Location struct {
	ID           uuid.UUID      `json:"id"`
	ZoneID       uuid.UUID      `json:"zone_id"`
	WarehouseID  uuid.UUID      `json:"warehouse_id"`
	Code         string         `json:"code"`    // e.g. "A-01-02-03" (aisle-rack-level-bin)
	Barcode      string         `json:"barcode"` // Barcode label on the physical location
	LocationType LocationType   `json:"location_type"`
	Capacity     *Capacity      `json:"capacity,omitempty"` // Max capacity (nil = unlimited)
	Status       LocationStatus `json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// LocationType classifies the type of storage location.
type LocationType string

const (
	LocationTypePallet   LocationType = "pallet"   // Full pallet position
	LocationTypeShelf    LocationType = "shelf"    // Shelf/bin location
	LocationTypeFloor    LocationType = "floor"    // Floor storage
	LocationTypeConveyor LocationType = "conveyor" // Conveyor position (WCS-controlled)
	LocationTypeAGV      LocationType = "agv"      // AGV docking point (RCS-controlled)
)

// LocationStatus represents the state of a location.
type LocationStatus string

const (
	LocationStatusEmpty    LocationStatus = "empty"
	LocationStatusOccupied LocationStatus = "occupied"
	LocationStatusReserved LocationStatus = "reserved"
	LocationStatusBlocked  LocationStatus = "blocked"
)

// Capacity defines the maximum storage capacity of a location.
type Capacity struct {
	MaxWeight float64 `json:"max_weight"` // Max weight in kg
	MaxVolume float64 `json:"max_volume"` // Max volume in m³
	MaxQty    int     `json:"max_qty"`    // Max quantity of units
}

// ── Location State Machine Methods ──────────────────────────────────────────

// IsTerminal returns true if the location is in a terminal (immutable) state.
func (l *Location) IsTerminal() bool {
	return l.Status == LocationStatusBlocked
}

// CanTransitionTo checks whether the location can transition from its current
// status to the target status. This is the authoritative location state machine.
//
// Valid transitions:
//
//	empty    → occupied, reserved, blocked
//	occupied → empty, blocked
//	reserved → occupied, empty, blocked
//	blocked  → empty
func (l *Location) CanTransitionTo(target LocationStatus) bool {
	if l.Status == target {
		return false
	}

	switch l.Status {
	case LocationStatusEmpty:
		return target == LocationStatusOccupied ||
			target == LocationStatusReserved ||
			target == LocationStatusBlocked
	case LocationStatusOccupied:
		return target == LocationStatusEmpty ||
			target == LocationStatusBlocked
	case LocationStatusReserved:
		return target == LocationStatusOccupied ||
			target == LocationStatusEmpty ||
			target == LocationStatusBlocked
	case LocationStatusBlocked:
		return target == LocationStatusEmpty
	default:
		return false
	}
}
