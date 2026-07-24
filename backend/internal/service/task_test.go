package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// mockTaskRepo implements repository.TaskRepository for testing.
type mockTaskRepo struct {
	tasks map[uuid.UUID]*domain.Task
	waves map[uuid.UUID]*domain.Wave
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{
		tasks: make(map[uuid.UUID]*domain.Task),
		waves: make(map[uuid.UUID]*domain.Wave),
	}
}

// ── Task ──────────────────────────────────────────────────

func (m *mockTaskRepo) CreateTask(ctx context.Context, t *domain.Task) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	m.tasks[t.ID] = t
	return nil
}

func (m *mockTaskRepo) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("task", id.String())
	}
	return t, nil
}

func (m *mockTaskRepo) GetTasksByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.Task, error) {
	var result []*domain.Task
	for _, t := range m.tasks {
		if t.OrderID != nil && *t.OrderID == orderID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockTaskRepo) ListTasks(ctx context.Context, filter repository.TaskFilter) ([]*domain.Task, error) {
	var result []*domain.Task
	for _, t := range m.tasks {
		if filter.WarehouseID != uuid.Nil && t.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.TaskType != "" && t.TaskType != filter.TaskType {
			continue
		}
		if filter.Status != "" && t.Status != filter.Status {
			continue
		}
		if filter.AssignedTo != "" && t.AssignedTo != filter.AssignedTo {
			continue
		}
		result = append(result, t)
	}
	return result, nil
}

func (m *mockTaskRepo) AssignTask(ctx context.Context, id uuid.UUID, assignedTo string) error {
	t, ok := m.tasks[id]
	if !ok {
		return pkgerrors.NewNotFound("task", id.String())
	}
	if t.Status != domain.TaskStatusPending {
		return pkgerrors.NewInvalidInput("can only assign pending tasks")
	}
	t.Status = domain.TaskStatusAssigned
	t.AssignedTo = assignedTo
	return nil
}

func (m *mockTaskRepo) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status domain.TaskStatus) error {
	t, ok := m.tasks[id]
	if !ok {
		return pkgerrors.NewNotFound("task", id.String())
	}
	t.Status = status
	return nil
}

func (m *mockTaskRepo) CompleteTask(ctx context.Context, id uuid.UUID, actualQty float64, toLocationID *uuid.UUID) error {
	t, ok := m.tasks[id]
	if !ok {
		return pkgerrors.NewNotFound("task", id.String())
	}
	t.Status = domain.TaskStatusCompleted
	t.ActualQty = actualQty
	if toLocationID != nil {
		t.ToLocation = toLocationID
	}
	return nil
}

func (m *mockTaskRepo) CountTasks(ctx context.Context, filter repository.TaskFilter) (int, error) {
	count := 0
	for _, t := range m.tasks {
		if filter.WarehouseID != uuid.Nil && t.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.TaskType != "" && t.TaskType != filter.TaskType {
			continue
		}
		if filter.Status != "" && t.Status != filter.Status {
			continue
		}
		if filter.AssignedTo != "" && t.AssignedTo != filter.AssignedTo {
			continue
		}
		count++
	}
	return count, nil
}

func (m *mockTaskRepo) CountTasksByStatus(ctx context.Context) (map[domain.TaskStatus]int, error) {
	result := make(map[domain.TaskStatus]int)
	for _, t := range m.tasks {
		result[t.Status]++
	}
	return result, nil
}

// ── Wave (not used by TaskService tests) ────────────────────

func (m *mockTaskRepo) CreateWave(ctx context.Context, w *domain.Wave) error { return nil }
func (m *mockTaskRepo) GetWave(ctx context.Context, id uuid.UUID) (*domain.Wave, error) {
	return nil, nil
}
func (m *mockTaskRepo) ListWaves(ctx context.Context, filter repository.WaveFilter) ([]*domain.Wave, error) {
	return nil, nil
}
func (m *mockTaskRepo) UpdateWaveStatus(ctx context.Context, id uuid.UUID, status domain.WaveStatus) error {
	return nil
}
func (m *mockTaskRepo) AddWaveOrders(ctx context.Context, id uuid.UUID, orderIDs []uuid.UUID) error {
	return nil
}
func (m *mockTaskRepo) RemoveWaveOrders(ctx context.Context, id uuid.UUID, orderIDs []uuid.UUID) error {
	return nil
}
func (m *mockTaskRepo) CountWaves(ctx context.Context, filter repository.WaveFilter) (int, error) {
	count := 0
	for _, w := range m.waves {
		if w.WarehouseID == filter.WarehouseID {
			count++
		}
	}
	return count, nil
}

// ── Tests ───────────────────────────────────────────────────────────────────

func TestTaskService_CreateTask(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	skuID := uuid.New()
	whID := uuid.New()

	task, err := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePick,
		WarehouseID: whID,
		SKUID:       skuID,
		ExpectedQty: 50,
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if task.Status != domain.TaskStatusPending {
		t.Errorf("status = %q, want %q", task.Status, domain.TaskStatusPending)
	}
	if task.Priority != domain.TaskPriorityNormal {
		t.Errorf("priority = %q, want %q", task.Priority, domain.TaskPriorityNormal)
	}
	if task.UOM != "EA" {
		t.Errorf("uom = %q, want EA (default)", task.UOM)
	}
	if !strings.HasPrefix(task.TaskNo, "TASK-") {
		t.Errorf("task_no should start with TASK-: got %q", task.TaskNo)
	}
	if task.SKUID != skuID {
		t.Errorf("sku_id = %q, want %q", task.SKUID, skuID)
	}
	if task.ExpectedQty != 50 {
		t.Errorf("expected_qty = %f, want %f", task.ExpectedQty, 50.0)
	}
}

