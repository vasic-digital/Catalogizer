#!/bin/bash
# Stress Test Runner - runs all stress tests without -short flag
set -e
echo "Running stress tests (this may take several minutes)..."
cd catalog-api
GOMAXPROCS=3 go test ./tests/stress/... -v -count=1 -timeout 30m -p 1 -parallel 2
echo "Stress tests complete."
