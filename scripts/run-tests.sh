# Test Configuration Files

## Go Test Configuration
# go-app/go.test.env
TEST_DB_URL=file:ent?mode=memory&cache=shared&_fk=1
NEO4J_URI=bolt://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=password
LOG_LEVEL=debug
TEST_TIMEOUT=10m

## Jest Configuration for Frontend Tests
# nextjs-app/jest.config.js
const nextJest = require('next/jest')

const createJestConfig = nextJest({
  // Provide the path to your Next.js app to load next.config.js and .env files
  dir: './',
})

const customJestConfig = {
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
  moduleNameMapping: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '^@/components/(.*)$': '<rootDir>/src/components/$1',
    '^@/hooks/(.*)$': '<rootDir>/src/hooks/$1',
    '^@/lib/(.*)$': '<rootDir>/src/lib/$1',
    '^@/pages/(.*)$': '<rootDir>/src/pages/$1',
  },
  testEnvironment: 'jest-environment-jsdom',
  collectCoverageFrom: [
    'src/**/*.{js,jsx,ts,tsx}',
    '!src/**/*.d.ts',
    '!src/pages/_app.tsx',
    '!src/pages/_document.tsx',
    '!src/pages/api/**/*',
  ],
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80,
    },
  },
  testMatch: [
    '<rootDir>/src/**/__tests__/**/*.{js,jsx,ts,tsx}',
    '<rootDir>/src/**/*.{test,spec}.{js,jsx,ts,tsx}',
  ],
  moduleDirectories: ['node_modules', '<rootDir>/'],
  testEnvironmentOptions: {
    url: 'http://localhost:3000',
  },
}

module.exports = createJestConfig(customJestConfig)

## Playwright Configuration
# go-app/playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './test/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [
    ['html'],
    ['json', { outputFile: 'test-results/results.json' }],
    ['junit', { outputFile: 'test-results/results.xml' }],
  ],
  use: {
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
    },
  ],
  webServer: {
    command: 'npm run dev',
    port: 3000,
    reuseExistingServer: !process.env.CI,
  },
});

## Docker Compose for Testing
# docker-compose.test.yml
version: '3.8'

services:
  postgres-test:
    image: postgres:15
    environment:
      POSTGRES_DB: testdb
      POSTGRES_USER: testuser
      POSTGRES_PASSWORD: testpass
    ports:
      - "5433:5432"
    volumes:
      - postgres_test_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U testuser -d testdb"]
      interval: 10s
      timeout: 5s
      retries: 5

  neo4j-test:
    image: neo4j:5.0-community
    environment:
      NEO4J_AUTH: neo4j/testpassword
      NEO4J_dbms_memory_pagecache_size: 256M
      NEO4J_dbms_memory_heap_initial__size: 256M
      NEO4J_dbms_memory_heap_max__size: 512M
    ports:
      - "7475:7474"
      - "7688:7687"
    volumes:
      - neo4j_test_data:/data
    healthcheck:
      test: ["CMD", "cypher-shell", "--username", "neo4j", "--password", "testpassword", "RETURN 1"]
      interval: 10s
      timeout: 5s
      retries: 10

  # 已废弃：Temporal 工作流引擎测试容器（历史参考，当前仓库不再提供/集成该服务）
  # temporal-test:
  #   image: temporalio/auto-setup:1.20.0
  #   environment:
  #     - DB=postgresql
  #     - DB_PORT=5432
  #     - POSTGRES_USER=testuser
  #     - POSTGRES_PWD=testpass
  #     - POSTGRES_SEEDS=postgres-test
  #   ports:
  #     - "7234:7233"
  #     - "8234:8233"
  #   depends_on:
  #     postgres-test:
  #       condition: service_healthy

  backend-test:
    build:
      context: ./go-app
      dockerfile: Dockerfile.test
    environment:
      DB_URL: postgres://testuser:testpass@postgres-test:5432/testdb?sslmode=disable
      NEO4J_URI: bolt://neo4j-test:7687
      NEO4J_USERNAME: neo4j
      NEO4J_PASSWORD: testpassword
      # 已废弃：工作流引擎地址（保留为历史示例，不再使用）
      # TEMPORAL_ADDRESS: temporal-test:7233
    depends_on:
      postgres-test:
        condition: service_healthy
      neo4j-test:
        condition: service_healthy
      # temporal-test:
      #   condition: service_started
    volumes:
      - ./go-app:/app
      - go_test_cache:/go/pkg/mod
    command: ["make", "-f", "Makefile.test", "test-all"]

  frontend-test:
    build:
      context: ./nextjs-app
      dockerfile: Dockerfile.test
    environment:
      NEXT_PUBLIC_API_URL: http://backend-test:8080
      NODE_ENV: test
    volumes:
      - ./nextjs-app:/app
      - node_test_modules:/app/node_modules
    command: ["npm", "run", "test:ci"]

