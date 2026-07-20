package domain

import (
	"testing"

	"github.com/google/uuid"
)

// ── Warehouse Status Tests ───────────────────────────────────────────────────

func TestWarehouseStatusValues(t *testing.T) {
	all := []WarehouseStatus{
		WarehouseStatusActive, WarehouseStatusInactive, WarehouseStatusArchived,
	}

	for _, s := range all {
		if s == "" {
			t.Error("warehouse status should not be empty")
		}
	}
}

func TestWarehouse_Struct(t *testing.T) {
	id := uuid.New()
	wh := &Warehouse{
		ID:     id,
		Code:   "WH-SH-01",
		Name:   "Shanghai Main Warehouse",
		Status: WarehouseStatusActive,
	}

	if wh.ID != id {
		t.Errorf("ID = %s, want %s", wh.ID, id)
	}
	if wh.Code != "WH-SH-01" {
		t.Errorf("Code = %s, want WH-SH-01", wh.Code)
	}
	if wh.Name != "Shanghai Main Warehouse" {
		t.Errorf("Name = %s, want Shanghai Main Warehouse", wh.Name)
	}
	if wh.Status != WarehouseStatusActive {
		t.Errorf("Status = %s, want %s", wh.Status, WarehouseStatusActive)
	}
}

// ── Zone Tests ────────────────────────────────────────────────────────────────

func TestZoneTypeValues(t *testing.T) {
	all := []ZoneType{
		ZoneTypeReceiving, ZoneTypeStorage, ZoneTypePicking,
		ZoneTypeShipping, ZoneTypeReturns, ZoneTypeStaging,
	}

	for _, zt := range all {
		if zt == "" {
			t.Error("zone type should not be empty")
		}
	}
}

func TestZoneStatusValues(t *testing.T) {
	all := []ZoneStatus{
		ZoneStatusActive, ZoneStatusInactive, ZoneStatusFull,
	}

	for _, s := range all {
		if s == "" {
			t.Error("zone status should not be empty")
		}
	}
}

func TestZone_Struct(t *testing.T) {
	id := uuid.New()
	whID := uuid.New()
	z := &Zone{
		ID:          id,
		WarehouseID: whID,
		Code:        "ZONE-RCV-01",
		Name:        "Receiving Zone A",
		ZoneType:    ZoneTypeReceiving,
		Status:      ZoneStatusActive,
	}

	if z.ID != id {
		t.Errorf("ID = %s, want %s", z.ID, id)
	}
	if z.WarehouseID != whID {
		t.Errorf("WarehouseID = %s, want %s", z.WarehouseID, whID)
	}
	if z.Code != "ZONE-RCV-01" {
		t.Errorf("Code = %s, want ZONE-RCV-01", z.Code)
	}
	if z.ZoneType != ZoneTypeReceiving {
		t.Errorf("ZoneType = %s, want %s", z.ZoneType, ZoneTypeReceiving)
	}
}

// ── Location Tests ────────────────────────────────────────────────────────────

func TestLocationTypeValues(t *testing.T) {
	all := []LocationType{
		LocationTypePallet, LocationTypeShelf, LocationTypeFloor,
		LocationTypeConveyor, LocationTypeAGV,
	}

	for _, lt := range all {
		if lt == "" {
			t.Error("location type should not be empty")
		}
	}
}

func TestLocationStatusValues(t *testing.T) {
	all := []LocationStatus{
		LocationStatusEmpty, LocationStatusOccupied,
		LocationStatusReserved, LocationStatusBlocked,
	}

	for _, s := range all {
		if s == "" {
			t.Error("location status should not be empty")
		}
	}
}

