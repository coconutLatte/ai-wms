package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
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

	_, err := r.db.Pool.Exec(ctx, query,
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
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
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
		argIdx++
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
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

	tag, err := r.db.Pool.Exec(ctx, query, w.Name, w.Address, w.Status, w.UpdatedAt, w.ID)
	if err != nil {
		return fmt.Errorf("update warehouse: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update warehouse %s: not found", w.ID)
	}
	return nil
}

// CountWarehouses returns the total number of warehouses.
func (r *WarehouseRepo) CountWarehouses(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM warehouses`
	var count int
	if err := r.db.Pool.QueryRow(ctx, query).Scan(&count); err != nil {
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

	_, err := r.db.Pool.Exec(ctx, query,
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
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
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
		argIdx++
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
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
	if err := r.db.Pool.QueryRow(ctx, query, warehouseID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count zones by warehouse: %w", err)
	}
	return count, nil
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

	_, err := r.db.Pool.Exec(ctx, query,
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

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
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

	err := r.db.Pool.QueryRow(ctx, query, barcode).Scan(
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
		argIdx++
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
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
	if err := r.db.Pool.QueryRow(ctx, query, zoneID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count locations by zone: %w", err)
	}
	return count, nil
}

// UpdateLocationStatus updates the status of a location.
func (r *WarehouseRepo) UpdateLocationStatus(ctx context.Context, id uuid.UUID, status domain.LocationStatus) error {
	const query = `UPDATE locations SET status=$1, updated_at=$2 WHERE id=$3`

	tag, err := r.db.Pool.Exec(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update location status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update location status %s: not found", id)
	}
	return nil
}
