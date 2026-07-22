// Package service implements business logic orchestration for the WMS domain.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// WaveService orchestrates business logic for picking waves.
type WaveService struct {
	repo repository.TaskRepository
}

// NewWaveService creates a new WaveService.
func NewWaveService(repo repository.TaskRepository) *WaveService {
	return &WaveService{repo: repo}
}

// ── Input Types ──────────────────────────────────────────────────────────────────────────

// CreateWaveInput is the input for creating a new wave.
type CreateWaveInput struct {
	WaveNo      string          `json:"wave_no,omitempty"` // Auto-generated if empty
	WaveType    domain.WaveType `json:"wave_type"`
	WarehouseID uuid.UUID       `json:"warehouse_id"`
	OrderIDs    []uuid.UUID     `json:"order_ids,omitempty"`
}

// Validate checks the input for business rule violations.
func (in *CreateWaveInput) Validate() error {
	if !isValidWaveType(in.WaveType) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid wave_type: %s", in.WaveType))
	}
	if in.WarehouseID == uuid.Nil {
		return pkgerrors.NewInvalidInput("warehouse_id is required")
	}
	return nil
}

// UpdateWaveStatusInput is the input for updating a wave's status.
type UpdateWaveStatusInput struct {
	Status domain.WaveStatus `json:"status"`
}

// Validate checks the input for business rule violations.
func (in *UpdateWaveStatusInput) Validate() error {
	if !isValidWaveStatus(in.Status) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid wave status: %s", in.Status))
	}
	return nil
}

// AddWaveOrdersInput is the input for adding orders to a wave.
type AddWaveOrdersInput struct {
	OrderIDs []uuid.UUID `json:"order_ids"`
}

// Validate checks the input for business rule violations.
func (in *AddWaveOrdersInput) Validate() error {
	if len(in.OrderIDs) == 0 {
		return pkgerrors.NewInvalidInput("at least one order_id is required")
	}
	return nil
}

// RemoveWaveOrdersInput is the input for removing orders from a wave.
type RemoveWaveOrdersInput struct {
	OrderIDs []uuid.UUID `json:"order_ids"`
}

// Validate checks the input for business rule violations.
func (in *RemoveWaveOrdersInput) Validate() error {
	if len(in.OrderIDs) == 0 {
		return pkgerrors.NewInvalidInput("at least one order_id is required")
	}
	return nil
}

// ── WaveFilterInput ─────────────────────────────────────────────────────────────────────

// WaveQueryParams holds optional query filters for listing waves.
type WaveQueryParams struct {
	WarehouseID uuid.UUID
	Status      domain.WaveStatus
	WaveType    domain.WaveType
	Page        int
	PageSize    int
}

// ToFilter converts query params to a repository filter.
func (q WaveQueryParams) ToFilter() repository.WaveFilter {
	limit := q.PageSize
	offset := (q.Page - 1) * q.PageSize
	if limit < 1 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return repository.WaveFilter{
		WarehouseID: q.WarehouseID,
		Status:      q.Status,
		WaveType:    q.WaveType,
		Limit:       limit,
		Offset:      offset,
	}
}

// ── Service Methods ──────────────────────────────────────────────────────────────────────

// CreateWave validates input and creates a new wave.
func (s *WaveService) CreateWave(ctx context.Context, input CreateWaveInput) (*domain.Wave, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Generate wave number if not provided.
	waveNo := input.WaveNo
	if waveNo == "" {
		waveNo = generateWaveNo()
	}

	orderIDs := input.OrderIDs
	if orderIDs == nil {
		orderIDs = []uuid.UUID{}
	}

	wave := &domain.Wave{
		WaveNo:      waveNo,
		WarehouseID: input.WarehouseID,
		WaveType:    input.WaveType,
		Status:      domain.WaveStatusCreated,
		OrderIDs:    orderIDs,
		TaskIDs:     []uuid.UUID{},
		TotalOrders: len(orderIDs),
	}

	if err := s.repo.CreateWave(ctx, wave); err != nil {
		return nil, fmt.Errorf("wave service: create: %w", err)
	}

	return wave, nil
}

// GetWave retrieves a wave by ID.
func (s *WaveService) GetWave(ctx context.Context, id uuid.UUID) (*domain.Wave, error) {
	wave, err := s.repo.GetWave(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("wave service: get %s: %w", id, err)
	}
	return wave, nil
}

