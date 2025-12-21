# Database Schema Migration Script
# Migrates from old schema (user_progress with status) to new schema (separate library and user_progress tables)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "MangaHub Database Schema Migration" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Detect if running from scripts directory or mangahub directory
$currentDir = Get-Location
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Determine base directory (mangahub root)
if ($currentDir.Path -match "scripts$") {
    # Running from scripts directory, go up one level
    $baseDir = Split-Path -Parent $currentDir
    Write-Host "Detected running from scripts directory" -ForegroundColor Yellow
    Write-Host "Using mangahub directory: $baseDir" -ForegroundColor Cyan
    Write-Host ""
} elseif (Test-Path (Join-Path $currentDir "scripts")) {
    # Running from mangahub directory
    $baseDir = $currentDir
    Write-Host "Detected running from mangahub directory" -ForegroundColor Cyan
    Write-Host ""
} else {
    Write-Host "Error: Cannot detect mangahub directory structure" -ForegroundColor Red
    Write-Host "Please run this script from either:" -ForegroundColor Yellow
    Write-Host "  - mangahub directory: .\scripts\migrate-library.ps1" -ForegroundColor White
    Write-Host "  - scripts directory: .\migrate-library.ps1" -ForegroundColor White
    exit 1
}

# Set paths relative to base directory
$dbPath = Join-Path $baseDir "data\mangahub.db"
$goScriptPath = Join-Path $scriptDir "migrate-library-schema.go"

# Check if database exists
if (-Not (Test-Path $dbPath)) {
    Write-Host "Error: Database not found at $dbPath" -ForegroundColor Red
    Write-Host "Expected location: $dbPath" -ForegroundColor Yellow
    Write-Host "Please make sure the database file exists" -ForegroundColor Yellow
    exit 1
}

Write-Host "Database found: $dbPath" -ForegroundColor Green
Write-Host ""

# Create backup
$backupPath = Join-Path $baseDir "data\mangahub_backup_$(Get-Date -Format 'yyyyMMdd_HHmmss').db"
Write-Host "Creating backup..." -ForegroundColor Yellow
Copy-Item $dbPath $backupPath
Write-Host "✅ Backup created: $backupPath" -ForegroundColor Green
Write-Host ""

# Ask for confirmation
Write-Host "This migration will:" -ForegroundColor Yellow
Write-Host "  1. Create new 'library' table (tracks manga in user's collection with status)" -ForegroundColor White
Write-Host "  2. Create new 'user_progress' table (tracks reading progress for ANY manga)" -ForegroundColor White
Write-Host "  3. Migrate all existing data to new tables" -ForegroundColor White
Write-Host "  4. Keep backup of original data in 'user_progress_backup' table" -ForegroundColor White
Write-Host ""
Write-Host "WARNING: This will modify your database!" -ForegroundColor Red
Write-Host "A backup has been created at: $backupPath" -ForegroundColor Yellow
Write-Host ""

$confirmation = Read-Host "Do you want to continue? (yes/no)"
if ($confirmation -ne "yes") {
    Write-Host "Migration cancelled" -ForegroundColor Yellow
    exit 0
}

Write-Host ""
Write-Host "Running migration..." -ForegroundColor Cyan
Write-Host ""

# Run the migration Go program
go run $goScriptPath $dbPath

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "✅ Migration completed successfully!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "New Database Schema:" -ForegroundColor Cyan
    Write-Host "  - library: Tracks which manga are in user's collection (with status)" -ForegroundColor White
    Write-Host "  - user_progress: Tracks reading progress for ANY manga (even non-library)" -ForegroundColor White
    Write-Host ""
    Write-Host "Benefits:" -ForegroundColor Cyan
    Write-Host "  ✅ TCP Progress: Now works for ANY manga read (not just library items)" -ForegroundColor Green
    Write-Host "  ✅ UDP Notifications: Triggers when manga added to library" -ForegroundColor Green
    Write-Host "  ✅ Better data separation: Reading progress vs Collection management" -ForegroundColor Green
    Write-Host ""
    Write-Host "Backup saved at: $backupPath" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "  1. Restart your API server to use the new schema" -ForegroundColor White
    Write-Host "  2. Test TCP progress updates (works with any manga now)" -ForegroundColor White
    Write-Host "  3. Test UDP notifications (triggers on library additions)" -ForegroundColor White
} else {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "❌ Migration failed!" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
    Write-Host ""
    Write-Host "Your original database is still intact" -ForegroundColor Yellow
    Write-Host "Backup available at: $backupPath" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Please check the error messages above and:" -ForegroundColor Yellow
    Write-Host "  1. Make sure the database is not in use (close API server)" -ForegroundColor White
    Write-Host "  2. Check file permissions" -ForegroundColor White
    Write-Host "  3. Try running the migration again" -ForegroundColor White
    exit 1
}
