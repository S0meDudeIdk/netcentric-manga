# Test MyAnimeList API Integration
Write-Host "==================================" -ForegroundColor Cyan
Write-Host "Testing MyAnimeList API Integration" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080/api/v1/manga"

# Test 1: Search MAL
Write-Host "Test 1: Search MyAnimeList for 'Naruto'" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/mal/search?q=Naruto&limit=5" -Method Get
    Write-Host "✓ Search successful! Found $($response.total) manga" -ForegroundColor Green
    Write-Host "First result: $($response.data[0].title)" -ForegroundColor Green
    $firstMalId = $response.data[0].id -replace "mal-", ""
} catch {
    Write-Host "✗ Search failed: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Test 2: Get Top MAL Manga
Write-Host "Test 2: Get Top Manga from MyAnimeList" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/mal/top?limit=5" -Method Get
    Write-Host "✓ Get top manga successful! Found $($response.total) manga" -ForegroundColor Green
    Write-Host "Top manga: $($response.data[0].title)" -ForegroundColor Green
} catch {
    Write-Host "✗ Get top manga failed: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Test 3: Get Specific MAL Manga by ID
Write-Host "Test 3: Get manga by MAL ID (One Piece - ID: 13)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/mal/13" -Method Get
    Write-Host "✓ Get manga by ID successful!" -ForegroundColor Green
    Write-Host "Title: $($response.title)" -ForegroundColor Green
    Write-Host "Author: $($response.author)" -ForegroundColor Green
    Write-Host "Status: $($response.status)" -ForegroundColor Green
    Write-Host "Genres: $($response.genres -join ', ')" -ForegroundColor Green
} catch {
    Write-Host "✗ Get manga by ID failed: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "==================================" -ForegroundColor Cyan
Write-Host "All tests passed! ✓" -ForegroundColor Green
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "MAL API integration is working correctly!" -ForegroundColor Green
Write-Host ""
Write-Host "Note: The API respects Jikan rate limits (1 request per second)" -ForegroundColor Yellow
Write-Host "If you see rate limit errors, wait a moment and try again." -ForegroundColor Yellow
