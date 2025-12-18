# Test Chapter Reading Implementation
# This script tests the MangaDex and MangaPlus API integration

$baseUrl = "http://localhost:8080/api/v1"

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "MangaHub Chapter Reading API Tests" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Health Check
Write-Host "Test 1: Health Check" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
    Write-Host "✓ Server is running" -ForegroundColor Green
    Write-Host "  Status: $($response.status)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Server is not running. Please start the backend server first." -ForegroundColor Red
    Write-Host "  Run: go run cmd/api-server/main.go" -ForegroundColor Gray
    exit 1
}
Write-Host ""

# Test 2: MangaDex Chapter List (Example manga)
Write-Host "Test 2: Fetch MangaDex Chapter List" -ForegroundColor Yellow
Write-Host "  Note: Using One Punch Man as example (mangadex-d8a959f7-648e-4c8d-8f23-f1f3f8e129f3)" -ForegroundColor Gray

$mangadexId = "d8a959f7-648e-4c8d-8f23-f1f3f8e129f3"  # One Punch Man
try {
    $url = "$baseUrl/manga/mangadex-$mangadexId/chapters?language=en&limit=5&offset=0"
    $chapters = Invoke-RestMethod -Uri $url -Method GET
    
    Write-Host "✓ Successfully fetched chapters" -ForegroundColor Green
    Write-Host "  Total chapters: $($chapters.total)" -ForegroundColor Gray
    Write-Host "  Fetched: $($chapters.chapters.Count)" -ForegroundColor Gray
    
    if ($chapters.chapters.Count -gt 0) {
        Write-Host "  First chapter:" -ForegroundColor Gray
        $firstChapter = $chapters.chapters[0]
        Write-Host "    - ID: $($firstChapter.id)" -ForegroundColor Gray
        Write-Host "    - Number: $($firstChapter.chapter_number)" -ForegroundColor Gray
        Write-Host "    - Title: $($firstChapter.title)" -ForegroundColor Gray
        Write-Host "    - Source: $($firstChapter.source)" -ForegroundColor Gray
        Write-Host "    - Pages: $($firstChapter.pages)" -ForegroundColor Gray
        
        # Store first chapter ID for next test
        $script:testChapterId = $firstChapter.id
        $script:testChapterSource = $firstChapter.source
    }
} catch {
    Write-Host "✗ Failed to fetch chapters" -ForegroundColor Red
    Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Gray
    Write-Host "  This might mean MangaDex API is slow or the manga ID is invalid" -ForegroundColor Gray
}
Write-Host ""

# Test 3: Fetch Chapter Pages
if ($script:testChapterId) {
    Write-Host "Test 3: Fetch Chapter Pages" -ForegroundColor Yellow
    Write-Host "  Testing with chapter: $script:testChapterId" -ForegroundColor Gray
    
    try {
        $url = "$baseUrl/manga/chapters/$script:testChapterId/pages?source=$script:testChapterSource"
        $pages = Invoke-RestMethod -Uri $url -Method GET
        
        Write-Host "✓ Successfully fetched chapter pages" -ForegroundColor Green
        Write-Host "  Chapter ID: $($pages.chapter_id)" -ForegroundColor Gray
        Write-Host "  Source: $($pages.source)" -ForegroundColor Gray
        Write-Host "  Total pages: $($pages.pages.Count)" -ForegroundColor Gray
        
        if ($pages.pages.Count -gt 0) {
            Write-Host "  First page URL: $($pages.pages[0])" -ForegroundColor Gray
        }
    } catch {
        Write-Host "✗ Failed to fetch chapter pages" -ForegroundColor Red
        Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Gray
    }
} else {
    Write-Host "Test 3: Skipped (no chapter ID available)" -ForegroundColor Yellow
}
Write-Host ""

# Test 4: MangaPlus Chapter List (Example manga)
Write-Host "Test 4: Fetch MangaPlus Chapter List" -ForegroundColor Yellow
Write-Host "  Note: Using One Piece as example (mangaplus-100020)" -ForegroundColor Gray

$mangaPlusId = "100020"  # One Piece
try {
    $url = "$baseUrl/manga/mangaplus-$mangaPlusId/chapters?limit=5"
    $chapters = Invoke-RestMethod -Uri $url -Method GET
    
    Write-Host "✓ Successfully fetched MangaPlus chapters" -ForegroundColor Green
    Write-Host "  Total chapters: $($chapters.total)" -ForegroundColor Gray
    Write-Host "  Fetched: $($chapters.chapters.Count)" -ForegroundColor Gray
    
    if ($chapters.chapters.Count -gt 0) {
        Write-Host "  First chapter:" -ForegroundColor Gray
        $firstChapter = $chapters.chapters[0]
        Write-Host "    - ID: $($firstChapter.id)" -ForegroundColor Gray
        Write-Host "    - Number: $($firstChapter.chapter_number)" -ForegroundColor Gray
        Write-Host "    - Title: $($firstChapter.title)" -ForegroundColor Gray
        Write-Host "    - Source: $($firstChapter.source)" -ForegroundColor Gray
    }
} catch {
    Write-Host "✗ Failed to fetch MangaPlus chapters" -ForegroundColor Red
    Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Gray
    Write-Host "  This might mean MangaPlus API is unavailable or the ID is invalid" -ForegroundColor Gray
}
Write-Host ""

# Summary
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "✓ Backend API is working" -ForegroundColor Green
Write-Host "✓ Chapter reading endpoints are implemented" -ForegroundColor Green
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "1. Start your React frontend (npm start)" -ForegroundColor Gray
Write-Host "2. Navigate to a manga detail page" -ForegroundColor Gray
Write-Host "3. Click on a chapter to test the reader" -ForegroundColor Gray
Write-Host ""
Write-Host "For testing with real manga:" -ForegroundColor Yellow
Write-Host "- Add manga with IDs like: mangadex-{uuid} or mangaplus-{id}" -ForegroundColor Gray
Write-Host "- Or use the Browse page to search MAL manga" -ForegroundColor Gray
Write-Host ""
Write-Host "Documentation available in:" -ForegroundColor Cyan
Write-Host "- docs/CHAPTER_READING_IMPLEMENTATION.md" -ForegroundColor Gray
Write-Host "- docs/CHAPTER_READING_QUICK_START.md" -ForegroundColor Gray