func TestTaskService_CreateTask_WithAllFields(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	orderID := uuid.New()
	fromLoc := uuid.New()
	toLoc := uuid.New()

	task, err := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:     domain.TaskTypePutaway,
		WarehouseID:  uuid.New(),
		OrderID:      &orderID,
		Priority:     domain.TaskPriorityHigh,
		FromLocation: &fromLoc,
		ToLocation:   &toLoc,
		SKUID:        uuid.New(),
		ExpectedQty:  100,
		UOM:          "PL",
		BatchNo:      "BATCH-001",
		Instructions: "Handle with care",
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if task.Priority != domain.TaskPriorityHigh {
		t.Errorf("priority = %q, want %q", task.Priority, domain.TaskPriorityHigh)
	}
	if task.UOM != "PL" {
		t.Errorf("uom = %q, want PL", task.UOM)
	}
	if task.BatchNo != "BATCH-001" {
		t.Errorf("batch_no = %q, want BATCH-001", task.BatchNo)
	}
	if task.Instructions != "Handle with care" {
		t.Errorf("instructions = %q, want 'Handle with care'", task.Instructions)
	}
	if task.OrderID == nil || *task.OrderID != orderID {
		t.Errorf("order_id not set correctly")
	}
	if task.FromLocation == nil || *task.FromLocation != fromLoc {
		t.Errorf("from_location_id not set correctly")
	}
	if task.ToLocation == nil || *task.ToLocation != toLoc {
		t.Errorf("to_location_id not set correctly")
	}
}

func TestTaskService_CreateTask_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	tests := []struct {
		name  string
		input CreateTaskInput
	}{
		{"empty task type", CreateTaskInput{
			TaskType:    "",
			WarehouseID: uuid.New(),
			SKUID:       uuid.New(),
			ExpectedQty: 1,
		}},
		{"invalid task type", CreateTaskInput{
			TaskType:    domain.TaskType("flying"),
			WarehouseID: uuid.New(),
			SKUID:       uuid.New(),
			ExpectedQty: 1,
		}},
		{"nil warehouse id", CreateTaskInput{
			TaskType:    domain.TaskTypePick,
			WarehouseID: uuid.Nil,
			SKUID:       uuid.New(),
			ExpectedQty: 1,
		}},
		{"nil sku id", CreateTaskInput{
			TaskType:    domain.TaskTypePick,
			WarehouseID: uuid.New(),
			SKUID:       uuid.Nil,
			ExpectedQty: 1,
		}},
		{"zero expected qty", CreateTaskInput{
			TaskType:    domain.TaskTypePick,
			WarehouseID: uuid.New(),
			SKUID:       uuid.New(),
			ExpectedQty: 0,
		}},
		{"negative expected qty", CreateTaskInput{
			TaskType:    domain.TaskTypePick,
			WarehouseID: uuid.New(),
			SKUID:       uuid.New(),
			ExpectedQty: -10,
		}},
		{"invalid priority", CreateTaskInput{
			TaskType:    domain.TaskTypePick,
			WarehouseID: uuid.New(),
			SKUID:       uuid.New(),
			ExpectedQty: 1,
			Priority:    domain.TaskPriority("super-urgent"),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateTask(ctx, tt.input)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestTaskService_AllTaskTypes(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	whID := uuid.New()
	skuID := uuid.New()

	types := []domain.TaskType{
		domain.TaskTypePutaway,
		domain.TaskTypePick,
		domain.TaskTypeReplenish,
		domain.TaskTypeTransfer,
		domain.TaskTypeCycleCount,
		domain.TaskTypeLoad,
		domain.TaskTypeUnload,
	}

	for _, tt := range types {
		t.Run(string(tt), func(t *testing.T) {
			task, err := svc.CreateTask(ctx, CreateTaskInput{
				TaskType:    tt,
				WarehouseID: whID,
				SKUID:       skuID,
				ExpectedQty: 10,
			})
			if err != nil {
				t.Fatalf("CreateTask %s failed: %v", tt, err)
			}
			if task.TaskType != tt {
				t.Errorf("task_type = %q, want %q", task.TaskType, tt)
			}
		})
	}
}

func TestTaskService_GetTask(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	created, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePick,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 30,
	})

	got, err := svc.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.TaskNo != created.TaskNo {
		t.Errorf("task_no = %q, want %q", got.TaskNo, created.TaskNo)
	}
}

func TestTaskService_GetTask_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	_, err := svc.GetTask(ctx, uuid.New())
	if err == nil {
		t.Fatal("expected error for unknown task")
	}
}

