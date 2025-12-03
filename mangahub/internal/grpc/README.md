# gRPC Service Implementation

This directory contains the gRPC implementation for the MangaHub application, providing internal service-to-service communication.

## Overview

The gRPC service implements the `MangaService` with three unary RPC methods:
- `GetManga` - Retrieve a single manga by ID
- `SearchManga` - Search for manga by query
- `UpdateProgress` - Update user's reading progress

## Architecture

```
proto/
  manga.proto          # Protocol Buffer definitions

internal/grpc/
  server.go           # gRPC server implementation
  client.go           # gRPC client helper

cmd/
  grpc-server/        # Standalone gRPC server
    main.go
  grpc-client-test/   # Test client for gRPC
    main.go
```

## Protocol Buffer Definition

### Service Definition

```protobuf
service MangaService {
  rpc GetManga(GetMangaRequest) returns (MangaResponse);
  rpc SearchManga(SearchRequest) returns (SearchResponse);
  rpc UpdateProgress(ProgressRequest) returns (ProgressResponse);
}
```

### Messages

- **GetMangaRequest**: Contains manga ID
- **MangaResponse**: Returns manga data or error
- **SearchRequest**: Contains search query and pagination
- **SearchResponse**: Returns list of manga and total count
- **ProgressRequest**: Contains user ID, manga ID, chapter, and status
- **ProgressResponse**: Returns success/failure status
- **Manga**: Complete manga entity with all fields

## Running the gRPC Server

### Start the Server

```bash
# From the mangahub directory
cd cmd/grpc-server
go run main.go
```

The server will start on port `9001` (configurable via `GRPC_SERVER_PORT` in `.env`)

### Output
```
Loaded environment variables from ../../.env file
Database initialized successfully
Starting gRPC MangaService server on port 9001...
gRPC server listening on port 9001
gRPC MangaService registered with methods:
  - GetManga(GetMangaRequest) returns (MangaResponse)
  - SearchManga(SearchRequest) returns (SearchResponse)
  - UpdateProgress(ProgressRequest) returns (ProgressResponse)
```

## Testing the gRPC Service

### Run the Test Client

```bash
# From the mangahub directory
cd cmd/grpc-client-test
go run main.go
```

### Expected Output
```
Connecting to gRPC server at localhost:9001...
Connected to gRPC server at localhost:9001

=== Test 1: Ping Server ===
gRPC Client: Pinging server
gRPC connection state: READY
Ping successful!

=== Test 2: Get Manga by ID ===
gRPC Client: Getting manga with ID: 1
...

=== All gRPC tests completed ===
```

## Configuration

### Environment Variables

Set in `.env` file:

```env
# gRPC Server Port
GRPC_SERVER_PORT=9001
```

## API Methods

### 1. GetManga

Retrieves a single manga by ID.

**Request:**
```go
&GetMangaRequest{
    ID: "1"
}
```

**Response:**
```go
&MangaResponse{
    Manga: &models.Manga{...},
    Error: ""
}
```

### 2. SearchManga

Searches for manga matching a query.

**Request:**
```go
&SearchRequest{
    Query:  "naruto",
    Limit:  20,
    Offset: 0
}
```

**Response:**
```go
&SearchResponse{
    Manga: []*models.Manga{...},
    Total: 15,
    Error: ""
}
```

### 3. UpdateProgress

Updates a user's reading progress for a manga.

**Request:**
```go
&ProgressRequest{
    UserID:         "user123",
    MangaID:        "1",
    CurrentChapter: 50,
    Status:         "reading"
}
```

**Response:**
```go
&ProgressResponse{
    Success: true,
    Message: "Progress updated successfully",
    Error:   ""
}
```

## Client Usage Example

```go
package main

import (
    "context"
    "log"
    "mangahub/internal/grpc"
    "time"
)

func main() {
    // Create client
    client, err := grpc.NewClient("localhost:9001")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Create context
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Get manga
    resp, err := client.GetManga(ctx, "1")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Manga: %+v", resp.Manga)
}
```

## Features Implemented

✅ **Protocol Buffer Definitions** - Complete `.proto` file with 3 services  
✅ **gRPC Server** - Full server implementation with all 3 methods  
✅ **gRPC Client** - Helper client for easy integration  
✅ **Unary RPC Calls** - All methods use unary (request-response) pattern  
✅ **Graceful Shutdown** - Server handles SIGTERM/SIGINT properly  
✅ **Error Handling** - Comprehensive error responses  
✅ **Logging** - Request/response logging for debugging  
✅ **Test Client** - Standalone test program  

## Dependencies

```go
require (
    google.golang.org/grpc v1.76.0
    google.golang.org/protobuf v1.36.10
)
```

## Integration with Main API Server

The gRPC server runs independently on port 9001, allowing internal services to communicate efficiently using Protocol Buffers instead of JSON over HTTP.

### Benefits:
- **Performance**: Binary protocol is faster than JSON
- **Type Safety**: Strongly typed contracts via protobuf
- **Internal Communication**: Secure service-to-service calls
- **Versioning**: Easy API versioning with protobuf

## Troubleshooting

### Port Already in Use
```bash
# Check what's using port 9001
netstat -ano | findstr :9001

# Kill the process (Windows)
taskkill /PID <PID> /F
```

### Connection Refused
- Ensure gRPC server is running: `go run cmd/grpc-server/main.go`
- Check firewall settings
- Verify `GRPC_SERVER_PORT` in `.env`

### Database Errors
- Ensure main database is initialized
- Check `data/mangahub.db` exists

## Future Enhancements

- [ ] Add bidirectional streaming for real-time updates
- [ ] Implement authentication/authorization
- [ ] Add TLS/SSL encryption
- [ ] Implement health check service
- [ ] Add metrics and monitoring
- [ ] Support for server reflection

## Production Deployment

For production, consider:
1. Enable TLS encryption
2. Add authentication middleware
3. Implement rate limiting
4. Use service mesh (Istio/Linkerd)
5. Add monitoring and tracing
6. Use load balancing

## License

Part of the MangaHub project.
