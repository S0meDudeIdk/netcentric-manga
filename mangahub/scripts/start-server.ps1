#!/usr/bin/env pwsh
# Start the MangaHub API Server (improved)
# This script will ensure the TCP progress sync server is running (optional),
# then start the API server. It prefers built binaries in `bin/` but falls
# back to `go run` when binaries are not available.

Write-Host "Starting MangaHub API Server (managed)..." -ForegroundColor Cyan
Write-Host "================================`n" -ForegroundColor Cyan

# Navigate to mangahub root (parent of scripts/)
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$mangahubRoot = Split-Path -Parent $scriptDir
Set-Location $mangahubRoot

# Enable CGO for SQLite
Write-Host "Configuring CGO..." -ForegroundColor Yellow
$env:CGO_ENABLED = "1"
$env:CC = "gcc"

# Check for GCC
try {
    $gccVersion = & gcc --version 2>$null | Select-Object -First 1
    Write-Host "✓ GCC found: $gccVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ ERROR: GCC not found in PATH!" -ForegroundColor Red
    Write-Host "\nPlease install MinGW or MSYS2 and ensure gcc is available in PATH." -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host "Working directory: $(Get-Location)" -ForegroundColor Yellow

# Helper to test TCP connectivity
function Test-TcpPort {
    param(
        [string]$hostname = "localhost",
        [int]$port = 9000
    )
    try {
        $c = New-Object System.Net.Sockets.TcpClient
        $async = $c.BeginConnect($hostname, $port, $null, $null)
        $wait = $async.AsyncWaitHandle.WaitOne(1000)
        if (-not $wait) {
            $c.Close()
            return $false
        }
        $c.EndConnect($async)
        $c.Close()
        return $true
    } catch {
        return $false
    }
}

# Start TCP server if not running
$tcpHost = $env:TCP_SERVER_HOST; if (-not $tcpHost) { $tcpHost = "localhost" }
$tcpPort = 9000

if (Test-TcpPort -hostname $tcpHost -port $tcpPort) {
    Write-Host "✓ TCP server already running on ${tcpHost}:${tcpPort}" -ForegroundColor Green
} else {
    Write-Host "TCP server not detected on ${tcpHost}:${tcpPort}. Starting one..." -ForegroundColor Yellow

    # Prefer a built binary if available
    if (Test-Path "bin/tcp-server.exe") {
        Write-Host "Starting built TCP server (bin/tcp-server.exe)..." -ForegroundColor Gray
        Start-Process -FilePath (Resolve-Path "bin/tcp-server.exe").Path -WindowStyle Hidden -PassThru | Out-Null
    } else {
        # Start via `go run` in a new process so it runs independently
        Write-Host "Launching TCP server via 'go run cmd/tcp-server/main.go'..." -ForegroundColor Gray
        $tcpJob = Start-Job -ScriptBlock {
            Set-Location $using:mangahubRoot
            $env:CGO_ENABLED = "1"
            $env:CC = "gcc"
            go run cmd/tcp-server/main.go
        }
        Start-Sleep -Seconds 2  # Give it a moment to start
    }

    # Wait for TCP server to come up (short timeout)
    $maxWait = 15
    $waited = 0
    while ($waited -lt $maxWait) {
        Start-Sleep -Seconds 1
        $waited++
        if (Test-TcpPort -hostname $tcpHost -port $tcpPort) {
            Write-Host "✓ TCP server is now accepting connections" -ForegroundColor Green
            break
        }
    }
    if ($waited -ge $maxWait) {
        Write-Host "⚠️  TCP server did not start within $maxWait seconds. API will start without broadcast support." -ForegroundColor Yellow
    }
}

Write-Host "Starting API server on port 8080..." -ForegroundColor Cyan
Write-Host "(Press Ctrl+C to stop)\n" -ForegroundColor Gray

# Prefer built binary if present
if (Test-Path "bin/api-server.exe") {
    Write-Host "Running binary: bin/api-server.exe" -ForegroundColor Gray
    & (Resolve-Path "bin/api-server.exe").Path
} else {
    Write-Host "Running via 'go run cmd/api-server/main.go'" -ForegroundColor Gray
    go run cmd/api-server/main.go
}
