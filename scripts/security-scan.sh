#!/bin/bash
# Comprehensive security scanning script for Catalogizer
# Runs all available security tools and generates unified report

set -e

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPORT_DIR="$ROOT_DIR/docs/security"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="$REPORT_DIR/security-scan-$TIMESTAMP.md"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_section() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

# Create report directory
mkdir -p "$REPORT_DIR"

# Initialize report
cat > "$REPORT_FILE" <<EOF
# Catalogizer Security Scan Report
**Generated:** $(date -u +"%Y-%m-%d %H:%M:%S UTC")
**Scan ID:** $TIMESTAMP

## Executive Summary

This report contains the results of automated security scanning across all Catalogizer components.

---

EOF

log_section "Catalogizer Security Scan"
log_info "Report will be saved to: $REPORT_FILE"

# Check for available tools
AVAILABLE_TOOLS=()
MISSING_TOOLS=()

check_tool() {
    if command -v "$1" &> /dev/null; then
        AVAILABLE_TOOLS+=("$1")
        log_info "✓ $1 is available"
        return 0
    else
        MISSING_TOOLS+=("$1")
        log_warn "✗ $1 is not installed"
        return 1
    fi
}

log_section "Checking Security Tools"

check_tool "snyk" || true
check_tool "trivy" || true
check_tool "gosec" || true
check_tool "nancy" || true
check_tool "npm" || true
check_tool "go" || true

if [ ${#AVAILABLE_TOOLS[@]} -eq 0 ]; then
    log_error "No security scanning tools available!"
    log_info "Install at least one tool:"
    log_info "  - Snyk: npm install -g snyk"
    log_info "  - Trivy: https://aquasecurity.github.io/trivy/latest/getting-started/installation/"
    log_info "  - Gosec: go install github.com/securego/gosec/v2/cmd/gosec@latest"
    exit 1
fi

# Add tool availability to report
cat >> "$REPORT_FILE" <<EOF
## Tools Used

EOF

for tool in "${AVAILABLE_TOOLS[@]}"; do
    version=$($tool --version 2>&1 | head -1 || echo "unknown")
    echo "- **$tool**: $version" >> "$REPORT_FILE"
done

if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
    echo -e "\n### Tools Not Available\n" >> "$REPORT_FILE"
    for tool in "${MISSING_TOOLS[@]}"; do
        echo "- $tool" >> "$REPORT_FILE"
    done
fi

echo -e "\n---\n" >> "$REPORT_FILE"

# Function to run Snyk scan
run_snyk() {
    log_section "Running Snyk Scan"

    cat >> "$REPORT_FILE" <<EOF
## Snyk Vulnerability Scan

### Go Dependencies (catalog-api)

EOF

    cd "$ROOT_DIR/catalog-api"

    if snyk test --severity-threshold=medium --json > "$REPORT_DIR/snyk-go-$TIMESTAMP.json" 2>&1; then
        log_info "Snyk Go scan: No vulnerabilities found"
        echo "✅ **No medium or higher vulnerabilities found**" >> "$REPORT_FILE"
    else
        log_warn "Snyk Go scan: Vulnerabilities detected"
        echo "⚠️ **Vulnerabilities detected** (see detailed JSON report)" >> "$REPORT_FILE"
    fi

    echo "" >> "$REPORT_FILE"

    # Scan npm projects
    for project in catalog-web catalogizer-desktop installer-wizard catalogizer-api-client; do
        if [ -d "$ROOT_DIR/$project" ] && [ -f "$ROOT_DIR/$project/package.json" ]; then
            echo "### npm Dependencies ($project)" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"

            cd "$ROOT_DIR/$project"
            if snyk test --severity-threshold=medium --json > "$REPORT_DIR/snyk-$project-$TIMESTAMP.json" 2>&1; then
                log_info "Snyk $project scan: No vulnerabilities found"
                echo "✅ **No medium or higher vulnerabilities found**" >> "$REPORT_FILE"
            else
                log_warn "Snyk $project scan: Vulnerabilities detected"
                echo "⚠️ **Vulnerabilities detected** (see detailed JSON report)" >> "$REPORT_FILE"
            fi
            echo "" >> "$REPORT_FILE"
        fi
    done

    echo "---" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
}

# Function to run Trivy scan
run_trivy() {
    log_section "Running Trivy Scan"

    cat >> "$REPORT_FILE" <<EOF
## Trivy Filesystem Scan

### Scanning Project Directory

EOF

    cd "$ROOT_DIR"

    if trivy fs --severity HIGH,CRITICAL --format json --output "$REPORT_DIR/trivy-fs-$TIMESTAMP.json" . 2>&1; then
        # Extract summary from JSON
        critical=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "CRITICAL")] | length' "$REPORT_DIR/trivy-fs-$TIMESTAMP.json" 2>/dev/null || echo "0")
        high=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "HIGH")] | length' "$REPORT_DIR/trivy-fs-$TIMESTAMP.json" 2>/dev/null || echo "0")

        echo "- **Critical vulnerabilities:** $critical" >> "$REPORT_FILE"
        echo "- **High vulnerabilities:** $high" >> "$REPORT_FILE"

        if [ "$critical" -eq 0 ] && [ "$high" -eq 0 ]; then
            log_info "Trivy scan: No critical or high vulnerabilities"
            echo "" >> "$REPORT_FILE"
            echo "✅ **No critical or high vulnerabilities found**" >> "$REPORT_FILE"
        else
            log_warn "Trivy scan: Found $critical critical and $high high vulnerabilities"
            echo "" >> "$REPORT_FILE"
            echo "⚠️ **Vulnerabilities detected** (see detailed JSON report)" >> "$REPORT_FILE"
        fi
    else
        log_error "Trivy scan failed"
        echo "❌ **Scan failed**" >> "$REPORT_FILE"
    fi

    echo "" >> "$REPORT_FILE"
    echo "---" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
}

