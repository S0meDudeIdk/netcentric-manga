# MangaHub - Network-Centric Manga Platform

A comprehensive manga reading and community platform showcasing modern network protocols and distributed systems. Features real-time chat, multi-protocol communication, and seamless MyAnimeList integration.

## 🌟 Overview

MangaHub is a full-stack manga platform built for a Network-Centric Computing course project. It demonstrates practical implementation of various network protocols, microservices architecture, and real-time communication systems.

### Key Features

#### 📚 Manga Reading & Discovery
- **External API Integration**: Fetch manga from MyAnimeList (via Jikan API), MangaDex, and Manga Plus
- **Browse & Search**: Explore extensive manga collections with filtering and sorting
- **Chapter Reading**: Built-in reader with progress tracking
- **Personal Library**: Organize your manga collection by reading status
- **Recommendations**: Get personalized manga suggestions

#### 💬 Real-Time Communication
- **WebSocket Chat**: Room-based chat system for each manga
- **General Chat**: Global community chat for all users
- **Live Notifications**: Real-time updates for library changes
- **User Presence**: See who's online in chat rooms

#### 🔌 Multi-Protocol Architecture
- **HTTP/REST**: Main API server (Port 8080)
- **gRPC**: High-performance RPC for library operations (Port 9003)
- **TCP**: Persistent sync server for real-time updates (Port 9001)
- **UDP**: Lightweight event notifications (Port 9002)
- **WebSocket**: Real-time bidirectional chat (Integrated in API server)

#### 🔐 Flexible Access Model
- **Public Access**: Browse and view manga without login
- **Authenticated Features**: Library management, reading progress, chat, and recommendations
- **JWT Authentication**: Secure token-based auth with 24-hour sessions

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- Node.js 14+
- SQLite (included)

### Option 1: Manual Setup (Development)

#### 1. Start Backend Servers
```bash
cd mangahub

# Terminal 1 - TCP Server
cd cmd/tcp-server && go run main.go

# Terminal 2 - UDP Server
cd cmd/udp-server && go run main.go

# Terminal 3 - gRPC Server
cd cmd/grpc-server && go run main.go

# Terminal 4 - Main API Server
cd cmd/api-server && go run main.go

# Terminal 5 - Fetch Manga Server (optional)
cd cmd/fetch-manga-server && go run main.go
```

#### 2. Start Web Client
```bash
cd mangahub/client/web-react
npm install
npm start
```

Visit http://localhost:3000

#### 3. Or Use CLI Client
```bash
cd mangahub/client/cli
go run main.go
```

### Option 2: Docker Compose (Production)

```bash
cd mangahub
docker-compose up -d
```

Services will be available at:
- **Web App**: http://localhost:3000
- **API Server**: http://localhost:8080
- **gRPC Server**: localhost:9003
- **TCP Server**: localhost:9001
- **UDP Server**: localhost:9002
- **Fetch Manga Server**: http://localhost:8082

## 🏗️ Architecture

### Protocol Usage

| Protocol | Purpose | Port | Use Case |
|----------|---------|------|----------|
| HTTP/REST | Main API, CRUD operations | 8080 | Manga search, user management, authentication |
| gRPC | Library operations | 9003 | Fast library queries, ratings, progress updates |
| TCP | Persistent sync | 9001 | Real-time library synchronization |
| UDP | Event notifications | 9002 | Lightweight status broadcasts |
| WebSocket | Chat system | 8080 | Real-time chat, user presence, notifications |

## 📁 Project Structure

```
mangahub/
├── cmd/                          # Server entry points
│   ├── api-server/               # Main HTTP API server
│   ├── grpc-server/              # gRPC service
│   ├── tcp-server/               # TCP sync server
│   ├── udp-server/               # UDP notification server
│   └── fetch-manga-server/       # External API aggregator
├── internal/                     # Business logic
│   ├── api/                      # HTTP handlers and routes
│   ├── auth/                     # JWT authentication
│   ├── external/                 # External API clients (MAL, MangaDex)
│   ├── grpc/                     # gRPC implementation
│   ├── manga/                    # Manga services
│   ├── tcp/                      # TCP protocol handlers
│   ├── udp/                      # UDP protocol handlers
│   ├── user/                     # User & library services
│   └── websocket/                # WebSocket chat hub
├── pkg/                          # Shared packages
│   ├── database/                 # SQLite database layer
│   ├── middleware/               # Rate limiting, validation
│   ├── models/                   # Data structures
│   └── utils/                    # Helper functions
├── proto/                        # Protocol Buffer definitions
├── client/
│   ├── cli/                      # Command-line client
│   │   └── protocol/             # Protocol implementations
│   └── web-react/                # React web application
│       ├── src/
│       │   ├── components/       # Reusable UI components
│       │   ├── pages/            # Route pages
│       │   └── services/         # API clients
│       └── Dockerfile            # Production build
├── docker-compose.yml            # Multi-service orchestration
└── Dockerfile                    # Go services image
```

## 🎯 Features in Detail

### Web Application (React)

