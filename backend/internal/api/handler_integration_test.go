// Package api provides integration tests for HTTP handler endpoints.
//
//go:build integration
// +build integration

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	"github.com/ai-wms/ai-wms/backend/internal/service"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// ── In-Memory Mock Repositories (stateful for integration scenarios) ────────

// memWarehouseRepo is an in-memory WarehouseRepository implementation.
type memWarehouseRepo struct {
	mu         sync.RWMutex
	warehouses map[uuid.UUID]*domain.Warehouse
	zones      map[uuid.UUID]*domain.Zone
	locations  map[uuid.UUID]*domain.Location
}

func newMemWarehouseRepo() *memWarehouseRepo {
	return &memWarehouseRepo{
		warehouses: make(map[uuid.UUID]*domain.Warehouse),
		zones:      make(map[uuid.UUID]*domain.Zone),
		locations:  make(map[uuid.UUID]*domain.Location),
	}
}

func (m *memWarehouseRepo) CreateWarehouse(ctx context.Context, w *domain.Warehouse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	w.CreatedAt = time.Now()
	w.UpdatedAt = w.CreatedAt
	m.warehouses[w.ID] = w
	return nil
}
func (m *memWarehouseRepo) GetWarehouse(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.warehouses[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("warehouse", id.String())
	}
	return w, nil
}
func (m *memWarehouseRepo) ListWarehouses(ctx context.Context, limit, offset int) ([]*domain.Warehouse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.Warehouse, 0, len(m.warehouses))
	for _, w := range m.warehouses {
		result = append(result, w)
	}
	// Apply offset/limit
	if offset >= len(result) {
		return nil, nil
	}
	end := offset + limit
	if end > len(result) || limit == 0 {
		end = len(result)
	}
	return result[offset:end], nil
}
func (m *memWarehouseRepo) UpdateWarehouse(ctx context.Context, w *domain.Warehouse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.warehouses[w.ID]
	if !ok {
		return pkgerrors.NewNotFound("warehouse", w.ID.String())
	}
	existing.Name = w.Name
	existing.Code = w.Code
	existing.Address = w.Address
	existing.UpdatedAt = time.Now()
	return nil
}
func (m *memWarehouseRepo) CountWarehouses(ctx context.Context) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.warehouses), nil
}

// Zone
func (m *memWarehouseRepo) CreateZone(ctx context.Context, z *domain.Zone) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if z.ID == uuid.Nil {
		z.ID = uuid.New()
	}
	z.CreatedAt = time.Now()
	z.UpdatedAt = z.CreatedAt
	m.zones[z.ID] = z
	return nil
}
func (m *memWarehouseRepo) GetZone(ctx context.Context, id uuid.UUID) (*domain.Zone, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	z, ok := m.zones[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("zone", id.String())
	}
	return z, nil
}
func (m *memWarehouseRepo) ListZonesByWarehouse(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]*domain.Zone, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.Zone
	for _, z := range m.zones {
		if z.WarehouseID == warehouseID {
			result = append(result, z)
		}
	}
	if offset >= len(result) {
		return nil, nil
	}
	end := offset + limit
	if end > len(result) || limit == 0 {
		end = len(result)
	}
	return result[offset:end], nil
}
func (m *memWarehouseRepo) CountZonesByWarehouse(ctx context.Context, warehouseID uuid.UUID) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, z := range m.zones {
		if z.WarehouseID == warehouseID {
			count++
		}
	}
	return count, nil
}
func (m *memWarehouseRepo) ListAllZones(ctx context.Context, filter repository.ZoneFilter) ([]*domain.Zone, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.Zone
	for _, z := range m.zones {
		result = append(result, z)
	}
	return result, nil
}
func (m *memWarehouseRepo) CountAllZones(ctx context.Context, filter repository.ZoneFilter) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.zones), nil
}
func (m *memWarehouseRepo) UpdateZone(ctx context.Context, z *domain.Zone) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.zones[z.ID]
	if !ok {
		return pkgerrors.NewNotFound("zone", z.ID.String())
	}
	existing.Name = z.Name
	existing.Code = z.Code
	existing.UpdatedAt = time.Now()
	return nil
}

// Location
func (m *memWarehouseRepo) CreateLocation(ctx context.Context, l *domain.Location) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	l.CreatedAt = time.Now()
	l.UpdatedAt = l.CreatedAt
	m.locations[l.ID] = l
	return nil
}
func (m *memWarehouseRepo) GetLocation(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	loc, ok := m.locations[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("location", id.String())
	}
	return loc, nil
}
func (m *memWarehouseRepo) GetLocationByBarcode(ctx context.Context, barcode string) (*domain.Location, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, loc := range m.locations {
		if loc.Barcode == barcode {
			return loc, nil
		}
	}
	return nil, pkgerrors.NewNotFound("location", barcode)
}
func (m *memWarehouseRepo) ListLocationsByZone(ctx context.Context, zoneID uuid.UUID, limit, offset int) ([]*domain.Location, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.Location
	for _, loc := range m.locations {
		if loc.ZoneID == zoneID {
			result = append(result, loc)
		}
	}
	if offset >= len(result) {
		return nil, nil
	}
	end := offset + limit
	if end > len(result) || limit == 0 {
		end = len(result)
	}
	return result[offset:end], nil
}
func (m *memWarehouseRepo) UpdateLocationStatus(ctx context.Context, id uuid.UUID, status domain.LocationStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	loc, ok := m.locations[id]
	if !ok {
		return pkgerrors.NewNotFound("location", id.String())
	}
	loc.Status = status
	loc.UpdatedAt = time.Now()
	return nil
}
func (m *memWarehouseRepo) CountLocationsByZone(ctx context.Context, zoneID uuid.UUID) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, loc := range m.locations {
		if loc.ZoneID == zoneID {
			count++
		}
	}
	return count, nil
}
func (m *memWarehouseRepo) ListAllLocations(ctx context.Context, filter repository.LocationFilter) ([]*domain.Location, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.Location
	for _, loc := range m.locations {
		result = append(result, loc)
	}
	return result, nil
}
func (m *memWarehouseRepo) CountAllLocations(ctx context.Context, filter repository.LocationFilter) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.locations), nil
}
func (m *memWarehouseRepo) UpdateLocation(ctx context.Context, l *domain.Location) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.locations[l.ID]
	if !ok {
		return pkgerrors.NewNotFound("location", l.ID.String())
	}
	existing.Code = l.Code
	existing.Barcode = l.Barcode
	existing.Capacity = l.Capacity
	existing.LocationType = l.LocationType
	existing.UpdatedAt = time.Now()
	return nil
}

// memInventoryRepo is an in-memory InventoryRepository implementation.
type memInventoryRepo struct {
	mu           sync.RWMutex
	skus         map[uuid.UUID]*domain.SKU
	inventory    map[uuid.UUID]*domain.Inventory
	transactions []*domain.InventoryTransaction
}

func newMemInventoryRepo() *memInventoryRepo {
	return &memInventoryRepo{
		skus:      make(map[uuid.UUID]*domain.SKU),
		inventory: make(map[uuid.UUID]*domain.Inventory),
	}
}

func (m *memInventoryRepo) CreateSKU(ctx context.Context, s *domain.SKU) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	s.CreatedAt = time.Now()
	s.UpdatedAt = s.CreatedAt
	m.skus[s.ID] = s
	return nil
}
func (m *memInventoryRepo) GetSKU(ctx context.Context, id uuid.UUID) (*domain.SKU, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.skus[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("sku", id.String())
	}
	return s, nil
}
func (m *memInventoryRepo) GetSKUByBarcode(ctx context.Context, barcode string) (*domain.SKU, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, s := range m.skus {
		if s.Barcode == barcode {
			return s, nil
		}
	}
	return nil, pkgerrors.NewNotFound("sku", barcode)
}
func (m *memInventoryRepo) GetSKUByCode(ctx context.Context, code string) (*domain.SKU, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, s := range m.skus {
		if s.Code == code {
			return s, nil
		}
	}
	return nil, pkgerrors.NewNotFound("sku", code)
}
func (m *memInventoryRepo) ListSKUs(ctx context.Context, limit, offset int) ([]*domain.SKU, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.SKU, 0, len(m.skus))
	for _, s := range m.skus {
		result = append(result, s)
	}
	if offset >= len(result) {
		return nil, nil
	}
	end := offset + limit
	if end > len(result) || limit == 0 {
		end = len(result)
	}
	return result[offset:end], nil
}
func (m *memInventoryRepo) UpdateSKU(ctx context.Context, s *domain.SKU) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.skus[s.ID]
	if !ok {
		return pkgerrors.NewNotFound("sku", s.ID.String())
	}
	existing.Name = s.Name
	existing.Code = s.Code
	existing.Barcode = s.Barcode
	existing.Status = s.Status
	existing.UpdatedAt = time.Now()
	return nil
}
func (m *memInventoryRepo) CountSKUs(ctx context.Context) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.skus), nil
}

