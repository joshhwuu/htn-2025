#!/bin/bash

# Test script for Vancouver Trip Planner
set -e

echo "🧪 Running Vancouver Trip Planner Tests..."
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

# Run unit tests
echo "🔍 Running unit tests..."
go test ./internal/... ./pkg/... -v -count=1
echo

# Run integration tests if API key is provided
if [ -n "$GOOGLE_MAPS_API_KEY" ]; then
    echo "🌐 Running integration tests (with Google Maps API)..."
    go test ./test/ -v -count=1
else
    echo "⚠️  Skipping integration tests (GOOGLE_MAPS_API_KEY not set)"
    echo "   To run integration tests: export GOOGLE_MAPS_API_KEY=your-key"
fi

echo
echo "✅ All tests completed!"