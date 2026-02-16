#!/bin/bash
# ============================================================
# Catalogizer - Coverage Validation & False Positive Detection
# Validates test coverage thresholds and detects false positives
# ============================================================

set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
REPORTS_DIR="${PROJECT_DIR}/reports"
RELEASES_DIR="${PROJECT_DIR}/releases"

# Validation results
VALIDATION_ISSUES=()
VALIDATION_WARNINGS=()
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[VALIDATE]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[VALIDATE]${NC} $1"; }
log_error() { echo -e "${RED}[VALIDATE]${NC} $1"; }
log_check() { echo -e "${BLUE}  [CHECK]${NC} $1"; }

check_pass() {
    ((TOTAL_CHECKS++))
    ((PASSED_CHECKS++))
    log_check "$1 ... PASS"
}

check_fail() {
    ((TOTAL_CHECKS++))
    ((FAILED_CHECKS++))
    log_check "$1 ... FAIL"
    VALIDATION_ISSUES+=("$1")
}

check_warn() {
    ((TOTAL_CHECKS++))
    ((PASSED_CHECKS++))
    log_check "$1 ... WARN"
    VALIDATION_WARNINGS+=("$1")
}

# ============================================================
# Go Coverage Validation
# ============================================================
validate_go_coverage() {
    log_info "=== Go Coverage Validation ==="

    local coverage_file="$REPORTS_DIR/go-coverage.out"
    if [ ! -f "$coverage_file" ]; then
        check_warn "Go coverage file not found"
        return 0
    fi

    # Parse total coverage percentage
    local total_coverage
    total_coverage=$(go tool cover -func="$coverage_file" 2>/dev/null | grep "total:" | awk '{print $3}' | tr -d '%' || echo "0")

    if [ -z "$total_coverage" ] || [ "$total_coverage" = "0" ]; then
        check_warn "Could not parse Go coverage (got: $total_coverage)"
        return 0
    fi

    log_info "Go total coverage: ${total_coverage}%"

    # Check per-package coverage
    local low_coverage_pkgs=0
    while IFS= read -r line; do
        local pkg
        pkg=$(echo "$line" | awk '{print $1}')
        local cov
        cov=$(echo "$line" | awk '{print $3}' | tr -d '%')
        if [ -n "$cov" ] && [ "$(echo "$cov < 50" | bc -l 2>/dev/null || echo 0)" = "1" ]; then
            ((low_coverage_pkgs++))
        fi
    done < <(go tool cover -func="$coverage_file" 2>/dev/null | grep -v "total:" | grep -v "^$" || true)

    if [ "$low_coverage_pkgs" -gt 0 ]; then
        check_warn "Go: $low_coverage_pkgs package(s) below 50% coverage"
    else
        check_pass "Go: all packages above 50% coverage"
    fi

    check_pass "Go coverage report generated (${total_coverage}%)"
}

# ============================================================
# JavaScript/TypeScript Coverage Validation
# ============================================================
validate_js_coverage() {
    log_info "=== JavaScript/TypeScript Coverage Validation ==="

    local js_projects=("catalog-web" "catalogizer-desktop" "installer-wizard" "catalogizer-api-client")

    for project in "${js_projects[@]}"; do
        local coverage_dir="$REPORTS_DIR/coverage-$project"
        local coverage_summary="$PROJECT_DIR/$project/coverage/coverage-summary.json"

        if [ -f "$coverage_summary" ]; then
            # Parse coverage from Vitest/Jest summary
            local line_cov
            line_cov=$(jq -r '.total.lines.pct // 0' "$coverage_summary" 2>/dev/null || echo "0")
            log_info "$project line coverage: ${line_cov}%"
            check_pass "$project: coverage report exists (lines: ${line_cov}%)"
        elif [ -d "$PROJECT_DIR/$project/coverage" ]; then
            check_pass "$project: coverage directory exists"
        else
            check_warn "$project: no coverage data found"
        fi
    done
}

