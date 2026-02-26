#!/bin/bash
#
# Gosec Security Scan Script
# Scans Go code for security vulnerabilities
#

set -e

echo "=== Running Gosec Security Scan ==="

cd "$(dirname "$0")/../catalog-api"

# Create reports directory
mkdir -p ../reports/security

# Run Gosec with SARIF output for GitHub integration
echo "[1/2] Running Gosec with SARIF output..."
gosec -fmt sarif -out ../reports/security/gosec-results.sarif \
  -stdout -verbose \
  -exclude-dir=vendor \
  -exclude-dir=internal/tests \
  -exclude-dir=mocks \
  ./... 2>/dev/null || true

# Run Gosec with JSON output for processing
echo "[2/2] Running Gosec with JSON output..."
gosec -fmt json -out ../reports/security/gosec-results.json \
  -exclude-dir=vendor \
  -exclude-dir=internal/tests \
  -exclude-dir=mocks \
  ./... 2>/dev/null || true

# Count issues
if [ -f ../reports/security/gosec-results.json ]; then
    ISSUES=$(jq '.Issues | length' ../reports/security/gosec-results.json 2>/dev/null || echo "0")
    echo ""
    echo "Scan complete. Found ${ISSUES} security issues."
    echo "Reports saved to: reports/security/"
else
    echo "Warning: Could not generate Gosec report"
fi
