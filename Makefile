MIGRATION_DIR := internal/app/infrastructure/migrations
SCHEMA_FILE := $(MIGRATION_DIR)/schema.hcl

# ---- Deployment / Docker variables ----
COMPOSE := docker compose
ENV_FILE ?= .env
PROD_ENV_FILE ?= .env.prod

IMAGE_NAME ?= dhcp2p
TAG ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo latest)
# Override IMAGE to include registry if needed, e.g. make IMAGE=ghcr.io/owner/dhcp2p:$(TAG)
IMAGE ?= $(IMAGE_NAME):$(TAG)
MIGRATE_IMAGE ?= $(IMAGE_NAME)-migrate:$(TAG)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  help                 Show this help"
	@echo "  hash                 Generate Atlas migration hash"
	@echo "  diff name=NAME       Create Atlas migration diff"
	@echo "  migrate              Apply migrations locally with Atlas (uses DB_URL)"
	@echo "  sqlc                 Generate sqlc code"
	@echo "  db                   Run migrate + sqlc"
	@echo "  setup                Run interactive project setup (.env, config)"
	@echo "  docker-build         Build app image (Dockerfile) -> $(IMAGE)"
	@echo "  docker-build-migrate Build migration image (Dockerfile.migrate) -> $(MIGRATE_IMAGE)"
	@echo "  docker-push          Push app image (requires IMAGE set with registry)"
	@echo "  docker-up            Start dev stack (docker-compose.yml)"
	@echo "  docker-down          Stop dev stack and remove volumes"
	@echo "  docker-logs          Follow app logs"
	@echo "  docker-ps            Show compose services"
	@echo "  docker-health        Check app health endpoint"
	@echo "  docker-up-prod       Start prod stack (-f docker-compose.prod.yml)"
	@echo "  docker-down-prod     Stop prod stack and remove volumes"
	@echo "  migrate-docker       Run DB migrations in container"
	@echo "  migrate-status       Show migration status"
	@echo "  test                 Run all tests"
	@echo "  test-unit            Run unit tests only"
	@echo "  test-integration     Run integration tests only"
	@echo "  test-e2e             Run end-to-end tests only"
	@echo "  test-coverage        Run tests with coverage report"
	@echo "  test-mocks           Generate mocks for testing"

hash:
	atlas migrate hash --dir "file://$(MIGRATION_DIR)"

diff:
	atlas migrate diff $(name) \
	  --to "file://$(SCHEMA_FILE)" \
	  --dir "file://$(MIGRATION_DIR)" \
	  --dev-url "docker://postgres/15"

migrate:
	atlas migrate apply \
	  --dir "file://$(MIGRATION_DIR)" \
	  --url "$(DB_URL)"

sqlc:
	sqlc generate

db: migrate sqlc

.PHONY: setup
setup:
	bash scripts/setup.sh -e $(ENV_FILE)

# ---- Docker build/push ----
.PHONY: docker-build docker-build-migrate docker-push docker-login
docker-build:
	docker build -f Dockerfile -t $(IMAGE) --build-arg BUILD_VERSION=$(TAG) .

docker-build-migrate:
	docker build -f Dockerfile.migrate -t $(MIGRATE_IMAGE) --build-arg BUILD_VERSION=$(TAG) .

docker-push:
	@if [ "$(findstring /,$(IMAGE))" = "" ]; then \
		echo "ERROR: IMAGE must include registry (e.g. ghcr.io/owner/$(IMAGE_NAME):$(TAG))"; \
		exit 1; \
	fi
	docker push $(IMAGE)

# ---- Docker Compose (dev) ----
.PHONY: docker-up docker-down docker-logs docker-ps docker-health
docker-up:
	$(COMPOSE) --env-file $(ENV_FILE) up -d --build

docker-down:
	$(COMPOSE) --env-file $(ENV_FILE) down -v

docker-logs:
	$(COMPOSE) --env-file $(ENV_FILE) logs -f dhcp2p

