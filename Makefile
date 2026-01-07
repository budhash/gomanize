# Gomanize - Hindi Transliteration Library
# Development workflow Makefile

.PHONY: help init hooks hooks-update build version test test-quick test-verbose test-cover test-dakshina test-analysis bench clean fmt fmt-check vet lint lint-fix check dev ci install run download-datasets

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet
GOMOD := $(GOCMD) mod
BINARY := gomanize

# Version info (injected at build time)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.date=$(BUILD_TIME) -X main.commit=$(GIT_COMMIT)"

help: ## Show this help message
	@echo "Gomanize - Available Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================================================
# Setup
# ============================================================================

init: ## Initialize development environment (first-time setup)
	@echo "Initializing development environment..."
	@$(GOMOD) download
	@echo "Installing pre-commit hooks..."
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit install; \
		echo "✓ Pre-commit hooks installed"; \
	else \
		echo "⚠ pre-commit not found. Install with: pip install pre-commit"; \
		echo "  Then run: make hooks"; \
	fi
	@echo "✓ Initialization complete"

hooks: ## Install pre-commit hooks
	@echo "Installing pre-commit hooks..."
	@pre-commit install
	@echo "✓ Hooks installed"

hooks-update: ## Update pre-commit hook versions
	@echo "Updating pre-commit hooks..."
	@pre-commit autoupdate
	@echo "✓ Hooks updated"

# ============================================================================
# Build
# ============================================================================

build: ## Build the gomanize binary
	@echo "Building $(BINARY) $(VERSION)..."
	@$(GOBUILD) $(LDFLAGS) -o $(BINARY) ./cmd/main.go
	@echo "✓ Build complete: ./$(BINARY)"

version: ## Show version info
	@echo "Version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git commit: $(GIT_COMMIT)"

install: build ## Install gomanize to GOPATH/bin
	@echo "Installing $(BINARY)..."
	@$(GOCMD) install ./cmd/...
	@echo "✓ Install complete"

# ============================================================================
# Testing
# ============================================================================

test: ## Run all tests
	@echo "Running all tests..."
	@$(GOTEST) ./... -v
	@echo "✓ Tests complete"

test-quick: ## Run tests without verbose output
	@$(GOTEST) ./...

test-verbose: ## Run tests with verbose output and coverage
	@echo "Running tests with coverage..."
	@$(GOTEST) ./... -v -cover
	@echo "✓ Tests complete"

test-cover: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "✓ Coverage report: coverage.out"

test-unit: ## Run unit tests only (fast, targeted)
	@echo "Running unit tests..."
	@$(GOTEST) ./internal/legacy_lang/... -v -run "^TestUnit"
	@echo "✓ Unit tests complete"

test-integration: ## Run integration tests (full datasets)
	@echo "Running integration tests..."
	@$(GOTEST) ./internal/legacy_lang/... -v -run "^TestIntegration"
	@echo "✓ Integration tests complete"

test-dakshina: ## Run Dakshina accuracy test
	@echo "Running Dakshina accuracy test..."
	@$(GOTEST) ./internal/legacy_lang/... -v -run "TestIntegrationDakshinaAccuracy"
	@echo "✓ Dakshina test complete"

test-analysis: ## Run failure analysis (shows breakdown of issues)
	@echo "Running failure pattern analysis..."
	@$(GOTEST) ./internal/legacy_lang/... -v -run "TestIntegrationFailureAnalysis"

test-original: ## Run original test suite (hindi-common.txt)
	@echo "Running original test suite..."
	@$(GOTEST) ./internal/legacy_lang/... -v -run "TestIntegrationOriginalTestSuite"

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	@$(GOTEST) ./internal/legacy_lang/... -bench=. -benchmem
	@echo "✓ Benchmarks complete"

# ============================================================================
# Code Quality
# ============================================================================

fmt: ## Format code
	@echo "Formatting code..."
	@$(GOFMT) ./...
	@echo "✓ Format complete"

fmt-check: ## Check if code is formatted (CI use)
	@echo "Checking code formatting..."
	@test -z "$$(gofmt -l .)" || (echo "Code not formatted. Run 'make fmt'" && gofmt -l . && exit 1)
	@echo "✓ Format check complete"

