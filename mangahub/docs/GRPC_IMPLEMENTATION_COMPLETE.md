# gRPC Implementation Guide

## Overview

This document describes the implementation of three gRPC use cases (UC-014, UC-015, UC-016) for the MangaHub application. The implementation provides gRPC-backed endpoints that integrate with the existing HTTP API, database, and TCP broadcasting system.

## Architecture

```
┌─────────────┐      HTTP/REST      ┌─────────────┐      gRPC       ┌─────────────┐
│   Frontend  │ ◄──────────────────► │  API Server │ ◄──────────────► │ gRPC Server │
│  (React)    │                      │  (Port 8080)│                  │ (Port 9001) │
└─────────────┘                      └──────┬──────┘                  └──────┬──────┘
                                             │                                │
                                             │                                │
                                     ┌───────▼────────┐              ┌───────▼────────┐
                                     │  TCP Server    │              │   Database     │
                                     │  (Port 9000)   │              │   (SQLite)     │
                                     └────────────────┘              └────────────────┘
```

## Use Cases Implemented

### UC-014: Retrieve Manga via gRPC

**Primary Actor:** Internal Service  
**Goal:** Fetch manga data through gRPC interface

**Implementation:**

1. **Proto Definition** (`proto/manga.proto`):
   ```protobuf
   rpc GetManga(GetMangaRequest) returns (MangaResponse);
   
   message GetMangaRequest {
     string id = 1;
   }
   
   message MangaResponse {
     Manga manga = 1;
     string error = 2;
   }
   ```

2. **gRPC Server** (`internal/grpc/server.go`):
   - Implements `GetManga` RPC method
   - Queries database through manga service
   - Converts models.Manga to protobuf format
   - Returns manga data or error

3. **HTTP Endpoint** (`/api/v1/grpc/manga/:id`):
   - Protected route (requires authentication)
   - Calls gRPC client with 5-second timeout
   - Returns JSON response with source indicator

4. **Frontend Service** (`services/grpcService.js`):
   ```javascript
   grpcService.getManga(mangaId)
   ```

**Testing:**
1. Navigate to `/grpc-test` page
2. Enter manga ID (e.g., "1")
3. Click "Get Manga" button
4. View manga details in JSON format

---

### UC-015: Search Manga via gRPC

**Primary Actor:** Internal Service  
**Goal:** Search manga using gRPC interface with pagination

**Implementation:**

1. **Proto Definition**:
   ```protobuf
   rpc SearchManga(SearchRequest) returns (SearchResponse);
   
   message SearchRequest {
     string query = 1;
     int32 limit = 2;
     int32 offset = 3;
   }
   
   message SearchResponse {
     repeated Manga manga = 1;
     int32 total = 2;
     string error = 3;
   }
   ```

2. **gRPC Server**:
   - Implements `SearchManga` RPC method
   - Executes database query with filters
   - Supports pagination (limit/offset)
   - Returns array of manga results

3. **HTTP Endpoint** (`/api/v1/grpc/manga/search`):
   - Query parameters: `q` (query), `limit` (default: 20)
   - Protected route
   - 10-second timeout for complex searches

4. **Frontend Service**:
   ```javascript
   grpcService.searchManga(query, limit)
   ```

**Testing:**
1. Enter search query (e.g., "One Piece")
2. Set result limit (1-100)
3. Click "Search Manga" button
4. View paginated results

---

### UC-016: Update Progress via gRPC

**Primary Actor:** Internal Service  
**Goal:** Update user reading progress through gRPC with TCP broadcast

**Implementation:**

1. **Proto Definition**:
   ```protobuf
   rpc UpdateProgress(ProgressRequest) returns (ProgressResponse);
   
   message ProgressRequest {
     string user_id = 1;
     string manga_id = 2;
     int32 current_chapter = 3;
     string status = 4;
   }
   
   message ProgressResponse {
     bool success = 1;
     string message = 2;
     string error = 3;
   }
   ```

2. **gRPC Server with TCP Integration**:
   - Updates user progress in database
   - **Triggers TCP broadcast** for real-time sync
   - Broadcasts to all connected TCP clients
   - Gracefully handles TCP connection failures

3. **TCP Broadcasting**:
   ```go
   // In grpc/server.go
   func (s *Server) broadcastProgress(userID, mangaID string, chapter int) {
       update := tcp.ProgressUpdate{
           UserID: userID,
           MangaID: mangaID,
           Chapter: chapter,
           Timestamp: time.Now().Unix(),
       }
       // Send to TCP server
   }
   ```

4. **HTTP Endpoint** (`/api/v1/grpc/progress/update`):
   - Body: `{ manga_id, current_chapter, status }`
   - User ID extracted from JWT token
   - Returns success/error response

5. **Frontend Service**:
   ```javascript
   grpcService.updateProgress(mangaId, currentChapter, status)
   ```

**Testing:**
1. Enter manga ID from library
2. Set current chapter number
3. Select reading status
4. Click "Update Progress"
5. Verify database update
6. Check TCP broadcast in TCP server logs

---

## Setup Instructions

### 1. Prerequisites

- Go 1.21+
- Node.js 16+
- SQLite3
- Protocol Buffers compiler (protoc)

### 2. Database Setup

Ensure the database is initialized with manga and user data:

