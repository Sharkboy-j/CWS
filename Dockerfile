# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary, useful for Alpine
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cws ./main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Copy binary from builder
COPY --from=builder /build/cws .

# Copy migrations directory
COPY --from=builder /build/internal/storage/migrations ./internal/storage/migrations

# Set ownership of all files
RUN chown -R appuser:appuser /app

USER appuser

# Expose port if needed (currently not used, but can be useful)
# EXPOSE 8080

# Run the application
CMD ["./cws"]

