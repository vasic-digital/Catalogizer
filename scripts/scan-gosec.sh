#!/bin/bash
# Gosec Security Scanner Script for Catalogizer
# Scans Go code for security issues

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORTS_DIR="$PROJECT_ROOT/reports"
GOSEC_DIR="$PROJECT_ROOT/catalog-api"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
SEVERITY="medium"
CONFIDENCE="medium"
OUTPUT_FORMAT="json"
OUTPUT_FILE="$REPORTS_DIR/gosec-results.json"
EXIT_CODE=0
VERBOSE=false

# Function to print usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -s, --severity LEVEL    Minimum severity: low, medium, high (default: medium)"
    echo "  -c, --confidence LEVEL  Minimum confidence: low, medium, high (default: medium)"
    echo "  -f, --format FORMAT     Output format: json, yaml, csv, junit-xml, sarif (default: json)"
    echo "  -o, --output FILE       Output file path"
    echo "  -e, --exit-code CODE    Exit code on findings (default: 0)"
    echo "  -v, --verbose           Enable verbose output"
    echo "  -t, --tests             Include tests in scan"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                               # Scan with defaults"
    echo "  $0 -s high -c high               # Scan high severity only"
    echo "  $0 -f sarif -o results.sarif     # SARIF output"
    echo "  $0 -v -t                         # Verbose, include tests"
}

# Parse command line arguments
INCLUDE_TESTS="-tests=false"
while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--severity)
            SEVERITY="$2"
            shift 2
            ;;
        -c|--confidence)
            CONFIDENCE="$2"
            shift 2
            ;;
        -f|--format)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -e|--exit-code)
            EXIT_CODE="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -t|--tests)
            INCLUDE_TESTS="-tests=true"
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

# Check if Gosec is installed
if ! command -v gosec &> /dev/null; then
    echo -e "${YELLOW}Gosec not found. Installing...${NC}"
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

# Verify Gosec installation
if ! command -v gosec &> /dev/null; then
    echo -e "${RED}Error: Gosec installation failed. Please install manually:${NC}"
    echo "  go install github.com/securego/gosec/v2/cmd/gosec@latest"
    exit 1
fi

GOSEC_VERSION=$(gosec --version 2>&1 | head -1)
echo -e "${GREEN}Using $GOSEC_VERSION${NC}"

# Create Gosec configuration if it doesn't exist
GOSEC_CONFIG="$PROJECT_ROOT/.gosec.json"
if [ ! -f "$GOSEC_CONFIG" ]; then
    echo "Creating default Gosec configuration..."
    cat > "$GOSEC_CONFIG" << 'EOF'
{
  "severity": "medium",
  "confidence": "medium",
  "exclude": [
    "G104",
    "G307"
  ],
  "tests": false,
  "nosec": true,
  "sort": true,
  "fmt": "json",
  "out": "reports/gosec-results.json"
}
EOF
    echo -e "${GREEN}Created $GOSEC_CONFIG${NC}"
fi

# Build scan arguments
SCAN_ARGS=(
    "-severity=$SEVERITY"
    "-confidence=$CONFIDENCE"
    "-fmt=$OUTPUT_FORMAT"
    "-out=$OUTPUT_FILE"
    "$INCLUDE_TESTS"
    "-nosec=true"
    "-sort=true"
)

# Add verbose flag if enabled
if [ "$VERBOSE" = true ]; then
    SCAN_ARGS+=("-verbose")
fi

# Add exclusions from config if it exists
if [ -f "$GOSEC_CONFIG" ]; then
    # Parse exclusions from config
    EXCLUSIONS=$(jq -r '.exclude | join(",")' "$GOSEC_CONFIG" 2>/dev/null || echo "")
    if [ -n "$EXCLUSIONS" ]; then
        SCAN_ARGS+=("-exclude=$EXCLUSIONS")
    fi
fi

echo -e "${GREEN}Starting Gosec scan...${NC}"
echo "Severity: $SEVERITY"
echo "Confidence: $CONFIDENCE"
echo "Format: $OUTPUT_FORMAT"
echo ""

# Change to catalog-api directory for scanning
cd "$GOSEC_DIR"

# Run Gosec
if [ "$VERBOSE" = true ]; then
    echo "Command: gosec ${SCAN_ARGS[*]} ./..."
fi

# Run the scan
gosec "${SCAN_ARGS[@]}" ./... || SCAN_EXIT=$?

# Check results
if [ -f "$OUTPUT_FILE" ]; then
    echo -e "${GREEN}Scan complete. Results saved to: $OUTPUT_FILE${NC}"
    
    # Parse results for summary
    if [ "$OUTPUT_FORMAT" = "json" ] && command -v jq &> /dev/null; then
        ISSUES_COUNT=$(jq '.Stats.found // 0' "$OUTPUT_FILE" 2>/dev/null || echo "0")
        FILES_COUNT=$(jq '.Stats.n_files // 0' "$OUTPUT_FILE" 2>/dev/null || echo "0")
        LINES_COUNT=$(jq '.Stats.n_lines // 0' "$OUTPUT_FILE" 2>/dev/null || echo "0")
        
        echo ""
        echo "========================================"
        echo "Scan Summary:"
        echo "========================================"
        echo "Files scanned: $FILES_COUNT"
        echo "Lines scanned: $LINES_COUNT"
        echo "Issues found: $ISSUES_COUNT"
        echo "========================================"
        
        # Show high severity issues
        if [ "$ISSUES_COUNT" -gt 0 ]; then
            echo ""
            echo "Issues by severity:"
            jq -r '.Issues? // [] | group_by(.severity) | .[] | "\(.[0].severity): \(length)"' "$OUTPUT_FILE" 2>/dev/null || true
            
            # Show first few issues
            echo ""
            echo "Sample issues:"
            jq -r '.Issues? // [] | .[0:5] | .[] | "  [\(.severity)] \(.rule_id): \(.details) (\(.file):\(.line))"' "$OUTPUT_FILE" 2>/dev/null || true
        fi
        
        # Exit with error code if issues found
        if [ "$ISSUES_COUNT" -gt 0 ] && [ "$EXIT_CODE" -ne 0 ]; then
            echo -e "${RED}Found $ISSUES_COUNT security issues${NC}"
            exit "$EXIT_CODE"
        fi
    fi
else
    echo -e "${YELLOW}Warning: Output file not created${NC}"
fi

echo -e "${GREEN}Gosec scan completed${NC}"
exit 0