func TestLocation_Struct(t *testing.T) {
	id := uuid.New()
	zoneID := uuid.New()
	whID := uuid.New()

	loc := &Location{
		ID:           id,
		ZoneID:       zoneID,
		WarehouseID:  whID,
		Code:         "A-01-02-03",
		Barcode:      "LOC-A010203",
		LocationType: LocationTypeShelf,
		Status:       LocationStatusEmpty,
	}

	if loc.ID != id {
		t.Errorf("ID = %s, want %s", loc.ID, id)
	}
	if loc.ZoneID != zoneID {
		t.Errorf("ZoneID = %s, want %s", loc.ZoneID, zoneID)
	}
	if loc.Code != "A-01-02-03" {
		t.Errorf("Code = %s, want A-01-02-03", loc.Code)
	}
	if loc.Barcode != "LOC-A010203" {
		t.Errorf("Barcode = %s, want LOC-A010203", loc.Barcode)
	}
	if loc.LocationType != LocationTypeShelf {
		t.Errorf("LocationType = %s, want %s", loc.LocationType, LocationTypeShelf)
	}
	if loc.Status != LocationStatusEmpty {
		t.Errorf("Status = %s, want %s", loc.Status, LocationStatusEmpty)
	}
}

// ── Capacity Tests ────────────────────────────────────────────────────────────

func TestCapacity_Struct(t *testing.T) {
	c := &Capacity{
		MaxWeight: 500.0,
		MaxVolume: 2.5,
		MaxQty:    100,
	}

	if c.MaxWeight != 500.0 {
		t.Errorf("MaxWeight = %f, want 500.0", c.MaxWeight)
	}
	if c.MaxVolume != 2.5 {
		t.Errorf("MaxVolume = %f, want 2.5", c.MaxVolume)
	}
	if c.MaxQty != 100 {
		t.Errorf("MaxQty = %d, want 100", c.MaxQty)
	}
}

func TestLocation_WithCapacity(t *testing.T) {
	capacity := &Capacity{MaxWeight: 1000, MaxVolume: 3.0, MaxQty: 50}
	loc := &Location{
		Code:         "A-02-01-01",
		LocationType: LocationTypePallet,
		Status:       LocationStatusOccupied,
		Capacity:     capacity,
	}

	if loc.Capacity == nil {
		t.Fatal("capacity should not be nil")
	}
	if loc.Capacity.MaxQty != 50 {
		t.Errorf("MaxQty = %d, want 50", loc.Capacity.MaxQty)
	}
}

func TestLocation_NilCapacity(t *testing.T) {
	loc := &Location{
		Code:         "UNLIMITED-SLOT",
		LocationType: LocationTypeFloor,
		Status:       LocationStatusEmpty,
	}

	if loc.Capacity != nil {
		t.Error("capacity should be nil (unlimited) by default")
	}
}

// ── Location State Machine Tests ──────────────────────────────────────────────

func TestLocation_IsTerminal(t *testing.T) {
	tests := []struct {
		status   LocationStatus
		terminal bool
	}{
		{LocationStatusEmpty, false},
		{LocationStatusOccupied, false},
		{LocationStatusReserved, false},
		{LocationStatusBlocked, true},
	}

	for _, tt := range tests {
		loc := &Location{Status: tt.status}
		if got := loc.IsTerminal(); got != tt.terminal {
			t.Errorf("IsTerminal(%s) = %v, want %v", tt.status, got, tt.terminal)
		}
	}
}

func TestLocation_CanTransitionTo_SameStatus(t *testing.T) {
	all := []LocationStatus{
		LocationStatusEmpty, LocationStatusOccupied,
		LocationStatusReserved, LocationStatusBlocked,
	}
	for _, s := range all {
		loc := &Location{Status: s}
		if loc.CanTransitionTo(s) {
			t.Errorf("CanTransitionTo(%s → %s) should be false (same status)", s, s)
		}
	}
}

func TestLocation_CanTransitionTo_FromEmpty(t *testing.T) {
	loc := &Location{Status: LocationStatusEmpty}

	valid := []LocationStatus{
		LocationStatusOccupied, LocationStatusReserved, LocationStatusBlocked,
	}
	for _, target := range valid {
		if !loc.CanTransitionTo(target) {
			t.Errorf("CanTransitionTo(empty → %s) should be true", target)
		}
	}
}

