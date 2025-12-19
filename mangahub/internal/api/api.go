package api

import (
	"log"
	"mangahub/internal/external"
	grpcClient "mangahub/internal/grpc"
	"mangahub/internal/manga"
	"mangahub/internal/user"
	internalWebsocket "mangahub/internal/websocket"
	"mangahub/pkg/database"
	"mangahub/pkg/middleware"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// APIServer represents the HTTP API server
type APIServer struct {
	Router         *gin.Engine
	UserService    *user.Service
	MangaService   *manga.Service
	ChapterService *manga.ChapterService
	RatingService  *manga.RatingService
	SyncService    *manga.SyncService
	MALClient      *external.MALClient
	JikanClient    *external.JikanClient
	Port           string
	// WebSocket chat hub for manga-specific chats
	ChatHub *internalWebsocket.ChatHub
	// WebSocket upgrader
	upgrader internalWebsocket.Upgrader
	// HTTP client for communicating with standalone UDP server
	udpServerURL string
	tcpServerURL string
	httpClient   *http.Client
	// gRPC client for internal service calls
	GRPCClient *grpcClient.Client
}

// NewAPIServer creates a new API server instance
func NewAPIServer() *APIServer {
	// Set Gin mode from environment
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug" // Default to debug mode
	}
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add security headers
	router.Use(middleware.SecurityHeaders())

	// Add rate limiting from environment (default: 100 requests per minute)
	rateLimitStr := os.Getenv("RATE_LIMIT_REQUESTS_PER_MINUTE")
	rateLimit := 100 // Default
	if rateLimitStr != "" {
		if rl, err := strconv.Atoi(rateLimitStr); err == nil && rl > 0 {
			rateLimit = rl
		}
	}
	router.Use(middleware.CreateRateLimiter(rateLimit))

	// Add request size limit from environment (default: 10MB)
	maxSizeMBStr := os.Getenv("MAX_REQUEST_SIZE_MB")
	maxSizeMB := 10 // Default
	if maxSizeMBStr != "" {
		if size, err := strconv.Atoi(maxSizeMBStr); err == nil && size > 0 {
			maxSizeMB = size
		}
	}
	router.Use(middleware.RequestSizeLimit(int64(maxSizeMB * 1024 * 1024)))

	// Add CORS middleware
	router.Use(corsMiddleware())

	// Add request and response validation
	router.Use(middleware.RequestValidator())
	router.Use(middleware.ResponseValidator())

	// Add recovery middleware
	router.Use(gin.Recovery())

	jikanClient := external.NewJikanClient()

	server := &APIServer{
		Router:         router,
		UserService:    user.NewService(),
		MangaService:   manga.NewService(),
		ChapterService: manga.NewChapterService(),
		RatingService:  manga.NewRatingService(),
		SyncService:    manga.NewSyncService(jikanClient),
		MALClient:      external.NewMALClient(),
		JikanClient:    jikanClient,
		Port:           getPort(),
		ChatHub:        internalWebsocket.NewChatHub(),
		upgrader: internalWebsocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for development
				// In production, check against allowed origins
				return true
			},
		},
	}

	// Set manga service reference for chapter service
	server.ChapterService.SetMangaService(server.MangaService)

	// Start WebSocket chat hub
	// WebSocket rooms are created on demand when users join
	log.Println("WebSocket ChatHub initialized")

	server.httpClient = &http.Client{
		Timeout: 5 * time.Second,
	}

	server.initializeTCP()
	server.initializeUDP()

	// Connect to gRPC server
	go server.connectToGRPCServer()

	// Auto-sync manga from MAL on startup (in background)
	go server.autoSyncManga()

	// Setup routes
	server.setupRoutes()

	return server
}

// getPort returns the port from environment or default
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

// Health check endpoint
func (s *APIServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "mangahub-api",
		"version": "1.0.0",
	})
}

// Start starts the HTTP server
func (s *APIServer) Start() error {
	log.Printf("Starting API server on port %s", s.Port)
	// return s.Router.Run(":" + s.Port)
	return s.Router.Run("0.0.0.0:8080")
}

func main() {
	// Load environment variables from .env file
	// Try current directory first, then parent directory
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../../.env"); err != nil {
			log.Println("Warning: .env file not found, using environment variables or defaults")
		} else {
			log.Println("Loaded environment variables from ../../.env file")
		}
	} else {
		log.Println("Loaded environment variables from .env file")
	}

	// Initialize database
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Print current working directory for debugging
	if cwd, err := os.Getwd(); err == nil {
		log.Printf("Current working directory: %s", cwd)
	}

	// Create and start server
	// Note: Manga data is now fetched from external APIs (MAL/Jikan, MangaDex)
	// Local manga.json is no longer used
	server := NewAPIServer()

	log.Println("MangaHub API Server starting...")
	log.Printf("Server configuration:")
	log.Printf("  - Port: %s", server.Port)
	log.Printf("  - Gin Mode: %s", os.Getenv("GIN_MODE"))
	log.Printf("  - CORS Origins: %s", os.Getenv("CORS_ALLOW_ORIGINS"))
	log.Printf("  - Rate Limit: %s requests/min", os.Getenv("RATE_LIMIT_REQUESTS_PER_MINUTE"))

	// Log MAL API configuration
	if server.MALClient.IsConfigured() {
		log.Printf("  - MyAnimeList: Official API (Client ID configured) ✓")
	} else {
		log.Printf("  - MyAnimeList: Jikan API (unofficial) - Configure MAL_CLIENT_ID for official API")
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
