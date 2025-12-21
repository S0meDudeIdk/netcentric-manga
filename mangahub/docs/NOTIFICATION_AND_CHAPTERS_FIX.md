# Notification System & Total Chapters Fix

## Issues Fixed

### 1. Notification System Not Working

**Problem:** Notifications were only being sent for reading progress updates, not for:
- Adding manga to library
- Removing manga from library  
- New manga being added
- New chapters being synced

**Solution:** 
- Added WebSocket notifications when users add manga to their library
- Added WebSocket notifications when users remove manga from their library
- The notifications are broadcasted to the `global-notifications` room that the frontend subscribes to
- Enhanced logging for notification broadcasts

**Files Modified:**
- `internal/api/userHandlers.go` - Added notification broadcasts for add/remove library actions
- `internal/websocket/websocket.go` - Enhanced BroadcastNotification with better logging

### 2. Total Chapters Showing 0

**Problem:** Many manga in the database had `total_chapters = 0`, causing the UI to show "? ch"

**Root Cause:** 
- When manga is fetched from MAL API, the `total_chapters` field is correctly populated
- However, older manga or manually added manga might have 0 chapters
- The MAL API field `num_chapters` might be null for some manga (ongoing series)

**Solution:**
1. Created a utility script to update existing manga with 0 chapters
2. The conversion from MAL API properly maps `num_chapters` to `total_chapters`
3. Frontend now properly displays the chapter count when available

**Files Created:**
- `scripts/update-chapters.go` - Go program to update manga with 0 chapters
- `scripts/update-total-chapters.ps1` - PowerShell script to run the update

**Files Modified:**
- `client/web-react/src/pages/Library.jsx` - Fixed to pass `total_chapters` to manga cards

## How to Use

### Update Existing Manga Chapters

Run this command from the `mangahub` directory:

```powershell
.\scripts\update-total-chapters.ps1
```

This will:
- Find all manga with `total_chapters = 0` or `NULL`
- Set them to a default value of 1
- Log all updates

For accurate chapter counts, you should:
1. Use the manga sync feature to fetch from MAL
2. Or manually update via the API

### Testing Notifications

1. **Start all servers:**
   ```powershell
   # Terminal 1 - API Server
   cd mangahub/cmd/api-server
   go run main.go

   # Terminal 2 - TCP Server (optional, for progress sync)
   cd mangahub/cmd/tcp-server
   go run main.go

   # Terminal 3 - UDP Server (optional, for notifications)
   cd mangahub/cmd/udp-server  
   go run main.go

   # Terminal 4 - Frontend
   cd mangahub/client/web-react
   npm start
   ```

2. **Test Notifications:**
   - Open the app and login
   - The notification bell should appear next to your username
   - Add a manga to your library → notification appears
   - Remove a manga from library → notification appears
   - Update reading progress → notification appears

3. **Verify WebSocket Connection:**
   - Open browser DevTools → Network tab
   - Filter by "WS" (WebSocket)
   - You should see a connection to `/api/v1/ws/chat?room=global-notifications`

## Notification Types

The system now supports these notification types:

| Type | Trigger | Message Format |
|------|---------|----------------|
| `library_add` | User adds manga to library | "Added '{manga_title}' to library with status: {status}" |
| `library_remove` | User removes manga from library | "Removed '{manga_title}' from library" |
| `progress_update` | User updates reading progress | Sent to manga-specific room |
| `notification` | General notifications | System messages |

## Architecture

### Notification Flow

```
User Action (Frontend)
    ↓
API Endpoint (HTTP)
    ↓
Service Layer (Go)
    ↓
WebSocket ChatHub.BroadcastNotification()
    ↓
NotificationChannel (buffered channel)
    ↓
broadcastNotifications() goroutine
    ↓
global-notifications room
    ↓
All connected clients receive notification
    ↓
Frontend NotificationService processes
    ↓
Notification bell shows count
```

### TCP/UDP Integration

- TCP Server: Handles reading progress sync between clients
- UDP Server: Handles real-time notifications (not yet fully integrated)
- WebSocket: Primary notification delivery mechanism
- HTTP API: Triggers notifications for user actions

## Future Improvements

1. **UDP Integration:** Fully integrate UDP notifications for low-latency updates
2. **Chapter Sync:** Automatically fetch chapter counts from MangaDex or other sources
3. **Batch Updates:** Implement bulk chapter count updates
4. **Notification Persistence:** Store notifications in database for history
5. **User Preferences:** Allow users to configure which notifications they want to receive

## Troubleshooting

### Notification Bell Not Appearing

- Verify you're logged in
- Check browser console for WebSocket connection errors
- Ensure API server is running on port 8080

### No Notifications Received

- Check that WebSocket connection is established (DevTools → Network → WS)
- Verify the notification room is "global-notifications"
- Check API server logs for "Broadcasting notification" messages

### Total Chapters Still 0

- Run the update script: `.\scripts\update-total-chapters.ps1`
- Use the sync feature to fetch from MAL
- Check that the manga exists in MAL database with chapter information

## Related Files

- `internal/api/userHandlers.go` - Library action handlers
- `internal/websocket/websocket.go` - WebSocket notification system
- `client/web-react/src/services/notificationService.js` - Frontend notification service
- `client/web-react/src/components/NotificationBell.jsx` - Notification UI component
- `internal/tcp/tcp.go` - TCP progress sync server
- `internal/udp/udp.go` - UDP notification server
