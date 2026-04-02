# Makefile for formcodec library

GO ?= go
GOFMT ?= gofmt

COVERAGE_OUT := coverage.out
COVERAGE_HTML := coverage.html

.PHONY: all test test-race test-cover bench fmt fmt-check vet lint license-dep license-fix license-check clean check help

# Default target
all: fmt vet test ## Run fmt, vet, and test

# ============================================================================
# Testing
# ============================================================================

test: ## Run all tests
	$(GO) test -v ./...

test-race: ## Run tests with race detection
	$(GO) test -race -v ./...

test-cover: ## Run tests and generate coverage report
	$(GO) test -coverprofile=$(COVERAGE_OUT) ./...
	$(GO) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

bench: ## Run benchmarks
	$(GO) test -bench=. -benchmem ./...

# ============================================================================
# Code Quality
# ============================================================================

fmt: ## Format code
	$(GOFMT) -w .

fmt-check: ## Check code format (for CI)
	@if [ -n "$$($(GOFMT) -l .)" ]; then \
		echo "Code is not formatted. Run 'make fmt' to fix."; \
		$(GOFMT) -l .; \
		exit 1; \
	fi

vet: ## Run go vet
	$(GO) vet ./...

lint: ## Run golangci-lint
	$(GO) tool golangci-lint run ./...

# ============================================================================
# License
# ============================================================================

license-dep: ## Generate dependency license files
	$(GO) tool license-eye dep resolve  -l LICENSE

license-fix: ## Fix source file license headers
	$(GO) tool license-eye header fix

license-check: ## Check source file license headers (for CI)
	$(GO) tool license-eye header check

# ============================================================================
# Cleanup
# ============================================================================

clean: ## Clean generated files
	@rm -f $(COVERAGE_OUT) $(COVERAGE_HTML)
	@echo "Cleaned generated files."

# ============================================================================
# Combined Targets
# ============================================================================

check: fmt-check vet license-check test ## Run fmt-check, vet, license-check, and test (for CI)

# ============================================================================
# Help
# ============================================================================

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
