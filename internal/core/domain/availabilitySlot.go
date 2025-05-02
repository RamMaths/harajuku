package domain

import (
  "github.com/google/uuid"
  "time"
)

// AvailabilitySlot is an entity that represents a user
type AvailabilitySlot struct {
	ID            uuid.UUID
	AdminID       uuid.UUID
	StartTime     time.Time
	EndTime       time.Time
  IsBooked      bool
}
