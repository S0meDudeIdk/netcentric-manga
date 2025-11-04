# MangaHub API Documentation

## Overview
MangaHub is a comprehensive manga tracking system built with Go and Gin framework. It provides RESTful APIs for user authentication, manga management, and reading progress tracking.

## Features
- User authentication with JWT tokens
- Manga CRUD operations (admin only)
- User library management
- Reading progress tracking
- Manga search and filtering
- User recommendations
- Rate limiting and security middleware

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login user

### User Management
- `GET /api/v1/users/profile` - Get user profile
- `GET /api/v1/users/library` - Get user's manga library
- `GET /api/v1/users/library/filtered` - Get filtered library with sorting
- `GET /api/v1/users/library/stats` - Get library statistics
- `GET /api/v1/users/recommendations` - Get manga recommendations
- `POST /api/v1/users/library` - Add manga to library
- `PUT /api/v1/users/progress` - Update reading progress
- `PUT /api/v1/users/progress/batch` - Batch update progress
- `DELETE /api/v1/users/library/:manga_id` - Remove manga from library

### Manga Management
- `GET /api/v1/manga` - Search manga with filters
- `GET /api/v1/manga/:id` - Get specific manga details
- `GET /api/v1/manga/genres` - Get all available genres
- `GET /api/v1/manga/popular` - Get popular manga
- `GET /api/v1/manga/stats` - Get manga statistics
- `POST /api/v1/manga` - Create new manga (admin only)
- `PUT /api/v1/manga/:id` - Update manga (admin only)
- `DELETE /api/v1/manga/:id` - Delete manga (admin only)

### Health Check
- `GET /health` - API health check

## Quick Start

### 1. Setup and Installation
```bash
# Clone the repository
git clone <repository-url>
cd netcentric-manga/mangahub

# Install dependencies
go mod tidy

# Run the server
go run cmd/api-server/main.go
```

### 2. Environment Variables
```bash
# Optional environment variables
export PORT=8080                    # Server port (default: 8080)
export GIN_MODE=release             # Gin mode (debug/release)
export JWT_SECRET=your-secret-key   # JWT secret key
```

### 3. Database
The application uses SQLite database which is automatically created in the `data/` directory when the server starts.

## API Usage Examples

### Authentication
```bash
# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "password123"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Manga Operations
```bash
# Search manga
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/manga?query=One%20Piece"

# Get manga details
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/manga/one-piece"

# Add manga to library
curl -X POST http://localhost:8080/api/v1/users/library \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "manga_id": "one-piece",
    "status": "reading"
  }'

# Update reading progress
curl -X PUT http://localhost:8080/api/v1/users/progress \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "manga_id": "one-piece",
    "current_chapter": 1095,
    "status": "reading"
  }'
```

## Testing

### Run Tests
```bash
# Run all tests
go test ./cmd/api-server -v

# Run tests with coverage
go test ./cmd/api-server -v -cover

# Run specific test
go test ./cmd/api-server -v -run TestHealthCheck
```

### Test Coverage
The test suite includes:
- Unit tests for all API endpoints
- Authentication and authorization tests
- Input validation tests
- Error handling tests
- Rate limiting tests

## Security Features

### Authentication
- JWT-based authentication
- Password hashing with bcrypt
- Token expiration (24 hours)

### Security Middleware
- Rate limiting (100 requests per minute)
- Request size limiting (10MB)
- Security headers (XSS protection, CSRF protection, etc.)
- CORS support

### Admin Access
Admin operations require special privileges. Currently, admin access is granted to:
- Users with email starting with "admin"
- Users with username ending with "admin"
- User with email "admin@mangahub.com"

## API Response Format

### Success Response
```json
{
  "data": {},
  "message": "Success"
}
```

### Error Response
```json
{
  "error": "Error message",
  "details": "Additional error details"
}
```

## Data Models

### User
```json
{
  "id": "string",
  "username": "string",
  "email": "string",
  "created_at": "timestamp"
}
```

### Manga
```json
{
  "id": "string",
  "title": "string",
  "author": "string",
  "genres": ["string"],
  "status": "ongoing|completed|hiatus|dropped|cancelled",
  "total_chapters": "number",
  "description": "string",
  "cover_url": "string"
}
```

### User Progress
```json
{
  "user_id": "string",
  "manga_id": "string",
  "current_chapter": "number",
  "status": "reading|completed|plan_to_read|dropped",
  "last_updated": "timestamp"
}
```

## Development

### Project Structure
```
mangahub/
├── cmd/
│   └── api-server/         # HTTP API server
├── internal/
│   ├── auth/              # Authentication logic
│   ├── manga/             # Manga service
│   ├── user/              # User service
├── pkg/
│   ├── database/          # Database utilities
│   ├── middleware/        # Custom middleware
│   ├── models/            # Data models
│   └── utils/             # Utility functions
├── data/                  # Data files and database
└── docs/                  # Documentation
```

### Adding New Features
1. Add new models in `pkg/models/`
2. Implement service logic in `internal/`
3. Add API endpoints in `cmd/api-server/`
4. Add tests in corresponding `*_test.go` files
5. Update documentation

## Troubleshooting

### Common Issues
1. **Database connection errors**: Ensure the `data/` directory exists and is writable
2. **Authentication failures**: Check JWT token format and expiration
3. **Rate limiting**: Reduce request frequency or increase rate limits
4. **CORS errors**: Update CORS configuration for your domain

### Logs
The application logs all requests and errors to stdout. For production, consider using a proper logging framework and log aggregation.

## Contributing
1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License
This project is licensed under the MIT License.