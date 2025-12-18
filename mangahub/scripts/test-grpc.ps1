# gRPC Implementation Test Script
# This script tests all three gRPC use cases

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "   gRPC Implementation Test" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080/api/v1"
$grpcUrl = "$baseUrl/grpc"

# Test credentials (you need to create this user first or modify)
$testUser = @{
    email = "test@example.com"
    password = "testpass123"
}

Write-Host "Step 1: Login and get token" -ForegroundColor Yellow
Write-Host "----------------------------" -ForegroundColor Gray

try {
    $loginBody = @{
        email = $testUser.email
        password = $testUser.password
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.token
    
    Write-Host "✓ Login successful" -ForegroundColor Green
    Write-Host "  User: $($loginResponse.user.username)" -ForegroundColor Gray
    Write-Host "  Token: $($token.Substring(0, 20))..." -ForegroundColor Gray
    Write-Host ""
}
catch {
    Write-Host "✗ Login failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please ensure:" -ForegroundColor Yellow
    Write-Host "  1. API server is running on port 8080" -ForegroundColor White
    Write-Host "  2. Test user exists (email: test@example.com, password: testpass123)" -ForegroundColor White
    Write-Host "  3. Or modify the test credentials in this script" -ForegroundColor White
    Write-Host ""
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Test UC-014: Get Manga via gRPC
Write-Host "Step 2: UC-014 - Get Manga via gRPC" -ForegroundColor Yellow
Write-Host "----------------------------" -ForegroundColor Gray

try {
    $mangaId = "1"
    $getMangaUrl = "$grpcUrl/manga/$mangaId"
    
    Write-Host "Calling: GET $getMangaUrl" -ForegroundColor Gray
    $manga = Invoke-RestMethod -Uri $getMangaUrl -Method Get -Headers $headers
    
    Write-Host "✓ UC-014 Passed" -ForegroundColor Green
    Write-Host "  Manga ID: $($manga.id)" -ForegroundColor Gray
    Write-Host "  Title: $($manga.title)" -ForegroundColor Gray
    Write-Host "  Author: $($manga.author)" -ForegroundColor Gray
    Write-Host "  Source: $($manga.source)" -ForegroundColor Gray
    Write-Host ""
}
catch {
    Write-Host "✗ UC-014 Failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
}

# Test UC-015: Search Manga via gRPC
Write-Host "Step 3: UC-015 - Search Manga via gRPC" -ForegroundColor Yellow
Write-Host "----------------------------" -ForegroundColor Gray

try {
    $searchQuery = "one"
    $searchUrl = "$grpcUrl/manga/search?q=$searchQuery&limit=5"
    
    Write-Host "Calling: GET $searchUrl" -ForegroundColor Gray
    $searchResult = Invoke-RestMethod -Uri $searchUrl -Method Get -Headers $headers
    
    Write-Host "✓ UC-015 Passed" -ForegroundColor Green
    Write-Host "  Query: $searchQuery" -ForegroundColor Gray
    Write-Host "  Results: $($searchResult.total)" -ForegroundColor Gray
    Write-Host "  Source: $($searchResult.source)" -ForegroundColor Gray
    
    if ($searchResult.manga.Count -gt 0) {
        Write-Host "  First result: $($searchResult.manga[0].title)" -ForegroundColor Gray
    }
    Write-Host ""
}
catch {
    Write-Host "✗ UC-015 Failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
}

# Test UC-016: Update Progress via gRPC
Write-Host "Step 4: UC-016 - Update Progress via gRPC" -ForegroundColor Yellow
Write-Host "----------------------------" -ForegroundColor Gray

try {
    # First, add manga to library if not already there
    $mangaId = "1"
    $addToLibraryBody = @{
        manga_id = $mangaId
        status = "reading"
    } | ConvertTo-Json
    
    Write-Host "Adding manga to library first..." -ForegroundColor Gray
    try {
        Invoke-RestMethod -Uri "$baseUrl/users/library" -Method Post -Body $addToLibraryBody -Headers $headers -ErrorAction SilentlyContinue | Out-Null
    } catch {
        # Ignore errors - manga might already be in library
    }
    
    # Now update progress via gRPC
    $progressBody = @{
        manga_id = $mangaId
        current_chapter = 10
        status = "reading"
    } | ConvertTo-Json
    
    $progressUrl = "$grpcUrl/progress/update"
    Write-Host "Calling: PUT $progressUrl" -ForegroundColor Gray
    $progressResult = Invoke-RestMethod -Uri $progressUrl -Method Put -Body $progressBody -Headers $headers
    
    Write-Host "✓ UC-016 Passed" -ForegroundColor Green
    Write-Host "  Success: $($progressResult.success)" -ForegroundColor Gray
    Write-Host "  Message: $($progressResult.message)" -ForegroundColor Gray
    Write-Host "  Source: $($progressResult.source)" -ForegroundColor Gray
    Write-Host "  TCP Broadcast: Triggered (check TCP server logs)" -ForegroundColor Gray
    Write-Host ""
}
catch {
    Write-Host "✗ UC-016 Failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
}

# Summary
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "   Test Summary" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "All tests completed! Check the results above." -ForegroundColor Yellow
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Check TCP server window for broadcast messages" -ForegroundColor White
Write-Host "  2. Check gRPC server logs for RPC calls" -ForegroundColor White
Write-Host "  3. Visit http://localhost:3000/grpc-test for UI testing" -ForegroundColor White
Write-Host ""
Write-Host "Press any key to exit..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
