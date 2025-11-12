# netcentric-manga

A comprehensive manga reading platform with MyAnimeList integration, featuring both web and CLI clients.

## Features

### 🌐 MyAnimeList Integration (NEW!)
- Browse real manga data from MyAnimeList
- Search MAL's extensive database
- Toggle between MAL and local data sources
- Available in both web and CLI clients
- No authentication required (public access)

See [MAL Integration Documentation](mangahub/docs/MAL_INTEGRATION.md) for details.

### 📚 Core Features
- Browse and search manga
- User library management
- Reading progress tracking
- Personalized recommendations
- Multi-protocol support (HTTP, gRPC, TCP, UDP)

### 🔓 Freemium Model
- **Public Access**: Browse, search, and view manga without login
- **Authenticated Features**: Library, reading progress, recommendations

## Quick Start

### Configuration (Optional)
The application comes with sensible defaults. To customize settings:
```bash
cd mangahub
# The .env file is already created with defaults
# Edit it to customize port, CORS, rate limits, etc.
```
See [Environment Variables Guide](mangahub/ENV_CONFIGURATION.md) for all options.

### Start the API Server
```bash
cd mangahub/cmd/api-server
go run main.go
```

### Web Client
```bash
cd mangahub/client/web-react
npm install
npm start
```
Visit http://localhost:3000

### CLI Client
```bash
cd mangahub/client/cli
go run main.go
```

## Documentation

- [Quick Start Guide](mangahub/QUICKSTART.md)
- [Environment Variables](mangahub/ENV_CONFIGURATION.md) ⚙️ NEW!
- [MAL Integration Guide](mangahub/docs/MAL_INTEGRATION.md)
- [MAL Implementation Summary](mangahub/MAL_IMPLEMENTATION_SUMMARY.md)
- [API Documentation](mangahub/docs/API_DOCUMENTATION.md)
- [Access Model](mangahub/client/web-react/ACCESS_MODEL.md)

## Testing

Test the MyAnimeList integration:
```powershell
cd mangahub
.\test-mal-api.ps1
```

## Technology Stack

- **Backend**: Go (Gin, gRPC)
- **Frontend**: React, Tailwind CSS, Framer Motion
- **External API**: Jikan API v4 (MyAnimeList)
- **Database**: PostgreSQL
- **Protocols**: HTTP/REST, gRPC, TCP, UDP, WebSocket
