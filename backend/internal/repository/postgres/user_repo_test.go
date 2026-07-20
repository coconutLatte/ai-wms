package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// setupUserTestDB creates a test database and cleans up user/role/audit_log test data.
func setupUserTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://wms:wms_dev_2026@localhost:5432/wms?sslmode=disable"
	}

	ctx := context.Background()
	db, err := NewDB(ctx, dsn)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
		return nil, nil
	}

	// Clean up previous test data (FK-aware order)
	db.Pool.Exec(ctx, "DELETE FROM audit_logs WHERE username LIKE 'TEST-%'")
	db.Pool.Exec(ctx, "DELETE FROM users WHERE username LIKE 'TEST-%'")
	db.Pool.Exec(ctx, "DELETE FROM roles WHERE name LIKE 'TEST-%'")

	cleanup := func() {
		db.Pool.Exec(ctx, "DELETE FROM audit_logs WHERE username LIKE 'TEST-%'")
		db.Pool.Exec(ctx, "DELETE FROM users WHERE username LIKE 'TEST-%'")
		db.Pool.Exec(ctx, "DELETE FROM roles WHERE name LIKE 'TEST-%'")
		db.Close()
	}

	return db, cleanup
}

// ── User Tests ──────────────────────────────────────────────

func TestUserRepo_CreateUser(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	roleID := uuid.New()
	user := &domain.User{
		Username:     "TEST-user-create-" + uuid.New().String()[:8],
		Email:        "test-create-" + uuid.New().String()[:8] + "@test.com",
		PasswordHash: "$2a$10$test_hash_placeholder",
		DisplayName:  "Test Create User",
		RoleIDs:      []uuid.UUID{roleID},
	}

	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.ID == uuid.Nil {
		t.Error("Expected ID to be assigned")
	}
	if user.Status != domain.UserStatusActive {
		t.Errorf("Expected status active, got %s", user.Status)
	}
	if user.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	// Verify retrieval
	fetched, err := repo.GetUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if fetched.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, fetched.Username)
	}
	if fetched.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, fetched.Email)
	}
	if fetched.DisplayName != user.DisplayName {
		t.Errorf("Expected display_name %s, got %s", user.DisplayName, fetched.DisplayName)
	}
	if len(fetched.RoleIDs) != 1 || fetched.RoleIDs[0] != roleID {
		t.Errorf("Expected role_ids [%s], got %v", roleID, fetched.RoleIDs)
	}
}

func TestUserRepo_CreateUser_Defaults(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	user := &domain.User{
		Username:     "TEST-user-defaults-" + uuid.New().String()[:8],
		Email:        "test-defaults-" + uuid.New().String()[:8] + "@test.com",
		PasswordHash: "$2a$10$test_hash",
	}

	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.Status != domain.UserStatusActive {
		t.Errorf("Expected default status active, got %s", user.Status)
	}
	if user.RoleIDs == nil || len(user.RoleIDs) != 0 {
		t.Errorf("Expected empty role_ids slice, got %v", user.RoleIDs)
	}
}

func TestUserRepo_GetUser_NotFound(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	_, err := repo.GetUser(ctx, uuid.New())
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestUserRepo_GetUserByUsername(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	user := &domain.User{
		Username:     "TEST-user-byuname-" + uuid.New().String()[:8],
		Email:        "test-byuname-" + uuid.New().String()[:8] + "@test.com",
		PasswordHash: "$2a$10$test_hash",
		DisplayName:  "Lookup User",
	}
	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	fetched, err := repo.GetUserByUsername(ctx, user.Username)
	if err != nil {
		t.Fatalf("GetUserByUsername failed: %v", err)
	}
	if fetched.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, fetched.ID)
	}

	// Not found
	_, err = repo.GetUserByUsername(ctx, "nonexistent-"+uuid.New().String())
	if err == nil {
		t.Error("Expected error for non-existent username")
	}
}

func TestUserRepo_GetUserByEmail(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	user := &domain.User{
		Username:     "TEST-user-byemail-" + uuid.New().String()[:8],
		Email:        "test-byemail-" + uuid.New().String()[:8] + "@test.com",
		PasswordHash: "$2a$10$test_hash",
	}
	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	fetched, err := repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if fetched.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, fetched.ID)
	}

	// Not found
	_, err = repo.GetUserByEmail(ctx, "nonexistent-"+uuid.New().String()+"@test.com")
	if err == nil {
		t.Error("Expected error for non-existent email")
	}
}

