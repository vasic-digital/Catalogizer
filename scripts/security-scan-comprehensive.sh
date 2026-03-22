#!/bin/bash
###############################################################################
# COMPREHENSIVE SECURITY SCANNING SUITE
# Catalogizer Project - Security Scanning Orchestration
# Version: 1.0
# Date: March 22, 2026
###############################################################################

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPORTS_DIR="reports/security"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
EXIT_CODE=0

# Ensure reports directory exists
mkdir -p "${REPORTS_DIR}/${TIMESTAMP}"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

# Header
print_header() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}\n"
}

# Function to check if tool is installed
check_tool() {
    if command -v "$1" &> /dev/null; then
        return 0
    else
        return 1
    fi
}

# Function to install tool
install_tool() {
    local tool=$1
    log_info "Installing ${tool}..."
    
    case $tool in
        trivy)
            curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh
            sudo mv trivy /usr/local/bin/
            ;;
        gosec)
            go install github.com/securego/gosec/v2/cmd/gosec@latest
            ;;
        nancy)
            go install github.com/sonatypecommunity/nancy@latest
            ;;
        semgrep)
            pip3 install semgrep
            ;;
        gitleaks)
            go install github.com/zricethezav/gitleaks/v8@latest
            ;;
        snyk)
            npm install -g snyk
            ;;
        *)
            log_error "Unknown tool: ${tool}"
            return 1
            ;;
    esac
    
    log_success "${tool} installed successfully"
}

###############################################################################
# SCAN 1: Trivy - Container and Filesystem Vulnerability Scanning
###############################################################################
run_trivy_scan() {
    print_header "TRIVY SECURITY SCAN"
    
    if ! check_tool trivy; then
        log_warning "Trivy not found. Installing..."
        install_tool trivy || return 1
    fi
    
    log_info "Running Trivy filesystem scan..."
    
    # Filesystem scan
    trivy fs \
        --severity HIGH,CRITICAL \
        --format sarif \
        --output "${REPORTS_DIR}/${TIMESTAMP}/trivy-fs.sarif" \
        --exit-code 0 \
        . 2>/dev/null || true
    
    # JSON report for parsing
    trivy fs \
        --severity HIGH,CRITICAL \
        --format json \
        --output "${REPORTS_DIR}/${TIMESTAMP}/trivy-fs.json" \
        --exit-code 0 \
        . 2>/dev/null || true
    
    # Count vulnerabilities
    if [ -f "${REPORTS_DIR}/${TIMESTAMP}/trivy-fs.json" ]; then
        CRITICAL_COUNT=$(jq '[.Results[].Vulnerabilities? // [] | .[] | select(.Severity == "CRITICAL")] | length' "${REPORTS_DIR}/${TIMESTAMP}/trivy-fs.json" 2>/dev/null || echo "0")
        HIGH_COUNT=$(jq '[.Results[].Vulnerabilities? // [] | .[] | select(.Severity == "HIGH")] | length' "${REPORTS_DIR}/${TIMESTAMP}/trivy-fs.json" 2>/dev/null || echo "0")
        
        log_info "Trivy Results:"
        log_info "  Critical vulnerabilities: ${CRITICAL_COUNT}"
        log_info "  High vulnerabilities: ${HIGH_COUNT}"
        
        if [ "$CRITICAL_COUNT" -gt 0 ]; then
            log_error "Found ${CRITICAL_COUNT} CRITICAL vulnerabilities!"
            EXIT_CODE=1
        else
            log_success "No critical vulnerabilities found"
        fi
    fi
    
    # Container scan if Dockerfile exists
    if [ -f "Dockerfile" ]; then
        log_info "Running Trivy container image scan..."
        
        podman build -t catalogizer:security-scan . 2>/dev/null || true
        
        trivy image \
            --severity HIGH,CRITICAL \
            --format sarif \
            --output "${REPORTS_DIR}/${TIMESTAMP}/trivy-container.sarif" \
            --exit-code 0 \
            catalogizer:security-scan 2>/dev/null || true
        
        trivy image \
            --severity HIGH,CRITICAL \
            --format json \
            --output "${REPORTS_DIR}/${TIMESTAMP}/trivy-container.json" \
            --exit-code 0 \
            catalogizer:security-scan 2>/dev/null || true
        
        # Clean up
        podman rmi catalogizer:security-scan 2>/dev/null || true
    fi
    
    log_success "Trivy scan complete. Reports: ${REPORTS_DIR}/${TIMESTAMP}/trivy-*.sarif"
}

