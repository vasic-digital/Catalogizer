# Production Dockerfile for Catalogizer API
# Multi-stage build for optimized image size

# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO is enabled for SQLite support
ENV CGO_ENABLED=1
RUN go build -a -installsuffix cgo -ldflags="-w -s" -o catalogizer-api .

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 catalogizer && \
    adduser -D -u 1000 -G catalogizer catalogizer

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/catalogizer-api /app/catalogizer-api

# Copy configuration files (if any)
COPY --from=builder /build/migrations /app/migrations
COPY --from=builder /build/config /app/config

# Create necessary directories with proper permissions
RUN mkdir -p /app/data /app/logs /app/temp && \
    chown -R catalogizer:catalogizer /app

# Switch to non-root user
USER catalogizer

# Expose API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["/app/catalogizer-api"]
