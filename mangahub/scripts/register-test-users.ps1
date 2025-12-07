# Register Test Users for WebSocket Chat Testing
Write-Host "`n=== REGISTERING TEST USERS ===" -ForegroundColor Cyan

$apiUrl = "http://localhost:8080/api/v1/auth/register"

$users = @(
    @{
        username = "testuser1"
        email = "testuser1@example.com"
        password = "password123"
    },
    @{
        username = "testuser2"
        email = "testuser2@example.com"
        password = "password123"
    }
)

foreach ($user in $users) {
    Write-Host "`nRegistering: $($user.username) ($($user.email))..." -ForegroundColor Yellow
    
    $body = @{
        username = $user.username
        email = $user.email
        password = $user.password
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri $apiUrl -Method POST -Body $body -ContentType "application/json" -ErrorAction Stop
        Write-Host "✅ Success! User registered." -ForegroundColor Green
    } catch {
        if ($_.Exception.Response.StatusCode.value__ -eq 409) {
            Write-Host "✅ User already exists (can login now)" -ForegroundColor Green
        } else {
            Write-Host "❌ Failed: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
}

Write-Host "`n=== TEST USERS READY ===" -ForegroundColor Green
Write-Host "`nYou can now login with:" -ForegroundColor Cyan
Write-Host "  Email: testuser1@example.com" -ForegroundColor White
Write-Host "  Password: password123" -ForegroundColor White
Write-Host "`n  Email: testuser2@example.com" -ForegroundColor White
Write-Host "  Password: password123" -ForegroundColor White
Write-Host "`nRefresh your browser and try logging in! 🚀" -ForegroundColor Yellow
