package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
)

// CycleCountRepository defines the data access interface for cycle count operations.
type CycleCountRepository interface {
	// CreateCycleCount creates a new cycle count.
	CreateCycleCount(ctx context.Context, cc *domain.CycleCount) error

	// GetCycleCount retrieves a cycle count by ID.
	GetCycleCount(ctx context.Context, id uuid.UUID) (*domain.CycleCount, error)

	// ListCycleCounts returns cycle counts matching the specified filter.
	ListCycleCounts(ctx context.Context, filter CycleCountFilter) ([]*domain.CycleCount, error)

	// CountCycleCounts returns the total count of cycle counts matching the filter.
	CountCycleCounts(ctx context.Context, filter CycleCountFilter) (int, error)

	// UpdateCycleCountStatus transitions a cycle count to a new status.
	UpdateCycleCountStatus(ctx context.Context, id uuid.UUID, status domain.CycleCountStatus) error

	// ApproveCycleCount approves a cycle count with optional approver information.
	ApproveCycleCount(ctx context.Context, id uuid.UUID, approvedBy string) error

	// CreateCycleCountLine creates a new line for a cycle count.
	CreateCycleCountLine(ctx context.Context, line *domain.CycleCountLine) error

	// GetCycleCountLines retrieves all lines for a cycle count.
	GetCycleCountLines(ctx context.Context, cycleCountID uuid.UUID) ([]*domain.CycleCountLine, error)

	// UpdateCycleCountLine updates a count line with the counted quantity and variance.
	UpdateCycleCountLine(ctx context.Context, id uuid.UUID, countedQty float64, variance float64) error

	// FinalizeCycleCount updates line counts and transitions count to pending_review.
	FinalizeCycleCount(ctx context.Context, id uuid.UUID, totalLines, matchedLines int) error

	// CreateCycleCountLines creates multiple lines in a single batch.
	CreateCycleCountLines(ctx context.Context, lines []*domain.CycleCountLine) error
}

// CycleCountFilter defines filter parameters for listing cycle counts.
type CycleCountFilter struct {
	WarehouseID uuid.UUID
	Status      string
	Limit       int
	Offset      int
}
