package domain

import (
  "github.com/google/uuid"
)

// UserRole is an enum for user's role
type AppointmentStatus string

// UserRole enum values
const (
	Booked   AppointmentStatus = "booked"
	Pending   AppointmentStatus = "pending"
	Cancelled AppointmentStatus = "cancelled"
	Completed AppointmentStatus = "completed"
)

// AvailabilitySlot is an entity that represents a user
type Appointment struct {
	ID            uuid.UUID
	UserID       	uuid.UUID
	SlotId     		uuid.UUID
	QuoteId       uuid.UUID
  Status      	AppointmentStatus
}
