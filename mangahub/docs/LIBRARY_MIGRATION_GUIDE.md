# Database Schema Migration Guide

## Overview

This migration restructures the database to separate **reading progress** from **library management**, enabling:

1. **TCP Progress Tracking for ANY manga** - Not just library items
2. **UDP Notifications on library additions** - Alert when users add manga to their collection
3. **Better data organization** - Clearer separation of concerns

## What Changes?

### Before (Old Schema)
```sql
user_progress:
  - user_id
  - manga_id
  - current_chapter
  - status           ← Mixed concern: both progress AND library
  - last_updated
```

**Problem**: User could only track progress for manga in their library. TCP progress updates only worked for library items.

### After (New Schema)

```sql
library:
  - user_id
  - manga_id
  - status           ← Library status: reading, completed, plan_to_read, etc.
  - added_at
  - last_updated

user_progress:
  - user_id
  - manga_id
  - current_chapter  ← Reading progress for ANY manga
  - last_read_at
```

**Benefits**:
- ✅ Track progress for manga you read but haven't added to library
- ✅ TCP broadcasts work for ALL reading activity
- ✅ UDP notifications trigger when adding to library
- ✅ Clearer data model: "What I'm reading" vs "What's in my collection"

## Migration Steps

### 1. Backup Your Database

The migration script automatically creates a backup, but you can also manually backup:

```powershell
Copy-Item .\data\mangahub.db .\data\mangahub_backup.db
```

### 2. Stop All Servers

**IMPORTANT**: Close all running servers before migration:

```powershell
.\scripts\stop-all-servers.ps1
```

Or manually close:
- API Server (port 8080)
- gRPC Server (port 9003)
- TCP Server (port 9001)
- UDP Server (port 8081)

### 3. Run Migration

```powershell
cd mangahub
.\scripts\migrate-library.ps1
```

The script will:
1. Create a timestamped backup
2. Create new `library` table
3. Migrate existing data:
   - Status info → `library` table
   - Progress info → `user_progress` table
4. Keep old data in `user_progress_backup` table

### 4. Restart Servers

```powershell
.\scripts\start-all-servers.ps1
```

## Migration Process Details

### Data Migration

```
Old user_progress row:
  user_id: "user123"
  manga_id: "manga456"
  current_chapter: 25
  status: "reading"
  last_updated: "2025-12-20"

Becomes:

library row:
  user_id: "user123"
  manga_id: "manga456"
  status: "reading"          ← Moved here
  added_at: "2025-12-20"
  last_updated: "2025-12-20"

user_progress row:
  user_id: "user123"
  manga_id: "manga456"
  current_chapter: 25         ← Kept here
  last_read_at: "2025-12-20"
```

### Safety Features

- ✅ Automatic backup before migration
- ✅ Original data saved in `user_progress_backup` table
- ✅ Transaction-based (all-or-nothing)
- ✅ Can restore from backup if needed

## How It Affects Features

### TCP Progress Sync

**Before**: Only worked if manga was in library
```go
// Old: Would fail if manga not in library
UpdateProgress(user, manga, chapter) 
  → Check if in library → Error if not
```

**After**: Works for ANY manga
```go
// New: Works for any manga, optionally updates library
UpdateProgress(user, manga, chapter, status?)
  → Always tracks progress
  → Optionally updates library status if provided
```

### UDP Notifications

**Before**: No specific trigger for library additions

**After**: Broadcasts when user adds manga to library
```go
AddToLibrary(user, manga, "reading")
  → Adds to library table
  → Triggers UDP notification: "📚 User added 'Title' to library"
```

### API Behavior

**UpdateProgress Endpoint** (`POST /api/v1/users/progress`):
```json
{
  "manga_id": "manga123",
  "current_chapter": 10,
  "status": "reading"  // Optional: updates library if provided
}
```

- Always updates `user_progress` (works for any manga)
- If `status` provided, also updates `library` table
- Triggers TCP broadcast for all updates

**AddToLibrary Endpoint** (`POST /api/v1/users/library`):
```json
{
  "manga_id": "manga123",
  "status": "reading"
}
```

- Adds/updates entry in `library` table
- Triggers UDP notification broadcast
- Doesn't affect progress (tracked separately)

## Rollback (If Needed)

If something goes wrong:

### Option 1: Restore from Auto-Backup

```powershell
# Find your backup
Get-ChildItem .\data\mangahub_backup_*.db

# Restore it
Copy-Item .\data\mangahub_backup_YYYYMMDD_HHMMSS.db .\data\mangahub.db -Force
```

### Option 2: Use Backup Table

The migration keeps your old data in `user_progress_backup`:

```sql
-- Drop new tables
DROP TABLE library;
DROP TABLE user_progress;

-- Restore from backup
ALTER TABLE user_progress_backup RENAME TO user_progress;
```

## Testing After Migration

### 1. Test Library Functionality

```powershell
# Add manga to library
curl -X POST http://localhost:8080/api/v1/users/library `
  -H "Authorization: Bearer YOUR_TOKEN" `
  -H "Content-Type: application/json" `
  -d '{"manga_id":"test-manga-1","status":"reading"}'

# Check UDP notification was triggered (check UDP server logs)
```

### 2. Test Progress Tracking (Non-Library Manga)

```powershell
# Update progress for manga NOT in library
curl -X POST http://localhost:8080/api/v1/users/progress `
  -H "Authorization: Bearer YOUR_TOKEN" `
  -H "Content-Type: application/json" `
  -d '{"manga_id":"random-manga-999","current_chapter":5}'

# Check TCP broadcast was triggered (check TCP server logs)
# This should work even though manga is not in library!
```

### 3. Test Combined Update

```powershell
# Update progress AND library status together
curl -X POST http://localhost:8080/api/v1/users/progress `
  -H "Authorization: Bearer YOUR_TOKEN" `
  -H "Content-Type: application/json" `
  -d '{"manga_id":"test-manga-1","current_chapter":15,"status":"completed"}'

# Updates both tables and triggers TCP broadcast
```

## Verification Queries

Check migration success with SQLite:

```bash
sqlite3 ./data/mangahub.db
```

```sql
-- Check library table structure
.schema library

-- Check user_progress table structure  
.schema user_progress

-- Count records in each table
SELECT COUNT(*) AS library_count FROM library;
SELECT COUNT(*) AS progress_count FROM user_progress;

-- Sample library entries
SELECT * FROM library LIMIT 5;

-- Sample progress entries
SELECT * FROM user_progress LIMIT 5;

-- Check backup exists
SELECT COUNT(*) FROM user_progress_backup;
```

## Common Issues

### Issue: "Database is locked"
**Solution**: Close all servers before running migration

### Issue: Migration failed mid-way
**Solution**: Restore from backup (see Rollback section)

### Issue: API server errors after migration
**Solution**: Make sure you're running the updated code with new schema

## Summary

This migration enables a more flexible system where:

- 📚 **Library** = Your collection (what manga you're tracking)
- 📖 **Progress** = What you're reading (any manga, library or not)
- 🔄 **TCP** = Broadcasts reading activity (works for everything)
- 📢 **UDP** = Notifies on collection changes (library additions)

The separation makes the system more powerful and flexible while maintaining all existing functionality.
