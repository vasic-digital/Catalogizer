#!/bin/bash
# Trivy Security Scanner Script for Catalogizer
# Supports filesystem, container, and repository scanning

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORTS_DIR="$PROJECT_ROOT/reports"
TRIVY_CACHE_DIR="$PROJECT_ROOT/.cache/trivy"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
SCAN_TYPE="fs"
SEVERITY="HIGH,CRITICAL"
FORMAT="json"
OUTPUT_FILE="$REPORTS_DIR/trivy-results.json"
EXIT_CODE=0

# Function to print usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -t, --type TYPE       Scan type: fs, image, repo (default: fs)"
    echo "  -s, --severity LEVEL  Severity levels: UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL"
    echo "  -f, --format FORMAT   Output format: json, sarif, table (default: json)"
    echo "  -o, --output FILE     Output file path"
    echo "  -i, --image IMAGE     Container image to scan (for type=image)"
    echo "  --exit-code CODE      Exit code on findings (default: 0)"
    echo "  --cache-dir DIR       Cache directory (default: .cache/trivy)"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Scan filesystem with defaults"
    echo "  $0 -t fs -s CRITICAL                  # Scan filesystem, critical only"
    echo "  $0 -t image -i catalogizer-api:latest # Scan container image"
    echo "  $0 -t repo -f sarif -o results.sarif  # Scan as repository, SARIF output"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            SCAN_TYPE="$2"
            shift 2
            ;;
        -s|--severity)
            SEVERITY="$2"
            shift 2
            ;;
        -f|--format)
            FORMAT="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -i|--image)
            IMAGE="$2"
            shift 2
            ;;
        --exit-code)
            EXIT_CODE="$2"
            shift 2
            ;;
        --cache-dir)
            TRIVY_CACHE_DIR="$2"
            shift 2
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

# Create necessary directories
mkdir -p "$REPORTS_DIR"
mkdir -p "$TRIVY_CACHE_DIR"

# Check if Trivy is installed
if ! command -v trivy &> /dev/null; then
    echo -e "${YELLOW}Trivy not found. Installing...${NC}"
    
    # Try to install using official installer
    if command -v curl &> /dev/null; then
        curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
    else
        echo -e "${RED}Error: curl not found. Please install Trivy manually.${NC}"
        exit 1
    fi
fi

# Verify Trivy installation
if ! command -v trivy &> /dev/null; then
    echo -e "${RED}Error: Trivy installation failed${NC}"
    exit 1
fi

TRIVY_VERSION=$(trivy version --format json 2>/dev/null | grep -o '"Version":"[^"]*"' | cut -d'"' -f4 || trivy version 2>&1 | head -1)
echo -e "${GREEN}Using Trivy version: $TRIVY_VERSION${NC}"

# Build scan command based on type
echo -e "${GREEN}Starting Trivy scan (type: $SCAN_TYPE, severity: $SEVERITY)${NC}"

 case $SCAN_TYPE in
    fs)
        echo "Scanning filesystem..."
        trivy filesystem \
            --scanners vuln,secret,config \
            --severity "$SEVERITY" \
            --format "$FORMAT" \
            --output "$OUTPUT_FILE" \
            --cache-dir "$TRIVY_CACHE_DIR" \
            --skip-dirs "node_modules,vendor,dist,build,target,.git,sonarqube" \
            --skip-files "*.test.js,*.test.ts,*_test.go" \
            "$PROJECT_ROOT"
        ;;
    
    image)
        if [ -z "$IMAGE" ]; then
            echo -e "${RED}Error: Image name required for image scan${NC}"
            usage
            exit 1
        fi
        echo "Scanning container image: $IMAGE"
        trivy image \
            --scanners vuln,secret,config \
            --severity "$SEVERITY" \
            --format "$FORMAT" \
            --output "$OUTPUT_FILE" \
            --cache-dir "$TRIVY_CACHE_DIR" \
            "$IMAGE"
        ;;
    
    repo)
        echo "Scanning repository..."
        trivy repository \
            --scanners vuln,secret,config \
            --severity "$SEVERITY" \
            --format "$FORMAT" \
            --output "$OUTPUT_FILE" \
            --cache-dir "$TRIVY_CACHE_DIR" \
            --skip-dirs "node_modules,vendor,dist,build,target,.git" \
            "$PROJECT_ROOT"
        ;;
    
    *)
        echo -e "${RED}Error: Unknown scan type: $SCAN_TYPE${NC}"
        usage
        exit 1
        ;;
esac

# Check scan results
if [ -f "$OUTPUT_FILE" ]; then
    echo -e "${GREEN}Scan complete. Results saved to: $OUTPUT_FILE${NC}"
    
    # Parse results for summary (JSON format)
    if [ "$FORMAT" = "json" ] && command -v jq &> /dev/null; then
        VULN_COUNT=$(jq '[.Results[]?.Vulnerabilities? // [] | length] | add // 0' "$OUTPUT_FILE" 2>/dev/null || echo "0")
        SECRET_COUNT=$(jq '[.Results[]?.Secrets? // [] | length] | add // 0' "$OUTPUT_FILE" 2>/dev/null || echo "0")
        CONFIG_COUNT=$(jq '[.Results[]?.Misconfigurations? // [] | length] | add // 0' "$OUTPUT_FILE" 2>/dev/null || echo "0")
        
        echo ""
        echo "========================================"
        echo "Scan Summary:"
        echo "========================================"
        echo "Vulnerabilities found: $VULN_COUNT"
        echo "Secrets found: $SECRET_COUNT"
        echo "Misconfigurations found: $CONFIG_COUNT"
        echo "========================================"
        
        # Generate HTML report if jq is available
        HTML_OUTPUT="${OUTPUT_FILE%.json}.html"
        if command -v trivy &> /dev/null; then
            echo "Generating HTML report..."
            trivy filesystem \
                --scanners vuln,secret,config \
                --severity "$SEVERITY" \
                --format template \
                --template "@/usr/local/share/trivy/templates/html.tpl" \
                --output "$HTML_OUTPUT" \
                --skip-dirs "node_modules,vendor,dist,build,target,.git" \
                "$PROJECT_ROOT" 2>/dev/null || echo "HTML template not available, skipping"
        fi
        
        # Exit with error code if vulnerabilities found and exit-code is set
        TOTAL_ISSUES=$((VULN_COUNT + SECRET_COUNT + CONFIG_COUNT))
        if [ "$TOTAL_ISSUES" -gt 0 ] && [ "$EXIT_CODE" -ne 0 ]; then
            echo -e "${RED}Found $TOTAL_ISSUES issues${NC}"
            exit "$EXIT_CODE"
        fi
    fi
else
    echo -e "${RED}Error: Scan failed or output file not created${NC}"
    exit 1
fi

echo -e "${GREEN}Trivy scan completed successfully${NC}"
exit 0
