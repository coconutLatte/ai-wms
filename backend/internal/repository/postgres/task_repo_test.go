package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// setupTaskTestDB creates a test database and cleans up task/wave-related test data.
func setupTaskTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	cfg := testConfig()

	ctx := context.Background()
	db, err := NewDB(ctx, cfg)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
		return nil, nil
	}

	// Clean up previous test data
	_, _ = db.Pool.Exec(ctx, "DELETE FROM tasks WHERE task_no LIKE 'TEST-%'")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM waves WHERE wave_no LIKE 'TEST-%'")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM order_lines WHERE order_id IN (SELECT id FROM orders WHERE order_no LIKE 'TEST-%')")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM orders WHERE order_no LIKE 'TEST-%'")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM inventory_transactions WHERE sku_id IN (SELECT id FROM skus WHERE code LIKE 'TEST-%')")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM inventory WHERE sku_id IN (SELECT id FROM skus WHERE code LIKE 'TEST-%')")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM skus WHERE code LIKE 'TEST-%'")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM locations WHERE code LIKE 'TEST-%'")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM zones WHERE code LIKE 'TEST-%'")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM warehouses WHERE code LIKE 'TEST-%'")

	cleanup := func() {
		_, _ = db.Pool.Exec(ctx, "DELETE FROM tasks WHERE task_no LIKE 'TEST-%'")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM waves WHERE wave_no LIKE 'TEST-%'")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM order_lines WHERE order_id IN (SELECT id FROM orders WHERE order_no LIKE 'TEST-%')")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM orders WHERE order_no LIKE 'TEST-%'")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM inventory_transactions WHERE sku_id IN (SELECT id FROM skus WHERE code LIKE 'TEST-%')")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM inventory WHERE sku_id IN (SELECT id FROM skus WHERE code LIKE 'TEST-%')")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM skus WHERE code LIKE 'TEST-%'")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM locations WHERE code LIKE 'TEST-%'")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM zones WHERE code LIKE 'TEST-%'")
		_, _ = db.Pool.Exec(ctx, "DELETE FROM warehouses WHERE code LIKE 'TEST-%'")
		db.Close()
	}

	return db, cleanup
}

// createTaskTestWarehouse creates a warehouse for task/wave tests.
func createTaskTestWarehouse(t *testing.T, ctx context.Context, repo *WarehouseRepo) *domain.Warehouse {
	t.Helper()

	wh := &domain.Warehouse{
		Code: "TEST-WH-TSK-" + uuid.New().String()[:8],
		Name: "Task Test Warehouse",
	}
	if err := repo.CreateWarehouse(ctx, wh); err != nil {
		t.Fatalf("CreateWarehouse failed: %v", err)
	}
	return wh
}

// createTaskTestSKU creates a SKU for task tests.
func createTaskTestSKU(t *testing.T, ctx context.Context, repo *InventoryRepo) *domain.SKU {
	t.Helper()

	sku := &domain.SKU{
		Code: "TEST-SKU-TSK-" + uuid.New().String()[:8],
		Name: "Task Test SKU",
		UOM:  domain.UOM{BaseUnit: "EA", PackQty: 1},
	}
	if err := repo.CreateSKU(ctx, sku); err != nil {
		t.Fatalf("CreateSKU failed: %v", err)
	}
	return sku
}

// createTaskTestLocation creates a location for task tests.
func createTaskTestLocation(t *testing.T, ctx context.Context, whRepo *WarehouseRepo, warehouseID, zoneID uuid.UUID, codePrefix string) *domain.Location {
	t.Helper()

	loc := &domain.Location{
		ZoneID:       zoneID,
		WarehouseID:  warehouseID,
		Code:         codePrefix + "-" + uuid.New().String()[:8],
		Barcode:      "BC-" + codePrefix + "-" + uuid.New().String()[:8],
		LocationType: domain.LocationTypeShelf,
	}
	if err := whRepo.CreateLocation(ctx, loc); err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}
	return loc
}

// createTaskTestZone creates a zone for task tests.
func createTaskTestZone(t *testing.T, ctx context.Context, whRepo *WarehouseRepo, warehouseID uuid.UUID) *domain.Zone {
	t.Helper()

	zone := &domain.Zone{
		WarehouseID: warehouseID,
		Code:        "TEST-ZONE-" + uuid.New().String()[:8],
		Name:        "Task Test Zone",
		ZoneType:    domain.ZoneTypeStorage,
	}
	if err := whRepo.CreateZone(ctx, zone); err != nil {
		t.Fatalf("CreateZone failed: %v", err)
	}
	return zone
}

// ── Task Tests ──────────────────────────────────────────────

