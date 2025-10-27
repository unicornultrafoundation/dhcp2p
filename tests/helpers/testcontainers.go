package helpers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestStack represents a complete test environment
type TestStack struct {
	PostgresContainer testcontainers.Container
	RedisContainer    testcontainers.Container
	PostgresConnStr   string
	RedisConnStr      string
}

// StartTestStack starts the complete test environment
func StartTestStack(ctx context.Context) (*TestStack, error) {
	// Start PostgreSQL
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("dhcp2p_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	// Start Redis
	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForLog("Ready to accept connections").
			WithStartupTimeout(30 * time.Second),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	if err != nil {
		postgresContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to start redis container: %w", err)
	}

	// Get connection strings
	postgresConnStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		postgresContainer.Terminate(ctx)
		redisContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get postgres connection string: %w", err)
	}

	redisHost, err := redisContainer.Host(ctx)
	if err != nil {
		postgresContainer.Terminate(ctx)
		redisContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get redis host: %w", err)
	}

	// Retry getting the port mapping with exponential backoff
	var redisPort nat.Port
	for i := 0; i < 15; i++ {
		redisPort, err = redisContainer.MappedPort(ctx, "6379")
		if err == nil {
			break
		}
		if i < 14 { // Don't sleep on the last attempt
			time.Sleep(time.Duration(300*(i+1)) * time.Millisecond)
		}
	}
	if err != nil {
		postgresContainer.Terminate(ctx)
		redisContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get redis port after retries: %w", err)
	}
	redisConnStr := fmt.Sprintf("%s:%s", redisHost, redisPort.Port())

	return &TestStack{
		PostgresContainer: postgresContainer,
		RedisContainer:    redisContainer,
		PostgresConnStr:   postgresConnStr,
		RedisConnStr:      redisConnStr,
	}, nil
}

// Terminate stops all containers
func (s *TestStack) Terminate(ctx context.Context) error {
	var errs []error

	if err := s.PostgresContainer.Terminate(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to terminate postgres: %w", err))
	}

	if err := s.RedisContainer.Terminate(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to terminate redis: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors terminating containers: %v", errs)
	}

	return nil
}

// RunMigrations runs database migrations
func RunMigrations(connStr string) error {
	// This would typically run your migration tool
	// For now, we'll create a simple implementation
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Create tables (simplified version)
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS alloc_state (
			id serial PRIMARY KEY,
			last_token_id bigint NOT NULL,
			max_token_id bigint NOT NULL DEFAULT 168162304
		)`,
		`CREATE TABLE IF NOT EXISTS leases (
			token_id bigint PRIMARY KEY,
			peer_id varchar(128) NOT NULL,
			expires_at timestamptz NOT NULL,
			created_at timestamptz NOT NULL DEFAULT now(),
			updated_at timestamptz NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS nonces (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			peer_id varchar(128) NOT NULL,
			issued_at timestamptz NOT NULL,
			expires_at timestamptz NOT NULL,
			used boolean NOT NULL DEFAULT false,
			used_at timestamptz NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_leases_expires_at ON leases (expires_at)`,
		`INSERT INTO alloc_state (id, last_token_id, max_token_id) VALUES (1, 167902209, 168162304) ON CONFLICT (id) DO NOTHING`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	return nil
}

// WaitForServices waits for all services to be ready
func WaitForServices(stack *TestStack, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Test PostgreSQL connection
	db, err := sql.Open("pgx", stack.PostgresConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer db.Close()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for services to be ready")
		default:
			if err := db.PingContext(ctx); err == nil {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// StartRedisContainer starts a Redis container for testing
func StartRedisContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForLog("Ready to accept connections").
			WithStartupTimeout(30 * time.Second),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start redis container: %w", err)
	}

	// Verify the container is ready and port mapping is available
	for i := 0; i < 15; i++ {
		_, err := redisContainer.MappedPort(ctx, "6379")
		if err == nil {
			break
		}
		if i == 14 {
			return nil, fmt.Errorf("failed to get redis port after retries: %w", err)
		}
		time.Sleep(time.Duration(300*(i+1)) * time.Millisecond)
	}

	return redisContainer, nil
}

// NewRedisClient creates a new Redis client for testing
func NewRedisClient(t *testing.T, host, port string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: "", // no password for test Redis
		DB:       0,  // default DB
	})
}

// NewRedisClientFromConnStr creates a new Redis client from connection string
func NewRedisClientFromConnStr(connStr string) (*redis.Client, error) {
	return redis.NewClient(&redis.Options{
		Addr: connStr,
	}), nil
}

// LoadTestData loads test data from JSON files
func LoadTestData(filename string, v interface{}) error {
	path := filepath.Join("tests", "fixtures", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read test data file %s: %w", filename, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal test data from %s: %w", filename, err)
	}

	return nil
}
