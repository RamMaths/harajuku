package http

import (
	"fmt"
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TypeOfServiceHandler representa el controlador HTTP para las solicitudes relacionadas con tipos de servicio
type TypeOfServiceHandler struct {
	svc port.TypeOfServiceService
}

// NewTypeOfServiceHandler crea una nueva instancia de TypeOfServiceHandler
func NewTypeOfServiceHandler(svc port.TypeOfServiceService) *TypeOfServiceHandler {
	return &TypeOfServiceHandler{
		svc,
	}
}

// typeOfServiceResponse representa la respuesta
type typeOfServiceResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Price float64   `json:"price"`
}

// newTypeOfServiceResponse convierte un objeto domain.TypeOfService en una respuesta de tipo de servicio
func newTypeOfServiceResponse(s *domain.TypeOfService) *typeOfServiceResponse {
	return &typeOfServiceResponse{
		ID:    s.ID,
		Name:  s.Name,
		Price: s.Price,
	}
}

// createTypeOfServiceRequest representa el cuerpo de la solicitud para crear un tipo de servicio
type createTypeOfServiceRequest struct {
	Name  string  `json:"name" binding:"required"`
	Price float64 `json:"price" binding:"required"`
}

// CreateTypeOfService godoc
//
// @Summary        Register a new type of service
// @Description    Create a new type of service
// @Tags           TypeOfServices
// @Accept         json
// @Produce        json
// @Param          name   body    string  true   "Name"
// @Param          price  body    float64 true  "Price"
// @Success        200    {object}  typeOfServiceResponse  "Type of service created"
// @Failure        400    {object}  errorResponse  "Validation error"
// @Failure        500    {object}  errorResponse  "Internal server error"
// @Router         /type-of-services [post]
func (tsh *TypeOfServiceHandler) CreateTypeOfService(ctx *gin.Context) {
	var req createTypeOfServiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validationError(ctx, err)
		return
	}

	service := &domain.TypeOfService{
		ID:    uuid.New(),
		Name:  req.Name,
		Price: req.Price,
	}

	createdService, err := tsh.svc.CreateTypeOfService(ctx, service)
	if err != nil {
		handleError(ctx, err)
		return
	}

	rsp := newTypeOfServiceResponse(createdService)
	handleSuccess(ctx, rsp)
}

// listTypeOfServicesRequest representa los parámetros de la consulta para listar tipos de servicio
type listTypeOfServicesRequest struct {
	Skip  uint64 `form:"skip" binding:"required,min=0"`
	Limit uint64 `form:"limit" binding:"required,min=5"`
}

// ListTypeOfServices godoc
//
// @Summary        List types of services
// @Description    List types of services with pagination
// @Tags           TypeOfServices
// @Accept         json
// @Produce        json
// @Param          skip   query   uint64 true   "Skip"
// @Param          limit  query   uint64 true   "Limit"
// @Success        200    {object}  meta  "Types of services displayed"
// @Failure        400    {object}  errorResponse  "Validation error"
// @Failure        500    {object}  errorResponse  "Internal server error"
// @Router         /type-of-services [get]
func (tsh *TypeOfServiceHandler) ListTypeOfServices(ctx *gin.Context) {
	var req listTypeOfServicesRequest
	var servicesList []typeOfServiceResponse

	if err := ctx.ShouldBindQuery(&req); err != nil {
		validationError(ctx, err)
		return
	}

	services, err := tsh.svc.ListTypeOfServices(ctx, req.Skip, req.Limit)
	if err != nil {
		handleError(ctx, err)
		return
	}

	for _, service := range services {
		servicesList = append(servicesList, *newTypeOfServiceResponse(&service))
	}

	total := uint64(len(servicesList))
	meta := newMeta(total, req.Limit, req.Skip)
	rsp := toMap(meta, servicesList, "typeOfServices")

	handleSuccess(ctx, rsp)
}

