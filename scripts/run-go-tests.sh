#!/bin/bash
# run-go-tests.sh - Run Go backend tests with coverage analysis

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CATALOG_API_DIR="$PROJECT_ROOT/catalog-api"
REPORTS_DIR="$PROJECT_ROOT/reports"
COVERAGE_DIR="$REPORTS_DIR/coverage"

echo "🧪 Running Go backend tests..."

# Create reports directory
mkdir -p "$COVERAGE_DIR"

# Change to catalog-api directory
cd "$CATALOG_API_DIR"

# Run tests with resource limits (respecting 30-40% host resource limits)
echo "📊 Running tests with resource limits..."
GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -timeout 10m -count=1 -coverprofile="$COVERAGE_DIR/coverage.out"

# Generate coverage report
echo "📈 Generating coverage report..."
go tool cover -func="$COVERAGE_DIR/coverage.out" -o "$COVERAGE_DIR/coverage.txt"

# Show summary
echo "📋 Coverage Summary:"
cat "$COVERAGE_DIR/coverage.txt" | tail -5

# Generate HTML coverage report
go tool cover -html="$COVERAGE_DIR/coverage.out" -o "$COVERAGE_DIR/coverage.html"

echo "✅ Tests completed. Coverage report: $COVERAGE_DIR/coverage.html"
echo "📊 Full coverage: $COVERAGE_DIR/coverage.txt"
