#!/bin/bash

# DHCP2P Migration Script
# This script helps run database migrations manually

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENV_FILE="./.env"
MIGRATION_MODE="auto"

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
    echo "  -e, --env-file FILE     Set environment file (default: ./.env)"
    echo "  --auto                  Run migrations in Docker container (default)"
    echo "  --manual                Run migrations locally with Atlas CLI"
    echo "  --status                Check migration status"
    echo ""
    echo "Examples:"
    echo "  $0                      # Run migrations in Docker"
    echo "  $0 --manual             # Run migrations locally"
    echo "  $0 --status             # Check migration status"
}

# Function to load environment variables
load_env() {
    if [[ -f "$ENV_FILE" ]]; then
        print_info "Loading environment from $ENV_FILE"
        set -a
        source "$ENV_FILE"
        set +a
    else
        print_error "Environment file not found: $ENV_FILE"
        exit 1
    fi
}

# Function to check prerequisites
check_prerequisites() {
    if [[ "$MIGRATION_MODE" == "manual" ]]; then
        if ! command -v atlas &> /dev/null; then
            print_error "Atlas CLI is not installed. Install it with: curl -sSf https://atlasgo.sh | sh"
            exit 1
        fi
    else
        if ! command -v docker &> /dev/null; then
            print_error "Docker is not installed or not in PATH"
            exit 1
        fi
    fi
}

# Function to run migrations in Docker
run_migrations_docker() {
    print_info "Running migrations in Docker container..."
    
    # Build migration image if it doesn't exist
    if ! docker image inspect dhcp2p-migrate:latest &> /dev/null; then
        print_info "Building migration image..."
        docker build -f Dockerfile.migrate -t dhcp2p-migrate:latest .
    fi
    
    # Run migrations
    docker run --rm \
        -e DATABASE_URL="$DATABASE_URL" \
        dhcp2p-migrate:latest
    
    print_success "Migrations completed successfully"
}

# Function to run migrations locally
run_migrations_local() {
    print_info "Running migrations locally with Atlas CLI..."
    
    if [[ -z "$DATABASE_URL" ]]; then
        print_error "DATABASE_URL is not set in environment"
        exit 1
    fi
    
    # Run Atlas migrations
    atlas migrate apply \
        --dir "file://internal/app/infrastructure/migrations" \
        --url "$DATABASE_URL"
    
    print_success "Migrations completed successfully"
}

# Function to check migration status
check_migration_status() {
    print_info "Checking migration status..."
    
    if [[ -z "$DATABASE_URL" ]]; then
        print_error "DATABASE_URL is not set in environment"
        exit 1
    fi
    
    if command -v atlas &> /dev/null; then
        atlas migrate status \
            --dir "file://internal/app/infrastructure/migrations" \
            --url "$DATABASE_URL"
    else
        print_warning "Atlas CLI not available locally. Use --auto to check status in Docker"
        docker run --rm \
            -e DATABASE_URL="$DATABASE_URL" \
            -v "$(pwd)/internal/app/infrastructure/migrations:/migrations" \
            alpine:latest \
            sh -c "apk add --no-cache curl && curl -sSf https://atlasgo.sh | sh && atlas migrate status --dir file:///migrations --url '$DATABASE_URL'"
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -e|--env-file)
            ENV_FILE="$2"
            shift 2
            ;;
        --auto)
            MIGRATION_MODE="auto"
            shift
            ;;
        --manual)
            MIGRATION_MODE="manual"
            shift
            ;;
        --status)
            MIGRATION_MODE="status"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

print_info "DHCP2P Migration Script"
print_info "======================="

# Load environment variables
load_env

# Check prerequisites
check_prerequisites

# Run appropriate migration command
case "$MIGRATION_MODE" in
    "auto")
        run_migrations_docker
        ;;
    "manual")
        run_migrations_local
        ;;
    "status")
        check_migration_status
        ;;
    *)
        print_error "Invalid migration mode: $MIGRATION_MODE"
        exit 1
        ;;
esac

print_success "Migration script completed"

