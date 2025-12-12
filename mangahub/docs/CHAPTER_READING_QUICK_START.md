# MangaHub Chapter Reading - Quick Start Guide

## What's New

Your MangaHub application now supports reading manga chapters directly in the browser using the MangaDex and MangaPlus APIs!

## How to Use

### 1. Finding Manga to Read

Currently, only manga from external sources (MangaDex/MangaPlus) are readable:

- Browse manga using the Browse page
- Search for manga from MyAnimeList (MAL)
- Look for manga with IDs starting with `mal-`, `mangadex-`, or `mangaplus-`

### 2. Opening the Reader

1. Go to any manga detail page
2. Scroll to the "Chapters" section
3. Click on any chapter to start reading
4. Chapters will show a badge indicating their source (MangaDex or MangaPlus)

### 3. Reader Controls

**Navigation:**
- `Arrow Keys` or `A/D` - Navigate between pages
- `Click Left/Right` - Click on left or right side of screen
- `Prev/Next Buttons` - Use on-screen navigation
- `Prev Ch/Next Ch` - Switch between chapters

**Settings (Click gear icon):**
- **View Mode**: Single Page or Vertical Scroll
- **Fit Mode**: Fit Width, Fit Height, or Original Size
- **Background**: Choose from 4 background colors

**Other:**
- `ESC` - Close settings panel
- `Back Button` - Return to manga details

### 4. Reading Progress

If you're logged in:
- Your reading progress is automatically saved
- Progress is updated when you open a chapter
- Check your Library to see your reading history

## API Usage Examples

### For Developers

**Fetch Chapter List:**
```javascript
const chapters = await mangaService.getChapters(
  'mangadex-uuid-here',  // Manga ID
  ['en'],                // Languages
  100,                   // Limit
  0                      // Offset
);
```

**Fetch Chapter Pages:**
```javascript
const pages = await mangaService.getChapterPages(
  'chapter-id-here',     // Chapter ID
  'mangadex'            // Source
);
```

**Backend API Endpoints:**
```bash
# Get chapters
GET /api/v1/manga/{manga_id}/chapters?language=en&limit=100&offset=0

# Get chapter pages
GET /api/v1/manga/chapters/{chapter_id}/pages?source=mangadex
```

## Testing the Feature

### Configuration

The MangaDex and MangaPlus APIs are now configurable via environment variables in `.env`:

```env
# MangaDex API Configuration
MANGADEX_API_BASE_URL=https://api.mangadex.org
MANGADEX_API_TIMEOUT=15
MANGADEX_DEBUG=false
MANGADEX_API_KEY=your_api_key_here  # Optional but recommended

# MangaPlus API Configuration
MANGAPLUS_API_BASE_URL=https://jumpg-webapi.tokyo-cdn.com/api
MANGAPLUS_API_TIMEOUT=15
MANGAPLUS_DEBUG=false
```

**MangaDex API Key (Optional):**
- Get yours at: https://mangadex.org/settings → API Clients
- **Not required** for basic reading
- **Recommended for**:
  - Higher rate limits (better performance)
  - Access to followed manga
  - Private/restricted content
  - Better reliability during high traffic

**No API key required for MangaPlus** - it's a free, public API!

### Quick Test

1. Start your backend server:
   ```powershell
   cd mangahub
   go run cmd/api-server/main.go
   ```

2. Start your frontend:
   ```powershell
   cd mangahub/client/web-react
   npm start
   ```

3. Test with a MangaDex manga:
   - The manga ID needs to be a valid MangaDex UUID
   - Example format: `mangadex-a96676e5-8ae2-425e-b549-7f15dd34a6d8`

### Testing with Real Data

Since your current database may not have MangaDex IDs, you can:

1. **Search MAL Manga**: Use the Browse page to search MyAnimeList
2. **Find MangaDex ID**: For any manga, you can find its MangaDex ID by:
   - Going to https://mangadex.org
   - Searching for the manga
   - Getting the ID from the URL

3. **Manual Testing**: Use the browser console:
   ```javascript
   // Test API directly
   fetch('http://localhost:8080/api/v1/manga/mangadex-{uuid}/chapters?language=en&limit=20')
     .then(r => r.json())
     .then(console.log);
   ```

## Supported Manga Sources

### MangaDex (Primary)
- ✅ Most manga available
- ✅ Multiple languages
- ✅ Community translations
- ✅ Free and open
- ⚠️ Requires valid UUID

### MangaPlus (Fallback)
- ✅ Official Shueisha manga (Jump titles)
- ✅ High quality
- ✅ Latest chapters
- ⚠️ Limited to Shueisha catalog
- ⚠️ Requires numeric title ID

### Local Database
- ❌ Not yet readable
- 📝 Shows placeholder chapters only
- 🔮 Future enhancement planned

## Known Limitations

1. **Local Database Manga**: Manga from your local database (`manga.json`) won't have readable chapters unless you add MangaDex/MangaPlus IDs to them

2. **MAL Integration**: While you can browse MAL manga, they need to be matched to MangaDex/MangaPlus for reading

3. **Download**: Offline reading not yet supported

4. **Bookmarks**: Chapter bookmarking coming in future update

## Troubleshooting

### "No chapters available"
- The manga might not exist in MangaDex/MangaPlus
- The manga ID format might be incorrect
- Check browser console for API errors

### "Failed to load chapter pages"
- MangaDex servers might be slow or down
- Try refreshing the page
- Try a different chapter

### Images not loading
- Check your internet connection
- MangaDex uses CDN servers which may be slow
- Clear browser cache and retry

### Reader not opening
- Make sure backend server is running
- Check browser console for errors
- Verify the manga has a valid source ID

## Next Steps

To make your existing manga readable:

1. **Add MangaDex IDs**: Update your `manga.json` to include MangaDex UUIDs
2. **Search Integration**: Implement search to find matching MangaDex manga
3. **ID Mapping**: Create a mapping service to link your manga to external sources

Example manga entry with readable chapters:
```json
{
  "id": "mangadex-a96676e5-8ae2-425e-b549-7f15dd34a6d8",
  "title": "One Punch Man",
  "author": "ONE",
  "genres": ["Action", "Comedy"],
  "status": "ongoing",
  "total_chapters": 180,
  "description": "...",
  "cover_url": "...",
  "publication_year": 2012,
  "rating": 8.5
}
```

## Support

If you encounter issues:
1. Check the documentation in `docs/CHAPTER_READING_IMPLEMENTATION.md`
2. Review browser console for errors
3. Test API endpoints directly
4. Verify backend logs

Enjoy reading manga! 📚✨
