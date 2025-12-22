# Docker Setup Validation Script
# Run this to verify your Docker environment is ready

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  MangaHub Docker Setup Validator" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$allGood = $true

# Check 1: Docker is installed and running
Write-Host "[1/7] Checking Docker installation..." -ForegroundColor Yellow
try {
    $dockerVersion = docker --version
    Write-Host "  ✓ Docker found: $dockerVersion" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Docker not found or not running!" -ForegroundColor Red
    Write-Host "    Install Docker Desktop: https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
    $allGood = $false
}

# Check 2: Docker Compose is available
Write-Host "[2/7] Checking Docker Compose..." -ForegroundColor Yellow
try {
    $composeVersion = docker-compose --version
    Write-Host "  ✓ Docker Compose found: $composeVersion" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Docker Compose not found!" -ForegroundColor Red
    $allGood = $false
}

# Check 3: Docker daemon is running
Write-Host "[3/7] Checking Docker daemon..." -ForegroundColor Yellow
try {
    docker info | Out-Null
    Write-Host "  ✓ Docker daemon is running" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Docker daemon is not running!" -ForegroundColor Red
    Write-Host "    Start Docker Desktop and try again" -ForegroundColor Yellow
    $allGood = $false
}

# Check 4: Required files exist
Write-Host "[4/7] Checking required files..." -ForegroundColor Yellow
$requiredFiles = @(
    "docker-compose.yml",
    "Dockerfile",
    ".env",
    "client\web-react\Dockerfile",
    "client\web-react\nginx.conf"
)

$missingFiles = @()
foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Host "  ✓ Found: $file" -ForegroundColor Green
    } else {
        Write-Host "  ✗ Missing: $file" -ForegroundColor Red
        $missingFiles += $file
        $allGood = $false
    }
}

# Check 5: Required ports are available
Write-Host "[5/7] Checking port availability..." -ForegroundColor Yellow
$requiredPorts = @(3000, 8080, 8082, 9001, 9002, 9003, 9010, 9020)
$portsInUse = @()

foreach ($port in $requiredPorts) {
    $connection = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
    if ($connection) {
        Write-Host "  ✗ Port $port is already in use" -ForegroundColor Red
        $portsInUse += $port
        $allGood = $false
    } else {
        Write-Host "  ✓ Port $port is available" -ForegroundColor Green
    }
}

# Check 6: Verify .env file has required variables
Write-Host "[6/7] Checking environment variables..." -ForegroundColor Yellow
if (Test-Path ".env") {
    $envContent = Get-Content ".env" -Raw
    $requiredVars = @("JWT_SECRET", "MAL_CLIENT_ID", "TCP_SERVER_PORT", "UDP_SERVER_PORT", "GRPC_SERVER_PORT")
    
    foreach ($var in $requiredVars) {
        if ($envContent -match "$var=") {
            Write-Host "  ✓ Found: $var" -ForegroundColor Green
        } else {
            Write-Host "  ⚠ Missing or empty: $var" -ForegroundColor Yellow
        }
    }
} else {
    Write-Host "  ✗ .env file not found!" -ForegroundColor Red
    $allGood = $false
}

# Check 7: Check available disk space
Write-Host "[7/7] Checking disk space..." -ForegroundColor Yellow
$drive = (Get-Location).Drive.Name
$disk = Get-PSDrive $drive
$freeSpaceGB = [math]::Round($disk.Free / 1GB, 2)

if ($freeSpaceGB -gt 5) {
    Write-Host "  ✓ Free space: ${freeSpaceGB}GB available" -ForegroundColor Green
} else {
    Write-Host "  ⚠ Low disk space: ${freeSpaceGB}GB (recommend 5GB+)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan

if ($allGood) {
    Write-Host "✓ All checks passed! Ready to start." -ForegroundColor Green
    Write-Host ""
    Write-Host "To start all services, run:" -ForegroundColor Cyan
    Write-Host "  .\docker-start.ps1 up -Detached" -ForegroundColor White
    Write-Host ""
    Write-Host "Or:" -ForegroundColor Cyan
    Write-Host "  docker-compose up -d --build" -ForegroundColor White
} else {
    Write-Host "✗ Some checks failed. Please fix the issues above." -ForegroundColor Red
    Write-Host ""
    
    if ($missingFiles.Count -gt 0) {
        Write-Host "Missing files:" -ForegroundColor Yellow
        foreach ($file in $missingFiles) {
            Write-Host "  - $file" -ForegroundColor Red
        }
        Write-Host ""
    }
    
    if ($portsInUse.Count -gt 0) {
        Write-Host "Ports in use:" -ForegroundColor Yellow
        foreach ($port in $portsInUse) {
            Write-Host "  - $port" -ForegroundColor Red
        }
        Write-Host ""
        Write-Host "Tip: You can modify ports in docker-compose.yml" -ForegroundColor Cyan
        Write-Host "Example: Change '3000:80' to '3001:80' to use port 3001" -ForegroundColor DarkGray
    }
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
