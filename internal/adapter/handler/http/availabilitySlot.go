package http

import (
	"fmt"
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
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
	if auth == nil || auth.Role != "admin" {
		handleError(ctx, domain.ErrUnauthorized)
		return
	}

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
	UserID string `form:"userId"` // No es required
	Month  string `form:"month"`  // No es required
	State  string `form:"state"`  // No es required (antes era IsBooked *bool)
	Skip   uint64 `form:"skip" binding:"required,min=0"`
	Limit  uint64 `form:"limit" binding:"required,min=5"`
}

func (h *AvailabilitySlotHandler) ListSlots(ctx *gin.Context) {
	var req listAvailabilitySlotRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		validationError(ctx, err)
		return
	}

	auth := getAuthPayload(ctx, authorizationPayloadKey)
	if auth == nil {
		handleError(ctx, domain.ErrUnauthorized)
		return
	}

	// Parsear UserID si viene en la solicitud
	var userID *uuid.UUID
	if req.UserID != "" {
		uid, err := uuid.Parse(req.UserID)
		if err != nil {
			validationError(ctx, fmt.Errorf("invalid userId format"))
			return
		}

		// Verificar permisos si el UserID no coincide con el usuario autenticado
		if uid != auth.UserID {
			user, err := h.userService.GetUser(ctx, auth.UserID)
			if err != nil {
				handleError(ctx, err)
				return
			}
			if user.Role != "admin" {
				handleError(ctx, domain.ErrUnauthorized)
				return
			}
		}
		userID = &uid
	}
	// Validar formato del mes si viene
	if req.Month != "" {
		if _, err := time.Parse("2006-01", req.Month); err != nil {
			validationError(ctx, fmt.Errorf("invalid month format, expected YYYY-MM"))
			return
		}
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

	// Construir el filtro
	filter := port.AvailabilitySlotFilter{
		UserID:  userID,
		Month:   nilIfEmpty(req.Month),
		ByState: state,
		Skip:    req.Skip,
		Limit:   req.Limit,
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
	auth := getAuthPayload(ctx, authorizationPayloadKey)
	if auth == nil || auth.Role != "admin" {
		handleError(ctx, domain.ErrUnauthorized)
		return
	}

	idStr := ctx.Param("id")
	if idStr == "" {
		validationError(ctx, fmt.Errorf("id is required"))
		return
	}

	slotID, err := uuid.Parse(idStr)
	if err != nil {
		validationError(ctx, fmt.Errorf("invalid id format"))
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

	slot, err := h.svc.GetAvailabilitySlot(ctx, slotID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	if slot.AdminID != auth.UserID {
		handleError(ctx, domain.ErrUnauthorized)
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
	auth := getAuthPayload(ctx, authorizationPayloadKey)
	if auth == nil || auth.Role != "admin" {
		handleError(ctx, domain.ErrUnauthorized)
		return
	}

	idStr := ctx.DefaultQuery("id", "")
	if idStr == "" {
		validationError(ctx, fmt.Errorf("id is required"))
		return
	}

	slotID, err := uuid.Parse(idStr)
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

	if slot.AdminID != auth.UserID {
		handleError(ctx, domain.ErrUnauthorized)
		return
	}

	if err := h.svc.DeleteAvailabilitySlot(ctx, slotID); err != nil {
		handleError(ctx, err)
		return
	}

	handleSuccess(ctx, "Availability slot deleted successfully")
}
