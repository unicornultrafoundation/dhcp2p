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