package http

import (
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// QuoteHandler representa el controlador HTTP para las solicitudes relacionadas con cotizaciones
type QuoteHandler struct {
	svc port.QuoteService
}

// NewQuoteHandler crea una nueva instancia de QuoteHandler
func NewQuoteHandler(svc port.QuoteService) *QuoteHandler {
	return &QuoteHandler{
		svc,
	}
}

// quoteResponse representa la respuesta
type quoteResponse struct {
	ID              uuid.UUID `json:"id"`
	TypeOfServiceID uuid.UUID `json:"typeOfServiceID"`
	ClientID        uuid.UUID `json:"clientID"`
	Description     string    `json:"description"`
	State           string    `json:"state"`
	Price           float64   `json:"price"`
	TestRequired    bool      `json:"testRequired"`
	Time            string    `json:"time"`
}

// newQuoteResponse convierte un objeto domain. Quote en una respuesta de cotización
func newQuoteResponse(q *domain.Quote) *quoteResponse {
	return &quoteResponse{
		ID:              q.ID,
		TypeOfServiceID: q.TypeOfServiceID,
		ClientID:        q.ClientID,
		Description:     q.Description,
		State:           string(q.State),
		Price:           q.Price,
		TestRequired:    q.TestRequired,
		Time:            q.Time.Format(time.RFC3339),
	}
}

// createQuoteRequest representa el cuerpo de la solicitud para crear una cotización
type createQuoteRequest struct {
	TypeOfServiceID uuid.UUID `json:"typeOfServiceID" binding:"required"`
	ClientID        uuid.UUID `json:"clientID" binding:"required"`
	Description     string    `json:"description" binding:"required"`
	State           string    `json:"state" binding:"required,oneof=pending approved rejected requires_proof"`
	Price           float64   `json:"price" binding:"required"`
	TestRequired    bool      `json:"testRequired" binding:"required"`
}

// CreateQuote godoc
//
//	@Summary		Register a new quote
//	@Description	create a new quote for a client
//	@Tags			Quotes
//	@Accept			json
//	@Produce		json
//	@Param			createQuoteRequest	body		createQuoteRequest	true	"Create quote request"
//	@Success		200				{object}	quoteResponse	"Quote created"
//	@Failure		400				{object}	errorResponse	"Validation error"
//	@Failure		500				{object}	errorResponse	"Internal server error"
//	@Router			/quotes [post]
func (qh *QuoteHandler) CreateQuote(ctx *gin.Context) {
	var req createQuoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validationError(ctx, err)
		return
	}

	quote := domain.Quote{
		ID:              uuid.New(),
		TypeOfServiceID: req.TypeOfServiceID,
		ClientID:        req.ClientID,
		Time:            time.Now(),
		Description:     req.Description,
		State:           domain.QuoteState(req.State),
		Price:           req.Price,
		TestRequired:    req.TestRequired,
	}

	createdQuote, err := qh.svc.CreateQuote(ctx, &quote)
	if err != nil {
		handleError(ctx, err)
		return
	}

	rsp := newQuoteResponse(createdQuote)
	handleSuccess(ctx, rsp)
}

// listQuotesRequest representa los parámetros de la consulta para listar cotizaciones
type listQuotesRequest struct {
	Skip  uint64 `form:"skip" binding:"required,min=0"`
	Limit uint64 `form:"limit" binding:"required,min=5"`
}

// ListQuotes godoc
//
//	@Summary		List quotes
//	@Description	List quotes with pagination
//	@Tags			Quotes
//	@Accept			json
//	@Produce		json
//	@Param			skip	query		uint64			true	"Skip"
//	@Param			limit	query		uint64			true	"Limit"
//	@Success		200		{object}	meta			"Quotes displayed"
//	@Failure		400		{object}	errorResponse	"Validation error"
//	@Failure		500		{object}	errorResponse	"Internal server error"
//	@Router			/quotes [get]
func (qh *QuoteHandler) ListQuotes(ctx *gin.Context) {
	var req listQuotesRequest
	var quotesList []quoteResponse

	if err := ctx.ShouldBindQuery(&req); err != nil {
		validationError(ctx, err)
		return
	}

	quotes, err := qh.svc.ListQuotes(ctx, req.Skip, req.Limit)
	if err != nil {
		handleError(ctx, err)
		return
	}

	for _, quote := range quotes {
		quotesList = append(quotesList, *newQuoteResponse(&quote))
	}

	total := uint64(len(quotesList))
	meta := newMeta(total, req.Limit, req.Skip)
	rsp := toMap(meta, quotesList, "quotes")

	handleSuccess(ctx, rsp)
}

