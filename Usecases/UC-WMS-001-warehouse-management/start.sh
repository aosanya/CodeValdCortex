#!/bin/bash

# Warehouse Management System - Startup Script
# This script starts the CodeValdCortex framework with UC-WMS-001 configuration

set -e  # Exit on error

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting UC-WMS-001 Warehouse Management System${NC}"

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Load environment variables
if [ -f .env ]; then
    echo -e "${GREEN}Loading environment variables from .env${NC}"
    set -a
    source .env
    set +a
else
    echo -e "${RED}ERROR: .env file not found${NC}"
    exit 1
fi

# Set use case configuration directory
export USECASE_CONFIG_DIR="$SCRIPT_DIR"

# Find framework binary
FRAMEWORK_DIR="${FRAMEWORK_DIR:-/workspaces/CodeValdCortex}"
BINARY="$FRAMEWORK_DIR/bin/codevaldcortex"

if [ ! -f "$BINARY" ]; then
    echo -e "${RED}ERROR: CodeValdCortex binary not found at $BINARY${NC}"
    echo -e "${YELLOW}Please run 'make build' in the framework directory${NC}"
    exit 1
fi

# Check if ArangoDB is accessible
echo -e "${GREEN}Checking ArangoDB connection...${NC}"
if ! curl -s "http://${CVXC_DATABASE_HOST}:${CVXC_DATABASE_PORT}/_api/version" > /dev/null 2>&1; then
    echo -e "${RED}WARNING: Cannot connect to ArangoDB at ${CVXC_DATABASE_HOST}:${CVXC_DATABASE_PORT}${NC}"
    echo -e "${YELLOW}Make sure ArangoDB is running before starting the application${NC}"
fi

# Display configuration
echo -e "${GREEN}Configuration:${NC}"
echo -e "  Use Case: ${USECASE_NAME}"
echo -e "  Database: ${CVXC_DATABASE_DATABASE}"
echo -e "  Server Port: ${CVXC_SERVER_PORT}"
echo -e "  Config Directory: ${USECASE_CONFIG_DIR}"
echo ""

# Check for agent configurations
AGENT_COUNT=$(find "$SCRIPT_DIR/config/agents" -name "*.json" 2>/dev/null | wc -l)
echo -e "${GREEN}Found ${AGENT_COUNT} agent type configurations${NC}"

if [ "$AGENT_COUNT" -eq 0 ]; then
    echo -e "${YELLOW}WARNING: No agent configurations found in config/agents/${NC}"
fi

# Start the application
echo -e "${GREEN}Starting CodeValdCortex framework...${NC}"
echo ""

exec "$BINARY"
