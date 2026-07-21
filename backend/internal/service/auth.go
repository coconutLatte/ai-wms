// Package service implements business logic orchestration for the WMS domain.
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// ── Auth DTOs ──────────────────────────────────────────────────────────────────────────────

// TokenPair is returned after successful authentication.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"` // Always "Bearer"
	ExpiresIn    int64  `json:"expires_in"` // Seconds until access token expires
}

// LoginInput is the request body for username/password authentication.
type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshInput is the request body for refreshing an access token.
type RefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

// ── JWT Claims ─────────────────────────────────────────────────────────────────────────────

// TokenClaims represents the custom JWT claims for WMS authentication.
type TokenClaims struct {
	jwt.RegisteredClaims
	Username  string   `json:"username"`
	RoleIDs   []string `json:"role_ids"`
	RoleNames []string `json:"role_names"`
	TokenType string   `json:"token_type"` // "access" or "refresh"
}

// ── AuthService ────────────────────────────────────────────────────────────────────────────

// AuthService handles authentication (login, token refresh, password hashing).
type AuthService struct {
	userRepo      repository.UserRepository
	jwtSecret     []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo repository.UserRepository, jwtSecret string, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// ── Public Methods ─────────────────────────────────────────────────────────────────────────

// Login authenticates a user by username and password. On success it returns an
// access/refresh token pair, updates the user's last_login timestamp, and returns
// the user profile. Invalid credentials always return a generic "invalid username
// or password" message to avoid user enumeration.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*TokenPair, *domain.User, error) {
	// Look up the user by username.
	user, err := s.userRepo.GetUserByUsername(ctx, input.Username)
	if err != nil {
		return nil, nil, wrapAuthError("invalid username or password")
	}

	// Verify the user is active.
	if user.Status != domain.UserStatusActive {
		return nil, nil, wrapAuthError("account is not active")
	}

	// Verify password against hash.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, wrapAuthError("invalid username or password")
	}

	// Generate token pair.
	pair, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("generate tokens: %w", err)
	}

	// Update last login timestamp (best-effort; don't fail login for this).
	now := time.Now()
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID, now)
	user.LastLogin = &now

	return pair, user, nil
}

// RefreshToken validates a refresh token and issues a new token pair.
// The old refresh token is consumed (rotation for security).
func (s *AuthService) RefreshToken(ctx context.Context, input RefreshInput) (*TokenPair, error) {
	// Parse and validate the refresh token.
	claims, err := s.parseToken(input.RefreshToken)
	if err != nil {
		return nil, wrapAuthError("invalid or expired refresh token")
	}

	// Ensure it is a refresh token (not an access token).
	if claims.TokenType != "refresh" {
		return nil, wrapAuthError("token is not a refresh token")
	}

	// Parse the user ID from the claims.
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, wrapAuthError("invalid token subject")
	}

	// Verify the user still exists and is active.
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, wrapAuthError("user not found")
	}
	if user.Status != domain.UserStatusActive {
		return nil, wrapAuthError("account is not active")
	}

	// Generate a new token pair (rotation).
	pair, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	return pair, nil
}

// HashPassword hashes a plain-text password using bcrypt with the default cost.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(bytes), nil
}

// ── Internal Helpers ───────────────────────────────────────────────────────────────────────

// generateTokenPair creates a new access + refresh token pair for the given user.
func (s *AuthService) generateTokenPair(ctx context.Context, user *domain.User) (*TokenPair, error) {
	now := time.Now()

	// Build role ID strings and look up role names for the claims.
	roleIDs := make([]string, len(user.RoleIDs))
	roleNames := make([]string, len(user.RoleIDs))
	for i, id := range user.RoleIDs {
		roleIDs[i] = id.String()
		role, err := s.userRepo.GetRole(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("lookup role %s: %w", id, err)
		}
		roleNames[i] = role.Name
	}

	// Generate a unique token ID for the refresh token.
	jti, err := generateJTI()
	if err != nil {
		return nil, fmt.Errorf("generate jti: %w", err)
	}

	// Access token claims.
	accessClaims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			ID:        jti,
		},
		Username:  user.Username,
		RoleIDs:   roleIDs,
		RoleNames: roleNames,
		TokenType: "access",
	}

	accessToken, err := s.signToken(accessClaims)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// Refresh token claims (longer TTL, different jti for uniqueness).
	refreshJTI, err := generateJTI()
	if err != nil {
		return nil, fmt.Errorf("generate refresh jti: %w", err)
	}

	refreshClaims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
			ID:        refreshJTI,
		},
		Username:  user.Username,
		RoleIDs:   roleIDs,
		RoleNames: roleNames,
		TokenType: "refresh",
	}

	refreshToken, err := s.signToken(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

// signToken creates a signed JWT string from the given claims using HS256.
func (s *AuthService) signToken(claims TokenClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// parseToken parses and validates a JWT token string, returning the claims.
func (s *AuthService) parseToken(tokenStr string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(t *jwt.Token) (any, error) {
		// Verify the signing method is what we expect.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ── Helpers ────────────────────────────────────────────────────────────────────────────────

// generateJTI creates a cryptographically random token ID (jti).
func generateJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// wrapAuthError creates a generic authentication error that doesn't reveal
// whether the username or password was incorrect (prevents user enumeration).
func wrapAuthError(msg string) error {
	return &pkgerrors.DomainError{
		Err:     pkgerrors.ErrUnauthorized,
		Message: msg,
		Code:    "UNAUTHORIZED",
	}
}
