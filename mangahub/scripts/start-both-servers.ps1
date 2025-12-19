# Start both API server and Fetch-Manga server
# API server starts first, then Fetch-Manga server

param(
    [switch]$NoBuild,
    [switch]$SkipFetch
)

Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "Starting MangaHub Multi-Server Environment" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""

# Get the mangahub directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$mangahubDir = Split-Path -Parent $scriptDir

# Check if servers are already running
$apiRunning = Get-Process -Name "api-server" -ErrorAction SilentlyContinue
$fetchRunning = Get-Process -Name "fetch-manga-server" -ErrorAction SilentlyContinue

if ($apiRunning -or $fetchRunning) {
    Write-Host "⚠️  Warning: Servers may already be running!" -ForegroundColor Yellow
    if ($apiRunning) { Write-Host "  - API Server (PID: $($apiRunning.Id))" -ForegroundColor Yellow }
    if ($fetchRunning) { Write-Host "  - Fetch-Manga Server (PID: $($fetchRunning.Id))" -ForegroundColor Yellow }
    Write-Host ""
    $continue = Read-Host "Continue anyway? (y/n)"
    if ($continue -ne "y") {
        Write-Host "Aborted." -ForegroundColor Red
        exit 1
    }
}

# Build servers if not skipped
if (-not $NoBuild) {
    Write-Host "🔨 Building servers..." -ForegroundColor Yellow
    
    # Build API Server
    Write-Host "  Building API Server..." -ForegroundColor Cyan
    Set-Location "$mangahubDir\cmd\api-server"
    go build -o "$mangahubDir\bin\api-server.exe" .
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Failed to build API Server" -ForegroundColor Red
        exit 1
    }
    
    # Build Fetch-Manga Server (only if not skipped)
    if (-not $SkipFetch) {
        Write-Host "  Building Fetch-Manga Server..." -ForegroundColor Cyan
        Set-Location "$mangahubDir\cmd\fetch-manga-server"
        go build -o "$mangahubDir\bin\fetch-manga-server.exe" .
        if ($LASTEXITCODE -ne 0) {
            Write-Host "❌ Failed to build Fetch-Manga Server" -ForegroundColor Red
            exit 1
        }
    }
    
    Write-Host "✓ Build completed successfully" -ForegroundColor Green
    Write-Host ""
}

# Start API Server
Write-Host "🚀 Starting API Server (Port 8080)..." -ForegroundColor Green
Set-Location "$mangahubDir\cmd\api-server"
Start-Process -FilePath "$mangahubDir\bin\api-server.exe" -WindowStyle Normal
Write-Host "  Waiting for API Server to initialize..." -ForegroundColor Cyan
Start-Sleep -Seconds 3

# Check if API Server started successfully
$apiProcess = Get-Process -Name "api-server" -ErrorAction SilentlyContinue
if (-not $apiProcess) {
    Write-Host "❌ API Server failed to start" -ForegroundColor Red
    exit 1
}
Write-Host "✓ API Server started (PID: $($apiProcess.Id))" -ForegroundColor Green
Write-Host ""

# Start Fetch-Manga Server (if not skipped)
if (-not $SkipFetch) {
    Write-Host "🚀 Starting Fetch-Manga Server (Port 8082)..." -ForegroundColor Green
    Set-Location "$mangahubDir\cmd\fetch-manga-server"
    Start-Process -FilePath "$mangahubDir\bin\fetch-manga-server.exe" -WindowStyle Normal
    Write-Host "  Waiting for Fetch-Manga Server to initialize..." -ForegroundColor Cyan
    Start-Sleep -Seconds 3
    
    # Check if Fetch-Manga Server started successfully
    $fetchProcess = Get-Process -Name "fetch-manga-server" -ErrorAction SilentlyContinue
    if (-not $fetchProcess) {
        Write-Host "❌ Fetch-Manga Server failed to start" -ForegroundColor Red
        Write-Host "   API Server is still running. Stop it manually if needed." -ForegroundColor Yellow
        exit 1
    }
    Write-Host "✓ Fetch-Manga Server started (PID: $($fetchProcess.Id))" -ForegroundColor Green
    Write-Host ""
}

# Summary
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "All servers started successfully!" -ForegroundColor Green
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Running Services:" -ForegroundColor White
Write-Host "  • API Server:          http://localhost:8080" -ForegroundColor Cyan
if (-not $SkipFetch) {
    Write-Host "  • Fetch-Manga Server:  http://localhost:8082" -ForegroundColor Cyan
}
Write-Host ""
Write-Host "Available Endpoints:" -ForegroundColor White
Write-Host "  API Server:" -ForegroundColor Yellow
Write-Host "    - Health:     http://localhost:8080/health" -ForegroundColor Gray
Write-Host "    - Manga API:  http://localhost:8080/api/v1/manga" -ForegroundColor Gray
Write-Host "    - Auth API:   http://localhost:8080/api/v1/auth" -ForegroundColor Gray
if (-not $SkipFetch) {
    Write-Host ""
    Write-Host "  Fetch-Manga Server:" -ForegroundColor Yellow
    Write-Host "    - Health:     http://localhost:8082/health" -ForegroundColor Gray
    Write-Host "    - MAL Search: http://localhost:8082/api/v1/manga/mal/search" -ForegroundColor Gray
    Write-Host "    - MAL Top:    http://localhost:8082/api/v1/manga/mal/top" -ForegroundColor Gray
    Write-Host "    - Sync:       http://localhost:8082/api/v1/manga/sync" -ForegroundColor Gray
}
Write-Host ""
Write-Host "To stop all servers:" -ForegroundColor Yellow
Write-Host "  Stop-Process -Name 'api-server','fetch-manga-server'" -ForegroundColor Gray
Write-Host ""
Write-Host "Press Ctrl+C to exit this script" -ForegroundColor DarkGray
Write-Host "=============================================" -ForegroundColor Cyan

# Keep script running to monitor servers
try {
    while ($true) {
        Start-Sleep -Seconds 5
        
        # Check if servers are still running
        $apiRunning = Get-Process -Name "api-server" -ErrorAction SilentlyContinue
        if (-not $SkipFetch) {
            $fetchRunning = Get-Process -Name "fetch-manga-server" -ErrorAction SilentlyContinue
        }
        
        if (-not $apiRunning) {
            Write-Host ""
            Write-Host "⚠️  API Server stopped unexpectedly" -ForegroundColor Red
            break
        }
        
        if (-not $SkipFetch -and -not $fetchRunning) {
            Write-Host ""
            Write-Host "⚠️  Fetch-Manga Server stopped unexpectedly" -ForegroundColor Red
            break
        }
    }
} catch {
    Write-Host ""
    Write-Host "Shutting down..." -ForegroundColor Yellow
}

# Cleanup on exit
Write-Host "Stopping servers..." -ForegroundColor Yellow
Stop-Process -Name "api-server" -ErrorAction SilentlyContinue -Force
if (-not $SkipFetch) {
    Stop-Process -Name "fetch-manga-server" -ErrorAction SilentlyContinue -Force
}
Write-Host "Servers stopped" -ForegroundColor Green