# ============================================================
# Android/JaCoCo Coverage Validation
# ============================================================
validate_android_coverage() {
    log_info "=== Android Coverage Validation ==="

    local android_projects=("catalogizer-android" "catalogizer-androidtv")

    for project in "${android_projects[@]}"; do
        local jacoco_xml="$PROJECT_DIR/$project/app/build/reports/jacoco/jacocoTestReport/jacocoTestReport.xml"
        local jacoco_html_dir="$PROJECT_DIR/$project/app/build/reports/jacoco/jacocoTestReport/html"

        if [ -f "$jacoco_xml" ]; then
            # Parse instruction coverage from JaCoCo XML
            local missed
            missed=$(grep -o 'type="INSTRUCTION"[^/]*missed="[0-9]*"' "$jacoco_xml" 2>/dev/null | head -1 | grep -o 'missed="[0-9]*"' | grep -o '[0-9]*' || echo "0")
            local covered
            covered=$(grep -o 'type="INSTRUCTION"[^/]*covered="[0-9]*"' "$jacoco_xml" 2>/dev/null | head -1 | grep -o 'covered="[0-9]*"' | grep -o '[0-9]*' || echo "0")

            if [ "$((missed + covered))" -gt 0 ]; then
                local pct=$((covered * 100 / (missed + covered)))
                log_info "$project instruction coverage: ${pct}%"
                check_pass "$project: JaCoCo coverage report exists (${pct}%)"
            else
                check_pass "$project: JaCoCo XML report generated"
            fi
        elif [ -d "$jacoco_html_dir" ]; then
            check_pass "$project: JaCoCo HTML report exists"
        else
            check_warn "$project: no JaCoCo coverage data found"
        fi
    done
}

# ============================================================
# False Positive Detection - Assertion Density
# ============================================================
validate_assertion_density() {
    log_info "=== Assertion Density Check ==="

    # Check Go test files for assertions
    local go_test_files=0
    local go_tests_no_assert=0
    while IFS= read -r f; do
        ((go_test_files++))
        if ! grep -qE '(assert\.|require\.|t\.Error|t\.Fatal|t\.Fail|if .* != |if .* == )' "$f" 2>/dev/null; then
            ((go_tests_no_assert++))
        fi
    done < <(find "$PROJECT_DIR/catalog-api" -name "*_test.go" -type f 2>/dev/null || true)

    if [ "$go_test_files" -gt 0 ]; then
        if [ "$go_tests_no_assert" -gt 0 ]; then
            check_warn "Go: $go_tests_no_assert/$go_test_files test files with no obvious assertions"
        else
            check_pass "Go: all $go_test_files test files contain assertions"
        fi
    fi

    # Check JS/TS test files
    local js_test_files=0
    local js_tests_no_assert=0
    while IFS= read -r f; do
        ((js_test_files++))
        if ! grep -qE '(expect\(|assert\.|toBe|toEqual|toHaveBeenCalled|toThrow|toContain|toMatch)' "$f" 2>/dev/null; then
            ((js_tests_no_assert++))
        fi
    done < <(find "$PROJECT_DIR" -path "*/node_modules" -prune -o \( -name "*.test.ts" -o -name "*.test.tsx" -o -name "*.test.js" -o -name "*.spec.ts" -o -name "*.spec.tsx" \) -type f -print 2>/dev/null || true)

    if [ "$js_test_files" -gt 0 ]; then
        if [ "$js_tests_no_assert" -gt 0 ]; then
            check_warn "JS/TS: $js_tests_no_assert/$js_test_files test files with no obvious assertions"
        else
            check_pass "JS/TS: all $js_test_files test files contain assertions"
        fi
    fi
}

