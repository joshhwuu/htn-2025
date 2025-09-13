#!/bin/bash

# Development script for Vancouver Trip Planner
set -e

echo "🛠️  Vancouver Trip Planner - Development Mode"
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

# Install dependencies
echo "📦 Installing dependencies..."
go mod tidy
echo

# Run tests first
echo "🧪 Running tests..."
./scripts/test.sh
echo

# Build the application
echo "🔨 Building application..."
go build -o vancouver-trip-planner ./cmd/
echo "✅ Built: vancouver-trip-planner"
echo

# Check if API key is set
if [ -z "$GOOGLE_MAPS_API_KEY" ]; then
    echo "⚠️  Warning: GOOGLE_MAPS_API_KEY not set"
    echo "   Some features may not work without a valid API key"
    echo "   Get yours at: https://console.cloud.google.com/apis/credentials"
    echo
fi

echo "🚀 Ready for development!"
echo
echo "Commands:"
echo "  ./scripts/run.sh    - Start the server"
echo "  ./scripts/test.sh   - Run tests"
echo "  ./vancouver-trip-planner - Run built binary"