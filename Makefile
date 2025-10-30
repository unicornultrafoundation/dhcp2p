MIGRATION_DIR := internal/app/infrastructure/migrations
SCHEMA_FILE := $(MIGRATION_DIR)/schema.hcl

# ---- Deployment / Docker variables ----
COMPOSE := docker compose
ENV_FILE ?= .env
PROD_ENV_FILE ?= .env.prod

# Docker build variables
IMAGE_NAME ?= depinnode/subnet-dhcp2p
TAG ?= 1.0.1
REGISTRY ?= 
PLATFORM ?= linux/amd64
NO_CACHE ?= false
QUIET ?= false

# Build full image name
ifeq ($(REGISTRY),)
    FULL_IMAGE_NAME = $(IMAGE_NAME):$(TAG)
else
    FULL_IMAGE_NAME = $(REGISTRY)/$(IMAGE_NAME):$(TAG)
endif

# Override IMAGE to include registry if needed, e.g. make IMAGE=ghcr.io/owner/dhcp2p:$(TAG)
IMAGE ?= $(FULL_IMAGE_NAME)

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
	@echo "  docker-build         Build combined image (Dockerfile) -> $(IMAGE)"
	@echo "  docker-build-push    Build and push image to registry"
	@echo "  docker-push          Push image to registry (requires REGISTRY set)"
	@echo "  docker-tag-latest    Tag current image as latest"
	@echo "  docker-info          Show Docker build information"
	@echo ""
	@echo "Docker build variables:"
	@echo "  IMAGE_NAME           Image name (default: depinnode/subnet-dhcp2p)"
	@echo "  TAG                  Image tag/version (default: 1.0.1)"
	@echo "  REGISTRY             Container registry (e.g., ghcr.io/username)"
	@echo "  PLATFORM             Target platform (default: linux/amd64)"
	@echo "  NO_CACHE             Build without cache (default: false)"
	@echo "  QUIET                Suppress build output (default: false)"
	@echo ""
	@echo "Examples:"
	@echo "  make docker-build"
	@echo "  make docker-build-push REGISTRY=ghcr.io/username"
	@echo "  make docker-build PLATFORM=linux/amd64,linux/arm64"
	@echo "  make docker-build TAG=v2.0.0"
	@echo "  docker-up            Start dev stack (docker-compose.yml)"
	@echo "  docker-down          Stop dev stack and remove volumes"
	@echo "  docker-logs          Follow app logs"
	@echo "  docker-ps            Show compose services"
	@echo "  docker-health        Check app health endpoint"
	@echo "  docker-up-prod       Start prod stack (-f docker-compose.prod.yml)"
	@echo "  docker-down-prod     Stop prod stack and remove volumes"
	@echo "  migrate-docker       Run DB migrations in container"
	@echo "  migrate-status       Show migration status"
	@echo ""
	@echo "Kubernetes deployment:"
	@echo "  k8s-deploy           Deploy to Kubernetes (requires kubectl)"
	@echo "  k8s-delete           Delete Kubernetes deployment"
	@echo "  k8s-status           Show Kubernetes deployment status"
	@echo "  k8s-logs             Show application logs from Kubernetes"
	@echo "  k8s-secret           Create/update Kubernetes secrets"
	@echo ""
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
.PHONY: docker-build docker-build-push docker-push docker-tag-latest docker-info docker-login

# Build arguments
BUILD_ARGS = --build-arg BUILD_VERSION=$(TAG)
ifeq ($(NO_CACHE),true)
    BUILD_ARGS += --no-cache
endif
ifeq ($(PLATFORM),linux/amd64)
    # Default platform, no need to specify
else
    BUILD_ARGS += --platform $(PLATFORM)
endif
ifeq ($(QUIET),true)
    BUILD_ARGS += --quiet
endif

docker-build:
	@echo "Building Docker image: $(FULL_IMAGE_NAME)"
	@echo "Tag: $(TAG)"
	@echo "Platform: $(PLATFORM)"
	@echo "Build args: $(BUILD_ARGS)"
	@if [ ! -f "Dockerfile" ]; then \
		echo "ERROR: Dockerfile not found. Please run from project root."; \
		exit 1; \
	fi
	@if ! docker info >/dev/null 2>&1; then \
		echo "ERROR: Docker is not running or accessible"; \
		exit 1; \
	fi
	docker build -t "$(FULL_IMAGE_NAME)" $(BUILD_ARGS) .
	@echo "Successfully built $(FULL_IMAGE_NAME)"

docker-build-push: docker-build docker-push
	@echo "Build and push completed successfully!"

docker-push:
	@echo "Pushing $(FULL_IMAGE_NAME)..."
	docker push "$(FULL_IMAGE_NAME)"
	@echo "Successfully pushed $(FULL_IMAGE_NAME)"

docker-tag-latest:
	@if [ "$(TAG)" != "latest" ]; then \
		LATEST_IMAGE="$(FULL_IMAGE_NAME)"; \
		LATEST_IMAGE="$${LATEST_IMAGE%:*}:latest"; \
		echo "Tagging as latest: $$LATEST_IMAGE"; \
		docker tag "$(FULL_IMAGE_NAME)" "$$LATEST_IMAGE"; \
		echo "Tagged as latest: $$LATEST_IMAGE"; \
	else \
		echo "Image is already tagged as latest"; \
	fi

