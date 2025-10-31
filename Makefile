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
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_PATH) ./cmd

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
run: ## Build and run the application 
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