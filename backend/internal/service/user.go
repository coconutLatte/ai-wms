// Package service implements business logic orchestration for the WMS domain.
package service

import (
	"context"
	"fmt"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// UserService orchestrates business logic for user operations.
type UserService struct {
	repo repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// ListUsers returns paginated users matching the specified filter,
// ordered by most recently created first.
func (s *UserService) ListUsers(ctx context.Context, filter repository.UserFilter) ([]*domain.User, int, error) {
	users, err := s.repo.ListUsers(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("user service: list: %w", err)
	}

	total, err := s.repo.CountUsers(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("user service: count: %w", err)
	}

	return users, total, nil
}
