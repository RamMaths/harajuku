package repository

import (
	"context"
	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// QuoteImageRepository implements port.QuoteImageRepository interface and provides access to the postgres database
type QuoteImageRepository struct {
	db *postgres.DB
}

// NewQuoteImageRepository creates a new quote image repository instance
func NewQuoteImageRepository(db *postgres.DB) *QuoteImageRepository {
	return &QuoteImageRepository{
		db,
	}
}

// CreateQuoteImage inserts a new quote image into the database
func (r *QuoteImageRepository) CreateQuoteImage(ctx context.Context, quoteImage *domain.QuoteImage) (*domain.QuoteImage, error) {
	query := r.db.QueryBuilder.Insert("\"QuoteImages\"").
		Columns("id", "\"quoteId\"", "url").
		Values(quoteImage.ID, quoteImage.QuoteID, quoteImage.URL).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(
		&quoteImage.ID,
		&quoteImage.QuoteID,
		&quoteImage.URL,
	)
	if err != nil {
		return nil, err
	}

	return quoteImage, nil
}

// GetQuoteImageByID selects a quote image by its ID
func (r *QuoteImageRepository) GetQuoteImageByID(ctx context.Context, id uuid.UUID) (*domain.QuoteImage, error) {
	var quoteImage domain.QuoteImage

	query := r.db.QueryBuilder.Select("*").
		From("\"QuoteImages\"").
		Where(sq.Eq{"id": id}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(
		&quoteImage.ID,
		&quoteImage.QuoteID,
		&quoteImage.URL,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

	return &quoteImage, nil
}

// GetQuoteImages selects all quote images with optional filtering by QuoteID
func (r *QuoteImageRepository) GetQuoteImages(ctx context.Context, skip, limit uint64, filters domain.QuoteImageFilters) ([]domain.QuoteImage, error) {
	var quoteImage domain.QuoteImage
	var quoteImages []domain.QuoteImage

	query := r.db.QueryBuilder.Select("*").
		From("\"QuoteImages\"").
		Limit(limit).
		Offset((skip - 1) * limit)

	// Apply filter for QuoteID
	if filters.QuoteID != nil {
		query = query.Where(sq.Eq{"\"quoteId\"": *filters.QuoteID})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	// Debug logging - useful for troubleshooting
	slog.DebugContext(ctx, "Executing query",
		"sql", sql,
		"args", args,
		"filters", filters)

	rows, err := r.db.Conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&quoteImage.ID,
			&quoteImage.QuoteID,
			&quoteImage.URL,
		)
		if err != nil {
			return nil, err
		}

		quoteImages = append(quoteImages, quoteImage)
	}

	return quoteImages, nil
}

// UpdateQuoteImage updates an existing quote image in the database
func (r *QuoteImageRepository) UpdateQuoteImage(ctx context.Context, quoteImage *domain.QuoteImage) (*domain.QuoteImage, error) {
	query := r.db.QueryBuilder.Update("\"QuoteImages\"").
		Set("\"quoteId\"", quoteImage.QuoteID).
		Set("url", quoteImage.URL).
		Where(sq.Eq{"id": quoteImage.ID}).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(
		&quoteImage.ID,
		&quoteImage.QuoteID,
		&quoteImage.URL,
	)
	if err != nil {
		return nil, err
	}

	return quoteImage, nil
}

// DeleteQuoteImage deletes a quote image by ID from the database
func (r *QuoteImageRepository) DeleteQuoteImage(ctx context.Context, id uuid.UUID) error {
	query := r.db.QueryBuilder.Delete("\"QuoteImages\"").
		Where(sq.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Conn.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *QuoteImageRepository) WithTx(
    ctx context.Context,
    fn func(repo port.QuoteImageRepository) error,
) error {
    return r.db.WithTx(ctx, func(txDB *postgres.DB) error {
        txRepo := NewQuoteImageRepository(txDB)
        return fn(txRepo)
    })
}