docker-ps:
	$(COMPOSE) --env-file $(ENV_FILE) ps

docker-health:
	curl -fsS http://localhost:8088/health || (echo "health check failed" && exit 1)

docker-readiness:
	curl -fsS http://localhost:8088/ready || (echo "readiness check failed" && exit 1)

# ---- Docker Compose (prod) ----
.PHONY: docker-up-prod docker-down-prod
docker-up-prod:
	$(COMPOSE) -f docker-compose.yml -f docker-compose.prod.yml --env-file $(PROD_ENV_FILE) up -d --build

docker-down-prod:
	$(COMPOSE) -f docker-compose.yml -f docker-compose.prod.yml --env-file $(PROD_ENV_FILE) down -v

# ---- Migrations via container/script ----
.PHONY: migrate-docker migrate-status
migrate-docker:
	bash scripts/migrate.sh --auto -e $(ENV_FILE)

migrate-status:
	bash scripts/migrate.sh --status -e $(ENV_FILE)

# ---- Testing ----
.PHONY: test test-unit test-integration test-e2e test-coverage test-mocks test-bench test-load test-contract test-security

test: test-unit test-integration test-e2e

test-unit:
	go test -v ./tests/unit/...

test-integration:
	go test -v ./tests/integration/... -tags=integration

test-e2e:
	go test -v ./tests/e2e/... -tags=e2e

test-benchmark:
	@echo "Starting Redis for benchmarks..."
	@docker stop benchmark-redis 2>/dev/null || true
	@docker rm benchmark-redis 2>/dev/null || true
	@echo "Creating Redis container..."
	@docker run -d --name benchmark-redis -p 127.0.0.1:6380:6379 redis:7-alpine
	@docker ps --filter "name=benchmark-redis" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
	@sleep 5
	@echo "Running benchmarks..."
	go test -v -bench=. -benchmem ./tests/benchmark/... -tags=benchmark
	@echo "Cleaning up Redis container..."
	@docker stop benchmark-redis | true
	@docker rm benchmark-redis | true

test-load:
	go test -v ./tests/load/... -tags=load

test-contract:
	go test -v ./tests/contract/... -tags=contract

test-security:
	go test -v ./tests/unit/security_test.go ./tests/unit/errors_test.go

test-all: test test-security test-benchmark test-contract

test-coverage:
	go clean -testcache
	go test -count=1 -coverprofile=coverage-unit.out ./tests/unit/...
	go test -count=1 -coverprofile=coverage-integration.out ./tests/integration/...
	go test -count=1 -tags=contract -coverprofile=coverage-contract.out ./tests/contract/...
	go test -count=1 -tags=load -coverprofile=coverage-load.out ./tests/load/...
	go test -count=1 -tags=e2e -short -coverprofile=coverage-e2e.out ./tests/e2e/... || true
	go test -count=1 -coverprofile=coverage-internal.out ./internal/... || true
	# Combine coverage files properly
	echo "mode: atomic" > coverage.out
	grep -h -v "mode:" coverage-unit.out coverage-integration.out coverage-contract.out coverage-load.out coverage-e2e.out coverage-internal.out >> coverage.out 2>/dev/null || true
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-unit:
	go test -coverprofile=coverage-unit.out -covermode=atomic ./tests/unit/...
	go tool cover -html=coverage-unit.out -o coverage-unit.html
	@echo "Unit test coverage report generated: coverage-unit.html"

test-coverage-integration:
	go test -coverprofile=coverage-integration.out -covermode=atomic ./tests/integration/... -tags=integration
	go tool cover -html=coverage-integration.out -o coverage-integration.html
	@echo "Integration test coverage report generated: coverage-integration.html"

test-race:
	go test -race -v ./tests/unit/...

test-mocks:
	cd tests/mocks && go generate
	@echo "Mock generation completed"

test-fast:
	go test -short ./tests/unit/...

test-clean:
	@echo "Cleaning test artifacts..."
	rm -f coverage*.out coverage*.html
	rm -f test.log
	go clean -testcache