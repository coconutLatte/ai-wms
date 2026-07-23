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

// CycleCountRepo implements repository.CycleCountRepository using PostgreSQL.
type CycleCountRepo struct {
	db *DB
}

// NewCycleCountRepo creates a new CycleCountRepo.
func NewCycleCountRepo(db *DB) *CycleCountRepo {
	return &CycleCountRepo{db: db}
}

// ── Cycle Count ──────────────────────────────────────────────

// CreateCycleCount inserts a new cycle count.
func (r *CycleCountRepo) CreateCycleCount(ctx context.Context, cc *domain.CycleCount) error {
	if cc.ID == uuid.Nil {
		cc.ID = uuid.New()
	}
	cc.CreatedAt = time.Now()
	if cc.Status == "" {
		cc.Status = domain.CycleCountStatusDraft
	}

	const query = `
		INSERT INTO cycle_counts (id, count_no, warehouse_id, location_id, zone_id,
		                          status, counted_by, notes, total_lines, matched_lines, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.Pool.Exec(ctx, query,
		cc.ID, cc.CountNo, cc.WarehouseID, cc.LocationID, cc.ZoneID,
		cc.Status, nullString(cc.CountedBy), nullString(cc.Notes),
		cc.TotalLines, cc.MatchedLines, cc.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create cycle count: %w", err)
	}
	return nil
}

// GetCycleCount retrieves a cycle count by ID.
func (r *CycleCountRepo) GetCycleCount(ctx context.Context, id uuid.UUID) (*domain.CycleCount, error) {
	const query = `
		SELECT id, count_no, warehouse_id, location_id, zone_id,
		       status, counted_by, notes, total_lines, matched_lines,
		       created_at, started_at, completed_at, approved_at, approved_by
		FROM cycle_counts WHERE id = $1`

	cc, err := r.scanCycleCount(r.db.Pool.QueryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get cycle count %s: %w", id, err)
		}
		return nil, fmt.Errorf("get cycle count: %w", err)
	}
	return cc, nil
}

// ListCycleCounts returns cycle counts matching the specified filter.
func (r *CycleCountRepo) ListCycleCounts(ctx context.Context, filter repository.CycleCountFilter) ([]*domain.CycleCount, error) {
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

	query := `
		SELECT id, count_no, warehouse_id, location_id, zone_id,
		       status, counted_by, notes, total_lines, matched_lines,
		       created_at, started_at, completed_at, approved_at, approved_by
		FROM cycle_counts`
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

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list cycle counts: %w", err)
	}
	defer rows.Close()

	var counts []*domain.CycleCount
	for rows.Next() {
		cc, err := r.scanCycleCountFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan cycle count: %w", err)
		}
		counts = append(counts, cc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cycle counts: %w", err)
	}
	return counts, nil
}

// CountCycleCounts returns the total count of cycle counts matching the filter.
func (r *CycleCountRepo) CountCycleCounts(ctx context.Context, filter repository.CycleCountFilter) (int, error) {
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

	query := `SELECT COUNT(*) FROM cycle_counts`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count cycle counts: %w", err)
	}
	return count, nil
}

// UpdateCycleCountStatus transitions a cycle count to a new status.
func (r *CycleCountRepo) UpdateCycleCountStatus(ctx context.Context, id uuid.UUID, status domain.CycleCountStatus) error {
	now := time.Now()
	var startedAt, completedAt *time.Time

	switch status {
	case domain.CycleCountStatusInProgress:
		startedAt = &now
	case domain.CycleCountStatusPendingReview:
		completedAt = &now
	}

	const query = `
		UPDATE cycle_counts SET status = $1,
		       started_at = COALESCE($2, started_at),
		       completed_at = $3
		WHERE id = $4`

	tag, err := r.db.Pool.Exec(ctx, query, status, startedAt, completedAt, id)
	if err != nil {
		return fmt.Errorf("update cycle count status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update cycle count status %s: not found", id)
	}
	return nil
}

// ApproveCycleCount approves a cycle count with approver information.
func (r *CycleCountRepo) ApproveCycleCount(ctx context.Context, id uuid.UUID, approvedBy string) error {
	now := time.Now()

	const query = `
		UPDATE cycle_counts SET status = $1, approved_at = $2, approved_by = $3
		WHERE id = $4`

	tag, err := r.db.Pool.Exec(ctx, query, domain.CycleCountStatusApproved, now, approvedBy, id)
	if err != nil {
		return fmt.Errorf("approve cycle count: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("approve cycle count %s: not found", id)
	}
	return nil
}

// FinalizeCycleCount updates line counts and transitions count to pending_review.
func (r *CycleCountRepo) FinalizeCycleCount(ctx context.Context, id uuid.UUID, totalLines, matchedLines int) error {
	now := time.Now()

	const query = `
		UPDATE cycle_counts SET status = $1, total_lines = $2, matched_lines = $3, completed_at = $4
		WHERE id = $5`

	tag, err := r.db.Pool.Exec(ctx, query, domain.CycleCountStatusPendingReview, totalLines, matchedLines, now, id)
	if err != nil {
		return fmt.Errorf("finalize cycle count: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("finalize cycle count %s: not found", id)
	}
	return nil
}

// ── Cycle Count Lines ────────────────────────────────────────

// CreateCycleCountLine inserts a new line for a cycle count.
func (r *CycleCountRepo) CreateCycleCountLine(ctx context.Context, line *domain.CycleCountLine) error {
	if line.ID == uuid.Nil {
		line.ID = uuid.New()
	}
	line.CreatedAt = time.Now()
	if line.Status == "" {
		line.Status = domain.CycleCountLineStatusPending
	}

	const query = `
		INSERT INTO cycle_count_lines (id, cycle_count_id, sku_id, location_id,
		                               batch_no, system_qty, counted_qty, variance,
		                               status, counted_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.Pool.Exec(ctx, query,
		line.ID, line.CycleCountID, line.SKUID, line.LocationID,
		nullString(line.BatchNo), line.SystemQty, line.CountedQty, line.Variance,
		line.Status, line.CountedAt, line.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create cycle count line: %w", err)
	}
	return nil
}

// CreateCycleCountLines creates multiple lines in a single batch.
func (r *CycleCountRepo) CreateCycleCountLines(ctx context.Context, lines []*domain.CycleCountLine) error {
	if len(lines) == 0 {
		return nil
	}

	// Build a multi-row insert using pgx's CopyFrom for efficiency, or use
	// a simple loop with a batch. For simplicity, use a single batch insert.
	const query = `
		INSERT INTO cycle_count_lines (id, cycle_count_id, sku_id, location_id,
		                               batch_no, system_qty, counted_qty, variance,
		                               status, counted_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	batch := &pgx.Batch{}
	for _, line := range lines {
		if line.ID == uuid.Nil {
			line.ID = uuid.New()
		}
		line.CreatedAt = time.Now()
		if line.Status == "" {
			line.Status = domain.CycleCountLineStatusPending
		}
		batch.Queue(query,
			line.ID, line.CycleCountID, line.SKUID, line.LocationID,
			nullString(line.BatchNo), line.SystemQty, line.CountedQty, line.Variance,
			line.Status, line.CountedAt, line.CreatedAt,
		)
	}

	br := r.db.Pool.SendBatch(ctx, batch)
	defer br.Close()

	for range lines {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("create cycle count lines batch: %w", err)
		}
	}

	return nil
}

// GetCycleCountLines retrieves all lines for a cycle count.
func (r *CycleCountRepo) GetCycleCountLines(ctx context.Context, cycleCountID uuid.UUID) ([]*domain.CycleCountLine, error) {
	const query = `
		SELECT id, cycle_count_id, sku_id, location_id,
		       batch_no, system_qty, counted_qty, variance,
		       status, counted_at, created_at
		FROM cycle_count_lines
		WHERE cycle_count_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.Pool.Query(ctx, query, cycleCountID)
	if err != nil {
		return nil, fmt.Errorf("get cycle count lines: %w", err)
	}
	defer rows.Close()

	var lines []*domain.CycleCountLine
	for rows.Next() {
		line, err := r.scanCycleCountLineFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan cycle count line: %w", err)
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cycle count lines: %w", err)
	}
	return lines, nil
}

// UpdateCycleCountLine updates a count line with the counted quantity and variance.
func (r *CycleCountRepo) UpdateCycleCountLine(ctx context.Context, id uuid.UUID, countedQty float64, variance float64) error {
	now := time.Now()

	const query = `
		UPDATE cycle_count_lines
		SET counted_qty = $1, variance = $2, status = $3, counted_at = $4
		WHERE id = $5`

	tag, err := r.db.Pool.Exec(ctx, query, countedQty, variance, domain.CycleCountLineStatusCounted, now, id)
	if err != nil {
		return fmt.Errorf("update cycle count line: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update cycle count line %s: not found", id)
	}
	return nil
}

// ── Scanners ─────────────────────────────────────────────────

func (r *CycleCountRepo) scanCycleCount(row pgx.Row) (*domain.CycleCount, error) {
	cc := &domain.CycleCount{}
	var countedBy, notes, approvedBy *string

	err := row.Scan(
		&cc.ID, &cc.CountNo, &cc.WarehouseID, &cc.LocationID, &cc.ZoneID,
		&cc.Status, &countedBy, &notes, &cc.TotalLines, &cc.MatchedLines,
		&cc.CreatedAt, &cc.StartedAt, &cc.CompletedAt, &cc.ApprovedAt, &approvedBy,
	)
	if err != nil {
		return nil, err
	}

	if countedBy != nil {
		cc.CountedBy = *countedBy
	}
	if notes != nil {
		cc.Notes = *notes
	}
	if approvedBy != nil {
		cc.ApprovedBy = *approvedBy
	}

	return cc, nil
}

func (r *CycleCountRepo) scanCycleCountFromRows(rows pgx.Rows) (*domain.CycleCount, error) {
	cc := &domain.CycleCount{}
	var countedBy, notes, approvedBy *string

	err := rows.Scan(
		&cc.ID, &cc.CountNo, &cc.WarehouseID, &cc.LocationID, &cc.ZoneID,
		&cc.Status, &countedBy, &notes, &cc.TotalLines, &cc.MatchedLines,
		&cc.CreatedAt, &cc.StartedAt, &cc.CompletedAt, &cc.ApprovedAt, &approvedBy,
	)
	if err != nil {
		return nil, err
	}

	if countedBy != nil {
		cc.CountedBy = *countedBy
	}
	if notes != nil {
		cc.Notes = *notes
	}
	if approvedBy != nil {
		cc.ApprovedBy = *approvedBy
	}

	return cc, nil
}

func (r *CycleCountRepo) scanCycleCountLineFromRows(rows pgx.Rows) (*domain.CycleCountLine, error) {
	line := &domain.CycleCountLine{}
	var batchNo *string

	err := rows.Scan(
		&line.ID, &line.CycleCountID, &line.SKUID, &line.LocationID,
		&batchNo, &line.SystemQty, &line.CountedQty, &line.Variance,
		&line.Status, &line.CountedAt, &line.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if batchNo != nil {
		line.BatchNo = *batchNo
	}

	return line, nil
}
