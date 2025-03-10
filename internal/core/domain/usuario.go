package domain

import (
	"github.com/google/uuid"
)

// User is an entity that represents a user
type Usuario struct {
	ID uuid.UUID
	Nombre    string
	ApellidoPaterno   string
	ApellidoMaterno   string
  Correo    string
  Contrasenia   string
}
