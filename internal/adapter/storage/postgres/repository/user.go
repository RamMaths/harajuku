package repository

import (
	"context"
	"fmt"
	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/core/domain"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

/**
 * UserRepository implements port.UserRepository interface
 * and provides an access to the postgres database
 */
type UserRepository struct {
    db *postgres.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *postgres.DB) *UserRepository {
    return &UserRepository{
        db,
    }
}

// CreateUser creates a new user in the database
func (ur *UserRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
    query := ur.db.QueryBuilder.Insert("users").
    Columns("id", "name", `"lastName"`, `"secondLastName"`, "email", "password"). // Use quoted identifiers
    Values(user.ID, user.Name, user.LastName, user.SecondLastName, user.Email, user.Password).
    Suffix("RETURNING *")

    sql, args, err := query.ToSql()
    if err != nil {
        return nil, err
    }

    err = ur.db.QueryRow(ctx, sql, args...).Scan(
        &user.ID,
        &user.Name,
        &user.LastName,
        &user.SecondLastName,
        &user.Email,
        &user.Password,
        &user.Role,
    )

    if err != nil {
        if errCode := ur.db.ErrorCode(err); errCode == "23505" {
            return nil, domain.ErrConflictingData
        }
        return nil, err
    }

    return user, nil
}

// GetUserByID gets a user by ID from the database
func (ur *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User

	query := ur.db.QueryBuilder.Select("*").
		From("users").
		Where(sq.Eq{"id": id}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = ur.db.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.Name,
		&user.LastName,
		&user.SecondLastName,
		&user.Email,
		&user.Password,
    &user.Role,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByEmailAndPassword gets a user by email from the database
func (ur *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	query := ur.db.QueryBuilder.Select("*").
		From("users").
		Where(sq.Eq{"email": email}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = ur.db.QueryRow(ctx, sql, args...).Scan(
    &user.ID,
    &user.Name,
    &user.LastName,
    &user.SecondLastName,
    &user.Email,
    &user.Password,
    &user.Role,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

	return &user, nil
}

// ListUsers lists users from the database with optional filters
func (ur *UserRepository) ListUsers(ctx context.Context, skip, limit uint64, filters domain.UserFilters) ([]domain.User, error) {
    var users []domain.User

    // Proper offset calculation
    offset := skip * limit
    
    query := ur.db.QueryBuilder.Select("*").
        From("users").
        OrderBy("\"id\"").  // Quoted identifier
        Limit(limit).
        Offset(offset)

    // Add filters with properly quoted column names
    if filters.Name != "" {
        query = query.Where(sq.ILike{"\"name\"": "%" + filters.Name + "%"})
    }
    if filters.LastName != "" {
        query = query.Where(sq.ILike{"\"lastName\"": "%" + filters.LastName + "%"})  // Exact case match
    }
    if filters.SecondLastName != "" {
        query = query.Where(sq.ILike{"\"secondLastName\"": "%" + filters.SecondLastName + "%"})
    }
    if filters.Role != "" {
        query = query.Where(sq.Eq{"\"role\"": filters.Role})
    }

    sql, args, err := query.ToSql()
    if err != nil {
        return nil, fmt.Errorf("query build failed: %w", err)
    }

    // Debug log the final query
    slog.DebugContext(ctx, "Executing query", 
        "sql", sql, 
        "args", args,
        "filters", filters)

    rows, err := ur.db.Query(ctx, sql, args...)
    if err != nil {
        return nil, fmt.Errorf("query execution failed: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var user domain.User
        err := rows.Scan(
            &user.ID,
            &user.Name,
            &user.LastName,
            &user.SecondLastName,
            &user.Email,
            &user.Password,
            &user.Role,
        )
        if err != nil {
            return nil, fmt.Errorf("row scan failed: %w", err)
        }
        users = append(users, user)
    }

    return users, nil
}

// UpdateUser updates a user by ID in the database
func (ur *UserRepository) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	name := nullString(user.Name)
	lastName := nullString(user.LastName)
	secondLastName := nullString(user.SecondLastName)
	email := nullString(user.Email)
	password := nullString(user.Password)
  role := nullString(string(user.Role))

	query := ur.db.QueryBuilder.Update("users").
		Set("name", sq.Expr("COALESCE(?, name)", name)).
		Set(`"lastName"`, sq.Expr("COALESCE(?, name)", lastName)).
		Set(`"secondLastName"`, sq.Expr("COALESCE(?, name)", secondLastName)).
		Set("email", sq.Expr("COALESCE(?, email)", email)).
		Set("password", sq.Expr("COALESCE(?, password)", password)).
    Set("role", sq.Expr("COALESCE(?, role)", role)).
		Where(sq.Eq{"id": user.ID}).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = ur.db.QueryRow(ctx, sql, args...).Scan(
    &user.ID,
    &user.Name,
    &user.LastName,
    &user.SecondLastName,
    &user.Email,
    &user.Password,
    &user.Role,
	)
	if err != nil {
		if errCode := ur.db.ErrorCode(err); errCode == "23505" {
			return nil, domain.ErrConflictingData
		}
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user by ID from the database
func (ur *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := ur.db.QueryBuilder.Delete("users").
		Where(sq.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = ur.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}