vet: ## Run go vet
	@echo "Running go vet..."
	@$(GOVET) ./...
	@echo "✓ Vet complete"

lint: ## Run linter (requires golangci-lint)
	@echo "Running golangci-lint..."
	@golangci-lint run || (echo "Install golangci-lint: https://golangci-lint.run/usage/install/"; exit 1)
	@echo "✓ Lint complete"

lint-fix: ## Run linter with auto-fix
	@echo "Running golangci-lint with fixes..."
	@golangci-lint run --fix
	@echo "✓ Lint fix complete"

check: fmt vet ## Run all checks (format + vet)
	@echo "✓ All checks complete"

# ============================================================================
# Development Workflows
# ============================================================================

dev: check test ## Full development workflow (format, vet, test)
	@echo "✓ Development workflow complete"

dev-quick: fmt test-quick ## Quick development workflow
	@echo "✓ Quick development workflow complete"

ci: fmt-check lint build test-cover ## Full CI pipeline (format, lint, build, test)
	@echo "✓ CI pipeline complete"

# ============================================================================
# Data Management
# ============================================================================

download-datasets: ## Download transliteration datasets (Dakshina)
	@echo "Downloading datasets..."
	@mkdir -p datasets
	@if [ ! -f datasets/dakshina.tar ]; then \
		echo "Downloading Dakshina dataset (1.9GB)..."; \
		curl -L -o datasets/dakshina.tar "https://storage.googleapis.com/gresearch/dakshina/dakshina_dataset_v1.0.tar"; \
		echo "Extracting..."; \
		cd datasets && tar -xf dakshina.tar; \
	else \
		echo "Dakshina dataset already downloaded"; \
	fi
	@echo "✓ Datasets ready"

setup-testdata: download-datasets ## Setup test data from datasets
	@echo "Setting up test data..."
	@mkdir -p testbed/dakshina
	@if [ -f datasets/dakshina_dataset_v1.0/hi/lexicons/hi.translit.sampled.train.tsv ]; then \
		awk -F'\t' '$$3 >= 4 {print $$1"\t"$$2"\t"$$3}' \
			datasets/dakshina_dataset_v1.0/hi/lexicons/hi.translit.sampled.train.tsv \
			> testbed/dakshina/all_high_conf.tsv; \
		echo "Created testbed/dakshina/all_high_conf.tsv ($$(wc -l < testbed/dakshina/all_high_conf.tsv) entries)"; \
	fi
	@echo "✓ Test data ready"

# ============================================================================
# Utilities
# ============================================================================

run: build ## Run gomanize with arguments (usage: make run ARGS="नमस्ते")
	@./$(BINARY) $(ARGS)

demo: build ## Demo transliteration with sample words
	@echo "=== Gomanize Demo ==="
	@echo ""
	@for word in "नमस्ते" "भारत" "हिंदी" "प्रकाश" "क्षत्रिय" "ज्ञान"; do \
		result=$$(./$(BINARY) "$$word"); \
		echo "$$word → $$result"; \
	done

clean: ## Clean build artifacts and caches
	@echo "Cleaning..."
	@rm -f $(BINARY)
	@rm -rf .cache
	@$(GOCMD) clean -cache -testcache
	@echo "✓ Clean complete"

clean-all: clean ## Clean everything including downloaded datasets
	@echo "Cleaning datasets..."
	@rm -rf datasets/
	@echo "✓ Full clean complete"

tidy: ## Tidy go modules
	@echo "Tidying modules..."
	@$(GOMOD) tidy
	@echo "✓ Tidy complete"

# ============================================================================
# Status / Info
# ============================================================================

status: ## Show current test status summary
	@echo "=== Gomanize Status ==="
	@echo ""
	@echo "Build:"
	@$(GOBUILD) ./... 2>&1 && echo "  ✓ Compiles successfully" || echo "  ✗ Build errors"
	@echo ""
	@echo "Tests:"
	@$(GOTEST) ./... -count=1 2>&1 | tail -5
	@echo ""
	@echo "Dakshina accuracy:"
	@$(GOTEST) ./internal/legacy_lang/... -v -run "TestAnalyzeAllFailures" 2>&1 | grep -E "(Passed|Failed|Total):" | head -3

.DEFAULT_GOAL := help
