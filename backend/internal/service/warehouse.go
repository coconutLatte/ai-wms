// Package service implements business logic orchestration for the WMS domain.
// Services take repository interfaces (not concrete implementations) following DDD principles.
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// WarehouseService orchestrates business logic for warehouses, zones, and locations.
type WarehouseService struct {
	repo repository.WarehouseRepository
}

// NewWarehouseService creates a new WarehouseService.
func NewWarehouseService(repo repository.WarehouseRepository) *WarehouseService {
	return &WarehouseService{repo: repo}
}

// ── Warehouse ────────────────────────────────────────────────────────────────────────────

// CreateWarehouseInput is the input for creating a new warehouse.
type CreateWarehouseInput struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

// Validate checks the input for business rule violations.
func (in *CreateWarehouseInput) Validate() error {
	if in.Code == "" {
		return pkgerrors.NewInvalidInput("warehouse code is required")
	}
	if in.Name == "" {
		return pkgerrors.NewInvalidInput("warehouse name is required")
	}
	return nil
}

// CreateWarehouse validates input and creates a new warehouse.
func (s *WarehouseService) CreateWarehouse(ctx context.Context, input CreateWarehouseInput) (*domain.Warehouse, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	w := &domain.Warehouse{
		Code:    input.Code,
		Name:    input.Name,
		Address: input.Address,
		Status:  domain.WarehouseStatusActive,
	}

	if err := s.repo.CreateWarehouse(ctx, w); err != nil {
		return nil, fmt.Errorf("warehouse service: create: %w", err)
	}

	return w, nil
}

// GetWarehouse retrieves a warehouse by ID.
func (s *WarehouseService) GetWarehouse(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	w, err := s.repo.GetWarehouse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("warehouse service: get %s: %w", id, err)
	}
	return w, nil
}

// ListWarehouses returns all warehouses with pagination support.
func (s *WarehouseService) ListWarehouses(ctx context.Context, limit, offset int) ([]*domain.Warehouse, int, error) {
	warehouses, err := s.repo.ListWarehouses(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("warehouse service: list: %w", err)
	}

	total, err := s.repo.CountWarehouses(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("warehouse service: count: %w", err)
	}

	return warehouses, total, nil
}

// CountWarehouses returns the total number of warehouses.
func (s *WarehouseService) CountWarehouses(ctx context.Context) (int, error) {
	total, err := s.repo.CountWarehouses(ctx)
	if err != nil {
		return 0, fmt.Errorf("warehouse service: count: %w", err)
	}
	return total, nil
}

// UpdateWarehouseInput is the input for updating a warehouse.
type UpdateWarehouseInput struct {
	Name    *string                 `json:"name,omitempty"`
	Address *string                 `json:"address,omitempty"`
	Status  *domain.WarehouseStatus `json:"status,omitempty"`
}

// UpdateWarehouse validates and updates an existing warehouse.
func (s *WarehouseService) UpdateWarehouse(ctx context.Context, id uuid.UUID, input UpdateWarehouseInput) (*domain.Warehouse, error) {
	w, err := s.repo.GetWarehouse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("warehouse service: update %s: %w", id, err)
	}

	if input.Name != nil {
		w.Name = *input.Name
	}
	if input.Address != nil {
		w.Address = *input.Address
	}
	if input.Status != nil {
		status := *input.Status
		if !isValidWarehouseStatus(status) {
			return nil, pkgerrors.NewInvalidInput(fmt.Sprintf("invalid warehouse status: %s", status))
		}
		w.Status = status
	}

	if err := s.repo.UpdateWarehouse(ctx, w); err != nil {
		return nil, fmt.Errorf("warehouse service: update %s: %w", id, err)
	}

	return w, nil
}

func isValidWarehouseStatus(s domain.WarehouseStatus) bool {
	switch s {
	case domain.WarehouseStatusActive, domain.WarehouseStatusInactive, domain.WarehouseStatusArchived:
		return true
	}
	return false
}

// ── Zone ─────────────────────────────────────────────────────────────────────────────────

// CreateZoneInput is the input for creating a new zone.
type CreateZoneInput struct {
	Code     string          `json:"code"`
	Name     string          `json:"name"`
	ZoneType domain.ZoneType `json:"zone_type"`
}

// Validate checks the input for business rule violations.
func (in *CreateZoneInput) Validate() error {
	if in.Code == "" {
		return pkgerrors.NewInvalidInput("zone code is required")
	}
	if in.Name == "" {
		return pkgerrors.NewInvalidInput("zone name is required")
	}
	if !isValidZoneType(in.ZoneType) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid zone type: %s", in.ZoneType))
	}
	return nil
}

func isValidZoneType(t domain.ZoneType) bool {
	switch t {
	case domain.ZoneTypeReceiving, domain.ZoneTypeStorage, domain.ZoneTypePicking,
		domain.ZoneTypeShipping, domain.ZoneTypeReturns, domain.ZoneTypeStaging:
		return true
	}
	return false
}

