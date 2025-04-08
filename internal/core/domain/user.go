package domain

import "github.com/google/uuid"

// UserRole is an enum for user's role
type UserRole string

// UserRole enum values
const (
	Admin   UserRole = "admin"
	Client UserRole = "client"
)

// UserFilters contains filter criteria for listing users
type UserFilters struct {
    Name            string
    LastName        string
    SecondLastName  string
    Role            UserRole
}

// User is an entity that represents a user
type User struct {
	ID        uuid.UUID
	Name      string
	LastName      string
	SecondLastName      string
	Role      UserRole
	Email     string
	Password  string
}
