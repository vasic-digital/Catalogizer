#!/bin/bash

# Test script for subtitle endpoints
# Usage: ./test_subtitles.sh

BASE_URL="http://localhost:8080"
USERNAME="testuser"
PASSWORD="testpass"

echo "Testing subtitle endpoints..."

# First, get a JWT token by logging in
echo "1. Logging in to get JWT token..."
LOGIN_REQUEST='{
  "username": "'$USERNAME'",
  "password": "'$PASSWORD'"
}'
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "$LOGIN_REQUEST")

echo "Login response: $LOGIN_RESPONSE"

# Extract token from response
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.session_token // .token // empty')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "Failed to get authentication token. Trying to register user first..."
  
  # Try to register the user first
  REGISTER_REQUEST='{
    "username": "'$USERNAME'",
    "password": "'$PASSWORD'",
    "email": "test@example.com",
    "first_name": "Test",
    "last_name": "User"
  }'
  REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "$REGISTER_REQUEST")
  
  echo "Register response: $REGISTER_RESPONSE"
  
  # Try login again
  LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"$USERNAME\", \"password\": \"$PASSWORD\"}")
  
  echo "Second login response: $LOGIN_RESPONSE"
  TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token // empty')
fi

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "ERROR: Could not get authentication token"
  exit 1
fi

echo "Authentication successful. Token: ${TOKEN:0:50}..."

# Set up authorization header
AUTH_HEADER="Authorization: Bearer $TOKEN"

# Test subtitle endpoints
echo ""
echo "2. Testing GET /api/v1/subtitles/languages"
RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/subtitles/languages" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
echo ""

echo "3. Testing GET /api/v1/subtitles/providers"
RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/subtitles/providers" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
echo ""

echo "4. Testing GET /api/v1/subtitles/search?media_path=/test/movie.mp4&languages=en,es"
RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/subtitles/search?media_path=/test/movie.mp4&languages=en,es" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
echo ""

echo "5. Testing GET /api/v1/subtitles/media/123"
RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/subtitles/media/123" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
echo ""

echo "6. Testing POST /api/v1/subtitles/download"
DOWNLOAD_REQUEST='{
  "media_item_id": 123,
  "result_id": "os_1",
  "language": "en",
  "verify_sync": true
}'
RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/subtitles/download" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "$DOWNLOAD_REQUEST")
echo "Response: $RESPONSE"
echo ""

echo "7. Testing POST /api/v1/subtitles/translate"
TRANSLATE_REQUEST='{
  "subtitle_id": "sub_1234567890",
  "source_language": "en",
  "target_language": "es",
  "use_cache": true
}'
RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/subtitles/translate" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "$TRANSLATE_REQUEST")
echo "Response: $RESPONSE"
echo ""

echo "8. Testing GET /api/v1/subtitles/sub_test_123/verify-sync/123"
RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/subtitles/sub_test_123/verify-sync/123" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
echo ""

echo "All subtitle endpoint tests completed!"