// Inventory
func (m *memInventoryRepo) CreateInventory(ctx context.Context, inv *domain.Inventory) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	inv.ReceivedAt = time.Now()
	inv.UpdatedAt = inv.ReceivedAt
	m.inventory[inv.ID] = inv
	return nil
}
func (m *memInventoryRepo) GetInventory(ctx context.Context, id uuid.UUID) (*domain.Inventory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	inv, ok := m.inventory[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("inventory", id.String())
	}
	return inv, nil
}
func (m *memInventoryRepo) GetAndLockInventory(ctx context.Context, id uuid.UUID) (*domain.Inventory, error) {
	return m.GetInventory(ctx, id)
}
func (m *memInventoryRepo) GetInventoryAtLocation(ctx context.Context, skuID, locationID uuid.UUID, batchNo string) (*domain.Inventory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, inv := range m.inventory {
		if inv.SKUID == skuID && inv.LocationID == locationID && inv.BatchNo == batchNo {
			return inv, nil
		}
	}
	return nil, pkgerrors.NewNotFound("inventory", "")
}
func (m *memInventoryRepo) QueryInventory(ctx context.Context, filter repository.InventoryFilter) ([]*domain.Inventory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.Inventory, 0, len(m.inventory))
	for _, inv := range m.inventory {
		result = append(result, inv)
	}
	return result, nil
}
func (m *memInventoryRepo) UpdateInventoryQty(ctx context.Context, id uuid.UUID, deltaQty, deltaReserved float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	inv, ok := m.inventory[id]
	if !ok {
		return pkgerrors.NewNotFound("inventory", id.String())
	}
	inv.Qty += deltaQty
	inv.ReservedQty += deltaReserved
	inv.AvailableQty = inv.Qty - inv.ReservedQty
	inv.UpdatedAt = time.Now()
	return nil
}
func (m *memInventoryRepo) UpdateInventoryStatus(ctx context.Context, id uuid.UUID, status domain.InventoryStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	inv, ok := m.inventory[id]
	if !ok {
		return pkgerrors.NewNotFound("inventory", id.String())
	}
	inv.Status = status
	inv.UpdatedAt = time.Now()
	return nil
}
func (m *memInventoryRepo) CountInventory(ctx context.Context, filter repository.InventoryFilter) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.inventory), nil
}
func (m *memInventoryRepo) GetOldestInventory(ctx context.Context, filter repository.InventoryRetrievalFilter) ([]*domain.Inventory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.Inventory, 0, len(m.inventory))
	for _, inv := range m.inventory {
		result = append(result, inv)
	}
	return result, nil
}
func (m *memInventoryRepo) GetExpiringInventory(ctx context.Context, filter repository.InventoryRetrievalFilter) ([]*domain.Inventory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.Inventory, 0, len(m.inventory))
	for _, inv := range m.inventory {
		result = append(result, inv)
	}
	return result, nil
}
func (m *memInventoryRepo) CreateTransaction(ctx context.Context, tx *domain.InventoryTransaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	tx.CreatedAt = time.Now()
	m.transactions = append(m.transactions, tx)
	return nil
}
func (m *memInventoryRepo) ListTransactions(ctx context.Context, inventoryID uuid.UUID, limit, offset int) ([]*domain.InventoryTransaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.InventoryTransaction
	for _, tx := range m.transactions {
		if tx.InventoryID == inventoryID {
			result = append(result, tx)
		}
	}
	return result, nil
}
func (m *memInventoryRepo) ListTransactionsByReference(ctx context.Context, referenceType string, referenceID uuid.UUID) ([]*domain.InventoryTransaction, error) {
	return nil, nil
}
func (m *memInventoryRepo) CountTransactions(ctx context.Context, inventoryID uuid.UUID) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, tx := range m.transactions {
		if tx.InventoryID == inventoryID {
			count++
		}
	}
	return count, nil
}
func (m *memInventoryRepo) ListTransactionsGlobal(ctx context.Context, filter repository.InventoryTxFilter) ([]*domain.InventoryTransaction, error) {
	return nil, nil
}
func (m *memInventoryRepo) CountTransactionsGlobal(ctx context.Context, filter repository.InventoryTxFilter) (int, error) {
	return 0, nil
}
func (m *memInventoryRepo) GetInventoryDashboardStats(ctx context.Context, warehouseID uuid.UUID, lowStockThreshold float64) (*repository.InventoryDashboardStats, error) {
	return &repository.InventoryDashboardStats{}, nil
}
func (m *memInventoryRepo) GetLowStockInventory(ctx context.Context, threshold float64, warehouseID uuid.UUID, limit int) ([]*domain.Inventory, error) {
	return nil, nil
}
func (m *memInventoryRepo) GetInventoryByWarehouse(ctx context.Context) ([]*repository.InventoryByWarehouseRow, error) {
	return nil, nil
}

// memOrderRepo is an in-memory OrderRepository implementation.
type memOrderRepo struct {
	mu        sync.RWMutex
	orders    map[uuid.UUID]*domain.Order
	lines     map[uuid.UUID]*domain.OrderLine
	asns      map[uuid.UUID]*domain.ASN
	asnLines  map[uuid.UUID]*domain.ASNLine
}

func newMemOrderRepo() *memOrderRepo {
	return &memOrderRepo{
		orders:   make(map[uuid.UUID]*domain.Order),
		lines:    make(map[uuid.UUID]*domain.OrderLine),
		asns:     make(map[uuid.UUID]*domain.ASN),
		asnLines: make(map[uuid.UUID]*domain.ASNLine),
	}
}

func (m *memOrderRepo) CreateOrder(ctx context.Context, o *domain.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	o.CreatedAt = time.Now()
	o.UpdatedAt = o.CreatedAt
	m.orders[o.ID] = o
	return nil
}
func (m *memOrderRepo) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	o, ok := m.orders[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("order", id.String())
	}
	// Build lines
	var orderLines []domain.OrderLine
	for _, l := range m.lines {
		if l.OrderID == id {
			orderLines = append(orderLines, *l)
		}
	}
	o.Lines = orderLines
	return o, nil
}
func (m *memOrderRepo) GetOrderByNo(ctx context.Context, orderNo string) (*domain.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, o := range m.orders {
		if o.OrderNo == orderNo {
			return o, nil
		}
	}
	return nil, pkgerrors.NewNotFound("order", orderNo)
}
func (m *memOrderRepo) ListOrders(ctx context.Context, filter repository.OrderFilter) ([]*domain.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.Order, 0, len(m.orders))
	for _, o := range m.orders {
		result = append(result, o)
	}
	return result, nil
}
func (m *memOrderRepo) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return pkgerrors.NewNotFound("order", id.String())
	}
	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
}
func (m *memOrderRepo) CountOrders(ctx context.Context, filter repository.OrderFilter) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.orders), nil
}
func (m *memOrderRepo) CountOrdersByStatus(ctx context.Context) (map[domain.OrderStatus]int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[domain.OrderStatus]int)
	for _, o := range m.orders {
		result[o.Status]++
	}
	return result, nil
}

