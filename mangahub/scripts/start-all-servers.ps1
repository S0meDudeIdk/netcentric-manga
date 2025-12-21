# Start all MangaHub servers in the correct order
# This script starts: TCP Server, UDP Server, gRPC Server, API Server, and Frontend

param(
    [switch]$NoFrontend,
    [switch]$NoBuild
)

Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "Starting Complete MangaHub Environment" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""

# Get the mangahub directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$mangahubDir = Split-Path -Parent $scriptDir

# First, stop any existing servers
Write-Host "🛑 Stopping any existing servers..." -ForegroundColor Yellow
& "$scriptDir\stop-all-servers.ps1"
Start-Sleep -Seconds 2

Write-Host ""
Write-Host "🚀 Starting servers in order..." -ForegroundColor Cyan
Write-Host ""

# Array to track background jobs
$jobs = @()

# 1. Start TCP Server (Port 9001, HTTP Trigger on 9010)
Write-Host "1️⃣  Starting TCP Server..." -ForegroundColor Green
Set-Location "$mangahubDir\cmd\tcp-server"
$tcpJob = Start-Job -ScriptBlock {
    param($dir)
    Set-Location $dir
    go run main.go
} -ArgumentList "$mangahubDir\cmd\tcp-server"
$jobs += @{Name="TCP Server"; Job=$tcpJob; Port=9001}
Write-Host "   TCP Server starting on port 9001 (HTTP trigger: 9010)" -ForegroundColor Gray
Start-Sleep -Seconds 2

# 2. Start UDP Server (Port 9002, HTTP Trigger on 9020)
Write-Host "2️⃣  Starting UDP Server..." -ForegroundColor Green
Set-Location "$mangahubDir\cmd\udp-server"
$udpJob = Start-Job -ScriptBlock {
    param($dir)
    Set-Location $dir
    go run main.go
} -ArgumentList "$mangahubDir\cmd\udp-server"
$jobs += @{Name="UDP Server"; Job=$udpJob; Port=9002}
Write-Host "   UDP Server starting on port 9002 (HTTP trigger: 9020)" -ForegroundColor Gray
Start-Sleep -Seconds 2

# 3. Start gRPC Server (Port 9003)
Write-Host "3️⃣  Starting gRPC Server..." -ForegroundColor Green
Set-Location "$mangahubDir\cmd\grpc-server"
$grpcJob = Start-Job -ScriptBlock {
    param($dir)
    Set-Location $dir
    go run main.go
} -ArgumentList "$mangahubDir\cmd\grpc-server"
$jobs += @{Name="gRPC Server"; Job=$grpcJob; Port=9003}
Write-Host "   gRPC Server starting on port 9003" -ForegroundColor Gray
Start-Sleep -Seconds 3

# 4. Start Main API Server (Port 8080)
Write-Host "4️⃣  Starting Main API Server..." -ForegroundColor Green
Set-Location "$mangahubDir\cmd\api-server"
$apiJob = Start-Job -ScriptBlock {
    param($dir)
    Set-Location $dir
    go run main.go
} -ArgumentList "$mangahubDir\cmd\api-server"
$jobs += @{Name="API Server"; Job=$apiJob; Port=8080}
Write-Host "   API Server starting on port 8080" -ForegroundColor Gray
Start-Sleep -Seconds 3

# 5. Start Frontend (Port 3000) - if not skipped
if (-not $NoFrontend) {
    Write-Host "5️⃣  Starting React Frontend..." -ForegroundColor Green
    Set-Location "$mangahubDir\client\web-react"
    
    # Check if node_modules exists
    if (-not (Test-Path "$mangahubDir\client\web-react\node_modules")) {
        Write-Host "   Installing npm packages..." -ForegroundColor Yellow
        npm install
    }
    
    $frontendJob = Start-Job -ScriptBlock {
        param($dir)
        Set-Location $dir
        $env:BROWSER = "none"  # Don't auto-open browser
        npm start
    } -ArgumentList "$mangahubDir\client\web-react"
    $jobs += @{Name="React Frontend"; Job=$frontendJob; Port=3000}
    Write-Host "   React Frontend starting on port 3000" -ForegroundColor Gray
    Start-Sleep -Seconds 3
}

Write-Host ""
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "✅ All servers started!" -ForegroundColor Green
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""

# Display job status
Write-Host "Running Services:" -ForegroundColor Cyan
foreach ($jobInfo in $jobs) {
    $status = $jobInfo.Job.State
    $color = if ($status -eq "Running") { "Green" } else { "Red" }
    Write-Host "  • $($jobInfo.Name) (Port $($jobInfo.Port)): " -NoNewline
    Write-Host $status -ForegroundColor $color
}

Write-Host ""
Write-Host "Access Points:" -ForegroundColor Cyan
Write-Host "  • Frontend:     http://localhost:3000" -ForegroundColor White
Write-Host "  • API Server:   http://localhost:8080" -ForegroundColor White
Write-Host "  • gRPC Server:  localhost:9003" -ForegroundColor White
Write-Host "  • TCP Server:   localhost:9001 (trigger: http://localhost:9010)" -ForegroundColor White
Write-Host "  • UDP Server:   localhost:9002 (trigger: http://localhost:9020)" -ForegroundColor White

Write-Host ""
Write-Host "📱 For mobile access:" -ForegroundColor Yellow
$ipAddress = (Get-NetIPAddress -AddressFamily IPv4 | Where-Object { $_.IPAddress -like "192.168.*" -or $_.IPAddress -like "10.*" } | Select-Object -First 1).IPAddress
if ($ipAddress) {
    Write-Host "  • Frontend:     http://${ipAddress}:3000" -ForegroundColor White
    Write-Host "  • API Server:   http://${ipAddress}:8080" -ForegroundColor White
}

Write-Host ""
Write-Host "Commands:" -ForegroundColor Cyan
Write-Host "  • View logs: Get-Job | Receive-Job -Keep" -ForegroundColor Gray
Write-Host "  • Stop all:  .\scripts\stop-all-servers.ps1" -ForegroundColor Gray
Write-Host ""
Write-Host "Press Ctrl+C to stop monitoring (servers will keep running)" -ForegroundColor Yellow
Write-Host ""

# Monitor jobs
try {
    while ($true) {
        Start-Sleep -Seconds 5
        
        # Check if any job has failed
        $failedJobs = $jobs | Where-Object { $_.Job.State -eq "Failed" -or $_.Job.State -eq "Stopped" }
        if ($failedJobs) {
            Write-Host ""
            Write-Host "⚠️  Warning: Some servers have stopped!" -ForegroundColor Red
            foreach ($failed in $failedJobs) {
                Write-Host "  • $($failed.Name): $($failed.Job.State)" -ForegroundColor Red
                $output = Receive-Job -Job $failed.Job -ErrorAction SilentlyContinue
                if ($output) {
                    Write-Host "    Last output: $($output | Select-Object -Last 3)" -ForegroundColor Gray
                }
            }
            break
        }
    }
} catch {
    Write-Host ""
    Write-Host "Monitoring stopped. Servers are still running in background." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "To stop all servers, run: .\scripts\stop-all-servers.ps1" -ForegroundColor Cyan
