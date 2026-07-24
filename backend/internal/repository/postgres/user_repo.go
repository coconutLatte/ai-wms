package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
)

// UserRepo implements repository.UserRepository using PostgreSQL.
type UserRepo struct {
	db *DB
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(db *DB) *UserRepo {
	return &UserRepo{db: db}
}

// ── User ────────────────────────────────────────────────────

// CreateUser inserts a new user.
func (r *UserRepo) CreateUser(ctx context.Context, u *domain.User) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	if u.Status == "" {
		u.Status = domain.UserStatusActive
	}
	if u.RoleIDs == nil {
		u.RoleIDs = []uuid.UUID{}
	}

	const query = `
		INSERT INTO users (id, username, email, password_hash, display_name, role_ids, status, last_login, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.exec(ctx, query,
		u.ID, u.Username, u.Email, u.PasswordHash,
		nullString(u.DisplayName), u.RoleIDs, u.Status,
		u.LastLogin, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// GetUser retrieves a user by ID.
func (r *UserRepo) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const query = `
		SELECT id, username, email, password_hash, display_name, role_ids, status, last_login, created_at, updated_at
		FROM users WHERE id = $1`

	u, err := r.scanUser(r.queryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get user %s: %w", id, err)
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

// GetUserByUsername retrieves a user by username.
func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	const query = `
		SELECT id, username, email, password_hash, display_name, role_ids, status, last_login, created_at, updated_at
		FROM users WHERE username = $1`

	u, err := r.scanUser(r.queryRow(ctx, query, username))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get user by username %s: %w", username, err)
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return u, nil
}

// GetUserByEmail retrieves a user by email.
func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT id, username, email, password_hash, display_name, role_ids, status, last_login, created_at, updated_at
		FROM users WHERE email = $1`

	u, err := r.scanUser(r.queryRow(ctx, query, email))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get user by email %s: %w", email, err)
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

// ListUsers returns users matching the specified filter.
func (r *UserRepo) ListUsers(ctx context.Context, filter repository.UserFilter) ([]*domain.User, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	query := `
		SELECT id, username, email, password_hash, display_name, role_ids, status, last_login, created_at, updated_at
		FROM users`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
	}

	rows, err := r.query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u, err := r.scanUserFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}

// UpdateUser updates an existing user's mutable fields.
func (r *UserRepo) UpdateUser(ctx context.Context, u *domain.User) error {
	u.UpdatedAt = time.Now()

	const query = `
		UPDATE users SET email=$1, display_name=$2, role_ids=$3, status=$4, updated_at=$5
		WHERE id=$6`

	tag, err := r.exec(ctx, query,
		u.Email, nullString(u.DisplayName), u.RoleIDs, u.Status, u.UpdatedAt, u.ID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("update user %s: not found", u.ID)
	}
	return nil
}

// UpdateUserStatus transitions a user to a new status.
func (r *UserRepo) UpdateUserStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error {
	const query = `UPDATE users SET status=$1, updated_at=$2 WHERE id=$3`

	tag, err := r.exec(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update user status: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("update user status %s: not found", id)
	}
	return nil
}

