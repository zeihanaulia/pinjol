#!/bin/bash

# Pinjol Smoke Test Simulation Script
# This script runs the smoke test simulation with various configurations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
DEFAULT_DURATION="30m"
DEFAULT_USERS=5
DEFAULT_MAX_REQUESTS=50
DEFAULT_URL="http://localhost:8081"

# Function to print usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Run smoke test simulation for Pinjol application"
    echo ""
    echo "OPTIONS:"
    echo "  -d, --duration DURATION    Simulation duration (default: ${DEFAULT_DURATION})"
    echo "  -u, --users NUM           Number of simulated users (default: ${DEFAULT_USERS})"
    echo "  -r, --requests NUM        Max requests per user (default: ${DEFAULT_MAX_REQUESTS})"
    echo "  -U, --url URL             Pinjol application URL (default: ${DEFAULT_URL})"
    echo "  -h, --help               Show this help message"
    echo "  -v, --verbose            Enable verbose output"
    echo ""
    echo "EXAMPLES:"
    echo "  $0                                    # Run with default settings"
    echo "  $0 -d 10m -u 3 -r 20               # 10 minutes, 3 users, 20 requests each"
    echo "  $0 --duration 1h --users 10        # 1 hour, 10 users"
    echo ""
    echo "ENVIRONMENT VARIABLES:"
    echo "  SIMULATION_DURATION     Same as --duration"
    echo "  SIMULATION_USERS        Same as --users"
    echo "  SIMULATION_MAX_REQUESTS Same as --requests"
    echo "  PINJOL_URL             Same as --url"
    echo ""
    echo "PRESETS (use with make):"
    echo "  make simulation-5m     # 5 minutes, 3 users"
    echo "  make simulation-30m    # 30 minutes, 5 users (default)"
    echo "  make simulation-1h     # 1 hour, 10 users"
}

# Function to check if application is running
check_app_health() {
    local url=$1
    echo -e "${BLUE}🔍 Checking application health at ${url}...${NC}"

    if curl -s "${url}/healthz" | grep -q "ok"; then
        echo -e "${GREEN}✅ Application is healthy${NC}"
        return 0
    else
        echo -e "${RED}❌ Application is not responding${NC}"
        echo -e "${YELLOW}💡 Make sure the application is running:${NC}"
        echo "   make run"
        echo "   # or"
        echo "   make compose"
        return 1
    fi
}

# Function to validate duration format
validate_duration() {
    local duration=$1
    if ! [[ $duration =~ ^[0-9]+[smhd]$ ]]; then
        echo -e "${RED}❌ Invalid duration format: ${duration}${NC}"
        echo -e "${YELLOW}💡 Valid formats: 30s, 5m, 1h, 2d${NC}"
        exit 1
    fi
}

# Parse command line arguments
DURATION=${DEFAULT_DURATION}
USERS=${DEFAULT_USERS}
MAX_REQUESTS=${DEFAULT_MAX_REQUESTS}
URL=${DEFAULT_URL}
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--duration)
            DURATION="$2"
            validate_duration "$DURATION"
            shift 2
            ;;
        -u|--users)
            USERS="$2"
            shift 2
            ;;
        -r|--requests)
            MAX_REQUESTS="$2"
            shift 2
            ;;
        -U|--url)
            URL="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Unknown option: $1${NC}"
            usage
            exit 1
            ;;
    esac
done

# Override with environment variables if set
DURATION=${SIMULATION_DURATION:-$DURATION}
USERS=${SIMULATION_USERS:-$USERS}
MAX_REQUESTS=${SIMULATION_MAX_REQUESTS:-$MAX_REQUESTS}
URL=${PINJOL_URL:-$URL}

# Validate inputs
validate_duration "$DURATION"

if ! [[ "$USERS" =~ ^[0-9]+$ ]] || [ "$USERS" -lt 1 ]; then
    echo -e "${RED}❌ Invalid number of users: ${USERS}${NC}"
    exit 1
fi

if ! [[ "$MAX_REQUESTS" =~ ^[0-9]+$ ]] || [ "$MAX_REQUESTS" -lt 1 ]; then
    echo -e "${RED}❌ Invalid number of max requests: ${MAX_REQUESTS}${NC}"
    exit 1
fi

# Print configuration
echo -e "${BLUE}🚀 Pinjol Smoke Test Simulation${NC}"
echo -e "${BLUE}================================${NC}"
echo -e "${YELLOW}Configuration:${NC}"
echo "  Duration: ${DURATION}"
echo "  Users: ${USERS}"
echo "  Max requests per user: ${MAX_REQUESTS}"
echo "  Application URL: ${URL}"
echo ""

# Check application health
if ! check_app_health "$URL"; then
    exit 1
fi

echo ""

# Set environment variables for the Go program
export SIMULATION_DURATION="$DURATION"
export SIMULATION_USERS="$USERS"
export SIMULATION_MAX_REQUESTS="$MAX_REQUESTS"
export PINJOL_URL="$URL"

# Run the simulation
echo -e "${GREEN}🎯 Starting simulation...${NC}"
echo ""

if [ "$VERBOSE" = true ]; then
    CGO_ENABLED=1 go run smoke_simulation.go
else
    CGO_ENABLED=1 go run smoke_simulation.go 2>&1
fi

echo ""
echo -e "${GREEN}✅ Simulation completed!${NC}"
echo ""
echo -e "${BLUE}📊 Check your dashboards:${NC}"
echo "  Business Dashboard: http://localhost:3000/d/pinjol-business-dashboard"
echo "  Logs Dashboard: http://localhost:3000/d/pinjol-logs-dashboard"
echo "  Application Logs: http://localhost:3000/d/pinjol-main-dashboard"
