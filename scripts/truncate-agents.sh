#!/bin/bash

# Script to truncate all agents from ArangoDB
# This will remove all data from the agents collection
# Usage: ./scripts/truncate-agents.sh [path-to-.env]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default to UC-INFRA-001 config if no argument provided
ENV_FILE="${1:-usecases/UC-INFRA-001-water-distribution-network/.env}"

# Load environment variables from specified file
if [ -f "$ENV_FILE" ]; then
    echo -e "${GREEN}Loading configuration from: ${ENV_FILE}${NC}"
    export $(grep -v '^#' "$ENV_FILE" | grep -v '^$' | xargs)
else
    echo -e "${RED}Error: Environment file not found: $ENV_FILE${NC}"
    exit 1
fi

# Use CVXC_ prefixed variables from .env file
ARANGO_HOST="${CVXC_DATABASE_HOST:-localhost}"
ARANGO_PORT="${CVXC_DATABASE_PORT:-8529}"
ARANGO_USER="${CVXC_DATABASE_USERNAME:-root}"
ARANGO_PASSWORD="${CVXC_DATABASE_PASSWORD:-rootpassword}"
ARANGO_DATABASE="${CVXC_DATABASE_DATABASE:-water_distribution_network}"

echo -e "${YELLOW}╔════════════════════════════════════════╗${NC}"
echo -e "${YELLOW}║  ArangoDB Agent Truncate Utility      ║${NC}"
echo -e "${YELLOW}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "Database: ${GREEN}${ARANGO_DATABASE}${NC}"
echo -e "Host:     ${GREEN}${ARANGO_HOST}:${ARANGO_PORT}${NC}"
echo ""

# Confirmation prompt
read -p "⚠️  This will DELETE ALL agents from the database. Are you sure? (yes/no): " confirmation

if [ "$confirmation" != "yes" ]; then
    echo -e "${YELLOW}Operation cancelled.${NC}"
    exit 0
fi

echo ""
echo -e "${YELLOW}🗑️  Truncating agents collection...${NC}"

# Execute AQL query to truncate agents collection
QUERY='FOR doc IN agents REMOVE doc IN agents'

curl -s -X POST \
  "http://${ARANGO_HOST}:${ARANGO_PORT}/_db/${ARANGO_DATABASE}/_api/cursor" \
  -u "${ARANGO_USER}:${ARANGO_PASSWORD}" \
  -H "Content-Type: application/json" \
  -d "{\"query\": \"${QUERY}\"}" | jq .

echo ""
echo -e "${GREEN}✅ All agents have been removed from the database.${NC}"
echo ""
echo -e "${YELLOW}To reload agents, run one of these commands:${NC}"
echo -e "  • ${GREEN}make run-water${NC}  - Load water distribution agents"
echo -e "  • ${GREEN}go run ./cmd --config config.yaml${NC}  - Load from default config"
echo ""