func TestTaskRepo_CreateAndGetTask(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)
	sku := createTaskTestSKU(t, ctx, invRepo)

	task := &domain.Task{
		TaskNo:       "TEST-TASK-001",
		TaskType:     domain.TaskTypePick,
		WarehouseID:  wh.ID,
		Priority:     domain.TaskPriorityHigh,
		SKUID:        sku.ID,
		ExpectedQty:  10.0,
		UOM:          "EA",
		BatchNo:      "BATCH-001",
		Instructions: "Pick from zone A, shelf 3",
	}

	err := taskRepo.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if task.ID == uuid.Nil {
		t.Error("expected task ID to be set")
	}
	if task.Status != domain.TaskStatusPending {
		t.Errorf("status = %q, want pending", task.Status)
	}

	got, err := taskRepo.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.TaskNo != task.TaskNo {
		t.Errorf("task_no = %q, want %q", got.TaskNo, task.TaskNo)
	}
	if got.TaskType != domain.TaskTypePick {
		t.Errorf("task_type = %q, want pick", got.TaskType)
	}
	if got.Priority != domain.TaskPriorityHigh {
		t.Errorf("priority = %q, want high", got.Priority)
	}
	if got.SKUID != sku.ID {
		t.Errorf("sku_id = %s, want %s", got.SKUID, sku.ID)
	}
	if got.ExpectedQty != 10.0 {
		t.Errorf("expected_qty = %f, want 10.0", got.ExpectedQty)
	}
	if got.BatchNo != "BATCH-001" {
		t.Errorf("batch_no = %q, want BATCH-001", got.BatchNo)
	}
	if got.Instructions != "Pick from zone A, shelf 3" {
		t.Errorf("instructions = %q, want 'Pick from zone A, shelf 3'", got.Instructions)
	}
	if got.OrderID != nil {
		t.Error("expected order_id to be nil")
	}
	if got.StartedAt != nil {
		t.Error("expected started_at to be nil for pending task")
	}
}

func TestTaskRepo_GetTask_NotFound(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	taskRepo := NewTaskRepo(db)

	_, err := taskRepo.GetTask(ctx, uuid.New())
	if err == nil {
		t.Error("expected error for nonexistent task")
	}
}

func TestTaskRepo_CreateTask_Defaults(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)
	sku := createTaskTestSKU(t, ctx, invRepo)

	// Minimal task with no explicit status, priority, or UOM
	task := &domain.Task{
		TaskNo:      "TEST-TASK-DEF-001",
		TaskType:    domain.TaskTypePutaway,
		WarehouseID: wh.ID,
		SKUID:       sku.ID,
		ExpectedQty: 5.0,
	}

	if err := taskRepo.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if task.Status != domain.TaskStatusPending {
		t.Errorf("status = %q, want pending (default)", task.Status)
	}
	if task.Priority != domain.TaskPriorityNormal {
		t.Errorf("priority = %q, want normal (default)", task.Priority)
	}
	if task.UOM != "EA" {
		t.Errorf("uom = %q, want EA (default)", task.UOM)
	}

	got, err := taskRepo.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.ActualQty != 0.0 {
		t.Errorf("actual_qty = %f, want 0.0 (default)", got.ActualQty)
	}
}

func TestTaskRepo_CreateTask_WithOrderReferences(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	orderRepo := NewOrderRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)
	sku := createTaskTestSKU(t, ctx, invRepo)

	// Create an order and order line for the task to reference
	order := &domain.Order{
		OrderNo:     "TEST-ORD-TSK-001",
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateOrder(ctx, order); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	line := &domain.OrderLine{
		OrderID:    order.ID,
		LineNo:     1,
		SKUID:      sku.ID,
		OrderedQty: 10.0,
	}
	if err := orderRepo.CreateOrderLine(ctx, line); err != nil {
		t.Fatalf("CreateOrderLine failed: %v", err)
	}

	task := &domain.Task{
		TaskNo:      "TEST-TASK-ORDREF-001",
		TaskType:    domain.TaskTypePick,
		WarehouseID: wh.ID,
		OrderID:     &order.ID,
		OrderLineID: &line.ID,
		SKUID:       sku.ID,
		ExpectedQty: 10.0,
	}

	if err := taskRepo.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	got, err := taskRepo.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.OrderID == nil || *got.OrderID != order.ID {
		t.Errorf("order_id = %v, want %s", got.OrderID, order.ID)
	}
	if got.OrderLineID == nil || *got.OrderLineID != line.ID {
		t.Errorf("order_line_id = %v, want %s", got.OrderLineID, line.ID)
	}
}

func TestTaskRepo_CreateTask_WithLocations(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)
	zone := createTaskTestZone(t, ctx, whRepo, wh.ID)
	sku := createTaskTestSKU(t, ctx, invRepo)
	fromLoc := createTaskTestLocation(t, ctx, whRepo, wh.ID, zone.ID, "TEST-LOC-FROM")
	toLoc := createTaskTestLocation(t, ctx, whRepo, wh.ID, zone.ID, "TEST-LOC-TO")

	task := &domain.Task{
		TaskNo:       "TEST-TASK-LOC-001",
		TaskType:     domain.TaskTypeTransfer,
		WarehouseID:  wh.ID,
		FromLocation: &fromLoc.ID,
		ToLocation:   &toLoc.ID,
		SKUID:        sku.ID,
		ExpectedQty:  20.0,
	}

	if err := taskRepo.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	got, err := taskRepo.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.FromLocation == nil || *got.FromLocation != fromLoc.ID {
		t.Errorf("from_location_id = %v, want %s", got.FromLocation, fromLoc.ID)
	}
	if got.ToLocation == nil || *got.ToLocation != toLoc.ID {
		t.Errorf("to_location_id = %v, want %s", got.ToLocation, toLoc.ID)
	}
}

