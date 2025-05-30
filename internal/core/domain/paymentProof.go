package domain

import "github.com/google/uuid"

// PaymentProof represents the proof of payment associated with a quote
type PaymentProof struct {
	ID         uuid.UUID
	QuoteID    uuid.UUID
	URL        string
	IsReviewed bool
}
