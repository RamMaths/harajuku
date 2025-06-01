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

// PaymentProofRepository implements port.PaymentProofRepository interface and provides access to the postgres database
type PaymentProofRepository struct {
	db *postgres.DB
}

// NewPaymentProofRepository creates a new payment proof repository instance
func NewPaymentProofRepository(db *postgres.DB) *PaymentProofRepository {
	return &PaymentProofRepository{
		db,
	}
}

// CreatePaymentProof inserts a new payment proof into the database
func (r *PaymentProofRepository) CreatePaymentProof(ctx context.Context, paymentProof *domain.PaymentProof) (*domain.PaymentProof, error) {
	query := r.db.QueryBuilder.Insert("\"PaymentProof\"").
		Columns("id", "\"quoteId\"", "url").
		Values(paymentProof.ID, paymentProof.QuoteID, paymentProof.URL).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(
		&paymentProof.ID,
		&paymentProof.QuoteID,
		&paymentProof.URL,
		&paymentProof.IsReviewed,
	)
	if err != nil {
		return nil, err
	}

	return paymentProof, nil
}

// GetPaymentProofByID selects a payment proof by its ID
func (r *PaymentProofRepository) GetPaymentProofByID(ctx context.Context, id uuid.UUID) (*domain.PaymentProof, error) {
	var paymentProof domain.PaymentProof

	query := r.db.QueryBuilder.Select("*").
		From("\"PaymentProof\"").
		Where(sq.Eq{"id": id}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(
		&paymentProof.ID,
		&paymentProof.QuoteID,
		&paymentProof.URL,
		&paymentProof.IsReviewed,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

	return &paymentProof, nil
}

// GetPaymentProofByQuoteID selecciona un comprobante por su QuoteID
func (r *PaymentProofRepository) GetPaymentProofByQuoteID(ctx context.Context, quoteID uuid.UUID) (*domain.PaymentProof, error) {
	var paymentProof domain.PaymentProof

	query := r.db.QueryBuilder.Select("*").
		From("\"PaymentProof\"").
		Where(sq.Eq{"\"quoteId\"": quoteID}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(
		&paymentProof.ID,
		&paymentProof.QuoteID,
		&paymentProof.URL,
		&paymentProof.IsReviewed,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No es error, simplemente no hay comprobante aún
		}
		return nil, err
	}

	return &paymentProof, nil
}

// GetPaymentProofs selects all payment proofs with optional filtering by QuoteID and IsReviewed
func (r *PaymentProofRepository) GetPaymentProofs(ctx context.Context, filter port.PaymentProofFilter) ([]domain.PaymentProof, error) {
	var paymentProofs []domain.PaymentProof

	query := r.db.QueryBuilder.Select("id", "\"quoteId\"", "url", "\"isReviewed\"").
		From("\"PaymentProof\"")

	// Filtros opcionales
	if filter.QuoteID != nil {
		query = query.Where(sq.Eq{`"quoteId"`: *filter.QuoteID})
	}
	if filter.IsReviewed != nil {
		query = query.Where(sq.Eq{`"isReviewed"`: *filter.IsReviewed})
	}

	// Paginación (skip = número de página - 1)
	if filter.Limit > 0 {
		offset := uint64(0)
		if filter.Skip > 0 {
			offset = (filter.Skip - 1) * filter.Limit
		}
		query = query.Limit(filter.Limit).Offset(offset)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Executing query",
		"sql", sql,
		"args", args,
		"filters", filter)

	rows, err := r.db.Conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p domain.PaymentProof
		err := rows.Scan(
			&p.ID,
			&p.QuoteID,
			&p.URL,
			&p.IsReviewed,
		)
		if err != nil {
			return nil, err
		}
		paymentProofs = append(paymentProofs, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return paymentProofs, nil
}

// UpdatePaymentProof updates only the IsReviewed field of a payment proof
func (r *PaymentProofRepository) UpdatePaymentProof(ctx context.Context, paymentProof *domain.PaymentProof) (*domain.PaymentProof, error) {
	query := r.db.QueryBuilder.Update("\"PaymentProof\"").
		Set("\"isReviewed\"", paymentProof.IsReviewed).
		Where(sq.Eq{"id": paymentProof.ID}).
		Suffix("RETURNING id, \"quoteId\", url, \"isReviewed\"")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(
		&paymentProof.ID,
		&paymentProof.QuoteID,
		&paymentProof.URL,
		&paymentProof.IsReviewed,
	)
	if err != nil {
		return nil, err
	}

	return paymentProof, nil
}

// DeletePaymentProof deletes a payment proof by ID from the database
func (r *PaymentProofRepository) DeletePaymentProof(ctx context.Context, id uuid.UUID) error {
	query := r.db.QueryBuilder.Delete("\"PaymentProof\"").
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

func (r *PaymentProofRepository) WithTx(
	ctx context.Context,
	fn func(repo port.PaymentProofRepository) error,
) error {
	return r.db.WithTx(ctx, func(txDB *postgres.DB) error {
		txRepo := NewPaymentProofRepository(txDB)
		return fn(txRepo)
	})
}

func (r *PaymentProofRepository) MarkAsReviewedByQuoteID(ctx context.Context, quoteID uuid.UUID) error {
	query := r.db.QueryBuilder.
		Update("\"PaymentProof\"").
		Set("\"isReviewed\"", true).
		Where(sq.Eq{"\"quoteId\"": quoteID})

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Conn.Exec(ctx, sql, args...)
	return err
}
