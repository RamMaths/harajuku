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
	{
		user := v1.Group("/users")
		{
			user.POST("/", userHandler.Register)
			user.POST("/login", authHandler.Login)

			authUser := user.Group("/").Use(authMiddleware(token))
			{
				authUser.GET("/", userHandler.ListUsers)
				authUser.GET("/:id", userHandler.GetUser)
			}
		}

		quote := v1.Group("/quotes").Use(authMiddleware(token))
		{
			quote.POST("", quoteHandler.CreateQuote)
			quote.GET("/all", quoteHandler.ListQuotes)
			quote.GET("", quoteHandler.GetQuote)
			quote.PUT("", quoteHandler.UpdateQuote)
			quote.PATCH("/state", quoteHandler.ChangeQuoteState).Use(adminMiddleware())
			quote.DELETE("", quoteHandler.DeleteQuote)
		}

		typeOfService := v1.Group("/typesofservice").Use(authMiddleware(token))
		{
			typeOfService.POST("", typeOfServiceHandler.CreateTypeOfService).Use(adminMiddleware())
			typeOfService.GET("/all", typeOfServiceHandler.ListTypeOfServices).Use(adminMiddleware())
			typeOfService.GET("", typeOfServiceHandler.GetTypeOfService).Use(adminMiddleware())
			typeOfService.PUT("", typeOfServiceHandler.UpdateTypeOfService).Use(adminMiddleware())
			typeOfService.DELETE("", typeOfServiceHandler.DeleteTypeOfService).Use(adminMiddleware())
		}

		availabilitySlot := v1.Group("/availabilityslots").Use(authMiddleware(token))
		{
			availabilitySlot.POST("", availabilitySlotHandler.CreateSlot).Use(adminMiddleware())
			availabilitySlot.GET("", availabilitySlotHandler.ListSlots).Use(adminMiddleware())
			availabilitySlot.PUT("", availabilitySlotHandler.UpdateSlot).Use(adminMiddleware())
			availabilitySlot.DELETE("", availabilitySlotHandler.DeleteSlot).Use(adminMiddleware())
		}

		appointments := v1.Group("/appointments").Use(authMiddleware(token))
		{
			appointments.POST("", appointmentHandler.CreateAppointment)
			appointments.GET("/all", appointmentHandler.ListAppointments)
			appointments.GET("", appointmentHandler.GetAppointment)
			appointments.PUT("", appointmentHandler.UpdateAppointment)
			appointments.DELETE("", appointmentHandler.DeleteAppointment)
		}
	}

	return &Router{
		router,
	}, nil
}

// Serve starts the HTTP server
func (r *Router) Serve(listenAddr string) error {
	return r.Run(listenAddr)
}
