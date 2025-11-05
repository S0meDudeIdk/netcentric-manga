# MangaHub API Test Scripts

This folder contains comprehensive test scripts for the MangaHub API, including all core features and new bulk import/validation functionality.

## Test Scripts

### 1. PowerShell Test Script (Windows)
**File:** `test-api.ps1`

**Usage:**
```powershell
# Make sure the API server is running first
cd mangahub
go run ./cmd/api-server

# In a new terminal, run the test script
cd test
./test-api.ps1
```

### 2. Bash Test Script (Linux/Mac)
**File:** `test-api.sh`

**Usage:**
```bash
# Make sure the API server is running first
cd mangahub
go run ./cmd/api-server

# In a new terminal, make the script executable and run it
cd test
chmod +x test-api.sh
./test-api.sh
```

**Note:** The bash script requires `jq` for JSON formatting. Install it with:
- Ubuntu/Debian: `sudo apt-get install jq`
- Mac: `brew install jq`

## What the Tests Cover

### Core Authentication & User Features
1. ✅ Health check endpoint
2. ✅ User registration
3. ✅ User login (JWT authentication)
4. ✅ Get user profile
5. ✅ Add manga to library
6. ✅ Update reading progress
7. ✅ Get user library
8. ✅ Get library statistics
9. ✅ Get filtered library
10. ✅ Get recommendations
11. ✅ Batch update progress

### Manga Management Features
12. ✅ Search manga
13. ✅ Get manga stats
14. ✅ Get popular manga
15. ✅ Get available genres
16. ✅ Create manga (admin only)
17. ✅ Update manga (admin only)

### New Bulk Import & Validation Features
18. ✅ **Validate manga data** - Test data validation with valid and invalid entries
19. ✅ **Bulk import manga** - Import multiple manga entries at once
20. ✅ **Get import statistics** - View import and data statistics
21. ✅ **Bulk delete manga** - Delete multiple manga entries at once

## Manual Testing with cURL

If you prefer to test individual endpoints manually, here are some common commands:

### Register User
```bash
curl -X POST "http://localhost:8080/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","email":"admin@mangahub.com","password":"admin123"}'
```

### Login
```bash
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@mangahub.com","password":"admin123"}'
```

### Validate Data (Admin Only)
```bash
# First, get the token from login response
TOKEN="your-jwt-token-here"

curl -X POST "http://localhost:8080/api/v1/manga/validate-data" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "manga": [
      {
        "id": "test-1",
        "title": "Test Manga",
        "author": "Test Author",
        "genres": ["Action"],
        "status": "ongoing"
      }
    ]
  }'
```

### Bulk Import (Admin Only)
```bash
curl -X POST "http://localhost:8080/api/v1/manga/bulk-import" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "manga": [
      {
        "id": "import-test-1",
        "title": "Import Test",
        "author": "Author",
        "genres": ["Action", "Adventure"],
        "status": "ongoing",
        "total_chapters": 50,
        "description": "Test manga for bulk import"
      }
    ],
    "skip_exists": true,
    "validate": true
  }'
```

### Get Import Stats (Admin Only)
```bash
curl -X GET "http://localhost:8080/api/v1/manga/import-stats" \
  -H "Authorization: Bearer $TOKEN"
```

### Bulk Delete (Admin Only)
```bash
curl -X DELETE "http://localhost:8080/api/v1/manga/bulk-delete" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "manga_ids": ["test-import-1", "test-import-2"],
    "confirm": true
  }'
```

## Expected Results

All tests should complete successfully with appropriate status codes:
- `200 OK` - Successful requests
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid data (expected for validation tests)
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions (non-admin accessing admin endpoints)
- `404 Not Found` - Resource not found

## Troubleshooting

### Server Not Running
If you get connection errors, make sure the API server is running:
```bash
cd mangahub
go run ./cmd/api-server
```

### Admin Authentication Failed
The admin user is created with:
- Username: `admin`
- Email: `admin@mangahub.com`
- Password: `admin123`

Admin privileges are granted to users with:
- Email containing "admin"
- Username containing "admin"

### Port Already in Use
If port 8080 is already in use, you can change it in `cmd/api-server/main.go` by modifying the port configuration.

## Additional Notes

- All test scripts include error handling for duplicate registrations and existing data
- The bulk import feature validates data before importing if `validate: true` is set
- The `skip_exists` flag prevents duplicate entries during bulk import
- All admin endpoints require JWT authentication with admin privileges
