# DHCP2P Deployment Guide

This guide provides comprehensive instructions for deploying DHCP2P in various environments, from development to production.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Environment Variables](#environment-variables)
- [Configuration Files](#configuration-files)
- [Docker Deployment](#docker-deployment)
- [Production Deployment](#production-deployment)
- [Environment-Specific Configurations](#environment-specific-configurations)
- [Monitoring and Observability](#monitoring-and-observability)
- [Backup and Recovery](#backup-and-recovery)
- [Scaling Considerations](#scaling-considerations)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

- **CPU**: 2+ cores recommended
- **Memory**: 2GB+ RAM recommended
- **Storage**: 10GB+ disk space
- **Network**: Port 8088 accessible

### Software Dependencies

- **Docker**: 20.10+ (for containerized deployment)
- **Docker Compose**: 2.0+ (for multi-container orchestration)
- **PostgreSQL**: 13+ (for database)
- **Redis**: 6.0+ (for caching)
- **Make**: Optional (for convenience commands)

### External Services

- **PostgreSQL Database**: Primary data store
- **Redis Cache**: Nonce storage and caching
- **Load Balancer**: For production deployments (optional)

## Environment Variables

### Required Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@host:5432/db` | Yes |
| `REDIS_URL` | Redis connection string | `redis://host:6379` | Yes |

### Optional Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `PORT` | HTTP server port | `8088` | `8088` |
| `LOG_LEVEL` | Logging level | `info` | `debug`, `info`, `warn`, `error` |
| `REDIS_PASSWORD` | Redis password | - | `secret123` |
| `NONCE_TTL` | Nonce TTL in minutes | `5` | `10` |
| `NONCE_CLEANER_INTERVAL` | Nonce cleanup interval in minutes | `5` | `10` |
| `LEASE_TTL` | Lease TTL in minutes | `120` | `240` |
| `MAX_LEASE_RETRIES` | Maximum lease allocation retries | `3` | `5` |
| `LEASE_RETRY_DELAY` | Lease retry delay in milliseconds | `500` | `1000` |

### Redis Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `REDIS_MAX_RETRIES` | Maximum Redis connection retries | `3` | `5` |
| `REDIS_POOL_SIZE` | Redis connection pool size | `10` | `20` |
| `REDIS_MIN_IDLE_CONNS` | Minimum idle connections | `5` | `10` |
| `REDIS_DIAL_TIMEOUT` | Dial timeout in seconds | `5` | `10` |
| `REDIS_READ_TIMEOUT` | Read timeout in seconds | `3` | `5` |
| `REDIS_WRITE_TIMEOUT` | Write timeout in seconds | `3` | `5` |

### Cache Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `CACHE_ENABLED` | Enable caching | `true` | `false` |
| `CACHE_DEFAULT_TTL` | Default cache TTL in minutes | `30` | `60` |

## Configuration Files

### config.yaml

The application supports YAML configuration files. Place `config.yaml` in the `config/` directory or specify the path via the `--config` flag.

```yaml
# Server Configuration
port: 8088
log_level: info

# Database Configuration
database_url: "postgres://dhcp2p:password@localhost:5432/dhcp2p?sslmode=disable"

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

# Nonce Configuration
nonce_ttl: 5
nonce_cleaner_interval: 5

# Lease Configuration
lease_ttl: 120
max_lease_retries: 3
lease_retry_delay: 500
```

### Environment File (.env)

For Docker deployments, use environment files:

```bash
# Database
DATABASE_URL=postgres://dhcp2p:your_password@postgres:5432/dhcp2p?sslmode=disable
REDIS_URL=redis://redis:6379

# Application
PORT=8088
LOG_LEVEL=info

# Optional
NONCE_TTL=5
LEASE_TTL=120
MAX_LEASE_RETRIES=3
LEASE_RETRY_DELAY=500
```

## Docker Deployment

### Quick Start

```bash
# Clone repository
git clone https://github.com/unicornultrafoundation/dhcp2p.git
cd dhcp2p

# Start development stack
make docker-up

# Check health
curl http://localhost:8088/health
```

### Production Deployment

```bash
# Setup production environment
cp .env.prod.example .env.prod
# Edit .env.prod with production values

# Start production stack
make docker-up-prod
```

### Custom Configuration

```bash
# Use custom environment file
make docker-up ENV_FILE=custom.env

# Use custom Docker Compose file
docker-compose -f docker-compose.custom.yml up -d
```

## Production Deployment

### Infrastructure Requirements

- **Load Balancer**: nginx, HAProxy, or cloud load balancer
- **Database**: Managed PostgreSQL service (AWS RDS, Google Cloud SQL, etc.)
- **Cache**: Managed Redis service (AWS ElastiCache, Google Memorystore, etc.)
- **Monitoring**: Prometheus, Grafana, or cloud monitoring
- **Logging**: Centralized logging (ELK stack, cloud logging)

### Deployment Steps

1. **Prepare Infrastructure**
   ```bash
   # Create production environment file
   cp .env.prod.example .env.prod
   
   # Configure production values
   DATABASE_URL=postgres://user:STRONG_PASSWORD@prod-db:5432/dhcp2p?sslmode=require
   REDIS_URL=redis://:STRONG_PASSWORD@prod-redis:6379
   LOG_LEVEL=info
   ```

2. **Deploy Application**
   ```bash
   # Build production image
   make docker-build IMAGE=your-registry/dhcp2p:latest
   
   # Push to registry
   make docker-push IMAGE=your-registry/dhcp2p:latest
   
   # Deploy to production
   make docker-up-prod
   ```

3. **Verify Deployment**
   ```bash
   # Check health
   curl https://your-domain.com/health
   
   # Check readiness
   curl https://your-domain.com/ready
   ```

### High Availability Setup

```yaml
# docker-compose.ha.yml
version: '3.8'
services:
  dhcp2p-1:
    image: dhcp2p:latest
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
```

## Environment-Specific Configurations

### Development

```yaml
# config/dev.yaml
port: 8088
log_level: debug
database_url: "postgres://dhcp2p:your_password@localhost:5432/dhcp2p_dev?sslmode=disable"
redis_url: "redis://localhost:6379"
nonce_ttl: 5
lease_ttl: 120
```

### Staging

```yaml
# config/staging.yaml
port: 8088
log_level: info
database_url: "postgres://user:pass@staging-db:5432/dhcp2p?sslmode=require"
redis_url: "redis://staging-redis:6379"
nonce_ttl: 5
lease_ttl: 120
max_lease_retries: 3
```

### Production

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
```

## Monitoring and Observability

### Health Checks

The application provides two health check endpoints:

- **`/health`**: Basic health check
- **`/ready`**: Readiness check (dependencies available)

### Metrics Collection

```bash
# Enable metrics endpoint (if implemented)
curl http://localhost:8088/metrics
```

### Logging

Structured logging with Zap:

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

### Monitoring Stack

```yaml
# monitoring/docker-compose.yml
version: '3.8'
services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
  
  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

## Backup and Recovery

### Database Backup

```bash
# Create backup
docker-compose exec postgres pg_dump -U dhcp2p dhcp2p > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore backup
docker-compose exec -T postgres psql -U dhcp2p dhcp2p < backup_20240115_103000.sql
```

### Automated Backups

```bash
#!/bin/bash
# backup.sh
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/dhcp2p_$DATE.sql"

# Create backup
docker-compose exec postgres pg_dump -U dhcp2p dhcp2p > "$BACKUP_FILE"

# Compress backup
gzip "$BACKUP_FILE"

# Remove old backups (keep 30 days)
find "$BACKUP_DIR" -name "dhcp2p_*.sql.gz" -mtime +30 -delete
```

### Redis Backup

```bash
# Redis persistence is handled automatically
# For manual backup:
docker-compose exec redis redis-cli BGSAVE
```

## Scaling Considerations

### Horizontal Scaling

```yaml
# docker-compose.scale.yml
version: '3.8'
services:
  dhcp2p:
    image: dhcp2p:latest
    deploy:
      replicas: 5
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
```

### Database Scaling

- **Read Replicas**: For read-heavy workloads
- **Connection Pooling**: Configure appropriate pool sizes
- **Query Optimization**: Monitor slow queries

### Redis Scaling

- **Redis Cluster**: For high availability
- **Memory Optimization**: Configure appropriate memory limits
- **Persistence**: Configure RDB and AOF as needed

### Load Balancer Configuration

```nginx
# nginx.conf
upstream dhcp2p {
    server dhcp2p-1:8088;
    server dhcp2p-2:8088;
    server dhcp2p-3:8088;
}

server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://dhcp2p;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
    
    location /health {
        proxy_pass http://dhcp2p;
        access_log off;
    }
}
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Failed

**Symptoms**: Application fails to start, database connection errors

**Solutions**:
```bash
# Check database connectivity
docker-compose exec dhcp2p ping postgres

# Check database logs
docker-compose logs postgres

# Verify connection string
echo $DATABASE_URL
```

#### 2. Redis Connection Failed

**Symptoms**: Cache errors, nonce storage failures

**Solutions**:
```bash
# Check Redis connectivity
docker-compose exec dhcp2p ping redis

# Check Redis logs
docker-compose logs redis

# Test Redis connection
docker-compose exec redis redis-cli ping
```

#### 3. High Memory Usage

**Symptoms**: Application crashes, OOM errors

**Solutions**:
```bash
# Monitor memory usage
docker stats

# Check for memory leaks
docker-compose exec dhcp2p ps aux

# Adjust resource limits
# In docker-compose.yml:
deploy:
  resources:
    limits:
      memory: 2G
```

#### 4. Slow Response Times

**Symptoms**: High latency, timeout errors

**Solutions**:
```bash
# Check database performance
docker-compose exec postgres psql -U dhcp2p -c "SELECT * FROM pg_stat_activity;"

# Check Redis performance
docker-compose exec redis redis-cli --latency

# Monitor application logs
docker-compose logs dhcp2p | grep -i "slow\|timeout\|error"
```

### Debug Commands

```bash
# Get application logs
make docker-logs

# Check container status
make docker-ps

# Get shell access
docker-compose exec dhcp2p sh

# Check configuration
docker-compose exec dhcp2p env | grep DHCP2P

# Test health endpoints
curl -v http://localhost:8088/health
curl -v http://localhost:8088/ready
```

### Performance Tuning

#### Database Tuning

```sql
-- PostgreSQL configuration
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
```

#### Redis Tuning

```bash
# Redis configuration
redis-server --maxmemory 512mb --maxmemory-policy allkeys-lru
```

#### Application Tuning

```yaml
# config/performance.yaml
redis_pool_size: 20
redis_min_idle_conns: 10
max_lease_retries: 5
lease_retry_delay: 1000
```

This deployment guide ensures DHCP2P can be deployed reliably across different environments with proper monitoring, backup, and scaling strategies.
