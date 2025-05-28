package http

import (
	"fmt"
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"io"
	"net/http"
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
		Time:            q.Time.Format(time.RFC3339),
	}
}

// quoteResponse representa la respuesta
type quoteImageResponse struct {
	ID              uuid.UUID `json:"id"`
	QuoteID 				uuid.UUID `json:"quoteId"`
	URL     				string    `json:"url"`
}


// quoteResponse representa la respuesta
type quoteResponseWithImages struct {
	ID              uuid.UUID 							`json:"id"`
	TypeOfServiceID uuid.UUID 							`json:"typeOfServiceID"`
	ClientID        uuid.UUID 							`json:"clientID"`
	Description     string    							`json:"description"`
	State           string    							`json:"state"`
	Price           float64   							`json:"price"`
	Time            string    							`json:"time"`
	Images          []quoteImageResponse    `json:"images"`
}

// newQuoteResponse convierte un objeto domain. Quote en una respuesta de cotización
func newQuoteResponseWithImages(q *domain.Quote, images []domain.QuoteImage) *quoteResponseWithImages {
	// build the slice of image responses
	respImgs := make([]quoteImageResponse, len(images))
	for i, img := range images {
			respImgs[i] = quoteImageResponse{
					ID:      img.ID,
					QuoteID: img.QuoteID,
					URL:     img.URL,
			}
	}

	return &quoteResponseWithImages{
		ID:              q.ID,
		TypeOfServiceID: q.TypeOfServiceID,
		ClientID:        q.ClientID,
		Description:     q.Description,
		State:           string(q.State),
		Price:           q.Price,
		Time:            q.Time.Format(time.RFC3339),
		Images:          respImgs,
	}
}


// createQuoteRequest representa el cuerpo de la solicitud para crear una cotización
type createQuoteRequest struct {
	TypeOfServiceID string `form:"typeOfServiceID" binding:"required"`
	Description     string `form:"description" binding:"required"`
}

// CreateQuote godoc
//
// @Summary        Register a new quote
// @Description    create a new quote for a client with an attached file
// @Tags           Quotes
// @Accept         multipart/form-data
// @Produce        json
// @Param          typeOfServiceID  formData  string  true  "Type of Service ID (UUID format)"
// @Param          description      formData  string  true  "Description"
// @Param          file             formData  file    true  "Attachment file"
// @Success        200              {object}  quoteResponse  "Quote created"
// @Failure        400              {object}  errorResponse  "Validation error"
// @Failure        500              {object}  errorResponse  "Internal server error"
// @Router         /quotes [post]
func (qh *QuoteHandler) CreateQuote(ctx *gin.Context) {
	// Get user ID from auth token
	authPayload := getAuthPayload(ctx, authorizationPayloadKey)

	if authPayload == nil {
		handleError(ctx, domain.ErrUnauthorized)
		return
	}

	userID := authPayload.UserID

	// Parse multipart form
	if err := ctx.Request.ParseMultipartForm(10 << 20); err != nil {
		validationError(ctx, fmt.Errorf("failed to parse multipart form: %v", err))
		return
	}

	// Bind the form data
	var req createQuoteRequest
	if err := ctx.ShouldBind(&req); err != nil {
		validationError(ctx, err)
		return
	}

	// Parse UUIDs
	typeOfServiceID, err := uuid.Parse(req.TypeOfServiceID)
	if err != nil {
		validationError(ctx, fmt.Errorf("invalid typeOfServiceID format"))
		return
	}

	// Get the file
	file, fileHeader, err := ctx.Request.FormFile("file")
	if err != nil {
		validationError(ctx, fmt.Errorf("file is required: %v", err))
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		handleError(ctx, fmt.Errorf("failed to read file: %v", err))
		return
	}

	quote := domain.Quote{
		ID:              uuid.New(),
		TypeOfServiceID: typeOfServiceID,
		ClientID:        userID,
		Time:            time.Now(),
		Description:     req.Description,
		State:           "pending",
		Price:           0,
	}

	createdQuote, err := qh.svc.CreateQuote(ctx, &quote, fileBytes, fileHeader.Filename)
	if err != nil {
		handleError(ctx, err)
		return
	}

	rsp := newQuoteResponse(createdQuote)
	handleSuccess(ctx, rsp)
}

