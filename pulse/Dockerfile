# Multi-stage build for the Pulse API with the embedded frontend.
#
# Build context MUST be the repository root (not ./pulse), because stage 1
# needs to copy the frontend sources:
#   docker build -f pulse/Dockerfile .
#
# Stage 1 (web): build the Vite production bundle.
# Stage 2 (go):  embed the bundle, regenerate Swagger, build the static binary.
# Stage 3:       minimal Alpine runtime.
#
# The frontend is //go:embed'd into the binary, so the final image serves both
# the SPA and the API from a single process — no nginx or separate frontend
# container is needed.

# ---- stage 1: build the frontend bundle ----
FROM node:25-alpine AS web
WORKDIR /web
# .npmrc carries legacy-peer-deps (react-simple-maps under React 19).
COPY frontend/package*.json frontend/.npmrc ./
RUN npm ci
COPY frontend/ ./
# Same-origin relative API URLs: the binary serves /api itself, so the SPA must
# not embed an absolute host (the .env.production default is localhost:6532).
RUN VITE_API_BASE_URL= npm run build

# ---- stage 2: build the Go binary with the embedded frontend ----
FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY pulse/go.mod pulse/go.sum ./
RUN go mod download

# Swagger generator (api/docs is git-ignored but imported by cmd/server).
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY pulse/ ./

# Place the built frontend where the web package expects it (web/dist), then
# regenerate Swagger docs.
COPY --from=web /web/dist ./web/dist
RUN swag init -g cmd/server/main.go -o api/docs --parseDependency --parseInternal -q

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -buildvcs=false -o pulse ./cmd/server

# ---- stage 3: runtime ----
FROM alpine:3.19
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /app/pulse .

EXPOSE 6532
HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null http://localhost:6532/api/v1/health || exit 1

CMD ["./pulse"]
