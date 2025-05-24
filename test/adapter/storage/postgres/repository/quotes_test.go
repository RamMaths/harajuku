package repository

import (
	"context"
	"harajuku/backend/internal/adapter/config"
	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/test/adapter/storage/postgres/helpers"
	"testing"
	"time"

	"github.com/google/uuid"
)

const QUOTE_PORT = "5435"

func TestCreateQuoteIntegration(t *testing.T) {
	testContainer := helpers.SetupTestDB(t)
	defer testContainer.Teardown()

	ctx := context.Background()

	cfg := &config.DB{
		Connection: "postgres",
		User:       "user",
		Password:   "secret",
		Host:       "localhost",
		Port:       QUOTE_PORT,
		Name:       "testdb",
	}

	// overwrite the db URL to point to container
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// apply migrations
	err = db.Migrate()
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := postgres.NewQuoteRepository(db)

	// Insert test data
	quote := domain.Quote{
		ID:              uuid.New(),
		TypeOfServiceID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ClientID:        uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Time:            time.Now(),
		Description:     "Test Create Quote",
		State:           "pending",
		Price:           100.50,
	}

	createdQuote, err := repo.CreateQuote(ctx, &quote)
	if err != nil {
		t.Fatalf("failed to create quote: %v", err)
	}

	t.Logf("Created quote with ID: %v", createdQuote.ID)
}

func TestGetQuoteByIDIntegration(t *testing.T) {
	testContainer := helpers.SetupTestDB(t)
	defer testContainer.Teardown()

	ctx := context.Background()

	cfg := &config.DB{
		Connection: "postgres",
		User:       "user",
		Password:   "secret",
		Host:       "localhost",
		Port:       PORT,
		Name:       "testdb",
	}

	// overwrite the db URL to point to container
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// apply migrations
	err = db.Migrate()
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := postgres.NewQuoteRepository(db)

	// Insert test data
	quote := domain.Quote{
		ID:              uuid.New(),
		TypeOfServiceID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ClientID:        uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Time:            time.Now(),
		Description:     "Test Get Quote",
		State:           "pending",
		Price:           200.75,
	}

	_, err = repo.CreateQuote(ctx, &quote)
	if err != nil {
		t.Fatalf("failed to create quote: %v", err)
	}

	// Retrieve quote by ID
	retrievedQuote, err := repo.GetQuoteByID(ctx, quote.ID)
	if err != nil {
		t.Fatalf("failed to get quote by ID: %v", err)
	}

	t.Logf("Retrieved quote: %+v", retrievedQuote)
}

func TestListQuotesIntegration(t *testing.T) {
	testContainer := helpers.SetupTestDB(t)
	defer testContainer.Teardown()

	ctx := context.Background()

	cfg := &config.DB{
		Connection: "postgres",
		User:       "user",
		Password:   "secret",
		Host:       "localhost",
		Port:       PORT,
		Name:       "testdb",
	}

	// overwrite the db URL to point to container
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// apply migrations
	err = db.Migrate()
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := postgres.NewQuoteRepository(db)

	// Insert multiple test data
	quotes := []domain.Quote{
		{
			ID:              uuid.New(),
			TypeOfServiceID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			ClientID:        uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			Time:            time.Now(),
			Description:     "Quote 1",
			State:           "pending",
			Price:           100.00,
		},
		{
			ID:              uuid.New(),
			TypeOfServiceID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			ClientID:        uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			Time:            time.Now(),
			Description:     "Quote 2",
			State:           "approved",
			Price:           200.00,
		},
	}

	for _, q := range quotes {
		_, err := repo.CreateQuote(ctx, &q)
		if err != nil {
			t.Fatalf("failed to create quote: %v", err)
		}
	}

	// List quotes
	listedQuotes, err := repo.ListQuotes(ctx, 1, 10)
	if err != nil {
		t.Fatalf("failed to list quotes: %v", err)
	}

	if len(listedQuotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(listedQuotes))
	}

	t.Logf("Listed quotes: %+v", listedQuotes)
}

func TestUpdateQuoteIntegration(t *testing.T) {
	testContainer := helpers.SetupTestDB(t)
	defer testContainer.Teardown()

	ctx := context.Background()

	cfg := &config.DB{
		Connection: "postgres",
		User:       "user",
		Password:   "secret",
		Host:       "localhost",
		Port:       PORT,
		Name:       "testdb",
	}

	// overwrite the db URL to point to container
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// apply migrations
	err = db.Migrate()
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := postgres.NewQuoteRepository(db)

	// Insert test data
	quote := domain.Quote{
		ID:              uuid.New(),
		TypeOfServiceID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		ClientID:        uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
		Time:            time.Now(),
		Description:     "Test Update Quote",
		State:           "pending",
		Price:           50.25,
	}

	createdQuote, err := repo.CreateQuote(ctx, &quote)
	if err != nil {
		t.Fatalf("failed to create quote: %v", err)
	}

	// Update quote
	createdQuote.Description = "Updated description"
	createdQuote.Price = 150.75

	updatedQuote, err := repo.UpdateQuote(ctx, createdQuote)
	if err != nil {
		t.Fatalf("failed to update quote: %v", err)
	}

	t.Logf("Updated quote: %+v", updatedQuote)
}

func TestDeleteQuoteIntegration(t *testing.T) {
	testContainer := helpers.SetupTestDB(t)
	defer testContainer.Teardown()

	ctx := context.Background()

	cfg := &config.DB{
		Connection: "postgres",
		User:       "user",
		Password:   "secret",
		Host:       "localhost",
		Port:       PORT,
		Name:       "testdb",
	}

	// overwrite the db URL to point to container
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// apply migrations
	err = db.Migrate()
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := postgres.NewQuoteRepository(db)

	// Insert test data
	quote := domain.Quote{
		ID:              uuid.New(),
		TypeOfServiceID: uuid.MustParse("44444444-4444-4444-4444-444444444444"),
		ClientID:        uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"),
		Time:            time.Now(),
		Description:     "Test Delete Quote",
		State:           "pending",
		Price:           75.50,
	}

	createdQuote, err := repo.CreateQuote(ctx, &quote)
	if err != nil {
		t.Fatalf("failed to create quote: %v", err)
	}

	// Delete quote
	err = repo.DeleteQuote(ctx, createdQuote.ID)
	if err != nil {
		t.Fatalf("failed to delete quote: %v", err)
	}

	t.Logf("Deleted quote with ID: %v", createdQuote.ID)
}
