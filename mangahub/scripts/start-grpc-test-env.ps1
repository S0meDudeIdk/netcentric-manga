# Start all servers for gRPC testing
# This script starts TCP, gRPC, and API servers in separate windows

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "   gRPC Test Environment Startup Script" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Get the script directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir

Write-Host "Project root: $projectRoot" -ForegroundColor Yellow
Write-Host ""

# Check if .env file exists
$envFile = Join-Path $projectRoot ".env"
if (-not (Test-Path $envFile)) {
    Write-Host "WARNING: .env file not found at $envFile" -ForegroundColor Red
    Write-Host "Creating a basic .env file..." -ForegroundColor Yellow
    
    $envContent = @"
# Database
DB_PATH=./manga.db

# Server Ports
PORT=8080
TCP_SERVER_PORT=9000
UDP_SERVER_PORT=9002
GRPC_SERVER_PORT=9001

# Server Addresses
TCP_SERVER_ADDRESS=localhost:9000
GRPC_SERVER_ADDR=localhost:9001

# Gin Mode
GIN_MODE=debug

# JWT Secret (change in production)
JWT_SECRET=your-secret-key-here-change-in-production

# CORS
CORS_ALLOW_ORIGINS=http://localhost:3000,http://localhost:8080

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=100
MAX_REQUEST_SIZE_MB=10
"@
    
    Set-Content -Path $envFile -Value $envContent
    Write-Host "Created .env file with default values" -ForegroundColor Green
    Write-Host ""
}

# Function to start a server in a new window
function Start-ServerWindow {
    param(
        [string]$Name,
        [string]$Command,
        [string]$WorkingDir,
        [string]$Color = "Green"
    )
    
    Write-Host "Starting $Name..." -ForegroundColor $Color
    
    $psCommand = "Set-Location '$WorkingDir'; Write-Host '=== $Name ===' -ForegroundColor $Color; $Command; Read-Host 'Press Enter to close'"
    
    Start-Process powershell -ArgumentList "-NoExit", "-Command", $psCommand
    
    Start-Sleep -Seconds 1
}

# 1. Start TCP Server
Write-Host "1. Starting TCP Server (Port 9000)..." -ForegroundColor Green
$tcpDir = Join-Path $projectRoot "cmd\tcp-server"
Start-ServerWindow -Name "TCP Server" -Command "go run main.go" -WorkingDir $tcpDir -Color "Green"

Write-Host "   Waiting for TCP server to initialize..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

# 2. Start gRPC Server
Write-Host "2. Starting gRPC Server (Port 9001)..." -ForegroundColor Cyan
$grpcDir = Join-Path $projectRoot "cmd\grpc-server"
Start-ServerWindow -Name "gRPC Server" -Command "go run main.go" -WorkingDir $grpcDir -Color "Cyan"

Write-Host "   Waiting for gRPC server to initialize..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

# 3. Start API Server
Write-Host "3. Starting API Server (Port 8080)..." -ForegroundColor Magenta
$apiDir = Join-Path $projectRoot "cmd\api-server"
Start-ServerWindow -Name "API Server" -Command "go run main.go" -WorkingDir $apiDir -Color "Magenta"

Write-Host "   Waiting for API server to initialize..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

Write-Host ""
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "   All Servers Started!" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Server Status:" -ForegroundColor Yellow
Write-Host "  - TCP Server:  http://localhost:9000" -ForegroundColor Green
Write-Host "  - gRPC Server: http://localhost:9001" -ForegroundColor Cyan
Write-Host "  - API Server:  http://localhost:8080" -ForegroundColor Magenta
Write-Host ""
Write-Host "Frontend:" -ForegroundColor Yellow
Write-Host "  Navigate to: http://localhost:3000/grpc-test" -ForegroundColor Blue
Write-Host ""
Write-Host "Test Use Cases:" -ForegroundColor Yellow
Write-Host "  UC-014: Get Manga via gRPC" -ForegroundColor White
Write-Host "  UC-015: Search Manga via gRPC" -ForegroundColor White
Write-Host "  UC-016: Update Progress via gRPC (with TCP broadcast)" -ForegroundColor White
Write-Host ""
Write-Host "To stop servers, close their respective PowerShell windows" -ForegroundColor Yellow
Write-Host ""
Write-Host "Press any key to exit this script..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
