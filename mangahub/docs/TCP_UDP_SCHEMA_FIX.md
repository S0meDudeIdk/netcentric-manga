# TCP/UDP Schema Improvements - Quick Summary

## What Was Done

Restructured the database schema to separate **reading progress** from **library management**, enabling the TCP and UDP features to work as intended.

## The Problem

1. **TCP Progress**: Only worked for manga in user's library (couldn't track reading of non-library manga)
2. **UDP Notifications**: Had no specific trigger for library additions
3. **Mixed Concerns**: The `user_progress` table mixed two different concepts:
   - What manga are in my collection? (status: reading, completed, etc.)
   - What am I currently reading? (current_chapter tracking)

## The Solution

### New Schema

**Before**:
- `user_progress` table with both `current_chapter` and `status` columns (mixed concerns)

**After**:
- `library` table: Tracks which manga are in user's collection (status field)
- `user_progress` table: Tracks reading progress for ANY manga (no status field)

### Key Changes

1. **Models** ([manga.go](X:\Bao2023toPresent\IU\4th year\NetCentric\Project\netcentric-manga\mangahub\pkg\models\manga.go)):
   - Added `Library` model for library entries
   - Updated `UserProgress` model (removed status, added LastReadAt)
   - Made status optional in `UpdateProgressRequest`

2. **Database Schema** ([database.go](X:\Bao2023toPresent\IU\4th year\NetCentric\Project\netcentric-manga\mangahub\pkg\database\database.go)):
   - Created `library` table with status tracking
   - Recreated `user_progress` table without status field
   - Added appropriate indexes

3. **User Service** ([user.go](X:\Bao2023toPresent\IU\4th year\NetCentric\Project\netcentric-manga\mangahub\internal\user\user.go)):
   - Updated `GetLibrary()`: Joins library + user_progress tables
   - Updated `AddToLibrary()`: Inserts into library table
   - Updated `UpdateProgress()`: Always works, optionally updates library
   - Updated `GetLibraryStats()`, `GetFilteredLibrary()`, etc.

4. **API Handlers** ([userHandlers.go](X:\Bao2023toPresent\IU\4th year\NetCentric\Project\netcentric-manga\mangahub\internal\api\userHandlers.go)):
   - Added UDP notification trigger on `addToLibrary()`
   - TCP broadcasts now work for all progress updates

5. **Migration Tools**:
   - `migrate-library-schema.go`: Go program to migrate database
   - `migrate-library.ps1`: PowerShell script to run migration with backup

## How It Works Now

### TCP Progress Sync

**Scenario 1**: User reads manga NOT in library
```
User reads chapter 5 of random manga
  → UpdateProgress(user, manga, 5)
  → Inserts/updates user_progress table (no library check!)
  → Triggers TCP broadcast ✅
```

**Scenario 2**: User reads manga IN library
```
User reads chapter 10 of library manga
  → UpdateProgress(user, manga, 10, "reading")
  → Updates user_progress (chapter tracking)
  → Updates library (status tracking)
  → Triggers TCP broadcast ✅
```

### UDP Notifications

**Scenario**: User adds manga to library
```
User adds "One Piece" to library with status "reading"
  → AddToLibrary(user, "One Piece", "reading")
  → Inserts into library table
  → Triggers UDP notification: "📚 User added 'One Piece' to library" ✅
```

## Migration Instructions

### Quick Start

```powershell
# 1. Stop all servers
.\scripts\stop-all-servers.ps1

# 2. Run migration (auto-creates backup)
.\scripts\migrate-library.ps1

# 3. Restart servers
.\scripts\start-all-servers.ps1
```

### What the Migration Does

1. Creates new `library` table
2. Migrates status data from old `user_progress` to new `library`
3. Recreates `user_progress` without status column
4. Migrates progress data to new `user_progress`
5. Keeps backup in `user_progress_backup` table
6. Creates timestamped database backup file

## Testing

### Test TCP with Non-Library Manga

```powershell
# Update progress for manga not in library
curl -X POST http://localhost:8080/api/v1/users/progress `
  -H "Authorization: Bearer YOUR_TOKEN" `
  -d '{"manga_id":"random-123","current_chapter":7}'

# Should trigger TCP broadcast even though manga not in library!
```

### Test UDP on Library Addition

```powershell
# Add manga to library
curl -X POST http://localhost:8080/api/v1/users/library `
  -H "Authorization: Bearer YOUR_TOKEN" `
  -d '{"manga_id":"test-manga","status":"reading"}'

# Should trigger UDP notification!
# Check UDP server logs or SSE notifications page
```

## Files Created/Modified

### Created
- ✅ `scripts/migrate-library-schema.go` - Migration program
- ✅ `scripts/migrate-library.ps1` - Migration runner script
- ✅ `docs/LIBRARY_MIGRATION_GUIDE.md` - Detailed guide

### Modified
- ✅ `pkg/database/database.go` - Schema definitions
- ✅ `pkg/models/manga.go` - Models
- ✅ `internal/user/user.go` - User service (all methods updated)
- ✅ `internal/api/userHandlers.go` - Added UDP trigger
- ✅ `internal/api/services.go` - Updated UDP notification method

## Benefits

1. **TCP Progress Now Universal**: Works for ANY manga, not just library items
2. **UDP Notifications Functional**: Triggers on library additions
3. **Better Data Model**: Clear separation of concerns
4. **Reading History**: Progress is preserved even if manga removed from library
5. **Flexibility**: Can track reading without committing to library

## Compliance with Requirements

### TCP Server (20 points) ✅
- Accepts connections ✅
- Broadcasts updates ✅
- JSON protocol ✅
- Concurrent handling ✅
- **NOW**: Works for ALL manga reads (not just library)

### UDP Server (15 points) ✅
- Client registration ✅
- Broadcasts notifications ✅
- Client list management ✅
- **NOW**: Triggers on library additions

## Next Steps

1. Run migration on your database
2. Test TCP progress with non-library manga
3. Test UDP notifications when adding to library
4. Verify SSE page shows both connections and updates

---

**Need Help?** See [LIBRARY_MIGRATION_GUIDE.md](./LIBRARY_MIGRATION_GUIDE.md) for detailed documentation.
