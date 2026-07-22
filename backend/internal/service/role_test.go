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

// mockRoleRepo implements repository.UserRepository for testing RoleService.
type mockRoleRepo struct {
	roles map[uuid.UUID]*domain.Role
}

func newMockRoleRepo() *mockRoleRepo {
	return &mockRoleRepo{roles: make(map[uuid.UUID]*domain.Role)}
}

func (m *mockRoleRepo) addRole(r *domain.Role) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.Permissions == nil {
		r.Permissions = []domain.Permission{}
	}
	clone := *r
	clone.Permissions = make([]domain.Permission, len(r.Permissions))
	copy(clone.Permissions, r.Permissions)
	m.roles[r.ID] = &clone
}

// ── Role ─────────────────────────────────────────────────────

func (m *mockRoleRepo) CreateRole(ctx context.Context, r *domain.Role) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.Permissions == nil {
		r.Permissions = []domain.Permission{}
	}
	clone := *r
	clone.Permissions = make([]domain.Permission, len(r.Permissions))
	copy(clone.Permissions, r.Permissions)
	m.roles[r.ID] = &clone
	return nil
}

func (m *mockRoleRepo) GetRole(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	r, ok := m.roles[id]
	if !ok {
		return nil, pkgerrors.NewNotFound("role", id.String())
	}
	clone := *r
	clone.Permissions = make([]domain.Permission, len(r.Permissions))
	copy(clone.Permissions, r.Permissions)
	return &clone, nil
}

func (m *mockRoleRepo) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	var result []*domain.Role
	for _, r := range m.roles {
		clone := *r
		clone.Permissions = make([]domain.Permission, len(r.Permissions))
		copy(clone.Permissions, r.Permissions)
		result = append(result, &clone)
	}
	return result, nil
}

func (m *mockRoleRepo) UpdateRole(ctx context.Context, r *domain.Role) error {
	existing, ok := m.roles[r.ID]
	if !ok {
		return pkgerrors.NewNotFound("role", r.ID.String())
	}
	existing.Name = r.Name
	existing.Description = r.Description
	existing.Permissions = make([]domain.Permission, len(r.Permissions))
	copy(existing.Permissions, r.Permissions)
	return nil
}

func (m *mockRoleRepo) DeleteRole(ctx context.Context, id uuid.UUID) error {
	if _, ok := m.roles[id]; !ok {
		return pkgerrors.NewNotFound("role", id.String())
	}
	delete(m.roles, id)
	return nil
}

func (m *mockRoleRepo) CountRoles(ctx context.Context) (int, error) {
	return len(m.roles), nil
}

// ── User (stubs) ──────────────────────────────────────────────

