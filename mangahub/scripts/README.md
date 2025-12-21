# MangaHub Scripts

This directory contains all PowerShell scripts for managing and testing the MangaHub application.

## Quick Start

### 🚀 Start All Servers (Recommended)
```powershell
.\scripts\start-all-servers.ps1
```
This starts all required servers in the correct order:
1. TCP Server (port 9001)
2. UDP Server (port 9002)
3. gRPC Server (port 9003)
4. API Server (port 8080)
5. React Frontend (port 3000)

### 🛑 Stop All Servers
```powershell
.\scripts\stop-all-servers.ps1
```

### 🔍 Check Port Status
```powershell
.\scripts\check-ports.ps1
```

## Server Architecture

MangaHub uses a **microservices architecture** with multiple servers:

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────┐
│  React Frontend │────▶│  API Server  │────▶│   Database  │
│    Port 3000    │     │  Port 8080   │     │   SQLite    │
└─────────────────┘     └──────┬───────┘     └─────────────┘
                               │
                ┌──────────────┼──────────────┐
                │              │              │
         ┌──────▼─────┐ ┌─────▼──────┐ ┌────▼─────┐
         │ gRPC Server│ │ TCP Server │ │UDP Server│
         │ Port 9003  │ │ Port 9001  │ │Port 9002 │
         └────────────┘ └────────────┘ └──────────┘
```

### Why Multiple Servers?

- **API Server (8080)**: Main HTTP REST API, handles most client requests
- **gRPC Server (9003)**: High-performance service for library management, progress tracking, and ratings
- **TCP Server (9001)**: Real-time progress broadcasting to connected clients
- **UDP Server (9002)**: Lightweight notifications for manga updates

### Port Reference

| Server | Port | Protocol | Purpose |
|--------|------|----------|---------|
| Frontend | 3000 | HTTP | React web interface |
| API Server | 8080 | HTTP/WS | Main REST API + WebSocket chat |
| TCP Server | 9001 | TCP | Progress broadcasts |
| TCP HTTP Trigger | 9010 | HTTP | Trigger TCP broadcasts |
| UDP Server | 9002 | UDP | Notification broadcasts |
| UDP HTTP Trigger | 9020 | HTTP | Trigger UDP broadcasts |
| gRPC Server | 9003 | gRPC | High-performance service calls |

## Script Reference

### Main Scripts

- **`start-server.ps1`** - Start the MangaHub API server (automatically starts TCP server if not running)
- **`build.ps1`** - Build all components and place binaries in `bin/`
- **`demo-cli.ps1`** - Demo the CLI client improvements
- **`demo-grpc.ps1`** - Demo gRPC service functionality
- **`register-test-users.ps1`** - Register test users for WebSocket chat testing
- **`test-mal-api.ps1`** - Test MyAnimeList API integration
- **`test-tcp-udp-integration.ps1`** - Test TCP/UDP integration

## Usage

From the `mangahub/` root directory, simply run:

```powershell
# Start the server (TCP + API)
./start-server.ps1

# Build all binaries
./build.ps1

# Run demos
./demo-cli.ps1
./demo-grpc.ps1

# Test functionality
./register-test-users.ps1
./test-mal-api.ps1
./test-tcp-udp-integration.ps1

# Update data
./update-total-chapters.ps1  # Fix manga with 0 chapters
./update-publication-years.ps1  # Update publication years from MAL
```

## Maintenance Scripts

### Update Total Chapters
```powershell
.\scripts\update-total-chapters.ps1
```
- Finds all manga with `total_chapters = 0` or `NULL`
- Sets them to a default value of 1
- Logs all updates
- For accurate counts, use the sync feature or manually update via API

**When to use:**
- After importing manga from external sources
- When chapter counts are missing or incorrect
- After database migrations

## Changes Made

### Before
- Scripts scattered in `mangahub/` root
- `.exe` files in both root and `bin/`
- Manual TCP server startup required

### After
- All actual scripts organized in `scripts/`
- All binaries consolidated in `bin/`
- Root scripts are thin wrappers for easy access
- `start-server.ps1` automatically starts TCP server if needed

## Benefits

1. **Cleaner root directory** - Only wrappers and essential files
2. **Automatic TCP server** - No more connection errors
3. **Centralized scripts** - Easy to find and maintain
4. **Binary management** - All executables in `bin/`
5. **Build script** - One command to build everything
