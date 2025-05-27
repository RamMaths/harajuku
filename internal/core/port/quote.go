package port

import (
	"context"
	"time"

	"harajuku/backend/internal/core/domain"

	"github.com/google/uuid"
)

//go:generate mockgen -source=quote.go -destination=mock/quote.go -package=mock

type QuoteFilter struct {
	TypeOfServiceID  *uuid.UUID
	ClientID     		 *uuid.UUID
	StartDate   *time.Time
	EndDate   	*time.Time
	ByState 		*domain.QuoteState
	Skip    		uint64
	Limit   		uint64
}

// QuoteRepository is an interface for interacting with quote-related data
type QuoteRepository interface {
	// CreateQuote inserts a new quote into the database
	CreateQuote(ctx context.Context, user *domain.Quote) (*domain.Quote, error)
	// GetQuoteByID selects a quote by id
	GetQuoteByID(ctx context.Context, id uuid.UUID) (*domain.Quote, error)
	// ListQuotes selects a list of quotes with pagination
	ListQuotes(ctx context.Context, filter QuoteFilter) ([]domain.Quote, error)
	// UpdateQuote updates a quote
	UpdateQuote(ctx context.Context, user *domain.Quote) (*domain.Quote, error)
	// DeleteQuote deletes a quote
	DeleteQuote(ctx context.Context, id uuid.UUID) error
  // Wrap a function in a DB transaction; if fn returns an error, rollback
  WithTx(ctx context.Context, fn func(repo QuoteRepository) error) error
}

// QuoteService is an interface for interacting with quote-related business logic
type QuoteService interface {
	// Creates a new quote
	CreateQuote(ctx context.Context, quote *domain.Quote, file []byte, fileName string) (*domain.Quote, error)
	// GetQuote returns a quote by id
	GetQuote(ctx context.Context, id uuid.UUID) (*domain.Quote, []domain.QuoteImage, error)
	// ListQuotes returns a list of quotes with pagination
	ListQuotes(ctx context.Context, filter QuoteFilter) ([]domain.Quote, error)
	// UpdateQuote updates a quote
	UpdateQuote(ctx context.Context, quote *domain.Quote) (*domain.Quote, error)
	// DeleteQuote deletes a quote
	DeleteQuote(ctx context.Context, id uuid.UUID) error
}
