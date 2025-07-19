# Build stage
FROM golang:1.24-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webserver ./cmd/webserver

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh webserveruser

# Set working directory
WORKDIR /app

# Create necessary directories
RUN mkdir -p configs static

# Copy binary from builder stage
COPY --from=builder /app/webserver .

# Copy configuration files
COPY --from=builder /app/configs/ ./configs/

# Copy static files (if any)
COPY --from=builder /app/static/ ./static/

# Change ownership to non-root user
RUN chown -R webserveruser:webserveruser /app

# Switch to non-root user
USER webserveruser

# Expose port 8080
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/stats || exit 1

# Set default command
CMD ["./webserver"] 