func TestTaskRepo_ListTasks(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)
	sku := createTaskTestSKU(t, ctx, invRepo)

	// Create tasks with different types and priorities
	tasks := []struct {
		no       string
		taskType domain.TaskType
		priority domain.TaskPriority
		status   domain.TaskStatus
	}{
		{"TEST-TASK-LIST-001", domain.TaskTypePick, domain.TaskPriorityHigh, domain.TaskStatusPending},
		{"TEST-TASK-LIST-002", domain.TaskTypePutaway, domain.TaskPriorityNormal, domain.TaskStatusPending},
		{"TEST-TASK-LIST-003", domain.TaskTypePick, domain.TaskPriorityUrgent, domain.TaskStatusAssigned},
	}

	for _, tc := range tasks {
		tk := &domain.Task{
			TaskNo:      tc.no,
			TaskType:    tc.taskType,
			WarehouseID: wh.ID,
			Priority:    tc.priority,
			Status:      tc.status,
			SKUID:       sku.ID,
			ExpectedQty: 10.0,
			AssignedTo:  "worker-1",
		}
		if err := taskRepo.CreateTask(ctx, tk); err != nil {
			t.Fatalf("CreateTask [%s] failed: %v", tc.no, err)
		}
	}

	// List all
	all, err := taskRepo.ListTasks(ctx, repository.TaskFilter{})
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(all) < 3 {
		t.Errorf("expected at least 3 tasks, got %d", len(all))
	}

	// Filter by warehouse
	byWH, err := taskRepo.ListTasks(ctx, repository.TaskFilter{WarehouseID: wh.ID})
	if err != nil {
		t.Fatalf("ListTasks by warehouse failed: %v", err)
	}
	if len(byWH) != 3 {
		t.Errorf("expected 3 tasks for warehouse, got %d", len(byWH))
	}

	// Filter by task type
	byType, err := taskRepo.ListTasks(ctx, repository.TaskFilter{TaskType: domain.TaskTypePick})
	if err != nil {
		t.Fatalf("ListTasks by type failed: %v", err)
	}
	if len(byType) != 2 {
		t.Errorf("expected 2 pick tasks, got %d", len(byType))
	}

	// Filter by status
	byStatus, err := taskRepo.ListTasks(ctx, repository.TaskFilter{Status: domain.TaskStatusAssigned})
	if err != nil {
		t.Fatalf("ListTasks by status failed: %v", err)
	}
	if len(byStatus) != 1 {
		t.Errorf("expected 1 assigned task, got %d", len(byStatus))
	}

	// Filter by assigned_to (empty means "not assigned")
	unassigned, err := taskRepo.ListTasks(ctx, repository.TaskFilter{})
	if err != nil {
		t.Fatalf("ListTasks unassigned failed: %v", err)
	}
	// All three tasks have assigned_to set to "worker-1"
	// Filter for tasks assigned to worker-1
	byWorker, err := taskRepo.ListTasks(ctx, repository.TaskFilter{AssignedTo: "worker-1"})
	if err != nil {
		t.Fatalf("ListTasks by assigned_to failed: %v", err)
	}
	if len(byWorker) != 3 {
		t.Errorf("expected 3 tasks for worker-1, got %d", len(byWorker))
	}

	// Filter with limit
	limited, err := taskRepo.ListTasks(ctx, repository.TaskFilter{Limit: 2})
	if err != nil {
		t.Fatalf("ListTasks with limit failed: %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("expected 2 tasks with limit, got %d", len(limited))
	}

	// Filter by non-matching warehouse
	noMatch, err := taskRepo.ListTasks(ctx, repository.TaskFilter{WarehouseID: uuid.New()})
	if err != nil {
		t.Fatalf("ListTasks by non-matching warehouse failed: %v", err)
	}
	if len(noMatch) != 0 {
		t.Errorf("expected 0 tasks for unknown warehouse, got %d", len(noMatch))
	}

	// Verify priority ordering: urgent > high > normal, then by created_at ASC
	// Unused var fix
	_ = unassigned
}

func TestTaskRepo_AssignTask(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)
	sku := createTaskTestSKU(t, ctx, invRepo)

	task := &domain.Task{
		TaskNo:      "TEST-TASK-ASGN-001",
		TaskType:    domain.TaskTypePick,
		WarehouseID: wh.ID,
		SKUID:       sku.ID,
		ExpectedQty: 5.0,
	}
	if err := taskRepo.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Assign to worker
	if err := taskRepo.AssignTask(ctx, task.ID, "worker-42"); err != nil {
		t.Fatalf("AssignTask failed: %v", err)
	}

	got, err := taskRepo.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.Status != domain.TaskStatusAssigned {
		t.Errorf("status = %q, want assigned", got.Status)
	}
	if got.AssignedTo != "worker-42" {
		t.Errorf("assigned_to = %q, want worker-42", got.AssignedTo)
	}
	if got.StartedAt == nil {
		t.Error("expected started_at to be set on assignment")
	}

	// Cannot re-assign an already assigned task (optimistic guard)
	err = taskRepo.AssignTask(ctx, task.ID, "worker-99")
	if err == nil {
		t.Error("expected error when re-assigning an already assigned task")
	}
}

func TestTaskRepo_AssignTask_NotFound(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	taskRepo := NewTaskRepo(db)

	err := taskRepo.AssignTask(ctx, uuid.New(), "worker-1")
	if err == nil {
		t.Error("expected error for nonexistent task")
	}
}

