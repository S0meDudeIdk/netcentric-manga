# Docker Compose Management Script for MangaHub
# This script helps manage the multi-service Docker environment

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet('up', 'down', 'restart', 'logs', 'status', 'build', 'clean')]
    [string]$Command = 'status',
    
    [Parameter(Mandatory=$false)]
    [string]$Service = ''
)

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = 'White'
    )
    Write-Host $Message -ForegroundColor $Color
}

function Show-Banner {
    Write-ColorOutput "`n╔════════════════════════════════════════╗" "Cyan"
    Write-ColorOutput "║     MangaHub Docker Management         ║" "Cyan"
    Write-ColorOutput "╚════════════════════════════════════════╝`n" "Cyan"
}

function Start-Services {
    Write-ColorOutput "Starting MangaHub services..." "Green"
    Write-ColorOutput "Services will start in order:" "Yellow"
    Write-ColorOutput "  1. TCP Server (port 9001, 9010)" "Gray"
    Write-ColorOutput "  2. UDP Server (port 9002, 9020)" "Gray"
    Write-ColorOutput "  3. gRPC Server (port 9003)" "Gray"
    Write-ColorOutput "  4. API Server (port 8080)" "Gray"
    Write-ColorOutput "  5. Web React (port 3000)" "Gray"
    Write-ColorOutput "  6. Fetch Manga Server (port 8081)`n" "Gray"
    
    docker-compose up -d
    
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "`n✓ All services started successfully!" "Green"
        Start-Sleep -Seconds 3
        Show-Status
    } else {
        Write-ColorOutput "`n✗ Failed to start services" "Red"
    }
}

function Stop-Services {
    Write-ColorOutput "Stopping MangaHub services..." "Yellow"
    docker-compose down
    
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "✓ All services stopped successfully!" "Green"
    } else {
        Write-ColorOutput "✗ Failed to stop services" "Red"
    }
}

function Restart-Services {
    if ($Service) {
        Write-ColorOutput "Restarting service: $Service..." "Yellow"
        docker-compose restart $Service
    } else {
        Write-ColorOutput "Restarting all services..." "Yellow"
        docker-compose restart
    }
    
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "✓ Services restarted successfully!" "Green"
    } else {
        Write-ColorOutput "✗ Failed to restart services" "Red"
    }
}

function Show-Logs {
    if ($Service) {
        Write-ColorOutput "Showing logs for: $Service" "Cyan"
        docker-compose logs -f $Service
    } else {
        Write-ColorOutput "Showing logs for all services (Ctrl+C to exit)" "Cyan"
        docker-compose logs -f
    }
}

function Show-Status {
    Write-ColorOutput "Service Status:" "Cyan"
    docker-compose ps
    
    Write-ColorOutput "`nService Health:" "Cyan"
    $services = @('tcp-server', 'udp-server', 'grpc-server', 'api-server', 'fetch-manga-server')
    
    foreach ($svc in $services) {
        $containerName = "mangahub-$svc"
        $health = docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}N/A{{end}}' $containerName 2>$null
        
        if ($health) {
            $color = switch ($health) {
                'healthy' { 'Green' }
                'unhealthy' { 'Red' }
                'starting' { 'Yellow' }
                default { 'Gray' }
            }
            Write-ColorOutput "  $svc : $health" $color
        }
    }
    
    Write-ColorOutput "`nAccess Points:" "Cyan"
    Write-ColorOutput "  Web UI        : http://localhost:3000" "White"
    Write-ColorOutput "  API Server    : http://localhost:8080" "White"
    Write-ColorOutput "  Fetch Server  : http://localhost:8081" "White"
    Write-ColorOutput "  TCP Server    : localhost:9001" "White"
    Write-ColorOutput "  UDP Server    : localhost:9002" "White"
    Write-ColorOutput "  gRPC Server   : localhost:9003`n" "White"
}

function Build-Services {
    Write-ColorOutput "Building services..." "Yellow"
    docker-compose build --no-cache
    
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "✓ Build completed successfully!" "Green"
    } else {
        Write-ColorOutput "✗ Build failed" "Red"
    }
}

function Clean-Environment {
    Write-ColorOutput "Cleaning Docker environment..." "Yellow"
    Write-ColorOutput "This will remove:" "Red"
    Write-ColorOutput "  - All stopped containers" "Red"
    Write-ColorOutput "  - All unused networks" "Red"
    Write-ColorOutput "  - All dangling images" "Red"
    
    $confirm = Read-Host "Continue? (y/N)"
    if ($confirm -eq 'y' -or $confirm -eq 'Y') {
        docker-compose down -v
        docker system prune -f
        Write-ColorOutput "✓ Cleanup completed!" "Green"
    } else {
        Write-ColorOutput "Cleanup cancelled" "Yellow"
    }
}

# Main execution
Show-Banner

# Check if Docker is running
try {
    docker info > $null 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-ColorOutput "✗ Docker is not running. Please start Docker Desktop." "Red"
        exit 1
    }
} catch {
    Write-ColorOutput "✗ Docker is not running. Please start Docker Desktop." "Red"
    exit 1
}

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-ColorOutput "⚠ Warning: .env file not found!" "Yellow"
    Write-ColorOutput "Creating .env from .env.example..." "Yellow"
    
    if (Test-Path ".env.example") {
        Copy-Item ".env.example" ".env"
        Write-ColorOutput "✓ Created .env file. Please review and update it." "Green"
    } else {
        Write-ColorOutput "✗ .env.example not found!" "Red"
        exit 1
    }
}

# Execute command
switch ($Command) {
    'up' { Start-Services }
    'down' { Stop-Services }
    'restart' { Restart-Services }
    'logs' { Show-Logs }
    'status' { Show-Status }
    'build' { Build-Services }
    'clean' { Clean-Environment }
    default { 
        Write-ColorOutput "Unknown command: $Command" "Red"
        Write-ColorOutput "Available commands: up, down, restart, logs, status, build, clean" "Yellow"
    }
}
