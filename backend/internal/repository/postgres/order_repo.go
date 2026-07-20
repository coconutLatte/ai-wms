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

// OrderRepo implements repository.OrderRepository using PostgreSQL.
type OrderRepo struct {
	db *DB
}

// NewOrderRepo creates a new OrderRepo.
func NewOrderRepo(db *DB) *OrderRepo {
	return &OrderRepo{db: db}
}

// ── Order ──────────────────────────────────────────────────

// CreateOrder inserts a new order.
func (r *OrderRepo) CreateOrder(ctx context.Context, o *domain.Order) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	o.CreatedAt = time.Now()
	o.UpdatedAt = o.CreatedAt
	if o.Status == "" {
		o.Status = domain.OrderStatusDraft
	}
	if o.Priority == "" {
		o.Priority = domain.OrderPriorityNormal
	}

	const query = `
		INSERT INTO orders (id, order_no, order_type, warehouse_id, status, priority,
		                    external_ref, external_type, notes, created_at, updated_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.Pool.Exec(ctx, query,
		o.ID, o.OrderNo, o.OrderType, o.WarehouseID, o.Status, o.Priority,
		nullString(o.ExternalRef), nullString(o.ExternalType), nullString(o.Notes),
		o.CreatedAt, o.UpdatedAt, nullString(o.CreatedBy),
	)
	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}
	return nil
}

// GetOrder retrieves an order by ID.
func (r *OrderRepo) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	const query = `
		SELECT id, order_no, order_type, warehouse_id, status, priority,
		       external_ref, external_type, notes, created_at, updated_at, completed_at, created_by
		FROM orders WHERE id = $1`

	o, err := r.scanOrder(r.db.Pool.QueryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get order %s: %w", id, err)
		}
		return nil, fmt.Errorf("get order: %w", err)
	}
	return o, nil
}

// GetOrderByNo retrieves an order by its business order number.
func (r *OrderRepo) GetOrderByNo(ctx context.Context, orderNo string) (*domain.Order, error) {
	const query = `
		SELECT id, order_no, order_type, warehouse_id, status, priority,
		       external_ref, external_type, notes, created_at, updated_at, completed_at, created_by
		FROM orders WHERE order_no = $1`

	o, err := r.scanOrder(r.db.Pool.QueryRow(ctx, query, orderNo))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get order by no %s: %w", orderNo, err)
		}
		return nil, fmt.Errorf("get order by no: %w", err)
	}
	return o, nil
}

// ListOrders returns orders matching the specified filter.
func (r *OrderRepo) ListOrders(ctx context.Context, filter repository.OrderFilter) ([]*domain.Order, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.WarehouseID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("warehouse_id = $%d", argIdx))
		args = append(args, filter.WarehouseID)
		argIdx++
	}
	if filter.OrderType != "" {
		conditions = append(conditions, fmt.Sprintf("order_type = $%d", argIdx))
		args = append(args, filter.OrderType)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	query := `
		SELECT id, order_no, order_type, warehouse_id, status, priority,
		       external_ref, external_type, notes, created_at, updated_at, completed_at, created_by
		FROM orders`
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
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		o, err := r.scanOrderFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}
	return orders, nil
}

// UpdateOrderStatus transitions an order to a new status.
func (r *OrderRepo) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	updatedAt := time.Now()
	var completedAt *time.Time
	if status == domain.OrderStatusCompleted {
		completedAt = &updatedAt
	}

	const query = `
		UPDATE orders SET status=$1, updated_at=$2, completed_at=$3 WHERE id=$4`

	tag, err := r.db.Pool.Exec(ctx, query, status, updatedAt, completedAt, id)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update order status %s: not found", id)
	}
	return nil
}

// ── OrderLine ───────────────────────────────────────────────

// CreateOrderLine inserts a new order line.
func (r *OrderRepo) CreateOrderLine(ctx context.Context, line *domain.OrderLine) error {
	if line.ID == uuid.Nil {
		line.ID = uuid.New()
	}
	if line.Status == "" {
		line.Status = domain.OrderLineStatusPending
	}
	if line.UOM == "" {
		line.UOM = "EA"
	}

	const query = `
		INSERT INTO order_lines (id, order_id, line_no, sku_id, ordered_qty, fulfilled_qty,
		                         uom, batch_no, status, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.Pool.Exec(ctx, query,
		line.ID, line.OrderID, line.LineNo, line.SKUID,
		line.OrderedQty, line.FulfilledQty,
		line.UOM, nullString(line.BatchNo), line.Status, nullString(line.Notes),
	)
	if err != nil {
		return fmt.Errorf("create order line: %w", err)
	}
	return nil
}

// GetOrderLines retrieves all lines for an order, ordered by line_no.
func (r *OrderRepo) GetOrderLines(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderLine, error) {
	const query = `
		SELECT id, order_id, line_no, sku_id, ordered_qty, fulfilled_qty,
		       uom, batch_no, status, notes
		FROM order_lines WHERE order_id = $1 ORDER BY line_no`

	rows, err := r.db.Pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order lines: %w", err)
	}
	defer rows.Close()

	var lines []*domain.OrderLine
	for rows.Next() {
		l, err := r.scanOrderLineFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan order line: %w", err)
		}
		lines = append(lines, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order lines: %w", err)
	}
	return lines, nil
}

