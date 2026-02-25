#!/bin/bash
# Run desktop app tests with coverage reporting

set -e

echo "💻 Running Catalogizer Desktop Tests"
echo "==================================="

cd catalogizer-desktop

echo "🔧 Installing dependencies if needed..."
if [ ! -d "node_modules" ]; then
    npm install
fi

echo "🧪 Running tests with coverage..."
npm run test:coverage

echo "📊 Coverage reports generated:"
echo "   - HTML: coverage/index.html"
echo "   - Text: coverage/coverage-final.json"

# Check if coverage meets threshold
COVERAGE_SUMMARY="coverage/coverage-summary.json"
if [ -f "$COVERAGE_SUMMARY" ]; then
    echo "📈 Checking coverage threshold..."
    
    # Extract coverage percentages (simplified)
    if command -v jq &> /dev/null; then
        LINES_COV=$(jq -r '.total.lines.pct' "$COVERAGE_SUMMARY")
        STATEMENTS_COV=$(jq -r '.total.statements.pct' "$COVERAGE_SUMMARY")
        FUNCTIONS_COV=$(jq -r '.total.functions.pct' "$COVERAGE_SUMMARY")
        BRANCHES_COV=$(jq -r '.total.branches.pct' "$COVERAGE_SUMMARY")
        
        echo "✅ Coverage Summary:"
        echo "   - Lines: ${LINES_COV}%"
        echo "   - Statements: ${STATEMENTS_COV}%"
        echo "   - Functions: ${FUNCTIONS_COV}%"
        echo "   - Branches: ${BRANCHES_COV}%"
        
        # Check against thresholds
        THRESHOLD=80
        if (( $(echo "$LINES_COV < $THRESHOLD" | bc -l) )); then
            echo "⚠️ Lines coverage below ${THRESHOLD}% target. Consider adding more tests."
        else
            echo "🎉 Lines coverage meets ${THRESHOLD}% target!"
        fi
    else
        echo "⚠️ jq not installed. Install jq to parse coverage summary."
        echo "   Coverage report available at: coverage/index.html"
    fi
else
    echo "⚠️ Coverage summary not found at $COVERAGE_SUMMARY"
    echo "   Raw coverage data available at: coverage/coverage-final.json"
fi

echo ""
echo "🚀 To view coverage report:"
echo "   open coverage/index.html"
