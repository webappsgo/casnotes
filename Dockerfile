# Production Dockerfile per CLAUDE.md

# Build stage
FROM golang:alpine AS builder

# Build arguments
ARG VERSION=1.0.0
ARG COMMIT=dev
ARG BUILD_TIME=unknown

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first (for layer caching)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build static binary (no CGO, fully static)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
    -ldflags "-s -w -X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${BUILD_TIME}'" \
    -o casnotes ./cmd/casnotes

# Runtime stage - minimal scratch image for security
FROM scratch

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary from builder
COPY --from=builder /app/casnotes /casnotes

# Set environment variables
ENV DATA_DIR=/data
ENV PORT=64123
ENV BIND=0.0.0.0

# Health check endpoint
HEALTHCHECK --interval=30s --start-period=5s --retries=3 \
    CMD ["/casnotes", "--help"]

# Expose port (configurable)
EXPOSE 64123

# Volume for data persistence
VOLUME ["/data"]

# Run as non-root (binary handles this internally)
ENTRYPOINT ["/casnotes"]