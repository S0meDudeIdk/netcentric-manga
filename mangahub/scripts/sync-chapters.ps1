# Test script to sync chapters for manga without chapters
# Run this if your manga don't have chapters stored

$baseUrl = "http://localhost:8080"

Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "Testing Chapter Sync" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""

# Test: Trigger chapter sync
Write-Host "Triggering chapter sync for manga without chapters..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/manga/sync-chapters" -Method POST -ContentType "application/json"
    Write-Host "✓ Sync triggered successfully!" -ForegroundColor Green
    Write-Host "  Total Fetched: $($response.total_fetched)" -ForegroundColor White
    Write-Host "  Synced: $($response.synced)" -ForegroundColor White
    Write-Host "  Skipped: $($response.skipped)" -ForegroundColor White
    Write-Host "  Failed: $($response.failed)" -ForegroundColor White
    Write-Host "  Message: $($response.message)" -ForegroundColor White
} catch {
    Write-Host "✗ Sync failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Response: $($_.Exception.Response)" -ForegroundColor Red
}

Write-Host ""
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "Testing Complete" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
