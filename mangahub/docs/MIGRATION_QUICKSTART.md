# 🚀 Quick Start: TCP/UDP Schema Migration

## What This Fixes

- ✅ **TCP Progress Sync**: Now works for ANY manga you read (not just library items)
- ✅ **UDP Notifications**: Triggers when you add manga to your library
- ✅ **Better Architecture**: Separates "reading progress" from "library collection"

## Run the Migration (3 Steps)

### Step 1: Stop Servers
```powershell
cd mangahub
.\scripts\stop-all-servers.ps1
```

### Step 2: Migrate Database
```powershell
.\scripts\migrate-library.ps1
```

This will:
- ✅ Create automatic backup
- ✅ Restructure database (library + user_progress tables)
- ✅ Migrate all your existing data
- ✅ Keep backup of original data

### Step 3: Start Servers
```powershell
.\scripts\start-all-servers.ps1
```

## Test It Works

### 1. Open React App
```
http://localhost:3000/realtime-sync
```

### 2. Test TCP (Read Non-Library Manga)
```powershell
# Update progress for ANY manga (doesn't need to be in library!)
curl -X POST http://localhost:8080/api/v1/users/progress `
  -H "Authorization: Bearer YOUR_TOKEN" `
  -H "Content-Type: application/json" `
  -d '{
    "manga_id": "random-manga-999",
    "current_chapter": 5
  }'
```

**Expected**: See TCP progress update appear in real-time on the page! 🎉

### 3. Test UDP (Add to Library)
```powershell
# Add manga to library
curl -X POST http://localhost:8080/api/v1/users/library `
  -H "Authorization: Bearer YOUR_TOKEN" `
  -H "Content-Type: application/json" `
  -d '{
    "manga_id": "test-manga-1",
    "status": "reading"
  }'
```

**Expected**: See UDP notification appear: "📚 User added 'Title' to library" 🎉

## What Changed?

### Before
```
user_progress table:
  - Had both progress AND library status mixed together
  - Could only track progress for library manga
```

### After
```
library table:
  - Tracks which manga are in your collection
  - Has status: reading, completed, plan_to_read, etc.

user_progress table:
  - Tracks reading progress for ANY manga
  - Works even if manga not in library
```

## Why This Matters

### TCP Progress Sync
**Before**: Only worked if manga in library
**Now**: Works for ANY manga you read! 🎉

Example: Browse random manga → Read a few chapters → TCP broadcasts your progress
(even though it's not in your library yet)

### UDP Notifications
**Before**: No specific trigger
**Now**: Broadcasts when you add manga to library! 🎉

Example: Add "One Piece" to library → Everyone gets notified

## Documentation

- 📖 [Full Migration Guide](./LIBRARY_MIGRATION_GUIDE.md) - Detailed docs
- 📖 [Technical Summary](./TCP_UDP_SCHEMA_FIX.md) - What was done

## Rollback (If Needed)

Auto-backup created at: `data/mangahub_backup_YYYYMMDD_HHMMSS.db`

```powershell
# Restore from backup
Copy-Item .\data\mangahub_backup_*.db .\data\mangahub.db -Force
```

## Need Help?

Check the full documentation:
- [LIBRARY_MIGRATION_GUIDE.md](./LIBRARY_MIGRATION_GUIDE.md)
- [TCP_UDP_SCHEMA_FIX.md](./TCP_UDP_SCHEMA_FIX.md)