func TestTaskRepo_UpdateTaskStatus(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)
	sku := createTaskTestSKU(t, ctx, invRepo)

	task := &domain.Task{
		TaskNo:      "TEST-TASK-STAT-001",
		TaskType:    domain.TaskTypePutaway,
		WarehouseID: wh.ID,
		SKUID:       sku.ID,
		ExpectedQty: 15.0,
	}
	if err := taskRepo.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Transition to in_progress — should set started_at
	if err := taskRepo.UpdateTaskStatus(ctx, task.ID, domain.TaskStatusInProgress); err != nil {
		t.Fatalf("UpdateTaskStatus -> in_progress failed: %v", err)
	}
	got, err := taskRepo.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.Status != domain.TaskStatusInProgress {
		t.Errorf("status = %q, want in_progress", got.Status)
	}
	if got.StartedAt == nil {
		t.Error("expected started_at to be set when transitioning to in_progress")
	}

	// Transition to paused
	if err := taskRepo.UpdateTaskStatus(ctx, task.ID, domain.TaskStatusPaused); err != nil {
		t.Fatalf("UpdateTaskStatus -> paused failed: %v", err)
	}
	got, err = taskRepo.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.Status != domain.TaskStatusPaused {
		t.Errorf("status = %q, want paused", got.Status)
	}

	// Transition to cancelled — should set cancelled_at
	if err := taskRepo.UpdateTaskStatus(ctx, task.ID, domain.TaskStatusCancelled); err != nil {
		t.Fatalf("UpdateTaskStatus -> cancelled failed: %v", err)
	}
	got, err = taskRepo.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.Status != domain.TaskStatusCancelled {
		t.Errorf("status = %q, want cancelled", got.Status)
	}
	if got.CancelledAt == nil {
		t.Error("expected cancelled_at to be set when cancelling")
	}
}

func TestTaskRepo_UpdateTaskStatus_Completed(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)
	sku := createTaskTestSKU(t, ctx, invRepo)

	task := &domain.Task{
		TaskNo:      "TEST-TASK-COMP-001",
		TaskType:    domain.TaskTypeReplenish,
		WarehouseID: wh.ID,
		SKUID:       sku.ID,
		ExpectedQty: 100.0,
	}
	if err := taskRepo.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Mark as completed
	if err := taskRepo.UpdateTaskStatus(ctx, task.ID, domain.TaskStatusCompleted); err != nil {
		t.Fatalf("UpdateTaskStatus -> completed failed: %v", err)
	}

	got, err := taskRepo.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.Status != domain.TaskStatusCompleted {
		t.Errorf("status = %q, want completed", got.Status)
	}
	if got.CompletedAt == nil {
		t.Error("expected completed_at to be set when completing")
	}
}

func TestTaskRepo_UpdateTaskStatus_NotFound(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	taskRepo := NewTaskRepo(db)

	err := taskRepo.UpdateTaskStatus(ctx, uuid.New(), domain.TaskStatusCompleted)
	if err == nil {
		t.Error("expected error for nonexistent task")
	}
}

func TestTaskRepo_CompleteTask(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)
	zone := createTaskTestZone(t, ctx, whRepo, wh.ID)
	sku := createTaskTestSKU(t, ctx, invRepo)
	toLoc := createTaskTestLocation(t, ctx, whRepo, wh.ID, zone.ID, "TEST-LOC-COMP")

	task := &domain.Task{
		TaskNo:      "TEST-TASK-CMPL-001",
		TaskType:    domain.TaskTypePutaway,
		WarehouseID: wh.ID,
		SKUID:       sku.ID,
		ExpectedQty: 50.0,
		ActualQty:   0.0,
	}
	if err := taskRepo.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	toLocationID := toLoc.ID
	if err := taskRepo.CompleteTask(ctx, task.ID, 48.0, &toLocationID); err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	got, err := taskRepo.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.Status != domain.TaskStatusCompleted {
		t.Errorf("status = %q, want completed", got.Status)
	}
	if got.ActualQty != 48.0 {
		t.Errorf("actual_qty = %f, want 48.0", got.ActualQty)
	}
	if got.ToLocation == nil || *got.ToLocation != toLoc.ID {
		t.Errorf("to_location_id = %v, want %s", got.ToLocation, toLoc.ID)
	}
	if got.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestTaskRepo_CompleteTask_WithoutLocation(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)
	sku := createTaskTestSKU(t, ctx, invRepo)

	task := &domain.Task{
		TaskNo:      "TEST-TASK-CMPL2-001",
		TaskType:    domain.TaskTypeCycleCount,
		WarehouseID: wh.ID,
		SKUID:       sku.ID,
		ExpectedQty: 100.0,
	}
	if err := taskRepo.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// Complete without providing a target location
	if err := taskRepo.CompleteTask(ctx, task.ID, 102.0, nil); err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	got, err := taskRepo.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if got.ActualQty != 102.0 {
		t.Errorf("actual_qty = %f, want 102.0", got.ActualQty)
	}
}

func TestTaskRepo_CompleteTask_NotFound(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	taskRepo := NewTaskRepo(db)

	err := taskRepo.CompleteTask(ctx, uuid.New(), 10.0, nil)
	if err == nil {
		t.Error("expected error for nonexistent task")
	}
}

func TestTaskRepo_CreateTask_AllTypes(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)
	sku := createTaskTestSKU(t, ctx, invRepo)

	types := []domain.TaskType{
		domain.TaskTypePutaway,
		domain.TaskTypePick,
		domain.TaskTypeReplenish,
		domain.TaskTypeTransfer,
		domain.TaskTypeCycleCount,
		domain.TaskTypeLoad,
		domain.TaskTypeUnload,
	}

	for i, tt := range types {
		task := &domain.Task{
			TaskNo:      "TEST-TASK-TYPES-00" + string(rune('1'+i)),
			TaskType:    tt,
			WarehouseID: wh.ID,
			SKUID:       sku.ID,
			ExpectedQty: 10.0,
		}
		if err := taskRepo.CreateTask(ctx, task); err != nil {
			t.Fatalf("CreateTask [%s] failed: %v", tt, err)
		}

		got, err := taskRepo.GetTask(ctx, task.ID)
		if err != nil {
			t.Fatalf("GetTask [%s] failed: %v", tt, err)
		}
		if got.TaskType != tt {
			t.Errorf("task_type = %q, want %q", got.TaskType, tt)
		}
	}
}

