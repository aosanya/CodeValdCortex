#!/bin/bash

# UC-INFRA-001: Water Distribution Network - Start Script
# This script starts the CodeValdCortex framework with this use case's configuration

set -e

# Get the absolute path to this script's directory (use case root)
USECASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Get the framework root (parent of Usecases directory)
FRAMEWORK_DIR="$(cd "$USECASE_DIR/../.." && pwd)"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  UC-INFRA-001: Water Distribution Network"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Framework Dir:  $FRAMEWORK_DIR"
echo "Use Case Dir:   $USECASE_DIR"
echo ""

# Check if framework binary exists
if [ ! -f "$FRAMEWORK_DIR/bin/codevaldcortex" ]; then
    echo "❌ Framework binary not found. Building..."
    cd "$FRAMEWORK_DIR"
    go build -o bin/codevaldcortex ./cmd/main.go
    echo "✅ Build complete"
    echo ""
fi

# Load environment variables from use case .env file
if [ -f "$USECASE_DIR/.env" ]; then
    echo "📝 Loading environment from .env"
    export $(cat "$USECASE_DIR/.env" | grep -v '^#' | xargs)
else
    echo "⚠️  No .env file found, using defaults"
fi

# Set the use case config directory
export USECASE_CONFIG_DIR="$USECASE_DIR"

echo ""
echo "Starting server..."
echo "  - Agent types will be loaded from: $USECASE_DIR/config/agents/"
echo "  - Database: ${CVXC_DATABASE_DATABASE:-codevaldcortex}"
echo "  - Server: http://${CVXC_SERVER_HOST:-0.0.0.0}:${CVXC_SERVER_PORT:-8080}"
echo ""
echo "Press Ctrl+C to stop"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Run the framework
cd "$FRAMEWORK_DIR"
exec ./bin/codevaldcortex
