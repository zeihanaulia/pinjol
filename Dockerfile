# ================================
# Build Stage
# ================================
FROM golang:1.24.5-alpine AS builder

# Install build dependencies for SQLite (CGO)
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /build

# Copy dependency files first for better caching
COPY go.mod go.sum ./

# Download dependencies (cached if go.mod/go.sum unchanged)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
# - CGO_ENABLED=1 for SQLite support
# - Static linking for distroless compatibility
# - Strip debug info and symbols
RUN CGO_ENABLED=1 GOOS=linux go build \
    -a \
    -installsuffix cgo \
    -ldflags="-s -w -extldflags '-static' -X main.version=${GIT_SHA:-dev} -X main.buildTime=$(date -u +%FT%TZ)" \
    -o pinjol \
    .

# ================================
# Runtime Stage (Alpine)
# ================================
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create app directory
RUN mkdir -p /app

# Copy the statically linked binary
COPY --from=builder /build/pinjol /pinjol

# Environment variables
ENV PORT=8080 \
    APP_ENV=prod \
    DATABASE_PATH=/app/pinjol.db

# Expose port
EXPOSE 8080

# Health check (using the binary itself)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/pinjol", "serve", "--help"] || exit 1

# Run as non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup
USER appuser

# Set the entrypoint and default command
ENTRYPOINT ["/pinjol"]
CMD ["serve"]
