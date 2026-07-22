package service

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// stubUserRepo is a simple stub for testing AuthService.
type stubUserRepo struct {
	users map[string]*domain.User
	roles map[uuid.UUID]*domain.Role
}

func newStubUserRepo() *stubUserRepo {
	return &stubUserRepo{
		users: make(map[string]*domain.User),
		roles: make(map[uuid.UUID]*domain.Role),
	}
}

func (r *stubUserRepo) addUser(u *domain.User) {
	r.users[u.Username] = u
}

func (r *stubUserRepo) addRole(role *domain.Role) {
	r.roles[role.ID] = role
}

// ── User ───────────────────────────────────────────────────

func (r *stubUserRepo) CreateUser(ctx context.Context, u *domain.User) error { return nil }
func (r *stubUserRepo) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (r *stubUserRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	u, ok := r.users[username]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return u, nil
}
func (r *stubUserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, fmt.Errorf("not found")
}
func (r *stubUserRepo) ListUsers(ctx context.Context, filter repository.UserFilter) ([]*domain.User, error) {
	return nil, nil
}
func (r *stubUserRepo) UpdateUser(ctx context.Context, u *domain.User) error { return nil }
func (r *stubUserRepo) UpdateUserStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error {
	return nil
}
func (r *stubUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error {
	return nil
}
func (r *stubUserRepo) CountUsers(ctx context.Context, filter repository.UserFilter) (int, error) {
	return 0, nil
}

// ── Role ───────────────────────────────────────────────────

func (r *stubUserRepo) CreateRole(ctx context.Context, role *domain.Role) error { return nil }
func (r *stubUserRepo) GetRole(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	role, ok := r.roles[id]
	if !ok {
		return nil, fmt.Errorf("role %s not found", id)
	}
	return role, nil
}
func (r *stubUserRepo) ListRoles(ctx context.Context) ([]*domain.Role, error) { return nil, nil }
func (r *stubUserRepo) UpdateRole(ctx context.Context, role *domain.Role) error { return nil }
func (r *stubUserRepo) DeleteRole(ctx context.Context, id uuid.UUID) error     { return nil }
func (r *stubUserRepo) CountRoles(ctx context.Context) (int, error)           { return 0, nil }

// ── AuditLog ───────────────────────────────────────────────

func (r *stubUserRepo) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error { return nil }
func (r *stubUserRepo) ListAuditLogs(ctx context.Context, filter repository.AuditLogFilter) ([]*domain.AuditLog, error) {
	return nil, nil
}
func (r *stubUserRepo) CountAuditLogs(ctx context.Context, filter repository.AuditLogFilter) (int, error) {
	return 0, nil
}

// ── TokenBlacklist ───────────────────────────────────────────

func (r *stubUserRepo) Add(ctx context.Context, entry *domain.TokenBlacklistEntry) error {
	return nil
}
func (r *stubUserRepo) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	return false, nil
}
func (r *stubUserRepo) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

// stubBlacklistRepo implements repository.TokenBlacklistRepository with an
// in-memory set to support tests for logout and blacklist-checked refresh.
type stubBlacklistRepo struct {
	sync.RWMutex
	blacklisted map[string]struct{}
}

func newStubBlacklistRepo() *stubBlacklistRepo {
	return &stubBlacklistRepo{blacklisted: make(map[string]struct{})}
}

func (r *stubBlacklistRepo) Add(ctx context.Context, entry *domain.TokenBlacklistEntry) error {
	r.Lock()
	defer r.Unlock()
	r.blacklisted[entry.JTI] = struct{}{}
	return nil
}

func (r *stubBlacklistRepo) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	r.RLock()
	defer r.RUnlock()
	_, ok := r.blacklisted[jti]
	return ok, nil
}

func (r *stubBlacklistRepo) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

// ── Tests ──────────────────────────────────────────────────────