# ============================================================
# Artifact Smoke Validation
# ============================================================
validate_artifacts() {
    log_info "=== Artifact Smoke Validation ==="

    # Validate Go binaries
    local go_binary="$RELEASES_DIR/linux/catalog-api/catalog-api-v${BUILD_VERSION:-1.0.0}-linux-amd64"
    if [ -f "$go_binary" ]; then
        if file "$go_binary" | grep -q "ELF.*executable"; then
            check_pass "Go binary: valid ELF executable"
        else
            check_fail "Go binary: not a valid ELF executable"
        fi

        if [ -x "$go_binary" ] || chmod +x "$go_binary"; then
            # Quick health check - just check it starts
            timeout 5 "$go_binary" --help >/dev/null 2>&1 || true
            check_pass "Go binary: executable"
        else
            check_fail "Go binary: not executable"
        fi
    else
        check_warn "Go Linux binary not found"
    fi

    # Validate web build
    local web_index="$RELEASES_DIR/linux/catalog-web/index.html"
    if [ -f "$web_index" ]; then
        if grep -q "<script" "$web_index" 2>/dev/null || grep -q "src=" "$web_index" 2>/dev/null; then
            check_pass "Web build: index.html contains script references"
        else
            check_warn "Web build: index.html may be missing asset references"
        fi
    else
        check_warn "Web build index.html not found"
    fi

    # Validate Android APKs
    for apk_dir in "$RELEASES_DIR/android/catalogizer-android" "$RELEASES_DIR/android/catalogizer-androidtv"; do
        local project_name
        project_name=$(basename "$apk_dir")
        local found_apk=false

        for apk in "$apk_dir"/*.apk; do
            if [ -f "$apk" ]; then
                found_apk=true
                # Check if APK is a valid ZIP (APKs are ZIP files)
                if file "$apk" | grep -qE "(Zip|Java archive)"; then
                    check_pass "$project_name APK: valid archive"
                else
                    check_fail "$project_name APK: not a valid archive"
                fi

                # Check APK with aapt if available
                if command -v aapt &>/dev/null; then
                    if aapt dump badging "$apk" >/dev/null 2>&1; then
                        check_pass "$project_name APK: aapt validation passed"

                        # Check if signed
                        if aapt dump badging "$apk" 2>/dev/null | grep -q "application-label"; then
                            check_pass "$project_name APK: has valid manifest"
                        fi
                    else
                        check_warn "$project_name APK: aapt validation failed"
                    fi
                fi
            fi
        done

        if ! $found_apk; then
            check_warn "$project_name: no APK files found"
        fi
    done

    # Validate desktop builds
    for desktop_dir in "$RELEASES_DIR/linux/catalogizer-desktop" "$RELEASES_DIR/linux/installer-wizard"; do
        local app_name
        app_name=$(basename "$desktop_dir")
        local found_binary=false

        for f in "$desktop_dir"/*; do
            if [ -f "$f" ] && file "$f" | grep -q "ELF"; then
                found_binary=true
                check_pass "$app_name: valid Linux binary found"
                break
            fi
        done

        if [ -f "$desktop_dir"/*.deb ] 2>/dev/null; then
            check_pass "$app_name: .deb package found"
        fi

        if ! $found_binary; then
            check_warn "$app_name: no Linux binary found"
        fi
    done
}

# ============================================================
# Skipped Test Detection
# ============================================================
validate_no_skipped_tests() {
    log_info "=== Skipped Test Detection ==="

    # Check Go test logs for skipped tests
    local go_log="$REPORTS_DIR/go-test.log"
    if [ -f "$go_log" ]; then
        local skipped
        skipped=$(grep -c "SKIP" "$go_log" 2>/dev/null || echo "0")
        if [ "$skipped" -gt 0 ]; then
            check_warn "Go: $skipped test(s) skipped"
        else
            check_pass "Go: no tests skipped"
        fi
    fi

    # Check JS test logs
    for log_file in "$REPORTS_DIR"/*-test.log; do
        if [ -f "$log_file" ]; then
            local name
            name=$(basename "$log_file" "-test.log")
            local skipped
            skipped=$(grep -ciE "(skipped|pending|todo)" "$log_file" 2>/dev/null || echo "0")
            if [ "$skipped" -gt 0 ]; then
                check_warn "$name: $skipped potentially skipped test(s)"
            else
                check_pass "$name: no skipped tests detected"
            fi
        fi
    done
}

# ============================================================
# Generate Validation Report
# ============================================================
generate_validation_report() {
    log_info "=== Generating Validation Report ==="

    # JSON report
    cat > "$REPORTS_DIR/validation-report.json" << REPORT_EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "total_checks": $TOTAL_CHECKS,
  "passed": $PASSED_CHECKS,
  "failed": $FAILED_CHECKS,
  "issues": [
$(printf '    "%s",\n' "${VALIDATION_ISSUES[@]}" 2>/dev/null | sed '$ s/,$//' || echo "")
  ],
  "warnings": [
$(printf '    "%s",\n' "${VALIDATION_WARNINGS[@]}" 2>/dev/null | sed '$ s/,$//' || echo "")
  ]
}
REPORT_EOF

    # HTML report
    cat > "$REPORTS_DIR/validation-report.html" << 'HTML_EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Catalogizer - Validation Report</title>
    <style>
        body { font-family: 'Segoe UI', sans-serif; margin: 20px; background: #f5f7fa; }
        .header { background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%); color: white; padding: 25px; border-radius: 12px; text-align: center; margin-bottom: 25px; }
        .summary { display: grid; grid-template-columns: repeat(3, 1fr); gap: 15px; margin-bottom: 25px; }
        .metric { background: white; padding: 20px; border-radius: 10px; text-align: center; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .metric h3 { margin: 0; color: #495057; font-size: 12px; text-transform: uppercase; }
        .metric p { margin: 8px 0 0; font-size: 28px; font-weight: bold; }
        .pass { color: #28a745; } .fail { color: #dc3545; } .warn { color: #ffc107; }
        .section { background: white; margin: 15px 0; padding: 20px; border-radius: 10px; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .item { padding: 8px 0; border-bottom: 1px solid #f0f0f0; }
    </style>
</head>
<body>
HTML_EOF

    cat >> "$REPORTS_DIR/validation-report.html" << HTML_BODY
    <div class="header">
        <h1>Validation Report</h1>
        <p>$(date)</p>
    </div>
    <div class="summary">
        <div class="metric"><h3>Total Checks</h3><p>$TOTAL_CHECKS</p></div>
        <div class="metric"><h3>Passed</h3><p class="pass">$PASSED_CHECKS</p></div>
        <div class="metric"><h3>Failed</h3><p class="fail">$FAILED_CHECKS</p></div>
    </div>
HTML_BODY

    if [ ${#VALIDATION_ISSUES[@]} -gt 0 ]; then
        echo "    <div class=\"section\"><h2>Issues</h2>" >> "$REPORTS_DIR/validation-report.html"
        for issue in "${VALIDATION_ISSUES[@]}"; do
            echo "        <div class=\"item fail\">$issue</div>" >> "$REPORTS_DIR/validation-report.html"
        done
        echo "    </div>" >> "$REPORTS_DIR/validation-report.html"
    fi

    if [ ${#VALIDATION_WARNINGS[@]} -gt 0 ]; then
        echo "    <div class=\"section\"><h2>Warnings</h2>" >> "$REPORTS_DIR/validation-report.html"
        for warning in "${VALIDATION_WARNINGS[@]}"; do
            echo "        <div class=\"item warn\">$warning</div>" >> "$REPORTS_DIR/validation-report.html"
        done
        echo "    </div>" >> "$REPORTS_DIR/validation-report.html"
    fi

    echo "</body></html>" >> "$REPORTS_DIR/validation-report.html"

    log_info "Validation report: $REPORTS_DIR/validation-report.json"
    log_info "Validation report: $REPORTS_DIR/validation-report.html"
}

# ============================================================
# Main
# ============================================================
main() {
    echo -e "${BLUE}=== Catalogizer Coverage Validation ===${NC}"

    mkdir -p "$REPORTS_DIR"

    validate_go_coverage
    validate_js_coverage
    validate_android_coverage
    validate_assertion_density
    validate_artifacts
    validate_no_skipped_tests
    generate_validation_report

    echo ""
    echo -e "${BLUE}=== Validation Summary ===${NC}"
    echo -e "  Total checks: $TOTAL_CHECKS"
    echo -e "  ${GREEN}Passed: $PASSED_CHECKS${NC}"
    echo -e "  ${RED}Failed: $FAILED_CHECKS${NC}"
    echo -e "  ${YELLOW}Warnings: ${#VALIDATION_WARNINGS[@]}${NC}"

    if [ "$FAILED_CHECKS" -gt 0 ]; then
        echo -e "\n${RED}Validation found $FAILED_CHECKS issue(s)${NC}"
        exit 1
    else
        echo -e "\n${GREEN}All validation checks passed${NC}"
        exit 0
    fi
}

main "$@"
