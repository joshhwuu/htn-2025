# Vancouver Trip Planner - Development Commands

.PHONY: help install dev backend frontend stop clean

# Default target
help:
	@echo "Vancouver Trip Planner - Available commands:"
	@echo ""
	@echo "  make install  - Install dependencies for both backend and frontend"
	@echo "  make dev      - Run both backend and frontend in development mode"
	@echo "  make backend  - Run only the backend server"
	@echo "  make frontend - Run only the frontend server"
	@echo "  make stop     - Stop all running servers"
	@echo "  make clean    - Clean build artifacts"
	@echo ""

# Install dependencies for both backend and frontend
install:
	@echo "📦 Installing backend dependencies..."
	cd app && go mod download
	@echo "📦 Installing frontend dependencies..."
	cd client && npm install
	@echo "✅ All dependencies installed!"

# Run both backend and frontend concurrently
dev:
	@echo "🚀 Starting Vancouver Trip Planner in development mode..."
	@echo "🔧 Backend will be available at: http://localhost:8080"
	@echo "🎨 Frontend will be available at: http://localhost:3000"
	@echo ""
	@echo "Press Ctrl+C to stop both servers"
	@echo ""
	@$(MAKE) -j2 backend frontend

# Run backend server
backend:
	@echo "🚗 Starting backend server..."
	cd app && make run

# Run frontend server
frontend:
	@echo "🎨 Starting frontend server..."
	cd client && npm run dev

# Stop all servers (for processes that might be running in background)
stop:
	@echo "🛑 Stopping servers..."
	-pkill -f "vancouver-trip-planner"
	-pkill -f "next dev"
	@echo "✅ Servers stopped!"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	cd app && go clean
	cd client && rm -rf .next
	@echo "✅ Clean complete!"