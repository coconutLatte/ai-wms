package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// WarehouseRepo implements repository.WarehouseRepository using PostgreSQL.
type WarehouseRepo struct {
	db *DB
}

// NewWarehouseRepo creates a new WarehouseRepo.
func NewWarehouseRepo(db *DB) *WarehouseRepo {
	return &WarehouseRepo{db: db}
}

// ── Warehouse ──────────────────────────────────────────────

// CreateWarehouse inserts a new warehouse.
func (r *WarehouseRepo) CreateWarehouse(ctx context.Context, w *domain.Warehouse) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	w.CreatedAt = time.Now()
	w.UpdatedAt = w.CreatedAt
	if w.Status == "" {
		w.Status = domain.WarehouseStatusActive
	}

	const query = `
		INSERT INTO warehouses (id, code, name, address, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.exec(ctx, query,
		w.ID, w.Code, w.Name, w.Address, w.Status, w.CreatedAt, w.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create warehouse: %w", err)
	}
	return nil
}

// GetWarehouse retrieves a warehouse by ID.
func (r *WarehouseRepo) GetWarehouse(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	const query = `
		SELECT id, code, name, address, status, created_at, updated_at
		FROM warehouses WHERE id = $1`

	w := &domain.Warehouse{}
	err := r.queryRow(ctx, query, id).Scan(
		&w.ID, &w.Code, &w.Name, &w.Address, &w.Status, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get warehouse %s: %w", id, err)
		}
		return nil, fmt.Errorf("get warehouse: %w", err)
	}
	return w, nil
}

// ListWarehouses returns warehouses with optional pagination.
func (r *WarehouseRepo) ListWarehouses(ctx context.Context, limit, offset int) ([]*domain.Warehouse, error) {
	query := `
		SELECT id, code, name, address, status, created_at, updated_at
		FROM warehouses ORDER BY created_at DESC`
	var args []any
	argIdx := 1
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
		argIdx++
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, offset)
	}

	rows, err := r.query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list warehouses: %w", err)
	}
	defer rows.Close()

	var warehouses []*domain.Warehouse
	for rows.Next() {
		w := &domain.Warehouse{}
		if err := rows.Scan(&w.ID, &w.Code, &w.Name, &w.Address, &w.Status, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan warehouse: %w", err)
		}
		warehouses = append(warehouses, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate warehouses: %w", err)
	}
	return warehouses, nil
}

// UpdateWarehouse updates an existing warehouse.
func (r *WarehouseRepo) UpdateWarehouse(ctx context.Context, w *domain.Warehouse) error {
	w.UpdatedAt = time.Now()

	const query = `
		UPDATE warehouses SET name=$1, address=$2, status=$3, updated_at=$4
		WHERE id=$5`

	tag, err := r.exec(ctx, query, w.Name, w.Address, w.Status, w.UpdatedAt, w.ID)
	if err != nil {
		return fmt.Errorf("update warehouse: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("update warehouse %s: not found", w.ID)
	}
	return nil
}

// CountWarehouses returns the total number of warehouses.
func (r *WarehouseRepo) CountWarehouses(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM warehouses`
	var count int
	if err := r.queryRow(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("count warehouses: %w", err)
	}
	return count, nil
}

// ── Zone ───────────────────────────────────────────────────