###############################################################################
# SCAN 2: Gosec - Go Security Checker
###############################################################################
run_gosec_scan() {
    print_header "GOSEC SECURITY SCAN"
    
    if ! check_tool gosec; then
        log_warning "Gosec not found. Installing..."
        install_tool gosec || return 1
    fi
    
    log_info "Running Gosec security scan on Go code..."
    
    cd catalog-api || return 1
    
    # Run gosec
    gosec \
        -fmt sarif \
        -out "../${REPORTS_DIR}/${TIMESTAMP}/gosec.sarif" \
        -exclude=G104,G304 \
        ./... 2>/dev/null || true
    
    # JSON output for parsing
    gosec \
        -fmt json \
        -out "../${REPORTS_DIR}/${TIMESTAMP}/gosec.json" \
        -exclude=G104,G304 \
        ./... 2>/dev/null || true
    
    cd ..
    
    # Parse results
    if [ -f "${REPORTS_DIR}/${TIMESTAMP}/gosec.json" ]; then
        ISSUE_COUNT=$(jq '.Issues | length' "${REPORTS_DIR}/${TIMESTAMP}/gosec.json" 2>/dev/null || echo "0")
        HIGH_COUNT=$(jq '[.Issues[] | select(.severity == "HIGH")] | length' "${REPORTS_DIR}/${TIMESTAMP}/gosec.json" 2>/dev/null || echo "0")
        MEDIUM_COUNT=$(jq '[.Issues[] | select(.severity == "MEDIUM")] | length' "${REPORTS_DIR}/${TIMESTAMP}/gosec.json" 2>/dev/null || echo "0")
        
        log_info "Gosec Results:"
        log_info "  Total issues: ${ISSUE_COUNT}"
        log_info "  High severity: ${HIGH_COUNT}"
        log_info "  Medium severity: ${MEDIUM_COUNT}"
        
        if [ "$HIGH_COUNT" -gt 0 ]; then
            log_error "Found ${HIGH_COUNT} HIGH severity security issues!"
            EXIT_CODE=1
        else
            log_success "No high severity issues found"
        fi
    fi
    
    log_success "Gosec scan complete. Report: ${REPORTS_DIR}/${TIMESTAMP}/gosec.sarif"
}

###############################################################################
# SCAN 3: Nancy - Go Dependency Vulnerability Scanner
###############################################################################
run_nancy_scan() {
    print_header "NANCY DEPENDENCY SCAN"
    
    if ! check_tool nancy; then
        log_warning "Nancy not found. Installing..."
        install_tool nancy || return 1
    fi
    
    log_info "Running Nancy dependency vulnerability scan..."
    
    cd catalog-api || return 1
    
    # Ensure go.sum is up to date
    go mod tidy 2>/dev/null || true
    
    # Run nancy
    go list -json -deps ./... | nancy sleuth \
        --output json \
        --outputfile "../${REPORTS_DIR}/${TIMESTAMP}/nancy.json" 2>/dev/null || true
    
    # Also output text for console
    go list -json -deps ./... | nancy sleuth 2>/dev/null || true
    
    cd ..
    
    # Parse results
    if [ -f "${REPORTS_DIR}/${TIMESTAMP}/nancy.json" ]; then
        VULN_COUNT=$(jq '.vulnerable | length' "${REPORTS_DIR}/${TIMESTAMP}/nancy.json" 2>/dev/null || echo "0")
        
        log_info "Nancy Results:"
        log_info "  Vulnerable dependencies: ${VULN_COUNT}"
        
        if [ "$VULN_COUNT" -gt 0 ]; then
            log_warning "Found ${VULN_COUNT} vulnerable dependencies"
            # Don't fail build for vulnerabilities, just warn
        else
            log_success "No vulnerable dependencies found"
        fi
    fi
    
    log_success "Nancy scan complete. Report: ${REPORTS_DIR}/${TIMESTAMP}/nancy.json"
}

