// Package repository defines data access interfaces for the WMS domain.
// Implementations live in subdirectories (e.g., postgres/).
package repository

import (
	"context"

	"time"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/google/uuid"
)

// WarehouseRepository manages warehouse, zone, and location persistence.
type WarehouseRepository interface {
	// Warehouse
	CreateWarehouse(ctx context.Context, w *domain.Warehouse) error
	GetWarehouse(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error)
	ListWarehouses(ctx context.Context, limit, offset int) ([]*domain.Warehouse, error)
	UpdateWarehouse(ctx context.Context, w *domain.Warehouse) error
	CountWarehouses(ctx context.Context) (int, error)

	// Zone
	CreateZone(ctx context.Context, z *domain.Zone) error
	GetZone(ctx context.Context, id uuid.UUID) (*domain.Zone, error)
	ListZonesByWarehouse(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]*domain.Zone, error)
	CountZonesByWarehouse(ctx context.Context, warehouseID uuid.UUID) (int, error)
	ListAllZones(ctx context.Context, filter ZoneFilter) ([]*domain.Zone, error)
	CountAllZones(ctx context.Context, filter ZoneFilter) (int, error)
	UpdateZone(ctx context.Context, z *domain.Zone) error

	// Location
	CreateLocation(ctx context.Context, l *domain.Location) error
	GetLocation(ctx context.Context, id uuid.UUID) (*domain.Location, error)
	GetLocationByBarcode(ctx context.Context, barcode string) (*domain.Location, error)
	ListLocationsByZone(ctx context.Context, zoneID uuid.UUID, limit, offset int) ([]*domain.Location, error)
	UpdateLocationStatus(ctx context.Context, id uuid.UUID, status domain.LocationStatus) error
	CountLocationsByZone(ctx context.Context, zoneID uuid.UUID) (int, error)
	ListAllLocations(ctx context.Context, filter LocationFilter) ([]*domain.Location, error)
	CountAllLocations(ctx context.Context, filter LocationFilter) (int, error)
	UpdateLocation(ctx context.Context, l *domain.Location) error
}

// ZoneFilter defines query parameters for global zone search.
type ZoneFilter struct {
	WarehouseID uuid.UUID
	Limit       int
	Offset      int
}

// LocationFilter defines query parameters for global location search.
type LocationFilter struct {
	ZoneID      uuid.UUID
	WarehouseID uuid.UUID
	Limit       int
	Offset      int
}

// InventoryRepository manages SKU, inventory, and transaction persistence.
type InventoryRepository interface {
	// SKU
	CreateSKU(ctx context.Context, s *domain.SKU) error
	GetSKU(ctx context.Context, id uuid.UUID) (*domain.SKU, error)
	GetSKUByBarcode(ctx context.Context, barcode string) (*domain.SKU, error)
	GetSKUByCode(ctx context.Context, code string) (*domain.SKU, error)
	ListSKUs(ctx context.Context, limit, offset int) ([]*domain.SKU, error)
	UpdateSKU(ctx context.Context, s *domain.SKU) error
	CountSKUs(ctx context.Context) (int, error)

	// Inventory
	CreateInventory(ctx context.Context, inv *domain.Inventory) error
	GetInventory(ctx context.Context, id uuid.UUID) (*domain.Inventory, error)
	GetAndLockInventory(ctx context.Context, id uuid.UUID) (*domain.Inventory, error)
	GetInventoryAtLocation(ctx context.Context, skuID, locationID uuid.UUID, batchNo string) (*domain.Inventory, error)
	QueryInventory(ctx context.Context, filter InventoryFilter) ([]*domain.Inventory, error)
	UpdateInventoryQty(ctx context.Context, id uuid.UUID, deltaQty, deltaReserved float64) error
	UpdateInventoryStatus(ctx context.Context, id uuid.UUID, status domain.InventoryStatus) error
	CountInventory(ctx context.Context, filter InventoryFilter) (int, error)

	// FEFO / FIFO retrieval strategies
	GetOldestInventory(ctx context.Context, filter InventoryRetrievalFilter) ([]*domain.Inventory, error)
	GetExpiringInventory(ctx context.Context, filter InventoryRetrievalFilter) ([]*domain.Inventory, error)

	// Inventory Transaction
	CreateTransaction(ctx context.Context, tx *domain.InventoryTransaction) error
	ListTransactions(ctx context.Context, inventoryID uuid.UUID, limit, offset int) ([]*domain.InventoryTransaction, error)
		ListTransactionsByReference(ctx context.Context, referenceType string, referenceID uuid.UUID) ([]*domain.InventoryTransaction, error)
	CountTransactions(ctx context.Context, inventoryID uuid.UUID) (int, error)

	// Dashboard queries
	GetInventoryDashboardStats(ctx context.Context, warehouseID uuid.UUID, lowStockThreshold float64) (*InventoryDashboardStats, error)
	GetLowStockInventory(ctx context.Context, threshold float64, warehouseID uuid.UUID, limit int) ([]*domain.Inventory, error)
	GetInventoryByWarehouse(ctx context.Context) ([]*InventoryByWarehouseRow, error)
}

