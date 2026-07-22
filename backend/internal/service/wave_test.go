package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// mockRepoForWave implements repository.TaskRepository for WaveService testing.
// It only provides the Wave-related methods; Task methods panic if called.
type mockRepoForWave struct {
	waves map[uuid.UUID]*domain.Wave
}

func newMockRepoForWave() *mockRepoForWave {
	return &mockRepoForWave{
		waves: make(map[uuid.UUID]*domain.Wave),
	}
}

// ── Task methods (unused, will panic) ────────────────────────

func (m *mockRepoForWave) CreateTask(ctx context.Context, t *domain.Task) error {
	panic("not implemented")
}
func (m *mockRepoForWave) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	panic("not implemented")
}
func (m *mockRepoForWave) GetTasksByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.Task, error) {
	panic("not implemented")
}
func (m *mockRepoForWave) ListTasks(ctx context.Context, filter repository.TaskFilter) ([]*domain.Task, error) {
	panic("not implemented")
}
func (m *mockRepoForWave) AssignTask(ctx context.Context, id uuid.UUID, assignedTo string) error {
	panic("not implemented")
}
func (m *mockRepoForWave) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status domain.TaskStatus) error {
	panic("not implemented")
}
func (m *mockRepoForWave) CompleteTask(ctx context.Context, id uuid.UUID, actualQty float64, toLocationID *uuid.UUID) error {
	panic("not implemented")
}
func (m *mockRepoForWave) CountTasks(ctx context.Context, filter repository.TaskFilter) (int, error) {
	panic("not implemented")
}

func (m *mockRepoForWave) CountTasksByStatus(ctx context.Context) (map[domain.TaskStatus]int, error) {
	panic("not implemented")
}

// ── Wave methods ────────────────────────────────────────────

func (m *mockRepoForWave) CreateWave(ctx context.Context, w *domain.Wave) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	clone := *w
	m.waves[w.ID] = &clone
	return nil
}

func (m *mockRepoForWave) GetWave(ctx context.Context, id uuid.UUID) (*domain.Wave, error) {
	w, ok := m.waves[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("wave", id.String())
	}
	clone := *w
	return &clone, nil
}

func (m *mockRepoForWave) ListWaves(ctx context.Context, filter repository.WaveFilter) ([]*domain.Wave, error) {
	var result []*domain.Wave
	for _, w := range m.waves {
		if filter.WarehouseID != uuid.Nil && w.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.Status != "" && w.Status != filter.Status {
			continue
		}
		if filter.WaveType != "" && w.WaveType != filter.WaveType {
			continue
		}
		clone := *w
		result = append(result, &clone)
	}
	return result, nil
}

func (m *mockRepoForWave) UpdateWaveStatus(ctx context.Context, id uuid.UUID, status domain.WaveStatus) error {
	w, ok := m.waves[id]
	if !ok {
		return pkgerrors.NewNotFound("wave", id.String())
	}
	w.Status = status
	return nil
}

func (m *mockRepoForWave) AddWaveOrders(ctx context.Context, id uuid.UUID, orderIDs []uuid.UUID) error {
	w, ok := m.waves[id]
	if !ok {
		return pkgerrors.NewNotFound("wave", id.String())
	}
	// Dedup: only add order IDs not already present.
	seen := make(map[uuid.UUID]bool)
	for _, oid := range w.OrderIDs {
		seen[oid] = true
	}
	for _, oid := range orderIDs {
		if !seen[oid] {
			w.OrderIDs = append(w.OrderIDs, oid)
			seen[oid] = true
		}
	}
	w.TotalOrders = len(w.OrderIDs)
	return nil
}

func (m *mockRepoForWave) RemoveWaveOrders(ctx context.Context, id uuid.UUID, orderIDs []uuid.UUID) error {
	w, ok := m.waves[id]
	if !ok {
		return pkgerrors.NewNotFound("wave", id.String())
	}
	toRemove := make(map[uuid.UUID]bool)
	for _, oid := range orderIDs {
		toRemove[oid] = true
	}
	filtered := w.OrderIDs[:0]
	for _, oid := range w.OrderIDs {
		if !toRemove[oid] {
			filtered = append(filtered, oid)
		}
	}
	w.OrderIDs = filtered
	w.TotalOrders = len(w.OrderIDs)
	return nil
}

