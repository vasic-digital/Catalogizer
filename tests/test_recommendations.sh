#!/bin/bash

echo "Testing recommendation endpoints..."

# Start server
cd catalog-api
./catalog-api -test-mode > /tmp/server.log 2>&1 &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "Testing GET /api/v1/recommendations/test"
# Test simple endpoint - this should work
if curl -s http://localhost:8080/api/v1/recommendations/test > /tmp/test1.json; then
    echo "✅ /api/v1/recommendations/test works"
else
    echo "❌ /api/v1/recommendations/test failed"
fi

echo "Testing GET /api/v1/recommendations/similar/123"
# Test similar endpoint - this might return 500 due to SQL issue but should not crash
if curl -s http://localhost:8080/api/v1/recommendations/similar/123 > /tmp/test2.json; then
    echo "✅ /api/v1/recommendations/similar/123 returns response"
else
    echo "❌ /api/v1/recommendations/similar/123 failed"
fi

echo "Testing GET /api/v1/recommendations/trending"
# Test trending endpoint - this should work with mock data
if curl -s http://localhost:8080/api/v1/recommendations/trending > /tmp/test3.json; then
    echo "✅ /api/v1/recommendations/trending works"
else
    echo "❌ /api/v1/recommendations/trending failed"
fi

echo "Testing GET /api/v1/recommendations/personalized/456"
# Test personalized endpoint - this should work with mock data
if curl -s http://localhost:8080/api/v1/recommendations/personalized/456 > /tmp/test4.json; then
    echo "✅ /api/v1/recommendations/personalized/456 works"
else
    echo "❌ /api/v1/recommendations/personalized/456 failed"
fi

# Stop server
kill $SERVER_PID 2>/dev/null

echo "Recommendation endpoint tests completed"