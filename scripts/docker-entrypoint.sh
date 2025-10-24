#!/bin/sh

# DHCP2P Docker Entrypoint Script
# This script handles the startup of the dhcp2p application in Docker

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Function to validate required environment variables
validate_environment() {
    print_info "Validating environment variables..."
    
    local missing_vars=""
    
    if [ -z "$DATABASE_URL" ]; then
        missing_vars="$missing_vars DATABASE_URL"
    fi
    
    if [ -z "$REDIS_URL" ]; then
        missing_vars="$missing_vars REDIS_URL"
    fi
    
    if [ -n "$missing_vars" ]; then
        print_error "Missing required environment variables:$missing_vars"
        print_error "Please set all required environment variables and try again"
        exit 1
    fi
    
    print_success "All required environment variables are set"
}

# Function to run migrations
run_migrations() {
    if [ "${RUN_MIGRATIONS:-true}" = "true" ]; then
        print_info "Running database migrations..."
        
        if [ -z "$DATABASE_URL" ]; then
            print_error "DATABASE_URL is required for migrations"
            exit 1
        fi
        
        # Check if Atlas CLI is available
        if command -v atlas >/dev/null 2>&1; then
            atlas migrate apply \
                --dir "file:///migrations" \
                --url "$DATABASE_URL"
            print_success "Migrations completed successfully"
        else
            print_error "Atlas CLI not found. Cannot run migrations."
            print_error "Please ensure migrations are run separately or Atlas CLI is installed."
            exit 1
        fi
    else
        print_warning "Skipping migrations (RUN_MIGRATIONS=false)"
    fi
}

# Function to start the application
start_application() {
    print_info "Starting dhcp2p application..."

    # Build argv safely (no word-splitting issues)
    set -- /dhcp2p serve \
        --database-url="$DATABASE_URL" \
        --redis-url="$REDIS_URL"

    # Add Redis password if set
    if [ -n "$REDIS_PASSWORD" ]; then
        set -- "$@" --redis-password="$REDIS_PASSWORD"
    fi

    # Add optional flags if set
    if [ -n "$PORT" ]; then
        set -- "$@" --port="$PORT"
    fi

    if [ -n "$LOG_LEVEL" ]; then
        set -- "$@" --log-level="$LOG_LEVEL"
    fi

    if [ -n "$NONCE_TTL" ]; then
        set -- "$@" --nonce-ttl="$NONCE_TTL"
    fi

    if [ -n "$NONCE_CLEANER_INTERVAL" ]; then
        set -- "$@" --nonce-cleaner-interval="$NONCE_CLEANER_INTERVAL"
    fi

    if [ -n "$LEASE_TTL" ]; then
        set -- "$@" --lease-ttl="$LEASE_TTL"
    fi

    if [ -n "$MAX_LEASE_RETRIES" ]; then
        set -- "$@" --max-lease-retries="$MAX_LEASE_RETRIES"
    fi

    if [ -n "$LEASE_RETRY_DELAY" ]; then
        set -- "$@" --lease-retry-delay="$LEASE_RETRY_DELAY"
    fi

    print_info "Starting with command: $*"

    # Execute the application
    exec "$@"
}

# Function to handle shutdown signals
setup_signal_handlers() {
    print_info "Setting up signal handlers for graceful shutdown..."
    
    # Function to handle shutdown
    shutdown_handler() {
        print_info "Received shutdown signal. Gracefully shutting down..."
        # The application should handle SIGTERM properly
        exit 0
    }
    
    # Set up signal handlers
    trap shutdown_handler SIGTERM SIGINT
}

# Main execution
main() {
    print_info "DHCP2P Docker Entrypoint"
    print_info "========================"
    
    # Set up signal handlers
    setup_signal_handlers
    
    # Validate environment
    validate_environment
    
    # Run migrations if enabled
    run_migrations
    
    # Start the application
    start_application
}

# Run main function
main "$@"