// UpdateOrderLineStatus transitions an order line to a new status.
func (r *OrderRepo) UpdateOrderLineStatus(ctx context.Context, id uuid.UUID, status domain.OrderLineStatus) error {
	const query = `UPDATE order_lines SET status=$1 WHERE id=$2`

	tag, err := r.db.Pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update order line status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update order line status %s: not found", id)
	}
	return nil
}

// UpdateOrderLineFulfilledQty sets the fulfilled quantity of an order line.
func (r *OrderRepo) UpdateOrderLineFulfilledQty(ctx context.Context, id uuid.UUID, qty float64) error {
	const query = `UPDATE order_lines SET fulfilled_qty=$1 WHERE id=$2`

	tag, err := r.db.Pool.Exec(ctx, query, qty, id)
	if err != nil {
		return fmt.Errorf("update order line fulfilled qty: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update order line fulfilled qty %s: not found", id)
	}
	return nil
}

// ── ASN ────────────────────────────────────────────────────

// CreateASN inserts a new Advanced Shipping Notice.
func (r *OrderRepo) CreateASN(ctx context.Context, asn *domain.ASN) error {
	if asn.ID == uuid.Nil {
		asn.ID = uuid.New()
	}
	asn.CreatedAt = time.Now()
	if asn.Status == "" {
		asn.Status = domain.ASNStatusPending
	}

	const query = `
		INSERT INTO asns (id, asn_no, warehouse_id, order_id, carrier, tracking_no,
		                  expected_at, arrived_at, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.Pool.Exec(ctx, query,
		asn.ID, asn.ASNNo, asn.WarehouseID, nullUUID(asn.OrderID),
		nullString(asn.Carrier), nullString(asn.TrackingNo),
		asn.ExpectedAt, asn.ArrivedAt, asn.Status, asn.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create asn: %w", err)
	}
	return nil
}

// GetASN retrieves an ASN by ID.
func (r *OrderRepo) GetASN(ctx context.Context, id uuid.UUID) (*domain.ASN, error) {
	const query = `
		SELECT id, asn_no, warehouse_id, order_id, carrier, tracking_no,
		       expected_at, arrived_at, status, created_at
		FROM asns WHERE id = $1`

	asn, err := r.scanASN(r.db.Pool.QueryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get asn %s: %w", id, err)
		}
		return nil, fmt.Errorf("get asn: %w", err)
	}
	return asn, nil
}

// GetASNByNo retrieves an ASN by its business ASN number.
func (r *OrderRepo) GetASNByNo(ctx context.Context, asnNo string) (*domain.ASN, error) {
	const query = `
		SELECT id, asn_no, warehouse_id, order_id, carrier, tracking_no,
		       expected_at, arrived_at, status, created_at
		FROM asns WHERE asn_no = $1`

	asn, err := r.scanASN(r.db.Pool.QueryRow(ctx, query, asnNo))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get asn by no %s: %w", asnNo, err)
		}
		return nil, fmt.Errorf("get asn by no: %w", err)
	}
	return asn, nil
}

// UpdateASNStatus transitions an ASN to a new status.
func (r *OrderRepo) UpdateASNStatus(ctx context.Context, id uuid.UUID, status domain.ASNStatus) error {
	var arrivedAt *time.Time
	if status == domain.ASNStatusArrived {
		now := time.Now()
		arrivedAt = &now
	}

	const query = `UPDATE asns SET status=$1, arrived_at=$2 WHERE id=$3`

	tag, err := r.db.Pool.Exec(ctx, query, status, arrivedAt, id)
	if err != nil {
		return fmt.Errorf("update asn status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update asn status %s: not found", id)
	}
	return nil
}

// ── ASNLine ─────────────────────────────────────────────────

// CreateASNLine inserts a new ASN line.
func (r *OrderRepo) CreateASNLine(ctx context.Context, line *domain.ASNLine) error {
	if line.ID == uuid.Nil {
		line.ID = uuid.New()
	}
	if line.Status == "" {
		line.Status = domain.ASNLineStatusPending
	}

	const query = `
		INSERT INTO asn_lines (id, asn_id, sku_id, expected_qty, received_qty,
		                       batch_no, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Pool.Exec(ctx, query,
		line.ID, line.ASNID, line.SKUID,
		line.ExpectedQty, line.ReceivedQty,
		nullString(line.BatchNo), line.Status,
	)
	if err != nil {
		return fmt.Errorf("create asn line: %w", err)
	}
	return nil
}

// GetASNLines retrieves all lines for an ASN.
func (r *OrderRepo) GetASNLines(ctx context.Context, asnID uuid.UUID) ([]*domain.ASNLine, error) {
	const query = `
		SELECT id, asn_id, sku_id, expected_qty, received_qty,
		       batch_no, status
		FROM asn_lines WHERE asn_id = $1 ORDER BY id`

	rows, err := r.db.Pool.Query(ctx, query, asnID)
	if err != nil {
		return nil, fmt.Errorf("get asn lines: %w", err)
	}
	defer rows.Close()

	var lines []*domain.ASNLine
	for rows.Next() {
		l, err := r.scanASNLineFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan asn line: %w", err)
		}
		lines = append(lines, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate asn lines: %w", err)
	}
	return lines, nil
}

// UpdateASNLineStatus transitions an ASN line to a new status.
func (r *OrderRepo) UpdateASNLineStatus(ctx context.Context, id uuid.UUID, status domain.ASNLineStatus) error {
	const query = `UPDATE asn_lines SET status=$1 WHERE id=$2`

	tag, err := r.db.Pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update asn line status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update asn line status %s: not found", id)
	}
	return nil
}

