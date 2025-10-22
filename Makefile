MIGRATION_DIR := internal/app/infrastructure/migrations
SCHEMA_FILE := $(MIGRATION_DIR)/schema.hcl

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