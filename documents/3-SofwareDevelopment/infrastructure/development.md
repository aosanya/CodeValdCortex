# CodeValdCortex - Development Environment

## Overview

The development environment provides a complete local setup for CodeValdCortex development, testing, and debugging. It includes containerized infrastructure services, development tools, and automation scripts for rapid iteration and testing.

## 1. Local Development Infrastructure

### Docker Compose Development Stack

```yaml
# docker-compose.dev.yaml
version: '3.8'
services:
  # ArangoDB for local development
  arangodb:
    image: arangodb:3.11.5
    container_name: pweza-arangodb-dev
    environment:
      ARANGO_ROOT_PASSWORD: devpassword
      ARANGO_NO_AUTH: 0
    ports:
      - "8529:8529"
    volumes:
      - arangodb_data:/var/lib/arangodb3
      - arangodb_apps:/var/lib/arangodb3-apps
    networks:
      - pweza-dev
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8529/_api/version"]
      interval: 30s
      timeout: 10s
      retries: 3
  
  # Redis for caching and session storage
  redis:
    image: redis:7-alpine
    container_name: pweza-redis-dev
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - pweza-dev
    command: redis-server --appendonly yes
  
  # Prometheus for metrics collection
  prometheus:
    image: prom/prometheus:latest
    container_name: pweza-prometheus-dev
    ports:
      - "9090:9090"
    volumes:
      - ./dev/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    networks:
      - pweza-dev
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--web.enable-lifecycle'
  
  # Grafana for visualization
  grafana:
    image: grafana/grafana:latest
    container_name: pweza-grafana-dev
    ports:
      - "3000:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: devpassword
    volumes:
      - grafana_data:/var/lib/grafana
      - ./dev/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./dev/grafana/datasources:/etc/grafana/provisioning/datasources
    networks:
      - pweza-dev
  
  # Jaeger for distributed tracing
  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: pweza-jaeger-dev
    ports:
      - "16686:16686"
      - "14268:14268"
    environment:
      COLLECTOR_OTLP_ENABLED: true
    networks:
      - pweza-dev

volumes:
  arangodb_data:
  arangodb_apps:
  redis_data:
  prometheus_data:
  grafana_data:

networks:
  pweza-dev:
    driver: bridge
```

## 2. Development Environment Setup

### Setup Script

```bash
#!/bin/bash
# setup-dev-env.sh

set -e

echo "Setting up CodeValdCortex development environment..."

# Check prerequisites
check_prerequisites() {
    echo "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        echo "Error: Docker is not installed"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        echo "Error: Docker Compose is not installed"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        echo "Error: Go is not installed"
        exit 1
    fi
    
    GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
    if [[ $(echo "$GO_VERSION 1.21" | tr " " "\n" | sort -V | head -n1) != "1.21" ]]; then
        echo "Error: Go version 1.21+ is required"
        exit 1
    fi
    
    if ! command -v kubectl &> /dev/null; then
        echo "Warning: kubectl is not installed (required for Kubernetes deployment)"
    fi
    
    if ! command -v helm &> /dev/null; then
        echo "Warning: Helm is not installed (required for Kubernetes deployment)"
    fi
    
    echo "Prerequisites check completed"
}

# Setup local infrastructure
setup_infrastructure() {
    echo "Starting local infrastructure..."
    
    # Create development configuration directories
    mkdir -p dev/{prometheus,grafana/{dashboards,datasources}}
    
    # Generate Prometheus configuration
    cat > dev/prometheus.yml << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'pweza-core'
    static_configs:
      - targets: ['host.docker.internal:8080']
    metrics_path: '/metrics'
    scrape_interval: 10s

  - job_name: 'arangodb'
    static_configs:
      - targets: ['arangodb:8529']
    metrics_path: '/_admin/metrics'
    scrape_interval: 30s
EOF
    
    # Generate Grafana datasource configuration
    cat > dev/grafana/datasources/prometheus.yml << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
EOF
    
    # Start infrastructure services
    docker-compose -f docker-compose.dev.yaml up -d
    
    echo "Waiting for services to be ready..."
    sleep 30
    
    # Verify services are running
    docker-compose -f docker-compose.dev.yaml ps
    
    echo "Local infrastructure setup completed"
    echo "- ArangoDB: http://localhost:8529 (root/devpassword)"
    echo "- Prometheus: http://localhost:9090"
    echo "- Grafana: http://localhost:3000 (admin/devpassword)"
    echo "- Jaeger: http://localhost:16686"
}

# Setup Go development environment
setup_go_environment() {
    echo "Setting up Go development environment..."
    
    # Initialize Go module if not exists
    if [ ! -f go.mod ]; then
        go mod init github.com/pweza/core
    fi
    
    # Install development dependencies
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install github.com/swaggo/swag/cmd/swag@latest
    go install github.com/air-verse/air@latest
    go install github.com/vektra/mockery/v2@latest
    
    # Download project dependencies
    go mod download
    go mod tidy
    
    echo "Go development environment setup completed"
}

# Setup project structure
setup_project_structure() {
    echo "Setting up project structure..."
    
    # Create project directories
    mkdir -p {cmd,internal,pkg,api,deployments,docs,examples,scripts,tests}
    mkdir -p internal/{agents,coordination,orchestration,auth,monitoring}
    mkdir -p pkg/{client,types,utils}
    mkdir -p deployments/{helm,kubernetes,docker}
    
    echo "Project structure setup completed"
}

# Main execution
main() {
    check_prerequisites
    setup_infrastructure
    setup_go_environment
    setup_project_structure
    
    echo ""
    echo "Development environment setup completed successfully!"
    echo ""
    echo "Next steps:"
    echo "1. Run 'make dev-start' to start the development environment"
    echo "2. Run 'make build' to build the application"
    echo "3. Run 'make test' to run tests"
    echo "4. Visit http://localhost:3000 for Grafana dashboards"
    echo ""
}

# Run main function
main "$@"
```

