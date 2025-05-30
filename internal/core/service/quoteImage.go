package service

import (
	"context"
	"log/slog"

	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/adapter/storage/postgres/repository"
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"harajuku/backend/internal/core/util"

	"github.com/google/uuid"
)

type QuoteImageService struct {
	repo      port.QuoteImageRepository
	file      port.FileRepository
	quoteRepo port.QuoteRepository
	db        postgres.DB
	cache     port.CacheRepository
}

func NewQuoteImageService(
	repo port.QuoteImageRepository,
	file port.FileRepository,
	quoteRepo port.QuoteRepository,
	db postgres.DB,
	cache port.CacheRepository,
) *QuoteImageService {
	return &QuoteImageService{
		repo:      repo,
		file:      file,
		quoteRepo: quoteRepo,
		db:        db,
		cache:     cache,
	}
}

// CreateQuoteImage creates a new quote image and stores it in the database and file storage
func (qs *QuoteImageService) GetQuoteImageByID(ctx context.Context, id uuid.UUID) (*domain.QuoteImage, []byte, error) {
	var image *domain.QuoteImage
	cacheKey := util.GenerateCacheKey("quoteImage", id)

	cached, err := qs.cache.Get(ctx, cacheKey)
	if err == nil {
		err = util.Deserialize(cached, &image)
		if err == nil {
			file, err := qs.file.Get(ctx, image.URL)
			if err != nil {
				return nil, nil, domain.ErrInternal
			}
			return image, file, nil
		}
	}

	image, err = qs.repo.GetQuoteImageByID(ctx, id)
	if err != nil {
		return nil, nil, domain.ErrInternal
	}

	data, _ := util.Serialize(image)
	_ = qs.cache.Set(ctx, cacheKey, data, 0)

	file, err := qs.file.Get(ctx, image.URL)
	if err != nil {
		return nil, nil, domain.ErrInternal
	}

	return image, file, nil
}

// GetQuoteImages retrieves a list of quote images, optionally filtered by quoteID, and supports pagination
func (qs *QuoteImageService) GetQuoteImages(ctx context.Context, quoteID *uuid.UUID, skip, limit uint64) ([]domain.QuoteImage, error) {
	filters := domain.QuoteImageFilters{QuoteID: quoteID}
	cacheKey := util.GenerateCacheKey("quoteImages", struct {
		Filters domain.QuoteImageFilters
		Skip    uint64
		Limit   uint64
	}{filters, skip, limit})

	var images []domain.QuoteImage
	cached, err := qs.cache.Get(ctx, cacheKey)
	if err == nil {
		err = util.Deserialize(cached, &images)
		if err == nil {
			return images, nil
		}
	}

	images, err = qs.repo.GetQuoteImages(ctx, skip, limit, filters)
	if err != nil {
		return nil, domain.ErrInternal
	}

	data, _ := util.Serialize(images)
	_ = qs.cache.Set(ctx, cacheKey, data, 0)

	return images, nil
}

// DeleteQuoteImage deletes a quote image by its ID, removing it from both the database and file storage
func (qs *QuoteImageService) DeleteQuoteImage(ctx context.Context, id uuid.UUID) error {
	image, err := qs.repo.GetQuoteImageByID(ctx, id)
	if err != nil {
		return err
	}

	err = qs.db.WithTx(ctx, func(txDB *postgres.DB) error {
		txRepo := repository.NewQuoteImageRepository(txDB)

		if err := qs.file.Delete(ctx, image.URL); err != nil {
			return err
		}

		return txRepo.DeleteQuoteImage(ctx, id)
	})

	if err != nil {
		slog.Error("transaction failed", "error", err)
		return domain.ErrInternal
	}

	cacheKey := util.GenerateCacheKey("quoteImage", id)
	_ = qs.cache.Delete(ctx, cacheKey)
	_ = qs.cache.DeleteByPrefix(ctx, "quoteImages:*")

	return nil
}
