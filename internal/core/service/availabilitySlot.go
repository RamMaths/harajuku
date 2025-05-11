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
type AvailabilitySlotService struct {
	repo  port.AvailabilitySlotRepository
	cache port.CacheRepository
}

// NewAvailabilitySlotService crea una nueva instancia del servicio AvailabilitySlot
func NewAvailabilitySlotService(repo port.AvailabilitySlotRepository, cache port.CacheRepository) *AvailabilitySlotService {
	return &AvailabilitySlotService{
		repo,
		cache,
	}
}

func (as *AvailabilitySlotService) CreateAvailabilitySlot(ctx context.Context, slot *domain.AvailabilitySlot) (*domain.AvailabilitySlot, error) {
	slot.ID = uuid.New()
	slot.IsBooked = false

	createdSlot, err := as.repo.CreateAvailabilitySlot(ctx, slot)
	if err != nil {
		slog.Error("AvailabilitySlot creation failed", "error", err)
		return nil, domain.ErrInternal
	}

	// Cache del slot creado
	cacheKey := util.GenerateCacheKey("availabilitySlot", createdSlot.ID)
	slotSerialized, err := util.Serialize(createdSlot)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = as.cache.Set(ctx, cacheKey, slotSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return createdSlot, nil
}

// GetAvailabilitySlot obtiene un availability slot por ID
func (as *AvailabilitySlotService) GetAvailabilitySlot(ctx context.Context, id uuid.UUID) (*domain.AvailabilitySlot, error) {
	var slot *domain.AvailabilitySlot

	// Revisar la caché primero
	cacheKey := util.GenerateCacheKey("availabilitySlot", id)
	cachedSlot, err := as.cache.Get(ctx, cacheKey)
	if err == nil {
		err := util.Deserialize(cachedSlot, &slot)
		if err != nil {
			return nil, domain.ErrInternal
		}
		return slot, nil
	}

	// Obtener del repositorio si no está en caché
	slot, err = as.repo.GetAvailabilitySlotByID(ctx, id)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	// Cache del slot
	slotSerialized, err := util.Serialize(slot)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = as.cache.Set(ctx, cacheKey, slotSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return slot, nil
}

// ListAvailabilitySlots lista todos los availability slots con opciones de filtrado
func (as *AvailabilitySlotService) ListAvailabilitySlots(ctx context.Context, filter port.AvailabilitySlotFilter) ([]domain.AvailabilitySlot, error) {
	var slots []domain.AvailabilitySlot

	params := util.GenerateCacheKeyParams(
		filter.UserID,
		filter.Month,
		filter.ByState,
		filter.Skip,
		filter.Limit,
	)
	cacheKey := util.GenerateCacheKey("availabilitySlots", params)

	// Revisar la caché primero
	cachedSlots, err := as.cache.Get(ctx, cacheKey)
	if err == nil {
		err := util.Deserialize(cachedSlots, &slots)
		if err != nil {
			return nil, domain.ErrInternal
		}
		return slots, nil
	}

	// Obtener del repositorio si no está en caché
	slots, err = as.repo.ListAvailabilitySlots(ctx, filter)
	if err != nil {
		return nil, domain.ErrInternal
	}

	// Cache de los slots
	slotsSerialized, err := util.Serialize(slots)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = as.cache.Set(ctx, cacheKey, slotsSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return slots, nil
}

// UpdateAvailabilitySlot actualiza los datos de un availability slot
func (as *AvailabilitySlotService) UpdateAvailabilitySlot(ctx context.Context, slot *domain.AvailabilitySlot) (*domain.AvailabilitySlot, error) {
	existingSlot, err := as.repo.GetAvailabilitySlotByID(ctx, slot.ID)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	// Revisar si los datos son los mismos (sin actualización)
	if existingSlot.StartTime == slot.StartTime && existingSlot.EndTime == slot.EndTime && existingSlot.IsBooked == slot.IsBooked {
		return nil, domain.ErrNoUpdatedData
	}

	// Actualizar el slot
	_, err = as.repo.UpdateAvailabilitySlot(ctx, slot)
	if err != nil {
		return nil, domain.ErrInternal
	}

	// Cache del slot actualizado
	cacheKey := util.GenerateCacheKey("availabilitySlot", slot.ID)

	err = as.cache.Delete(ctx, cacheKey)
	if err != nil {
		return nil, domain.ErrInternal
	}

	slotSerialized, err := util.Serialize(slot)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = as.cache.Set(ctx, cacheKey, slotSerialized, 0)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return slot, nil
}

// DeleteAvailabilitySlot elimina un availability slot por ID
func (as *AvailabilitySlotService) DeleteAvailabilitySlot(ctx context.Context, id uuid.UUID) error {
	_, err := as.repo.GetAvailabilitySlotByID(ctx, id)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return err
		}
		return domain.ErrInternal
	}

	// Eliminar de la caché
	cacheKey := util.GenerateCacheKey("availabilitySlot", id)

	err = as.cache.Delete(ctx, cacheKey)
	if err != nil {
		return domain.ErrInternal
	}

	// Eliminar del repositorio
	return as.repo.DeleteAvailabilitySlot(ctx, id)
}
