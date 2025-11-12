# MangaHub Clients

This directory contains two client implementations for the MangaHub API:

## 1. CLI Client (Command Line Interface)

A terminal-based client with colorful UI and emoji support.

### Features
- ✅ **User Authentication**: Login and registration with validation
- 📚 **Manga Browsing**: Browse popular manga with customizable results
- 🔍 **Search Functionality**: Search by title, author, or genre
- 📖 **Library Management**: Track reading progress and status
- 💡 **Personalized Recommendations**: Get manga suggestions based on your library
- 📊 **Statistics**: View your reading statistics
- 📡 **Real-time Sync**: TCP-based live progress updates across clients

### Installation & Usage

#### Prerequisites
- Go 1.19 or higher
- MangaHub API server running on `localhost:8080`
- (Optional) TCP server running on `localhost:9000` for real-time sync

#### Running the CLI Client

**Windows (PowerShell):**
```powershell
cd client\cli
go run main.go
```

**Linux/Mac:**
```bash
cd client/cli
go run main.go
```

Or build and run:
```bash
# Build
go build -o mangahub-cli main.go

# Run (Windows)
.\mangahub-cli.exe

# Run (Linux/Mac)
./mangahub-cli
```

### User Guide

#### Registration
1. Select "Register" from the authentication menu
2. Requirements:
   - **Username**: 3-30 characters
   - **Email**: Valid email format
   - **Password**: Minimum 6 characters

The client validates input before sending to the server and provides clear error messages.

#### Login
1. Select "Login" from the authentication menu
2. Enter your registered email and password
3. Upon successful login, the client will automatically attempt to connect to the TCP sync server

#### Browse Manga
- View popular manga with customizable result limits
- See title, author, chapter count, and genres
- Select any manga to view detailed information

#### Search Manga
- Search by:
  - Title (partial match)
  - Author
  - Genre
- Filter results and view details

#### My Library
- View all manga in your library
- Update reading progress
- Mark manga as reading or completed
- View library statistics:
  - Total manga in library
  - Currently reading count
  - Completed count
  - Total chapters read

#### Real-Time Sync (TCP)
When connected to the TCP server:
- Your progress updates are broadcast to other connected clients
- Receive notifications when others update their progress
- Status shown in main menu: "Real-time sync: ENABLED" or "OFFLINE"

### Password Validation

The client enforces password requirements:
- ✅ Minimum 6 characters
- ❌ "guy" - Too short (only 3 characters)
- ✅ "password123" - Valid
- ✅ "admin123" - Valid

Clear error messages guide you through validation issues:
- "❌ Password must be at least 6 characters"
- "❌ Username must be between 3 and 30 characters"
- "❌ Please enter a valid email address"

---

## 2. Web Client (Browser-based)

A modern, responsive web application with a clean UI.

### Features
- 🔐 **User Authentication**: Login and registration with form validation
- 📚 **Manga Grid Display**: Beautiful card-based manga browser
- 🔍 **Advanced Search**: Filter by title, author, and genre
- 📖 **Library Management**: Track and update reading progress
- 💡 **Recommendations**: Personalized manga suggestions
- 📊 **Statistics Dashboard**: Visual reading statistics
- 📱 **Responsive Design**: Works on desktop, tablet, and mobile

### Installation & Usage

#### Prerequisites
- Modern web browser (Chrome, Firefox, Safari, Edge)
- MangaHub API server running on `localhost:8080`

#### Running the Web Client

Simply open `client/web/index.html` in your web browser, or serve it with a local server:

**Using Python:**
```bash
cd client/web
python -m http.server 3000
# Visit http://localhost:3000
```

**Using Node.js:**
```bash
cd client/web
npx http-server -p 3000
# Visit http://localhost:3000
```

**Using VS Code:**
- Install "Live Server" extension
- Right-click on `index.html` → "Open with Live Server"

### User Guide

#### Registration
1. Click "Register" tab
2. Fill in:
   - Username (3-30 characters)
   - Email (valid format)
   - Password (minimum 6 characters)
3. Click "Register" button
4. Upon success, automatically switched to login

#### Login
1. Enter your email and password
2. Click "Login" button
3. Session is saved in browser localStorage (persists after refresh)

#### Browse & Search
- View all manga in grid layout with cover images
- Use search box to filter by title, author, or genre
- Click on any manga card to view full details
- Add manga to your library with one click

#### Library Management
- Click "My Library" to view your collection
- Update progress with chapter number
- Change reading status (Reading/Completed)
- View reading statistics

#### Session Management
- Login state persists across page refreshes
- Click "Logout" to end session
- Automatic token management

---

## Development

