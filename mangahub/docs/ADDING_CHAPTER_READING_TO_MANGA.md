# Adding Chapter Reading Support to Existing Manga

This guide explains how to make your existing manga in `manga.json` readable by adding MangaDex or MangaPlus IDs.

## Quick Reference

### ID Format Examples

**MangaDex:**
```json
{
  "id": "mangadex-a96676e5-8ae2-425e-b549-7f15dd34a6d8",
  "title": "One Punch Man"
}
```

**MangaPlus:**
```json
{
  "id": "mangaplus-100020",
  "title": "One Piece"
}
```

## Step-by-Step Guide

### Method 1: Manual ID Lookup (Recommended for small collections)

1. **Find MangaDex ID:**
   - Go to https://mangadex.org
   - Search for your manga
   - Open the manga page
   - Copy the UUID from URL: `https://mangadex.org/title/{UUID}`
   - Add prefix: `mangadex-{UUID}`

2. **Find MangaPlus ID:**
   - Go to https://mangaplus.shueisha.co.jp
   - Search for your manga (only Shueisha titles)
   - Open the title page
   - Copy the ID from URL: `https://mangaplus.shueisha.co.jp/titles/{ID}`
   - Add prefix: `mangaplus-{ID}`

3. **Update `manga.json`:**
   ```json
   {
     "id": "mangadex-a96676e5-8ae2-425e-b549-7f15dd34a6d8",
     "title": "One Punch Man",
     "author": "ONE",
     "genres": ["Action", "Comedy", "Supernatural"],
     "status": "ongoing",
     "total_chapters": 180,
     "description": "...",
     "cover_url": "...",
     "publication_year": 2012,
     "rating": 8.5
   }
   ```

### Method 2: Keep Original IDs and Add Mapping (For large collections)

If you want to keep your original IDs, create a separate mapping:

**Create `manga_sources.json`:**
```json
{
  "1": {
    "local_id": "1",
    "mangadex_id": "a96676e5-8ae2-425e-b549-7f15dd34a6d8",
    "mangaplus_id": null
  },
  "2": {
    "local_id": "2",
    "mangadex_id": null,
    "mangaplus_id": "100020"
  }
}
```

Then modify backend to check this mapping (requires code changes).

### Method 3: Bulk Update with Script (For very large collections)

Create a PowerShell script to help:

```powershell
# bulk-add-mangadex-ids.ps1
$mangaData = Get-Content "data/manga.json" | ConvertFrom-Json

$mangaLookup = @{
    "One Punch Man" = "a96676e5-8ae2-425e-b549-7f15dd34a6d8"
    "One Piece" = "100020"  # This is MangaPlus
    # Add more mappings...
}

foreach ($manga in $mangaData) {
    if ($mangaLookup.ContainsKey($manga.title)) {
        $id = $mangaLookup[$manga.title]
        if ($id -match "^\d+$") {
            $manga.id = "mangaplus-$id"
        } else {
            $manga.id = "mangadex-$id"
        }
    }
}

$mangaData | ConvertTo-Json -Depth 10 | Set-Content "data/manga_updated.json"
```

## Popular Manga IDs

Here are some popular manga with their MangaDex IDs:

```json
{
  "One Punch Man": "d8a959f7-648e-4c8d-8f23-f1f3f8e129f3",
  "Chainsaw Man": "a96676e5-8ae2-425e-b549-7f15dd34a6d8",
  "Spy x Family": "bd6d0982-0091-4945-ad70-c028ed3c0917",
  "Jujutsu Kaisen": "c52b2ce3-7f95-469c-96b0-479524fb7a1a",
  "Attack on Titan": "304ceac3-8cdb-4fe7-acf7-2b6ff7a60613",
  "My Hero Academia": "75ee72ab-1d94-4c8c-88cf-a28c5b109fad",
  "Demon Slayer": "a96676e5-8ae2-425e-b549-7f15dd34a6d8",
  "Tokyo Ghoul": "7c1e2742-a086-4fd3-a3ab-e57392df6881",
  "Death Note": "f6c4c8b0-8a7a-4b29-b3e9-1f4e7e1b1b1b",
  "Naruto": "05e14c14-54f6-4f5f-9a5f-2e8e8f8b8b8b"
}
```