// ── Wave Tests ──────────────────────────────────────────────

func TestTaskRepo_CreateAndGetWave(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)

	orderID1 := uuid.New()
	orderID2 := uuid.New()

	wave := &domain.Wave{
		WaveNo:      "TEST-WAVE-001",
		WarehouseID: wh.ID,
		WaveType:    domain.WaveTypeBatch,
		OrderIDs:    []uuid.UUID{orderID1, orderID2},
		TaskIDs:     []uuid.UUID{},
		TotalOrders: 2,
		TotalLines:  5,
		TotalQty:    250.0,
	}

	err := taskRepo.CreateWave(ctx, wave)
	if err != nil {
		t.Fatalf("CreateWave failed: %v", err)
	}
	if wave.ID == uuid.Nil {
		t.Error("expected wave ID to be set")
	}
	if wave.Status != domain.WaveStatusCreated {
		t.Errorf("status = %q, want created", wave.Status)
	}

	got, err := taskRepo.GetWave(ctx, wave.ID)
	if err != nil {
		t.Fatalf("GetWave failed: %v", err)
	}
	if got.WaveNo != wave.WaveNo {
		t.Errorf("wave_no = %q, want %q", got.WaveNo, wave.WaveNo)
	}
	if got.WaveType != domain.WaveTypeBatch {
		t.Errorf("wave_type = %q, want batch", got.WaveType)
	}
	if got.TotalOrders != 2 {
		t.Errorf("total_orders = %d, want 2", got.TotalOrders)
	}
	if got.TotalLines != 5 {
		t.Errorf("total_lines = %d, want 5", got.TotalLines)
	}
	if got.TotalQty != 250.0 {
		t.Errorf("total_qty = %f, want 250.0", got.TotalQty)
	}
	if len(got.OrderIDs) != 2 {
		t.Errorf("expected 2 order IDs, got %d", len(got.OrderIDs))
	}
	if len(got.TaskIDs) != 0 {
		t.Errorf("expected 0 task IDs, got %d", len(got.TaskIDs))
	}
	if got.ReleasedAt != nil {
		t.Error("expected released_at to be nil for created wave")
	}
	if got.CompletedAt != nil {
		t.Error("expected completed_at to be nil for created wave")
	}
}

func TestTaskRepo_GetWave_NotFound(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	taskRepo := NewTaskRepo(db)

	_, err := taskRepo.GetWave(ctx, uuid.New())
	if err == nil {
		t.Error("expected error for nonexistent wave")
	}
}

func TestTaskRepo_CreateWave_Defaults(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)

	// Minimal wave with no explicit status or arrays
	wave := &domain.Wave{
		WaveNo:      "TEST-WAVE-DEF-001",
		WarehouseID: wh.ID,
		WaveType:    domain.WaveTypeSingleOrder,
	}

	if err := taskRepo.CreateWave(ctx, wave); err != nil {
		t.Fatalf("CreateWave failed: %v", err)
	}
	if wave.Status != domain.WaveStatusCreated {
		t.Errorf("status = %q, want created (default)", wave.Status)
	}
	if wave.OrderIDs == nil {
		t.Error("expected order_ids to default to empty slice, got nil")
	}
	if wave.TaskIDs == nil {
		t.Error("expected task_ids to default to empty slice, got nil")
	}

	got, err := taskRepo.GetWave(ctx, wave.ID)
	if err != nil {
		t.Fatalf("GetWave failed: %v", err)
	}
	if got.TotalOrders != 0 {
		t.Errorf("total_orders = %d, want 0 (default)", got.TotalOrders)
	}
	if got.TotalQty != 0.0 {
		t.Errorf("total_qty = %f, want 0.0 (default)", got.TotalQty)
	}
}

func TestTaskRepo_ListWaves(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)

	// Create multiple waves for the warehouse
	for i := 1; i <= 3; i++ {
		wave := &domain.Wave{
			WaveNo:      "TEST-WAVE-LIST-00" + string(rune('0'+i)),
			WarehouseID: wh.ID,
			WaveType:    domain.WaveTypeBatch,
			TotalOrders: i * 2,
		}
		if err := taskRepo.CreateWave(ctx, wave); err != nil {
			t.Fatalf("CreateWave [%d] failed: %v", i, err)
		}
	}

	waves, err := taskRepo.ListWaves(ctx, repository.WaveFilter{WarehouseID: wh.ID})
	if err != nil {
		t.Fatalf("ListWaves failed: %v", err)
	}
	if len(waves) != 3 {
		t.Errorf("expected 3 waves, got %d", len(waves))
	}

	// List waves for a different warehouse — should return empty
	empty, err := taskRepo.ListWaves(ctx, repository.WaveFilter{WarehouseID: uuid.New()})
	if err != nil {
		t.Fatalf("ListWaves for unknown warehouse failed: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected 0 waves for unknown warehouse, got %d", len(empty))
	}
}

func TestTaskRepo_ListWaves_Empty(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)

	waves, err := taskRepo.ListWaves(ctx, repository.WaveFilter{WarehouseID: wh.ID})
	if err != nil {
		t.Fatalf("ListWaves failed: %v", err)
	}
	if len(waves) != 0 {
		t.Errorf("expected 0 waves for warehouse with no waves, got %d", len(waves))
	}
}