// CreateZone inserts a new zone.
func (r *WarehouseRepo) CreateZone(ctx context.Context, z *domain.Zone) error {
	if z.ID == uuid.Nil {
		z.ID = uuid.New()
	}
	z.CreatedAt = time.Now()
	z.UpdatedAt = z.CreatedAt
	if z.Status == "" {
		z.Status = domain.ZoneStatusActive
	}

	const query = `
		INSERT INTO zones (id, warehouse_id, code, name, zone_type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.exec(ctx, query,
		z.ID, z.WarehouseID, z.Code, z.Name, z.ZoneType, z.Status, z.CreatedAt, z.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create zone: %w", err)
	}
	return nil
}

// GetZone retrieves a zone by ID.
func (r *WarehouseRepo) GetZone(ctx context.Context, id uuid.UUID) (*domain.Zone, error) {
	const query = `
		SELECT id, warehouse_id, code, name, zone_type, status, created_at, updated_at
		FROM zones WHERE id = $1`

	z := &domain.Zone{}
	err := r.queryRow(ctx, query, id).Scan(
		&z.ID, &z.WarehouseID, &z.Code, &z.Name, &z.ZoneType, &z.Status, &z.CreatedAt, &z.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get zone %s: %w", id, err)
		}
		return nil, fmt.Errorf("get zone: %w", err)
	}
	return z, nil
}

// ListZonesByWarehouse returns all zones in a warehouse with optional pagination.
func (r *WarehouseRepo) ListZonesByWarehouse(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]*domain.Zone, error) {
	query := `
		SELECT id, warehouse_id, code, name, zone_type, status, created_at, updated_at
		FROM zones WHERE warehouse_id = $1 ORDER BY code`
	var args []any
	args = append(args, warehouseID)
	argIdx := 2
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
		argIdx++
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, offset)
	}

	rows, err := r.query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list zones: %w", err)
	}
	defer rows.Close()

	var zones []*domain.Zone
	for rows.Next() {
		z := &domain.Zone{}
		if err := rows.Scan(&z.ID, &z.WarehouseID, &z.Code, &z.Name, &z.ZoneType, &z.Status, &z.CreatedAt, &z.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan zone: %w", err)
		}
		zones = append(zones, z)
	}
	return zones, rows.Err()
}

// CountZonesByWarehouse returns the total number of zones in a warehouse.
func (r *WarehouseRepo) CountZonesByWarehouse(ctx context.Context, warehouseID uuid.UUID) (int, error) {
	const query = `SELECT COUNT(*) FROM zones WHERE warehouse_id = $1`
	var count int
	if err := r.queryRow(ctx, query, warehouseID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count zones by warehouse: %w", err)
	}
	return count, nil
}

// ListAllZones returns zones, optionally filtered by warehouse_id, with pagination.
func (r *WarehouseRepo) ListAllZones(ctx context.Context, filter repository.ZoneFilter) ([]*domain.Zone, error) {
	query := `SELECT id, warehouse_id, code, name, zone_type, status, created_at, updated_at FROM zones`
	var args []any
	argIdx := 1

	if filter.WarehouseID != uuid.Nil {
		query += fmt.Sprintf(" WHERE warehouse_id = $%d", argIdx)
		args = append(args, filter.WarehouseID)
		argIdx++
	}
	query += " ORDER BY code"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)

	}

	rows, err := r.query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list all zones: %w", err)
	}
	defer rows.Close()

	var zones []*domain.Zone
	for rows.Next() {
		z := &domain.Zone{}
		if err := rows.Scan(&z.ID, &z.WarehouseID, &z.Code, &z.Name, &z.ZoneType, &z.Status, &z.CreatedAt, &z.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan zone: %w", err)
		}
		zones = append(zones, z)
	}
	return zones, rows.Err()
}

// CountAllZones returns the total number of zones, optionally filtered by warehouse_id.
func (r *WarehouseRepo) CountAllZones(ctx context.Context, filter repository.ZoneFilter) (int, error) {
	query := `SELECT COUNT(*) FROM zones`
	var args []any
	if filter.WarehouseID != uuid.Nil {
		query += " WHERE warehouse_id = $1"
		args = append(args, filter.WarehouseID)
	}
	var count int
	var err error
	if len(args) > 0 {
		err = r.queryRow(ctx, query, args...).Scan(&count)
	} else {
		err = r.queryRow(ctx, query).Scan(&count)
	}
	if err != nil {
		return 0, fmt.Errorf("count all zones: %w", err)
	}
	return count, nil
}

// UpdateZone updates an existing zone.
func (r *WarehouseRepo) UpdateZone(ctx context.Context, z *domain.Zone) error {
	z.UpdatedAt = time.Now()

	const query = `
		UPDATE zones SET name=$1, zone_type=$2, status=$3, updated_at=$4
		WHERE id=$5`

	tag, err := r.exec(ctx, query, z.Name, z.ZoneType, z.Status, z.UpdatedAt, z.ID)
	if err != nil {
		return fmt.Errorf("update zone: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("update zone %s: not found", z.ID)
	}
	return nil
}

// ── Location ───────────────────────────────────────────────

// CreateLocation inserts a new location.
func (r *WarehouseRepo) CreateLocation(ctx context.Context, loc *domain.Location) error {
	if loc.ID == uuid.Nil {
		loc.ID = uuid.New()
	}
	loc.CreatedAt = time.Now()
	loc.UpdatedAt = loc.CreatedAt
	if loc.Status == "" {
		loc.Status = domain.LocationStatusEmpty
	}

	const query = `
		INSERT INTO locations (id, zone_id, warehouse_id, code, barcode, location_type, status,
		                       max_weight, max_volume, max_qty, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	var maxWeight, maxVolume *float64
	var maxQty *int
	if loc.Capacity != nil {
		maxWeight = &loc.Capacity.MaxWeight
		maxVolume = &loc.Capacity.MaxVolume
		maxQty = &loc.Capacity.MaxQty
	}

	_, err := r.exec(ctx, query,
		loc.ID, loc.ZoneID, loc.WarehouseID, loc.Code, loc.Barcode,
		loc.LocationType, loc.Status, maxWeight, maxVolume, maxQty,
		loc.CreatedAt, loc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create location: %w", err)
	}
	return nil
}

// GetLocation retrieves a location by ID.
func (r *WarehouseRepo) GetLocation(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	const query = `
		SELECT id, zone_id, warehouse_id, code, barcode, location_type, status,
		       max_weight, max_volume, max_qty, created_at, updated_at
		FROM locations WHERE id = $1`

	loc := &domain.Location{}
	var maxWeight, maxVolume *float64
	var maxQty *int

	err := r.queryRow(ctx, query, id).Scan(
		&loc.ID, &loc.ZoneID, &loc.WarehouseID, &loc.Code, &loc.Barcode,
		&loc.LocationType, &loc.Status, &maxWeight, &maxVolume, &maxQty,
		&loc.CreatedAt, &loc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get location %s: %w", id, err)
		}
		return nil, fmt.Errorf("get location: %w", err)
	}

	if maxWeight != nil || maxVolume != nil || maxQty != nil {
		loc.Capacity = &domain.Capacity{}
		if maxWeight != nil {
			loc.Capacity.MaxWeight = *maxWeight
		}
		if maxVolume != nil {
			loc.Capacity.MaxVolume = *maxVolume
		}
		if maxQty != nil {
			loc.Capacity.MaxQty = *maxQty
		}
	}

	return loc, nil
}

