package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// TaskRepo implements repository.TaskRepository using PostgreSQL.
type TaskRepo struct {
	db *DB
}

// NewTaskRepo creates a new TaskRepo.
func NewTaskRepo(db *DB) *TaskRepo {
	return &TaskRepo{db: db}
}

// ── Task ────────────────────────────────────────────────────

// CreateTask inserts a new task.
func (r *TaskRepo) CreateTask(ctx context.Context, t *domain.Task) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	t.CreatedAt = time.Now()
	if t.Status == "" {
		t.Status = domain.TaskStatusPending
	}
	if t.Priority == "" {
		t.Priority = domain.TaskPriorityNormal
	}
	if t.UOM == "" {
		t.UOM = "EA"
	}

	const query = `
		INSERT INTO tasks (id, task_no, task_type, warehouse_id, order_id, order_line_id,
		                   priority, status, assigned_to, from_location_id, to_location_id,
		                   sku_id, expected_qty, actual_qty, uom, batch_no, instructions, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	_, err := r.exec(ctx, query,
		t.ID, t.TaskNo, t.TaskType, t.WarehouseID,
		t.OrderID, t.OrderLineID,
		t.Priority, t.Status, nullString(t.AssignedTo),
		t.FromLocation, t.ToLocation,
		t.SKUID, t.ExpectedQty, t.ActualQty,
		t.UOM, nullString(t.BatchNo), nullString(t.Instructions),
		t.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}
	return nil
}

// GetTask retrieves a task by ID.
func (r *TaskRepo) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	const query = `
		SELECT id, task_no, task_type, warehouse_id, order_id, order_line_id,
		       priority, status, assigned_to, from_location_id, to_location_id,
		       sku_id, expected_qty, actual_qty, uom, batch_no, instructions,
		       created_at, started_at, completed_at, cancelled_at
		FROM tasks WHERE id = $1`

	t, err := r.scanTask(r.queryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get task %s: %w", id, err)
		}
		return nil, fmt.Errorf("get task: %w", err)
	}
	return t, nil
}

// GetTasksByOrderID retrieves all tasks associated with a given order.
func (r *TaskRepo) GetTasksByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.Task, error) {
	const query = `
		SELECT id, task_no, task_type, warehouse_id, order_id, order_line_id,
		       priority, status, assigned_to, from_location_id, to_location_id,
		       sku_id, expected_qty, actual_qty, uom, batch_no, instructions,
		       created_at, started_at, completed_at, cancelled_at
		FROM tasks WHERE order_id = $1
		ORDER BY created_at ASC`

	rows, err := r.query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("get tasks by order: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		t, err := r.scanTaskFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}
	return tasks, nil
}

// ListTasks returns tasks matching the specified filter.
func (r *TaskRepo) ListTasks(ctx context.Context, filter repository.TaskFilter) ([]*domain.Task, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.WarehouseID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("warehouse_id = $%d", argIdx))
		args = append(args, filter.WarehouseID)
		argIdx++
	}
	if filter.TaskType != "" {
		conditions = append(conditions, fmt.Sprintf("task_type = $%d", argIdx))
		args = append(args, filter.TaskType)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.AssignedTo != "" {
		conditions = append(conditions, fmt.Sprintf("assigned_to = $%d", argIdx))
		args = append(args, filter.AssignedTo)
		argIdx++
	}

	query := `
		SELECT id, task_no, task_type, warehouse_id, order_id, order_line_id,
		       priority, status, assigned_to, from_location_id, to_location_id,
		       sku_id, expected_qty, actual_qty, uom, batch_no, instructions,
		       created_at, started_at, completed_at, cancelled_at
		FROM tasks`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY priority DESC, created_at ASC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
		argIdx++
	}

	rows, err := r.query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		t, err := r.scanTaskFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}
	return tasks, nil
}

// AssignTask assigns a task to a worker or robot.
func (r *TaskRepo) AssignTask(ctx context.Context, id uuid.UUID, assignedTo string) error {
	const query = `
		UPDATE tasks SET assigned_to=$1, status=$2, started_at=$3
		WHERE id=$4 AND status = 'pending'`

	now := time.Now()
	tag, err := r.exec(ctx, query, assignedTo, domain.TaskStatusAssigned, now, id)
	if err != nil {
		return fmt.Errorf("assign task: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("assign task %s: not found or not in assignable status", id)
	}
	return nil
}

// UpdateTaskStatus transitions a task to a new status.
func (r *TaskRepo) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status domain.TaskStatus) error {
	now := time.Now()
	var startedAt, completedAt, cancelledAt *time.Time

	switch status {
	case domain.TaskStatusInProgress:
		startedAt = &now
	case domain.TaskStatusCompleted:
		completedAt = &now
	case domain.TaskStatusCancelled:
		cancelledAt = &now
	}

	const query = `
		UPDATE tasks SET status=$1, started_at=COALESCE($2, started_at),
		                 completed_at=$3, cancelled_at=$4
		WHERE id=$5`

	tag, err := r.exec(ctx, query, status, startedAt, completedAt, cancelledAt, id)
	if err != nil {
		return fmt.Errorf("update task status: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("update task status %s: not found", id)
	}
	return nil
}

// CompleteTask marks a task as completed with actual quantity and optional target location.
func (r *TaskRepo) CompleteTask(ctx context.Context, id uuid.UUID, actualQty float64, toLocationID *uuid.UUID) error {
	now := time.Now()

	const query = `
		UPDATE tasks SET status=$1, actual_qty=$2, to_location_id=COALESCE($3, to_location_id),
		                 completed_at=$4
		WHERE id=$5`

	tag, err := r.exec(ctx, query,
		domain.TaskStatusCompleted, actualQty, toLocationID, now, id,
	)
	if err != nil {
		return fmt.Errorf("complete task: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("complete task %s: not found", id)
	}
	return nil
}

// CountTasks returns the total count of tasks matching the filter.
func (r *TaskRepo) CountTasks(ctx context.Context, filter repository.TaskFilter) (int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.WarehouseID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("warehouse_id = $%d", argIdx))
		args = append(args, filter.WarehouseID)
		argIdx++
	}
	if filter.TaskType != "" {
		conditions = append(conditions, fmt.Sprintf("task_type = $%d", argIdx))
		args = append(args, filter.TaskType)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.AssignedTo != "" {
		conditions = append(conditions, fmt.Sprintf("assigned_to = $%d", argIdx))
		args = append(args, filter.AssignedTo)
		argIdx++
	}

	query := `SELECT COUNT(*) FROM tasks`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.queryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count tasks: %w", err)
	}
	return count, nil
}

// CountTasksByStatus returns task counts grouped by status.
// Used by the admin dashboard to show task status distribution.
func (r *TaskRepo) CountTasksByStatus(ctx context.Context) (map[domain.TaskStatus]int, error) {
	const query = `SELECT status, COUNT(*) FROM tasks GROUP BY status`

	rows, err := r.query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("count tasks by status: %w", err)
	}
	defer rows.Close()

	result := make(map[domain.TaskStatus]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan task status count: %w", err)
		}
		result[domain.TaskStatus(status)] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate task status counts: %w", err)
	}
	return result, nil
}

// ── Wave ────────────────────────────────────────────────────

// CreateWave inserts a new wave.
func (r *TaskRepo) CreateWave(ctx context.Context, w *domain.Wave) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	w.CreatedAt = time.Now()
	if w.Status == "" {
		w.Status = domain.WaveStatusCreated
	}
	if w.OrderIDs == nil {
		w.OrderIDs = []uuid.UUID{}
	}
	if w.TaskIDs == nil {
		w.TaskIDs = []uuid.UUID{}
	}

	const query = `
		INSERT INTO waves (id, wave_no, warehouse_id, wave_type, status,
		                   order_ids, task_ids, total_orders, total_lines, total_qty, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.exec(ctx, query,
		w.ID, w.WaveNo, w.WarehouseID, w.WaveType, w.Status,
		w.OrderIDs, w.TaskIDs,
		w.TotalOrders, w.TotalLines, w.TotalQty,
		w.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create wave: %w", err)
	}
	return nil
}

