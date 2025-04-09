package postgres

import (
	"context"
	"harajuku/backend/internal/core/domain"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type QuoteRepository struct {
	db *DB
}

func NewQuoteRepository(db *DB) *QuoteRepository {
	return &QuoteRepository{
		db,
	}
}

// CreateQuote creates a new quote in the database
func (r *QuoteRepository) CreateQuote(ctx context.Context, quote *domain.Quote) (*domain.Quote, error) {
	query := r.db.QueryBuilder.Insert("\"Quote\"").
		Columns("id", "\"typeOfServiceId\"", "\"clientId\"", "\"time\"", "\"description\"", "\"state\"", "\"price\"", "\"testRequired\"").
		Values(quote.ID, quote.TypeOfServiceID, quote.ClientID, quote.Time, quote.Description, quote.State, quote.Price, quote.TestRequired).
		Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(&quote.ID)
	if err != nil {
		return nil, err
	}

	return quote, nil
}

// GetQuoteByID retrieves a quote by ID from the database
func (r *QuoteRepository) GetQuoteByID(ctx context.Context, id uuid.UUID) (*domain.Quote, error) {
	var q domain.Quote

	query := r.db.QueryBuilder.Select("id", "\"typeOfServiceId\"", "\"clientId\"", "\"time\"", "\"description\"", "\"state\"", "\"price\"", "\"testRequired\"").
		From("\"Quote\"").
		Where(sq.Eq{"id": id}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(&q.ID, &q.TypeOfServiceID, &q.ClientID, &q.Time, &q.Description, &q.State, &q.Price, &q.TestRequired)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

	return &q, nil
}

// ListQuotes retrieves a list of quotes from the database
func (r *QuoteRepository) ListQuotes(ctx context.Context, skip, limit uint64) ([]domain.Quote, error) {
	var quotes []domain.Quote

	query := r.db.QueryBuilder.Select("id", "\"typeOfServiceId\"", "\"clientId\"", "time", "description", "state", "price", "\"testRequired\"").
		From("\"Quote\"").
		Limit(limit).
		Offset((skip - 1) * limit)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	// Debug logging - crucial for troubleshooting
	slog.DebugContext(ctx, "Executing query", "sql", sql, "args", args)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
	query := r.db.QueryBuilder.Update("\"Quote\"").
		Set("\"typeOfServiceId\"", quote.TypeOfServiceID).
		Set("\"clientId\"", quote.ClientID).
		Set("time", quote.Time).
		Set("description", quote.Description).
		Set("state", quote.State).
		Set("price", quote.Price).
		Set("\"testRequired\"", quote.TestRequired).
		Where(sq.Eq{"id": quote.ID}).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(&quote.ID, &quote.TypeOfServiceID, &quote.ClientID, &quote.Time, &quote.Description, &quote.State, &quote.Price, &quote.TestRequired)
	if err != nil {
		return nil, err
	}

	return quote, nil
}

// DeleteQuote deletes a quote by ID from the database
func (r *QuoteRepository) DeleteQuote(ctx context.Context, id uuid.UUID) error {
	query := r.db.QueryBuilder.Delete("\"Quote\"").
		Where(sq.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}
