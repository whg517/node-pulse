.PHONY: test test-unit test-integration test-quick setup-test-db cleanup-test-db help

# Default target
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nPulse API Test Suite\n\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

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
	go test -short ./...

test-integration: ## Run integration tests only
	@echo "🧪 Running integration tests..."
	go test ./tests/integration/...

test: setup-test-db test-unit test-integration cleanup-test-db ## Run full test suite (setup + tests + cleanup)
	@echo "✅ All tests passed"

test-quick: ## Run all tests (assumes DB is running)
	@echo "🧪 Quick test (assumes DB is running)..."
	go test ./...
