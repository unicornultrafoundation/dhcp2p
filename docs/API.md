# DHCP2P API Reference

This document provides comprehensive API documentation for the DHCP2P service, including authentication flow, endpoints, request/response schemas, and examples.

## Table of Contents

- [Authentication](#authentication)
- [Base URL](#base-url)
- [Error Handling](#error-handling)
- [Endpoints](#endpoints)
  - [Authentication Endpoints](#authentication-endpoints)
  - [Lease Management Endpoints](#lease-management-endpoints)
  - [Health Check Endpoints](#health-check-endpoints)
- [Data Models](#data-models)
- [Examples](#examples)

## Authentication

DHCP2P uses libp2p cryptographic signatures for authentication. The authentication flow is nonce-based to prevent replay attacks.

### Authentication Flow

1. **Request Nonce**: Send a POST request to `/request-auth` with your public key
2. **Receive Nonce**: Server returns a time-limited nonce (default: 5 minutes)
3. **Sign Nonce**: Sign the nonce with your private key using libp2p crypto
4. **Include Signature**: Include the signature in the `Authorization` header for protected endpoints
5. **Server Verification**: Server verifies the signature using your public key

### Authorization Header Format

```
Authorization: Bearer <base64-encoded-signature>
```

The signature should be the raw bytes of the libp2p signature, base64-encoded.

## Base URL

- **Development**: `http://localhost:8088`
- **Production**: Configure according to your deployment

## Error Handling

All API endpoints return consistent error responses:

```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "details": "Additional error details (optional)"
}
```

### Common HTTP Status Codes

- `200 OK` - Request successful
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Authentication required or invalid
- `403 Forbidden` - Valid authentication but insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists or conflict
- `500 Internal Server Error` - Server error

## Endpoints

### Authentication Endpoints

#### Request Authentication Nonce

**POST** `/request-auth`

Request a nonce for authentication.

**Request Body:**
```json
{
  "pubkey": "base64-encoded-public-key"
}
```

**Response:**
```json
{
  "pubkey": "base64-encoded-public-key",
  "nonce": "nonce-id-uuid"
}
```

**Example:**
```bash
curl -X POST http://localhost:8088/request-auth \
  -H "Content-Type: application/json" \
  -d '{
    "pubkey": "CAESIK...base64-encoded-public-key"
  }'
```

**Response:**
```json
{
  "pubkey": "CAESIK...base64-encoded-public-key",
  "nonce": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Lease Management Endpoints

#### Allocate IP Lease

**POST** `/allocate-ip`

Allocate a new IP lease for a peer. This endpoint is protected and requires authentication.

**Request Body:**
```json
{
  "peer_id": "peer-id-string"
}
```

**Response:**
```json
{
  "token_id": 12345,
  "peer_id": "peer-id-string",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "expires_at": "2024-01-15T12:30:00Z",
  "ttl": 120
}
```

**Example:**
```bash
curl -X POST http://localhost:8088/allocate-ip \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer base64-encoded-signature" \
  -d '{
    "peer_id": "12D3KooWExamplePeerID"
  }'
```

#### Renew Lease

**POST** `/renew-lease`

Renew an existing lease. This endpoint is protected and requires authentication.

**Request Body:**
```json
{
  "token_id": 12345,
  "peer_id": "peer-id-string"
}
```

**Response:**
```json
{
  "token_id": 12345,
  "peer_id": "peer-id-string",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T11:30:00Z",
  "expires_at": "2024-01-15T13:30:00Z",
  "ttl": 120
}
```

**Example:**
```bash
curl -X POST http://localhost:8088/renew-lease \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer base64-encoded-signature" \
  -d '{
    "token_id": 12345,
    "peer_id": "12D3KooWExamplePeerID"
  }'
```

#### Release Lease

**POST** `/release-lease`

Release an existing lease. This endpoint is protected and requires authentication.

**Request Body:**
```json
{
  "token_id": 12345,
  "peer_id": "peer-id-string"
}
```

**Response:**
```json
{
  "status": "success"
}
```

**Example:**
```bash
curl -X POST http://localhost:8088/release-lease \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer base64-encoded-signature" \
  -d '{
    "token_id": 12345,
    "peer_id": "12D3KooWExamplePeerID"
  }'
```

#### Get Lease by Peer ID

**GET** `/lease/peer-id/{peerID}`

Retrieve lease information by peer ID. This endpoint is public and does not require authentication.

**Path Parameters:**
- `peerID` (string): The peer ID to look up

**Response:**
```json
{
  "token_id": 12345,
  "peer_id": "12D3KooWExamplePeerID",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T11:30:00Z",
  "expires_at": "2024-01-15T13:30:00Z",
  "ttl": 120
}
```

**Example:**
```bash
curl http://localhost:8088/lease/peer-id/12D3KooWExamplePeerID
```

#### Get Lease by Token ID

**GET** `/lease/token-id/{tokenID}`

Retrieve lease information by token ID. This endpoint is public and does not require authentication.

**Path Parameters:**
- `tokenID` (integer): The token ID to look up

**Response:**
```json
{
  "token_id": 12345,
  "peer_id": "12D3KooWExamplePeerID",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T11:30:00Z",
  "expires_at": "2024-01-15T13:30:00Z",
  "ttl": 120
}
```

**Example:**
```bash
curl http://localhost:8088/lease/token-id/12345
```

### Health Check Endpoints

#### Health Check

**GET** `/health`

Check if the service is running and healthy.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

**Example:**
```bash
curl http://localhost:8088/health
```

#### Readiness Check

**GET** `/ready`

Check if the service is ready to accept requests (dependencies are available).

**Response:**
```json
{
  "status": "ready",
  "timestamp": "2024-01-15T10:30:00Z",
  "dependencies": {
    "database": "connected",
    "redis": "connected"
  }
}
```

**Example:**
```bash
curl http://localhost:8088/ready
```

## Data Models

### Lease

Represents an IP lease allocation.

```json
{
  "token_id": 12345,           // Unique token ID (int64)
  "peer_id": "peer-id-string", // Peer identifier (string)
  "created_at": "2024-01-15T10:30:00Z", // Creation timestamp (ISO 8601)
  "updated_at": "2024-01-15T11:30:00Z", // Last update timestamp (ISO 8601)
  "expires_at": "2024-01-15T13:30:00Z", // Expiration timestamp (ISO 8601)
  "ttl": 120                   // Time to live in minutes (int32)
}
```

### AuthRequest

Request for authentication nonce.

```json
{
  "pubkey": "base64-encoded-public-key" // libp2p public key (base64)
}
```

### AuthResponse

Response containing authentication nonce.

```json
{
  "pubkey": "base64-encoded-public-key", // libp2p public key (base64)
  "nonce": "nonce-id-uuid"               // Nonce identifier (UUID string)
}
```

### Error Response

Standard error response format.

```json
{
  "error": "error_code",        // Machine-readable error code
  "message": "Error message",   // Human-readable error message
  "details": "Additional info"  // Optional additional details
}
```

## Examples

### Complete Authentication and Lease Allocation Flow

```bash
# Step 1: Request authentication nonce
NONCE_RESPONSE=$(curl -s -X POST http://localhost:8088/request-auth \
  -H "Content-Type: application/json" \
  -d '{"pubkey": "CAESIK...your-public-key"}')

# Extract nonce ID
NONCE_ID=$(echo $NONCE_RESPONSE | jq -r '.nonce')

# Step 2: Sign the nonce with your private key (pseudo-code)
# SIGNATURE=$(sign_with_libp2p_private_key $NONCE_ID $PRIVATE_KEY)

# Step 3: Allocate IP lease
curl -X POST http://localhost:8088/allocate-ip \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SIGNATURE" \
  -d '{"peer_id": "12D3KooWYourPeerID"}'
```

### Error Handling Example

```bash
# Request with invalid peer ID
curl -X POST http://localhost:8088/allocate-ip \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer invalid-signature" \
  -d '{"peer_id": ""}'
```

**Error Response:**
```json
{
  "error": "validation_error",
  "message": "Peer ID cannot be empty",
  "details": "Field 'peer_id' is required and must not be empty"
}
```

### Rate Limiting

The API implements rate limiting (1000 requests per second by default). When rate limited:

**Response:**
```json
{
  "error": "rate_limit_exceeded",
  "message": "Too many requests",
  "details": "Rate limit exceeded. Please try again later."
}
```

**HTTP Status:** `429 Too Many Requests`

## Middleware

The API includes several middleware components:

- **Security Middleware**: CORS, security headers, rate limiting
- **Authentication Middleware**: libp2p signature verification
- **Logging Middleware**: Request/response logging
- **Recovery Middleware**: Panic recovery
- **Timeout Middleware**: Request timeout (60 seconds)

## SDK and Client Libraries

Currently, clients need to implement libp2p signature verification manually. Future versions may include:

- Go client library
- JavaScript/TypeScript client library
- Python client library

## Versioning

The API uses semantic versioning. Current version: `v1.0.0`

- **Major version changes**: Breaking changes to the API
- **Minor version changes**: New features, backward compatible
- **Patch version changes**: Bug fixes, backward compatible

Version information is available in the health check response.
