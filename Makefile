-include .env

# Check if goose is installed, if not use go run
GOOSE := $(shell command -v goose 2> /dev/null)
ifndef GOOSE
	GOOSE = go run github.com/pressly/goose/v3/cmd/goose@latest
else
	GOOSE = goose
endif

DB_DRIVER=postgres
DB_STRING="host=${POSTGRES_HOST} port=${POSTGRES_PORT} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} sslmode=disable"

migrate-up:
	$(GOOSE) -dir ./migrations $(DB_DRIVER) $(DB_STRING) up

migrate-down:
	$(GOOSE) -dir ./migrations $(DB_DRIVER) $(DB_STRING) down

migrate-status:
	$(GOOSE) -dir ./migrations $(DB_DRIVER) $(DB_STRING) status

migrate-reset:
	$(GOOSE) -dir ./migrations $(DB_DRIVER) $(DB_STRING) reset

lint:
	golangci-lint run ./...

security:
	gosec -exclude-dir=bin/database ./...

test:
	go test -cover -coverprofile=./test.out ./...

coverage:
	go tool cover -func ./test.out

coverage-html:
	go tool cover -html=./test.out

validate: lint security test

run:
	go run cmd/main.go

db:
	docker compose up -d db

db-test:
	@echo "Starting test database on port 5433 (production DB uses 5432)..."
	docker compose up -d db-test

db-test-clean:
	@echo "Cleaning and restarting test database..."
	docker compose down db-test
	docker volume rm go-hw_postgres_test_data 2>/dev/null || true
	docker compose up -d db-test

migrate-test-up: db-test
	@echo "Waiting for database to be ready..."
	@timeout=30; \
	while [ $$timeout -gt 0 ]; do \
		if docker inspect --format='{{.State.Health.Status}}' database_test 2>/dev/null | grep -q "healthy" || docker exec database_test pg_isready -U postgres > /dev/null 2>&1; then \
			echo "Database is ready!"; \
			break; \
		fi; \
		echo "Waiting for database... ($$timeout seconds remaining)"; \
		sleep 1; \
		timeout=$$((timeout - 1)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "Database failed to start in time"; \
		exit 1; \
	fi
	TEST_POSTGRES_DB=postgres_test TEST_POSTGRES_PORT=5433 \
	$(GOOSE) -dir ./migrations postgres "host=localhost port=5433 user=postgres password=postgres dbname=postgres_test sslmode=disable" up

migrate-test-down:
	TEST_POSTGRES_DB=postgres_test TEST_POSTGRES_PORT=5433 \
	$(GOOSE) -dir ./migrations postgres "host=localhost port=5433 user=postgres password=postgres dbname=postgres_test sslmode=disable" down

test-handler:
	@echo "Running handler tests..."
	go test ./internal/handler -v

test-integration: db-test migrate-test-up test-handler
	@echo "Cleaning up test database..."
	-docker compose down db-test
	@echo "Integration tests completed"

test-integration-keep: db-test migrate-test-up test-handler
	@echo "Integration tests completed (database kept running for debugging)"

up:
	docker compose up --build

down:
	docker compose down

restart:
	docker compose down
	docker compose up --build

create-migration:
	@read -p "Enter migration name: " name; \
	$(GOOSE) -dir ./migrations create $$name sql

swagger:
	swag init -g ./cmd/api/v1/main.go -o ./docs