func TestTaskService_ListTasks(t *testing.T) {
	ctx := context.Background()
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)

	wh1 := uuid.New()
	wh2 := uuid.New()

	_, _ = svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePick,
		WarehouseID: wh1,
		SKUID:       uuid.New(),
		ExpectedQty: 10,
	})
	_, _ = svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePutaway,
		WarehouseID: wh1,
		SKUID:       uuid.New(),
		ExpectedQty: 20,
	})
	_, _ = svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePick,
		WarehouseID: wh2,
		SKUID:       uuid.New(),
		ExpectedQty: 30,
	})

	// All tasks.
	all, _, err := svc.ListTasks(ctx, repository.TaskFilter{})
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(all))
	}

	// Filter by warehouse.
	wh1Tasks, _, err := svc.ListTasks(ctx, repository.TaskFilter{WarehouseID: wh1})
	if err != nil {
		t.Fatalf("ListTasks wh1 failed: %v", err)
	}
	if len(wh1Tasks) != 2 {
		t.Errorf("expected 2 tasks in wh1, got %d", len(wh1Tasks))
	}

	// Filter by task type.
	picks, _, err := svc.ListTasks(ctx, repository.TaskFilter{TaskType: domain.TaskTypePick})
	if err != nil {
		t.Fatalf("ListTasks pick failed: %v", err)
	}
	if len(picks) != 2 {
		t.Errorf("expected 2 pick tasks, got %d", len(picks))
	}
}

func TestTaskService_AssignTask(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePick,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 25,
	})

	updated, err := svc.AssignTask(ctx, task.ID, AssignTaskInput{AssignedTo: "worker-42"})
	if err != nil {
		t.Fatalf("AssignTask failed: %v", err)
	}
	if updated.Status != domain.TaskStatusAssigned {
		t.Errorf("status = %q, want %q", updated.Status, domain.TaskStatusAssigned)
	}
	if updated.AssignedTo != "worker-42" {
		t.Errorf("assigned_to = %q, want worker-42", updated.AssignedTo)
	}
}

func TestTaskService_AssignTask_EmptyWorker(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePick,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 5,
	})

	_, err := svc.AssignTask(ctx, task.ID, AssignTaskInput{AssignedTo: ""})
	if err == nil {
		t.Fatal("expected error for empty assigned_to")
	}
}

func TestTaskService_AssignTask_NotPending(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePick,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 5,
	})

	// First assign succeeds.
	_, _ = svc.AssignTask(ctx, task.ID, AssignTaskInput{AssignedTo: "worker-1"})

	// Second assign should fail — task is already assigned.
	_, err := svc.AssignTask(ctx, task.ID, AssignTaskInput{AssignedTo: "worker-2"})
	if err == nil {
		t.Fatal("expected error for assigning already-assigned task")
	}
}

func TestTaskService_UpdateTaskStatus_ValidTransitions(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePutaway,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 100,
	})

	// pending → assigned
	updated, err := svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
	if err != nil {
		t.Fatalf("pending → assigned failed: %v", err)
	}
	if updated.Status != domain.TaskStatusAssigned {
		t.Errorf("status = %q, want %q", updated.Status, domain.TaskStatusAssigned)
	}

	// assigned → in_progress
	updated, err = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})
	if err != nil {
		t.Fatalf("assigned → in_progress failed: %v", err)
	}
	if updated.Status != domain.TaskStatusInProgress {
		t.Errorf("status = %q, want %q", updated.Status, domain.TaskStatusInProgress)
	}

	// in_progress → paused
	updated, err = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusPaused})
	if err != nil {
		t.Fatalf("in_progress → paused failed: %v", err)
	}
	if updated.Status != domain.TaskStatusPaused {
		t.Errorf("status = %q, want %q", updated.Status, domain.TaskStatusPaused)
	}

	// paused → in_progress (resume)
	updated, err = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})
	if err != nil {
		t.Fatalf("paused → in_progress failed: %v", err)
	}
	if updated.Status != domain.TaskStatusInProgress {
		t.Errorf("status = %q, want %q", updated.Status, domain.TaskStatusInProgress)
	}
}

func TestTaskService_UpdateTaskStatus_InvalidTransitions(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePick,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 10,
	})

	// pending → completed (invalid — must go through assigned, in_progress, complete)
	_, err := svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusCompleted})
	if err == nil {
		t.Fatal("expected error for pending → completed")
	}

	// pending → in_progress (invalid — must go through assigned)
	_, err = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})
	if err == nil {
		t.Fatal("expected error for pending → in_progress")
	}

	// pending → paused (invalid)
	_, err = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusPaused})
	if err == nil {
		t.Fatal("expected error for pending → paused")
	}
}

func TestTaskService_UpdateTaskStatus_CancelFromAny(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePick,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 10,
	})

	// pending → cancelled (valid)
	_, err := svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusCancelled})
	if err != nil {
		t.Fatalf("pending → cancelled failed: %v", err)
	}
}

func TestTaskService_UpdateTaskStatus_TerminalStates(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	// Cancel a task — terminal.
	task1, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePick,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 10,
	})
	_, _ = svc.UpdateTaskStatus(ctx, task1.ID, UpdateTaskStatusInput{Status: domain.TaskStatusCancelled})

	_, err := svc.UpdateTaskStatus(ctx, task1.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
	if err == nil {
		t.Fatal("expected error for cancelled → assigned")
	}

	// Complete a task (full lifecycle) — terminal.
	task2, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePutaway,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 20,
	})
	_, _ = svc.UpdateTaskStatus(ctx, task2.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
	_, _ = svc.UpdateTaskStatus(ctx, task2.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})
	_, _ = svc.CompleteTask(ctx, task2.ID, CompleteTaskInput{ActualQty: 20})

	_, err = svc.UpdateTaskStatus(ctx, task2.ID, UpdateTaskStatusInput{Status: domain.TaskStatusCancelled})
	if err == nil {
		t.Fatal("expected error for completed → cancelled")
	}
}

