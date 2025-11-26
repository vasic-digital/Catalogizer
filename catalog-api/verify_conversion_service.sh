#!/bin/bash

# Conversion API Verification Script
# Tests the complete conversion service implementation

echo "=== Catalogizer Conversion API Verification ==="
echo "Date: $(date)"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASS_COUNT=0
TOTAL_TESTS=0

# Helper function to report test results
test_result() {
    local test_name="$1"
    local expected="$2"
    local actual="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [[ "$expected" == "$actual" ]]; then
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name"
        echo "  Expected: $expected"
        echo "  Actual: $actual"
    fi
}

# Helper function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

echo "1. Building Application..."
if go build -o catalogizer-test .; then
    test_result "Application builds successfully" "0" "$?"
else
    test_result "Application builds successfully" "0" "1"
    echo -e "${YELLOW}Build failed, skipping remaining tests${NC}"
    exit 1
fi

echo
echo "2. Running Unit Tests..."

# Handler tests
if go test ./handlers -v -run "TestCreateJob|TestGetJob|TestListJobs|TestCancelJob|TestGetSupportedFormats" >/dev/null 2>&1; then
    test_result "Conversion handler tests pass" "0" "$?"
else
    test_result "Conversion handler tests pass" "0" "1"
fi

# Structure tests
if go test ./tests -v -run "TestConversionAPIEndpoints|TestConversionPermissionConstants|TestConversionStatusConstants" >/dev/null 2>&1; then
    test_result "API structure tests pass" "0" "$?"
else
    test_result "API structure tests pass" "0" "1"
fi

echo
echo "3. Database Schema Verification..."

# Check if database exists
if [[ -f "catalog.db" ]]; then
    test_result "Database file exists" "0" "0"
    
    # Check if conversion_jobs table exists
    if sqlite3 catalog.db "SELECT name FROM sqlite_master WHERE type='table' AND name='conversion_jobs';" | grep -q conversion_jobs; then
        test_result "conversion_jobs table exists" "0" "0"
        
        # Check table structure (should match current migration)
        column_count=$(sqlite3 catalog.db "PRAGMA table_info(conversion_jobs);" | wc -l | tr -d ' ')
        test_result "conversion_jobs has correct columns" "17" "$column_count"
    else
        test_result "conversion_jobs table exists" "0" "1"
    fi
    
    # Check if auth tables exist
    if sqlite3 catalog.db "SELECT name FROM sqlite_master WHERE type='table' AND name='users';" | grep -q users; then
        test_result "users table exists" "0" "0"
    else
        test_result "users table exists" "0" "1"
    fi
else
    test_result "Database file exists" "0" "1"
fi

echo
echo "4. API Route Registration..."

# Start server in background and capture routes
./catalogizer-test > /tmp/catalog-server.log 2>&1 &
SERVER_PID=$!
sleep 2

# Check if server started
if kill -0 $SERVER_PID 2>/dev/null; then
    test_result "Server starts successfully" "0" "0"
    
    # Check if conversion routes are registered (by checking logs)
    if grep -q "POST   /api/v1/conversion/jobs" /tmp/catalog-server.log; then
        test_result "Create job route registered" "0" "0"
    else
        test_result "Create job route registered" "0" "1"
    fi
    
    if grep -q "GET    /api/v1/conversion/formats" /tmp/catalog-server.log; then
        test_result "Get formats route registered" "0" "0"
    else
        test_result "Get formats route registered" "0" "1"
    fi
    
    # Test health endpoint
    if curl -s http://localhost:8080/health >/dev/null 2>&1; then
        test_result "Health endpoint accessible" "0" "0"
        
        # Test formats endpoint (should return 401 without auth)
        status=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/conversion/formats 2>/dev/null)
        test_result "Formats endpoint requires auth" "401" "$status"
    else
        test_result "Health endpoint accessible" "0" "1"
    fi
    
    # Stop server
    kill $SERVER_PID 2>/dev/null
    wait $SERVER_PID 2>/dev/null
else
    test_result "Server starts successfully" "0" "1"
fi

echo
echo "5. External Dependencies Check..."

# Check for FFmpeg
if command_exists ffmpeg; then
    test_result "FFmpeg available" "0" "0"
    ffmpeg_version=$(ffmpeg -version 2>/dev/null | head -n1 | cut -d' ' -f3)
    echo "  FFmpeg version: $ffmpeg_version"
else
    test_result "FFmpeg available" "0" "1"
    echo -e "${YELLOW}  Note: FFmpeg is required for video/audio conversion${NC}"
fi

# Check for ImageMagick
if command_exists convert || command_exists magick; then
    test_result "ImageMagick available" "0" "0"
    if command_exists convert; then
        im_version=$(convert -version 2>/dev/null | head -n1 | cut -d' ' -f3)
    else
        im_version=$(magick -version 2>/dev/null | head -n1 | cut -d' ' -f3)
    fi
    echo "  ImageMagick version: $im_version"
else
    test_result "ImageMagick available" "0" "1"
    echo -e "${YELLOW}  Note: ImageMagick is required for image conversion${NC}"
fi

echo
echo "6. Configuration Verification..."

# Check if conversion handler exists
if [[ -f "handlers/conversion_handler.go" ]]; then
    test_result "Conversion handler file exists" "0" "0"
    
    # Check if handler has required methods (excluding constructor and private methods)
    public_method_count=$(grep "func (h \*ConversionHandler)" handlers/conversion_handler.go | grep -v "getCurrentUser" | wc -l | tr -d ' ')
    test_result "Handler has required methods" "5" "$public_method_count"
else
    test_result "Conversion handler file exists" "0" "1"
fi

# Check if models are defined
if [[ -f "models/user.go" ]]; then
    test_result "Models file exists" "0" "0"
    
    # Check for ConversionJob model
    if grep -q "type ConversionJob struct" models/user.go; then
        test_result "ConversionJob model defined" "0" "0"
    else
        test_result "ConversionJob model defined" "0" "1"
    fi
    
    # Check for permission constants
    if grep -q "PermissionConversionCreate" models/user.go; then
        test_result "Permission constants defined" "0" "0"
    else
        test_result "Permission constants defined" "0" "1"
    fi
else
    test_result "Models file exists" "0" "1"
fi

echo
echo "=== Test Summary ==="
echo "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASS_COUNT${NC}"
echo -e "Failed: ${RED}$((TOTAL_TESTS - PASS_COUNT))${NC}"

if [[ $PASS_COUNT -eq $TOTAL_TESTS ]]; then
    echo -e "${GREEN}✓ All tests passed! Conversion API is fully functional.${NC}"
    exit_code=0
else
    echo -e "${YELLOW}⚠ Some tests failed. Review the output above.${NC}"
    exit_code=1
fi

# Cleanup
rm -f catalogizer-test
rm -f /tmp/catalog-server.log

exit $exit_code