#!/usr/bin/env pwsh
# Build script for MangaHub executables
# Creates all necessary binaries in bin/ folder

Write-Host "Building MangaHub components..." -ForegroundColor Cyan
Write-Host "================================`n" -ForegroundColor Cyan

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

# Enable CGO for SQLite
$env:CGO_ENABLED = "1"
$env:CC = "gcc"

# Create bin directory if it doesn't exist
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Build components
$builds = @(
    @{Name="API Server"; Path="cmd/api-server"; Output="bin/api-server.exe"},
    @{Name="TCP Server"; Path="cmd/tcp-server"; Output="bin/tcp-server.exe"},
    @{Name="UDP Server"; Path="cmd/udp-server"; Output="bin/udp-server.exe"},
    @{Name="gRPC Server"; Path="cmd/grpc-server"; Output="bin/grpc-server.exe"},
    @{Name="gRPC Client Test"; Path="cmd/grpc-client-test"; Output="bin/grpc-client-test.exe"},
    @{Name="CLI Client"; Path="client/cli"; Output="bin/cli-client.exe"}
)

$success = 0
$failed = 0

foreach ($build in $builds) {
    Write-Host "Building $($build.Name)..." -ForegroundColor Yellow
    go build -o $build.Output "./$($build.Path)"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ $($build.Name) built successfully" -ForegroundColor Green
        $success++
    } else {
        Write-Host "✗ Failed to build $($build.Name)" -ForegroundColor Red
        $failed++
    }
}

Write-Host "`n================================" -ForegroundColor Cyan
Write-Host "Build Summary" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host "Success: $success" -ForegroundColor Green
Write-Host "Failed: $failed" -ForegroundColor $(if ($failed -gt 0) { "Red" } else { "Gray" })
Write-Host "`nBinaries location: bin/" -ForegroundColor Yellow