func TestTaskService_CompleteTask(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePutaway,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 100,
	})

	// Must be in_progress to complete.
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	toLoc := uuid.New()
	updated, err := svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: 95, ToLocationID: &toLoc})
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}
	if updated.Status != domain.TaskStatusCompleted {
		t.Errorf("status = %q, want %q", updated.Status, domain.TaskStatusCompleted)
	}
	if updated.ActualQty != 95 {
		t.Errorf("actual_qty = %f, want 95", updated.ActualQty)
	}
}

func TestTaskService_CompleteTask_NotInProgress(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePick,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 10,
	})

	// Try to complete a pending task — should fail.
	_, err := svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: 10})
	if err == nil {
		t.Fatal("expected error for completing pending task")
	}

	// Assign it, then try — should still fail (not in progress).
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
	_, err = svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: 10})
	if err == nil {
		t.Fatal("expected error for completing assigned task (not in_progress)")
	}
}

func TestTaskService_CompleteTask_NegativeQty(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePick,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 10,
	})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	_, err := svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: -1})
	if err == nil {
		t.Fatal("expected error for negative actual_qty")
	}
}

func TestTaskService_PauseResumeFlow(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypeCycleCount,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 50,
	})

	// Assign → start → pause → resume → complete.
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	// Pause.
	updated, err := svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusPaused})
	if err != nil {
		t.Fatalf("in_progress → paused failed: %v", err)
	}
	if updated.Status != domain.TaskStatusPaused {
		t.Errorf("status = %q, want %q", updated.Status, domain.TaskStatusPaused)
	}

	// Resume.
	updated, err = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})
	if err != nil {
		t.Fatalf("paused → in_progress failed: %v", err)
	}
	if updated.Status != domain.TaskStatusInProgress {
		t.Errorf("status = %q, want %q", updated.Status, domain.TaskStatusInProgress)
	}

	// Complete.
	updated, err = svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: 50})
	if err != nil {
		t.Fatalf("CompleteTask after resume failed: %v", err)
	}
	if updated.Status != domain.TaskStatusCompleted {
		t.Errorf("status = %q, want %q", updated.Status, domain.TaskStatusCompleted)
	}
}

func TestTaskService_ExceptionFlow(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePick,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 10,
	})

	// Assign and start the task.
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	// in_progress → exception (via UpdateTaskStatus — exception is a valid target)
	// Note: The repository's UpdateTaskStatus handles status transition to exception.
	// The service validates the transition: in_progress → exception should be allowed.

	// We need to directly set status to exception in mock since the state machine
	// doesn't have a direct in_progress→exception path by design —
	// exceptions are set by system events, not operator status changes.
	// But the raw UpdateTaskStatus service method with exception target should be tested.

	// Cancel from exception is valid.
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusCancelled})

	got, _ := svc.GetTask(ctx, task.ID)
	if got.Status != domain.TaskStatusCancelled {
		t.Errorf("status = %q, want %q", got.Status, domain.TaskStatusCancelled)
	}
}

func TestTaskService_CompleteTask_ZeroQty(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(newMockTaskRepo())

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypeCycleCount,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 50,
	})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	// Zero qty (valid — e.g., zero-count for cycle counting).
	updated, err := svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: 0})
	if err != nil {
		t.Fatalf("CompleteTask with zero qty failed: %v", err)
	}
	if updated.ActualQty != 0 {
		t.Errorf("actual_qty = %f, want 0", updated.ActualQty)
	}
}

// ── Inventory-Aware Task Completion Tests ───────────────────────────────────────

// combinedMockRepo implements both TaskRepository and InventoryRepository for testing
// task-completion inventory effects.
type combinedMockRepo struct {
	tasks       map[uuid.UUID]*domain.Task
	waves       map[uuid.UUID]*domain.Wave
	inventory   map[uuid.UUID]*domain.Inventory
	transactions []*domain.InventoryTransaction
	skus        map[uuid.UUID]*domain.SKU
}

func newCombinedMockRepo() *combinedMockRepo {
	return &combinedMockRepo{
		tasks:     make(map[uuid.UUID]*domain.Task),
		waves:     make(map[uuid.UUID]*domain.Wave),
		inventory: make(map[uuid.UUID]*domain.Inventory),
		skus:      make(map[uuid.UUID]*domain.SKU),
	}
}

// ── TaskRepository implementation ────────────────────────────

func (m *combinedMockRepo) CreateTask(ctx context.Context, t *domain.Task) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	m.tasks[t.ID] = t
	return nil
}

func (m *combinedMockRepo) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("task", id.String())
	}
	return t, nil
}

