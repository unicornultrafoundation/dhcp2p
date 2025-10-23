# DHCP2P Development Guide

This guide provides comprehensive instructions for setting up a local development environment, understanding the project structure, and contributing to DHCP2P.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Local Setup](#local-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Database Management](#database-management)
- [Code Generation](#code-generation)
- [Testing](#testing)
- [Coding Standards](#coding-standards)
- [Makefile Commands](#makefile-commands)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Software

- **Go**: 1.25+ ([Download](https://golang.org/dl/))
- **Docker**: 20.10+ ([Download](https://www.docker.com/get-started))
- **Docker Compose**: 2.0+ (included with Docker Desktop)
- **Make**: 4.0+ (optional, for convenience commands)
- **Atlas CLI**: Latest ([Installation Guide](https://atlasgo.io/getting-started/installation))

### Optional Tools

- **sqlc**: For SQL code generation ([Installation](https://docs.sqlc.dev/en/latest/overview/install.html))
- **golangci-lint**: For code linting ([Installation](https://golangci-lint.run/usage/install/))
- **mockgen**: For mock generation (included with Go)

### IDE/Editor Setup

Recommended editors with Go support:
- **VS Code**: With Go extension
- **GoLand**: JetBrains IDE
- **Vim/Neovim**: With vim-go plugin
- **Emacs**: With go-mode

## Local Setup

### 1. Clone Repository

```bash
git clone https://github.com/duchuongnguyen/dhcp2p.git
cd dhcp2p
```

### 2. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install Atlas CLI
curl -sSf https://atlasgo.sh | sh

# Install sqlc (optional)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### 3. Environment Setup

```bash
# Interactive setup
make setup

# Or manually create environment file
cp .env.example .env
# Edit .env with your configuration
```

### 4. Start Development Services

```bash
# Start PostgreSQL and Redis
make docker-up

# Verify services are running
make docker-health
```

### 5. Run Database Migrations

```bash
# Apply migrations
make migrate

# Generate SQL code
make sqlc

# Or run both
make db
```

### 6. Run Application

```bash
# Run locally (requires PostgreSQL and Redis)
go run cmd/dhcp2p/main.go serve

# Or run with Docker
make docker-up
```

## Project Structure

```
dhcp2p/
├── cmd/                          # Application entry points
│   └── dhcp2p/
│       ├── cmd/                  # CLI commands
│       │   ├── root.go          # Root command
│       │   ├── serve.go         # Serve command
│       │   └── version.go       # Version command
│       └── main.go              # Main entry point
├── config/                      # Configuration files
│   └── config.yaml             # Default configuration
├── docs/                        # Documentation
├── docker/                      # Docker-related files
├── internal/                    # Private application code
│   └── app/
│       ├── adapters/            # External interface adapters
│       │   ├── auth/            # Authentication adapters
│       │   │   └── libp2p/      # libp2p signature verification
│       │   ├── handlers/        # HTTP handlers
│       │   │   └── http/        # HTTP-specific handlers
│       │   └── repositories/    # Data access adapters
│       │       ├── hybrid/      # Hybrid repository (Postgres + Redis)
│       │       ├── postgres/    # PostgreSQL implementation
│       │       └── redis/        # Redis implementation
│       ├── application/         # Application layer
│       │   ├── jobs/           # Background jobs
│       │   ├── services/       # Application services
│       │   └── utils/          # Application utilities
│       ├── domain/             # Domain layer
│       │   ├── errors/         # Domain errors
│       │   ├── models/         # Domain models
│       │   └── ports/          # Domain interfaces (ports)
│       ├── infrastructure/     # Infrastructure layer
│       │   ├── config/         # Configuration management
│       │   ├── flag/           # CLI flags
│       │   ├── logger/         # Logging
│       │   ├── migrations/     # Database migrations
│       │   └── server/         # HTTP server
│       └── app_module.go       # Application module
├── tests/                       # Test suites
│   ├── unit/                   # Unit tests
│   ├── integration/            # Integration tests
│   ├── e2e/                    # End-to-end tests
│   ├── benchmark/              # Benchmark tests
│   ├── contract/               # Contract tests
│   ├── fixtures/               # Test data
│   ├── helpers/                 # Test utilities
│   └── mocks/                  # Generated mocks
├── scripts/                     # Utility scripts
├── docker-compose.yml          # Docker Compose configuration
├── Dockerfile                  # Application Docker image
├── Dockerfile.migrate          # Migration Docker image
├── Makefile                    # Build and development commands
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
└── sqlc.yaml                   # SQL code generation config
```

### Layer Responsibilities

#### Domain Layer (`internal/app/domain/`)
- **Models**: Core business entities
- **Ports**: Interfaces for external dependencies
- **Errors**: Domain-specific error types

#### Application Layer (`internal/app/application/`)
- **Services**: Business logic and use cases
- **Jobs**: Background processes
- **Utils**: Application-specific utilities

#### Adapters Layer (`internal/app/adapters/`)
- **Handlers**: HTTP request handlers
- **Repositories**: Data access implementations
- **Auth**: Authentication implementations

#### Infrastructure Layer (`internal/app/infrastructure/`)
- **Config**: Configuration management
- **Server**: HTTP server setup
- **Logger**: Logging configuration
- **Migrations**: Database schema management

## Development Workflow

### 1. Feature Development

```bash
# Create feature branch
git checkout -b feature/new-feature

# Make changes
# ... edit code ...

# Run tests
make test

# Run linter
golangci-lint run

# Commit changes
git add .
git commit -m "feat: add new feature"

# Push branch
git push origin feature/new-feature
```

### 2. Database Changes

```bash
# Create migration
atlas migrate diff add_new_table \
  --to "file://internal/app/infrastructure/migrations/schema.hcl" \
  --dir "file://internal/app/infrastructure/migrations" \
  --dev-url "docker://postgres/15"

# Apply migration
make migrate

# Generate SQL code
make sqlc

# Update tests
# ... update test fixtures ...
```

### 3. Testing Changes

```bash
# Run specific test
go test -v ./tests/unit/application/services/lease_service_test.go

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration

# Run e2e tests
make test-e2e
```

### 4. Code Review Process

```bash
# Create pull request
# ... via GitHub UI ...

# Address review comments
# ... make changes ...

# Update PR
git add .
git commit -m "fix: address review comments"
git push origin feature/new-feature
```

## Database Management

### Migrations with Atlas

Atlas is used for database schema management:

```bash
# Create new migration
atlas migrate diff migration_name \
  --to "file://internal/app/infrastructure/migrations/schema.hcl" \
  --dir "file://internal/app/infrastructure/migrations" \
  --dev-url "docker://postgres/15"

# Apply migrations
atlas migrate apply \
  --dir "file://internal/app/infrastructure/migrations" \
  --url "$DATABASE_URL"

# Check migration status
atlas migrate status \
  --dir "file://internal/app/infrastructure/migrations" \
  --url "$DATABASE_URL"
```

### Schema Definition

Edit `internal/app/infrastructure/migrations/schema.hcl` to define schema changes:

```hcl
table "new_table" {
  schema = schema.public
  column "id" {
    type = serial
    null = false
  }
  column "name" {
    type = varchar(255)
    null = false
  }
  
  primary_key {
    columns = [column.id]
  }
}
```

### SQL Code Generation

sqlc generates type-safe Go code from SQL queries:

```bash
# Generate code
sqlc generate

# Or use Make
make sqlc
```

## Code Generation

### Mock Generation

Generate mocks for interfaces:

```bash
# Generate all mocks
make test-mocks

# Or manually
cd tests/mocks
go generate
```

### SQL Code Generation

Generate Go code from SQL queries:

```bash
# Generate SQL code
make sqlc
```

## Testing

### Test Categories

#### Unit Tests
- Test individual functions and methods
- Use mocks for external dependencies
- Fast execution, no external dependencies

```bash
# Run unit tests
make test-unit

# Run specific unit test
go test -v ./tests/unit/application/services/
```

#### Integration Tests
- Test with real database and Redis
- Use testcontainers for isolated environments
- Slower execution, requires Docker

```bash
# Run integration tests
make test-integration

# Run specific integration test
go test -v ./tests/integration/database/postgres/ -tags=integration
```

#### End-to-End Tests
- Test complete API workflows
- Test real HTTP requests and responses
- Comprehensive testing of user scenarios

```bash
# Run e2e tests
make test-e2e

# Run specific e2e test
go test -v ./tests/e2e/api/ -tags=e2e
```

### Test Helpers

The `tests/helpers/` package provides utilities:

- **Database**: Database setup and cleanup
- **HTTP Client**: HTTP testing utilities
- **Testcontainers**: Container management
- **Assertions**: Custom assertion helpers

### Writing Tests

#### Unit Test Example

```go
func TestLeaseService_AllocateIP(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockRepo := mocks.NewMockLeaseRepository(ctrl)
    service := services.NewLeaseService(&config.AppConfig{}, mockRepo, zap.NewNop())
    
    expectedLease := &models.Lease{
        TokenID: 12345,
        PeerID:  "peer123",
    }
    
    mockRepo.EXPECT().GetLeaseByPeerID(gomock.Any(), "peer123").Return(nil, nil)
    mockRepo.EXPECT().AllocateNewLease(gomock.Any(), "peer123").Return(expectedLease, nil)
    
    result, err := service.AllocateIP(context.Background(), "peer123")
    
    assert.NoError(t, err)
    assert.Equal(t, expectedLease, result)
}
```

#### Integration Test Example

```go
func TestLeaseRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    ctx := context.Background()
    
    // Start PostgreSQL container
    postgresContainer, err := postgres.RunContainer(ctx, ...)
    require.NoError(t, err)
    defer postgresContainer.Terminate(ctx)
    
    // Get connection string
    connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
    require.NoError(t, err)
    
    // Create repository
    repo, err := postgres.NewLeaseRepository(connStr)
    require.NoError(t, err)
    
    // Test functionality
    lease, err := repo.AllocateNewLease(ctx, "peer123")
    assert.NoError(t, err)
    assert.NotNil(t, lease)
    assert.Equal(t, "peer123", lease.PeerID)
}
```

## Coding Standards

### Go Code Style

Follow standard Go conventions:

- Use `gofmt` for formatting
- Use `golint` for linting
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use meaningful variable and function names

### Project-Specific Conventions

#### Error Handling

```go
// Use domain errors
if err != nil {
    return nil, errors.ErrLeaseNotFound
}

// Wrap errors with context
if err != nil {
    return nil, fmt.Errorf("failed to allocate lease: %w", err)
}
```

#### Dependency Injection

```go
// Use constructor functions
func NewLeaseService(config *config.AppConfig, repo ports.LeaseRepository, logger *zap.Logger) *LeaseService {
    return &LeaseService{
        repo:   repo,
        logger: logger,
        config: config,
    }
}
```

#### Interface Design

```go
// Keep interfaces small and focused
type LeaseRepository interface {
    AllocateNewLease(ctx context.Context, peerID string) (*models.Lease, error)
    GetLeaseByPeerID(ctx context.Context, peerID string) (*models.Lease, error)
}
```

### Code Organization

- **One package per file**: Keep related functionality together
- **Clear separation**: Domain, application, adapters, infrastructure
- **Interface segregation**: Small, focused interfaces
- **Dependency inversion**: Depend on abstractions, not concretions

## Makefile Commands

### Development Commands

```bash
# Setup and start
make setup              # Interactive project setup
make docker-up          # Start development stack
make docker-down        # Stop development stack
make docker-logs        # View application logs
make docker-health      # Check application health
```

### Database Commands

```bash
# Migrations
make migrate            # Apply migrations
make sqlc              # Generate SQL code
make db                # Run migrations + generate code

# Manual migration commands
make hash              # Generate Atlas migration hash
make diff name=NAME     # Create Atlas migration diff
```

### Testing Commands

```bash
# Test execution
make test              # Run all tests
make test-unit         # Run unit tests only
make test-integration  # Run integration tests only
make test-e2e          # Run end-to-end tests only
make test-coverage     # Generate coverage report

# Test utilities
make test-mocks        # Generate mocks for testing
```

### Building Commands

```bash
# Docker builds
make docker-build      # Build application image
make docker-build-migrate # Build migration image
make docker-push       # Push image to registry

# Production
make docker-up-prod    # Start production stack
make docker-down-prod  # Stop production stack
```

### Utility Commands

```bash
# Help
make help              # Show all available commands

# Status and debugging
make docker-ps         # Show running containers
make migrate-status     # Show migration status
make migrate-docker     # Run migrations in container
```

## Troubleshooting

### Common Development Issues

#### 1. Go Module Issues

```bash
# Clean module cache
go clean -modcache

# Download dependencies
go mod download

# Tidy dependencies
go mod tidy
```

#### 2. Database Connection Issues

```bash
# Check database status
make docker-ps

# Check database logs
docker-compose logs postgres

# Test database connection
docker-compose exec postgres psql -U dhcp2p -c "SELECT 1;"
```

#### 3. Redis Connection Issues

```bash
# Check Redis status
docker-compose logs redis

# Test Redis connection
docker-compose exec redis redis-cli ping
```

#### 4. Migration Issues

```bash
# Check migration status
make migrate-status

# Reset migrations (development only)
docker-compose exec postgres psql -U dhcp2p -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
make migrate
```

#### 5. Test Failures

```bash
# Run tests with verbose output
go test -v ./tests/unit/...

# Run tests with race detection
go test -race ./tests/unit/...

# Check test coverage
make test-coverage
```

### Debug Mode

```bash
# Run with debug logging
LOG_LEVEL=debug go run cmd/dhcp2p/main.go serve

# Run with race detection
go run -race cmd/dhcp2p/main.go serve
```

### Performance Profiling

```bash
# CPU profiling
go run cmd/dhcp2p/main.go serve &
curl http://localhost:8088/health
go tool pprof http://localhost:8088/debug/pprof/profile

# Memory profiling
go tool pprof http://localhost:8088/debug/pprof/heap
```

This development guide ensures developers can effectively contribute to DHCP2P with proper understanding of the codebase, testing practices, and development workflow.
