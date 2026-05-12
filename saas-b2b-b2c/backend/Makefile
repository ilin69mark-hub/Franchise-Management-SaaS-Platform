.PHONY: test test-pg test-unit test-all lint clean

# Run all unit tests (fast, no external deps)
test:
	go test -short ./...

# Run PostgreSQL integration tests (requires docker)
test-pg:
	docker-compose -f docker-compose.test.yml up -d
	sleep 2
	go test -v -run "Postgres" ./...
	docker-compose -f docker-compose.test.yml down

# Run all tests including PostgreSQL
test-all:
	docker-compose -f docker-compose.test.yml up -d
	sleep 2
	go test -v ./...
	docker-compose -f docker-compose.test.yml down

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Run linter
lint:
	golangci-lint run

# Clean up
clean:
	docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true
	rm -f coverage.out

# Run specific package tests
test-handlers:
	go test -v ./internal/handlers/...

test-services:
	go test -v ./internal/services/...

test-repos:
	go test -v ./internal/repository/...