package domain

import "github.com/google/uuid"

// TypeOfService is an entity that represents a type of service
type TypeOfService struct {
	ID    uuid.UUID
	Name  string
	Price float64
}
