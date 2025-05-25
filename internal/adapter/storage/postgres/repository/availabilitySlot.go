package repository

import (
	"context"
	"fmt"
	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"log"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AvailabilitySlotRepository struct {
	db *postgres.DB
}

func NewAvailabilitySlotRepository(db *postgres.DB) *AvailabilitySlotRepository {
	return &AvailabilitySlotRepository{
		db,
	}
}

// CreateAvailabilitySlot crea un nuevo availability slot en la base de datos
func (r *AvailabilitySlotRepository) CreateAvailabilitySlot(ctx context.Context, slot *domain.AvailabilitySlot) (*domain.AvailabilitySlot, error) {
	query := r.db.QueryBuilder.Insert("\"AvailabilitySlot\""). // Ajustar el nombre de la tabla si es necesario
									Columns("id", "\"adminId\"", "\"startTime\"", "\"endTime\"", "\"isBooked\"").
									Values(slot.ID, slot.AdminID, slot.StartTime, slot.EndTime, slot.IsBooked).
									Suffix("RETURNING id")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(&slot.ID)
	if err != nil {
		return nil, err
	}

	return slot, nil
}

// GetAvailabilitySlotByID obtiene un availability slot por su ID
func (r *AvailabilitySlotRepository) GetAvailabilitySlotByID(ctx context.Context, id uuid.UUID) (*domain.AvailabilitySlot, error) {
	var slot domain.AvailabilitySlot

	query := r.db.QueryBuilder.Select("id", "\"adminId\"", "\"startTime\"", "\"endTime\"", "\"isBooked\"").
		From("\"AvailabilitySlot\"").
		Where(sq.Eq{"id": id}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(&slot.ID, &slot.AdminID, &slot.StartTime, &slot.EndTime, &slot.IsBooked)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

	return &slot, nil
}

// ListAvailabilitySlots obtiene una lista de availability slots de la base de datos
func (r *AvailabilitySlotRepository) ListAvailabilitySlots(ctx context.Context, filter port.AvailabilitySlotFilter) ([]domain.AvailabilitySlot, error) {
	log.Printf("Iniciando ListAvailabilitySlots con filtro: %+v", filter)
	var slots []domain.AvailabilitySlot

	query := r.db.QueryBuilder.
		Select(
			`"AvailabilitySlot"."id"`,
			`"AvailabilitySlot"."adminId"`,
			`"AvailabilitySlot"."startTime"`,
			`"AvailabilitySlot"."endTime"`,
			`"AvailabilitySlot"."isBooked"`,
		).
		From(`"AvailabilitySlot"`)

	// Paginación (skip = número de página - 1)
	if filter.Limit > 0 {
		offset := ((filter.Skip - 1) * filter.Limit)
		log.Printf("Configurando paginación - Limit: %d, Offset: %d", filter.Limit, offset)
		query = query.Limit(filter.Limit).Offset(offset)
	}

	// Filtro por adminID
	if filter.UserID != nil {
		log.Printf("Aplicando filtro por UserID: %v", *filter.UserID)
		query = query.Where(sq.Eq{`"AvailabilitySlot"."adminId"`: *filter.UserID})
	}

	// Filtro por mes
	if filter.Month != nil {
		if _, err := time.Parse("2006-01", *filter.Month); err != nil {
			return nil, fmt.Errorf("formato de mes debe ser YYYY-MM")
		}

		parts := strings.Split(*filter.Month, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("formato de mes inválido")
		}

		year, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("año inválido")
		}

		month, err := strconv.Atoi(parts[1])
		if err != nil || month < 1 || month > 12 {
			return nil, fmt.Errorf("mes inválido")
		}

		// Corrección clave: Usar And() para combinar ambas condiciones
		query = query.Where(
			sq.And{
				sq.Expr(`EXTRACT(YEAR FROM "AvailabilitySlot"."startTime") = ?`, year),
				sq.Expr(`EXTRACT(MONTH FROM "AvailabilitySlot"."startTime") = ?`, month),
			},
		)
	}

	// Filtro por estado
	if filter.ByState != nil {
		switch *filter.ByState {
		case port.SlotStateFree:
			log.Println("Configurando JOIN para slots libres")
			query = query.
				LeftJoin(`"Appointment" ON "AvailabilitySlot"."id" = "Appointment"."slotId"`).
				Where(sq.Or{
					// Slots sin appointment asociado
					sq.Expr(`"Appointment"."id" IS NULL`),
					// Slots con appointment en estado 'needs_review'
					sq.Eq{`"Appointment"."status"`: "needs_review"},
				}).
				// Asegurar que no estén marcados como booked
				Where(sq.Eq{`"AvailabilitySlot"."isBooked"`: false})

		case port.SlotStateBooked:
			log.Println("Configurando JOIN para slots reservados")
			query = query.
				Join(`"Appointment" ON "AvailabilitySlot"."id" = "Appointment"."slotId"`).
				Where(sq.Or{
					sq.Eq{`"Appointment"."status"`: "booked"},
					sq.Eq{`"Appointment"."status"`: "completed"},
				})
		}
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error al construir la consulta: %w", err)
	}
	log.Printf("SQL generado: %s", sql)
	log.Printf("Parámetros SQL: %v", args)

	rows, err := r.db.Conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar consulta: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var slot domain.AvailabilitySlot
		if err := rows.Scan(
			&slot.ID,
			&slot.AdminID,
			&slot.StartTime,
			&slot.EndTime,
			&slot.IsBooked,
		); err != nil {
			return nil, fmt.Errorf("error al leer datos: %w", err)
		}
		slots = append(slots, slot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error al procesar resultados: %w", err)
	}

	log.Printf("Consulta completada exitosamente. Slots encontrados: %d", len(slots))
	return slots, nil
}

// UpdateAvailabilitySlot actualiza un availability slot existente en la base de datos
func (r *AvailabilitySlotRepository) UpdateAvailabilitySlot(ctx context.Context, slot *domain.AvailabilitySlot) (*domain.AvailabilitySlot, error) {
	query := r.db.QueryBuilder.Update("\"AvailabilitySlot\"").
		Set("\"adminId\"", slot.AdminID).
		Set("\"startTime\"", slot.StartTime).
		Set("\"endTime\"", slot.EndTime).
		Set("\"isBooked\"", slot.IsBooked).
		Where(sq.Eq{"id": slot.ID}).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.Conn.QueryRow(ctx, sql, args...).Scan(&slot.ID, &slot.AdminID, &slot.StartTime, &slot.EndTime, &slot.IsBooked)
	if err != nil {
		return nil, err
	}

	return slot, nil
}

// DeleteAvailabilitySlot elimina un availability slot por ID desde la base de datos
func (r *AvailabilitySlotRepository) DeleteAvailabilitySlot(ctx context.Context, id uuid.UUID) error {
	query := r.db.QueryBuilder.Delete("\"AvailabilitySlot\"").
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
