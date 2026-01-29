.PHONY: build install test clean run-init run-backtest run-interactive help

# Build the CLI binary
build:
	@echo "🔨 Building Wisp CLI..."
	@go build -o wisp .
	@echo "✅ Build complete: ./wisp"

# Install the CLI to $GOPATH/bin
install:
	@echo "📦 Installing Wisp CLI..."
	@go build -o $(shell go env GOPATH)/bin/wisp .
	@echo "✅ Installed to $(shell go env GOPATH)/bin/wisp"
	@echo "💡 Run 'wisp --help' to get started"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test ./... -v

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -f wisp
	@rm -rf dist/
	@echo "✅ Clean complete"

# Run init command with project name (usage: make run-init PROJECT=my-project)
run-init: build
	@if [ -z "$(PROJECT)" ]; then \
		echo "❌ Error: PROJECT is required"; \
		echo "Usage: make run-init PROJECT=my-project"; \
		exit 1; \
	fi
	@echo "🚀 Running wisp init $(PROJECT)..."
	@./wisp init $(PROJECT)

# Run backtest command
run-backtest: build
	@echo "🚀 Running wisp backtest..."
	@./wisp backtest

# Run interactive backtest
run-interactive: build
	@echo "🚀 Running wisp backtest --interactive..."
	@./wisp backtest --interactive

# Run dry-run
run-dry: build
	@echo "🚀 Running wisp backtest --dry-run..."
	@./wisp backtest --dry-run

# Tidy dependencies
tidy:
	@echo "📦 Tidying dependencies..."
	@go mod tidy
	@echo "✅ Dependencies tidied"

# Format code
fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Run linter
lint:
	@echo "🔍 Running linter..."
	@golangci-lint run
	@echo "✅ Linting complete"

# Show help
help:
	@echo "Wisp CLI - Makefile targets:"
	@echo ""
	@echo "  build              Build the CLI binary"
	@echo "  install            Install to \$$GOPATH/bin"
	@echo "  test               Run tests"
	@echo "  clean              Clean build artifacts"
	@echo "  run-init           Run wisp init (usage: make run-init PROJECT=my-project)"
	@echo "  run-backtest       Run wisp backtest"
	@echo "  run-interactive    Run wisp backtest --interactive"
	@echo "  run-dry            Run wisp backtest --dry-run"
	@echo "  tidy               Tidy go.mod dependencies"
	@echo "  fmt                Format code"
	@echo "  lint               Run linter"
	@echo "  help               Show this help message"
	@echo ""

