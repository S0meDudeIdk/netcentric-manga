# Rating System & MangaPlus Fallback Implementation

## Overview
This document describes the implementation of two major features:
1. **Custom Rating System**: User-based manga ratings (0-10 scale)
2. **MangaPlus Fallback**: Automatic handling of licensed manga with external URLs

## Feature 1: Custom Rating System

### Backend Implementation

#### Database Schema
```sql
CREATE TABLE manga_ratings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    manga_id TEXT NOT NULL,
    rating INTEGER NOT NULL CHECK(rating >= 0 AND rating <= 10),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, manga_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
)
```

#### Models
- `MangaRating`: Represents a single user rating
- `MangaRatingStats`: Aggregated rating statistics (average, count, user's rating)
- `RateMangaRequest`: Request payload for rating submission

#### API Endpoints

**Rate a Manga** (Protected)
```
POST /api/v1/manga/:manga_id/ratings
Authorization: Bearer <token>
Body: { "rating": 7 }
Response: MangaRatingStats
```

**Get Rating Stats** (Public)
```
GET /api/v1/manga/:id/ratings
Optional Auth: Bearer <token> (to include user's rating)
Response: {
  "manga_id": "mal-1",
  "average_rating": 8.5,
  "total_ratings": 42,
  "user_rating": 9  // Only if authenticated
}
```

**Delete Rating** (Protected)
```
DELETE /api/v1/manga/:manga_id/ratings
Authorization: Bearer <token>
Response: { "message": "Rating deleted successfully" }
```

#### Service Layer
`RatingService` (`internal/manga/rating_service.go`):
- `RateManga()`: Add/update user rating
- `GetUserRating()`: Get specific user's rating
- `GetMangaRatingStats()`: Get aggregated statistics
- `DeleteRating()`: Remove user's rating
- `GetAllRatingsForManga()`: Get all ratings for a manga (paginated)

### Frontend Implementation

#### Components
**Star Rating UI**:
- 10-star rating system
- Interactive hover effects
- Visual feedback for current rating
- Disabled state during submission

**Rating Display**:
- Large average rating number
- Star visualization (filled/unfilled)
- Total rating count
- User's personal rating (if authenticated)

#### Features
- Automatic rating fetch on manga detail load
- Real-time rating updates
- Authentication required to rate
- Redirect to login if not authenticated

## Feature 2: MangaPlus Fallback for Licensed Manga

### Problem Statement
Some manga (e.g., One Piece, Naruto) are officially licensed and not available on MangaDex for reading. These chapters have `externalUrl` fields pointing to official sources like MangaPlus.

### Solution Architecture

#### Backend Changes

**1. Enhanced MangaDex API Models**
```go
type MangaDexChapterAttributes struct {
    // ... existing fields
    ExternalUrl *string `json:"externalUrl"` // New field
}
```

**2. Updated ChapterInfo Model**
```go
type ChapterInfo struct {
    // ... existing fields
    ExternalUrl *string `json:"external_url"`
    IsExternal  bool    `json:"is_external"`
}
```

**3. Chapter Service Logic**
```go
// In getMangaDexChapters()
if mdChapter.Attributes.ExternalUrl != nil && *mdChapter.Attributes.ExternalUrl != "" {
    chapterInfo.ExternalUrl = mdChapter.Attributes.ExternalUrl
    chapterInfo.IsExternal = true
    // Detect MangaPlus URLs
    if strings.Contains(*mdChapter.Attributes.ExternalUrl, "mangaplus.shueisha.co.jp") {
        chapterInfo.Source = "mangaplus"
    }
}
```

#### Frontend Changes

**1. Chapter Click Handler**
```javascript
const handleChapterClick = (chapter) => {
  // Check if this is an external chapter (licensed manga)
  if (chapter.is_external && chapter.external_url) {
    // Open external URL in new tab
    window.open(chapter.external_url, '_blank', 'noopener,noreferrer');
    return;
  }
  // ... normal chapter navigation
};
```

**2. Visual Indicators**
- External chapters show "External" badge (blue)
- Regular chapters show "MangaDex" or "MangaPlus" badge (gray)

**3. Chapter Data Mapping**
```javascript
return realChapters.map(ch => ({
  // ... existing fields
  external_url: ch.external_url || null,
  is_external: ch.is_external || false
}));
```

### User Experience Flow

#### For Licensed Manga (e.g., One Piece):
1. User navigates to manga detail page
2. System fetches chapters from MangaDex
3. MangaDex returns chapters with `externalUrl` for licensed content
4. Frontend displays chapters with "External" badge
5. User clicks on external chapter
6. New tab opens to official MangaPlus reader
7. User reads on official platform

#### Example URLs:
- One Piece Chapter 1: `https://mangaplus.shueisha.co.jp/viewer/1000486`
- Format: `https://mangaplus.shueisha.co.jp/viewer/{chapter_id}`

### Security Considerations

1. **CORS**: External links use `noopener,noreferrer` for security
2. **URL Validation**: Only trusted domains (mangaplus.shueisha.co.jp)
3. **Authentication**: Rating requires JWT token
4. **Input Validation**: 
   - Rating must be 0-10
   - SQL injection prevented via prepared statements
   - Unique constraint on (user_id, manga_id)

## Database Migration

To apply these changes to existing databases:

```sql
-- Add rating table
CREATE TABLE IF NOT EXISTS manga_ratings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    manga_id TEXT NOT NULL,
    rating INTEGER NOT NULL CHECK(rating >= 0 AND rating <= 10),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, manga_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
);

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_ratings_manga ON manga_ratings(manga_id);
CREATE INDEX IF NOT EXISTS idx_ratings_user ON manga_ratings(user_id);
```

## Testing Recommendations

### Rating System Tests
1. **Create Rating**: Rate a manga (0-10)
2. **Update Rating**: Change existing rating
3. **View Stats**: Check average calculation
4. **Delete Rating**: Remove user's rating
5. **Unauthorized**: Try rating without login
6. **Invalid Rating**: Try rating with value < 0 or > 10

### MangaPlus Fallback Tests
1. **Licensed Manga**: Test with One Piece (mal-13)
2. **External Link**: Verify MangaPlus URL opens in new tab
3. **Badge Display**: Check "External" badge appears
4. **Mixed Chapters**: Some external, some internal
5. **Regular Manga**: Verify non-licensed manga still works normally

## API Integration Notes

### MangaPlus API
- Library: `github.com/luevano/mangoplus`
- Documentation: https://pkg.go.dev/github.com/luevano/mangoplus
- Status: Installed but not directly used yet
- Future: Can be used for direct chapter fetching if needed

### MangaDex API
- Already integrated
- Returns `externalUrl` in chapter attributes for licensed content
- No changes required to API calls

## Known Limitations

1. **MangaDex API Limit**: Maximum 500 chapters per request
2. **Rating Scale**: Fixed 0-10 (not customizable)
3. **External URLs**: No validation of URL availability
4. **MangaPlus Access**: Depends on user's geographic location
5. **MAL Rating Removed**: No longer fetches ratings from MAL API

## Future Enhancements

1. **Rating Distribution**: Show histogram of rating distribution
2. **Rating Comments**: Allow users to write reviews
3. **Rating Filters**: Filter manga by rating range
4. **Direct MangaPlus Integration**: Use mangoplus library for direct chapter fetching
5. **Fallback Chain**: MangaDex → MangaPlus → Other sources
6. **Region Detection**: Warn users if external content not available in their region
7. **Rating Analytics**: Track rating trends over time
8. **Moderation**: Flag/remove inappropriate ratings

## Migration from External API Ratings

### What Changed:
- **Before**: Manga ratings fetched from MAL/Jikan API
- **After**: Ratings stored in local database, calculated from user submissions

### Data Consistency:
- Old MAL ratings are ignored
- Start fresh with user-generated ratings
- Average rating starts at 0 with 0 ratings
- Gradually builds up as users rate manga

### Benefits:
- Full control over rating system
- No dependency on external API
- Custom rating scale (0-10 vs MAL's 0-10)
- Can add features like reviews, moderation, etc.
