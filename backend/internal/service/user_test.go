package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// mockUserRepo implements repository.UserRepository for testing UserService.
type mockUserRepo struct {
	users []*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{}
}

func (m *mockUserRepo) addUser(u *domain.User) {
	m.users = append(m.users, u)
}

// ── User ─────────────────────────────────────────────────────

func (m *mockUserRepo) CreateUser(ctx context.Context, u *domain.User) error { return nil }
func (m *mockUserRepo) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) ListUsers(ctx context.Context, filter repository.UserFilter) ([]*domain.User, error) {
	var result []*domain.User
	for _, u := range m.users {
		if filter.Status != "" && u.Status != filter.Status {
			continue
		}
		result = append(result, u)
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

func (m *mockUserRepo) UpdateUser(ctx context.Context, u *domain.User) error { return nil }
func (m *mockUserRepo) UpdateUserStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error {
	return nil
}
func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error {
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

// ── Tests ─────────────────────────────────────────────────────

func TestUserService_ListUsers(t *testing.T) {
	ctx := context.Background()
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	// Seed users.
	repo.addUser(&domain.User{
		ID: uuid.New(), Username: "alice", Email: "alice@example.com",
		Status: domain.UserStatusActive,
	})
	repo.addUser(&domain.User{
		ID: uuid.New(), Username: "bob", Email: "bob@example.com",
		Status: domain.UserStatusInactive,
	})
	repo.addUser(&domain.User{
		ID: uuid.New(), Username: "carol", Email: "carol@example.com",
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
			ID: uuid.New(), Username: "user-" + string(rune('0'+i%10)),
			Email: "user@example.com", Status: domain.UserStatusActive,
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
