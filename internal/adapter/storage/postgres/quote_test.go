package postgres_test

import (
	"context"
	"testing"
	"time"

	"harajuku/backend/internal/adapter/storage/postgres"
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

func TestCreateQuote(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewQuoteRepository(db)
	ctx := context.Background()

	quote := &domain.Quote{
		ID:              uuid.New(),
		TypeOfServiceID: uuid.MustParse("b3f9e12e-74b7-4f64-a57f-3943f7065b3e"),
		ClientID:        uuid.MustParse("0f75e0b1-134e-42f9-92ae-d1f0e3a63a1f"),
		Time:            time.Now(),
		Description:     "Test create",
		State:           "pending",
		Price:           123.45,
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

	id := uuid.MustParse("5bdf6192-d3a1-4579-9614-39caad8b4397")

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

	quotes, err := repo.ListQuotes(ctx, 1, 10)
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
		TypeOfServiceID: uuid.MustParse("b3f9e12e-74b7-4f64-a57f-3943f7065b3e"),
		ClientID:        uuid.MustParse("0f75e0b1-134e-42f9-92ae-d1f0e3a63a1f"),
		Time:            time.Now(),
		Description:     "Before update",
		State:           "pending",
		Price:           150.0,
	}
	t.Logf("Creating original quote with ID: %v", id)
	_, err := repo.CreateQuote(ctx, original)
	if err != nil {
		t.Fatalf("CreateQuote failed: %v", err)
	}

	t.Logf("Updating quote with ID: %v", id)
	original.Description = "After update"
	original.State = "approved"
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

	id := uuid.MustParse("5bdf6192-d3a1-4579-9614-39caad8b4397")

	t.Logf("Deleting quote with ID: %v", id)
	err := repo.DeleteQuote(ctx, id)
	if err != nil {
		t.Fatalf("DeleteQuote failed: %v", err)
	}
}