// GetLocationByBarcode retrieves a location by its barcode.
func (r *WarehouseRepo) GetLocationByBarcode(ctx context.Context, barcode string) (*domain.Location, error) {
	const query = `
		SELECT id, zone_id, warehouse_id, code, barcode, location_type, status,
		       max_weight, max_volume, max_qty, created_at, updated_at
		FROM locations WHERE barcode = $1`

	loc := &domain.Location{}
	var maxWeight, maxVolume *float64
	var maxQty *int

	err := r.queryRow(ctx, query, barcode).Scan(
		&loc.ID, &loc.ZoneID, &loc.WarehouseID, &loc.Code, &loc.Barcode,
		&loc.LocationType, &loc.Status, &maxWeight, &maxVolume, &maxQty,
		&loc.CreatedAt, &loc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get location by barcode %s: %w", barcode, err)
		}
		return nil, fmt.Errorf("get location by barcode: %w", err)
	}

	if maxWeight != nil || maxVolume != nil || maxQty != nil {
		loc.Capacity = &domain.Capacity{}
		if maxWeight != nil {
			loc.Capacity.MaxWeight = *maxWeight
		}
		if maxVolume != nil {
			loc.Capacity.MaxVolume = *maxVolume
		}
		if maxQty != nil {
			loc.Capacity.MaxQty = *maxQty
		}
	}

	return loc, nil
}

// ListLocationsByZone returns all locations in a zone with optional pagination.
func (r *WarehouseRepo) ListLocationsByZone(ctx context.Context, zoneID uuid.UUID, limit, offset int) ([]*domain.Location, error) {
	query := `
		SELECT id, zone_id, warehouse_id, code, barcode, location_type, status,
		       max_weight, max_volume, max_qty, created_at, updated_at
		FROM locations WHERE zone_id = $1 ORDER BY code`
	var args []any
	args = append(args, zoneID)
	argIdx := 2
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
		argIdx++
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, offset)

	}

	rows, err := r.query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list locations: %w", err)
	}
	defer rows.Close()

	var locations []*domain.Location
	for rows.Next() {
		loc := &domain.Location{}
		var maxWeight, maxVolume *float64
		var maxQty *int

		if err := rows.Scan(
			&loc.ID, &loc.ZoneID, &loc.WarehouseID, &loc.Code, &loc.Barcode,
			&loc.LocationType, &loc.Status, &maxWeight, &maxVolume, &maxQty,
			&loc.CreatedAt, &loc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}

		if maxWeight != nil || maxVolume != nil || maxQty != nil {
			loc.Capacity = &domain.Capacity{}
			if maxWeight != nil {
				loc.Capacity.MaxWeight = *maxWeight
			}
			if maxVolume != nil {
				loc.Capacity.MaxVolume = *maxVolume
			}
			if maxQty != nil {
				loc.Capacity.MaxQty = *maxQty
			}
		}

		locations = append(locations, loc)
	}
	return locations, rows.Err()
}

// CountLocationsByZone returns the total number of locations in a zone.
func (r *WarehouseRepo) CountLocationsByZone(ctx context.Context, zoneID uuid.UUID) (int, error) {
	const query = `SELECT COUNT(*) FROM locations WHERE zone_id = $1`
	var count int
	if err := r.queryRow(ctx, query, zoneID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count locations by zone: %w", err)
	}
	return count, nil
}

