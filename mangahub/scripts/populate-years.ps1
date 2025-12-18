#!/usr/bin/env pwsh
# Script to re-sync manga and populate publication_year

Write-Host "=== Re-syncing Manga to Populate Publication Year ===" -ForegroundColor Cyan
Write-Host ""

# Check if API server is running
$apiRunning = $false
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/health" -Method GET -TimeoutSec 2 -ErrorAction SilentlyContinue
    if ($response.StatusCode -eq 200) {
        $apiRunning = $true
    }
} catch {
    $apiRunning = $false
}

if (-not $apiRunning) {
    Write-Host "❌ API server is not running!" -ForegroundColor Red
    Write-Host "Please start the API server first with: .\scripts\start-server.ps1" -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ API server is running" -ForegroundColor Green
Write-Host ""

# Prompt user
Write-Host "This will re-sync manga from external APIs to populate publication_year." -ForegroundColor Yellow
Write-Host "Choose sync source:" -ForegroundColor White
Write-Host "  1. MAL (MyAnimeList) - Syncs manga with chapters" -ForegroundColor White
Write-Host "  2. Quick test - Sync just 5 manga" -ForegroundColor White
Write-Host ""
$choice = Read-Host "Enter choice (1-2)"

$query = ""
$limit = 100

switch ($choice) {
    "1" {
        Write-Host "Enter search query (e.g., 'one piece', 'naruto', etc.):" -ForegroundColor Cyan
        $query = Read-Host
        if ([string]::IsNullOrWhiteSpace($query)) {
            Write-Host "❌ Query cannot be empty" -ForegroundColor Red
            exit 1
        }
    }
    "2" {
        Write-Host "Enter search query for quick test:" -ForegroundColor Cyan
        $query = Read-Host
        if ([string]::IsNullOrWhiteSpace($query)) {
            $query = "naruto"
        }
        $limit = 5
    }
    default {
        Write-Host "❌ Invalid choice" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "Starting sync..." -ForegroundColor Cyan

try {
    $url = "http://localhost:8080/api/v1/admin/sync?source=mal&query=$([uri]::EscapeDataString($query))&limit=$limit"
    Write-Host "Calling: $url" -ForegroundColor Gray
    
    $response = Invoke-RestMethod -Uri $url -Method POST -TimeoutSec 300
    
    Write-Host ""
    Write-Host "=== Sync Results ===" -ForegroundColor Green
    Write-Host "Total Fetched: $($response.total_fetched)" -ForegroundColor White
    Write-Host "Synced: $($response.synced)" -ForegroundColor Green
    Write-Host "Skipped: $($response.skipped)" -ForegroundColor Yellow
    Write-Host "Failed: $($response.failed)" -ForegroundColor Red
    
    if ($response.details -and $response.details.Count -gt 0) {
        Write-Host ""
        Write-Host "Details:" -ForegroundColor Cyan
        foreach ($detail in $response.details) {
            Write-Host "  $detail" -ForegroundColor Gray
        }
    }
    
    Write-Host ""
    Write-Host "✓ Sync completed successfully!" -ForegroundColor Green
    Write-Host "Publication years have been populated for synced manga." -ForegroundColor White
    
} catch {
    Write-Host ""
    Write-Host "❌ Sync failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
