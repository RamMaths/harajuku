package postgres_test

import (
	"context"
	"testing"

	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/core/domain"

	"github.com/google/uuid"
)

func TestCreateTypeOfService(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewTypeOfServiceRepository(db)
	ctx := context.Background()

	service := &domain.TypeOfService{
		ID:    uuid.New(),
		Name:  "Test Service",
		Price: 99.99,
	}

	t.Logf("Creating TypeOfService with ID: %v", service.ID)

	created, err := repo.CreateTypeOfService(ctx, service)
	if err != nil {
		t.Fatalf("CreateTypeOfService failed: %v", err)
	}

	t.Logf("Created TypeOfService with ID: %v", created.ID)
}

func TestGetTypeOfServiceByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewTypeOfServiceRepository(db)
	ctx := context.Background()

	id := uuid.MustParse("00b35c18-d807-4625-a8e3-511131528a41")

	t.Logf("Fetching TypeOfService with ID: %v", id)

	_, err := repo.GetTypeOfServiceByID(ctx, id)
	if err != nil {
		t.Fatalf("GetTypeOfServiceByID failed: %v", err)
	}
}

func TestListTypeOfServices(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewTypeOfServiceRepository(db)
	ctx := context.Background()

	t.Logf("Listing TypeOfServices")

	services, err := repo.ListTypeOfServices(ctx, 1, 10)
	if err != nil {
		t.Fatalf("ListTypeOfServices failed: %v", err)
	}
	t.Logf("Retrieved %d TypeOfServices", len(services))
}

func TestUpdateTypeOfService(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewTypeOfServiceRepository(db)
	ctx := context.Background()

	id := uuid.New()
	service := &domain.TypeOfService{
		ID:    id,
		Name:  "Initial Name",
		Price: 45.0,
	}

	t.Logf("Creating original TypeOfService with ID: %v", id)
	_, err := repo.CreateTypeOfService(ctx, service)
	if err != nil {
		t.Fatalf("CreateTypeOfService failed: %v", err)
	}

	service.Name = "Updated Name"
	service.Price = 99.99

	t.Logf("Updating TypeOfService with ID: %v", id)
	updated, err := repo.UpdateTypeOfService(ctx, service)
	if err != nil {
		t.Fatalf("UpdateTypeOfService failed: %v", err)
	}

	t.Logf("Updated TypeOfService: %v", updated)
}

func TestDeleteTypeOfService(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewTypeOfServiceRepository(db)
	ctx := context.Background()

	service := &domain.TypeOfService{
		ID:    uuid.New(),
		Name:  "To Be Deleted",
		Price: 20.0,
	}

	_, err := repo.CreateTypeOfService(ctx, service)
	if err != nil {
		t.Fatalf("CreateTypeOfService (for deletion) failed: %v", err)
	}

	t.Logf("Deleting TypeOfService with ID: %v", service.ID)
	err = repo.DeleteTypeOfService(ctx, service.ID)
	if err != nil {
		t.Fatalf("DeleteTypeOfService failed: %v", err)
	}
}