func TestTaskRepo_UpdateWaveStatus(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)

	wave := &domain.Wave{
		WaveNo:      "TEST-WAVE-STAT-001",
		WarehouseID: wh.ID,
		WaveType:    domain.WaveTypeZone,
	}
	if err := taskRepo.CreateWave(ctx, wave); err != nil {
		t.Fatalf("CreateWave failed: %v", err)
	}

	// Transition to released — should set released_at
	if err := taskRepo.UpdateWaveStatus(ctx, wave.ID, domain.WaveStatusReleased); err != nil {
		t.Fatalf("UpdateWaveStatus -> released failed: %v", err)
	}
	got, err := taskRepo.GetWave(ctx, wave.ID)
	if err != nil {
		t.Fatalf("GetWave failed: %v", err)
	}
	if got.Status != domain.WaveStatusReleased {
		t.Errorf("status = %q, want released", got.Status)
	}
	if got.ReleasedAt == nil {
		t.Error("expected released_at to be set when releasing")
	}

	// Transition to in_progress
	if err := taskRepo.UpdateWaveStatus(ctx, wave.ID, domain.WaveStatusInProgress); err != nil {
		t.Fatalf("UpdateWaveStatus -> in_progress failed: %v", err)
	}
	got, err = taskRepo.GetWave(ctx, wave.ID)
	if err != nil {
		t.Fatalf("GetWave failed: %v", err)
	}
	if got.Status != domain.WaveStatusInProgress {
		t.Errorf("status = %q, want in_progress", got.Status)
	}

	// Transition to completed — should set completed_at
	if err := taskRepo.UpdateWaveStatus(ctx, wave.ID, domain.WaveStatusCompleted); err != nil {
		t.Fatalf("UpdateWaveStatus -> completed failed: %v", err)
	}
	got, err = taskRepo.GetWave(ctx, wave.ID)
	if err != nil {
		t.Fatalf("GetWave failed: %v", err)
	}
	if got.Status != domain.WaveStatusCompleted {
		t.Errorf("status = %q, want completed", got.Status)
	}
	if got.CompletedAt == nil {
		t.Error("expected completed_at to be set when completing")
	}
}

func TestTaskRepo_UpdateWaveStatus_NotFound(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	taskRepo := NewTaskRepo(db)

	err := taskRepo.UpdateWaveStatus(ctx, uuid.New(), domain.WaveStatusCompleted)
	if err == nil {
		t.Error("expected error for nonexistent wave")
	}
}

func TestTaskRepo_CreateWave_AllTypes(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)

	types := []domain.WaveType{
		domain.WaveTypeSingleOrder,
		domain.WaveTypeBatch,
		domain.WaveTypeZone,
		domain.WaveTypeCarrier,
	}

	for i, wt := range types {
		wave := &domain.Wave{
			WaveNo:      "TEST-WAVE-TYPE-00" + string(rune('1'+i)),
			WarehouseID: wh.ID,
			WaveType:    wt,
		}
		if err := taskRepo.CreateWave(ctx, wave); err != nil {
			t.Fatalf("CreateWave [%s] failed: %v", wt, err)
		}

		got, err := taskRepo.GetWave(ctx, wave.ID)
		if err != nil {
			t.Fatalf("GetWave [%s] failed: %v", wt, err)
		}
		if got.WaveType != wt {
			t.Errorf("wave_type = %q, want %q", got.WaveType, wt)
		}
	}
}

func TestTaskRepo_CreateWave_WithTaskIDs(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)

	taskID1 := uuid.New()
	taskID2 := uuid.New()
	taskID3 := uuid.New()

	wave := &domain.Wave{
		WaveNo:      "TEST-WAVE-TIDS-001",
		WarehouseID: wh.ID,
		WaveType:    domain.WaveTypeBatch,
		TaskIDs:     []uuid.UUID{taskID1, taskID2, taskID3},
		TotalOrders: 3,
		TotalLines:  8,
		TotalQty:    500.0,
	}

	if err := taskRepo.CreateWave(ctx, wave); err != nil {
		t.Fatalf("CreateWave failed: %v", err)
	}

	got, err := taskRepo.GetWave(ctx, wave.ID)
	if err != nil {
		t.Fatalf("GetWave failed: %v", err)
	}
	if len(got.TaskIDs) != 3 {
		t.Errorf("expected 3 task IDs, got %d", len(got.TaskIDs))
	}
	if got.TaskIDs[0] != taskID1 {
		t.Errorf("task_ids[0] = %s, want %s", got.TaskIDs[0], taskID1)
	}
	if got.TaskIDs[2] != taskID3 {
		t.Errorf("task_ids[2] = %s, want %s", got.TaskIDs[2], taskID3)
	}
}

func TestTaskRepo_CountWaves(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)

	// Initially zero waves
	count, err := taskRepo.CountWaves(ctx, repository.WaveFilter{WarehouseID: wh.ID})
	if err != nil {
		t.Fatalf("CountWaves failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 waves, got %d", count)
	}

	// Create 3 waves
	for i := 1; i <= 3; i++ {
		wave := &domain.Wave{
			WaveNo:      "TEST-WAVE-CNT-00" + string(rune('0'+i)),
			WarehouseID: wh.ID,
			WaveType:    domain.WaveTypeBatch,
		}
		if err := taskRepo.CreateWave(ctx, wave); err != nil {
			t.Fatalf("CreateWave [%d] failed: %v", i, err)
		}
	}

	count, err = taskRepo.CountWaves(ctx, repository.WaveFilter{WarehouseID: wh.ID})
	if err != nil {
		t.Fatalf("CountWaves failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 waves, got %d", count)
	}

	// Different warehouse should return 0
	count, err = taskRepo.CountWaves(ctx, repository.WaveFilter{WarehouseID: uuid.New()})
	if err != nil {
		t.Fatalf("CountWaves for unknown warehouse failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 waves for unknown warehouse, got %d", count)
	}
}

