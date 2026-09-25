.PHONY: test test-pg test-unit test-all lint clean

COMPOSE := docker compose

# Run all unit tests (fast, no external deps)
test:
	go test -short ./...

# Run PostgreSQL integration tests (requires docker)
test-pg:
	$(COMPOSE) -f docker-compose.test.yml up -d --wait
	TEST_DB_DSN='host=127.0.0.1 port=5433 user=postgres password=postgres dbname=franchise_test sslmode=disable' go test -v -run "Postgres" ./...
	$(COMPOSE) -f docker-compose.test.yml down

# Run all tests including PostgreSQL
test-all:
	$(COMPOSE) -f docker-compose.test.yml up -d --wait
	TEST_DB_DSN='host=127.0.0.1 port=5433 user=postgres password=postgres dbname=franchise_test sslmode=disable' go test -v ./...
	$(COMPOSE) -f docker-compose.test.yml down

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Run linter
lint:
	golangci-lint run

# Clean up
clean:
	$(COMPOSE) -f docker-compose.test.yml down -v 2>/dev/null || true
	rm -f coverage.out

# Run specific package tests
test-handlers:
	go test -v ./internal/handlers/...

test-services:
	go test -v ./internal/services/...

test-repos:
	go test -v ./internal/repository/...