// CreateZone validates the input and parent warehouse, then creates the zone.
func (s *WarehouseService) CreateZone(ctx context.Context, warehouseID uuid.UUID, input CreateZoneInput) (*domain.Zone, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Verify warehouse exists.
	if _, err := s.repo.GetWarehouse(ctx, warehouseID); err != nil {
		return nil, fmt.Errorf("warehouse service: create zone: %w", err)
	}

	z := &domain.Zone{
		WarehouseID: warehouseID,
		Code:        input.Code,
		Name:        input.Name,
		ZoneType:    input.ZoneType,
		Status:      domain.ZoneStatusActive,
	}

	if err := s.repo.CreateZone(ctx, z); err != nil {
		return nil, fmt.Errorf("warehouse service: create zone: %w", err)
	}

	return z, nil
}

// GetZone retrieves a zone by ID.
func (s *WarehouseService) GetZone(ctx context.Context, id uuid.UUID) (*domain.Zone, error) {
	z, err := s.repo.GetZone(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("warehouse service: get zone %s: %w", id, err)
	}
	return z, nil
}

// ListZones returns all zones in a warehouse with pagination support.
func (s *WarehouseService) ListZones(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]*domain.Zone, int, error) {
	// Verify warehouse exists.
	if _, err := s.repo.GetWarehouse(ctx, warehouseID); err != nil {
		return nil, 0, fmt.Errorf("warehouse service: list zones: %w", err)
	}

	zones, err := s.repo.ListZonesByWarehouse(ctx, warehouseID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("warehouse service: list zones: %w", err)
	}

	total, err := s.repo.CountZonesByWarehouse(ctx, warehouseID)
	if err != nil {
		return nil, 0, fmt.Errorf("warehouse service: count zones: %w", err)
	}

	return zones, total, nil
}

// ListAllZones returns zones globally, optionally filtered by warehouse, with pagination.
func (s *WarehouseService) ListAllZones(ctx context.Context, filter repository.ZoneFilter) ([]*domain.Zone, int, error) {
	zones, err := s.repo.ListAllZones(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("warehouse service: list all zones: %w", err)
	}

	total, err := s.repo.CountAllZones(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("warehouse service: count all zones: %w", err)
	}

	return zones, total, nil
}

// UpdateZoneInput is the input for updating a zone.
type UpdateZoneInput struct {
	Name     *string           `json:"name,omitempty"`
	ZoneType *domain.ZoneType  `json:"zone_type,omitempty"`
	Status   *domain.ZoneStatus `json:"status,omitempty"`
}

// UpdateZone validates and updates an existing zone.
func (s *WarehouseService) UpdateZone(ctx context.Context, id uuid.UUID, input UpdateZoneInput) (*domain.Zone, error) {
	z, err := s.repo.GetZone(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("warehouse service: update zone %s: %w", id, err)
	}

	if input.Name != nil {
		z.Name = *input.Name
	}
	if input.ZoneType != nil {
		zt := *input.ZoneType
		if !isValidZoneType(zt) {
			return nil, pkgerrors.NewInvalidInput(fmt.Sprintf("invalid zone type: %s", zt))
		}
		z.ZoneType = zt
	}
	if input.Status != nil {
		s := *input.Status
		if !isValidZoneStatus(s) {
			return nil, pkgerrors.NewInvalidInput(fmt.Sprintf("invalid zone status: %s", s))
		}
		z.Status = s
	}

	if err := s.repo.UpdateZone(ctx, z); err != nil {
		return nil, fmt.Errorf("warehouse service: update zone %s: %w", id, err)
	}

	return z, nil
}

func isValidZoneStatus(s domain.ZoneStatus) bool {
	switch s {
	case domain.ZoneStatusActive, domain.ZoneStatusInactive, domain.ZoneStatusFull:
		return true
	}
	return false
}

// ── Location ─────────────────────────────────────────────────────────────────────────────

// CreateLocationInput is the input for creating a new location.
type CreateLocationInput struct {
	Code         string             `json:"code"`
	Barcode      string             `json:"barcode,omitempty"`
	LocationType domain.LocationType `json:"location_type"`
	Capacity     *domain.Capacity   `json:"capacity,omitempty"`
}

// Validate checks the input for business rule violations.
func (in *CreateLocationInput) Validate() error {
	if in.Code == "" {
		return pkgerrors.NewInvalidInput("location code is required")
	}
	if !isValidLocationType(in.LocationType) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid location type: %s", in.LocationType))
	}
	return nil
}

func isValidLocationType(t domain.LocationType) bool {
	switch t {
	case domain.LocationTypePallet, domain.LocationTypeShelf, domain.LocationTypeFloor,
		domain.LocationTypeConveyor, domain.LocationTypeAGV:
		return true
	}
	return false
}

