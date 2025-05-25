package repository

import (
	"context"
	"fmt"
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"harajuku/backend/internal/adapter/storage/postgres"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AppointmentRepository struct {
	db *postgres.DB
}

func NewAppointmentRepository(db *postgres.DB) *AppointmentRepository {
	return &AppointmentRepository{
		db,
	}
}

// CreateAppointment crea un nuevo availability appointment en la base de datos
func (r *AppointmentRepository) CreateAppointment(ctx context.Context, appointment *domain.Appointment) (*domain.Appointment, error) {
	query := r.db.QueryBuilder.Insert("\"Appointment\""). // Ajustar el nombre de la tabla si es necesario
									Columns("id", "\"clientId\"", "\"slotId\"", "\"quoteId\"", "\"status\"").
									Values(appointment.ID, appointment.UserID, appointment.SlotId, appointment.QuoteId, appointment.Status).
									Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil { return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(&appointment.ID)
	if err != nil {
		return nil, err
	}

	return appointment, nil
}

// GetAppointmentByID obtiene un availability appointment por su ID
func (r *AppointmentRepository) GetAppointmentByID(ctx context.Context, id uuid.UUID) (*domain.Appointment, error) {
	var appointment domain.Appointment

	query := r.db.QueryBuilder.Select("id", "\"clientId\"", "\"slotId\"", "\"quoteId\"", "\"status\"").
		From("\"Appointment\"").
		Where(sq.Eq{"id": id}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(&appointment.ID, &appointment.UserID, &appointment.SlotId, &appointment.QuoteId, &appointment.Status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

	return &appointment, nil
}

// ListAppointments obtiene una lista de availability appointments de la base de datos
func (r *AppointmentRepository) ListAppointments(ctx context.Context, filter port.AppointmentFilter) ([]domain.Appointment, error) {
	log.Printf("Iniciando ListAppointments con filtro: %+v", filter)
	var appointments []domain.Appointment

	query := r.db.QueryBuilder.
		Select(
			`"Appointment"."id"`,
			`"Appointment"."clientId"`,
			`"Appointment"."slotId"`,
			`"Appointment"."quoteId"`,
			`"Appointment"."status"`,
		).
		From(`"Appointment"`).
		Join(`"AvailabilitySlot" ON "Appointment"."slotId" = "AvailabilitySlot"."id"`)

	// Filter by Customer ID (Appointment.clientId)
	if filter.CustomerID != nil {
		query = query.Where(sq.Eq{"Appointment.clientId": *filter.CustomerID})
	}

	// Filter by Appointment status
	if filter.ByState != nil {
		query = query.Where(sq.Eq{"Appointment.status": *filter.ByState})
	}

	// Filter by AvailabilitySlot.startTime
	if filter.StartDate != nil {
		query = query.Where(sq.GtOrEq{"AvailabilitySlot.startTime": *filter.StartDate})
	}
	if filter.EndDate != nil {
		query = query.Where(sq.LtOrEq{"AvailabilitySlot.startTime": *filter.EndDate})
	}

	// Pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Skip > 0 {
		query = query.Offset(filter.Skip)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	rows, err := r.db.Conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar consulta: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var appointment domain.Appointment
		if err := rows.Scan(
			&appointment.ID,
			&appointment.UserID,
			&appointment.SlotId,
			&appointment.QuoteId,
			&appointment.Status,
		); err != nil {
			return nil, fmt.Errorf("Error while reading data: %w", err)
		}
		appointments = append(appointments, appointment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error while processing results: %w", err)
	}

	return appointments, nil
}

// UpdateAppointment actualiza un availability appointment existente en la base de datos
func (r *AppointmentRepository) UpdateAppointment(ctx context.Context, appointment *domain.Appointment) (*domain.Appointment, error) {
	query := r.db.QueryBuilder.Update("\"Appointment\"").
		Set("\"clientId\"", appointment.UserID).
		Set("\"slotId\"", appointment.SlotId).
		Set("\"quoteId\"", appointment.QuoteId).
		Set("\"status\"", appointment.Status).
		Where(sq.Eq{"id": appointment.ID}).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(&appointment.ID, &appointment.UserID, &appointment.SlotId, &appointment.QuoteId, &appointment.Status)
	if err != nil {
		return nil, err
	}

	return appointment, nil
}

// DeleteAppointment elimina un availability appointment por ID desde la base de datos
func (r *AppointmentRepository) DeleteAppointment(ctx context.Context, id uuid.UUID) error {
	query := r.db.QueryBuilder.Delete("\"Appointment\"").
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
