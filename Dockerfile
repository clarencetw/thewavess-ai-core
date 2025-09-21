# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies (cached unless go.mod/go.sum change)
RUN go mod download

# Install swag for API documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy source code
COPY . .

# Generate Swagger docs
RUN swag init --parseDependency --parseInternal

# Set build arguments
ARG VERSION=unknown
ARG BUILD_TIME=unknown  
ARG GIT_COMMIT=unknown

# Build the application with optimizations and real build info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -a -installsuffix cgo -o main .

# Runtime stage - use minimal alpine for better security and functionality
FROM alpine:latest

# Install minimal runtime dependencies
RUN apk --no-cache add ca-certificates wget tzdata && \
    addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy docs directory
COPY --from=builder /app/docs ./docs

# Copy public directory for web interface  
COPY --from=builder /app/public ./public

# Copy templates directory for HTML rendering
COPY --from=builder /app/templates ./templates

# Copy configs directory for configuration files
COPY --from=builder /app/configs ./configs

# Set correct ownership
RUN chown -R appuser:appgroup /app

# Use non-root user
USER appuser:appgroup

# Expose port
EXPOSE 8080

# Enhanced health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/monitor/health || exit 1

# Run the application
ENTRYPOINT ["./main"]