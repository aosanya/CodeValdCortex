#!/bin/bash

# UC-TRACK-001: Safiri Salama - Safe Journey Tracking System Startup Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=========================================="
echo "UC-TRACK-001: Safiri Salama"
echo "Safe Journey Tracking System"
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
        echo "  You can start it with: docker-compose up -d arangodb"
    fi
else
    echo "⚠ curl not found, skipping database connectivity check"
fi

echo ""
echo "Agent Configuration:"
echo "  Vehicle Agent:         $CVXC_AGENT_VEHICLE_ENABLED"
echo "  Parent Agent:          $CVXC_AGENT_PARENT_ENABLED"
echo "  Passenger Agent:       $CVXC_AGENT_PASSENGER_ENABLED"
echo "  Route Manager Agent:   $CVXC_AGENT_ROUTE_MANAGER_ENABLED"
echo "  Fleet Operator Agent:  $CVXC_AGENT_FLEET_OPERATOR_ENABLED"
echo ""

echo "Broadcasting Configuration:"
echo "  Enabled:               $CVXC_BROADCASTING_ENABLED"
echo "  At Stop Interval:      ${CVXC_BROADCAST_AT_STOP_INTERVAL}s"
echo "  Approaching Stop:      ${CVXC_BROADCAST_APPROACHING_STOP_INTERVAL}s"
echo "  En Route:              ${CVXC_BROADCAST_EN_ROUTE_INTERVAL}s"
echo "  Emergency:             ${CVXC_BROADCAST_EMERGENCY_INTERVAL}s"
echo "  Privacy Mode:          $CVXC_BROADCAST_PRIVACY_MODE_ENABLED"
echo ""

echo "Feature Flags:"
echo "  School Mode:           $CVXC_SCHOOL_MODE_ENABLED"
echo "  Matatu Mode:           $CVXC_MATATU_MODE_ENABLED"
echo "  Favorite Matatus:      $CVXC_FEATURE_FAVORITE_MATATUS"
echo "  Trip Ratings:          $CVXC_FEATURE_TRIP_RATINGS"
echo "  Loyalty Points:        $CVXC_FEATURE_LOYALTY_POINTS"
echo "  Smart Alerts:          $CVXC_FEATURE_SMART_ALERTS"
echo ""

# Check if message broker is accessible
echo "Checking message broker connectivity..."
if [ "$CVXC_BROKER_TYPE" = "nats" ]; then
    if command -v nc &> /dev/null; then
        if nc -z $CVXC_BROKER_HOST $CVXC_BROKER_PORT 2>/dev/null; then
            echo "✓ NATS broker is accessible at $CVXC_BROKER_HOST:$CVXC_BROKER_PORT"
        else
            echo "⚠ WARNING: Cannot connect to NATS broker at $CVXC_BROKER_HOST:$CVXC_BROKER_PORT"
            echo "  Please ensure NATS is running"
            echo "  You can start it with: docker-compose up -d nats"
        fi
    else
        echo "⚠ nc (netcat) not found, skipping broker connectivity check"
    fi
fi

echo ""
echo "Agent Type Schemas:"
if [ -d "$SCRIPT_DIR/config/agents" ]; then
    SCHEMA_COUNT=$(find "$SCRIPT_DIR/config/agents" -name "*.json" | wc -l)
    echo "  Found $SCHEMA_COUNT agent type schema(s)"
    find "$SCRIPT_DIR/config/agents" -name "*.json" -exec basename {} \; | sed 's/^/    - /'
else
    echo "  ⚠ WARNING: Agent config directory not found"
fi

echo ""
echo "Starting Safiri Salama Tracking System..."
echo ""

# Check if binary exists
if [ -f "$PROJECT_ROOT/bin/codevaldcortex" ]; then
    echo "Using binary: $PROJECT_ROOT/bin/codevaldcortex"
    cd "$PROJECT_ROOT"
    exec ./bin/codevaldcortex
elif [ -f "$PROJECT_ROOT/main" ]; then
    echo "Using binary: $PROJECT_ROOT/main"
    cd "$PROJECT_ROOT"
    exec ./main
else
    echo "Binary not found. Building from source..."
    cd "$PROJECT_ROOT"
    
    if [ -f "go.mod" ]; then
        echo "Building Go application..."
        go build -o bin/codevaldcortex cmd/main.go
        echo "✓ Build complete"
        echo ""
        echo "Starting application..."
        exec ./bin/codevaldcortex
    else
        echo "ERROR: Cannot find Go module or binary"
        echo "Please build the application first:"
        echo "  cd $PROJECT_ROOT"
        echo "  make build"
        exit 1
    fi
fi
