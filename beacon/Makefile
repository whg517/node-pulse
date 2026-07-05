.PHONY: build clean test install lint

# Build variables
BINARY_NAME=beacon
BUILD_DIR=build
GOOS=linux
GOARCH=amd64

##@ Build-time version stamping (D-G5)
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   ?= $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION_PKG := github.com/whg517/node-pulse/beacon/internal/version
VERSION_LDFLAGS := -X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).Commit=$(COMMIT) -X $(VERSION_PKG).BuildDate=$(BUILD_DATE)

# Build for Linux AMD64 static binary
build:
	@echo "Building $(BINARY_NAME) (version=$(VERSION) commit=$(COMMIT)) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags "-extldflags '-static' $(VERSION_LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		./cmd/beacon
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME) ($(VERSION) @ $(COMMIT))"

# Build for current platform
build-local:
	@echo "Building $(BINARY_NAME) for current platform (version=$(VERSION) commit=$(COMMIT))..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(VERSION_LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/beacon
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME) ($(VERSION) @ $(COMMIT))"

# Run tests
test:
	@echo "Running tests..."
	go test -v -cover ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@echo "Running golangci-lint..."
	golangci-lint run --timeout 5m

# Fix linting issues automatically
lint-fix:
	@echo "Running golangci-lint with fixes..."
	golangci-lint run --fix --timeout 5m

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install binary to /usr/local/bin (binary-only; for systemd setup see
# install-systemd, which also creates the unit + user + config dirs).
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Install complete (binary only). For systemd service: make install-systemd"

# Install as a managed systemd service (idempotent): builds the binary,
# creates the beacon user, installs the unit, seeds /etc/beacon/beacon.yaml,
# and enables the service. Closes the D-G4 deployment gap.
install-systemd:
	@echo "Installing beacon as a systemd service..."
	@sudo ./deploy/install-systemd.sh

# Run the beacon
run:
	@echo "Running $(BINARY_NAME)..."
	go run ./cmd/beacon
