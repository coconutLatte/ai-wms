// Package repository defines data access interfaces for the WMS domain.
// Implementations live in subdirectories (e.g., postgres/).
package repository

import (
	"context"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/google/uuid"
)

// WarehouseRepository manages warehouse, zone, and location persistence.
type WarehouseRepository interface {
	// Warehouse
	CreateWarehouse(ctx context.Context, w *domain.Warehouse) error
	GetWarehouse(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error)
	ListWarehouses(ctx context.Context) ([]*domain.Warehouse, error)
	UpdateWarehouse(ctx context.Context, w *domain.Warehouse) error

	// Zone
	CreateZone(ctx context.Context, z *domain.Zone) error
	GetZone(ctx context.Context, id uuid.UUID) (*domain.Zone, error)
	ListZonesByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]*domain.Zone, error)

	// Location
	CreateLocation(ctx context.Context, l *domain.Location) error
	GetLocation(ctx context.Context, id uuid.UUID) (*domain.Location, error)
	GetLocationByBarcode(ctx context.Context, barcode string) (*domain.Location, error)
	ListLocationsByZone(ctx context.Context, zoneID uuid.UUID) ([]*domain.Location, error)
	UpdateLocationStatus(ctx context.Context, id uuid.UUID, status domain.LocationStatus) error
}

// InventoryRepository manages SKU, inventory, and transaction persistence.
type InventoryRepository interface {
	// SKU
	CreateSKU(ctx context.Context, s *domain.SKU) error
	GetSKU(ctx context.Context, id uuid.UUID) (*domain.SKU, error)
	GetSKUByBarcode(ctx context.Context, barcode string) (*domain.SKU, error)
	GetSKUByCode(ctx context.Context, code string) (*domain.SKU, error)
	ListSKUs(ctx context.Context) ([]*domain.SKU, error)
	UpdateSKU(ctx context.Context, s *domain.SKU) error

	// Inventory
	CreateInventory(ctx context.Context, inv *domain.Inventory) error
	GetInventory(ctx context.Context, id uuid.UUID) (*domain.Inventory, error)
	GetInventoryAtLocation(ctx context.Context, skuID, locationID uuid.UUID, batchNo string) (*domain.Inventory, error)
	QueryInventory(ctx context.Context, filter InventoryFilter) ([]*domain.Inventory, error)
	UpdateInventoryQty(ctx context.Context, id uuid.UUID, deltaQty, deltaReserved float64) error

	// Inventory Transaction
	CreateTransaction(ctx context.Context, tx *domain.InventoryTransaction) error
	ListTransactions(ctx context.Context, inventoryID uuid.UUID) ([]*domain.InventoryTransaction, error)
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

// OrderRepository manages order, order line, and ASN persistence.
type OrderRepository interface {
	// Order
	CreateOrder(ctx context.Context, o *domain.Order) error
	GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetOrderByNo(ctx context.Context, orderNo string) (*domain.Order, error)
	ListOrders(ctx context.Context, filter OrderFilter) ([]*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error

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

	// Wave
	CreateWave(ctx context.Context, w *domain.Wave) error
	GetWave(ctx context.Context, id uuid.UUID) (*domain.Wave, error)
	ListWaves(ctx context.Context, warehouseID uuid.UUID) ([]*domain.Wave, error)
	UpdateWaveStatus(ctx context.Context, id uuid.UUID, status domain.WaveStatus) error
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
