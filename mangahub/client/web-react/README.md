# MangaHub React Web Client

A modern React-based web client for the MangaHub API, built with React, Tailwind CSS, Framer Motion, and Lucide Icons.

## 📋 Prerequisites

- Node.js 14+ and npm
- MangaHub API server running on `localhost:8080`

## 🚀 Quick Start

### 1. Install Dependencies

```bash
cd client/web-react
npm install
```

### 2. Start Development Server

```bash
npm start
```

The app will open at `http://localhost:3000`

### 3. Build for Production

```bash
npm run build
```

## ✅ Complete Component List

All components have been created and are ready to use!

### Components (`src/components/`)
- ✅ **Header.jsx** - Navigation with auth state, user menu, mobile responsive
- ✅ **Footer.jsx** - Footer with links and copyright
- ✅ **MangaCard.jsx** - Manga display card with hover animations
- ✅ **LoadingSpinner.jsx** - Animated loading indicator

### Pages (`src/pages/`)
- ✅ **Home.jsx** - Landing page with featured manga and CTAs
- ✅ **Login.jsx** - User login form with validation
- ✅ **Register.jsx** - User registration with password confirmation
- ✅ **Browse.jsx** - Browse all manga with genre filters and sorting
- ✅ **Search.jsx** - Search manga by title, author, or genre
- ✅ **Library.jsx** - User's manga library with stats and status filters
- ✅ **MangaDetail.jsx** - Individual manga details with library management

### Services (`src/services/`)
- ✅ **authService.js** - Authentication (login, register, logout, token management)
- ✅ **mangaService.js** - Manga API calls (search, browse, details, stats)
- ✅ **userService.js** - User/library operations (add, update, progress tracking)

## 📁 Project Structure

```
web-react/
├── public/              # Static files
├── src/
│   ├── components/      # Reusable components
│   │   ├── Header.jsx
│   │   ├── Footer.jsx
│   │   ├── MangaCard.jsx
│   │   └── LoadingSpinner.jsx
│   ├── pages/          # Page components
│   │   ├── Home.jsx
│   │   ├── Login.jsx
│   │   ├── Register.jsx
│   │   ├── Browse.jsx
│   │   ├── Search.jsx
│   │   ├── Library.jsx
│   │   └── MangaDetail.jsx
│   ├── services/       # API services
│   │   ├── authService.js
│   │   ├── mangaService.js
│   │   └── userService.js
│   ├── App.js          # Main app with routing
│   ├── index.js        # Entry point
│   └── index.css       # Global styles
├── package.json
└── tailwind.config.js
```

## 🎨 Features

### For Everyone (No Login Required)
- ✅ **Browse Manga**: View entire manga collection freely
- ✅ **Search**: Find manga by title, author, or genre
- ✅ **View Details**: See full manga information, descriptions, and stats
- ✅ **Filter & Sort**: Organize manga by genre, popularity, etc.
- ✅ **Modern UI**: Beautiful interface with Tailwind CSS
- ✅ **Smooth Animations**: Enhanced UX with Framer Motion
- ✅ **Responsive Design**: Works perfectly on mobile, tablet, and desktop

### With an Account (Free Registration)
- ✅ **Personal Library**: Save manga to your collection
- ✅ **Track Progress**: Mark chapters read and current progress
- ✅ **Continue Reading**: Pick up exactly where you left off
- ✅ **Reading Status**: Organize by reading, completed, plan to read, etc.
- ✅ **Reading Lists**: Create custom playlists and collections
- ✅ **Statistics**: View your reading stats and history
- ✅ **Recommendations**: Get personalized manga suggestions

## 📦 Dependencies

```json
{
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.20.0",
    "axios": "^1.6.2",
    "lucide-react": "^0.294.0",
    "framer-motion": "^10.16.5"
  },
  "devDependencies": {
    "tailwindcss": "^3.3.5",
    "postcss": "^8.4.32",
    "autoprefixer": "^10.4.16"
  }
}
```

## 🔧 Configuration

### API Base URL

The app automatically detects the environment:
- **Development**: `http://localhost:8080/api/v1`
- **Production**: Uses `REACT_APP_BACKEND_URL` environment variable

To set a custom API URL, create a `.env` file:

```env
REACT_APP_BACKEND_URL=https://your-api-url.com
```

### Tailwind CSS

Configuration in `tailwind.config.js`:
```javascript
module.exports = {
  content: ["./src/**/*.{js,jsx,ts,tsx}"],
  darkMode: 'class',
  theme: {
    extend: {},
  },
  plugins: [],
}
```

## 🎯 Usage Guide

### Browsing Without an Account

1. **Browse Manga**: Visit `/browse` to see all available manga
   - Filter by genre
   - Sort by popularity, title, chapters, or year
   - No login required!

2. **Search**: Use `/search` to find specific manga
   - Search by title, author, or genre
   - Instant results
   - Available to everyone

3. **View Details**: Click any manga card
   - See full description, author, genres
   - View publication info and stats
   - Check total chapters available
   - **Login prompt** appears for library features

