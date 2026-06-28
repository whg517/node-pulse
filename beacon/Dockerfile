# Production Dockerfile for the Beacon monitoring agent.
# Multi-stage: builds a CGO-free static binary, then ships it on a minimal
# Alpine base. Alpine (rather than scratch) is used because the agent writes
# rotating log files via lumberjack by default (/var/log/beacon/beacon.log),
# which needs a writable filesystem and a mkdir at startup.
#
# Build context: ./beacon   (run from repo root: docker build -f beacon/Dockerfile beacon/)
# Configuration: mount/write beacon.yaml at /etc/node-pulse/beacon.yaml
#                (path overridable via BEACON_CONFIG_PATH)
# Logs: written to /var/log/beacon (mount a volume to persist beyond restarts)

# ---- build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

# git for module fetches; ca-certificates/tzdata carried into runtime
RUN apk add --no-cache git ca-certificates tzdata

# Cache module downloads
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary (matches `make build` flags). Prometheus metrics port 2112
# is documented via EXPOSE below.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-extldflags '-static'" \
    -buildvcs=false \
    -o /out/beacon ./cmd/beacon

# ---- runtime stage ----
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

# Pre-create the default log directory (lumberjack needs it to exist).
RUN mkdir -p /var/log/beacon

COPY --from=builder /out/beacon /usr/local/bin/beacon

# Beacon config: mounted at runtime. The agent reads ./beacon.yaml or the
# path in BEACON_CONFIG_PATH by default.
ENV BEACON_CONFIG_PATH=/etc/node-pulse/beacon.yaml

# Prometheus /metrics endpoint (configurable, default 2112)
EXPOSE 2112

ENTRYPOINT ["beacon"]