// CreateLocation validates input and parent zone, then creates the location.
func (s *WarehouseService) CreateLocation(ctx context.Context, zoneID uuid.UUID, input CreateLocationInput) (*domain.Location, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Verify zone exists and get its warehouse ID.
	z, err := s.repo.GetZone(ctx, zoneID)
	if err != nil {
		return nil, fmt.Errorf("warehouse service: create location: %w", err)
	}

	loc := &domain.Location{
		ZoneID:       zoneID,
		WarehouseID:  z.WarehouseID,
		Code:         input.Code,
		Barcode:      input.Barcode,
		LocationType: input.LocationType,
		Capacity:     input.Capacity,
		Status:       domain.LocationStatusEmpty,
	}

	if err := s.repo.CreateLocation(ctx, loc); err != nil {
		return nil, fmt.Errorf("warehouse service: create location: %w", err)
	}

	return loc, nil
}

// GetLocation retrieves a location by ID.
func (s *WarehouseService) GetLocation(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	loc, err := s.repo.GetLocation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("warehouse service: get location %s: %w", id, err)
	}
	return loc, nil
}

// ListLocations returns all locations in a zone with pagination support.
func (s *WarehouseService) ListLocations(ctx context.Context, zoneID uuid.UUID, limit, offset int) ([]*domain.Location, int, error) {
	// Verify zone exists.
	if _, err := s.repo.GetZone(ctx, zoneID); err != nil {
		return nil, 0, fmt.Errorf("warehouse service: list locations: %w", err)
	}

	locs, err := s.repo.ListLocationsByZone(ctx, zoneID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("warehouse service: list locations: %w", err)
	}

	total, err := s.repo.CountLocationsByZone(ctx, zoneID)
	if err != nil {
		return nil, 0, fmt.Errorf("warehouse service: count locations: %w", err)
	}

	return locs, total, nil
}

// ListAllLocations returns locations globally, optionally filtered by zone or warehouse, with pagination.
func (s *WarehouseService) ListAllLocations(ctx context.Context, filter repository.LocationFilter) ([]*domain.Location, int, error) {
	locs, err := s.repo.ListAllLocations(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("warehouse service: list all locations: %w", err)
	}

	total, err := s.repo.CountAllLocations(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("warehouse service: count all locations: %w", err)
	}

	return locs, total, nil
}

// UpdateLocationInput is the input for updating a location's metadata.
type UpdateLocationInput struct {
	Code         *string             `json:"code,omitempty"`
	Barcode      *string             `json:"barcode,omitempty"`
	LocationType *domain.LocationType `json:"location_type,omitempty"`
	Capacity     *domain.Capacity    `json:"capacity,omitempty"`
}

// UpdateLocation validates and updates an existing location's metadata (not status).
func (s *WarehouseService) UpdateLocation(ctx context.Context, id uuid.UUID, input UpdateLocationInput) (*domain.Location, error) {
	loc, err := s.repo.GetLocation(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("warehouse service: update location %s: %w", id, err)
	}

	if input.Code != nil {
		loc.Code = *input.Code
	}
	if input.Barcode != nil {
		loc.Barcode = *input.Barcode
	}
	if input.LocationType != nil {
		lt := *input.LocationType
		if !isValidLocationType(lt) {
			return nil, pkgerrors.NewInvalidInput(fmt.Sprintf("invalid location type: %s", lt))
		}
		loc.LocationType = lt
	}
	if input.Capacity != nil {
		loc.Capacity = input.Capacity
	}

	if err := s.repo.UpdateLocation(ctx, loc); err != nil {
		return nil, fmt.Errorf("warehouse service: update location %s: %w", id, err)
	}

	return loc, nil
}

// UpdateLocationStatus updates the status of a location, enforcing the
// Location state machine via CanTransitionTo.
func (s *WarehouseService) UpdateLocationStatus(ctx context.Context, id uuid.UUID, status domain.LocationStatus) error {
	if !isValidLocationStatus(status) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid location status: %s", status))
	}

	loc, err := s.repo.GetLocation(ctx, id)
	if err != nil {
		return fmt.Errorf("warehouse service: update location status %s: %w", id, err)
	}

	if !loc.CanTransitionTo(status) {
		return pkgerrors.NewInvalidStatus(string(loc.Status), string(status))
	}

	if err := s.repo.UpdateLocationStatus(ctx, id, status); err != nil {
		return fmt.Errorf("warehouse service: update location status %s: %w", id, err)
	}
	return nil
}

// GetLocationByBarcode retrieves a location by its barcode.
func (s *WarehouseService) GetLocationByBarcode(ctx context.Context, barcode string) (*domain.Location, error) {
	loc, err := s.repo.GetLocationByBarcode(ctx, barcode)
	if err != nil {
		return nil, fmt.Errorf("warehouse service: get location by barcode %s: %w", barcode, err)
	}
	return loc, nil
}

func isValidLocationStatus(s domain.LocationStatus) bool {
	switch s {
	case domain.LocationStatusEmpty, domain.LocationStatusOccupied,
		domain.LocationStatusReserved, domain.LocationStatusBlocked:
		return true
	}
	return false
}