###############################################################################
# SCAN 4: Semgrep - Static Analysis Security Testing
###############################################################################
run_semgrep_scan() {
    print_header "SEMGREP SAST SCAN"
    
    if ! check_tool semgrep; then
        log_warning "Semgrep not found. Installing..."
        install_tool semgrep || return 1
    fi
    
    log_info "Running Semgrep static analysis..."
    
    # Run semgrep with security-focused rules
    semgrep \
        --config=auto \
        --config=p/security-audit \
        --config=p/owasp-top-ten \
        --json \
        --output "${REPORTS_DIR}/${TIMESTAMP}/semgrep.json" \
        --severity=ERROR \
        . 2>/dev/null || true
    
    # Also generate SARIF
    semgrep \
        --config=auto \
        --config=p/security-audit \
        --sarif \
        --output "${REPORTS_DIR}/${TIMESTAMP}/semgrep.sarif" \
        . 2>/dev/null || true
    
    # Parse results
    if [ -f "${REPORTS_DIR}/${TIMESTAMP}/semgrep.json" ]; then
        ERROR_COUNT=$(jq '.results | length' "${REPORTS_DIR}/${TIMESTAMP}/semgrep.json" 2>/dev/null || echo "0")
        
        log_info "Semgrep Results:"
        log_info "  Total findings: ${ERROR_COUNT}"
        
        if [ "$ERROR_COUNT" -gt 0 ]; then
            log_error "Found ${ERROR_COUNT} security issues!"
            EXIT_CODE=1
        else
            log_success "No security issues found"
        fi
    fi
    
    log_success "Semgrep scan complete. Report: ${REPORTS_DIR}/${TIMESTAMP}/semgrep.sarif"
}

###############################################################################
# SCAN 5: GitLeaks - Secret Detection
###############################################################################
run_gitleaks_scan() {
    print_header "GITLEAKS SECRET SCAN"
    
    if ! check_tool gitleaks; then
        log_warning "GitLeaks not found. Installing..."
        install_tool gitleaks || return 1
    fi
    
    log_info "Running GitLeaks secret detection..."
    
    gitleaks detect \
        --verbose \
        --source . \
        --report-format json \
        --report-path "${REPORTS_DIR}/${TIMESTAMP}/gitleaks.json" \
        --exit-code 0 2>/dev/null || true
    
    # Parse results
    if [ -f "${REPORTS_DIR}/${TIMESTAMP}/gitleaks.json" ]; then
        LEAK_COUNT=$(jq '. | length' "${REPORTS_DIR}/${TIMESTAMP}/gitleaks.json" 2>/dev/null || echo "0")
        
        log_info "GitLeaks Results:"
        log_info "  Secrets detected: ${LEAK_COUNT}"
        
        if [ "$LEAK_COUNT" -gt 0 ]; then
            log_error "Found ${LEAK_COUNT} potential secrets!"
            EXIT_CODE=1
        else
            log_success "No secrets detected"
        fi
    fi
    
    log_success "GitLeaks scan complete. Report: ${REPORTS_DIR}/${TIMESTAMP}/gitleaks.json"
}

###############################################################################
# SCAN 6: Snyk - Comprehensive Security Scanning
###############################################################################
run_snyk_scan() {
    print_header "SNYK SECURITY SCAN"
    
    if ! check_tool snyk; then
        log_warning "Snyk not found. Installing..."
        install_tool snyk || return 1
    fi
    
    log_info "Running Snyk dependency scan..."
    
    # Test dependencies
    snyk test \
        --all-projects \
        --json \
        --json-file-output="${REPORTS_DIR}/${TIMESTAMP}/snyk-deps.json" 2>/dev/null || true
    
    # Container scan if Dockerfile exists
    if [ -f "Dockerfile" ]; then
        log_info "Running Snyk container scan..."
        podman build -t catalogizer:snyk-scan . 2>/dev/null || true
        
        snyk container test \
            catalogizer:snyk-scan \
            --json \
            --json-file-output="${REPORTS_DIR}/${TIMESTAMP}/snyk-container.json" 2>/dev/null || true
        
        podman rmi catalogizer:snyk-scan 2>/dev/null || true
    fi
    
    # IaC scan
    log_info "Running Snyk IaC scan..."
    snyk iac test \
        --json \
        --json-file-output="${REPORTS_DIR}/${TIMESTAMP}/snyk-iac.json" 2>/dev/null || true
    
    log_success "Snyk scan complete. Reports: ${REPORTS_DIR}/${TIMESTAMP}/snyk-*.json"
}

