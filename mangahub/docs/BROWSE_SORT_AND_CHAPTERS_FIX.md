# Browse Sort and Chapter Display Fixes

## Issues Fixed

### 1. Browse Page Sorting Not Working
**Problem**: The sort dropdown on the Browse page was not actually sorting the manga. Changing the sort option had no effect on the displayed order.

**Root Cause**: 
- Frontend had sort UI and state (`sortBy`) that updated on selection
- But the `searchLocal` API call didn't include the sort parameter
- Backend `SearchManga` had hardcoded `ORDER BY title` that couldn't be changed

**Solution**:
1. **Backend Model** (`pkg/models/manga.go`):
   - Added `Sort` field to `MangaSearchRequest` struct

2. **Backend Handler** (`cmd/api-server/main.go`):
   - Modified `searchManga` handler to accept `sort` query parameter
   - Parse and pass sort to `MangaService.SearchManga`

3. **Backend Service** (`internal/manga/manga.go`):
   - Modified `SearchManga` to build dynamic ORDER BY clause based on `req.Sort`:
     - `title`: ORDER BY title ASC (alphabetical)
     - `chapters`: ORDER BY total_chapters DESC (most chapters first)
     - `year`: ORDER BY created_at DESC (newest first)
     - `popular`: ORDER BY total_chapters DESC (using chapters as popularity proxy)
     - Default: ORDER BY title ASC

4. **Frontend Service** (`client/web-react/src/services/mangaService.js`):
   - Updated `searchLocal` signature to accept `sort` parameter
   - Pass sort to backend API call as query param

5. **Frontend UI** (`client/web-react/src/pages/Browse.jsx`):
   - Modified `fetchData` to pass `currentSort` to `searchLocal`
   - Modified `handleSearch` to pass `currentSort` to `searchLocal`

### 2. Chapters Not Displaying on Manga Detail Page
**Problem**: Manga detail pages showed "No chapters found" message despite chapters being stored in the database.

**Root Cause**:
- The chapter fetch logic had a conditional check:
  ```javascript
  if (!idStr.startsWith('mal-') && !idStr.includes('mangadex') && !idStr.includes('mangaplus'))
  ```
- MangaDex manga from database have IDs like `"md-12345..."` (not "mangadex")
- The check would fail and return early without fetching chapters

**Solution** (`client/web-react/src/pages/MangaDetail.jsx`):
- Removed the restrictive ID format check
- Now attempts to fetch chapters for all manga regardless of ID format
- The backend `GetChapterList` checks database first, so it works for local manga

## Changes Summary

### Modified Files:
1. `pkg/models/manga.go` - Added Sort field to MangaSearchRequest
2. `cmd/api-server/main.go` - Accept and parse sort parameter
3. `internal/manga/manga.go` - Dynamic ORDER BY based on sort parameter
4. `client/web-react/src/services/mangaService.js` - Accept and pass sort parameter
5. `client/web-react/src/pages/Browse.jsx` - Pass sort to API calls
6. `client/web-react/src/pages/MangaDetail.jsx` - Remove restrictive ID check for chapters

## Testing

### Test Sort Functionality:
1. Navigate to Browse page
2. Select different sort options from dropdown:
   - **Most Popular** (popular) - Should sort by total_chapters DESC
   - **Title A-Z** (title) - Should sort alphabetically ASC
   - **Most Chapters** (chapters) - Should sort by total_chapters DESC
   - **Newest** (year) - Should sort by created_at DESC
3. Verify manga order changes with each selection

### Test Chapter Display:
1. Navigate to Browse page
2. Click on any manga (with ID like "md-...")
3. Scroll down to "Available Chapters" section
4. Verify chapters are displayed (no "No chapters found" message)
5. Click on a chapter to verify reading works

## Database Structure Reference

### Manga ID Formats:
- **MangaDex**: `md-{uuid}` (e.g., "md-abc123...")
- **MAL**: `mal-{id}` (e.g., "mal-12345")

### Relevant Tables:
- `manga`: Main manga table (id, title, author, genres, total_chapters, etc.)
- `manga_chapters`: Chapter metadata (id, manga_id, chapter_number, title, source_chapter_id)
- `manga_sources`: Source mappings (manga_id, source, source_id)

## API Endpoints

### Search Manga (GET /api/manga/)
**Query Parameters**:
- `query`: Search term (optional)
- `sort`: Sort order - "title", "chapters", "year", "popular" (optional, default: "title")
- `limit`: Results per page (optional, default: 20)
- `offset`: Pagination offset (optional, default: 0)
- `genres`: Comma-separated genre list (optional)
- `status`: Manga status filter (optional)
- `author`: Author filter (optional)

**Response**:
```json
{
  "manga": [...],
  "count": 1634,
  "limit": 25,
  "offset": 0
}
```

### Get Chapters (GET /api/manga/:id/chapters)
**Query Parameters**:
- `languages`: Comma-separated language codes (default: "en")
- `limit`: Max chapters to return (default: 100)
- `offset`: Pagination offset (default: 0)

**Response**:
```json
{
  "chapters": [
    {
      "id": "chapter-id",
      "chapter_number": "1",
      "title": "Chapter Title",
      "volume": "1",
      "language": "en",
      "pages": 20
    }
  ]
}
```

## Notes

- Sort uses `total_chapters` as a proxy for popularity (could be enhanced with ratings/views)
- Chapter fetching now works for all manga types (MangaDex, MAL, local)
- Backend checks database first before fetching from external sources
- All manga from MangaDex sync have format `md-{uuid}`