// UpdateLocationStatus updates the status of a location.
func (r *WarehouseRepo) UpdateLocationStatus(ctx context.Context, id uuid.UUID, status domain.LocationStatus) error {
	const query = `UPDATE locations SET status=$1, updated_at=$2 WHERE id=$3`

	tag, err := r.exec(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update location status: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("update location status %s: not found", id)
	}
	return nil
}

// ListAllLocations returns locations, optionally filtered by zone_id or warehouse_id, with pagination.
func (r *WarehouseRepo) ListAllLocations(ctx context.Context, filter repository.LocationFilter) ([]*domain.Location, error) {
	query := `
		SELECT id, zone_id, warehouse_id, code, barcode, location_type, status,
		       max_weight, max_volume, max_qty, created_at, updated_at
		FROM locations`
	var args []any
	argIdx := 1
	var conditions []string

	if filter.ZoneID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("zone_id = $%d", argIdx))
		args = append(args, filter.ZoneID)
		argIdx++
	}
	if filter.WarehouseID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("warehouse_id = $%d", argIdx))
		args = append(args, filter.WarehouseID)
		argIdx++
	}
	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}
	query += " ORDER BY code"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)

	}

	rows, err := r.query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list all locations: %w", err)
	}
	defer rows.Close()

	var locations []*domain.Location
	for rows.Next() {
		loc := &domain.Location{}
		var maxWeight, maxVolume *float64
		var maxQty *int

		if err := rows.Scan(
			&loc.ID, &loc.ZoneID, &loc.WarehouseID, &loc.Code, &loc.Barcode,
			&loc.LocationType, &loc.Status, &maxWeight, &maxVolume, &maxQty,
			&loc.CreatedAt, &loc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}

		if maxWeight != nil || maxVolume != nil || maxQty != nil {
			loc.Capacity = &domain.Capacity{}
			if maxWeight != nil {
				loc.Capacity.MaxWeight = *maxWeight
			}
			if maxVolume != nil {
				loc.Capacity.MaxVolume = *maxVolume
			}
			if maxQty != nil {
				loc.Capacity.MaxQty = *maxQty
			}
		}

		locations = append(locations, loc)
	}
	return locations, rows.Err()
}

// CountAllLocations returns the total number of locations, optionally filtered.
func (r *WarehouseRepo) CountAllLocations(ctx context.Context, filter repository.LocationFilter) (int, error) {
	query := `SELECT COUNT(*) FROM locations`
	var args []any
	argIdx := 1
	var conditions []string

	if filter.ZoneID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("zone_id = $%d", argIdx))
		args = append(args, filter.ZoneID)
		argIdx++
	}
	if filter.WarehouseID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("warehouse_id = $%d", argIdx))
		args = append(args, filter.WarehouseID)
	}
	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	var count int
	var err error
	if len(args) > 0 {
		err = r.queryRow(ctx, query, args...).Scan(&count)
	} else {
		err = r.queryRow(ctx, query).Scan(&count)
	}
	if err != nil {
		return 0, fmt.Errorf("count all locations: %w", err)
	}
	return count, nil
}

// UpdateLocation updates an existing location's code, barcode, type, and capacity.
func (r *WarehouseRepo) UpdateLocation(ctx context.Context, l *domain.Location) error {
	l.UpdatedAt = time.Now()

	var maxWeight, maxVolume *float64
	var maxQty *int
	if l.Capacity != nil {
		maxWeight = &l.Capacity.MaxWeight
		maxVolume = &l.Capacity.MaxVolume
		maxQty = &l.Capacity.MaxQty
	}

	const query = `
		UPDATE locations SET code=$1, barcode=$2, location_type=$3,
		       max_weight=$4, max_volume=$5, max_qty=$6, updated_at=$7
		WHERE id=$8`

	tag, err := r.exec(ctx, query,
		l.Code, l.Barcode, l.LocationType,
		maxWeight, maxVolume, maxQty,
		l.UpdatedAt, l.ID,
	)
	if err != nil {
		return fmt.Errorf("update location: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("update location %s: not found", l.ID)
	}
	return nil
}

// ── Transaction-aware dispatch helpers ─────────────────────

// exec dispatches to the active pgx.Tx if one exists in the context,
// otherwise falls back to the connection pool.
func (r *WarehouseRepo) exec(ctx context.Context, sql string, args ...any) (int64, error) {
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
func (r *WarehouseRepo) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.Query(ctx, sql, args...)
	}
	return r.db.Pool.Query(ctx, sql, args...)
}

// queryRow dispatches to the active pgx.Tx if one exists in the context,
// otherwise falls back to the connection pool.
func (r *WarehouseRepo) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.QueryRow(ctx, sql, args...)
	}
	return r.db.Pool.QueryRow(ctx, sql, args...)
}
