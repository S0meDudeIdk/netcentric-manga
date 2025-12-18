# Test the manga sync endpoint
$baseUrl = "http://localhost:8080/api/v1"

Write-Host "Testing Manga Sync Endpoint..." -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Test sync with a popular manga
$query = "naruto"
$limit = 2

Write-Host "Syncing manga with query: '$query', limit: $limit" -ForegroundColor Yellow
Write-Host ""

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/manga/sync?query=$query&limit=$limit" -Method POST -ContentType "application/json"
    
    Write-Host "Response:" -ForegroundColor Green
    Write-Host "Total Fetched: $($response.total_fetched)" -ForegroundColor White
    Write-Host "Synced: $($response.synced)" -ForegroundColor Green
    Write-Host "Skipped: $($response.skipped)" -ForegroundColor Yellow
    Write-Host "Failed: $($response.failed)" -ForegroundColor Red
    Write-Host ""
    
    Write-Host "Details:" -ForegroundColor Cyan
    foreach ($detail in $response.details) {
        Write-Host "  $detail"
    }
    Write-Host ""
    
    # Check database
    Write-Host "Checking database..." -ForegroundColor Cyan
    $searchResponse = Invoke-RestMethod -Uri "$baseUrl/manga/?query=$query&limit=10" -Method GET
    Write-Host "Found $($searchResponse.manga.Count) manga in local database" -ForegroundColor Green
    
    if ($searchResponse.manga.Count -gt 0) {
        Write-Host ""
        Write-Host "First manga in database:" -ForegroundColor Yellow
        $firstManga = $searchResponse.manga[0]
        Write-Host "  ID: $($firstManga.id)"
        Write-Host "  Title: $($firstManga.title)"
        Write-Host "  Author: $($firstManga.author)"
        Write-Host "  Status: $($firstManga.status)"
    }
    
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Response: $($_.ErrorDetails.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "Test complete!" -ForegroundColor Cyan
