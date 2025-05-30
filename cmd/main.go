package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	paseto "harajuku/backend/internal/adapter/auth"
	"harajuku/backend/internal/adapter/communication/email"
	"harajuku/backend/internal/adapter/config"
	"harajuku/backend/internal/adapter/handler/http"
	"harajuku/backend/internal/adapter/logger"
	"harajuku/backend/internal/adapter/storage/awsS3"
	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/adapter/storage/postgres/repository"
	"harajuku/backend/internal/adapter/storage/redis"
	"harajuku/backend/internal/core/service"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
	// Load environment variables
	config, err := config.New()
	if err != nil {
		slog.Error("Error loading environment variables", "error", err)
		os.Exit(1)
	}

	// Set logger
	logger.Set(config.App)

	slog.Info("Starting the application", "app", config.App.Name, "env", config.App.Env)

	// Init database
	ctx := context.Background()
	db, err := postgres.New(ctx, config.DB)
	if err != nil {
		slog.Error("Error initializing database connection", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("Successfully connected to the database", "db", config.DB.Connection)

	// Migrate database
	err = db.Migrate()
	if err != nil {
		slog.Error("Error migrating database", "error", err)
		os.Exit(1)
	}

	slog.Info("Successfully migrated the database")

	// Init token service
	token, err := paseto.New(config.Token)
	if err != nil {
		slog.Error("Error initializing token service", "error", err)
		os.Exit(1)
	}

	// Init cache service
	cache, err := redis.New(ctx, config.Redis)
	if err != nil {
		slog.Error("Error initializing cache connection", "error", err)
		os.Exit(1)
	}
	defer cache.Close()

	slog.Info("Successfully connected to the cache server")

	// // Init token service
	// token, err := paseto.New(config.Token)
	// if err != nil {
	//     slog.Error("Error initializing token service", "error", err)
	//     os.Exit(1)
	// }

	// Dependency injection
	// User
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cache)
	userHandler := http.NewUserHandler(userService)

	// Auth
	authService := service.NewAuthService(userRepo, token)
	authHandler := http.NewAuthHandler(authService)

	// S3

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(config.AwsS3.Region),
	}))

	s3 := awsS3.NewAwsS3(sess, config.AwsS3.Bucket)

	// Email

	email, err := email.New(ctx, config.Email)

	if err != nil {
		slog.Error("Error initializing the email service", "error", err)
		os.Exit(1)
	}

	// TypeOfService
	typeOfServiceRepo := repository.NewTypeOfServiceRepository(db)
	typeOfServiceService := service.NewTypeOfServiceService(typeOfServiceRepo, cache)
	typeOfServiceHandler := http.NewTypeOfServiceHandler(typeOfServiceService)

	// Quote
	quoteRepo := repository.NewQuoteRepository(db)
	quoteImageRepo := repository.NewQuoteImageRepository(db)
	quoteService := service.NewQuoteService(quoteRepo, s3, userRepo, email, quoteImageRepo, typeOfServiceRepo, *db, cache)
	quoteHandler := http.NewQuoteHandler(quoteService)

	// AvailabilitySlot
	availabilitySlotRepo := repository.NewAvailabilitySlotRepository(db)
	availabilitySlotService := service.NewAvailabilitySlotService(availabilitySlotRepo, cache)
	availabilitySlotHandler := http.NewAvailabilitySlotHandler(availabilitySlotService, userService)

	// Appointment
	appointmentRepo := repository.NewAppointmentRepository(db)
	appointmentService := service.NewAppointmentService(appointmentRepo, quoteRepo, availabilitySlotRepo, cache)
	appointmentHandler := http.NewAppointmentHandler(appointmentService, userService)

	// PaymentProof
	paymentProofRepo := repository.NewPaymentProofRepository(db)
	paymentProofService := service.NewPaymentProofService(
		paymentProofRepo, // port.PaymentProofRepository
		s3,               // port.FileRepository
		quoteRepo,        // port.QuoteRepository
		email,            // port.EmailRepository
		*db,              // postgres.DB
		cache,            // port.CacheRepository
	)
	paymentProofHandler := http.NewPaymentProofHandler(paymentProofService)

	// Init router
	router, err := http.NewRouter(
		config.HTTP,
		token,
		*userHandler,
		*authHandler,
		*quoteHandler,
		*typeOfServiceHandler,
		*availabilitySlotHandler,
		*appointmentHandler,
		*paymentProofHandler,
	)

	if err != nil {
		slog.Error("Error initializing router", "error", err)
		os.Exit(1)
	}
	// Start server
	listenAddr := fmt.Sprintf("%s:%s", config.HTTP.URL, config.HTTP.Port)
	slog.Info("Starting the HTTP server", "listen_address", listenAddr)
	err = router.Serve(listenAddr)
	if err != nil {
		slog.Error("Error starting the HTTP server", "error", err)
		os.Exit(1)
	}
}