// getTypeOfServiceRequest representa el cuerpo de la solicitud para obtener un tipo de servicio por ID

// GetTypeOfService godoc
//
// @Summary        Get a type of service
// @Description    Get a type of service by id
// @Tags           TypeOfServices
// @Accept         json
// @Produce        json
// @Param          id   query   string true   "Type of Service ID"
// @Success        200  {object}  typeOfServiceResponse  "Type of service displayed"
// @Failure        400  {object}  errorResponse  "Validation error"
// @Failure        404  {object}  errorResponse  "Data not found error"
// @Failure        500  {object}  errorResponse  "Internal server error"
// @Router         /type-of-services [get]
func (tsh *TypeOfServiceHandler) GetTypeOfService(ctx *gin.Context) {
	id := ctx.DefaultQuery("id", "")

	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID parameter is required"})
		return
	}

	serviceID, err := uuid.Parse(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Llamar al servicio para obtener el tipo de servicio
	service, err := tsh.svc.GetTypeOfService(ctx, serviceID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	// Responder con el tipo de servicio
	rsp := newTypeOfServiceResponse(service)
	handleSuccess(ctx, rsp)
}

// updateTypeOfServiceRequest representa el cuerpo de la solicitud para actualizar un tipo de servicio
type updateTypeOfServiceRequest struct {
	Name  string  `json:"name" binding:"required"`
	Price float64 `json:"price" binding:"required"`
}

// UpdateTypeOfService godoc
//
// @Summary        Update a type of service
// @Description    Update a type of service by id
// @Tags           TypeOfServices
// @Accept         json
// @Produce        json
//
// @Param         id     query   string               true   "Type of Service ID"
// @Param         service body    updateTypeOfServiceRequest true   "Service Data"
//
// @Success        200    {object}  typeOfServiceResponse  "Type of service updated"
// @Failure        400    {object}  errorResponse  "Validation error"
// @Failure        404    {object}  errorResponse  "Data not found error"
// @Failure        500    {object}  errorResponse  "Internal server error"
// @Router         /type-of-services [put]
func (tsh *TypeOfServiceHandler) UpdateTypeOfService(ctx *gin.Context) {
	id := ctx.DefaultQuery("id", "")

	if id == "" {
		validationError(ctx, fmt.Errorf("ID is required"))
		return
	}

	var req updateTypeOfServiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validationError(ctx, err)
		return
	}

	service := &domain.TypeOfService{
		ID:    uuid.MustParse(id),
		Name:  req.Name,
		Price: req.Price,
	}

	updatedService, err := tsh.svc.UpdateTypeOfService(ctx, service)
	if err != nil {
		handleError(ctx, err)
		return
	}

	rsp := newTypeOfServiceResponse(updatedService)
	handleSuccess(ctx, rsp)
}

// DeleteTypeOfService godoc
//
// @Summary        Delete a type of service
// @Description    Delete a type of service by id
// @Tags           TypeOfServices
// @Accept         json
// @Produce        json
//
// @Param         id   query   string true   "Type of Service ID"
//
// @Success        200    {object}  string  "Type of service deleted successfully"
// @Failure        400    {object}  errorResponse  "Validation error"
// @Failure        404    {object}  errorResponse  "Data not found error"
// @Failure        500    {object}  errorResponse  "Internal server error"
// @Router         /type-of-services [delete]
func (tsh *TypeOfServiceHandler) DeleteTypeOfService(ctx *gin.Context) {
	idStr := ctx.DefaultQuery("id", "")
	if idStr == "" {
		validationError(ctx, fmt.Errorf("ID is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		validationError(ctx, fmt.Errorf("invalid UUID format"))
		return
	}

	err = tsh.svc.DeleteTypeOfService(ctx, id)
	if err != nil {
		handleError(ctx, err)
		return
	}

	handleSuccess(ctx, "Type of service deleted successfully")
}
