// Package service implements business logic orchestration for the WMS domain.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// TaskService orchestrates business logic for warehouse tasks.
type TaskService struct {
	repo          repository.TaskRepository
	inventoryRepo repository.InventoryRepository
	txManager     repository.TxManager
}

// NewTaskService creates a new TaskService without inventory effects.
// Use NewTaskServiceWithTx when inventory effects and transaction support are needed.
func NewTaskService(repo repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

// NewTaskServiceWithTx creates a new TaskService with inventory support.
// When inventoryRepo and txManager are provided, completing tasks triggers
// inventory effects (putaway creates/increments inventory, pick decrements it)
// within an atomic database transaction.
func NewTaskServiceWithTx(repo repository.TaskRepository, inventoryRepo repository.InventoryRepository, txManager repository.TxManager) *TaskService {
	return &TaskService{repo: repo, inventoryRepo: inventoryRepo, txManager: txManager}
}

// ── Input Types ──────────────────────────────────────────────────────────────────────────

// CreateTaskInput is the input for creating a new task.
type CreateTaskInput struct {
	TaskType     domain.TaskType     `json:"task_type"`
	WarehouseID  uuid.UUID           `json:"warehouse_id"`
	OrderID      *uuid.UUID          `json:"order_id,omitempty"`
	OrderLineID  *uuid.UUID          `json:"order_line_id,omitempty"`
	Priority     domain.TaskPriority `json:"priority,omitempty"` // Default "normal"
	FromLocation *uuid.UUID          `json:"from_location_id,omitempty"`
	ToLocation   *uuid.UUID          `json:"to_location_id,omitempty"`
	SKUID        uuid.UUID           `json:"sku_id"`
	ExpectedQty  float64             `json:"expected_qty"`
	UOM          string              `json:"uom,omitempty"` // Default "EA"
	BatchNo      string              `json:"batch_no,omitempty"`
	Instructions string              `json:"instructions,omitempty"`
}

// Validate checks the input for business rule violations.
func (in *CreateTaskInput) Validate() error {
	if !isValidTaskType(in.TaskType) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid task_type: %s", in.TaskType))
	}
	if in.WarehouseID == uuid.Nil {
		return pkgerrors.NewInvalidInput("warehouse_id is required")
	}
	if in.Priority != "" && !isValidTaskPriority(in.Priority) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid priority: %s", in.Priority))
	}
	if in.SKUID == uuid.Nil {
		return pkgerrors.NewInvalidInput("sku_id is required")
	}
	if in.ExpectedQty <= 0 {
		return pkgerrors.NewInvalidInput("expected_qty must be positive")
	}
	return nil
}

// AssignTaskInput is the input for assigning a task to a worker.
type AssignTaskInput struct {
	AssignedTo string `json:"assigned_to"`
}

// Validate checks the input for business rule violations.
func (in *AssignTaskInput) Validate() error {
	if in.AssignedTo == "" {
		return pkgerrors.NewInvalidInput("assigned_to is required")
	}
	return nil
}

// UpdateTaskStatusInput is the input for updating a task's status.
type UpdateTaskStatusInput struct {
	Status domain.TaskStatus `json:"status"`
}

// Validate checks the input for business rule violations.
func (in *UpdateTaskStatusInput) Validate() error {
	if !isValidTaskStatus(in.Status) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid task status: %s", in.Status))
	}
	return nil
}

// CompleteTaskInput is the input for completing a task.
type CompleteTaskInput struct {
	ActualQty    float64    `json:"actual_qty"`
	ToLocationID *uuid.UUID `json:"to_location_id,omitempty"`
}

// Validate checks the input for business rule violations.
func (in *CompleteTaskInput) Validate() error {
	if in.ActualQty < 0 {
		return pkgerrors.NewInvalidInput("actual_qty must be >= 0")
	}
	return nil
}

// ── Service Methods ──────────────────────────────────────────────────────────────────────

