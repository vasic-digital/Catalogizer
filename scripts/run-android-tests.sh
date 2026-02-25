#!/bin/bash
# Run Android tests with coverage reporting

set -e

echo "📱 Running Catalogizer Android Tests"
echo "==================================="

cd catalogizer-android

echo "🔧 Cleaning build..."
./gradlew clean

echo "🧪 Running unit tests..."
./gradlew testDebugUnitTest --info

echo "📊 Generating test coverage report..."
./gradlew jacocoTestReport

echo "📁 Coverage reports generated:"
echo "   - HTML: app/build/reports/jacoco/jacocoTestReport/html/index.html"
echo "   - XML:  app/build/reports/jacoco/jacocoTestReport/jacocoTestReport.xml"

# Check if coverage meets threshold (70%)
COVERAGE_FILE="app/build/reports/jacoco/jacocoTestReport/jacocoTestReport.xml"
if [ -f "$COVERAGE_FILE" ]; then
    echo "📈 Checking coverage threshold..."
    # Extract line coverage percentage (simplified)
    COVERAGE=$(grep -o 'linecoverage="[0-9]*\.[0-9]*"' "$COVERAGE_FILE" | head -1 | sed 's/linecoverage="//' | sed 's/"//')
    if [ -n "$COVERAGE" ]; then
        echo "✅ Line coverage: ${COVERAGE}%"
        if (( $(echo "$COVERAGE < 70" | bc -l) )); then
            echo "⚠️ Coverage below 70% target. Consider adding more tests."
        else
            echo "🎉 Coverage meets 70% target!"
        fi
    else
        echo "⚠️ Could not parse coverage from report"
    fi
else
    echo "⚠️ Coverage report not found at $COVERAGE_FILE"
fi

echo ""
echo "🚀 To view coverage report:"
echo "   open app/build/reports/jacoco/jacocoTestReport/html/index.html"