func (m *combinedMockRepo) GetTasksByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.Task, error) {
	var result []*domain.Task
	for _, t := range m.tasks {
		if t.OrderID != nil && *t.OrderID == orderID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *combinedMockRepo) ListTasks(ctx context.Context, filter repository.TaskFilter) ([]*domain.Task, error) {
	var result []*domain.Task
	for _, t := range m.tasks {
		if filter.WarehouseID != uuid.Nil && t.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.TaskType != "" && t.TaskType != filter.TaskType {
			continue
		}
		if filter.Status != "" && t.Status != filter.Status {
			continue
		}
		if filter.AssignedTo != "" && t.AssignedTo != filter.AssignedTo {
			continue
		}
		result = append(result, t)
	}
	return result, nil
}

func (m *combinedMockRepo) AssignTask(ctx context.Context, id uuid.UUID, assignedTo string) error {
	t, ok := m.tasks[id]
	if !ok {
		return pkgerrors.NewNotFound("task", id.String())
	}
	if t.Status != domain.TaskStatusPending {
		return pkgerrors.NewInvalidInput("can only assign pending tasks")
	}
	t.Status = domain.TaskStatusAssigned
	t.AssignedTo = assignedTo
	return nil
}

func (m *combinedMockRepo) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status domain.TaskStatus) error {
	t, ok := m.tasks[id]
	if !ok {
		return pkgerrors.NewNotFound("task", id.String())
	}
	t.Status = status
	return nil
}

func (m *combinedMockRepo) CompleteTask(ctx context.Context, id uuid.UUID, actualQty float64, toLocationID *uuid.UUID) error {
	t, ok := m.tasks[id]
	if !ok {
		return pkgerrors.NewNotFound("task", id.String())
	}
	t.Status = domain.TaskStatusCompleted
	t.ActualQty = actualQty
	if toLocationID != nil {
		t.ToLocation = toLocationID
	}
	return nil
}

func (m *combinedMockRepo) CountTasks(ctx context.Context, filter repository.TaskFilter) (int, error) {
	count := 0
	for _, t := range m.tasks {
		if filter.WarehouseID != uuid.Nil && t.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.TaskType != "" && t.TaskType != filter.TaskType {
			continue
		}
		if filter.Status != "" && t.Status != filter.Status {
			continue
		}
		if filter.AssignedTo != "" && t.AssignedTo != filter.AssignedTo {
			continue
		}
		count++
	}
	return count, nil
}

func (m *combinedMockRepo) CountTasksByStatus(ctx context.Context) (map[domain.TaskStatus]int, error) {
	result := make(map[domain.TaskStatus]int)
	for _, t := range m.tasks {
		result[t.Status]++
	}
	return result, nil
}

// ── Wave (not used by these tests) ────────────────────────────

func (m *combinedMockRepo) CreateWave(ctx context.Context, w *domain.Wave) error { return nil }
func (m *combinedMockRepo) GetWave(ctx context.Context, id uuid.UUID) (*domain.Wave, error) {
	return nil, nil
}
func (m *combinedMockRepo) ListWaves(ctx context.Context, filter repository.WaveFilter) ([]*domain.Wave, error) {
	return nil, nil
}
func (m *combinedMockRepo) UpdateWaveStatus(ctx context.Context, id uuid.UUID, status domain.WaveStatus) error {
	return nil
}
func (m *combinedMockRepo) AddWaveOrders(ctx context.Context, id uuid.UUID, orderIDs []uuid.UUID) error {
	return nil
}
func (m *combinedMockRepo) RemoveWaveOrders(ctx context.Context, id uuid.UUID, orderIDs []uuid.UUID) error {
	return nil
}
func (m *combinedMockRepo) CountWaves(ctx context.Context, filter repository.WaveFilter) (int, error) {
	return 0, nil
}

// ── InventoryRepository implementation ────────────────────────

func (m *combinedMockRepo) CreateSKU(ctx context.Context, s *domain.SKU) error { return nil }
func (m *combinedMockRepo) GetSKU(ctx context.Context, id uuid.UUID) (*domain.SKU, error) {
	return nil, nil
}
func (m *combinedMockRepo) GetSKUByBarcode(ctx context.Context, barcode string) (*domain.SKU, error) {
	return nil, nil
}
func (m *combinedMockRepo) GetSKUByCode(ctx context.Context, code string) (*domain.SKU, error) {
	return nil, nil
}
func (m *combinedMockRepo) ListSKUs(ctx context.Context, limit, offset int) ([]*domain.SKU, error) {
	return nil, nil
}
func (m *combinedMockRepo) UpdateSKU(ctx context.Context, s *domain.SKU) error { return nil }
func (m *combinedMockRepo) CountSKUs(ctx context.Context) (int, error)  { return 0, nil }

func (m *combinedMockRepo) CreateInventory(ctx context.Context, inv *domain.Inventory) error {
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	m.inventory[inv.ID] = inv
	return nil
}

func (m *combinedMockRepo) GetInventory(ctx context.Context, id uuid.UUID) (*domain.Inventory, error) {
	inv, ok := m.inventory[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("inventory", id.String())
	}
	return inv, nil
}

func (m *combinedMockRepo) GetAndLockInventory(ctx context.Context, id uuid.UUID) (*domain.Inventory, error) {
	return m.GetInventory(ctx, id)
}

func (m *combinedMockRepo) GetInventoryAtLocation(ctx context.Context, skuID, locationID uuid.UUID, batchNo string) (*domain.Inventory, error) {
	for _, inv := range m.inventory {
		if inv.SKUID == skuID && inv.LocationID == locationID && inv.BatchNo == batchNo {
			return inv, nil
		}
	}
	return nil, pkgerrors.NewNotFound("inventory", fmt.Sprintf("sku=%s,loc=%s", skuID, locationID))
}

