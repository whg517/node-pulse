package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrUserExists is returned when a user with the same username already exists
	ErrUserExists = errors.New("user already exists")
	// ErrLastAdmin is returned when attempting to delete the last admin user
	ErrLastAdmin = errors.New("cannot delete the last admin user")
)

// UserQuerier defines user database operations
type UserQuerier interface {
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, int, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User, passwordHash string) error
	UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	CountAdmins(ctx context.Context) (int, error)
}

type userQuerier struct {
	pool *pgxpool.Pool
}

// NewUserQuerier creates a new user querier
func NewUserQuerier(pool *pgxpool.Pool) UserQuerier {
	return &userQuerier{pool: pool}
}

// ListUsers retrieves a paginated list of users
func (q *userQuerier) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, int, error) {
	// Get total count
	var totalCount int
	err := q.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Query users with pagination
	query := `
		SELECT user_id, username, email, role, failed_login_attempts,
		       locked_until, mfa_enabled, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := q.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := []*models.User{}
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.FailedLoginAttempts,
			&user.LockedUntil,
			&user.MFAEnabled,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating users: %w", err)
	}

	return users, totalCount, nil
}

// GetUserByID retrieves a user by ID
func (q *userQuerier) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user := &models.User{}

	query := `
		SELECT user_id, username, email, role, failed_login_attempts,
		       locked_until, mfa_enabled, created_at, updated_at
		FROM users
		WHERE user_id = $1
	`

	err := q.pool.QueryRow(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.FailedLoginAttempts,
		&user.LockedUntil,
		&user.MFAEnabled,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (q *userQuerier) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}

	query := `
		SELECT user_id, username, email, role, failed_login_attempts,
		       locked_until, mfa_enabled, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	err := q.pool.QueryRow(ctx, query, username).Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.FailedLoginAttempts,
		&user.LockedUntil,
		&user.MFAEnabled,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// CreateUser creates a new user
func (q *userQuerier) CreateUser(ctx context.Context, user *models.User, passwordHash string) error {
	// Check if username already exists
	var exists bool
	err := q.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", user.Username).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check username existence: %w", err)
	}
	if exists {
		return ErrUserExists
	}

	// Generate UUID if not provided
	if user.UserID == "" {
		user.UserID = uuid.New().String()
	} else {
		// Validate UUID format
		_, err := uuid.Parse(user.UserID)
		if err != nil {
			return fmt.Errorf("invalid user ID format: %w", err)
		}
	}

	query := `
		INSERT INTO users (user_id, username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	err = q.pool.QueryRow(ctx, query,
		user.UserID,
		user.Username,
		user.Email,
		passwordHash,
		user.Role,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UpdateUser updates an existing user
func (q *userQuerier) UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
	// Build dynamic UPDATE query
	setClauses := []string{}
	args := []interface{}{}
	argCount := 1

	// Allowed update fields
	allowedFields := map[string]bool{
		"username": true,
		"email":    true,
		"password": true,
		"role":     true,
	}

	for field, value := range updates {
		if !allowedFields[field] {
			continue
		}

		var placeholder string
		if field == "password" {
			// Password field maps to password_hash
			placeholder = fmt.Sprintf("password_hash = $%d", argCount)
		} else {
			placeholder = fmt.Sprintf("%s = $%d", field, argCount)
		}

		setClauses = append(setClauses, placeholder)
		args = append(args, value)
		argCount++
	}

	if len(setClauses) == 0 {
		// No updates to apply
		return nil
	}

	// Add updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = NOW()"))

	// Build query
	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE user_id = $%d
	`, joinClauses(setClauses, ", "), argCount)

	args = append(args, userID)

	result, err := q.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeleteUser deletes a user
func (q *userQuerier) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// Check if user exists and is admin
	var role string
	err := q.pool.QueryRow(ctx, "SELECT role FROM users WHERE user_id = $1", userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to get user role: %w", err)
	}

	// If deleting an admin, check if there are other admins
	if role == "admin" {
		adminCount, err := q.CountAdmins(ctx)
		if err != nil {
			return fmt.Errorf("failed to count admins: %w", err)
		}
		if adminCount <= 1 {
			return ErrLastAdmin
		}
	}

	// Delete the user
	query := `DELETE FROM users WHERE user_id = $1`
	result, err := q.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// CountAdmins returns the number of admin users
func (q *userQuerier) CountAdmins(ctx context.Context) (int, error) {
	var count int
	err := q.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count admins: %w", err)
	}
	return count, nil
}

// joinClauses joins string clauses with a separator
func joinClauses(clauses []string, sep string) string {
	if len(clauses) == 0 {
		return ""
	}
	result := clauses[0]
	for i := 1; i < len(clauses); i++ {
		result += sep + clauses[i]
	}
	return result
}

// MockUserQuerier is a mock implementation for testing
type MockUserQuerier struct {
	Users map[uuid.UUID]*models.User
}

func (m *MockUserQuerier) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, int, error) {
	users := make([]*models.User, 0, len(m.Users))
	for _, u := range m.Users {
		users = append(users, u)
	}
	return users, len(users), nil
}

func (m *MockUserQuerier) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, exists := m.Users[userID]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserQuerier) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	for _, u := range m.Users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *MockUserQuerier) CreateUser(ctx context.Context, user *models.User, passwordHash string) error {
	if m.Users == nil {
		m.Users = make(map[uuid.UUID]*models.User)
	}
	userID, _ := uuid.Parse(user.UserID)
	if _, exists := m.Users[userID]; exists {
		return ErrUserExists
	}
	m.Users[userID] = user
	return nil
}

func (m *MockUserQuerier) UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
	user, exists := m.Users[userID]
	if !exists {
		return ErrUserNotFound
	}
	if username, ok := updates["username"].(string); ok {
		user.Username = username
	}
	if email, ok := updates["email"].(*string); ok {
		user.Email = email
	}
	if role, ok := updates["role"].(string); ok {
		user.Role = role
	}
	return nil
}

func (m *MockUserQuerier) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if _, exists := m.Users[userID]; !exists {
		return ErrUserNotFound
	}
	delete(m.Users, userID)
	return nil
}

func (m *MockUserQuerier) CountAdmins(ctx context.Context) (int, error) {
	count := 0
	for _, u := range m.Users {
		if u.Role == "admin" {
			count++
		}
	}
	return count, nil
}