### Project Structure
```
client/
├── README.md           # This file
├── cli/               # Command-line client
│   └── main.go        # Go CLI implementation
└── web/               # Web client
    └── index.html     # Single-page web app
```

### CLI Client Architecture
- **main.go**: Single-file implementation with:
  - HTTP client for API communication
  - TCP client for real-time sync
  - Terminal UI with colors and emojis
  - Input validation and error handling

### Web Client Architecture
- **index.html**: Single-page application with:
  - HTML structure
  - CSS styling (embedded)
  - JavaScript logic (embedded)
  - localStorage for session persistence

---

## Testing

### CLI Client Testing

1. **Start API Server:**
```bash
cd mangahub
go run cmd/api-server/main.go
```

2. **Start TCP Server (Optional):**
```bash
go run cmd/tcp-server/main.go
```

3. **Run CLI Client:**
```bash
cd client/cli
go run main.go
```

4. **Test Registration:**
- Try short password (< 6 chars) → Should show validation error
- Try valid password (>= 6 chars) → Should succeed

5. **Test TCP Sync:**
- Open two CLI clients
- Login on both
- Update progress on one → See notification on the other

### Web Client Testing

1. Start API server (same as above)
2. Open `client/web/index.html` in browser
3. Test registration with various inputs
4. Test login and session persistence (refresh page)
5. Test all features: browse, search, library, recommendations

---

## Troubleshooting

### CLI Client

**Problem**: "TCP sync unavailable (server offline)"
- **Solution**: Start the TCP server: `go run cmd/tcp-server/main.go`
- **Note**: TCP sync is optional; the client works without it

**Problem**: "Password must be at least 6 characters"
- **Solution**: Use a password with 6 or more characters
- **Example**: "password123" instead of "guy"

**Problem**: Colors not showing correctly
- **Solution**: Use a terminal with ANSI color support (Windows Terminal, iTerm2, etc.)

**Problem**: Connection refused
- **Solution**: Make sure API server is running on port 8080

### Web Client

**Problem**: CORS errors in browser console
- **Solution**: API server has CORS enabled; try using a local server instead of opening HTML directly

**Problem**: Login doesn't persist
- **Solution**: Check browser localStorage is enabled

**Problem**: Can't connect to API
- **Solution**: Verify API server is running and accessible at `http://localhost:8080`

---

## API Compatibility

Both clients are compatible with MangaHub API v1.0.0:
- Base URL: `http://localhost:8080`
- API Prefix: `/api/v1`
- Authentication: JWT Bearer tokens
- Content-Type: `application/json`

---

## Future Enhancements

### Planned Features
- [ ] WebSocket support for real-time updates in web client
- [ ] Offline mode with local caching
- [ ] Multi-language support
- [ ] Dark/light theme toggle
- [ ] Advanced filtering options
- [ ] Export library to JSON/CSV
- [ ] Batch operations (add multiple manga)
- [ ] Reading history timeline

---

## License

Part of the MangaHub project. See LICENSE file in the root directory.

## 2. Web Client (Frontend)

A beautiful, modern web interface built with vanilla HTML, CSS, and JavaScript.

### Features
- ✅ Responsive design with gradient backgrounds
- ✅ User authentication with JWT tokens
- ✅ Browse and search manga
- ✅ Personal library management
- ✅ Reading statistics dashboard
- ✅ Modal popups for manga details
- ✅ Real-time updates
- ✅ Local storage for session persistence

### Running the Web Client

#### Option 1: Simple HTTP Server (Python)
```bash
# Navigate to the web client directory
cd client/web

# Start a simple HTTP server
python -m http.server 8000

# Open browser to http://localhost:8000
```

#### Option 2: Live Server (VS Code Extension)
1. Install "Live Server" extension in VS Code
2. Right-click on `index.html`
3. Select "Open with Live Server"

#### Option 3: Direct File Access
Simply open `client/web/index.html` in your web browser.

### Web Client Features

#### 📱 Responsive Design
- Works on desktop, tablet, and mobile
- Beautiful gradient backgrounds
- Smooth animations and transitions

#### 🎨 Modern UI Components
- Card-based layouts
- Tab navigation
- Modal dialogs
- Loading spinners
- Success/Error alerts

#### 🔐 Authentication
- Login/Register tabs
- JWT token management
- Session persistence with localStorage

#### 📚 Manga Management
- **Browse Tab**: View popular manga in a grid layout
- **Search Tab**: Search manga by title
- **Library Tab**: Manage your personal collection
- **Stats Tab**: View reading statistics

### Web Client Screenshots

**Login Screen:**
```
┌────────────────────────────────────────┐
│  📚 MangaHub                           │
├────────────────────────────────────────┤
│  Login | Register                      │
│                                        │
│  Email: [________________]             │
│  Password: [________________]          │
│                                        │
│  [🔐 Login]                            │
└────────────────────────────────────────┘
```

