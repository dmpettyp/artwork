# Database migration Makefile
# Requires golang-migrate CLI: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

PASSWORD = foofoofoo
DBNAME = artwork
DATABASE_URL ?= postgres://postgres:$(PASSWORD)@localhost:5432/$(DBNAME)?sslmode=disable
MIGRATIONS_DIR = .

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

.PHONY: up
up: ## Run all pending migrations
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

.PHONY: down
down: ## Rollback the last migration
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

.PHONY: down-all
down-all: ## Rollback all migrations (WARNING: destructive)
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down

.PHONY: version
version: ## Show current migration version
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

.PHONY: force
force: ## Force set migration version (use: make force VERSION=1)
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION not set. Usage: make force VERSION=1"; exit 1; fi
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(VERSION)

.PHONY: create
create: ## Create a new migration (use: make create NAME=add_users_table)
	@if [ -z "$(NAME)" ]; then echo "Error: NAME not set. Usage: make create NAME=migration_name"; exit 1; fi
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

.PHONY: drop
drop: ## Drop everything in the database (WARNING: very destructive)
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" drop

.PHONY: docker-start
docker-start: ## Start postgres in Docker
	docker run --name postgres-artwork -e POSTGRES_PASSWORD=$(PASSWORD) -v pgdata:/tmp -p 5432:5432 -d postgres

.PHONY: docker-stop
docker-stop: ## Stop and remove postgres container
	docker stop postgres-artwork || true
	docker rm postgres-artwork || true

.PHONY: docker-create-db
docker-create-db: ## Create the database in the running postgres container
	@echo "Creating database $(DBNAME)..."
	docker exec -it postgres-artwork psql -U postgres -c "CREATE DATABASE $(DBNAME);"
	@echo "Database $(DBNAME) created successfully!"

.PHONY: docker-drop-db
docker-drop-db: ## Drop the database (WARNING: destructive)
	@echo "Dropping database $(DBNAME)..."
	docker exec -it postgres-artwork psql -U postgres -c "DROP DATABASE IF EXISTS $(DBNAME);"
	@echo "Database $(DBNAME) dropped!"

.PHONY: docker-setup
docker-setup: docker-start ## Start postgres and create database (waits 3s for postgres to start)
	@echo "Waiting for postgres to start..."
	@sleep 3
	@$(MAKE) docker-create-db

.PHONY: docker-reset
docker-reset: docker-stop docker-setup up ## Stop, restart, recreate DB, and run migrations

.PHONY: psql
psql: ## Connect to the database with local psql
	PGPASSWORD=$(PASSWORD) psql -h localhost -U postgres -d $(DBNAME)