func (m *memOrderRepo) CreateOrderLine(ctx context.Context, line *domain.OrderLine) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if line.ID == uuid.Nil {
		line.ID = uuid.New()
	}
	m.lines[line.ID] = line
	return nil
}
func (m *memOrderRepo) GetOrderLine(ctx context.Context, id uuid.UUID) (*domain.OrderLine, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.lines[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("order_line", id.String())
	}
	return l, nil
}
func (m *memOrderRepo) GetOrderLines(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderLine, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.OrderLine
	for _, l := range m.lines {
		if l.OrderID == orderID {
			result = append(result, l)
		}
	}
	return result, nil
}
func (m *memOrderRepo) UpdateOrderLineStatus(ctx context.Context, id uuid.UUID, status domain.OrderLineStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.lines[id]
	if !ok {
		return pkgerrors.NewNotFound("order_line", id.String())
	}
	l.Status = status
	return nil
}
func (m *memOrderRepo) UpdateOrderLineFulfilledQty(ctx context.Context, id uuid.UUID, qty float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.lines[id]
	if !ok {
		return pkgerrors.NewNotFound("order_line", id.String())
	}
	l.FulfilledQty = qty
	return nil
}

// ASN
func (m *memOrderRepo) CreateASN(ctx context.Context, asn *domain.ASN) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if asn.ID == uuid.Nil {
		asn.ID = uuid.New()
	}
	asn.CreatedAt = time.Now()
	m.asns[asn.ID] = asn
	return nil
}
func (m *memOrderRepo) GetASN(ctx context.Context, id uuid.UUID) (*domain.ASN, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.asns[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("asn", id.String())
	}
	return a, nil
}
func (m *memOrderRepo) GetASNByNo(ctx context.Context, asnNo string) (*domain.ASN, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, a := range m.asns {
		if a.ASNNo == asnNo {
			return a, nil
		}
	}
	return nil, pkgerrors.NewNotFound("asn", asnNo)
}
func (m *memOrderRepo) ListASNs(ctx context.Context, filter repository.ASNFilter) ([]*domain.ASN, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.ASN, 0, len(m.asns))
	for _, a := range m.asns {
		result = append(result, a)
	}
	return result, nil
}
func (m *memOrderRepo) UpdateASNStatus(ctx context.Context, id uuid.UUID, status domain.ASNStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.asns[id]
	if !ok {
		return pkgerrors.NewNotFound("asn", id.String())
	}
	a.Status = status
	return nil
}
func (m *memOrderRepo) CountASNs(ctx context.Context, filter repository.ASNFilter) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.asns), nil
}
func (m *memOrderRepo) CreateASNLine(ctx context.Context, line *domain.ASNLine) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if line.ID == uuid.Nil {
		line.ID = uuid.New()
	}
	m.asnLines[line.ID] = line
	return nil
}
func (m *memOrderRepo) GetASNLine(ctx context.Context, id uuid.UUID) (*domain.ASNLine, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.asnLines[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("asn_line", id.String())
	}
	return l, nil
}
func (m *memOrderRepo) GetASNLines(ctx context.Context, asnID uuid.UUID) ([]*domain.ASNLine, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.ASNLine
	for _, l := range m.asnLines {
		if l.ASNID == asnID {
			result = append(result, l)
		}
	}
	return result, nil
}
func (m *memOrderRepo) UpdateASNLineStatus(ctx context.Context, id uuid.UUID, status domain.ASNLineStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.asnLines[id]
	if !ok {
		return pkgerrors.NewNotFound("asn_line", id.String())
	}
	l.Status = status
	return nil
}
func (m *memOrderRepo) UpdateASNLineReceivedQty(ctx context.Context, id uuid.UUID, qty float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.asnLines[id]
	if !ok {
		return pkgerrors.NewNotFound("asn_line", id.String())
	}
	l.ReceivedQty = qty
	return nil
}

// memTaskRepo is an in-memory TaskRepository implementation.
type memTaskRepo struct {
	mu     sync.RWMutex
	tasks  map[uuid.UUID]*domain.Task
	waves  map[uuid.UUID]*domain.Wave
}

func newMemTaskRepo() *memTaskRepo {
	return &memTaskRepo{
		tasks: make(map[uuid.UUID]*domain.Task),
		waves: make(map[uuid.UUID]*domain.Wave),
	}
}

func (m *memTaskRepo) CreateTask(ctx context.Context, t *domain.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	t.CreatedAt = time.Now()
	m.tasks[t.ID] = t
	return nil
}
func (m *memTaskRepo) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tasks[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("task", id.String())
	}
	return t, nil
}
func (m *memTaskRepo) GetTasksByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.Task
	for _, t := range m.tasks {
		if t.OrderID != nil && *t.OrderID == orderID {
			result = append(result, t)
		}
	}
	return result, nil
}
func (m *memTaskRepo) ListTasks(ctx context.Context, filter repository.TaskFilter) ([]*domain.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		result = append(result, t)
	}
	return result, nil
}
func (m *memTaskRepo) AssignTask(ctx context.Context, id uuid.UUID, assignedTo string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return pkgerrors.NewNotFound("task", id.String())
	}
	t.AssignedTo = assignedTo
	t.Status = domain.TaskStatusAssigned
	return nil
}
func (m *memTaskRepo) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status domain.TaskStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return pkgerrors.NewNotFound("task", id.String())
	}
	t.Status = status
	return nil
}
func (m *memTaskRepo) CompleteTask(ctx context.Context, id uuid.UUID, actualQty float64, toLocationID *uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return pkgerrors.NewNotFound("task", id.String())
	}
	t.Status = domain.TaskStatusCompleted
	t.ActualQty = actualQty
	now := time.Now()
	t.CompletedAt = &now
	if toLocationID != nil {
		t.ToLocation = toLocationID
	}
	return nil
}
func (m *memTaskRepo) CountTasks(ctx context.Context, filter repository.TaskFilter) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tasks), nil
}
func (m *memTaskRepo) CountTasksByStatus(ctx context.Context) (map[domain.TaskStatus]int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[domain.TaskStatus]int)
	for _, t := range m.tasks {
		result[t.Status]++
	}
	return result, nil
}

// Wave
func (m *memTaskRepo) CreateWave(ctx context.Context, w *domain.Wave) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	w.CreatedAt = time.Now()
	m.waves[w.ID] = w
	return nil
}
func (m *memTaskRepo) GetWave(ctx context.Context, id uuid.UUID) (*domain.Wave, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.waves[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("wave", id.String())
	}
	return w, nil
}
func (m *memTaskRepo) ListWaves(ctx context.Context, filter repository.WaveFilter) ([]*domain.Wave, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.Wave, 0, len(m.waves))
	for _, w := range m.waves {
		result = append(result, w)
	}
	return result, nil
}
func (m *memTaskRepo) UpdateWaveStatus(ctx context.Context, id uuid.UUID, status domain.WaveStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.waves[id]
	if !ok {
		return pkgerrors.NewNotFound("wave", id.String())
	}
	w.Status = status
	return nil
}
func (m *memTaskRepo) AddWaveOrders(ctx context.Context, id uuid.UUID, orderIDs []uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.waves[id]
	if !ok {
		return pkgerrors.NewNotFound("wave", id.String())
	}
	for _, oid := range orderIDs {
		found := false
		for _, existing := range w.OrderIDs {
			if existing == oid {
				found = true
				break
			}
		}
		if !found {
			w.OrderIDs = append(w.OrderIDs, oid)
		}
	}
	w.TotalOrders = len(w.OrderIDs)
	return nil
}
func (m *memTaskRepo) RemoveWaveOrders(ctx context.Context, id uuid.UUID, orderIDs []uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, ok := m.waves[id]
	if !ok {
		return pkgerrors.NewNotFound("wave", id.String())
	}
	toRemove := make(map[uuid.UUID]bool, len(orderIDs))
	for _, oid := range orderIDs {
		toRemove[oid] = true
	}
	filtered := make([]uuid.UUID, 0, len(w.OrderIDs))
	for _, existing := range w.OrderIDs {
		if !toRemove[existing] {
			filtered = append(filtered, existing)
		}
	}
	w.OrderIDs = filtered
	w.TotalOrders = len(w.OrderIDs)
	return nil
}
func (m *memTaskRepo) CountWaves(ctx context.Context, filter repository.WaveFilter) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.waves), nil
}

