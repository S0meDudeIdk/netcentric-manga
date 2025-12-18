# Chapter Pagination Implementation Guide

## Overview
This guide explains the new chapter pagination feature implemented in the MangaDetail page. The feature includes volume-based pagination, chapter-based pagination, index modal, and chapter sorting.

## Features Implemented

### 1. **Smart Pagination**
- **Volume-based**: When volumes are detected, displays 1 volume per page
- **Chapter-based**: When no volumes are found, displays 20 chapters per page
- Automatically determines the pagination method based on available data

### 2. **Index Modal** (Magnifier Button)
- Click the magnifier/search icon to open the index
- Shows all volumes or chapter ranges
- Click on any volume/chapter range to jump directly to that page
- Current page is highlighted
- Modal can be closed by clicking outside or the X button

### 3. **Chapter Sorting** (Arrow Button)
- Click the arrow button to toggle between ascending and descending order
- **Descending** (default): Shows newest chapters first (Chapter 821 → Chapter 1)
- **Ascending**: Shows oldest chapters first (Chapter 1 → Chapter 821)
- Works for both volume-based and chapter-based pagination

### 4. **Pagination Controls**
- Previous/Next buttons to navigate between pages
- Current page indicator showing:
  - "Volume X of Y" for volume-based pagination
  - "Page X of Y" for chapter-based pagination
- Buttons are disabled when at the first or last page

## Implementation Details

### Service Layer (`mangaService.js`)
Added `getChapterList` method that:
- Generates chapter data based on total chapters
- Organizes chapters into volumes (assuming ~18 chapters per volume)
- Returns structured chapter objects with number, volume, and title

### Component Layer (`MangaDetail.jsx`)

#### New State Variables:
```javascript
const [showIndexModal, setShowIndexModal] = useState(false);
const [sortAscending, setSortAscending] = useState(false);
const [currentPage, setCurrentPage] = useState(1);
```

#### Key Functions:
- `chapters`: Memoized chapter list generation
- `volumeGroups`: Groups chapters by volume
- `paginatedContent`: Returns the chapters for the current page
- `handlePageChange`: Navigate between pages
- `handleVolumeSelect`: Jump to specific volume from index
- `toggleSort`: Switch between ascending/descending order

#### UI Components:
1. **Chapter List Header**: Shows volume number when applicable
2. **Action Buttons**: Magnifier (index) and Arrow (sort)
3. **Chapter Items**: Display chapter number and read status
4. **Pagination Controls**: Previous/Next buttons with page indicator
5. **Index Modal**: Full-screen overlay with volume/chapter navigation

## Usage

### For Users:
1. **View Chapters**: Scroll through the current page of chapters
2. **Navigate Pages**: Use Previous/Next buttons
3. **Jump to Volume**: Click magnifier icon → Select volume from index
4. **Sort Order**: Click arrow icon to reverse chapter order
5. **Read Status**: Chapters you've read are marked with a checkmark

### For Developers:
The implementation is modular and can be enhanced with:
- Real chapter data from API endpoints
- Chapter titles and publication dates
- Chapter thumbnails
- Reading links
- Download functionality

## Future Enhancements

1. **API Integration**: Connect to real chapter data endpoints
2. **Search**: Add chapter search functionality within the index
3. **Filters**: Filter by read/unread status
4. **Bookmarks**: Allow users to bookmark specific chapters
5. **Reading Interface**: Add chapter reader functionality
6. **Chapter Details**: Show scanlation groups, upload dates, etc.

## Example Scenarios

### Scenario 1: Manga with Volumes (e.g., Doraemon - 821 chapters)
- System detects 821 chapters
- Organizes into ~46 volumes (18 chapters each)
- Shows 1 volume per page
- Index displays all 46 volumes

### Scenario 2: Manga without Volume Data
- System detects total chapters
- Shows 20 chapters per page
- Index displays chapter ranges (1-20, 21-40, etc.)
- User can navigate through all pages

## Notes

- The current implementation generates mock chapter data based on total chapters
- Chapter-to-volume mapping uses an average of 18 chapters per volume
- Real volume data should be integrated when available from the API
- The index modal is designed to match the reference image provided
- All animations use Framer Motion for smooth transitions
