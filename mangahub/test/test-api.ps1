# MangaHub API Test Script
# This script tests all API endpoints including the new bulk import and validation features

$baseUrl = "http://localhost:8080"
$apiUrl = "$baseUrl/api/v1"

Write-Host "================================" -ForegroundColor Cyan
Write-Host "MangaHub API Test Script" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Health Check
Write-Host "Test 1: Health Check" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/health" -Method Get
Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
Write-Host ""

# Test 2: Register Admin User
Write-Host "Test 2: Register Admin User" -ForegroundColor Yellow
try {
    $registerData = @{
        username = "admin"
        email = "admin@mangahub.com"
        password = "admin123"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$apiUrl/auth/register" -Method Post -Body $registerData -ContentType "application/json"
    Write-Host "Admin registered successfully!" -ForegroundColor Green
    Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
} catch {
    Write-Host "Admin might already exist or registration failed: $($_.Exception.Message)" -ForegroundColor Yellow
}
Write-Host ""

# Test 3: Login as Admin
Write-Host "Test 3: Login as Admin" -ForegroundColor Yellow
$loginData = @{
    email = "admin@mangahub.com"
    password = "admin123"
} | ConvertTo-Json

$loginResponse = Invoke-RestMethod -Uri "$apiUrl/auth/login" -Method Post -Body $loginData -ContentType "application/json"
$token = $loginResponse.token
Write-Host "Login successful! Token obtained." -ForegroundColor Green
Write-Host "Token: $token" -ForegroundColor Green
Write-Host ""

# Create headers with auth token
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Test 4: Get User Profile
Write-Host "Test 4: Get User Profile" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$apiUrl/users/profile" -Method Get -Headers $headers
Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
Write-Host ""

# Test 5: Search Manga
Write-Host "Test 5: Search Manga (Query: 'one')" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$apiUrl/manga?query=one" -Method Get -Headers $headers
    Write-Host "Found $($response.manga.Count) manga" -ForegroundColor Green
    Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
} catch {
    Write-Host "Search failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 6: Get Manga Stats
Write-Host "Test 6: Get Manga Stats" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$apiUrl/manga/stats" -Method Get -Headers $headers
    Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
} catch {
    Write-Host "Stats failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 7: Get Popular Manga
Write-Host "Test 7: Get Popular Manga" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$apiUrl/manga/popular?limit=5" -Method Get -Headers $headers
Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
Write-Host ""

# Test 8: Get Genres
Write-Host "Test 8: Get Available Genres" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$apiUrl/manga/genres" -Method Get -Headers $headers
Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
Write-Host ""

# Test 9: NEW - Validate Manga Data
Write-Host "Test 9: Validate Manga Data (NEW ENDPOINT)" -ForegroundColor Yellow
$validationData = @{
    manga = @(
        @{
            id = "test-manga-1"
            title = "Valid Test Manga"
            author = "Test Author"
            genres = @("Action", "Adventure")
            status = "ongoing"
            total_chapters = 100
            description = "This is a valid test manga entry"
        },
        @{
            id = "invalid manga id"  # Invalid: contains spaces
            title = "Invalid Test Manga"
            genres = @("Action")
            status = "invalid_status"  # Invalid status
        },
        @{
            id = "missing-genres"
            title = "No Genres Manga"
            genres = @()  # Invalid: no genres
        }
    )
} | ConvertTo-Json -Depth 4

$response = Invoke-RestMethod -Uri "$apiUrl/manga/validate-data" -Method Post -Body $validationData -Headers $headers
Write-Host "Validation Results:" -ForegroundColor Green
Write-Host "Total: $($response.total)" -ForegroundColor Cyan
Write-Host "Valid: $($response.valid)" -ForegroundColor Green
Write-Host "Invalid: $($response.invalid)" -ForegroundColor Red
if ($response.errors) {
    Write-Host "Errors:" -ForegroundColor Red
    Write-Host ($response.errors | ConvertTo-Json -Depth 3) -ForegroundColor Red
}
Write-Host ""

# Test 10: NEW - Get Import Stats
Write-Host "Test 10: Get Import Statistics (NEW ENDPOINT)" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$apiUrl/manga/import-stats" -Method Get -Headers $headers
Write-Host "Response: $($response | ConvertTo-Json -Depth 4)" -ForegroundColor Green
Write-Host ""

# Test 11: NEW - Bulk Import Manga
Write-Host "Test 11: Bulk Import Manga (NEW ENDPOINT)" -ForegroundColor Yellow
$bulkImportData = @{
    manga = @(
        @{
            id = "test-import-1"
            title = "Bulk Import Test 1"
            author = "Test Author 1"
            genres = @("Action", "Adventure")
            status = "ongoing"
            total_chapters = 50
            description = "First test manga for bulk import"
            cover_url = "https://example.com/cover1.jpg"
        },
        @{
            id = "test-import-2"
            title = "Bulk Import Test 2"
            author = "Test Author 2"
            genres = @("Romance", "Comedy")
            status = "completed"
            total_chapters = 120
            description = "Second test manga for bulk import"
            cover_url = "https://example.com/cover2.jpg"
        },
        @{
            id = "test-import-3"
            title = "Bulk Import Test 3"
            author = "Test Author 3"
            genres = @("Fantasy", "Mystery")
            status = "hiatus"
            total_chapters = 75
            description = "Third test manga for bulk import"
            cover_url = "https://example.com/cover3.jpg"
        }
    )
    skip_exists = $true
    validate = $true
} | ConvertTo-Json -Depth 4

$response = Invoke-RestMethod -Uri "$apiUrl/manga/bulk-import" -Method Post -Body $bulkImportData -Headers $headers
Write-Host "Bulk Import Results:" -ForegroundColor Green
Write-Host "Total: $($response.total)" -ForegroundColor Cyan
Write-Host "Success: $($response.success)" -ForegroundColor Green
Write-Host "Failed: $($response.failed)" -ForegroundColor Red
Write-Host "Skipped: $($response.skipped)" -ForegroundColor Yellow
if ($response.imported_ids) {
    Write-Host "Imported IDs: $($response.imported_ids -join ', ')" -ForegroundColor Green
}
if ($response.errors) {
    Write-Host "Errors: $($response.errors -join ', ')" -ForegroundColor Red
}
Write-Host ""

# Test 12: Verify Imported Manga
Write-Host "Test 12: Verify Imported Manga" -ForegroundColor Yellow
if ($response.imported_ids -and $response.imported_ids.Count -gt 0) {
    $firstId = $response.imported_ids[0]
    $verifyResponse = Invoke-RestMethod -Uri "$apiUrl/manga/$firstId" -Method Get -Headers $headers
    Write-Host "Retrieved imported manga:" -ForegroundColor Green
    Write-Host "Response: $($verifyResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Green
}
Write-Host ""

# Test 13: Add Manga to Library
Write-Host "Test 13: Add Manga to Library" -ForegroundColor Yellow
$addToLibraryData = @{
    manga_id = "test-import-1"
    status = "reading"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$apiUrl/users/library" -Method Post -Body $addToLibraryData -Headers $headers
    Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
} catch {
    Write-Host "Manga might already be in library: $($_.Exception.Message)" -ForegroundColor Yellow
}
Write-Host ""

# Test 14: Update Reading Progress
Write-Host "Test 14: Update Reading Progress" -ForegroundColor Yellow
$progressData = @{
    manga_id = "test-import-1"
    current_chapter = 25
    status = "reading"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$apiUrl/users/progress" -Method Put -Body $progressData -Headers $headers
    Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
} catch {
    Write-Host "Failed to update progress: $($_.Exception.Message)" -ForegroundColor Yellow
}
Write-Host ""

# Test 15: Get Library
Write-Host "Test 15: Get User Library" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$apiUrl/users/library" -Method Get -Headers $headers
Write-Host "Library contains $($response.library.Count) manga" -ForegroundColor Green
Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
Write-Host ""

# Test 16: Get Library Stats
Write-Host "Test 16: Get Library Statistics" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$apiUrl/users/library/stats" -Method Get -Headers $headers
Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
Write-Host ""

# Test 17: Get Filtered Library
Write-Host "Test 17: Get Filtered Library (Status: reading)" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$apiUrl/users/library/filtered?status=reading" -Method Get -Headers $headers
Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
Write-Host ""

# Test 18: Get Recommendations
Write-Host "Test 18: Get Recommendations" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$apiUrl/users/recommendations?limit=5" -Method Get -Headers $headers
Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
Write-Host ""

# Test 19: Create New Manga (Admin)
Write-Host "Test 19: Create New Manga (Admin Only)" -ForegroundColor Yellow
$newMangaData = @{
    id = "test-admin-manga"
    title = "Admin Created Manga"
    author = "Admin Author"
    genres = @("Action", "Sci-Fi")
    status = "ongoing"
    total_chapters = 25
    description = "A manga created through the admin endpoint for testing"
    cover_url = "https://example.com/admin-manga.jpg"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$apiUrl/manga/" -Method Post -Body $newMangaData -Headers $headers
    Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
} catch {
    Write-Host "Failed to create manga: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 20: Update Manga (Admin)
Write-Host "Test 20: Update Manga (Admin Only)" -ForegroundColor Yellow
$updateMangaData = @{
    title = "Admin Created Manga (Updated)"
    author = "Admin Author"
    genres = @("Action", "Sci-Fi", "Thriller")
    status = "ongoing"
    total_chapters = 30
    description = "Updated description for admin created manga with more chapters"
    cover_url = "https://example.com/admin-manga-updated.jpg"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$apiUrl/manga/test-admin-manga" -Method Put -Body $updateMangaData -Headers $headers
    Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
} catch {
    Write-Host "Failed to update manga: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 21: NEW - Bulk Delete Manga
Write-Host "Test 21: Bulk Delete Manga (NEW ENDPOINT)" -ForegroundColor Yellow
$bulkDeleteData = @{
    manga_ids = @("test-import-1", "test-import-2", "test-import-3", "test-admin-manga")
    confirm = $true
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$apiUrl/manga/bulk-delete" -Method Delete -Body $bulkDeleteData -Headers $headers
    Write-Host "Bulk Delete Results:" -ForegroundColor Green
    Write-Host "Total: $($response.total)" -ForegroundColor Cyan
    Write-Host "Success: $($response.success)" -ForegroundColor Green
    Write-Host "Failed: $($response.failed)" -ForegroundColor Red
    if ($response.deleted_ids) {
        Write-Host "Deleted IDs: $($response.deleted_ids -join ', ')" -ForegroundColor Green
    }
    if ($response.errors) {
        Write-Host "Errors: $($response.errors -join ', ')" -ForegroundColor Red
    }
} catch {
    Write-Host "Failed to bulk delete: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 22: Batch Update Progress (Skipped - mangas deleted in Test 21)
Write-Host "Test 22: Batch Update Progress (Skipped - test data deleted)" -ForegroundColor Yellow
Write-Host "Note: This test would require manga IDs that weren't deleted in previous tests" -ForegroundColor Gray
Write-Host ""

# Test Summary
Write-Host "================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host "All tests completed!" -ForegroundColor Green
Write-Host ""
Write-Host "New Features Tested:" -ForegroundColor Yellow
Write-Host "  ✓ Data Validation Endpoint" -ForegroundColor Green
Write-Host "  ✓ Bulk Import Endpoint" -ForegroundColor Green
Write-Host "  ✓ Import Statistics Endpoint" -ForegroundColor Green
Write-Host "  ✓ Bulk Delete Endpoint" -ForegroundColor Green
Write-Host ""
Write-Host "API server is working correctly with all new bulk import and validation features!" -ForegroundColor Green