func TestLocation_CanTransitionTo_FromOccupied(t *testing.T) {
	loc := &Location{Status: LocationStatusOccupied}

	valid := []LocationStatus{LocationStatusEmpty, LocationStatusBlocked}
	for _, target := range valid {
		if !loc.CanTransitionTo(target) {
			t.Errorf("CanTransitionTo(occupied → %s) should be true", target)
		}
	}

	invalid := []LocationStatus{LocationStatusReserved}
	for _, target := range invalid {
		if loc.CanTransitionTo(target) {
			t.Errorf("CanTransitionTo(occupied → %s) should be false", target)
		}
	}
}

func TestLocation_CanTransitionTo_FromReserved(t *testing.T) {
	loc := &Location{Status: LocationStatusReserved}

	valid := []LocationStatus{
		LocationStatusOccupied, LocationStatusEmpty, LocationStatusBlocked,
	}
	for _, target := range valid {
		if !loc.CanTransitionTo(target) {
			t.Errorf("CanTransitionTo(reserved → %s) should be true", target)
		}
	}
}

func TestLocation_CanTransitionTo_FromBlocked(t *testing.T) {
	loc := &Location{Status: LocationStatusBlocked}

	// Blocked can only transition to empty (unblock).
	if !loc.CanTransitionTo(LocationStatusEmpty) {
		t.Errorf("CanTransitionTo(blocked → empty) should be true")
	}

	invalid := []LocationStatus{
		LocationStatusOccupied, LocationStatusReserved,
	}
	for _, target := range invalid {
		if loc.CanTransitionTo(target) {
			t.Errorf("CanTransitionTo(blocked → %s) should be false", target)
		}
	}
}

func TestLocation_CanTransitionTo_FullTable(t *testing.T) {
	// Exhaustive test of all 4×4 = 16 possible transitions.
	type testCase struct {
		from   LocationStatus
		to     LocationStatus
		expect bool
	}

	cases := []testCase{
		// empty → *
		{LocationStatusEmpty, LocationStatusEmpty, false},
		{LocationStatusEmpty, LocationStatusOccupied, true},
		{LocationStatusEmpty, LocationStatusReserved, true},
		{LocationStatusEmpty, LocationStatusBlocked, true},
		// occupied → *
		{LocationStatusOccupied, LocationStatusEmpty, true},
		{LocationStatusOccupied, LocationStatusOccupied, false},
		{LocationStatusOccupied, LocationStatusReserved, false},
		{LocationStatusOccupied, LocationStatusBlocked, true},
		// reserved → *
		{LocationStatusReserved, LocationStatusEmpty, true},
		{LocationStatusReserved, LocationStatusOccupied, true},
		{LocationStatusReserved, LocationStatusReserved, false},
		{LocationStatusReserved, LocationStatusBlocked, true},
		// blocked → *
		{LocationStatusBlocked, LocationStatusEmpty, true},
		{LocationStatusBlocked, LocationStatusOccupied, false},
		{LocationStatusBlocked, LocationStatusReserved, false},
		{LocationStatusBlocked, LocationStatusBlocked, false},
	}

	for _, c := range cases {
		loc := &Location{Status: c.from}
		got := loc.CanTransitionTo(c.to)
		if got != c.expect {
			t.Errorf("CanTransitionTo(%s → %s) = %v, want %v", c.from, c.to, got, c.expect)
		}
	}
}

// ── UOM Tests ─────────────────────────────────────────────────────────────────

func TestUOM_Struct(t *testing.T) {
	u := UOM{
		BaseUnit: "EA",
		PackUnit: "BOX",
		PackQty:  12,
		Weight:   0.5,
		Volume:   0.001,
		Length:   10.0,
		Width:    8.0,
		Height:   5.0,
	}

	if u.BaseUnit != "EA" {
		t.Errorf("BaseUnit = %s, want EA", u.BaseUnit)
	}
	if u.PackUnit != "BOX" {
		t.Errorf("PackUnit = %s, want BOX", u.PackUnit)
	}
	if u.PackQty != 12 {
		t.Errorf("PackQty = %d, want 12", u.PackQty)
	}
}
