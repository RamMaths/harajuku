package repository_test

import (
	"context"
	"testing"

	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/adapter/storage/postgres/repository"
	"harajuku/backend/internal/core/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *postgres.DB {
	// Aquí estableces la conexión con la base de datos
	connConfig := "postgres://postgres:123@127.0.0.1:5432/harajuku?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connConfig)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}

	// Inicializa el QueryBuilder usando squirrel
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Crea y devuelve una instancia de postgres.DB
	return &postgres.DB{
		Pool:         pool,
		QueryBuilder: &psql,
	}
}

func TestCreateQuoteImage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewQuoteImageRepository(db)
	ctx := context.Background()

	quoteImage := &domain.QuoteImage{
		ID:      uuid.New(),
		QuoteID: uuid.MustParse("66d54352-e614-427c-bba5-7c8092a6147e"),
		URL:     "https://example.com/image.jpg",
	}

	t.Logf("Creating quote image with ID: %v", quoteImage.ID)

	created, err := repo.CreateQuoteImage(ctx, quoteImage)
	if err != nil {
		t.Fatalf("CreateQuoteImage failed: %v", err)
	}

	t.Logf("Created quote image with ID: %v", created.ID)
}

func TestGetQuoteImageByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewQuoteImageRepository(db)
	ctx := context.Background()

	id := uuid.MustParse("193f3b8e-fa67-405b-9f34-5031331343d3")

	t.Logf("Fetching quote image with ID: %v", id)

	quoteImage, err := repo.GetQuoteImageByID(ctx, id)
	if err != nil {
		t.Fatalf("GetQuoteImageByID failed: %v", err)
	}

	t.Logf("Fetched quote image with ID: %v, URL: %v", quoteImage.ID, quoteImage.URL)
}

func TestGetQuoteImages(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewQuoteImageRepository(db)
	ctx := context.Background()

	t.Logf("Fetching list of quote images")

	filters := domain.QuoteImageFilters{
		QuoteID: func() *uuid.UUID {
			id := uuid.MustParse("66d54352-e614-427c-bba5-7c8092a6147e")
			return &id
		}(),
	}

	quoteImages, err := repo.GetQuoteImages(ctx, 1, 10, filters)
	if err != nil {
		t.Fatalf("GetQuoteImages failed: %v", err)
	}
	t.Logf("Retrieved %d quote images", len(quoteImages))
}

func TestUpdateQuoteImage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewQuoteImageRepository(db)
	ctx := context.Background()

	id := uuid.New()
	original := &domain.QuoteImage{
		ID:      id,
		QuoteID: uuid.MustParse("66d54352-e614-427c-bba5-7c8092a6147e"),
		URL:     "https://example.com/original-image.jpg",
	}

	t.Logf("Creating original quote image with ID: %v", id)
	_, err := repo.CreateQuoteImage(ctx, original)
	if err != nil {
		t.Fatalf("CreateQuoteImage failed: %v", err)
	}

	original.URL = "https://example.com/updated-imageeeee.jpg"
	t.Logf("Updating quote image with ID: %v", id)

	updated, err := repo.UpdateQuoteImage(ctx, original)
	if err != nil {
		t.Fatalf("UpdateQuoteImage failed: %v", err)
	}

	t.Logf("Updated quote image with ID: %v, URL: %v", updated.ID, updated.URL)
}

func TestDeleteQuoteImage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewQuoteImageRepository(db)
	ctx := context.Background()

	id := uuid.MustParse("95c29f21-8205-4c25-8536-6ba3c10c75e8")

	t.Logf("Deleting quote image with ID: %v", id)
	err := repo.DeleteQuoteImage(ctx, id)
	if err != nil {
		t.Fatalf("DeleteQuoteImage failed: %v", err)
	}
}