// ── Additional Task Repo Tests ─────────────────────────────

func TestTaskRepo_GetTasksByOrderID(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	orderRepo := NewOrderRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)
	sku := createTaskTestSKU(t, ctx, invRepo)

	order := &domain.Order{
		OrderNo:     "TEST-ORD-TBY-" + uuid.New().String()[:8],
		OrderType:   domain.OrderTypeOutbound,
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateOrder(ctx, order); err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Create 3 tasks for this order
	for i := range 3 {
		task := &domain.Task{
			TaskNo:      "TEST-TASK-TBY-00" + string(rune('1'+i)),
			TaskType:    domain.TaskTypePick,
			WarehouseID: wh.ID,
			OrderID:     &order.ID,
			SKUID:       sku.ID,
			ExpectedQty: 10.0,
		}
		if err := taskRepo.CreateTask(ctx, task); err != nil {
			t.Fatalf("CreateTask failed: %v", err)
		}
	}

	// Create 1 task for a different order
	otherOrder := &domain.Order{
		OrderNo:     "TEST-ORD-TBY-OTHER-" + uuid.New().String()[:8],
		OrderType:   domain.OrderTypeInbound,
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateOrder(ctx, otherOrder); err != nil {
		t.Fatalf("CreateOrder (other) failed: %v", err)
	}
	otherTask := &domain.Task{
		TaskNo:      "TEST-TASK-TBY-OTHER",
		TaskType:    domain.TaskTypePutaway,
		WarehouseID: wh.ID,
		OrderID:     &otherOrder.ID,
		SKUID:       sku.ID,
		ExpectedQty: 20.0,
	}
	if err := taskRepo.CreateTask(ctx, otherTask); err != nil {
		t.Fatalf("CreateTask (other) failed: %v", err)
	}

	// Get tasks for the first order
	tasks, err := taskRepo.GetTasksByOrderID(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetTasksByOrderID failed: %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks for order, got %d", len(tasks))
	}
	for _, task := range tasks {
		if task.OrderID == nil || *task.OrderID != order.ID {
			t.Error("task has wrong order_id")
		}
	}

	// Get tasks for order with no tasks
	emptyOrder := &domain.Order{
		OrderNo:     "TEST-ORD-TBY-EMPTY-" + uuid.New().String()[:8],
		OrderType:   domain.OrderTypeTransfer,
		WarehouseID: wh.ID,
	}
	if err := orderRepo.CreateOrder(ctx, emptyOrder); err != nil {
		t.Fatalf("CreateOrder (empty) failed: %v", err)
	}
	tasks, err = taskRepo.GetTasksByOrderID(ctx, emptyOrder.ID)
	if err != nil {
		t.Fatalf("GetTasksByOrderID for empty order failed: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks for empty order, got %d", len(tasks))
	}
}

func TestTaskRepo_CountTasks(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)
	sku := createTaskTestSKU(t, ctx, invRepo)

	// Create 2 pick tasks and 1 putaway task
	for i := range 3 {
		taskType := domain.TaskTypePick
		if i == 2 {
			taskType = domain.TaskTypePutaway
		}
		task := &domain.Task{
			TaskNo:      "TEST-TASK-CNT-00" + string(rune('1'+i)),
			TaskType:    taskType,
			WarehouseID: wh.ID,
			SKUID:       sku.ID,
			ExpectedQty: 10.0,
		}
		if err := taskRepo.CreateTask(ctx, task); err != nil {
			t.Fatalf("CreateTask failed: %v", err)
		}
	}

	// Count all
	count, err := taskRepo.CountTasks(ctx, repository.TaskFilter{})
	if err != nil {
		t.Fatalf("CountTasks failed: %v", err)
	}
	if count < 3 {
		t.Errorf("expected at least 3 tasks, got %d", count)
	}

	// Count by warehouse
	count, err = taskRepo.CountTasks(ctx, repository.TaskFilter{WarehouseID: wh.ID})
	if err != nil {
		t.Fatalf("CountTasks by warehouse failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 tasks for warehouse, got %d", count)
	}

	// Count by type
	count, err = taskRepo.CountTasks(ctx, repository.TaskFilter{TaskType: domain.TaskTypePick})
	if err != nil {
		t.Fatalf("CountTasks by type failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 pick tasks, got %d", count)
	}

	// Zero for unknown warehouse
	count, err = taskRepo.CountTasks(ctx, repository.TaskFilter{WarehouseID: uuid.New()})
	if err != nil {
		t.Fatalf("CountTasks for unknown warehouse failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tasks for unknown warehouse, got %d", count)
	}
}

