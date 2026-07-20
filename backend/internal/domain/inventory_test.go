package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// ── Inventory Business Rule Tests ────────────────────────────────────────────

func TestInventory_Available(t *testing.T) {
	tests := []struct {
		name        string
		qty         float64
		reservedQty float64
		want        float64
	}{
		{"no reserve", 100, 0, 100},
		{"partial reserve", 100, 30, 70},
		{"fully reserved", 100, 100, 0},
		{"zero qty", 0, 0, 0},
		{"fractional", 10.5, 3.2, 7.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Inventory{Qty: tt.qty, ReservedQty: tt.reservedQty}
			got := inv.Available()
			if got != tt.want {
				t.Errorf("Available() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestInventory_CanDeduct(t *testing.T) {
	tests := []struct {
		name string
		qty  float64
		res  float64
		ded  float64
		want bool
	}{
		{"enough available", 100, 0, 50, true},
		{"exact available", 100, 0, 100, true},
		{"not enough (reserved)", 100, 80, 50, false}, // 20 available, 50 needed
		{"not enough (qty)", 50, 0, 100, false},
		{"zero deduction", 100, 0, 0, false},
		{"negative deduction", 100, 0, -10, false},
		{"fractional enough", 10.5, 3.0, 7.5, true},
		{"fractional not enough", 10.5, 3.0, 7.6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Inventory{Qty: tt.qty, ReservedQty: tt.res}
			got := inv.CanDeduct(tt.ded)
			if got != tt.want {
				t.Errorf("CanDeduct(%f) = %v, want %v", tt.ded, got, tt.want)
			}
		})
	}
}

func TestInventory_CanReserve(t *testing.T) {
	tests := []struct {
		name       string
		qty        float64
		reservedQty float64
		reserve    float64
		want       bool
	}{
		{"can reserve", 100, 0, 40, true},
		{"can reserve partial", 100, 20, 60, true},
		{"can reserve all available", 100, 0, 100, true},
		{"cannot reserve more than available", 100, 30, 80, false}, // 70 avail, 80 req
		{"cannot reserve zero", 100, 0, 0, false},
		{"cannot reserve negative", 100, 0, -5, false},
		{"no inventory left to reserve", 100, 100, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Inventory{Qty: tt.qty, ReservedQty: tt.reservedQty}
			got := inv.CanReserve(tt.reserve)
			if got != tt.want {
				t.Errorf("CanReserve(%f) = %v, want %v", tt.reserve, got, tt.want)
			}
		})
	}
}

func TestInventory_CanAdjustTo(t *testing.T) {
	tests := []struct {
		name     string
		qty      float64
		deltaQty float64
		want     bool
	}{
		{"increase", 100, 50, true},
		{"decrease within bounds", 100, -50, true},
		{"decrease to exactly zero", 100, -100, true},
		{"decrease below zero", 100, -150, false},
		{"zero delta", 100, 0, true}, // Arithmetic ok, but business rules should reject
		{"from zero, increase", 0, 10, true},
		{"from zero, decrease", 0, -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Inventory{Qty: tt.qty}
			got := inv.CanAdjustTo(tt.deltaQty)
			if got != tt.want {
				t.Errorf("CanAdjustTo(%f) = %v, want %v", tt.deltaQty, got, tt.want)
			}
		})
	}
}

func TestInventory_ResultingQty(t *testing.T) {
	tests := []struct {
		name     string
		qty      float64
		deltaQty float64
		want     float64
	}{
		{"positive delta", 100, 50, 150},
		{"negative delta", 100, -30, 70},
		{"zero delta", 100, 0, 100},
		{"from zero", 0, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Inventory{Qty: tt.qty}
			got := inv.ResultingQty(tt.deltaQty)
			if got != tt.want {
				t.Errorf("ResultingQty(%f) = %f, want %f", tt.deltaQty, got, tt.want)
			}
		})
	}
}

// ── Inventory FEFO/FIFO Tests ────────────────────────────────────────────────

