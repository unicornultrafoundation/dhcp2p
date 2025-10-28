# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

# Install git and ca-certificates for fetching dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH \
    go build -ldflags="-w -s -X main.Build=${BUILD_VERSION:-dev}" \
    -o dhcp2p ./cmd/dhcp2p

# Final stage - Alpine with Atlas CLI and application
FROM alpine:latest

# Install Atlas CLI and other dependencies
RUN apk add --no-cache \
    curl \
    postgresql-client \
    ca-certificates \
    tzdata

# Install Atlas CLI
RUN curl -sSfL https://atlasgo.sh | sh -s -- -b /usr/local/bin

# Copy the binary
COPY --from=builder /app/dhcp2p /dhcp2p

# Copy migration files
COPY --from=builder /app/internal/app/infrastructure/migrations /migrations

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy entrypoint script
COPY scripts/docker-entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Run as non-root user
RUN adduser -D -s /bin/sh dhcp2p

# Create directories and set permissions for dhcp2p user
RUN mkdir -p /home/dhcp2p/logs && \
    chown -R dhcp2p:dhcp2p /home/dhcp2p && \
    chown -R dhcp2p:dhcp2p /migrations && \
    chown dhcp2p:dhcp2p /dhcp2p && \
    chown dhcp2p:dhcp2p /entrypoint.sh

USER dhcp2p

# Prepare writable workspace (for logs)
WORKDIR /home/dhcp2p

# Expose port
EXPOSE 8088

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl --fail --silent http://localhost:8088/health || exit 1

# Use entrypoint script
ENTRYPOINT ["/entrypoint.sh"]
