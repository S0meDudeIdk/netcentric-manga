# Automatic MangaDex Integration

## Overview

The MangaHub system now automatically searches and fetches chapters from MangaDex when users try to read manga that were originally fetched from MyAnimeList (MAL) API.

## How It Works

### Architecture

1. **Metadata Source**: MAL/Jikan API
   - Used for browsing, searching, and displaying manga information
   - Provides title, description, genres, cover images, ratings, etc.

2. **Chapter Source**: MangaDex API
   - Automatically used for reading chapters
   - Searched by manga title when manga ID doesn't have `mangadex-` or `mangaplus-` prefix

### Flow Diagram

```
User clicks "Read Chapter"
         ↓
Check manga ID format
         ↓
    ┌────────────────────────┐
    │ Has MangaDex prefix?   │
    │ (mangadex-UUID)        │
    └────────┬───────────────┘
             │
        No   │   Yes
    ┌────────┴────────┐
    ↓                 ↓
Get manga title    Use MangaDex ID
from database      directly
    ↓                 ↓
Search MangaDex      │
by title             │
    ↓                 ↓
Use first/best    ───┘
match                 
    ↓
Fetch chapters from MangaDex
    ↓
Display in reader
```

## API Endpoints

### 1. Get Chapter List
```http
GET /api/v1/manga/:id/chapters
```

**Parameters:**
- `id` - Manga ID (can be MAL ID, MangaDex UUID, or prefixed ID)
- `language[]` (optional) - Array of language codes (default: ["en"])
- `limit` (optional) - Number of chapters to fetch (default: 100)
- `offset` (optional) - Pagination offset (default: 0)

**Behavior:**
- If `id` starts with `mangadex-` or is a valid UUID → Direct MangaDex fetch
- If `id` is a MAL ID → Search MangaDex by manga title, use best match
- If `id` starts with `mangaplus-` → Fetch from MangaPlus

**Response:**
```json
{
  "chapters": [
    {
      "id": "chapter-uuid",
      "manga_id": "mal-id",
      "chapter_number": "1",
      "volume_number": "1",
      "title": "Chapter Title",
      "language": "en",
      "pages": 42,
      "published_at": "2024-01-15",
      "source": "mangadex"
    }
  ],
  "total": 150,
  "limit": 100,
  "offset": 0
}
```

### 2. Search MangaDex
```http
GET /api/v1/manga/mangadex/search?title=Doraemon
```

**Parameters:**
- `title` (required) - Search query
- `limit` (optional) - Number of results (default: 10, max: 100)

**Response:**
```json
{
  "results": [
    {
      "id": "mangadex-uuid",
      "title": "Doraemon",
      "description": "A robotic cat from the future...",
      "status": "completed",
      "year": 1969,
      "genres": ["Comedy", "Sci-Fi", "Slice of Life"]
    }
  ],
  "total": 5,
  "limit": 10,
  "offset": 0
}
```

### 3. Get Chapter Pages
```http
GET /api/v1/manga/chapters/:chapter_id/pages?source=mangadex
```

**Parameters:**
- `chapter_id` (required) - Chapter UUID from MangaDex
- `source` (optional) - "mangadex" or "mangaplus" (default: "mangadex")

**Response:**
```json
{
  "chapter_id": "chapter-uuid",
  "manga_id": "manga-uuid",
  "chapter_number": "1",
  "pages": [
    "https://uploads.mangadex.org/data/hash/page1.jpg",
    "https://uploads.mangadex.org/data/hash/page2.jpg"
  ],
  "source": "mangadex",
  "base_url": "https://uploads.mangadex.org",
  "hash": "chapter-hash"
}
```

## Backend Implementation

### ChapterService Changes

```go
// SetMangaService allows chapter service to look up manga metadata
func (s *ChapterService) SetMangaService(mangaService *Service) {
    s.mangaService = mangaService
}

// GetChapterList now automatically searches MangaDex
func (s *ChapterService) GetChapterList(mangaID string, languages []string, limit, offset int) (*models.ChapterListResponse, error) {
    // Check if it's already a MangaDex/MangaPlus ID
    if strings.HasPrefix(mangaID, "mangadex-") || isMangaDexUUID(mangaID) {
        return s.getMangaDexChapters(mangaID, languages, limit, offset)
    }
    
    // For MAL IDs, get manga title and search MangaDex
    if s.mangaService != nil {
        manga, err := s.mangaService.GetMangaByID(mangaID)
        if err == nil && manga != nil {
            searchResults, err := s.mangaDexClient.SearchManga(manga.Title, 5)
            if err == nil && len(searchResults.Data) > 0 {
                // Use most relevant result
                mdManga := searchResults.Data[0]
                return s.getMangaDexChapters(mdManga.ID, languages, limit, offset)
            }
        }
    }
    
    // Fallback attempts
    return s.tryFallbackSources(mangaID, languages, limit, offset)
}
```

### MangaDex Client Addition

```go
// SearchManga searches for manga on MangaDex by title
func (c *MangaDexClient) SearchManga(title string, limit int) (*MangaDexMangaResponse, error) {
    params := url.Values{}
    params.Add("title", title)
    params.Add("limit", fmt.Sprintf("%d", limit))
    params.Add("order[relevance]", "desc")
    
    // Make API request and return results
}
```

## Frontend Changes

### User Experience