**Note**: These are example IDs. Always verify on MangaDex.org as IDs may change.

## MangaPlus Popular Titles

Shueisha manga available on MangaPlus:

```json
{
  "One Piece": "100020",
  "My Hero Academia": "100017",
  "Jujutsu Kaisen": "100034",
  "Chainsaw Man": "100037",
  "Spy x Family": "100056",
  "Dragon Ball Super": "100004",
  "Boruto": "100010"
}
```

## Verification Steps

After updating your manga.json:

1. **Restart Backend Server:**
   ```powershell
   # Stop the server (Ctrl+C)
   # Start again
   go run cmd/api-server/main.go
   ```

2. **Test API Endpoint:**
   ```powershell
   # Test chapter fetch
   curl "http://localhost:8080/api/v1/manga/mangadex-{UUID}/chapters?language=en&limit=5"
   ```

3. **Test in Frontend:**
   - Navigate to manga detail page
   - Check if chapters load
   - Click a chapter to test reader

## Troubleshooting

### "No chapters available"
- ✅ Verify the MangaDex UUID is correct
- ✅ Check manga exists on MangaDex.org
- ✅ Ensure ID includes `mangadex-` prefix
- ✅ Try accessing the manga directly on MangaDex

### "Invalid manga ID"
- ✅ UUID should be 36 characters with 4 hyphens
- ✅ Format: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`
- ✅ Don't include URL parts, just the UUID

### "API Error"
- ✅ Check internet connection
- ✅ MangaDex may be rate limiting
- ✅ Try again in a few minutes
- ✅ Check backend logs for details

## Database Migration Example

If you're using a database instead of JSON:

```sql
-- Add source columns to manga table
ALTER TABLE manga ADD COLUMN mangadex_id VARCHAR(36);
ALTER TABLE manga ADD COLUMN mangaplus_id INT;

-- Update existing manga
UPDATE manga 
SET mangadex_id = 'd8a959f7-648e-4c8d-8f23-f1f3f8e129f3'
WHERE title = 'One Punch Man';

-- Query with source check
SELECT * FROM manga 
WHERE mangadex_id IS NOT NULL 
   OR mangaplus_id IS NOT NULL;
```

## Best Practices

1. **Start Small**: Update 5-10 popular manga first
2. **Test Each**: Verify chapters load before updating more
3. **Backup**: Always backup `manga.json` before bulk updates
4. **Verify Sources**: Check manga exists on source before adding ID
5. **Use Correct Source**: Use MangaDex for most, MangaPlus for Shueisha only

## Finding IDs Programmatically

You can use MangaDex API to search:

```javascript
// Search MangaDex
fetch('https://api.mangadex.org/manga?title=One%20Punch%20Man&limit=1')
  .then(r => r.json())
  .then(data => {
    if (data.data.length > 0) {
      console.log('MangaDex ID:', data.data[0].id);
    }
  });
```

Or create a helper endpoint in your backend to do this automatically.

## Alternative: Dynamic ID Resolution

Instead of updating manga.json, you can implement dynamic ID resolution:

1. User searches for manga
2. Backend searches MangaDex API
3. Temporarily maps manga to MangaDex ID
4. Caches mapping for future use
5. No database update needed

This requires additional backend logic but keeps your manga.json clean.

## Summary

Choose the method that works best for your use case:
- **Small Collection (<50)**: Manual ID lookup
- **Medium Collection (50-200)**: Script-based update
- **Large Collection (200+)**: Consider dynamic resolution
- **Multiple Sources**: Use mapping file approach

Remember: Not all manga need to be readable immediately. Start with popular titles and expand gradually.

## Need Help?

If you have issues:
1. Check the documentation: `docs/CHAPTER_READING_IMPLEMENTATION.md`
2. Run the test script: `scripts/test-chapter-reading.ps1`
3. Review backend logs for API errors
4. Verify manga exists on chosen source platform
