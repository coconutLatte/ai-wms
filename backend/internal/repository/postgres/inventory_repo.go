package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// InventoryRepo implements repository.InventoryRepository using PostgreSQL.
type InventoryRepo struct {
	db *DB
}

// NewInventoryRepo creates a new InventoryRepo.
func NewInventoryRepo(db *DB) *InventoryRepo {
	return &InventoryRepo{db: db}
}

// ── SKU ────────────────────────────────────────────────────

// CreateSKU inserts a new SKU.
func (r *InventoryRepo) CreateSKU(ctx context.Context, s *domain.SKU) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	s.CreatedAt = time.Now()
	s.UpdatedAt = s.CreatedAt
	if s.Status == "" {
		s.Status = domain.SKUStatusActive
	}
	if s.UOM.BaseUnit == "" {
		s.UOM.BaseUnit = "EA"
	}
	if s.UOM.PackQty == 0 {
		s.UOM.PackQty = 1
	}

	attrsJSON, err := json.Marshal(s.Attributes)
	if err != nil {
		return fmt.Errorf("marshal attributes: %w", err)
	}

	const query = `
		INSERT INTO skus (id, code, name, description, barcode, base_unit, pack_unit, pack_qty,
		                  weight, volume, length, width, height, category, attributes, status,
		                  created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	_, err = r.exec(ctx, query,
		s.ID, s.Code, s.Name, s.Description, nullString(s.Barcode),
		s.UOM.BaseUnit, s.UOM.PackUnit, s.UOM.PackQty,
		s.UOM.Weight, s.UOM.Volume, s.UOM.Length, s.UOM.Width, s.UOM.Height,
		s.Category, attrsJSON, s.Status, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create sku: %w", err)
	}
	return nil
}

// GetSKU retrieves a SKU by ID.
func (r *InventoryRepo) GetSKU(ctx context.Context, id uuid.UUID) (*domain.SKU, error) {
	const query = `
		SELECT id, code, name, description, barcode, base_unit, pack_unit, pack_qty,
		       weight, volume, length, width, height, category, attributes, status,
		       created_at, updated_at
		FROM skus WHERE id = $1`

	s, err := r.scanSKU(r.queryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get sku %s: %w", id, err)
		}
		return nil, fmt.Errorf("get sku: %w", err)
	}
	return s, nil
}

// GetSKUByBarcode retrieves a SKU by its barcode.
func (r *InventoryRepo) GetSKUByBarcode(ctx context.Context, barcode string) (*domain.SKU, error) {
	const query = `
		SELECT id, code, name, description, barcode, base_unit, pack_unit, pack_qty,
		       weight, volume, length, width, height, category, attributes, status,
		       created_at, updated_at
		FROM skus WHERE barcode = $1`

	s, err := r.scanSKU(r.queryRow(ctx, query, barcode))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get sku by barcode %s: %w", barcode, err)
		}
		return nil, fmt.Errorf("get sku by barcode: %w", err)
	}
	return s, nil
}

// GetSKUByCode retrieves a SKU by its code.
func (r *InventoryRepo) GetSKUByCode(ctx context.Context, code string) (*domain.SKU, error) {
	const query = `
		SELECT id, code, name, description, barcode, base_unit, pack_unit, pack_qty,
		       weight, volume, length, width, height, category, attributes, status,
		       created_at, updated_at
		FROM skus WHERE code = $1`

	s, err := r.scanSKU(r.queryRow(ctx, query, code))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get sku by code %s: %w", code, err)
		}
		return nil, fmt.Errorf("get sku by code: %w", err)
	}
	return s, nil
}

// ListSKUs returns all SKUs.
func (r *InventoryRepo) ListSKUs(ctx context.Context) ([]*domain.SKU, error) {
	const query = `
		SELECT id, code, name, description, barcode, base_unit, pack_unit, pack_qty,
		       weight, volume, length, width, height, category, attributes, status,
		       created_at, updated_at
		FROM skus ORDER BY created_at DESC`

	rows, err := r.query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list skus: %w", err)
	}
	defer rows.Close()

	var skus []*domain.SKU
	for rows.Next() {
		s, err := r.scanSKUFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan sku: %w", err)
		}
		skus = append(skus, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate skus: %w", err)
	}
	return skus, nil
}

// UpdateSKU updates an existing SKU.
func (r *InventoryRepo) UpdateSKU(ctx context.Context, s *domain.SKU) error {
	s.UpdatedAt = time.Now()

	attrsJSON, err := json.Marshal(s.Attributes)
	if err != nil {
		return fmt.Errorf("marshal attributes: %w", err)
	}

	const query = `
		UPDATE skus SET name=$1, description=$2, barcode=$3, base_unit=$4, pack_unit=$5,
		                pack_qty=$6, weight=$7, volume=$8, length=$9, width=$10, height=$11,
		                category=$12, attributes=$13, status=$14, updated_at=$15
		WHERE id=$16`

	rowsAffected, err := r.exec(ctx, query,
		s.Name, s.Description, s.Barcode,
		s.UOM.BaseUnit, s.UOM.PackUnit, s.UOM.PackQty,
		s.UOM.Weight, s.UOM.Volume, s.UOM.Length, s.UOM.Width, s.UOM.Height,
		s.Category, attrsJSON, s.Status, s.UpdatedAt, s.ID,
	)
	if err != nil {
		return fmt.Errorf("update sku: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("update sku %s: not found", s.ID)
	}
	return nil
}

// ── Inventory ──────────────────────────────────────────────

// CreateInventory inserts a new inventory record.
func (r *InventoryRepo) CreateInventory(ctx context.Context, inv *domain.Inventory) error {
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	inv.ReceivedAt = time.Now()
	inv.UpdatedAt = inv.ReceivedAt
	if inv.Status == "" {
		inv.Status = domain.InventoryStatusAvailable
	}

	const query = `
		INSERT INTO inventory (id, sku_id, location_id, warehouse_id, batch_no,
		                       qty, reserved_qty, status, production_date, expiry_date,
		                       received_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.exec(ctx, query,
		inv.ID, inv.SKUID, inv.LocationID, inv.WarehouseID,
		nullString(inv.BatchNo),
		inv.Qty, inv.ReservedQty, inv.Status,
		inv.ProductionDate, inv.ExpiryDate,
		inv.ReceivedAt, inv.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create inventory: %w", err)
	}
	return nil
}

