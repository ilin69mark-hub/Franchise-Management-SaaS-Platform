.PHONY: test test-coverage test-watch test-e2e test-e2e-ui lint build clean

# Install dependencies
install:
	npm install

# Run unit tests
test:
	npm test

# Run tests with coverage
test-coverage:
	npm test -- --coverage

# Run tests in watch mode
test-watch:
	npm run test:watch

# Run E2E tests
test-e2e:
	npm run test:e2e

# Run E2E tests with UI
test-e2e-ui:
	npm run test:e2e:ui

# Run E2E tests in headed mode
test-e2e-headed:
	npm run test:e2e:headed

# Lint code
lint:
	npm run lint

# TypeScript check
typecheck:
	npx tsc --noEmit

# Build production
build:
	npm run build

# Clean
clean:
	rm -rf .next coverage node_modules/.cache