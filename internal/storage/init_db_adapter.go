package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Transaction represents a micro-transaction record.
type Transaction struct {
	ID        string
	FromAcct  string
	ToAcct    string
	Amount    int64
	Currency  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PostgresAdapter wraps the DB connection.
type PostgresAdapter struct {
	db *sql.DB
}

// NewPostgresAdapter opens a connection using the provided DSN.
func NewPostgresAdapter(ctx context.Context, dsn string) (*PostgresAdapter, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &PostgresAdapter{db: db}, nil
}

// Migrate ensures the schema exists.
func (p *PostgresAdapter) Migrate(ctx context.Context) error {
	query := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	CREATE TABLE IF NOT EXISTS transactions (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		from_acct TEXT NOT NULL,
		to_acct TEXT NOT NULL,
		amount BIGINT NOT NULL,
		currency VARCHAR(3) NOT NULL,
		status VARCHAR(20) NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`
	_, err := p.db.ExecContext(ctx, query)
	return err
}

// Create inserts a new transaction.
func (p *PostgresAdapter) Create(ctx context.Context, t *Transaction) (string, error) {
	query := `
	INSERT INTO transactions (from_acct, to_acct, amount, currency, status)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at, updated_at;`

	row := p.db.QueryRowContext(ctx, query, t.FromAcct, t.ToAcct, t.Amount, t.Currency, t.Status)
	var id string
	if err := row.Scan(&id, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return "", fmt.Errorf("scan create: %w", err)
	}
	t.ID = id
	return id, nil
}

// Get retrieves a transaction by ID.
func (p *PostgresAdapter) Get(ctx context.Context, id string) (*Transaction, error) {
	query := `
	SELECT id, from_acct, to_acct, amount, currency, status, created_at, updated_at
	FROM transactions WHERE id = $1;`
	row := p.db.QueryRowContext(ctx, query, id)
	var t Transaction
	if err := row.Scan(&t.ID, &t.FromAcct, &t.ToAcct, &t.Amount, &t.Currency, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, fmt.Errorf("scan get: %w", err)
	}
	return &t, nil
}
