package domain

import (
	"time"

	"github.com/google/uuid"
)

// TokenBlacklistEntry represents a revoked refresh token.
// When a user logs out, the refresh token's JTI (JWT ID) is stored here
// with the original token's expiry time. The entry is automatically
// eligible for cleanup after expires_at passes.
type TokenBlacklistEntry struct {
	ID        uuid.UUID `json:"id"`
	JTI       string    `json:"jti"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// NewTokenBlacklistEntry creates a new TokenBlacklistEntry with a generated ID.
// jti is the JWT Token ID claim value, userID identifies the token owner,
// expiresAt is when the original refresh token would have expired.
func NewTokenBlacklistEntry(jti string, userID uuid.UUID, expiresAt time.Time) *TokenBlacklistEntry {
	return &TokenBlacklistEntry{
		ID:        uuid.New(),
		JTI:       jti,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
}
