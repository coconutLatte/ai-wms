// Package middleware provides HTTP middleware for the WMS API layer.
// The rbac middleware enforces role-based access control after JWT authentication.
package middleware

import (
	"net/http"
	"strings"
)

// RequireRole returns a middleware that checks the authenticated user has at least
// one of the specified role names. If the user has none of the allowed roles,
// a 403 Forbidden response is returned. Must be used after the Auth middleware.
//
// Usage:
//
//	// Admin-only endpoint
//	mux.Handle("GET /api/v1/admin/users", RequireRole("admin")(handler))
//
//	// Multiple roles allowed
//	mux.Handle("GET /api/v1/tasks", RequireRole("admin", "operator")(handler))
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	// Build a lookup set for O(1) membership check.
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[strings.ToLower(r)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRoles := GetUserRoleNames(r.Context())

			// Check if any of the user's roles match the allowed set.
			authorized := false
			for _, ur := range userRoles {
				if _, ok := allowed[strings.ToLower(ur)]; ok {
					authorized = true
					break
				}
			}

			if !authorized {
				writeForbidden(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// writeForbidden writes a 403 Forbidden JSON response.
func writeForbidden(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
_, _ = w.Write([]byte(`{"error":"forbidden","detail":"` + msg + `"}`))
}
