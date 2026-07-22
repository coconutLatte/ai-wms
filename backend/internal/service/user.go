// Package service implements business logic orchestration for the WMS domain.
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// UserService orchestrates business logic for user operations.
type UserService struct {
	repo repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// ── Input Types ──────────────────────────────────────────────────────────────────────────

// CreateUserInput is the input for creating a new user.
type CreateUserInput struct {
	Username    string      `json:"username"`
	Email       string      `json:"email"`
	Password    string      `json:"password"`
	DisplayName string      `json:"display_name,omitempty"`
	RoleIDs     []uuid.UUID `json:"role_ids,omitempty"`
}

// Validate checks the input for business rule violations.
func (in *CreateUserInput) Validate() error {
	if in.Username == "" {
		return pkgerrors.NewInvalidInput("username is required")
	}
	if in.Email == "" {
		return pkgerrors.NewInvalidInput("email is required")
	}
	if in.Password == "" {
		return pkgerrors.NewInvalidInput("password is required")
	}
	if len(in.Password) < 6 {
		return pkgerrors.NewInvalidInput("password must be at least 6 characters")
	}
	return nil
}

// UpdateUserInput is the input for updating an existing user.
type UpdateUserInput struct {
	Email       *string     `json:"email,omitempty"`
	DisplayName *string     `json:"display_name,omitempty"`
	RoleIDs     []uuid.UUID `json:"role_ids,omitempty"`
}

// UpdateUserStatusInput is the input for updating a user's status.
type UpdateUserStatusInput struct {
	Status domain.UserStatus `json:"status"`
}

// Validate checks the input for business rule violations.
func (in *UpdateUserStatusInput) Validate() error {
	if !isValidUserStatus(in.Status) {
		return pkgerrors.NewInvalidInput(fmt.Sprintf("invalid user status: %s", in.Status))
	}
	return nil
}

// ── Service Methods ──────────────────────────────────────────────────────────────────────

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

// CreateUser validates input, hashes the password, and creates a new user.
func (s *UserService) CreateUser(ctx context.Context, input CreateUserInput) (*domain.User, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Check for duplicate username.
	existing, err := s.repo.GetUserByUsername(ctx, input.Username)
	if err != nil && !pkgerrors.IsNotFound(err) {
		return nil, fmt.Errorf("user service: check username: %w", err)
	}
	if existing != nil {
		return nil, pkgerrors.NewAlreadyExists("user", input.Username)
	}

	// Check for duplicate email.
	existing, err = s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil && !pkgerrors.IsNotFound(err) {
		return nil, fmt.Errorf("user service: check email: %w", err)
	}
	if existing != nil {
		return nil, pkgerrors.NewAlreadyExists("user", input.Email)
	}

	// Hash the password.
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("user service: hash password: %w", err)
	}

	roleIDs := input.RoleIDs
	if roleIDs == nil {
		roleIDs = []uuid.UUID{}
	}

	user := &domain.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hash),
		DisplayName:  input.DisplayName,
		RoleIDs:      roleIDs,
		Status:       domain.UserStatusActive,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("user service: create: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by ID.
func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user service: get %s: %w", id, err)
	}
	return user, nil
}

// UpdateUser updates an existing user's mutable fields (email, display name, role IDs).
func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*domain.User, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user service: update %s: %w", id, err)
	}

	if input.Email != nil {
		// Check for duplicate email (unless it's the same user).
		existing, err := s.repo.GetUserByEmail(ctx, *input.Email)
		if err != nil && !pkgerrors.IsNotFound(err) {
			return nil, fmt.Errorf("user service: check email: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, pkgerrors.NewAlreadyExists("user", *input.Email)
		}
		user.Email = *input.Email
	}
	if input.DisplayName != nil {
		user.DisplayName = *input.DisplayName
	}
	if input.RoleIDs != nil {
		user.RoleIDs = input.RoleIDs
	}

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("user service: update %s: %w", id, err)
	}

	// Re-fetch to get updated state.
	updated, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user service: re-fetch after update %s: %w", id, err)
	}

	return updated, nil
}

// UpdateUserStatus validates the state transition and updates the user's status.
func (s *UserService) UpdateUserStatus(ctx context.Context, id uuid.UUID, input UpdateUserStatusInput) (*domain.User, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user service: update status %s: %w", id, err)
	}

	// Validate the state transition.
	if !user.CanTransitionTo(input.Status) {
		return nil, pkgerrors.NewInvalidStatus(string(user.Status), string(input.Status))
	}

	if err := s.repo.UpdateUserStatus(ctx, id, input.Status); err != nil {
		return nil, fmt.Errorf("user service: update status %s: %w", id, err)
	}

	// Re-fetch to get updated state.
	updated, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user service: re-fetch after status update %s: %w", id, err)
	}

	return updated, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────────────────

func isValidUserStatus(s domain.UserStatus) bool {
	switch s {
	case domain.UserStatusActive, domain.UserStatusInactive, domain.UserStatusLocked:
		return true
	}
	return false
}
