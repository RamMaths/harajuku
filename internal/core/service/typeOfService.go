package service

import (
	"context"
	"log/slog"

	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"harajuku/backend/internal/core/util"

	"github.com/google/uuid"
)

/**
 * TypeOfServiceService implements port.TypeOfServiceService interface
 * and provides access to the type of service repository and cache service
 */
type TypeOfServiceService struct {
	repo  port.TypeOfServiceRepository
	cache port.CacheRepository
}

// NewTypeOfServiceService creates a new TypeOfService service instance
func NewTypeOfServiceService(repo port.TypeOfServiceRepository, cache port.CacheRepository) *TypeOfServiceService {
	return &TypeOfServiceService{
		repo,
		cache,
	}
}

// CreateTypeOfService creates a new type of service
func (s *TypeOfServiceService) CreateTypeOfService(ctx context.Context, t *domain.TypeOfService) (*domain.TypeOfService, error) {
	// Save the TypeOfService using the repository
	created, err := s.repo.CreateTypeOfService(ctx, t)
	if err != nil {
		slog.Error("TypeOfService creation failed", "error", err)
		if err == domain.ErrConflictingData {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	// Cache the newly created TypeOfService
	cacheKey := util.GenerateCacheKey("typeofservice", created.ID)
	serialized, err := util.Serialize(created)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = s.cache.Set(ctx, cacheKey, serialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return created, nil
}

// GetTypeOfService retrieves a type of service by ID
func (s *TypeOfServiceService) GetTypeOfService(ctx context.Context, id uuid.UUID) (*domain.TypeOfService, error) {
	var t *domain.TypeOfService

	// Check cache for TypeOfService
	cacheKey := util.GenerateCacheKey("typeofservice", id)
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		err := util.Deserialize(cachedData, &t)
		if err != nil {
			return nil, domain.ErrInternal
		}
		return t, nil
	}

	// If not found in cache, fetch from repository
	t, err = s.repo.GetTypeOfServiceByID(ctx, id)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	// Cache the retrieved TypeOfService
	serialized, err := util.Serialize(t)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = s.cache.Set(ctx, cacheKey, serialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return t, nil
}

// ListTypeOfServices lists all types of service
func (s *TypeOfServiceService) ListTypeOfServices(ctx context.Context, skip, limit uint64) ([]domain.TypeOfService, error) {
	var services []domain.TypeOfService

	// Generate cache key for paginated list
	params := util.GenerateCacheKeyParams(skip, limit)
	cacheKey := util.GenerateCacheKey("typeofservices", params)

	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		err := util.Deserialize(cachedData, &services)
		if err != nil {
			return nil, domain.ErrInternal
		}
		return services, nil
	}

	// Fetch list from repository if not found in cache
	services, err = s.repo.ListTypeOfServices(ctx, skip, limit)
	if err != nil {
		return nil, domain.ErrInternal
	}

	// Cache the list of TypeOfServices
	serialized, err := util.Serialize(services)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = s.cache.Set(ctx, cacheKey, serialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return services, nil
}

// UpdateTypeOfService updates an existing type of service
func (s *TypeOfServiceService) UpdateTypeOfService(ctx context.Context, t *domain.TypeOfService) (*domain.TypeOfService, error) {
	// Check if the type of service exists
	existingService, err := s.repo.GetTypeOfServiceByID(ctx, t.ID)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	// If no data was changed, return early
	if existingService.Name == t.Name && existingService.Price == t.Price {
		return nil, domain.ErrNoUpdatedData
	}

	// Update the type of service in the repository
	updated, err := s.repo.UpdateTypeOfService(ctx, t)
	if err != nil {
		if err == domain.ErrConflictingData {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	// Invalidate the cache
	cacheKey := util.GenerateCacheKey("typeofservice", t.ID)
	err = s.cache.Delete(ctx, cacheKey)
	if err != nil {
		return nil, domain.ErrInternal
	}

	// Cache the updated TypeOfService
	serialized, err := util.Serialize(updated)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = s.cache.Set(ctx, cacheKey, serialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	// Invalidate general cache for lists
	err = s.cache.DeleteByPrefix(ctx, "typeofservices:*")
	if err != nil {
		return nil, domain.ErrInternal
	}

	return updated, nil
}

// DeleteTypeOfService deletes a type of service by ID
func (s *TypeOfServiceService) DeleteTypeOfService(ctx context.Context, id uuid.UUID) error {
	// Check if the type of service exists
	_, err := s.repo.GetTypeOfServiceByID(ctx, id)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return err
		}
		return domain.ErrInternal
	}

	// Invalidate the cache for this type of service
	cacheKey := util.GenerateCacheKey("typeofservice", id)
	err = s.cache.Delete(ctx, cacheKey)
	if err != nil {
		return domain.ErrInternal
	}

	// Invalidate the general cache for lists of types of service
	err = s.cache.DeleteByPrefix(ctx, "typeofservices:*")
	if err != nil {
		return domain.ErrInternal
	}

	// Delete the type of service from the repository
	return s.repo.DeleteTypeOfService(ctx, id)
}