// listQuotesRequest representa los parámetros de la consulta para listar cotizaciones
type listQuotesRequest struct {
	TypeOfServiceID  *string 	`form:"typeOfServiceId"`
	ClientID     		 *string 	`form:"clientId"`
	StartDate   		 *string 	`form:"startDate"`
	EndDate   			 *string 	`form:"endDate"`
	State 		     	 *string 	`form:"state"`
	Skip  					 uint64 	`form:"skip" binding:"required,min=0"`
	Limit 					 uint64 	`form:"limit" binding:"required,min=5"`
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

	// Get user ID from auth token
	authPayload := getAuthPayload(ctx, authorizationPayloadKey)

	if authPayload.Role == domain.Client && req.ClientID != nil {
		handleError(ctx, domain.ErrUnauthorized)
		return
	}

	var typeOfServiceId *uuid.UUID

	if req.TypeOfServiceID != nil {
		id, err := uuid.Parse(*req.TypeOfServiceID)
		if err != nil {
			validationError(ctx, fmt.Errorf("invalid typeOfServiceId: %v", err))
			return
		}
		typeOfServiceId = &id
	}

	var clientId *uuid.UUID

	if req.ClientID != nil {
		id, err := uuid.Parse(*req.ClientID)
		if err != nil {
			validationError(ctx, fmt.Errorf("invalid clientId: %v", err))
			return
		}
		clientId = &id
	}

	if authPayload.Role == domain.Client {
		clientId = &authPayload.UserID
	}

	var startDate *time.Time

	if req.StartDate != nil {
		t, err := time.Parse(time.RFC3339, *req.StartDate)
		if err != nil {
			validationError(ctx, fmt.Errorf("invalid startDate format (must be RFC3339)"))
			return
		}
		startDate = &t
	}

	var endDate *time.Time

	if req.EndDate != nil {
		t, err := time.Parse(time.RFC3339, *req.EndDate)
		if err != nil {
			validationError(ctx, fmt.Errorf("invalid endDate format (must be RFC3339)"))
			return
		}
		endDate = &t
	}

	var state *domain.QuoteState

	if req.State != nil {
		s := domain.QuoteState(*req.State)
		if s != domain.QuotePending && s != domain.QuoteApproved &&
			s != domain.QuoteRejected && s != domain.QuoteRequiresProof {
			validationError(ctx, fmt.Errorf("invalid state value"))
			return
		}
		state = &s
	}

	// Build the filter
	filter := port.QuoteFilter{
		TypeOfServiceID: typeOfServiceId,
		ClientID: clientId,
		StartDate:  startDate,
		EndDate:    endDate,
		ByState:    state,
		Skip:       req.Skip,
		Limit:      req.Limit,
	}

	quotes, err := qh.svc.ListQuotes(ctx, filter)
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
//type getQuoteRequest struct {
//	ID string `form:"id" binding:"required"`
//}

// GetQuote godoc
//
//	@Summary		Get a quote
//	@Description	Get a quote by id
//	@Tags			Quotes
//	@Accept			json
//	@Produce		json
//	@Param			id	query		string		true	"Quote ID"
//	@Success		200	{object}	quoteResponse	"Quote displayed"
//	@Failure		400	{object}	errorResponse	"Validation error"
//	@Failure		404	{object}	errorResponse	"Data not found error"
//	@Failure		500	{object}	errorResponse	"Internal server error"
//	@Router			/quotes [get]

func (qh *QuoteHandler) GetQuote(ctx *gin.Context) {

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

	// Llamar al servicio para obtener la cotización
	quote, images, err := qh.svc.GetQuote(ctx, quoteID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	// Responder con la cotización
	rsp := newQuoteResponseWithImages(quote, images)
	handleSuccess(ctx, rsp)
}

// updateQuoteRequest representa el cuerpo de la solicitud para actualizar una cotización

type updateQuoteRequest struct {
	TypeOfServiceID *uuid.UUID `json:"typeOfServiceId,omitempty"`
	Description     *string    `json:"description,omitempty"`
	Price           *float64   `json:"price,omitempty"`
}

// UpdateQuote godoc
//
//	@Summary		Update a quote
//	@Description	Update a quote by id
//	@Tags			Quotes
//	@Accept			json
//	@Produce		json
//
// @Param         id     query   string        true   "Quote ID"
// @Param         quote  body    updateQuoteRequest   true   "Quote Data" (sin el ID)
//
//	@Success		200	{object}	quoteResponse	"Quote updated"
//	@Failure		400	{object}	errorResponse	"Validation error"
//	@Failure		404	{object}	errorResponse	"Data not found error"
//	@Failure		500	{object}	errorResponse	"Internal server error"
//	@Router			/quotes [put]
func (qh *QuoteHandler) UpdateQuote(ctx *gin.Context) {
	// Get user ID from auth token
	authPayload := getAuthPayload(ctx, authorizationPayloadKey)
	idStr := ctx.DefaultQuery("id", "")

	if idStr == "" {
		validationError(ctx, fmt.Errorf("ID is required"))
		return
	}

	// Convertir el string a un UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		validationError(ctx, fmt.Errorf("invalid UUID format"))
		return
	}

	quote, _, err := qh.svc.GetQuote(ctx, id)

	if err != nil {
		handleError(ctx, err)
		return
	}

	if authPayload.Role == domain.Client && quote.ClientID != authPayload.UserID {
		handleError(ctx, domain.ErrUnauthorized)
		return
	}

	// Inicializar la estructura para la solicitud
	var req updateQuoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validationError(ctx, err)
		return
	}

	if authPayload.Role == domain.Client && req.Price != nil {
		handleError(ctx, domain.ErrUnauthorized)
		return
	}

	quote = &domain.Quote{ ID: id }

	if req.TypeOfServiceID != nil {
		quote.TypeOfServiceID = *req.TypeOfServiceID
	}

	if req.Description != nil {
		quote.Description = *req.Description
	}

	if req.Price != nil {
		quote.Price = *req.Price
	}

	_, err = qh.svc.UpdateQuote(ctx, quote)
	if err != nil {
		handleError(ctx, err)
		return
	}

	// Responder con el resultado
	rsp := newQuoteResponse(quote)
	handleSuccess(ctx, rsp)
}