**Pages:**
- **Home**: Landing page with featured manga
- **Browse**: Grid view with genre filters and sorting
- **Search**: Search by title, author, or genre
- **Library**: Personal collection with status filters (Reading, Completed, Plan to Read, etc.)
- **Manga Detail**: Full information with cover, description, chapters, and ratings
- **Chapter Reader**: Full-screen reading experience
- **Chat**: Real-time manga-specific and general chat rooms
- **Profile**: User settings and library statistics

**Components:**
- Responsive navigation with auth state
- Animated manga cards with hover effects
- Loading spinners and error states
- Real-time notification toasts
- Mobile-responsive design

### CLI Client (Go)

- Interactive menu-driven interface
- Support for all network protocols (HTTP, gRPC, TCP, UDP, WebSocket)
- Manga browsing and search
- Library management
- Protocol selection for operations
- Colored output for better UX

### Authentication System

- **Registration**: Username, email, password (min 6 chars)
- **Login**: Email or username + password
- **JWT Tokens**: 24-hour expiration, HS256 signing
- **Protected Routes**: Middleware validation
- **Optional Auth**: Public endpoints work without login

### External API Integration

#### MyAnimeList (via Jikan API v4)
- Search manga by title
- Get manga details, statistics, and recommendations
- Browse top/popular manga
- Rate limiting: 1 request per second

#### MangaDex
- Extensive manga catalog
- Chapter listings
- Cover art and metadata

#### Manga Plus
- Official Shueisha releases
- Latest chapters

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the `mangahub` directory:

```bash
# Server Configuration
PORT=8080
GIN_MODE=release

# Authentication
JWT_SECRET=your-secret-key-here

# CORS
CORS_ALLOW_ORIGINS=*
CORS_ALLOW_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOW_HEADERS=Origin,Content-Type,Authorization

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=100
MAX_REQUEST_SIZE_MB=10

# External APIs
JIKAN_API_BASE_URL=https://api.jikan.moe/v4
JIKAN_RATE_LIMIT_SECONDS=1
MANGADEX_API_BASE_URL=https://api.mangadex.org
MANGADEX_API_TIMEOUT=15
MANGAPLUS_API_BASE_URL=https://jumpg-webapi.tokyo-cdn.com/api

# Optional: MyAnimeList Official API
MAL_CLIENT_ID=your-client-id
MAL_CLIENT_SECRET=your-client-secret

# Server Addresses (for Docker)
TCP_SERVER_ADDR=tcp-server:9001
UDP_SERVER_ADDR=udp-server:9002
GRPC_SERVER_ADDR=grpc-server:9003
```

## 🧪 Testing

### Manual Testing
```bash
# Test HTTP API
curl http://localhost:8080/api/v1/manga

# Test WebSocket (requires websocat or similar)
websocat ws://localhost:8080/ws/chat?token=YOUR_JWT_TOKEN

# Use CLI for protocol testing
cd mangahub/client/cli
go run main.go
```

### Automated Tests
```bash
cd mangahub
go test ./...
```

## 🛠️ Technology Stack

### Backend
- **Language**: Go 1.24
- **Web Framework**: Gin (HTTP server)
- **Database**: SQLite with database/sql
- **Authentication**: JWT (golang-jwt/jwt)
- **gRPC**: Protocol Buffers, google.golang.org/grpc
- **WebSocket**: gorilla/websocket
- **Password Hashing**: bcrypt (golang.org/x/crypto)

### Frontend
- **Framework**: React 19
- **Routing**: React Router v7
- **Styling**: Tailwind CSS 3.4
- **Animations**: Framer Motion 12
- **Icons**: Lucide React
- **HTTP Client**: Axios
- **Build Tool**: React Scripts (Create React App)

### DevOps
- **Containerization**: Docker & Docker Compose
- **Reverse Proxy**: Nginx (for React production)
- **Environment**: godotenv for configuration

## 📚 API Documentation

### Authentication Endpoints
- `POST /api/v1/auth/register` - Create new account
- `POST /api/v1/auth/login` - Login and get JWT

### Manga Endpoints (Public)
- `GET /api/v1/manga` - List all manga
- `GET /api/v1/manga/search` - Search manga
- `GET /api/v1/manga/:id` - Get manga details
- `GET /api/v1/manga/:id/chapters` - Get chapters

### User/Library Endpoints (Protected)
- `GET /api/v1/users/profile` - Get user profile
- `PUT /api/v1/users/profile` - Update profile
- `GET /api/v1/users/library` - Get user's library
- `POST /api/v1/users/library` - Add manga to library
- `PUT /api/v1/users/library/:id` - Update reading progress
- `DELETE /api/v1/users/library/:id` - Remove from library

### WebSocket Endpoints
- `WS /ws/chat?token=JWT` - General chat room
- `WS /ws/manga/:id?token=JWT` - Manga-specific chat room

## 🤝 Contributing

This is an educational project for a Network-Centric Computing course. Contributions are welcome for learning purposes!

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👥 Authors

Created as a Network-Centric Computing course project demonstrating:
- Microservices architecture
- Multi-protocol communication
- Real-time systems
- RESTful API design
- gRPC implementation
- WebSocket chat systems
- External API integration
- Modern web development practices