docker-info:
	@echo "Docker Build Information:"
	@echo "  Image Name: $(IMAGE_NAME)"
	@echo "  Tag: $(TAG)"
	@echo "  Registry: $(REGISTRY)"
	@echo "  Full Image Name: $(FULL_IMAGE_NAME)"
	@echo "  Platform: $(PLATFORM)"
	@echo "  No Cache: $(NO_CACHE)"
	@echo "  Quiet: $(QUIET)"
	@echo "  Build Args: $(BUILD_ARGS)"

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
	@echo "Migrations are now handled automatically by the Docker container"
	@echo "Set RUN_MIGRATIONS=false in environment to disable auto-migration"

migrate-status:
	@echo "Migration status can be checked via the application health endpoint"

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

# ---- Kubernetes deployment ----
.PHONY: k8s-deploy k8s-delete k8s-status k8s-logs k8s-secret k8s-update-image
K8S_NAMESPACE ?= dhcp2p
K8S_DIR ?= k8s

k8s-deploy:
	@echo "Deploying to Kubernetes namespace: $(K8S_NAMESPACE)"
	@if ! kubectl get namespace $(K8S_NAMESPACE) >/dev/null 2>&1; then \
		echo "Creating namespace..."; \
		kubectl apply -f $(K8S_DIR)/namespace.yaml; \
	fi
	@echo "Applying ConfigMap..."
	@kubectl apply -f $(K8S_DIR)/configmap.yaml -n $(K8S_NAMESPACE)
	@if [ ! -f "$(K8S_DIR)/secret.yaml" ]; then \
		echo "ERROR: secret.yaml not found. Please create it from secret.yaml.example"; \
		echo "  cp $(K8S_DIR)/secret.yaml.example $(K8S_DIR)/secret.yaml"; \
		echo "  # Edit $(K8S_DIR)/secret.yaml with your actual secrets"; \
		exit 1; \
	fi
	@echo "Applying Secrets..."
	@kubectl apply -f $(K8S_DIR)/secret.yaml -n $(K8S_NAMESPACE)
	@echo "Applying Deployment..."
	@kubectl apply -f $(K8S_DIR)/deployment.yaml -n $(K8S_NAMESPACE)
	@echo "Applying Service..."
	@kubectl apply -f $(K8S_DIR)/service.yaml -n $(K8S_NAMESPACE)
	@echo "Waiting for deployment to be ready..."
	@kubectl wait --for=condition=available --timeout=300s deployment/dhcp2p -n $(K8S_NAMESPACE) || true
	@echo "Deployment completed!"

k8s-delete:
	@echo "Deleting Kubernetes deployment from namespace: $(K8S_NAMESPACE)"
	@kubectl delete deployment dhcp2p -n $(K8S_NAMESPACE) || true
	@kubectl delete service dhcp2p-service -n $(K8S_NAMESPACE) || true
	@kubectl delete configmap dhcp2p-config -n $(K8S_NAMESPACE) || true
	@kubectl delete secret dhcp2p-secrets -n $(K8S_NAMESPACE) || true
	@echo "Deployment deleted!"

k8s-status:
	@echo "Kubernetes Deployment Status:"
	@echo "=============================="
	@kubectl get all -n $(K8S_NAMESPACE) -l app=dhcp2p
	@echo ""
	@echo "Pods:"
	@kubectl get pods -n $(K8S_NAMESPACE) -l app=dhcp2p
	@echo ""
	@echo "Recent events:"
	@kubectl get events -n $(K8S_NAMESPACE) --sort-by='.lastTimestamp' | tail -10 || true

k8s-logs:
	@kubectl logs -f deployment/dhcp2p -n $(K8S_NAMESPACE) || kubectl logs -f -l app=dhcp2p -n $(K8S_NAMESPACE)

k8s-secret:
	@echo "Creating/updating Kubernetes secret..."
	@echo ""
	@echo "Example usage:"
	@echo "  kubectl create secret generic dhcp2p-secrets \\"
	@echo "    --from-literal=DATABASE_URL='postgres://user:pass@host:5432/db?sslmode=require' \\"
	@echo "    --from-literal=REDIS_URL='redis-service:6379' \\"
	@echo "    --from-literal=REDIS_PASSWORD='your-password' \\"
	@echo "    -n $(K8S_NAMESPACE)"
	@echo ""
	@echo "Or create from file:"
	@echo "  cp $(K8S_DIR)/secret.yaml.example $(K8S_DIR)/secret.yaml"
	@echo "  # Edit $(K8S_DIR)/secret.yaml with your secrets"
	@echo "  kubectl apply -f $(K8S_DIR)/secret.yaml"

k8s-update-image:
	@if [ -z "$(IMAGE)" ]; then \
		echo "ERROR: IMAGE must be specified"; \
		echo "Usage: make k8s-update-image IMAGE=your-registry/dhcp2p:v1.0.0"; \
		exit 1; \
	fi
	@echo "Updating deployment image to: $(IMAGE)"
	@kubectl set image deployment/dhcp2p dhcp2p=$(IMAGE) -n $(K8S_NAMESPACE)
	@echo "Image updated. Deployment will rollout automatically."