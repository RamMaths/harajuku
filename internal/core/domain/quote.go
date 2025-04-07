package domain

import (
	"time"

	"github.com/google/uuid"
)

// QuoteState is an enum for the state of a quote
type QuoteState string

// QuoteState enum values
const (
	QuotePending       QuoteState = "pending"
	QuoteApproved      QuoteState = "approved"
	QuoteRejected      QuoteState = "rejected"
	QuoteRequiresProof QuoteState = "requires_proof"
)

// Quote is an entity that represents a service quote
type Quote struct {
	ID              uuid.UUID
	TypeOfServiceID uuid.UUID
	ClientID        uuid.UUID
	Time            time.Time
	Description     string
	State           QuoteState
	Price           float64
	TestRequired    bool
}
