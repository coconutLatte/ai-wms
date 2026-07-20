// Package middleware provides HTTP middleware for the WMS API layer.
// The auth middleware validates JWT access tokens and injects user info into the request context.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ── Context Keys ────────────────────────────────────────────────────────────────────────────

const (
	// UserIDKey is the context key for the authenticated user's UUID.
	UserIDKey contextKey = "user_id"
	// UsernameKey is the context key for the authenticated user's username.
	UsernameKey contextKey = "username"
	// UserRoleIDsKey is the context key for the authenticated user's role ID strings.
	UserRoleIDsKey contextKey = "user_role_ids"
)

// ── Context Helpers ─────────────────────────────────────────────────────────────────────────

// GetUserID extracts the authenticated user ID from the context.
// Returns uuid.Nil if no user is authenticated.
func GetUserID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// GetUsername extracts the authenticated username from the context.
// Returns an empty string if no user is authenticated.
func GetUsername(ctx context.Context) string {
	if name, ok := ctx.Value(UsernameKey).(string); ok {
		return name
	}
	return ""
}

// GetUserRoleIDs extracts the authenticated user's role IDs from the context.
// Returns nil if no user is authenticated.
func GetUserRoleIDs(ctx context.Context) []string {
	if roles, ok := ctx.Value(UserRoleIDsKey).([]string); ok {
		return roles
	}
	return nil
}

// ── Token Claims ────────────────────────────────────────────────────────────────────────────

// tokenClaims represents the custom JWT claims for WMS authentication middleware.
type tokenClaims struct {
	jwt.RegisteredClaims
	Username  string   `json:"username"`
	RoleIDs   []string `json:"role_ids"`
	TokenType string   `json:"token_type"`
}

// ── Auth Middleware ─────────────────────────────────────────────────────────────────────────

// Auth returns a middleware that validates JWT Bearer tokens from the Authorization header.
// Valid tokens inject user information into the request context.
// Missing or invalid tokens return 401 Unauthorized.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	secret := []byte(jwtSecret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the token from the Authorization header.
			tokenStr, err := extractBearerToken(r)
			if err != nil {
				writeUnauthorized(w, err.Error())
				return
			}

			// Parse and validate the JWT token.
			claims, err := parseAndValidate(tokenStr, secret)
			if err != nil {
				writeUnauthorized(w, "invalid or expired token")
				return
			}

			// Ensure it is an access token (not a refresh token).
			if claims.TokenType != "access" {
				writeUnauthorized(w, "token is not an access token")
				return
			}

			// Parse user ID from the subject claim.
			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				writeUnauthorized(w, "invalid token subject")
				return
			}

			// Inject user information into the request context.
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, userID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)
			ctx = context.WithValue(ctx, UserRoleIDsKey, claims.RoleIDs)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth returns a middleware that optionally validates JWT tokens.
// If a valid token is present, user info is injected into the context.
// If no token or an invalid token, the request proceeds without user context.
func OptionalAuth(jwtSecret string) func(http.Handler) http.Handler {
	secret := []byte(jwtSecret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, err := extractBearerToken(r)
			if err != nil {
				// No token or malformed — proceed without auth context.
				next.ServeHTTP(w, r)
				return
			}

			claims, err := parseAndValidate(tokenStr, secret)
			if err != nil {
				// Invalid token — proceed without auth context.
				next.ServeHTTP(w, r)
				return
			}

			if claims.TokenType != "access" {
				next.ServeHTTP(w, r)
				return
			}

			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, userID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)
			ctx = context.WithValue(ctx, UserRoleIDsKey, claims.RoleIDs)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ── Internal Helpers ───────────────────────────────────────────────────────────────────────

// extractBearerToken extracts a Bearer token from the Authorization header.
func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	// Must be "Bearer <token>".
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("invalid authorization header format")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fmt.Errorf("empty token")
	}

	return token, nil
}

// parseAndValidate parses a JWT token string and validates its signature and expiry.
func parseAndValidate(tokenStr string, secret []byte) (*tokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &tokenClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// writeUnauthorized writes a 401 Unauthorized JSON response.
func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, `{"error":"unauthorized","detail":"%s"}`, msg)
}
