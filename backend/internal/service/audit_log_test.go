package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// mockAuditLogRepo implements repository.UserRepository for testing AuditLogService.
type mockAuditLogRepo struct {
	logs []*domain.AuditLog
}

func newMockAuditLogRepo() *mockAuditLogRepo {
	return &mockAuditLogRepo{}
}

func (m *mockAuditLogRepo) addLog(l *domain.AuditLog) {
	m.logs = append(m.logs, l)
}

// ── User (stubs) ─────────────────────────────────────────────

func (m *mockAuditLogRepo) CreateUser(ctx context.Context, u *domain.User) error   { return nil }
func (m *mockAuditLogRepo) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}
func (m *mockAuditLogRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return nil, nil
}
func (m *mockAuditLogRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
func (m *mockAuditLogRepo) ListUsers(ctx context.Context, filter repository.UserFilter) ([]*domain.User, error) {
	return nil, nil
}
func (m *mockAuditLogRepo) UpdateUser(ctx context.Context, u *domain.User) error { return nil }
func (m *mockAuditLogRepo) UpdateUserStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error {
	return nil
}
func (m *mockAuditLogRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error {
	return nil
}
func (m *mockAuditLogRepo) CountUsers(ctx context.Context, filter repository.UserFilter) (int, error) {
	return 0, nil
}

// ── Role (stubs) ─────────────────────────────────────────────

func (m *mockAuditLogRepo) CreateRole(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockAuditLogRepo) GetRole(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return nil, nil
}
func (m *mockAuditLogRepo) ListRoles(ctx context.Context) ([]*domain.Role, error) { return nil, nil }
func (m *mockAuditLogRepo) UpdateRole(ctx context.Context, role *domain.Role) error { return nil }
func (m *mockAuditLogRepo) CountRoles(ctx context.Context) (int, error)            { return 0, nil }

// ── AuditLog ─────────────────────────────────────────────────

func (m *mockAuditLogRepo) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	return nil
}

func (m *mockAuditLogRepo) ListAuditLogs(ctx context.Context, filter repository.AuditLogFilter) ([]*domain.AuditLog, error) {
	var result []*domain.AuditLog
	for _, l := range m.logs {
		if filter.UserID != uuid.Nil && l.UserID != filter.UserID {
			continue
		}
		if filter.Action != "" && l.Action != filter.Action {
			continue
		}
		if filter.Resource != "" && l.Resource != filter.Resource {
			continue
		}

		// Apply limit/offset (simplified — mock doesn't do real pagination,
		// it returns all matching records; the service uses CountAuditLogs
		// for pagination metadata so this is fine for testing).
		result = append(result, l)
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

func (m *mockAuditLogRepo) CountAuditLogs(ctx context.Context, filter repository.AuditLogFilter) (int, error) {
	count := 0
	for _, l := range m.logs {
		if filter.UserID != uuid.Nil && l.UserID != filter.UserID {
			continue
		}
		if filter.Action != "" && l.Action != filter.Action {
			continue
		}
		if filter.Resource != "" && l.Resource != filter.Resource {
			continue
		}
		count++
	}
	return count, nil
}

// ── Tests ────────────────────────────────────────────────────

func TestAuditLogService_ListAuditLogs(t *testing.T) {
	ctx := context.Background()
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	uid1 := uuid.New()
	uid2 := uuid.New()

	// Seed logs.
	repo.addLog(&domain.AuditLog{
		ID: uuid.New(), UserID: uid1, Username: "alice",
		Action: "order.create", Resource: "order", ResourceID: "order-1",
	})
	repo.addLog(&domain.AuditLog{
		ID: uuid.New(), UserID: uid1, Username: "alice",
		Action: "inventory.adjust", Resource: "inventory", ResourceID: "inv-1",
	})
	repo.addLog(&domain.AuditLog{
		ID: uuid.New(), UserID: uid2, Username: "bob",
		Action: "order.create", Resource: "order", ResourceID: "order-2",
	})

	// All logs.
	logs, total, err := svc.ListAuditLogs(ctx, repository.AuditLogFilter{})
	if err != nil {
		t.Fatalf("ListAuditLogs failed: %v", err)
	}
	if len(logs) != 3 {
		t.Errorf("expected 3 logs, got %d", len(logs))
	}
	if total != 3 {
		t.Errorf("expected total=3, got %d", total)
	}

	// Filter by UserID.
	logs, total, err = svc.ListAuditLogs(ctx, repository.AuditLogFilter{UserID: uid1})
	if err != nil {
		t.Fatalf("ListAuditLogs by user failed: %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 logs for uid1, got %d", len(logs))
	}
	if total != 2 {
		t.Errorf("expected total=2 for uid1, got %d", total)
	}

	// Filter by Action.
	logs, total, err = svc.ListAuditLogs(ctx, repository.AuditLogFilter{Action: "order.create"})
	if err != nil {
		t.Fatalf("ListAuditLogs by action failed: %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 logs for action=order.create, got %d", len(logs))
	}

	// Filter by Resource.
	logs, total, err = svc.ListAuditLogs(ctx, repository.AuditLogFilter{Resource: "inventory"})
	if err != nil {
		t.Fatalf("ListAuditLogs by resource failed: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log for resource=inventory, got %d", len(logs))
	}

	// Combined filter.
	logs, total, err = svc.ListAuditLogs(ctx, repository.AuditLogFilter{
		UserID: uid2, Action: "order.create", Resource: "order",
	})
	if err != nil {
		t.Fatalf("ListAuditLogs combined filter failed: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log for combined filter, got %d", len(logs))
	}

	// Filter with no matches.
	logs, total, err = svc.ListAuditLogs(ctx, repository.AuditLogFilter{Action: "nonexistent"})
	if err != nil {
		t.Fatalf("ListAuditLogs no-match failed: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs for nonexistent action, got %d", len(logs))
	}
	if total != 0 {
		t.Errorf("expected total=0, got %d", total)
	}
}

func TestAuditLogService_ListAuditLogs_Pagination(t *testing.T) {
	ctx := context.Background()
	repo := newMockAuditLogRepo()
	svc := NewAuditLogService(repo)

	// Seed 10 logs.
	for i := range 10 {
		repo.addLog(&domain.AuditLog{
			ID: uuid.New(), UserID: uuid.New(), Username: fmt.Sprintf("user-%d", i),
			Action: "order.create", Resource: "order", ResourceID: fmt.Sprintf("order-%d", i),
		})
	}

	// Page 1, size 5.
	logs, total, err := svc.ListAuditLogs(ctx, repository.AuditLogFilter{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("ListAuditLogs page 1 failed: %v", err)
	}
	if len(logs) != 5 {
		t.Errorf("expected 5 logs on page 1, got %d", len(logs))
	}
	if total != 10 {
		t.Errorf("expected total=10, got %d", total)
	}

	// Page 2, size 5.
	logs, total, err = svc.ListAuditLogs(ctx, repository.AuditLogFilter{Limit: 5, Offset: 5})
	if err != nil {
		t.Fatalf("ListAuditLogs page 2 failed: %v", err)
	}
	if len(logs) != 5 {
		t.Errorf("expected 5 logs on page 2, got %d", len(logs))
	}
	if total != 10 {
		t.Errorf("expected total=10, got %d", total)
	}

	// Page 3 (empty).
	logs, total, err = svc.ListAuditLogs(ctx, repository.AuditLogFilter{Limit: 5, Offset: 10})
	if err != nil {
		t.Fatalf("ListAuditLogs page 3 failed: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs on page 3, got %d", len(logs))
	}
	if total != 10 {
		t.Errorf("expected total=10 on page 3, got %d", total)
	}
}

func TestAuditLogService_ListAuditLogs_EmptyRepo(t *testing.T) {
	ctx := context.Background()
	svc := NewAuditLogService(newMockAuditLogRepo())

	logs, total, err := svc.ListAuditLogs(ctx, repository.AuditLogFilter{})
	if err != nil {
		t.Fatalf("ListAuditLogs empty repo failed: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs, got %d", len(logs))
	}
	if total != 0 {
		t.Errorf("expected total=0, got %d", total)
	}
}
