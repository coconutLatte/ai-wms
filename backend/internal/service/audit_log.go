// Package service implements business logic orchestration for the WMS domain.
package service

import (
	"context"
	"fmt"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// AuditLogService orchestrates business logic for audit log queries.
// Audit logs are read-only — creation is handled directly at the repository
// layer by middleware and other services.
type AuditLogService struct {
	repo repository.UserRepository
}

// NewAuditLogService creates a new AuditLogService.
func NewAuditLogService(repo repository.UserRepository) *AuditLogService {
	return &AuditLogService{repo: repo}
}

// ListAuditLogs returns paginated audit logs matching the specified filter,
// ordered by most recent first.
func (s *AuditLogService) ListAuditLogs(ctx context.Context, filter repository.AuditLogFilter) ([]*domain.AuditLog, int, error) {
	logs, err := s.repo.ListAuditLogs(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("audit log service: list: %w", err)
	}

	total, err := s.repo.CountAuditLogs(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("audit log service: count: %w", err)
	}

	return logs, total, nil
}
