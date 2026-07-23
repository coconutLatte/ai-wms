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

// CycleCountService orchestrates business logic for cycle count operations.
type CycleCountService struct {
	repo          repository.CycleCountRepository
	inventoryRepo repository.InventoryRepository
	warehouseRepo repository.WarehouseRepository
	txManager     repository.TxManager
}

// NewCycleCountService creates a new CycleCountService.
func NewCycleCountService(
	repo repository.CycleCountRepository,
	inventoryRepo repository.InventoryRepository,
	warehouseRepo repository.WarehouseRepository,
	txManager repository.TxManager,
) *CycleCountService {
	return &CycleCountService{
		repo:          repo,
		inventoryRepo: inventoryRepo,
		warehouseRepo: warehouseRepo,
		txManager:     txManager,
	}
}

// ── Input Types ──────────────────────────────────────────────────────────────────

// StartCycleCountInput is the input for starting a new cycle count.
type StartCycleCountInput struct {
	WarehouseID uuid.UUID  `json:"warehouse_id"`
	LocationID  *uuid.UUID `json:"location_id,omitempty"`
	ZoneID      *uuid.UUID `json:"zone_id,omitempty"`
	CountedBy   string     `json:"counted_by,omitempty"`
	Notes       string     `json:"notes,omitempty"`
}

// Validate checks the input for business rule violations.
func (in *StartCycleCountInput) Validate() error {
	if in.WarehouseID == uuid.Nil {
		return pkgerrors.NewInvalidInput("warehouse_id is required")
	}
	if in.LocationID == nil && in.ZoneID == nil {
		return pkgerrors.NewInvalidInput("either location_id or zone_id is required")
	}
	return nil
}

// SubmitLineInput is the input for submitting a single count line.
type SubmitLineInput struct {
	LineID     uuid.UUID `json:"line_id"`
	CountedQty float64   `json:"counted_qty"`
}

// Validate checks the input.
func (in *SubmitLineInput) Validate() error {
	if in.LineID == uuid.Nil {
		return pkgerrors.NewInvalidInput("line_id is required")
	}
	if in.CountedQty < 0 {
		return pkgerrors.NewInvalidInput("counted_qty must be >= 0")
	}
	return nil
}

// FinalizeCountInput is the input for finalizing a cycle count.
type FinalizeCountInput struct {
	Notes string `json:"notes,omitempty"`
}

// ApproveCountInput is the input for approving a cycle count.
type ApproveCountInput struct {
	ApprovedBy string `json:"approved_by,omitempty"`
	Action     string `json:"action"` // "approve" or "adjust"
}

// Validate checks the input.
func (in *ApproveCountInput) Validate() error {
	if in.Action != "approve" && in.Action != "adjust" {
		return pkgerrors.NewInvalidInput("action must be 'approve' or 'adjust'")
	}
	return nil
}

// ── Service Methods ──────────────────────────────────────────────────────────────