// GetWave retrieves a wave by ID.
func (r *TaskRepo) GetWave(ctx context.Context, id uuid.UUID) (*domain.Wave, error) {
	const query = `
		SELECT id, wave_no, warehouse_id, wave_type, status,
		       order_ids, task_ids, total_orders, total_lines, total_qty,
		       created_at, released_at, completed_at
		FROM waves WHERE id = $1`

	w, err := r.scanWave(r.queryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get wave %s: %w", id, err)
		}
		return nil, fmt.Errorf("get wave: %w", err)
	}
	return w, nil
}

// ListWaves returns waves matching the specified filter.
func (r *TaskRepo) ListWaves(ctx context.Context, filter repository.WaveFilter) ([]*domain.Wave, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.WarehouseID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("warehouse_id = $%d", argIdx))
		args = append(args, filter.WarehouseID)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.WaveType != "" {
		conditions = append(conditions, fmt.Sprintf("wave_type = $%d", argIdx))
		args = append(args, filter.WaveType)
		argIdx++
	}

	query := `
		SELECT id, wave_no, warehouse_id, wave_type, status,
		       order_ids, task_ids, total_orders, total_lines, total_qty,
		       created_at, released_at, completed_at
		FROM waves`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
		argIdx++
	}

	rows, err := r.query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list waves: %w", err)
	}
	defer rows.Close()

	var waves []*domain.Wave
	for rows.Next() {
		w, err := r.scanWaveFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan wave: %w", err)
		}
		waves = append(waves, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate waves: %w", err)
	}
	return waves, nil
}

