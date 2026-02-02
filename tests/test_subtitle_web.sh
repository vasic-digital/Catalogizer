#!/bin/bash

# Test Subtitle API Endpoints
# This script tests all the subtitle endpoints implemented for Phase 2.3

echo "=== Testing Subtitle API Endpoints ==="
echo

# Base URL
BASE_URL="http://localhost:8080/api/v1"

# Test 1: Get supported languages
echo "1. Testing GET /api/v1/subtitles/languages"
response=$(curl -s -X GET "$BASE_URL/subtitles/languages")
if [[ $response == *"code"* ]]; then
  echo "✅ Languages endpoint working"
else
  echo "❌ Languages endpoint failed"
  echo "Response: $response"
fi
echo

# Test 2: Get supported providers
echo "2. Testing GET /api/v1/subtitles/providers"
response=$(curl -s -X GET "$BASE_URL/subtitles/providers")
if [[ $response == *"name"* ]]; then
  echo "✅ Providers endpoint working"
else
  echo "❌ Providers endpoint failed"
  echo "Response: $response"
fi
echo

# Test 3: Search subtitles (no auth)
echo "3. Testing GET /api/v1/subtitles/search"
response=$(curl -s -X GET "$BASE_URL/subtitles/search?query=test")
if [[ $response == *"results"* ]]; then
  echo "✅ Search endpoint working (no results expected without proper query)"
else
  echo "❌ Search endpoint failed"
  echo "Response: $response"
fi
echo

# Test 4: Download subtitle (should fail without proper ID)
echo "4. Testing POST /api/v1/subtitles/download"
response=$(curl -s -X POST "$BASE_URL/subtitles/download" \
  -H "Content-Type: application/json" \
  -d '{"id":"test-id","language":"en"}')
if [[ $response == *"error"* ]]; then
  echo "✅ Download endpoint working (expected failure with test ID)"
else
  echo "❌ Download endpoint failed"
  echo "Response: $response"
fi
echo

# Test 5: Get media subtitles (without media_id)
echo "5. Testing GET /api/v1/subtitles/media/999"
response=$(curl -s -X GET "$BASE_URL/subtitles/media/999")
if [[ $response == *"subtitles"* ]]; then
  echo "✅ Media subtitles endpoint working"
else
  echo "❌ Media subtitles endpoint failed"
  echo "Response: $response"
fi
echo

# Test 6: Verify sync (should fail without proper IDs)
echo "6. Testing GET /api/v1/subtitles/test/verify-sync/999"
response=$(curl -s -X GET "$BASE_URL/subtitles/test/verify-sync/999")
if [[ $response == *"error"* ]]; then
  echo "✅ Sync verification endpoint working (expected failure with test IDs)"
else
  echo "❌ Sync verification endpoint failed"
  echo "Response: $response"
fi
echo

# Test 7: Translate (should fail without proper text)
echo "7. Testing POST /api/v1/subtitles/translate"
response=$(curl -s -X POST "$BASE_URL/subtitles/translate" \
  -H "Content-Type: application/json" \
  -d '{"text":"","from_language":"en","to_language":"es"}')
if [[ $response == *"error"* ]]; then
  echo "✅ Translate endpoint working (expected failure with empty text)"
else
  echo "❌ Translate endpoint failed"
  echo "Response: $response"
fi
echo

echo "=== Subtitle API Test Complete ==="
echo "Note: Some endpoints may require authentication or valid data for full functionality"