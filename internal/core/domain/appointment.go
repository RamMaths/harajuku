package domain

import "github.com/google/uuid"

// Status is an enum for appointment's status
type Status string

// Status enum values
const (
	Booked      Status = "booked"
	Cancelled   Status = "cancelled"
	Completed   Status = "completed"
)

// Appointment is an entity that represents a user
type Appointment struct {
	ID          uuid.UUID
	ClientID    uuid.UUID
	SlotID      uuid.UUID
	QuoteID     uuid.UUID
  Status      Status
}