// UpdateLastLogin updates the last_login timestamp for a user.
func (r *UserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error {
	const query = `UPDATE users SET last_login=$1, updated_at=$2 WHERE id=$3`

	tag, err := r.exec(ctx, query, t, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("update last login %s: not found", id)
	}
	return nil
}

// CountUsers returns the total number of users matching the filter.
func (r *UserRepo) CountUsers(ctx context.Context, filter repository.UserFilter) (int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
	}

	query := "SELECT COUNT(*) FROM users"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.queryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

// ── Role ────────────────────────────────────────────────────

// CreateRole inserts a new role.
func (r *UserRepo) CreateRole(ctx context.Context, role *domain.Role) error {
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	role.CreatedAt = time.Now()

	permsJSON, err := json.Marshal(role.Permissions)
	if err != nil {
		return fmt.Errorf("marshal permissions: %w", err)
	}
	if role.Permissions == nil {
		permsJSON = []byte("[]")
	}

	const query = `
		INSERT INTO roles (id, name, description, permissions, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err = r.exec(ctx, query,
		role.ID, role.Name, nullString(role.Description), permsJSON, role.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create role: %w", err)
	}
	return nil
}

// GetRole retrieves a role by ID.
func (r *UserRepo) GetRole(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	const query = `
		SELECT id, name, description, permissions, created_at
		FROM roles WHERE id = $1`

	role, err := r.scanRole(r.queryRow(ctx, query, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("get role %s: %w", id, err)
		}
		return nil, fmt.Errorf("get role: %w", err)
	}
	return role, nil
}

// ListRoles returns all roles.
func (r *UserRepo) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	const query = `
		SELECT id, name, description, permissions, created_at
		FROM roles ORDER BY name ASC`

	rows, err := r.query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		role, err := r.scanRoleFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roles: %w", err)
	}
	return roles, nil
}

// UpdateRole updates an existing role.
func (r *UserRepo) UpdateRole(ctx context.Context, role *domain.Role) error {
	permsJSON, err := json.Marshal(role.Permissions)
	if err != nil {
		return fmt.Errorf("marshal permissions: %w", err)
	}
	if role.Permissions == nil {
		permsJSON = []byte("[]")
	}

	const query = `
		UPDATE roles SET name=$1, description=$2, permissions=$3
		WHERE id=$4`

	tag, err := r.exec(ctx, query,
		role.Name, nullString(role.Description), permsJSON, role.ID,
	)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("update role %s: not found", role.ID)
	}
	return nil
}

// CountRoles returns the total number of roles.
func (r *UserRepo) CountRoles(ctx context.Context) (int, error) {
	const query = "SELECT COUNT(*) FROM roles"

	var count int
	err := r.queryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count roles: %w", err)
	}
	return count, nil
	}

// DeleteRole deletes a role by ID.
func (r *UserRepo) DeleteRole(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM roles WHERE id = $1`

	tag, err := r.exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	if tag == 0 {
		return fmt.Errorf("delete role %s: not found", id)
	}
	return nil
}

// ── AuditLog ────────────────────────────────────────────────

// CreateAuditLog inserts a new audit log entry.
func (r *UserRepo) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	log.CreatedAt = time.Now()

	const query = `
		INSERT INTO audit_logs (id, user_id, username, action, resource, resource_id, details, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.exec(ctx, query,
		log.ID, log.UserID, log.Username,
		log.Action, log.Resource, log.ResourceID,
		log.Details, log.IPAddress, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

// ListAuditLogs returns audit logs matching the specified filter.
func (r *UserRepo) ListAuditLogs(ctx context.Context, filter repository.AuditLogFilter) ([]*domain.AuditLog, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.UserID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, filter.UserID)
		argIdx++
	}
	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIdx))
		args = append(args, filter.Action)
		argIdx++
	}
	if filter.Resource != "" {
		conditions = append(conditions, fmt.Sprintf("resource = $%d", argIdx))
		args = append(args, filter.Resource)
		argIdx++
	}
	if filter.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d::timestamptz", argIdx))
		args = append(args, filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d::timestamptz", argIdx))
		args = append(args, filter.DateTo)
		argIdx++
	}

	query := `
		SELECT id, user_id, username, action, resource, resource_id, details, ip_address, created_at
		FROM audit_logs`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
	}

	rows, err := r.query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		l, err := r.scanAuditLogFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit logs: %w", err)
	}
	return logs, nil
}
// CountAuditLogs returns the total count of audit logs matching the filter.
func (r *UserRepo) CountAuditLogs(ctx context.Context, filter repository.AuditLogFilter) (int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.UserID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, filter.UserID)
		argIdx++
	}
	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIdx))
		args = append(args, filter.Action)
		argIdx++
	}
	if filter.Resource != "" {
		conditions = append(conditions, fmt.Sprintf("resource = $%d", argIdx))
		args = append(args, filter.Resource)
		argIdx++
	}
	if filter.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d::timestamptz", argIdx))
		args = append(args, filter.DateFrom)
		argIdx++
	}
	if filter.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d::timestamptz", argIdx))
		args = append(args, filter.DateTo)
	}

	query := "SELECT COUNT(*) FROM audit_logs"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.queryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count audit logs: %w", err)
	}
	return count, nil
}

// ── Scan Helpers ────────────────────────────────────────────

// scanUser scans a single user row.
func (r *UserRepo) scanUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
	var displayName *string

	err := row.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&displayName, &u.RoleIDs, &u.Status,
		&u.LastLogin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if displayName != nil {
		u.DisplayName = *displayName
	}
	if u.RoleIDs == nil {
		u.RoleIDs = []uuid.UUID{}
	}

	return u, nil
}

// scanUserFromRows scans a user row from a Rows iterator.
func (r *UserRepo) scanUserFromRows(rows pgx.Rows) (*domain.User, error) {
	u := &domain.User{}
	var displayName *string

	err := rows.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&displayName, &u.RoleIDs, &u.Status,
		&u.LastLogin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if displayName != nil {
		u.DisplayName = *displayName
	}
	if u.RoleIDs == nil {
		u.RoleIDs = []uuid.UUID{}
	}

	return u, nil
}

// scanRole scans a single role row.
func (r *UserRepo) scanRole(row pgx.Row) (*domain.Role, error) {
	role := &domain.Role{}
	var description *string
	var permsJSON []byte

	err := row.Scan(
		&role.ID, &role.Name, &description, &permsJSON, &role.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if description != nil {
		role.Description = *description
	}
	if len(permsJSON) > 0 {
		if err := json.Unmarshal(permsJSON, &role.Permissions); err != nil {
			return nil, fmt.Errorf("unmarshal permissions: %w", err)
		}
	}
	if role.Permissions == nil {
		role.Permissions = []domain.Permission{}
	}

	return role, nil
}

// scanRoleFromRows scans a role row from a Rows iterator.
func (r *UserRepo) scanRoleFromRows(rows pgx.Rows) (*domain.Role, error) {
	role := &domain.Role{}
	var description *string
	var permsJSON []byte

	err := rows.Scan(
		&role.ID, &role.Name, &description, &permsJSON, &role.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if description != nil {
		role.Description = *description
	}
	if len(permsJSON) > 0 {
		if err := json.Unmarshal(permsJSON, &role.Permissions); err != nil {
			return nil, fmt.Errorf("unmarshal permissions: %w", err)
		}
	}
	if role.Permissions == nil {
		role.Permissions = []domain.Permission{}
	}

	return role, nil
}

// scanAuditLogFromRows scans an audit log row from a Rows iterator.
func (r *UserRepo) scanAuditLogFromRows(rows pgx.Rows) (*domain.AuditLog, error) {
	l := &domain.AuditLog{}

	err := rows.Scan(
		&l.ID, &l.UserID, &l.Username,
		&l.Action, &l.Resource, &l.ResourceID,
		&l.Details, &l.IPAddress, &l.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// ── Transaction-aware dispatch helpers ─────────────────────

// exec dispatches to the active pgx.Tx if one exists in the context,
// otherwise falls back to the connection pool.
func (r *UserRepo) exec(ctx context.Context, sql string, args ...any) (int64, error) {
	if tx := TxFromContext(ctx); tx != nil {
		tag, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			return 0, err
		}
		return tag.RowsAffected(), nil
	}
	tag, err := r.db.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// query dispatches to the active pgx.Tx if one exists in the context,
// otherwise falls back to the connection pool.
func (r *UserRepo) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.Query(ctx, sql, args...)
	}
	return r.db.Pool.Query(ctx, sql, args...)
}

// queryRow dispatches to the active pgx.Tx if one exists in the context,
// otherwise falls back to the connection pool.
func (r *UserRepo) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.QueryRow(ctx, sql, args...)
	}
	return r.db.Pool.QueryRow(ctx, sql, args...)
}