func (m *mockRepoForWave) CountWaves(ctx context.Context, filter repository.WaveFilter) (int, error) {
	count := 0
	for _, w := range m.waves {
		if filter.WarehouseID != uuid.Nil && w.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.Status != "" && w.Status != filter.Status {
			continue
		}
		if filter.WaveType != "" && w.WaveType != filter.WaveType {
			continue
		}
		count++
	}
	return count, nil
}

// ── Tests ───────────────────────────────────────────────────

func TestWaveService_CreateWave(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepoForWave()
	svc := NewWaveService(repo)

	whID := uuid.New()
	input := CreateWaveInput{
		WaveType:    domain.WaveTypeBatch,
		WarehouseID: whID,
		OrderIDs:    []uuid.UUID{uuid.New(), uuid.New()},
	}

	wave, err := svc.CreateWave(ctx, input)
	if err != nil {
		t.Fatalf("CreateWave failed: %v", err)
	}

	if wave.WaveNo == "" {
		t.Error("expected WaveNo to be auto-generated")
	}
	if wave.WarehouseID != whID {
		t.Errorf("expected warehouse_id %s, got %s", whID, wave.WarehouseID)
	}
	if wave.Status != domain.WaveStatusCreated {
		t.Errorf("expected status created, got %s", wave.Status)
	}
	if wave.TotalOrders != 2 {
		t.Errorf("expected TotalOrders 2, got %d", wave.TotalOrders)
	}
	if len(wave.OrderIDs) != 2 {
		t.Errorf("expected 2 order IDs, got %d", len(wave.OrderIDs))
	}

	// Verify it's retrievable.
	retrieved, err := svc.GetWave(ctx, wave.ID)
	if err != nil {
		t.Fatalf("GetWave after create failed: %v", err)
	}
	if retrieved.ID != wave.ID {
		t.Error("retrieved wave ID mismatch")
	}
}

