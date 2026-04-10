#!/bin/bash
# Unified Security Scanning Script for Catalogizer
# Runs all security scanners and generates a unified report

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORTS_DIR="$PROJECT_ROOT/reports"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
UNIFIED_REPORT="$REPORTS_DIR/security-scan-unified-$TIMESTAMP.md"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Scan options
FAIL_ON_CRITICAL=true
SKIP_TRIVY=false
SKIP_GOSEC=false
SKIP_NANCY=false
SKIP_SONAR=false
SKIP_SNYK=false
VERBOSE=false

# Function to print usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --skip-trivy          Skip Trivy scan"
    echo "  --skip-gosec          Skip Gosec scan"
    echo "  --skip-nancy          Skip Nancy scan"
    echo "  --skip-sonar          Skip SonarQube scan"
    echo "  --skip-snyk           Skip Snyk scan"
    echo "  --no-fail             Don't fail on critical findings"
    echo "  -v, --verbose         Enable verbose output"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                    # Run all scans"
    echo "  $0 --skip-snyk        # Skip Snyk (if no token)"
    echo "  $0 --skip-sonar       # Skip SonarQube (if not running)"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-trivy)
            SKIP_TRIVY=true
            shift
            ;;
        --skip-gosec)
            SKIP_GOSEC=true
            shift
            ;;
        --skip-nancy)
            SKIP_NANCY=true
            shift
            ;;
        --skip-sonar)
            SKIP_SONAR=true
            shift
            ;;
        --skip-snyk)
            SKIP_SNYK=true
            shift
            ;;
        --no-fail)
            FAIL_ON_CRITICAL=false
            shift
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
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Create reports directory
mkdir -p "$REPORTS_DIR"

# Print banner
echo -e "${CYAN}"
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║        CATALOGIZER UNIFIED SECURITY SCANNER                  ║"
echo "║        $(date '+%Y-%m-%d %H:%M:%S')                              ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Initialize counters
TOTAL_SCANS=0
SUCCESSFUL_SCANS=0
FAILED_SCANS=0
CRITICAL_FINDINGS=0
HIGH_FINDINGS=0
MEDIUM_FINDINGS=0
LOW_FINDINGS=0

# Start unified report
cat > "$UNIFIED_REPORT" << EOF
# Catalogizer Security Scan Report

**Date:** $(date '+%Y-%m-%d %H:%M:%S')  
**Scan ID:** $TIMESTAMP  
**Project:** Catalogizer  
**Version:** $(cat "$PROJECT_ROOT/versions.json" 2>/dev/null | grep '"version"' | head -1 | cut -d'"' -f4 || echo "Unknown")

---

## Executive Summary

| Scanner | Status | Findings |
|---------|--------|----------|
EOF

# Function to run a scan
run_scan() {
    local name="$1"
    local script="$2"
    local output_file="$3"
    local skip_flag="$4"
    
    if [ "$skip_flag" = "true" ]; then
        echo -e "${YELLOW}⏭ Skipping $name scan${NC}"
        echo "| $name | ⏭ Skipped | N/A |" >> "$UNIFIED_REPORT"
        return 0
    fi
    
    TOTAL_SCANS=$((TOTAL_SCANS + 1))
    echo ""
    echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  Running $name scan${NC}"
    echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
    
    local start_time=$(date +%s)
    
    if [ "$VERBOSE" = true ]; then
        "$script" -v
    else
        "$script" 2>&1 | tee "$REPORTS_DIR/${name,,}-scan-$TIMESTAMP.log"
    fi
    
    local exit_code=${PIPESTATUS[0]}
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✓ $name scan completed successfully (${duration}s)${NC}"
        SUCCESSFUL_SCANS=$((SUCCESSFUL_SCANS + 1))
        echo "| $name | ✅ Passed | 0 |" >> "$UNIFIED_REPORT"
    else
        echo -e "${YELLOW}⚠ $name scan completed with findings (${duration}s)${NC}"
        SUCCESSFUL_SCANS=$((SUCCESSFUL_SCANS + 1))
        # Count findings from output file
        local findings=0
        if [ -f "$output_file" ]; then
            findings=$(jq 'length' "$output_file" 2>/dev/null || echo "0")
        fi
        echo "| $name | ⚠ Findings | $findings |" >> "$UNIFIED_REPORT"
    fi
    
    return $exit_code
}

