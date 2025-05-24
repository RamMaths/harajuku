package repository

import (
	"context"
	"harajuku/backend/internal/adapter/config"
	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"harajuku/backend/test/adapter/storage/postgres/helpers"
	"testing"
	"time"

	"github.com/google/uuid"
)

const AS_PORT = "5435"

func TestCreateAvailabilitySlotIntegration(t *testing.T) {
	testContainer := helpers.SetupTestDB(t)
	defer testContainer.Teardown()

	ctx := context.Background()

	cfg := &config.DB{
		Connection: "postgres",
		User:       "user",
		Password:   "secret",
		Host:       "localhost",
		Port:       PORT,
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

	repo := postgres.NewAvailabilitySlotRepository(db)

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

	slot := domain.AvailabilitySlot{
		ID:        uuid.New(),
		AdminID:   adminID,
		StartTime: time.Now().UTC(),
		EndTime:   time.Now().UTC().Add(1 * time.Hour),
		IsBooked:  false,
	}

	createdSlot, err := repo.CreateAvailabilitySlot(ctx, &slot)
	if err != nil {
		t.Fatalf("failed to create slot: %v", err)
	}

	if createdSlot.ID != slot.ID {
		t.Errorf("expected slot ID %v, got %v", slot.ID, createdSlot.ID)
	}
}

func TestGetAvailabilitySlotByIDIntegration(t *testing.T) {
	testContainer := helpers.SetupTestDB(t)
	defer testContainer.Teardown()

	ctx := context.Background()

	cfg := &config.DB{
		Connection: "postgres",
		User:       "user",
		Password:   "secret",
		Host:       "localhost",
		Port:       PORT,
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

	repo := postgres.NewAvailabilitySlotRepository(db)

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

	slotID := uuid.New()
	_, err = db.Exec(ctx,
		`INSERT INTO "AvailabilitySlot" (id, "adminId", "startTime", "endTime", "isBooked") 
		VALUES ($1, $2, $3, $4, $5)`,
		slotID, adminID, time.Now().UTC(), time.Now().UTC().Add(1*time.Hour), false,
	)
	if err != nil {
		t.Fatalf("failed to insert test slot: %v", err)
	}

	retrievedSlot, err := repo.GetAvailabilitySlotByID(ctx, slotID)
	if err != nil {
		t.Fatalf("failed to get slot by ID: %v", err)
	}

	if retrievedSlot.ID != slotID {
		t.Errorf("expected slot ID %v, got %v", slotID, retrievedSlot.ID)
	}
}

func TestListAvailabilitySlotsIntegration(t *testing.T) {
	testContainer := helpers.SetupTestDB(t)
	defer testContainer.Teardown()

	ctx := context.Background()

	cfg := &config.DB{
		Connection: "postgres",
		User:       "user",
		Password:   "secret",
		Host:       "localhost",
		Port:       PORT,
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

	repo := postgres.NewAvailabilitySlotRepository(db)

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

	// Insert test slots
	slots := []domain.AvailabilitySlot{
		{
			ID:        uuid.New(),
			AdminID:   adminID,
			StartTime: time.Date(2025, 5, 1, 9, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC),
			IsBooked:  false,
		},
		{
			ID:        uuid.New(),
			AdminID:   adminID,
			StartTime: time.Date(2025, 5, 2, 9, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 5, 2, 10, 0, 0, 0, time.UTC),
			IsBooked:  true,
		},
	}

	for _, s := range slots {
		_, err := repo.CreateAvailabilitySlot(ctx, &s)
		if err != nil {
			t.Fatalf("failed to create slot: %v", err)
		}
	}

	// Test listing with month filter
	filter := port.AvailabilitySlotFilter{
		UserID: &adminID,
		Month:  ptrString("2025-05"),
		Skip:   1,
		Limit:  10,
	}

	listedSlots, err := repo.ListAvailabilitySlots(ctx, filter)
	if err != nil {
		t.Fatalf("failed to list slots: %v", err)
	}

	if len(listedSlots) != 2 {
		t.Errorf("expected 2 slots, got %d", len(listedSlots))
	}
}

func TestUpdateAvailabilitySlotIntegration(t *testing.T) {
	testContainer := helpers.SetupTestDB(t)
	defer testContainer.Teardown()

	ctx := context.Background()

	cfg := &config.DB{
		Connection: "postgres",
		User:       "user",
		Password:   "secret",
		Host:       "localhost",
		Port:       PORT,
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

	repo := postgres.NewAvailabilitySlotRepository(db)

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

	slot := domain.AvailabilitySlot{
		ID:        uuid.New(),
		AdminID:   adminID,
		StartTime: time.Now().UTC(),
		EndTime:   time.Now().UTC().Add(1 * time.Hour),
		IsBooked:  false,
	}

	createdSlot, err := repo.CreateAvailabilitySlot(ctx, &slot)
	if err != nil {
		t.Fatalf("failed to create slot: %v", err)
	}

	// Update slot
	newEndTime := createdSlot.EndTime.Add(30 * time.Minute)
	createdSlot.EndTime = newEndTime
	createdSlot.IsBooked = true

	_, err = repo.UpdateAvailabilitySlot(ctx, createdSlot)
	if err != nil {
		t.Fatalf("failed to update slot: %v", err)
	}

}

func TestDeleteAvailabilitySlotIntegration(t *testing.T) {
	testContainer := helpers.SetupTestDB(t)
	defer testContainer.Teardown()

	ctx := context.Background()

	cfg := &config.DB{
		Connection: "postgres",
		User:       "user",
		Password:   "secret",
		Host:       "localhost",
		Port:       PORT,
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

	repo := postgres.NewAvailabilitySlotRepository(db)

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

	slot := domain.AvailabilitySlot{
		ID:        uuid.New(),
		AdminID:   adminID,
		StartTime: time.Now().UTC(),
		EndTime:   time.Now().UTC().Add(1 * time.Hour),
		IsBooked:  false,
	}

	createdSlot, err := repo.CreateAvailabilitySlot(ctx, &slot)
	if err != nil {
		t.Fatalf("failed to create slot: %v", err)
	}

	// Delete slot
	err = repo.DeleteAvailabilitySlot(ctx, createdSlot.ID)
	if err != nil {
		t.Fatalf("failed to delete slot: %v", err)
	}

	// Verify deletion
	_, err = repo.GetAvailabilitySlotByID(ctx, createdSlot.ID)
	if err == nil {
		t.Error("expected error when getting deleted slot, got nil")
	}
}

func ptrString(s string) *string {
	return &s
}
