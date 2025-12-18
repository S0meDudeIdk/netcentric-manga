# MangaHub Chapter Reading Implementation

## Overview

This document describes the implementation of the manga chapter reading feature using MangaDex API as the primary source and MangaPlus API as a fallback.

## Features

- **Multi-Source Support**: Fetches chapters from MangaDex and MangaPlus APIs
- **Full Chapter Reader**: Interactive reader with multiple viewing modes
- **Progress Tracking**: Automatically updates user's reading progress
- **Keyboard Navigation**: Use arrow keys or A/D for navigation
- **Customizable Reader**: Adjust view mode, fit mode, and background color

## Architecture

### Backend Components

#### 1. API Clients

**MangaDex Client** (`internal/external/mangadex.go`)
- Searches manga by title
- Retrieves manga details by ID
- Fetches chapter lists with pagination
- Gets chapter pages from at-home servers
- Supports multiple languages

**MangaPlus Client** (`internal/external/mangaplus.go`)
- Retrieves title details
- Fetches chapter lists
- Gets chapter pages
- Fallback for manga not available on MangaDex

#### 2. Chapter Service (`internal/manga/chapter_service.go`)

Provides unified interface for chapter operations:
- `GetChapterList(mangaID, languages, limit, offset)` - Retrieves chapter list
- `GetChapterPages(chapterID, source)` - Gets page URLs for a chapter
- Automatically determines source based on manga ID
- Implements fallback logic

#### 3. API Endpoints

**GET** `/api/v1/manga/:manga_id/chapters`
- Query params: `language[]`, `limit`, `offset`
- Returns: Chapter list with metadata

**GET** `/api/v1/manga/chapters/:chapter_id/pages`
- Query params: `source` (mangadex|mangaplus)
- Returns: Array of page image URLs

### Frontend Components

#### 1. Chapter Reader (`pages/ChapterReader.jsx`)

Features:
- **View Modes**:
  - Single Page: Navigate page by page
  - Vertical Scroll: Continuous scrolling
  
- **Fit Modes**:
  - Fit Width: Scale to browser width
  - Fit Height: Scale to viewport height
  - Original Size: No scaling

- **Navigation**:
  - Keyboard: Arrow keys, A/D
  - Click: Left/right thirds of screen
  - Buttons: Previous/Next page and chapter
  
- **Settings**:
  - Customizable background color
  - View and fit mode preferences
  - Persist across sessions (future enhancement)

#### 2. Updated MangaDetail (`pages/MangaDetail.jsx`)

Enhancements:
- Fetches real chapter data from backend
- Displays chapter source (MangaDex/MangaPlus)
- Shows page count when available
- Click to read functionality
- Fallback to placeholder chapters for unsupported manga

#### 3. Manga Service (`services/mangaService.js`)

New methods:
- `getChapters(mangaID, languages, limit, offset)` - Fetch chapter list
- `getChapterPages(chapterID, source)` - Fetch chapter pages

## Usage

### Reading a Chapter

1. Navigate to a manga detail page
2. Click on any chapter in the chapter list
3. The reader will open with the chapter loaded
4. Use navigation controls or keyboard shortcuts to read

### Supported Manga Sources

- **MangaDex**: Primary source, supports most manga
- **MangaPlus**: Official Shueisha manga (One Piece, Naruto, etc.)
- **Local Database**: Placeholder chapters (not readable yet)

### Manga ID Formats

- MangaDex: `mangadex-{uuid}` or raw UUID
- MangaPlus: `mangaplus-{title_id}` or numeric ID
- Local: Numeric ID from local database

## API Integration Details

### MangaDex API

Base URL: `https://api.mangadex.org`

Key Endpoints:
- `/manga/{id}` - Manga details
- `/manga/{id}/feed` - Chapter list
- `/at-home/server/{chapterId}` - Chapter pages

Features:
- Free and open API
- Supports multiple languages
- Community-driven translations
- Rate limiting: Be respectful

