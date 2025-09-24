# Production Dockerfile per CLAUDE.md

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-s -w -X 'main.Version=1.0.0' -X 'main.BuildTime=$(date -u +%Y%m%d-%H%M%S)'" \
    -o casnotes ./cmd/casnotes

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl \
    && addgroup -S casnotes \
    && adduser -D -S -s /bin/sh -G casnotes casnotes

# Copy binary
COPY --from=builder /app/casnotes /usr/local/bin/casnotes
RUN chmod +x /usr/local/bin/casnotes

# Create data directory
RUN mkdir -p /data && chown casnotes:casnotes /data

# Switch to non-root user
USER casnotes

# Set environment
ENV DATA_DIR=/data
ENV PORT=64123
ENV BIND=0.0.0.0

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:64123/healthz || exit 1

# Expose port
EXPOSE 64123

# Volume
VOLUME ["/data"]

# Run
CMD ["casnotes"]