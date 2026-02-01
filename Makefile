MIGRATE_PATH=./db/migrations

.PHONY: migration
migration:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: Please provide NAME. Usage: make migration NAME=your_migration_name"; \
		exit 1; \
	fi
	@goose -dir $(MIGRATE_PATH) create $(NAME) sql

.PHONY: migrate-up
migrate-up:
	@read -p "Enter DB host [localhost]: " DB_HOST; \
	DB_HOST=$${DB_HOST:-localhost}; \
	read -p "Enter DB port [3306]: " DB_PORT; \
	DB_PORT=$${DB_PORT:-3306}; \
	read -p "Enter DB username [root]: " DB_USER; \
	DB_USER=$${DB_USER:-root}; \
	read -s -p "Enter DB password: " DB_PASS; echo; \
	read -p "Enter DB name [dev]: " DB_NAME; \
	DB_NAME=$${DB_NAME:-dev}; \
	DB_URL="$$DB_USER:$$DB_PASS@tcp($$DB_HOST:$$DB_PORT)/$$DB_NAME"; \
	echo "Running goose up..."; \
	goose -dir $(MIGRATE_PATH) mysql "$$DB_URL" up

.PHONY: migrate-down
migrate-down:
	@read -p "Enter DB host [localhost]: " DB_HOST; \
	DB_HOST=$${DB_HOST:-localhost}; \
	read -p "Enter DB port [3306]: " DB_PORT; \
	DB_PORT=$${DB_PORT:-3306}; \
	read -p "Enter DB username [root]: " DB_USER; \
	DB_USER=$${DB_USER:-root}; \
	read -s -p "Enter DB password: " DB_PASS; echo; \
	read -p "Enter DB name [dev]: " DB_NAME; \
	DB_NAME=$${DB_NAME:-dev}; \
	DB_URL="$$DB_USER:$$DB_PASS@tcp($$DB_HOST:$$DB_PORT)/$$DB_NAME"; \
	echo "Running goose down..."; \
	goose -dir $(MIGRATE_PATH) mysql "$$DB_URL" down

.PHONY: migrate-status
migrate-status:
	@read -p "Enter DB host [localhost]: " DB_HOST; \
	DB_HOST=$${DB_HOST:-localhost}; \
	read -p "Enter DB port [3306]: " DB_PORT; \
	DB_PORT=$${DB_PORT:-3306}; \
	read -p "Enter DB username [root]: " DB_USER; \
	DB_USER=$${DB_USER:-root}; \
	read -s -p "Enter DB password: " DB_PASS; echo; \
	read -p "Enter DB name [dev]: " DB_NAME; \
	DB_NAME=$${DB_NAME:-dev}; \
	DB_URL="$$DB_USER:$$DB_PASS@tcp($$DB_HOST:$$DB_PORT)/$$DB_NAME"; \
	goose -dir $(MIGRATE_PATH) mysql "$$DB_URL" status

.PHONY: migrate-version
migrate-version:
	@read -p "Enter DB host [localhost]: " DB_HOST; \
	DB_HOST=$${DB_HOST:-localhost}; \
	read -p "Enter DB port [3306]: " DB_PORT; \
	DB_PORT=$${DB_PORT:-3306}; \
	read -p "Enter DB username [root]: " DB_USER; \
	DB_USER=$${DB_USER:-root}; \
	read -s -p "Enter DB password: " DB_PASS; echo; \
	read -p "Enter DB name [dev]: " DB_NAME; \
	DB_NAME=$${DB_NAME:-dev}; \
	DB_URL="$$DB_USER:$$DB_PASS@tcp($$DB_HOST:$$DB_PORT)/$$DB_NAME"; \
	goose -dir $(MIGRATE_PATH) mysql "$$DB_URL" version

.PHONY: migrate-reset
migrate-reset:
	@read -p "Enter DB host [localhost]: " DB_HOST; \
	DB_HOST=$${DB_HOST:-localhost}; \
	read -p "Enter DB port [3306]: " DB_PORT; \
	DB_PORT=$${DB_PORT:-3306}; \
	read -p "Enter DB username [root]: " DB_USER; \
	DB_USER=$${DB_USER:-root}; \
	read -s -p "Enter DB password: " DB_PASS; echo; \
	read -p "Enter DB name [dev]: " DB_NAME; \
	DB_NAME=$${DB_NAME:-dev}; \
	DB_URL="$$DB_USER:$$DB_PASS@tcp($$DB_HOST:$$DB_PORT)/$$DB_NAME"; \
	echo "WARNING: This will reset all migrations!"; \
	read -p "Are you sure? (y/N): " CONFIRM; \
	if [ "$$CONFIRM" = "y" ] || [ "$$CONFIRM" = "Y" ]; then \
		goose -dir $(MIGRATE_PATH) mysql "$$DB_URL" reset; \
		echo "Reset all migration is success."; \
	else \
		echo "Reset cancelled."; \
	fi