func TestUserRepo_ListUsers(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	// Create 3 users: 2 active, 1 inactive
	for i := 0; i < 3; i++ {
		user := &domain.User{
			Username:     "TEST-user-list-" + uuid.New().String()[:8],
			Email:        "test-list-" + uuid.New().String()[:8] + "@test.com",
			PasswordHash: "$2a$10$test_hash",
		}
		if i == 2 {
			user.Status = domain.UserStatusInactive
		}
		if err := repo.CreateUser(ctx, user); err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}
	}

	// List all
	all, err := repo.ListUsers(ctx, repository.UserFilter{})
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(all) < 3 {
		t.Errorf("Expected at least 3 users, got %d", len(all))
	}

	// Filter by status: active only
	active, err := repo.ListUsers(ctx, repository.UserFilter{Status: domain.UserStatusActive})
	if err != nil {
		t.Fatalf("ListUsers(active) failed: %v", err)
	}
	for _, u := range active {
		if u.Status != domain.UserStatusActive {
			t.Errorf("Expected only active users, got %s", u.Status)
		}
	}

	// Filter by status: inactive only
	inactive, err := repo.ListUsers(ctx, repository.UserFilter{Status: domain.UserStatusInactive})
	if err != nil {
		t.Fatalf("ListUsers(inactive) failed: %v", err)
	}
	for _, u := range inactive {
		if u.Status != domain.UserStatusInactive {
			t.Errorf("Expected only inactive users, got %s", u.Status)
		}
	}
}

func TestUserRepo_ListUsers_LimitOffset(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	for i := 0; i < 3; i++ {
		user := &domain.User{
			Username:     "TEST-user-lim-" + uuid.New().String()[:8],
			Email:        "test-lim-" + uuid.New().String()[:8] + "@test.com",
			PasswordHash: "$2a$10$test_hash",
		}
		if err := repo.CreateUser(ctx, user); err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}
	}

	// Limit
	limited, err := repo.ListUsers(ctx, repository.UserFilter{Limit: 1})
	if err != nil {
		t.Fatalf("ListUsers(limit=1) failed: %v", err)
	}
	if len(limited) > 1 {
		t.Errorf("Expected at most 1 user with limit, got %d", len(limited))
	}

	// Offset
	all, err := repo.ListUsers(ctx, repository.UserFilter{})
	if err != nil {
		t.Fatalf("ListUsers() failed: %v", err)
	}
	if len(all) >= 2 {
		offset, err := repo.ListUsers(ctx, repository.UserFilter{Offset: 1})
		if err != nil {
			t.Fatalf("ListUsers(offset=1) failed: %v", err)
		}
		if len(offset) >= len(all) {
			t.Errorf("Expected fewer results with offset=1")
		}
	}
}

func TestUserRepo_UpdateUser(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	user := &domain.User{
		Username:     "TEST-user-update-" + uuid.New().String()[:8],
		Email:        "test-update-" + uuid.New().String()[:8] + "@test.com",
		PasswordHash: "$2a$10$test_hash",
		DisplayName:  "Original Name",
	}
	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Update fields
	user.Email = "updated-" + user.Email
	user.DisplayName = "Updated Name"
	user.Status = domain.UserStatusInactive

	if err := repo.UpdateUser(ctx, user); err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	// Verify
	fetched, err := repo.GetUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if fetched.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, fetched.Email)
	}
	if fetched.DisplayName != user.DisplayName {
		t.Errorf("Expected display_name %s, got %s", user.DisplayName, fetched.DisplayName)
	}
	if fetched.Status != domain.UserStatusInactive {
		t.Errorf("Expected status inactive, got %s", fetched.Status)
	}
}

func TestUserRepo_UpdateUser_NotFound(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	err := repo.UpdateUser(ctx, &domain.User{ID: uuid.New(), Email: "no@test.com"})
	if err == nil {
		t.Error("Expected error for non-existent user update")
	}
}

func TestUserRepo_UpdateUserStatus(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	user := &domain.User{
		Username:     "TEST-user-status-" + uuid.New().String()[:8],
		Email:        "test-status-" + uuid.New().String()[:8] + "@test.com",
		PasswordHash: "$2a$10$test_hash",
	}
	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := repo.UpdateUserStatus(ctx, user.ID, domain.UserStatusLocked); err != nil {
		t.Fatalf("UpdateUserStatus failed: %v", err)
	}

	fetched, err := repo.GetUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if fetched.Status != domain.UserStatusLocked {
		t.Errorf("Expected status locked, got %s", fetched.Status)
	}
}

func TestUserRepo_UpdateUserStatus_NotFound(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	err := repo.UpdateUserStatus(ctx, uuid.New(), domain.UserStatusActive)
	if err == nil {
		t.Error("Expected error for non-existent user status update")
	}
}

