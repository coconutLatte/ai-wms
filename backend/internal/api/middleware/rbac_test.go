package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRequireRole_Authorized(t *testing.T) {
	userID := uuid.New()
	token := createTestToken(
		userID.String(), "adminuser", "access",
		[]string{uuid.New().String()},
		[]string{"admin"},
		15*time.Minute,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler := Auth(testSecret)(
		RequireRole("admin")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	userID := uuid.New()
	token := createTestToken(
		userID.String(), "operator", "access",
		[]string{uuid.New().String()},
		[]string{"operator"},
		15*time.Minute,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler := Auth(testSecret)(
		RequireRole("admin")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("handler should not be called for forbidden request")
			}),
		),
	)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d: %s", rr.Code, rr.Body.String())
	}

	body := rr.Body.String()
	if body == "" {
		t.Error("expected error body")
	}
}

func TestRequireRole_MultipleAllowedRoles(t *testing.T) {
	userID := uuid.New()
	token := createTestToken(
		userID.String(), "picker", "access",
		[]string{uuid.New().String()},
		[]string{"picker"},
		15*time.Minute,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler := Auth(testSecret)(
		RequireRole("admin", "operator", "picker")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK (picker in allowed list), got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRequireRole_NoAuthContext(t *testing.T) {
	// No auth middleware — context has no user roles.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	rr := httptest.NewRecorder()

	handler := RequireRole("admin")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("handler should not be called without auth context")
		}),
	)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden (no roles in context), got %d", rr.Code)
	}
}

func TestRequireRole_CaseInsensitive(t *testing.T) {
	userID := uuid.New()
	token := createTestToken(
		userID.String(), "adminuser", "access",
		[]string{uuid.New().String()},
		[]string{"Admin"}, // Capital A
		15*time.Minute,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler := Auth(testSecret)(
		RequireRole("admin")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK (case-insensitive match), got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRequireRole_MultipleRolesInToken_OneMatches(t *testing.T) {
	userID := uuid.New()
	token := createTestToken(
		userID.String(), "multiuser", "access",
		[]string{uuid.New().String(), uuid.New().String()},
		[]string{"viewer", "admin"},
		15*time.Minute,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler := Auth(testSecret)(
		RequireRole("admin")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK (user has admin role), got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRequireRole_MultipleRolesInToken_NoneMatch(t *testing.T) {
	userID := uuid.New()
	token := createTestToken(
		userID.String(), "multiuser", "access",
		[]string{uuid.New().String(), uuid.New().String()},
		[]string{"viewer", "operator"},
		15*time.Minute,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler := Auth(testSecret)(
		RequireRole("admin")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("handler should not be called")
			}),
		),
	)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden (no matching role), got %d", rr.Code)
	}
}

func TestGetUserRoleNames_NoAuth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	names := GetUserRoleNames(req.Context())
	if names != nil {
		t.Errorf("expected nil role names on empty context, got %v", names)
	}
}

func TestAuthMiddleware_SetsRoleNames(t *testing.T) {
	userID := uuid.New()
	roleNames := []string{"admin", "operator"}
	token := createTestToken(
		userID.String(), "testuser", "access",
		[]string{uuid.New().String(), uuid.New().String()},
		roleNames,
		15*time.Minute,
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	var capturedNames []string
	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedNames = GetUserRoleNames(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if len(capturedNames) != 2 {
		t.Fatalf("expected 2 role names, got %d: %v", len(capturedNames), capturedNames)
	}
	if capturedNames[0] != "admin" {
		t.Errorf("expected first role name 'admin', got %q", capturedNames[0])
	}
	if capturedNames[1] != "operator" {
		t.Errorf("expected second role name 'operator', got %q", capturedNames[1])
	}
}

func BenchmarkRequireRole_Authorized(b *testing.B) {
	userID := uuid.New()
	token := createTestToken(
		userID.String(), "admin", "access",
		[]string{uuid.New().String()},
		[]string{"admin"},
		15*time.Minute,
	)

	handler := Auth(testSecret)(
		RequireRole("admin")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
