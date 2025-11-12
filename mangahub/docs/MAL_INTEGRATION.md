# MyAnimeList API Integration

This document describes the MyAnimeList (MAL) API integration using the Jikan API v4.

## Overview

The MangaHub application now fetches real manga data from MyAnimeList via the unofficial Jikan API. This provides access to a vast database of manga information including:

- Manga titles, authors, and descriptions
- Cover images
- Genres, themes, and demographics
- Publication status and chapter counts
- Ratings and popularity rankings

## Architecture

### Backend Components

1. **Jikan Client** (`internal/external/jikan.go`)
   - HTTP client for Jikan API v4
   - Rate limiting (1 request per second)
   - Methods:
     - `SearchManga(query, page, limit)` - Search for manga
     - `GetTopManga(page, limit)` - Get top-ranked manga
     - `GetMangaByID(malID)` - Get specific manga by MyAnimeList ID
     - `GetMangaRecommendations(malID)` - Get recommendations (not yet exposed)

2. **Data Converter** (`internal/external/converter.go`)
   - Converts Jikan API responses to internal Manga model
   - Maps MAL statuses to internal statuses
   - Extracts and combines genres, themes, and demographics
   - Generates IDs in format `mal-{malID}`

3. **API Endpoints** (`cmd/api-server/main.go`)
   - `GET /api/v1/manga/mal/search?q={query}&page={page}&limit={limit}` - Search MAL
   - `GET /api/v1/manga/mal/top?page={page}&limit={limit}` - Get top manga
   - `GET /api/v1/manga/mal/{mal_id}` - Get specific manga

### Frontend Components

1. **Web Client** (`client/web-react`)
   - **mangaService.js**: Added MAL API methods
     - `searchMAL(query, page, limit)`
     - `getTopMAL(page, limit)`
     - `getMALMangaById(malId)`
   
   - **Browse.jsx**: Toggle between Local and MyAnimeList data sources
   - **Search.jsx**: Toggle between Local and MyAnimeList search

2. **CLI Client** (`client/cli/main.go`)
   - Option 3: "Search MyAnimeList"
   - Displays MAL manga with source indicator

## Usage

### Starting the Server

```bash
cd mangahub/cmd/api-server
go run main.go
```

The server will start on port 8080 with MAL API endpoints enabled.

### Testing the API

Run the test script:

```powershell
cd mangahub
.\test-mal-api.ps1
```

### Using the Web Client

1. Start the React app:
   ```powershell
   cd client/web-react
   npm start
   ```

2. Navigate to Browse or Search page

3. Use the toggle buttons at the top to switch between:
   - **MyAnimeList**: Real data from MAL (default)
   - **Local**: Data from local database

### Using the CLI Client

1. Run the CLI:
   ```bash
   cd client/cli
   go run main.go
   ```

2. Login or browse as guest

3. Select option 3: "Search MyAnimeList"

4. Enter a search query (e.g., "Naruto", "One Piece")

5. View results and details

## Rate Limiting

The Jikan API has strict rate limits:
- **1 request per second**
- **3 requests per second** for some endpoints (not used in this implementation)

The client automatically respects these limits by:
- Tracking the last request time
- Sleeping if necessary before making a new request

## Data Model Extensions

The `Manga` model was extended with new fields from MAL:

```go
type Manga struct {
    // ... existing fields ...
    PublicationYear int     `json:"publication_year" db:"publication_year"`
    Rating          float64 `json:"rating" db:"rating"`
}
```

## API Examples

### Search Manga

**Request:**
```http
GET /api/v1/manga/mal/search?q=naruto&page=1&limit=10
```

**Response:**
```json
{
  "data": [
    {
      "id": "mal-11",
      "title": "Naruto",
      "author": "Kishimoto, Masashi",
      "genres": ["Action", "Adventure", "Shounen"],
      "status": "completed",
      "total_chapters": 700,
      "description": "...",
      "cover_url": "https://cdn.myanimelist.net/images/manga/3/249658.jpg",
      "publication_year": 1999,
      "rating": 8.1
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```

### Get Top Manga

**Request:**
```http
GET /api/v1/manga/mal/top?page=1&limit=20
```

**Response:**
```json
{
  "data": [
    {
      "id": "mal-2",
      "title": "Berserk",
      "author": "Miura, Kentarou",
      "genres": ["Action", "Adventure", "Drama", "Fantasy", "Horror", "Seinen"],
      "status": "ongoing",
      "total_chapters": 0,
      "description": "...",
      "cover_url": "https://cdn.myanimelist.net/images/manga/1/157897.jpg",
      "publication_year": 1989,
      "rating": 9.46
    }
  ],
  "total": 20,
  "page": 1,
  "limit": 20
}
```

### Get Manga by ID

**Request:**
```http
GET /api/v1/manga/mal/13
```

**Response:**
```json
{
  "id": "mal-13",
  "title": "One Piece",
  "author": "Oda, Eiichiro",
  "genres": ["Action", "Adventure", "Comedy", "Drama", "Fantasy", "Shounen"],
  "status": "ongoing",
  "total_chapters": 1100,
  "description": "...",
  "cover_url": "https://cdn.myanimelist.net/images/manga/2/253146.jpg",
  "publication_year": 1997,
  "rating": 9.21
}
```

## Error Handling

The integration handles various error scenarios:

1. **Network Errors**: Displayed as "Failed to search/get MyAnimeList"
2. **Rate Limiting**: Automatically handled with delays
3. **404 Not Found**: "Manga not found on MyAnimeList"
4. **Invalid Parameters**: "Invalid MyAnimeList ID" or "Search query is required"

## Freemium Model Compatibility

The MAL API endpoints are **public** (no authentication required), aligning with the freemium model:

- ✅ Browse MAL manga without login
- ✅ Search MAL manga without login
- ✅ View MAL manga details without login
- 🔒 Library features still require authentication

## Future Enhancements

Potential improvements:

1. **Caching**: Cache MAL responses to reduce API calls
2. **Recommendations**: Expose the recommendations endpoint
3. **Batch Operations**: Fetch multiple manga in one go
4. **Advanced Filters**: Genre, year, status filtering
5. **User Reviews**: Fetch and display MAL user reviews
6. **Related Manga**: Show related/similar manga
7. **Seasonal Manga**: Get manga from specific seasons
8. **Character Info**: Fetch character details

## Limitations

1. **Read-only**: Cannot modify MAL data (library features use local DB)
2. **Rate Limits**: Slow browsing due to 1 req/sec limit
3. **Pagination**: Jikan API limits to 25 results per page
4. **No Real-time Updates**: MAL data may be slightly outdated
5. **English Titles**: Some manga may have Japanese-only titles

## Resources

- [Jikan API Documentation](https://docs.api.jikan.moe/)
- [MyAnimeList](https://myanimelist.net/)
- [Jikan GitHub](https://github.com/jikan-me/jikan)

## Troubleshooting

**Problem**: "Failed to search MyAnimeList"
- **Solution**: Check if API server is running and Jikan API is accessible

**Problem**: Rate limit errors
- **Solution**: Wait 1 second between requests (handled automatically)

**Problem**: Empty results
- **Solution**: Try different search terms or check MAL website

**Problem**: No cover images
- **Solution**: Some manga don't have images in MAL database
