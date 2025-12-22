# MangaHub Docker Setup

This Docker configuration provides a complete multi-service containerized setup for the MangaHub application.

## Architecture

The setup includes 6 services that start in the following order:

1. **TCP Server** (Port 9001, 9010) - Progress synchronization server
2. **UDP Server** (Port 9002, 9020) - Notification broadcast server
3. **gRPC Server** (Port 9003) - High-performance RPC service
4. **API Server** (Port 8080) - Main REST API backend
5. **Web React** (Port 3000) - Frontend web application
6. **Fetch Manga Server** (Port 8082) - Manga data fetching service

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- 4GB+ RAM recommended

## Quick Start

### 1. Environment Configuration

Create a `.env` file in the mangahub directory (or use the existing one):

```bash
# The .env file should already exist
# Make sure to update these values if needed:
# MAL_CLIENT_ID=your_mal_client_id
# MAL_CLIENT_SECRET=your_mal_client_secret
# MANGADEX_API_KEY=your_mangadex_key (optional)
```

### 2. Build and Start All Services

```bash
# From the mangahub directory
docker-compose up --build
```

Or start in detached mode:

```bash
docker-compose up -d --build
```

### 3. Access the Application

- **Frontend**: http://localhost:3000
- **API Server**: http://localhost:8080
- **Fetch Manga Server**: http://localhost:8082
- **gRPC Server**: localhost:9003
- **TCP Server**: localhost:9001 (HTTP API: 9010)
- **UDP Server**: localhost:9002 (HTTP API: 9020)

## Service Management

### Start Services
```bash
docker-compose up
```

### Stop Services
```bash
docker-compose down
```

### Stop and Remove Volumes
```bash
docker-compose down -v
```

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api-server
docker-compose logs -f web-react
```

### Restart a Service
```bash
docker-compose restart api-server
```

### Scale Services (if needed)
```bash
docker-compose up --scale fetch-manga-server=2
```

## Service Dependencies

The services use health checks and `depends_on` conditions to ensure proper startup order:

```
tcp-server (healthy)
    ↓
udp-server (healthy)
    ↓
grpc-server (healthy)
    ↓
api-server (healthy)
    ↓
web-react (healthy)
    ↓
fetch-manga-server
```

## Volume Management

### Persistent Data
The `manga-data` volume stores:
- SQLite database (`database.db`)
- Downloaded manga data
- User preferences
- Chapter information

### Backup Data
```bash
# Create backup
docker run --rm -v mangahub_manga-data:/data -v $(pwd):/backup alpine tar czf /backup/manga-data-backup.tar.gz -C /data .

# Restore backup
docker run --rm -v mangahub_manga-data:/data -v $(pwd):/backup alpine sh -c "cd /data && tar xzf /backup/manga-data-backup.tar.gz"
```

## Network Configuration

All services run on the `mangahub-network` bridge network, allowing internal communication using service names:

- `tcp-server:9001`
- `udp-server:9002`
- `grpc-server:9003`
- `api-server:8080`
- `web-react:80`
- `fetch-manga-server:8082`

## Database

The application uses **SQLite** as the database, which is stored in the `manga-data` volume at `/app/data/database.db`. This makes the setup simple and portable without requiring a separate database server.

## Troubleshooting

### Service Won't Start
```bash
# Check logs
docker-compose logs <service-name>

# Check health status
docker-compose ps
```

### Port Conflicts
If ports are already in use, modify the port mappings in `docker-compose.yml`:
```yaml
ports:
  - "8090:8080"  # Change 8090 to any available port
```

### Database Issues
```bash
# Reset database
docker-compose down -v
docker-compose up --build
```

### Rebuild Specific Service
```bash
docker-compose build --no-cache api-server
docker-compose up -d api-server
```

## Development Mode

For development with hot reload:

1. Use volume mounts to sync source code
2. Set `GIN_MODE=debug` in environment
3. Use `docker-compose -f docker-compose.dev.yml up`

## Production Deployment

For production:

1. Set `GIN_MODE=release`
2. Update `JWT_SECRET` with a secure random value
3. Configure proper CORS origins (don't use `*`)
4. Set up reverse proxy (nginx/traefik) for SSL
5. Use Docker secrets for sensitive data

## Environment Variables

KeyFETCH_MANGA_PORT`: Port for fetch manga server (default: 8082)
- `GIN_MODE`: `debug` or `release`
- `CORS_ALLOW_ORIGINS`: Allowed CORS originsWT tokens
- `MAL_CLIENT_ID`, `MAL_CLIENT_SECRET`: MyAnimeList API credentials
- `MANGADEX_API_KEY`: MangaDex API key (optional)
- `GIN_MODE`: `debug` or `release`
- `CORS_ALLOW_ORIGINS`: Allowed CORS origins
- `DB_*`: Database configuration

## Health Checks

Each service implements health checks:

- **TCP/UDP/API Servers**: HTTP endpoint `/health`
- **gRPC Server**: Port availability check
- **Web React**: Nginx health check

Health check interval: 10s
Timeout: 5s
Retries: 3-5 (varies by service)

## Updating Services

### Pull Latest Code
```bash
git pull
docker-compose down
docker-compose up --build
```

### Update Single Service
```bash
docker-compose build --no-cache api-server
docker-compose up -d api-server
```

## Monitoring

### View Resource Usage
```bash
docker stats
```

### Check Service Health
```bash
docker-compose ps
```

## Support

For issues or questions:
1. Check service logs: `docker-compose logs -f`
2. Verify environment configuration
3. Ensure all ports are available
4. Check Docker daemon status

## License

See LICENSE file in the root directory.
