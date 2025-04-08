package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"

	"github.com/google/uuid"
)

type QuoteRepository struct {
	db *sql.DB
}

func NewQuoteRepository(db *sql.DB) port.QuoteRepository {
	return &QuoteRepository{db: db}
}

// CreateQuote creates a new quote in the database
func (r *QuoteRepository) CreateQuote(ctx context.Context, quote *domain.Quote) (*domain.Quote, error) {
	query := `
        INSERT INTO "Quote" ("id", "typeOfServiceId", "clientId", "time", "description", "state", "price", "testRequired")
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING "id"
    `
	err := r.db.QueryRowContext(ctx, query, quote.ID, quote.TypeOfServiceID, quote.ClientID, quote.Time, quote.Description, quote.State, quote.Price, quote.TestRequired).Scan(&quote.ID)
	if err != nil {
		return nil, err
	}
	return quote, nil
}

// GetQuoteByID retrieves a quote by ID from the database
func (r *QuoteRepository) GetQuoteByID(ctx context.Context, id uuid.UUID) (*domain.Quote, error) {
	query := `
        SELECT "id", "typeOfServiceId", "clientId", "time", "description", "state", "price", "testRequired"
        FROM "Quote"
        WHERE "id" = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	var q domain.Quote
	err := row.Scan(&q.ID, &q.TypeOfServiceID, &q.ClientID, &q.Time, &q.Description, &q.State, &q.Price, &q.TestRequired)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}
	return &q, nil
}

// ListQuotes retrieves a list of quotes from the database
func (r *QuoteRepository) ListQuotes(ctx context.Context, skip, limit uint64) ([]domain.Quote, error) {
	query := `
        SELECT "id", "typeOfServiceId", "clientId", "time", "description", "state", "price", "testRequired"
        FROM "Quote"
        LIMIT $1 OFFSET $2
    `
	rows, err := r.db.QueryContext(ctx, query, limit, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotes []domain.Quote
	for rows.Next() {
		var q domain.Quote
		if err := rows.Scan(&q.ID, &q.TypeOfServiceID, &q.ClientID, &q.Time, &q.Description, &q.State, &q.Price, &q.TestRequired); err != nil {
			return nil, err
		}
		quotes = append(quotes, q)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return quotes, nil
}

// UpdateQuote updates an existing quote in the database
func (r *QuoteRepository) UpdateQuote(ctx context.Context, quote *domain.Quote) (*domain.Quote, error) {
	query := `
        UPDATE "Quote" SET
            "typeOfServiceId" = $1,
            "clientId" = $2,
            "time" = $3,
            "description" = $4,
            "state" = $5,
            "price" = $6,
            "testRequired" = $7
        WHERE "id" = $8
    `
	_, err := r.db.ExecContext(ctx, query, quote.TypeOfServiceID, quote.ClientID, quote.Time, quote.Description, quote.State, quote.Price, quote.TestRequired, quote.ID)
	if err != nil {
		return nil, err
	}
	return quote, nil
}

// DeleteQuote deletes a quote by ID from the database
func (r *QuoteRepository) DeleteQuote(ctx context.Context, id uuid.UUID) error {
	var exists bool
	queryCheck := `SELECT EXISTS(SELECT 1 FROM "Quote" WHERE "id" = $1)`
	err := r.db.QueryRowContext(ctx, queryCheck, id).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("quote with id %v does not exist", id)
	}

	query := `DELETE FROM "Quote" WHERE "id" = $1`
	_, err = r.db.ExecContext(ctx, query, id)
	return err
}
