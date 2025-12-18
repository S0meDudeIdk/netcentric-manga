# Chapter Storage and Display Fix

## Problem Analysis

The manga database has 1,634 manga synced from MangaDex, but chapters are not displaying on the manga detail pages. The issue was identified in the sync logic:

### Root Cause
1. **During initial sync**: When manga is new, both manga and chapters are stored ✅
2. **During subsequent syncs**: When manga already exists, the entire sync was skipped - including chapter storage ❌
3. **After database corruption**: If chapters were lost, re-running sync wouldn't restore them because existing manga were skipped completely

### The Broken Logic
```go
// OLD CODE - BROKEN
if exists {
    log.Printf("Already exists, skipping")
    result.Skipped++
    continue  // Skips checking/adding chapters!
}
```

## Solution Implemented

### 1. Smart Skip Logic
Now the sync checks BOTH manga existence AND chapter count:

```go
// NEW CODE - FIXED
// Check if manga exists
var mangaExists bool
err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM manga WHERE id = ?)", mangaID).Scan(&mangaExists)

// Check if chapters exist
var chapterCount int
err = s.db.QueryRow("SELECT COUNT(*) FROM manga_chapters WHERE manga_id = ?)", mangaID).Scan(&chapterCount)

// Only skip if BOTH manga and chapters exist
if mangaExists && chapterCount > 0 {
    log.Printf("Already exists with %d chapters, skipping", chapterCount)
    result.Skipped++
    continue
}
```

### 2. Conditional Manga Storage
- If manga exists: Only sync chapters
- If manga doesn't exist: Store both manga and chapters

```go
// Store manga only if it doesn't exist
if !mangaExists {
    if err := s.storeMangaDirect(manga, mdManga.ID); err != nil {
        // error handling
    }
    log.Printf("Manga stored successfully")
} else {
    log.Printf("Manga already exists, updating chapters only")
}

// Store chapters (whether manga is new or existing)
stored := 0
for _, ch := range chapters.Data {
    // Store chapter logic
}
```

### 3. Frontend Fix
Removed restrictive ID check that prevented chapter fetching for "md-" prefixed manga IDs:

```javascript
// OLD CODE - BROKEN
if (!idStr.startsWith('mal-') && !idStr.includes('mangadex') && !idStr.includes('mangaplus')) {
    setUseRealChapters(false);
    return; // Skips chapter fetching for "md-" IDs!
}

// NEW CODE - FIXED
// Fetch chapters for all manga (local database now has chapters)
const idStr = id.toString();
try {
    setLoadingChapters(true);
    const chaptersData = await mangaService.getChapters(id, ['en'], 500, 0);
    // ... rest of logic
}
```

### 4. Manual Chapter Sync Endpoint
Added a new API endpoint to force chapter syncing for manga without chapters:

**Endpoint**: `POST /api/v1/manga/sync-chapters`

**Response**:
```json
{
  "success": true,
  "total_fetched": 100,
  "synced": 85,
  "skipped": 15,
  "failed": 0,
  "message": "Synced chapters for 85 manga"
}
```

## Files Modified

### Backend Changes:
1. **internal/manga/sync_service.go** (Lines 273-325)
   - Added chapter count check before skipping manga
   - Conditional manga storage (skip if exists, always sync chapters)
   - Better logging for debugging

2. **cmd/api-server/main.go**
   - Added `/api/v1/manga/sync-chapters` endpoint (Line 342)
   - Implemented `syncMangaChapters` handler (Lines 2508-2532)

### Frontend Changes:
3. **client/web-react/src/pages/MangaDetail.jsx** (Lines 115-145)
   - Removed restrictive ID format check
   - Now fetches chapters for all manga regardless of ID prefix

### Scripts:
4. **scripts/sync-chapters.ps1** (NEW)
   - PowerShell script to manually trigger chapter sync
   - Useful for recovering from database issues

## How to Use

### Automatic (Recommended)
The sync now runs automatically on server startup and will:
1. Fetch new manga from MangaDex
2. Add chapters for any manga without chapters
3. Skip manga that already have chapters

### Manual Sync
If you need to force a chapter sync (e.g., after database recovery):

**Option 1: Using PowerShell Script**
```powershell
cd mangahub/scripts
.\sync-chapters.ps1
```

**Option 2: Using curl/Postman**
```bash
curl -X POST http://localhost:8080/api/v1/manga/sync-chapters
```

**Option 3: Using JavaScript (Browser Console)**
```javascript
fetch('http://localhost:8080/api/v1/manga/sync-chapters', {
  method: 'POST'
})
.then(r => r.json())
.then(data => console.log(data));
```

## Testing

### 1. Verify Chapters Appear
1. Start the server: `.\scripts\start-server.ps1`
2. Open browser: http://localhost:3000
3. Navigate to Browse page
4. Click on any manga (e.g., "One Piece")
5. Scroll to "Chapters" section
6. ✅ Chapters should now be displayed