## 3. Development Tools Configuration

### Makefile for Development Tasks

```makefile
.PHONY: help build test lint docker-build docker-push deploy clean

# Variables
APP_NAME := pweza-core
VERSION := $(shell git describe --tags --always --dirty)
REGISTRY := your-registry.com
IMAGE := $(REGISTRY)/$(APP_NAME)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(APP_NAME) version $(VERSION)"
	go build -ldflags "-X main.version=$(VERSION)" -o bin/$(APP_NAME) cmd/manager/main.go

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	golangci-lint run

docker-build: ## Build Docker image
	docker build -t $(IMAGE):$(VERSION) .
	docker tag $(IMAGE):$(VERSION) $(IMAGE):latest

docker-push: docker-build ## Push Docker image
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest

deploy: ## Deploy to Kubernetes
	helm upgrade --install $(APP_NAME) deployments/helm/pweza-core --set image.tag=$(VERSION)

clean: ## Clean build artifacts
	rm -rf bin/ dist/ coverage.out coverage.html

dev-start: ## Start development environment
	./scripts/setup-dev-env.sh
	docker-compose -f docker-compose.dev.yaml up -d

dev-stop: ## Stop development environment
	docker-compose -f docker-compose.dev.yaml down

dev-logs: ## Show development environment logs
	docker-compose -f docker-compose.dev.yaml logs -f

dev-reset: ## Reset development environment
	docker-compose -f docker-compose.dev.yaml down -v
	docker-compose -f docker-compose.dev.yaml up -d

test-integration: ## Run integration tests
	go test -v -tags=integration ./tests/integration/...

benchmark: ## Run benchmarks
	go test -v -bench=. -benchmem ./...

generate: ## Generate code (mocks, docs, etc.)
	go generate ./...
	swag init -g cmd/manager/main.go

install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/air-verse/air@latest
	go install github.com/vektra/mockery/v2@latest
```

### Air Configuration for Live Reload

```toml
# .air.toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/manager/"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "docs", "deployments"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

## 4. Testing Configuration

### Go Test Configuration

```go
// internal/test/setup.go
package test