// UpdateASNLineReceivedQty sets the received quantity of an ASN line.
func (r *OrderRepo) UpdateASNLineReceivedQty(ctx context.Context, id uuid.UUID, qty float64) error {
	const query = `UPDATE asn_lines SET received_qty=$1 WHERE id=$2`

	tag, err := r.db.Pool.Exec(ctx, query, qty, id)
	if err != nil {
		return fmt.Errorf("update asn line received qty: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update asn line received qty %s: not found", id)
	}
	return nil
}

// ── Helpers ────────────────────────────────────────────────

// scanOrder scans a single order row.
func (r *OrderRepo) scanOrder(row pgx.Row) (*domain.Order, error) {
	o := &domain.Order{}
	var externalRef, externalType, notes, createdBy *string

	err := row.Scan(
		&o.ID, &o.OrderNo, &o.OrderType, &o.WarehouseID, &o.Status, &o.Priority,
		&externalRef, &externalType, &notes,
		&o.CreatedAt, &o.UpdatedAt, &o.CompletedAt, &createdBy,
	)
	if err != nil {
		return nil, err
	}

	if externalRef != nil {
		o.ExternalRef = *externalRef
	}
	if externalType != nil {
		o.ExternalType = *externalType
	}
	if notes != nil {
		o.Notes = *notes
	}
	if createdBy != nil {
		o.CreatedBy = *createdBy
	}

	return o, nil
}

// scanOrderFromRows scans an order row from a Rows iterator.
func (r *OrderRepo) scanOrderFromRows(rows pgx.Rows) (*domain.Order, error) {
	o := &domain.Order{}
	var externalRef, externalType, notes, createdBy *string

	err := rows.Scan(
		&o.ID, &o.OrderNo, &o.OrderType, &o.WarehouseID, &o.Status, &o.Priority,
		&externalRef, &externalType, &notes,
		&o.CreatedAt, &o.UpdatedAt, &o.CompletedAt, &createdBy,
	)
	if err != nil {
		return nil, err
	}

	if externalRef != nil {
		o.ExternalRef = *externalRef
	}
	if externalType != nil {
		o.ExternalType = *externalType
	}
	if notes != nil {
		o.Notes = *notes
	}
	if createdBy != nil {
		o.CreatedBy = *createdBy
	}

	return o, nil
}

// scanOrderLineFromRows scans an order line row from a Rows iterator.
func (r *OrderRepo) scanOrderLineFromRows(rows pgx.Rows) (*domain.OrderLine, error) {
	l := &domain.OrderLine{}
	var batchNo, notes *string

	err := rows.Scan(
		&l.ID, &l.OrderID, &l.LineNo, &l.SKUID,
		&l.OrderedQty, &l.FulfilledQty,
		&l.UOM, &batchNo, &l.Status, &notes,
	)
	if err != nil {
		return nil, err
	}

	if batchNo != nil {
		l.BatchNo = *batchNo
	}
	if notes != nil {
		l.Notes = *notes
	}

	return l, nil
}

// scanASN scans a single ASN row.
func (r *OrderRepo) scanASN(row pgx.Row) (*domain.ASN, error) {
	a := &domain.ASN{}
	var orderID *uuid.UUID
	var carrier, trackingNo *string

	err := row.Scan(
		&a.ID, &a.ASNNo, &a.WarehouseID, &orderID,
		&carrier, &trackingNo,
		&a.ExpectedAt, &a.ArrivedAt, &a.Status, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if orderID != nil {
		a.OrderID = *orderID
	}
	if carrier != nil {
		a.Carrier = *carrier
	}
	if trackingNo != nil {
		a.TrackingNo = *trackingNo
	}

	return a, nil
}

// scanASNLineFromRows scans an ASN line row from a Rows iterator.
func (r *OrderRepo) scanASNLineFromRows(rows pgx.Rows) (*domain.ASNLine, error) {
	l := &domain.ASNLine{}
	var batchNo *string

	err := rows.Scan(
		&l.ID, &l.ASNID, &l.SKUID,
		&l.ExpectedQty, &l.ReceivedQty,
		&batchNo, &l.Status,
	)
	if err != nil {
		return nil, err
	}

	if batchNo != nil {
		l.BatchNo = *batchNo
	}

	return l, nil
}