// deleteQuoteRequest representa el cuerpo de la solicitud para eliminar una cotización
//type deleteQuoteRequest struct {
//	ID uuid.UUID `query:"id" binding:"required"`
//}

// DeleteQuote godoc
//
//	@Summary		Delete a quote
//	@Description	Delete a quote by id
//	@Tags			Quotes
//	@Accept			json
//	@Produce		json
//
// @Param         id   query   string         true   "Quote ID"
//
//	@Success		200	{object}	string		"Quote deleted successfully"
//	@Failure		400	{object}	errorResponse	"Validation error"
//	@Failure		404	{object}	errorResponse	"Data not found error"
//	@Failure		500	{object}	errorResponse	"Internal server error"
//	@Router			/quotes [delete]
func (qh *QuoteHandler) DeleteQuote(ctx *gin.Context) {
	// Get user ID from auth token
	authPayload := getAuthPayload(ctx, authorizationPayloadKey)

	idStr := ctx.DefaultQuery("id", "")
	if idStr == "" {
		validationError(ctx, fmt.Errorf("ID is required"))
		return
	}

	// Convertir el string a un UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		validationError(ctx, fmt.Errorf("invalid UUID format"))
		return
	}

	quote, _, err := qh.svc.GetQuote(ctx, id)

	if err != nil {
		handleError(ctx, err)
		return
	}

	if authPayload.Role == domain.Client && quote.ClientID != authPayload.UserID {
		handleError(ctx, domain.ErrUnauthorized)
		return
	}

	// Llamar al servicio para eliminar la cotización
	err = qh.svc.DeleteQuote(ctx, id)
	if err != nil {
		handleError(ctx, err)
		return
	}
	handleSuccess(ctx, "Quote deleted successfully")
}

type changeQuoteStateRequest struct {
	State *string `json:"state,omitempty"`
}

// UpdateQuote godoc
//
//	@Summary		Update a quote
//	@Description	Update a quote by id
//	@Tags			Quotes
//	@Accept			json
//	@Produce		json
//
// @Param         id     query   string        true   "Quote ID"
// @Param         quote  body    updateQuoteRequest   true   "Quote Data" (sin el ID)
//
//	@Success		200	{object}	quoteResponse	"Quote updated"
//	@Failure		400	{object}	errorResponse	"Validation error"
//	@Failure		404	{object}	errorResponse	"Data not found error"
//	@Failure		500	{object}	errorResponse	"Internal server error"
//	@Router			/quotes [put]
func (qh *QuoteHandler) ChangeQuoteState(ctx *gin.Context) {
	idStr := ctx.DefaultQuery("id", "")

	if idStr == "" {
		validationError(ctx, fmt.Errorf("ID is required"))
		return
	}

	// Convertir el string a un UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		validationError(ctx, fmt.Errorf("invalid UUID format"))
		return
	}

	// Inicializar la estructura para la solicitud
	var req changeQuoteStateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validationError(ctx, err)
		return
	}

	var state domain.QuoteState

	if req.State != nil {
		s := domain.QuoteState(*req.State)
		if s != domain.QuoteApproved && s != domain.QuoteRequiresProof && s != domain.QuoteRejected {
				validationError(ctx, fmt.Errorf("invalid state value it has to be either 'approved', 'rejected' or 'requires_proof'"))
			return
		}
		state = s
	}

	quote, err := qh.svc.ChangeQuoteState(ctx, id, state)
	if err != nil {
		handleError(ctx, err)
		return
	}

	// Responder con el resultado
	rsp := newQuoteResponse(quote)
	handleSuccess(ctx, rsp)
}
