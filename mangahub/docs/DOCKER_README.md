# Docker Management Scripts for MangaHub

## Quick Start

### Start all services (in order)
```powershell
docker-compose up -d
```

### View logs
```powershell
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f tcp-server
docker-compose logs -f udp-server
docker-compose logs -f grpc-server
docker-compose logs -f api-server
docker-compose logs -f web-react
docker-compose logs -f fetch-manga-server
```

### Stop all services
```powershell
docker-compose down
```

### Rebuild and restart
```powershell
docker-compose up -d --build
```

### Check service status
```powershell
docker-compose ps
```

## Service Startup Order

The services start in the following order with health checks:

1. **TCP Server** (port 9001, 9010) - Progress synchronization
2. **UDP Server** (port 9002, 9020) - Notifications
3. **gRPC Server** (port 9003) - Remote procedure calls
4. **API Server** (port 8080) - Main HTTP REST API
5. **Web React** (port 3000) - Frontend UI
6. **Fetch Manga Server** (port 8081) - Background manga fetching

## Port Mapping

- `3000` - React Web UI
- `8080` - API Server (REST API)
- `8081` - Fetch Manga Server
- `9001` - TCP Server
- `9010` - TCP HTTP Trigger API
- `9002` - UDP Server
- `9020` - UDP HTTP Trigger API
- `9003` - gRPC Server

## Environment Variables

Copy `.env.example` to `.env` and configure:

```powershell
Copy-Item .env.example .env
# Edit .env with your configurations
```

Key variables:
- `JWT_SECRET` - JWT authentication secret
- `MAL_CLIENT_ID` / `MAL_CLIENT_SECRET` - MyAnimeList API credentials
- `GIN_MODE` - Set to `release` for production
- `CORS_ALLOW_ORIGINS` - Allowed CORS origins

## Useful Commands

### Clean up everything (including volumes)
```powershell
docker-compose down -v
```

### Rebuild specific service
```powershell
docker-compose up -d --build tcp-server
```

### View resource usage
```powershell
docker stats
```

### Access container shell
```powershell
docker exec -it mangahub-api-server sh
```

### View network
```powershell
docker network inspect mangahub_mangahub-network
```

## Troubleshooting

### Check if ports are in use
```powershell
netstat -ano | findstr "3000 8080 8081 9001 9002 9003 9010 9020"
```

### View service health status
```powershell
docker inspect --format='{{json .State.Health}}' mangahub-tcp-server
docker inspect --format='{{json .State.Health}}' mangahub-udp-server
docker inspect --format='{{json .State.Health}}' mangahub-grpc-server
docker inspect --format='{{json .State.Health}}' mangahub-api-server
```

### Restart specific service
```powershell
docker-compose restart api-server
```

### Remove all stopped containers
```powershell
docker container prune
```

### Remove unused images
```powershell
docker image prune -a
```

## Development Workflow

### Hot reload for React (development mode)
For development with hot reload, modify the React service in docker-compose.yml:

```yaml
web-react:
  command: npm start
  volumes:
    - ./client/web-react/src:/app/src
    - ./client/web-react/public:/app/public
```

### View real-time logs during development
```powershell
docker-compose logs -f api-server grpc-server
```

## Production Deployment

1. Set environment variables:
```powershell
$env:GIN_MODE="release"
$env:JWT_SECRET="your-secure-secret-here"
```

2. Build and start:
```powershell
docker-compose up -d --build
```

3. Verify all services are healthy:
```powershell
docker-compose ps
```

## Backup and Restore Database

### Backup
```powershell
docker cp mangahub-api-server:/data/mangahub.db ./backup/mangahub-$(Get-Date -Format "yyyyMMdd-HHmmss").db
```

### Restore
```powershell
docker cp ./backup/mangahub-backup.db mangahub-api-server:/data/mangahub.db
docker-compose restart api-server grpc-server fetch-manga-server
```