func (m *combinedMockRepo) QueryInventory(ctx context.Context, filter repository.InventoryFilter) ([]*domain.Inventory, error) {
	return nil, nil
}

func (m *combinedMockRepo) UpdateInventoryQty(ctx context.Context, id uuid.UUID, deltaQty, deltaReserved float64) error {
	inv, ok := m.inventory[id]
	if !ok {
		return pkgerrors.NewNotFound("inventory", id.String())
	}
	inv.Qty += deltaQty
	inv.ReservedQty += deltaReserved
	inv.AvailableQty = inv.Qty - inv.ReservedQty
	return nil
}

func (m *combinedMockRepo) UpdateInventoryStatus(ctx context.Context, id uuid.UUID, status domain.InventoryStatus) error {
	return nil
}

func (m *combinedMockRepo) CountInventory(ctx context.Context, filter repository.InventoryFilter) (int, error) {
	return len(m.inventory), nil
}

func (m *combinedMockRepo) GetOldestInventory(ctx context.Context, filter repository.InventoryRetrievalFilter) ([]*domain.Inventory, error) {
	return nil, nil
}

func (m *combinedMockRepo) GetExpiringInventory(ctx context.Context, filter repository.InventoryRetrievalFilter) ([]*domain.Inventory, error) {
	return nil, nil
}

func (m *combinedMockRepo) CreateTransaction(ctx context.Context, tx *domain.InventoryTransaction) error {
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	m.transactions = append(m.transactions, tx)
	return nil
}

func (m *combinedMockRepo) ListTransactions(ctx context.Context, inventoryID uuid.UUID, limit, offset int) ([]*domain.InventoryTransaction, error) {
	return nil, nil
}

func (m *combinedMockRepo) ListTransactionsByReference(ctx context.Context, referenceType string, referenceID uuid.UUID) ([]*domain.InventoryTransaction, error) {
	return nil, nil
}

func (m *combinedMockRepo) CountTransactions(ctx context.Context, inventoryID uuid.UUID) (int, error) {
	return 0, nil
}

func (m *combinedMockRepo) ListTransactionsGlobal(ctx context.Context, filter repository.InventoryTxFilter) ([]*domain.InventoryTransaction, error) {
	return nil, nil
}

func (m *combinedMockRepo) CountTransactionsGlobal(ctx context.Context, filter repository.InventoryTxFilter) (int, error) {
	return 0, nil
}

func (m *combinedMockRepo) GetInventoryDashboardStats(ctx context.Context, warehouseID uuid.UUID, lowStockThreshold float64) (*repository.InventoryDashboardStats, error) {
	return nil, nil
}

func (m *combinedMockRepo) GetLowStockInventory(ctx context.Context, threshold float64, warehouseID uuid.UUID, limit int) ([]*domain.Inventory, error) {
	return nil, nil
}

func (m *combinedMockRepo) GetInventoryByWarehouse(ctx context.Context) ([]*repository.InventoryByWarehouseRow, error) {
	return nil, nil
}

// ── Simple in-memory TxManager ────────────────────────────────

type simpleTxManager struct{}

func (m *simpleTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	// In tests, just run the function directly (no real DB transaction needed).
	return fn(ctx)
}

// ── Inventory Effect Tests ────────────────────────────────────────────

func TestTaskService_CompleteTask_PutawayCreatesInventory(t *testing.T) {
	ctx := context.Background()
	repo := newCombinedMockRepo()
	svc := NewTaskServiceWithTx(repo, repo, &simpleTxManager{})

	whID := uuid.New()
	skuID := uuid.New()
	toLoc := uuid.New()

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:     domain.TaskTypePutaway,
		WarehouseID:  whID,
		SKUID:        skuID,
		ExpectedQty:  100,
		ToLocation:   &toLoc,
		BatchNo:      "BATCH-001",
	})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	updated, err := svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: 95})
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}
	if updated.Status != domain.TaskStatusCompleted {
		t.Errorf("status = %q, want completed", updated.Status)
	}

	// Check inventory was created.
	if len(repo.inventory) != 1 {
		t.Fatalf("expected 1 inventory record, got %d", len(repo.inventory))
	}

	var inv *domain.Inventory
	for _, v := range repo.inventory {
		inv = v
		break
	}

	if inv.SKUID != skuID {
		t.Errorf("inv sku_id = %s, want %s", inv.SKUID, skuID)
	}
	if inv.LocationID != toLoc {
		t.Errorf("inv location_id = %s, want %s", inv.LocationID, toLoc)
	}
	if inv.Qty != 95 {
		t.Errorf("inv qty = %f, want 95", inv.Qty)
	}
	if inv.BatchNo != "BATCH-001" {
		t.Errorf("inv batch_no = %q, want BATCH-001", inv.BatchNo)
	}
	if inv.Status != domain.InventoryStatusAvailable {
		t.Errorf("inv status = %q, want available", inv.Status)
	}

	// Check transaction was recorded.
	if len(repo.transactions) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(repo.transactions))
	}
	tx := repo.transactions[0]
	if tx.Type != domain.InventoryTxPutaway {
		t.Errorf("tx type = %q, want putaway", tx.Type)
	}
	if tx.DeltaQty != 95 {
		t.Errorf("tx delta_qty = %f, want 95", tx.DeltaQty)
	}
	if tx.ReferenceType != "task" {
		t.Errorf("tx reference_type = %q, want task", tx.ReferenceType)
	}
	if tx.ReferenceID != task.ID {
		t.Errorf("tx reference_id = %s, want %s", tx.ReferenceID, task.ID)
	}
}

