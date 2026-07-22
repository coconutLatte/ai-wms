package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// mockUserRepo implements repository.UserRepository for testing UserService.
type mockUserRepo struct {
	users map[uuid.UUID]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[uuid.UUID]*domain.User)}
}

func (m *mockUserRepo) addUser(u *domain.User) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.Status == "" {
		u.Status = domain.UserStatusActive
	}
	if u.RoleIDs == nil {
		u.RoleIDs = []uuid.UUID{}
	}
	clone := *u
	m.users[u.ID] = &clone
}

// ── User ─────────────────────────────────────────────────────

func (m *mockUserRepo) CreateUser(ctx context.Context, u *domain.User) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	if u.Status == "" {
		u.Status = domain.UserStatusActive
	}
	if u.RoleIDs == nil {
		u.RoleIDs = []uuid.UUID{}
	}
	clone := *u
	m.users[u.ID] = &clone
	return nil
}

func (m *mockUserRepo) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("user", id.String())
	}
	clone := *u
	return &clone, nil
}

func (m *mockUserRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			clone := *u
			return &clone, nil
		}
	}
	return nil, pkgerrors.NewNotFound("user", username)
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			clone := *u
			return &clone, nil
		}
	}
	return nil, pkgerrors.NewNotFound("user", email)
}

func (m *mockUserRepo) ListUsers(ctx context.Context, filter repository.UserFilter) ([]*domain.User, error) {
	var result []*domain.User
	for _, u := range m.users {
		if filter.Status != "" && u.Status != filter.Status {
			continue
		}
		clone := *u
		result = append(result, &clone)
	}

	// Apply offset + limit for realistic pagination simulation.
	if filter.Offset > 0 {
		if filter.Offset >= len(result) {
			return nil, nil
		}
		result = result[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}

	return result, nil
}

func (m *mockUserRepo) UpdateUser(ctx context.Context, u *domain.User) error {
	existing, ok := m.users[u.ID]
	if !ok {
		return pkgerrors.NewNotFound("user", u.ID.String())
	}
	existing.Email = u.Email
	existing.DisplayName = u.DisplayName
	existing.RoleIDs = u.RoleIDs
	existing.Status = u.Status
	existing.UpdatedAt = time.Now()
	return nil
}

func (m *mockUserRepo) UpdateUserStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error {
	u, ok := m.users[id]
	if !ok {
		return pkgerrors.NewNotFound("user", id.String())
	}
	u.Status = status
	u.UpdatedAt = time.Now()
	return nil
}

func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error {
	u, ok := m.users[id]
	if !ok {
		return pkgerrors.NewNotFound("user", id.String())
	}
	u.LastLogin = &t
	u.UpdatedAt = time.Now()
	return nil
}

func (m *mockUserRepo) CountUsers(ctx context.Context, filter repository.UserFilter) (int, error) {
	count := 0
	for _, u := range m.users {
		if filter.Status != "" && u.Status != filter.Status {
			continue
		}
		count++
	}
	return count, nil
}

// ── Role (stubs) ──────────────────────────────────────────────

func (m *mockUserRepo) CreateRole(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockUserRepo) GetRole(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return nil, nil
}
func (m *mockUserRepo) ListRoles(ctx context.Context) ([]*domain.Role, error) { return nil, nil }
func (m *mockUserRepo) UpdateRole(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockUserRepo) CountRoles(ctx context.Context) (int, error)            { return 0, nil }

// ── AuditLog (stubs) ──────────────────────────────────────────

func (m *mockUserRepo) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	return nil
}
func (m *mockUserRepo) ListAuditLogs(ctx context.Context, filter repository.AuditLogFilter) ([]*domain.AuditLog, error) {
	return nil, nil
}
func (m *mockUserRepo) CountAuditLogs(ctx context.Context, filter repository.AuditLogFilter) (int, error) {
	return 0, nil
}

// ── Tests: ListUsers ──────────────────────────────────────────

