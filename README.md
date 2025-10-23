# DHCP2P

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/duchuongnguyen/dhcp2p)
[![Go Version](https://img.shields.io/badge/go-1.25+-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Coverage](https://img.shields.io/badge/coverage-80%25-yellow)](tests/)

A decentralized IP lease management service that provides token-based IP allocation with libp2p authentication. DHCP2P enables secure, peer-to-peer IP address management through cryptographic signatures and nonce-based authentication.

## 🚀 Key Features

- **Token-based IP Leases**: Allocate unique token IDs for IP address management
- **libp2p Authentication**: Secure peer-to-peer authentication using cryptographic signatures
- **Nonce-based Security**: Time-limited nonces prevent replay attacks
- **Redis Caching**: High-performance caching for nonces and lease data
- **PostgreSQL Persistence**: Reliable data storage with ACID compliance
- **Clean Architecture**: Hexagonal architecture with dependency injection
- **Docker Ready**: Complete containerization with Docker Compose
- **Comprehensive Testing**: Unit, integration, and end-to-end test suites

## 🏗️ Technology Stack

- **Backend**: Go 1.25+ with Uber Fx dependency injection
- **Database**: PostgreSQL with Atlas migrations
- **Cache**: Redis for nonce and lease caching
- **Authentication**: libp2p cryptographic signatures
- **HTTP Framework**: Chi router with middleware
- **Testing**: Testcontainers for integration tests
- **Containerization**: Docker & Docker Compose

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Make (optional, for convenience commands)

### Run with Docker

```bash
# Clone the repository
git clone https://github.com/duchuongnguyen/dhcp2p.git
cd dhcp2p

# Start the application stack
make docker-up

# Check health
curl http://localhost:8088/health
```

The application will be available at:
- **API**: http://localhost:8088
- **Health Check**: http://localhost:8088/health
- **Readiness Check**: http://localhost:8088/ready

### Environment Setup

```bash
# Interactive setup
make setup

# Or manually create .env file
cp .env.example .env
# Edit .env with your configuration
```

## 📚 Documentation

- **[API Reference](docs/API.md)** - Complete API documentation with examples
- **[Architecture Guide](docs/ARCHITECTURE.md)** - System design and clean architecture
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Production deployment and configuration
- **[Development Guide](docs/DEVELOPMENT.md)** - Local development setup and workflow
- **[Configuration Reference](docs/CONFIGURATION.md)** - All configuration options
- **[Security Guide](docs/SECURITY.md)** - Authentication and security practices
- **[Contributing Guide](docs/CONTRIBUTING.md)** - How to contribute to the project
- **[Docker Deployment](docker/README.md)** - Comprehensive Docker deployment guide
- **[Testing Guide](tests/README.md)** - Testing setup and best practices

## 🔧 Common Commands

```bash
# Development
make docker-up          # Start development stack
make docker-down        # Stop development stack
make docker-logs        # View application logs
make docker-health      # Check application health

# Testing
make test              # Run all tests
make test-unit         # Run unit tests only
make test-integration  # Run integration tests
make test-e2e         # Run end-to-end tests
make test-coverage    # Generate coverage report

# Database
make migrate           # Run database migrations
make sqlc             # Generate SQL code
make db               # Run migrations + generate code

# Building
make docker-build     # Build application image
make docker-push      # Push image to registry
```

## 🏛️ Architecture Overview

DHCP2P follows Clean Architecture principles with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Handlers                            │
├─────────────────────────────────────────────────────────────┤
│                 Application Services                        │
├─────────────────────────────────────────────────────────────┤
│                   Domain Models                             │
├─────────────────────────────────────────────────────────────┤
│              Repository Adapters (Postgres/Redis)          │
└─────────────────────────────────────────────────────────────┘
```

## 🔐 Authentication Flow

1. **Request Nonce**: Client sends public key to `/request-auth`
2. **Receive Nonce**: Server returns a time-limited nonce
3. **Sign Nonce**: Client signs the nonce with their private key
4. **Authenticate**: Client includes signature in subsequent requests
5. **Verify**: Server verifies signature using libp2p crypto

## 📊 API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/request-auth` | Request authentication nonce | No |
| POST | `/allocate-ip` | Allocate new IP lease | Yes |
| POST | `/renew-lease` | Renew existing lease | Yes |
| POST | `/release-lease` | Release lease | Yes |
| GET | `/lease/peer-id/{peerID}` | Get lease by peer ID | No |
| GET | `/lease/token-id/{tokenID}` | Get lease by token ID | No |
| GET | `/health` | Health check | No |
| GET | `/ready` | Readiness check | No |

## 🗄️ Database Schema

- **leases**: Token-based IP lease records
- **nonces**: Authentication nonces with expiration
- **alloc_state**: Allocation state tracking

## 🧪 Testing

The project includes comprehensive testing:

- **Unit Tests**: Business logic with mocked dependencies
- **Integration Tests**: Real database and Redis using testcontainers
- **End-to-End Tests**: Complete API workflows
- **Benchmark Tests**: Performance testing
- **Contract Tests**: API contract validation

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/CONTRIBUTING.md) for details on:

- Development setup
- Code standards
- Pull request process
- Issue reporting

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📖 Check the [documentation](docs/)
- 🐛 Report issues on [GitHub Issues](https://github.com/duchuongnguyen/dhcp2p/issues)
- 💬 Join discussions in [GitHub Discussions](https://github.com/duchuongnguyen/dhcp2p/discussions)

## 🔗 Related Projects

- [libp2p](https://libp2p.io/) - Peer-to-peer networking stack
- [Atlas](https://atlasgo.io/) - Database schema management
- [Chi](https://github.com/go-chi/chi) - HTTP router for Go
- [Uber Fx](https://uber-go.github.io/fx/) - Dependency injection framework