**Browse Manga:**
```
┌─────────┬─────────┬─────────┐
│ One     │ Naruto  │ Bleach  │
│ Piece   │         │         │
│ ⭐⭐⭐   │ ⭐⭐⭐   │ ⭐⭐⭐   │
├─────────┼─────────┼─────────┤
│ Attack  │ Death   │ Hunter  │
│ on      │ Note    │ x       │
│ Titan   │         │ Hunter  │
│ ⭐⭐⭐   │ ⭐⭐⭐   │ ⭐⭐⭐   │
└─────────┴─────────┴─────────┘
```

## Usage Examples

### CLI Client Example Session

```bash
$ go run main.go

╔════════════════════════════════════════╗
║         MangaHub CLI Client            ║
╚════════════════════════════════════════╝

📚 Authentication Menu
1. Login
2. Register
3. Exit

Select an option: 1

🔐 Login
Email: admin@mangahub.com
Password: ********
✅ Login successful!

📚 Main Menu
1. Browse Manga
2. Search Manga
3. My Library
4. Get Recommendations
5. Logout

Select an option: 2

🔍 Search Manga
Enter search query: one

🔍 Found 14 manga matching 'one':

1. One Piece
   ✍️  Oda Eiichiro | 📚 1100 chapters | 🏷️  Action, Adventure, Shounen

2. Mob Psycho 100
   ✍️  ONE | 📚 101 chapters | 🏷️  Action, Comedy, Shounen
```

### Web Client Example Usage

1. **Open the web client** in your browser
2. **Register** a new account or **login** with existing credentials
3. **Browse** popular manga from the main screen
4. **Click on a manga card** to view detailed information
5. **Add to library** by clicking the "➕ Add to Library" button
6. **Switch to Library tab** to view your collection
7. **Check Stats tab** to see your reading statistics

## API Endpoints Used

Both clients use the following API endpoints:

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user

### Manga
- `GET /api/v1/manga/` - Search manga
- `GET /api/v1/manga/popular` - Get popular manga
- `GET /api/v1/manga/:id` - Get manga details

### User Library
- `GET /api/v1/users/library` - Get user's library
- `POST /api/v1/users/library` - Add manga to library
- `PUT /api/v1/users/progress` - Update reading progress
- `GET /api/v1/users/library/stats` - Get library statistics
- `GET /api/v1/users/recommendations` - Get recommendations

## Configuration

Both clients are configured to connect to:
- **API URL**: `http://localhost:8080/api/v1`

To change the API URL, modify:
- **CLI**: Change `baseURL` constant in `cli/main.go`
- **Web**: Change `API_URL` constant in `web/index.html`

## Troubleshooting

### Connection Refused
**Problem**: Cannot connect to API server

**Solution**: Make sure the API server is running:
```bash
cd mangahub
go run ./cmd/api-server
```

### CORS Errors (Web Client)
**Problem**: CORS policy blocking requests

**Solution**: The API server already has CORS middleware enabled. If issues persist:
1. Use the simple HTTP server method instead of opening the HTML file directly
2. Check that the API server's CORS middleware is properly configured

### Authentication Errors
**Problem**: "Unauthorized" or "Invalid token" errors

**Solution**:
- **CLI**: Logout and login again
- **Web**: Clear localStorage and refresh the page

### Token Expired
**Problem**: Token expires after 24 hours

**Solution**: Login again to get a new token

## Development

### Adding Features to CLI Client

Edit `client/cli/main.go` and add new menu options:

```go
func (c *Client) userMenu() {
    fmt.Println("1. Browse Manga")
    fmt.Println("2. Search Manga")
    fmt.Println("3. My Library")
    fmt.Println("4. Your New Feature")  // Add here
    
    choice := c.readInput()
    switch choice {
    case "4":
        c.yourNewFeature()  // Implement here
    }
}
```

### Adding Features to Web Client

Edit `client/web/index.html` and add new tabs or functionality:

```javascript
// Add new tab button
<button class="tab" onclick="switchAppTab('newfeature')">🆕 New Feature</button>

// Add new tab content
<div id="newfeatureTab" class="tab-content">
    <!-- Your content here -->
</div>

// Add JavaScript function
async function loadNewFeature() {
    // Your implementation
}
```

## Next Steps

- [ ] Add manga rating system
- [ ] Implement review/comment functionality
- [ ] Add reading history timeline
- [ ] Create manga comparison feature
- [ ] Add export library to JSON/CSV
- [ ] Implement dark/light theme toggle
- [ ] Add mobile app version
- [ ] Create browser extension

## License

Part of the MangaHub project.