// UpdateWaveStatus transitions a wave to a new status.
func (r *TaskRepo) UpdateWaveStatus(ctx context.Context, id uuid.UUID, status domain.WaveStatus) error {
	now := time.Now()
	var releasedAt, completedAt *time.Time

	switch status {
	case domain.WaveStatusReleased:
		releasedAt = &now
	case domain.WaveStatusCompleted:
		completedAt = &now
	}

	const query = `
		UPDATE waves SET status=$1, released_at=COALESCE($2, released_at),
		                 completed_at=$3
		WHERE id=$4`

	tag, err := r.exec(ctx, query, status, releasedAt, completedAt, id)
	if err != nil {
		return fmt.Errorf("update wave status: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("update wave status %s: not found", id)
	}
	return nil
}

// CountWaves returns the total count of waves matching the filter.
func (r *TaskRepo) CountWaves(ctx context.Context, filter repository.WaveFilter) (int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.WarehouseID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("warehouse_id = $%d", argIdx))
		args = append(args, filter.WarehouseID)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.WaveType != "" {
		conditions = append(conditions, fmt.Sprintf("wave_type = $%d", argIdx))
		args = append(args, filter.WaveType)
		argIdx++
	}

	query := `SELECT COUNT(*) FROM waves`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.queryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count waves: %w", err)
	}
	return count, nil
}

