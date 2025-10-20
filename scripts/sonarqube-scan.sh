#!/bin/bash

# SonarQube Scanner Script for Catalogizer
# This script performs comprehensive code quality analysis

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SONAR_HOST_URL="${SONAR_HOST_URL:-http://localhost:9000}"
SONAR_TOKEN="${SONAR_TOKEN:?Sonar token required}"
PROJECT_KEY="${PROJECT_KEY:-catalogizer}"
REPORTS_DIR="$PROJECT_ROOT/reports"

echo "🔍 Starting SonarQube Analysis for Catalogizer"
echo "🌐 SonarQube Server: $SONAR_HOST_URL"
echo "📁 Project Root: $PROJECT_ROOT"

# Create reports directory
mkdir -p "$REPORTS_DIR"

# Function to check SonarQube server availability
check_sonarqube() {
    echo "🔍 Checking SonarQube server availability..."
    for i in {1..30}; do
        if curl -f -s "$SONAR_HOST_URL/api/system/status" > /dev/null 2>&1; then
            echo "✅ SonarQube server is ready"
            return 0
        fi
        echo "⏳ Waiting for SonarQube server... ($i/30)"
        sleep 10
    done
    echo "❌ SonarQube server is not available"
    return 1
}

# Function to install SonarScanner
install_scanner() {
    if ! command -v sonar-scanner &> /dev/null; then
        echo "📦 Installing SonarScanner..."
        
        # Detect OS and architecture
        OS=$(uname -s | tr '[:upper:]' '[:lower:]')
        ARCH=$(uname -m)
        
        case $ARCH in
            x86_64) ARCH="x64" ;;
            aarch64|arm64) ARCH="arm64" ;;
            *) echo "❌ Unsupported architecture: $ARCH"; exit 1 ;;
        esac
        
        SONAR_SCANNER_VERSION="5.0.1.3006"
        SONAR_SCANNER_URL="https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-${SONAR_SCANNER_VERSION}-${OS}-${ARCH}.zip"
        
        cd /tmp
        wget -q "$SONAR_SCANNER_URL" -O sonar-scanner.zip
        unzip -q sonar-scanner.zip
        sudo mv sonar-scanner-${SONAR_SCANNER_VERSION}-${OS}-${ARCH} /opt/sonar-scanner
        sudo ln -sf /opt/sonar-scanner/bin/sonar-scanner /usr/local/bin/sonar-scanner
        rm sonar-scanner.zip
        
        echo "✅ SonarScanner installed successfully"
    else
        echo "✅ SonarScanner is already installed"
    fi
}

# Function to prepare Go coverage
prepare_go_coverage() {
    echo "🐹 Preparing Go coverage report..."
    cd "$PROJECT_ROOT/catalog-api"
    
    if [ -f "go.mod" ]; then
        go mod tidy
        go test -v -race -coverprofile=coverage.out ./... 2>/dev/null || true
        go tool cover -html=coverage.out -o coverage.html 2>/dev/null || true
        
        # Generate test results in JSON format
        go test -json ./... > test-results.json 2>/dev/null || true
        
        echo "✅ Go coverage prepared"
    fi
}

# Function to prepare JavaScript/TypeScript coverage
prepare_js_coverage() {
    echo "🟢 Preparing JavaScript/TypeScript coverage..."
    
    for project_dir in catalog-web catalogizer-desktop catalogizer-api-client installer-wizard; do
        if [ -d "$PROJECT_ROOT/$project_dir" ] && [ -f "$PROJECT_ROOT/$project_dir/package.json" ]; then
            echo "📦 Processing $project_dir..."
            cd "$PROJECT_ROOT/$project_dir"
            
            if npm run test:coverage 2>/dev/null || npm run test 2>/dev/null; then
                echo "✅ Coverage generated for $project_dir"
            else
                echo "⚠️  Coverage generation failed for $project_dir"
            fi
        fi
    done
}

# Function to prepare Android coverage
prepare_android_coverage() {
    echo "📱 Preparing Android coverage..."
    
    for project_dir in catalogizer-android catalogizer-androidtv; do
        if [ -d "$PROJECT_ROOT/$project_dir" ] && [ -f "$PROJECT_ROOT/$project_dir/build.gradle.kts" ]; then
            echo "📱 Processing $project_dir..."
            cd "$PROJECT_ROOT/$project_dir"
            
            if ./gradlew testDebugUnitTest 2>/dev/null; then
                echo "✅ Android tests completed for $project_dir"
            else
                echo "⚠️  Android tests failed for $project_dir"
            fi
        fi
    done
}