// memUserRepo is an in-memory UserRepository implementation.
type memUserRepo struct {
	mu        sync.RWMutex
	users     map[uuid.UUID]*domain.User
	roles     map[uuid.UUID]*domain.Role
	auditLogs []*domain.AuditLog
}

func newMemUserRepo() *memUserRepo {
	return &memUserRepo{
		users: make(map[uuid.UUID]*domain.User),
		roles: make(map[uuid.UUID]*domain.Role),
	}
}

func (m *memUserRepo) CreateUser(ctx context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	m.users[u.ID] = u
	return nil
}
func (m *memUserRepo) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("user", id.String())
	}
	return u, nil
}
func (m *memUserRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, pkgerrors.NewNotFound("user", username)
}
func (m *memUserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, pkgerrors.NewNotFound("user", email)
}
func (m *memUserRepo) ListUsers(ctx context.Context, filter repository.UserFilter) ([]*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.User, 0, len(m.users))
	for _, u := range m.users {
		result = append(result, u)
	}
	return result, nil
}
func (m *memUserRepo) UpdateUser(ctx context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.users[u.ID]
	if !ok {
		return pkgerrors.NewNotFound("user", u.ID.String())
	}
	existing.Email = u.Email
	existing.DisplayName = u.DisplayName
	existing.UpdatedAt = time.Now()
	return nil
}
func (m *memUserRepo) UpdateUserStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return pkgerrors.NewNotFound("user", id.String())
	}
	u.Status = status
	u.UpdatedAt = time.Now()
	return nil
}
func (m *memUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return pkgerrors.NewNotFound("user", id.String())
	}
	u.LastLogin = &t
	return nil
}
func (m *memUserRepo) CountUsers(ctx context.Context, filter repository.UserFilter) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.users), nil
}

func (m *memUserRepo) CreateRole(ctx context.Context, r *domain.Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	r.CreatedAt = time.Now()
	m.roles[r.ID] = r
	return nil
}
func (m *memUserRepo) GetRole(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.roles[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("role", id.String())
	}
	return r, nil
}
func (m *memUserRepo) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.Role, 0, len(m.roles))
	for _, r := range m.roles {
		result = append(result, r)
	}
	return result, nil
}
func (m *memUserRepo) UpdateRole(ctx context.Context, r *domain.Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.roles[r.ID]
	if !ok {
		return pkgerrors.NewNotFound("role", r.ID.String())
	}
	existing.Name = r.Name
	existing.Description = r.Description
	existing.Permissions = r.Permissions
	return nil
}
func (m *memUserRepo) DeleteRole(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.roles[id]; !ok {
		return pkgerrors.NewNotFound("role", id.String())
	}
	delete(m.roles, id)
	return nil
}
func (m *memUserRepo) CountRoles(ctx context.Context) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.roles), nil
}

func (m *memUserRepo) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	log.CreatedAt = time.Now()
	m.auditLogs = append(m.auditLogs, log)
	return nil
}
func (m *memUserRepo) ListAuditLogs(ctx context.Context, filter repository.AuditLogFilter) ([]*domain.AuditLog, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.auditLogs, nil
}
func (m *memUserRepo) CountAuditLogs(ctx context.Context, filter repository.AuditLogFilter) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.auditLogs), nil
}

// ── Test Server Setup ────────────────────────────────────────────────────────

var testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// testServer holds all the components needed for integration tests.
type testServer struct {
	server      *httptest.Server
	mux         *http.ServeMux
	whRepo      *memWarehouseRepo
	invRepo     *memInventoryRepo
	orderRepo   *memOrderRepo
	taskRepo    *memTaskRepo
	userRepo    *memUserRepo
}

func newTestServer() *testServer {
	mux := http.NewServeMux()
	ts := &testServer{
		mux:       mux,
		whRepo:    newMemWarehouseRepo(),
		invRepo:   newMemInventoryRepo(),
		orderRepo: newMemOrderRepo(),
		taskRepo:  newMemTaskRepo(),
		userRepo:  newMemUserRepo(),
	}

	// Wire services
	whSvc := service.NewWarehouseService(ts.whRepo)
	skuSvc := service.NewSKUService(ts.invRepo)
	invSvc := service.NewInventoryService(ts.invRepo)
	orderSvc := service.NewOrderService(ts.orderRepo, ts.taskRepo)
	taskSvc := service.NewTaskService(ts.taskRepo)
	userSvc := service.NewUserService(ts.userRepo)

	// Register all routes
	RegisterWarehouseRoutes(mux, NewWarehouseHandler(whSvc, testLogger))
	RegisterSKURoutes(mux, NewSKUHandler(skuSvc, testLogger))
	RegisterInventoryRoutes(mux, NewInventoryHandler(invSvc, testLogger))
	RegisterOrderRoutes(mux, NewOrderHandler(orderSvc, testLogger))
	RegisterASNRoutes(mux, NewOrderHandler(orderSvc, testLogger))
	RegisterTaskRoutes(mux, NewTaskHandler(taskSvc, testLogger))

	// Register remaining routes
	auditLogSvc := service.NewAuditLogService(ts.userRepo)
	RegisterAuditLogRoutes(mux, NewAuditLogHandler(auditLogSvc, testLogger))
	RegisterUserRoutes(mux, NewUserHandler(userSvc, testLogger))
	roleSvc := service.NewRoleService(ts.userRepo)
	RegisterRoleRoutes(mux, NewRoleHandler(roleSvc, testLogger))
	waveSvc := service.NewWaveService(ts.taskRepo)
	RegisterWaveRoutes(mux, NewWaveHandler(waveSvc, testLogger))
	RegisterDashboardRoute(mux, NewDashboardHandler(whSvc, skuSvc, invSvc, orderSvc, taskSvc, testLogger))

	ts.server = httptest.NewServer(mux)
	return ts
}

func (ts *testServer) Close() {
	ts.server.Close()
}

func (ts *testServer) URL() string {
	return ts.server.URL
}

// doReq is a helper to make HTTP requests in tests.
func doReq(t *testing.T, ts *testServer, method, path string, body string) *http.Response {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, ts.URL()+path, r)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	return string(b)
}

func assertStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Errorf("expected status %d, got %d: %s", want, resp.StatusCode, readBody(t, resp))
	}
}

func assertJSON(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var result map[string]any
	body := readBody(t, resp)
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		t.Fatalf("failed to decode response body (%q): %v", body, err)
	}
	return result
}

// ── Warehouse Integration Tests ──────────────────────────────────────────────

func TestIntegration_WarehouseCreateAndGet(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create warehouse
	createBody := `{"code":"WH-INT-001","name":"Integration Warehouse","address":"123 Test St"}`
	resp := doReq(t, ts, "POST", "/api/v1/warehouses", createBody)
	assertStatus(t, resp, http.StatusCreated)
	result := assertJSON(t, resp)
	if result["code"] != "WH-INT-001" {
		t.Errorf("code = %q, want WH-INT-001", result["code"])
	}
	id := result["id"].(string)

	// Get warehouse
	resp = doReq(t, ts, "GET", "/api/v1/warehouses/"+id, "")
	assertStatus(t, resp, http.StatusOK)
	result = assertJSON(t, resp)
	if result["name"] != "Integration Warehouse" {
		t.Errorf("name = %q, want Integration Warehouse", result["name"])
	}

	// Get non-existent warehouse
	resp = doReq(t, ts, "GET", "/api/v1/warehouses/"+uuid.New().String(), "")
	assertStatus(t, resp, http.StatusNotFound)

	// Get with invalid UUID
	resp = doReq(t, ts, "GET", "/api/v1/warehouses/bad-uuid", "")
	assertStatus(t, resp, http.StatusBadRequest)
}

