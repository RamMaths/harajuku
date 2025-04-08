package port

import (
	"context"

	"harajuku/backend/internal/core/domain"

	"github.com/google/uuid"
)

//go:generate mockgen -source=quote.go -destination=mock/quote.go -package=mock

// QuoteRepository is an interface for interacting with quote-related data
type QuoteRepository interface {
	// CreateQuote inserts a new quote into the database
	CreateQuote(ctx context.Context, user *domain.Quote) (*domain.Quote, error)
	// GetQuoteByID selects a quote by id
	GetQuoteByID(ctx context.Context, id uuid.UUID) (*domain.Quote, error)
	// ListQuotes selects a list of quotes with pagination
	ListQuotes(ctx context.Context, skip, limit uint64) ([]domain.Quote, error)
	// UpdateQuote updates a quote
	UpdateQuote(ctx context.Context, user *domain.Quote) (*domain.Quote, error)
	// DeleteQuote deletes a quote
	DeleteQuote(ctx context.Context, id uuid.UUID) error
}

// QuoteService is an interface for interacting with quote-related business logic
type QuoteService interface {
	// Creates a new quote
	CreateQuote(ctx context.Context, quote *domain.Quote) (*domain.Quote, error)
	// GetQuote returns a quote by id
	GetQuote(ctx context.Context, id uuid.UUID) (*domain.Quote, error)
	// ListQuotes returns a list of quotes with pagination
	ListQuotes(ctx context.Context, skip, limit uint64) ([]domain.Quote, error)
	// UpdateQuote updates a quote
	UpdateQuote(ctx context.Context, quote *domain.Quote) (*domain.Quote, error)
	// DeleteQuote deletes a quote
	DeleteQuote(ctx context.Context, id uuid.UUID) error
}
