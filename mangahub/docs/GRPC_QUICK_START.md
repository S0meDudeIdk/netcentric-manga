# Quick Start Guide: Testing gRPC Implementation

## Quick Setup (5 minutes)

### Step 1: Start All Servers
```powershell
cd mangahub\scripts
.\start-grpc-test-env.ps1
```

This will open 3 PowerShell windows for:
- TCP Server (Port 9000)
- gRPC Server (Port 9001)
- API Server (Port 8080)

Wait for all servers to show "listening" or "started" messages.

### Step 2: Start Frontend
```powershell
cd mangahub\client\web-react
npm start
```

Browser should open to http://localhost:3000

### Step 3: Login
1. Go to http://localhost:3000/login
2. Use existing account or register new one
3. After login, you'll be redirected to home page

### Step 4: Access gRPC Test Page
1. Click on your username (top right)
2. Select "gRPC Test" from dropdown menu
3. Or navigate directly to http://localhost:3000/grpc-test

## Testing Each Use Case

### UC-014: Get Manga via gRPC

**Test Steps:**
1. In the "Get Manga via gRPC" section (blue)
2. Enter manga ID: `1` (or any valid ID from your database)
3. Click "Get Manga" button
4. **Expected Result:** JSON response with manga details including "source": "grpc"

**Success Indicators:**
- ✓ Manga details displayed (title, author, genres, etc.)
- ✓ "source": "grpc" field present in response
- ✓ No errors shown

### UC-015: Search Manga via gRPC

**Test Steps:**
1. In the "Search Manga via gRPC" section (green)
2. Enter search query: `naruto` (or any keyword)
3. Set limit: `10`
4. Click "Search Manga" button
5. **Expected Result:** List of matching manga with total count

**Success Indicators:**
- ✓ Results array displayed
- ✓ "total" field shows number of matches
- ✓ "source": "grpc" in response
- ✓ Each manga has complete details

### UC-016: Update Progress via gRPC

**Prerequisites:** 
- Manga must be in your library first
- Go to http://localhost:3000/browse
- Add a manga to your library using the "+ Add to Library" button

**Test Steps:**
1. In the "Update Progress via gRPC" section (purple)
2. Enter manga ID: `1` (the manga you added to library)
3. Set chapter number: `5`
4. Select status: `reading`
5. Click "Update Progress" button
6. **Expected Result:** Success message with TCP broadcast confirmation

**Success Indicators:**
- ✓ "success": true in response
- ✓ "Progress updated successfully" message
- ✓ Green text: "✓ Progress updated and broadcast via TCP"
- ✓ Check TCP server window - should show broadcast message

**Verify in Database:**
```powershell
# In API server window, you'll see logs like:
# "gRPC UpdateProgress called for user: <user_id>, manga: 1"
# "Broadcasted progress update via TCP"
```

## Troubleshooting

### ❌ "gRPC service unavailable"

**Problem:** gRPC server not running or not connected

**Solution:**
```powershell
# Check if gRPC server window shows:
# "gRPC server listening on port 9001"
# "Connected to TCP server at localhost:9000"
```

If not, restart servers using the start script.

### ❌ "Authentication Required"

**Problem:** Not logged in

**Solution:**
1. Navigate to http://localhost:3000/login
2. Login or register
3. Return to http://localhost:3000/grpc-test

### ❌ "manga not found in user's library"

**Problem:** Trying to update progress for manga not in library

**Solution:**
1. Go to http://localhost:3000/browse
2. Find a manga
3. Click "Add to Library"
4. Then try update progress again

### ❌ No results in search

**Problem:** Database empty or search term doesn't match

**Solution:**
- Try searching for "" (empty) to see all manga
- Or browse http://localhost:3000/browse to see what's available

## Expected Server Logs

### TCP Server (Port 9000)
```
TCP Server listening on :9000
New connection from 127.0.0.1:xxxxx
Broadcasted update to N clients
```

### gRPC Server (Port 9001)
```
gRPC server listening on port 9001
Connected to TCP server at localhost:9000
gRPC GetManga called with ID: 1
gRPC SearchManga called with query: naruto
gRPC UpdateProgress called for user: xxx, manga: 1
Broadcasted progress update via TCP
```

### API Server (Port 8080)
```
[GIN-debug] Listening and serving HTTP on :8080
Connected to gRPC server at localhost:9001
GET /api/v1/grpc/manga/1
PUT /api/v1/grpc/progress/update
```

## Test Data

If you need test data, here are some manga IDs that should exist:
- ID: `1` - One Piece
- ID: `2` - Naruto
- ID: `3` - Bleach

Search terms that should work:
- "one piece"
- "naruto"
- "action" (genre)
- "" (empty - returns all)

## Quick Validation Checklist

- [ ] All 3 servers running (TCP, gRPC, API)
- [ ] Frontend running on port 3000
- [ ] Logged in with valid account
- [ ] gRPC Test page shows "✓ Available" status
- [ ] UC-014: Can retrieve manga by ID
- [ ] UC-015: Can search manga
- [ ] UC-016: Can update progress (manga in library)
- [ ] TCP broadcast visible in server logs

## Video Demo Script

1. **Show all servers running** (3 PowerShell windows)
2. **Navigate to test page** (show "Available" status)
3. **UC-014 Demo:**
   - Get manga ID 1
   - Show JSON response
   - Highlight "source": "grpc"
4. **UC-015 Demo:**
   - Search "naruto"
   - Show results array
   - Point out pagination
5. **UC-016 Demo:**
   - Update progress for manga 1
   - Show success message
   - **Switch to TCP server window**
   - Show broadcast log entry
   - Back to browser, verify update

## Need Help?

Check the full documentation:
- `docs/GRPC_IMPLEMENTATION_COMPLETE.md` - Complete implementation guide
- `docs/API_DOCUMENTATION.md` - API reference
- `GRPC_IMPLEMENTATION.md` - Original gRPC setup guide

## Screenshots Location

Save screenshots of:
1. All servers running
2. gRPC test page - Available status
3. UC-014 response
4. UC-015 search results
5. UC-016 success with TCP broadcast
6. TCP server logs showing broadcast