# Function to run Gosec scan
run_gosec() {
    log_section "Running Gosec Scan"

    cat >> "$REPORT_FILE" <<EOF
## Gosec Security Audit (Go Code)

### Scanning catalog-api

EOF

    cd "$ROOT_DIR/catalog-api"

    if gosec -fmt json -out "$REPORT_DIR/gosec-$TIMESTAMP.json" -exclude-generated ./... 2>&1 | tee /dev/null; then
        # Extract summary
        issues=$(jq '.Stats.found // 0' "$REPORT_DIR/gosec-$TIMESTAMP.json" 2>/dev/null || echo "unknown")

        echo "- **Issues found:** $issues" >> "$REPORT_FILE"

        if [ "$issues" -eq 0 ] 2>/dev/null; then
            log_info "Gosec scan: No security issues found"
            echo "" >> "$REPORT_FILE"
            echo "✅ **No security issues found**" >> "$REPORT_FILE"
        else
            log_warn "Gosec scan: Found $issues security issues"
            echo "" >> "$REPORT_FILE"
            echo "⚠️ **Security issues detected** (see detailed JSON report)" >> "$REPORT_FILE"
        fi
    else
        log_warn "Gosec scan completed with findings"
        echo "⚠️ **See detailed JSON report for findings**" >> "$REPORT_FILE"
    fi

    echo "" >> "$REPORT_FILE"
    echo "---" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
}

# Function to run Nancy (Go dependency check)
run_nancy() {
    log_section "Running Nancy Go Dependency Check"

    cat >> "$REPORT_FILE" <<EOF
## Nancy Go Dependency Vulnerability Check

### Scanning Go modules

EOF

    cd "$ROOT_DIR/catalog-api"

    if go list -json -m all | nancy sleuth --output json > "$REPORT_DIR/nancy-$TIMESTAMP.json" 2>&1; then
        log_info "Nancy scan: No vulnerable dependencies"
        echo "✅ **No vulnerable Go dependencies found**" >> "$REPORT_FILE"
    else
        log_warn "Nancy scan: Vulnerable dependencies detected"
        echo "⚠️ **Vulnerable dependencies detected** (see detailed JSON report)" >> "$REPORT_FILE"
    fi

    echo "" >> "$REPORT_FILE"
    echo "---" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
}

