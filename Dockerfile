# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o nfl-scores

# Final stage
FROM alpine:3.18

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/nfl-scores .

# Create directory for NATS store
RUN mkdir -p /tmp/nats-store

# Run as non-root user
RUN adduser -D appuser
USER appuser

# Expose Prometheus metrics port
EXPOSE 2112

# Run the application
CMD ["./nfl-scores"] 