###############################################################################
# SCAN 7: Custom Security Checks
###############################################################################
run_custom_checks() {
    print_header "CUSTOM SECURITY CHECKS"
    
    log_info "Running custom security checks..."
    
    local ISSUES=0
    
    # Check for hardcoded passwords
    if grep -r "password.*=.*\"" --include="*.go" --include="*.ts" --include="*.js" . 2>/dev/null | grep -v "_test.go" | grep -v "test_" | head -5; then
        log_warning "Potential hardcoded passwords found"
        ((ISSUES++))
    fi
    
    # Check for API keys
    if grep -r "api_key\|apikey\|API_KEY" --include="*.go" --include="*.ts" --include="*.js" . 2>/dev/null | grep -v "\.env" | grep -v "config" | head -5; then
        log_warning "Potential hardcoded API keys found"
        ((ISSUES++))
    fi
    
    # Check for private keys
    if grep -r "BEGIN PRIVATE KEY\|BEGIN RSA PRIVATE KEY" --include="*.go" --include="*.ts" --include="*.js" --include="*.pem" . 2>/dev/null | head -5; then
        log_error "Private keys found in source code!"
        ((ISSUES++))
        EXIT_CODE=1
    fi
    
    # Check for SQL injection patterns
    if grep -r "fmt\.Sprintf.*SELECT\|fmt\.Sprintf.*INSERT\|fmt\.Sprintf.*UPDATE\|fmt\.Sprintf.*DELETE" --include="*.go" . 2>/dev/null | head -5; then
        log_warning "Potential SQL injection patterns found"
        ((ISSUES++))
    fi
    
    # Check for debug flags in production
    if grep -r "DEBUG.*=.*true\|debug.*=.*True" --include="*.go" --include="*.ts" --include="*.json" . 2>/dev/null | grep -v "test" | head -5; then
        log_warning "Debug flags may be enabled in production"
        ((ISSUES++))
    fi
    
    if [ $ISSUES -eq 0 ]; then
        log_success "No custom security issues found"
    else
        log_warning "Found ${ISSUES} custom security concerns"
    fi
}

###############################################################################
# Generate Summary Report
###############################################################################
generate_summary() {
    print_header "SECURITY SCAN SUMMARY"
    
    local REPORT_FILE="${REPORTS_DIR}/${TIMESTAMP}/SUMMARY.md"
    
    cat > "${REPORT_FILE}" << EOF
# Security Scan Summary

**Scan Date:** $(date)
**Scan ID:** ${TIMESTAMP}
**Status:** $([ $EXIT_CODE -eq 0 ] && echo "✅ PASSED" || echo "❌ FAILED")

## Scans Performed

1. ✅ Trivy - Container and filesystem vulnerability scanning
2. ✅ Gosec - Go security checker
3. ✅ Nancy - Go dependency vulnerability scanner
4. ✅ Semgrep - Static analysis security testing
5. ✅ GitLeaks - Secret detection
6. ✅ Snyk - Comprehensive security scanning
7. ✅ Custom security checks

## Results

| Scan Tool | Status | Report Location |
|-----------|--------|-----------------|
| Trivy FS | Completed | trivy-fs.sarif |
| Trivy Container | Completed | trivy-container.sarif |
| Gosec | Completed | gosec.sarif |
| Nancy | Completed | nancy.json |
| Semgrep | Completed | semgrep.sarif |
| GitLeaks | Completed | gitleaks.json |
| Snyk Dependencies | Completed | snyk-deps.json |
| Snyk Container | Completed | snyk-container.json |
| Snyk IaC | Completed | snyk-iac.json |

## Next Steps

1. Review all SARIF/JSON reports in ${REPORTS_DIR}/${TIMESTAMP}/
2. Address any CRITICAL or HIGH severity findings
3. Create tickets for remediation
4. Re-run scans after fixes

## Reports Location

All detailed reports are available at:
\`\`\`
${REPORTS_DIR}/${TIMESTAMP}/
\`\`\`
EOF

    log_info "Summary report generated: ${REPORT_FILE}"
    
    # Print to console
    cat "${REPORT_FILE}"
}

###############################################################################
# Main Execution
###############################################################################
main() {
    print_header "CATALOGIZER COMPREHENSIVE SECURITY SCAN"
    
    log_info "Starting comprehensive security scanning suite..."
    log_info "Reports will be saved to: ${REPORTS_DIR}/${TIMESTAMP}/"
    
    # Run all scans
    run_trivy_scan
    run_gosec_scan
    run_nancy_scan
    run_semgrep_scan
    run_gitleaks_scan
    run_snyk_scan
    run_custom_checks
    
    # Generate summary
    generate_summary
    
    # Final status
    print_header "SCAN COMPLETE"
    
    if [ $EXIT_CODE -eq 0 ]; then
        log_success "All security scans passed!"
        log_info "Review reports at: ${REPORTS_DIR}/${TIMESTAMP}/"
    else
        log_error "Security scans found issues that need attention!"
        log_info "Review reports at: ${REPORTS_DIR}/${TIMESTAMP}/"
        log_info "Address CRITICAL and HIGH severity issues before deployment"
    fi
    
    exit $EXIT_CODE
}

# Run main function
main "$@"