// getQuoteRequest representa el cuerpo de la solicitud para obtener una cotización por ID
type getQuoteRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,min=1"`
}

// GetQuote godoc
//
//	@Summary		Get a quote
//	@Description	Get a quote by id
//	@Tags			Quotes
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64			true	"Quote ID"
//	@Success		200	{object}	quoteResponse	"Quote displayed"
//	@Failure		400	{object}	errorResponse	"Validation error"
//	@Failure		404	{object}	errorResponse	"Data not found error"
//	@Failure		500	{object}	errorResponse	"Internal server error"
//	@Router			/quotes/{id} [get]
func (qh *QuoteHandler) GetQuote(ctx *gin.Context) {
	var req getQuoteRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		validationError(ctx, err)
		return
	}

	quote, err := qh.svc.GetQuote(ctx, req.ID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	rsp := newQuoteResponse(quote)

	handleSuccess(ctx, rsp)
}

// updateQuoteRequest representa el cuerpo de la solicitud para actualizar una cotización
type updateQuoteRequest struct {
	Description  string  `json:"description" binding:"omitempty,required"`
	State        string  `json:"state" binding:"omitempty,required,oneof=pending approved rejected requires_proof"`
	Price        float64 `json:"price" binding:"omitempty,required"`
	TestRequired bool    `json:"testRequired" binding:"omitempty,required"`
}

// UpdateQuote godoc
//
//	@Summary		Update a quote
//	@Description	Update a quote by id
//	@Tags			Quotes
//	@Accept			json
//	@Produce		json
//	@Param			id					path		uint64				true	"Quote ID"
//	@Param			updateQuoteRequest	body		updateQuoteRequest	true	"Update quote request"
//	@Success		200					{object}	quoteResponse		"Quote updated"
//	@Failure		400					{object}	errorResponse		"Validation error"
//	@Failure		500					{object}	errorResponse		"Internal server error"
//	@Router			/quotes/{id} [put]
func (qh *QuoteHandler) UpdateQuote(ctx *gin.Context) {
	var req updateQuoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validationError(ctx, err)
		return
	}

	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		validationError(ctx, err)
		return
	}

	quote := domain.Quote{
		ID:           id,
		Description:  req.Description,
		State:        domain.QuoteState(req.State),
		Price:        req.Price,
		TestRequired: req.TestRequired,
	}

	updatedQuote, err := qh.svc.UpdateQuote(ctx, &quote)
	if err != nil {
		handleError(ctx, err)
		return
	}

	rsp := newQuoteResponse(updatedQuote)

	handleSuccess(ctx, rsp)
}

// deleteQuoteRequest representa el cuerpo de la solicitud para eliminar una cotización
type deleteQuoteRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,min=1"`
}

// DeleteQuote godoc
//
//	@Summary		Delete a quote
//	@Description	Delete a quote by id
//	@Tags			Quotes
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64			true	"Quote ID"
//	@Success		200	{object}	response		"Quote deleted"
//	@Failure		400	{object}	errorResponse	"Validation error"
//	@Failure		404	{object}	errorResponse	"Data not found error"
//	@Failure		500	{object}	errorResponse	"Internal server error"
//	@Router			/quotes/{id} [delete]
func (qh *QuoteHandler) DeleteQuote(ctx *gin.Context) {
	var req deleteQuoteRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		validationError(ctx, err)
		return
	}

	err := qh.svc.DeleteQuote(ctx, req.ID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	handleSuccess(ctx, nil)
}