# Run Trivy
if [ "$SKIP_TRIVY" = false ]; then
    run_scan "Trivy" "$SCRIPT_DIR/scan-trivy.sh" "$REPORTS_DIR/trivy-results.json" "$SKIP_TRIVY"
    # Count findings
    if [ -f "$REPORTS_DIR/trivy-results.json" ]; then
        VULNS=$(jq '[.Results[]?.Vulnerabilities? // [] | length] | add // 0' "$REPORTS_DIR/trivy-results.json" 2>/dev/null || echo "0")
        CRITICAL_FINDINGS=$((CRITICAL_FINDINGS + $(jq '[.Results[]?.Vulnerabilities? // [] | .[] | select(.Severity == "CRITICAL")] | length' "$REPORTS_DIR/trivy-results.json" 2>/dev/null || echo "0")))
        HIGH_FINDINGS=$((HIGH_FINDINGS + $(jq '[.Results[]?.Vulnerabilities? // [] | .[] | select(.Severity == "HIGH")] | length' "$REPORTS_DIR/trivy-results.json" 2>/dev/null || echo "0")))
    fi
fi

# Run Gosec
if [ "$SKIP_GOSEC" = false ]; then
    run_scan "Gosec" "$SCRIPT_DIR/scan-gosec.sh" "$REPORTS_DIR/gosec-results.json" "$SKIP_GOSEC"
    if [ -f "$REPORTS_DIR/gosec-results.json" ]; then
        ISSUES=$(jq '.Stats.found // 0' "$REPORTS_DIR/gosec-results.json" 2>/dev/null || echo "0")
        HIGH_FINDINGS=$((HIGH_FINDINGS + ISSUES))
    fi
fi

# Run Nancy
if [ "$SKIP_NANCY" = false ]; then
    run_scan "Nancy" "$SCRIPT_DIR/scan-nancy.sh" "$REPORTS_DIR/nancy-results.json" "$SKIP_NANCY"
    if [ -f "$REPORTS_DIR/nancy-results.json" ]; then
        VULNS=$(jq 'length' "$REPORTS_DIR/nancy-results.json" 2>/dev/null || echo "0")
        CRITICAL_FINDINGS=$((CRITICAL_FINDINGS + $(jq '[.[] | select(.Vulnerability.Severity == "Critical")] | length' "$REPORTS_DIR/nancy-results.json" 2>/dev/null || echo "0")))
        HIGH_FINDINGS=$((HIGH_FINDINGS + $(jq '[.[] | select(.Vulnerability.Severity == "High")] | length' "$REPORTS_DIR/nancy-results.json" 2>/dev/null || echo "0")))
    fi
fi

# Run SonarQube (if configured)
if [ "$SKIP_SONAR" = false ] && [ -f "$SCRIPT_DIR/run-sonarqube-scan.sh" ]; then
    echo ""
    echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  Running SonarQube scan${NC}"
    echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
    
    TOTAL_SCANS=$((TOTAL_SCANS + 1))
    
    if "$SCRIPT_DIR/run-sonarqube-scan.sh" 2>&1 | tee "$REPORTS_DIR/sonarqube-scan-$TIMESTAMP.log"; then
        echo -e "${GREEN}✓ SonarQube scan completed${NC}"
        SUCCESSFUL_SCANS=$((SUCCESSFUL_SCANS + 1))
        echo "| SonarQube | ✅ Passed | - |" >> "$UNIFIED_REPORT"
    else
        echo -e "${YELLOW}⚠ SonarQube scan completed with findings${NC}"
        SUCCESSFUL_SCANS=$((SUCCESSFUL_SCANS + 1))
        echo "| SonarQube | ⚠ Findings | - |" >> "$UNIFIED_REPORT"
    fi
fi

# Run Snyk (if configured)
if [ "$SKIP_SNYK" = false ]; then
    if [ -n "$SNYK_TOKEN" ]; then
        echo ""
        echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
        echo -e "${BLUE}  Running Snyk scan${NC}"
        echo -e "${BLUE}══════════════════════════════════════════════════════════════${NC}"
        
        TOTAL_SCANS=$((TOTAL_SCANS + 1))
        
        if command -v snyk &> /dev/null; then
            if snyk test --json > "$REPORTS_DIR/snyk-results.json" 2>&1; then
                echo -e "${GREEN}✓ Snyk scan completed${NC}"
                SUCCESSFUL_SCANS=$((SUCCESSFUL_SCANS + 1))
                echo "| Snyk | ✅ Passed | 0 |" >> "$UNIFIED_REPORT"
            else
                echo -e "${YELLOW}⚠ Snyk scan completed with findings${NC}"
                SUCCESSFUL_SCANS=$((SUCCESSFUL_SCANS + 1))
                VULNS=$(jq '.vulnerabilities | length' "$REPORTS_DIR/snyk-results.json" 2>/dev/null || echo "0")
                echo "| Snyk | ⚠ Findings | $VULNS |" >> "$UNIFIED_REPORT"
            fi
        else
            echo -e "${YELLOW}⚠ Snyk CLI not installed, skipping${NC}"
            echo "| Snyk | ⏭ Skipped | N/A |" >> "$UNIFIED_REPORT"
        fi
    else
        echo -e "${YELLOW}⏭ Skipping Snyk (SNYK_TOKEN not set)${NC}"
        echo "| Snyk | ⏭ Skipped | N/A |" >> "$UNIFIED_REPORT"
    fi
