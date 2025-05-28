package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/adapter/storage/postgres/repository"
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
	repo       			port.QuoteRepository
	file       			port.FileRepository
	user       			port.UserRepository
	email      			port.EmailRepository
	quoteImage 			port.QuoteImageRepository
	typeOfService 	port.TypeOfServiceRepository
	db 							postgres.DB
	cache      			port.CacheRepository
}

// NewQuoteService creates a new quote service instance
func NewQuoteService(
	repo port.QuoteRepository,
	file port.FileRepository,
	user port.UserRepository,
	email port.EmailRepository,
	quoteImage port.QuoteImageRepository,
	typeOfService port.TypeOfServiceRepository,
	db postgres.DB,
	cache port.CacheRepository,
) *QuoteService {
	return &QuoteService{
		repo,
		file,
		user,
		email,
		quoteImage,
		typeOfService,
		db,
		cache,
	}
}

// Register creates a new quote
func (us *QuoteService) CreateQuote(ctx context.Context, quote *domain.Quote, file []byte, fileName string) (*domain.Quote, error) {
		// 1) Validate IDs
		_, err := us.typeOfService.GetTypeOfServiceByID(ctx, quote.TypeOfServiceID)

		if err != nil {
			return nil, domain.ErrDataNotFound
		}

		_, err = us.user.GetUserByID(ctx, quote.ClientID)

		if err != nil {
			return nil, domain.ErrDataNotFound
		}

		// 2) Upload the image first. Fail fast if this errors.
    path, err := us.file.Save(ctx, file, fileName)
    if err != nil {
        slog.Error("file save failed", "error", err)
        return nil, domain.ErrInternal
    }

    // 3) Atomic DB transaction: create quote + image row
    var created *domain.Quote

		err = us.db.WithTx(ctx, func(txDB *postgres.DB) error {
			// build both repos on txDB
			txQuoteRepo := repository.NewQuoteRepository(txDB)
			txImageRepo := repository.NewQuoteImageRepository(txDB)

			// insert quote
			q, err := txQuoteRepo.CreateQuote(ctx, quote)
			if err != nil {
				return err
			}
			created = q

			// insert image in *same* tx
			_, err = txImageRepo.CreateQuoteImage(ctx, &domain.QuoteImage{
				ID:      uuid.New(),
				QuoteID: created.ID,
				URL:     path,
			})

			return err
		})
   
    if err != nil {
        slog.Error("transaction failed", "error", err)
				err := us.file.Delete(ctx, path)
				if err != nil {
					slog.Error("deleting file failed", "error", err)
					return nil, domain.ErrInternal
				}

        return nil, domain.ErrInternal
    }


    // 4) Cache the new quote (best-effort)
    cacheKey := util.GenerateCacheKey("quote", created.ID)
    data, _ := util.Serialize(created)
    if err := us.cache.Set(ctx, cacheKey, data, 0); err != nil {
        slog.Warn("cache set failed", "error", err)
    }
		err = us.cache.DeleteByPrefix(ctx, "quotes:*")
		if err != nil {
			return nil, domain.ErrInternal
		}
		err = us.cache.DeleteByPrefix(ctx, "quoteImages:*")
		if err != nil {
			return nil, domain.ErrInternal
		}

    // 5) Notify admins (best-effort)
    emails, err := us.user.GetAdminsEmails(ctx)

    if err == nil {
        client, _ := us.user.GetUserByID(ctx, created.ClientID)
        if err := us.email.SendEmail(
            ctx,
            emails,
            "Se ha creado una nueva cotización",
            fmt.Sprintf(
                "Una nueva cotización se ha creado\n\tid: %s\n\tDescripción: %s\n\tCliente: %s %s",
                created.ID, created.Description, client.Name, client.LastName,
            ),
            "",
        ); err != nil {
            slog.Warn("email send failed", "quote_id", created.ID, "error", err)
        }
    } else {
        slog.Warn("could not fetch admin emails", "error", err)
    }

    return created, nil
}

