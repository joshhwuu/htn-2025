# Vancouver Trip Planner Makefile

.PHONY: help test run dev demo build clean

help: ## Show this help message
	@echo "Vancouver Trip Planner - Available Commands:"
	@echo
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

test: ## Run all tests
	@./scripts/test.sh

run: ## Start the server (requires GOOGLE_MAPS_API_KEY)
	@./scripts/run.sh

dev: ## Setup development environment
	@./scripts/dev.sh

demo: ## Run API demo (requires server to be running)
	@./scripts/demo.sh

build: ## Build the binary
	@echo "🔨 Building vancouver-trip-planner..."
	@go build -o vancouver-trip-planner ./cmd/
	@echo "✅ Built: vancouver-trip-planner"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning up..."
	@rm -f vancouver-trip-planner
	@go clean
	@echo "✅ Cleaned up build artifacts"

install: ## Install dependencies
	@echo "📦 Installing dependencies..."
	@go mod tidy
	@echo "✅ Dependencies installed"

fmt: ## Format Go code
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

lint: ## Run linter (requires golangci-lint)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "🔍 Running linter..."; \
		golangci-lint run; \
		echo "✅ Linting completed"; \
	else \
		echo "⚠️  golangci-lint not installed. Install with:"; \
		echo "   brew install golangci-lint"; \
	fi