// StartCycleCount creates a new cycle count with lines for each inventory record
// in the specified location or zone.
func (s *CycleCountService) StartCycleCount(ctx context.Context, input StartCycleCountInput) (*domain.CycleCount, []*domain.CycleCountLine, error) {
	if err := input.Validate(); err != nil {
		return nil, nil, err
	}

	// Fetch current inventory for the location or zone.
	inventories, err := s.inventoryRepo.QueryInventory(ctx, repository.InventoryFilter{
		WarehouseID: input.WarehouseID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("cycle count service: start: query inventory: %w", err)
	}

	// Filter inventories by location or zone.
	var matched []*domain.Inventory
	if input.LocationID != nil {
		for _, inv := range inventories {
			if inv.LocationID == *input.LocationID {
				matched = append(matched, inv)
			}
		}
	} else if input.ZoneID != nil {
		// For zone-based counts, we need to look up locations in the zone.
		// Then match inventory records for those locations.
		locations, err := s.warehouseRepo.ListLocationsByZone(ctx, *input.ZoneID, 10000, 0)
		if err != nil {
			return nil, nil, fmt.Errorf("cycle count service: start: list locations in zone: %w", err)
		}
		locSet := make(map[uuid.UUID]bool, len(locations))
		for _, loc := range locations {
			locSet[loc.ID] = true
		}
		for _, inv := range inventories {
			if locSet[inv.LocationID] {
				matched = append(matched, inv)
			}
		}
	}

	if len(matched) == 0 {
		return nil, nil, pkgerrors.NewInvalidInput("no inventory found in the specified location/zone to count")
	}

	// Create the cycle count.
	cc := &domain.CycleCount{
		CountNo:     generateCycleCountNo(),
		WarehouseID: input.WarehouseID,
		LocationID:  input.LocationID,
		ZoneID:      input.ZoneID,
		Status:      domain.CycleCountStatusDraft,
		CountedBy:   input.CountedBy,
		Notes:       input.Notes,
		TotalLines:  len(matched),
	}

	// Create lines within a transaction.
	var lines []*domain.CycleCountLine
	doWrites := func(ctx context.Context) error {
		if err := s.repo.CreateCycleCount(ctx, cc); err != nil {
			return fmt.Errorf("create cycle count: %w", err)
		}

		for _, inv := range matched {
			lines = append(lines, &domain.CycleCountLine{
				CycleCountID: cc.ID,
				SKUID:        inv.SKUID,
				LocationID:   inv.LocationID,
				BatchNo:      inv.BatchNo,
				SystemQty:    inv.Qty,
			})
		}

		if err := s.repo.CreateCycleCountLines(ctx, lines); err != nil {
			return fmt.Errorf("create cycle count lines: %w", err)
		}

		// Transition to in_progress.
		if err := s.repo.UpdateCycleCountStatus(ctx, cc.ID, domain.CycleCountStatusInProgress); err != nil {
			return fmt.Errorf("start cycle count: %w", err)
		}
		return nil
	}

	if err := s.txManager.WithTx(ctx, doWrites); err != nil {
		return nil, nil, fmt.Errorf("cycle count service: start: %w", err)
	}

	cc.Status = domain.CycleCountStatusInProgress
	now := time.Now()
	cc.StartedAt = &now

	return cc, lines, nil
}

// GetCycleCount retrieves a cycle count by ID.
func (s *CycleCountService) GetCycleCount(ctx context.Context, id uuid.UUID) (*domain.CycleCount, []*domain.CycleCountLine, error) {
	cc, err := s.repo.GetCycleCount(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("cycle count service: get %s: %w", id, err)
	}

	lines, err := s.repo.GetCycleCountLines(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("cycle count service: get lines %s: %w", id, err)
	}

	return cc, lines, nil
}

// ListCycleCounts returns cycle counts matching the specified filter.
func (s *CycleCountService) ListCycleCounts(ctx context.Context, filter repository.CycleCountFilter) ([]*domain.CycleCount, int, error) {
	counts, err := s.repo.ListCycleCounts(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("cycle count service: list: %w", err)
	}

	total, err := s.repo.CountCycleCounts(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("cycle count service: count: %w", err)
	}

	return counts, total, nil
}

// SubmitLine records a counted quantity for a single line.
func (s *CycleCountService) SubmitLine(ctx context.Context, cycleCountID uuid.UUID, input SubmitLineInput) (*domain.CycleCountLine, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Verify the cycle count is in_progress.
	cc, err := s.repo.GetCycleCount(ctx, cycleCountID)
	if err != nil {
		return nil, fmt.Errorf("cycle count service: submit line: %w", err)
	}
	if cc.Status != domain.CycleCountStatusInProgress {
		return nil, pkgerrors.NewInvalidInput(
			fmt.Sprintf("can only submit lines for counts in progress (current: %s)", cc.Status))
	}

	// Get the existing lines to find the one being updated.
	lines, err := s.repo.GetCycleCountLines(ctx, cycleCountID)
	if err != nil {
		return nil, fmt.Errorf("cycle count service: get lines: %w", err)
	}

	var targetLine *domain.CycleCountLine
	for _, l := range lines {
		if l.ID == input.LineID {
			targetLine = l
			break
		}
	}
	if targetLine == nil {
		return nil, pkgerrors.NewInvalidInput(
			fmt.Sprintf("line %s not found in cycle count %s", input.LineID, cycleCountID))
	}

	// Calculate variance.
	variance := input.CountedQty - targetLine.SystemQty

	// Update the line.
	if err := s.repo.UpdateCycleCountLine(ctx, input.LineID, input.CountedQty, variance); err != nil {
		return nil, fmt.Errorf("cycle count service: update line: %w", err)
	}

	// Return updated line.
	updatedLines, err := s.repo.GetCycleCountLines(ctx, cycleCountID)
	if err != nil {
		return nil, fmt.Errorf("cycle count service: re-fetch lines: %w", err)
	}
	for _, l := range updatedLines {
		if l.ID == input.LineID {
			return l, nil
		}
	}
	return nil, fmt.Errorf("cycle count service: line %s disappeared after update", input.LineID)
}

// FinalizeCount completes the counting phase and moves the cycle count to pending_review.
func (s *CycleCountService) FinalizeCount(ctx context.Context, id uuid.UUID, input FinalizeCountInput) (*domain.CycleCount, error) {
	cc, err := s.repo.GetCycleCount(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("cycle count service: finalize: %w", err)
	}

	if cc.Status != domain.CycleCountStatusInProgress {
		return nil, pkgerrors.NewInvalidInput(
			fmt.Sprintf("can only finalize counts in progress (current: %s)", cc.Status))
	}

	// Count total lines and matched lines (variance == 0).
	lines, err := s.repo.GetCycleCountLines(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("cycle count service: get lines: %w", err)
	}

	totalLines := len(lines)
	matchedLines := 0
	for _, l := range lines {
		if l.Variance != nil && *l.Variance == 0 {
			matchedLines++
		}
	}

	if input.Notes != "" {
		cc.Notes = input.Notes
	}

	if err := s.repo.FinalizeCycleCount(ctx, id, totalLines, matchedLines); err != nil {
		return nil, fmt.Errorf("cycle count service: finalize: %w", err)
	}

	cc.Status = domain.CycleCountStatusPendingReview
	cc.TotalLines = totalLines
	cc.MatchedLines = matchedLines
	return cc, nil
}

// ApproveCount approves a cycle count and applies inventory adjustments for variances.
// When action is "approve", inventory is adjusted to match the counted quantities.
// When action is "adjust", the count is approved without inventory adjustment.
func (s *CycleCountService) ApproveCount(ctx context.Context, id uuid.UUID, input ApproveCountInput) (*domain.CycleCount, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	cc, err := s.repo.GetCycleCount(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("cycle count service: approve: %w", err)
	}

	if cc.Status != domain.CycleCountStatusPendingReview {
		return nil, pkgerrors.NewInvalidInput(
			fmt.Sprintf("can only approve counts in pending_review (current: %s)", cc.Status))
	}

	lines, err := s.repo.GetCycleCountLines(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("cycle count service: get lines: %w", err)
	}

	targetStatus := domain.CycleCountStatusApproved
	if input.Action == "adjust" {
		targetStatus = domain.CycleCountStatusAdjusted
	}

	doWrites := func(ctx context.Context) error {
		// For "approve", apply inventory adjustments for variances.
		if input.Action == "approve" {
			for _, line := range lines {
				if line.Variance == nil || *line.Variance == 0 {
					continue
				}
				// Find the inventory record and adjust it.
				inv, err := s.inventoryRepo.GetInventoryAtLocation(ctx, line.SKUID, line.LocationID, line.BatchNo)
				if err != nil {
					// Inventory not found — skip (may have been moved since count started).
					continue
				}
				// Adjust to the counted quantity.
				newQty := line.CountedQty
				if newQty == nil {
					continue
				}
				deltaQty := *newQty - inv.Qty
				if deltaQty == 0 {
					continue
				}
				if err := s.inventoryRepo.UpdateInventoryQty(ctx, inv.ID, deltaQty, 0); err != nil {
					return fmt.Errorf("adjust inventory for line %s: %w", line.ID, err)
				}
				// Record transaction.
				tx := &domain.InventoryTransaction{
					InventoryID:   inv.ID,
					SKUID:         line.SKUID,
					LocationID:    line.LocationID,
					Type:          domain.InventoryTxAdjustment,
					DeltaQty:      deltaQty,
					ResultingQty:  *newQty,
					ReferenceType: "cycle_count",
					ReferenceID:   id,
				}
				if err := s.inventoryRepo.CreateTransaction(ctx, tx); err != nil {
					return fmt.Errorf("create adjustment transaction for line %s: %w", line.ID, err)
				}
			}
		}

		if err := s.repo.ApproveCycleCount(ctx, id, input.ApprovedBy); err != nil {
			return fmt.Errorf("approve cycle count: %w", err)
		}
		return nil
	}

	if err := s.txManager.WithTx(ctx, doWrites); err != nil {
		return nil, fmt.Errorf("cycle count service: approve: %w", err)
	}

	cc.Status = targetStatus
	if input.ApprovedBy != "" {
		cc.ApprovedBy = input.ApprovedBy
	}
	now := time.Now()
	cc.ApprovedAt = &now
	return cc, nil
}

// CancelCycleCount cancels a cycle count.
func (s *CycleCountService) CancelCycleCount(ctx context.Context, id uuid.UUID) (*domain.CycleCount, error) {
	cc, err := s.repo.GetCycleCount(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("cycle count service: cancel: %w", err)
	}

	if cc.IsTerminal() {
		return nil, pkgerrors.NewInvalidInput(
			fmt.Sprintf("cannot cancel a terminal cycle count (current: %s)", cc.Status))
	}

	if err := s.repo.UpdateCycleCountStatus(ctx, id, domain.CycleCountStatusCancelled); err != nil {
		return nil, fmt.Errorf("cycle count service: cancel: %w", err)
	}

	cc.Status = domain.CycleCountStatusCancelled
	return cc, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────────

func generateCycleCountNo() string {
	now := time.Now()
	return fmt.Sprintf("CC-%s-%06d", now.Format("20060102"), now.UnixMilli()%1000000)
}
