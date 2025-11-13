#!/usr/bin/env pwsh
# Start the MangaHub API Server
# This script ensures the server runs from the correct directory

Write-Host "Starting MangaHub API Server..." -ForegroundColor Cyan
Write-Host "================================`n" -ForegroundColor Cyan

# Navigate to the mangahub directory (project root)
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

# Enable CGO for SQLite
Write-Host "Configuring CGO..." -ForegroundColor Yellow
$env:CGO_ENABLED = "1"
$env:CC = "gcc"

# Check for GCC
try {
    $gccVersion = gcc --version 2>$null | Select-Object -First 1
    Write-Host "✓ GCC found: $gccVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ ERROR: GCC not found in PATH!" -ForegroundColor Red
    Write-Host "`nPlease install MinGW:" -ForegroundColor Yellow
    Write-Host "  Option 1 (Admin PowerShell): choco install mingw" -ForegroundColor White
    Write-Host "  Option 2: Install MSYS2 from https://www.msys2.org" -ForegroundColor White
    Write-Host "`nAfter installation, restart PowerShell and run this script again." -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host "Working directory: $(Get-Location)" -ForegroundColor Yellow
Write-Host "Checking for manga data file..." -ForegroundColor Yellow

if (Test-Path "data/manga.json") {
    $mangaCount = (Get-Content "data/manga.json" | ConvertFrom-Json).Count
    Write-Host "✓ Found manga.json with $mangaCount entries`n" -ForegroundColor Green
} else {
    Write-Host "✗ WARNING: manga.json not found!`n" -ForegroundColor Red
}

Write-Host "Starting API server on port 8080..." -ForegroundColor Cyan
Write-Host "(Press Ctrl+C to stop)`n" -ForegroundColor Gray

# Run the server with CGO enabled
go run cmd/api-server/main.go
