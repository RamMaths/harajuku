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

type AvailabilitySlotHandler struct {
	svc         port.AvailabilitySlotService
	userService port.UserService
}

func NewAvailabilitySlotHandler(svc port.AvailabilitySlotService, userService port.UserService) *AvailabilitySlotHandler {
	return &AvailabilitySlotHandler{
		svc:         svc,
		userService: userService,
	}
}

type createAvailabilitySlotRequest struct {
	StartTime string `json:"startTime" binding:"required"`
	EndTime   string `json:"endTime" binding:"required"`
}

type availabilitySlotResponse struct {
	ID        uuid.UUID `json:"id"`
	AdminID   uuid.UUID `json:"adminId"`
	StartTime string    `json:"startTime"`
	EndTime   string    `json:"endTime"`
	IsBooked  bool      `json:"isBooked"`
}

func newAvailabilitySlotResponse(slot *domain.AvailabilitySlot) *availabilitySlotResponse {
	return &availabilitySlotResponse{
		ID:        slot.ID,
		AdminID:   slot.AdminID,
		StartTime: slot.StartTime.Format(time.RFC3339),
		EndTime:   slot.EndTime.Format(time.RFC3339),
		IsBooked:  slot.IsBooked,
	}
}

func (h *AvailabilitySlotHandler) CreateSlot(ctx *gin.Context) {
	auth := getAuthPayload(ctx, authorizationPayloadKey)

	var req createAvailabilitySlotRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validationError(ctx, err)
		return
	}

	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		validationError(ctx, fmt.Errorf("invalid startTime format"))
		return
	}

	end, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		validationError(ctx, fmt.Errorf("invalid endTime format"))
		return
	}

	if !end.After(start) {
		validationError(ctx, fmt.Errorf("endTime must be after startTime"))
		return
	}

	slot := &domain.AvailabilitySlot{
		ID:        uuid.New(),
		AdminID:   auth.UserID,
		StartTime: start,
		EndTime:   end,
		IsBooked:  false,
	}

	created, err := h.svc.CreateAvailabilitySlot(ctx, slot)
	if err != nil {
		handleError(ctx, err)
		return
	}

	handleSuccess(ctx, newAvailabilitySlotResponse(created))
}

type listAvailabilitySlotRequest struct {
	StartDate  string `form:"start_date"`  // No es required
	EndDate  	 string `form:"end_date"`  // No es required
	State  		 string `form:"state"`  // No es required (antes era IsBooked *bool)
	Skip   		 uint64 `form:"skip" binding:"required,min=0"`
	Limit  		 uint64 `form:"limit" binding:"required,min=5"`
}

func (h *AvailabilitySlotHandler) ListSlots(ctx *gin.Context) {
	var req listAvailabilitySlotRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		validationError(ctx, err)
		return
	}

	var (
		startDate *time.Time
		endDate   *time.Time
		err       error
	)

	// Parse StartDate
	if req.StartDate != "" {
		t, err := time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			validationError(ctx, fmt.Errorf("invalid start_date format (must be RFC3339)"))
			return
		}
		startDate = &t
	}

	// Parse EndDate
	if req.EndDate != "" {
		t, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			validationError(ctx, fmt.Errorf("invalid end_date format (must be RFC3339)"))
			return
		}
		endDate = &t
	}

	// Convertir el estado string a SlotState
	var state *port.SlotState
	if req.State != "" {
		s := port.SlotState(req.State)
		if s != port.SlotStateFree && s != port.SlotStateBooked {
			validationError(ctx, fmt.Errorf("invalid state value, must be 'free' or 'booked'"))
			return
		}
		state = &s
	}

	// Construir el filtro con *time.Time
	filter := port.AvailabilitySlotFilter{
		StartDate: startDate,
		EndDate:   endDate,
		ByState:   state,
		Skip:      req.Skip,
		Limit:     req.Limit,
	}

	// Obtener los slots
	slots, err := h.svc.ListAvailabilitySlots(ctx, filter)
	if err != nil {
		handleError(ctx, err)
		return
	}

	// Preparar la respuesta
	responses := make([]availabilitySlotResponse, 0, len(slots))
	for _, s := range slots {
		responses = append(responses, *newAvailabilitySlotResponse(&s))
	}

	meta := newMeta(uint64(len(slots)), req.Limit, req.Skip)
	handleSuccess(ctx, toMap(meta, responses, "slots"))
}

// Función auxiliar para convertir string vacío a nil
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type updateAvailabilitySlotRequest struct {
	StartTime string `json:"startTime" binding:"required"`
	EndTime   string `json:"endTime" binding:"required"`
}

func (h *AvailabilitySlotHandler) UpdateSlot(ctx *gin.Context) {
	id := ctx.DefaultQuery("id", "")

	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID parameter is required"})
		return
	}

	slotId, err := uuid.Parse(id)

	if err != nil {
		handleError(ctx, err)
		return
	}

	var req updateAvailabilitySlotRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validationError(ctx, err)
		return
	}

	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		validationError(ctx, fmt.Errorf("invalid startTime format"))
		return
	}

	end, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		validationError(ctx, fmt.Errorf("invalid endTime format"))
		return
	}

	if !end.After(start) {
		validationError(ctx, fmt.Errorf("endTime must be after startTime"))
		return
	}

	slot, err := h.svc.GetAvailabilitySlot(ctx, slotId)
	if err != nil {
		handleError(ctx, err)
		return
	}

	slot.StartTime = start
	slot.EndTime = end

	updatedSlot, err := h.svc.UpdateAvailabilitySlot(ctx, slot)
	if err != nil {
		handleError(ctx, err)
		return
	}

	handleSuccess(ctx, newAvailabilitySlotResponse(updatedSlot))
}

func (h *AvailabilitySlotHandler) DeleteSlot(ctx *gin.Context) {
	id := ctx.DefaultQuery("id", "")

	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID parameter is required"})
		return
	}

	slotID, err := uuid.Parse(id)
	if err != nil {
		validationError(ctx, fmt.Errorf("invalid id format"))
		return
	}

	slot, err := h.svc.GetAvailabilitySlot(ctx, slotID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	if slot.IsBooked {
		handleError(ctx, fmt.Errorf("cannot delete a booked slot"))
		return
	}

	if err := h.svc.DeleteAvailabilitySlot(ctx, slotID); err != nil {
		handleError(ctx, err)
		return
	}

	handleSuccess(ctx, "Availability slot deleted successfully")
}
