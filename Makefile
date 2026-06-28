# Root Makefile for the Node-Pulse monorepo.
#
# Convenience targets that fan out to the per-component Makefiles so a single
# command runs every completion gate locally. The component Makefiles remain
# the source of truth; this file only orchestrates them.
#
# Run `make help` for the full target list.

.DEFAULT_GOAL := help

.PHONY: help lint lint-pulse lint-beacon lint-frontend
.PHONY: build build-pulse build-beacon build-frontend
.PHONY: test test-pulse test-beacon test-frontend
.PHONY: ci-local docker-build docker-up docker-down tidy check-mods

help: ## Show all available targets
	@awk 'BEGIN {FS = ":.*##"; printf "\nNode-Pulse monorepo\n\nUsage:\n  make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

##@ Lint (all components)

lint: lint-pulse lint-beacon lint-frontend ## Lint every component

lint-pulse: ## Lint the Pulse backend
	@cd pulse && make lint

lint-beacon: ## Lint the Beacon agent
	@cd beacon && make lint

lint-frontend: ## Lint the frontend
	@cd frontend && npm run lint

##@ Build (all components)

build: build-pulse build-beacon build-frontend ## Build every component

build-pulse: ## Build the Pulse API binary (regenerates Swagger if missing)
	@cd pulse && make build

build-beacon: ## Build the Beacon agent (current platform)
	@cd beacon && make build-local

build-frontend: ## Build the frontend production bundle
	@cd frontend && npm run build

##@ Test

test: test-pulse test-beacon test-frontend ## Run every component's unit tests

test-pulse: ## Run Pulse unit tests (short mode, no DB)
	@cd pulse && make test-unit

test-beacon: ## Run Beacon tests
	@cd beacon && make test

test-frontend: ## Run frontend unit/component tests
	@cd frontend && npm run test -- --run

##@ Local CI gate

ci-local: lint build test ## Run the full local gate (lint + build + test) mirroring CI

##@ Dependency hygiene

tidy: ## Run go mod tidy on both Go modules (regenerates pulse swagger first)
	@cd pulse && make swag-ensure && go mod tidy
	@cd beacon && go mod tidy

check-mods: ## Fail if go.mod/go.sum are not tidy (run after `make tidy`)
	@echo "Checking pulse modules..."
	@cd pulse && make swag-ensure && go mod tidy && go mod verify && git diff --exit-code -- go.mod go.sum || (echo "::error::pulse go.mod/go.sum not tidy; run 'make tidy'"; exit 1)
	@echo "Checking beacon modules..."
	@cd beacon && go mod tidy && go mod verify && git diff --exit-code -- go.mod go.sum || (echo "::error::beacon go.mod/go.sum not tidy; run 'make tidy'"; exit 1)

##@ Docker

docker-build: ## Build the production images (pulse embeds the frontend)
	docker build -t node-pulse-api    -f pulse/Dockerfile  .
	docker build -t node-pulse-beacon -f beacon/Dockerfile ./beacon

docker-up: ## Start the production stack (requires .env; see .env.example)
	docker compose -f docker-compose.prod.yml up -d --build

docker-down: ## Stop the production stack
	docker compose -f docker-compose.prod.yml down