func TestInventory_IsOlderThan(t *testing.T) {
	now := time.Now()
	older := &Inventory{ReceivedAt: now.Add(-2 * time.Hour)}
	newer := &Inventory{ReceivedAt: now}

	if !older.IsOlderThan(newer) {
		t.Error("older inventory should be older than newer")
	}
	if newer.IsOlderThan(older) {
		t.Error("newer inventory should not be older than older")
	}
}

func TestInventory_HasEarlierExpiry(t *testing.T) {
	now := time.Now()
	soon := now.Add(24 * time.Hour)
	later := now.Add(7 * 24 * time.Hour)

	expiresSoon := &Inventory{ExpiryDate: &soon}
	expiresLater := &Inventory{ExpiryDate: &later}
	noExpiry := &Inventory{ExpiryDate: nil}

	// Expires sooner < expires later.
	if !expiresSoon.HasEarlierExpiry(expiresLater) {
		t.Error("soon-expiring should be earlier than later-expiring")
	}
	if expiresLater.HasEarlierExpiry(expiresSoon) {
		t.Error("later-expiring should not be earlier than soon-expiring")
	}

	// Any expiry < no expiry (FEFO: expiring items get priority).
	if !expiresSoon.HasEarlierExpiry(noExpiry) {
		t.Error("expiring item should be earlier than never-expiring")
	}
	if noExpiry.HasEarlierExpiry(expiresSoon) {
		t.Error("never-expiring should not be earlier than expiring")
	}

	// No expiry vs no expiry.
	if noExpiry.HasEarlierExpiry(noExpiry) {
		t.Error("no-expiry vs no-expiry: should be false")
	}
}

func TestInventory_IsExpiredAt(t *testing.T) {
	now := time.Now()
	pastExpiry := now.Add(-1 * time.Hour)
	futureExpiry := now.Add(1 * time.Hour)

	expired := &Inventory{ExpiryDate: &pastExpiry}
	notExpired := &Inventory{ExpiryDate: &futureExpiry}
	noExpiry := &Inventory{ExpiryDate: nil}

	if !expired.IsExpiredAt(now) {
		t.Error("past-expiry inventory should be expired")
	}
	if notExpired.IsExpiredAt(now) {
		t.Error("future-expiry inventory should not be expired")
	}
	if noExpired := noExpiry.IsExpiredAt(now); noExpired {
		t.Error("no-expiry inventory should not be expired")
	}
}

// ── Inventory Status Tests ───────────────────────────────────────────────────

func TestInventoryStatusValues(t *testing.T) {
	all := []InventoryStatus{
		InventoryStatusAvailable, InventoryStatusQuarantine,
		InventoryStatusDamaged, InventoryStatusExpired,
	}

	for _, s := range all {
		if s == "" {
			t.Error("inventory status should not be empty")
		}
	}
}

// ── InventoryTransaction Tests ───────────────────────────────────────────────

func TestInventoryTransaction_IsIncrease(t *testing.T) {
	increase := &InventoryTransaction{DeltaQty: 10}
	if !increase.IsIncrease() {
		t.Error("positive delta should be increase")
	}
	decrease := &InventoryTransaction{DeltaQty: -5}
	if decrease.IsIncrease() {
		t.Error("negative delta should not be increase")
	}
	zero := &InventoryTransaction{DeltaQty: 0}
	if zero.IsIncrease() {
		t.Error("zero delta should not be increase")
	}
}

func TestInventoryTransaction_IsDecrease(t *testing.T) {
	decrease := &InventoryTransaction{DeltaQty: -10}
	if !decrease.IsDecrease() {
		t.Error("negative delta should be decrease")
	}
	increase := &InventoryTransaction{DeltaQty: 5}
	if increase.IsDecrease() {
		t.Error("positive delta should not be decrease")
	}
	zero := &InventoryTransaction{DeltaQty: 0}
	if zero.IsDecrease() {
		t.Error("zero delta should not be decrease")
	}
}

