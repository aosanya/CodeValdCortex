#!/bin/bash

# UC-FRA-001: Financial Risk Analysis System Startup Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=========================================="
echo "UC-FRA-001: Financial Risk Analysis System"
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
echo "  Entity Agent:             $CVXC_AGENT_ENTITY_ENABLED"
echo "  Ratio Calculator Agent:   $CVXC_AGENT_RATIO_CALCULATOR_ENABLED"
echo "  Risk Scorer Agent:        $CVXC_AGENT_RISK_SCORER_ENABLED"
echo "  Covenant Monitor Agent:   $CVXC_AGENT_COVENANT_MONITOR_ENABLED"
echo "  Portfolio Aggregator:     $CVXC_AGENT_PORTFOLIO_AGGREGATOR_ENABLED"
echo ""

# Check if agent config files exist
echo "Verifying agent configuration files..."
AGENTS_DIR="$SCRIPT_DIR/config/agents"
REQUIRED_AGENTS=("entity.json" "ratio_calculator.json" "risk_scorer.json" "covenant_monitor.json" "portfolio_aggregator.json")

all_agents_found=true
for agent in "${REQUIRED_AGENTS[@]}"; do
    if [ -f "$AGENTS_DIR/$agent" ]; then
        echo "  ✓ $agent"
    else
        echo "  ✗ $agent (NOT FOUND)"
        all_agents_found=false
    fi
done

if [ "$all_agents_found" = false ]; then
    echo ""
    echo "ERROR: Some required agent configuration files are missing"
    exit 1
fi

echo ""
echo "Compliance Configuration:"
echo "  Basel III:     $CVXC_COMPLIANCE_BASEL_III"
echo "  IFRS 9:        $CVXC_COMPLIANCE_IFRS_9"
echo "  CECL:          $CVXC_COMPLIANCE_CECL"
echo "  Dodd-Frank:    $CVXC_COMPLIANCE_DODD_FRANK"
echo ""

# Build the application
echo "Building Financial Risk Analysis application..."
cd "$PROJECT_ROOT"

if [ -f "$PROJECT_ROOT/go.mod" ]; then
    echo "  Building Go application..."
    go build -o "$PROJECT_ROOT/bin/fra_server" "$PROJECT_ROOT/cmd/main.go"
    if [ $? -eq 0 ]; then
        echo "✓ Build successful"
    else
        echo "ERROR: Build failed"
        exit 1
    fi
else
    echo "ERROR: go.mod not found in project root"
    exit 1
fi

echo ""
echo "Starting Financial Risk Analysis System..."
echo "  API will be available at: http://localhost:$CVXC_API_PORT"
echo "  Metrics available at:     http://localhost:$CVXC_METRICS_PORT/metrics"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Start the application
cd "$SCRIPT_DIR"
"$PROJECT_ROOT/bin/fra_server"
