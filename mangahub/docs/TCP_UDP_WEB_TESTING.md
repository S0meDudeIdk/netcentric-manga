# Testing TCP/UDP Web Integration

## Pre-requisites

Ensure you have:
- Go 1.21+ installed
- Node.js 16+ installed
- GCC compiler (for SQLite)
- All dependencies installed (`go mod tidy`, `npm install`)

## Test Scenario 1: TCP Progress Sync

### Setup (4 terminals)

**Terminal 1 - API Server:**
```powershell
cd mangahub
$env:CGO_ENABLED = "1"
go run ./cmd/api-server
```

**Terminal 2 - TCP Server:**
```powershell
cd mangahub
go run ./cmd/tcp-server
```

**Terminal 3 - Web Client:**
```powershell
cd mangahub/client/web-react
npm start
```

**Terminal 4 - CLI Client (for testing):**
```powershell
cd mangahub
go run ./client/cli
```

### Test Steps

1. **Open web browser** → `http://localhost:3000`
2. **Login** to the web client
3. **Navigate** to "Real-time" page (click menu)
4. **Verify** green "TCP Progress Sync - Connected" indicator

5. **In CLI client (Terminal 4)**:
   - Login with test user
   - Browse manga
   - Update progress on any manga (read a chapter)

6. **Expected Result**:
   - Web client shows progress update in left panel
   - Shows: username, manga title, chapter number, timestamp
   - Update appears within 1-2 seconds

### Success Criteria ✅
- ✅ Connection indicator shows green
- ✅ Progress updates appear in real-time
- ✅ Updates show correct user and manga info
- ✅ Multiple updates accumulate in the list
- ✅ Connection stays stable for 5+ minutes

## Test Scenario 2: UDP Notifications

### Additional Setup

**Terminal 5 - UDP Server:**
```powershell
cd mangahub
go run ./cmd/udp-server
```

### Test Steps

1. **Verify** green "UDP Notifications - Connected" indicator on web page

2. **Method A - Use broadcast test utility:**
   ```powershell
   # Terminal 6
   cd mangahub
   go run ./cmd/udp-broadcast-test
   ```
   - Enter notification details when prompted
   - Select type (chapter_release, status_update, etc.)

3. **Method B - Use HTTP trigger:**
   ```powershell
   curl -X POST http://localhost:9020/trigger `
     -H "Content-Type: application/json" `
     -d '{
       "type": "chapter_release",
       "manga_id": "one-piece",
       "message": "Chapter 1101 is now available!",
       "timestamp": 1734710400
     }'
   ```

4. **Expected Result**:
   - Notification appears in right panel
   - Shows type badge, message, manga ID
   - Browser notification may appear (if permissions granted)

### Success Criteria ✅
- ✅ Connection indicator shows green
- ✅ Notifications appear in real-time
- ✅ Notification type is displayed correctly
- ✅ Multiple notifications accumulate in the list
- ✅ Browser notifications work (after granting permission)

## Test Scenario 3: Web Progress Updates → TCP Broadcast

### Test Steps

1. **Keep web "Real-time" page open** in Browser 1
2. **Open another browser tab** (Browser 2)
3. **In Browser 2**:
   - Navigate to any manga detail page
   - Update your reading progress
   - OR use browser console:
     ```javascript
     fetch('http://localhost:8080/api/v1/users/progress', {
       method: 'PUT',
       headers: {
         'Authorization': 'Bearer ' + localStorage.getItem('token'),
         'Content-Type': 'application/json'
       },
       body: JSON.stringify({
         manga_id: 'test-manga',
         current_chapter: 42
       })
     });
     ```

4. **Expected Result**:
   - Browser 1 (Real-time page) shows your progress update
   - CLI client (if connected) also receives the update
   - Update shows within 1-2 seconds

### Success Criteria ✅
- ✅ Web client can trigger TCP broadcasts
- ✅ Broadcasts are received by other web clients
- ✅ Broadcasts are received by CLI clients
- ✅ Round-trip latency < 2 seconds

## Test Scenario 4: Connection Recovery

### Test Steps

1. **Start all servers and web client**
2. **Navigate to Real-time page**
3. **Verify both services are connected** (green indicators)
4. **Stop TCP server** (Ctrl+C in Terminal 2)
5. **Observe**:
   - Indicator turns yellow
   - Console shows reconnection attempts
6. **Restart TCP server** (Terminal 2)
7. **Observe**:
   - Indicator returns to green
   - Updates resume working

8. **Repeat for UDP server** (Terminal 5)

### Success Criteria ✅
- ✅ Service detects disconnection within 5 seconds
- ✅ Reconnection attempts logged in console
- ✅ Auto-reconnects when server is available
- ✅ No manual page refresh needed
- ✅ Max 5 reconnection attempts before giving up

## Test Scenario 5: Multiple Web Clients

### Test Steps

1. **Open 3 browser windows** (or incognito/different browsers)
2. **Login with different users** in each
3. **Navigate all to Real-time page**
4. **In CLI client**, update progress
5. **Verify** all 3 web clients receive the update simultaneously

6. **Trigger UDP notification**
7. **Verify** all 3 web clients receive the notification

### Success Criteria ✅
- ✅ All clients receive broadcasts
- ✅ No message loss
- ✅ No significant delay differences
- ✅ Server handles 3+ concurrent connections

## Test Scenario 6: Browser Notifications

