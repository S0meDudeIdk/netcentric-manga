# Database Migration Script - Add External URL Support
# Run this to update existing database with external URL columns

$dbPath = "mangahub.db"

Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "Database Migration: Add External URL Support" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path $dbPath)) {
    Write-Host "✗ Database not found at: $dbPath" -ForegroundColor Red
    Write-Host "  This is normal if starting fresh. The server will create the schema." -ForegroundColor Yellow
    exit 0
}

Write-Host "Adding external URL columns to manga_chapters table..." -ForegroundColor Yellow

# Using sqlite3 command
$sql = @"
ALTER TABLE manga_chapters ADD COLUMN external_url TEXT;
ALTER TABLE manga_chapters ADD COLUMN is_external INTEGER DEFAULT 0;
"@

try {
    # Try using sqlite3 if available
    $sql | sqlite3 $dbPath 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Columns added successfully!" -ForegroundColor Green
    } else {
        throw "sqlite3 command failed"
    }
} catch {
    Write-Host "✗ Could not add columns automatically" -ForegroundColor Red
    Write-Host ""
    Write-Host "Manual Migration Required:" -ForegroundColor Yellow
    Write-Host "1. Install SQLite tools or a SQLite browser" -ForegroundColor White
    Write-Host "2. Run these commands:" -ForegroundColor White
    Write-Host ""
    Write-Host "   ALTER TABLE manga_chapters ADD COLUMN external_url TEXT;" -ForegroundColor Cyan
    Write-Host "   ALTER TABLE manga_chapters ADD COLUMN is_external INTEGER DEFAULT 0;" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "OR simply delete the database and restart the server:" -ForegroundColor Yellow
    Write-Host "   Remove-Item mangahub.db" -ForegroundColor Cyan
    Write-Host "   .\scripts\start-server.ps1" -ForegroundColor Cyan
    Write-Host ""
}

Write-Host ""
Write-Host "After migration, run sync to update chapters:" -ForegroundColor Yellow
Write-Host "  .\scripts\sync-chapters.ps1" -ForegroundColor Cyan
Write-Host ""
Write-Host "==================================================" -ForegroundColor Cyan
