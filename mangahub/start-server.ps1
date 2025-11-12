#!/usr/bin/env pwsh
# Start the MangaHub API Server
# This script ensures the server runs from the correct directory

Write-Host "Starting MangaHub API Server..." -ForegroundColor Cyan
Write-Host "================================`n" -ForegroundColor Cyan

# Navigate to the mangahub directory (project root)
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

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

# Run the server
go run cmd/api-server/main.go
