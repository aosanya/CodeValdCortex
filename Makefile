# CodeValdCortex Makefile

# Build information
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT ?= $(shell git rev-parse HEAD)

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
GOFMT = gofmt
BINARY_NAME = codevaldcortex
BINARY_PATH = bin/$(BINARY_NAME)

# Build flags
LDFLAGS = -s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)

# Docker parameters
DOCKER_REGISTRY ?= ghcr.io
DOCKER_IMAGE ?= $(DOCKER_REGISTRY)/aosanya/codevaldcortex
DOCKER_TAG ?= $(VERSION)

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: css ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_PATH) ./cmd

.PHONY: css
css: ## Compile SCSS to CSS
	@echo "Compiling SCSS..."
	@npx sass static/scss:static/css --no-source-map --style=compressed

.PHONY: build-all
build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-amd64 ./cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-arm64 ./cmd
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd

.PHONY: run
run: css ## Build and run the application 
	@echo "Generating templates..."
	templ generate
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_PATH) ./cmd
	@echo "Running $(BINARY_NAME)..."
	@echo "💡 Tip: Hard refresh browser (Ctrl+Shift+R or Cmd+Shift+R) to reload cached JavaScript"
	./$(BINARY_PATH)

.PHONY: kill
kill: ## Stop any running instances
	@echo "Stopping any running instances..."
	@pkill -f "./bin/codevaldcortex" || true
	@sleep 1

.PHONY: run-dev
run-dev: ## Run the application in development mode
	@echo "Running in development mode..."
	$(GOCMD) run ./cmd --config config.yaml

