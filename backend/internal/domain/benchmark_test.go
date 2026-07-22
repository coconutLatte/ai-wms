// Package domain — WMS standard benchmark tests for core domain operations.
// Benchmarks measure throughput of state machines, inventory operations,
// FEFO/FIFO sorting, and task lifecycle transitions.
package domain

import (
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ── State Machine Benchmarks ──────────────────────────────────────────────────

// BenchmarkOrderStateMachine measures throughput of order status transitions.
func BenchmarkOrderStateMachine(b *testing.B) {
	orders := make([]*Order, b.N)
	for i := 0; i < b.N; i++ {
		orders[i] = &Order{Status: OrderStatusDraft}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o := orders[i]
		o.CanTransitionTo(OrderStatusConfirmed)
		o.Status = OrderStatusConfirmed
		o.CanTransitionTo(OrderStatusProcessing)
		o.Status = OrderStatusProcessing
		o.CanTransitionTo(OrderStatusCompleted)
		o.IsTerminal()
	}
}

// BenchmarkTaskStateMachine measures throughput of task lifecycle transitions.
func BenchmarkTaskStateMachine(b *testing.B) {
	tasks := make([]*Task, b.N)
	for i := 0; i < b.N; i++ {
		tasks[i] = &Task{Status: TaskStatusPending}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := tasks[i]
		t.CanTransitionTo(TaskStatusAssigned)
		t.Status = TaskStatusAssigned
		t.CanTransitionTo(TaskStatusInProgress)
		t.Status = TaskStatusInProgress
		t.CanTransitionTo(TaskStatusCompleted)
		t.IsTerminal()
	}
}

// BenchmarkInventoryStateMachine measures throughput of inventory status transitions.
func BenchmarkInventoryStateMachine(b *testing.B) {
	invs := make([]*Inventory, b.N)
	for i := 0; i < b.N; i++ {
		invs[i] = &Inventory{Status: InventoryStatusAvailable}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inv := invs[i]
		inv.CanTransitionTo(InventoryStatusQuarantine)
		inv.Status = InventoryStatusQuarantine
		inv.CanTransitionTo(InventoryStatusAvailable)
		inv.IsTerminal()
	}
}

// ── Inventory Operation Benchmarks ────────────────────────────────────────────

// BenchmarkInventoryCanDeduct measures throughput of deduction checks.
func BenchmarkInventoryCanDeduct(b *testing.B) {
	inv := &Inventory{Qty: 1000, ReservedQty: 200}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inv.CanDeduct(10)
	}
}

// BenchmarkInventoryCanReserve measures throughput of reservation checks.
func BenchmarkInventoryCanReserve(b *testing.B) {
	inv := &Inventory{Qty: 1000, ReservedQty: 100}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inv.CanReserve(50)
	}
}

// BenchmarkInventoryAvailable measures throughput of available qty calculation.
func BenchmarkInventoryAvailable(b *testing.B) {
	inv := &Inventory{Qty: 1000, ReservedQty: 300}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = inv.Available()
	}
}

// BenchmarkInventoryCanAdjustTo measures throughput of adjustment validation.
func BenchmarkInventoryCanAdjustTo(b *testing.B) {
	inv := &Inventory{Qty: 500}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inv.CanAdjustTo(50)
	}
}

// ── FEFO / FIFO Benchmarks ────────────────────────────────────────────────────

// BenchmarkFEFOSorting measures throughput of FEFO-ordered inventory sorting.
func BenchmarkFEFOSorting(b *testing.B) {
	b.StopTimer()
	now := time.Now()
	batches := make([]*Inventory, 1000)
	for i := 0; i < 1000; i++ {
		offset := time.Duration(rand.Intn(365*24)) * time.Hour
		expiry := now.Add(offset)
		batches[i] = &Inventory{
			ID:             uuid.New(),
			SKUID:          uuid.New(),
			Qty:            float64(rand.Intn(1000)),
			ReceivedAt:     now.Add(-time.Duration(i) * time.Hour),
			ExpiryDate:     &expiry,
		}
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		// Sort by FEFO: earlier expiry first, nil expiry last.
		sort.Slice(batches, func(a, b int) bool {
			return batches[a].HasEarlierExpiry(batches[b])
		})
	}
}