volumes:
  postgres_test_data:
  neo4j_test_data:
  go_test_cache:
  node_test_modules:

## Test Scripts
# scripts/run-tests.sh
#!/bin/bash

set -e

echo "🚀 Starting Employee Model Test Suite"
echo "====================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "SUCCESS")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "ERROR")
            echo -e "${RED}❌ $message${NC}"
            ;;
        "INFO")
            echo -e "${YELLOW}ℹ️  $message${NC}"
            ;;
    esac
}

# Function to run tests with error handling
run_test() {
    local test_name=$1
    local test_command=$2
    
    print_status "INFO" "Running $test_name..."
    
    if eval "$test_command"; then
        print_status "SUCCESS" "$test_name completed successfully"
        return 0
    else
        print_status "ERROR" "$test_name failed"
        return 1
    fi
}

# Parse command line arguments
RUN_UNIT=true
RUN_INTEGRATION=true
RUN_E2E=false
RUN_PERFORMANCE=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --unit-only)
            RUN_INTEGRATION=false
            RUN_E2E=false
            shift
            ;;
        --integration-only)
            RUN_UNIT=false
            RUN_E2E=false
            shift
            ;;
        --e2e)
            RUN_E2E=true
            shift
            ;;
        --performance)
            RUN_PERFORMANCE=true
            shift
            ;;
        --all)
            RUN_E2E=true
            RUN_PERFORMANCE=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--unit-only|--integration-only|--e2e|--performance|--all] [--verbose]"
            exit 1
            ;;
    esac
done

# Set verbose flag
if [ "$VERBOSE" = true ]; then
    export GO_TEST_FLAGS="-v -race -timeout=10m"
else
    export GO_TEST_FLAGS="-race -timeout=10m"
fi

print_status "INFO" "Test configuration:"
echo "  Unit tests: $RUN_UNIT"
echo "  Integration tests: $RUN_INTEGRATION"
echo "  E2E tests: $RUN_E2E"
echo "  Performance tests: $RUN_PERFORMANCE"
echo "  Verbose: $VERBOSE"
echo ""

# Start test services
print_status "INFO" "Starting test services..."
docker-compose -f docker-compose.test.yml up -d postgres-test neo4j-test

# Wait for services to be ready
print_status "INFO" "Waiting for test services to be ready..."
timeout 60 bash -c 'until docker-compose -f docker-compose.test.yml exec postgres-test pg_isready -U testuser -d testdb; do sleep 2; done'
timeout 90 bash -c 'until docker-compose -f docker-compose.test.yml exec neo4j-test cypher-shell --username neo4j --password testpassword "RETURN 1"; do sleep 2; done'

# Change to go-app directory
cd go-app

# Track test results
FAILED_TESTS=()

# Run unit tests
if [ "$RUN_UNIT" = true ]; then
    if ! run_test "Unit Tests" "make -f Makefile.test test-unit"; then
        FAILED_TESTS+=("Unit Tests")
    fi
fi

# Run integration tests
if [ "$RUN_INTEGRATION" = true ]; then
    if ! run_test "Integration Tests" "make -f Makefile.test test-integration"; then
        FAILED_TESTS+=("Integration Tests")
    fi
fi

# Run E2E tests
if [ "$RUN_E2E" = true ]; then
    print_status "INFO" "Starting application for E2E tests..."
    
    # Build and start the application
    docker-compose -f ../docker-compose.test.yml up -d backend-test frontend-test
    
    # Wait for application to be ready
    timeout 120 bash -c 'until curl -f http://localhost:8080/health; do sleep 2; done'
    timeout 120 bash -c 'until curl -f http://localhost:3000; do sleep 2; done'
    
    if ! run_test "E2E Tests" "make -f Makefile.test test-e2e"; then
        FAILED_TESTS+=("E2E Tests")
    fi
fi

# Run performance tests
if [ "$RUN_PERFORMANCE" = true ]; then
    if ! run_test "Performance Tests" "make -f Makefile.test test-performance"; then
        FAILED_TESTS+=("Performance Tests")
    fi
fi

# Generate coverage report
if [ "$RUN_UNIT" = true ] || [ "$RUN_INTEGRATION" = true ]; then
    print_status "INFO" "Generating coverage report..."
    make -f Makefile.test test-coverage
fi

# Cleanup
print_status "INFO" "Cleaning up test services..."
docker-compose -f ../docker-compose.test.yml down -v

# Summary
echo ""
echo "====================================="
echo "🏁 Test Suite Summary"
echo "====================================="

if [ ${#FAILED_TESTS[@]} -eq 0 ]; then
    print_status "SUCCESS" "All tests passed! 🎉"
    exit 0
else
    print_status "ERROR" "The following tests failed:"
    for test in "${FAILED_TESTS[@]}"; do
        echo "  - $test"
    done
    exit 1
fi
