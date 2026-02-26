#!/bin/bash
#
# Security Gates Script
# Validates security thresholds
#

MAX_CRITICAL=0
MAX_HIGH=10

echo "=== Security Gates ==="
echo "Max Critical Issues: ${MAX_CRITICAL}"
echo "Max High Issues: ${MAX_HIGH}"
echo ""

# Check for critical vulnerabilities
CRITICAL_COUNT=0
HIGH_COUNT=0

# Parse latest scan results
LATEST_SCAN=$(ls -td reports/security/*/ 2>/dev/null | head -1)

if [ -z "${LATEST_SCAN}" ]; then
    echo "Warning: No security scan results found"
    echo "Run: ./scripts/security-scan-full.sh"
    exit 0
fi

echo "Checking scan results in: ${LATEST_SCAN}"

# Try to count issues from various reports
if [ -f "${LATEST_SCAN}/trivy-fs.json" ]; then
    CRITICAL_COUNT=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity=="CRITICAL")] | length' "${LATEST_SCAN}/trivy-fs.json" 2>/dev/null || echo "0")
    HIGH_COUNT=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity=="HIGH")] | length' "${LATEST_SCAN}/trivy-fs.json" 2>/dev/null || echo "0")
fi

echo ""
echo "Results:"
echo "  Critical Issues: ${CRITICAL_COUNT} (max: ${MAX_CRITICAL})"
echo "  High Issues: ${HIGH_COUNT} (max: ${MAX_HIGH})"
echo ""

# Check gates
FAILED=0

if [ "${CRITICAL_COUNT}" -gt "${MAX_CRITICAL}" ]; then
    echo "FAIL: Critical vulnerabilities exceed threshold"
    FAILED=1
fi

if [ "${HIGH_COUNT}" -gt "${MAX_HIGH}" ]; then
    echo "FAIL: High vulnerabilities exceed threshold"
    FAILED=1
fi

if [ ${FAILED} -eq 1 ]; then
    echo ""
    echo "SECURITY GATES FAILED"
    exit 1
else
    echo "SECURITY GATES PASSED"
    exit 0
fi