func TestAuthService_Login_Success(t *testing.T) {
	// Setup: create a user with a known password.
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hash,
		Status:       domain.UserStatusActive,
		RoleIDs:      []uuid.UUID{uuid.New()},
	}

	repo := newStubUserRepo()
	repo.addUser(user)
	repo.addRole(&domain.Role{ID: user.RoleIDs[0], Name: "admin"})

	svc := NewAuthService(repo, "test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	// Act
	pair, returnedUser, err := svc.Login(context.Background(), LoginInput{
		Username: "testuser",
		Password: "correct-password",
	})

	// Assert
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if pair == nil {
		t.Fatal("expected token pair, got nil")
	}
	if pair.AccessToken == "" {
		t.Error("access token is empty")
	}
	if pair.RefreshToken == "" {
		t.Error("refresh token is empty")
	}
	if pair.TokenType != "Bearer" {
		t.Errorf("expected token_type 'Bearer', got %q", pair.TokenType)
	}
	if pair.ExpiresIn <= 0 {
		t.Errorf("expected positive expires_in, got %d", pair.ExpiresIn)
	}
	if returnedUser == nil {
		t.Fatal("expected user, got nil")
	}
	if returnedUser.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %q", returnedUser.Username)
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hash,
		Status:       domain.UserStatusActive,
	}

	repo := newStubUserRepo()
	repo.addUser(user)

	svc := NewAuthService(repo, "test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	_, _, err = svc.Login(context.Background(), LoginInput{
		Username: "testuser",
		Password: "wrong-password",
	})

	if err == nil {
		t.Fatal("expected error for invalid password, got nil")
	}
}

