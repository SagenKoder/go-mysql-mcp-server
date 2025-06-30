#!/bin/bash

# Build all three executables
echo "Building mysql-mcp-http..."
go build -o mysql-mcp-http ./cmd/mysql-mcp-http

echo "Building mysql-mcp-interactive..."
go build -o mysql-mcp-interactive ./cmd/mysql-mcp-interactive

echo "Building mysql-mcp-stdio..."
go build -o mysql-mcp-stdio ./cmd/mysql-mcp-stdio

echo "All builds completed!"

# Optional: build Docker images
if [ "$1" = "docker" ]; then
    echo "Building Docker images..."
    
    # Build stdio version (default)
    echo "Building stdio version..."
    docker build --build-arg MODE=stdio -t mysql-mcp:stdio .
    docker tag mysql-mcp:stdio mysql-mcp:latest
    
    # Build HTTP version
    echo "Building HTTP version..."
    docker build --build-arg MODE=http -t mysql-mcp:http .
    
    # Build interactive version
    echo "Building interactive version..."
    docker build --build-arg MODE=interactive -t mysql-mcp:interactive .
    
    echo "Docker builds completed!"
    echo "Available images:"
    echo "  - mysql-mcp:stdio (also tagged as mysql-mcp:latest)"
    echo "  - mysql-mcp:http"
    echo "  - mysql-mcp:interactive"
fi