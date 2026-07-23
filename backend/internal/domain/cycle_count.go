package domain

import (
	"time"

	"github.com/google/uuid"
)

// CycleCount represents a physical inventory count for a specific warehouse area.
// It tracks the lifecycle from creation → in_progress → pending_review → approved/adjusted.
type CycleCount struct {
	ID           uuid.UUID        `json:"id"`
	CountNo      string           `json:"count_no"`      // e.g. "CC-20260724-001"
	WarehouseID  uuid.UUID        `json:"warehouse_id"`
	LocationID   *uuid.UUID       `json:"location_id,omitempty"` // Specific location or nil for zone
	ZoneID       *uuid.UUID       `json:"zone_id,omitempty"`     // Specific zone or nil for location
	Status       CycleCountStatus `json:"status"`
	CountedBy    string           `json:"counted_by,omitempty"` // Operator ID
	Notes        string           `json:"notes,omitempty"`
	TotalLines   int              `json:"total_lines"`   // Total count lines
	MatchedLines int              `json:"matched_lines"` // Lines with no variance
	CreatedAt    time.Time        `json:"created_at"`
	StartedAt    *time.Time       `json:"started_at,omitempty"`
	CompletedAt  *time.Time       `json:"completed_at,omitempty"`
	ApprovedAt   *time.Time       `json:"approved_at,omitempty"`
	ApprovedBy   string           `json:"approved_by,omitempty"`
}

// CycleCountLine represents a single counted item (SKU/batch at a location) within a cycle count.
type CycleCountLine struct {
	ID             uuid.UUID           `json:"id"`
	CycleCountID   uuid.UUID           `json:"cycle_count_id"`
	SKUID          uuid.UUID           `json:"sku_id"`
	LocationID     uuid.UUID           `json:"location_id"`
	BatchNo        string              `json:"batch_no,omitempty"`
	SystemQty      float64             `json:"system_qty"`   // What the system says is there
	CountedQty     *float64            `json:"counted_qty,omitempty"` // What the operator counted
	Variance       *float64            `json:"variance,omitempty"`    // counted_qty - system_qty
	Status         CycleCountLineStatus `json:"status"`
	CountedAt      *time.Time          `json:"counted_at,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
}

// CycleCountStatus represents the lifecycle state of a cycle count.
type CycleCountStatus string

const (
	CycleCountStatusDraft        CycleCountStatus = "draft"         // Created but not started
	CycleCountStatusInProgress   CycleCountStatus = "in_progress"   // Operator is counting
	CycleCountStatusPendingReview CycleCountStatus = "pending_review" // Awaiting supervisor review
	CycleCountStatusApproved     CycleCountStatus = "approved"      // Variance approved, inventory updated
	CycleCountStatusAdjusted     CycleCountStatus = "adjusted"      // Adjusted (override) applied
	CycleCountStatusCancelled    CycleCountStatus = "cancelled"     // Cancelled
)

// CycleCountLineStatus represents the state of an individual count line.
type CycleCountLineStatus string

const (
	CycleCountLineStatusPending  CycleCountLineStatus = "pending"  // Not yet counted
	CycleCountLineStatusCounted  CycleCountLineStatus = "counted"  // Operator has entered count
	CycleCountLineStatusReviewed CycleCountLineStatus = "reviewed" // Reviewed by supervisor
)

// ── CycleCount State Machine ─────────────────────────────────────────────────

// IsTerminal returns true if the cycle count is in a terminal state.
func (cc *CycleCount) IsTerminal() bool {
	return cc.Status == CycleCountStatusApproved ||
		cc.Status == CycleCountStatusAdjusted ||
		cc.Status == CycleCountStatusCancelled
}

// CanTransitionTo checks whether the cycle count can transition from its current
// status to the target status.
//
// Valid transitions:
//
//	draft          → in_progress, cancelled
//	in_progress    → pending_review, cancelled
//	pending_review → approved, adjusted, cancelled
//	approved       → (terminal)
//	adjusted       → (terminal)
//	cancelled      → (terminal)
func (cc *CycleCount) CanTransitionTo(target CycleCountStatus) bool {
	if cc.Status == target {
		return false
	}
	if cc.IsTerminal() {
		return false
	}
	// Any non-terminal status can be cancelled.
	if target == CycleCountStatusCancelled {
		return true
	}

	switch cc.Status {
	case CycleCountStatusDraft:
		return target == CycleCountStatusInProgress
	case CycleCountStatusInProgress:
		return target == CycleCountStatusPendingReview
	case CycleCountStatusPendingReview:
		return target == CycleCountStatusApproved || target == CycleCountStatusAdjusted
	default:
		return false
	}
}

// CanBeStarted returns true if the cycle count can be started.
func (cc *CycleCount) CanBeStarted() bool {
	return cc.Status == CycleCountStatusDraft
}

// CanBeSubmitted returns true if the cycle count can be submitted for review.
func (cc *CycleCount) CanBeSubmitted() bool {
	return cc.Status == CycleCountStatusInProgress
}

// CanBeReviewed returns true if the cycle count is ready for approval.
func (cc *CycleCount) CanBeReviewed() bool {
	return cc.Status == CycleCountStatusPendingReview
}

// HasVariances returns true if any line has a non-zero variance.
func (cc *CycleCount) HasVariances(lines []*CycleCountLine) bool {
	for _, l := range lines {
		if l.Variance != nil && *l.Variance != 0 {
			return true
		}
	}
	return false
}

// ── CycleCountLine Methods ───────────────────────────────────────────────────

// IsCounted returns true if the line has been counted.
func (l *CycleCountLine) IsCounted() bool {
	return l.Status == CycleCountLineStatusCounted || l.Status == CycleCountLineStatusReviewed
}

// SetCountedQty sets the counted quantity and computes variance.
func (l *CycleCountLine) SetCountedQty(qty float64) {
	l.CountedQty = &qty
	v := qty - l.SystemQty
	l.Variance = &v
	l.Status = CycleCountLineStatusCounted
	now := time.Now()
	l.CountedAt = &now
}
