package port

import (
	"context"

	"harajuku/backend/internal/core/domain"

	"github.com/google/uuid"
)

//go:generate mockgen -source=user.go -destination=mock/user.go -package=mock

// UserRepository is an interface for interacting with user-related data
type UserRepository interface {
	// CreateUser inserts a new user into the database
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	// GetUserByID selects a user by id
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	// GetUserByEmail selects a user by email
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	// GetAdminsEmail returns an array containing all the emails of amdmin users
	GetAdminsEmails(ctx context.Context) ([]string, error)
	// ListUsers selects a list of users with pagination
	ListUsers(ctx context.Context, skip, limit uint64, filters domain.UserFilters) ([]domain.User, error)
	// UpdateUser updates a user
	UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

// UserService is an interface for interacting with user-related business logic
type UserService interface {
	// Register registers a new user
	Register(ctx context.Context, user *domain.User) (*domain.User, error)
	// GetUser returns a user by id
	GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error)
	// ListUsers returns a list of users with pagination
	ListUsers(ctx context.Context, skip, limit uint64, filters domain.UserFilters) ([]domain.User, error)
	// UpdateUser updates a user
	UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
