package service

import (
	"context"
	"log/slog"
	"time"

	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"harajuku/backend/internal/core/util"

	"github.com/google/uuid"
)

/**
 * UserService implements port.UserService interface
 * and provides an access to the user repository
 * and cache service
 */
type UserService struct {
	repo  port.UserRepository
	cache port.CacheRepository
}

// NewUserService creates a new user service instance
func NewUserService(repo port.UserRepository, cache port.CacheRepository) *UserService {
	return &UserService{
		repo,
		cache,
	}
}

// Register creates a new user
func (us *UserService) Register(ctx context.Context, user *domain.User) (*domain.User, error) {
	hashedPassword, err := util.HashPassword(user.Password)
	if err != nil {
		return nil, domain.ErrInternal
      }

	user.Password = hashedPassword

	user, err = us.repo.CreateUser(ctx, user)
	if err != nil {
    slog.Error("User registration failed", "error", err)
		if err == domain.ErrConflictingData {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	cacheKey := util.GenerateCacheKey("user", user.ID)
	userSerialized, err := util.Serialize(user)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = us.cache.Set(ctx, cacheKey, userSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = us.cache.DeleteByPrefix(ctx, "users:*")
	if err != nil {
		return nil, domain.ErrInternal
	}

	return user, nil
}

// GetUser gets a user by ID
func (us *UserService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user *domain.User

	cacheKey := util.GenerateCacheKey("user", id)
	cachedUser, err := us.cache.Get(ctx, cacheKey)
	if err == nil {
		err := util.Deserialize(cachedUser, &user)
		if err != nil {
			return nil, domain.ErrInternal
		}
		return user, nil
	}

	user, err = us.repo.GetUserByID(ctx, id)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	userSerialized, err := util.Serialize(user)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = us.cache.Set(ctx, cacheKey, userSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return user, nil
}

// ListUsers lists all users
func (us *UserService) ListUsers(ctx context.Context, skip, limit uint64, filters domain.UserFilters) ([]domain.User, error) {
    var users []domain.User

    // Include filters in cache key
    params := util.GenerateCacheKeyParams(skip, limit, filters)
    cacheKey := util.GenerateCacheKey("users", params)

    // Try cache first
    cachedUsers, err := us.cache.Get(ctx, cacheKey)
    if err == nil {
        err := util.Deserialize(cachedUsers, &users)
        if err != nil {
            slog.Error("cache deserialization failed", "error", err)
            // Fall through to database query
        } else {
            return users, nil
        }
    }

    // Cache miss - query database
    users, err = us.repo.ListUsers(ctx, skip, limit, filters)
    if err != nil {
        return nil, domain.ErrInternal
    }

    // Cache results
    usersSerialized, err := util.Serialize(users)
    if err != nil {
        slog.Error("serialization failed", "error", err)
        return users, nil // Return results without caching
    }

    if err := us.cache.Set(ctx, cacheKey, usersSerialized, 10*time.Minute); err != nil {
        slog.Error("cache set failed", "error", err)
    }

    return users, nil
}

// UpdateUser updates a user's name, email, and password
func (us *UserService) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	existingUser, err := us.repo.GetUserByID(ctx, user.ID)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	emptyData := user.Name == "" &&
		user.LastName == "" &&
		user.SecondLastName == "" &&
		user.Email == "" &&
		user.Password == "" &&
    user.Role == ""

	sameData := existingUser.Name == user.Name &&
		existingUser.LastName == user.LastName &&
		existingUser.SecondLastName == user.SecondLastName &&
		existingUser.Email == user.Email &&
    existingUser.Role == user.Role
	if emptyData || sameData {
		return nil, domain.ErrNoUpdatedData
	}

	var hashedPassword string

	if user.Password != "" {
		hashedPassword, err = util.HashPassword(user.Password)
		if err != nil {
			return nil, domain.ErrInternal
		}
	}

	user.Password = hashedPassword

	_, err = us.repo.UpdateUser(ctx, user)
	if err != nil {
		if err == domain.ErrConflictingData {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	cacheKey := util.GenerateCacheKey("user", user.ID)

	err = us.cache.Delete(ctx, cacheKey)
	if err != nil {
		return nil, domain.ErrInternal
	}

	userSerialized, err := util.Serialize(user)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = us.cache.Set(ctx, cacheKey, userSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = us.cache.DeleteByPrefix(ctx, "users:*")
	if err != nil {
		return nil, domain.ErrInternal
	}

	return user, nil
}

// DeleteUser deletes a user by ID
func (us *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := us.repo.GetUserByID(ctx, id)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return err
		}
		return domain.ErrInternal
	}

	cacheKey := util.GenerateCacheKey("user", id)

	err = us.cache.Delete(ctx, cacheKey)
	if err != nil {
		return domain.ErrInternal
	}

	err = us.cache.DeleteByPrefix(ctx, "users:*")
	if err != nil {
		return domain.ErrInternal
	}

	return us.repo.DeleteUser(ctx, id)
}
