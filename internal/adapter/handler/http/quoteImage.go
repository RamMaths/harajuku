package http

import (
	"fmt"
	"net/http"
	"path/filepath"

	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type QuoteImageHandler struct {
	svc port.QuoteImageService
}

func NewQuoteImageHandler(svc port.QuoteImageService) *QuoteImageHandler {
	return &QuoteImageHandler{svc: svc}
}

func newQuoteImageResponse(q *domain.QuoteImage) quoteImageResponse {
	return quoteImageResponse{
		ID:      q.ID,
		QuoteID: q.QuoteID,
		URL:     q.URL,
	}
}

// GetQuoteImageByID descarga la imagen asociada al ID
func (h *QuoteImageHandler) GetQuoteImageByID(ctx *gin.Context) {
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

	// Obtener metadatos + archivo desde el servicio
	quoteImage, fileData, err := h.svc.GetQuoteImageByID(ctx, id)
	if err != nil {
		handleError(ctx, err)
		return
	}

	// Detectar tipo MIME para respuesta correcta
	mimeType := http.DetectContentType(fileData)

	// Establecer headers para descargar el archivo con el nombre original
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(quoteImage.URL)))
	ctx.Data(http.StatusOK, mimeType, fileData)
}

// GetQuoteImages lista imágenes filtrando por quoteId
func (h *QuoteImageHandler) GetQuoteImages(ctx *gin.Context) {
	var quoteID *uuid.UUID
	var skip, limit uint64

	if quoteIDStr := ctx.Query("quoteId"); quoteIDStr != "" {
		parsedID, err := uuid.Parse(quoteIDStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "quoteId inválido"})
			return
		}
		quoteID = &parsedID
	}

	if skipStr := ctx.Query("skip"); skipStr != "" {
		_, err := fmt.Sscan(skipStr, &skip)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "skip inválido"})
			return
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		_, err := fmt.Sscan(limitStr, &limit)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "limit inválido"})
			return
		}
	}

	quoteImages, err := h.svc.GetQuoteImages(ctx, quoteID, skip, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []quoteImageResponse
	for _, q := range quoteImages {
		response = append(response, newQuoteImageResponse(&q))
	}

	ctx.JSON(http.StatusOK, response)
}

// DeleteQuoteImage elimina una imagen de cotización por ID
func (h *QuoteImageHandler) DeleteQuoteImage(ctx *gin.Context) {
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

	if err := h.svc.DeleteQuoteImage(ctx, id); err != nil {
		handleError(ctx, err)
		return
	}

	handleSuccess(ctx, "Quote image deleted successfully")
}