// CreateTask validates input and creates a new task.
func (s *TaskService) CreateTask(ctx context.Context, input CreateTaskInput) (*domain.Task, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	priority := input.Priority
	if priority == "" {
		priority = domain.TaskPriorityNormal
	}
	uom := input.UOM
	if uom == "" {
		uom = "EA"
	}

	task := &domain.Task{
		TaskNo:       generateTaskNo(),
		TaskType:     input.TaskType,
		WarehouseID:  input.WarehouseID,
		OrderID:      input.OrderID,
		OrderLineID:  input.OrderLineID,
		Priority:     priority,
		Status:       domain.TaskStatusPending,
		FromLocation: input.FromLocation,
		ToLocation:   input.ToLocation,
		SKUID:        input.SKUID,
		ExpectedQty:  input.ExpectedQty,
		UOM:          uom,
		BatchNo:      input.BatchNo,
		Instructions: input.Instructions,
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("task service: create: %w", err)
	}

	return task, nil
}

// GetTask retrieves a task by ID.
func (s *TaskService) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("task service: get %s: %w", id, err)
	}
	return task, nil
}

// ListTasks returns tasks matching the specified filter, ordered by priority desc then age asc.
func (s *TaskService) ListTasks(ctx context.Context, filter repository.TaskFilter) ([]*domain.Task, int, error) {
	tasks, err := s.repo.ListTasks(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("task service: list: %w", err)
	}

	total, err := s.repo.CountTasks(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("task service: count: %w", err)
	}

	return tasks, total, nil
}

// AssignTask assigns a pending task to a worker.
// Only tasks in "pending" status can be assigned.
func (s *TaskService) AssignTask(ctx context.Context, id uuid.UUID, input AssignTaskInput) (*domain.Task, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("task service: assign %s: %w", id, err)
	}

	if task.Status != domain.TaskStatusPending {
		return nil, pkgerrors.NewInvalidInput(
			fmt.Sprintf("can only assign pending tasks (current: %s)", task.Status))
	}

	if err := s.repo.AssignTask(ctx, id, input.AssignedTo); err != nil {
		return nil, fmt.Errorf("task service: assign %s: %w", id, err)
	}

	updated, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("task service: re-fetch after assign %s: %w", id, err)
	}
	return updated, nil
}

// UpdateTaskStatus validates the state transition and updates the task status.
// Used for status transitions like: start, pause, resume, cancel.
func (s *TaskService) UpdateTaskStatus(ctx context.Context, id uuid.UUID, input UpdateTaskStatusInput) (*domain.Task, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("task service: update status %s: %w", id, err)
	}

	// Validate the state transition.
	if !task.CanTransitionTo(input.Status) {
		return nil, pkgerrors.NewInvalidStatus(string(task.Status), string(input.Status))
	}

	if err := s.repo.UpdateTaskStatus(ctx, id, input.Status); err != nil {
		return nil, fmt.Errorf("task service: update status %s: %w", id, err)
	}

	updated, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("task service: re-fetch after status update %s: %w", id, err)
	}
	return updated, nil
}

// CompleteTask marks a task as completed with the actual quantity performed.
// Only tasks in "in_progress" status can be completed.
//
// When inventory support is configured (via NewTaskServiceWithTx), completing a task
// also applies inventory effects within an atomic transaction:
//   - Putaway / Transfer / Return: creates or increments inventory at the to_location.
//     Records an InventoryTransaction of type "putaway".
//   - Pick: decrements inventory at the from_location.
//     Records an InventoryTransaction of type "pick".
//     Fails if the deduction would result in negative quantity.
func (s *TaskService) CompleteTask(ctx context.Context, id uuid.UUID, input CompleteTaskInput) (*domain.Task, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("task service: complete %s: %w", id, err)
	}

	if task.Status != domain.TaskStatusInProgress {
		return nil, pkgerrors.NewInvalidInput(
			fmt.Sprintf("can only complete in-progress tasks (current: %s)", task.Status))
	}

	// If no inventory support, just complete the task directly.
	if s.inventoryRepo == nil || s.txManager == nil {
		if err := s.repo.CompleteTask(ctx, id, input.ActualQty, input.ToLocationID); err != nil {
			return nil, fmt.Errorf("task service: complete %s: %w", id, err)
		}

		updated, err := s.repo.GetTask(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("task service: re-fetch after complete %s: %w", id, err)
		}
		return updated, nil
	}

	// With inventory support: perform task completion + inventory effects atomically.
	doWrites := func(ctx context.Context) error {
		// 1. Complete the task.
		if err := s.repo.CompleteTask(ctx, id, input.ActualQty, input.ToLocationID); err != nil {
			return fmt.Errorf("complete task: %w", err)
		}

		// 2. Apply inventory effects based on task type.
		if err := s.applyInventoryEffect(ctx, task, input.ActualQty, input.ToLocationID); err != nil {
			return err
		}

		return nil
	}

	if err := s.txManager.WithTx(ctx, doWrites); err != nil {
		return nil, fmt.Errorf("task service: complete: %w", err)
	}

	updated, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("task service: re-fetch after complete %s: %w", id, err)
	}
	return updated, nil
}

