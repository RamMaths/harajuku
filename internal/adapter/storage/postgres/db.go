package postgres

import (
	"context"
	"embed"
	"fmt"

	"harajuku/backend/internal/adapter/config"

	"github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// migrationsFS embeds the migrations folder
//go:embed migrations/*.sql
var migrationsFS embed.FS

// Conn defines the minimal interface for executing queries
// both pgxpool.Pool and pgx.Tx implement these methods.
type Conn interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// DB wraps a connection (pool or transaction) and squirrel builder
type DB struct {
	Conn         Conn
	QueryBuilder squirrel.StatementBuilderType
	url          string
}

// New creates a new database pool instance
type DBPool struct { *DB }

func New(ctx context.Context, cfg *config.DB) (*DB, error) {
	url := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Connection, cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	return &DB{Conn: pool, QueryBuilder: builder, url: url}, nil
}

// BeginTx starts a new transaction and returns a new DB wrapping it
func (db *DB) BeginTx(ctx context.Context) (*DB, error) {
	tx, err := db.Conn.(*pgxpool.Pool).Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	return &DB{Conn: tx, QueryBuilder: builder, url: db.url}, nil
}

// WithTx executes fn inside a transaction, rolling back on error
func (db *DB) WithTx(ctx context.Context, fn func(txDB *DB) error) error {
	// Start transaction
	txDB, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	// Ensure rollback on panic or error
	defer func() {
		_ = txDB.Conn.(interface{ Rollback(ctx context.Context) error }).Rollback(ctx)
	}()
	// Execute user function
	if err := fn(txDB); err != nil {
		return err
	}
	// Commit
	return txDB.Conn.(interface{ Commit(ctx context.Context) error }).Commit(ctx)
}

// Migrate runs DB migrations
func (db *DB) Migrate() error {
	driver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", driver, db.url)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// ErrorCode extracts Postgres error code
func (db *DB) ErrorCode(err error) string {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code
	}
	return ""
}

// Close closes the underlying pool
func (db *DB) Close() {
	if pool, ok := db.Conn.(*pgxpool.Pool); ok {
		pool.Close()
	}
}
