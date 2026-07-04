.PHONY: test test-unit test-integration test-quick setup-test-db cleanup-test-db help
.PHONY: build run clean deps swag swag-clean swag-force swag-ensure lint fmt
.PHONY: migrate-up migrate-down migrate-force migrate-version migrate-create
.PHONY: web-build

# Default target
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nPulse API - Node Pulse Backend Server\n\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies ready"

# swag-ensure regenerates the Swagger docs only when they are missing, so a
# fresh checkout still builds. The committed source annotations are the source
# of truth; run `make swag-force` deliberately to refresh them.
swag-ensure:
	@if [ ! -f api/docs/docs.go ]; then \
		echo "📚 Swagger docs missing, generating..."; \
		if ! command -v swag > /dev/null 2>&1; then \
			echo "⚠️  swag not found. Installing..."; \
			go install github.com/swaggo/swag/cmd/swag@latest; \
		fi; \
		swag init -g cmd/server/main.go -o api/docs --parseDependency --parseInternal -q; \
	fi

run: swag-ensure ## Run the server (default port 6532)
	@echo "🚀 Starting server..."
	@go run cmd/server/main.go

build: swag-ensure ## Build the server binary
	@echo "🔨 Building server..."
	@go build -o bin/pulse-api cmd/server/main.go
	@echo "✅ Binary created: bin/pulse-api"

build-debug: swag-ensure ## Build debug binary with optimizations disabled
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

lint: swag-ensure ## Run linters (requires golangci-lint)
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
	@echo "📖 Access at: http://localhost:6532/swagger/index.html"

swag-clean: ## Clean generated Swagger documentation
	@echo "🧹 Cleaning Swagger docs..."
	@rm -f api/docs/docs.go api/docs/swagger.json api/docs/swagger.yaml
	@echo "✅ Swagger docs cleaned"

swag-force: deps swag-clean swag ## Force rebuild Swagger docs (clean + generate)

##@ Embedded Frontend

# Path to the sibling frontend module and the embedded copy inside this binary.
FRONTEND_DIR := ../frontend
WEB_DIST_DIR := web/dist

web-build: ## Build the frontend and copy its dist into the embedded web/dist
	@echo "🌐 Building frontend..."
	@cd $(FRONTEND_DIR) && npm ci && VITE_API_BASE_URL= npm run build
	@echo "📦 Copying dist into $(WEB_DIST_DIR)..."
	@mkdir -p $(WEB_DIST_DIR)
	@cp -r $(FRONTEND_DIR)/dist/. $(WEB_DIST_DIR)/
	@echo "✅ Frontend embedded into $(WEB_DIST_DIR)"

##@ Database Management
setup-test-db: ## Start test database container
	@echo "🐳 Starting test database..."
	docker compose -f ../deploy/docker/docker-compose.test.yml up -d postgres-test
	@echo "⏳ Waiting for database to be ready..."
	@until docker compose -f ../deploy/docker/docker-compose.test.yml exec postgres-test pg_isready -U testuser > /dev/null 2>&1; do \
		echo "Waiting for database..."; \
		sleep 2; \
	done
	@echo "✅ Test database ready"

cleanup-test-db: ## Stop test database container
	@echo "🛑 Stopping test database..."
	docker compose -f ../deploy/docker/docker-compose.test.yml down

##@ Schema Migrations (golang-migrate)
# Migrations live under internal/db/migrations/ as versioned .up.sql/.down.sql
# pairs, embedded into the binary. The server applies them on startup; these
# targets are for manual operations against a running database.
MIGRATE_DSN ?= postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable
MIGRATE_DIR = internal/db/migrations

migrate-up: ## Apply pending migrations to $(MIGRATE_DSN)
	@echo "⬆️  Applying migrations to $(MIGRATE_DSN)..."
	@migrate -path $(MIGRATE_DIR) -database "$(MIGRATE_DSN)" up