// applyInventoryEffect determines and applies the correct inventory change for the task type.
func (s *TaskService) applyInventoryEffect(ctx context.Context, task *domain.Task, actualQty float64, toLocationID *uuid.UUID) error {
	switch task.TaskType {
	case domain.TaskTypePutaway, domain.TaskTypeTransfer:
		return s.applyPutawayEffect(ctx, task, actualQty, toLocationID)
	case domain.TaskTypePick:
		return s.applyPickEffect(ctx, task, actualQty)
	case domain.TaskTypeReplenish:
		// Replenish: moves inventory from reserve to pick location.
		// Decrement from from_location, increment at to_location.
		if task.FromLocation != nil {
			if err := s.deductInventory(ctx, task, *task.FromLocation, actualQty, domain.InventoryTxTransfer); err != nil {
				return err
			}
		}
		targetLoc := task.ToLocation
		if toLocationID != nil {
			targetLoc = toLocationID
		}
		if targetLoc != nil {
			return s.addInventory(ctx, task, *targetLoc, actualQty, domain.InventoryTxTransfer)
		}
		return nil
	default:
		// cycle_count, load, unload: no automatic inventory effect.
		return nil
	}
}

// applyPutawayEffect creates or increments inventory at the target location.
func (s *TaskService) applyPutawayEffect(ctx context.Context, task *domain.Task, actualQty float64, toLocationID *uuid.UUID) error {
	if actualQty <= 0 {
		return nil // Nothing to add.
	}
	targetLoc := task.ToLocation
	if toLocationID != nil {
		targetLoc = toLocationID
	}
	if targetLoc == nil {
		return pkgerrors.NewInvalidInput("to_location_id is required for putaway tasks")
	}
	return s.addInventory(ctx, task, *targetLoc, actualQty, domain.InventoryTxPutaway)
}

// applyPickEffect decrements inventory at the source (from) location.
func (s *TaskService) applyPickEffect(ctx context.Context, task *domain.Task, actualQty float64) error {
	if actualQty <= 0 {
		return nil // Nothing to deduct.
	}
	fromLoc := task.FromLocation
	if fromLoc == nil {
		return pkgerrors.NewInvalidInput("from_location_id is required for pick tasks")
	}
	return s.deductInventory(ctx, task, *fromLoc, actualQty, domain.InventoryTxPick)
}

