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

	// Location
	CreateLocation(ctx context.Context, l *domain.Location) error
	GetLocation(ctx context.Context, id uuid.UUID) (*domain.Location, error)
	GetLocationByBarcode(ctx context.Context, barcode string) (*domain.Location, error)
	ListLocationsByZone(ctx context.Context, zoneID uuid.UUID, limit, offset int) ([]*domain.Location, error)
	UpdateLocationStatus(ctx context.Context, id uuid.UUID, status domain.LocationStatus) error
	CountLocationsByZone(ctx context.Context, zoneID uuid.UUID) (int, error)
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
	CountInventory(ctx context.Context, filter InventoryFilter) (int, error)

	// FEFO / FIFO retrieval strategies
	GetOldestInventory(ctx context.Context, filter InventoryRetrievalFilter) ([]*domain.Inventory, error)
	GetExpiringInventory(ctx context.Context, filter InventoryRetrievalFilter) ([]*domain.Inventory, error)

	// Inventory Transaction
	CreateTransaction(ctx context.Context, tx *domain.InventoryTransaction) error
	ListTransactions(ctx context.Context, inventoryID uuid.UUID, limit, offset int) ([]*domain.InventoryTransaction, error)
	CountTransactions(ctx context.Context, inventoryID uuid.UUID) (int, error)
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

// OrderRepository manages order, order line, and ASN persistence.
type OrderRepository interface {
	// Order
	CreateOrder(ctx context.Context, o *domain.Order) error
	GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetOrderByNo(ctx context.Context, orderNo string) (*domain.Order, error)
	ListOrders(ctx context.Context, filter OrderFilter) ([]*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
	CountOrders(ctx context.Context, filter OrderFilter) (int, error)

	// OrderLine
	CreateOrderLine(ctx context.Context, line *domain.OrderLine) error
	GetOrderLines(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderLine, error)
	UpdateOrderLineStatus(ctx context.Context, id uuid.UUID, status domain.OrderLineStatus) error
	UpdateOrderLineFulfilledQty(ctx context.Context, id uuid.UUID, qty float64) error

	// ASN
	CreateASN(ctx context.Context, asn *domain.ASN) error
	GetASN(ctx context.Context, id uuid.UUID) (*domain.ASN, error)
	GetASNByNo(ctx context.Context, asnNo string) (*domain.ASN, error)
	UpdateASNStatus(ctx context.Context, id uuid.UUID, status domain.ASNStatus) error

	// ASNLine
	CreateASNLine(ctx context.Context, line *domain.ASNLine) error
	GetASNLines(ctx context.Context, asnID uuid.UUID) ([]*domain.ASNLine, error)
	UpdateASNLineStatus(ctx context.Context, id uuid.UUID, status domain.ASNLineStatus) error
	UpdateASNLineReceivedQty(ctx context.Context, id uuid.UUID, qty float64) error
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
	ListTasks(ctx context.Context, filter TaskFilter) ([]*domain.Task, error)
	AssignTask(ctx context.Context, id uuid.UUID, assignedTo string) error
	UpdateTaskStatus(ctx context.Context, id uuid.UUID, status domain.TaskStatus) error
	CompleteTask(ctx context.Context, id uuid.UUID, actualQty float64, toLocationID *uuid.UUID) error
	CountTasks(ctx context.Context, filter TaskFilter) (int, error)

	// Wave
	CreateWave(ctx context.Context, w *domain.Wave) error
	GetWave(ctx context.Context, id uuid.UUID) (*domain.Wave, error)
	ListWaves(ctx context.Context, warehouseID uuid.UUID) ([]*domain.Wave, error)
	UpdateWaveStatus(ctx context.Context, id uuid.UUID, status domain.WaveStatus) error
	CountWaves(ctx context.Context, warehouseID uuid.UUID) (int, error)
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

// AuditLogFilter defines query parameters for audit log search.
type AuditLogFilter struct {
	UserID   uuid.UUID
	Action   string
	Resource string
	Limit    int
	Offset   int
}