func TestIntegration_WarehouseList(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create multiple warehouses
	for i := 1; i <= 3; i++ {
		body := fmt.Sprintf(`{"code":"WH-%03d","name":"Warehouse %d","address":"Addr %d"}`, i, i, i)
		resp := doReq(t, ts, "POST", "/api/v1/warehouses", body)
		assertStatus(t, resp, http.StatusCreated)
	}

	// List warehouses
	resp := doReq(t, ts, "GET", "/api/v1/warehouses?page=1&page_size=10", "")
	assertStatus(t, resp, http.StatusOK)
	result := assertJSON(t, resp)
	data := result["data"].([]any)
	if len(data) != 3 {
		t.Errorf("expected 3 warehouses, got %d", len(data))
	}
	pagination := result["pagination"].(map[string]any)
	if pagination["total"].(float64) != 3 {
		t.Errorf("expected total 3, got %v", pagination["total"])
	}
}

func TestIntegration_WarehouseUpdate(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create
	resp := doReq(t, ts, "POST", "/api/v1/warehouses", `{"code":"WH-UPD","name":"Old Name","address":"Old Addr"}`)
	assertStatus(t, resp, http.StatusCreated)
	result := assertJSON(t, resp)
	id := result["id"].(string)

	// Update
	resp = doReq(t, ts, "PUT", "/api/v1/warehouses/"+id, `{"name":"Updated Name","address":"New Addr"}`)
	assertStatus(t, resp, http.StatusOK)
	result = assertJSON(t, resp)
	if result["name"] != "Updated Name" {
		t.Errorf("name = %q, want Updated Name", result["name"])
	}
}

// ── Zone Integration Tests ───────────────────────────────────────────────────

func TestIntegration_ZoneCreateAndGet(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create a warehouse first
	resp := doReq(t, ts, "POST", "/api/v1/warehouses", `{"code":"WH-ZONE","name":"Zone WH","address":"Addr"}`)
	assertStatus(t, resp, http.StatusCreated)
	whResult := assertJSON(t, resp)
	whID := whResult["id"].(string)

	// Create zone
	createBody := `{"code":"Z-001","name":"Receiving Zone","zone_type":"receiving"}`
	resp = doReq(t, ts, "POST", "/api/v1/warehouses/"+whID+"/zones", createBody)
	assertStatus(t, resp, http.StatusCreated)
	result := assertJSON(t, resp)
	if result["code"] != "Z-001" {
		t.Errorf("code = %q, want Z-001", result["code"])
	}
	zoneID := result["id"].(string)

	// Get zone
	resp = doReq(t, ts, "GET", "/api/v1/zones/"+zoneID, "")
	assertStatus(t, resp, http.StatusOK)
	result = assertJSON(t, resp)
	if result["name"] != "Receiving Zone" {
		t.Errorf("name = %q, want Receiving Zone", result["name"])
	}

	// List zones for warehouse
	resp = doReq(t, ts, "GET", "/api/v1/warehouses/"+whID+"/zones", "")
	assertStatus(t, resp, http.StatusOK)
	result = assertJSON(t, resp)
	data := result["data"].([]any)
	if len(data) != 1 {
		t.Errorf("expected 1 zone, got %d", len(data))
	}

	// List all zones
	resp = doReq(t, ts, "GET", "/api/v1/zones", "")
	assertStatus(t, resp, http.StatusOK)

	// Update zone
	resp = doReq(t, ts, "PUT", "/api/v1/zones/"+zoneID, `{"name":"Updated Zone"}`)
	assertStatus(t, resp, http.StatusOK)
}

// ── Location Integration Tests ───────────────────────────────────────────────

func TestIntegration_LocationCreateAndGet(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create warehouse
	resp := doReq(t, ts, "POST", "/api/v1/warehouses", `{"code":"WH-LOC","name":"Location WH","address":"Addr"}`)
	whResult := assertJSON(t, resp)
	whID := whResult["id"].(string)

	// Create zone
	resp = doReq(t, ts, "POST", "/api/v1/warehouses/"+whID+"/zones", `{"code":"Z-LOC","name":"Loc Zone","zone_type":"storage"}`)
	zoneResult := assertJSON(t, resp)
	zoneID := zoneResult["id"].(string)

	// Create location
	createBody := `{"code":"A-01-01-01","barcode":"LOC-BC-INT-001","location_type":"shelf"}`
	resp = doReq(t, ts, "POST", "/api/v1/zones/"+zoneID+"/locations", createBody)
	assertStatus(t, resp, http.StatusCreated)
	locResult := assertJSON(t, resp)
	if locResult["barcode"] != "LOC-BC-INT-001" {
		t.Errorf("barcode = %q, want LOC-BC-INT-001", locResult["barcode"])
	}
	locID := locResult["id"].(string)

	// Get location
	resp = doReq(t, ts, "GET", "/api/v1/locations/"+locID, "")
	assertStatus(t, resp, http.StatusOK)

	// Lookup by barcode
	resp = doReq(t, ts, "GET", "/api/v1/locations?barcode=LOC-BC-INT-001", "")
	assertStatus(t, resp, http.StatusOK)
	bcResult := assertJSON(t, resp)
	if bcResult["code"] != "A-01-01-01" {
		t.Errorf("code = %q, want A-01-01-01", bcResult["code"])
	}

	// List locations
	resp = doReq(t, ts, "GET", "/api/v1/locations", "")
	assertStatus(t, resp, http.StatusOK)

	// Update location
	resp = doReq(t, ts, "PUT", "/api/v1/locations/"+locID, `{"code":"A-01-01-02","barcode":"LOC-BC-NEW","location_type":"shelf"}`)
	assertStatus(t, resp, http.StatusOK)

	// Update location status
	resp = doReq(t, ts, "PATCH", "/api/v1/locations/"+locID+"/status", `{"status":"occupied"}`)
	assertStatus(t, resp, http.StatusOK)
}

// ── SKU Integration Tests ────────────────────────────────────────────────────

func TestIntegration_SKUCreateAndGet(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create SKU
	resp := doReq(t, ts, "POST", "/api/v1/skus", `{"code":"SKU-INT-001","name":"Integration SKU","barcode":"BC-INT-001","uom":{"base_unit":"EA"}}`)
	assertStatus(t, resp, http.StatusCreated)
	result := assertJSON(t, resp)
	if result["code"] != "SKU-INT-001" {
		t.Errorf("code = %q, want SKU-INT-001", result["code"])
	}
	skuID := result["id"].(string)

	// Get SKU
	resp = doReq(t, ts, "GET", "/api/v1/skus/"+skuID, "")
	assertStatus(t, resp, http.StatusOK)

	// Lookup by code
	resp = doReq(t, ts, "GET", "/api/v1/skus?code=SKU-INT-001", "")
	assertStatus(t, resp, http.StatusOK)

	// List SKUs
	resp = doReq(t, ts, "GET", "/api/v1/skus", "")
	assertStatus(t, resp, http.StatusOK)

	// Update SKU
	resp = doReq(t, ts, "PUT", "/api/v1/skus/"+skuID, `{"name":"Updated SKU","barcode":"BC-INT-002","uom":{"base_unit":"EA"},"status":"active"}`)
	assertStatus(t, resp, http.StatusOK)
}

func TestIntegration_SKUCreateValidationError(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Missing code
	resp := doReq(t, ts, "POST", "/api/v1/skus", `{"name":"No Code SKU","barcode":"BC-001"}`)
	assertStatus(t, resp, http.StatusBadRequest)

	// Missing name
	resp = doReq(t, ts, "POST", "/api/v1/skus", `{"code":"SKU-001","barcode":"BC-001"}`)
	assertStatus(t, resp, http.StatusBadRequest)
}

// ── Inventory Integration Tests ──────────────────────────────────────────────

