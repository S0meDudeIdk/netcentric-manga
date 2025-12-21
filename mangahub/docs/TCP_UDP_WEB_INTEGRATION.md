# Real-time TCP/UDP Integration for Web-React

## Overview

This implementation bridges TCP Progress Sync and UDP Notifications to the web-react client using **Server-Sent Events (SSE)**. Since browsers cannot directly connect to TCP/UDP sockets, the API server acts as a bridge that receives broadcasts from TCP/UDP servers and streams them to web clients via SSE.

## Architecture

```
┌─────────────────┐         ┌─────────────────┐         ┌──────────────────┐
│  TCP Server     │◄────────┤   API Server    │◄────────┤  Web-React       │
│  (Port 9001)    │  TCP    │   (Port 8080)   │  SSE    │  Client          │
│  Progress Sync  │         │                 │  HTTP   │  (Browser)       │
└─────────────────┘         │  ┌───────────┐  │         └──────────────────┘
                            │  │ SSE Hub   │  │
┌─────────────────┐         │  │           │  │         ┌──────────────────┐
│  UDP Server     │◄────────┤  │ TCP Client│  │◄────────┤  CLI Client      │
│  (Port 8081)    │  UDP    │  │ UDP Client│  │  TCP    │                  │
│  Notifications  │         │  └───────────┘  │  UDP    │  (Direct socket) │
└─────────────────┘         └─────────────────┘         └──────────────────┘
```

## Components

### Backend (Go)

#### 1. SSE Hub (`internal/api/sse.go`)
- **Purpose**: Manages SSE connections and broadcasts
- **Features**:
  - Progress client management (TCP bridge)
  - Notification client management (UDP bridge)
  - Concurrent broadcasting with goroutines
  - Keep-alive pings every 30 seconds
  - Automatic client cleanup

```go
type SSEHub struct {
    progressClients       map[string]*SSEClient
    notificationClients   map[string]*SSEClient
    ProgressBroadcast     chan interface{}
    NotificationBroadcast chan interface{}
}
```

#### 2. TCP Client (`internal/api/tcp_client.go`)
- **Purpose**: Connects to TCP server to receive progress broadcasts
- **Features**:
  - Auto-reconnection with exponential backoff
  - JSON message parsing
  - Thread-safe operations
  - Forwards updates to SSE hub

```go
type TCPProgressUpdate struct {
    UserID     string `json:"user_id"`
    Username   string `json:"username"`
    MangaTitle string `json:"manga_title"`
    Chapter    int    `json:"chapter"`
    Timestamp  int64  `json:"timestamp"`
}
```

#### 3. UDP Client (`internal/api/udp_client.go`)
- **Purpose**: Listens for UDP notifications
- **Features**:
  - Client registration with heartbeat
  - Message filtering (ignores pings)
  - Thread-safe operations
  - Forwards notifications to SSE hub

```go
type UDPNotification struct {
    Type      string `json:"type"`
    MangaID   string `json:"manga_id"`
    Message   string `json:"message"`
    Timestamp int64  `json:"timestamp"`
}
```

#### 4. SSE Endpoints (`internal/api/sseHandlers.go`)
- `GET /api/v1/sse/progress` - Stream TCP progress updates
- `GET /api/v1/sse/notifications` - Stream UDP notifications
- **Authentication**: JWT token via query parameter
- **Format**: Server-Sent Events (text/event-stream)

### Frontend (React)

#### 1. Progress Sync Service (`services/progressSyncService.js`)
- **Purpose**: Consumes TCP progress updates via SSE
- **Features**:
  - EventSource-based connection
  - Multiple listeners support
  - Auto-reconnection (max 5 attempts)
  - Connection state tracking

```javascript
progressSyncService.connect(token, onProgress, onError);
```

#### 2. UDP Notification Service (`services/udpNotificationService.js`)
- **Purpose**: Consumes UDP notifications via SSE
- **Features**:
  - EventSource-based connection
  - Notification history (last 50)
  - Browser notifications support
  - Read/unread tracking