# Function to run npm audit
run_npm_audit() {
    log_section "Running npm audit"

    cat >> "$REPORT_FILE" <<EOF
## npm Audit Results

EOF

    for project in catalog-web catalogizer-desktop installer-wizard catalogizer-api-client; do
        if [ -d "$ROOT_DIR/$project" ] && [ -f "$ROOT_DIR/$project/package.json" ]; then
            echo "### $project" >> "$REPORT_FILE"
            echo "" >> "$REPORT_FILE"

            cd "$ROOT_DIR/$project"

            if [ ! -d "node_modules" ]; then
                log_warn "Skipping npm audit for $project (node_modules not found)"
                echo "⚠️ **Skipped** (dependencies not installed)" >> "$REPORT_FILE"
                echo "" >> "$REPORT_FILE"
                continue
            fi

            if npm audit --json > "$REPORT_DIR/npm-audit-$project-$TIMESTAMP.json" 2>&1; then
                log_info "npm audit $project: No vulnerabilities"
                echo "✅ **No vulnerabilities found**" >> "$REPORT_FILE"
            else
                # Extract summary
                critical=$(jq '.metadata.vulnerabilities.critical // 0' "$REPORT_DIR/npm-audit-$project-$TIMESTAMP.json" 2>/dev/null || echo "0")
                high=$(jq '.metadata.vulnerabilities.high // 0' "$REPORT_DIR/npm-audit-$project-$TIMESTAMP.json" 2>/dev/null || echo "0")
                moderate=$(jq '.metadata.vulnerabilities.moderate // 0' "$REPORT_DIR/npm-audit-$project-$TIMESTAMP.json" 2>/dev/null || echo "0")

                log_warn "npm audit $project: Found vulnerabilities (Critical: $critical, High: $high, Moderate: $moderate)"

                echo "- **Critical:** $critical" >> "$REPORT_FILE"
                echo "- **High:** $high" >> "$REPORT_FILE"
                echo "- **Moderate:** $moderate" >> "$REPORT_FILE"
                echo "" >> "$REPORT_FILE"
                echo "⚠️ **Run \`npm audit fix\` to resolve**" >> "$REPORT_FILE"
            fi
            echo "" >> "$REPORT_FILE"
        fi
    done

    echo "---" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
}

# Run available scans
for tool in "${AVAILABLE_TOOLS[@]}"; do
    case $tool in
        snyk)
            run_snyk
            ;;
        trivy)
            run_trivy
            ;;
        gosec)
            run_gosec
            ;;
        nancy)
            run_nancy
            ;;
        npm)
            run_npm_audit
            ;;
    esac
done

# Add recommendations section
cat >> "$REPORT_FILE" <<EOF
## Recommendations

### Immediate Actions
1. Review all CRITICAL and HIGH severity vulnerabilities
2. Update vulnerable dependencies where patches are available
3. Evaluate workarounds for vulnerabilities without patches
4. Re-run security scans after remediation

### Continuous Security
1. Integrate security scanning into CI/CD pipeline (run locally)
2. Schedule weekly dependency updates
3. Enable automated security alerts
4. Conduct periodic manual security reviews

### Missing Tools
EOF

if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
    echo "" >> "$REPORT_FILE"
    echo "Consider installing these additional security tools:" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    for tool in "${MISSING_TOOLS[@]}"; do
        case $tool in
            snyk)
                echo "- **Snyk**: \`npm install -g snyk\` - Comprehensive vulnerability scanner" >> "$REPORT_FILE"
                ;;
            trivy)
                echo "- **Trivy**: See https://aquasecurity.github.io/trivy/latest/getting-started/installation/ - Filesystem and container scanner" >> "$REPORT_FILE"
                ;;
            gosec)
                echo "- **Gosec**: \`go install github.com/securego/gosec/v2/cmd/gosec@latest\` - Go security checker" >> "$REPORT_FILE"
                ;;
            nancy)
                echo "- **Nancy**: \`go install github.com/sonatype-nexus-community/nancy@latest\` - Go dependency vulnerability scanner" >> "$REPORT_FILE"
                ;;
        esac
    done
else
    echo "" >> "$REPORT_FILE"
    echo "✅ All recommended security tools are installed." >> "$REPORT_FILE"
fi

cat >> "$REPORT_FILE" <<EOF

---

## Detailed Reports

All detailed JSON reports are saved in: \`docs/security/\`

- Scan ID: $TIMESTAMP
- View JSON reports for comprehensive vulnerability details
- Use tool-specific commands for interactive remediation

---

**End of Report**
EOF

# Print summary
log_section "Security Scan Complete"
log_info "Report saved to: $REPORT_FILE"
log_info "Detailed JSON reports in: $REPORT_DIR/"

echo ""
log_info "Tools used: ${AVAILABLE_TOOLS[*]}"
if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
    log_warn "Missing tools: ${MISSING_TOOLS[*]}"
fi

echo ""
log_info "Next steps:"
log_info "1. Review the report: cat $REPORT_FILE"
log_info "2. Address critical and high severity findings"
log_info "3. Re-run scan after remediation: $0"

exit 0
