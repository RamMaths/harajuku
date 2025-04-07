package domain

import "github.com/google/uuid"

// QuoteImage is an entity that represents a images associated with a quote
type QuoteImage struct {
	ID      uuid.UUID
	QuoteID uuid.UUID
	URL     string
}
