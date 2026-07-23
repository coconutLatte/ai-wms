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

// ShipmentRepo implements repository.ShipmentRepository using PostgreSQL.
type ShipmentRepo struct {
	db *DB
}

// NewShipmentRepo creates a new ShipmentRepo.
func NewShipmentRepo(db *DB) *ShipmentRepo {
	return &ShipmentRepo{db: db}
}

// ── Shipment CRUD ──────────────────────────────────────────────

// CreateShipment inserts a new shipment.
func (r *ShipmentRepo) CreateShipment(ctx context.Context, s *domain.Shipment) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now
	if s.Status == "" {
		s.Status = domain.ShipmentStatusPending
	}

	const query = `
		INSERT INTO shipments (id, shipment_no, order_id, warehouse_id,
		                       status, carrier, tracking_no, carrier_service,
		                       estimated_delivery, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.Pool.Exec(ctx, query,
		s.ID, s.ShipmentNo, s.OrderID, s.WarehouseID,
		s.Status, s.Carrier, nullString(s.TrackingNo), nullString(s.CarrierService),
		s.EstimatedDelivery, nullString(s.Notes), s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create shipment: %w", err)
	}
	return nil
}

// GetShipment retrieves a shipment by ID.
func (r *ShipmentRepo) GetShipment(ctx context.Context, id uuid.UUID) (*domain.Shipment, error) {
	const query = `
		SELECT id, shipment_no, order_id, warehouse_id,
		       status, carrier, tracking_no, carrier_service,
		       estimated_delivery, actual_delivery, notes,
		       created_at, updated_at, shipped_at, delivered_at
		FROM shipments WHERE id = $1`

	s, err := r.scanShipment(r.db.Pool.QueryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get shipment %s: %w", id, err)
		}
		return nil, fmt.Errorf("get shipment: %w", err)
	}
	return s, nil
}

// ListShipments returns shipments matching the specified filter.
func (r *ShipmentRepo) ListShipments(ctx context.Context, filter repository.ShipmentFilter) ([]*domain.Shipment, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.WarehouseID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("warehouse_id = $%d", argIdx))
		args = append(args, filter.WarehouseID)
		argIdx++
	}
	if filter.OrderID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("order_id = $%d", argIdx))
		args = append(args, filter.OrderID)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Carrier != "" {
		conditions = append(conditions, fmt.Sprintf("carrier = $%d", argIdx))
		args = append(args, filter.Carrier)
		argIdx++
	}

	query := `
		SELECT id, shipment_no, order_id, warehouse_id,
		       status, carrier, tracking_no, carrier_service,
		       estimated_delivery, actual_delivery, notes,
		       created_at, updated_at, shipped_at, delivered_at
		FROM shipments`
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
		return nil, fmt.Errorf("list shipments: %w", err)
	}
	defer rows.Close()

	var shipments []*domain.Shipment
	for rows.Next() {
		s, err := r.scanShipmentFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan shipment: %w", err)
		}
		shipments = append(shipments, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate shipments: %w", err)
	}
	return shipments, nil
}

// CountShipments returns the total count of shipments matching the filter.
func (r *ShipmentRepo) CountShipments(ctx context.Context, filter repository.ShipmentFilter) (int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.WarehouseID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("warehouse_id = $%d", argIdx))
		args = append(args, filter.WarehouseID)
		argIdx++
	}
	if filter.OrderID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("order_id = $%d", argIdx))
		args = append(args, filter.OrderID)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Carrier != "" {
		conditions = append(conditions, fmt.Sprintf("carrier = $%d", argIdx))
		args = append(args, filter.Carrier)
		argIdx++
	}

	query := `SELECT COUNT(*) FROM shipments`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	if err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count shipments: %w", err)
	}
	return count, nil
}

// UpdateShipmentStatus transitions a shipment to a new status.
func (r *ShipmentRepo) UpdateShipmentStatus(ctx context.Context, id uuid.UUID, status domain.ShipmentStatus) error {
	now := time.Now()
	var shippedAt, deliveredAt *time.Time

	switch status {
	case domain.ShipmentStatusInTransit:
		shippedAt = &now
	case domain.ShipmentStatusDelivered:
		deliveredAt = &now
	}

	const query = `
		UPDATE shipments SET status = $1, updated_at = $2,
		       shipped_at = COALESCE($3, shipped_at),
		       delivered_at = $4
		WHERE id = $5`

	tag, err := r.db.Pool.Exec(ctx, query, status, now, shippedAt, deliveredAt, id)
	if err != nil {
		return fmt.Errorf("update shipment status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update shipment status %s: not found", id)
	}
	return nil
}

// UpdateShipmentTracking updates the tracking information for a shipment.
func (r *ShipmentRepo) UpdateShipmentTracking(ctx context.Context, id uuid.UUID, carrier, trackingNo, carrierService string) error {
	now := time.Now()

	const query = `
		UPDATE shipments SET carrier = $1, tracking_no = $2, carrier_service = $3, updated_at = $4
		WHERE id = $5`

	tag, err := r.db.Pool.Exec(ctx, query, carrier, nullString(trackingNo), nullString(carrierService), now, id)
	if err != nil {
		return fmt.Errorf("update shipment tracking: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update shipment tracking %s: not found", id)
	}
	return nil
}

// DeliverShipment marks a shipment as delivered.
func (r *ShipmentRepo) DeliverShipment(ctx context.Context, id uuid.UUID) error {
	now := time.Now()

	const query = `
		UPDATE shipments SET status = $1, delivered_at = $2, actual_delivery = $3, updated_at = $4
		WHERE id = $5`

	tag, err := r.db.Pool.Exec(ctx, query, domain.ShipmentStatusDelivered, now, now, now, id)
	if err != nil {
		return fmt.Errorf("deliver shipment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("deliver shipment %s: not found", id)
	}
	return nil
}

// ── Scanners ─────────────────────────────────────────────────

func (r *ShipmentRepo) scanShipment(row pgx.Row) (*domain.Shipment, error) {
	s := &domain.Shipment{}
	var trackingNo, carrierService, notes *string

	err := row.Scan(
		&s.ID, &s.ShipmentNo, &s.OrderID, &s.WarehouseID,
		&s.Status, &s.Carrier, &trackingNo, &carrierService,
		&s.EstimatedDelivery, &s.ActualDelivery, &notes,
		&s.CreatedAt, &s.UpdatedAt, &s.ShippedAt, &s.DeliveredAt,
	)
	if err != nil {
		return nil, err
	}

	if trackingNo != nil {
		s.TrackingNo = *trackingNo
	}
	if carrierService != nil {
		s.CarrierService = *carrierService
	}
	if notes != nil {
		s.Notes = *notes
	}

	return s, nil
}

func (r *ShipmentRepo) scanShipmentFromRows(rows pgx.Rows) (*domain.Shipment, error) {
	s := &domain.Shipment{}
	var trackingNo, carrierService, notes *string

	err := rows.Scan(
		&s.ID, &s.ShipmentNo, &s.OrderID, &s.WarehouseID,
		&s.Status, &s.Carrier, &trackingNo, &carrierService,
		&s.EstimatedDelivery, &s.ActualDelivery, &notes,
		&s.CreatedAt, &s.UpdatedAt, &s.ShippedAt, &s.DeliveredAt,
	)
	if err != nil {
		return nil, err
	}

	if trackingNo != nil {
		s.TrackingNo = *trackingNo
	}
	if carrierService != nil {
		s.CarrierService = *carrierService
	}
	if notes != nil {
		s.Notes = *notes
	}

	return s, nil
}
