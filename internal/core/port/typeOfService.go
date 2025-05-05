package port

import (
	"context"

	"harajuku/backend/internal/core/domain"

	"github.com/google/uuid"
)

//go:generate mockgen -source=type_of_service.go -destination=mock/type_of_service.go -package=mock

// TypeOfServiceRepository is an interface for interacting with type-of-service-related data
type TypeOfServiceRepository interface {
	// CreateTypeOfService inserts a new type of service into the database
	CreateTypeOfService(ctx context.Context, service *domain.TypeOfService) (*domain.TypeOfService, error)
	// GetTypeOfServiceByID selects a type of service by id
	GetTypeOfServiceByID(ctx context.Context, id uuid.UUID) (*domain.TypeOfService, error)
	// ListTypeOfServices selects a list of types of service with pagination
	ListTypeOfServices(ctx context.Context, skip, limit uint64) ([]domain.TypeOfService, error)
	// UpdateTypeOfService updates a type of service
	UpdateTypeOfService(ctx context.Context, service *domain.TypeOfService) (*domain.TypeOfService, error)
	// DeleteTypeOfService deletes a type of service
	DeleteTypeOfService(ctx context.Context, id uuid.UUID) error
}

// TypeOfServiceService is an interface for interacting with type-of-service-related business logic
type TypeOfServiceService interface {
	// Creates a new type of service
	CreateTypeOfService(ctx context.Context, service *domain.TypeOfService) (*domain.TypeOfService, error)
	// GetTypeOfService returns a type of service by id
	GetTypeOfService(ctx context.Context, id uuid.UUID) (*domain.TypeOfService, error)
	// ListTypeOfServices returns a list of types of service with pagination
	ListTypeOfServices(ctx context.Context, skip, limit uint64) ([]domain.TypeOfService, error)
	// UpdateTypeOfService updates a type of service
	UpdateTypeOfService(ctx context.Context, service *domain.TypeOfService) (*domain.TypeOfService, error)
	// DeleteTypeOfService deletes a type of service
	DeleteTypeOfService(ctx context.Context, id uuid.UUID) error
}
