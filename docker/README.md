# DHCP2P Docker Deployment Guide

This guide provides comprehensive instructions for deploying DHCP2P using Docker and Docker Compose.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Architecture Overview](#architecture-overview)
- [Quick Start](#quick-start)
- [Setup Scenarios](#setup-scenarios)
- [Migration Strategies](#migration-strategies)
- [Environment Configuration](#environment-configuration)
- [Common Commands](#common-commands)
- [Upgrade Guide](#upgrade-guide)
- [Rollback Procedure](#rollback-procedure)
- [Troubleshooting](#troubleshooting)
- [Security Best Practices](#security-best-practices)
- [Performance Tuning](#performance-tuning)

## Prerequisites

- **Docker**: Version 20.10 or higher
- **Docker Compose**: Version 2.0 or higher
- **Make**: For using convenience commands (optional)
- **Atlas CLI**: For manual migrations (optional)

### Installation

#### Docker and Docker Compose
```bash
# macOS (using Homebrew)
brew install docker docker-compose

# Ubuntu/Debian
sudo apt-get update
sudo apt-get install docker.io docker-compose

# Verify installation
docker --version
docker-compose --version
```

#### Atlas CLI (Optional)
```bash
curl -sSf https://atlasgo.sh | sh
```

## Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   DHCP2P App    │    │   PostgreSQL    │    │     Redis       │
│   Port: 8088    │◄──►│   Port: 5432    │    │   Port: 6379    │
│                 │    │                 │    │                 │
│  - HTTP API     │    │  - Leases       │    │  - Nonces       │
│  - Health Check │    │  - Nonces       │    │  - Cache        │
│  - Auth         │    │  - State        │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │
         ▼
┌─────────────────┐
│ Migration Init  │
│   Container     │
│                 │
│  - Atlas CLI    │
│  - DB Setup     │
└─────────────────┘
```

### Service Dependencies

1. **PostgreSQL** starts first (health check required)
2. **Redis** starts in parallel (health check required)
3. **Migration Container** runs after DB is healthy
4. **DHCP2P Application** starts after migration completes

### Port Reference

| Service | Port | Description |
|---------|------|-------------|
| DHCP2P | 8088 | HTTP API, Health checks |
| PostgreSQL | 5432 | Database (internal) |
| Redis | 6379 | Cache (internal) |

## Quick Start

### Development Environment

1. **Clone and setup**:
   ```bash
   git clone <repository>
   cd dhcp2p
   make dev
   ```

2. **Access the application**:
   - API: http://localhost:8088
   - Health: http://localhost:8088/health
   - Readiness: http://localhost:8088/ready

### Production Environment

1. **Setup production**:
   ```bash
   cp .env.prod.example .env.prod
   # Edit .env.prod with production values
   make prod
   ```

## Setup Scenarios

### Scenario 1: New Deployment

```bash
# Interactive setup
./scripts/setup.sh

# Or using Make
make docker-setup
```

This will:
- Create necessary directories
- Set up environment configuration
- Guide you through database and Redis configuration

### Scenario 2: Using External Databases

1. **Update environment file**:
   ```bash
   # .env
   DATABASE_URL=postgres://user:pass@external-db:5432/dhcp2p?sslmode=require
   REDIS_URL=redis://user:pass@external-redis:6379
   ```

2. **Start with external services**:
   ```bash
   # Development
   docker-compose up dhcp2p-migrate dhcp2p

   # Production
   docker-compose -f docker-compose.prod.yml up dhcp2p-migrate dhcp2p
   ```

## Migration Strategies

### Strategy 1: Automatic Migrations (Default)

Migrations run automatically via init container:

```bash
# Development
make docker-up

# Production
make docker-up-prod
```

### Strategy 2: Manual Migrations

Run migrations before starting the application:

```bash
# Using Docker
./scripts/migrate.sh --auto

# Using local Atlas CLI
./scripts/migrate.sh --manual

# Check status
./scripts/migrate.sh --status
```

### Strategy 3: Separate Migration Container

```bash
# Run migrations only
docker-compose run --rm dhcp2p-migrate

# Then start application
docker-compose up dhcp2p
```

## Environment Configuration

### Development (.env)

```bash
# Database
DATABASE_URL=postgres://dhcp2p:your_password@postgres:5432/dhcp2p?sslmode=disable
REDIS_URL=redis://redis:6379

# Application
PORT=8088
LOG_LEVEL=debug
# PASSWORD and ACCOUNT are no longer required since Ethereum/keystore functionality was removed
# PASSWORD=your_keystore_password
# ACCOUNT=0xyour_account_address

# Optional
NONCE_TTL=5
LEASE_TTL=120
```

### Production (.env.prod)

```bash
# Database (use strong passwords)
DATABASE_URL=postgres://user:STRONG_PASSWORD@external-db:5432/dhcp2p?sslmode=require
REDIS_URL=redis://:STRONG_PASSWORD@external-redis:6379

# Application
PORT=8088
LOG_LEVEL=info
# PASSWORD and ACCOUNT are no longer required since Ethereum/keystore functionality was removed
# PASSWORD=STRONG_KEYSTORE_PASSWORD
# ACCOUNT=0xACCOUNT_ADDRESS

# Production settings
NONCE_TTL=5
LEASE_TTL=120
MAX_LEASE_RETRIES=3
```

## Common Commands

### Using Make (Recommended)

```bash
# Help
make help

# Development
make dev              # Complete setup
make dev-up           # Start services
make dev-logs         # View logs
make dev-down         # Stop services

# Production
make prod             # Complete setup
make prod-up          # Start services
make prod-logs        # View logs

# Management
make docker-health    # Check health
make docker-ps        # Show containers
make docker-clean     # Clean up
make docker-shell     # Get shell access
```

### Using Docker Compose Directly

```bash
# Development
docker-compose up -d
docker-compose logs -f
docker-compose down

# Production
docker-compose -f docker-compose.prod.yml up -d
docker-compose -f docker-compose.prod.yml logs -f
docker-compose -f docker-compose.prod.yml down
```

## Upgrade Guide

### Step-by-Step Upgrade Process

1. **Stop current version gracefully**:
   ```bash
   make docker-down
   ```

2. **Backup data**:
   ```bash
   # Backup database
   docker-compose exec postgres pg_dump -U dhcp2p dhcp2p > backup.sql

   # Backup application data
   cp -r data backup-data/
   ```

3. **Pull new images**:
   ```bash
   make docker-build
   ```

4. **Run migrations**:
   ```bash
   make docker-migrate
   ```

5. **Start new version**:
   ```bash
   make docker-up
   ```

6. **Verify health**:
   ```bash
   make docker-health
   curl http://localhost:8088/health
   ```

7. **Rollback if needed** (see Rollback Procedure)

## Rollback Procedure

### If Upgrade Fails

1. **Stop new version**:
   ```bash
   make docker-down
   ```

2. **Restore database**:
   ```bash
   # Restore from backup
   docker-compose up -d postgres
   docker-compose exec postgres psql -U dhcp2p dhcp2p < backup.sql
   ```

3. **Restore application data**:
   ```bash
   cp -r backup-data/* data/
   ```

4. **Revert to previous image**:
   ```bash
   # Tag previous image as latest
   docker tag dhcp2p:previous dhcp2p:latest
   ```

5. **Start previous version**:
   ```bash
   make docker-up
   ```

6. **Verify functionality**:
   ```bash
   make docker-health
   curl http://localhost:8088/health
   ```

## Troubleshooting

### Common Issues

#### 1. Container Won't Start

**Symptoms**: Container exits immediately
**Solutions**:
```bash
# Check logs
make docker-logs

# Check environment variables
docker-compose config

# Verify required files exist
ls -la data/keystore/
```

#### 2. Database Connection Failed

**Symptoms**: "database connection failed" errors
**Solutions**:
```bash
# Check database health
make docker-health

# Test database connection
docker-compose exec postgres pg_isready -U dhcp2p

# Check database logs
make docker-logs-db
```

#### 3. Migration Failures

**Symptoms**: Migration container fails
**Solutions**:
```bash
# Check migration status
make docker-migrate-status

# Run migrations manually
make docker-migrate

# Check database permissions
docker-compose exec postgres psql -U dhcp2p -c "\du"
```

#### 4. Health Check Failures

**Symptoms**: Health endpoint returns 503
**Solutions**:
```bash
# Check application logs
make docker-logs-app

# Test health endpoint manually
curl -v http://localhost:8088/health

# Check service dependencies
make docker-ps
```

### Debug Commands

```bash
# Get shell access
make docker-shell

# Check container resources
docker stats

# Inspect container configuration
docker inspect dhcp2p-app

# Check network connectivity
docker-compose exec dhcp2p ping postgres
docker-compose exec dhcp2p ping redis
```

## Security Best Practices

### Production Security Checklist

- [ ] **Use strong passwords** for all services
- [ ] **Enable SSL/TLS** for database connections
- [ ] **Use external databases** instead of containerized ones
- [ ] **Implement secrets management** (Docker secrets, external vaults)
- [ ] **Enable read-only filesystems** where possible
- [ ] **Use non-root users** (already implemented)
- [ ] **Regular security updates** for base images
- [ ] **Network isolation** between services
- [ ] **Backup encryption** for sensitive data
- [ ] **Monitor access logs** and failed attempts

### Secrets Management

```bash
# Using Docker secrets (production)
echo "strong_password" | docker secret create keystore_password -
echo "db_password" | docker secret create db_password -

# Update docker-compose.prod.yml to use secrets
secrets:
  - keystore_password
  - db_password
```

### Network Security

```bash
# Create isolated network
docker network create --driver bridge dhcp2p-network

# Use in docker-compose.yml
networks:
  dhcp2p-network:
    driver: bridge
    internal: true  # No external access
```

## Performance Tuning

### Resource Limits

```yaml
# docker-compose.prod.yml
services:
  dhcp2p:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '1.0'
          memory: 1G
```

### Database Optimization

```bash
# PostgreSQL tuning
POSTGRES_SHARED_BUFFERS=256MB
POSTGRES_EFFECTIVE_CACHE_SIZE=1GB
POSTGRES_MAINTENANCE_WORK_MEM=64MB
```

### Redis Optimization

```bash
# Redis configuration
redis-server --maxmemory 512mb --maxmemory-policy allkeys-lru
```

### Connection Pooling

```bash
# Application configuration
DATABASE_MAX_CONNECTIONS=20
DATABASE_MIN_CONNECTIONS=5
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=5
```

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Reference](https://docs.docker.com/compose/)
- [Atlas CLI Documentation](https://atlasgo.io/)
- [PostgreSQL Docker Image](https://hub.docker.com/_/postgres)
- [Redis Docker Image](https://hub.docker.com/_/redis)

## Support

For issues and questions:
1. Check this documentation
2. Review troubleshooting section
3. Check application logs: `make docker-logs`
4. Create an issue in the repository