// ── Role Tests ──────────────────────────────────────────────

func TestUserRepo_CreateRole(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	role := &domain.Role{
		Name:        "TEST-role-create-" + uuid.New().String()[:8],
		Description: "Test Role Description",
		Permissions: []domain.Permission{
			{Resource: "order", Actions: []string{"read", "create"}},
			{Resource: "inventory", Actions: []string{"read"}},
		},
	}

	if err := repo.CreateRole(ctx, role); err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	if role.ID == uuid.Nil {
		t.Error("Expected ID to be assigned")
	}
	if role.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	// Verify retrieval
	fetched, err := repo.GetRole(ctx, role.ID)
	if err != nil {
		t.Fatalf("GetRole failed: %v", err)
	}
	if fetched.Name != role.Name {
		t.Errorf("Expected name %s, got %s", role.Name, fetched.Name)
	}
	if len(fetched.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(fetched.Permissions))
	}
	if !fetched.Permissions[0].Can("order", "read") {
		t.Error("Expected permission to allow order.read")
	}
}

func TestUserRepo_CreateRole_Defaults(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	role := &domain.Role{
		Name: "TEST-role-defaults-" + uuid.New().String()[:8],
	}

	if err := repo.CreateRole(ctx, role); err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	if len(role.Permissions) != 0 {
		t.Errorf("Expected empty permissions, got %v", role.Permissions)
	}
}

func TestUserRepo_GetRole_NotFound(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	_, err := repo.GetRole(ctx, uuid.New())
	if err == nil {
		t.Error("Expected error for non-existent role")
	}
}

func TestUserRepo_ListRoles(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	// Create 2 test roles
	for i := 0; i < 2; i++ {
		role := &domain.Role{
			Name: "TEST-role-list-" + uuid.New().String()[:8],
		}
		if err := repo.CreateRole(ctx, role); err != nil {
			t.Fatalf("CreateRole failed: %v", err)
		}
	}

	roles, err := repo.ListRoles(ctx)
	if err != nil {
		t.Fatalf("ListRoles failed: %v", err)
	}
	if len(roles) < 2 {
		t.Errorf("Expected at least 2 roles, got %d", len(roles))
	}

	// Verify seeded roles exist
	hasAdmin := false
	hasOperator := false
	for _, r := range roles {
		if r.Name == "admin" {
			hasAdmin = true
		}
		if r.Name == "operator" {
			hasOperator = true
		}
	}
	if !hasAdmin {
		t.Error("Expected seeded 'admin' role")
	}
	if !hasOperator {
		t.Error("Expected seeded 'operator' role")
	}
}

func TestUserRepo_UpdateRole(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	role := &domain.Role{
		Name:        "TEST-role-update-" + uuid.New().String()[:8],
		Description: "Original Description",
		Permissions: []domain.Permission{
			{Resource: "order", Actions: []string{"read"}},
		},
	}
	if err := repo.CreateRole(ctx, role); err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	// Update
	role.Name = "TEST-role-updated-" + uuid.New().String()[:8]
	role.Description = "Updated Description"
	role.Permissions = []domain.Permission{
		{Resource: "order", Actions: []string{"read", "create", "update", "delete"}},
	}

	if err := repo.UpdateRole(ctx, role); err != nil {
		t.Fatalf("UpdateRole failed: %v", err)
	}

	fetched, err := repo.GetRole(ctx, role.ID)
	if err != nil {
		t.Fatalf("GetRole failed: %v", err)
	}
	if fetched.Name != role.Name {
		t.Errorf("Expected name %s, got %s", role.Name, fetched.Name)
	}
	if fetched.Description != role.Description {
		t.Errorf("Expected description %s, got %s", role.Description, fetched.Description)
	}
	if len(fetched.Permissions) != 1 || !fetched.Permissions[0].Can("order", "delete") {
		t.Error("Expected updated permissions including order.delete")
	}
}

func TestUserRepo_UpdateRole_NotFound(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	err := repo.UpdateRole(ctx, &domain.Role{ID: uuid.New(), Name: "no"})
	if err == nil {
		t.Error("Expected error for non-existent role update")
	}
}

// ── AuditLog Tests ──────────────────────────────────────────

