# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Install swag to regenerate the Swagger docs from annotations. The generated
# api/docs/*.go files are git-ignored (they are build artifacts), but
# cmd/server imports the package, so they must exist before `go build`.
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy source code
COPY . .

# Regenerate Swagger docs (no-op effect if already present; idempotent).
RUN swag init -g cmd/server/main.go -o api/docs --parseDependency --parseInternal -q

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -buildvcs=false -o pulse ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata wget

# Copy binary from builder
COPY --from=builder /app/pulse .

# Expose port
EXPOSE 6532

# Health check
HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null http://localhost:6532/api/v1/health || exit 1

# Run the binary
CMD ["./pulse"]