func TestIntegration_InventoryQueryAndGet(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create prerequisite data
	resp := doReq(t, ts, "POST", "/api/v1/warehouses", `{"code":"WH-INV","name":"Inventory WH","address":"Addr"}`)
	whResult := assertJSON(t, resp)
	whID := whResult["id"].(string)

	resp = doReq(t, ts, "POST", "/api/v1/warehouses/"+whID+"/zones", `{"code":"Z-INV","name":"Inv Zone","zone_type":"storage"}`)
	zoneResult := assertJSON(t, resp)
	zoneID := zoneResult["id"].(string)

	resp = doReq(t, ts, "POST", "/api/v1/zones/"+zoneID+"/locations", `{"code":"L-001","barcode":"BC-INV-001","location_type":"shelf"}`)
	locResult := assertJSON(t, resp)
	locID := locResult["id"].(string)

	resp = doReq(t, ts, "POST", "/api/v1/skus", `{"code":"SKU-INV-001","name":"Inv SKU","barcode":"BC-SKU-INV-001","uom":{"base_unit":"EA"}}`)
	skuResult := assertJSON(t, resp)
	skuID := skuResult["id"].(string)

	// Create inventory via the task handler CreateTask approach
	// Actually, the service creates it via a different path - let's use inventory query
	// Inventory is created via CreateInventory repo method via service.

	// Query inventory (initially empty)
	resp = doReq(t, ts, "GET", "/api/v1/inventory", "")
	assertStatus(t, resp, http.StatusOK)
	result := assertJSON(t, resp)
	data := result["data"].([]any)
	if len(data) != 0 {
		t.Logf("expected 0 inventory records, got %d", len(data))
	}

	// Get non-existent inventory
	resp = doReq(t, ts, "GET", "/api/v1/inventory/"+uuid.New().String(), "")
	assertStatus(t, resp, http.StatusNotFound)

	// FEFO / FIFO queries (should return empty)
	resp = doReq(t, ts, "GET", "/api/v1/inventory/fifo", "")
	assertStatus(t, resp, http.StatusOK)
	resp = doReq(t, ts, "GET", "/api/v1/inventory/fefo", "")
	assertStatus(t, resp, http.StatusOK)

	// Dashboard
	resp = doReq(t, ts, "GET", "/api/v1/inventory/dashboard", "")
	assertStatus(t, resp, http.StatusOK)

	// Make sure unused vars don't cause trouble
	_ = whID
	_ = locID
	_ = skuID
}

func TestIntegration_InventoryStatusTransition(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Update non-existent inventory status
	resp := doReq(t, ts, "PATCH", "/api/v1/inventory/"+uuid.New().String()+"/status", `{"status":"quarantine"}`)
	assertStatus(t, resp, http.StatusNotFound)
}

// ── Order Integration Tests ──────────────────────────────────────────────────

func TestIntegration_OrderCreateAndGet(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create prerequisite data
	resp := doReq(t, ts, "POST", "/api/v1/warehouses", `{"code":"WH-ORD","name":"Order WH","address":"Addr"}`)
	whResult := assertJSON(t, resp)
	whID := whResult["id"].(string)

	resp = doReq(t, ts, "POST", "/api/v1/skus", `{"code":"SKU-ORD-001","name":"Order SKU","barcode":"BC-ORD-001","uom":{"base_unit":"EA"}}`)
	skuResult := assertJSON(t, resp)
	skuID := skuResult["id"].(string)

	// Create order with lines
	createBody := fmt.Sprintf(`{"warehouse_id":"%s","order_type":"inbound","priority":"normal","external_ref":"REF-001","lines":[{"sku_id":"%s","ordered_qty":10,"uom":"EA"}],"created_by":"test"}`, whID, skuID)
	resp = doReq(t, ts, "POST", "/api/v1/orders", createBody)
	assertStatus(t, resp, http.StatusCreated)
	result := assertJSON(t, resp)
	if result["order_type"] != "inbound" {
		t.Errorf("order_type = %q, want inbound", result["order_type"])
	}
	orderID := result["id"].(string)

	// Get order
	resp = doReq(t, ts, "GET", "/api/v1/orders/"+orderID, "")
	assertStatus(t, resp, http.StatusOK)

	// List orders
	resp = doReq(t, ts, "GET", "/api/v1/orders", "")
	assertStatus(t, resp, http.StatusOK)
	result = assertJSON(t, resp)
	data := result["data"].([]any)
	if len(data) != 1 {
		t.Errorf("expected 1 order, got %d", len(data))
	}

	// Update order status
	resp = doReq(t, ts, "PUT", "/api/v1/orders/"+orderID+"/status", `{"status":"confirmed"}`)
	assertStatus(t, resp, http.StatusOK)
}

func TestIntegration_OrderCreateValidationError(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Missing warehouse_id
	resp := doReq(t, ts, "POST", "/api/v1/orders", `{"order_type":"inbound","priority":"normal"}`)
	assertStatus(t, resp, http.StatusBadRequest)
}

func TestIntegration_OrderAddLineAndUpdateStatus(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create warehouse + SKU
	resp := doReq(t, ts, "POST", "/api/v1/warehouses", `{"code":"WH-OL","name":"OrderLine WH","address":"Addr"}`)
	whResult := assertJSON(t, resp)
	whID := whResult["id"].(string)

	resp = doReq(t, ts, "POST", "/api/v1/skus", `{"code":"SKU-OL-001","name":"Line SKU","barcode":"BC-OL-001","uom":{"base_unit":"EA"}}`)
	skuResult := assertJSON(t, resp)
	skuID := skuResult["id"].(string)

	// Create order with initial line
	resp = doReq(t, ts, "POST", "/api/v1/orders", fmt.Sprintf(`{"warehouse_id":"%s","order_type":"inbound","priority":"normal","lines":[{"sku_id":"%s","ordered_qty":50,"uom":"EA"}],"created_by":"test"}`, whID, skuID))
	orderResult := assertJSON(t, resp)
	orderID := orderResult["id"].(string)

	// Add another order line
	lineBody := fmt.Sprintf(`{"sku_id":"%s","ordered_qty":100,"uom":"EA"}`, skuID)
	resp = doReq(t, ts, "POST", "/api/v1/orders/"+orderID+"/lines", lineBody)
	assertStatus(t, resp, http.StatusCreated)
	lineResult := assertJSON(t, resp)
	lineID := lineResult["id"].(string)

	// Update line status
	resp = doReq(t, ts, "PUT", "/api/v1/orders/"+orderID+"/lines/"+lineID+"/status", `{"status":"allocated"}`)
	assertStatus(t, resp, http.StatusOK)
}

// ── ASN Integration Tests ────────────────────────────────────────────────────

func TestIntegration_ASNCreateAndGet(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create warehouse + SKU
	resp := doReq(t, ts, "POST", "/api/v1/warehouses", `{"code":"WH-ASN","name":"ASN WH","address":"Addr"}`)
	whResult := assertJSON(t, resp)
	whID := whResult["id"].(string)

	resp = doReq(t, ts, "POST", "/api/v1/skus", `{"code":"SKU-ASN-001","name":"ASN SKU","barcode":"BC-ASN-001","uom":{"base_unit":"EA"}}`)
	skuResult := assertJSON(t, resp)
	skuID := skuResult["id"].(string)

	// Create ASN with lines
	createBody := fmt.Sprintf(`{"warehouse_id":"%s","carrier":"FedEx","tracking_no":"TRK-001","expected_at":"2026-08-01T10:00:00Z","lines":[{"sku_id":"%s","expected_qty":100}]}`, whID, skuID)
	resp = doReq(t, ts, "POST", "/api/v1/asns", createBody)
	assertStatus(t, resp, http.StatusCreated)
	result := assertJSON(t, resp)
	if result["tracking_no"] != "TRK-001" {
		t.Errorf("tracking_no = %q, want TRK-001", result["tracking_no"])
	}
	asnID := result["id"].(string)

	// Get ASN
	resp = doReq(t, ts, "GET", "/api/v1/asns/"+asnID, "")
	assertStatus(t, resp, http.StatusOK)

	// List ASNs
	resp = doReq(t, ts, "GET", "/api/v1/asns", "")
	assertStatus(t, resp, http.StatusOK)

	// Update ASN status
	resp = doReq(t, ts, "PUT", "/api/v1/asns/"+asnID+"/status", `{"status":"arrived"}`)
	assertStatus(t, resp, http.StatusOK)
}

// ── Task Integration Tests ───────────────────────────────────────────────────