### 2. Check Database
Query the database to verify chapters are stored:
```sql
-- Count manga
SELECT COUNT(*) FROM manga;

-- Count chapters
SELECT COUNT(*) FROM manga_chapters;

-- Check specific manga's chapters
SELECT COUNT(*) FROM manga_chapters WHERE manga_id = 'md-{uuid}';

-- List some chapters
SELECT manga_id, chapter_number, title, language 
FROM manga_chapters 
LIMIT 10;
```

### 3. Test Chapter Reading
1. Click on a chapter from the list
2. Reader should load chapter pages
3. Pages should display correctly

## Database Schema

### manga_chapters Table
```sql
CREATE TABLE manga_chapters (
    id TEXT PRIMARY KEY,
    manga_id TEXT NOT NULL,
    chapter_number TEXT NOT NULL,
    title TEXT,
    volume TEXT,
    language TEXT NOT NULL,
    pages INTEGER,
    source TEXT NOT NULL,
    source_chapter_id TEXT NOT NULL,
    FOREIGN KEY (manga_id) REFERENCES manga(id)
);
```

### Key Fields:
- **id**: Internal chapter ID (format: `{manga_id}-ch-{chapter_number}`)
- **manga_id**: Foreign key to manga table (format: `md-{uuid}` for MangaDex)
- **source_chapter_id**: Original chapter ID from MangaDex (used for fetching pages)
- **source**: Source provider ("mangadex" or "mangaplus")

## API Endpoints

### Get Chapters
**GET** `/api/v1/manga/:id/chapters`

**Query Parameters**:
- `language`: Language filter (array, default: none)
- `limit`: Max chapters to return (default: 100)
- `offset`: Pagination offset (default: 0)

**Example**:
```
GET /api/v1/manga/md-abc123.../chapters?language=en&limit=50&offset=0
```

**Response**:
```json
{
  "chapters": [
    {
      "id": "chapter-uuid",
      "manga_id": "md-abc123...",
      "chapter_number": "1",
      "title": "Chapter Title",
      "volume_number": "1",
      "language": "en",
      "pages": 20,
      "source": "mangadex"
    }
  ],
  "total": 150,
  "limit": 50,
  "offset": 0
}
```

### Sync Chapters
**POST** `/api/v1/manga/sync-chapters`

Triggers a full sync of chapters for manga without chapters.

**Response**:
```json
{
  "success": true,
  "total_fetched": 100,
  "synced": 85,
  "skipped": 15,
  "failed": 0,
  "message": "Synced chapters for 85 manga"
}
```

## Sync Logic Flow

```
1. Fetch manga list from MangaDex (100 per batch)
   ↓
2. For each manga:
   ├─ Check if manga exists in DB
   ├─ Check chapter count for manga
   ├─ If manga exists AND has chapters → SKIP
   └─ Else:
      ├─ Fetch chapters from MangaDex (500 max)
      ├─ If no chapters found → SKIP
      ├─ If manga doesn't exist → Store manga
      └─ Store all chapters (INSERT OR IGNORE)
   ↓
3. Rate limit: 300ms between requests
   ↓
4. Continue until all manga processed or limit reached
```

## Performance Considerations

- **Rate Limiting**: 300ms delay between MangaDex API requests
- **Batch Size**: 100 manga per batch
- **Chapter Limit**: 500 chapters per manga (can be adjusted)
- **Duplicate Prevention**: Uses `INSERT OR IGNORE` to avoid duplicates
- **Database Queries**: Optimized to check existence before expensive operations

## Troubleshooting

### Chapters Still Not Showing
1. **Check server logs** for errors during sync
2. **Verify manga ID format**: Should be `md-{uuid}` for MangaDex manga
3. **Run manual sync**: `.\scripts\sync-chapters.ps1`
4. **Check database**: Query `manga_chapters` table directly

### Database Corruption
If SQLite shows corruption errors:
```powershell
# Stop server
# Delete database
Remove-Item mangahub.db

# Restart server (will trigger auto-sync)
.\scripts\start-server.ps1
```

### Slow Sync
- **Normal**: Syncing 1,634 manga takes ~8 minutes (300ms × 1,634 = 490s)
- **Solution**: Sync runs in background, server is still responsive
- **Monitoring**: Check logs for progress updates

## Benefits

✅ **Automatic Recovery**: Manga without chapters are auto-fixed on next sync  
✅ **No Data Loss**: Existing chapters are preserved (INSERT OR IGNORE)  
✅ **Efficient**: Only syncs what's needed (skips manga with chapters)  
✅ **Manual Control**: Can trigger sync anytime via API endpoint  
✅ **Better UX**: Chapters now display correctly in frontend  
✅ **Debug Friendly**: Detailed logging for troubleshooting  

## Next Steps

### Potential Enhancements:
1. **Progress Tracking**: Add WebSocket for real-time sync progress
2. **Selective Sync**: Sync specific manga by ID
3. **Chapter Updates**: Detect and sync new chapters for existing manga
4. **Background Jobs**: Use a job queue for large syncs
5. **Admin Panel**: UI for triggering and monitoring syncs
6. **Metrics**: Track sync success rates and performance