// addInventory finds existing inventory at the given location for the task's SKU+batch,
// increments its quantity, or creates a new inventory record if none exists.
func (s *TaskService) addInventory(ctx context.Context, task *domain.Task, locationID uuid.UUID, qty float64, txType domain.InventoryTxType) error {
	// Try to find existing inventory at this location for this SKU and batch.
	inv, err := s.inventoryRepo.GetInventoryAtLocation(ctx, task.SKUID, locationID, task.BatchNo)
	if err != nil {
		// Not found — create new inventory record.
		newInv := &domain.Inventory{
			SKUID:       task.SKUID,
			LocationID:  locationID,
			WarehouseID: task.WarehouseID,
			BatchNo:     task.BatchNo,
			Qty:         qty,
			ReservedQty: 0,
			Status:      domain.InventoryStatusAvailable,
		}
		if err := s.inventoryRepo.CreateInventory(ctx, newInv); err != nil {
			return fmt.Errorf("create inventory: %w", err)
		}

		// Record transaction.
		tx := &domain.InventoryTransaction{
			InventoryID:   newInv.ID,
			SKUID:         task.SKUID,
			LocationID:    locationID,
			Type:          txType,
			DeltaQty:      qty,
			ResultingQty:  qty,
			ReferenceType: "task",
			ReferenceID:   task.ID,
		}
		if err := s.inventoryRepo.CreateTransaction(ctx, tx); err != nil {
			return fmt.Errorf("create inventory transaction: %w", err)
		}
		return nil
	}

	// Existing inventory found — increment quantity.
	newQty := inv.Qty + qty
	if err := s.inventoryRepo.UpdateInventoryQty(ctx, inv.ID, qty, 0); err != nil {
		return fmt.Errorf("update inventory qty: %w", err)
	}

	// Record transaction.
	tx := &domain.InventoryTransaction{
		InventoryID:   inv.ID,
		SKUID:         task.SKUID,
		LocationID:    locationID,
		Type:          txType,
		DeltaQty:      qty,
		ResultingQty:  newQty,
		ReferenceType: "task",
		ReferenceID:   task.ID,
	}
	if err := s.inventoryRepo.CreateTransaction(ctx, tx); err != nil {
		return fmt.Errorf("create inventory transaction: %w", err)
	}

	return nil
}

// deductInventory decrements inventory at the given location for the task's SKU+batch.
// Fails if the deduction would result in negative quantity.
func (s *TaskService) deductInventory(ctx context.Context, task *domain.Task, locationID uuid.UUID, qty float64, txType domain.InventoryTxType) error {
	// Find existing inventory at this location for this SKU and batch.
	inv, err := s.inventoryRepo.GetInventoryAtLocation(ctx, task.SKUID, locationID, task.BatchNo)
	if err != nil {
		return fmt.Errorf("inventory not found at location for SKU %s: %w", task.SKUID, err)
	}

	// Check negative qty constraint against domain model.
	if !inv.CanDeduct(qty) {
		return pkgerrors.NewInvalidInput(
			fmt.Sprintf("insufficient inventory: location has %.2f available (qty=%.2f, reserved=%.2f), need to deduct %.2f",
				inv.Available(), inv.Qty, inv.ReservedQty, qty),
		)
	}

	newQty := inv.ResultingQty(-qty)
	if err := s.inventoryRepo.UpdateInventoryQty(ctx, inv.ID, -qty, 0); err != nil {
		return fmt.Errorf("update inventory qty: %w", err)
	}

	// Record transaction.
	tx := &domain.InventoryTransaction{
		InventoryID:   inv.ID,
		SKUID:         task.SKUID,
		LocationID:    locationID,
		Type:          txType,
		DeltaQty:      -qty,
		ResultingQty:  newQty,
		ReferenceType: "task",
		ReferenceID:   task.ID,
	}
	if err := s.inventoryRepo.CreateTransaction(ctx, tx); err != nil {
		return fmt.Errorf("create inventory transaction: %w", err)
	}

	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────────────────

// generateTaskNo creates a business task number: TASK-YYYYMMDD-NNNNNN.
func generateTaskNo() string {
	now := time.Now()
	return fmt.Sprintf("TASK-%s-%06d", now.Format("20060102"), now.UnixMilli()%1000000)
}

func isValidTaskType(t domain.TaskType) bool {
	switch t {
	case domain.TaskTypePutaway, domain.TaskTypePick, domain.TaskTypeReplenish,
		domain.TaskTypeTransfer, domain.TaskTypeCycleCount, domain.TaskTypeLoad, domain.TaskTypeUnload:
		return true
	}
	return false
}

func isValidTaskPriority(p domain.TaskPriority) bool {
	switch p {
	case domain.TaskPriorityLow, domain.TaskPriorityNormal,
		domain.TaskPriorityHigh, domain.TaskPriorityUrgent:
		return true
	}
	return false
}

func isValidTaskStatus(s domain.TaskStatus) bool {
	switch s {
	case domain.TaskStatusPending, domain.TaskStatusAssigned, domain.TaskStatusInProgress,
		domain.TaskStatusPaused, domain.TaskStatusCompleted, domain.TaskStatusCancelled,
		domain.TaskStatusException:
		return true
	}
	return false
}
