package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DatabaseHelper provides utilities for database testing
type DatabaseHelper struct {
	DB *sql.DB
}

// NewDatabaseHelper creates a new database helper
func NewDatabaseHelper(connStr string) (*DatabaseHelper, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseHelper{DB: db}, nil
}

// Close closes the database connection
func (h *DatabaseHelper) Close() error {
	return h.DB.Close()
}

// CleanupTables removes all data from test tables
func (h *DatabaseHelper) CleanupTables(ctx context.Context) error {
	tables := []string{"leases", "nonces", "alloc_state"}

	for _, table := range tables {
		if _, err := h.DB.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			return fmt.Errorf("failed to cleanup table %s: %w", table, err)
		}
	}

	// Reset alloc_state to initial value
	if _, err := h.DB.ExecContext(ctx, "INSERT INTO alloc_state (id, last_token_id) VALUES (1, 184418304) ON CONFLICT (id) DO UPDATE SET last_token_id = 184418304"); err != nil {
		return fmt.Errorf("failed to reset alloc_state: %w", err)
	}

	return nil
}

// InsertTestLease inserts a test lease
func (h *DatabaseHelper) InsertTestLease(ctx context.Context, tokenID int64, peerID string, expiresAt time.Time) error {
	query := `INSERT INTO leases (token_id, peer_id, expires_at, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5)`

	now := time.Now()
	_, err := h.DB.ExecContext(ctx, query, tokenID, peerID, expiresAt, now, now)
	if err != nil {
		return fmt.Errorf("failed to insert test lease: %w", err)
	}

	return nil
}

// InsertTestNonce inserts a test nonce
func (h *DatabaseHelper) InsertTestNonce(ctx context.Context, id, peerID string, expiresAt time.Time, used bool) error {
	query := `INSERT INTO nonces (id, peer_id, issued_at, expires_at, used, used_at) 
			  VALUES ($1, $2, $3, $4, $5, $6)`

	now := time.Now()
	var usedAt *time.Time
	if used {
		usedAt = &now
	}

	_, err := h.DB.ExecContext(ctx, query, id, peerID, now, expiresAt, used, usedAt)
	if err != nil {
		return fmt.Errorf("failed to insert test nonce: %w", err)
	}

	return nil
}

// GetLeaseCount returns the number of leases in the database
func (h *DatabaseHelper) GetLeaseCount(ctx context.Context) (int, error) {
	var count int
	err := h.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM leases").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get lease count: %w", err)
	}
	return count, nil
}

// GetNonceCount returns the number of nonces in the database
func (h *DatabaseHelper) GetNonceCount(ctx context.Context) (int, error) {
	var count int
	err := h.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM nonces").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce count: %w", err)
	}
	return count, nil
}
