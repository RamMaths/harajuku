package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"harajuku/backend/internal/core/util"

	"github.com/google/uuid"
)

/**
 * QuoteService implements port.QuoteService interface
 * and provides an access to the quote repository
 * and cache service
 */
type QuoteService struct {
	repo  port.QuoteRepository
  file port.FileRepository
  user port.UserRepository
  email port.EmailRepository
  quoteImage port.QuoteImageRepository
	cache port.CacheRepository
}

// NewQuoteService creates a new quote service instance
func NewQuoteService(repo port.QuoteRepository, file port.FileRepository, user port.UserRepository, email port.EmailRepository, quoteImage port.QuoteImageRepository, cache port.CacheRepository) *QuoteService {
	return &QuoteService {
		repo,
    file,
    user,
    email,
    quoteImage,
		cache,
	}
}

// Register creates a new quote
func (us *QuoteService) CreateQuote(ctx context.Context, quote *domain.Quote, file []byte, fileName string) (*domain.Quote, error) {

  // QuoteRepository

  quote, err := us.repo.CreateQuote(ctx, quote)

	if err != nil {
    slog.Error("Quote registration failed", "error", err)
		if err == domain.ErrConflictingData {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

  // Cache

	cacheKey := util.GenerateCacheKey("quote", quote.ID)
	quoteSerialized, err := util.Serialize(quote)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = us.cache.Set(ctx, cacheKey, quoteSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = us.cache.DeleteByPrefix(ctx, "quotes:*")
	if err != nil {
		return nil, domain.ErrInternal
	}

  // Handle File Saving

  path, err := us.file.Save(ctx, file, fileName)

  if err != nil {
    return nil, domain.ErrInternal
  }

  // QuoteImage Repository

  _, err = us.quoteImage.CreateQuoteImage(ctx, &domain.QuoteImage{
    ID: uuid.New(),
    QuoteID: quote.ID,
    URL: path,
  })

  if err != nil {
    return nil, domain.ErrInternal
  }

  // Send email

  emails, err := us.user.GetAdminsEmails(ctx)

  if err != nil {
    return nil, domain.ErrInternal
  }

  client, err := us.user.GetUserByID(ctx, quote.ClientID)

  err = us.email.SendEmail(
    ctx,
    emails,
    "Se ha creado una nueva cotización",
    fmt.Sprintf(
      "Una nueva cotización se ha creado\n\t\tid: %s\n\t\tDescripción: %s\n\t\tCliente: %s %s",
      quote.ID.String(),
      quote.Description,
      client.Name,
      client.LastName,
    ),
    "",
  )

  if err != nil {
    // Log detailed error and continue since email failure doesn't block quote creation
    slog.Error("failed to send email notification for quote %s: %v", quote.ID.String(), err)
    return nil, domain.ErrInternal
  }

  // Return the new quote

	return quote, nil
}

// GetQuote gets a quote by ID
func (us *QuoteService) GetQuote(ctx context.Context, id uuid.UUID) (*domain.Quote, error) {
	var quote *domain.Quote

	cacheKey := util.GenerateCacheKey("quote", id)
	cachedQuote, err := us.cache.Get(ctx, cacheKey)
	if err == nil {
		err := util.Deserialize(cachedQuote, &quote)
		if err != nil {
			return nil, domain.ErrInternal
		}
		return quote, nil
	}

	quote, err = us.repo.GetQuoteByID(ctx, id)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	quoteSerialized, err := util.Serialize(quote)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = us.cache.Set(ctx, cacheKey, quoteSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return quote, nil
}

// ListQuotes lists all quotes
func (us *QuoteService) ListQuotes(ctx context.Context, skip, limit uint64) ([]domain.Quote, error) {
	var quotes []domain.Quote

	params := util.GenerateCacheKeyParams(skip, limit)
	cacheKey := util.GenerateCacheKey("quotes", params)

	cachedQuotes, err := us.cache.Get(ctx, cacheKey)
	if err == nil {
		err := util.Deserialize(cachedQuotes, &quotes)
		if err != nil {
			return nil, domain.ErrInternal
		}
		return quotes, nil
	}

	quotes, err = us.repo.ListQuotes(ctx, skip, limit)
	if err != nil {
		return nil, domain.ErrInternal
	}

	quotesSerialized, err := util.Serialize(quotes)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = us.cache.Set(ctx, cacheKey, quotesSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return quotes, nil
}

// UpdateQuote updates a quote's content, author, and associated metadata.
func (us *QuoteService) UpdateQuote(ctx context.Context, quote *domain.Quote) (*domain.Quote, error) {
	existingQuote, err := us.repo.GetQuoteByID(ctx, quote.ID)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

  // Create zero values for comparison
  zeroUUID := uuid.UUID{}
  zeroTime := time.Time{}

  emptyData := quote.TypeOfServiceID == zeroUUID &&
    quote.ClientID == zeroUUID &&
    quote.Time == zeroTime &&
    quote.Description == "" &&
    quote.State == "" &&
    quote.Price == 0.0 &&
    !quote.TestRequired

  sameData := existingQuote.TypeOfServiceID == quote.TypeOfServiceID &&
    existingQuote.ClientID == quote.ClientID &&
    existingQuote.Time.Equal(quote.Time) &&
    existingQuote.Description == quote.Description &&
    existingQuote.State == quote.State &&
    existingQuote.Price == quote.Price &&
    existingQuote.TestRequired == quote.TestRequired

  if emptyData || sameData {
      return nil, domain.ErrNoUpdatedData
  }

	_, err = us.repo.UpdateQuote(ctx, quote)
	if err != nil {
		if err == domain.ErrConflictingData {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	cacheKey := util.GenerateCacheKey("quote", quote.ID)

	err = us.cache.Delete(ctx, cacheKey)
	if err != nil {
		return nil, domain.ErrInternal
	}

	quoteSerialized, err := util.Serialize(quote)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = us.cache.Set(ctx, cacheKey, quoteSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = us.cache.DeleteByPrefix(ctx, "quotes:*")
	if err != nil {
		return nil, domain.ErrInternal
	}

	return quote, nil
}

// DeleteQuote deletes a quote by ID
func (us *QuoteService) DeleteQuote(ctx context.Context, id uuid.UUID) error {
	_, err := us.repo.GetQuoteByID(ctx, id)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return err
		}
		return domain.ErrInternal
	}

	cacheKey := util.GenerateCacheKey("quote", id)

	err = us.cache.Delete(ctx, cacheKey)
	if err != nil {
		return domain.ErrInternal
	}

	err = us.cache.DeleteByPrefix(ctx, "quotes:*")
	if err != nil {
		return domain.ErrInternal
	}

	return us.repo.DeleteQuote(ctx, id)
}
