# Gomanize - Hindi Transliteration Library
# Development workflow Makefile

.PHONY: help init hooks hooks-update build version wasm wasm-serve test test-quick test-verbose test-cover test-unit test-integration test-dakshina test-analysis bench benchmark clean fmt fmt-check vet lint lint-fix check dev ci install run download-datasets tasks

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
	@$(GOBUILD) $(LDFLAGS) -o $(BINARY) ./cmd/gomanize
	@echo "✓ Build complete: ./$(BINARY)"

version: ## Show version info
	@echo "Version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git commit: $(GIT_COMMIT)"

WASM_DIR := web

wasm: ## Build the WebAssembly demo into web/ (gomanize.wasm + wasm_exec.js)
	@echo "Building WebAssembly demo..."
	@GOOS=js GOARCH=wasm $(GOBUILD) -o $(WASM_DIR)/gomanize.wasm ./cmd/gomanize-wasm
	@cp "$$($(GOCMD) env GOROOT)/lib/wasm/wasm_exec.js" $(WASM_DIR)/wasm_exec.js 2>/dev/null \
		|| cp "$$($(GOCMD) env GOROOT)/misc/wasm/wasm_exec.js" $(WASM_DIR)/wasm_exec.js
	@echo "✓ WASM demo built. Serve locally with: make wasm-serve"

wasm-serve: wasm ## Build and serve the WASM demo at http://localhost:8080
	@echo "Serving $(WASM_DIR)/ at http://localhost:8080 (Ctrl-C to stop)..."
	@cd $(WASM_DIR) && python3 -m http.server 8080

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
	@$(GOTEST) -race -coverpkg=./... -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -func=coverage.out | tail -1
	@echo "✓ Coverage report: coverage.out"

test-unit: ## Run unit tests only (fast, all packages except benchmark)
	@echo "Running unit tests..."
	@$(GOTEST) $(shell $(GOCMD) list ./... | grep -v /benchmark) -count=1
	@echo "✓ Unit tests complete"

test-integration: ## Run integration tests (full Dakshina + Aksharantar datasets)
	@echo "Running integration tests..."
	@$(GOTEST) ./benchmark/... -v -run "TestBenchmarkDakshinaHindi|TestBenchmarkAksharantarHindi"
	@echo "✓ Integration tests complete"

test-dakshina: ## Run Dakshina accuracy test (curated high-confidence subset)
	@echo "Running Dakshina accuracy test..."
	@$(GOTEST) ./benchmark/... -v -run "TestBenchmarkCuratedHindi"
	@echo "✓ Dakshina test complete"

test-analysis: ## Run failure analysis (shows breakdown of issues)
	@echo "Running failure pattern analysis..."
	@$(GOTEST) ./benchmark/... -v -run "TestBenchmarkFailureAnalysis"

bench: ## Run performance benchmarks
	@echo "Running performance benchmarks..."
	@$(GOTEST) ./benchmark/... -bench=. -benchmem
	@echo "✓ Benchmarks complete"

benchmark: ## Run accuracy benchmarks (must pass threshold)
	@echo "Running accuracy benchmarks..."
	@$(GOTEST) ./benchmark/... -v -run "TestBenchmark"
	@echo "✓ Accuracy benchmarks complete"

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

ci: fmt-check lint build test-cover benchmark ## Full CI pipeline (format, lint, build, test, benchmark)
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

tasks: ## Show the task tracker (usage: make tasks ARGS="next" — see ./tools/tasks help)
	@./tools/tasks $(ARGS)

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
	@echo "Dakshina accuracy (curated):"
	@$(GOTEST) ./benchmark/... -v -run "TestBenchmarkCuratedHindi" 2>&1 | grep -E "Pure|With overrides" | head -3

.DEFAULT_GOAL := help
