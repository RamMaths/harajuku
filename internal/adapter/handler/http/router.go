package http

import (
	"log/slog"
	"net/http"
	"strings"

	"harajuku/backend/internal/adapter/config"
	"harajuku/backend/internal/core/port"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	sloggin "github.com/samber/slog-gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Router is a wrapper for HTTP router
type Router struct {
	*gin.Engine
}

// NewRouter creates a new HTTP router
func NewRouter(
	config *config.HTTP,
	token port.TokenService,
	userHandler UserHandler,
	authHandler AuthHandler,
	quoteHandler QuoteHandler,
	typeOfServiceHandler TypeOfServiceHandler,
	availabilitySlotHandler AvailabilitySlotHandler,
	appointmentHandler AppointmentHandler,
	paymentProofHandler PaymentProofHandler,
	quoteImageHandler QuoteImageHandler,
) (*Router, error) {
	// Disable debug mode in production
	if config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Custom validators
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		if err := v.RegisterValidation("user_role", userRoleValidator); err != nil {
			return nil, err
		}
	}

	// CORS
	ginConfig := cors.DefaultConfig()
	allowedOrigins := config.AllowedOrigins
	originsList := strings.Split(allowedOrigins, ",")
	ginConfig.AllowOrigins = originsList
	ginConfig.AllowHeaders = append(
		ginConfig.AllowHeaders,
		"Authorization",
		"Content-Type",
		"X-Requested-With",
	)
	ginConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	// If you need the frontend to read the token or other headers back:
	ginConfig.ExposeHeaders = []string{"Authorization"}
	ginConfig.AllowCredentials = true
	router := gin.New()
	router.Use(cors.New(ginConfig), sloggin.New(slog.Default()), gin.Recovery())

	// Swagger
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//Options
	router.OPTIONS("/*any", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	// API
	v1 := router.Group("/v1")

	// Users (unauthenticated + authenticated)
	v1.POST("/users/", userHandler.Register)
	v1.POST("/users/login", authHandler.Login)
	v1.GET("/users/", authMiddleware(token), userHandler.ListUsers)
	v1.GET("/users/:id", authMiddleware(token), userHandler.GetUser)

	// Quotes (authenticated, admin for PATCH)
	v1.POST("/quotes", authMiddleware(token), quoteHandler.CreateQuote)
	v1.GET("/quotes/all", authMiddleware(token), quoteHandler.ListQuotes)
	v1.GET("/quotes", authMiddleware(token), quoteHandler.GetQuote)
	v1.PUT("/quotes", authMiddleware(token), quoteHandler.UpdateQuote)
	v1.PATCH("/quotes/state", authMiddleware(token), adminMiddleware(), quoteHandler.ChangeQuoteState)
	v1.DELETE("/quotes", authMiddleware(token), quoteHandler.DeleteQuote)

	// TypeOfService (authenticated, admin for write ops)
	v1.GET("/typesofservice/all", authMiddleware(token), typeOfServiceHandler.ListTypeOfServices)
	v1.GET("/typesofservice", authMiddleware(token), typeOfServiceHandler.GetTypeOfService)
	v1.POST("/typesofservice", authMiddleware(token), adminMiddleware(), typeOfServiceHandler.CreateTypeOfService)
	v1.PUT("/typesofservice", authMiddleware(token), adminMiddleware(), typeOfServiceHandler.UpdateTypeOfService)
	v1.DELETE("/typesofservice", authMiddleware(token), adminMiddleware(), typeOfServiceHandler.DeleteTypeOfService)

	// AvailabilitySlots
	v1.POST("/availabilityslots", authMiddleware(token), adminMiddleware(), availabilitySlotHandler.CreateSlot)
	v1.GET("/availabilityslots", authMiddleware(token), availabilitySlotHandler.ListSlots)
	v1.PUT("/availabilityslots", authMiddleware(token), adminMiddleware(), availabilitySlotHandler.UpdateSlot)
	v1.DELETE("/availabilityslots", authMiddleware(token), adminMiddleware(), availabilitySlotHandler.DeleteSlot)

	// Appointments (authenticated)
	v1.POST("/appointments", authMiddleware(token), appointmentHandler.CreateAppointment)
	v1.GET("/appointments/all", authMiddleware(token), appointmentHandler.ListAppointments)
	v1.GET("/appointments", authMiddleware(token), appointmentHandler.GetAppointment)
	v1.PUT("/appointments", authMiddleware(token), appointmentHandler.UpdateAppointment)
	v1.DELETE("/appointments", authMiddleware(token), appointmentHandler.DeleteAppointment)

	// PaymentProofs (authenticated, admin for write ops)
	v1.POST("/paymentproofs", authMiddleware(token), paymentProofHandler.CreatePaymentProof)
	v1.GET("/paymentproofs", authMiddleware(token), paymentProofHandler.GetPaymentProofByID)
	v1.GET("/paymentproofs/all", authMiddleware(token), paymentProofHandler.GetPaymentProofs)
	v1.PUT("/paymentproofs", authMiddleware(token), adminMiddleware(), paymentProofHandler.UpdatePaymentProof)
	v1.DELETE("/paymentproofs", authMiddleware(token), adminMiddleware(), paymentProofHandler.DeletePaymentProof)

	// QuoteImages (authenticated, admin for delete)
	v1.GET("/quoteimages/all", authMiddleware(token), quoteImageHandler.GetQuoteImages)
	v1.GET("/quoteimages", authMiddleware(token), quoteImageHandler.GetQuoteImageByID)
	v1.DELETE("/quoteimages", authMiddleware(token), adminMiddleware(), quoteImageHandler.DeleteQuoteImage)
	// If adding these later:
	// v1.POST("/quoteimages", authMiddleware(token), quoteImageHandler.CreateQuoteImage)
	// v1.PUT("/quoteimages", authMiddleware(token), quoteImageHandler.UpdateQuoteImage)

	return &Router{
		router,
	}, nil
}

// Serve starts the HTTP server
func (r *Router) Serve(listenAddr string) error {
	return r.Run(listenAddr)
}