import (
    "context"
    "fmt"
    "os"
    "testing"
    "time"

    "github.com/arangodb/go-driver"
    "github.com/arangodb/go-driver/http"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

// TestDatabase provides a test database instance
type TestDatabase struct {
    Container testcontainers.Container
    Database  driver.Database
    URL       string
}

// SetupTestDatabase creates a test ArangoDB instance
func SetupTestDatabase(t *testing.T) *TestDatabase {
    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image:        "arangodb:3.11.5",
        ExposedPorts: []string{"8529/tcp"},
        Env: map[string]string{
            "ARANGO_ROOT_PASSWORD": "testpassword",
            "ARANGO_NO_AUTH":      "0",
        },
        WaitingFor: wait.ForLog("ArangoDB (version").WithStartupTimeout(2 * time.Minute),
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        t.Fatalf("Failed to start ArangoDB container: %v", err)
    }

    host, err := container.Host(ctx)
    if err != nil {
        t.Fatalf("Failed to get container host: %v", err)
    }

    port, err := container.MappedPort(ctx, "8529")
    if err != nil {
        t.Fatalf("Failed to get container port: %v", err)
    }

    dbURL := fmt.Sprintf("http://%s:%s", host, port.Port())

    // Connect to database
    conn, err := http.NewConnection(http.ConnectionConfig{
        Endpoints: []string{dbURL},
    })
    if err != nil {
        t.Fatalf("Failed to create ArangoDB connection: %v", err)
    }

    client, err := driver.NewClient(driver.ClientConfig{
        Connection:     conn,
        Authentication: driver.BasicAuthentication("root", "testpassword"),
    })
    if err != nil {
        t.Fatalf("Failed to create ArangoDB client: %v", err)
    }

    // Create test database
    testDBName := fmt.Sprintf("test_%d", time.Now().Unix())
    db, err := client.CreateDatabase(ctx, testDBName, nil)
    if err != nil {
        t.Fatalf("Failed to create test database: %v", err)
    }

    t.Cleanup(func() {
        if err := container.Terminate(ctx); err != nil {
            t.Logf("Failed to terminate container: %v", err)
        }
    })

    return &TestDatabase{
        Container: container,
        Database:  db,
        URL:       dbURL,
    }
}

// Helper function to get test database URL from environment or container
func GetTestDatabaseURL() string {
    if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
        return url
    }
    return "http://localhost:8529" // fallback to local instance
}
```

### Integration Test Example

```go
// tests/integration/agent_test.go
//go:build integration

package integration

import (
    "context"
    "testing"
    "time"

    "github.com/pweza/core/internal/agents"
    "github.com/pweza/core/internal/test"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAgentLifecycle(t *testing.T) {
    // Setup test database
    testDB := test.SetupTestDatabase(t)
    
    // Create agent manager
    manager := agents.NewManager(agents.Config{
        Database: testDB.Database,
    })
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Test agent creation
    config := agents.AgentConfig{
        Name: "test-agent",
        Type: "processor",
        ResourceRequirements: agents.ResourceRequirements{
            CPU:    "100m",
            Memory: "256Mi",
        },
    }
    
    agent, err := manager.CreateAgent(ctx, config)
    require.NoError(t, err)
    assert.NotEmpty(t, agent.ID)
    assert.Equal(t, "test-agent", agent.Name)
    assert.Equal(t, "processor", agent.Type)
    
    // Test agent retrieval
    retrievedAgent, err := manager.GetAgent(ctx, agent.ID)
    require.NoError(t, err)
    assert.Equal(t, agent.ID, retrievedAgent.ID)
    
    // Test agent deletion
    err = manager.DeleteAgent(ctx, agent.ID)
    require.NoError(t, err)
    
    // Verify agent is deleted
    _, err = manager.GetAgent(ctx, agent.ID)
    assert.Error(t, err)
}
```

## 5. Code Quality Tools

### Linting Configuration

```yaml
# .golangci.yml
run:
  timeout: 5m
  issues-exit-code: 1
  tests: true

output:
  format: colored-line-number
  print-issued-lines: true
  print-linter-name: true

linters-settings:
  govet:
    check-shadowing: true
  gocyclo:
    min-complexity: 15
  dupl:
    threshold: 100
  goconst:
    min-len: 3
    min-occurrences: 3
  misspell:
    locale: US
  lll:
    line-length: 120
  goimports:
    local-prefixes: github.com/pweza/core
  gocritic:
    enabled-tags:
      - performance
      - style
      - experimental
    disabled-checks:
      - wrapperFunc

linters:
  enable:
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
    - typecheck
    - goimports
    - gofmt
    - gocyclo
    - goconst
    - misspell
    - lll
    - gosec
    - gocritic
  disable:
    - maligned

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gocyclo
        - errcheck
        - dupl
        - gosec
```

### Git Hooks

```bash
#!/bin/sh
# .git/hooks/pre-commit

set -e

echo "Running pre-commit checks..."

# Run linting
echo "Running linter..."
make lint

# Run tests
echo "Running tests..."
make test

# Check formatting
echo "Checking formatting..."
if [ -n "$(gofmt -l .)" ]; then
    echo "Code is not formatted. Run 'go fmt ./...' to fix."
    exit 1
fi

echo "All pre-commit checks passed!"
```

This development environment provides a complete setup for efficient CodeValdCortex development with automated testing, quality checks, and local infrastructure services.