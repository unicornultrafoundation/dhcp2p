# Testing Guide for DHCP2P

This document provides a comprehensive guide to the testing setup and implementation for the DHCP2P project.

## Table of Contents

- [Test Structure](#test-structure)
- [Running Tests](#running-tests)
- [Test Configuration](#test-configuration)
- [Mock Generation](#mock-generation)
- [CI/CD Integration](#cicd-integration)
- [Test Helpers](#test-helpers)
- [Writing Tests](#writing-tests)
- [Benchmark Tests](#benchmark-tests)
- [Contract Tests](#contract-tests)
- [Test Fixtures](#test-fixtures)
- [Best Practices](#best-practices)
- [Coverage Goals](#coverage-goals)
- [Troubleshooting](#troubleshooting)

## Test Structure

The tests are organized into five main categories:

### 1. Unit Tests (`tests/unit/`)
- **Domain Models**: Test business logic and validation
- **Application Services**: Test business logic with mocked dependencies
- **HTTP Handlers**: Test API endpoints with mocked services
- **Utilities**: Test pure functions
- **Repositories**: Test repository interfaces with mocks

### 2. Integration Tests (`tests/integration/`)
- **Database Integration**: Test with real PostgreSQL using testcontainers
- **Redis Integration**: Test caching and nonce management
- **Service Integration**: Test service interactions with real repositories
- **Cache Integration**: Test Redis caching behavior

### 3. End-to-End Tests (`tests/e2e/`)
- **API Workflows**: Complete lease allocation/renewal/release flows
- **Error Handling**: Test error responses and edge cases
- **Authentication Flow**: Test complete authentication workflow

### 4. Benchmark Tests (`tests/benchmark/`)
- **Performance Testing**: Measure performance of critical operations
- **Load Testing**: Test system under various loads
- **Memory Profiling**: Identify memory usage patterns

### 5. Contract Tests (`tests/contract/`)
- **API Contracts**: Validate API contract compliance
- **Schema Validation**: Test request/response schemas
- **Version Compatibility**: Test backward compatibility

## Running Tests

### Prerequisites
- Go 1.25+
- Docker (for integration and e2e tests)
- Make

### Commands

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Run only e2e tests
make test-e2e

# Generate coverage report
make test-coverage

# Generate mocks
make test-mocks
```

### Individual Test Categories

```bash
# Unit tests
go test -v ./tests/unit/...

# Integration tests (requires Docker)
go test -v ./tests/integration/... -tags=integration

# E2E tests (requires Docker)
go test -v ./tests/e2e/... -tags=e2e
```

## Test Configuration

### Environment Variables
- `DB_URL`: PostgreSQL connection string for integration tests
- `REDIS_URL`: Redis connection string for integration tests

### Test Data
Test fixtures are located in `tests/fixtures/`:
- `test_config.yaml`: Test configuration
- `leases.json`: Sample lease data
- `nonces.json`: Sample nonce data

## Mock Generation

Mocks are automatically generated using `go generate`:

```bash
make test-mocks
```

This generates mocks for all interfaces in the `domain/ports` package.

## CI/CD Integration

The project includes GitHub Actions workflow (`.github/workflows/test.yml`) that:
- Runs all test categories
- Generates coverage reports
- Uploads coverage to Codecov

## Test Helpers

The `tests/helpers/` package provides utilities for:
- **Testcontainers**: Start PostgreSQL and Redis containers
- **Database**: Database setup and cleanup
- **HTTP Client**: HTTP testing utilities

## Writing Tests

### Unit Test Example

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

### Integration Test Example

```go
func TestLeaseRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    ctx := context.Background()
    
    // Start PostgreSQL container
    postgresContainer, err := postgres.RunContainer(ctx, 
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("testuser"),
        postgres.WithPassword("testpass"),
    )
    require.NoError(t, err)
    defer postgresContainer.Terminate(ctx)

    // Get connection string
    connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
    require.NoError(t, err)

    // Test with real database
    repo, err := postgres.NewLeaseRepository(connStr)
    require.NoError(t, err)

    lease, err := repo.AllocateNewLease(ctx, "peer123")
    assert.NoError(t, err)
    assert.NotNil(t, lease)
    assert.Equal(t, "peer123", lease.PeerID)
}
```

### E2E Test Example

```go
func TestLeaseWorkflow_E2E(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping e2e test")
    }

    // Start test environment
    env := testhelpers.StartTestEnvironment(t)
    defer env.Cleanup()

    client := testhelpers.NewHTTPClient(env.BaseURL)

    // Test complete lease workflow
    t.Run("Complete Lease Workflow", func(t *testing.T) {
        // 1. Request authentication
        authResp, err := client.RequestAuth("test-peer-id")
        require.NoError(t, err)
        assert.NotEmpty(t, authResp.Nonce)

        // 2. Allocate lease
        lease, err := client.AllocateLease("test-peer-id", authResp.Signature)
        require.NoError(t, err)
        assert.NotNil(t, lease)
        assert.Equal(t, "test-peer-id", lease.PeerID)

        // 3. Renew lease
        renewedLease, err := client.RenewLease(lease.TokenID, "test-peer-id", authResp.Signature)
        require.NoError(t, err)
        assert.NotNil(t, renewedLease)
        assert.True(t, renewedLease.ExpiresAt.After(lease.ExpiresAt))

        // 4. Release lease
        err = client.ReleaseLease(lease.TokenID, "test-peer-id", authResp.Signature)
        require.NoError(t, err)
    })
}
```

## Benchmark Tests

### Performance Benchmark Example

```go
func BenchmarkLeaseService_AllocateIP(b *testing.B) {
    ctrl := gomock.NewController(b)
    defer ctrl.Finish()
    
    mockRepo := mocks.NewMockLeaseRepository(ctrl)
    service := services.NewLeaseService(&config.AppConfig{}, mockRepo, zap.NewNop())

    expectedLease := &models.Lease{
        TokenID: 12345,
        PeerID:  "peer123",
    }

    mockRepo.EXPECT().GetLeaseByPeerID(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
    mockRepo.EXPECT().AllocateNewLease(gomock.Any(), gomock.Any()).Return(expectedLease, nil).AnyTimes()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.AllocateIP(context.Background(), "peer123")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Memory Benchmark Example

```go
func BenchmarkLeaseRepository_MemoryUsage(b *testing.B) {
    ctx := context.Background()
    
    // Start PostgreSQL container
    postgresContainer, err := postgres.RunContainer(ctx)
    require.NoError(b, err)
    defer postgresContainer.Terminate(ctx)

    connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
    require.NoError(b, err)

    repo, err := postgres.NewLeaseRepository(connStr)
    require.NoError(b, err)

    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        lease, err := repo.AllocateNewLease(ctx, fmt.Sprintf("peer-%d", i))
        if err != nil {
            b.Fatal(err)
        }
        _ = lease // Prevent optimization
    }
}
```

## Contract Tests

### API Contract Test Example

```go
func TestLeaseAPI_Contract(t *testing.T) {
    env := testhelpers.StartTestEnvironment(t)
    defer env.Cleanup()

    client := testhelpers.NewHTTPClient(env.BaseURL)

    tests := []struct {
        name           string
        endpoint       string
        method         string
        requestBody    interface{}
        expectedStatus int
        expectedSchema string
    }{
        {
            name:           "Request Auth",
            endpoint:       "/request-auth",
            method:         "POST",
            requestBody:    map[string]string{"pubkey": "test-key"},
            expectedStatus: 200,
            expectedSchema: "auth-response-schema.json",
        },
        {
            name:           "Allocate IP",
            endpoint:       "/allocate-ip",
            method:         "POST",
            requestBody:    map[string]string{"peer_id": "test-peer"},
            expectedStatus: 200,
            expectedSchema: "lease-schema.json",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp, err := client.MakeRequest(tt.method, tt.endpoint, tt.requestBody)
            require.NoError(t, err)
            assert.Equal(t, tt.expectedStatus, resp.StatusCode)

            // Validate response schema
            err = client.ValidateSchema(resp.Body, tt.expectedSchema)
            assert.NoError(t, err)
        })
    }
}
```

## Test Fixtures

### Fixture Usage Example

```go
func TestLeaseService_WithFixtures(t *testing.T) {
    // Load test fixtures
    leases := fixtures.LoadLeases(t)
    nonces := fixtures.LoadNonces(t)

    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockRepo := mocks.NewMockLeaseRepository(ctrl)
    service := services.NewLeaseService(&config.AppConfig{}, mockRepo, zap.NewNop())

    // Use fixture data in tests
    for _, lease := range leases {
        t.Run(fmt.Sprintf("Lease_%d", lease.TokenID), func(t *testing.T) {
            mockRepo.EXPECT().GetLeaseByTokenID(gomock.Any(), lease.TokenID).Return(&lease, nil)
            
            result, err := service.GetLeaseByTokenID(context.Background(), lease.TokenID)
            assert.NoError(t, err)
            assert.Equal(t, &lease, result)
        })
    }
}
```

### Fixture Builder Example

```go
func TestLeaseService_WithBuilders(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockRepo := mocks.NewMockLeaseRepository(ctrl)
    service := services.NewLeaseService(&config.AppConfig{}, mockRepo, zap.NewNop())

    // Use builders for test data
    lease := fixtures.NewLeaseBuilder().
        WithTokenID(12345).
        WithPeerID("test-peer").
        WithTTL(120).
        Build()

    mockRepo.EXPECT().AllocateNewLease(gomock.Any(), "test-peer").Return(lease, nil)

    result, err := service.AllocateIP(context.Background(), "test-peer")
    assert.NoError(t, err)
    assert.Equal(t, lease, result)
}
```

## Best Practices

### Test Organization
1. **Use Table-Driven Tests**: For testing multiple scenarios
2. **Mock External Dependencies**: Use mocks for unit tests
3. **Test Edge Cases**: Include error conditions and boundary cases
4. **Use Testcontainers**: For integration tests with real databases
5. **Clean Up Resources**: Always defer cleanup in integration tests
6. **Skip Long Tests**: Use `testing.Short()` for integration/e2e tests

### Test Data Management
1. **Use Fixtures**: Load test data from JSON files
2. **Use Builders**: Create test data with builder pattern
3. **Isolate Tests**: Each test should be independent
4. **Clean State**: Reset state between tests

### Performance Testing
1. **Benchmark Critical Paths**: Focus on performance-critical operations
2. **Memory Profiling**: Use `b.ReportAllocs()` for memory benchmarks
3. **Load Testing**: Test under various load conditions
4. **Profile Analysis**: Use `go tool pprof` for detailed analysis

### Test Quality
1. **Clear Test Names**: Use descriptive test function names
2. **Arrange-Act-Assert**: Structure tests clearly
3. **Single Responsibility**: Each test should test one thing
4. **Readable Assertions**: Use clear assertion messages

## Coverage Goals

### Coverage Targets
- **Unit Tests**: 80%+ coverage for business logic
- **Integration Tests**: Critical paths covered (database, Redis)
- **E2E Tests**: Main user workflows covered
- **Contract Tests**: All API endpoints covered

### Coverage Commands
```bash
# Generate coverage report
make test-coverage

# View coverage in browser
go tool cover -html=coverage.out

# Check coverage threshold
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total
```

## Troubleshooting

### Common Issues

#### 1. Docker Issues
```bash
# Docker not running
docker ps
# Start Docker Desktop or Docker daemon

# Port conflicts
netstat -tulpn | grep :5432
netstat -tulpn | grep :6379
# Stop conflicting services or change ports
```

#### 2. Test Environment Issues
```bash
# Mock generation
make test-mocks

# Module issues
go mod tidy
go mod download

# Clean test cache
go clean -testcache
```

#### 3. Database Issues
```bash
# Reset test database
docker-compose down -v
docker-compose up -d postgres redis

# Check database connectivity
docker-compose exec postgres pg_isready -U dhcp2p
```

### Debug Mode

#### Verbose Testing
```bash
# Run tests with verbose output
go test -v ./tests/unit/...

# Run tests with race detection
go test -race ./tests/unit/...

# Run specific test
go test -v -run TestLeaseService_AllocateIP ./tests/unit/application/services/
```

#### Test Debugging
```bash
# Run single test with debug output
go test -v -run TestSpecificTest ./tests/unit/...

# Run tests with timeout
go test -timeout 30s ./tests/integration/...

# Run tests with build tags
go test -tags=integration ./tests/integration/...
```

#### Performance Debugging
```bash
# Run benchmarks
go test -bench=. ./tests/benchmark/...

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./tests/benchmark/...

# Profile CPU usage
go test -bench=. -cpuprofile=cpu.prof ./tests/benchmark/...
go tool pprof cpu.prof
```

### Test Environment Debugging

#### Container Debugging
```bash
# Check container logs
docker-compose logs postgres
docker-compose logs redis

# Access container shell
docker-compose exec postgres psql -U dhcp2p dhcp2p
docker-compose exec redis redis-cli

# Check container status
docker-compose ps
```

#### Network Debugging
```bash
# Test connectivity
docker-compose exec dhcp2p ping postgres
docker-compose exec dhcp2p ping redis

# Check port accessibility
telnet localhost 5432
telnet localhost 6379
```

### Test Data Debugging

#### Fixture Issues
```bash
# Validate JSON fixtures
cat tests/fixtures/leases.json | jq .

# Check fixture loading
go test -v -run TestFixtureLoading ./tests/unit/...
```

#### Mock Issues
```bash
# Regenerate mocks
make test-mocks

# Check mock usage
go test -v -run TestWithMocks ./tests/unit/...
```

This enhanced testing guide provides comprehensive coverage of all testing aspects in DHCP2P, including examples, best practices, and troubleshooting guidance.
