# MangaHub Scripts

This directory contains all PowerShell scripts for managing and testing the MangaHub application.

## Structure

All `.ps1` files in the **root `mangahub/` directory** are now **lightweight wrappers** that call the actual scripts in `scripts/`.

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
```

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
