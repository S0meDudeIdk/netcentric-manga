# Stop all MangaHub server processes
Write-Host "Stopping all MangaHub servers..." -ForegroundColor Cyan

$ports = @(8080, 9001, 9002, 9003, 9010, 9020)
$stopped = 0

foreach ($port in $ports) {
    $connections = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
    
    if ($connections) {
        foreach ($conn in $connections) {
            $process = Get-Process -Id $conn.OwningProcess -ErrorAction SilentlyContinue
            if ($process) {
                Write-Host "Stopping $($process.Name) (PID: $($process.Id)) on port $port..." -ForegroundColor Yellow
                Stop-Process -Id $process.Id -Force
                $stopped++
            }
        }
    }
}

if ($stopped -eq 0) {
    Write-Host "No running servers found." -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "Stopped $stopped process(es)." -ForegroundColor Green
}

# Also kill any remaining go.exe processes related to the project
$goProcesses = Get-Process -Name "go" -ErrorAction SilentlyContinue | Where-Object {
    $_.Path -like "*NetCentric*"
}

if ($goProcesses) {
    Write-Host ""
    Write-Host "Stopping remaining Go processes..." -ForegroundColor Yellow
    $goProcesses | ForEach-Object {
        Stop-Process -Id $_.Id -Force
        Write-Host "  Stopped PID: $($_.Id)" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "Done! All servers stopped." -ForegroundColor Green
