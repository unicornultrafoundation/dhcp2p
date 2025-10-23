# DHCP2P Security Guide

This document provides comprehensive information about DHCP2P's security architecture, authentication mechanisms, and security best practices.

## Table of Contents

- [Security Overview](#security-overview)
- [Authentication Architecture](#authentication-architecture)
- [libp2p Signature Verification](#libp2p-signature-verification)
- [Nonce-based Security](#nonce-based-security)
- [Security Middleware](#security-middleware)
- [Data Protection](#data-protection)
- [Network Security](#network-security)
- [Production Security](#production-security)
- [Threat Model](#threat-model)
- [Security Best Practices](#security-best-practices)
- [Security Monitoring](#security-monitoring)
- [Incident Response](#incident-response)

## Security Overview

DHCP2P implements a multi-layered security architecture designed to protect against common threats in distributed systems:

- **Cryptographic Authentication**: libp2p signature-based authentication
- **Nonce-based Security**: Time-limited tokens prevent replay attacks
- **Rate Limiting**: Protection against DoS attacks
- **Input Validation**: Comprehensive request validation
- **Secure Communication**: TLS/SSL support for database connections
- **Access Control**: Peer-based authorization

## Authentication Architecture

### Authentication Flow

```
1. Client generates libp2p key pair (secp256k1)
   ↓
2. Client sends public key to /request-auth
   ↓
3. Server generates time-limited nonce
   ↓
4. Server stores nonce in Redis with TTL
   ↓
5. Server returns nonce to client
   ↓
6. Client signs nonce with private key
   ↓
7. Client includes signature in Authorization header
   ↓
8. Server verifies signature using public key
   ↓
9. Server extracts peer ID from signature
   ↓
10. Request proceeds to handler
```

### Key Components

- **libp2p Keys**: secp256k1 elliptic curve cryptography
- **Nonce Generation**: Cryptographically secure random UUIDs
- **Signature Verification**: ECDSA signature verification
- **Peer ID Extraction**: Derived from public key hash

## libp2p Signature Verification

### Key Generation

Clients generate libp2p key pairs using secp256k1 elliptic curve:

```go
// Generate new key pair
privKey, pubKey, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
if err != nil {
    return err
}

// Marshal public key for transmission
pubKeyBytes, err := crypto.MarshalPublicKey(pubKey)
if err != nil {
    return err
}
```

### Signature Creation

```go
// Sign nonce with private key
signature, err := privKey.Sign(nonceBytes)
if err != nil {
    return err
}

// Encode signature for transmission
signatureB64 := base64.StdEncoding.EncodeToString(signature)
```

### Signature Verification

```go
// Unmarshal public key
pubKey, err := crypto.UnmarshalPublicKey(pubKeyBytes)
if err != nil {
    return err
}

// Verify signature
ok, err := pubKey.Verify(nonceBytes, signature)
if err != nil {
    return err
}
if !ok {
    return errors.ErrInvalidSignature
}
```

### Security Properties

- **Cryptographic Strength**: secp256k1 provides 128-bit security
- **Non-repudiation**: Signatures prove key ownership
- **Integrity**: Signatures detect tampering
- **Authentication**: Signatures verify identity

## Nonce-based Security

### Nonce Generation

```go
// Generate cryptographically secure nonce
nonceID := uuid.New().String()

// Set expiration time
expiresAt := time.Now().Add(time.Duration(nonceTTL) * time.Minute)

// Store nonce with expiration
nonce := &models.Nonce{
    ID:        nonceID,
    PeerID:    peerID,
    IssuedAt:  time.Now(),
    ExpiresAt: expiresAt,
    Used:      false,
}
```

### Nonce Properties

- **Uniqueness**: UUID-based nonces are globally unique
- **Time Limitation**: Default 5-minute expiration
- **One-time Use**: Nonces marked as used after verification
- **Secure Storage**: Stored in Redis with TTL

### Replay Attack Prevention

1. **Time Window**: Nonces expire after 5 minutes
2. **Single Use**: Nonces can only be used once
3. **Cleanup**: Expired nonces are automatically removed
4. **Verification**: Server checks nonce validity before processing

## Security Middleware

### Combined Security Middleware

```go
func CombinedSecurityMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // CORS headers
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            
            // Security headers
            w.Header().Set("X-Content-Type-Options", "nosniff")
            w.Header().Set("X-Frame-Options", "DENY")
            w.Header().Set("X-XSS-Protection", "1; mode=block")
            w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
            
            // Handle preflight requests
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Authentication Middleware

```go
func WithAuth(authService ports.AuthService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract signature from Authorization header
            authHeader := r.Header.Get("Authorization")
            if !strings.HasPrefix(authHeader, "Bearer ") {
                http.Error(w, "Missing or invalid authorization header", http.StatusUnauthorized)
                return
            }
            
            signatureB64 := strings.TrimPrefix(authHeader, "Bearer ")
            signature, err := base64.StdEncoding.DecodeString(signatureB64)
            if err != nil {
                http.Error(w, "Invalid signature format", http.StatusUnauthorized)
                return
            }
            
            // Verify signature and extract peer ID
            peerID, err := authService.VerifyAuth(r.Context(), signature)
            if err != nil {
                http.Error(w, "Authentication failed", http.StatusUnauthorized)
                return
            }
            
            // Add peer ID to request context
            ctx := context.WithValue(r.Context(), "peer_id", peerID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Rate Limiting

```go
// Chi middleware for rate limiting
r.Use(middleware.Throttle(1000)) // 1000 requests per second
```

### Request Validation

```go
func ValidateLeaseRequest(r *http.Request) (interface{}, error) {
    var req LeaseRequestData
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return nil, fmt.Errorf("invalid JSON: %w", err)
    }
    
    if req.PeerID == "" {
        return nil, fmt.Errorf("peer_id is required")
    }
    
    if len(req.PeerID) > 128 {
        return nil, fmt.Errorf("peer_id too long")
    }
    
    return &req, nil
}
```

## Data Protection

### Database Security

- **Connection Encryption**: SSL/TLS for database connections
- **Authentication**: Strong database credentials
- **Access Control**: Database user with minimal privileges
- **Query Protection**: Parameterized queries prevent SQL injection

### Redis Security

- **Authentication**: Password protection for Redis
- **Network Isolation**: Redis accessible only from application
- **Data Encryption**: Sensitive data encrypted in transit
- **Access Control**: Redis user with minimal privileges

### Data Encryption

- **In Transit**: TLS/SSL for all external communications
- **At Rest**: Database and Redis encryption
- **Sensitive Data**: Nonces and signatures handled securely

## Network Security

### TLS/SSL Configuration

```yaml
# Database connection with SSL
database_url: "postgres://user:pass@host:5432/db?sslmode=require"

# Redis connection with TLS
redis_url: "rediss://user:pass@host:6380"
```

### Firewall Configuration

```bash
# Allow only necessary ports
# Application port
iptables -A INPUT -p tcp --dport 8088 -j ACCEPT

# Database port (internal only)
iptables -A INPUT -p tcp --dport 5432 -s 10.0.0.0/8 -j ACCEPT

# Redis port (internal only)
iptables -A INPUT -p tcp --dport 6379 -s 10.0.0.0/8 -j ACCEPT
```

### Network Isolation

- **Internal Networks**: Database and Redis on private networks
- **Load Balancer**: Public access through load balancer only
- **VPN Access**: Administrative access through VPN

## Production Security

### Security Checklist

- [ ] **Strong Passwords**: Use strong, unique passwords for all services
- [ ] **SSL/TLS**: Enable SSL for all external connections
- [ ] **Firewall**: Configure appropriate firewall rules
- [ ] **Updates**: Keep all software updated
- [ ] **Monitoring**: Implement security monitoring
- [ ] **Backups**: Encrypt and secure backups
- [ ] **Access Control**: Implement least privilege access
- [ ] **Logging**: Enable comprehensive security logging

### Secrets Management

```bash
# Use Docker secrets
echo "strong_password" | docker secret create db_password -
echo "redis_password" | docker secret create redis_password -

# Use environment variables
export DATABASE_URL="postgres://user:$(cat /run/secrets/db_password)@host:5432/db"
export REDIS_URL="redis://:$(cat /run/secrets/redis_password)@host:6379"
```

### Container Security

```dockerfile
# Use non-root user
RUN adduser -D -s /bin/sh dhcp2p
USER dhcp2p

# Use minimal base image
FROM alpine:latest

# Remove unnecessary packages
RUN apk del build-dependencies
```

### Kubernetes Security

```yaml
# Security context
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop:
      - ALL
```

## Threat Model

### Identified Threats

#### 1. Replay Attacks
- **Threat**: Attacker replays valid requests
- **Mitigation**: Nonce expiration and one-time use
- **Risk Level**: Low

#### 2. Man-in-the-Middle Attacks
- **Threat**: Attacker intercepts and modifies communications
- **Mitigation**: TLS encryption and signature verification
- **Risk Level**: Low

#### 3. Denial of Service (DoS)
- **Threat**: Attacker overwhelms service with requests
- **Mitigation**: Rate limiting and resource management
- **Risk Level**: Medium

#### 4. Data Tampering
- **Threat**: Attacker modifies data in transit or storage
- **Mitigation**: Cryptographic signatures and TLS
- **Risk Level**: Low

#### 5. Unauthorized Access
- **Threat**: Attacker gains access without proper authentication
- **Mitigation**: libp2p signature verification
- **Risk Level**: Low

#### 6. Information Disclosure
- **Threat**: Sensitive information exposed
- **Mitigation**: Access control and data encryption
- **Risk Level**: Medium

### Attack Vectors

#### Network Attacks
- **Port Scanning**: Firewall rules prevent unauthorized access
- **DDoS**: Rate limiting and load balancing mitigate impact
- **Packet Sniffing**: TLS encryption protects data in transit

#### Application Attacks
- **SQL Injection**: Parameterized queries prevent injection
- **XSS**: Input validation and output encoding
- **CSRF**: Same-origin policy and token validation

#### Infrastructure Attacks
- **Container Escape**: Non-root user and minimal privileges
- **Privilege Escalation**: Least privilege access model
- **Data Exfiltration**: Network isolation and access control

## Security Best Practices

### Development Security

1. **Secure Coding**: Follow secure coding practices
2. **Input Validation**: Validate all inputs
3. **Error Handling**: Don't expose sensitive information
4. **Dependency Management**: Keep dependencies updated
5. **Code Review**: Security-focused code reviews

### Operational Security

1. **Monitoring**: Monitor for security events
2. **Logging**: Log security-relevant events
3. **Incident Response**: Have incident response procedures
4. **Backup Security**: Encrypt and secure backups
5. **Access Management**: Implement least privilege access

### Configuration Security

1. **Default Settings**: Change default passwords and settings
2. **Minimal Configuration**: Enable only necessary features
3. **Regular Updates**: Keep configurations updated
4. **Documentation**: Document security configurations
5. **Testing**: Test security configurations

## Security Monitoring

### Log Analysis

```json
{
  "level": "warn",
  "timestamp": "2024-01-15T10:30:00Z",
  "caller": "middleware/auth.go:45",
  "msg": "authentication failed",
  "peer_id": "12D3KooWAttacker",
  "error": "invalid signature",
  "ip_address": "192.168.1.100"
}
```

### Security Metrics

- **Authentication Failures**: Monitor failed authentication attempts
- **Rate Limit Violations**: Track rate limit violations
- **Error Rates**: Monitor error rates for anomalies
- **Access Patterns**: Analyze access patterns for anomalies

### Alerting

```yaml
# Prometheus alerting rules
groups:
  - name: dhcp2p_security
    rules:
      - alert: HighAuthenticationFailures
        expr: rate(authentication_failures_total[5m]) > 10
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "High authentication failure rate"
          
      - alert: RateLimitViolations
        expr: rate(rate_limit_violations_total[5m]) > 5
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "High rate limit violation rate"
```

## Incident Response

### Incident Classification

#### Low Severity
- Single authentication failure
- Minor configuration issues
- Non-critical security warnings

#### Medium Severity
- Multiple authentication failures
- Rate limit violations
- Suspicious access patterns

#### High Severity
- Successful unauthorized access
- Data breach indicators
- System compromise

### Response Procedures

1. **Detection**: Identify security incident
2. **Assessment**: Assess severity and impact
3. **Containment**: Contain the incident
4. **Investigation**: Investigate root cause
5. **Recovery**: Restore normal operations
6. **Documentation**: Document incident and response
7. **Prevention**: Implement preventive measures

### Emergency Contacts

- **Security Team**: security@company.com
- **On-call Engineer**: +1-555-0123
- **Management**: management@company.com

### Recovery Procedures

```bash
# Emergency shutdown
docker-compose down

# Isolate affected systems
iptables -A INPUT -p tcp --dport 8088 -j DROP

# Restore from backup
docker-compose exec postgres psql -U dhcp2p dhcp2p < backup.sql

# Restart services
docker-compose up -d
```

This security guide ensures DHCP2P is deployed and operated securely, protecting against common threats and providing guidance for security best practices.
