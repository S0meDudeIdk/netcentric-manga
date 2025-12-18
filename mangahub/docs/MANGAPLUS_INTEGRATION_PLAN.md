# MangaPlus Integration Plan

## Current Status

Currently, the system uses:
- **MAL/Jikan API**: For manga metadata (titles, descriptions, covers, ratings)
- **MangaDex API**: For reading chapters (primary source)
- **MangaPlus API**: Partially implemented but needs enhancement

## Problem Statement

Some manga available on MangaDex don't return chapters when searched by title. This happens because:
1. Title mismatch between MAL and MangaDex
2. Different romanization (e.g., "Berserk" vs "ベルセルク")
3. Manga not uploaded to MangaDex yet
4. Regional availability restrictions

## Solution: Enhanced Search Strategy

### Current Implementation

The system now uses an improved MangaDex search with multiple strategies:

```go
// Strategy 1: Exact title match
// Strategy 2: Case-insensitive contains
// Strategy 3: First result (most relevant by MangaDex)
// Strategy 4: Try with cleaned title (remove suffixes)
```

### When Chapters Are Not Found

The system now returns an empty chapter list instead of an error, allowing the frontend to display:
- Amber notice: "No chapters found - This manga is not currently available"
- Green notice: "Chapters available from MangaDex - Found X chapters"

## Future Enhancement: MangaPlus Fallback

### Reference Implementation

The `mangoplus` repository (https://github.com/luevano/mangoplus) provides:
- Python-based MangaPlus API client
- Chapter listing and page fetching
- Image URL generation

### Integration Steps

#### 1. Analyze MangaPlus API Structure

From the mangoplus repository, we need to understand:
- API endpoints structure
- Authentication requirements (if any)
- Rate limiting
- Response formats

#### 2. Create Go MangaPlus Client

Enhance `internal/external/mangaplus.go`:

```go
type MangaPlusClient struct {
    BaseURL    string
    HTTPClient *http.Client
    UserAgent  string
}

// Methods to implement:
// - SearchManga(title string) - Search by title
// - GetTitleByID(id int) - Get manga details
// - GetChapterList(titleID int) - Get chapters
// - GetChapterPages(chapterID int) - Get page URLs
```

#### 3. Update ChapterService Logic

Modify `internal/manga/chapter_service.go`:

```go
func (s *ChapterService) GetChapterList(mangaID string, languages []string, limit, offset int) (*models.ChapterListResponse, error) {
    // 1. Try MangaDex first (existing implementation)
    mdChapters, mdErr := s.tryMangaDex(mangaID, languages, limit, offset)
    if mdErr == nil && len(mdChapters.Chapters) > 0 {
        return mdChapters, nil
    }
    
    // 2. Fallback to MangaPlus
    mpChapters, mpErr := s.tryMangaPlus(mangaID)
    if mpErr == nil && len(mpChapters.Chapters) > 0 {
        return mpChapters, nil
    }
    
    // 3. Return empty list if both fail
    return &models.ChapterListResponse{
        Chapters: []models.ChapterInfo{},
        Total:    0,
    }, nil
}
```

#### 4. MangaPlus Title Matching

MangaPlus uses numeric IDs, so we need a mapping strategy:

**Option A: Title-based Search**
```go
func (s *ChapterService) tryMangaPlus(mangaID string) (*models.ChapterListResponse, error) {
    // Get manga title from MAL
    title := s.getMangaTitleFromMAL(mangaID)
    
    // Search MangaPlus by title
    mpResults, err := s.mangaPlusClient.SearchByTitle(title)
    if err != nil || len(mpResults) == 0 {
        return nil, fmt.Errorf("not found on MangaPlus")
    }
    
    // Use first result
    titleID := mpResults[0].ID
    return s.getMangaPlusChapters(titleID)
}
```

**Option B: Manual Mapping**
Create a mapping file `data/mangaplus_mappings.json`:
```json
{
  "mal-2": "mangaplus-100020",
  "mal-21": "mangaplus-100028"
}
```

#### 5. Update Frontend

Add MangaPlus source indicators:

```jsx
{chapter.source === 'mangadex' && (
  <span className="text-xs px-2 py-1 bg-blue-500 text-white rounded">
    MangaDex
  </span>
)}
{chapter.source === 'mangaplus' && (
  <span className="text-xs px-2 py-1 bg-orange-500 text-white rounded">
    MangaPlus Official
  </span>
)}
```

## Implementation Priority

### Phase 1: Current (✅ Completed)
- ✅ Improved MangaDex search with multiple strategies
- ✅ Better error handling (return empty list instead of error)
- ✅ Enhanced UI feedback (amber/green notices)

### Phase 2: Optional Enhancement
- ⏳ Analyze mangoplus repository API structure
- ⏳ Implement MangaPlus search by title
- ⏳ Add MangaPlus as fallback source
- ⏳ Create mapping system for popular manga
- ⏳ Add source indicators in UI

### Phase 3: Advanced Features
- ⏳ Support for multiple language scanlations
- ⏳ Chapter download queue
- ⏳ Offline reading support
- ⏳ User preference for preferred source

## Testing Strategy

### Test Cases

1. **Manga available on MangaDex only**
   - Should fetch from MangaDex
   - Green notice displayed
   - Chapters load correctly

2. **Manga available on MangaPlus only**
   - Should fetch from MangaPlus (after implementation)
   - Orange "Official" badge shown
   - Chapters load correctly

3. **Manga available on both**
   - Should prefer MangaDex (faster, more scanlations)
   - Can manually switch sources (future feature)

4. **Manga not available anywhere**
   - Should show amber notice
   - No chapter list displayed
   - No error in console

### Manual Testing

```bash
# Test with manga known to be on MangaDex
curl "http://localhost:8080/api/v1/manga/mal-2/chapters"

# Test with manga only on MangaPlus (e.g., One Piece, Naruto)
curl "http://localhost:8080/api/v1/manga/mal-13/chapters"

# Test with obscure manga not on either
curl "http://localhost:8080/api/v1/manga/mal-99999/chapters"
```

## MangaPlus API Research Needed

Before full implementation, research these aspects from the mangoplus repo:

1. **API Endpoints**
   - Search endpoint URL and parameters
   - Title details endpoint
   - Chapter list endpoint
   - Page/image endpoints

2. **Authentication**
   - Does MangaPlus require authentication?
   - API keys needed?
   - Rate limits and quotas

3. **Response Format**
   - JSON structure for search results
   - Chapter metadata format
   - Image URL generation logic

4. **Limitations**
   - Regional restrictions (CORS, geo-blocking)
   - Available manga (only Shueisha titles?)
   - Language support

5. **Image Handling**
   - Are images scrambled/encrypted?
   - Need for descrambling algorithm?
   - CORS headers for image loading

## Alternative: Use Existing MangaPlus Client

If the mangoplus Python client is stable, we could:
1. Run it as a microservice
2. Create HTTP wrapper around it
3. Call it from our Go backend

This would avoid reimplementing the entire client in Go.

## Conclusion

The current implementation provides a solid foundation with improved MangaDex search. MangaPlus integration is optional and can be added later when needed for specific popular manga that aren't available on MangaDex.

For now, the system will:
- ✅ Try MangaDex with improved search
- ✅ Show clear feedback when chapters are/aren't available
- ✅ Gracefully handle missing chapters without errors

MangaPlus can be added in Phase 2 when there's a clear need for specific manga.
