package service

import (
	"context"
	"log/slog"

	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"harajuku/backend/internal/core/util"

	"github.com/google/uuid"
)

// AvailabilitySlotService implementa la interfaz port.AvailabilitySlotService
// y proporciona acceso al repositorio de availability slot y al servicio de caché
type AppointmentService struct {
	repo  port.AppointmentRepository
	cache port.CacheRepository
}

// NewAppointmentService crea una nueva instancia del servicio Appointment
func NewAppointmentService(repo port.AppointmentRepository, cache port.CacheRepository) *AppointmentService {
	return &AppointmentService{
		repo,
		cache,
	}
}

func (as *AppointmentService) CreateAppointment(ctx context.Context, appointment *domain.Appointment) (*domain.Appointment, error) {
	appointment.ID = uuid.New()
	appointment.Status = domain.Pending

	createdappointment, err := as.repo.CreateAppointment(ctx, appointment)
	if err != nil {
		slog.Error("Appointment creation failed", "error", err)
		return nil, domain.ErrInternal
	}

	// Cache del appointment creado
	cacheKey := util.GenerateCacheKey("appointment", createdappointment.ID)
	appointmentSerialized, err := util.Serialize(createdappointment)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = as.cache.Set(ctx, cacheKey, appointmentSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = as.cache.DeleteByPrefix(ctx, "appointments:*")
	if err != nil {
		return nil, domain.ErrInternal
	}

	return createdappointment, nil
}

// GetAppointment obtiene un availability appointment por ID
func (as *AppointmentService) GetAppointment(ctx context.Context, id uuid.UUID) (*domain.Appointment, error) {
	var appointment *domain.Appointment

	// Revisar la caché primero
	cacheKey := util.GenerateCacheKey("appointment", id)
	cachedappointment, err := as.cache.Get(ctx, cacheKey)
	if err == nil {
		err := util.Deserialize(cachedappointment, &appointment)
		if err != nil {
			return nil, domain.ErrInternal
		}
		return appointment, nil
	}

	// Obtener del repositorio si no está en caché
	appointment, err = as.repo.GetAppointmentByID(ctx, id)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	// Cache del appointment
	appointmentSerialized, err := util.Serialize(appointment)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = as.cache.Set(ctx, cacheKey, appointmentSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return appointment, nil
}

// ListAppointments lista todos los availability appointments con opciones de filtrado
func (as *AppointmentService) ListAppointments(ctx context.Context, filter port.AppointmentFilter) ([]domain.Appointment, error) {
	var appointments []domain.Appointment

	params := util.GenerateCacheKeyParams(
		filter.CustomerID,
		filter.StartDate,
		filter.EndDate,
		filter.ByState,
		filter.Skip,
		filter.Limit,
	)
	cacheKey := util.GenerateCacheKey("appointments", params)

	// Revisar la caché primero
	cachedappointments, err := as.cache.Get(ctx, cacheKey)
	if err == nil {
		err := util.Deserialize(cachedappointments, &appointments)
		if err != nil {
			return nil, domain.ErrInternal
		}
		return appointments, nil
	}

	// Obtener del repositorio si no está en caché
	appointments, err = as.repo.ListAppointments(ctx, filter)
	if err != nil {
		return nil, domain.ErrInternal
	}

	// Cache de los appointments
	appointmentsSerialized, err := util.Serialize(appointments)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = as.cache.Set(ctx, cacheKey, appointmentsSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return appointments, nil
}

// UpdateAppointment actualiza los datos de un availability appointment
func (as *AppointmentService) UpdateAppointment(ctx context.Context, appointment *domain.Appointment) (*domain.Appointment, error) {
	existingAppointment, err := as.repo.GetAppointmentByID(ctx, appointment.ID)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	zeroUUID := uuid.UUID{}

	emptyData := appointment.UserID == zeroUUID &&
		appointment.SlotId == zeroUUID &&
		appointment.QuoteId == zeroUUID &&
		appointment.Status == "" 

	sameData := existingAppointment.UserID == appointment.UserID &&
		existingAppointment.SlotId == appointment.SlotId &&
		existingAppointment.QuoteId == appointment.QuoteId &&
		existingAppointment.Status == appointment.Status

	if emptyData || sameData {
		return nil, domain.ErrNoUpdatedData
	}

	// Actualizar el appointment
	_, err = as.repo.UpdateAppointment(ctx, appointment)
	if err != nil {
		return nil, domain.ErrInternal
	}

	// Cache del appointment actualizado
	cacheKey := util.GenerateCacheKey("appointment", appointment.ID)

	err = as.cache.Delete(ctx, cacheKey)
	if err != nil {
		return nil, domain.ErrInternal
	}

	appointmentSerialized, err := util.Serialize(appointment)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = as.cache.Set(ctx, cacheKey, appointmentSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = as.cache.DeleteByPrefix(ctx, "appointments:*")
	if err != nil {
		return nil, domain.ErrInternal
	}

	return appointment, nil
}

// DeleteAppointment elimina un availability appointment por ID
func (as *AppointmentService) DeleteAppointment(ctx context.Context, id uuid.UUID) error {
	_, err := as.repo.GetAppointmentByID(ctx, id)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return err
		}
		return domain.ErrInternal
	}

	// Eliminar de la caché
	cacheKey := util.GenerateCacheKey("appointment", id)

	err = as.cache.Delete(ctx, cacheKey)
	if err != nil {
		return domain.ErrInternal
	}

	err = as.cache.DeleteByPrefix(ctx, "appointments:*")
	if err != nil {
		return domain.ErrInternal
	}

	// Eliminar del repositorio
	return as.repo.DeleteAppointment(ctx, id)
}