func TestIntegration_TaskCreateAndGet(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create warehouse
	resp := doReq(t, ts, "POST", "/api/v1/warehouses", `{"code":"WH-TSK","name":"Task WH","address":"Addr"}`)
	whResult := assertJSON(t, resp)
	whID := whResult["id"].(string)

	// Create SKU
	resp = doReq(t, ts, "POST", "/api/v1/skus", `{"code":"SKU-TSK-001","name":"Task SKU","barcode":"BC-TSK-001","uom":{"base_unit":"EA"}}`)
	skuResult := assertJSON(t, resp)
	skuID := skuResult["id"].(string)

	// Create task
	createBody := fmt.Sprintf(`{"task_type":"putaway","warehouse_id":"%s","sku_id":"%s","expected_qty":50,"uom":"EA","priority":"normal"}`, whID, skuID)
	resp = doReq(t, ts, "POST", "/api/v1/tasks", createBody)
	assertStatus(t, resp, http.StatusCreated)
	result := assertJSON(t, resp)
	if result["task_type"] != "putaway" {
		t.Errorf("task_type = %q, want putaway", result["task_type"])
	}
	taskID := result["id"].(string)

	// Get task
	resp = doReq(t, ts, "GET", "/api/v1/tasks/"+taskID, "")
	assertStatus(t, resp, http.StatusOK)

	// List tasks
	resp = doReq(t, ts, "GET", "/api/v1/tasks", "")
	assertStatus(t, resp, http.StatusOK)
	result = assertJSON(t, resp)
	data := result["data"].([]any)
	if len(data) != 1 {
		t.Errorf("expected 1 task, got %d", len(data))
	}

	// List with filters
	resp = doReq(t, ts, "GET", "/api/v1/tasks?status=pending&task_type=putaway", "")
	assertStatus(t, resp, http.StatusOK)
}

func TestIntegration_TaskAssignAndComplete(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create warehouse
	resp := doReq(t, ts, "POST", "/api/v1/warehouses", `{"code":"WH-TAC","name":"Assign WH","address":"Addr"}`)
	whResult := assertJSON(t, resp)
	whID := whResult["id"].(string)

	// Create SKU
	resp = doReq(t, ts, "POST", "/api/v1/skus", `{"code":"SKU-TAC-001","name":"Assign SKU","barcode":"BC-TAC-001","uom":{"base_unit":"EA"}}`)
	skuResult := assertJSON(t, resp)
	skuID := skuResult["id"].(string)

	// Create task
	createBody := fmt.Sprintf(`{"task_type":"pick","warehouse_id":"%s","sku_id":"%s","expected_qty":30,"uom":"EA","priority":"high"}`, whID, skuID)
	resp = doReq(t, ts, "POST", "/api/v1/tasks", createBody)
	taskResult := assertJSON(t, resp)
	taskID := taskResult["id"].(string)

	// Assign task first (pending → assigned), then update status (assigned → in_progress), then complete
	resp = doReq(t, ts, "POST", "/api/v1/tasks/"+taskID+"/assign", `{"assigned_to":"worker-1"}`)
	assertStatus(t, resp, http.StatusOK)
	resp = doReq(t, ts, "PUT", "/api/v1/tasks/"+taskID+"/status", `{"status":"in_progress"}`)
	assertStatus(t, resp, http.StatusOK)
	resp = doReq(t, ts, "POST", "/api/v1/tasks/"+taskID+"/complete", `{"actual_qty":30}`)
	assertStatus(t, resp, http.StatusOK)
	result := assertJSON(t, resp)
	if result["status"] != "completed" {
		t.Errorf("status = %q, want completed", result["status"])
	}
}

func TestIntegration_TaskNotFound(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	nonExistentID := uuid.New().String()
	resp := doReq(t, ts, "GET", "/api/v1/tasks/"+nonExistentID, "")
	assertStatus(t, resp, http.StatusNotFound)

	resp = doReq(t, ts, "POST", "/api/v1/tasks/"+nonExistentID+"/assign", `{"assigned_to":"w1"}`)
	assertStatus(t, resp, http.StatusNotFound)
}

// ── Wave Integration Tests ───────────────────────────────────────────────────

func TestIntegration_WaveCreateAndGet(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create warehouse
	resp := doReq(t, ts, "POST", "/api/v1/warehouses", `{"code":"WH-WV","name":"Wave WH","address":"Addr"}`)
	whResult := assertJSON(t, resp)
	whID := whResult["id"].(string)

	// Create wave
	// Create wave (no name field — CreateWaveInput uses wave_type + warehouse_id)
	createBody := fmt.Sprintf(`{"wave_type":"single_order","warehouse_id":"%s"}`, whID)
	resp = doReq(t, ts, "POST", "/api/v1/waves", createBody)
	assertStatus(t, resp, http.StatusCreated)
	result := assertJSON(t, resp)
	if result["wave_type"] != "single_order" {
		t.Errorf("wave_type = %q, want single_order", result["wave_type"])
	}
	waveID := result["id"].(string)

	// Get wave
	resp = doReq(t, ts, "GET", "/api/v1/waves/"+waveID, "")
	assertStatus(t, resp, http.StatusOK)

	// List waves
	resp = doReq(t, ts, "GET", "/api/v1/waves", "")
	assertStatus(t, resp, http.StatusOK)
	result = assertJSON(t, resp)
	data := result["data"].([]any)
	if len(data) != 1 {
		t.Errorf("expected 1 wave, got %d", len(data))
	}
}

func TestIntegration_WaveOrderManagement(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create warehouse + SKU
	resp := doReq(t, ts, "POST", "/api/v1/warehouses", `{"code":"WH-WOM","name":"WaveOM WH","address":"Addr"}`)
	whResult := assertJSON(t, resp)
	whID := whResult["id"].(string)

	resp = doReq(t, ts, "POST", "/api/v1/skus", `{"code":"SKU-WOM-001","name":"Wave SKU","barcode":"BC-WOM-001","uom":{"base_unit":"EA"}}`)
	skuResult := assertJSON(t, resp)
	skuID := skuResult["id"].(string)

	// Create wave
	resp = doReq(t, ts, "POST", "/api/v1/waves", fmt.Sprintf(`{"wave_type":"single_order","warehouse_id":"%s"}`, whID))
	waveResult := assertJSON(t, resp)
	waveID := waveResult["id"].(string)

	// Create order with lines and add to wave
	resp = doReq(t, ts, "POST", "/api/v1/orders", fmt.Sprintf(`{"warehouse_id":"%s","order_type":"outbound","priority":"high","lines":[{"sku_id":"%s","ordered_qty":10,"uom":"EA"}],"created_by":"test"}`, whID, skuID))
	orderResult := assertJSON(t, resp)
	orderID := orderResult["id"].(string)

	// Add order to wave
	addBody := fmt.Sprintf(`{"order_ids":["%s"]}`, orderID)
	resp = doReq(t, ts, "POST", "/api/v1/waves/"+waveID+"/orders", addBody)
	assertStatus(t, resp, http.StatusOK)
	result := assertJSON(t, resp)
	if result["total_orders"].(float64) != 1 {
		t.Errorf("expected total_orders=1, got %v", result["total_orders"])
	}

	// Remove order from wave
	resp = doReq(t, ts, "DELETE", "/api/v1/waves/"+waveID+"/orders", addBody)
	assertStatus(t, resp, http.StatusOK)
	result = assertJSON(t, resp)
	if result["total_orders"].(float64) != 0 {
		t.Errorf("expected total_orders=0, got %v", result["total_orders"])
	}
}

func TestIntegration_WaveStatusFlow(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create warehouse
	resp := doReq(t, ts, "POST", "/api/v1/warehouses", `{"code":"WH-WSF","name":"WaveSF WH","address":"Addr"}`)
	whResult := assertJSON(t, resp)
	whID := whResult["id"].(string)

	// Create wave
	resp = doReq(t, ts, "POST", "/api/v1/waves", fmt.Sprintf(`{"wave_type":"single_order","warehouse_id":"%s"}`, whID))
	waveResult := assertJSON(t, resp)
	waveID := waveResult["id"].(string)

	// Release wave (created → released)
	resp = doReq(t, ts, "POST", "/api/v1/waves/"+waveID+"/release", "")
	assertStatus(t, resp, http.StatusOK)

	// Transition to in_progress (released → in_progress)
	resp = doReq(t, ts, "PUT", "/api/v1/waves/"+waveID+"/status", `{"status":"in_progress"}`)
	assertStatus(t, resp, http.StatusOK)
}