# Function to run SonarQube scan
run_sonar_scan() {
    echo "🔍 Running SonarQube analysis..."
    cd "$PROJECT_ROOT"
    
    # Prepare all coverage reports
    prepare_go_coverage
    prepare_js_coverage
    prepare_android_coverage
    
    # Run SonarQube scan
    sonar-scanner \
        -Dsonar.projectKey="$PROJECT_KEY" \
        -Dsonar.host.url="$SONAR_HOST_URL" \
        -Dsonar.login="$SONAR_TOKEN" \
        -Dsonar.projectVersion="1.0.0" \
        -Dsonar.sources="." \
        -Dsonar.exclusions="**/node_modules/**,**/target/**,**/build/**,**/dist/**,**/vendor/**,**/releases/**,**/.git/**,**/reports/**" \
        -Dsonar.test.inclusions="**/*test*.go,**/*test*.js,**/*test*.ts,**/*Test.java,**/*Test.kt,**/*_test.go" \
        -Dsonar.coverage.exclusions="**/*_test.go,**/*test*.js,**/*test*.ts,**/mocks/**,**/stubs/**,**/generated/**" \
        -Dsonar.java.binaries="**/target/classes/**,**/build/classes/**" \
        -Dsonar.go.coverage.reportPaths="catalog-api/coverage.out" \
        -Dsonar.javascript.lcov.reportPaths="catalog-web/coverage/lcov.info,catalogizer-desktop/coverage/lcov.info,catalogizer-api-client/coverage/lcov.info,installer-wizard/coverage/lcov.info" \
        -Dsonar.typescript.lcov.reportPaths="catalog-web/coverage/lcov.info,catalogizer-desktop/coverage/lcov.info,catalogizer-api-client/coverage/lcov.info,installer-wizard/coverage/lcov.info" \
        -Dsonar.qualitygate.wait=true \
        -Dsonar.sourceEncoding=UTF-8
    
    echo "✅ SonarQube analysis completed"
}

# Function to generate report
generate_report() {
    echo "📊 Generating SonarQube report..."
    
    # Get analysis results from SonarQube API
    ANALYSIS_ID=$(curl -s -u "$SONAR_TOKEN:" "$SONAR_HOST_URL/api/project_analyses/search?project=$PROJECT_KEY&ps=1" | jq -r '.analyses[0].key')
    
    if [ -n "$ANALYSIS_ID" ] && [ "$ANALYSIS_ID" != "null" ]; then
        # Get quality gate status
        QUALITY_GATE_STATUS=$(curl -s -u "$SONAR_TOKEN:" "$SONAR_HOST_URL/api/qualitygates/project_status?analysisId=$ANALYSIS_ID" | jq -r '.projectStatus.status')
        
        # Get metrics
        METRICS=$(curl -s -u "$SONAR_TOKEN:" "$SONAR_HOST_URL/api/measures/component?component=$PROJECT_KEY&metricKeys=ncloc,coverage,duplicated_lines_density,violations,bugs,vulnerabilities,code_smells,security_hotspots")
        
        # Generate JSON report
        cat > "$REPORTS_DIR/sonarqube-report.json" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "project_key": "$PROJECT_KEY",
  "analysis_id": "$ANALYSIS_ID",
  "quality_gate_status": "$QUALITY_GATE_STATUS",
  "metrics": $METRICS,
  "sonarqube_url": "$SONAR_HOST_URL/dashboard?id=$PROJECT_KEY"
}
EOF
        
        echo "📊 SonarQube report generated: $REPORTS_DIR/sonarqube-report.json"
        echo "🌐 View results at: $SONAR_HOST_URL/dashboard?id=$PROJECT_KEY"
        
        if [ "$QUALITY_GATE_STATUS" = "OK" ]; then
            echo "✅ Quality Gate: PASSED"
            return 0
        else
            echo "❌ Quality Gate: FAILED"
            return 1
        fi
    else
        echo "❌ Failed to get analysis results"
        return 1
    fi
}

# Main execution
main() {
    echo "🚀 Starting SonarQube Security Analysis..."
    
    # Check prerequisites
    check_sonarqube
    install_scanner
    
    # Run analysis
    run_sonar_scan
    
    # Generate report
    if generate_report; then
        echo "🎉 SonarQube analysis completed successfully!"
        exit 0
    else
        echo "❌ SonarQube analysis failed!"
        exit 1
    fi
}

# Run main function
main "$@"