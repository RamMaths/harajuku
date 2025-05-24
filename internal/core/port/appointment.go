package port

import (
	"context"
	"harajuku/backend/internal/core/domain"

	"github.com/google/uuid"
)

// AppointmentFilter contiene los filtros para listar Appointments
type AppointmentFilter struct {
	CustomerID  *uuid.UUID
	StartDate   *string
	EndDate   	*string
	ByState 		*SlotState
	Skip    		uint64
	Limit   		uint64
}

// ApointmentRepository es la interfaz para interactuar con los datos de Appointment
type ApointmentRepository interface {
	CreateApointment(ctx context.Context, slot *domain.Appointment) (*domain.Appointment, error)
	GetApointmentByID(ctx context.Context, id uuid.UUID) (*domain.Appointment, error)
	ListApointments(ctx context.Context, filter AppointmentFilter) ([]domain.Appointment, error)
	UpdateApointment(ctx context.Context, slot *domain.Appointment) (*domain.Appointment, error)
	DeleteApointment(ctx context.Context, id uuid.UUID) error
}

// ApointmentService es la interfaz para interactuar con la lógica de negocio de Appointment
type ApointmentService interface {
	CreateApointment(ctx context.Context, slot *domain.Appointment) (*domain.Appointment, error)
	GetApointment(ctx context.Context, id uuid.UUID) (*domain.Appointment, error)
	ListApointments(ctx context.Context, filter AppointmentFilter) ([]domain.Appointment, error)
	UpdateApointment(ctx context.Context, slot *domain.Appointment) (*domain.Appointment, error)
	DeleteApointment(ctx context.Context, id uuid.UUID) error
}