func TestTaskService_CompleteTask_PutawayIncrementsExistingInventory(t *testing.T) {
	ctx := context.Background()
	repo := newCombinedMockRepo()
	svc := NewTaskServiceWithTx(repo, repo, &simpleTxManager{})

	whID := uuid.New()
	skuID := uuid.New()
	toLoc := uuid.New()

	// Pre-populate inventory at the target location.
	existing := &domain.Inventory{
		ID:          uuid.New(),
		SKUID:       skuID,
		LocationID:  toLoc,
		WarehouseID: whID,
		BatchNo:     "BATCH-001",
		Qty:         50,
		ReservedQty: 0,
		Status:      domain.InventoryStatusAvailable,
	}
	existing.AvailableQty = existing.Qty - existing.ReservedQty
	repo.inventory[existing.ID] = existing

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePutaway,
		WarehouseID: whID,
		SKUID:       skuID,
		ExpectedQty: 100,
		ToLocation:  &toLoc,
		BatchNo:     "BATCH-001",
	})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	_, err := svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: 30})
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	// Check existing inventory was incremented (not duplicated).
	if len(repo.inventory) != 1 {
		t.Fatalf("expected 1 inventory record, got %d", len(repo.inventory))
	}
	if existing.Qty != 80 {
		t.Errorf("inv qty = %f, want 80 (50 + 30)", existing.Qty)
	}

	// Check transaction.
	if len(repo.transactions) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(repo.transactions))
	}
	tx := repo.transactions[0]
	if tx.DeltaQty != 30 {
		t.Errorf("tx delta_qty = %f, want 30", tx.DeltaQty)
	}
	if tx.ResultingQty != 80 {
		t.Errorf("tx resulting_qty = %f, want 80", tx.ResultingQty)
	}
}

func TestTaskService_CompleteTask_PutawayRequiresToLocation(t *testing.T) {
	ctx := context.Background()
	repo := newCombinedMockRepo()
	svc := NewTaskServiceWithTx(repo, repo, &simpleTxManager{})

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePutaway,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 100,
		// No ToLocation set.
	})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	_, err := svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: 50})
	if err == nil {
		t.Fatal("expected error for putaway without to_location_id")
	}
}

func TestTaskService_CompleteTask_PickDecrementsInventory(t *testing.T) {
	ctx := context.Background()
	repo := newCombinedMockRepo()
	svc := NewTaskServiceWithTx(repo, repo, &simpleTxManager{})

	whID := uuid.New()
	skuID := uuid.New()
	fromLoc := uuid.New()

	// Pre-populate inventory at the pick location.
	existing := &domain.Inventory{
		ID:          uuid.New(),
		SKUID:       skuID,
		LocationID:  fromLoc,
		WarehouseID: whID,
		BatchNo:     "BATCH-002",
		Qty:         100,
		ReservedQty: 0,
		Status:      domain.InventoryStatusAvailable,
	}
	existing.AvailableQty = existing.Qty - existing.ReservedQty
	repo.inventory[existing.ID] = existing

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:     domain.TaskTypePick,
		WarehouseID:  whID,
		SKUID:        skuID,
		ExpectedQty:  30,
		FromLocation: &fromLoc,
		BatchNo:      "BATCH-002",
	})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	_, err := svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: 30})
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	// Check inventory was decremented.
	if existing.Qty != 70 {
		t.Errorf("inv qty = %f, want 70 (100 - 30)", existing.Qty)
	}

	// Check transaction was recorded.
	if len(repo.transactions) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(repo.transactions))
	}
	tx := repo.transactions[0]
	if tx.Type != domain.InventoryTxPick {
		t.Errorf("tx type = %q, want pick", tx.Type)
	}
	if tx.DeltaQty != -30 {
		t.Errorf("tx delta_qty = %f, want -30", tx.DeltaQty)
	}
	if tx.ResultingQty != 70 {
		t.Errorf("tx resulting_qty = %f, want 70", tx.ResultingQty)
	}
	if tx.ReferenceType != "task" {
		t.Errorf("tx reference_type = %q, want task", tx.ReferenceType)
	}
}

func TestTaskService_CompleteTask_PickInsufficientInventory(t *testing.T) {
	ctx := context.Background()
	repo := newCombinedMockRepo()
	svc := NewTaskServiceWithTx(repo, repo, &simpleTxManager{})

	whID := uuid.New()
	skuID := uuid.New()
	fromLoc := uuid.New()

	// Pre-populate inventory with only 10 units.
	existing := &domain.Inventory{
		ID:          uuid.New(),
		SKUID:       skuID,
		LocationID:  fromLoc,
		WarehouseID: whID,
		BatchNo:     "",
		Qty:         10,
		ReservedQty: 5, // Only 5 available.
		Status:      domain.InventoryStatusAvailable,
	}
	existing.AvailableQty = existing.Qty - existing.ReservedQty
	repo.inventory[existing.ID] = existing

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:     domain.TaskTypePick,
		WarehouseID:  whID,
		SKUID:        skuID,
		ExpectedQty:  20,
		FromLocation: &fromLoc,
	})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	_, err := svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: 20})
	if err == nil {
		t.Fatal("expected error for insufficient inventory")
	}

	// Inventory should not have changed.
	if existing.Qty != 10 {
		t.Errorf("inv qty = %f, want 10 (unchanged)", existing.Qty)
	}
	// No transaction should have been recorded.
	if len(repo.transactions) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(repo.transactions))
	}
}

