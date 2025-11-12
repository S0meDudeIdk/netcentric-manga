#!/usr/bin/env bash
# MangaHub API Test Script (Bash/Git Bash version)
# This script tests all API endpoints including the new bulk import and validation features
# Compatible with Git Bash on Windows

BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

echo "================================"
echo "MangaHub API Test Script"
echo "================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Use curl.exe on Windows for better compatibility
CURL_CMD="curl.exe"

# Test 1: Health Check
echo -e "${YELLOW}Test 1: Health Check${NC}"
$CURL_CMD -s "$BASE_URL/health"
echo -e "\n"

# Test 2: Register Admin User
echo -e "${YELLOW}Test 2: Register Admin User${NC}"
$CURL_CMD -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","email":"admin@mangahub.com","password":"admin123"}'
echo -e "\n"

# Test 3: Login as Admin
echo -e "${YELLOW}Test 3: Login as Admin${NC}"
LOGIN_RESPONSE=$($CURL_CMD -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@mangahub.com","password":"admin123"}')
echo "$LOGIN_RESPONSE"
TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo -e "${GREEN}Token: $TOKEN${NC}"
echo ""

# Test 4: Get User Profile
echo -e "${YELLOW}Test 4: Get User Profile${NC}"
$CURL_CMD -s "$API_URL/users/profile" \
  -H "Authorization: Bearer $TOKEN"
echo -e "\n"

# Test 5: Search Manga
echo -e "${YELLOW}Test 5: Search Manga (Query: 'one')${NC}"
$CURL_CMD -s "$API_URL/manga?query=one" \
  -H "Authorization: Bearer $TOKEN"
echo -e "\n"

# Test 6: Get Manga Stats
echo -e "${YELLOW}Test 6: Get Manga Stats${NC}"
$CURL_CMD -s "$API_URL/manga/stats" \
  -H "Authorization: Bearer $TOKEN"
echo -e "\n"

# Test 7: Get Popular Manga
echo -e "${YELLOW}Test 7: Get Popular Manga${NC}"
$CURL_CMD -s "$API_URL/manga/popular?limit=5" \
  -H "Authorization: Bearer $TOKEN"
echo -e "\n"

# Test 8: Get Genres
echo -e "${YELLOW}Test 8: Get Available Genres${NC}"
$CURL_CMD -s "$API_URL/manga/genres" \
  -H "Authorization: Bearer $TOKEN"
echo -e "\n"

# Test 9: NEW - Validate Manga Data
echo -e "${YELLOW}Test 9: Validate Manga Data (NEW ENDPOINT)${NC}"
$CURL_CMD -s -X POST "$API_URL/manga/validate-data" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "manga": [
      {
        "id": "test-manga-1",
        "title": "Valid Test Manga",
        "author": "Test Author",
        "genres": ["Action", "Adventure"],
        "status": "ongoing",
        "total_chapters": 100,
        "description": "This is a valid test manga entry"
      },
      {
        "id": "invalid manga id",
        "title": "Invalid Test Manga",
        "genres": ["Action"],
        "status": "invalid_status"
      },
      {
        "id": "missing-genres",
        "title": "No Genres Manga",
        "genres": []
      }
    ]
  }'
echo -e "\n"

# Test 10: NEW - Get Import Stats
echo -e "${YELLOW}Test 10: Get Import Statistics (NEW ENDPOINT)${NC}"
$CURL_CMD -s "$API_URL/manga/import-stats" \
  -H "Authorization: Bearer $TOKEN"
echo -e "\n"

# Test 11: NEW - Bulk Import Manga
echo -e "${YELLOW}Test 11: Bulk Import Manga (NEW ENDPOINT)${NC}"
$CURL_CMD -s -X POST "$API_URL/manga/bulk-import" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "manga": [
      {
        "id": "test-import-1",
        "title": "Bulk Import Test 1",
        "author": "Test Author 1",
        "genres": ["Action", "Adventure"],
        "status": "ongoing",
        "total_chapters": 50,
        "description": "First test manga for bulk import",
        "cover_url": "https://example.com/cover1.jpg"
      },
      {
        "id": "test-import-2",
        "title": "Bulk Import Test 2",
        "author": "Test Author 2",
        "genres": ["Romance", "Comedy"],
        "status": "completed",
        "total_chapters": 120,
        "description": "Second test manga for bulk import",
        "cover_url": "https://example.com/cover2.jpg"
      },
      {
        "id": "test-import-3",
        "title": "Bulk Import Test 3",
        "author": "Test Author 3",
        "genres": ["Fantasy", "Mystery"],
        "status": "hiatus",
        "total_chapters": 75,
        "description": "Third test manga for bulk import",
        "cover_url": "https://example.com/cover3.jpg"
      }
    ],
    "skip_exists": true,
    "validate": true
  }'
echo -e "\n"

# Test 12: Add Manga to Library
echo -e "${YELLOW}Test 12: Add Manga to Library${NC}"
$CURL_CMD -s -X POST "$API_URL/users/library" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"manga_id":"one-piece","status":"reading"}'
echo -e "\n"

# Test 13: Update Progress
echo -e "${YELLOW}Test 13: Update Reading Progress${NC}"
$CURL_CMD -s -X PUT "$API_URL/users/progress" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"manga_id":"one-piece","current_chapter":1050,"status":"reading"}'
echo -e "\n"

# Test 14: Get Library
echo -e "${YELLOW}Test 14: Get User Library${NC}"
$CURL_CMD -s "$API_URL/users/library" \
  -H "Authorization: Bearer $TOKEN"
echo -e "\n"

# Test 15: Get Library Stats
echo -e "${YELLOW}Test 15: Get Library Statistics${NC}"
$CURL_CMD -s "$API_URL/users/library/stats" \
  -H "Authorization: Bearer $TOKEN"
echo -e "\n"

# Test 16: NEW - Bulk Delete Manga
echo -e "${YELLOW}Test 16: Bulk Delete Manga (NEW ENDPOINT)${NC}"
$CURL_CMD -s -X DELETE "$API_URL/manga/bulk-delete" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "manga_ids": ["test-import-1", "test-import-2", "test-import-3"],
    "confirm": true
  }'
echo -e "\n"

# Test Summary
echo "================================"
echo -e "${CYAN}Test Summary${NC}"
echo "================================"
echo -e "${GREEN}All tests completed!${NC}"
echo ""
echo -e "${YELLOW}New Features Tested:${NC}"
echo -e "${GREEN}  ✓ Data Validation Endpoint${NC}"
echo -e "${GREEN}  ✓ Bulk Import Endpoint${NC}"
echo -e "${GREEN}  ✓ Import Statistics Endpoint${NC}"
echo -e "${GREEN}  ✓ Bulk Delete Endpoint${NC}"
echo ""
