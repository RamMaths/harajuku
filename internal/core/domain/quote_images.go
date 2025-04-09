package domain

import "github.com/google/uuid"

// QuoteImage represents the image associated with a quote
type QuoteImage struct {
	ID      uuid.UUID
	QuoteID uuid.UUID
	URL     string
}

// QuoteImageFilters represents the filters to use when fetching quote images
type QuoteImageFilters struct {
	QuoteID *uuid.UUID
}