func (m *mockRoleRepo) CreateUser(ctx context.Context, u *domain.User) error                          { return nil }
func (m *mockRoleRepo) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error)               { return nil, nil }
func (m *mockRoleRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error)   { return nil, nil }
func (m *mockRoleRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error)         { return nil, nil }
func (m *mockRoleRepo) ListUsers(ctx context.Context, filter repository.UserFilter) ([]*domain.User, error) { return nil, nil }
func (m *mockRoleRepo) UpdateUser(ctx context.Context, u *domain.User) error                           { return nil }
func (m *mockRoleRepo) UpdateUserStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error { return nil }
func (m *mockRoleRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error         { return nil }
func (m *mockRoleRepo) CountUsers(ctx context.Context, filter repository.UserFilter) (int, error)      { return 0, nil }

// ── AuditLog (stubs) ──────────────────────────────────────────

func (m *mockRoleRepo) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error                       { return nil }
func (m *mockRoleRepo) ListAuditLogs(ctx context.Context, filter repository.AuditLogFilter) ([]*domain.AuditLog, error) { return nil, nil }
func (m *mockRoleRepo) CountAuditLogs(ctx context.Context, filter repository.AuditLogFilter) (int, error)    { return 0, nil }

// ── Tests: ListRoles ──────────────────────────────────────────

func TestRoleService_ListRoles(t *testing.T) {
	ctx := context.Background()
	repo := newMockRoleRepo()
	svc := NewRoleService(repo)

	// Seed roles.
	repo.addRole(&domain.Role{Name: "admin", Description: "Admin role"})
	repo.addRole(&domain.Role{Name: "operator", Description: "Operator role"})
	repo.addRole(&domain.Role{Name: "picker", Description: "Picker role"})

	roles, total, err := svc.ListRoles(ctx)
	if err != nil {
		t.Fatalf("ListRoles failed: %v", err)
	}
	if len(roles) != 3 {
		t.Errorf("expected 3 roles, got %d", len(roles))
	}
	if total != 3 {
		t.Errorf("expected total=3, got %d", total)
	}
}

func TestRoleService_ListRoles_Empty(t *testing.T) {
	ctx := context.Background()
	svc := NewRoleService(newMockRoleRepo())

	roles, total, err := svc.ListRoles(ctx)
	if err != nil {
		t.Fatalf("ListRoles empty repo failed: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("expected 0 roles, got %d", len(roles))
	}
	if total != 0 {
		t.Errorf("expected total=0, got %d", total)
	}
}

// ── Tests: CreateRole ─────────────────────────────────────────

func TestRoleService_CreateRole(t *testing.T) {
	ctx := context.Background()
	repo := newMockRoleRepo()
	svc := NewRoleService(repo)

	input := CreateRoleInput{
		Name:        "supervisor",
		Description: "Supervisor role",
		Permissions: []domain.Permission{
			{Resource: "warehouse", Actions: []string{"read"}},
			{Resource: "inventory", Actions: []string{"read", "update"}},
		},
	}

	role, err := svc.CreateRole(ctx, input)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	if role.Name != "supervisor" {
		t.Errorf("expected name 'supervisor', got %q", role.Name)
	}
	if role.Description != "Supervisor role" {
		t.Errorf("expected description 'Supervisor role', got %q", role.Description)
	}
	if len(role.Permissions) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(role.Permissions))
	}
	if role.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}

	// Verify retrievable.
	retrieved, err := svc.GetRole(ctx, role.ID)
	if err != nil {
		t.Fatalf("GetRole after create failed: %v", err)
	}
	if retrieved.ID != role.ID {
		t.Error("retrieved role ID mismatch")
	}
}

func TestRoleService_CreateRole_Defaults(t *testing.T) {
	ctx := context.Background()
	repo := newMockRoleRepo()
	svc := NewRoleService(repo)

	role, err := svc.CreateRole(ctx, CreateRoleInput{Name: "minimal"})
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	if role.Name != "minimal" {
		t.Errorf("expected name 'minimal', got %q", role.Name)
	}
	if role.Description != "" {
		t.Errorf("expected empty description, got %q", role.Description)
	}
	if len(role.Permissions) != 0 {
		t.Errorf("expected 0 permissions, got %d", len(role.Permissions))
	}
}