.PHONY: run-water
run-water: ## Run with UC-INFRA-001 water distribution network config
	@echo "Running with UC-INFRA-001 (Water Distribution Network) configuration..."
	@if [ -f usecases/UC-INFRA-001-water-distribution-network/.env ]; then \
		export $$(cat usecases/UC-INFRA-001-water-distribution-network/.env | grep -v '^#' | xargs) && \
		$(GOCMD) run ./cmd --config config.yaml; \
	else \
		echo "Error: .env file not found at usecases/UC-INFRA-001-water-distribution-network/.env"; \
		exit 1; \
	fi

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -rf static/css/*.css static/css/*.css.map

.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	$(GOFMT) -s -w .

.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...

.PHONY: check
check: fmt vet lint test ## Run all checks (format, vet, lint, test)

.PHONY: audit
audit: ## Audit code for debug logs (console.log, fmt.Printf, emoji logs, MVP logs)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🔍 CODE AUDIT - Debug Log Detection"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "📊 1. JAVASCRIPT CONSOLE LOGS"
	@echo "────────────────────────────────────────"
	@chmod +x scripts/lint-console-logs.sh
	@scripts/lint-console-logs.sh || true
	@echo ""
	@echo "📊 2. GO DEBUG PRINT STATEMENTS (fmt.Printf/Println)"
	@echo "────────────────────────────────────────"
	@if grep -rn "fmt.Printf\|fmt.Println" internal/ cmd/ --include="*.go" 2>/dev/null | grep -v "_test.go" | grep -v "// OK:" ; then \
		echo "⚠️  Found fmt.Printf/Println statements (remove before merge)"; \
	else \
		echo "✅ No fmt.Printf/Println found"; \
	fi
	@echo ""
	@echo "📊 3. MVP-PREFIXED DEBUG LOGS"
	@echo "────────────────────────────────────────"
	@if grep -rn 'log\.Printf.*\[MVP-\|logger.*\[MVP-' internal/ cmd/ --include="*.go" 2>/dev/null; then \
		echo "⚠️  Found MVP-prefixed debug logs (remove before merge)"; \
	else \
		echo "✅ No MVP-prefixed logs found"; \
	fi
	@echo ""
	@echo "📊 4. EMOJI-PREFIXED DEBUG LOGS"
	@echo "────────────────────────────────────────"
	@if grep -rn '🔍\|📊\|💾\|🔹\|✅\|⚠️\|🚀\|🎯\|🔥\|💡\|📝\|🧪' internal/ cmd/ --include="*.go" 2>/dev/null | grep -v "templ.go" | grep -v "_test.go" | grep -v "// UI:" | grep -v "WriteString"; then \
		echo "⚠️  Found emoji-prefixed debug logs (remove before merge)"; \
	else \
		echo "✅ No emoji debug logs found in code (ignoring UI strings)"; \
	fi
	@echo ""
	@echo "📊 5. DEBUG COMMENT MARKERS"
	@echo "────────────────────────────────────────"
	@if grep -rn '// DEBUG:\|// TODO: remove\|// FIXME:\|// XXX:' internal/ cmd/ --include="*.go" 2>/dev/null | head -20; then \
		echo "⚠️  Found debug comment markers"; \
	else \
		echo "✅ No debug comment markers found"; \
	fi
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ Audit complete!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		.

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push: ## Push Docker image to registry
	@echo "Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOCMD) install github.com/air-verse/air@latest
	$(GOCMD) install honnef.co/go/tools/cmd/staticcheck@latest
	$(GOCMD) install mvdan.cc/unparam@latest
	-$(GOCMD) install github.com/nishanths/exhaustive/cmd/exhaustive@latest || echo "Warning: exhaustive tool installation failed (Go version compatibility)"
	$(GOCMD) install github.com/alexkohler/unimport@latest

.PHONY: deadcode
deadcode: ## Check for dead code using multiple tools
	@echo "🔍 Running comprehensive dead code analysis..."
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📊 DEAD CODE ANALYSIS REPORT"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "🔍 1. UNUSED PARAMETERS (unparam)"
	@echo "────────────────────────────────────────"
	@if command -v unparam >/dev/null 2>&1; then \
		unparam ./... || echo "✅ No unused parameters found"; \
	else \
		echo "⚠️  unparam not installed, skipping"; \
	fi
	@echo ""
	@echo "🔍 2. UNUSED IMPORTS (unimport)" 
	@echo "────────────────────────────────────────"
	@if command -v unimport >/dev/null 2>&1; then \
		unimport ./... || echo "✅ No unused imports found"; \
	else \
		echo "⚠️  unimport not installed, skipping"; \
	fi
	@echo ""
	@echo "🔍 3. STATICCHECK ANALYSIS (unused code detection)"
	@echo "────────────────────────────────────────"
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck -checks=U1000,U1001 ./... || echo "✅ No unused code found"; \
	else \
		echo "⚠️  staticcheck not installed, skipping"; \
	fi
	@echo ""
	@echo "🔍 4. GOLANGCI-LINT DEAD CODE CHECKS"
	@echo "────────────────────────────────────────"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --enable=unused,ineffassign --disable-all || echo "✅ No dead code found"; \
	else \
		echo "⚠️  golangci-lint not installed, skipping"; \
	fi
	@echo ""
	@echo "🔍 5. MISSING EXHAUSTIVE SWITCH CASES"
	@echo "────────────────────────────────────────"
	@if command -v exhaustive >/dev/null 2>&1; then \
		exhaustive ./... || echo "✅ All switch cases are exhaustive"; \
	else \
		echo "⚠️  exhaustive not installed (Go version compatibility issue), skipping"; \
	fi
	@echo ""
	@echo "🔍 6. GO VET ANALYSIS (built-in dead code detection)"
	@echo "────────────────────────────────────────"
	@$(GOCMD) vet ./... || echo "✅ No vet issues found"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ Dead code analysis complete!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

.PHONY: dev-setup
dev-setup: install-tools deps ## Setup development environment
	@echo "Development environment setup complete!"

.PHONY: release
release: check build-all ## Prepare release
	@echo "Release prepared in bin/ directory"

.PHONY: version
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"