#!/bin/bash

# Test Coverage and Benchmark Script for Go Meta-Contract Editor Backend
# Usage: ./test_coverage.sh [options]
# Options:
#   --coverage    Generate coverage report
#   --benchmark   Run benchmark tests
#   --all         Run all tests with coverage and benchmarks
#   --html        Generate HTML coverage report
#   --verbose     Verbose output

set -e

# Configuration
COVERAGE_DIR="coverage"
COVERAGE_FILE="$COVERAGE_DIR/coverage.out"
COVERAGE_HTML="$COVERAGE_DIR/coverage.html"
BENCHMARK_FILE="$COVERAGE_DIR/benchmarks.txt"
TEST_TIMEOUT="10m"
MIN_COVERAGE_THRESHOLD=85

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

create_coverage_dir() {
    mkdir -p "$COVERAGE_DIR"
}

run_unit_tests() {
    print_info "Running unit tests..."
    
    if [ "$VERBOSE" = true ]; then
        go test -v -timeout="$TEST_TIMEOUT" ./internal/...
    else
        go test -timeout="$TEST_TIMEOUT" ./internal/...
    fi
    
    print_success "Unit tests completed"
}

run_tests_with_coverage() {
    print_info "Running tests with coverage analysis..."
    
    create_coverage_dir
    
    # Run tests with coverage
    go test -timeout="$TEST_TIMEOUT" -coverprofile="$COVERAGE_FILE" -covermode=atomic ./internal/...
    
    if [ ! -f "$COVERAGE_FILE" ]; then
        print_error "Coverage file not generated"
        exit 1
    fi
    
    # Generate coverage summary
    local coverage_percent=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}' | sed 's/%//')
    
    print_info "Coverage Summary:"
    go tool cover -func="$COVERAGE_FILE" | grep -E "(total|internal/)" | tail -10
    
    # Check coverage threshold
    if (( $(echo "$coverage_percent >= $MIN_COVERAGE_THRESHOLD" | bc -l) )); then
        print_success "Coverage: ${coverage_percent}% (above threshold of ${MIN_COVERAGE_THRESHOLD}%)"
    else
        print_warning "Coverage: ${coverage_percent}% (below threshold of ${MIN_COVERAGE_THRESHOLD}%)"
    fi
    
    # Generate detailed coverage report
    print_info "Generating detailed coverage report..."
    go tool cover -func="$COVERAGE_FILE" > "$COVERAGE_DIR/coverage_detailed.txt"
    
    print_success "Coverage analysis completed"
}

generate_html_coverage() {
    if [ ! -f "$COVERAGE_FILE" ]; then
        print_error "Coverage file not found. Run with --coverage first."
        exit 1
    fi
    
    print_info "Generating HTML coverage report..."
    go tool cover -html="$COVERAGE_FILE" -o "$COVERAGE_HTML"
    
    if [ -f "$COVERAGE_HTML" ]; then
        print_success "HTML coverage report generated: $COVERAGE_HTML"
        
        # Try to open in browser on macOS/Linux
        if command -v open &> /dev/null; then
            open "$COVERAGE_HTML"
        elif command -v xdg-open &> /dev/null; then
            xdg-open "$COVERAGE_HTML"
        fi
    else
        print_error "Failed to generate HTML coverage report"
        exit 1
    fi
}

run_benchmarks() {
    print_info "Running benchmark tests..."
    
    create_coverage_dir
    
    # Run benchmarks and save to file
    go test -bench=. -benchmem -timeout="$TEST_TIMEOUT" ./internal/... | tee "$BENCHMARK_FILE"
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        print_success "Benchmark tests completed"
        print_info "Benchmark results saved to: $BENCHMARK_FILE"
        
        # Extract and display summary
        print_info "Benchmark Summary:"
        grep "Benchmark" "$BENCHMARK_FILE" | head -10
    else
        print_error "Benchmark tests failed"
        exit 1
    fi
}

run_race_detection() {
    print_info "Running tests with race detection..."
    
    go test -race -timeout="$TEST_TIMEOUT" ./internal/...
    
    if [ $? -eq 0 ]; then
        print_success "Race detection tests passed"
    else
        print_error "Race conditions detected"
        exit 1
    fi
}

run_memory_tests() {
    print_info "Running memory tests..."
    
    # Test with different memory limits
    GOMAXPROCS=1 go test -timeout="$TEST_TIMEOUT" ./internal/...
    
    print_success "Memory tests completed"
}

generate_test_report() {
    print_info "Generating comprehensive test report..."
    
    local report_file="$COVERAGE_DIR/test_report.md"
    
    cat > "$report_file" << EOF
# Go Meta-Contract Editor Backend Test Report

Generated on: $(date)

## Test Summary

### Unit Tests
$(go test -json ./internal/... 2>/dev/null | jq -r '. | select(.Action == "pass" or .Action == "fail") | "\(.Package): \(.Action)"' | sort | uniq -c || echo "Test summary not available")

EOF

    if [ -f "$COVERAGE_FILE" ]; then
        local coverage_percent=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}')
        cat >> "$report_file" << EOF
