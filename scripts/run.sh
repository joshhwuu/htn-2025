#!/bin/bash

# Run script for Vancouver Trip Planner
set -e

echo "🚗 Starting Vancouver Trip Planner..."
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

# API key will be loaded from .env file by the Go application

# Install dependencies
echo "📦 Installing dependencies..."
go mod tidy
echo

# Set default port if not provided
if [ -z "$PORT" ]; then
    export PORT=8080
fi

echo "🌐 Server will start on http://localhost:$PORT"
echo "📋 API docs: http://localhost:$PORT/health"
echo
echo "Press Ctrl+C to stop the server"
echo

# Start the server
go run cmd/main.go