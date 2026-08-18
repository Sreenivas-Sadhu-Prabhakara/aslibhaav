package backend

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStore persists evaluations in PostgreSQL.
type PostgresStore struct {
	pool *pgxpool.Pool
}

const schemaDDL = `
CREATE TABLE IF NOT EXISTS evaluations (
	id         BIGSERIAL PRIMARY KEY,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	input_json JSONB NOT NULL,
	effective  DOUBLE PRECISION NOT NULL,
	note       TEXT NOT NULL DEFAULT ''
);`

// NewPostgresStore connects, verifies the connection, and ensures the schema.
func NewPostgresStore(ctx context.Context, url string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if _, err := pool.Exec(ctx, schemaDDL); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

// Close releases the connection pool.
func (s *PostgresStore) Close() { s.pool.Close() }

// Save inserts one evaluation and returns it with its assigned id/created_at.
func (s *PostgresStore) Save(ev Evaluation) (Evaluation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	raw, err := json.Marshal(ev.Input)
	if err != nil {
		return Evaluation{}, err
	}
	row := s.pool.QueryRow(ctx,
		`INSERT INTO evaluations (input_json, effective, note) VALUES ($1, $2, $3) RETURNING id, created_at`,
		raw, ev.Effective, ev.Note)
	if err := row.Scan(&ev.ID, &ev.CreatedAt); err != nil {
		return Evaluation{}, err
	}
	return ev, nil
}

// List returns the most recent evaluations, newest first.
func (s *PostgresStore) List(limit int) ([]Evaluation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx,
		`SELECT id, created_at, input_json, effective, note FROM evaluations ORDER BY id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Evaluation
	for rows.Next() {
		var ev Evaluation
		var raw []byte
		if err := rows.Scan(&ev.ID, &ev.CreatedAt, &raw, &ev.Effective, &ev.Note); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &ev.Input); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}