func TestUserService_ListUsers(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	// Seed users.
	repo.addUser(&domain.User{
		Username: "alice", Email: "alice@example.com",
		Status: domain.UserStatusActive,
	})
	repo.addUser(&domain.User{
		Username: "bob", Email: "bob@example.com",
		Status: domain.UserStatusInactive,
	})
	repo.addUser(&domain.User{
		Username: "carol", Email: "carol@example.com",
		Status: domain.UserStatusActive,
	})

	// All users.
	users, total, err := svc.ListUsers(ctx, repository.UserFilter{})
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}
	if total != 3 {
		t.Errorf("expected total=3, got %d", total)
	}

	// Filter by status.
	users, total, err = svc.ListUsers(ctx, repository.UserFilter{Status: domain.UserStatusActive})
	if err != nil {
		t.Fatalf("ListUsers by status failed: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 active users, got %d", len(users))
	}
	if total != 2 {
		t.Errorf("expected total=2 for active, got %d", total)
	}

	// Filter by status with no matches.
	users, total, err = svc.ListUsers(ctx, repository.UserFilter{Status: domain.UserStatusLocked})
	if err != nil {
		t.Fatalf("ListUsers by locked status failed: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 locked users, got %d", len(users))
	}
	if total != 0 {
		t.Errorf("expected total=0 for locked, got %d", total)
	}
}

