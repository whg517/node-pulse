.PHONY: test test-unit test-integration test-quick setup-test-db cleanup-test-db help
.PHONY: build run clean deps swag swag-clean swag-force lint fmt

# Default target
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nPulse API - Node Pulse Backend Server\n\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies ready"

run: ## Run the server (default port 8080)
	@echo "🚀 Starting server..."
	@go run cmd/server/main.go

build: ## Build the server binary
	@echo "🔨 Building server..."
	@go build -o bin/pulse-api cmd/server/main.go
	@echo "✅ Binary created: bin/pulse-api"

build-debug: ## Build debug binary with optimizations disabled
	@echo "🔨 Building debug binary..."
	@go build -gcflags="all=-N -l" -o bin/pulse-api-debug cmd/server/main.go
	@echo "✅ Debug binary created: bin/pulse-api-debug"

clean: ## Clean build artifacts and generated files
	@echo "🧹 Cleaning..."
	@rm -rf bin/
	@rm -f server api.test
	@echo "✅ Clean complete"

fmt: ## Format Go code
	@echo "🎨 Formatting code..."
	@gofmt -w -s .
	@echo "✅ Code formatted"

lint: ## Run linters (requires golangci-lint)
	@echo "🔍 Running linters..."
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint not found. Install from: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

##@ Swagger Documentation

swag: ## Generate/update Swagger documentation
	@echo "📚 Generating Swagger documentation..."
	@if ! command -v swag > /dev/null 2>&1; then \
		echo "⚠️  swag not found. Installing..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@swag init -g cmd/server/main.go -o api/docs --parseDependency --parseInternal
	@echo "✅ Swagger documentation generated"
	@echo "📖 Access at: http://localhost:8080/swagger/index.html"

swag-clean: ## Clean generated Swagger documentation
	@echo "🧹 Cleaning Swagger docs..."
	@rm -f api/docs/docs.go api/docs/swagger.json api/docs/swagger.yaml
	@echo "✅ Swagger docs cleaned"

swag-force: deps swag-clean swag ## Force rebuild Swagger docs (clean + generate)

##@ Database Management
setup-test-db: ## Start test database container
	@echo "🐳 Starting test database..."
	docker compose -f docker-compose.test.yml up -d postgres-test
	@echo "⏳ Waiting for database to be ready..."
	@until docker compose -f docker-compose.test.yml exec postgres-test pg_isready -U testuser > /dev/null 2>&1; do \
		echo "Waiting for database..."; \
		sleep 2; \
	done
	@echo "✅ Test database ready"

cleanup-test-db: ## Stop test database container
	@echo "🛑 Stopping test database..."
	docker compose -f docker-compose.test.yml down

##@ Testing
test-unit: ## Run unit tests only (short mode)
	@echo "🧪 Running unit tests..."
	go test -short -p 1 ./...

test-integration: ## Run integration tests only
	@echo "🧪 Running integration tests..."
	go test -p 1 ./tests/integration/...

test: setup-test-db test-unit test-integration cleanup-test-db ## Run full test suite (setup + tests + cleanup)
	@echo "✅ All tests passed"

test-quick: ## Run all tests (assumes DB is running)
	@echo "🧪 Quick test (assumes DB is running)..."
	go test -p 1 ./...
