# DHCP2P Architecture Guide

This document provides a comprehensive overview of the DHCP2P system architecture, design patterns, and implementation details.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Clean Architecture Implementation](#clean-architecture-implementation)
- [Layer Breakdown](#layer-breakdown)
- [Dependency Injection](#dependency-injection)
- [Data Flow](#data-flow)
- [Database Schema](#database-schema)
- [Authentication Flow](#authentication-flow)
- [Lease Allocation Algorithm](#lease-allocation-algorithm)
- [Nonce Management](#nonce-management)
- [Error Handling](#error-handling)
- [Performance Considerations](#performance-considerations)

## Architecture Overview

DHCP2P follows Clean Architecture (Hexagonal Architecture) principles, ensuring separation of concerns, testability, and maintainability.

```
┌─────────────────────────────────────────────────────────────────┐
│                         External World                          │
│  HTTP Clients     │        PostgreSQL         │     Redis       │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Infrastructure Layer                         │
│  HTTP Server  │  Database Adapters  │  Redis Adapters  │ Auth   │ 
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Adapters Layer                            │
│  HTTP Handlers  │  Repository Implementations  │  Auth Adapters │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Application Layer                            │
│  Services  │  Use Cases  │  Domain Services  │  Background Jobs │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Domain Layer                              │
│  Models  │  Ports (Interfaces)  │  Errors  │   Business Rules   │
└─────────────────────────────────────────────────────────────────┘
```

## Clean Architecture Implementation

### Core Principles

1. **Dependency Inversion**: High-level modules don't depend on low-level modules
2. **Interface Segregation**: Small, focused interfaces
3. **Single Responsibility**: Each layer has a single, well-defined purpose
4. **Testability**: Easy to test with mocks and stubs

### Layer Responsibilities

| Layer | Responsibility | Dependencies |
|-------|---------------|--------------|
| Domain | Business rules, entities, interfaces | None |
| Application | Use cases, services, orchestration | Domain only |
| Adapters | External interfaces, implementations | Application + Domain |
| Infrastructure | External systems, configuration | Adapters + Application + Domain |

## Layer Breakdown

### Domain Layer (`internal/app/domain/`)

The innermost layer containing business logic and rules.

#### Models (`models/`)
- **Lease**: Core lease entity with token ID, peer ID, timestamps, and TTL
- **AuthRequest/AuthResponse**: Authentication request/response models
- **Nonce**: Nonce entity for authentication

#### Ports (`ports/`)
Interfaces defining contracts for external dependencies:

```go
type LeaseRepository interface {
    AllocateNewLease(ctx context.Context, peerID string) (*models.Lease, error)
    GetLeaseByPeerID(ctx context.Context, peerID string) (*models.Lease, error)
    GetLeaseByTokenID(ctx context.Context, tokenID int64) (*models.Lease, error)
    RenewLease(ctx context.Context, tokenID int64, peerID string) (*models.Lease, error)
    ReleaseLease(ctx context.Context, tokenID int64, peerID string) error
    FindAndReuseExpiredLease(ctx context.Context, peerID string) (*models.Lease, error)
}

type AuthService interface {
    RequestAuth(ctx context.Context, req *models.AuthRequest) (*models.Nonce, error)
    VerifyAuth(ctx context.Context, req *models.AuthVerifyRequest) (*models.AuthVerifyResponse, error)
}

type SignatureVerifier interface {
    VerifySignature(ctx context.Context, publicKey []byte, payload []byte, signature []byte) error
}
```

#### Errors (`errors/`)
Domain-specific error types:

```go
var (
    ErrInvalidSignature = errors.New("invalid signature")
    ErrNonceExpired     = errors.New("nonce expired")
    ErrNonceNotFound    = errors.New("nonce not found")
    ErrLeaseNotFound    = errors.New("lease not found")
    ErrLeaseExpired     = errors.New("lease expired")
)
```

### Application Layer (`internal/app/application/`)

Contains use cases and application services.

#### Services (`services/`)
- **LeaseService**: Core lease management logic
- **AuthService**: Authentication orchestration
- **NonceService**: Nonce management

#### Jobs (`jobs/`)
Background processes:
- **NonceCleaner**: Removes expired nonces

### Adapters Layer (`internal/app/adapters/`)

Implements external interfaces and adapts external systems.

#### Handlers (`handlers/http/`)
HTTP request handlers with validation and error handling:

```go
type LeaseHandler struct {
    leaseService ports.LeaseService
}

func (h *LeaseHandler) AllocateIP(w http.ResponseWriter, r *http.Request) {
    sc := &ServiceCall{Handler: w, Request: r}
    sc.ExecuteWithValidation(
        h.handleAllocateIP,
        ValidateLeaseRequest,
    )
}
```

#### Repositories (`repositories/`)
- **PostgreSQL**: Primary data persistence
- **Redis**: Caching and nonce storage
- **Hybrid**: Combines PostgreSQL and Redis

#### Auth (`auth/libp2p/`)
libp2p signature verification implementation.

### Infrastructure Layer (`internal/app/infrastructure/`)

External system integrations and configuration.

#### Components
- **Config**: Configuration management with Viper
- **Server**: HTTP server setup with Chi router
- **Logger**: Structured logging with Zap
- **Migrations**: Database schema management with Atlas

## Dependency Injection

Uses Uber Fx for dependency injection and lifecycle management.

### Module Structure

```go
func NewApp() *fx.App {
    return fx.New(
        fx.NopLogger, // Disable Fx logging
        
        // Add modules
        adapters.Module,
        application.Module,
        infrastructure.Module,
        
        // Invoke servers and jobs
        fx.Invoke(func(server *server.HTTPServer) {}),
        fx.Invoke(func(nonceCleaner ports.NonceCleaner) {}),
    )
}
```

### Module Registration

Each layer provides a module that registers its dependencies:

```go
var Module = fx.Module("adapters",
    fx.Provide(
        http.NewHTTPRouter,
        http.NewAuthHandler,
        http.NewLeaseHandler,
        postgres.NewLeaseRepository,
        redis.NewNonceRepository,
    ),
)
```

## Data Flow

### Request Processing Flow

```
HTTP Request
    ↓
Router (Chi)
    ↓
Middleware (Security, Auth, Logging)
    ↓
Handler (Validation)
    ↓
Service (Business Logic)
    ↓
Repository (Data Access)
    ↓
Database/Redis
```

### Authentication Flow

```
1. Client → POST /request-auth (pubkey)
    ↓
2. AuthService → Generate nonce
    ↓
3. Store nonce in Redis (TTL: 5 minutes)
    ↓
4. Return nonce to client
    ↓
5. Client signs nonce with private key
    ↓
6. Client → Protected endpoint (signature in header)
    ↓
7. AuthMiddleware → Verify signature
    ↓
8. Extract peer ID from signature
    ↓
9. Continue to handler
```

## Database Schema

### Tables

#### `leases`
```sql
CREATE TABLE leases (
    token_id BIGINT PRIMARY KEY,
    peer_id VARCHAR(128) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_leases_expires_at ON leases(expires_at);
```

#### `nonces`
```sql
CREATE TABLE nonces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    peer_id VARCHAR(128) NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    used_at TIMESTAMPTZ
);
```

#### `alloc_state`
```sql
CREATE TABLE alloc_state (
    id SERIAL PRIMARY KEY,
    last_token_id BIGINT NOT NULL
);
```

### Relationships

- **leases**: Independent table with token_id as primary key
- **nonces**: Independent table for authentication
- **alloc_state**: Singleton table for tracking allocation state

## Authentication Flow

### libp2p Signature Verification

1. **Key Generation**: Client generates libp2p key pair
2. **Nonce Request**: Client sends public key to `/request-auth`
3. **Nonce Generation**: Server generates UUID nonce, stores with expiration
4. **Nonce Signing**: Client signs nonce with private key
5. **Signature Verification**: Server verifies signature using public key

### Security Features

- **Time-limited nonces**: Default 5-minute expiration
- **One-time use**: Nonces marked as used after verification
- **Cryptographic signatures**: libp2p secp256k1 signatures
- **Rate limiting**: 1000 requests per second
- **CORS protection**: Configurable CORS policies

## Lease Allocation Algorithm

### Allocation Strategy

1. **Check Existing Lease**: Look for active lease for peer ID
2. **Reuse Expired Lease**: Find and reuse expired lease (retry logic)
3. **Allocate New Lease**: Generate new token ID if no expired lease found
4. **Retry Logic**: Configurable retries with exponential backoff

### Implementation

```go
func (s *LeaseService) AllocateIP(ctx context.Context, peerID string) (*models.Lease, error) {
    // Check existing lease
    if lease, err := s.repo.GetLeaseByPeerID(ctx, peerID); lease != nil && err == nil {
        return lease, nil
    }
    
    // Try to reuse expired lease (with retries)
    for retries := 0; retries < s.maxRetries; retries++ {
        if lease, err := s.repo.FindAndReuseExpiredLease(ctx, peerID); lease != nil {
            return lease, nil
        }
        time.Sleep(s.retryDelay)
    }
    
    // Allocate new lease (with retries)
    for retries := 0; retries < s.maxRetries; retries++ {
        if lease, err := s.repo.AllocateNewLease(ctx, peerID); lease != nil {
            return lease, nil
        }
        time.Sleep(s.retryDelay)
    }
    
    return nil, fmt.Errorf("failed to allocate lease after %d retries", s.maxRetries)
}
```

## Nonce Management

### Nonce Lifecycle

1. **Generation**: UUID-based nonce with expiration
2. **Storage**: Redis with TTL (default: 5 minutes)
3. **Verification**: Signature verification against nonce
4. **Cleanup**: Background job removes expired nonces

### Background Cleanup

```go
type NonceCleaner struct {
    repo   ports.NonceRepository
    logger *zap.Logger
}

func (c *NonceCleaner) CleanupExpiredNonces(ctx context.Context) error {
    return c.repo.CleanupExpiredNonces(ctx)
}
```

## Error Handling

### Error Hierarchy

```
Domain Errors (Business Logic)
    ↓
Application Errors (Use Cases)
    ↓
Adapter Errors (External Systems)
    ↓
HTTP Errors (Client Response)
```

### Error Propagation

1. **Domain Layer**: Returns domain-specific errors
2. **Application Layer**: Wraps domain errors with context
3. **Adapter Layer**: Converts to HTTP status codes
4. **Handler Layer**: Formats error responses

## Performance Considerations

### Caching Strategy

- **Redis**: Nonce storage and lease caching
- **TTL-based**: Automatic expiration
- **Cache-aside**: Read-through cache pattern

### Database Optimization

- **Indexes**: Optimized queries with proper indexes
- **Connection Pooling**: Configurable pool sizes
- **Query Optimization**: Efficient SQL queries

### Scalability

- **Horizontal Scaling**: Stateless application design
- **Database Scaling**: Read replicas for queries
- **Cache Scaling**: Redis cluster support
- **Load Balancing**: Multiple application instances

### Monitoring

- **Health Checks**: `/health` and `/ready` endpoints
- **Metrics**: Request/response metrics
- **Logging**: Structured logging with Zap
- **Tracing**: Request tracing capabilities

## Security Architecture

### Defense in Depth

1. **Network Security**: CORS, rate limiting
2. **Authentication**: libp2p cryptographic signatures
3. **Authorization**: Peer-based access control
4. **Data Protection**: Encrypted connections, secure storage
5. **Monitoring**: Security event logging

### Threat Model

- **Replay Attacks**: Prevented by nonce expiration
- **Man-in-the-Middle**: Prevented by cryptographic signatures
- **DoS Attacks**: Mitigated by rate limiting
- **Data Tampering**: Prevented by signature verification

This architecture ensures DHCP2P is secure, scalable, maintainable, and follows industry best practices for distributed systems.