func TestUserService_ListUsers_Pagination(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	// Seed 10 users.
	for i := range 10 {
		repo.addUser(&domain.User{
			Username: "user-" + string(rune('0'+i%10)),
			Email:    "user@example.com",
			Status:   domain.UserStatusActive,
		})
	}

	// Page 1, size 5.
	users, total, err := svc.ListUsers(ctx, repository.UserFilter{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("ListUsers page 1 failed: %v", err)
	}
	if len(users) != 5 {
		t.Errorf("expected 5 users on page 1, got %d", len(users))
	}
	if total != 10 {
		t.Errorf("expected total=10, got %d", total)
	}

	// Page 2, size 5.
	users, total, err = svc.ListUsers(ctx, repository.UserFilter{Limit: 5, Offset: 5})
	if err != nil {
		t.Fatalf("ListUsers page 2 failed: %v", err)
	}
	if len(users) != 5 {
		t.Errorf("expected 5 users on page 2, got %d", len(users))
	}
	if total != 10 {
		t.Errorf("expected total=10, got %d", total)
	}

	// Page 3 (empty).
	users, total, err = svc.ListUsers(ctx, repository.UserFilter{Limit: 5, Offset: 10})
	if err != nil {
		t.Fatalf("ListUsers page 3 failed: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users on page 3, got %d", len(users))
	}
	if total != 10 {
		t.Errorf("expected total=10 on page 3, got %d", total)
	}
}

func TestUserService_ListUsers_EmptyRepo(t *testing.T) {
	ctx := context.Background()
	svc := NewUserService(newMockUserRepo())

	users, total, err := svc.ListUsers(ctx, repository.UserFilter{})
	if err != nil {
		t.Fatalf("ListUsers empty repo failed: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
	if total != 0 {
		t.Errorf("expected total=0, got %d", total)
	}
}

// ── Tests: CreateUser ─────────────────────────────────────────

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	roleID := uuid.New()
	input := CreateUserInput{
		Username:    "newuser",
		Email:       "newuser@example.com",
		Password:    "secure123",
		DisplayName: "New User",
		RoleIDs:     []uuid.UUID{roleID},
	}

	user, err := svc.CreateUser(ctx, input)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.Username != "newuser" {
		t.Errorf("expected username 'newuser', got %q", user.Username)
	}
	if user.Email != "newuser@example.com" {
		t.Errorf("expected email 'newuser@example.com', got %q", user.Email)
	}
	if user.PasswordHash == "" {
		t.Error("expected password hash to be set")
	}
	if user.PasswordHash == "secure123" {
		t.Error("password should be hashed, not plaintext")
	}
	if user.Status != domain.UserStatusActive {
		t.Errorf("expected status 'active', got %s", user.Status)
	}
	if len(user.RoleIDs) != 1 || user.RoleIDs[0] != roleID {
		t.Errorf("expected role IDs to contain %s", roleID)
	}
	if user.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	// Verify retrievable.
	retrieved, err := svc.GetUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUser after create failed: %v", err)
	}
	if retrieved.ID != user.ID {
		t.Error("retrieved user ID mismatch")
	}
}

func TestUserService_CreateUser_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewUserService(newMockUserRepo())

	tests := []struct {
		name  string
		input CreateUserInput
	}{
		{"empty username", CreateUserInput{Email: "a@b.com", Password: "123456"}},
		{"empty email", CreateUserInput{Username: "test", Password: "123456"}},
		{"empty password", CreateUserInput{Username: "test", Email: "a@b.com"}},
		{"short password", CreateUserInput{Username: "test", Email: "a@b.com", Password: "12345"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateUser(ctx, tt.input)
			if err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

func TestUserService_CreateUser_DuplicateUsername(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	// Create first user.
	_, err := svc.CreateUser(ctx, CreateUserInput{
		Username: "dupe", Email: "first@example.com", Password: "123456",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Try to create with same username.
	_, err = svc.CreateUser(ctx, CreateUserInput{
		Username: "dupe", Email: "second@example.com", Password: "123456",
	})
	if err == nil {
		t.Error("expected duplicate username error")
	}
	if !pkgerrors.Is(err, pkgerrors.ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestUserService_CreateUser_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	// Create first user.
	_, err := svc.CreateUser(ctx, CreateUserInput{
		Username: "user1", Email: "dupe@example.com", Password: "123456",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Try to create with same email.
	_, err = svc.CreateUser(ctx, CreateUserInput{
		Username: "user2", Email: "dupe@example.com", Password: "123456",
	})
	if err == nil {
		t.Error("expected duplicate email error")
	}
	if !pkgerrors.Is(err, pkgerrors.ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestUserService_CreateUser_DefaultRoleIDs(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	user, err := svc.CreateUser(ctx, CreateUserInput{
		Username: "noroles", Email: "noroles@example.com", Password: "123456",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.RoleIDs == nil {
		t.Error("expected RoleIDs to be empty slice, not nil")
	}
	if len(user.RoleIDs) != 0 {
		t.Errorf("expected 0 role IDs, got %d", len(user.RoleIDs))
	}
}

// ── Tests: GetUser ────────────────────────────────────────────

func TestUserService_GetUser(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	repo.addUser(&domain.User{
		Username:    "getme",
		Email:       "getme@example.com",
		DisplayName: "Get Me",
		Status:      domain.UserStatusActive,
	})

	// Find the user we just added to get its ID.
	var userID uuid.UUID
	for id := range repo.users {
		userID = id
		break
	}

	user, err := svc.GetUser(ctx, userID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if user.Username != "getme" {
		t.Errorf("expected username 'getme', got %q", user.Username)
	}
	if user.Email != "getme@example.com" {
		t.Errorf("expected email 'getme@example.com', got %q", user.Email)
	}
}

func TestUserService_GetUser_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewUserService(newMockUserRepo())

	_, err := svc.GetUser(ctx, uuid.New())
	if err == nil {
		t.Error("expected error for non-existent user")
	}
	if !pkgerrors.IsNotFound(err) {
		t.Errorf("expected NotFound error, got %v", err)
	}
}

// ── Tests: UpdateUser ─────────────────────────────────────────

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	// Create a user first.
	created, err := svc.CreateUser(ctx, CreateUserInput{
		Username:    "updateme",
		Email:       "old@example.com",
		Password:    "123456",
		DisplayName: "Old Name",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Update email and display name.
	newEmail := "new@example.com"
	newName := "New Name"
	newRoleIDs := []uuid.UUID{uuid.New()}

	updated, err := svc.UpdateUser(ctx, created.ID, UpdateUserInput{
		Email:       &newEmail,
		DisplayName: &newName,
		RoleIDs:     newRoleIDs,
	})
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	if updated.Email != newEmail {
		t.Errorf("expected email %q, got %q", newEmail, updated.Email)
	}
	if updated.DisplayName != newName {
		t.Errorf("expected display_name %q, got %q", newName, updated.DisplayName)
	}
	if len(updated.RoleIDs) != 1 || updated.RoleIDs[0] != newRoleIDs[0] {
		t.Errorf("expected role IDs to be updated")
	}
}

func TestUserService_UpdateUser_Partial(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	created, err := svc.CreateUser(ctx, CreateUserInput{
		Username:    "partial",
		Email:       "partial@example.com",
		Password:    "123456",
		DisplayName: "Partial",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	originalEmail := created.Email

	// Update only display name.
	newName := "Updated Name"
	updated, err := svc.UpdateUser(ctx, created.ID, UpdateUserInput{
		DisplayName: &newName,
	})
	if err != nil {
		t.Fatalf("UpdateUser partial failed: %v", err)
	}

	if updated.DisplayName != newName {
		t.Errorf("expected display_name %q, got %q", newName, updated.DisplayName)
	}
	if updated.Email != originalEmail {
		t.Errorf("expected email unchanged %q, got %q", originalEmail, updated.Email)
	}
}

func TestUserService_UpdateUser_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewUserService(newMockUserRepo())

	newName := "test"
	_, err := svc.UpdateUser(ctx, uuid.New(), UpdateUserInput{DisplayName: &newName})
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestUserService_UpdateUser_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	// Create two users.
	_, err := svc.CreateUser(ctx, CreateUserInput{
		Username: "user1", Email: "user1@example.com", Password: "123456",
	})
	if err != nil {
		t.Fatalf("CreateUser user1 failed: %v", err)
	}
	user2, err := svc.CreateUser(ctx, CreateUserInput{
		Username: "user2", Email: "user2@example.com", Password: "123456",
	})
	if err != nil {
		t.Fatalf("CreateUser user2 failed: %v", err)
	}

	// Try to update user2's email to user1's email.
	newEmail := "user1@example.com"
	_, err = svc.UpdateUser(ctx, user2.ID, UpdateUserInput{Email: &newEmail})
	if err == nil {
		t.Error("expected duplicate email error")
	}
}

func TestUserService_UpdateUser_EmailToSelf(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	created, err := svc.CreateUser(ctx, CreateUserInput{
		Username: "self", Email: "self@example.com", Password: "123456",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Update email to same value (should succeed because it's the same user).
	sameEmail := "self@example.com"
	updated, err := svc.UpdateUser(ctx, created.ID, UpdateUserInput{Email: &sameEmail})
	if err != nil {
		t.Fatalf("UpdateUser same email failed: %v", err)
	}
	if updated.Email != "self@example.com" {
		t.Errorf("expected email unchanged")
	}
}

// ── Tests: UpdateUserStatus ───────────────────────────────────

func TestUserService_UpdateUserStatus(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	created, err := svc.CreateUser(ctx, CreateUserInput{
		Username: "statususer", Email: "status@example.com", Password: "123456",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Active → Inactive
	updated, err := svc.UpdateUserStatus(ctx, created.ID, UpdateUserStatusInput{Status: domain.UserStatusInactive})
	if err != nil {
		t.Fatalf("UpdateUserStatus active→inactive failed: %v", err)
	}
	if updated.Status != domain.UserStatusInactive {
		t.Errorf("expected inactive, got %s", updated.Status)
	}

	// Inactive → Locked
	updated, err = svc.UpdateUserStatus(ctx, created.ID, UpdateUserStatusInput{Status: domain.UserStatusLocked})
	if err != nil {
		t.Fatalf("UpdateUserStatus inactive→locked failed: %v", err)
	}
	if updated.Status != domain.UserStatusLocked {
		t.Errorf("expected locked, got %s", updated.Status)
	}

	// Locked → Active
	updated, err = svc.UpdateUserStatus(ctx, created.ID, UpdateUserStatusInput{Status: domain.UserStatusActive})
	if err != nil {
		t.Fatalf("UpdateUserStatus locked→active failed: %v", err)
	}
	if updated.Status != domain.UserStatusActive {
		t.Errorf("expected active, got %s", updated.Status)
	}
}

func TestUserService_UpdateUserStatus_SameStatus(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	created, err := svc.CreateUser(ctx, CreateUserInput{
		Username: "samestatus", Email: "same@example.com", Password: "123456",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Active → Active should fail (self-transition).
	_, err = svc.UpdateUserStatus(ctx, created.ID, UpdateUserStatusInput{Status: domain.UserStatusActive})
	if err == nil {
		t.Error("expected error for self-transition")
	}
	if !pkgerrors.IsInvalidStatus(err) {
		t.Errorf("expected InvalidStatus error, got %v", err)
	}
}

func TestUserService_UpdateUserStatus_InvalidInput(t *testing.T) {
	ctx := context.Background()
	svc := NewUserService(newMockUserRepo())

	_, err := svc.UpdateUserStatus(ctx, uuid.New(), UpdateUserStatusInput{Status: "invalid_status"})
	if err == nil {
		t.Error("expected validation error for invalid status")
	}
}

func TestUserService_UpdateUserStatus_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewUserService(newMockUserRepo())

	_, err := svc.UpdateUserStatus(ctx, uuid.New(), UpdateUserStatusInput{Status: domain.UserStatusInactive})
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}