// InventoryFilter defines query parameters for inventory search.
type InventoryFilter struct {
	WarehouseID uuid.UUID
	SKUID       uuid.UUID
	LocationID  uuid.UUID
	BatchNo     string
	Status      domain.InventoryStatus
	Limit       int
	Offset      int
}

// InventoryRetrievalFilter defines query parameters for FEFO / FIFO retrieval strategies.
// Unlike InventoryFilter (which is a general-purpose query), this filter is designed for
// picking decisions: it returns only available, non-zero inventory sorted by the
// appropriate strategy key (received_at for FIFO, expiry_date for FEFO).
type InventoryRetrievalFilter struct {
	WarehouseID uuid.UUID
	SKUID       uuid.UUID
	Limit       int
}

// InventoryDashboardStats holds aggregate inventory statistics for the dashboard.
type InventoryDashboardStats struct {
	TotalRecords      int     `json:"total_records"`
	TotalQty          float64 `json:"total_qty"`
	TotalReservedQty  float64 `json:"total_reserved_qty"`
	TotalAvailableQty float64 `json:"total_available_qty"`
	AvailableCount    int     `json:"available_count"`
	QuarantineCount   int     `json:"quarantine_count"`
	DamagedCount      int     `json:"damaged_count"`
	ExpiredCount      int     `json:"expired_count"`
	LowStockCount     int     `json:"low_stock_count"`
}

// InventoryByWarehouseRow holds inventory grouped by warehouse for the dashboard.
type InventoryByWarehouseRow struct {
	WarehouseID   uuid.UUID `json:"warehouse_id"`
	WarehouseName string    `json:"warehouse_name"`
	WarehouseCode string    `json:"warehouse_code"`
	TotalQty      float64   `json:"total_qty"`
	ReservedQty   float64   `json:"reserved_qty"`
	AvailableQty  float64   `json:"available_qty"`
	RecordCount   int       `json:"record_count"`
}

// OrderRepository manages order, order line, and ASN persistence.
type OrderRepository interface {
	// Order
	CreateOrder(ctx context.Context, o *domain.Order) error
	GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetOrderByNo(ctx context.Context, orderNo string) (*domain.Order, error)
	ListOrders(ctx context.Context, filter OrderFilter) ([]*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
	CountOrders(ctx context.Context, filter OrderFilter) (int, error)
	CountOrdersByStatus(ctx context.Context) (map[domain.OrderStatus]int, error)

	// OrderLine
	CreateOrderLine(ctx context.Context, line *domain.OrderLine) error
	GetOrderLine(ctx context.Context, id uuid.UUID) (*domain.OrderLine, error)
	GetOrderLines(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderLine, error)
	UpdateOrderLineStatus(ctx context.Context, id uuid.UUID, status domain.OrderLineStatus) error
	UpdateOrderLineFulfilledQty(ctx context.Context, id uuid.UUID, qty float64) error

	// ASN
	CreateASN(ctx context.Context, asn *domain.ASN) error
	GetASN(ctx context.Context, id uuid.UUID) (*domain.ASN, error)
	GetASNByNo(ctx context.Context, asnNo string) (*domain.ASN, error)
	ListASNs(ctx context.Context, filter ASNFilter) ([]*domain.ASN, error)
	UpdateASNStatus(ctx context.Context, id uuid.UUID, status domain.ASNStatus) error
	CountASNs(ctx context.Context, filter ASNFilter) (int, error)

	// ASNLine
	CreateASNLine(ctx context.Context, line *domain.ASNLine) error
	GetASNLine(ctx context.Context, id uuid.UUID) (*domain.ASNLine, error)
	GetASNLines(ctx context.Context, asnID uuid.UUID) ([]*domain.ASNLine, error)
	UpdateASNLineStatus(ctx context.Context, id uuid.UUID, status domain.ASNLineStatus) error
	UpdateASNLineReceivedQty(ctx context.Context, id uuid.UUID, qty float64) error
}

// ASNFilter defines query parameters for ASN search.
type ASNFilter struct {
	WarehouseID uuid.UUID
	ASNNo       string
	Status      domain.ASNStatus
	Limit       int
	Offset      int
}

// OrderFilter defines query parameters for order search.
type OrderFilter struct {
	WarehouseID uuid.UUID
	OrderType   domain.OrderType
	Status      domain.OrderStatus
	Limit       int
	Offset      int
}

// TaskRepository manages task and wave persistence.
type TaskRepository interface {
	// Task
	CreateTask(ctx context.Context, t *domain.Task) error
	GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	GetTasksByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.Task, error)
	ListTasks(ctx context.Context, filter TaskFilter) ([]*domain.Task, error)
	AssignTask(ctx context.Context, id uuid.UUID, assignedTo string) error
	UpdateTaskStatus(ctx context.Context, id uuid.UUID, status domain.TaskStatus) error
	CompleteTask(ctx context.Context, id uuid.UUID, actualQty float64, toLocationID *uuid.UUID) error
	CountTasks(ctx context.Context, filter TaskFilter) (int, error)
	CountTasksByStatus(ctx context.Context) (map[domain.TaskStatus]int, error)

	// Wave
	CreateWave(ctx context.Context, w *domain.Wave) error
	GetWave(ctx context.Context, id uuid.UUID) (*domain.Wave, error)
	ListWaves(ctx context.Context, filter WaveFilter) ([]*domain.Wave, error)
	UpdateWaveStatus(ctx context.Context, id uuid.UUID, status domain.WaveStatus) error
	AddWaveOrders(ctx context.Context, id uuid.UUID, orderIDs []uuid.UUID) error
	RemoveWaveOrders(ctx context.Context, id uuid.UUID, orderIDs []uuid.UUID) error
	CountWaves(ctx context.Context, filter WaveFilter) (int, error)
}