### Coverage Analysis
- **Total Coverage**: $coverage_percent
- **Coverage Threshold**: ${MIN_COVERAGE_THRESHOLD}%
- **Status**: $(if (( $(echo "${coverage_percent%\%} >= $MIN_COVERAGE_THRESHOLD" | bc -l) )); then echo "✅ PASSED"; else echo "❌ BELOW THRESHOLD"; fi)

#### Coverage by Package
\`\`\`
$(go tool cover -func="$COVERAGE_FILE" | grep -E "internal/" | tail -20)
\`\`\`

EOF
    fi

    if [ -f "$BENCHMARK_FILE" ]; then
        cat >> "$report_file" << EOF
### Benchmark Results
\`\`\`
$(grep "Benchmark" "$BENCHMARK_FILE" | head -10)
\`\`\`

EOF
    fi

    cat >> "$report_file" << EOF
## Test Categories

### 1. Meta-Contract Compiler Tests
- Parser tests (YAML parsing, validation)
- Validator tests (field validation, constraint checking)
- Compiler tests (code generation, error handling)

### 2. LocalAI Service Tests
- Service integration tests
- NLP engine tests
- Code analyzer tests
- Performance benchmarks

### 3. WebSocket Communication Tests
- Hub management tests
- Client connection tests
- Message broadcasting tests
- Concurrent operation tests

### 4. Integration Tests
- End-to-end workflow tests
- Cross-module integration tests
- Performance under load tests

## Coverage Requirements

- **Minimum Coverage**: ${MIN_COVERAGE_THRESHOLD}%
- **Critical Paths**: 95%+ coverage required
- **Error Handling**: All error paths must be tested
- **Edge Cases**: Boundary conditions and edge cases covered

## Performance Benchmarks

All performance-critical functions include benchmark tests to track:
- Memory allocation patterns
- Processing time performance
- Concurrency behavior
- Resource utilization

EOF

    print_success "Test report generated: $report_file"
}

cleanup() {
    print_info "Cleaning up temporary files..."
    # Add cleanup logic if needed
}

show_help() {
    cat << EOF
Go Meta-Contract Editor Backend Test Script

Usage: $0 [OPTIONS]

Options:
    --coverage      Run tests with coverage analysis
    --benchmark     Run benchmark tests
    --html          Generate HTML coverage report (requires --coverage)
    --race          Run tests with race detection
    --memory        Run memory tests
    --report        Generate comprehensive test report
    --all           Run all tests, coverage, and benchmarks
    --verbose       Enable verbose output
    --help          Show this help message

Examples:
    $0 --coverage --html          # Run coverage tests and generate HTML report
    $0 --benchmark                # Run only benchmark tests
    $0 --all                      # Run everything
    $0 --coverage --race --report # Run coverage with race detection and generate report

Environment Variables:
    MIN_COVERAGE_THRESHOLD  Minimum coverage percentage required (default: 85)
    TEST_TIMEOUT           Test timeout duration (default: 10m)

EOF
}

# Parse command line arguments
COVERAGE=false
HTML=false
BENCHMARK=false
RACE=false
MEMORY=false
REPORT=false
ALL=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --coverage)
            COVERAGE=true
            shift
            ;;
        --html)
            HTML=true
            shift
            ;;
        --benchmark)
            BENCHMARK=true
            shift
            ;;
        --race)
            RACE=true
            shift
            ;;
        --memory)
            MEMORY=true
            shift
            ;;
        --report)
            REPORT=true
            shift
            ;;
        --all)
            ALL=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Main execution
trap cleanup EXIT

print_info "Starting Go Meta-Contract Editor Backend Test Suite"
print_info "======================================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Check if bc is available for float comparison
if ! command -v bc &> /dev/null; then
    print_warning "bc not found, coverage threshold checking disabled"
fi

# Run all tests if --all is specified
if [ "$ALL" = true ]; then
    COVERAGE=true
    HTML=true
    BENCHMARK=true
    RACE=true
    MEMORY=true
    REPORT=true
fi

# If no specific options, run basic unit tests
if [ "$COVERAGE" = false ] && [ "$BENCHMARK" = false ] && [ "$RACE" = false ] && [ "$MEMORY" = false ]; then
    run_unit_tests
    exit 0
fi

# Run tests based on options
if [ "$COVERAGE" = true ]; then
    run_tests_with_coverage
fi

if [ "$HTML" = true ]; then
    generate_html_coverage
fi

if [ "$BENCHMARK" = true ]; then
    run_benchmarks
fi

if [ "$RACE" = true ]; then
    run_race_detection
fi

if [ "$MEMORY" = true ]; then
    run_memory_tests
fi

if [ "$REPORT" = true ]; then
    generate_test_report
fi

print_success "Test suite completed successfully!"
print_info "Coverage files available in: $COVERAGE_DIR/"