func TestInventoryTxTypeValues(t *testing.T) {
	all := []InventoryTxType{
		InventoryTxReceipt, InventoryTxPutaway, InventoryTxPick,
		InventoryTxShip, InventoryTxTransfer,
		InventoryTxAdjustment, InventoryTxReturn,
	}

	for _, txt := range all {
		if txt == "" {
			t.Error("inventory tx type should not be empty")
		}
	}
}

// ── SKU Tests ────────────────────────────────────────────────────────────────

func TestSKU_IsActive(t *testing.T) {
	active := &SKU{Status: SKUStatusActive}
	if !active.IsActive() {
		t.Error("active status should be active")
	}

	inactive := &SKU{Status: SKUStatusInactive}
	if inactive.IsActive() {
		t.Error("inactive status should not be active")
	}

	discontinued := &SKU{Status: SKUStatusDiscontinued}
	if discontinued.IsActive() {
		t.Error("discontinued status should not be active")
	}
}

func TestSKU_IsDiscontinued(t *testing.T) {
	discontinued := &SKU{Status: SKUStatusDiscontinued}
	if !discontinued.IsDiscontinued() {
		t.Error("discontinued status should be discontinued")
	}

	active := &SKU{Status: SKUStatusActive}
	if active.IsDiscontinued() {
		t.Error("active status should not be discontinued")
	}
}

func TestSKUStatusValues(t *testing.T) {
	all := []SKUStatus{
		SKUStatusActive, SKUStatusInactive, SKUStatusDiscontinued,
	}

	for _, s := range all {
		if s == "" {
			t.Error("sku status should not be empty")
		}
	}
}

// ── Inventory Struct Tests ───────────────────────────────────────────────────

func TestInventory_ZeroValues(t *testing.T) {
	inv := &Inventory{}
	if inv.Available() != 0 {
		t.Errorf("zero-value inventory available should be 0, got %f", inv.Available())
	}
	if inv.CanDeduct(1) {
		t.Error("zero-value inventory should not allow deduction")
	}
}

func TestInventory_StructFields(t *testing.T) {
	id := uuid.New()
	skuID := uuid.New()
	locationID := uuid.New()
	warehouseID := uuid.New()

	inv := &Inventory{
		ID:          id,
		SKUID:       skuID,
		LocationID:  locationID,
		WarehouseID: warehouseID,
		BatchNo:     "LOT-001",
		Qty:         150.0,
		ReservedQty: 25.0,
		Status:      InventoryStatusAvailable,
	}

	if inv.ID != id {
		t.Errorf("ID = %s, want %s", inv.ID, id)
	}
	if inv.SKUID != skuID {
		t.Errorf("SKUID = %s, want %s", inv.SKUID, skuID)
	}
	if inv.LocationID != locationID {
		t.Errorf("LocationID = %s, want %s", inv.LocationID, locationID)
	}
	if inv.WarehouseID != warehouseID {
		t.Errorf("WarehouseID = %s, want %s", inv.WarehouseID, warehouseID)
	}
	if inv.BatchNo != "LOT-001" {
		t.Errorf("BatchNo = %s, want LOT-001", inv.BatchNo)
	}
	if inv.Qty != 150.0 {
		t.Errorf("Qty = %f, want 150.0", inv.Qty)
	}
	if inv.ReservedQty != 25.0 {
		t.Errorf("ReservedQty = %f, want 25.0", inv.ReservedQty)
	}
	// Available() is a method, verified by TestInventory_Available.
	if inv.Available() != 125.0 {
		t.Errorf("Available() = %f, want 125.0", inv.Available())
	}
}

func TestSKU_StructFields(t *testing.T) {
	id := uuid.New()
	sku := &SKU{
		ID:      id,
		Code:    "SKU-001",
		Name:    "Test Product",
		Barcode: "1234567890123",
		Status:  SKUStatusActive,
	}

	if sku.ID != id {
		t.Errorf("ID = %s, want %s", sku.ID, id)
	}
	if sku.Code != "SKU-001" {
		t.Errorf("Code = %s, want SKU-001", sku.Code)
	}
	if sku.Name != "Test Product" {
		t.Errorf("Name = %s, want Test Product", sku.Name)
	}
}
