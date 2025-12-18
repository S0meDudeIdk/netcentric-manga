#!/usr/bin/env pwsh
# Test script for gRPC library and rating features

$API_BASE = "http://localhost:8080/api/v1/grpc"
$TOKEN = ""

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "gRPC Library & Rating Test Script" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Check if user provided token
if ($args.Length -eq 0) {
    Write-Host "Usage: .\test-grpc-features.ps1 <JWT_TOKEN>" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "To get a token:" -ForegroundColor Yellow
    Write-Host "1. Login at http://localhost:3000/login" -ForegroundColor Yellow
    Write-Host "2. Open browser DevTools > Application > Local Storage" -ForegroundColor Yellow
    Write-Host "3. Copy the 'token' value" -ForegroundColor Yellow
    Write-Host ""
    exit 1
}

$TOKEN = $args[0]
$headers = @{
    "Authorization" = "Bearer $TOKEN"
    "Content-Type" = "application/json"
}

Write-Host "Testing gRPC features..." -ForegroundColor Green
Write-Host ""

# Test 1: Get Library Stats
Write-Host "[TEST 1] Get Library Stats" -ForegroundColor Magenta
try {
    $response = Invoke-RestMethod -Uri "$API_BASE/library/stats" -Method GET -Headers $headers
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Total Manga: $($response.total_manga)" -ForegroundColor White
    Write-Host "Reading: $($response.reading)" -ForegroundColor White
    Write-Host "Completed: $($response.completed)" -ForegroundColor White
    Write-Host "Source: $($response.source)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 2: Add to Library
Write-Host "[TEST 2] Add Manga to Library" -ForegroundColor Magenta
$addBody = @{
    manga_id = "test-manga-1"
    status = "reading"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$API_BASE/library" -Method POST -Headers $headers -Body $addBody
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Message: $($response.message)" -ForegroundColor White
    Write-Host "Source: $($response.source)" -ForegroundColor Cyan
} catch {
    Write-Host "⚠️  Note: $($_.Exception.Message)" -ForegroundColor Yellow
}
Write-Host ""

# Test 3: Get Library
Write-Host "[TEST 3] Get Full Library" -ForegroundColor Magenta
try {
    $response = Invoke-RestMethod -Uri "$API_BASE/library" -Method GET -Headers $headers
    Write-Host "✅ Success!" -ForegroundColor Green
    $totalManga = ($response.reading.Count + $response.completed.Count + $response.plan_to_read.Count + 
                   $response.dropped.Count + $response.on_hold.Count + $response.re_reading.Count)
    Write-Host "Total manga in library: $totalManga" -ForegroundColor White
    Write-Host "Source: $($response.source)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 4: Update Progress
Write-Host "[TEST 4] Update Reading Progress" -ForegroundColor Magenta
$progressBody = @{
    manga_id = "test-manga-1"
    current_chapter = 5
    status = "reading"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$API_BASE/progress/update" -Method PUT -Headers $headers -Body $progressBody
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Message: $($response.message)" -ForegroundColor White
    Write-Host "Source: $($response.source)" -ForegroundColor Cyan
    Write-Host "Note: Check TCP server logs for broadcast message!" -ForegroundColor Yellow
} catch {
    Write-Host "⚠️  Note: $($_.Exception.Message)" -ForegroundColor Yellow
}
Write-Host ""

# Test 5: Rate Manga
Write-Host "[TEST 5] Rate Manga" -ForegroundColor Magenta
$ratingBody = @{
    manga_id = "test-manga-1"
    rating = 9
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$API_BASE/rating" -Method POST -Headers $headers -Body $ratingBody
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Message: $($response.message)" -ForegroundColor White
    Write-Host "Average Rating: $($response.average_rating)" -ForegroundColor White
    Write-Host "Total Ratings: $($response.total_ratings)" -ForegroundColor White
    Write-Host "Source: $($response.source)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 6: Get Manga Ratings
Write-Host "[TEST 6] Get Manga Ratings" -ForegroundColor Magenta
try {
    $response = Invoke-RestMethod -Uri "$API_BASE/rating/test-manga-1" -Method GET -Headers $headers
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Average Rating: $($response.average_rating)" -ForegroundColor White
    Write-Host "Total Ratings: $($response.total_ratings)" -ForegroundColor White
    Write-Host "Your Rating: $($response.user_rating)" -ForegroundColor White
    Write-Host "Source: $($response.source)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 7: Delete Rating
Write-Host "[TEST 7] Delete Rating" -ForegroundColor Magenta
try {
    $response = Invoke-RestMethod -Uri "$API_BASE/rating/test-manga-1" -Method DELETE -Headers $headers
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Message: $($response.message)" -ForegroundColor White
    Write-Host "Source: $($response.source)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 8: Remove from Library
Write-Host "[TEST 8] Remove from Library" -ForegroundColor Magenta
try {
    $response = Invoke-RestMethod -Uri "$API_BASE/library/test-manga-1" -Method DELETE -Headers $headers
    Write-Host "✅ Success!" -ForegroundColor Green
    Write-Host "Message: $($response.message)" -ForegroundColor White
    Write-Host "Source: $($response.source)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "All tests completed!" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Check that all tests show 'Source: grpc'" -ForegroundColor White
Write-Host "2. Verify TCP broadcast in tcp-server terminal (Test 4)" -ForegroundColor White
Write-Host "3. Enable gRPC in web app with .env file" -ForegroundColor White
Write-Host "4. Test features in the web interface" -ForegroundColor White
