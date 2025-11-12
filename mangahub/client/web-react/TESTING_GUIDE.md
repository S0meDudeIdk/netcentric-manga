# Testing Your React App - Quick Guide

## 🧪 Step-by-Step Testing Guide

### Prerequisites Check
1. ✅ API server running on port 8080
2. ✅ Node.js installed
3. ✅ Dependencies installed (`npm install`)

### Test 1: Start the App
```powershell
cd client\web-react
npm start
```
**Expected**: Browser opens to http://localhost:3000

### Test 2: Home Page
- ✅ See "MangaHub" logo and title
- ✅ See "Browse Manga" and "Search" buttons
- ✅ See "Popular Manga" section with manga cards
- ✅ See features section (Track Progress, Personal Library, Recommendations)
- ✅ If not logged in, see "Ready to start your manga journey?" CTA

### Test 3: Browse Page
1. Click "Browse Manga" button or nav link
2. ✅ See manga grid
3. ✅ See genre filter dropdown
4. ✅ See sort dropdown (Popular, Title, Chapters, Year)
5. ✅ Filter by a genre - grid updates
6. ✅ Change sort order - grid reorders
7. ✅ Click manga card - goes to detail page

### Test 4: Search Page
1. Click "Search" in navigation
2. ✅ See search bar with placeholder
3. ✅ See quick suggestions (One Piece, Naruto, etc.)
4. Type "one" and click Search
5. ✅ See results with manga matching "one"
6. ✅ Click X to clear search
7. ✅ Click a suggestion - auto-searches

### Test 5: Registration
1. Click "Register" button
2. ✅ See registration form
3. Fill out:
   - Username: testuser123
   - Email: test@test.com
   - Password: test123
   - Confirm: test123
4. Click "Create Account"
5. ✅ Redirects to Library page
6. ✅ See user menu in header with username

### Test 6: Login (if you already have an account)
1. Click "Login" button
2. ✅ See login form
3. Enter:
   - Email: test@test.com
   - Password: test123
4. Click "Sign In"
5. ✅ Redirects to Library
6. ✅ User menu appears in header

### Test 7: Manga Detail Page
1. Navigate to Browse or Search
2. Click any manga card
3. ✅ See large cover image
4. ✅ See title, author, publication year
5. ✅ See description
6. ✅ See genres as colored tags
7. ✅ See stats (rating, chapters, year, genres)
8. ✅ See "Add to Library" button (if logged in)

### Test 8: Add to Library
1. On manga detail page (while logged in)
2. Click "Add to Library"
3. ✅ Button changes to "In Your Library" with checkmark
4. ✅ See status dropdown (Reading, Completed, etc.)
5. ✅ See current chapter input
6. ✅ See progress bar

### Test 9: Update Progress
1. On manga detail with manga in library
2. Change status dropdown to "Reading"
3. ✅ Status updates (may need to refresh to see in library)
4. Enter a chapter number (e.g., 5)
5. Click outside the input (blur)
6. ✅ Progress bar updates

### Test 10: Library Page
1. Click "Library" in navigation
2. ✅ See stats cards (Total Manga, Chapters Read, Currently Reading, Completed)
3. ✅ See status filter buttons
4. ✅ See your manga in the grid
5. Click "Reading" filter
6. ✅ Only shows manga with "reading" status
7. ✅ Click manga card - goes to detail

### Test 11: Header Navigation
1. ✅ Logo links to Home
2. ✅ All nav links work (Home, Browse, Search, Library)
3. ✅ User menu shows username
4. Click user menu dropdown
5. ✅ See "My Library" and "Logout" options

### Test 12: Logout
1. Click user dropdown
2. Click "Logout"
3. ✅ Redirects to login page
4. ✅ User menu replaced with Login/Register buttons
5. ✅ Library nav link hidden

### Test 13: Protected Routes
1. While logged out, try to visit: http://localhost:3000/library
2. ✅ Automatically redirects to /login

### Test 14: Mobile Responsive
1. Open browser DevTools (F12)
2. Toggle device toolbar (Ctrl+Shift+M)
3. Select "iPhone 12 Pro" or similar
4. ✅ Header shows hamburger menu
5. ✅ Click hamburger - mobile menu slides down
6. ✅ All links visible and working
7. ✅ Manga cards stack vertically
8. ✅ Forms are readable and usable

### Test 15: Animations
1. ✅ Hover over manga cards - slight lift effect
2. ✅ Loading spinners animate when fetching data
3. ✅ Page transitions smooth
4. ✅ Button hover effects work

## 🐛 Common Test Failures & Solutions

### No manga showing on Home/Browse
**Problem**: API server not running or no data
**Solution**: 
- Check API: http://localhost:8080/api/v1/manga
- Run: `.\start-server.ps1` from mangahub directory

### "Network Error" or CORS error
**Problem**: API server not accessible
**Solution**:
- Verify server running on port 8080
- Check console for specific error
- Ensure CORS enabled in API (already configured)

### Login/Register not working
**Problem**: API auth endpoints not responding
**Solution**:
- Check API: http://localhost:8080/api/v1/auth/register
- Check browser console for error details
- Verify request payload in Network tab

### Token expired / Auto-logout
**Expected**: JWT tokens expire after time
**Solution**: This is normal - just log in again

### Library not loading
**Problem**: Not authenticated or API issue
**Solution**:
- Verify logged in (check localStorage: `localStorage.getItem('token')`)
- Check browser console for errors
- Try logging out and back in

### Images not showing
**Expected**: Some manga may not have cover_url
**Solution**: Placeholder will show - this is normal

## ✅ Success Criteria

Your React app is working correctly if:
- ✅ All pages load without errors
- ✅ Navigation works smoothly
- ✅ Can search and browse manga
- ✅ Can register and login
- ✅ Can add manga to library
- ✅ Can update reading progress
- ✅ Responsive on mobile devices
- ✅ No console errors (except expected 404s for missing images)

## 📊 Browser Console Check

Open DevTools (F12) → Console tab

**Should NOT see**:
- ❌ Red errors about failed API calls
- ❌ CORS errors
- ❌ "Cannot read property of undefined" errors
- ❌ React rendering errors

**OK to see**:
- ⚠️ Warnings about keys in lists (minor)
- ⚠️ 404 errors for missing manga cover images
- ℹ️ Info about React development mode

## 🎯 Quick Smoke Test (2 minutes)

1. Start app → Home page loads ✅
2. Click Browse → Manga grid shows ✅
3. Click Search → Can search ✅
4. Register new account → Succeeds ✅
5. Add manga to library → Appears in library ✅
6. Logout → Returns to login ✅

If all ✅, your app is working! 🎉

## 📝 Notes

- **First load**: May be slow while fetching all manga
- **Token storage**: localStorage (persists across sessions)
- **Auto-refresh**: Library data refreshes after updates
- **Concurrent users**: Each user has their own library

## 🚀 Performance Tips

- Use Chrome DevTools Lighthouse for performance audit
- Check Network tab for slow API calls
- Monitor memory usage in Performance tab
- Optimize images if custom cover_url added

---

Happy testing! If all tests pass, your React frontend is ready to use! 🎊
