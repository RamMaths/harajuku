package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// QuoteState is an enum for the state of a quote
type QuoteState string

func (q QuoteState) String() string {
	return string(q)
}

// QuoteState enum values
const (
	QuotePending       QuoteState = "pending"
	QuoteApproved      QuoteState = "approved"
	QuoteRejected      QuoteState = "rejected"
	QuoteRequiresProof QuoteState = "requires_proof"
)

// IsValidState checks if a QuoteState is valid
func (q QuoteState) IsValidState() bool {
	switch q {
	case QuotePending, QuoteApproved, QuoteRejected, QuoteRequiresProof:
		return true
	}
	return false
}

// SetState sets the state of the quote and validates it
func (q *Quote) SetState(state QuoteState) error {
	if !state.IsValidState() {
		return errors.New("invalid state for quote")
	}
	q.State = state
	return nil
}

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
