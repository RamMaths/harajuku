package port

import (
	"context"
	"harajuku/backend/internal/core/domain"

	"github.com/google/uuid"
)

//go:generate mockgen -source=quoteImage.go -destination=mock/quoteImage.go -package=mock

// QuoteImageRepository is an interface for interacting with quote-image-related data
type QuoteImageRepository interface {
	// CreateQuoteImage inserts a new quote image into the database
	CreateQuoteImage(ctx context.Context, quoteImage *domain.QuoteImage) (*domain.QuoteImage, error)
	// UpdateQuoteImage updates an existing quote image
	UpdateQuoteImage(ctx context.Context, quoteImage *domain.QuoteImage) (*domain.QuoteImage, error)
	// DeleteQuoteImage deletes a quote image by ID
	DeleteQuoteImage(ctx context.Context, id uuid.UUID) error
	// GetQuoteImageByID selects a quote image by its ID
	GetQuoteImageByID(ctx context.Context, id uuid.UUID) (*domain.QuoteImage, error)
	// GetQuoteImages selects all quote images with optional filtering by QuoteID
	GetQuoteImages(ctx context.Context, skip, limit uint64, filters domain.QuoteImageFilters) ([]domain.QuoteImage, error)
}

// QuoteImageService is an interface for interacting with quote-image-related business logic
type QuoteImageService interface {
	// CreateQuoteImage creates a new quote image
	CreateQuoteImage(ctx context.Context, quoteImage *domain.QuoteImage) (*domain.QuoteImage, error)
	// UpdateQuoteImage updates an existing quote image
	UpdateQuoteImage(ctx context.Context, quoteImage *domain.QuoteImage) (*domain.QuoteImage, error)
	// DeleteQuoteImage deletes a quote image by its ID
	DeleteQuoteImage(ctx context.Context, id uuid.UUID) error
	// GetQuoteImageByID returns a quote image by its ID
	GetQuoteImageByID(ctx context.Context, id uuid.UUID) (*domain.QuoteImage, error)
	// GetQuoteImages returns a list of quote images, with optional filtering by quoteId
	GetQuoteImages(ctx context.Context, quoteId *uuid.UUID, skip, limit uint64) ([]domain.QuoteImage, error)
}