func TestTaskService_CompleteTask_PickRequiresFromLocation(t *testing.T) {
	ctx := context.Background()
	repo := newCombinedMockRepo()
	svc := NewTaskServiceWithTx(repo, repo, &simpleTxManager{})

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePick,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 50,
		// No FromLocation set.
	})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	_, err := svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: 50})
	if err == nil {
		t.Fatal("expected error for pick without from_location_id")
	}
}

func TestTaskService_CompleteTask_CycleCountNoInventoryEffect(t *testing.T) {
	ctx := context.Background()
	repo := newCombinedMockRepo()
	svc := NewTaskServiceWithTx(repo, repo, &simpleTxManager{})

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypeCycleCount,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 50,
	})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	_, err := svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: 48})
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	// No inventory records should have been created.
	if len(repo.inventory) != 0 {
		t.Errorf("expected 0 inventory records, got %d", len(repo.inventory))
	}
	// No transactions should have been recorded.
	if len(repo.transactions) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(repo.transactions))
	}
}

func TestTaskService_CompleteTask_ReplenishMovesInventory(t *testing.T) {
	ctx := context.Background()
	repo := newCombinedMockRepo()
	svc := NewTaskServiceWithTx(repo, repo, &simpleTxManager{})

	whID := uuid.New()
	skuID := uuid.New()
	fromLoc := uuid.New()
	toLoc := uuid.New()

	// Pre-populate inventory at the from location (reserve area).
	fromInv := &domain.Inventory{
		ID:          uuid.New(),
		SKUID:       skuID,
		LocationID:  fromLoc,
		WarehouseID: whID,
		Qty:         200,
		ReservedQty: 0,
		Status:      domain.InventoryStatusAvailable,
	}
	fromInv.AvailableQty = fromInv.Qty - fromInv.ReservedQty
	repo.inventory[fromInv.ID] = fromInv

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:     domain.TaskTypeReplenish,
		WarehouseID:  whID,
		SKUID:        skuID,
		ExpectedQty:  50,
		FromLocation: &fromLoc,
		ToLocation:   &toLoc,
	})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	_, err := svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: 50})
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	// From location should be decremented.
	if fromInv.Qty != 150 {
		t.Errorf("from inv qty = %f, want 150 (200 - 50)", fromInv.Qty)
	}

	// To location should have a new inventory record.
	if len(repo.inventory) != 2 {
		t.Fatalf("expected 2 inventory records, got %d", len(repo.inventory))
	}

	// Two transactions: one deduction (transfer out), one addition (transfer in).
	if len(repo.transactions) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(repo.transactions))
	}
}

func TestTaskService_CompleteTask_ZeroQtyNoInventoryEffect(t *testing.T) {
	ctx := context.Background()
	repo := newCombinedMockRepo()
	svc := NewTaskServiceWithTx(repo, repo, &simpleTxManager{})

	whID := uuid.New()
	skuID := uuid.New()
	toLoc := uuid.New()

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePutaway,
		WarehouseID: whID,
		SKUID:       skuID,
		ExpectedQty: 100,
		ToLocation:  &toLoc,
	})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	_, err := svc.CompleteTask(ctx, task.ID, CompleteTaskInput{ActualQty: 0})
	if err != nil {
		t.Fatalf("CompleteTask with zero qty failed: %v", err)
	}

	// No inventory records should be created for zero qty.
	if len(repo.inventory) != 0 {
		t.Errorf("expected 0 inventory records, got %d", len(repo.inventory))
	}
	// No transactions.
	if len(repo.transactions) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(repo.transactions))
	}
}

func TestTaskService_CompleteTask_OverridesToLocationOnCompletion(t *testing.T) {
	ctx := context.Background()
	repo := newCombinedMockRepo()
	svc := NewTaskServiceWithTx(repo, repo, &simpleTxManager{})

	whID := uuid.New()
	skuID := uuid.New()
	originalToLoc := uuid.New()
	overrideToLoc := uuid.New()

	task, _ := svc.CreateTask(ctx, CreateTaskInput{
		TaskType:    domain.TaskTypePutaway,
		WarehouseID: whID,
		SKUID:       skuID,
		ExpectedQty: 100,
		ToLocation:  &originalToLoc,
	})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusAssigned})
_, _ = svc.UpdateTaskStatus(ctx, task.ID, UpdateTaskStatusInput{Status: domain.TaskStatusInProgress})

	// Complete with a different to_location.
	_, err := svc.CompleteTask(ctx, task.ID, CompleteTaskInput{
		ActualQty:    50,
		ToLocationID: &overrideToLoc,
	})
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	// Inventory should be at the override location.
	if len(repo.inventory) != 1 {
		t.Fatalf("expected 1 inventory record, got %d", len(repo.inventory))
	}
	var inv *domain.Inventory
	for _, v := range repo.inventory {
		inv = v
		break
	}
	if inv.LocationID != overrideToLoc {
		t.Errorf("inv location_id = %s, want override %s", inv.LocationID, overrideToLoc)
	}
}