1. **Chapter List Display**
   - Shows info notice for MAL-sourced manga: "Chapters sourced from MangaDex"
   - Explains automatic search behavior
   - Real MangaDex chapters show source badge

2. **Read Button Behavior**
   - MAL manga: Triggers automatic MangaDex search
   - MangaDex manga: Direct chapter navigation
   - Clear messaging about chapter availability

### Implementation

```jsx
// Info Notice Component
{!useRealChapters && !loadingChapters && (
  <div className="mb-4 p-4 bg-blue-50 dark:bg-blue-900/20">
    <BookOpen className="w-5 h-5 text-blue-600" />
    <p className="font-semibold">Chapters sourced from MangaDex</p>
    <p>We'll automatically search and fetch chapters from MangaDex.</p>
  </div>
)}

// Updated click handler
const handleChapterClick = (chapter) => {
  if (useRealChapters && chapter.id) {
    navigate(`/read/${id}?chapter=${chapter.id}&source=${chapter.source}`);
  } else {
    alert('We will search MangaDex automatically when you click "Read"');
  }
};
```

## Configuration

### Environment Variables

```env
# MangaDex Configuration
MANGADEX_API_BASE_URL=https://api.mangadex.org
MANGADEX_API_TIMEOUT=15
MANGADEX_DEBUG=false
MANGADEX_API_KEY=  # Optional - for higher rate limits
```

### Rate Limiting

- **Without API Key**: 5 requests/second, burst up to 40
- **With API Key**: Higher limits for authenticated requests
- Search operations are cached where possible

## Search Matching Strategy

The system uses MangaDex's relevance-based search:

1. **Exact Title Match**: Highest priority
2. **Alternate Titles**: Japanese, romaji, other languages
3. **Partial Matches**: Substring matching
4. **Fuzzy Matching**: MangaDex's built-in fuzzy search

**Example Searches:**
- "Doraemon" → Finds "Doraemon", "ドラえもん"
- "One Piece" → Finds "One Piece", "ワンピース"
- "Attack on Titan" → Finds "Shingeki no Kyojin", "進撃の巨人"

## Error Handling

### Common Scenarios

1. **No MangaDex Results Found**
   ```
   User sees: "No chapters found on MangaDex for this manga"
   ```

2. **MangaDex API Error**
   ```
   User sees: "Failed to fetch chapters. Please try again later"
   Backend logs: Full error details for debugging
   ```

3. **Rate Limit Exceeded**
   ```
   User sees: "Too many requests. Please wait a moment"
   Backend: Automatic retry with backoff
   ```

## Benefits

### For Users
- ✅ Seamless experience - no manual linking required
- ✅ Browse using comprehensive MAL database
- ✅ Read using MangaDex's extensive chapter library
- ✅ Automatic fallback to best available source

### For Developers
- ✅ Clean separation of concerns (metadata vs content)
- ✅ Leverages strengths of both APIs
- ✅ Extensible to additional sources
- ✅ Minimal maintenance required

## Future Enhancements

1. **Manual Linking**
   - Allow users to manually select correct MangaDex manga if auto-search fails
   - Store user-confirmed links to improve future searches

2. **Caching**
   - Cache MangaDex search results per manga title
   - Reduce API calls for popular manga

3. **Multiple Source Support**
   - Try MangaPlus if MangaDex fails
   - Support other legal manga sources

4. **Smart Matching**
   - Use MAL ID mapping databases
   - Machine learning for better title matching
   - Consider author names, publication years for disambiguation

## Troubleshooting

### Chapter Not Loading

**Problem**: "No chapters found" message appears

**Solutions:**
1. Check manga title accuracy in MAL
2. Try searching MangaDex manually to verify availability
3. Check MangaDex API status
4. Review backend logs for search query used

### Wrong Manga Matched

**Problem**: Chapters from different manga appear

**Solutions:**
1. Implement manual linking feature (future)
2. Report issue with manga title and expected MangaDex ID
3. Check for similar manga titles causing confusion

### Rate Limiting

**Problem**: "Too many requests" errors

**Solutions:**
1. Add MangaDex API key to environment variables
2. Implement request caching
3. Add exponential backoff retry logic

## Testing

### Manual Testing

```bash
# Test MAL manga chapter fetch (auto-search)
curl "http://localhost:8080/api/v1/manga/101/chapters"

# Test direct MangaDex chapter fetch
curl "http://localhost:8080/api/v1/manga/mangadex-uuid/chapters"

# Test MangaDex search
curl "http://localhost:8080/api/v1/manga/mangadex/search?title=Doraemon&limit=5"

# Test chapter pages
curl "http://localhost:8080/api/v1/manga/chapters/chapter-uuid/pages?source=mangadex"
```

### Automated Testing

```go
func TestAutoMangaDexSearch(t *testing.T) {
    service := NewChapterService()
    service.SetMangaService(mockMangaService)
    
    // Test with MAL ID
    chapters, err := service.GetChapterList("101", []string{"en"}, 10, 0)
    assert.NoError(t, err)
    assert.NotEmpty(t, chapters.Chapters)
    assert.Equal(t, "mangadex", chapters.Chapters[0].Source)
}
```

## References

- [MangaDex API Documentation](https://api.mangadex.org/docs/)
- [MyAnimeList API](https://myanimelist.net/apiconfig/references/api/v2)
- [Jikan API](https://docs.api.jikan.moe/)
