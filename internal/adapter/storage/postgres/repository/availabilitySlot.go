package repository

import (
	"context"
	"fmt"
	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"log"

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

	// Filter by AvailabilitySlot.startTime
	if filter.StartDate != nil {
		query = query.Where(sq.GtOrEq{`"AvailabilitySlot"."startTime"`: *filter.StartDate})
	}

	if filter.EndDate != nil {
		query = query.Where(sq.LtOrEq{`"AvailabilitySlot"."startTime"`: *filter.EndDate})
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
					sq.Eq{`"Appointment"."status"`: "pending"},
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

// MarkSlotsAsBookedByQuoteID actualiza todos los AvailabilitySlot relacionados a una Quote aprobada
func (r *AvailabilitySlotRepository) MarkSlotsAsBookedByQuoteID(ctx context.Context, quoteID uuid.UUID) error {
	log.Printf("Marking slots as booked for quote ID: %s\n", quoteID)

	sql := `
        UPDATE "AvailabilitySlot"
        SET "isBooked" = TRUE
        WHERE "id" IN (
            SELECT "slotId" FROM "Appointment" WHERE "quoteId" = $1
        )
        AND "isBooked" = FALSE
    `

	cmdTag, err := r.db.Conn.Exec(ctx, sql, quoteID)
	if err != nil {
		return fmt.Errorf("failed to mark slots as booked: %w", err)
	}

	rowsAffected := cmdTag.RowsAffected()
	log.Printf("Rows affected: %d\n", rowsAffected)

	// Opcional: Decide si 0 filas afectadas es realmente un error en tu caso de uso
	if rowsAffected == 0 {
		log.Println("No availability slots found to update for quote ID:", quoteID)
		return nil // o mantener el error si es realmente un caso anómalo
	}

	return nil
}