### MangaPlus API

Base URL: `https://jumpg-webapi.tokyo-cdn.com/api`

Key Endpoints:
- `/title_detailV3?title_id={id}` - Title details
- `/manga_viewer?chapter_id={id}` - Chapter pages

Features:
- Official Shueisha manga
- High-quality images
- Latest chapters
- Limited to Shueisha titles

## Configuration

The APIs are publicly accessible and don't require API keys. Configuration is done through environment variables in `.env`.

### Environment Variables

```env
# MangaDex API Configuration
MANGADEX_API_BASE_URL=https://api.mangadex.org
MANGADEX_API_TIMEOUT=15          # Timeout in seconds
MANGADEX_DEBUG=false             # Enable verbose logging
MANGADEX_API_KEY=                # Optional: Your MangaDex API key

# MangaPlus API Configuration
MANGAPLUS_API_BASE_URL=https://jumpg-webapi.tokyo-cdn.com/api
MANGAPLUS_API_TIMEOUT=15         # Timeout in seconds
MANGAPLUS_DEBUG=false            # Enable verbose logging
```

### Configuration Details

- **Base URLs**: Can be changed to use alternative endpoints or proxies
- **Timeouts**: Adjust based on network conditions (default: 15 seconds)
- **Debug Mode**: When enabled, logs detailed API request/response information
- **MangaDex API Key** (Optional):
  - Get yours at: https://mangadex.org/settings (API Clients section)
  - **Not required** for reading public manga
  - **Benefits when provided**:
    - Higher rate limits
    - Access to private/restricted content
    - User-specific features (follows, reading history)
    - Better priority during high traffic
- **No API Key for MangaPlus**: MangaPlus is a free, public API

See `.env.example` for a complete list of all environment variables.

## Error Handling

The implementation includes robust error handling:

1. **Chapter List Errors**: Falls back to placeholder chapters
2. **Page Load Errors**: Shows placeholder image
3. **API Errors**: Displays error message with details
4. **Network Issues**: Timeout after 15 seconds

## Future Enhancements

1. **Download Support**: Allow offline reading
2. **Bookmarks**: Mark favorite chapters
3. **Reading History**: Track all read chapters
4. **Comments**: Chapter-specific discussions
5. **Image Caching**: Faster page loads
6. **Preloading**: Load next chapter in background
7. **Custom Sources**: Add more manga sources
8. **Reader Themes**: More color schemes
9. **Double Page View**: For manga spreads
10. **RTL Support**: Right-to-left reading mode

## Testing

### Manual Testing

1. **MangaDex Manga**:
   - Search for a manga
   - Open details page
   - Click a chapter
   - Verify reader loads and pages display

2. **Navigation**:
   - Test keyboard shortcuts
   - Test click navigation
   - Test chapter navigation
   - Test settings panel

3. **Progress Tracking**:
   - Add manga to library
   - Read a chapter
   - Verify progress updates

### API Testing

```bash
# Test chapter list endpoint
curl "http://localhost:8080/api/v1/manga/mangadex-{id}/chapters?language=en&limit=20"

# Test chapter pages endpoint
curl "http://localhost:8080/api/v1/manga/chapters/{chapter-id}/pages?source=mangadex"
```

## Troubleshooting

### Chapters Not Loading

- Check if manga ID is valid
- Verify internet connection
- Check browser console for errors
- Try refreshing the page

### Images Not Displaying

- MangaDex at-home servers may be slow
- Try a different chapter
- Check browser CORS settings
- Verify image URLs in network tab

### Reader Not Working

- Clear browser cache
- Check JavaScript console for errors
- Verify backend server is running
- Test API endpoints directly

## Contributing

To add a new manga source:

1. Create client in `internal/external/{source}.go`
2. Implement required interfaces
3. Add to chapter service fallback logic
4. Update frontend to display source
5. Add tests and documentation

## License

This feature is part of the MangaHub project and follows the same license.
