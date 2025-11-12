# MangaHub React Web Client - Quick Start Script
# This script helps you start the React development server

Write-Host "==================================" -ForegroundColor Cyan
Write-Host "  MangaHub React Web Client" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

# Check if we're in the right directory
if (-not (Test-Path "package.json")) {
    Write-Host "Error: package.json not found!" -ForegroundColor Red
    Write-Host "Please run this script from the web-react directory" -ForegroundColor Yellow
    Write-Host "Expected path: client/web-react" -ForegroundColor Yellow
    pause
    exit 1
}

# Check if node_modules exists
if (-not (Test-Path "node_modules")) {
    Write-Host "Installing dependencies..." -ForegroundColor Yellow
    Write-Host "This may take a few minutes on first run..." -ForegroundColor Gray
    npm install
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "`nError: Failed to install dependencies!" -ForegroundColor Red
        pause
        exit 1
    }
    
    Write-Host "`nDependencies installed successfully!" -ForegroundColor Green
} else {
    Write-Host "Dependencies already installed ✓" -ForegroundColor Green
}

Write-Host ""
Write-Host "Starting React development server..." -ForegroundColor Yellow
Write-Host ""
Write-Host "The app will open at: http://localhost:3000" -ForegroundColor Cyan
Write-Host ""
Write-Host "IMPORTANT: Make sure the API server is running on port 8080!" -ForegroundColor Yellow
Write-Host "Run start-server.ps1 from the mangahub directory if not running" -ForegroundColor Gray
Write-Host ""
Write-Host "Press Ctrl+C to stop the server" -ForegroundColor Gray
Write-Host ""

# Start the development server
npm start
