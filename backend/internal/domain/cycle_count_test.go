package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestCycleCount_StateMachine(t *testing.T) {
	cc := &CycleCount{
		ID:     uuid.New(),
		Status: CycleCountStatusDraft,
	}

	t.Run("draft → in_progress", func(t *testing.T) {
		if !cc.CanTransitionTo(CycleCountStatusInProgress) {
			t.Error("draft should transition to in_progress")
		}
	})

	t.Run("draft → pending_review (invalid)", func(t *testing.T) {
		if cc.CanTransitionTo(CycleCountStatusPendingReview) {
			t.Error("draft should NOT transition directly to pending_review")
		}
	})

	t.Run("draft → cancelled", func(t *testing.T) {
		if !cc.CanTransitionTo(CycleCountStatusCancelled) {
			t.Error("draft should allow cancellation")
		}
	})

	t.Run("can be started", func(t *testing.T) {
		if !cc.CanBeStarted() {
			t.Error("draft should be startable")
		}
	})

	// Transition to in_progress
	cc.Status = CycleCountStatusInProgress

	t.Run("in_progress → pending_review", func(t *testing.T) {
		if !cc.CanTransitionTo(CycleCountStatusPendingReview) {
			t.Error("in_progress should transition to pending_review")
		}
	})

	t.Run("in_progress → approved (invalid)", func(t *testing.T) {
		if cc.CanTransitionTo(CycleCountStatusApproved) {
			t.Error("in_progress should NOT transition directly to approved")
		}
	})

	t.Run("can be submitted", func(t *testing.T) {
		if !cc.CanBeSubmitted() {
			t.Error("in_progress should be submittable")
		}
	})

	// Transition to pending_review
	cc.Status = CycleCountStatusPendingReview

	t.Run("pending_review → approved", func(t *testing.T) {
		if !cc.CanTransitionTo(CycleCountStatusApproved) {
			t.Error("pending_review should transition to approved")
		}
	})

	t.Run("pending_review → adjusted", func(t *testing.T) {
		if !cc.CanTransitionTo(CycleCountStatusAdjusted) {
			t.Error("pending_review should transition to adjusted")
		}
	})

	t.Run("can be reviewed", func(t *testing.T) {
		if !cc.CanBeReviewed() {
			t.Error("pending_review should be reviewable")
		}
	})

	// Transition to approved (terminal)
	cc.Status = CycleCountStatusApproved

	t.Run("approved → in_progress (invalid - terminal)", func(t *testing.T) {
		if cc.CanTransitionTo(CycleCountStatusInProgress) {
			t.Error("approved should be terminal, no further transitions")
		}
	})

	t.Run("is terminal", func(t *testing.T) {
		if !cc.IsTerminal() {
			t.Error("approved should be terminal")
		}
	})
}

func TestCycleCount_HasVariances(t *testing.T) {
	cc := &CycleCount{ID: uuid.New()}

	t.Run("no variance", func(t *testing.T) {
		lines := []*CycleCountLine{
			{SKUID: uuid.New(), SystemQty: 10},
			{SKUID: uuid.New(), SystemQty: 5},
		}
		// Set counted quantities matching system (zero variance)
		qty10 := 10.0
		lines[0].CountedQty = &qty10
		v0 := 0.0
		lines[0].Variance = &v0

		qty5 := 5.0
		lines[1].CountedQty = &qty5
		v02 := 0.0
		lines[1].Variance = &v02

		if cc.HasVariances(lines) {
			t.Error("should not have variances when counted matches system")
		}
	})

	t.Run("has variance", func(t *testing.T) {
		lines := []*CycleCountLine{
			{SKUID: uuid.New(), SystemQty: 10},
		}
		qty8 := 8.0
		lines[0].CountedQty = &qty8
		v := -2.0
		lines[0].Variance = &v

		if !cc.HasVariances(lines) {
			t.Error("should detect variance when counted differs from system")
		}
	})
}

func TestCycleCountLine_SetCountedQty(t *testing.T) {
	line := &CycleCountLine{
		ID:        uuid.New(),
		SystemQty: 100,
	}

	t.Run("counted matches system", func(t *testing.T) {
		line.SetCountedQty(100)
		if !line.IsCounted() {
			t.Error("line should be marked as counted")
		}
		if line.Status != CycleCountLineStatusCounted {
			t.Errorf("expected status counted, got %s", line.Status)
		}
		if line.CountedQty == nil || *line.CountedQty != 100 {
			t.Error("counted_qty should be 100")
		}
		if line.Variance == nil || *line.Variance != 0 {
			t.Errorf("variance should be 0, got %v", line.Variance)
		}
		if line.CountedAt == nil {
			t.Error("counted_at should be set")
		}
	})

	t.Run("counted less than system (negative variance)", func(t *testing.T) {
		line2 := &CycleCountLine{ID: uuid.New(), SystemQty: 50}
		line2.SetCountedQty(45)
		if line2.Variance == nil || *line2.Variance != -5 {
			t.Errorf("variance should be -5, got %v", line2.Variance)
		}
	})

	t.Run("counted more than system (positive variance)", func(t *testing.T) {
		line3 := &CycleCountLine{ID: uuid.New(), SystemQty: 50}
		line3.SetCountedQty(55)
		if line3.Variance == nil || *line3.Variance != 5 {
			t.Errorf("variance should be 5, got %v", line3.Variance)
		}
	})
}

func TestCycleCount_AdditionalTransitions(t *testing.T) {
	t.Run("draft → draft (same - no transition)", func(t *testing.T) {
		cc := &CycleCount{ID: uuid.New(), Status: CycleCountStatusDraft}
		if cc.CanTransitionTo(CycleCountStatusDraft) {
			t.Error("should not allow transition to same status")
		}
	})

	t.Run("in_progress → draft (invalid - backward)", func(t *testing.T) {
		cc := &CycleCount{ID: uuid.New(), Status: CycleCountStatusInProgress}
		if cc.CanTransitionTo(CycleCountStatusDraft) {
			t.Error("in_progress should NOT transition backward to draft")
		}
	})

	t.Run("pending_review → in_progress (invalid - backward)", func(t *testing.T) {
		cc := &CycleCount{ID: uuid.New(), Status: CycleCountStatusPendingReview}
		if cc.CanTransitionTo(CycleCountStatusInProgress) {
			t.Error("pending_review should NOT transition backward")
		}
	})

	t.Run("adjusted is terminal", func(t *testing.T) {
		cc := &CycleCount{ID: uuid.New(), Status: CycleCountStatusAdjusted}
		if !cc.IsTerminal() {
			t.Error("adjusted should be terminal")
		}
	})

	t.Run("cancelled is terminal", func(t *testing.T) {
		cc := &CycleCount{ID: uuid.New(), Status: CycleCountStatusCancelled}
		if !cc.IsTerminal() {
			t.Error("cancelled should be terminal")
		}
	})
}
