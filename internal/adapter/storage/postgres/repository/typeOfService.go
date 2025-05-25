package repository

import (
	"context"
	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/core/domain"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TypeOfServiceRepository struct {
	db *postgres.DB
}

func NewTypeOfServiceRepository(db *postgres.DB) *TypeOfServiceRepository {
	return &TypeOfServiceRepository{
		db,
	}
}

// CreateTypeOfService inserts a new type of service into the database
func (r *TypeOfServiceRepository) CreateTypeOfService(ctx context.Context, service *domain.TypeOfService) (*domain.TypeOfService, error) {
	query := r.db.QueryBuilder.Insert("\"TypeOfService\"").
		Columns("id", "name", "price").
		Values(service.ID, service.Name, service.Price).
		Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(&service.ID)
	if err != nil {
		return nil, err
	}

	return service, nil
}

// GetTypeOfServiceByID retrieves a type of service by ID
func (r *TypeOfServiceRepository) GetTypeOfServiceByID(ctx context.Context, id uuid.UUID) (*domain.TypeOfService, error) {
	var s domain.TypeOfService

	query := r.db.QueryBuilder.Select("id", "name", "price").
		From("\"TypeOfService\"").
		Where(sq.Eq{"id": id}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(&s.ID, &s.Name, &s.Price)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

	return &s, nil
}

// ListTypeOfServices retrieves a list of types of service
func (r *TypeOfServiceRepository) ListTypeOfServices(ctx context.Context, skip, limit uint64) ([]domain.TypeOfService, error) {
	var services []domain.TypeOfService

	query := r.db.QueryBuilder.Select("id", "name", "price").
		From("\"TypeOfService\"").
		Limit(limit).
		Offset((skip - 1) * limit)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Executing query", "sql", sql, "args", args)

	rows, err := r.db.Conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s domain.TypeOfService
		if err := rows.Scan(&s.ID, &s.Name, &s.Price); err != nil {
			return nil, err
		}
		services = append(services, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

// UpdateTypeOfService updates an existing type of service
func (r *TypeOfServiceRepository) UpdateTypeOfService(ctx context.Context, service *domain.TypeOfService) (*domain.TypeOfService, error) {
	query := r.db.QueryBuilder.Update("\"TypeOfService\"").
		Set("name", service.Name).
		Set("price", service.Price).
		Where(sq.Eq{"id": service.ID}).
		Suffix("RETURNING id, name, price")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(&service.ID, &service.Name, &service.Price)
	if err != nil {
		return nil, err
	}

	return service, nil
}

// DeleteTypeOfService deletes a type of service by ID
func (r *TypeOfServiceRepository) DeleteTypeOfService(ctx context.Context, id uuid.UUID) error {
	query := r.db.QueryBuilder.Delete("\"TypeOfService\"").
		Where(sq.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Conn.Exec(ctx, sql, args...)
	return err
}
