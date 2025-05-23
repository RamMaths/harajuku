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
			quote.DELETE("", quoteHandler.DeleteQuote)
		}

		typeOfService := v1.Group("/typesofservice").Use(authMiddleware(token))
		{
			typeOfService.POST("", typeOfServiceHandler.CreateTypeOfService)
			typeOfService.GET("/all", typeOfServiceHandler.ListTypeOfServices)
			typeOfService.GET("", typeOfServiceHandler.GetTypeOfService)
			typeOfService.PUT("", typeOfServiceHandler.UpdateTypeOfService)
			typeOfService.DELETE("", typeOfServiceHandler.DeleteTypeOfService)
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
