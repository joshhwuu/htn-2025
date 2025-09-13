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

# Check if API key is set
if [ -z "$GOOGLE_MAPS_API_KEY" ]; then
    echo "❌ GOOGLE_MAPS_API_KEY environment variable is required"
    echo
    echo "Get your API key at: https://console.cloud.google.com/apis/credentials"
    echo "Then run: export GOOGLE_MAPS_API_KEY=your-key-here"
    exit 1
fi

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