func TestRoleService_CreateRole_DuplicateName(t *testing.T) {
	ctx := context.Background()
	repo := newMockRoleRepo()
	svc := NewRoleService(repo)

	_, err := svc.CreateRole(ctx, CreateRoleInput{Name: "admin"})
	if err != nil {
		t.Fatalf("CreateRole first failed: %v", err)
	}

	_, err = svc.CreateRole(ctx, CreateRoleInput{Name: "admin"})
	if err == nil {
		t.Error("expected duplicate name error")
	}
	if !pkgerrors.Is(err, pkgerrors.ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestRoleService_CreateRole_Validation(t *testing.T) {
	ctx := context.Background()
	svc := NewRoleService(newMockRoleRepo())

	_, err := svc.CreateRole(ctx, CreateRoleInput{Name: ""})
	if err == nil {
		t.Error("expected validation error for empty name")
	}
	if !pkgerrors.IsInvalidInput(err) {
		t.Errorf("expected InvalidInput error, got %v", err)
	}
}

// ── Tests: GetRole ────────────────────────────────────────────

func TestRoleService_GetRole(t *testing.T) {
	ctx := context.Background()
	repo := newMockRoleRepo()
	svc := NewRoleService(repo)

	repo.addRole(&domain.Role{Name: "admin", Description: "Administrator"})

	// Find the role we just added.
	var roleID uuid.UUID
	for id := range repo.roles {
		roleID = id
		break
	}

	role, err := svc.GetRole(ctx, roleID)
	if err != nil {
		t.Fatalf("GetRole failed: %v", err)
	}
	if role.Name != "admin" {
		t.Errorf("expected name 'admin', got %q", role.Name)
	}
}

func TestRoleService_GetRole_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewRoleService(newMockRoleRepo())

	_, err := svc.GetRole(ctx, uuid.New())
	if err == nil {
		t.Error("expected error for non-existent role")
	}
	if !pkgerrors.IsNotFound(err) {
		t.Errorf("expected NotFound error, got %v", err)
	}
}

// ── Tests: UpdateRole ─────────────────────────────────────────

func TestRoleService_UpdateRole(t *testing.T) {
	ctx := context.Background()
	repo := newMockRoleRepo()
	svc := NewRoleService(repo)

	created, err := svc.CreateRole(ctx, CreateRoleInput{
		Name:        "old-name",
		Description: "Old description",
		Permissions: []domain.Permission{
			{Resource: "warehouse", Actions: []string{"read"}},
		},
	})
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	newName := "new-name"
	newDesc := "New description"
	newPerms := []domain.Permission{
		{Resource: "inventory", Actions: []string{"read", "create", "update", "delete"}},
	}

	updated, err := svc.UpdateRole(ctx, created.ID, UpdateRoleInput{
		Name:        &newName,
		Description: &newDesc,
		Permissions: newPerms,
	})
	if err != nil {
		t.Fatalf("UpdateRole failed: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("expected name %q, got %q", newName, updated.Name)
	}
	if updated.Description != newDesc {
		t.Errorf("expected description %q, got %q", newDesc, updated.Description)
	}
	if len(updated.Permissions) != 1 || updated.Permissions[0].Resource != "inventory" {
		t.Errorf("expected permissions to be updated")
	}
}

func TestRoleService_UpdateRole_Partial(t *testing.T) {
	ctx := context.Background()
	repo := newMockRoleRepo()
	svc := NewRoleService(repo)

	created, err := svc.CreateRole(ctx, CreateRoleInput{
		Name:        "partial-role",
		Description: "Original",
	})
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	newDesc := "Updated description"
	updated, err := svc.UpdateRole(ctx, created.ID, UpdateRoleInput{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("UpdateRole partial failed: %v", err)
	}

	if updated.Description != newDesc {
		t.Errorf("expected description %q, got %q", newDesc, updated.Description)
	}
	if updated.Name != "partial-role" {
		t.Errorf("expected name unchanged %q, got %q", "partial-role", updated.Name)
	}
}

func TestRoleService_UpdateRole_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewRoleService(newMockRoleRepo())

	newName := "test"
	_, err := svc.UpdateRole(ctx, uuid.New(), UpdateRoleInput{Name: &newName})
	if err == nil {
		t.Error("expected error for non-existent role")
	}
}

// ── Tests: DeleteRole ─────────────────────────────────────────

func TestRoleService_DeleteRole(t *testing.T) {
	ctx := context.Background()
	repo := newMockRoleRepo()
	svc := NewRoleService(repo)

	created, err := svc.CreateRole(ctx, CreateRoleInput{Name: "temporary"})
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	err = svc.DeleteRole(ctx, created.ID)
	if err != nil {
		t.Fatalf("DeleteRole failed: %v", err)
	}

	// Verify it's gone.
	_, err = svc.GetRole(ctx, created.ID)
	if err == nil {
		t.Error("expected error for deleted role")
	}
	if !pkgerrors.IsNotFound(err) {
		t.Errorf("expected NotFound error, got %v", err)
	}
}

func TestRoleService_DeleteRole_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewRoleService(newMockRoleRepo())

	err := svc.DeleteRole(ctx, uuid.New())
	if err == nil {
		t.Error("expected error for non-existent role")
	}
}
