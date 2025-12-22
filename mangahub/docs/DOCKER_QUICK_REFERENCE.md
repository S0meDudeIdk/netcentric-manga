# Docker Quick Reference

## Quick Commands

### Start Everything
```powershell
# From mangahub directory
.\docker-start.ps1 up -Detached

# Or using docker-compose directly
docker-compose up -d --build
```

### Stop Everything
```powershell
.\docker-start.ps1 down

# Or
docker-compose down
```

### View Logs
```powershell
# All services
.\docker-start.ps1 logs

# Specific service
.\docker-start.ps1 logs -Service api-server

# Or
docker-compose logs -f api-server
```

### Check Status
```powershell
.\docker-start.ps1 status

# Or
docker-compose ps
```

## Service URLs

| Service | URL | Description |
|---------|-----|-------------|
| Frontend | http://localhost:3000 | React web application |
| API Server | http://localhost:8080 | Main REST API |
| Fetch Manga | http://localhost:8082 | Manga fetching service |
| gRPC Server | localhost:9003 | gRPC service |
| TCP Server | localhost:9001 | Progress sync |
| TCP HTTP API | http://localhost:9010 | TCP trigger API |
| UDP Server | localhost:9002 | Notifications |
| UDP HTTP API | http://localhost:9020 | UDP trigger API |

## Troubleshooting

### Container Won't Start
```powershell
# Check logs
docker-compose logs <service-name>

# Rebuild specific service
docker-compose build --no-cache <service-name>
docker-compose up -d <service-name>
```

### Port Already in Use
Edit `docker-compose.yml` and change the external port:
```yaml
ports:
  - "8090:8080"  # Change 8090 to any available port
```

### Database Issues
```powershell
# Reset everything including volumes
docker-compose down -v
docker-compose up -d --build
```

### Network Issues
```powershell
# Recreate network
docker-compose down
docker network prune
docker-compose up -d
```

## Development Tips

### Restart Single Service
```powershell
docker-compose restart api-server
```

### Execute Commands in Container
```powershell
# Open shell in container
docker exec -it mangahub-api-server sh

# Run specific command
docker exec mangahub-api-server ls /app/data
```

### View Container Resource Usage
```powershell
docker stats
```

### Clean Up Unused Resources
```powershell
# Remove stopped containers, unused networks, images
docker system prune -a

# Also remove volumes (WARNING: deletes data!)
docker system prune -a --volumes
```

## Data Management

### Backup Database
```powershell
# Create backup
docker cp mangahub-api-server:/app/data/database.db ./backup-database.db

# Or backup entire volume
docker run --rm -v mangahub_manga-data:/data -v ${PWD}:/backup alpine tar czf /backup/manga-data-backup.tar.gz -C /data .
```

### Restore Database
```powershell
# Restore single file
docker cp ./backup-database.db mangahub-api-server:/app/data/database.db
docker-compose restart api-server

# Or restore entire volume
docker run --rm -v mangahub_manga-data:/data -v ${PWD}:/backup alpine sh -c "cd /data && tar xzf /backup/manga-data-backup.tar.gz"
```

## Service Dependencies

```
tcp-server (9001, 9010)
    ↓ waits for healthy
udp-server (9002, 9020)
    ↓ waits for healthy
grpc-server (9003)
    ↓ waits for healthy
api-server (8080)
    ↓ waits for healthy
web-react (3000)
    ↓ waits for healthy
fetch-manga-server (8082)
```

## Environment Variables

Key environment variables can be set in `.env` file:

```bash
# API Credentials
MAL_CLIENT_ID=your_client_id
MAL_CLIENT_SECRET=your_client_secret
MANGADEX_API_KEY=your_api_key

# Security
JWT_SECRET=your_secret_key

# Configuration
GIN_MODE=release
CORS_ALLOW_ORIGINS=*
```

## Production Checklist

- [ ] Change `JWT_SECRET` to a secure random value
- [ ] Set `GIN_MODE=release`
- [ ] Configure proper `CORS_ALLOW_ORIGINS` (not `*`)
- [ ] Add API credentials (MAL, MangaDex)
- [ ] Set up SSL/TLS with reverse proxy
- [ ] Configure backup schedule
- [ ] Set up monitoring and logging
- [ ] Review resource limits in docker-compose.yml
