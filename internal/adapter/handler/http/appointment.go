package http

import (
	"fmt"
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AppointmentHandler struct {
	svc         port.AppointmentService
}

func NewAppointmentHandler(svc port.AppointmentService, userService port.UserService) *AppointmentHandler {
	return &AppointmentHandler{
		svc:         svc,
	}
}

type createAppointmentRequest struct {
	SlotID 		uuid.UUID `json:"slotId" binding:"required"`
	QuoteID   uuid.UUID `json:"quoteId" binding:"required"`
}

type appointmentResponse struct {
	ID        uuid.UUID 									`json:"id"`
	UserID   	uuid.UUID 									`json:"userId"`
	SlotID 		uuid.UUID    								`json:"slotId"`
	QuoteID   uuid.UUID    								`json:"quoteId"`
	Status  	domain.AppointmentStatus    `json:"status"`
}

func newAppointmentResponse(appointment *domain.Appointment) *appointmentResponse {
	return &appointmentResponse{
		ID:        	appointment.ID,
		UserID:   	appointment.UserID,
		SlotID: 		appointment.SlotID,
		QuoteID:   	appointment.QuoteID,
		Status:  		appointment.Status,
	}
}

func (h *AppointmentHandler) CreateAppointment(ctx *gin.Context) {
	auth := getAuthPayload(ctx, authorizationPayloadKey)

	var req createAppointmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validationError(ctx, err)
		return
	}

	appointment := &domain.Appointment{
		ID:        	uuid.New(),
		UserID: 		auth.UserID,
		SlotID: 		req.SlotID,
		QuoteID:   	req.QuoteID,
		Status:  		domain.Pending,
	}

	created, err := h.svc.CreateAppointment(ctx, appointment)
	if err != nil {
		handleError(ctx, err)
		return
	}

	handleSuccess(ctx, newAppointmentResponse(created))
}

type listAppointmentRequest struct {
	CustomerID  string		`form:"customerId"`
	QuoteID  	 	string		`form:"quoteId"`
	StartDate  	string		`form:"startDate"`
	EndDate  		string		`form:"endDate"`
	ByState 		string		`form:"state"`
	Skip   		 	uint64		`form:"skip" binding:"required,min=0"`
	Limit  		 	uint64		`form:"limit" binding:"required,min=5"`
}

func (h *AppointmentHandler) ListAppointments(ctx *gin.Context) {
	var req listAppointmentRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		validationError(ctx, err)
		return
	}

	// Parse optional UUIDs
	var (
		customerID *uuid.UUID
		quoteID    *uuid.UUID
		err        error
	)

	if req.CustomerID != "" {
		id, err := uuid.Parse(req.CustomerID)
		if err != nil {
			validationError(ctx, fmt.Errorf("invalid customerId: %v", err))
			return
		}
		customerID = &id
	}

	if req.QuoteID != "" {
		id, err := uuid.Parse(req.QuoteID)
		if err != nil {
			validationError(ctx, fmt.Errorf("invalid quoteId: %v", err))
			return
		}
		quoteID = &id
	}

	// Parse optional dates
	var (
		startDate *time.Time
		endDate   *time.Time
	)
	if req.StartDate != "" {
		t, err := time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			validationError(ctx, fmt.Errorf("invalid startDate format (must be RFC3339)"))
			return
		}
		startDate = &t
	}
	if req.EndDate != "" {
		t, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			validationError(ctx, fmt.Errorf("invalid endDate format (must be RFC3339)"))
			return
		}
		endDate = &t
	}

	// Convert state if provided
	var state *domain.AppointmentStatus
	if req.ByState != "" {
		s := domain.AppointmentStatus(req.ByState)
		if s != domain.Booked && s != domain.Cancelled && s != domain.Pending && s != domain.Completed {
			validationError(ctx, fmt.Errorf("invalid state value, must be 'booked', 'pending', 'cancelled' or 'completed'"))
		}
		// validate against your enum if needed...
	}

	// Build the filter
	filter := port.AppointmentFilter{
		CustomerID: customerID,
		QuoteID:    quoteID,
		StartDate:  startDate,
		EndDate:    endDate,
		ByState:    state,
		Skip:       req.Skip,
		Limit:      req.Limit,
	}

	// Obtener los appointments
	appointments, err := h.svc.ListAppointments(ctx, filter)
	if err != nil {
		handleError(ctx, err)
		return
	}

	// Preparar la respuesta
	responses := make([]appointmentResponse, 0, len(appointments))
	for _, s := range appointments {
		responses = append(responses, *newAppointmentResponse(&s))
	}

	meta := newMeta(uint64(len(appointments)), req.Limit, req.Skip)
	handleSuccess(ctx, toMap(meta, responses, "appointments"))
}

func (qh *AppointmentHandler) GetAppointment(ctx *gin.Context) {
	id := ctx.DefaultQuery("id", "")

	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID parameter is required"})
		return
	}

	quoteID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	appointment, err := qh.svc.GetAppointment(ctx, quoteID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	rsp := newAppointmentResponse(appointment)
	handleSuccess(ctx, rsp)
}

type updateAppointmentRequest struct {
	SlotID 		uuid.UUID `json:"slotId" binding:"required"`
}

func (h *AppointmentHandler) UpdateAppointment(ctx *gin.Context) {
	id := ctx.DefaultQuery("id", "")

	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID parameter is required"})
		return
	}

	appointmentId, err := uuid.Parse(id)

	if err != nil {
		handleError(ctx, err)
		return
	}

	var req updateAppointmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validationError(ctx, err)
		return
	}

	updatedAppointment, err := h.svc.UpdateAppointment(ctx, &domain.Appointment{ID: appointmentId, SlotID: req.SlotID})
	if err != nil {
		handleError(ctx, err)
		return
	}

	handleSuccess(ctx, newAppointmentResponse(updatedAppointment))
}

func (h *AppointmentHandler) DeleteAppointment(ctx *gin.Context) {
	id := ctx.DefaultQuery("id", "")

	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID parameter is required"})
		return
	}

	appointmentID, err := uuid.Parse(id)
	if err != nil {
		validationError(ctx, fmt.Errorf("invalid id format"))
		return
	}

	if err := h.svc.DeleteAppointment(ctx, appointmentID); err != nil {
		handleError(ctx, err)
		return
	}

	handleSuccess(ctx, "Availability appointment deleted successfully")
}
