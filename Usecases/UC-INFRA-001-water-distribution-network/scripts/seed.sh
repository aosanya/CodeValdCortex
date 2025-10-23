#!/bin/bash

# INFRA-007: Seed Agent Instances Script
# This script creates 27 agent instances for the water distribution network demo

set -e

# Get the absolute path to this script's directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
USECASE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
FRAMEWORK_DIR="$(cd "$USECASE_DIR/../.." && pwd)"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  INFRA-007: Create Infrastructure Agent Instances"
echo "  Seeding 27 agents for water distribution network demo"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Framework Dir:  $FRAMEWORK_DIR"
echo "Use Case Dir:   $USECASE_DIR"
echo "Script Dir:     $SCRIPT_DIR"
echo ""

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
echo "Database: ${CVXC_DATABASE_DATABASE:-codevaldcortex}"
echo ""

# Run the seeding script
cd "$FRAMEWORK_DIR"
echo "Running agent seeder..."
echo ""

go run "$SCRIPT_DIR/seed_agents.go"

echo ""
echo "✅ Complete! You can now:"
echo "  1. Start the server: cd $USECASE_DIR && ./start.sh"
echo "  2. View agents at: http://localhost:8083/agents"
echo ""