fi

# Complete unified report
cat >> "$UNIFIED_REPORT" << EOF

---

## Findings Summary

| Severity | Count |
|----------|-------|
| 🔴 Critical | $CRITICAL_FINDINGS |
| 🟠 High | $HIGH_FINDINGS |
| 🟡 Medium | $MEDIUM_FINDINGS |
| 🟢 Low | $LOW_FINDINGS |

---

## Detailed Reports

Individual scan reports are available in the \`reports/\` directory:

EOF

# Add links to individual reports
for report in "$REPORTS_DIR"/*-results*.json; do
    if [ -f "$report" ]; then
        basename "$report" >> "$UNIFIED_REPORT"
    fi
done

cat >> "$UNIFIED_REPORT" << EOF

---

## Recommendations

EOF

if [ $CRITICAL_FINDINGS -gt 0 ]; then
    echo "1. 🔴 **Critical vulnerabilities found** - Immediate attention required" >> "$UNIFIED_REPORT"
    echo "   - Review critical findings immediately" >> "$UNIFIED_REPORT"
    echo "   - Apply patches or updates" >> "$UNIFIED_REPORT"
    echo "   - Re-scan after fixes" >> "$UNIFIED_REPORT"
fi

if [ $HIGH_FINDINGS -gt 0 ]; then
    echo "2. 🟠 **High severity issues found** - Address within 24 hours" >> "$UNIFIED_REPORT"
fi

cat >> "$UNIFIED_REPORT" << EOF

3. Review all findings in individual scan reports
4. Update \`.trivyignore\`, \`.gosec.json\`, \`.nancy-ignore\` for accepted risks
5. Schedule regular security scans (recommend: weekly)

---

*Report generated by Catalogizer Security Scanner*  
*Timestamp: $TIMESTAMP*
EOF

# Print summary
echo ""
echo -e "${CYAN}══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}                    SCAN SUMMARY                              ${NC}"
echo -e "${CYAN}══════════════════════════════════════════════════════════════${NC}"
echo ""
echo "Scanners run:     $TOTAL_SCANS"
echo -e "Successful:       ${GREEN}$SUCCESSFUL_SCANS${NC}"
echo -e "Failed:           ${RED}$FAILED_SCANS${NC}"
echo ""
echo "Findings:"
if [ $CRITICAL_FINDINGS -gt 0 ]; then
    echo -e "  🔴 Critical:    ${RED}$CRITICAL_FINDINGS${NC}"
else
    echo -e "  🔴 Critical:    ${GREEN}$CRITICAL_FINDINGS${NC}"
fi

if [ $HIGH_FINDINGS -gt 0 ]; then
    echo -e "  🟠 High:        ${YELLOW}$HIGH_FINDINGS${NC}"
else
    echo -e "  🟠 High:        ${GREEN}$HIGH_FINDINGS${NC}"
fi

echo -e "  🟡 Medium:      $MEDIUM_FINDINGS"
echo -e "  🟢 Low:         $LOW_FINDINGS"
echo ""
echo -e "Unified report:   ${BLUE}$UNIFIED_REPORT${NC}"
echo ""
echo -e "${CYAN}══════════════════════════════════════════════════════════════${NC}"

# Exit with appropriate code
if [ "$FAIL_ON_CRITICAL" = true ] && [ $CRITICAL_FINDINGS -gt 0 ]; then
    echo -e "${RED}❌ Security scan failed: Critical vulnerabilities found${NC}"
    exit 1
fi

if [ "$FAIL_ON_CRITICAL" = true ] && [ $HIGH_FINDINGS -gt 5 ]; then
    echo -e "${YELLOW}⚠ Security scan warning: High number of high-severity issues${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Security scan completed${NC}"
exit 0
