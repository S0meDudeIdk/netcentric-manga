# Environment Variable Updates for Chapter Reading

## What Changed

The MangaDex and MangaPlus API clients now use environment variables for configuration, making them more flexible and easier to customize.

## New Environment Variables

Add these to your `.env` file:

```env
# MangaDex API Configuration
MANGADEX_API_BASE_URL=https://api.mangadex.org
MANGADEX_API_TIMEOUT=15
MANGADEX_DEBUG=false

# MangaPlus API Configuration
MANGAPLUS_API_BASE_URL=https://jumpg-webapi.tokyo-cdn.com/api
MANGAPLUS_API_TIMEOUT=15
MANGAPLUS_DEBUG=false
```

## Benefits

1. **Configurable Base URLs**: Easy to use proxies or alternative endpoints
2. **Adjustable Timeouts**: Customize based on your network conditions
3. **Debug Mode**: Enable verbose logging when troubleshooting
4. **No Breaking Changes**: Works with default values if not configured

## Migration Steps

1. Copy `.env.example` to `.env` if you haven't already
2. Add the new MangaDex and MangaPlus configuration variables
3. Restart your backend server
4. Test the chapter reading functionality

## Default Values

If you don't add these variables, the system will use these defaults:

- **MangaDex Base URL**: `https://api.mangadex.org`
- **MangaPlus Base URL**: `https://jumpg-webapi.tokyo-cdn.com/api`
- **Timeout**: 15 seconds
- **Debug**: false

## Debugging

To enable debug logging for API calls:

```env
MANGADEX_DEBUG=true
MANGAPLUS_DEBUG=true
```

This will log:
- API request URLs
- Response status codes
- Error details
- Timing information

## No API Keys Required

Both MangaDex and MangaPlus are free, public APIs that don't require authentication or API keys. You can start using them immediately!

## Questions?

Check the documentation:
- `docs/CHAPTER_READING_IMPLEMENTATION.md` - Full technical details
- `docs/CHAPTER_READING_QUICK_START.md` - User guide
- `.env.example` - Complete environment variable reference