```javascript
udpNotificationService.connect(token, onNotification, onError);
```

#### 3. Real-time Sync Page (`pages/RealtimeSyncPage.jsx`)
- **Purpose**: Demo page showcasing real-time features
- **Features**:
  - Live progress update feed
  - Live notification feed
  - Connection status indicators
  - Auto-updates without page refresh

## Setup Instructions

### 1. Backend Setup

**Environment Variables** (optional):
```bash
# In mangahub/.env
TCP_PROGRESS_ADDR=127.0.0.1:9001      # TCP server address
UDP_NOTIFICATION_ADDR=127.0.0.1:8081  # UDP server address
```

**Start all servers:**

Terminal 1 - API Server:
```powershell
cd mangahub
$env:CGO_ENABLED = "1"
go run ./cmd/api-server
```

Terminal 2 - TCP Server:
```powershell
cd mangahub
go run ./cmd/tcp-server
```

Terminal 3 - UDP Server:
```powershell
cd mangahub
go run ./cmd/udp-server
```

Terminal 4 - Web Client:
```powershell
cd mangahub/client/web-react
npm start
```

### 2. Frontend Setup

No additional configuration needed. The services automatically connect when users navigate to `/realtime` page.

## Usage

### Web Client

1. **Navigate to Real-time Page**: 
   - Login to the application
   - Click "Real-time" in the navigation menu
   - Or visit `http://localhost:3000/realtime`

2. **View Live Updates**:
   - Left panel: TCP progress updates (when users read manga)
   - Right panel: UDP notifications (chapter releases, etc.)

3. **Connection Status**:
   - Green indicator = Connected and receiving updates
   - Yellow indicator = Connecting or reconnecting

### Programmatic Usage

**In any React component:**

```javascript
import progressSyncService from '../services/progressSyncService';
import udpNotificationService from '../services/udpNotificationService';

// Connect to progress sync
useEffect(() => {
  const token = localStorage.getItem('token');
  
  progressSyncService.connect(
    token,
    (update) => {
      console.log('Progress update:', update);
      // Handle update in your component
    },
    (error) => {
      console.error('Error:', error);
    }
  );

  return () => {
    progressSyncService.disconnect();
  };
}, []);

// Connect to notifications
useEffect(() => {
  const token = localStorage.getItem('token');
  
  udpNotificationService.connect(
    token,
    (notification) => {
      console.log('Notification:', notification);
      // Handle notification in your component
    }
  );

  return () => {
    udpNotificationService.disconnect();
  };
}, []);
```

## Testing

### Test TCP Progress Updates

**Method 1: Use CLI Client**
```powershell
cd mangahub
go run ./client/cli
# Login and update progress on any manga
```

**Method 2: Use Web Client**
```javascript
// In browser console or any page with authentication
fetch('http://localhost:8080/api/v1/users/progress', {
  method: 'PUT',
  headers: {
    'Authorization': 'Bearer ' + localStorage.getItem('token'),
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    manga_id: 'one-piece',
    current_chapter: 1050
  })
});
```

### Test UDP Notifications

**Use UDP broadcast test utility:**
```powershell
cd mangahub
go run ./cmd/udp-broadcast-test
```

Or trigger via HTTP:
```bash
curl -X POST http://localhost:9020/trigger \
  -H "Content-Type: application/json" \
  -d '{
    "type": "chapter_release",
    "manga_id": "one-piece",
    "message": "New chapter 1101 released!",
    "timestamp": 1734710400
  }'
```

## Message Formats

### TCP Progress Update
```json
{
  "user_id": "user123",
  "username": "JohnDoe",
  "manga_title": "One Piece",
  "chapter": 1050,
  "timestamp": 1734710400
}
```

### UDP Notification
```json
{
  "type": "chapter_release",
  "manga_id": "one-piece",
  "message": "New chapter 1101 released for One Piece",
  "timestamp": 1734710400
}
```

