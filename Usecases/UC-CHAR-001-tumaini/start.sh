#!/bin/bash

# UC-CHAR-001: CodeValdTumaini - Charity Distribution Network Startup Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=========================================="
echo "UC-CHAR-001: CodeValdTumaini"
echo "Charity Distribution Network"
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
echo "  Donor Agent:              $CVXC_AGENT_DONOR_ENABLED"
echo "  Recipient Agent:          $CVXC_AGENT_RECIPIENT_ENABLED"
echo "  Item Agent:               $CVXC_AGENT_ITEM_ENABLED"
echo "  Volunteer Agent:          $CVXC_AGENT_VOLUNTEER_ENABLED"
echo "  Logistics Coordinator:    $CVXC_AGENT_LOGISTICS_ENABLED"
echo "  Storage Facility Agent:   $CVXC_AGENT_STORAGE_ENABLED"
echo "  Need Matcher Agent:       $CVXC_AGENT_MATCHER_ENABLED"
echo "  Impact Tracker Agent:     $CVXC_AGENT_IMPACT_TRACKER_ENABLED"
echo ""

echo "Charity Configuration:"
echo "  Match Algorithm:          $CVXC_CHARITY_MATCH_ALGORITHM"
echo "  Delivery SLA:             $CVXC_CHARITY_DELIVERY_SLA_HOURS hours"
echo "  Waste Reduction Target:   $CVXC_CHARITY_WASTE_REDUCTION_TARGET%"
echo "  Donor Retention Target:   $CVXC_CHARITY_DONOR_RETENTION_TARGET%"
echo ""

# Build the application
echo "Building Tumaini system..."
cd "$PROJECT_ROOT"

if [ ! -f "$PROJECT_ROOT/bin/codevaldcortex" ]; then
    echo "Building CodeValdCortex binary..."
    make build
    echo "✓ Build complete"
else
    echo "✓ Using existing binary"
fi

echo ""
echo "Starting CodeValdTumaini Charity Distribution Network..."
echo "Access the system at: http://localhost:$CVXC_API_PORT"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Start the application
cd "$SCRIPT_DIR"
exec "$PROJECT_ROOT/bin/codevaldcortex"
