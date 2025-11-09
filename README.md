# Simple CRUD Interface

A RESTful API for user management built with Go, Gin, and PostgreSQL.

The task itself can be found [here](/TASK.md)

## Prerequisites

- [Go](https://go.dev/dl/) 1.25 or later
- [Docker](https://www.docker.com/get-started/) and Docker Compose (v1 or v2)
- [Goose](https://github.com/pressly/goose) (optional - can use `go run` instead)

## Getting Started

1. Start database

```
## Via Makefile
make db

## Via Docker (works with both docker-compose v1 and docker compose v2)
docker compose up -d db
# or
docker-compose up -d db
```

2. Run migrations

```
## Via Makefile
make migrate-up

## Via Goose
DB_DRIVER=postgres
DB_STRING="host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
goose -dir ./migrations $(DB_DRIVER) $(DB_STRING) up
```

3. Run application

```bash
# Via Makefile
make run

# Directly
go run cmd/main.go
```

**Environment Variables (optional):**

- `POSTGRES_DSN` - Database connection string (defaults to localhost:5432)
- `API_KEY` - API key for X-API-Key authentication (optional for development)
  - If not set, all requests are allowed (development mode)
  - If set, requests must include `X-API-Key: <API_KEY>` header

**Example with API key:**

```bash
API_KEY=your-secret-key go run cmd/main.go
```

The application runs on `http://localhost:8080` by default. All requests are logged in JSON format to stderr.

## API Endpoints

All endpoints are under `/api/v1/users`:

- `GET /api/v1/users` - Get all users
- `GET /api/v1/users/username/:username` - Get user by username
- `GET /api/v1/users/id/:id` - Get user by ID
- `POST /api/v1/users` - Create a new user
- `PATCH /api/v1/users/:uuid` - Update user by UUID
- `DELETE /api/v1/users/:uuid` - Delete user by UUID

**Authentication:**
- If `API_KEY` environment variable is set, all requests must include `X-API-Key: <API_KEY>` header
- Missing header returns `401 Unauthorized`
- Invalid key returns `403 Forbidden`
- If `API_KEY` is not set, all requests are allowed (development mode)

## Docker Deployment

The application can be run using Docker Compose:

```bash
# Start database and application
docker compose up -d db
docker compose up --build app

# Or start everything at once
docker compose up --build
```

The application container will:
- Build from the Dockerfile (multi-stage build, minimal scratch image)
- Connect to the `db` service
- Use `API_KEY` from environment or default to `test-api-key-12345`
- Expose port 8080

## Running Tests

### Integration Tests

Integration tests require a test database. The Makefile provides convenient commands to run all integration tests:

```bash
# Run all integration tests (starts test DB, runs migrations, runs tests, cleans up)
make test-integration

# Run integration tests but keep the test database running (useful for debugging)
make test-integration-keep
```

The `test-integration` target will:
1. Start a test PostgreSQL container on port 5433
2. Wait for the database to be ready
3. Run database migrations
4. Execute all handler tests
5. Clean up the test database container

**Manual setup (if not using Makefile):**

```bash
# Start test database
docker compose up -d db-test

# Run migrations
make migrate-test-up

# Run tests
go test ./internal/handler -v

# Clean up (optional)
make migrate-test-down
docker compose down db-test
```