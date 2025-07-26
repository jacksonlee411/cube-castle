# Makefile for Cube Castle Go Application

.PHONY: help build test test-integration test-e2e clean run-server run-temporal docker-up docker-down lint fmt

# Variables
APP_NAME := cube-castle
BUILD_DIR := ./build
CMD_DIR := ./cmd
TEST_TIMEOUT := 300s

# Help target
help: ## Show this help message
	@echo "Cube Castle - Available Make Targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build targets
build: ## Build the application
	@echo "Building application..."
	@go mod tidy
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/test-server $(CMD_DIR)/test-server/main.go
	@echo "Build completed: $(BUILD_DIR)/test-server"

build-all: ## Build all components
	@echo "Building all components..."
	@go mod tidy
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/test-server $(CMD_DIR)/test-server/main.go
	@echo "All builds completed"

# Development targets
run-server: build ## Run the test server
	@echo "Starting test server..."
	@./$(BUILD_DIR)/test-server

run-temporal: ## Start Temporal services using Docker Compose
	@echo "Starting Temporal services..."
	@docker-compose -f ../docker-compose.temporal-optimized.yml up -d
	@echo "Waiting for services to be ready..."
	@sleep 30
	@echo "Temporal services started"

# Testing targets
test: ## Run unit tests
	@echo "Running unit tests..."
	@go test -v -timeout $(TEST_TIMEOUT) ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -timeout $(TEST_TIMEOUT) -tags=integration ./tests/...

test-e2e: ## Run E2E tests (requires running server)
	@echo "Running E2E tests..."
	@echo "Make sure the test server is running on localhost:8080"
	@go test -v -timeout $(TEST_TIMEOUT) -tags=e2e ./tests/e2e_test.go

test-all: test test-integration ## Run all tests except E2E

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Docker targets
docker-up: ## Start all services with Docker Compose
	@echo "Starting all services..."
	@docker-compose -f ../docker-compose.temporal-optimized.yml up -d
	@echo "Services started"

docker-down: ## Stop all services
	@echo "Stopping all services..."
	@docker-compose -f ../docker-compose.temporal-optimized.yml down
	@echo "Services stopped"

docker-logs: ## Show Docker Compose logs
	@docker-compose -f ../docker-compose.temporal-optimized.yml logs -f

# Development tools
lint: ## Run linters
	@echo "Running linters..."
	@go fmt ./...
	@go vet ./...
	@echo "Linting completed"

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .
	@echo "Code formatted"

mod-tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	@go mod tidy
	@echo "Modules tidied"

# Clean targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean completed"

clean-all: clean ## Clean everything including Docker volumes
	@echo "Cleaning Docker volumes..."
	@docker-compose -f ../docker-compose.temporal-optimized.yml down -v
	@echo "All cleaned"

# Development workflow
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	@go mod tidy
	@mkdir -p $(BUILD_DIR)
	@echo "Development environment ready"

dev-test: lint test-integration ## Run development tests
	@echo "Development testing completed"

# CI/CD targets
ci-test: ## Run tests suitable for CI environment
	@echo "Running CI tests..."
	@go mod tidy
	@go test -v -race -timeout $(TEST_TIMEOUT) ./...
	@echo "CI tests completed"

# Monitoring and debugging
monitor: ## Monitor application logs (requires running server)
	@echo "Monitoring application..."
	@curl -s http://localhost:8080/monitor/live

health-check: ## Check application health
	@echo "Checking application health..."
	@curl -s http://localhost:8080/health | jq .

metrics: ## Show application metrics
	@echo "Fetching application metrics..."
	@curl -s http://localhost:8080/metrics | jq .

# Default target
.DEFAULT_GOAL := help