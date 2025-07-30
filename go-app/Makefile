# Makefile for Cube Castle Employee Model System

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
COMPILER_BINARY=metacontract-compiler

.PHONY: build test clean compile-person validate-person dev-setup help

# Build the meta-contract compiler
build-compiler:
	$(GOBUILD) -o $(COMPILER_BINARY) ./cmd/metacontract-compiler

# Build all binaries
build: build-compiler
	@echo "✅ Build completed successfully"

# Test the system
test:
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(COMPILER_BINARY)
	rm -rf generated/

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Compile the person meta-contract example
compile-person: build-compiler
	@echo "🔧 Compiling Person meta-contract..."
	./$(COMPILER_BINARY) -input test-data/person.yaml -output ./generated -verbose

# Validate person meta-contract
validate-person: build-compiler
	@echo "✅ Validating Person meta-contract..."
	./$(COMPILER_BINARY) -input test-data/person.yaml -validate -verbose

# Full development workflow
dev-setup: build-compiler compile-person
	@echo "🚀 Development environment ready!"
	@echo ""
	@echo "📋 Generated Files:"
	@ls -la generated/schema/ generated/api/
	@echo ""
	@echo "🔧 Next Steps:"
	@echo "  1. Implement business logic in generated handlers"
	@echo "  2. Set up database connections"
	@echo "  3. Add authentication and authorization"

# Show help
help:
	@echo "Available targets:"
	@echo "  build-compiler   - Build the meta-contract compiler"
	@echo "  build           - Build all binaries"
	@echo "  test            - Run tests"
	@echo "  clean           - Clean build artifacts"
	@echo "  tidy            - Tidy Go dependencies"
	@echo "  compile-person  - Compile Person meta-contract example"
	@echo "  validate-person - Validate Person meta-contract"
	@echo "  dev-setup       - Complete development setup"
	@echo "  help            - Show this help"