// BenchmarkFIFOSorting measures throughput of FIFO-ordered inventory sorting.
func BenchmarkFIFOSorting(b *testing.B) {
	b.StopTimer()
	now := time.Now()
	batches := make([]*Inventory, 1000)
	for i := 0; i < 1000; i++ {
		batches[i] = &Inventory{
			ID:         uuid.New(),
			SKUID:      uuid.New(),
			Qty:        float64(rand.Intn(1000)),
			ReceivedAt: now.Add(-time.Duration(i*17+rand.Intn(100)) * time.Minute),
		}
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		sort.Slice(batches, func(a, b int) bool {
			return batches[a].IsOlderThan(batches[b])
		})
	}
}

// ── Large-Scale Allocation Benchmarks ─────────────────────────────────────────

// BenchmarkAllocateInventory simulates the core allocation logic at scale.
// Given N order lines and M inventory batches, match using FEFO strategy.
func BenchmarkAllocateInventory(b *testing.B) {
	b.StopTimer()
	now := time.Now()
	skuID := uuid.New()

	// 100 order lines requesting various quantities.
	type demand struct {
		qty float64
	}
	demands := make([]demand, 100)
	for i := 0; i < 100; i++ {
		demands[i] = demand{qty: float64(1 + rand.Intn(50))}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		// Create fresh inventory batches each iteration.
		batches := make([]*Inventory, 500)
		for j := 0; j < 500; j++ {
			expiry := now.Add(time.Duration(1+rand.Intn(365)) * 24 * time.Hour)
			batches[j] = &Inventory{
				ID:         uuid.New(),
				SKUID:      skuID,
				Qty:        float64(1 + rand.Intn(200)),
				ReservedQty: 0,
				ReceivedAt: now.Add(-time.Duration(j) * time.Hour),
				ExpiryDate: &expiry,
			}
		}

		// Allocate each demand against available inventory (FEFO).
		for _, d := range demands {
			remaining := d.qty
			for _, batch := range batches {
				if batch.SKUID != skuID {
					continue
				}
				avail := batch.Available()
				if avail <= 0 {
					continue
				}
				if avail >= remaining {
					batch.ReservedQty += remaining
					remaining = 0
					break
				}
				batch.ReservedQty += avail
				remaining -= avail
			}
			_ = remaining
		}
	}
}

// ── Concurrent Operation Benchmarks ───────────────────────────────────────────

// BenchmarkParallelInventoryAdjust simulates concurrent inventory adjustments
// across many SKU-location pairs (parallel goroutine scenario).
func BenchmarkParallelInventoryAdjust(b *testing.B) {
	b.StopTimer()
	invs := make([]*Inventory, 1000)
	now := time.Now()
	for i := 0; i < 1000; i++ {
		invs[i] = &Inventory{
			ID:         uuid.New(),
			SKUID:      uuid.New(),
			LocationID: uuid.New(),
			Qty:        100,
			ReservedQty: 0,
			Status:     InventoryStatusAvailable,
			ReceivedAt: now,
		}
	}

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			inv := invs[idx%len(invs)]
			// Simulate an adjustment: validate and apply.
			if inv.CanAdjustTo(-1) {
				inv.Qty--
			}
			idx++
		}
	})
}

// BenchmarkParallelTaskTransition simulates concurrent task status transitions.
func BenchmarkParallelTaskTransition(b *testing.B) {
	b.StopTimer()
	tasks := make([]*Task, 1000)
	for i := 0; i < 1000; i++ {
		tasks[i] = &Task{
			ID:       uuid.New(),
			Status:   TaskStatusPending,
			TaskType: TaskTypePick,
		}
	}

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			t := tasks[idx%len(tasks)]
			// Simple lifecycle: pending → assigned → in_progress.
			switch t.Status {
			case TaskStatusPending:
				if t.CanTransitionTo(TaskStatusAssigned) {
					t.Status = TaskStatusAssigned
				}
			case TaskStatusAssigned:
				if t.CanTransitionTo(TaskStatusInProgress) {
					t.Status = TaskStatusInProgress
				}
			default:
				t.IsTerminal()
			}
			idx++
		}
	})
}