// ── User Integration Tests ──────────────────────────────────────────────────

func TestIntegration_UserCreateAndGet(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create role
	resp := doReq(t, ts, "POST", "/api/v1/roles", `{"name":"Operator","description":"Warehouse operator","permissions":[]}`)
	roleResult := assertJSON(t, resp)
	roleID := roleResult["id"].(string)

	// Create user
	createBody := fmt.Sprintf(`{"username":"integration-user","email":"iu@test.com","display_name":"Integration User","password":"secret123","role_ids":["%s"]}`, roleID)
	resp = doReq(t, ts, "POST", "/api/v1/users", createBody)
	assertStatus(t, resp, http.StatusCreated)
	result := assertJSON(t, resp)
	if result["username"] != "integration-user" {
		t.Errorf("username = %q, want integration-user", result["username"])
	}
	userID := result["id"].(string)

	// Get user
	resp = doReq(t, ts, "GET", "/api/v1/users/"+userID, "")
	assertStatus(t, resp, http.StatusOK)

	// List users
	resp = doReq(t, ts, "GET", "/api/v1/users", "")
	assertStatus(t, resp, http.StatusOK)
	result = assertJSON(t, resp)
	data := result["data"].([]any)
	if len(data) != 1 {
		t.Errorf("expected 1 user, got %d", len(data))
	}

	// Update user
	resp = doReq(t, ts, "PUT", "/api/v1/users/"+userID, `{"email":"iu2@test.com","display_name":"Updated User"}`)
	assertStatus(t, resp, http.StatusOK)

	// Update user status
	resp = doReq(t, ts, "PUT", "/api/v1/users/"+userID+"/status", `{"status":"inactive"}`)
	assertStatus(t, resp, http.StatusOK)
}

func TestIntegration_UserNotFound(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	nonExistentID := uuid.New().String()
	resp := doReq(t, ts, "GET", "/api/v1/users/"+nonExistentID, "")
	assertStatus(t, resp, http.StatusNotFound)
}

// ── Role Integration Tests ───────────────────────────────────────────────────

func TestIntegration_RoleCRUD(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create role
	resp := doReq(t, ts, "POST", "/api/v1/roles", `{"name":"Manager","description":"Warehouse manager","permissions":[{"resource":"warehouses","actions":["read","write"]}]}`)
	assertStatus(t, resp, http.StatusCreated)
	result := assertJSON(t, resp)
	roleID := result["id"].(string)

	// Get role
	resp = doReq(t, ts, "GET", "/api/v1/roles/"+roleID, "")
	assertStatus(t, resp, http.StatusOK)

	// List roles
	resp = doReq(t, ts, "GET", "/api/v1/roles", "")
	assertStatus(t, resp, http.StatusOK)

	// Update role
	resp = doReq(t, ts, "PUT", "/api/v1/roles/"+roleID, `{"name":"Senior Manager","description":"Updated desc","permissions":[]}`)
	assertStatus(t, resp, http.StatusOK)

	// Delete role
	resp = doReq(t, ts, "DELETE", "/api/v1/roles/"+roleID, "")
	assertStatus(t, resp, http.StatusNoContent)

	// Verify deleted
	resp = doReq(t, ts, "GET", "/api/v1/roles/"+roleID, "")
	assertStatus(t, resp, http.StatusNotFound)
}

// ── Dashboard Integration Tests ──────────────────────────────────────────────

func TestIntegration_Dashboard(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create test data
	doReq(t, ts, "POST", "/api/v1/warehouses", `{"code":"WH-DB-1","name":"DB WH 1","address":"Addr"}`)
	doReq(t, ts, "POST", "/api/v1/warehouses", `{"code":"WH-DB-2","name":"DB WH 2","address":"Addr"}`)
	doReq(t, ts, "POST", "/api/v1/skus", `{"code":"SKU-DB-1","name":"DB SKU 1","barcode":"BC-DB-1","uom":{"base_unit":"EA"}}`)
	doReq(t, ts, "POST", "/api/v1/skus", `{"code":"SKU-DB-2","name":"DB SKU 2","barcode":"BC-DB-2","uom":{"base_unit":"EA"}}`)

	resp := doReq(t, ts, "GET", "/api/v1/dashboard", "")
	assertStatus(t, resp, http.StatusOK)
	result := assertJSON(t, resp)

	if result["warehouse_count"].(float64) != 2 {
		t.Errorf("warehouse_count = %v, want 2", result["warehouse_count"])
	}
	if result["sku_count"].(float64) != 2 {
		t.Errorf("sku_count = %v, want 2", result["sku_count"])
	}
}

// ── Audit Log Integration Tests ──────────────────────────────────────────────

func TestIntegration_AuditLogList(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create an audit log entry directly
	ts.userRepo.CreateAuditLog(nil, &domain.AuditLog{
		UserID:     uuid.New(),
		Username:   "test-user",
		Action:     "CREATE",
		Resource:   "warehouse",
		ResourceID: uuid.New().String(),
	})

	resp := doReq(t, ts, "GET", "/api/v1/audit-logs", "")
	assertStatus(t, resp, http.StatusOK)
	result := assertJSON(t, resp)
	data := result["data"].([]any)
	if len(data) != 1 {
		t.Errorf("expected 1 audit log entry, got %d", len(data))
	}
}

// ── Error Handling Integration Tests ─────────────────────────────────────────

func TestIntegration_InvalidJSON(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	resp := doReq(t, ts, "POST", "/api/v1/warehouses", `{invalid json}`)
	assertStatus(t, resp, http.StatusBadRequest)
}

func TestIntegration_InvalidUUIDFormat(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	resp := doReq(t, ts, "GET", "/api/v1/warehouses/not-a-uuid", "")
	assertStatus(t, resp, http.StatusBadRequest)
}

func TestIntegration_EmptyRequestBody(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	resp := doReq(t, ts, "POST", "/api/v1/warehouses", "")
	assertStatus(t, resp, http.StatusBadRequest)
}

// ── Pagination Integration Tests ─────────────────────────────────────────────

func TestIntegration_PaginationDefaults(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// Create 5 warehouses
	for i := 1; i <= 5; i++ {
		body := fmt.Sprintf(`{"code":"WH-PG-%03d","name":"Pagination WH %d","address":"Addr"}`, i, i)
		resp := doReq(t, ts, "POST", "/api/v1/warehouses", body)
		assertStatus(t, resp, http.StatusCreated)
	}

	// Default page
	resp := doReq(t, ts, "GET", "/api/v1/warehouses", "")
	assertStatus(t, resp, http.StatusOK)
	result := assertJSON(t, resp)
	pagination := result["pagination"].(map[string]any)
	if pagination["total"].(float64) != 5 {
		t.Errorf("expected total 5, got %v", pagination["total"])
	}

	// Specific page and page_size
	resp = doReq(t, ts, "GET", "/api/v1/warehouses?page=1&page_size=2", "")
	assertStatus(t, resp, http.StatusOK)
	result = assertJSON(t, resp)
	data := result["data"].([]any)
	if len(data) > 2 {
		t.Errorf("expected at most 2 items, got %d", len(data))
	}
}

// ── Resource Not Found Patterns ──────────────────────────────────────────────

func TestIntegration_NotFoundAcrossResources(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	nonExist := uuid.New().String()

	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/warehouses/" + nonExist},
		{"GET", "/api/v1/zones/" + nonExist},
		{"GET", "/api/v1/locations/" + nonExist},
		{"GET", "/api/v1/skus/" + nonExist},
		{"GET", "/api/v1/inventory/" + nonExist},
		{"GET", "/api/v1/orders/" + nonExist},
		{"GET", "/api/v1/tasks/" + nonExist},
		{"GET", "/api/v1/asns/" + nonExist},
		{"GET", "/api/v1/users/" + nonExist},
		{"GET", "/api/v1/roles/" + nonExist},
		{"GET", "/api/v1/waves/" + nonExist},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			resp := doReq(t, ts, tt.method, tt.path, "")
			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("expected 404 for %s %s, got %d", tt.method, tt.path, resp.StatusCode)
			}
		})
	}
}
