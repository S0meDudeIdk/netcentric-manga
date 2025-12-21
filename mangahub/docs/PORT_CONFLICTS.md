# Troubleshooting Port Conflicts

## Problem: "bind: Only one usage of each socket address is normally permitted"

This error occurs when you try to start a server on a port that's already in use.

### Quick Fix

**Option 1: Stop All Servers and Restart**
```powershell
# From mangahub directory
.\scripts\stop-all-servers.ps1
.\scripts\start-all-servers.ps1
```

**Option 2: Check Which Process is Using the Port**
```powershell
.\scripts\check-ports.ps1
```

**Option 3: Manually Kill Process**
```powershell
# Find process using port 9003 (example)
Get-NetTCPConnection -LocalPort 9003 | Select-Object OwningProcess

# Kill the process
Stop-Process -Id <PID> -Force
```

## Common Scenarios

### Scenario 1: Servers Already Running
**Problem**: You tried to start a server that's already running from a previous session.

**Solution**: 
```powershell
.\scripts\stop-all-servers.ps1
.\scripts\start-all-servers.ps1
```

### Scenario 2: Starting Individual Servers
**Problem**: You started the TCP/UDP/gRPC servers individually, but they're already started by the main API server.

**Important**: The API server **connects to** (not starts) these servers. They must be running separately.

**Correct order**:
1. Start TCP Server: `cd cmd\tcp-server && go run main.go`
2. Start UDP Server: `cd cmd\udp-server && go run main.go`
3. Start gRPC Server: `cd cmd\grpc-server && go run main.go`
4. Start API Server: `cd cmd\api-server && go run main.go`

**Better**: Use `.\scripts\start-all-servers.ps1` to start everything in the correct order.

### Scenario 3: Previous Go Processes Didn't Exit Cleanly
**Problem**: You closed terminal windows without properly stopping servers, leaving orphaned Go processes.

**Solution**:
```powershell
# Find all Go processes
Get-Process -Name "go" | Where-Object { $_.Path -like "*NetCentric*" }

# Kill them
Get-Process -Name "go" | Where-Object { $_.Path -like "*NetCentric*" } | Stop-Process -Force

# Or use the stop script
.\scripts\stop-all-servers.ps1
```

### Scenario 4: Another Application Using the Port
**Problem**: A different application (not MangaHub) is using one of the required ports.

**Solution**:
```powershell
# Check what's using the port
netstat -ano | findstr :9003

# Find the process
Get-Process -Id <PID>

# If it's not MangaHub, either:
# 1. Stop that application
# 2. Change MangaHub's port configuration
```

## Port Configuration

If you need to change the default ports, update these environment variables in `.env`:

```env
# Main API Server
PORT=8080

# gRPC Server (in cmd/grpc-server)
GRPC_PORT=9003

# TCP Server
TCP_PORT=9001
TCP_HTTP_PORT=9010

# UDP Server
UDP_PORT=9002
UDP_HTTP_PORT=9020
```

**Note**: After changing ports, you'll also need to update the connection addresses:
- API Server connects to gRPC at `GRPC_SERVER_ADDR` (default: localhost:9003)
- API Server triggers TCP at `TCP_SERVER_ADDR` (default: http://localhost:9010)
- API Server triggers UDP at `UDP_SERVER_ADDR` (default: http://localhost:9020)

## Prevention Tips

### 1. Always Use the Startup Script
```powershell
.\scripts\start-all-servers.ps1
```
This handles starting servers in the correct order and checking for conflicts.

### 2. Always Stop Properly
Don't just close terminal windows. Use:
```powershell
.\scripts\stop-all-servers.ps1
```
Or press `Ctrl+C` in each terminal.

### 3. Check Ports Before Starting
```powershell
.\scripts\check-ports.ps1
```

### 4. Use PowerShell Jobs (Advanced)
The `start-all-servers.ps1` script uses PowerShell jobs to manage all servers in one terminal:
```powershell
# View server logs
Get-Job | Receive-Job -Keep

# Stop a specific server
Stop-Job -Name "TCP Server"

# Stop all servers
Get-Job | Stop-Job
Get-Job | Remove-Job
```

## Understanding the Error

The error message:
```
listen tcp :9003: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted.
```

Means:
- Windows only allows **one process** to listen on a specific port
- Another process is already using port 9003
- The new process cannot bind to the same port
- You must stop the existing process first

## Network Diagram

```
Your Computer
├── Port 3000: React Frontend ──┐
├── Port 8080: API Server ──────┼─▶ Connected via HTTP
├── Port 9001: TCP Server ◀─────┤
├── Port 9002: UDP Server ◀─────┤
├── Port 9003: gRPC Server ◀────┤
├── Port 9010: TCP HTTP Trigger ◀┤
└── Port 9020: UDP HTTP Trigger ◀┘
```

Each port can only have **one listener**. If you try to start two TCP servers on port 9001, the second one will fail.