func TestUserRepo_CreateAuditLog(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	// First create a test user so the FK is satisfied
	user := &domain.User{
		Username:     "TEST-audituser-" + uuid.New().String()[:8],
		Email:        "test-audituser-" + uuid.New().String()[:8] + "@test.com",
		PasswordHash: "$2a$10$test_hash",
	}
	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	log := &domain.AuditLog{
		UserID:     user.ID,
		Username:   user.Username,
		Action:     "order.create",
		Resource:   "order",
		ResourceID: uuid.New().String(),
		Details:    `{"order_no": "ORD-001"}`,
		IPAddress:  "192.168.1.1",
	}

	if err := repo.CreateAuditLog(ctx, log); err != nil {
		t.Fatalf("CreateAuditLog failed: %v", err)
	}

	if log.ID == uuid.Nil {
		t.Error("Expected ID to be assigned")
	}
	if log.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestUserRepo_ListAuditLogs(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	// Create a test user
	user := &domain.User{
		Username:     "TEST-auditlist-" + uuid.New().String()[:8],
		Email:        "test-auditlist-" + uuid.New().String()[:8] + "@test.com",
		PasswordHash: "$2a$10$test_hash",
	}
	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Create 3 audit logs for this user
	for i := 0; i < 3; i++ {
		log := &domain.AuditLog{
			UserID:     user.ID,
			Username:   user.Username,
			Action:     "order.create",
			Resource:   "order",
			ResourceID: uuid.New().String(),
			Details:    `{"idx": ` + string(rune('0'+i)) + `}`,
			IPAddress:  "10.0.0.1",
		}
		if err := repo.CreateAuditLog(ctx, log); err != nil {
			t.Fatalf("CreateAuditLog failed: %v", err)
		}
	}

	// Also create a log with a different action
	log := &domain.AuditLog{
		UserID:     user.ID,
		Username:   user.Username,
		Action:     "inventory.adjust",
		Resource:   "inventory",
		ResourceID: uuid.New().String(),
		Details:    `{"qty": 10}`,
		IPAddress:  "10.0.0.2",
	}
	if err := repo.CreateAuditLog(ctx, log); err != nil {
		t.Fatalf("CreateAuditLog failed: %v", err)
	}

	// List all for this user
	logs, err := repo.ListAuditLogs(ctx, repository.AuditLogFilter{UserID: user.ID})
	if err != nil {
		t.Fatalf("ListAuditLogs failed: %v", err)
	}
	if len(logs) != 4 {
		t.Errorf("Expected 4 audit logs, got %d", len(logs))
	}

	// Filter by action
	orderLogs, err := repo.ListAuditLogs(ctx, repository.AuditLogFilter{UserID: user.ID, Action: "order.create"})
	if err != nil {
		t.Fatalf("ListAuditLogs(action=order.create) failed: %v", err)
	}
	if len(orderLogs) != 3 {
		t.Errorf("Expected 3 order.create logs, got %d", len(orderLogs))
	}

	// Filter by resource
	invLogs, err := repo.ListAuditLogs(ctx, repository.AuditLogFilter{Resource: "inventory"})
	if err != nil {
		t.Fatalf("ListAuditLogs(resource=inventory) failed: %v", err)
	}
	if len(invLogs) < 1 {
		t.Errorf("Expected at least 1 inventory log, got %d", len(invLogs))
	}
}

func TestUserRepo_ListAuditLogs_LimitOffset(t *testing.T) {
	db, cleanup := setupUserTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepo(db)

	user := &domain.User{
		Username:     "TEST-auditlim-" + uuid.New().String()[:8],
		Email:        "test-auditlim-" + uuid.New().String()[:8] + "@test.com",
		PasswordHash: "$2a$10$test_hash",
	}
	if err := repo.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		log := &domain.AuditLog{
			UserID: user.ID, Username: user.Username,
			Action: "test.action", Resource: "test", ResourceID: uuid.New().String(),
		}
		if err := repo.CreateAuditLog(ctx, log); err != nil {
			t.Fatalf("CreateAuditLog failed: %v", err)
		}
	}

	// Limit
	limited, err := repo.ListAuditLogs(ctx, repository.AuditLogFilter{UserID: user.ID, Limit: 1})
	if err != nil {
		t.Fatalf("ListAuditLogs(limit=1) failed: %v", err)
	}
	if len(limited) > 1 {
		t.Errorf("Expected at most 1 log with limit, got %d", len(limited))
	}

	// Offset
	all, err := repo.ListAuditLogs(ctx, repository.AuditLogFilter{UserID: user.ID})
	if err != nil {
		t.Fatalf("ListAuditLogs() failed: %v", err)
	}
	if len(all) >= 2 {
		offset, err := repo.ListAuditLogs(ctx, repository.AuditLogFilter{UserID: user.ID, Offset: 1})
		if err != nil {
			t.Fatalf("ListAuditLogs(offset=1) failed: %v", err)
		}
		if len(offset) >= len(all) {
			t.Errorf("Expected fewer results with offset=1")
		}
	}
}