```powershell
cd mangahub/cmd/api-server
go run main.go
# Database will be created automatically if it doesn't exist
```

### 3. Start All Servers

Use the automated script:

```powershell
cd mangahub/scripts
.\start-grpc-test-env.ps1
```

This will start:
1. TCP Server on port 9000
2. gRPC Server on port 9001 (connected to TCP)
3. API Server on port 8080 (connected to gRPC)

### 4. Start Frontend

```powershell
cd mangahub/client/web-react
npm install
npm start
```

Frontend will run on http://localhost:3000

### 5. Access Test Page

Navigate to: http://localhost:3000/grpc-test

## Environment Variables

Required environment variables (`.env` file):

```env
# Server Ports
PORT=8080
TCP_SERVER_PORT=9000
GRPC_SERVER_PORT=9001

# Server Addresses
TCP_SERVER_ADDRESS=localhost:9000
GRPC_SERVER_ADDR=localhost:9001

# Database
DB_PATH=./manga.db

# JWT Secret
JWT_SECRET=your-secret-key

# CORS
CORS_ALLOW_ORIGINS=http://localhost:3000
```

## Testing Scenarios

### Scenario 1: Retrieve Specific Manga

1. Start all servers
2. Login to the application
3. Navigate to `/grpc-test`
4. Test UC-014:
   - Enter manga ID: `1`
   - Click "Get Manga"
   - Verify response contains manga details
   - Check "source": "grpc" field

### Scenario 2: Search Functionality

1. Test UC-015:
   - Enter query: "naruto"
   - Set limit: 10
   - Click "Search Manga"
   - Verify results array
   - Check total count

### Scenario 3: Progress Update with TCP Broadcast

1. Add manga to library first (via regular UI)
2. Test UC-016:
   - Enter manga ID from library
   - Set chapter: 5
   - Select status: "reading"
   - Click "Update Progress"
   - Check response: `success: true`
   - Verify TCP server logs show broadcast
   - Check database for updated progress

### Scenario 4: Error Handling

Test error cases:
- Invalid manga ID in UC-014
- Empty search query in UC-015
- Non-existent manga in UC-016
- gRPC server offline

## API Endpoints

### Get Manga via gRPC
```
GET /api/v1/grpc/manga/:id
Authorization: Bearer <token>
Response: { id, title, author, genres, ... , source: "grpc" }
```

### Search Manga via gRPC
```
GET /api/v1/grpc/manga/search?q=query&limit=20
Authorization: Bearer <token>
Response: { manga: [...], total: 10, source: "grpc" }
```

### Update Progress via gRPC
```
PUT /api/v1/grpc/progress/update
Authorization: Bearer <token>
Body: {
  "manga_id": "1",
  "current_chapter": 5,
  "status": "reading"
}
Response: { message: "...", source: "grpc", success: true }
```

## Architecture Components

### 1. gRPC Server (`cmd/grpc-server/main.go`)
- Standalone Go server
- Connects to database
- Connects to TCP server for broadcasts
- Exposes gRPC service on port 9001

### 2. gRPC Client (`internal/grpc/client.go`)
- Used by API server
- Maintains persistent connection
- Implements retry logic
- 5-10 second timeouts

### 3. API Server Integration
- gRPC client initialized on startup
- Graceful handling of gRPC unavailability
- Fallback to direct database queries if needed

### 4. Frontend Service
- Axios-based HTTP client
- Calls gRPC-backed HTTP endpoints
- Error handling with user feedback

## Troubleshooting

### gRPC Server Won't Start

```powershell
# Check if port 9001 is available
netstat -ano | findstr :9001

# Check database initialization
cd mangahub
go run cmd/api-server/main.go
```

### TCP Connection Failed

- gRPC server logs: "Warning: Failed to connect to TCP server"
- Solution: Start TCP server first
- gRPC will work without TCP, but no broadcasts

### API Server Can't Connect to gRPC

- API server logs: "WARNING: gRPC server not available"
- Check gRPC server is running
- Verify GRPC_SERVER_ADDR in .env
- API will show "Service Unavailable" but continue running

### Frontend Shows "Service Unavailable"

1. Check all servers are running
2. Verify authentication (login required)
3. Check browser console for errors
4. Verify CORS settings

## Performance Considerations

- **Timeouts**: 5s for GetManga, 10s for SearchManga
- **Connection Pooling**: gRPC maintains persistent connection
- **TCP Broadcasting**: Async (goroutine) to avoid blocking
- **Database Queries**: Use prepared statements
- **Pagination**: Default limit 20, max 100

## Security

- All gRPC endpoints require JWT authentication
- Token validated by API server before gRPC call
- User ID extracted from token for progress updates
- SQL injection protection via parameterized queries

## Future Enhancements

1. **Health Checks**: Add gRPC health check service
2. **Metrics**: Prometheus metrics for gRPC calls
3. **Load Balancing**: Multiple gRPC server instances
4. **Circuit Breaker**: Prevent cascading failures
5. **Caching**: Redis for frequently accessed manga
6. **Streaming**: Server-side streaming for large result sets

## Conclusion

The gRPC implementation provides a robust, scalable architecture for manga operations with real-time synchronization via TCP broadcasting. All three use cases (UC-014, UC-015, UC-016) are fully implemented and tested with both backend and frontend integration.
