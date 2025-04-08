package repository

import (
	"context"
	"harajuku/backend/internal/adapter/config"
	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/adapter/storage/postgres/repository"
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/test/adapter/storage/postgres/helpers"
	"testing"

	"github.com/google/uuid"
)

const PORT = "5435"

func TestCreateUsersIntegration(t *testing.T) {
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

	// overwrite the db URL to point to container
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// apply migrations
	err = db.Migrate()
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

  repo := repository.NewUserRepository(db)

	// insert test data
  users_arr := []domain.User {
    {
      ID: uuid.New(),
      Name: "Ramses",
      Email: "ramses.hdz30@gmail.com",
      LastName: "Mata",
    },
    {
      ID: uuid.New(),
      Name: "Mariana",
      Email: "m.mata@gmail.com",
      LastName: "Mata",
    },
  }

  for _, a := range users_arr {
    _, err := repo.CreateUser(ctx, &a)
    if err != nil {
      t.Fatalf("failed to crate user: %v", err)
    }
  }
}

func TestListUsersIntegration(t *testing.T) {
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

	// overwrite the db URL to point to container
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// apply migrations
	err = db.Migrate()
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

  // Verify migrations were applied
    var exists bool
    err = db.Pool.QueryRow(ctx, 
        `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'users')`).Scan(&exists)
    if err != nil || !exists {
        t.Fatalf("migrations verification failed: %v (exists: %v)", err, exists)
    }

  repo := repository.NewUserRepository(db)

  tx, err := db.Pool.Begin(ctx)

  if err != nil {
      t.Fatalf("failed to begin transaction: %v", err)
  }

  defer tx.Rollback(ctx) // Safe rollback if commit fails

	// insert test data
  users_arr := []domain.User {
    {
      ID: uuid.New(),
      Name: "Ramses",
      Email: "ramses.hdz30@gmail.com",
      LastName: "Mata",
    },
    {
      ID: uuid.New(),
      Name: "Mariana",
      Email: "m.mata@gmail.com",
      LastName: "Mata",
    },
  }

  for _, a := range users_arr {
    _, err := repo.CreateUser(ctx, &a)
    if err != nil {
      t.Fatalf("failed to crate user: %v", err)
    }
  }

  // Commit the test data
  if err := tx.Commit(ctx); err != nil {
      t.Fatalf("failed to commit test data: %v", err)
  }

  // Verify data exists
  var count int
  err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
  if err != nil || count != 2 {
      t.Fatalf("data verification failed: %v (count: %d)", err, count)
  }

  // Now test ListUsers
  filters := domain.UserFilters{}
  result, err := repo.ListUsers(ctx, 0, 10, filters) // Skip=0 to get first page
  if err != nil {
      t.Fatalf("failed to list users: %v", err)
  }

  if len(result) != 2 {
      t.Errorf("expected 2 users, got %d", len(result))
      t.Logf("Actual users: %+v", result) // Debug output
  }
}

func TestListUsersIntegrationAndCheckFilters(t *testing.T) {
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

	// overwrite the db URL to point to container
	db, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// apply migrations
	err = db.Migrate()
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

  // Verify migrations were applied
    var exists bool
    err = db.Pool.QueryRow(ctx, 
        `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'users')`).Scan(&exists)
    if err != nil || !exists {
        t.Fatalf("migrations verification failed: %v (exists: %v)", err, exists)
    }

  repo := repository.NewUserRepository(db)

  tx, err := db.Pool.Begin(ctx)

  if err != nil {
      t.Fatalf("failed to begin transaction: %v", err)
  }

  defer tx.Rollback(ctx) // Safe rollback if commit fails

	// insert test data
  users_arr := []domain.User {
    {
      ID: uuid.New(),
      Name: "Ramses",
      Email: "ramses.hdz30@gmail.com",
      LastName: "Mata",
    },
    {
      ID: uuid.New(),
      Name: "Mariana",
      Email: "m.mata@gmail.com",
      LastName: "Mata",
    },
  }

  for _, a := range users_arr {
    _, err := repo.CreateUser(ctx, &a)
    if err != nil {
      t.Fatalf("failed to crate user: %v", err)
    }
  }

  // Commit the test data
  if err := tx.Commit(ctx); err != nil {
      t.Fatalf("failed to commit test data: %v", err)
  }

  // Verify data exists
  var count int
  err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
  if err != nil || count != 2 {
      t.Fatalf("data verification failed: %v (count: %d)", err, count)
  }

  // Now test ListUsers
  filters := domain.UserFilters{
    Role: "client",
  }
  result, err := repo.ListUsers(ctx, 0, 10, filters) // Skip=0 to get first page
  if err != nil {
      t.Fatalf("failed to list users: %v", err)
  }

  if len(result) != 2 {
      t.Errorf("expected 2 users, got %d", len(result))
      t.Logf("Actual users: %+v", result) // Debug output
  }
}
