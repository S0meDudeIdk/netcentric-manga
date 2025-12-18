# Quick Start: Reading Manga from MAL Database

## What Changed?

Your MangaHub system now **automatically searches MangaDex** when you try to read manga from your MAL database. No manual configuration needed!

## How It Works

1. **Browse manga** using MAL/Jikan API (titles, covers, descriptions, ratings)
2. **Click "Read"** on any manga
3. **System automatically**:
   - Searches MangaDex using the manga title
   - Finds the best matching manga
   - Fetches real chapters
   - Opens the reader

## Example: Reading Doraemon

Based on your screenshot showing Doraemon:

1. Browse to Doraemon from MAL database
2. You'll see a blue notice: "Chapters sourced from MangaDex"
3. Click any chapter to read
4. System searches MangaDex for "Doraemon"
5. Loads the actual chapter pages from MangaDex
6. You can now read!

## Testing

```bash
# 1. Start the API server
cd mangahub
go run cmd/api-server/main.go

# 2. Start the React frontend
cd client/web-react
npm start

# 3. Browse to a manga (e.g., Doraemon)
# 4. Click on any chapter
# 5. Should automatically fetch from MangaDex!
```

## New API Endpoints

### Search MangaDex Directly
```
GET /api/v1/manga/mangadex/search?title=Doraemon&limit=10
```

### Get Chapters (Auto-search enabled)
```
GET /api/v1/manga/:id/chapters
```
- Works with MAL IDs (auto-searches MangaDex)
- Works with MangaDex UUIDs (direct fetch)
- Works with MangaPlus IDs (direct fetch)

## What You See in the UI

### Before Reading
- Blue info box: "Chapters sourced from MangaDex"
- Explains automatic search will happen

### When Real Chapters Load
- Chapter list shows actual chapter data
- Source badge: "MangaDex" or "MangaPlus"
- Page counts displayed
- Publication dates shown

## Advantages

✅ **Best of Both Worlds**
- MAL: Comprehensive metadata, ratings, recommendations
- MangaDex: Extensive chapter library, high-quality scans

✅ **No Manual Work**
- Automatic title-based matching
- No need to link manga manually

✅ **Fallback Support**
- Tries MangaPlus if MangaDex doesn't have chapters
- Clear error messages if nothing found

## Common Scenarios

### Scenario 1: Popular Manga (e.g., Doraemon)
✅ MangaDex has it → Chapters load automatically

### Scenario 2: Obscure Manga
⚠️ Not on MangaDex → Message: "No chapters found"

### Scenario 3: Multiple Matches
✅ System picks most relevant result (by MangaDex relevance score)

## Troubleshooting

**"No chapters found"**
- Manga might not be on MangaDex yet
- Try searching MangaDex.org manually to verify

**Wrong manga chapters loading**
- Title might be ambiguous
- Future update will add manual linking option

**Rate limit errors**
- Add your MangaDex API key to `.env` file
- Wait a few seconds and try again

## Next Steps

1. Test with the Doraemon manga from your screenshot
2. Try other manga from your MAL database
3. Check the new API endpoints
4. Read the full documentation in `docs/AUTOMATIC_MANGADEX_INTEGRATION.md`

## Summary

You can now:
- Browse 10,000+ manga from MAL
- Read chapters automatically from MangaDex
- Enjoy seamless integration between both APIs
- No manual manga linking required!
