package port

import (
	"context"
	"harajuku/backend/internal/core/domain"

	"github.com/google/uuid"
)

//go:generate mockgen -source=availability_slot.go -destination=mock/availability_slot.go -package=mock

type SlotState string

const (
	SlotStateFree   SlotState = "free"
	SlotStateBooked SlotState = "booked"
)

// AvailabilitySlotFilter contiene los filtros para listar AvailabilitySlots
type AvailabilitySlotFilter struct {
	UserID  *uuid.UUID
	Month   *string
	ByState *SlotState
	Skip    uint64
	Limit   uint64
}

// AvailabilitySlotRepository es la interfaz para interactuar con los datos de AvailabilitySlot
type AvailabilitySlotRepository interface {
	CreateAvailabilitySlot(ctx context.Context, slot *domain.AvailabilitySlot) (*domain.AvailabilitySlot, error)
	GetAvailabilitySlotByID(ctx context.Context, id uuid.UUID) (*domain.AvailabilitySlot, error)
	ListAvailabilitySlots(ctx context.Context, filter AvailabilitySlotFilter) ([]domain.AvailabilitySlot, error)
	UpdateAvailabilitySlot(ctx context.Context, slot *domain.AvailabilitySlot) (*domain.AvailabilitySlot, error)
	DeleteAvailabilitySlot(ctx context.Context, id uuid.UUID) error
}

// AvailabilitySlotService es la interfaz para interactuar con la lógica de negocio de AvailabilitySlot
type AvailabilitySlotService interface {
	CreateAvailabilitySlot(ctx context.Context, slot *domain.AvailabilitySlot) (*domain.AvailabilitySlot, error)
	GetAvailabilitySlot(ctx context.Context, id uuid.UUID) (*domain.AvailabilitySlot, error)
	ListAvailabilitySlots(ctx context.Context, filter AvailabilitySlotFilter) ([]domain.AvailabilitySlot, error)
	UpdateAvailabilitySlot(ctx context.Context, slot *domain.AvailabilitySlot) (*domain.AvailabilitySlot, error)
	DeleteAvailabilitySlot(ctx context.Context, id uuid.UUID) error
}