// GetInventory retrieves an inventory record by ID.
func (r *InventoryRepo) GetInventory(ctx context.Context, id uuid.UUID) (*domain.Inventory, error) {
	const query = `
		SELECT id, sku_id, location_id, warehouse_id, batch_no,
		       qty, reserved_qty, status, production_date, expiry_date,
		       received_at, updated_at
		FROM inventory WHERE id = $1`

	inv, err := r.scanInventory(r.queryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get inventory %s: %w", id, err)
		}
		return nil, fmt.Errorf("get inventory: %w", err)
	}
	return inv, nil
}

// GetInventoryAtLocation retrieves inventory for a specific SKU at a specific location and batch.
func (r *InventoryRepo) GetInventoryAtLocation(ctx context.Context, skuID, locationID uuid.UUID, batchNo string) (*domain.Inventory, error) {
	const query = `
		SELECT id, sku_id, location_id, warehouse_id, batch_no,
		       qty, reserved_qty, status, production_date, expiry_date,
		       received_at, updated_at
		FROM inventory WHERE sku_id = $1 AND location_id = $2 AND batch_no IS NOT DISTINCT FROM $3`

	row := r.queryRow(ctx, query, skuID, locationID, nullString(batchNo))
	inv, err := r.scanInventory(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get inventory at location (sku=%s, loc=%s): %w", skuID, locationID, err)
		}
		return nil, fmt.Errorf("get inventory at location: %w", err)
	}
	return inv, nil
}