func TestTaskRepo_CountTasksByStatus(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	invRepo := NewInventoryRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)
	sku := createTaskTestSKU(t, ctx, invRepo)

	// Create tasks with different statuses
	pending := &domain.Task{
		TaskNo: "TEST-TASK-CNTSTAT-001", TaskType: domain.TaskTypePick,
		WarehouseID: wh.ID, SKUID: sku.ID, ExpectedQty: 10.0,
	}
	assigned := &domain.Task{
		TaskNo: "TEST-TASK-CNTSTAT-002", TaskType: domain.TaskTypePutaway,
		WarehouseID: wh.ID, SKUID: sku.ID, ExpectedQty: 10.0,
		Status: domain.TaskStatusAssigned, AssignedTo: "worker-1",
	}
	if err := taskRepo.CreateTask(ctx, pending); err != nil {
		t.Fatalf("CreateTask pending failed: %v", err)
	}
	if err := taskRepo.CreateTask(ctx, assigned); err != nil {
		t.Fatalf("CreateTask assigned failed: %v", err)
	}

	counts, err := taskRepo.CountTasksByStatus(ctx)
	if err != nil {
		t.Fatalf("CountTasksByStatus failed: %v", err)
	}
	if counts[domain.TaskStatusPending] < 1 {
		t.Errorf("expected at least 1 pending task, got %d", counts[domain.TaskStatusPending])
	}
	if counts[domain.TaskStatusAssigned] < 1 {
		t.Errorf("expected at least 1 assigned task, got %d", counts[domain.TaskStatusAssigned])
	}
}

func TestTaskRepo_AddWaveOrders(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)

	orderID1 := uuid.New()
	orderID2 := uuid.New()
	orderID3 := uuid.New()

	wave := &domain.Wave{
		WaveNo:      "TEST-WAVE-ADD-" + uuid.New().String()[:8],
		WarehouseID: wh.ID,
		WaveType:    domain.WaveTypeBatch,
		OrderIDs:    []uuid.UUID{orderID1},
		TotalOrders: 1,
	}
	if err := taskRepo.CreateWave(ctx, wave); err != nil {
		t.Fatalf("CreateWave failed: %v", err)
	}

	// Add 2 more orders
	if err := taskRepo.AddWaveOrders(ctx, wave.ID, []uuid.UUID{orderID2, orderID3}); err != nil {
		t.Fatalf("AddWaveOrders failed: %v", err)
	}

	got, err := taskRepo.GetWave(ctx, wave.ID)
	if err != nil {
		t.Fatalf("GetWave failed: %v", err)
	}
	if got.TotalOrders != 3 {
		t.Errorf("total_orders = %d, want 3", got.TotalOrders)
	}
	if len(got.OrderIDs) != 3 {
		t.Errorf("expected 3 order IDs, got %d", len(got.OrderIDs))
	}

	// Add duplicate — should deduplicate
	if err := taskRepo.AddWaveOrders(ctx, wave.ID, []uuid.UUID{orderID1}); err != nil {
		t.Fatalf("AddWaveOrders (duplicate) failed: %v", err)
	}

	got, err = taskRepo.GetWave(ctx, wave.ID)
	if err != nil {
		t.Fatalf("GetWave failed: %v", err)
	}
	if got.TotalOrders != 3 {
		t.Errorf("total_orders = %d, want 3 (no duplicates)", got.TotalOrders)
	}

	// Not found
	err = taskRepo.AddWaveOrders(ctx, uuid.New(), []uuid.UUID{orderID1})
	if err == nil {
		t.Error("expected error for nonexistent wave")
	}
}

func TestTaskRepo_RemoveWaveOrders(t *testing.T) {
	db, cleanup := setupTaskTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	whRepo := NewWarehouseRepo(db)
	taskRepo := NewTaskRepo(db)

	wh := createTaskTestWarehouse(t, ctx, whRepo)

	orderID1 := uuid.New()
	orderID2 := uuid.New()
	orderID3 := uuid.New()

	wave := &domain.Wave{
		WaveNo:      "TEST-WAVE-REM-" + uuid.New().String()[:8],
		WarehouseID: wh.ID,
		WaveType:    domain.WaveTypeZone,
		OrderIDs:    []uuid.UUID{orderID1, orderID2, orderID3},
		TotalOrders: 3,
	}
	if err := taskRepo.CreateWave(ctx, wave); err != nil {
		t.Fatalf("CreateWave failed: %v", err)
	}

	// Remove 1 order
	if err := taskRepo.RemoveWaveOrders(ctx, wave.ID, []uuid.UUID{orderID2}); err != nil {
		t.Fatalf("RemoveWaveOrders failed: %v", err)
	}

	got, err := taskRepo.GetWave(ctx, wave.ID)
	if err != nil {
		t.Fatalf("GetWave failed: %v", err)
	}
	if got.TotalOrders != 2 {
		t.Errorf("total_orders = %d, want 2", got.TotalOrders)
	}
	if len(got.OrderIDs) != 2 {
		t.Errorf("expected 2 order IDs, got %d", len(got.OrderIDs))
	}
	// Verify orderID2 is removed
	for _, oid := range got.OrderIDs {
		if oid == orderID2 {
			t.Error("orderID2 should have been removed")
		}
	}

	// Remove another 2
	if err := taskRepo.RemoveWaveOrders(ctx, wave.ID, []uuid.UUID{orderID1, orderID3}); err != nil {
		t.Fatalf("RemoveWaveOrders (remaining) failed: %v", err)
	}
	got, err = taskRepo.GetWave(ctx, wave.ID)
	if err != nil {
		t.Fatalf("GetWave failed: %v", err)
	}
	if got.TotalOrders != 0 {
		t.Errorf("total_orders = %d, want 0", got.TotalOrders)
	}
}
