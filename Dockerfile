# Build stage
FROM golang:1.23-alpine AS builder

# Build argument to specify which mode to build (stdio, http, or interactive)
ARG MODE=stdio

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application based on MODE argument
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/mysql-mcp-${MODE}

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -g '' appuser

# Copy the binary from builder
COPY --from=builder /app/app /usr/local/bin/mysql-mcp

# Switch to non-root user
USER appuser

# Default environment variables for MySQL connection
ENV MYSQL_HOST=localhost
ENV MYSQL_PORT=3306
ENV MYSQL_USER=root
# Note: MYSQL_PASSWORD should be provided at runtime for security

# Run the application
CMD ["/usr/local/bin/mysql-mcp"]