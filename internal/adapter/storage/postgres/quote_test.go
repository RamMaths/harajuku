package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/core/domain"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", "postgres://postgres:123@127.0.0.1:5432/harajuku?sslmode=disable")
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	return db
}

func TestCreateQuote(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewQuoteRepository(db)
	ctx := context.Background()

	quote := &domain.Quote{
		ID:              uuid.New(),
		TypeOfServiceID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ClientID:        uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Time:            time.Now(),
		Description:     "Test create",
		State:           "pending",
		Price:           123.45,
		TestRequired:    true,
	}

	t.Logf("Quote creation with ID: %v", quote.ID)

	created, err := repo.CreateQuote(ctx, quote)
	if err != nil {
		t.Fatalf("CreateQuote failed: %v", err)
	}

	t.Logf("Created quote with ID: %v", created.ID)
}

func TestGetQuoteByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewQuoteRepository(db)
	ctx := context.Background()

	id := uuid.MustParse("ba040759-f58a-47f6-ba01-c77029fb35e6")

	t.Logf("Fetching quote with ID: %v", id)

	_, err := repo.GetQuoteByID(ctx, id)
	if err != nil {
		t.Fatalf("GetQuoteByID failed: %v", err)
	}
}

func TestListQuotes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewQuoteRepository(db)
	ctx := context.Background()

	t.Logf("Fetching list of quotes")

	quotes, err := repo.ListQuotes(ctx, 0, 10)
	if err != nil {
		t.Fatalf("ListQuotes failed: %v", err)
	}
	t.Logf("Retrieved %d quotes", len(quotes))
}

func TestUpdateQuote(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewQuoteRepository(db)
	ctx := context.Background()

	id := uuid.New()
	original := &domain.Quote{
		ID:              id,
		TypeOfServiceID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ClientID:        uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Time:            time.Now(),
		Description:     "Before update",
		State:           "pending",
		Price:           150.0,
		TestRequired:    true,
	}
	t.Logf("Creating original quote with ID: %v", id)
	_, err := repo.CreateQuote(ctx, original)
	if err != nil {
		t.Fatalf("CreateQuote failed: %v", err)
	}

	t.Logf("Updating quote with ID: %v", id)
	original.Description = "After update"
	original.State = "completed"
	original.Price = 300.0
	updated, err := repo.UpdateQuote(ctx, original)
	if err != nil {
		t.Fatalf("UpdateQuote failed: %v", err)
	}

	t.Logf("Updated quote with ID: %v", updated.ID)
}

func TestDeleteQuote(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewQuoteRepository(db)
	ctx := context.Background()

	id := uuid.MustParse("ba040759-f58a-47f6-ba01-c77029fb35e6")

	t.Logf("Deleting quote with ID: %v", id)
	err := repo.DeleteQuote(ctx, id)
	if err != nil {
		t.Fatalf("DeleteQuote failed: %v", err)
	}
}
