package service

import (
	"context"

	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"harajuku/backend/internal/core/util"
)

/**
 * AuthService implements port.AuthService interface
 * and provides an access to the user repository
 * and token service
 */
type AuthService struct {
	repo port.UserRepository
	ts   port.TokenService
}

// NewAuthService creates a new auth service instance
func NewAuthService(repo port.UserRepository, ts port.TokenService) *AuthService {
	return &AuthService{
		repo,
		ts,
	}
}

// Login gives a registered user an access token if the credentials are valid
func (as *AuthService) Login(ctx context.Context, email, password string) (token string, role domain.UserRole, err error) {
	user, err := as.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return "", "", domain.ErrInvalidCredentials
		}
		return "", "", domain.ErrInvalidCredentials
	}

	err = util.ComparePassword(password, user.Password)
	if err != nil {
		return "", "", domain.ErrInvalidCredentials
	}

	accessToken, err := as.ts.CreateToken(user)
	if err != nil {
		return "", "", domain.ErrInvalidCredentials
	}

	return accessToken, user.Role, nil
}