## Features

### TCP Progress Sync ✅
- ✅ Real-time progress broadcasting
- ✅ Multiple concurrent connections
- ✅ JSON message protocol
- ✅ Auto-reconnection
- ✅ SSE bridge for web clients
- ✅ Thread-safe operations

### UDP Notifications ✅
- ✅ Client registration mechanism
- ✅ Broadcast notifications
- ✅ Chapter release notifications
- ✅ Heartbeat keep-alive
- ✅ SSE bridge for web clients
- ✅ Browser notifications support

### SSE Infrastructure ✅
- ✅ Connection management
- ✅ Keep-alive pings
- ✅ Automatic cleanup
- ✅ Multiple listener support
- ✅ Error handling and reconnection

## Troubleshooting

### Issue: SSE not connecting

**Solution:**
1. Check if API server is running
2. Verify JWT token is valid
3. Check browser console for errors
4. Ensure CORS is properly configured

### Issue: No progress updates received

**Solution:**
1. Verify TCP server is running on port 9000
2. Check API server logs for TCP connection status
3. Test TCP server with CLI client
4. Ensure progress updates are being triggered

### Issue: No UDP notifications

**Solution:**
1. Verify UDP server is running on port 8081
2. Check API server logs for UDP registration
3. Test with UDP broadcast utility
4. Check firewall settings for UDP traffic

### Issue: Connection keeps dropping

**Solution:**
1. Check network stability
2. Verify server keep-alive settings
3. Check for proxy/firewall interference
4. Monitor server logs for errors

## Performance Considerations

- **SSE Connections**: Each web client creates 2 SSE connections (progress + notifications)
- **Memory**: SSE hub stores last 50 messages per type
- **Bandwidth**: Minimal - only sends data when events occur
- **Scalability**: Suitable for 100+ concurrent web clients

## Security

- **Authentication**: All SSE endpoints require valid JWT tokens
- **CORS**: Configured to allow web client origin
- **Rate Limiting**: Inherited from API server middleware
- **Data Validation**: JSON parsing with error handling

## Future Enhancements

1. **Message Persistence**: Store messages in Redis for reconnection recovery
2. **User-Specific Filters**: Subscribe to specific manga or user updates
3. **Compression**: Implement message compression for large payloads
4. **WebSocket Fallback**: Use WebSocket when SSE is not supported
5. **Health Monitoring**: Add metrics and monitoring dashboards

## Files Created/Modified

### Backend
- `internal/api/sse.go` - SSE hub implementation
- `internal/api/tcp_client.go` - TCP client for receiving broadcasts
- `internal/api/udp_client.go` - UDP client for receiving notifications
- `internal/api/sseHandlers.go` - SSE endpoint handlers
- `internal/api/api.go` - Updated to initialize SSE components
- `internal/api/routes.go` - Added SSE routes
- `internal/api/services.go` - Added TCP/UDP client initialization

### Frontend
- `services/progressSyncService.js` - TCP progress sync service
- `services/udpNotificationService.js` - UDP notification service
- `pages/RealtimeSyncPage.jsx` - Real-time demo page
- `components/Header.jsx` - Added "Real-time" navigation link
- `App.js` - Added real-time route

### Documentation
- `docs/TCP_UDP_WEB_INTEGRATION.md` - This file

## API Endpoints Summary

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/api/v1/sse/progress` | GET | Required | Stream TCP progress updates |
| `/api/v1/sse/notifications` | GET | Required | Stream UDP notifications |

## References

- [Server-Sent Events Specification](https://html.spec.whatwg.org/multipage/server-sent-events.html)
- [EventSource API](https://developer.mozilla.org/en-US/docs/Web/API/EventSource)
- [TCP Protocol Documentation](docs/TCP_UDP_INTEGRATION.md)
- [UDP Protocol Documentation](docs/TCP_UDP_README.md)
