#!/bin/bash

# UC-COMM-001: CodeValdDiraMoja - Community Chatter Management Startup Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=========================================="
echo "UC-COMM-001: CodeValdDiraMoja"
echo "Community Engagement Platform"
echo "=========================================="
echo ""

# Load environment variables
if [ -f "$SCRIPT_DIR/.env" ]; then
    echo "Loading environment configuration..."
    export $(grep -v '^#' "$SCRIPT_DIR/.env" | xargs)
    echo "✓ Environment variables loaded"
else
    echo "ERROR: .env file not found at $SCRIPT_DIR/.env"
    exit 1
fi

echo ""
echo "Configuration:"
echo "  Use Case ID:    $CVXC_USE_CASE_ID"
echo "  Use Case Name:  $CVXC_USE_CASE_NAME"
echo "  API Port:       $CVXC_API_PORT"
echo "  Database:       $CVXC_DB_NAME"
echo "  Environment:    $CVXC_ENVIRONMENT"
echo ""

# Check if database is accessible
echo "Checking database connectivity..."
if command -v curl &> /dev/null; then
    DB_CHECK=$(curl -s -o /dev/null -w "%{http_code}" http://$CVXC_DB_HOST:$CVXC_DB_PORT/_api/version 2>/dev/null || echo "000")
    if [ "$DB_CHECK" = "200" ]; then
        echo "✓ ArangoDB is accessible at $CVXC_DB_HOST:$CVXC_DB_PORT"
    else
        echo "⚠ WARNING: Cannot connect to ArangoDB at $CVXC_DB_HOST:$CVXC_DB_PORT"
        echo "  Please ensure ArangoDB is running"
    fi
else
    echo "⚠ curl not found, skipping database connectivity check"
fi

echo ""
echo "Agent Configuration:"
echo "  Member Agent:             $CVXC_AGENT_MEMBER_ENABLED"
echo "  Topic Agent:              $CVXC_AGENT_TOPIC_ENABLED"
echo "  Event Agent:              $CVXC_AGENT_EVENT_ENABLED"
echo "  Voting Agent:             $CVXC_AGENT_VOTING_ENABLED"
echo "  Moderator Agent:          $CVXC_AGENT_MODERATOR_ENABLED"
echo "  Sentiment Analyzer:       $CVXC_AGENT_SENTIMENT_ANALYZER_ENABLED"
echo ""

# Build the application
echo "Building DiraMoja system..."
cd "$PROJECT_ROOT"

if [ ! -f "$PROJECT_ROOT/bin/codevaldcortex" ]; then
    echo "Building CodeValdCortex binary..."
    make build
    echo "✓ Build complete"
else
    echo "✓ Using existing binary"
fi

echo ""
echo "Starting CodeValdDiraMoja Community Engagement Platform..."
echo "Access the system at: http://localhost:$CVXC_API_PORT"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Start the application
cd "$SCRIPT_DIR"
exec "$PROJECT_ROOT/bin/codevaldcortex"
