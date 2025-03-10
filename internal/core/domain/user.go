package domain

import "github.com/google/uuid"

// User is an entity that represents a user
type User struct {
	ID        uuid.UUID
	Name      string
	LastName      string
	SecondLastName      string
	Email     string
	Password  string
}
