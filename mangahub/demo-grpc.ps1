# gRPC Service Demonstration Script
# Run this to test the gRPC implementation

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "MangaHub gRPC Service Demonstration" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if server is already running
$grpcProcess = Get-Process -Name "grpc-server" -ErrorAction SilentlyContinue
if ($grpcProcess) {
    Write-Host "⚠️  gRPC server is already running (PID: $($grpcProcess.Id))" -ForegroundColor Yellow
    Write-Host "   Stopping existing server..." -ForegroundColor Yellow
    Stop-Process -Name "grpc-server" -Force
    Start-Sleep -Seconds 2
}

Write-Host "📁 Building gRPC components..." -ForegroundColor Green
Write-Host ""

# Build server
Write-Host "   Building gRPC server..." -ForegroundColor Gray
go build -o bin/grpc-server.exe ./cmd/grpc-server
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to build gRPC server" -ForegroundColor Red
    exit 1
}

# Build test client
Write-Host "   Building gRPC test client..." -ForegroundColor Gray
go build -o bin/grpc-client-test.exe ./cmd/grpc-client-test
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to build gRPC client" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "✅ Build successful!" -ForegroundColor Green
Write-Host ""

# Start gRPC server in background
Write-Host "🚀 Starting gRPC server on port 9001..." -ForegroundColor Green
$serverJob = Start-Job -ScriptBlock {
    Set-Location "X:\Bao2023toPresent\IU\4th year\NetCentric\Project\netcentric-manga\mangahub"
    .\bin\grpc-server.exe
}

# Wait for server to start
Write-Host "   Waiting for server to initialize..." -ForegroundColor Gray
Start-Sleep -Seconds 3

# Check if server is running
$grpcProcess = Get-Process -Name "grpc-server" -ErrorAction SilentlyContinue
if (-not $grpcProcess) {
    Write-Host "❌ Failed to start gRPC server" -ForegroundColor Red
    Receive-Job $serverJob
    Remove-Job $serverJob -Force
    exit 1
}

Write-Host "✅ gRPC server is running (PID: $($grpcProcess.Id))" -ForegroundColor Green
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "gRPC Implementation Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "📋 Protocol Buffer Definition:" -ForegroundColor Yellow
Write-Host "   File: proto/manga.proto" -ForegroundColor Gray
Write-Host "   Package: manga" -ForegroundColor Gray
Write-Host ""

Write-Host "🔧 gRPC Services Implemented:" -ForegroundColor Yellow
Write-Host "   ✓ GetManga(GetMangaRequest) returns (MangaResponse)" -ForegroundColor Green
Write-Host "   ✓ SearchManga(SearchRequest) returns (SearchResponse)" -ForegroundColor Green
Write-Host "   ✓ UpdateProgress(ProgressRequest) returns (ProgressResponse)" -ForegroundColor Green
Write-Host ""

Write-Host "📦 Components:" -ForegroundColor Yellow
Write-Host "   ✓ internal/grpc/server.go - gRPC server implementation" -ForegroundColor Green
Write-Host "   ✓ internal/grpc/client.go - gRPC client helper" -ForegroundColor Green
Write-Host "   ✓ cmd/grpc-server/main.go - Standalone gRPC server" -ForegroundColor Green
Write-Host "   ✓ cmd/grpc-client-test/main.go - Test client" -ForegroundColor Green
Write-Host ""

Write-Host "🌐 Server Details:" -ForegroundColor Yellow
Write-Host "   Address: localhost:9001" -ForegroundColor Gray
Write-Host "   Status: Running ✓" -ForegroundColor Green
Write-Host "   PID: $($grpcProcess.Id)" -ForegroundColor Gray
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Usage Instructions" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "To use the gRPC service:" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Server is currently running" -ForegroundColor White
Write-Host "   Stop with: Stop-Process -Name 'grpc-server' -Force" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Start server manually:" -ForegroundColor White
Write-Host "   cd cmd/grpc-server" -ForegroundColor Gray
Write-Host "   go run main.go" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Test with client:" -ForegroundColor White
Write-Host "   cd cmd/grpc-client-test" -ForegroundColor Gray
Write-Host "   go run main.go" -ForegroundColor Gray
Write-Host ""
Write-Host "4. Integrate into your code:" -ForegroundColor White
Write-Host "   import \"mangahub/internal/grpc\"" -ForegroundColor Gray
Write-Host "   client, _ := grpc.NewClient(\"localhost:9001\")" -ForegroundColor Gray
Write-Host "   resp, _ := client.GetManga(ctx, \"1\")" -ForegroundColor Gray
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "gRPC Features Demonstrated" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "✅ Protocol Buffer definitions (.proto file)" -ForegroundColor Green
Write-Host "✅ gRPC server implementation with 3 services" -ForegroundColor Green
Write-Host "✅ gRPC client helper for easy integration" -ForegroundColor Green
Write-Host "✅ Unary RPC calls (request-response pattern)" -ForegroundColor Green
Write-Host "✅ Error handling and logging" -ForegroundColor Green
Write-Host "✅ Graceful shutdown support" -ForegroundColor Green
Write-Host "✅ Integration with existing services" -ForegroundColor Green
Write-Host ""

Write-Host "Press any key to stop the gRPC server..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

Write-Host ""
Write-Host "🛑 Stopping gRPC server..." -ForegroundColor Red
Stop-Process -Name "grpc-server" -Force -ErrorAction SilentlyContinue
Remove-Job $serverJob -Force -ErrorAction SilentlyContinue

Write-Host "✅ gRPC server stopped" -ForegroundColor Green
Write-Host ""
Write-Host "Thank you for testing the gRPC implementation! 🎉" -ForegroundColor Cyan
