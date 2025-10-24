# DHCP2P Configuration Reference

This document provides a comprehensive reference for all configuration options available in DHCP2P, including environment variables, configuration files, and their relationships.

## Table of Contents

- [Configuration Overview](#configuration-overview)
- [Environment Variables](#environment-variables)
- [Configuration File](#configuration-file)
- [Configuration Precedence](#configuration-precedence)
- [Server Configuration](#server-configuration)
- [Database Configuration](#database-configuration)
- [Redis Configuration](#redis-configuration)
- [Cache Configuration](#cache-configuration)
- [Authentication Configuration](#authentication-configuration)
- [Lease Configuration](#lease-configuration)
- [Logging Configuration](#logging-configuration)
- [Security Configuration](#security-configuration)
- [Performance Configuration](#performance-configuration)
- [Configuration Examples](#configuration-examples)

## Configuration Overview

DHCP2P supports multiple configuration methods with the following precedence (highest to lowest):

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration file** (`config.yaml`)
4. **Default values** (lowest priority)

### Configuration Sources

- **Environment Variables**: Prefixed with `DHCP2P_`
- **Configuration File**: YAML format in `config/config.yaml`
- **Command-line Flags**: Override all other sources
- **Default Values**: Built-in application defaults

## Environment Variables

### Required Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `DHCP2P_DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@host:5432/db` | Yes |
| `DHCP2P_REDIS_URL` | Redis connection string | `redis://host:6379` | Yes |

### Server Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DHCP2P_PORT` | HTTP server port | `8088` | `8088` |
| `DHCP2P_LOG_LEVEL` | Logging level | `info` | `debug`, `info`, `warn`, `error` |

### Database Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DHCP2P_DATABASE_URL` | PostgreSQL connection string | - | `postgres://user:pass@host:5432/db?sslmode=require` |

### PostgreSQL Pool Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DHCP2P_DB_MAX_CONNS` | Maximum number of connections in the pool | `4` | `25` |
| `DHCP2P_DB_MIN_CONNS` | Minimum number of connections in the pool | `0` | `5` |
| `DHCP2P_DB_MAX_CONN_LIFETIME` | Maximum lifetime of a connection in minutes | `60` | `30` |
| `DHCP2P_DB_MAX_CONN_IDLE_TIME` | Maximum idle time of a connection in minutes | `30` | `5` |
| `DHCP2P_DB_HEALTH_CHECK_PERIOD` | Health check period in seconds | `60` | `30` |

### Redis Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DHCP2P_REDIS_URL` | Redis connection string | - | `redis://host:6379` |
| `DHCP2P_REDIS_PASSWORD` | Redis password | - | `secret123` |
| `DHCP2P_REDIS_MAX_RETRIES` | Maximum connection retries | `3` | `5` |
| `DHCP2P_REDIS_POOL_SIZE` | Connection pool size | `10` | `20` |
| `DHCP2P_REDIS_MIN_IDLE_CONNS` | Minimum idle connections | `5` | `10` |
| `DHCP2P_REDIS_DIAL_TIMEOUT` | Dial timeout in seconds | `5` | `10` |
| `DHCP2P_REDIS_READ_TIMEOUT` | Read timeout in seconds | `3` | `5` |
| `DHCP2P_REDIS_WRITE_TIMEOUT` | Write timeout in seconds | `3` | `5` |

### Cache Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DHCP2P_CACHE_ENABLED` | Enable caching | `true` | `false` |
| `DHCP2P_CACHE_DEFAULT_TTL` | Default cache TTL in minutes | `30` | `60` |

### Authentication Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DHCP2P_NONCE_TTL` | Nonce TTL in minutes | `5` | `10` |
| `DHCP2P_NONCE_CLEANER_INTERVAL` | Nonce cleanup interval in minutes | `5` | `10` |

### Lease Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DHCP2P_LEASE_TTL` | Lease TTL in minutes | `120` | `240` |
| `DHCP2P_MAX_LEASE_RETRIES` | Maximum lease allocation retries | `3` | `5` |
| `DHCP2P_LEASE_RETRY_DELAY` | Lease retry delay in milliseconds | `500` | `1000` |

## Configuration File

### File Location

The configuration file can be located in several places:

1. **Specified path**: `--config /path/to/config.yaml`
2. **Current directory**: `./config/config.yaml`
3. **System directory**: `/etc/dhcp2p/config.yaml`

### Configuration File Format

```yaml
# Server Configuration
port: 8088
log_level: info

# Database Configuration
database_url: "postgres://dhcp2p:password@localhost:5432/dhcp2p?sslmode=disable"

# PostgreSQL Pool Configuration (recommended values)
db_max_conns: 25
db_min_conns: 5
db_max_conn_lifetime: 30  # minutes
db_max_conn_idle_time: 5  # minutes
db_health_check_period: 30 # seconds

# Redis Configuration
redis_url: "redis://localhost:6379"
redis_password: ""
redis_max_retries: 3
redis_pool_size: 10
redis_min_idle_conns: 5
redis_dial_timeout: 5
redis_read_timeout: 3
redis_write_timeout: 3

# Cache Configuration
cache_enabled: true
cache_default_ttl: 30

# Authentication Configuration
nonce_ttl: 5
nonce_cleaner_interval: 5

# Lease Configuration
lease_ttl: 120
max_lease_retries: 3
lease_retry_delay: 500
```

## Configuration Precedence

### Example Configuration Resolution

Given the following configuration sources:

**1. Default values:**
```go
Port: 8088
LogLevel: "info"
```

**2. Configuration file (`config.yaml`):**
```yaml
port: 9000
log_level: "debug"
```

**3. Environment variables:**
```bash
DHCP2P_PORT=7000
DHCP2P_LOG_LEVEL=warn
```

**4. Command-line flags:**
```bash
--port 6000 --log-level error
```

**Final resolved values:**
- Port: `6000` (command-line flag wins)
- Log Level: `error` (command-line flag wins)

## Server Configuration

### Port Configuration

```yaml
# Configuration file
port: 8088

# Environment variable
DHCP2P_PORT=8088

# Command-line flag
--port 8088
```

### Logging Configuration

```yaml
# Configuration file
log_level: info  # debug, info, warn, error

# Environment variable
DHCP2P_LOG_LEVEL=info

# Command-line flag
--log-level info
```

## Database Configuration

### PostgreSQL Connection

```yaml
# Configuration file
database_url: "postgres://user:password@host:port/database?sslmode=mode"

# Environment variable
DHCP2P_DATABASE_URL="postgres://user:password@host:port/database?sslmode=mode"
```

### Connection String Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `user` | Database username | `dhcp2p` |
| `password` | Database password | `secret123` |
| `host` | Database host | `localhost`, `db.example.com` |
| `port` | Database port | `5432` |
| `database` | Database name | `dhcp2p` |
| `sslmode` | SSL mode | `disable`, `require`, `verify-full` |

### SSL Mode Options

- **`disable`**: No SSL connection
- **`require`**: SSL connection required
- **`verify-full`**: SSL connection with certificate verification

### PostgreSQL Pool Configuration

The PostgreSQL connection pool settings control how the application manages database connections for optimal performance and resource usage.

**Note**: The default values shown in the table below come from the pgxpool library itself. If these configuration values are not set (or set to 0), the pgxpool library defaults will be used.

```yaml
# Configuration file
db_max_conns: 25
db_min_conns: 5
db_max_conn_lifetime: 30  # minutes
db_max_conn_idle_time: 5   # minutes
db_health_check_period: 30 # seconds

# Environment variables
DHCP2P_DB_MAX_CONNS=25
DHCP2P_DB_MIN_CONNS=5
DHCP2P_DB_MAX_CONN_LIFETIME=30
DHCP2P_DB_MAX_CONN_IDLE_TIME=5
DHCP2P_DB_HEALTH_CHECK_PERIOD=30
```

#### Pool Configuration Parameters

| Parameter | Description | Recommended Value | Impact |
|-----------|-------------|-------------------|---------|
| `db_max_conns` | Maximum connections in pool | 25-50 | Higher = more concurrent operations, but more memory usage |
| `db_min_conns` | Minimum connections in pool | 5-10 | Higher = faster response, but more resource usage |
| `db_max_conn_lifetime` | Connection lifetime | 30-60 minutes | Prevents stale connections, reduces memory leaks |
| `db_max_conn_idle_time` | Idle connection timeout | 5-15 minutes | Frees unused connections, saves resources |
| `db_health_check_period` | Health check frequency | 30-60 seconds | Detects dead connections, improves reliability |

#### Performance Tuning Guidelines

**High Traffic Applications:**
```yaml
db_max_conns: 50
db_min_conns: 10
db_max_conn_lifetime: 60
db_max_conn_idle_time: 10
db_health_check_period: 30
```

**Low Traffic Applications:**
```yaml
db_max_conns: 10
db_min_conns: 2
db_max_conn_lifetime: 30
db_max_conn_idle_time: 5
db_health_check_period: 60
```

**Resource Constrained:**
```yaml
db_max_conns: 15
db_min_conns: 3
db_max_conn_lifetime: 45
db_max_conn_idle_time: 8
db_health_check_period: 45
```

## Redis Configuration

### Basic Redis Configuration

```yaml
# Configuration file
redis_url: "redis://host:port"
redis_password: "optional_password"

# Environment variables
DHCP2P_REDIS_URL="redis://localhost:6379"
DHCP2P_REDIS_PASSWORD="secret123"
```

### Advanced Redis Configuration

```yaml
# Connection pool settings
redis_pool_size: 10
redis_min_idle_conns: 5

# Timeout settings
redis_dial_timeout: 5
redis_read_timeout: 3
redis_write_timeout: 3

# Retry settings
redis_max_retries: 3
```

### Redis URL Formats

| Format | Description | Example |
|--------|-------------|---------|
| `redis://host:port` | Basic connection | `redis://localhost:6379` |
| `redis://:password@host:port` | With password | `redis://:secret@localhost:6379` |
| `redis://user:password@host:port` | With username and password | `redis://user:pass@localhost:6379` |
| `redis://host:port/db` | With database number | `redis://localhost:6379/0` |

## Cache Configuration

### Cache Settings

```yaml
# Enable/disable caching
cache_enabled: true

# Default TTL for cached items (minutes)
cache_default_ttl: 30
```

### Cache Behavior

- **Enabled**: Nonces and lease data cached in Redis
- **Disabled**: Direct database access for all operations
- **TTL**: Automatic expiration of cached items

## Authentication Configuration

### Nonce Configuration

```yaml
# Nonce time-to-live (minutes)
nonce_ttl: 5

# Nonce cleanup interval (minutes)
nonce_cleaner_interval: 5
```

### Nonce Lifecycle

1. **Generation**: Nonce created with expiration time
2. **Storage**: Stored in Redis with TTL
3. **Verification**: Signature verification against nonce
4. **Cleanup**: Background job removes expired nonces

## Lease Configuration

### Lease Settings

```yaml
# Lease time-to-live (minutes)
lease_ttl: 120

# Maximum retry attempts for lease allocation
max_lease_retries: 3

# Delay between retry attempts (milliseconds)
lease_retry_delay: 500
```

### Lease Allocation Strategy

1. **Check existing lease**: Look for active lease for peer ID
2. **Reuse expired lease**: Find and reuse expired lease
3. **Allocate new lease**: Generate new token ID
4. **Retry logic**: Configurable retries with delay

## Logging Configuration

### Log Levels

| Level | Description | Use Case |
|-------|-------------|----------|
| `debug` | Detailed debugging information | Development |
| `info` | General information messages | Production |
| `warn` | Warning messages | Production |
| `error` | Error messages only | Troubleshooting |

### Log Format

Structured JSON logging with Zap:

```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "caller": "handlers/http/lease.go:45",
  "msg": "lease allocated",
  "peer_id": "12D3KooWExample",
  "token_id": 12345,
  "ttl": 120
}
```

## Security Configuration

### Rate Limiting

```yaml
# Rate limiting is configured in the HTTP server middleware
# Default: 1000 requests per second
```

### CORS Configuration

```yaml
# CORS is configured in the HTTP server middleware
# Default: Allow all origins (development)
```

### Security Headers

```yaml
# Security headers are configured in the HTTP server middleware
# Includes: X-Content-Type-Options, X-Frame-Options, etc.
```

## Performance Configuration

### Connection Pooling

```yaml
# Redis connection pool
redis_pool_size: 10
redis_min_idle_conns: 5

# PostgreSQL connection pool
db_max_conns: 25
db_min_conns: 5
db_max_conn_lifetime: 30  # minutes
db_max_conn_idle_time: 5   # minutes
db_health_check_period: 30 # seconds
```

### Timeout Configuration

```yaml
# Redis timeouts
redis_dial_timeout: 5
redis_read_timeout: 3
redis_write_timeout: 3

# HTTP server timeout (configured in middleware)
# Default: 60 seconds
```

### Retry Configuration

```yaml
# Lease allocation retries
max_lease_retries: 3
lease_retry_delay: 500

# Redis connection retries
redis_max_retries: 3
```

## Configuration Examples

### Development Configuration

```yaml
# config/dev.yaml
port: 8088
log_level: debug
database_url: "postgres://dhcp2p:your_password@localhost:5432/dhcp2p_dev?sslmode=disable"
redis_url: "redis://localhost:6379"
nonce_ttl: 5
lease_ttl: 120
cache_enabled: true
```

### Staging Configuration

```yaml
# config/staging.yaml
port: 8088
log_level: info
database_url: "postgres://user:pass@staging-db:5432/dhcp2p?sslmode=require"
redis_url: "redis://staging-redis:6379"
nonce_ttl: 5
lease_ttl: 120
max_lease_retries: 3
cache_enabled: true
```

### Production Configuration

```yaml
# config/prod.yaml
port: 8088
log_level: warn
database_url: "postgres://user:STRONG_PASS@prod-db:5432/dhcp2p?sslmode=require"
redis_url: "redis://:STRONG_PASS@prod-redis:6379"
nonce_ttl: 5
lease_ttl: 120
max_lease_retries: 5
lease_retry_delay: 1000
redis_pool_size: 20
redis_min_idle_conns: 10
cache_enabled: true
cache_default_ttl: 60

# PostgreSQL Pool Configuration
db_max_conns: 50
db_min_conns: 10
db_max_conn_lifetime: 60
db_max_conn_idle_time: 10
db_health_check_period: 30
```

### Docker Environment Configuration

```bash
# .env file for Docker Compose
DATABASE_URL=postgres://dhcp2p:your_password@postgres:5432/dhcp2p?sslmode=disable
REDIS_URL=redis://redis:6379
PORT=8088
LOG_LEVEL=info
NONCE_TTL=5
LEASE_TTL=120
```

### Kubernetes Configuration

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: dhcp2p-config
data:
  config.yaml: |
    port: 8088
    log_level: info
    nonce_ttl: 5
    lease_ttl: 120
    cache_enabled: true
```

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: dhcp2p-secrets
data:
  database-url: cG9zdGdyZXM6Ly91c2VyOnBhc3NAaG9zdDo1NDMyL2Ricw==
  redis-url: cmVkaXM6Ly9ob3N0OjYzNzk=
```

### Environment-Specific Overrides

```bash
# Override specific settings for different environments
export DHCP2P_LOG_LEVEL=debug
export DHCP2P_NONCE_TTL=10
export DHCP2P_CACHE_ENABLED=false

# Run with overrides
go run cmd/dhcp2p/main.go serve
```

### Configuration Validation

The application validates configuration on startup:

```bash
# Invalid configuration will cause startup failure
DHCP2P_DATABASE_URL="invalid-url" go run cmd/dhcp2p/main.go serve
# Error: invalid database URL format

# Missing required configuration will cause startup failure
go run cmd/dhcp2p/main.go serve
# Error: DATABASE_URL is required
```

This configuration reference ensures administrators can properly configure DHCP2P for their specific environment and requirements.
