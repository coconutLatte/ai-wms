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
func (r *InventoryRepo) ListSKUs(ctx context.Context, limit, offset int) ([]*domain.SKU, error) {
	query := `
		SELECT id, code, name, description, barcode, base_unit, pack_unit, pack_qty,
		       weight, volume, length, width, height, category, attributes, status,
		       created_at, updated_at
		FROM skus ORDER BY created_at DESC`
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

	rows, err := r.query(ctx, query, args...)
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

// CountSKUs returns the total number of SKUs.
func (r *InventoryRepo) CountSKUs(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM skus`

	var count int
	err := r.queryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count skus: %w", err)
	}
	return count, nil
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

// GetAndLockInventory retrieves an inventory record by ID with a row-level
// lock (SELECT ... FOR UPDATE). Must be called inside a transaction; the
// lock is held until the transaction commits or rolls back.
//
// This prevents race conditions in multi-step operations like
// AdjustInventory where the business-rule check (e.g. "qty must not go
// negative") must be performed against a stable snapshot.
func (r *InventoryRepo) GetAndLockInventory(ctx context.Context, id uuid.UUID) (*domain.Inventory, error) {
	const query = `
		SELECT id, sku_id, location_id, warehouse_id, batch_no,
		       qty, reserved_qty, status, production_date, expiry_date,
		       received_at, updated_at
		FROM inventory WHERE id = $1
		FOR UPDATE`

	inv, err := r.scanInventory(r.queryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get and lock inventory %s: %w", id, err)
		}
		return nil, fmt.Errorf("get and lock inventory: %w", err)
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


// UpdateInventoryStatus updates the status of an inventory record.
func (r *InventoryRepo) UpdateInventoryStatus(ctx context.Context, id uuid.UUID, status domain.InventoryStatus) error {
	const query = `
		UPDATE inventory
		SET status = $2, updated_at = $3
		WHERE id = $1`

	rowsAffected, err := r.exec(ctx, query, id, string(status), time.Now())
	if err != nil {
		return fmt.Errorf("update inventory status: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("update inventory status %s: not found", id)
	}
	return nil
}

// CountInventory returns the total count of inventory records matching the filter.
func (r *InventoryRepo) CountInventory(ctx context.Context, filter repository.InventoryFilter) (int, error) {
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

	query := `SELECT COUNT(*) FROM inventory`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.queryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count inventory: %w", err)
	}
	return count, nil
}

// ── FEFO / FIFO Retrieval ─────────────────────────────────

// GetOldestInventory returns available, non-zero inventory records for the given
// filter, sorted by received_at ASC (oldest first — FIFO). This is the default
// retrieval strategy for non-perishable goods.
func (r *InventoryRepo) GetOldestInventory(ctx context.Context, filter repository.InventoryRetrievalFilter) ([]*domain.Inventory, error) {
	return r.queryWithOrder(ctx, filter, "received_at ASC")
}

// GetExpiringInventory returns available, non-zero inventory records for the given
// filter, sorted by expiry_date ASC NULLS LAST (earliest expiring first — FEFO).
// This is the preferred retrieval strategy for perishable goods.
func (r *InventoryRepo) GetExpiringInventory(ctx context.Context, filter repository.InventoryRetrievalFilter) ([]*domain.Inventory, error) {
	return r.queryWithOrder(ctx, filter, "expiry_date ASC NULLS LAST")
}

// queryWithOrder is a shared helper that builds a parameterised query for
// FEFO / FIFO retrieval with a given ORDER BY clause.
func (r *InventoryRepo) queryWithOrder(ctx context.Context, filter repository.InventoryRetrievalFilter, orderClause string) ([]*domain.Inventory, error) {
	var conditions []string
	var args []any
	argIdx := 1

	// Only return available inventory with positive on-hand quantity.
	conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
	args = append(args, domain.InventoryStatusAvailable)
	argIdx++

	conditions = append(conditions, fmt.Sprintf("qty > $%d", argIdx))
	args = append(args, 0.0)
	argIdx++

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

	query := `
		SELECT id, sku_id, location_id, warehouse_id, batch_no,
		       qty, reserved_qty, status, production_date, expiry_date,
		       received_at, updated_at
		FROM inventory
		WHERE ` + strings.Join(conditions, " AND ") + `
		ORDER BY ` + orderClause

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}

	rows, err := r.query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("fefo/fifo query: %w", err)
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
func (r *InventoryRepo) ListTransactions(ctx context.Context, inventoryID uuid.UUID, limit, offset int) ([]*domain.InventoryTransaction, error) {
	query := `
		SELECT id, inventory_id, sku_id, location_id,
		       type, delta_qty, resulting_qty,
		       reference_type, reference_id,
		       created_at, created_by
		FROM inventory_transactions
		WHERE inventory_id = $1
		ORDER BY created_at DESC`
	var args []any
	args = append(args, inventoryID)
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

	rows, err := r.query(ctx, query, args...)
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

// ListTransactionsByReference returns all inventory transactions for a given
// reference (e.g., an order line). Used when unreserving inventory to find
// which inventory records were reserved for a specific order allocation.
func (r *InventoryRepo) ListTransactionsByReference(ctx context.Context, referenceType string, referenceID uuid.UUID) ([]*domain.InventoryTransaction, error) {
	const query = `
			SELECT id, inventory_id, sku_id, location_id,
			       type, delta_qty, resulting_qty,
			       reference_type, reference_id,
			       created_at, created_by
			FROM inventory_transactions
			WHERE reference_type = $1 AND reference_id = $2
			ORDER BY created_at DESC`

	rows, err := r.query(ctx, query, referenceType, referenceID)
	if err != nil {
		return nil, fmt.Errorf("list transactions by reference: %w", err)
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
		return nil, fmt.Errorf("iterate transactions by reference: %w", err)
	}
	return txs, nil
}

// CountTransactions returns the total number of transactions for an inventory record.
func (r *InventoryRepo) CountTransactions(ctx context.Context, inventoryID uuid.UUID) (int, error) {
	const query = `SELECT COUNT(*) FROM inventory_transactions WHERE inventory_id = $1`

	var count int
	err := r.queryRow(ctx, query, inventoryID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count transactions: %w", err)
	}
	return count, nil
}

// ListTransactionsGlobal returns inventory transactions matching the given filter
// across all inventory records, with optional warehouse filter (via JOIN).
func (r *InventoryRepo) ListTransactionsGlobal(ctx context.Context, filter repository.InventoryTxFilter) ([]*domain.InventoryTransaction, error) {
	conditions, args := r.buildTxGlobalConditions(filter, nil)

	query := `
		SELECT t.id, t.inventory_id, t.sku_id, t.location_id,
		       t.type, t.delta_qty, t.resulting_qty,
		       t.reference_type, t.reference_id,
		       t.created_at, t.created_by
		FROM inventory_transactions t`
	if filter.WarehouseID != uuid.Nil {
		query += ` JOIN inventory i ON t.inventory_id = i.id`
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY t.created_at DESC"

	argIdx := len(args) + 1
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
		return nil, fmt.Errorf("list transactions global: %w", err)
	}
	defer rows.Close()

	return r.scanTxRows(rows)
}

// CountTransactionsGlobal returns the total count of inventory transactions matching the filter.
func (r *InventoryRepo) CountTransactionsGlobal(ctx context.Context, filter repository.InventoryTxFilter) (int, error) {
	conditions, args := r.buildTxGlobalConditions(filter, nil)

	query := `SELECT COUNT(*) FROM inventory_transactions t`
	if filter.WarehouseID != uuid.Nil {
		query += ` JOIN inventory i ON t.inventory_id = i.id`
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.queryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count transactions global: %w", err)
	}
	return count, nil
}

// buildTxGlobalConditions builds WHERE clauses and argument list for
// inventory transaction global queries.
func (r *InventoryRepo) buildTxGlobalConditions(filter repository.InventoryTxFilter, _ *int) ([]string, []any) {
	var conditions []string
	var args []any
	idx := 1

	if filter.SKUID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("t.sku_id = $%d", idx))
		args = append(args, filter.SKUID)
		idx++
	}
	if filter.WarehouseID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("i.warehouse_id = $%d", idx))
		args = append(args, filter.WarehouseID)
		idx++
	}
	if filter.TxType != "" {
		conditions = append(conditions, fmt.Sprintf("t.type = $%d", idx))
		args = append(args, filter.TxType)
		idx++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("t.created_at >= $%d", idx))
		args = append(args, *filter.DateFrom)
		idx++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("t.created_at <= $%d", idx))
		args = append(args, *filter.DateTo)
		idx++
	}

	return conditions, args
}

// scanTxRows scans inventory transaction rows from an iterator.
func (r *InventoryRepo) scanTxRows(rows pgx.Rows) ([]*domain.InventoryTransaction, error) {
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

// ── Dashboard Queries ────────────────────────────────────────

// GetInventoryDashboardStats returns aggregate inventory statistics.
// When warehouseID is uuid.Nil, stats cover all warehouses.
func (r *InventoryRepo) GetInventoryDashboardStats(ctx context.Context, warehouseID uuid.UUID, lowStockThreshold float64) (*repository.InventoryDashboardStats, error) {
	const query = `
		SELECT
			COUNT(*) as total_records,
			COALESCE(SUM(qty), 0) as total_qty,
			COALESCE(SUM(reserved_qty), 0) as total_reserved_qty,
			COALESCE(SUM(qty - reserved_qty), 0) as total_available_qty,
			COUNT(*) FILTER (WHERE status = 'available') as available_count,
			COUNT(*) FILTER (WHERE status = 'quarantine') as quarantine_count,
			COUNT(*) FILTER (WHERE status = 'damaged') as damaged_count,
			COUNT(*) FILTER (WHERE status = 'expired') as expired_count,
			COUNT(*) FILTER (WHERE (qty - reserved_qty) > 0 AND (qty - reserved_qty) <= $2) as low_stock_count
		FROM inventory
		WHERE ($1::uuid = '00000000-0000-0000-0000-000000000000' OR warehouse_id = $1)`

	stats := &repository.InventoryDashboardStats{}
	err := r.queryRow(ctx, query, warehouseID, lowStockThreshold).Scan(
		&stats.TotalRecords,
		&stats.TotalQty,
		&stats.TotalReservedQty,
		&stats.TotalAvailableQty,
		&stats.AvailableCount,
		&stats.QuarantineCount,
		&stats.DamagedCount,
		&stats.ExpiredCount,
		&stats.LowStockCount,
	)
	if err != nil {
		return nil, fmt.Errorf("get inventory dashboard stats: %w", err)
	}
	return stats, nil
}

// GetLowStockInventory returns inventory records where available quantity is
// positive but at or below the given threshold. When warehouseID is uuid.Nil,
// all warehouses are included.
func (r *InventoryRepo) GetLowStockInventory(ctx context.Context, threshold float64, warehouseID uuid.UUID, limit int) ([]*domain.Inventory, error) {
	query := `
		SELECT id, sku_id, location_id, warehouse_id, batch_no,
		       qty, reserved_qty, status, production_date, expiry_date,
		       received_at, updated_at
		FROM inventory
		WHERE (qty - reserved_qty) > 0
		  AND (qty - reserved_qty) <= $1
		  AND ($2::uuid = '00000000-0000-0000-0000-000000000000' OR warehouse_id = $2)
		ORDER BY (qty - reserved_qty) ASC`

	args := []any{threshold, warehouseID}
	argIdx := 3

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
		argIdx++
	}

	rows, err := r.query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get low stock inventory: %w", err)
	}
	defer rows.Close()

	var results []*domain.Inventory
	for rows.Next() {
		inv, err := r.scanInventoryFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan low stock inventory: %w", err)
		}
		results = append(results, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate low stock inventory: %w", err)
	}
	return results, nil
}

// GetInventoryByWarehouse returns inventory aggregated by warehouse.
func (r *InventoryRepo) GetInventoryByWarehouse(ctx context.Context) ([]*repository.InventoryByWarehouseRow, error) {
	const query = `
		SELECT
			w.id, w.name, w.code,
			COALESCE(SUM(i.qty), 0),
			COALESCE(SUM(i.reserved_qty), 0),
			COALESCE(SUM(i.qty - i.reserved_qty), 0),
			COUNT(i.id)
		FROM warehouses w
		LEFT JOIN inventory i ON w.id = i.warehouse_id
		GROUP BY w.id, w.name, w.code
		ORDER BY w.name`

	rows, err := r.query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get inventory by warehouse: %w", err)
	}
	defer rows.Close()

	var results []*repository.InventoryByWarehouseRow
	for rows.Next() {
		row := &repository.InventoryByWarehouseRow{}
		if err := rows.Scan(
			&row.WarehouseID,
			&row.WarehouseName,
			&row.WarehouseCode,
			&row.TotalQty,
			&row.ReservedQty,
			&row.AvailableQty,
			&row.RecordCount,
		); err != nil {
			return nil, fmt.Errorf("scan inventory by warehouse row: %w", err)
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inventory by warehouse: %w", err)
	}
	return results, nil
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