// GetQuote gets a quote by ID
func (us *QuoteService) GetQuote(ctx context.Context, id uuid.UUID) (*domain.Quote, []domain.QuoteImage, error) {
	var quote *domain.Quote
	cacheKey:= util.GenerateCacheKey("quote", id)
	cachedQuote, err := us.cache.Get(ctx, cacheKey)
	if err == nil {
		err := util.Deserialize(cachedQuote, &quote)
		if err != nil {
			return nil, nil, domain.ErrInternal
		}
	}

	var images []domain.QuoteImage
	cacheKeyQuoteImage := util.GenerateCacheKey("quoteImages", id)
	cachedImages, err := us.cache.Get(ctx, cacheKeyQuoteImage)
	if err == nil {
		err := util.Deserialize(cachedImages, &images)
		if err != nil {
			return nil, nil, domain.ErrInternal
		}
	}

	if quote != nil && images != nil {
		return quote, images, nil
	}

	//

	quote, err = us.repo.GetQuoteByID(ctx, id)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return nil, nil, err
		}
		return nil, nil, domain.ErrInternal
	}

	quoteSerialized, err := util.Serialize(quote)
	if err != nil {
		return nil, nil, domain.ErrInternal
	}

	err = us.cache.Set(ctx, cacheKey, quoteSerialized, 0)
	if err != nil {
		return nil, nil, domain.ErrInternal
	}

	//

	images, err = us.quoteImage.GetQuoteImages(ctx, 1, 10, domain.QuoteImageFilters{QuoteID: &quote.ID})
	if err != nil {
		return nil, nil, domain.ErrInternal
	}

	quoteImageSerialized, err := util.Serialize(images)
	if err != nil {
		return nil, nil, domain.ErrInternal
	}

	err = us.cache.Set(ctx, cacheKeyQuoteImage, quoteImageSerialized, 0)
	if err != nil {
		return nil, nil, domain.ErrInternal
	}

	return quote, images, nil
}

// ListQuotes lists all quotes
func (us *QuoteService) ListQuotes(ctx context.Context, filter port.QuoteFilter) ([]domain.Quote, error) {
	var quotes []domain.Quote

	params := util.GenerateCacheKeyParams(filter)
	cacheKey := util.GenerateCacheKey("quotes", params)

	cachedQuotes, err := us.cache.Get(ctx, cacheKey)
	if err == nil {
		err := util.Deserialize(cachedQuotes, &quotes)
		if err != nil {
			return nil, domain.ErrInternal
		}
		return quotes, nil
	}

	quotes, err = us.repo.ListQuotes(ctx, filter)
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
		quote.Price == 0.0

	sameData := existingQuote.TypeOfServiceID == quote.TypeOfServiceID &&
		existingQuote.ClientID == quote.ClientID &&
		existingQuote.Time.Equal(quote.Time) &&
		existingQuote.Description == quote.Description &&
		existingQuote.State == quote.State &&
		existingQuote.Price == quote.Price

	if emptyData || sameData {
		return nil, domain.ErrNoUpdatedData
	}

	quote, err = us.repo.UpdateQuote(ctx, quote)
	if err != nil {
		if err == domain.ErrConflictingData {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

  if quote.State == domain.QuoteRequiresProof {
    client, err := us.user.GetUserByID(ctx, quote.ClientID)

    if err != nil {
      if err == domain.ErrDataNotFound {
        return nil, err
      }
      return nil, domain.ErrInternal
    }

    emails := []string{client.Email}

    if err := us.email.SendEmail(
      ctx,
      emails,
      "Resupesta a su cotización",
      fmt.Sprintf(
        "Estimado cliente su Cotización requiere una prueba de mechón, para esto necesitamos que realice una cita en nuestro sistema.",
      ),
      "",
    ); err != nil {
      slog.Warn("email send failed", "quote_id", quote.ID, "error", err)
    }
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
	err = us.cache.DeleteByPrefix(ctx, "quoteImages:*")
	if err != nil {
		return nil, domain.ErrInternal
	}


	return quote, nil
}

// DeleteQuote deletes a quote by ID
func (us *QuoteService) DeleteQuote(ctx context.Context, id uuid.UUID) error {
	quote, err := us.repo.GetQuoteByID(ctx, id)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return err
		}
		return domain.ErrInternal
	}

	images, err := us.quoteImage.GetQuoteImages(ctx, 1, 10, domain.QuoteImageFilters{QuoteID: &quote.ID})
	if err != nil {
		return domain.ErrInternal
	}

	err = us.db.WithTx(ctx, func(txDB *postgres.DB) error {
		// build both repos on txDB
		txQuoteRepo := repository.NewQuoteRepository(txDB)
		txImageRepo := repository.NewQuoteImageRepository(txDB)
		for _, image := range images {
			err = us.file.Delete(ctx, image.URL) 
			if err != nil {
				return err
			}
			err := txImageRepo.DeleteQuoteImage(ctx, image.ID)
			if err != nil {
				return err
			}
		}

		err = txQuoteRepo.DeleteQuote(ctx, quote.ID)
		return err
	})
 
	if err != nil {
			slog.Error("transaction failed", "error", err)
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
	err = us.cache.DeleteByPrefix(ctx, "quoteImages:*")
	if err != nil {
		return domain.ErrInternal
	}

	return us.repo.DeleteQuote(ctx, id)
}