migrate-down: ## Roll back the last migration on $(MIGRATE_DSN)
	@echo "⬇️  Rolling back last migration on $(MIGRATE_DSN)..."
	@migrate -path $(MIGRATE_DIR) -database "$(MIGRATE_DSN)" down 1

migrate-force: ## Force-set migration version (usage: make migrate-force V=N)
	@if [ -z "$(V)" ]; then echo "Usage: make migrate-force V=<version>"; exit 1; fi
	@migrate -path $(MIGRATE_DIR) -database "$(MIGRATE_DSN)" force $(V)

migrate-version: ## Show current migration version
	@migrate -path $(MIGRATE_DIR) -database "$(MIGRATE_DSN)" version

migrate-create: ## Create a new migration pair (usage: make migrate-create NAME=add_foo)
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=<name>"; exit 1; fi
	@next=$$(ls $(MIGRATE_DIR) | grep -oE '^[0-9]+' | sort -n | tail -1); \
	next=$$((next+1)); \
	name=$$(printf "%04d_%s" $$next "$(NAME)"); \
	touch $(MIGRATE_DIR)/$${name}.up.sql $(MIGRATE_DIR)/$${name}.down.sql; \
	echo "✅ Created $(MIGRATE_DIR)/$${name}.up.sql and .down.sql"

##@ Testing
test-unit: swag-ensure ## Run unit tests only (short mode)
	@echo "🧪 Running unit tests..."
	go test -short -p 1 ./...

test-integration: swag-ensure ## Run integration tests only
	@echo "🧪 Running integration tests..."
	go test -p 1 ./tests/integration/...

test: setup-test-db test-unit test-integration cleanup-test-db ## Run full test suite (setup + tests + cleanup)
	@echo "✅ All tests passed"

test-quick: swag-ensure ## Run all tests (assumes DB is running)
	@echo "🧪 Quick test (assumes DB is running)..."
	go test -p 1 ./...

test-coverage: swag-ensure ## Run unit tests with coverage (report-only, no threshold)
	@echo "🧪 Running unit tests with coverage (report-only)..."
	@# Scope to internal/ + cmd/ so this runs without a test database. The
	@# integration suite (tests/integration) needs a live Postgres and is
	@# excluded from this report-only target.
	go test -short -p 1 -coverprofile=coverage.out ./internal/... ./cmd/...
	@go tool cover -func=coverage.out | tail -1
	@echo "📊 Full report: go tool cover -html=coverage.out (artifacts: coverage.out, git-ignored)"

# COVERAGE_BASELINE is the ratchet floor for the internal/ unit-test coverage
# total (statements). Run `make test-coverage-check` locally before raising the
# floor; bump this number only after improving coverage.
# Lowered from 31.0 → 28.0 with the Viper migration (refactor-pulse-config-viper):
# removing ~210 lines of hand-written env-binding boilerplate (mergeFromEnv)
# shrank the covered-statement denominator; config-package coverage itself rose
# to ~80%. Re-raise as new code lands.
COVERAGE_BASELINE := 28.0

test-coverage-check: swag-ensure ## Run coverage and fail if it dropped below the baseline floor
	@echo "🧪 Checking coverage against baseline ($(COVERAGE_BASELINE)%)..."
	@go test -short -p 1 -coverprofile=coverage.out ./internal/... ./cmd/... >/dev/null
	@total=$$(go tool cover -func=coverage.out | tail -1 | grep -oE '[0-9]+\.[0-9]+' | tail -1); \
	echo "   current: $${total}%   baseline: $(COVERAGE_BASELINE)%"; \
	awk -v cur="$$total" -v base="$(COVERAGE_BASELINE)" 'BEGIN { exit !(cur >= base) }' || { \
		echo "::error::Coverage regression: $${total}% < baseline $(COVERAGE_BASELINE)%. Improve coverage or lower the floor deliberately."; \
		exit 1; \
	}; \
	echo "✅ Coverage $${total}% meets the $(COVERAGE_BASELINE)% baseline."