// TaskFilter defines query parameters for task search.
type TaskFilter struct {
	WarehouseID uuid.UUID
	TaskType    domain.TaskType
	Status      domain.TaskStatus
	AssignedTo  string
	Limit       int
	Offset      int
}

// WaveFilter defines query parameters for wave search.
type WaveFilter struct {
	WarehouseID uuid.UUID
	Status      domain.WaveStatus
	WaveType    domain.WaveType
	Limit       int
	Offset      int
}

// UserRepository manages user, role, and audit log persistence.
type UserRepository interface {
	// User
	CreateUser(ctx context.Context, u *domain.User) error
	GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	ListUsers(ctx context.Context, filter UserFilter) ([]*domain.User, error)
	UpdateUser(ctx context.Context, u *domain.User) error
	UpdateUserStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error
	CountUsers(ctx context.Context, filter UserFilter) (int, error)

	// Role
	CreateRole(ctx context.Context, r *domain.Role) error
	GetRole(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	ListRoles(ctx context.Context) ([]*domain.Role, error)
	UpdateRole(ctx context.Context, r *domain.Role) error
	DeleteRole(ctx context.Context, id uuid.UUID) error
	CountRoles(ctx context.Context) (int, error)

	// AuditLog
	CreateAuditLog(ctx context.Context, log *domain.AuditLog) error
	ListAuditLogs(ctx context.Context, filter AuditLogFilter) ([]*domain.AuditLog, error)
		CountAuditLogs(ctx context.Context, filter AuditLogFilter) (int, error)
}

// UserFilter defines query parameters for user search.
type UserFilter struct {
	Status domain.UserStatus
	Limit  int
	Offset int
}

// TokenBlacklistRepository manages revoked token persistence.
type TokenBlacklistRepository interface {
	// Add adds a JTI to the blacklist. Returns an error if the JTI already exists.
	Add(ctx context.Context, entry *domain.TokenBlacklistEntry) error

	// IsBlacklisted checks whether a JTI has been revoked.
	IsBlacklisted(ctx context.Context, jti string) (bool, error)

	// DeleteExpired removes entries whose expires_at has passed.
	DeleteExpired(ctx context.Context) (int64, error)
}

// AuditLogFilter defines query parameters for audit log search.
type AuditLogFilter struct {
	UserID   uuid.UUID
	Action   string
	Resource string
	Limit    int
	Offset   int
}

// MigrationRepository manages schema migration tracking.
// It tracks which SQL migration files have been applied to ensure
// each .sql file in the migrations/ directory runs exactly once.
type MigrationRepository interface {
	// GetApplied returns all migrations that have been applied, ordered by version.
	GetApplied(ctx context.Context) ([]*domain.SchemaMigration, error)

	// IsApplied checks whether a specific migration version has been applied.
	IsApplied(ctx context.Context, version string) (bool, error)

	// RecordApplied records that a migration has been successfully applied.
	RecordApplied(ctx context.Context, m *domain.SchemaMigration) error
}