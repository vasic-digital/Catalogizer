#!/bin/bash
# Boot script for Catalogizer - starts all services

set -e

export SERVER_PORT=28080
export HTTPS_PORT=28443

cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer/catalog-api

echo "Starting Catalogizer API on ports $SERVER_PORT (HTTP) and $HTTPS_PORT (HTTPS/HTTP3)..."

# Start the API binary in background, redirect output to log file
nohup ./api-server > /tmp/catalogizer-api.log 2>&1 &
API_PID=$!

echo "API started with PID: $API_PID"

# Wait for API to be ready
sleep 3

# Check if running
if ps -p $API_PID > /dev/null 2>&1; then
    echo "API is running (PID: $API_PID)"
else
    echo "API failed to start"
    cat /tmp/catalogizer-api.log
    exit 1
fi

echo "=== API Health Check ==="
curl -s http://localhost:$SERVER_PORT/api/v1/health || echo "Health check pending..."

echo ""
echo "To stop: kill $API_PID"