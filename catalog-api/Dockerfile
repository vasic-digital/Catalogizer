# Production Dockerfile for Catalogizer API
# Multi-stage build for optimized image size
#
# Build from repo root:
#   podman build --network host -f catalog-api/Dockerfile -t catalogizer-api .

# Stage 1: Build (Debian-based for glibc — required by go-fitz/MuPDF)
FROM docker.io/library/golang:1.25 AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git make gcc libc6-dev && rm -rf /var/lib/apt/lists/*

# Prevent Go from auto-downloading newer toolchain
ENV GOTOOLCHAIN=local
ENV CGO_ENABLED=1

WORKDIR /build

# Copy submodules (required by go.mod replace directives)
COPY Assets/ /build/Assets/
COPY Auth/ /build/Auth/
COPY Cache/ /build/Cache/
COPY Challenges/ /build/Challenges/
COPY Concurrency/ /build/Concurrency/
COPY Config/ /build/Config/
COPY Containers/ /build/Containers/
COPY Database/ /build/Database/
COPY Discovery/ /build/Discovery/
COPY Entities/ /build/Entities/
COPY EventBus/ /build/EventBus/
COPY Filesystem/ /build/Filesystem/
COPY Lazy/ /build/Lazy/
COPY Media/ /build/Media/
COPY Memory/ /build/Memory/
COPY Middleware/ /build/Middleware/
COPY Observability/ /build/Observability/
COPY RateLimiter/ /build/RateLimiter/
COPY Recovery/ /build/Recovery/
COPY Security/ /build/Security/
COPY Storage/ /build/Storage/
COPY Streaming/ /build/Streaming/
COPY Watcher/ /build/Watcher/

# Copy catalog-api source
COPY catalog-api/go.mod catalog-api/go.sum /build/catalog-api/
WORKDIR /build/catalog-api
RUN go mod download

COPY catalog-api/ /build/catalog-api/

# Build the application
RUN go build -a -ldflags="-w -s" -o catalogizer-api .

# Stage 2: Runtime (must match builder's glibc version — both use trixie)
FROM docker.io/library/debian:trixie-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata wget && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/catalog-api/catalogizer-api /app/catalogizer-api

# Copy runtime configuration files
# Note: config.json is NOT copied — the app auto-creates defaults and reads
# auth settings (JWT_SECRET, ADMIN_USERNAME, ADMIN_PASSWORD) from env vars.
COPY --from=builder /build/catalog-api/migrations /app/migrations
COPY --from=builder /build/catalog-api/challenges/config /app/challenges/config

# Create a dedicated non-root user
RUN groupadd --system --gid 1001 appgroup && \
    useradd --system --uid 1001 --gid appgroup --no-create-home appuser

# Create necessary directories with proper permissions
RUN mkdir -p /app/data /app/logs /app/temp && \
    chown -R appuser:appgroup /app

# Expose API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Switch to non-root user
USER appuser

# Run the application
ENTRYPOINT ["/app/catalogizer-api"]