### Test Steps

1. **Open Real-time page**
2. **When prompted**, grant notification permissions
3. **Trigger a UDP notification**
4. **Expected**:
   - Desktop notification appears
   - Notification shows in browser
   - Sound may play (browser dependent)

5. **Minimize browser**
6. **Trigger another notification**
7. **Verify** you still receive desktop notification

### Success Criteria ✅
- ✅ Permission prompt appears
- ✅ Notifications show when browser is visible
- ✅ Notifications show when browser is minimized
- ✅ Notification content is correct

## Test Scenario 7: Performance Test

### Test Steps

1. **Start all servers**
2. **Open Real-time page**
3. **Run load test**:
   ```powershell
   # Send 100 progress updates rapidly
   for ($i=1; $i -le 100; $i++) {
     curl -X POST http://localhost:9010/trigger `
       -H "Content-Type: application/json" `
       -d "{\"user_id\":\"user$i\",\"username\":\"User$i\",\"manga_title\":\"Test Manga\",\"chapter\":$i,\"timestamp\":$(Get-Date -UFormat %s)}"
     Start-Sleep -Milliseconds 100
   }
   ```

4. **Observe**:
   - Page remains responsive
   - Updates appear smoothly
   - No browser freezing
   - Memory usage stays reasonable

### Success Criteria ✅
- ✅ Handles 100 updates without issues
- ✅ Page stays responsive
- ✅ No JavaScript errors
- ✅ Memory usage < 200MB

## Test Scenario 8: API Server Logs

### Verify Logs Show

**On startup:**
```
✅ SSE Hub initialized for real-time updates
Attempting to connect to TCP Progress Sync Server at localhost:9000...
✅ Connected to TCP Progress Sync Server - Real-time updates enabled
UDP client listening on [::]:xxxxx
✅ Connected to UDP Notification Server - Push notifications enabled
```

**During operation:**
```
SSE Progress client connected: progress-abc123 (Total: 1)
📡 TCP Progress Update: User=JohnDoe, Manga=One Piece, Chapter=1050
🔔 UDP Notification: Type=chapter_release, Manga=one-piece, Message=New chapter!
SSE Progress client disconnected: progress-abc123 (Remaining: 0)
```

## Troubleshooting Common Issues

### Issue: "Failed to connect to TCP server"

**Diagnosis:**
```powershell
# Check if TCP server is running
netstat -an | findstr 9000
```

**Solution:**
- Ensure TCP server is started
- Check port is not blocked by firewall
- Verify correct port in environment variables

### Issue: "No updates appearing in web client"

**Diagnosis:**
1. Open browser DevTools → Console
2. Look for errors
3. Check Network tab for SSE connection

**Solution:**
- Verify JWT token is valid (check localStorage)
- Ensure SSE endpoint is accessible
- Check CORS configuration

### Issue: "Connection keeps dropping"

**Diagnosis:**
- Check server logs for errors
- Monitor network stability
- Check firewall/proxy settings

**Solution:**
- Increase keep-alive interval
- Check for network interference
- Verify server resources (CPU, memory)

## Validation Checklist

Before considering implementation complete:

- [ ] All servers start without errors
- [ ] Web client connects to both SSE endpoints
- [ ] TCP progress updates appear in web client
- [ ] UDP notifications appear in web client
- [ ] CLI client can trigger web updates
- [ ] Web client can trigger CLI updates
- [ ] Multiple web clients receive broadcasts
- [ ] Auto-reconnection works after server restart
- [ ] Browser notifications work
- [ ] No memory leaks after 30+ minutes
- [ ] No JavaScript errors in console
- [ ] All documentation is accurate
- [ ] Code is properly commented

## Performance Metrics

### Expected Values

| Metric | Expected | Acceptable |
|--------|----------|------------|
| TCP connection time | < 100ms | < 500ms |
| UDP registration time | < 100ms | < 500ms |
| SSE connection time | < 200ms | < 1s |
| Progress update latency | < 500ms | < 2s |
| Notification latency | < 200ms | < 1s |
| Reconnection time | < 5s | < 10s |
| Memory per client | < 10MB | < 50MB |

## Test Report Template

```markdown
# Test Report: TCP/UDP Web Integration

**Date**: YYYY-MM-DD
**Tester**: Your Name
**Environment**: Windows/Mac/Linux

## Results

### Scenario 1: TCP Progress Sync
- [ ] PASS / [ ] FAIL
- Notes: _____________________

### Scenario 2: UDP Notifications
- [ ] PASS / [ ] FAIL
- Notes: _____________________

### Scenario 3: Web → TCP Broadcast
- [ ] PASS / [ ] FAIL
- Notes: _____________________

### Scenario 4: Connection Recovery
- [ ] PASS / [ ] FAIL
- Notes: _____________________

### Scenario 5: Multiple Clients
- [ ] PASS / [ ] FAIL
- Notes: _____________________

### Scenario 6: Browser Notifications
- [ ] PASS / [ ] FAIL
- Notes: _____________________

### Scenario 7: Performance
- [ ] PASS / [ ] FAIL
- Notes: _____________________

## Issues Found
1. _____________________
2. _____________________

## Overall Status
- [ ] All tests passed
- [ ] Minor issues (acceptable)
- [ ] Major issues (needs fixing)
```

---

**Ready to test? Start with Scenario 1 and work your way through! 🚀**