func TestWaveService_CreateWave_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewWaveService(newMockRepoForWave())

	tests := []struct {
		name  string
		input CreateWaveInput
	}{
		{"empty warehouse_id", CreateWaveInput{WaveType: domain.WaveTypeBatch}},
		{"invalid wave_type", CreateWaveInput{WarehouseID: uuid.New(), WaveType: "invalid"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateWave(ctx, tt.input)
			if err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

func TestWaveService_GetWave_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewWaveService(newMockRepoForWave())

	_, err := svc.GetWave(ctx, uuid.New())
	if err == nil {
		t.Error("expected error for non-existent wave")
	}
}

func TestWaveService_ListWaves(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepoForWave()
	svc := NewWaveService(repo)

	wh1 := uuid.New()
	wh2 := uuid.New()

	// Create 3 waves in wh1, 2 in wh2.
	for i := 0; i < 3; i++ {
		_, err := svc.CreateWave(ctx, CreateWaveInput{
			WaveType:    domain.WaveTypeBatch,
			WarehouseID: wh1,
		})
		if err != nil {
			t.Fatalf("CreateWave failed: %v", err)
		}
	}
	for i := 0; i < 2; i++ {
		_, err := svc.CreateWave(ctx, CreateWaveInput{
			WaveType:    domain.WaveTypeZone,
			WarehouseID: wh2,
		})
		if err != nil {
			t.Fatalf("CreateWave failed: %v", err)
		}
	}

	// List wh1.
	waves, total, err := svc.ListWaves(ctx, WaveQueryParams{WarehouseID: wh1, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListWaves failed: %v", err)
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	if len(waves) != 3 {
		t.Errorf("expected 3 waves, got %d", len(waves))
	}

	// List wh2.
	waves, total, err = svc.ListWaves(ctx, WaveQueryParams{WarehouseID: wh2, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListWaves failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}

	// List all (no warehouse filter).
	waves, total, err = svc.ListWaves(ctx, WaveQueryParams{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListWaves all failed: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5 across all warehouses, got %d", total)
	}
}

func TestWaveService_UpdateWaveStatus(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepoForWave()
	svc := NewWaveService(repo)

	wave, err := svc.CreateWave(ctx, CreateWaveInput{
		WaveType:    domain.WaveTypeBatch,
		WarehouseID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("CreateWave failed: %v", err)
	}

	// Valid: created → released
	updated, err := svc.UpdateWaveStatus(ctx, wave.ID, UpdateWaveStatusInput{Status: domain.WaveStatusReleased})
	if err != nil {
		t.Fatalf("UpdateWaveStatus to released failed: %v", err)
	}
	if updated.Status != domain.WaveStatusReleased {
		t.Errorf("expected released, got %s", updated.Status)
	}

	// Valid: released → in_progress
	updated, err = svc.UpdateWaveStatus(ctx, wave.ID, UpdateWaveStatusInput{Status: domain.WaveStatusInProgress})
	if err != nil {
		t.Fatalf("UpdateWaveStatus to in_progress failed: %v", err)
	}
	if updated.Status != domain.WaveStatusInProgress {
		t.Errorf("expected in_progress, got %s", updated.Status)
	}

	// Valid: in_progress → completed
	updated, err = svc.UpdateWaveStatus(ctx, wave.ID, UpdateWaveStatusInput{Status: domain.WaveStatusCompleted})
	if err != nil {
		t.Fatalf("UpdateWaveStatus to completed failed: %v", err)
	}
	if updated.Status != domain.WaveStatusCompleted {
		t.Errorf("expected completed, got %s", updated.Status)
	}
}

func TestWaveService_UpdateWaveStatus_InvalidTransition(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepoForWave()
	svc := NewWaveService(repo)

	wave, err := svc.CreateWave(ctx, CreateWaveInput{
		WaveType:    domain.WaveTypeBatch,
		WarehouseID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("CreateWave failed: %v", err)
	}

	// Invalid: created → in_progress (must go through released first)
	_, err = svc.UpdateWaveStatus(ctx, wave.ID, UpdateWaveStatusInput{Status: domain.WaveStatusInProgress})
	if err == nil {
		t.Error("expected error for invalid transition created → in_progress")
	}
	if !pkgerrors.IsInvalidStatus(err) {
		t.Errorf("expected InvalidStatus error, got %v", err)
	}
}

func TestWaveService_UpdateWaveStatus_Terminal(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepoForWave()
	svc := NewWaveService(repo)

	wave, err := svc.CreateWave(ctx, CreateWaveInput{
		WaveType:    domain.WaveTypeBatch,
		WarehouseID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("CreateWave failed: %v", err)
	}

	// Progress to completed.
	svc.UpdateWaveStatus(ctx, wave.ID, UpdateWaveStatusInput{Status: domain.WaveStatusReleased})
	svc.UpdateWaveStatus(ctx, wave.ID, UpdateWaveStatusInput{Status: domain.WaveStatusInProgress})
	svc.UpdateWaveStatus(ctx, wave.ID, UpdateWaveStatusInput{Status: domain.WaveStatusCompleted})

	// Terminal: completed → anything should fail.
	_, err = svc.UpdateWaveStatus(ctx, wave.ID, UpdateWaveStatusInput{Status: domain.WaveStatusReleased})
	if err == nil {
		t.Error("expected error for transition from terminal completed")
	}
}

func TestWaveService_ReleaseWave(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepoForWave()
	svc := NewWaveService(repo)

	wave, err := svc.CreateWave(ctx, CreateWaveInput{
		WaveType:    domain.WaveTypeBatch,
		WarehouseID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("CreateWave failed: %v", err)
	}

	// ReleaseWave convenience method.
	released, err := svc.ReleaseWave(ctx, wave.ID)
	if err != nil {
		t.Fatalf("ReleaseWave failed: %v", err)
	}
	if released.Status != domain.WaveStatusReleased {
		t.Errorf("expected released, got %s", released.Status)
	}
}

func TestWaveService_AddWaveOrders(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepoForWave()
	svc := NewWaveService(repo)

	order1 := uuid.New()
	order2 := uuid.New()
	order3 := uuid.New()

	wave, err := svc.CreateWave(ctx, CreateWaveInput{
		WaveType:    domain.WaveTypeBatch,
		WarehouseID: uuid.New(),
		OrderIDs:    []uuid.UUID{order1},
	})
	if err != nil {
		t.Fatalf("CreateWave failed: %v", err)
	}

	// Add orders.
	updated, err := svc.AddWaveOrders(ctx, wave.ID, AddWaveOrdersInput{OrderIDs: []uuid.UUID{order2, order3}})
	if err != nil {
		t.Fatalf("AddWaveOrders failed: %v", err)
	}
	if updated.TotalOrders != 3 {
		t.Errorf("expected TotalOrders 3, got %d", updated.TotalOrders)
	}
	if len(updated.OrderIDs) != 3 {
		t.Errorf("expected 3 OrderIDs, got %d", len(updated.OrderIDs))
	}
}

func TestWaveService_AddWaveOrders_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewWaveService(newMockRepoForWave())

	_, err := svc.AddWaveOrders(ctx, uuid.New(), AddWaveOrdersInput{OrderIDs: nil})
	if err == nil {
		t.Error("expected validation error for empty order_ids")
	}

	_, err = svc.AddWaveOrders(ctx, uuid.New(), AddWaveOrdersInput{OrderIDs: []uuid.UUID{}})
	if err == nil {
		t.Error("expected validation error for empty order_ids slice")
	}
}

func TestWaveService_AddWaveOrders_WrongStatus(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepoForWave()
	svc := NewWaveService(repo)

	wave, err := svc.CreateWave(ctx, CreateWaveInput{
		WaveType:    domain.WaveTypeBatch,
		WarehouseID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("CreateWave failed: %v", err)
	}

	// Release the wave first.
	svc.ReleaseWave(ctx, wave.ID)

	// Now try to add orders — should fail because wave is released.
	_, err = svc.AddWaveOrders(ctx, wave.ID, AddWaveOrdersInput{OrderIDs: []uuid.UUID{uuid.New()}})
	if err == nil {
		t.Error("expected error when adding orders to non-created wave")
	}
}

func TestWaveService_RemoveWaveOrders(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepoForWave()
	svc := NewWaveService(repo)

	order1 := uuid.New()
	order2 := uuid.New()
	order3 := uuid.New()

	wave, err := svc.CreateWave(ctx, CreateWaveInput{
		WaveType:    domain.WaveTypeBatch,
		WarehouseID: uuid.New(),
		OrderIDs:    []uuid.UUID{order1, order2, order3},
	})
	if err != nil {
		t.Fatalf("CreateWave failed: %v", err)
	}

	// Remove order2.
	updated, err := svc.RemoveWaveOrders(ctx, wave.ID, RemoveWaveOrdersInput{OrderIDs: []uuid.UUID{order2}})
	if err != nil {
		t.Fatalf("RemoveWaveOrders failed: %v", err)
	}
	if updated.TotalOrders != 2 {
		t.Errorf("expected TotalOrders 2, got %d", updated.TotalOrders)
	}

	// Verify order2 is gone.
	found := false
	for _, oid := range updated.OrderIDs {
		if oid == order2 {
			found = true
			break
		}
	}
	if found {
		t.Error("order2 should have been removed")
	}
}

func TestWaveService_RemoveWaveOrders_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewWaveService(newMockRepoForWave())

	_, err := svc.RemoveWaveOrders(ctx, uuid.New(), RemoveWaveOrdersInput{OrderIDs: nil})
	if err == nil {
		t.Error("expected validation error for empty order_ids")
	}
}
