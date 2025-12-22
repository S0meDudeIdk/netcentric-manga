# MangaHub Docker Startup Script
# This script helps you quickly start, stop, and manage the Docker containers

param(
    [Parameter(Position=0)]
    [ValidateSet('up', 'down', 'restart', 'logs', 'build', 'status', 'clean')]
    [string]$Action = 'up',
    
    [Parameter()]
    [switch]$Detached,
    
    [Parameter()]
    [string]$Service
)

$ErrorActionPreference = "Stop"

Write-Host "==================================" -ForegroundColor Cyan
Write-Host "   MangaHub Docker Manager" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

function Show-Help {
    Write-Host "Usage: .\docker-start.ps1 [action] [options]" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Actions:" -ForegroundColor Green
    Write-Host "  up        - Start all services (default)"
    Write-Host "  down      - Stop all services"
    Write-Host "  restart   - Restart all services"
    Write-Host "  logs      - View logs"
    Write-Host "  build     - Rebuild containers"
    Write-Host "  status    - Show service status"
    Write-Host "  clean     - Stop and remove all containers and volumes"
    Write-Host ""
    Write-Host "Options:" -ForegroundColor Green
    Write-Host "  -Detached - Run in background (for 'up' action)"
    Write-Host "  -Service  - Target specific service (for logs/restart)"
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\docker-start.ps1 up -Detached"
    Write-Host "  .\docker-start.ps1 logs -Service api-server"
    Write-Host "  .\docker-start.ps1 restart"
    Write-Host ""
}

# Check if Docker is running
try {
    docker info | Out-Null
} catch {
    Write-Host "ERROR: Docker is not running!" -ForegroundColor Red
    Write-Host "Please start Docker Desktop and try again." -ForegroundColor Yellow
    exit 1
}

# Check if docker-compose.yml exists
if (-not (Test-Path "docker-compose.yml")) {
    Write-Host "ERROR: docker-compose.yml not found!" -ForegroundColor Red
    Write-Host "Please run this script from the mangahub directory." -ForegroundColor Yellow
    exit 1
}

switch ($Action) {
    'up' {
        Write-Host "Starting MangaHub services..." -ForegroundColor Green
        Write-Host ""
        Write-Host "Service startup order:" -ForegroundColor Cyan
        Write-Host "  1. TCP Server       (Port 9001, 9010)" -ForegroundColor White
        Write-Host "  2. UDP Server       (Port 9002, 9020)" -ForegroundColor White
        Write-Host "  3. gRPC Server      (Port 9003)" -ForegroundColor White
        Write-Host "  4. API Server       (Port 8080)" -ForegroundColor White
        Write-Host "  5. Web React        (Port 3000)" -ForegroundColor White
        Write-Host "  6. Fetch Manga Srv  (Port 8082)" -ForegroundColor White
        Write-Host ""
        
        if ($Detached) {
            Write-Host "Starting in detached mode..." -ForegroundColor Yellow
            docker-compose up -d --build
        } else {
            Write-Host "Starting in foreground mode (Ctrl+C to stop)..." -ForegroundColor Yellow
            Write-Host "Tip: Use -Detached flag to run in background" -ForegroundColor DarkGray
            Write-Host ""
            docker-compose up --build
        }
        
        if ($LASTEXITCODE -eq 0 -and $Detached) {
            Write-Host ""
            Write-Host "✓ All services started successfully!" -ForegroundColor Green
            Write-Host ""
            Write-Host "Access the application at:" -ForegroundColor Cyan
            Write-Host "  Frontend:     http://localhost:3000" -ForegroundColor White
            Write-Host "  API Server:   http://localhost:8080" -ForegroundColor White
            Write-Host "  Fetch Server: http://localhost:8082" -ForegroundColor White
            Write-Host ""
            Write-Host "Useful commands:" -ForegroundColor Yellow
            Write-Host "  .\docker-start.ps1 logs        - View logs"
            Write-Host "  .\docker-start.ps1 status      - Check status"
            Write-Host "  .\docker-start.ps1 down        - Stop services"
        }
    }
    
    'down' {
        Write-Host "Stopping MangaHub services..." -ForegroundColor Yellow
        docker-compose down
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host ""
            Write-Host "✓ All services stopped successfully!" -ForegroundColor Green
        }
    }
    
    'restart' {
        Write-Host "Restarting MangaHub services..." -ForegroundColor Yellow
        
        if ($Service) {
            Write-Host "Restarting service: $Service" -ForegroundColor Cyan
            docker-compose restart $Service
        } else {
            docker-compose restart
        }
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host ""
            Write-Host "✓ Services restarted successfully!" -ForegroundColor Green
        }
    }
    
    'logs' {
        if ($Service) {
            Write-Host "Viewing logs for: $Service" -ForegroundColor Cyan
            Write-Host "Press Ctrl+C to exit" -ForegroundColor DarkGray
            Write-Host ""
            docker-compose logs -f $Service
        } else {
            Write-Host "Viewing logs for all services" -ForegroundColor Cyan
            Write-Host "Press Ctrl+C to exit" -ForegroundColor DarkGray
            Write-Host ""
            docker-compose logs -f
        }
    }
    
    'build' {
        Write-Host "Rebuilding MangaHub containers..." -ForegroundColor Yellow
        docker-compose build --no-cache
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host ""
            Write-Host "✓ Containers rebuilt successfully!" -ForegroundColor Green
            Write-Host "Run '.\docker-start.ps1 up' to start the services" -ForegroundColor Cyan
        }
    }
    
    'status' {
        Write-Host "MangaHub Service Status:" -ForegroundColor Cyan
        Write-Host ""
        docker-compose ps
        
        Write-Host ""
        Write-Host "Resource Usage:" -ForegroundColor Cyan
        docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}"
    }
    
    'clean' {
        Write-Host "WARNING: This will stop all services and remove volumes!" -ForegroundColor Red
        Write-Host "All data will be lost!" -ForegroundColor Red
        Write-Host ""
        $confirm = Read-Host "Are you sure? (yes/no)"
        
        if ($confirm -eq "yes") {
            Write-Host "Cleaning up..." -ForegroundColor Yellow
            docker-compose down -v
            docker system prune -f
            
            if ($LASTEXITCODE -eq 0) {
                Write-Host ""
                Write-Host "✓ Cleanup completed!" -ForegroundColor Green
            }
        } else {
            Write-Host "Cleanup cancelled." -ForegroundColor Yellow
        }
    }
    
    default {
        Show-Help
    }
}

Write-Host ""
