package helpers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ContainerPool manages a pool of reusable test containers
type ContainerPool struct {
	mu            sync.RWMutex
	postgresPool  chan testcontainers.Container
	redisPool     chan testcontainers.Container
	maxPoolSize   int
	createdCount  int
	logger        interface{} // Use interface to avoid import cycles
}

// NewContainerPool creates a new container pool
func NewContainerPool(maxPoolSize int) *ContainerPool {
	return &ContainerPool{
		postgresPool: make(chan testcontainers.Container, maxPoolSize),
		redisPool:    make(chan testcontainers.Container, maxPoolSize),
		maxPoolSize:  maxPoolSize,
	}
}

// GetPostgresContainer gets a PostgreSQL container from the pool or creates a new one
func (p *ContainerPool) GetPostgresContainer(ctx context.Context) (testcontainers.Container, string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case container := <-p.postgresPool:
		// Return existing container
		connStr, err := p.getPostgresConnectionString(ctx, container)
		if err != nil {
			container.Terminate(ctx)
			return p.createPostgresContainer(ctx)
		}
		return container, connStr, nil
	default:
		// Create new container if pool is empty
		return p.createPostgresContainer(ctx)
	}
}

// GetRedisContainer gets a Redis container from the pool or creates a new one
func (p *ContainerPool) GetRedisContainer(ctx context.Context) (testcontainers.Container, string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case container := <-p.redisPool:
		// Return existing container
		connStr, err := p.getRedisConnectionString(ctx, container)
		if err != nil {
			container.Terminate(ctx)
			return p.createRedisContainer(ctx)
		}
		return container, connStr, nil
	default:
		// Create new container if pool is empty
		return p.createRedisContainer(ctx)
	}
}

// ReturnPostgresContainer returns a PostgreSQL container to the pool
func (p *ContainerPool) ReturnPostgresContainer(container testcontainers.Container) {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case p.postgresPool <- container:
		// Container returned to pool
	default:
		// Pool is full, terminate the container
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			container.Terminate(ctx)
		}()
	}
}

// ReturnRedisContainer returns a Redis container to the pool
func (p *ContainerPool) ReturnRedisContainer(container testcontainers.Container) {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case p.redisPool <- container:
		// Container returned to pool
	default:
		// Pool is full, terminate the container
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			container.Terminate(ctx)
		}()
	}
}

// createPostgresContainer creates a new PostgreSQL container
func (p *ContainerPool) createPostgresContainer(ctx context.Context) (testcontainers.Container, string, error) {
	if p.createdCount >= p.maxPoolSize {
		return nil, "", fmt.Errorf("maximum pool size reached")
	}

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("dhcp2p_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create postgres container: %w", err)
	}

	p.createdCount++
	connStr, err := p.getPostgresConnectionString(ctx, container)
	if err != nil {
		container.Terminate(ctx)
		return nil, "", err
	}

	return container, connStr, nil
}

// createRedisContainer creates a new Redis container
func (p *ContainerPool) createRedisContainer(ctx context.Context) (testcontainers.Container, string, error) {
	if p.createdCount >= p.maxPoolSize {
		return nil, "", fmt.Errorf("maximum pool size reached")
	}

	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForLog("Ready to accept connections").
			WithOccurrence(1).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create redis container: %w", err)
	}

	// Give the container additional time to initialize
	time.Sleep(100 * time.Millisecond)

	p.createdCount++
	connStr, err := p.getRedisConnectionString(ctx, container)
	if err != nil {
		container.Terminate(ctx)
		return nil, "", err
	}

	return container, connStr, nil
}

// getPostgresConnectionString returns the PostgreSQL connection string
func (p *ContainerPool) getPostgresConnectionString(ctx context.Context, container testcontainers.Container) (string, error) {
	// Try to cast to postgres container to get connection string
	if pgContainer, ok := container.(*postgres.PostgresContainer); ok {
		return pgContainer.ConnectionString(ctx, "sslmode=disable")
	}
	
	// Fallback: try to get host and port manually
	host, err := container.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get postgres host: %w", err)
	}
	
	// Retry getting the port mapping with exponential backoff
	var port nat.Port
	for i := 0; i < 10; i++ {
		port, err = container.MappedPort(ctx, "5432")
		if err == nil {
			break
		}
		if i == 9 {
			return "", fmt.Errorf("failed to get postgres port after retries: %w", err)
		}
		time.Sleep(time.Duration(200*(i+1)) * time.Millisecond)
	}
	
	return fmt.Sprintf("postgres://test:test@%s:%s/dhcp2p_test?sslmode=disable", host, port.Port()), nil
}

// getRedisConnectionString returns the Redis connection string
func (p *ContainerPool) getRedisConnectionString(ctx context.Context, container testcontainers.Container) (string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get redis host: %w", err)
	}

	// Retry getting the port mapping with exponential backoff
	var port nat.Port
	for i := 0; i < 10; i++ {
		port, err = container.MappedPort(ctx, "6379")
		if err == nil {
			break
		}
		if i == 9 {
			return "", fmt.Errorf("failed to get redis port after retries: %w", err)
		}
		time.Sleep(time.Duration(200*(i+1)) * time.Millisecond)
	}

	return fmt.Sprintf("%s:%s", host, port.Port()), nil
}

// Cleanup terminates all containers in the pool
func (p *ContainerPool) Cleanup(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errors []error

	// Terminate all PostgreSQL containers
	close(p.postgresPool)
	for container := range p.postgresPool {
		if err := container.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to terminate postgres container: %w", err))
		}
	}

	// Terminate all Redis containers
	close(p.redisPool)
	for container := range p.redisPool {
		if err := container.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to terminate redis container: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}

// Global container pool instance
var globalPool *ContainerPool
var poolOnce sync.Once

// GetGlobalPool returns the global container pool (singleton pattern)
func GetGlobalPool() *ContainerPool {
	poolOnce.Do(func() {
		globalPool = NewContainerPool(3) // Max 3 containers of each type
	})
	return globalPool
}
