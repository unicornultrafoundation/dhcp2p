#!/bin/bash

# DHCP2P Setup Script
# This script helps initialize the dhcp2p application

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
CONFIG_DIR="./config"
ENV_FILE="./.env"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -c, --config-dir DIR    Set config directory (default: ./config)"
    echo "  -e, --env-file FILE     Set environment file (default: ./.env)"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -c|--config-dir)
            CONFIG_DIR="$2"
            shift 2
            ;;
        -e|--env-file)
            ENV_FILE="$2"
            shift 2
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

print_info "DHCP2P Setup Script"
print_info "==================="

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed or not in PATH"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed or not in PATH"
    exit 1
fi

# Create necessary directories
print_info "Creating directories..."
mkdir -p "$CONFIG_DIR"

print_success "Directories created: $CONFIG_DIR"

# Check if .env file already exists
if [[ -f "$ENV_FILE" ]]; then
    print_warning "Environment file $ENV_FILE already exists"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Keeping existing environment file"
        exit 0
    fi
fi

# Copy .env.example to .env if it doesn't exist
if [[ ! -f "$ENV_FILE" ]]; then
    if [[ -f ".env.example" ]]; then
        cp .env.example "$ENV_FILE"
        print_success "Created $ENV_FILE from .env.example"
    else
        print_error ".env.example not found. Please create it first."
        exit 1
    fi
fi

# Display configuration instructions
echo ""
print_info "Configuration Required:"
echo "Please update the following in $ENV_FILE:"
echo ""
echo "1. Database Configuration:"
echo "   - DATABASE_URL: PostgreSQL connection string"
echo "   - POSTGRES_DB, POSTGRES_USER, POSTGRES_PASSWORD"
echo ""
echo "2. Redis Configuration:"
echo "   - REDIS_URL: Redis connection string"
echo "   - REDIS_PASSWORD (if required)"
echo ""
echo "3. Application Configuration:"
echo "   - PORT: Application port (default: 8088)"
echo "   - LOG_LEVEL: Logging level (debug/info/warn/error)"
echo "   - NONCE_TTL, LEASE_TTL: Timeout configurations"
echo ""

# Prompt for basic configuration
read -p "Do you want to configure basic settings now? (y/N): " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Prompt for database configuration
    echo ""
    print_info "Database Configuration:"
    read -p "PostgreSQL Database Name [dhcp2p]: " POSTGRES_DB
    POSTGRES_DB=${POSTGRES_DB:-dhcp2p}
    
    read -p "PostgreSQL Username [dhcp2p]: " POSTGRES_USER
    POSTGRES_USER=${POSTGRES_USER:-dhcp2p}
    
    read -s -p "PostgreSQL Password: " POSTGRES_PASSWORD
    echo
    
    # Prompt for Redis configuration
    echo ""
    print_info "Redis Configuration:"
    read -p "Redis Password (optional): " REDIS_PASSWORD
    
    # Prompt for application configuration
    echo ""
    print_info "Application Configuration:"
    read -p "Application Port [8088]: " PORT
    PORT=${PORT:-8088}
    
    read -p "Log Level [debug]: " LOG_LEVEL
    LOG_LEVEL=${LOG_LEVEL:-debug}
    
    # Update .env file
    print_info "Updating environment file..."
    
    # Use sed to update the .env file
    sed -i.bak "s/^POSTGRES_DB=.*/POSTGRES_DB=$POSTGRES_DB/" "$ENV_FILE"
    sed -i.bak "s/^POSTGRES_USER=.*/POSTGRES_USER=$POSTGRES_USER/" "$ENV_FILE"
    sed -i.bak "s/^POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=$POSTGRES_PASSWORD/" "$ENV_FILE"
    sed -i.bak "s/^DATABASE_URL=.*/DATABASE_URL=postgres:\/\/$POSTGRES_USER:$POSTGRES_PASSWORD@postgres:5432\/$POSTGRES_DB?sslmode=disable/" "$ENV_FILE"
    
    # Redis URL should always be just host:port format (no scheme or password)
    sed -i.bak "s/^REDIS_URL=.*/REDIS_URL=redis:6379/" "$ENV_FILE"
    
    if [[ -n "$REDIS_PASSWORD" ]]; then
        sed -i.bak "s/^REDIS_PASSWORD=.*/REDIS_PASSWORD=$REDIS_PASSWORD/" "$ENV_FILE"
    fi
    
    sed -i.bak "s/^PORT=.*/PORT=$PORT/" "$ENV_FILE"
    sed -i.bak "s/^LOG_LEVEL=.*/LOG_LEVEL=$LOG_LEVEL/" "$ENV_FILE"
    
    # Clean up backup file
    rm -f "$ENV_FILE.bak"
    
    print_success "Environment file updated with your configuration"
fi

print_success "Environment file ready: $ENV_FILE"

# Display next steps
echo ""
print_success "Setup completed successfully!"
echo ""
print_info "Next steps:"
echo "1. Review and update $ENV_FILE if needed"
echo "2. Start the application:"
echo "   Development: make docker-up"
echo "   Production:  make docker-up-prod"
echo ""
print_info "Useful commands:"
echo "  make docker-logs     # View application logs"
echo "  make docker-health   # Check health status"
echo "  make docker-down     # Stop the application"
echo ""
print_warning "Important:"
echo "- Use strong passwords in production"
echo "- Consider using external databases for production"