func TestAuthService_Login_UnknownUser(t *testing.T) {
	repo := newStubUserRepo()
	svc := NewAuthService(repo, "test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	_, _, err := svc.Login(context.Background(), LoginInput{
		Username: "nonexistent",
		Password: "any-password",
	})

	if err == nil {
		t.Fatal("expected error for unknown user, got nil")
	}
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "inactiveuser",
		Email:        "inactive@example.com",
		PasswordHash: hash,
		Status:       domain.UserStatusInactive,
	}

	repo := newStubUserRepo()
	repo.addUser(user)

	svc := NewAuthService(repo, "test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	_, _, err = svc.Login(context.Background(), LoginInput{
		Username: "inactiveuser",
		Password: "correct-password",
	})

	if err == nil {
		t.Fatal("expected error for inactive user, got nil")
	}
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hash,
		Status:       domain.UserStatusActive,
		RoleIDs:      []uuid.UUID{uuid.New()},
	}

	repo := newStubUserRepo()
	repo.addUser(user)
	repo.addRole(&domain.Role{ID: user.RoleIDs[0], Name: "operator"})

	svc := NewAuthService(repo, "test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	// Login first to get a refresh token.
	pair, _, err := svc.Login(context.Background(), LoginInput{
		Username: "testuser",
		Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Use the refresh token.
	newPair, err := svc.RefreshToken(context.Background(), RefreshInput{
		RefreshToken: pair.RefreshToken,
	})

	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if newPair == nil {
		t.Fatal("expected token pair, got nil")
	}
	if newPair.AccessToken == "" {
		t.Error("access token is empty after refresh")
	}
	if newPair.RefreshToken == "" {
		t.Error("refresh token is empty after refresh")
	}
	// Tokens should be rotated (new ones).
	if newPair.AccessToken == pair.AccessToken {
		t.Error("access token should be different after refresh")
	}
	if newPair.RefreshToken == pair.RefreshToken {
		t.Error("refresh token should be different after refresh")
	}
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	repo := newStubUserRepo()
	svc := NewAuthService(repo, "test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	_, err := svc.RefreshToken(context.Background(), RefreshInput{
		RefreshToken: "not-a-valid-jwt",
	})

	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestAuthService_RefreshToken_AccessTokenUsedAsRefresh(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hash,
		Status:       domain.UserStatusActive,
	}

	repo := newStubUserRepo()
	repo.addUser(user)

	svc := NewAuthService(repo, "test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	pair, _, err := svc.Login(context.Background(), LoginInput{
		Username: "testuser",
		Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Try to use access token as refresh token.
	_, err = svc.RefreshToken(context.Background(), RefreshInput{
		RefreshToken: pair.AccessToken,
	})

	if err == nil {
		t.Fatal("expected error when using access token as refresh token, got nil")
	}
}

func TestHashPassword(t *testing.T) {
	hash1, err := HashPassword("my-password")
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if hash1 == "" {
		t.Fatal("hash is empty")
	}

	// Same password should produce a different hash (different salt).
	hash2, err := HashPassword("my-password")
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if hash1 == hash2 {
		t.Error("bcrypt hashes for the same password should differ due to random salt")
	}

	// Verify the hash works.
	if err := bcrypt.CompareHashAndPassword([]byte(hash1), []byte("my-password")); err != nil {
		t.Errorf("hash verification failed: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash1), []byte("wrong-password")); err == nil {
		t.Error("hash should not verify for wrong password")
	}
}

func TestAuthService_TokenContainsUserInfo(t *testing.T) {
	hash, err := HashPassword("password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	roleID := uuid.New()
	user := &domain.User{
		ID:           uuid.New(),
		Username:     "infouser",
		Email:        "info@example.com",
		PasswordHash: hash,
		Status:       domain.UserStatusActive,
		RoleIDs:      []uuid.UUID{roleID},
	}

	repo := newStubUserRepo()
	repo.addUser(user)
	repo.addRole(&domain.Role{ID: roleID, Name: "picker"})

	svc := NewAuthService(repo, "test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	pair, _, err := svc.Login(context.Background(), LoginInput{
		Username: "infouser",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Parse the access token and verify claims.
	claims, err := svc.parseToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}

	if claims.Subject != user.ID.String() {
		t.Errorf("expected sub %s, got %s", user.ID.String(), claims.Subject)
	}
	if claims.Username != "infouser" {
		t.Errorf("expected username 'infouser', got %q", claims.Username)
	}
	if claims.TokenType != "access" {
		t.Errorf("expected token_type 'access', got %q", claims.TokenType)
	}
	if len(claims.RoleIDs) != 1 || claims.RoleIDs[0] != roleID.String() {
		t.Errorf("expected role_ids [%s], got %v", roleID.String(), claims.RoleIDs)
	}
	if len(claims.RoleNames) != 1 || claims.RoleNames[0] != "picker" {
		t.Errorf("expected role_names [picker], got %v", claims.RoleNames)
	}
}

func TestAuthService_AccessTokenExpiry(t *testing.T) {
	hash, err := HashPassword("password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "expiryuser",
		Email:        "expiry@example.com",
		PasswordHash: hash,
		Status:       domain.UserStatusActive,
	}

	repo := newStubUserRepo()
	repo.addUser(user)

	// Use a very short access TTL (1 ms) to test expiry.
	svc := NewAuthService(repo, "test-secret-key-32-bytes-long!", 1*time.Millisecond, 7*24*time.Hour)

	pair, _, err := svc.Login(context.Background(), LoginInput{
		Username: "expiryuser",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Wait for the token to expire.
	time.Sleep(10 * time.Millisecond)

	// Token should be expired now.
	_, err = svc.parseToken(pair.AccessToken)
	if err == nil {
		t.Error("expected expired token error, got nil")
	}
}

// ── Logout Tests ──────────────────────────────────────────────────

func TestAuthService_Logout_RevokesToken(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hash,
		Status:       domain.UserStatusActive,
		RoleIDs:      []uuid.UUID{uuid.New()},
	}

	repo := newStubUserRepo()
	repo.addUser(user)
	repo.addRole(&domain.Role{ID: user.RoleIDs[0], Name: "admin"})

	blRepo := newStubBlacklistRepo()
	svc := NewAuthServiceWithBlacklist(repo, blRepo, "test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	// Login to get a refresh token.
	pair, _, err := svc.Login(context.Background(), LoginInput{
		Username: "testuser",
		Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Logout — should add refresh token JTI to blacklist.
	err = svc.Logout(context.Background(), RefreshInput{
		RefreshToken: pair.RefreshToken,
	})
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	// Trying to refresh with the revoked token should fail.
	_, err = svc.RefreshToken(context.Background(), RefreshInput{
		RefreshToken: pair.RefreshToken,
	})
	if err == nil {
		t.Fatal("expected error when refreshing a revoked token, got nil")
	}
}

func TestAuthService_Logout_InvalidToken_Succeeds(t *testing.T) {
	repo := newStubUserRepo()
	blRepo := newStubBlacklistRepo()
	svc := NewAuthServiceWithBlacklist(repo, blRepo, "test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	// Logout with an invalid token should succeed (nothing to revoke).
	err := svc.Logout(context.Background(), RefreshInput{
		RefreshToken: "not-a-valid-jwt",
	})
	if err != nil {
		t.Fatalf("logout with invalid token should succeed (no-op), got: %v", err)
	}
}

func TestAuthService_Logout_AccessTokenTreatedAsRefresh_NoOp(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hash,
		Status:       domain.UserStatusActive,
		RoleIDs:      []uuid.UUID{uuid.New()},
	}

	repo := newStubUserRepo()
	repo.addUser(user)
	repo.addRole(&domain.Role{ID: user.RoleIDs[0], Name: "admin"})

	blRepo := newStubBlacklistRepo()
	svc := NewAuthServiceWithBlacklist(repo, blRepo, "test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	pair, _, err := svc.Login(context.Background(), LoginInput{
		Username: "testuser",
		Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Logout with access token (not refresh token) — should no-op.
	err = svc.Logout(context.Background(), RefreshInput{
		RefreshToken: pair.AccessToken,
	})
	if err != nil {
		t.Fatalf("logout with access token should no-op, got: %v", err)
	}

	// Refresh should still work since access token wasn't revoked.
	_, err = svc.RefreshToken(context.Background(), RefreshInput{
		RefreshToken: pair.RefreshToken,
	})
	if err != nil {
		t.Fatalf("refresh should still work after access token logout, got: %v", err)
	}
}

func TestAuthService_RefreshToken_RejectsRevokedToken(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hash,
		Status:       domain.UserStatusActive,
		RoleIDs:      []uuid.UUID{uuid.New()},
	}

	repo := newStubUserRepo()
	repo.addUser(user)
	repo.addRole(&domain.Role{ID: user.RoleIDs[0], Name: "admin"})

	blRepo := newStubBlacklistRepo()
	svc := NewAuthServiceWithBlacklist(repo, blRepo, "test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	pair, _, err := svc.Login(context.Background(), LoginInput{
		Username: "testuser",
		Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Revoke the refresh token.
	err = svc.Logout(context.Background(), RefreshInput{
		RefreshToken: pair.RefreshToken,
	})
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	// Attempt to refresh with revoked token.
	_, err = svc.RefreshToken(context.Background(), RefreshInput{
		RefreshToken: pair.RefreshToken,
	})
	if err == nil {
		t.Fatal("expected error for revoked refresh token, got nil")
	}
}

func TestAuthService_Logout_WithoutBlacklist_NoOp(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hash,
		Status:       domain.UserStatusActive,
		RoleIDs:      []uuid.UUID{uuid.New()},
	}

	repo := newStubUserRepo()
	repo.addUser(user)
	repo.addRole(&domain.Role{ID: user.RoleIDs[0], Name: "admin"})

	// Create service WITHOUT blacklist support.
	svc := NewAuthService(repo, "test-secret-key-32-bytes-long!", 15*time.Minute, 7*24*time.Hour)

	pair, _, err := svc.Login(context.Background(), LoginInput{
		Username: "testuser",
		Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Logout should not error even without blacklist.
	err = svc.Logout(context.Background(), RefreshInput{
		RefreshToken: pair.RefreshToken,
	})
	if err != nil {
		t.Fatalf("logout without blacklist should no-op, got: %v", err)
	}
}
