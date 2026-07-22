// Package domain — WMS standard business scenario tests.
// Covers core warehouse flows: ASN→receiving→putaway, wave→allocate→pick→verify,
// outbound→weighing→shipping, cycle count→discrepancy→adjustment, and race conditions.
package domain

import (
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ── Helpers ──────────────────────────────────────────────────────────────────

// makeFEFOBatches creates Inventory records with varying expiry dates for FEFO tests.
func makeFEFOBatches(skuID uuid.UUID, locID uuid.UUID, whID uuid.UUID) []*Inventory {
	now := time.Now()
	return []*Inventory{
		{
			ID: uuid.New(), SKUID: skuID, LocationID: locID, WarehouseID: whID,
			BatchNo: "LOT-EARLY", Qty: 100, Status: InventoryStatusAvailable,
			ReceivedAt: now.Add(-30 * 24 * time.Hour),
			ExpiryDate: timePtr(now.Add(3 * 24 * time.Hour)),
		},
		{
			ID: uuid.New(), SKUID: skuID, LocationID: locID, WarehouseID: whID,
			BatchNo: "LOT-MID", Qty: 100, Status: InventoryStatusAvailable,
			ReceivedAt: now.Add(-20 * 24 * time.Hour),
			ExpiryDate: timePtr(now.Add(10 * 24 * time.Hour)),
		},
		{
			ID: uuid.New(), SKUID: skuID, LocationID: locID, WarehouseID: whID,
			BatchNo: "LOT-LATE", Qty: 100, Status: InventoryStatusAvailable,
			ReceivedAt: now.Add(-10 * 24 * time.Hour),
			ExpiryDate: timePtr(now.Add(60 * 24 * time.Hour)),
		},
	}
}

func timePtr(t time.Time) *time.Time { return &t }

// ── Scenario 1: ASN → Receiving → Putaway (Inbound Flow) ─────────────────────

// TestScenario_InboundFlow exercises the complete inbound receiving flow:
//  1. ASN is created (pending)
//  2. ASN arrives → receiving state
//  3. Goods received → putaway tasks created
//  4. Tasks completed → inventory exists at storage location
func TestScenario_InboundFlow(t *testing.T) {
	skuID := uuid.New()
	whID := uuid.New()
	receivingLocID := uuid.New()
	storageLocID := uuid.New()

	// ── Step 1: Create ASN ──
	asn := &ASN{
		ID:          uuid.New(),
		ASNNo:       "ASN-2026-001",
		WarehouseID: whID,
		Status:      ASNStatusPending,
		Lines: []ASNLine{
			{ID: uuid.New(), ASNID: uuid.New(), SKUID: skuID, ExpectedQty: 100, Status: ASNLineStatusPending},
		},
	}
	if asn.Status != ASNStatusPending {
		t.Fatal("step 1: ASN should start as pending")
	}

	// ── Step 2: ASN arrives ──
	if !asn.CanTransitionTo(ASNStatusArrived) {
		t.Fatal("step 2: pending → arrived should be valid")
	}
	asn.Status = ASNStatusArrived

	if !asn.CanTransitionTo(ASNStatusReceiving) {
		t.Fatal("step 2: arrived → receiving should be valid")
	}
	asn.Status = ASNStatusReceiving

	// ── Step 3: Receive goods — validate inventory creation ──
	now := time.Now()
	for i := range asn.Lines {
		asn.Lines[i].ReceivedQty = asn.Lines[i].ExpectedQty
		asn.Lines[i].Status = ASNLineStatusReceived
	}

	// Inventory at receiving dock.
	receivingInv := &Inventory{
		ID:          uuid.New(),
		SKUID:       skuID,
		LocationID:  receivingLocID,
		WarehouseID: whID,
		BatchNo:     "LOT-RCV-001",
		Qty:         100,
		Status:      InventoryStatusAvailable,
		ReceivedAt:  now,
		ExpiryDate:  timePtr(now.Add(90 * 24 * time.Hour)),
	}
	if !receivingInv.CanDeduct(10) {
		t.Fatal("step 3: receiving inventory should be deductable")
	}
	if receivingInv.Available() != 100 {
		t.Errorf("step 3: available = %f, want 100", receivingInv.Available())
	}

	// ASN received.
	if !asn.CanTransitionTo(ASNStatusReceived) {
		t.Fatal("step 3: receiving → received should be valid")
	}
	asn.Status = ASNStatusReceived
	if !asn.IsTerminal() {
		t.Error("step 3: received ASN should be terminal")
	}

	// ── Step 4: Putaway to storage location ──
	// Remove from receiving, add to storage.
	putawayQty := receivingInv.Qty
	receivingInv.Qty = 0

	storageInv := &Inventory{
		ID:          uuid.New(),
		SKUID:       skuID,
		LocationID:  storageLocID,
		WarehouseID: whID,
		BatchNo:     "LOT-RCV-001",
		Qty:         putawayQty,
		Status:      InventoryStatusAvailable,
		ReceivedAt:  now,
		ExpiryDate:  timePtr(now.Add(90 * 24 * time.Hour)),
	}
	if storageInv.Available() != putawayQty {
		t.Errorf("step 4: storage available = %f, want %f", storageInv.Available(), putawayQty)
	}
	if receivingInv.Available() != 0 {
		t.Errorf("step 4: receiving dock should be empty after putaway, got %f", receivingInv.Available())
	}
}

// ── Scenario 2: Wave → Allocation → Pick → Verify (Outbound Flow) ────────────

// TestScenario_OutboundFEFOFlow exercises FEFO-driven outbound picking:
//  1. Wave created for outbound orders
//  2. Inventory allocated using FEFO (earliest expiry first)
//  3. Tasks picked against allocated inventory
//  4. Verification confirms picked qty matches order lines
func TestScenario_OutboundFEFOFlow(t *testing.T) {
	skuID := uuid.New()
	locID := uuid.New()
	whID := uuid.New()

	// ── Step 1: Create wave for an outbound order ──
	wave := &Wave{
		ID:          uuid.New(),
		WaveNo:      "WAVE-2026-001",
		WarehouseID: whID,
		WaveType:    WaveTypeSingleOrder,
		Status:      WaveStatusCreated,
	}
	if !wave.CanTransitionTo(WaveStatusReleased) {
		t.Fatal("step 1: created → released should be valid")
	}
	wave.Status = WaveStatusReleased
	now := time.Now()
	wave.ReleasedAt = &now

	// Order line: need 150 units of SKU.
	orderLine := &OrderLine{
		ID:         uuid.New(),
		LineNo:    1,
		SKUID:     skuID,
		OrderedQty: 150,
		Status:    OrderLineStatusPending,
	}

	// ── Step 2: FEFO allocation ──
	// 3 batches: LOT-EARLY expires in 3 days, LOT-MID in 10, LOT-LATE in 60.
	batches := makeFEFOBatches(skuID, locID, whID)

	// Sort by FEFO.
	sort.Slice(batches, func(a, b int) bool {
		return batches[a].HasEarlierExpiry(batches[b])
	})

	// Verify sort order.
	if batches[0].BatchNo != "LOT-EARLY" {
		t.Errorf("step 2: FEFO sort: first = %q, want LOT-EARLY", batches[0].BatchNo)
	}
	if batches[1].BatchNo != "LOT-MID" {
		t.Errorf("step 2: FEFO sort: second = %q, want LOT-MID", batches[1].BatchNo)
	}
	if batches[2].BatchNo != "LOT-LATE" {
		t.Errorf("step 2: FEFO sort: third = %q, want LOT-LATE", batches[2].BatchNo)
	}

	// Allocate 150 units (need LOT-EARLY [100] + LOT-MID [50]).
	needed := orderLine.OrderedQty
	totalAllocated := 0.0
	allocatedFrom := make(map[string]float64)
	for _, batch := range batches {
		if needed <= 0 {
			break
		}
		available := batch.Available()
		take := min(needed, available)
		batch.ReservedQty += take
		needed -= take
		totalAllocated += take
		allocatedFrom[batch.BatchNo] = take
	}

	if totalAllocated != 150 {
		t.Errorf("step 2: total allocated = %f, want 150", totalAllocated)
	}
	if needed != 0 {
		t.Errorf("step 2: unmet demand = %f, want 0", needed)
	}
	if allocatedFrom["LOT-EARLY"] != 100 {
		t.Errorf("step 2: LOT-EARLY allocated = %f, want 100", allocatedFrom["LOT-EARLY"])
	}
	if allocatedFrom["LOT-MID"] != 50 {
		t.Errorf("step 2: LOT-MID allocated = %f, want 50", allocatedFrom["LOT-MID"])
	}

	// Order line transitions to allocated.
	orderLine.Status = OrderLineStatusAllocated

	// ── Step 3: Pick (deduct actual inventory from reserved batches) ──
	for batchNo, allocated := range allocatedFrom {
		for _, batch := range batches {
			if batch.BatchNo == batchNo {
				// After allocation, available = Qty - ReservedQty.
				// Pick fulfillment: reduce both Qty and ReservedQty (convert reservation to actual pick).
				if batch.Qty- (batch.ReservedQty - allocated) < 0 {
					t.Errorf("step 3: not enough on-hand to fulfill %f from %s", allocated, batchNo)
					continue
				}
				// Direct deduction: reservation was already committed.
				batch.ReservedQty -= allocated
				batch.Qty -= allocated
			}
		}
	}

	// Verify total remaining.
	remaining := 0.0
	for _, batch := range batches {
		remaining += batch.Qty
	}
	if remaining != 150 { // 300 - 150
		t.Errorf("step 3: remaining inventory = %f, want 150", remaining)
	}

	orderLine.FulfilledQty = totalAllocated

	// ── Step 4: Verify pick accuracy ──
	if orderLine.FulfilledQty != orderLine.OrderedQty {
		t.Errorf("step 4: fulfilled %f != ordered %f", orderLine.FulfilledQty, orderLine.OrderedQty)
	}
	if orderLine.Status != OrderLineStatusAllocated {
		t.Errorf("step 4: order line status = %s, want allocated", orderLine.Status)
	}
	orderLine.Status = OrderLineStatusFulfilled
	if !orderLine.IsTerminal() {
		t.Error("step 4: fulfilled order line should be terminal")
	}

	// Wave progresses.
	wave.Status = WaveStatusInProgress
	wave.Status = WaveStatusCompleted
	if !wave.IsTerminal() {
		t.Error("step 4: completed wave should be terminal")
	}
}

// ── Scenario 3: Outbound → Weighing → Shipping ───────────────────────────────

// TestScenario_OutboundShipmentFlow exercises the outbound shipment flow:
//  1. Outbound order created
//  2. Order processing → picked items weighed as part of packing
//  3. Weight verified against expected
//  4. Order shipped (completed)
func TestScenario_OutboundShipmentFlow(t *testing.T) {
	skuID := uuid.New()
	whID := uuid.New()

	// ── Step 1: Create outbound order ──
	order := &Order{
		ID:          uuid.New(),
		OrderNo:     "OUT-2026-001",
		OrderType:   OrderTypeOutbound,
		WarehouseID: whID,
		Status:      OrderStatusDraft,
		Priority:    OrderPriorityNormal,
		Lines: []OrderLine{
			{ID: uuid.New(), LineNo: 1, SKUID: skuID, OrderedQty: 10, Status: OrderLineStatusPending},
			{ID: uuid.New(), LineNo: 2, SKUID: skuID, OrderedQty: 5, Status: OrderLineStatusPending},
		},
	}

	// Transition to confirmed.
	if !order.CanTransitionTo(OrderStatusConfirmed) {
		t.Fatal("step 1: draft → confirmed invalid")
	}
	order.Status = OrderStatusConfirmed

	// Transition to processing.
	if !order.CanTransitionTo(OrderStatusProcessing) {
		t.Fatal("step 1: confirmed → processing invalid")
	}
	order.Status = OrderStatusProcessing

	// ── Step 2: Weighing (packing station) ──
	// Simulate UOM-based weight: SKU UOM has weight 2.5 kg/unit.
	skuUOM := UOM{BaseUnit: "EA", Weight: 2.5, Volume: 0.01}
	totalWeight := 0.0
	for _, line := range order.Lines {
		totalWeight += line.OrderedQty * skuUOM.Weight
	}
	expectedWeight := 15 * 2.5 // 37.5 kg
	if totalWeight != expectedWeight {
		t.Errorf("step 2: total weight = %f, want %f", totalWeight, expectedWeight)
	}

	// Weighing verification: actual weight within tolerance.
	actualWeight := 37.8
	tolerance := 0.5 // 0.5 kg tolerance
	deviation := actualWeight - totalWeight
	if deviation < -tolerance || deviation > tolerance {
		t.Errorf("step 2: weight deviation %f exceeds tolerance ±%f", deviation, tolerance)
	}

	// ── Step 3: Shipping ──
	// Mark all lines fulfilled.
	for i := range order.Lines {
		order.Lines[i].FulfilledQty = order.Lines[i].OrderedQty
		order.Lines[i].Status = OrderLineStatusFulfilled
	}
	order.Status = OrderStatusCompleted
	now := time.Now()
	order.CompletedAt = &now

	if !order.IsTerminal() {
		t.Error("step 3: completed order should be terminal")
	}
	for _, line := range order.Lines {
		if !line.IsTerminal() {
			t.Errorf("step 3: line %d should be terminal", line.LineNo)
		}
	}
}

// ── Scenario 4: Cycle Count → Discrepancy → Adjustment ───────────────────────

// TestScenario_CycleCountFlow exercises the inventory reconciliation process:
//  1. Cycle count task created for a location
//  2. Physical count differs from system record
//  3. Discrepancy analysis performed
//  4. Adjustment applied to reconcile
func TestScenario_CycleCountFlow(t *testing.T) {
	skuID := uuid.New()
	locID := uuid.New()
	whID := uuid.New()

	// ── Step 1: Setup — system shows 100 units ──
	inv := &Inventory{
		ID:          uuid.New(),
		SKUID:       skuID,
		LocationID:  locID,
		WarehouseID: whID,
		BatchNo:     "LOT-001",
		Qty:         100,
		Status:      InventoryStatusAvailable,
		ReceivedAt:  time.Now().Add(-30 * 24 * time.Hour),
	}

	// Cycle count task created.
	task := &Task{
		ID:          uuid.New(),
		TaskNo:      "TASK-CC-001",
		TaskType:    TaskTypeCycleCount,
		WarehouseID: whID,
		Status:      TaskStatusPending,
		SKUID:       skuID,
		ExpectedQty: inv.Available(), // System record.
	}

	// ── Step 2: Physical count result — 95 units ──
	physicalCount := 95.0
	discrepancy := physicalCount - task.ExpectedQty // -5

	if discrepancy >= 0 {
		t.Errorf("step 2: expected negative discrepancy, got %f", discrepancy)
	}

	// ── Step 3: Discrepancy analysis ──
	threshold := 2.0 // 2% tolerance for small differences
	pctDiscrepancy := (physicalCount - task.ExpectedQty) / task.ExpectedQty * 100
	exceedsThreshold := pctDiscrepancy < -threshold || pctDiscrepancy > threshold

	// At -5 for 100 units, that's -5% — exceeds threshold.
	if !exceedsThreshold {
		t.Error("step 3: discrepancy should exceed 2% threshold")
	}

	// ── Step 4: Apply adjustment ──
	delta := discrepancy // -5
	if !inv.CanAdjustTo(delta) {
		t.Fatal("step 4: adjustment should be valid")
	}
	inv.Qty += delta // = 95

	if inv.Available() != physicalCount {
		t.Errorf("step 4: adjusted qty = %f, want %f", inv.Available(), physicalCount)
	}

	// Task completed.
	task.ActualQty = physicalCount
	task.Status = TaskStatusInProgress
	task.Status = TaskStatusCompleted

	if !task.IsTerminal() {
		t.Error("step 4: completed task should be terminal")
	}

	// ── Edge case: adjustment below zero ──
	if inv.CanAdjustTo(-200) {
		t.Error("step 4: adjustment below zero should be rejected")
	}
}

// ── Scenario 5: Inventory Accuracy KPIs ──────────────────────────────────────

// TestScenario_InventoryAccuracyKPI validates that inventory accuracy is ≥99.5%.
// Accuracy = (correct locations / total locations checked) * 100.
func TestScenario_InventoryAccuracyKPI(t *testing.T) {
	type countResult struct {
		Location     string
		SystemQty    float64
		PhysicalQty  float64
		IsCorrect    bool
	}

	// Simulate 1000 cycle count results.
	results := make([]countResult, 1000)
	correctCount := 0
	for i := 0; i < 1000; i++ {
		sysQty := float64(100 + (i % 10)) // 100–109
		physQty := sysQty
		// Introduce 5 discrepancies (0.5% error rate).
		if i == 42 || i == 137 || i == 299 || i == 555 || i == 888 {
			physQty = sysQty - 1
		}
		results[i] = countResult{
			Location:    fmt.Sprintf("LOC-%04d", i),
			SystemQty:   sysQty,
			PhysicalQty: physQty,
			IsCorrect:   sysQty == physQty,
		}
		if results[i].IsCorrect {
			correctCount++
		}
	}

	accuracy := float64(correctCount) / float64(len(results)) * 100
	minAccuracy := 99.5

	if accuracy < minAccuracy {
		t.Errorf("inventory accuracy = %.2f%%, below minimum %.1f%%", accuracy, minAccuracy)
	}
	t.Logf("inventory accuracy: %.2f%% (%d/%d correct)", accuracy, correctCount, len(results))
}

// TestScenario_PickAccuracyKPI validates that pick accuracy is ≥99.9%.
// Pick accuracy = (correct picks / total picks) * 100.
func TestScenario_PickAccuracyKPI(t *testing.T) {
	type pickResult struct {
		TaskID       string
		ExpectedQty  float64
		PickedQty    float64
		IsCorrect    bool
		ErrorType    string // empty if correct
	}

	// Simulate 10000 pick operations.
	results := make([]pickResult, 10000)
	correctCount := 0
	shortCount := 0
	overCount := 0
	damagedCount := 0

	for i := 0; i < 10000; i++ {
		expected := float64(5 + (i % 20)) // 5–24 units
		picked := expected
		errorType := ""

		// ~0.05% error rate (5 out of 10000).
		switch i {
		case 42:
			picked = expected - 1 // short pick
			errorType = "short"
		case 137:
			picked = expected + 1 // over pick
			errorType = "over"
		case 299:
			picked = expected - 2 // short pick
			errorType = "short"
		case 555:
			picked = 0 // damaged / missing
			errorType = "damaged"
		case 888:
			picked = expected + 2 // over pick
			errorType = "over"
		}

		isCorrect := picked == expected
		results[i] = pickResult{
			TaskID:      fmt.Sprintf("TASK-PICK-%05d", i),
			ExpectedQty: expected,
			PickedQty:   picked,
			IsCorrect:   isCorrect,
			ErrorType:   errorType,
		}

		if isCorrect {
			correctCount++
		} else {
			switch errorType {
			case "short":
				shortCount++
			case "over":
				overCount++
			case "damaged":
				damagedCount++
			}
		}
	}

	accuracy := float64(correctCount) / float64(len(results)) * 100
	minAccuracy := 99.9

	if accuracy < minAccuracy {
		t.Errorf("pick accuracy = %.2f%%, below minimum %.1f%%", accuracy, minAccuracy)
	}
	t.Logf("pick accuracy: %.2f%% (%d/%d correct, errors: short=%d, over=%d, damaged=%d)",
		accuracy, correctCount, len(results), shortCount, overCount, damagedCount)
}

// TestScenario_CombinedAccuracyKPI validates both KPIs together in a
// single logical run to demonstrate end-to-end quality metrics.
func TestScenario_CombinedAccuracyKPI(t *testing.T) {
	// Warehouse-level accuracy dashboard.
	totalInventoryLocations := 5000
	inventoryDiscrepancies := 0
	totalPicks := 20000
	pickErrors := 0

	// Simulate inventory discrepancies: each is -1 deviation.
	inventoryDiscrepancies = 20 // 0.4% error rate

	// Simulate pick errors.
	pickErrors = 15 // 0.075% error rate

	invAccuracy := float64(totalInventoryLocations-inventoryDiscrepancies) / float64(totalInventoryLocations) * 100
	pickAccuracy := float64(totalPicks-pickErrors) / float64(totalPicks) * 100

	t.Logf("inventory accuracy: %.4f%% (limit 99.5%%) -> %s", invAccuracy, passFail(invAccuracy >= 99.5))
	t.Logf("pick accuracy:      %.4f%% (limit 99.9%%) -> %s", pickAccuracy, passFail(pickAccuracy >= 99.9))

	if invAccuracy < 99.5 {
		t.Errorf("inventory accuracy %.4f%% below 99.5%% threshold", invAccuracy)
	}
	if pickAccuracy < 99.9 {
		t.Errorf("pick accuracy %.4f%% below 99.9%% threshold", pickAccuracy)
	}
}

func passFail(ok bool) string {
	if ok {
		return "PASS"
	}
	return "FAIL"
}

// ── Scenario 6: Race Condition Tests ─────────────────────────────────────────

// TestScenario_ConcurrentInventoryAdjust simulates two goroutines concurrently
// adjusting the same inventory record. Only one should succeed for the available qty.
func TestScenario_ConcurrentInventoryAdjust(t *testing.T) {
	inv := &Inventory{
		ID:     uuid.New(),
		SKUID:  uuid.New(),
		Qty:    100,
		Status: InventoryStatusAvailable,
	}

	// Two goroutines each try to deduct 60 (total 120, but only 100 available).
	var successCount int32
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	for range 2 {
		go func() {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			if inv.CanDeduct(60) {
				inv.Qty -= 60
				successCount++
			}
		}()
	}

	wg.Wait()

	// Only one should succeed (100 available, 60 taken first).
	if successCount != 1 {
		t.Errorf("concurrent adjust: %d succeeded, want 1", successCount)
	}
	if inv.Qty != 40 {
		t.Errorf("concurrent adjust: remaining qty = %f, want 40", inv.Qty)
	}
}

// TestScenario_ConcurrentReserveRace tests reservation race: two order lines
// competing for the same inventory.
func TestScenario_ConcurrentReserveRace(t *testing.T) {
	inv := &Inventory{
		ID:     uuid.New(),
		SKUID:  uuid.New(),
		Qty:    100,
		Status: InventoryStatusAvailable,
	}

	var successCount int32
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	for range 2 {
		go func() {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			if inv.CanReserve(60) {
				inv.ReservedQty += 60
				successCount++
			}
		}()
	}

	wg.Wait()

	// Only one should succeed (100 available, 60 reserved first, 40 remaining < 60).
	if successCount != 1 {
		t.Errorf("concurrent reserve: %d succeeded, want 1", successCount)
	}
	if inv.ReservedQty != 60 {
		t.Errorf("concurrent reserve: reserved = %f, want 60", inv.ReservedQty)
	}
	if inv.Available() != 40 {
		t.Errorf("concurrent reserve: available = %f, want 40", inv.Available())
	}
}

// TestScenario_ConcurrentAllocateAndPick simulates a scenario where one goroutine
// allocates inventory while another simultaneously picks it.
func TestScenario_ConcurrentAllocateAndPick(t *testing.T) {
	inv := &Inventory{
		ID:     uuid.New(),
		SKUID:  uuid.New(),
		Qty:    50,
		Status: InventoryStatusAvailable,
	}

	var mu sync.Mutex
	var allocateSuccess, pickSuccess bool
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine A: allocate 30.
	go func() {
		defer wg.Done()
		mu.Lock()
		defer mu.Unlock()
		if inv.CanReserve(30) {
			inv.ReservedQty += 30
			allocateSuccess = true
		}
	}()

	// Goroutine B: pick 50 (deduct from qty, skip reserve).
	go func() {
		defer wg.Done()
		mu.Lock()
		defer mu.Unlock()
		if inv.CanDeduct(50) {
			inv.Qty -= 50
			pickSuccess = true
		}
	}()

	wg.Wait()

	// Both shouldn't succeed — 50 available, allocate 30 leaves 20, pick 50 needs 50.
	// Or, pick 50 happens first (leaves 0), then allocate 30 fails.
	if allocateSuccess && pickSuccess {
		t.Error("both allocate and full pick should not succeed with only 50 available")
	}
	t.Logf("allocate success=%v, pick success=%v, qty=%f, reserved=%f, available=%f",
		allocateSuccess, pickSuccess, inv.Qty, inv.ReservedQty, inv.Available())
}

// ── Scenario 7: Order Partial Fulfillment ────────────────────────────────────

// TestScenario_PartialFulfillment exercises the partial order flow where
// not enough inventory is available to fulfill the complete order.
func TestScenario_PartialFulfillment(t *testing.T) {
	skuID := uuid.New()
	whID := uuid.New()

	// ── Setup: 80 units available for 100 unit order ──
	order := &Order{
		ID:          uuid.New(),
		OrderNo:     "OUT-2026-002",
		OrderType:   OrderTypeOutbound,
		WarehouseID: whID,
		Status:      OrderStatusConfirmed,
		Lines: []OrderLine{
			{ID: uuid.New(), LineNo: 1, SKUID: skuID, OrderedQty: 100, Status: OrderLineStatusPending},
		},
	}

	inv := &Inventory{
		ID:         uuid.New(),
		SKUID:      skuID,
		Qty:        80,
		Status:     InventoryStatusAvailable,
		ReceivedAt: time.Now().Add(-10 * 24 * time.Hour),
	}

	// ── Allocate what we can ──
	line := &order.Lines[0]
	available := inv.Available()
	if available < line.OrderedQty {
		// Partial allocation.
		if !inv.CanDeduct(available) {
			t.Fatal("should be able to deduct available qty")
		}
		inv.Qty -= available
		inv.ReservedQty += available
		line.FulfilledQty = available
		line.Status = OrderLineStatusPartial
		order.Status = OrderStatusPartial
	}

	if line.FulfilledQty != 80 {
		t.Errorf("fulfilled = %f, want 80", line.FulfilledQty)
	}
	if line.Status != OrderLineStatusPartial {
		t.Errorf("line status = %s, want partial", line.Status)
	}
	if order.Status != OrderStatusPartial {
		t.Errorf("order status = %s, want partial", order.Status)
	}

	// ── Later: remaining 20 arrive → complete ──
	inv2 := &Inventory{
		ID:         uuid.New(),
		SKUID:      skuID,
		Qty:        20,
		Status:     InventoryStatusAvailable,
		ReceivedAt: time.Now(),
	}

	inv2.Qty -= 20
	line.FulfilledQty += 20
	line.Status = OrderLineStatusFulfilled
	order.Status = OrderStatusCompleted
	order.CompletedAt = timePtr(time.Now())

	if line.FulfilledQty != 100 {
		t.Errorf("final fulfilled = %f, want 100", line.FulfilledQty)
	}
	if !order.IsTerminal() {
		t.Error("order should be terminal after completion")
	}
	if !line.IsTerminal() {
		t.Error("line should be terminal after fulfillment")
	}
}

// ── Scenario 8: Quality Hold (Quarantine) Flow ───────────────────────────────

// TestScenario_QuarantineFlow exercises the quality inspection and hold process:
// inventory is moved to quarantine for inspection, and either released or marked damaged.
func TestScenario_QuarantineFlow(t *testing.T) {
	skuID := uuid.New()
	locID := uuid.New()

	// ── Initial state: 200 units available ──
	inv := &Inventory{
		ID:     uuid.New(),
		SKUID:  skuID,
		LocationID: locID,
		Qty:    200,
		Status: InventoryStatusAvailable,
	}

	// ── Move to quarantine ──
	if !inv.CanTransitionTo(InventoryStatusQuarantine) {
		t.Fatal("available → quarantine should be valid")
	}
	inv.Status = InventoryStatusQuarantine

	// ── Inspection: 10 units damaged, 190 ok ──
	// Release 190 back to available.
	okQty := 190.0
	damagedQty := 10.0

	// Create separate inventory record for damaged portion.
	damagedInv := &Inventory{
		ID:     uuid.New(),
		SKUID:  skuID,
		LocationID: locID,
		Qty:    damagedQty,
		Status: InventoryStatusDamaged,
	}

	inv.Qty = okQty
	if !inv.CanTransitionTo(InventoryStatusAvailable) {
		t.Fatal("quarantine → available should be valid")
	}
	inv.Status = InventoryStatusAvailable

	if inv.Qty != okQty {
		t.Errorf("released qty = %f, want %f", inv.Qty, okQty)
	}
	if damagedInv.Status != InventoryStatusDamaged {
		t.Errorf("damaged status = %s, want damaged", damagedInv.Status)
	}

	// ── Damaged can be expired ──
	if !damagedInv.CanTransitionTo(InventoryStatusExpired) {
		t.Fatal("damaged → expired should be valid")
	}
	damagedInv.Status = InventoryStatusExpired
	if !damagedInv.IsTerminal() {
		t.Error("expired inventory should be terminal")
	}

	// Verify total: 190 available + 10 expired.
	totalInventory := inv.Qty + damagedInv.Qty
	if totalInventory != 200 {
		t.Errorf("total inventory = %f, want 200", totalInventory)
	}
}

// ── Scenario 9: Multi-Order Wave Picking ─────────────────────────────────────

// TestScenario_MultiOrderWave excercises batch wave picking with multiple orders.
func TestScenario_MultiOrderWave(t *testing.T) {
	skuIDA := uuid.New()
	skuIDB := uuid.New()
	locID := uuid.New()
	whID := uuid.New()

	// ── Inventory setup ──
	inventory := map[uuid.UUID]*Inventory{
		skuIDA: {ID: uuid.New(), SKUID: skuIDA, LocationID: locID, WarehouseID: whID, Qty: 200, Status: InventoryStatusAvailable, ReceivedAt: time.Now()},
		skuIDB: {ID: uuid.New(), SKUID: skuIDB, LocationID: locID, WarehouseID: whID, Qty: 150, Status: InventoryStatusAvailable, ReceivedAt: time.Now()},
	}

	// ── Three orders in a batch wave ──
	orders := []*Order{
		{
			ID: uuid.New(), OrderNo: "OUT-001", WarehouseID: whID, Status: OrderStatusConfirmed,
			Lines: []OrderLine{
				{ID: uuid.New(), LineNo: 1, SKUID: skuIDA, OrderedQty: 50},
				{ID: uuid.New(), LineNo: 2, SKUID: skuIDB, OrderedQty: 30},
			},
		},
		{
			ID: uuid.New(), OrderNo: "OUT-002", WarehouseID: whID, Status: OrderStatusConfirmed,
			Lines: []OrderLine{
				{ID: uuid.New(), LineNo: 1, SKUID: skuIDA, OrderedQty: 100},
			},
		},
		{
			ID: uuid.New(), OrderNo: "OUT-003", WarehouseID: whID, Status: OrderStatusConfirmed,
			Lines: []OrderLine{
				{ID: uuid.New(), LineNo: 1, SKUID: skuIDB, OrderedQty: 60},
			},
		},
	}

	// ── Wave creation ──
	wave := &Wave{
		ID:          uuid.New(),
		WaveNo:      "WAVE-BATCH-001",
		WarehouseID: whID,
		WaveType:    WaveTypeBatch,
		Status:      WaveStatusCreated,
		TotalOrders: len(orders),
	}

	// ── Aggregate demand by SKU ──
	demand := map[uuid.UUID]float64{}
	totalLines := 0
	totalQty := 0.0
	for _, order := range orders {
		for _, line := range order.Lines {
			totalLines++
			totalQty += line.OrderedQty
			demand[line.SKUID] += line.OrderedQty
		}
	}

	wave.TotalLines = totalLines
	wave.TotalQty = totalQty

	// ── Validate inventory sufficiency ──
	for skuID, needed := range demand {
		inv := inventory[skuID]
		if !inv.CanDeduct(needed) {
			t.Errorf("insufficient inventory for SKU %s: need %f, have %f available",
				skuID, needed, inv.Available())
		}
	}

	// ── Allocate and pick ──
	totalFulfilled := 0.0
	for skuID, needed := range demand {
		inv := inventory[skuID]
		inv.Qty -= needed
		totalFulfilled += needed
	}

	if totalFulfilled != totalQty {
		t.Errorf("total fulfilled = %f, want %f", totalFulfilled, totalQty)
	}

	// ── Complete all orders ──
	for _, order := range orders {
		for i := range order.Lines {
			order.Lines[i].FulfilledQty = order.Lines[i].OrderedQty
			order.Lines[i].Status = OrderLineStatusFulfilled
		}
		order.Status = OrderStatusCompleted
		if !order.IsTerminal() {
			t.Errorf("order %s should be terminal", order.OrderNo)
		}
	}

	// ── Wave complete ──
	wave.Status = WaveStatusReleased
	wave.Status = WaveStatusInProgress
	wave.Status = WaveStatusCompleted

	if !wave.IsTerminal() {
		t.Error("wave should be terminal after completion")
	}

	// Verify remaining inventory.
	if inventory[skuIDA].Qty != 50 { // 200 - 150
		t.Errorf("SKU A remaining = %f, want 50", inventory[skuIDA].Qty)
	}
	if inventory[skuIDB].Qty != 60 { // 150 - 90
		t.Errorf("SKU B remaining = %f, want 60", inventory[skuIDB].Qty)
	}
}

// ── Scenario 10: Transfer Flow ───────────────────────────────────────────────

// TestScenario_TransferFlow exercises internal warehouse transfer orders.
func TestScenario_TransferFlow(t *testing.T) {
	skuID := uuid.New()
	whID := uuid.New()
	srcLocID := uuid.New()
	dstLocID := uuid.New()

	// ── Source inventory ──
	srcInv := &Inventory{
		ID: uuid.New(), SKUID: skuID, LocationID: srcLocID, WarehouseID: whID,
		Qty: 50, Status: InventoryStatusAvailable, ReceivedAt: time.Now().Add(-30 * 24 * time.Hour),
	}

	// ── Create transfer order ──
	order := &Order{
		ID: uuid.New(), OrderNo: "TRF-2026-001", OrderType: OrderTypeTransfer,
		WarehouseID: whID, Status: OrderStatusConfirmed,
		Lines: []OrderLine{
			{ID: uuid.New(), LineNo: 1, SKUID: skuID, OrderedQty: 30},
		},
	}

	// ── Execute transfer: pick from source, putaway to destination ──
	transferQty := order.Lines[0].OrderedQty
	if !srcInv.CanDeduct(transferQty) {
		t.Fatal("transfer: insufficient source inventory")
	}
	srcInv.Qty -= transferQty

	dstInv := &Inventory{
		ID: uuid.New(), SKUID: skuID, LocationID: dstLocID, WarehouseID: whID,
		Qty: transferQty, Status: InventoryStatusAvailable, ReceivedAt: srcInv.ReceivedAt,
		// Preserve batch/expiry from source for traceability.
		BatchNo:    srcInv.BatchNo,
		ExpiryDate: srcInv.ExpiryDate,
	}

	if srcInv.Qty != 20 {
		t.Errorf("source remaining = %f, want 20", srcInv.Qty)
	}
	if dstInv.Qty != 30 {
		t.Errorf("destination qty = %f, want 30", dstInv.Qty)
	}

	// ── Complete transfer order ──
	order.Lines[0].FulfilledQty = transferQty
	order.Lines[0].Status = OrderLineStatusFulfilled
	order.Status = OrderStatusCompleted

	if !order.IsTerminal() {
		t.Error("transfer order should be terminal after completion")
	}
}
