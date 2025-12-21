# Check which processes are using the required ports
Write-Host "Checking ports used by MangaHub..." -ForegroundColor Cyan
Write-Host ""

$ports = @(8080, 9001, 9002, 9003, 9010, 9020, 3000)
$portNames = @{
    8080 = "API Server (Main)"
    9001 = "TCP Server"
    9002 = "UDP Server"
    9003 = "gRPC Server"
    9010 = "TCP HTTP Trigger"
    9020 = "UDP HTTP Trigger"
    3000 = "React Frontend"
}

foreach ($port in $ports) {
    $connections = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
    
    if ($connections) {
        Write-Host "Port $port ($($portNames[$port])):" -ForegroundColor Yellow
        foreach ($conn in $connections) {
            $process = Get-Process -Id $conn.OwningProcess -ErrorAction SilentlyContinue
            if ($process) {
                Write-Host "  ✗ IN USE by: $($process.Name) (PID: $($process.Id))" -ForegroundColor Red
            }
        }
    } else {
        Write-Host "Port $port ($($portNames[$port])): ✓ Available" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "To kill a process, run: Stop-Process -Id <PID> -Force" -ForegroundColor Cyan
