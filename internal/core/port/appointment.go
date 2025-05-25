package port

import (
	"context"
	"harajuku/backend/internal/core/domain"

	"github.com/google/uuid"
	"go.starlark.net/lib/time"
)

// AppointmentFilter contiene los filtros para listar Appointments
type AppointmentFilter struct {
	CustomerID  *uuid.UUID
	QuoteID     *uuid.UUID
	StartDate   *time.Time
	EndDate   	*time.Time
	ByState 		*domain.AppointmentStatus
	Skip    		uint64
	Limit   		uint64
}

// AppointmentRepository es la interfaz para interactuar con los datos de Appointment
type AppointmentRepository interface {
	CreateAppointment(ctx context.Context, slot *domain.Appointment) (*domain.Appointment, error)
	GetAppointmentByID(ctx context.Context, id uuid.UUID) (*domain.Appointment, error)
	ListAppointments(ctx context.Context, filter AppointmentFilter) ([]domain.Appointment, error)
	UpdateAppointment(ctx context.Context, slot *domain.Appointment) (*domain.Appointment, error)
	DeleteAppointment(ctx context.Context, id uuid.UUID) error
}

// AppointmentService es la interfaz para interactuar con la lógica de negocio de Appointment
type AppointmentService interface {
	CreateAppointment(ctx context.Context, slot *domain.Appointment) (*domain.Appointment, error)
	GetAppointment(ctx context.Context, id uuid.UUID) (*domain.Appointment, error)
	ListAppointments(ctx context.Context, filter AppointmentFilter) ([]domain.Appointment, error)
	UpdateAppointment(ctx context.Context, slot *domain.Appointment) (*domain.Appointment, error)
	DeleteAppointment(ctx context.Context, id uuid.UUID) error
}
