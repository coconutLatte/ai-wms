package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// RoleService orchestrates business logic for role operations.
type RoleService struct {
	repo repository.UserRepository
}

// NewRoleService creates a new RoleService.
func NewRoleService(repo repository.UserRepository) *RoleService {
	return &RoleService{repo: repo}
}

// ── Input Types ──────────────────────────────────────────────────────────────────────────

// CreateRoleInput is the input for creating a new role.
type CreateRoleInput struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Permissions []domain.Permission `json:"permissions,omitempty"`
}

// Validate checks the input for business rule violations.
func (in *CreateRoleInput) Validate() error {
	if in.Name == "" {
		return pkgerrors.NewInvalidInput("name is required")
	}
	return nil
}

// UpdateRoleInput is the input for updating an existing role.
type UpdateRoleInput struct {
	Name        *string             `json:"name,omitempty"`
	Description *string             `json:"description,omitempty"`
	Permissions []domain.Permission `json:"permissions,omitempty"`
}

// ── Service Methods ──────────────────────────────────────────────────────────────────────

// ListRoles returns all roles with pagination support.
func (s *RoleService) ListRoles(ctx context.Context) ([]*domain.Role, int, error) {
	roles, err := s.repo.ListRoles(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("role service: list: %w", err)
	}

	total, err := s.repo.CountRoles(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("role service: count: %w", err)
	}

	if roles == nil {
		roles = []*domain.Role{}
	}

	return roles, total, nil
}

// GetRole retrieves a role by ID.
func (s *RoleService) GetRole(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	role, err := s.repo.GetRole(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("role service: get %s: %w", id, err)
	}
	return role, nil
}

// CreateRole validates input and creates a new role.
func (s *RoleService) CreateRole(ctx context.Context, input CreateRoleInput) (*domain.Role, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Check for duplicate name by listing and scanning.
	roles, err := s.repo.ListRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("role service: check name: %w", err)
	}
	for _, r := range roles {
		if r.Name == input.Name {
			return nil, pkgerrors.NewAlreadyExists("role", input.Name)
		}
	}

	perms := input.Permissions
	if perms == nil {
		perms = []domain.Permission{}
	}

	role := &domain.Role{
		Name:        input.Name,
		Description: input.Description,
		Permissions: perms,
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("role service: create: %w", err)
	}

	return role, nil
}

// UpdateRole updates an existing role's mutable fields.
func (s *RoleService) UpdateRole(ctx context.Context, id uuid.UUID, input UpdateRoleInput) (*domain.Role, error) {
	role, err := s.repo.GetRole(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("role service: update %s: %w", id, err)
	}

	if input.Name != nil {
		role.Name = *input.Name
	}
	if input.Description != nil {
		role.Description = *input.Description
	}
	if input.Permissions != nil {
		role.Permissions = input.Permissions
	}

	if err := s.repo.UpdateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("role service: update %s: %w", id, err)
	}

	// Re-fetch to get persisted state.
	updated, err := s.repo.GetRole(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("role service: re-fetch after update %s: %w", id, err)
	}

	return updated, nil
}

// DeleteRole deletes a role by ID.
func (s *RoleService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteRole(ctx, id); err != nil {
		return fmt.Errorf("role service: delete %s: %w", id, err)
	}
	return nil
}
