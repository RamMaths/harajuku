package http

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentProofHandler struct {
	svc port.PaymentProofService
}

func NewPaymentProofHandler(svc port.PaymentProofService) *PaymentProofHandler {
	return &PaymentProofHandler{svc: svc}
}

// Request para creación
type createPaymentProofRequest struct {
	QuoteID string `json:"quoteId" binding:"required,uuid"`
	URL     string `json:"url" binding:"required,url"`
	// IsReviewed no se envía al crear, siempre empieza en false
}

// Request para actualización (puede actualizar URL e IsReviewed)
type updatePaymentProofRequest struct {
	ID         string `json:"id" binding:"required,uuid"`
	QuoteID    string `json:"quoteId" binding:"required,uuid"`
	URL        string `json:"url" binding:"required,url"`
	IsReviewed *bool  `json:"isReviewed"` // Opcional para actualizar estado
}

// Response común
type paymentProofResponse struct {
	ID         uuid.UUID `json:"id"`
	QuoteID    uuid.UUID `json:"quoteId"`
	URL        string    `json:"url"`
	IsReviewed bool      `json:"isReviewed"`
}

func newPaymentProofResponse(p *domain.PaymentProof) paymentProofResponse {
	return paymentProofResponse{
		ID:         p.ID,
		QuoteID:    p.QuoteID,
		URL:        p.URL,
		IsReviewed: p.IsReviewed,
	}
}

// CreatePaymentProof crea un comprobante nuevo
func (h *PaymentProofHandler) CreatePaymentProof(ctx *gin.Context) {
	// Parse multipart form
	if err := ctx.Request.ParseMultipartForm(10 << 20); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse multipart form: " + err.Error()})
		return
	}

	quoteIDStr := ctx.Request.FormValue("quoteId")
	if quoteIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "quoteId is required"})
		return
	}

	quoteID, err := uuid.Parse(quoteIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid quoteId format"})
		return
	}

	file, fileHeader, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required: " + err.Error()})
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file: " + err.Error()})
		return
	}

	paymentProof := &domain.PaymentProof{
		ID:         uuid.New(),
		QuoteID:    quoteID,
		IsReviewed: false, // default
	}

	created, err := h.svc.CreatePaymentProof(ctx, paymentProof, fileBytes, fileHeader.Filename)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, newPaymentProofResponse(created))
}

// GetPaymentProof obtiene un comprobante por ID y descarga su archivo
func (h *PaymentProofHandler) GetPaymentProofByID(ctx *gin.Context) {
	idStr := ctx.DefaultQuery("id", "")
	if idStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID parameter is required"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Obtener metadatos + archivo desde S3
	paymentProof, fileData, err := h.svc.GetPaymentProofByID(ctx, id)
	if err != nil {
		handleError(ctx, err)
		return
	}

	// Detectar tipo MIME
	mimeType := http.DetectContentType(fileData)

	// Establecer headers para descarga
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(paymentProof.URL)))
	ctx.Data(http.StatusOK, mimeType, fileData)
}

// ListPaymentProofs lista comprobantes con filtros opcionales: quoteId e isReviewed
func (h *PaymentProofHandler) GetPaymentProofs(ctx *gin.Context) {
	filter := port.PaymentProofFilter{}

	if quoteIDStr := ctx.Query("quoteId"); quoteIDStr != "" {
		if quoteID, err := uuid.Parse(quoteIDStr); err == nil {
			filter.QuoteID = &quoteID
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "quoteId inválido"})
			return
		}
	}

	if isReviewedStr := ctx.Query("isReviewed"); isReviewedStr != "" {
		if isReviewedStr == "true" {
			val := true
			filter.IsReviewed = &val
		} else if isReviewedStr == "false" {
			val := false
			filter.IsReviewed = &val
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "isReviewed debe ser true o false"})
			return
		}
	}

	// Paginación opcional
	if skipStr := ctx.Query("skip"); skipStr != "" {
		var skip uint64
		_, err := fmt.Sscan(skipStr, &skip)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "skip inválido"})
			return
		}
		filter.Skip = skip
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		var limit uint64
		_, err := fmt.Sscan(limitStr, &limit)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "limit inválido"})
			return
		}
		filter.Limit = limit
	}

	paymentProofs, err := h.svc.GetPaymentProofs(ctx, filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []paymentProofResponse
	for _, p := range paymentProofs {
		response = append(response, newPaymentProofResponse(&p))
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdatePaymentProof actualiza solo el campo IsReviewed de un comprobante existente
func (h *PaymentProofHandler) UpdatePaymentProof(ctx *gin.Context) {
	var req struct {
		ID         string `json:"id" binding:"required"`
		IsReviewed *bool  `json:"is_reviewed" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	existing, _, err := h.svc.GetPaymentProofByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No encontrado"})
		return
	}

	// Actualiza solo IsReviewed
	existing.IsReviewed = *req.IsReviewed

	updated, err := h.svc.UpdatePaymentProof(ctx, existing)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, newPaymentProofResponse(updated))
}

// DeletePaymentProof elimina un comprobante por ID
func (h *PaymentProofHandler) DeletePaymentProof(ctx *gin.Context) {
	idStr := ctx.DefaultQuery("id", "")
	if idStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID parameter is required"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		validationError(ctx, fmt.Errorf("invalid id format"))
		return
	}

	if err := h.svc.DeletePaymentProof(ctx, id); err != nil {
		handleError(ctx, err)
		return
	}

	handleSuccess(ctx, "Payment proof deleted successfully")
}