### Creating an Account (Optional but Recommended)

1. **Register**: Create a free account at `/register`
   - Username (3-30 characters)
   - Email (valid format)
   - Password (minimum 6 characters)

2. **Login**: Sign in at `/login`
   - Email and password
   - Token stored in localStorage
   - Stay logged in across sessions

3. **Logout**: Click user menu → Logout

### Library Management (Requires Account)

- **Add to Library**: Click "Add to Library" on any manga detail page
- **Track Progress**: Update current chapter as you read
- **Update Status**: Mark as reading, completed, plan to read, on hold, or dropped
- **Continue Reading**: Your progress is saved automatically
- **View Stats**: See your reading statistics on library page
- **Filter by Status**: View manga by reading, completed, plan to read, etc.
- **View Stats**: See your reading statistics on library page
- **Filter by Status**: View manga by reading, completed, plan to read, etc.

## 🚀 Available Scripts

### `npm start`
Runs the app in development mode at http://localhost:3000

### `npm test`
Launches the test runner in interactive watch mode

### `npm run build`
Builds the app for production to the `build` folder

### `npm run eject`
**Note: this is a one-way operation!** Ejects from Create React App

## 🐛 Troubleshooting

### Port 3000 Already in Use
```bash
# Windows PowerShell
$env:PORT=3001; npm start

# Linux/Mac
PORT=3001 npm start
```

### CORS Errors
Make sure your Go API server has CORS enabled (it already does in `cmd/api-server/main.go`).

### Authentication Issues
- Check browser console for errors
- Verify token in localStorage: `localStorage.getItem('token')`
- Ensure API server is running on port 8080

### API Connection Issues
- Verify API server is running: `http://localhost:8080/api/v1/manga`
- Check network tab in browser DevTools
- Ensure no firewall blocking localhost connections

## 🔗 Integration with Backend

The React app is configured to work with your existing Go backend:

- **Auth**: `/api/v1/auth/login`, `/api/v1/auth/register`
- **Manga**: `/api/v1/manga/*`
- **Users**: `/api/v1/users/*`

No backend changes needed! The app uses the same API endpoints as the CLI and HTML clients.

## 📊 Comparison with HTML Version

| Feature | HTML Version | React Version |
|---------|--------------|---------------|
| Framework | Vanilla JS | React + Router |
| Styling | CSS | Tailwind CSS |
| Animations | None | Framer Motion |
| State Management | localStorage | React state + Context |
| Routing | Hash-based | React Router |
| Code Organization | Single file | Component-based |
| Build Process | None | Webpack via CRA |
| Icons | None | Lucide React |
| Mobile Responsive | Basic | Full responsive design |

## 🚀 Deployment

### Build for Production
```bash
npm run build
```

### Serve Built Files
```bash
# Using serve
npx serve -s build -p 3000

# Or copy build/ folder to your web server
```

### Deploy to Production
1. Build the app: `npm run build`
2. Set environment variable: `REACT_APP_BACKEND_URL=https://your-api.com`
3. Upload `build/` folder to your hosting service
4. Configure web server to serve `index.html` for all routes (for React Router)

## 📚 Resources

- [React Documentation](https://react.dev/)
- [React Router](https://reactrouter.com/)
- [Tailwind CSS](https://tailwindcss.com/)
- [Framer Motion](https://www.framer.com/motion/)
- [Lucide Icons](https://lucide.dev/)
- [Axios](https://axios-http.com/)

## 🤝 Development Tips

### Adding New Pages
1. Create component in `src/pages/`
2. Add route in `src/App.js`
3. Add navigation link in `src/components/Header.jsx`

### Adding New API Calls
1. Add method to appropriate service in `src/services/`
2. Use in components with `async/await`
3. Handle loading and error states

### Styling Guidelines
- Use Tailwind utility classes
- Follow mobile-first responsive design
- Use Framer Motion for animations
- Use Lucide React for icons

## 📄 License

Same license as the main MangaHub project.

---

**Ready to use!** All components are complete. Just run `npm start` and your React app is ready! 🎉

### Analyzing the Bundle Size

This section has moved here: [https://facebook.github.io/create-react-app/docs/analyzing-the-bundle-size](https://facebook.github.io/create-react-app/docs/analyzing-the-bundle-size)

### Making a Progressive Web App

This section has moved here: [https://facebook.github.io/create-react-app/docs/making-a-progressive-web-app](https://facebook.github.io/create-react-app/docs/making-a-progressive-web-app)

### Advanced Configuration

This section has moved here: [https://facebook.github.io/create-react-app/docs/advanced-configuration](https://facebook.github.io/create-react-app/docs/advanced-configuration)

### Deployment

This section has moved here: [https://facebook.github.io/create-react-app/docs/deployment](https://facebook.github.io/create-react-app/docs/deployment)

### `npm run build` fails to minify

This section has moved here: [https://facebook.github.io/create-react-app/docs/troubleshooting#npm-run-build-fails-to-minify](https://facebook.github.io/create-react-app/docs/troubleshooting#npm-run-build-fails-to-minify)
