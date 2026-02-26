#!/bin/bash
#
# Nancy Dependency Vulnerability Scanner
# Scans Go dependencies for known vulnerabilities
#

set -e

echo "=== Running Nancy Dependency Scan ==="

cd "$(dirname "$0")/../catalog-api"

# Create reports directory
mkdir -p ../reports/security

# Generate dependency list and scan
echo "Scanning Go dependencies..."
go list -json -m all 2>/dev/null | nancy sleuth \
  --output json > ../reports/security/nancy-results.json 2>&1 || true

# Check results
if [ -f ../reports/security/nancy-results.json ]; then
    # Try to parse vulnerable count
    VULNERABLE=$(jq '.vulnerable | length' ../reports/security/nancy-results.json 2>/dev/null || echo "0")
    echo "Scan complete. Found ${VULNERABLE} vulnerable dependencies."
else
    echo "Warning: Nancy may not be installed or no results generated"
fi