// AddWaveOrders appends order IDs to a wave and recalculates totals.
// Only allowed when wave status is "created".
func (r *TaskRepo) AddWaveOrders(ctx context.Context, id uuid.UUID, orderIDs []uuid.UUID) error {
	const query = `
		UPDATE waves
		SET order_ids = ARRAY(
			SELECT DISTINCT UNNEST(order_ids || $2::uuid[])
		),
		total_orders = (
			SELECT COUNT(DISTINCT oid) FROM UNNEST(order_ids || $2::uuid[]) AS oid
		)
		WHERE id = $1 AND status = 'created'`

	tag, err := r.exec(ctx, query, id, orderIDs)
	if err != nil {
		return fmt.Errorf("add wave orders: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("add wave orders %s: wave not found or not in created status", id)
	}
	return nil
}

// RemoveWaveOrders removes order IDs from a wave and recalculates totals.
// Only allowed when wave status is "created".
func (r *TaskRepo) RemoveWaveOrders(ctx context.Context, id uuid.UUID, orderIDs []uuid.UUID) error {
	const query = `
		UPDATE waves
		SET order_ids = ARRAY(
			SELECT oid FROM UNNEST(order_ids) AS oid
			WHERE oid <> ALL($2::uuid[])
		),
		total_orders = (
			SELECT COUNT(*) FROM (
				SELECT oid FROM UNNEST(order_ids) AS oid
				WHERE oid <> ALL($2::uuid[])
			) t
		)
		WHERE id = $1 AND status = 'created'`

	tag, err := r.exec(ctx, query, id, orderIDs)
	if err != nil {
		return fmt.Errorf("remove wave orders: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("remove wave orders %s: wave not found or not in created status", id)
	}
	return nil
}

// ── Helpers ────────────────────────────────────────────────

// scanTask scans a single task row.
func (r *TaskRepo) scanTask(row pgx.Row) (*domain.Task, error) {
	t := &domain.Task{}
	var orderID, orderLineID, fromLocation, toLocation *uuid.UUID
	var assignedTo, batchNo, instructions *string

	err := row.Scan(
		&t.ID, &t.TaskNo, &t.TaskType, &t.WarehouseID,
		&orderID, &orderLineID,
		&t.Priority, &t.Status, &assignedTo,
		&fromLocation, &toLocation,
		&t.SKUID, &t.ExpectedQty, &t.ActualQty,
		&t.UOM, &batchNo, &instructions,
		&t.CreatedAt, &t.StartedAt, &t.CompletedAt, &t.CancelledAt,
	)
	if err != nil {
		return nil, err
	}

	if orderID != nil {
		t.OrderID = orderID
	}
	if orderLineID != nil {
		t.OrderLineID = orderLineID
	}
	if assignedTo != nil {
		t.AssignedTo = *assignedTo
	}
	if fromLocation != nil {
		t.FromLocation = fromLocation
	}
	if toLocation != nil {
		t.ToLocation = toLocation
	}
	if batchNo != nil {
		t.BatchNo = *batchNo
	}
	if instructions != nil {
		t.Instructions = *instructions
	}

	return t, nil
}

// scanTaskFromRows scans a task row from a Rows iterator.
func (r *TaskRepo) scanTaskFromRows(rows pgx.Rows) (*domain.Task, error) {
	t := &domain.Task{}
	var orderID, orderLineID, fromLocation, toLocation *uuid.UUID
	var assignedTo, batchNo, instructions *string

	err := rows.Scan(
		&t.ID, &t.TaskNo, &t.TaskType, &t.WarehouseID,
		&orderID, &orderLineID,
		&t.Priority, &t.Status, &assignedTo,
		&fromLocation, &toLocation,
		&t.SKUID, &t.ExpectedQty, &t.ActualQty,
		&t.UOM, &batchNo, &instructions,
		&t.CreatedAt, &t.StartedAt, &t.CompletedAt, &t.CancelledAt,
	)
	if err != nil {
		return nil, err
	}

	if orderID != nil {
		t.OrderID = orderID
	}
	if orderLineID != nil {
		t.OrderLineID = orderLineID
	}
	if assignedTo != nil {
		t.AssignedTo = *assignedTo
	}
	if fromLocation != nil {
		t.FromLocation = fromLocation
	}
	if toLocation != nil {
		t.ToLocation = toLocation
	}
	if batchNo != nil {
		t.BatchNo = *batchNo
	}
	if instructions != nil {
		t.Instructions = *instructions
	}

	return t, nil
}

// scanWave scans a single wave row.
func (r *TaskRepo) scanWave(row pgx.Row) (*domain.Wave, error) {
	w := &domain.Wave{}

	err := row.Scan(
		&w.ID, &w.WaveNo, &w.WarehouseID, &w.WaveType, &w.Status,
		&w.OrderIDs, &w.TaskIDs,
		&w.TotalOrders, &w.TotalLines, &w.TotalQty,
		&w.CreatedAt, &w.ReleasedAt, &w.CompletedAt,
	)
	if err != nil {
		return nil, err
	}

	if w.OrderIDs == nil {
		w.OrderIDs = []uuid.UUID{}
	}
	if w.TaskIDs == nil {
		w.TaskIDs = []uuid.UUID{}
	}

	return w, nil
}

// scanWaveFromRows scans a wave row from a Rows iterator.
func (r *TaskRepo) scanWaveFromRows(rows pgx.Rows) (*domain.Wave, error) {
	w := &domain.Wave{}

	err := rows.Scan(
		&w.ID, &w.WaveNo, &w.WarehouseID, &w.WaveType, &w.Status,
		&w.OrderIDs, &w.TaskIDs,
		&w.TotalOrders, &w.TotalLines, &w.TotalQty,
		&w.CreatedAt, &w.ReleasedAt, &w.CompletedAt,
	)
	if err != nil {
		return nil, err
	}

	if w.OrderIDs == nil {
		w.OrderIDs = []uuid.UUID{}
	}
	if w.TaskIDs == nil {
		w.TaskIDs = []uuid.UUID{}
	}

	return w, nil
}

// ── Transaction-aware dispatch helpers ─────────────────────

// exec dispatches to the active pgx.Tx if one exists in the context,
// otherwise falls back to the connection pool.
func (r *TaskRepo) exec(ctx context.Context, sql string, args ...any) (int64, error) {
	if tx := TxFromContext(ctx); tx != nil {
		tag, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			return 0, err
		}
		return tag.RowsAffected(), nil
	}
	tag, err := r.db.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// query dispatches to the active pgx.Tx if one exists in the context,
// otherwise falls back to the connection pool.
func (r *TaskRepo) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.Query(ctx, sql, args...)
	}
	return r.db.Pool.Query(ctx, sql, args...)
}

// queryRow dispatches to the active pgx.Tx if one exists in the context,
// otherwise falls back to the connection pool.
func (r *TaskRepo) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.QueryRow(ctx, sql, args...)
	}
	return r.db.Pool.QueryRow(ctx, sql, args...)
}
