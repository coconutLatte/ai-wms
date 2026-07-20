package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user (admin, warehouse operator, PDA user).
type User struct {
	ID        uuid.UUID  `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	PasswordHash string  `json:"-"` // Never serialized to JSON
	DisplayName string   `json:"display_name"`
	RoleIDs   []uuid.UUID `json:"role_ids"`
	Status    UserStatus `json:"status"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// UserStatus represents the account status.
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusLocked   UserStatus = "locked"
)

// Role defines a set of permissions.
type Role struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`         // e.g. "admin", "operator", "picker"
	Description string         `json:"description"`
	Permissions []Permission   `json:"permissions"`
	CreatedAt   time.Time      `json:"created_at"`
}

// Permission represents a single access right.
type Permission struct {
	Resource   string   `json:"resource"`    // e.g. "warehouse", "inventory", "order"
	Actions    []string `json:"actions"`     // e.g. ["read", "create", "update", "delete"]
}

// Can checks if this permission allows a specific action on a resource.
func (p Permission) Can(resource, action string) bool {
	if p.Resource != resource && p.Resource != "*" {
		return false
	}
	for _, a := range p.Actions {
		if a == action || a == "*" {
			return true
		}
	}
	return false
}

// AuditLog records an audited operation for compliance and traceability.
type AuditLog struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`     // e.g. "order.create", "inventory.adjust"
	Resource   string    `json:"resource"`   // e.g. "order", "inventory"
	ResourceID string    `json:"resource_id"`
	Details    string    `json:"details"`    // JSON-encoded details
	IPAddress  string    `json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
}