// QueryInventory searches inventory with optional filters.
func (r *InventoryRepo) QueryInventory(ctx context.Context, filter repository.InventoryFilter) ([]*domain.Inventory, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.WarehouseID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("warehouse_id = $%d", argIdx))
		args = append(args, filter.WarehouseID)
		argIdx++
	}
	if filter.SKUID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("sku_id = $%d", argIdx))
		args = append(args, filter.SKUID)
		argIdx++
	}
	if filter.LocationID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("location_id = $%d", argIdx))
		args = append(args, filter.LocationID)
		argIdx++
	}
	if filter.BatchNo != "" {
		conditions = append(conditions, fmt.Sprintf("batch_no = $%d", argIdx))
		args = append(args, filter.BatchNo)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	query := `
		SELECT id, sku_id, location_id, warehouse_id, batch_no,
		       qty, reserved_qty, status, production_date, expiry_date,
		       received_at, updated_at
		FROM inventory`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY received_at DESC"

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
		return nil, fmt.Errorf("query inventory: %w", err)
	}
	defer rows.Close()

	var results []*domain.Inventory
	for rows.Next() {
		inv, err := r.scanInventoryFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan inventory: %w", err)
		}
		results = append(results, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inventory: %w", err)
	}
	return results, nil
}

// UpdateInventoryQty atomically adjusts inventory quantity and reserved quantity.
func (r *InventoryRepo) UpdateInventoryQty(ctx context.Context, id uuid.UUID, deltaQty, deltaReserved float64) error {
	const query = `
		UPDATE inventory
		SET qty = qty + $2, reserved_qty = reserved_qty + $3, updated_at = $4
		WHERE id = $1`

	rowsAffected, err := r.exec(ctx, query, id, deltaQty, deltaReserved, time.Now())
	if err != nil {
		return fmt.Errorf("update inventory qty: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("update inventory qty %s: not found", id)
	}
	return nil
}

// ── Inventory Transaction ──────────────────────────────────

// CreateTransaction records a new inventory transaction.
func (r *InventoryRepo) CreateTransaction(ctx context.Context, tx *domain.InventoryTransaction) error {
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	tx.CreatedAt = time.Now()

	const query = `
		INSERT INTO inventory_transactions (id, inventory_id, sku_id, location_id,
		                                    type, delta_qty, resulting_qty,
		                                    reference_type, reference_id,
		                                    created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.exec(ctx, query,
		tx.ID, tx.InventoryID, tx.SKUID, tx.LocationID,
		tx.Type, tx.DeltaQty, tx.ResultingQty,
		nullString(tx.ReferenceType), nullUUID(tx.ReferenceID),
		tx.CreatedAt, nullString(tx.CreatedBy),
	)
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}
	return nil
}

// ListTransactions returns all transactions for an inventory record.
func (r *InventoryRepo) ListTransactions(ctx context.Context, inventoryID uuid.UUID) ([]*domain.InventoryTransaction, error) {
	const query = `
		SELECT id, inventory_id, sku_id, location_id,
		       type, delta_qty, resulting_qty,
		       reference_type, reference_id,
		       created_at, created_by
		FROM inventory_transactions
		WHERE inventory_id = $1
		ORDER BY created_at DESC`

	rows, err := r.query(ctx, query, inventoryID)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var txs []*domain.InventoryTransaction
	for rows.Next() {
		tx := &domain.InventoryTransaction{}
		var refType, createdBy *string
		var refID *uuid.UUID

		if err := rows.Scan(
			&tx.ID, &tx.InventoryID, &tx.SKUID, &tx.LocationID,
			&tx.Type, &tx.DeltaQty, &tx.ResultingQty,
			&refType, &refID,
			&tx.CreatedAt, &createdBy,
		); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}

		if refType != nil {
			tx.ReferenceType = *refType
		}
		if refID != nil {
			tx.ReferenceID = *refID
		}
		if createdBy != nil {
			tx.CreatedBy = *createdBy
		}

		txs = append(txs, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transactions: %w", err)
	}
	return txs, nil
}

// ── Helpers ────────────────────────────────────────────────

// scanSKU scans a single SKU row.
func (r *InventoryRepo) scanSKU(row pgx.Row) (*domain.SKU, error) {
	s := &domain.SKU{}
	var barcode, packUnit *string
	var weight, volume, length, width, height *float64
	var attrsBytes []byte

	err := row.Scan(
		&s.ID, &s.Code, &s.Name, &s.Description, &barcode,
		&s.UOM.BaseUnit, &packUnit, &s.UOM.PackQty,
		&weight, &volume, &length, &width, &height,
		&s.Category, &attrsBytes, &s.Status,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if barcode != nil {
		s.Barcode = *barcode
	}
	if packUnit != nil {
		s.UOM.PackUnit = *packUnit
	}
	if weight != nil {
		s.UOM.Weight = *weight
	}
	if volume != nil {
		s.UOM.Volume = *volume
	}
	if length != nil {
		s.UOM.Length = *length
	}
	if width != nil {
		s.UOM.Width = *width
	}
	if height != nil {
		s.UOM.Height = *height
	}

	s.Attributes = make(domain.Attributes)
	if len(attrsBytes) > 0 {
		if err := json.Unmarshal(attrsBytes, &s.Attributes); err != nil {
			return nil, fmt.Errorf("unmarshal attributes: %w", err)
		}
	}

	return s, nil
}

// scanSKUFromRows scans a SKU row from a Rows iterator.
func (r *InventoryRepo) scanSKUFromRows(rows pgx.Rows) (*domain.SKU, error) {
	s := &domain.SKU{}
	var barcode, packUnit *string
	var weight, volume, length, width, height *float64
	var attrsBytes []byte

	err := rows.Scan(
		&s.ID, &s.Code, &s.Name, &s.Description, &barcode,
		&s.UOM.BaseUnit, &packUnit, &s.UOM.PackQty,
		&weight, &volume, &length, &width, &height,
		&s.Category, &attrsBytes, &s.Status,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if barcode != nil {
		s.Barcode = *barcode
	}
	if packUnit != nil {
		s.UOM.PackUnit = *packUnit
	}
	if weight != nil {
		s.UOM.Weight = *weight
	}
	if volume != nil {
		s.UOM.Volume = *volume
	}
	if length != nil {
		s.UOM.Length = *length
	}
	if width != nil {
		s.UOM.Width = *width
	}
	if height != nil {
		s.UOM.Height = *height
	}

	s.Attributes = make(domain.Attributes)
	if len(attrsBytes) > 0 {
		if err := json.Unmarshal(attrsBytes, &s.Attributes); err != nil {
			return nil, fmt.Errorf("unmarshal attributes: %w", err)
		}
	}

	return s, nil
}

// scanInventory scans a single inventory row into a domain.Inventory.
func (r *InventoryRepo) scanInventory(row pgx.Row) (*domain.Inventory, error) {
	inv := &domain.Inventory{}
	var batchNo *string

	err := row.Scan(
		&inv.ID, &inv.SKUID, &inv.LocationID, &inv.WarehouseID,
		&batchNo,
		&inv.Qty, &inv.ReservedQty, &inv.Status,
		&inv.ProductionDate, &inv.ExpiryDate,
		&inv.ReceivedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if batchNo != nil {
		inv.BatchNo = *batchNo
	}
	inv.AvailableQty = inv.Qty - inv.ReservedQty

	return inv, nil
}

// scanInventoryFromRows scans an inventory row from a Rows iterator.
func (r *InventoryRepo) scanInventoryFromRows(rows pgx.Rows) (*domain.Inventory, error) {
	inv := &domain.Inventory{}
	var batchNo *string

	err := rows.Scan(
		&inv.ID, &inv.SKUID, &inv.LocationID, &inv.WarehouseID,
		&batchNo,
		&inv.Qty, &inv.ReservedQty, &inv.Status,
		&inv.ProductionDate, &inv.ExpiryDate,
		&inv.ReceivedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if batchNo != nil {
		inv.BatchNo = *batchNo
	}
	inv.AvailableQty = inv.Qty - inv.ReservedQty

	return inv, nil
}

// ── Transaction-aware dispatch helpers ─────────────────────

// exec dispatches to the active pgx.Tx if one exists in the context,
// otherwise falls back to the connection pool.
func (r *InventoryRepo) exec(ctx context.Context, sql string, args ...any) (int64, error) {
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
func (r *InventoryRepo) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.Query(ctx, sql, args...)
	}
	return r.db.Pool.Query(ctx, sql, args...)
}

// queryRow dispatches to the active pgx.Tx if one exists in the context,
// otherwise falls back to the connection pool.
func (r *InventoryRepo) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.QueryRow(ctx, sql, args...)
	}
	return r.db.Pool.QueryRow(ctx, sql, args...)
}

// nullString returns nil for empty strings, so PostgreSQL stores NULL.
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nullUUID returns nil for uuid.Nil, so PostgreSQL stores NULL.
func nullUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}
