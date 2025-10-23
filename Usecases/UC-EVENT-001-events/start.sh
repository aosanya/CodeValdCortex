#!/bin/bash

# UC-EVENT-001: CodeValdEvents (Nuruyetu) - AI-Powered Event Info Desk Startup Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=========================================="
echo "UC-EVENT-001: CodeValdEvents (Nuruyetu)"
echo "AI-Powered Event Info Desk"
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
echo "  Info Desk Agent:          $CVXC_AGENT_INFO_DESK_ENABLED"
echo "  Attendee Agent:           $CVXC_AGENT_ATTENDEE_ENABLED"
echo "  Incident Manager:         $CVXC_AGENT_INCIDENT_MANAGER_ENABLED"
echo "  Staff Agent:              $CVXC_AGENT_STAFF_ENABLED"
echo "  Emergency Coordinator:    $CVXC_AGENT_EMERGENCY_COORDINATOR_ENABLED"
echo "  Analytics Agent:          $CVXC_AGENT_ANALYTICS_ENABLED"
echo ""

echo "AI Configuration:"
echo "  RAG Enabled:              $CVXC_AI_RAG_ENABLED"
echo "  Confidence Threshold:     $CVXC_AI_CONFIDENCE_THRESHOLD"
echo "  Citation Mode:            $CVXC_AI_CITATION_MODE"
echo "  Multilingual:             $CVXC_AI_MULTILINGUAL_ENABLED"
echo "  Supported Languages:      $CVXC_AI_SUPPORTED_LANGUAGES"
echo ""

# Build the application
echo "Building Nuruyetu system..."
cd "$PROJECT_ROOT"

if [ ! -f "$PROJECT_ROOT/bin/codevaldcortex" ]; then
    echo "Building CodeValdCortex binary..."
    make build
    echo "✓ Build complete"
else
    echo "✓ Using existing binary"
fi

echo ""
echo "Starting CodeValdEvents (Nuruyetu) AI Info Desk..."
echo "Access the system at: http://localhost:$CVXC_API_PORT"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Start the application
cd "$SCRIPT_DIR"
exec "$PROJECT_ROOT/bin/codevaldcortex"
