# MangaHub Access Model

## 🌐 Public Access (No Login Required)

Users can freely explore and read manga without creating an account!

### Available Features:
- ✅ **Browse All Manga** - View the entire collection at `/browse`
- ✅ **Search** - Find manga by title, author, or genre at `/search`
- ✅ **View Details** - Click any manga to see full information
- ✅ **Filter & Sort** - Organize manga by genre, popularity, chapters, year
- ✅ **View Stats** - See ratings, chapter counts, publication years
- ✅ **Responsive UI** - Works on all devices

### Public Routes:
- `/` - Home page
- `/browse` - Browse all manga
- `/search` - Search functionality
- `/manga/:id` - Manga details page

## 🔒 Account Features (Free Registration Required)

Create a free account to unlock these additional features:

### Library Management:
- ⭐ **Save to Library** - Add manga to your personal collection
- ⭐ **Track Progress** - Save which chapter you're on
- ⭐ **Continue Reading** - Pick up exactly where you left off
- ⭐ **Reading Status** - Mark as:
  - Reading
  - Completed
  - Plan to Read
  - On Hold
  - Dropped
- ⭐ **Reading Lists** - Create custom playlists
- ⭐ **Statistics** - View your reading stats and history
- ⭐ **Recommendations** - Get personalized suggestions

### Protected Routes:
- `/library` - Personal manga library (requires login)

## 🎯 User Flow

### Guest User Journey:
1. Visit homepage → Browse freely
2. Search for manga → View results
3. Click manga → See details
4. Try to add to library → Prompted to login/register

### Registered User Journey:
1. Visit homepage → Browse freely
2. Search for manga → View results
3. Click manga → See details
4. Add to library → Saved to collection
5. Update progress → Track chapters read
6. View library → See all saved manga with stats

## 🔄 Converting from Guest to Member

When a guest user tries to use account-only features:
1. **"Add to Library" button** shows "Login to Add" for guests
2. Clicking it redirects to `/login` page
3. After login/registration, they can immediately add manga
4. All progress is saved to their account

## 📊 Benefits

### For Users:
- **Try before registering** - Explore the entire collection first
- **No barriers** - Browse and read without signup friction
- **Value proposition** - See what you get before creating account
- **Privacy** - Browse anonymously if preferred

### For the Platform:
- **Lower barrier to entry** - More users can explore
- **Better conversion** - Users understand value before registering
- **Reduced bounce rate** - Users don't leave immediately
- **Natural upgrade path** - Clear incentive to create account

## 🛡️ Technical Implementation

### Authentication:
- **Optional JWT tokens** - Sent only if user is logged in
- **Public endpoints** - Manga browsing works without auth
- **Protected endpoints** - Library operations require auth token

### Service Layer:
```javascript
// mangaService.js - Auth headers are optional
const getAuthHeaders = () => {
  const token = authService.getToken();
  if (token) {
    return {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    };
  }
  return {
    'Content-Type': 'application/json'
  };
};
```

### Routing:
```javascript
// App.js - Public vs Protected routes
// Public - No login needed
<Route path="/browse" element={<Browse />} />
<Route path="/search" element={<Search />} />
<Route path="/manga/:id" element={<MangaDetail />} />

// Protected - Login required
<Route path="/library" element={
  <ProtectedRoute>
    <Library />
  </ProtectedRoute>
} />
```

## 🎨 UI Patterns

### For Guests:
- Navigation shows "Login" and "Register" buttons
- Manga detail page shows "Login to Add" button
- Library link hidden in navigation
- CTA sections encourage account creation

### For Logged-In Users:
- Navigation shows username and user menu
- Manga detail page shows "Add to Library" with progress tracking
- Library link visible in navigation
- User can logout from dropdown menu

## 📝 Summary

**Anyone can browse and explore manga freely. Creating an account unlocks the ability to save progress, build a library, and get personalized recommendations.**

This model provides the best of both worlds:
- 🌍 **Open access** for discovery
- 🔐 **Account value** for engagement
- 📈 **Natural conversion** path
