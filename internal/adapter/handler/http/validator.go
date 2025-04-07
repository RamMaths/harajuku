package http

import (
	"harajuku/backend/internal/core/domain"
	"github.com/go-playground/validator/v10"
)

// userRoleValidator is a custom validator for validating user roles
var userRoleValidator validator.Func = func(fl validator.FieldLevel) bool {
	userRole := fl.Field().Interface().(domain.UserRole)

	switch userRole {
	case "admin", "client":
		return true
	default:
		return false
	}
}
