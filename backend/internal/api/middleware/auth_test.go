package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "test-secret-key-for-middleware-tests"

// createTestToken creates a signed JWT for testing.
func createTestToken(userID, username, tokenType string, roleIDs []string, ttl time.Duration) string {
	now := time.Now()
	claims := tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        "test-jti",
		},
		Username:  username,
		RoleIDs:   roleIDs,
		TokenType: tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testSecret))
	return signed
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	userID := uuid.New()
	roleIDs := []string{uuid.New().String(), uuid.New().String()}
	token := createTestToken(userID.String(), "testuser", "access", roleIDs, 15*time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()

	// Handler that echoes user info from context.
	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		gotUserID := GetUserID(ctx)
		gotUsername := GetUsername(ctx)
		gotRoles := GetUserRoleIDs(ctx)

		if gotUserID != userID {
			t.Errorf("expected user_id %s, got %s", userID, gotUserID)
		}
		if gotUsername != "testuser" {
			t.Errorf("expected username 'testuser', got %q", gotUsername)
		}
		if len(gotRoles) != 2 {
			t.Errorf("expected 2 role IDs, got %d", len(gotRoles))
		}
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "NotBearer abc123")
	rr := httptest.NewRecorder()

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_EmptyToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	rr := httptest.NewRecorder()

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	// Create a token that's already expired.
	token := createTestToken(uuid.New().String(), "testuser", "access", nil, -1*time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for expired token, got %d", rr.Code)
	}
}

func TestAuthMiddleware_RefreshTokenRejected(t *testing.T) {
	token := createTestToken(uuid.New().String(), "testuser", "refresh", nil, 15*time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	// Refresh tokens should be rejected by the auth middleware.
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 when using refresh token, got %d", rr.Code)
	}
}

func TestAuthMiddleware_TamperedToken(t *testing.T) {
	token := createTestToken(uuid.New().String(), "testuser", "access", nil, 15*time.Minute)
	// Tamper with the token by changing the last character.
	tampered := token[:len(token)-1] + "X"

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tampered)
	rr := httptest.NewRecorder()

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for tampered token, got %d", rr.Code)
	}
}

func TestAuthMiddleware_DifferentSecret(t *testing.T) {
	// Create a token with a different secret.
	userID := uuid.New()
	now := time.Now()
	claims := tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
		Username:  "testuser",
		TokenType: "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("different-secret"))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	rr := httptest.NewRecorder()

	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for token with wrong secret, got %d", rr.Code)
	}
}

func TestOptionalAuthMiddleware_NoToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler := OptionalAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if id := GetUserID(ctx); id != uuid.Nil {
			t.Error("expected nil user ID when no token present")
		}
		if name := GetUsername(ctx); name != "" {
			t.Error("expected empty username when no token present")
		}
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestOptionalAuthMiddleware_ValidToken(t *testing.T) {
	userID := uuid.New()
	token := createTestToken(userID.String(), "optuser", "access", nil, 15*time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler := OptionalAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if id := GetUserID(ctx); id != userID {
			t.Errorf("expected user_id %s, got %s", userID, id)
		}
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestOptionalAuthMiddleware_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	handler := OptionalAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Should still have no user info with invalid token.
		if id := GetUserID(ctx); id != uuid.Nil {
			t.Error("expected nil user ID with invalid token")
		}
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 (proceeds without auth), got %d", rr.Code)
	}
}

func TestContextHelpers_NoAuth(t *testing.T) {
	ctx := context.Background()

	if id := GetUserID(ctx); id != uuid.Nil {
		t.Error("GetUserID should return nil on empty context")
	}
	if name := GetUsername(ctx); name != "" {
		t.Error("GetUsername should return empty on empty context")
	}
	if roles := GetUserRoleIDs(ctx); roles != nil {
		t.Error("GetUserRoleIDs should return nil on empty context")
	}
}