// ListWaves returns waves matching the specified query params.
func (s *WaveService) ListWaves(ctx context.Context, params WaveQueryParams) ([]*domain.Wave, int, error) {
	filter := params.ToFilter()

	waves, err := s.repo.ListWaves(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("wave service: list: %w", err)
	}

	total, err := s.repo.CountWaves(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("wave service: count: %w", err)
	}

	return waves, total, nil
}

// UpdateWaveStatus validates the state transition and updates the wave status.
func (s *WaveService) UpdateWaveStatus(ctx context.Context, id uuid.UUID, input UpdateWaveStatusInput) (*domain.Wave, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	wave, err := s.repo.GetWave(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("wave service: update status %s: %w", id, err)
	}

	// Validate the state transition.
	if !wave.CanTransitionTo(input.Status) {
		return nil, pkgerrors.NewInvalidStatus(string(wave.Status), string(input.Status))
	}

	if err := s.repo.UpdateWaveStatus(ctx, id, input.Status); err != nil {
		return nil, fmt.Errorf("wave service: update status %s: %w", id, err)
	}

	// Re-fetch to get updated state.
	updated, err := s.repo.GetWave(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("wave service: re-fetch after status update %s: %w", id, err)
	}

	return updated, nil
}

// ReleaseWave is a convenience method that transitions a wave from "created" to "released".
func (s *WaveService) ReleaseWave(ctx context.Context, id uuid.UUID) (*domain.Wave, error) {
	return s.UpdateWaveStatus(ctx, id, UpdateWaveStatusInput{Status: domain.WaveStatusReleased})
}

// AddWaveOrders adds orders to an existing wave.
// Only allowed when the wave is in "created" status.
func (s *WaveService) AddWaveOrders(ctx context.Context, waveID uuid.UUID, input AddWaveOrdersInput) (*domain.Wave, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	wave, err := s.repo.GetWave(ctx, waveID)
	if err != nil {
		return nil, fmt.Errorf("wave service: add orders %s: %w", waveID, err)
	}

	if wave.Status != domain.WaveStatusCreated {
		return nil, pkgerrors.NewInvalidInput(
			fmt.Sprintf("can only add orders to created waves (current: %s)", wave.Status))
	}

	if err := s.repo.AddWaveOrders(ctx, waveID, input.OrderIDs); err != nil {
		return nil, fmt.Errorf("wave service: add orders %s: %w", waveID, err)
	}

	// Re-fetch to get updated state.
	updated, err := s.repo.GetWave(ctx, waveID)
	if err != nil {
		return nil, fmt.Errorf("wave service: re-fetch after add orders %s: %w", waveID, err)
	}

	return updated, nil
}

// RemoveWaveOrders removes orders from an existing wave.
// Only allowed when the wave is in "created" status.
func (s *WaveService) RemoveWaveOrders(ctx context.Context, waveID uuid.UUID, input RemoveWaveOrdersInput) (*domain.Wave, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	wave, err := s.repo.GetWave(ctx, waveID)
	if err != nil {
		return nil, fmt.Errorf("wave service: remove orders %s: %w", waveID, err)
	}

	if wave.Status != domain.WaveStatusCreated {
		return nil, pkgerrors.NewInvalidInput(
			fmt.Sprintf("can only remove orders from created waves (current: %s)", wave.Status))
	}

	if err := s.repo.RemoveWaveOrders(ctx, waveID, input.OrderIDs); err != nil {
		return nil, fmt.Errorf("wave service: remove orders %s: %w", waveID, err)
	}

	// Re-fetch to get updated state.
	updated, err := s.repo.GetWave(ctx, waveID)
	if err != nil {
		return nil, fmt.Errorf("wave service: re-fetch after remove orders %s: %w", waveID, err)
	}

	return updated, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────────────────

// generateWaveNo creates a business wave number: WAVE-YYYYMMDD-NNNNNN.
func generateWaveNo() string {
	now := time.Now()
	return fmt.Sprintf("WAVE-%s-%06d", now.Format("20060102"), now.UnixMilli()%1000000)
}

func isValidWaveType(t domain.WaveType) bool {
	switch t {
	case domain.WaveTypeSingleOrder, domain.WaveTypeBatch,
		domain.WaveTypeZone, domain.WaveTypeCarrier:
		return true
	}
	return false
}

func isValidWaveStatus(s domain.WaveStatus) bool {
	switch s {
	case domain.WaveStatusCreated, domain.WaveStatusReleased,
		domain.WaveStatusInProgress, domain.WaveStatusCompleted:
		return true
	}
	return false
}
