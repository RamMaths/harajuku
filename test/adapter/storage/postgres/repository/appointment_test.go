package repository

import (
	"context"
	"harajuku/backend/internal/adapter/config"
	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/test/adapter/storage/postgres/helpers"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateAppointmentIntegration(t *testing.T) {
	testContainer := helpers.SetupTestDB(t)
	defer testContainer.Teardown()

	ctx := context.Background()

	cfg := &config.DB{
		Connection: "postgres",
		User:       "user",
		Password:   "secret",
		Host:       "localhost",
		Port:       testContainer.PORT,
		Name:       "testdb",
	}

	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	err = db.Migrate()
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := postgres.NewAppointmentRepository(db)

	// Insert test admin user first
	adminID := uuid.New()
	_, err = db.Exec(ctx, `INSERT INTO "users" ("id", "name", "lastName", "secondLastName", "email", "password", "role")
	VALUES (
		$1,
		'Juan',
		'Pérez',
		'González',
		'juan.perez@example.com',
		'hashed_password_aqui',
		'admin' 
	);
	`, adminID)
	if err != nil {
		t.Fatalf("failed to insert test admin: %v", err)
	}

	// Insert customer user first
	clientID := uuid.New()
	_, err = db.Exec(ctx, `INSERT INTO "users" ("id", "name", "lastName", "email", "password", "role")
	VALUES (
		$1,
		'Kevin',
		'Rodríguez',
		'kevin.rdz@example.com',
		'hashed_password_aqui',
		'client' 
	);
	`, clientID)
	if err != nil {
		t.Fatalf("failed to insert test admin: %v", err)
	}

	quote := domain.Quote{
		ID:              uuid.New(),
		TypeOfServiceID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ClientID:        uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Time:            time.Now(),
		Description:     "Test Create Quote",
		State:           "pending",
		Price:           100.50,
	}

	slot := domain.AvailabilitySlot{
		ID:        uuid.New(),
		AdminID:   adminID,
		StartTime: time.Now().UTC(),
		EndTime:   time.Now().UTC().Add(1 * time.Hour),
		IsBooked:  false,
	}

	appointment := domain.Appointment{
		ID:        uuid.New(),
		UserID:   clientID,
		SlotId: slot.ID,
		QuoteId:   quote.ID,
		Status:  domain.Booked,
	}

	createdSlot, err := repo.CreateAppointment(ctx, &appointment)
	if err != nil {
		t.Fatalf("failed to create slot: %v", err)
	}

	if createdSlot.ID != slot.ID {
		t.Errorf("expected slot ID %v, got %v", slot.ID, createdSlot.ID)
	}
}

