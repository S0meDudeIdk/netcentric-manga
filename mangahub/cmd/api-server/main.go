package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mangahub/internal/auth"
	"mangahub/internal/external"
	grpcClient "mangahub/internal/grpc"
	"mangahub/internal/manga"
	"mangahub/internal/user"
	internalWebsocket "mangahub/internal/websocket"
	"mangahub/pkg/database"
	"mangahub/pkg/middleware"
	"mangahub/pkg/models"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// APIServer represents the HTTP API server
type APIServer struct {
	Router         *gin.Engine
	UserService    *user.Service
	MangaService   *manga.Service
	ChapterService *manga.ChapterService
	RatingService  *manga.RatingService
	MALClient      *external.MALClient
	JikanClient    *external.JikanClient
	Port           string
	// TCP client connection to broadcast progress updates
	tcpConn net.Conn
	tcpMu   sync.Mutex
	// WebSocket room hub for manga-specific chats
	RoomHub *internalWebsocket.RoomHub
	// WebSocket upgrader
	upgrader websocket.Upgrader
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

	// Add request logging middleware
	router.Use(gin.Logger())

	// Add recovery middleware
	router.Use(gin.Recovery())

	server := &APIServer{
		Router:         router,
		UserService:    user.NewService(),
		MangaService:   manga.NewService(),
		ChapterService: manga.NewChapterService(),
		RatingService:  manga.NewRatingService(),
		MALClient:      external.NewMALClient(),
		JikanClient:    external.NewJikanClient(),
		Port:           getPort(),
		RoomHub:        internalWebsocket.NewRoomHub(),
		upgrader: websocket.Upgrader{
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
	log.Println("WebSocket RoomHub initialized")

	// Connect to TCP server for broadcasting progress updates
	go server.connectToTCPServer()

	// Connect to gRPC server
	go server.connectToGRPCServer()

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

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request origin
		origin := c.Request.Header.Get("Origin")

		// Get allowed origins from environment or use default
		allowedOriginsStr := os.Getenv("CORS_ALLOW_ORIGINS")
		if allowedOriginsStr == "" {
			allowedOriginsStr = "*" // Default to allow all
		}

		// Determine which origin to allow
		allowOrigin := ""
		if allowedOriginsStr == "*" {
			allowOrigin = "*"
		} else {
			// Split comma-separated origins and check if request origin is allowed
			allowedOrigins := strings.Split(allowedOriginsStr, ",")
			for _, allowed := range allowedOrigins {
				allowed = strings.TrimSpace(allowed)
				if allowed == origin {
					allowOrigin = origin
					break
				}
			}
			// If origin not found in allowed list, don't set the header
			if allowOrigin == "" && len(allowedOrigins) > 0 {
				// For development, allow the first origin if origin is not in the list
				allowOrigin = strings.TrimSpace(allowedOrigins[0])
			}
		}

		// Get allowed methods from environment or use default
		allowedMethods := os.Getenv("CORS_ALLOW_METHODS")
		if allowedMethods == "" {
			allowedMethods = "POST, OPTIONS, GET, PUT, DELETE"
		}

		// Get allowed headers from environment or use default
		allowedHeaders := os.Getenv("CORS_ALLOW_HEADERS")
		if allowedHeaders == "" {
			allowedHeaders = "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"
		}

		// Set CORS headers - only set one origin value
		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Headers", allowedHeaders)
		c.Header("Access-Control-Allow-Methods", allowedMethods)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// optionalAuthMiddleware extracts user info from token if present, but doesn't require it
func optionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// Try to get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				token = tokenParts[1]
			}
		}

		// If token exists, validate and set user info
		if token != "" {
			claims, err := auth.ValidateToken(token)
			if err == nil {
				// Store user info in context
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("email", claims.Email)
			}
		}

		// Continue regardless of authentication status
		c.Next()
	}
}

// authMiddleware validates JWT tokens
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// Try to get token from Authorization header first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Extract token from "Bearer <token>"
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				token = tokenParts[1]
			}
		}

		// If no token in header, try query parameter (for WebSocket connections)
		if token == "" {
			token = c.Query("token")
		}

		// If still no token, return error
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := auth.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// adminMiddleware checks if user has admin privileges
func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For now, we'll use a simple check for admin email or username
		// In production, you'd have proper role-based access control
		email := c.GetString("email")
		username := c.GetString("username")

		// Simple admin check - you can customize this logic
		isAdmin := email == "admin@mangahub.com" || username == "admin" ||
			strings.HasPrefix(email, "admin") || strings.HasSuffix(username, "admin")

		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// setupRoutes configures all API routes
func (s *APIServer) setupRoutes() {
	// Health check
	s.Router.GET("/health", s.healthCheck)

	// API version 1
	v1 := s.Router.Group("/api/v1")
	{
		// Auth routes (no middleware)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", s.register)
			auth.POST("/login", s.login)
		}

		// Public manga browsing routes (no auth required)
		publicManga := v1.Group("/manga")
		{
			publicManga.GET("/", s.searchManga)
			publicManga.GET("/genres", s.getGenres)
			publicManga.GET("/popular", s.getPopularManga)
			publicManga.GET("/stats", s.getMangaStats)

			// MAL/Jikan API integration routes
			publicManga.GET("/mal/search", s.searchMAL)
			publicManga.GET("/mal/top", s.getTopMAL)
			publicManga.GET("/mal/:mal_id", s.getMALManga)
			publicManga.GET("/mal/:mal_id/recommendations", s.getMALRecommendations)

			// MangaDex search routes
			publicManga.GET("/mangadex/search", s.searchMangaDex)

			// Chapter routes (public - no auth required for reading)
			// These must come before /:id to avoid route conflicts
			publicManga.GET("/chapters/:chapter_id/pages", s.getChapterPages)

			// This must be last to avoid conflicts with specific routes above
			publicManga.GET("/:id", s.getManga)
			publicManga.GET("/:id/chapters", s.getChapterList)
			// Use optional auth for ratings to return user-specific rating if authenticated
			publicManga.GET("/:id/ratings", optionalAuthMiddleware(), s.getMangaRatings)
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(authMiddleware())
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/profile", s.getProfile)
				users.GET("/library", s.getLibrary)
				users.GET("/library/filtered", s.getFilteredLibrary)
				users.GET("/library/stats", s.getLibraryStats)
				users.GET("/recommendations", s.getRecommendations)
				users.POST("/library", s.addToLibrary)
				users.PUT("/progress", s.updateProgress)
				users.PUT("/progress/batch", s.batchUpdateProgress)
				users.DELETE("/library/:manga_id", s.removeFromLibrary)
				// Rating routes (protected)
				users.POST("/manga/:manga_id/rating", s.rateManga)
				users.DELETE("/manga/:manga_id/rating", s.deleteRating)
			}

			// Admin routes for manga management
			manga := protected.Group("/manga")
			{
				adminManga := manga.Group("/")
				adminManga.Use(adminMiddleware())
				{
					adminManga.POST("/", s.createManga)
					adminManga.PUT("/:id", s.updateManga)
					adminManga.DELETE("/:id", s.deleteManga)

					// Bulk data import and management operations
					adminManga.POST("/bulk-import", s.bulkImportManga)
					adminManga.POST("/validate-data", s.validateMangaData)
					adminManga.GET("/import-stats", s.getImportStats)
					adminManga.DELETE("/bulk-delete", s.bulkDeleteManga)
				}
			}

			// WebSocket chat endpoint (protected - requires authentication)
			protected.GET("/ws/chat", s.handleWebSocketChat)

			// gRPC-backed endpoints (protected)
			grpcRoutes := protected.Group("/grpc")
			{
				grpcRoutes.GET("/manga/:id", s.getMangaViaGRPC)
				grpcRoutes.GET("/manga/search", s.searchMangaViaGRPC)
				grpcRoutes.PUT("/progress/update", s.updateProgressViaGRPC)
			}
		}

		// WebSocket stats endpoint (public for monitoring)
		v1.GET("/ws/stats", s.getWebSocketStats)
	}
}

// Health check endpoint
func (s *APIServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "mangahub-api",
		"version": "1.0.0",
	})
}

// Register endpoint
func (s *APIServer) register(c *gin.Context) {
	var req models.UserRegistration
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := s.UserService.Register(req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			log.Printf("Registration error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login endpoint
func (s *APIServer) login(c *gin.Context) {
	var req models.UserLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := s.UserService.Login(req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		} else {
			log.Printf("Login error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// Get profile endpoint
func (s *APIServer) getProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	profile, err := s.UserService.GetProfile(userID)
	if err != nil {
		log.Printf("Get profile error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// Get library endpoint
func (s *APIServer) getLibrary(c *gin.Context) {
	userID := c.GetString("user_id")

	library, err := s.UserService.GetLibrary(userID)
	if err != nil {
		log.Printf("Get library error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, library)
}

// Add to library endpoint
func (s *APIServer) addToLibrary(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.AddToLibraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.UserService.AddToLibrary(userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("Add to library error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Manga added to library successfully"})
}

// Update progress endpoint
func (s *APIServer) updateProgress(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.UserService.UpdateProgress(userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("Update progress error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Broadcast progress update to TCP server
	go s.broadcastProgressUpdate(userID, req.MangaID, req.CurrentChapter)

	c.JSON(http.StatusOK, gin.H{"message": "Progress updated successfully"})
}

// Get filtered library endpoint
func (s *APIServer) getFilteredLibrary(c *gin.Context) {
	userID := c.GetString("user_id")

	// Parse query parameters
	req := models.LibraryFilterRequest{
		Status: c.Query("status"),
		SortBy: c.Query("sort_by"),
		Limit:  20,
		Offset: 0,
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	progressList, err := s.UserService.GetFilteredLibrary(userID, req.Status, req.SortBy, req.Limit, req.Offset)
	if err != nil {
		log.Printf("Get filtered library error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"progress": progressList,
		"count":    len(progressList),
	})
}

// Get library stats endpoint
func (s *APIServer) getLibraryStats(c *gin.Context) {
	userID := c.GetString("user_id")

	stats, err := s.UserService.GetLibraryStats(userID)
	if err != nil {
		log.Printf("Get library stats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Get recommendations endpoint
func (s *APIServer) getRecommendations(c *gin.Context) {
	userID := c.GetString("user_id")

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 20 {
			limit = l
		}
	}

	recommendations, err := s.UserService.GetReadingRecommendations(userID, limit)
	if err != nil {
		log.Printf("Get recommendations error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
		"count":           len(recommendations),
	})
}

// Batch update progress endpoint
func (s *APIServer) batchUpdateProgress(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.BatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.UserService.BatchUpdateProgress(userID, req.Updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("Batch update progress error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Progress updated successfully",
		"updated": len(req.Updates),
	})
}

// Remove from library endpoint
func (s *APIServer) removeFromLibrary(c *gin.Context) {
	userID := c.GetString("user_id")
	mangaID := c.Param("manga_id")

	err := s.UserService.RemoveFromLibrary(userID, mangaID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("Remove from library error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Manga removed from library successfully"})
}

// Search manga endpoint
func (s *APIServer) searchManga(c *gin.Context) {
	// Parse query parameters
	req := models.MangaSearchRequest{
		Query:  c.Query("query"),
		Author: c.Query("author"),
		Status: c.Query("status"),
		Limit:  20,
		Offset: 0,
	}

	// Parse limit and offset
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	// Parse genres
	if genresStr := c.Query("genres"); genresStr != "" {
		req.Genres = strings.Split(genresStr, ",")
		for i := range req.Genres {
			req.Genres[i] = strings.TrimSpace(req.Genres[i])
		}
	}

	// Search manga
	mangaList, err := s.MangaService.SearchManga(req)
	if err != nil {
		log.Printf("Search manga error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"manga": mangaList,
		"count": len(mangaList),
	})
}

// Get manga endpoint
func (s *APIServer) getManga(c *gin.Context) {
	mangaID := c.Param("id")

	manga, err := s.MangaService.GetManga(mangaID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			log.Printf("Get manga error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, manga)
}

// Get genres endpoint
func (s *APIServer) getGenres(c *gin.Context) {
	genres, err := s.MangaService.GetAllGenres()
	if err != nil {
		log.Printf("Get genres error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"genres": genres,
		"count":  len(genres),
	})
}

// Get popular manga endpoint
func (s *APIServer) getPopularManga(c *gin.Context) {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	mangaList, err := s.MangaService.GetPopularManga(limit)
	if err != nil {
		log.Printf("Get popular manga error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"manga": mangaList,
		"count": len(mangaList),
	})
}

// Get manga stats endpoint
func (s *APIServer) getMangaStats(c *gin.Context) {
	stats, err := s.MangaService.GetMangaStats()
	if err != nil {
		log.Printf("Get manga stats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Create manga endpoint (admin only)
func (s *APIServer) createManga(c *gin.Context) {
	var req models.CreateMangaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert request to manga model
	manga := models.Manga{
		ID:            req.ID,
		Title:         req.Title,
		Author:        req.Author,
		Genres:        req.Genres,
		Status:        req.Status,
		TotalChapters: req.TotalChapters,
		Description:   req.Description,
		CoverURL:      req.CoverURL,
	}

	createdManga, err := s.MangaService.CreateManga(manga)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			log.Printf("Create manga error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, createdManga)
}

// Update manga endpoint (admin only)
func (s *APIServer) updateManga(c *gin.Context) {
	mangaID := c.Param("id")

	var req models.UpdateMangaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert request to manga model
	manga := models.Manga{
		Title:         req.Title,
		Author:        req.Author,
		Genres:        req.Genres,
		Status:        req.Status,
		TotalChapters: req.TotalChapters,
		Description:   req.Description,
		CoverURL:      req.CoverURL,
	}

	updatedManga, err := s.MangaService.UpdateManga(mangaID, manga)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			log.Printf("Update manga error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, updatedManga)
}

// Delete manga endpoint (admin only)
func (s *APIServer) deleteManga(c *gin.Context) {
	mangaID := c.Param("id")

	err := s.MangaService.DeleteManga(mangaID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found"})
		} else {
			log.Printf("Delete manga error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Manga deleted successfully"})
}

// MAL/Jikan API handlers

// searchMAL searches manga from MyAnimeList via Jikan API
// searchMAL searches manga from MyAnimeList (tries official API first, falls back to Jikan)
func (s *APIServer) searchMAL(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// For pagination-heavy operations like searching, prefer Jikan API as it provides
	// complete pagination info (total count, last_visible_page, etc.)
	// Official MAL API only provides next/previous links without totals

	// Commented out MAL API to prefer Jikan for better pagination
	// if s.MALClient.IsConfigured() {
	// 	// Get page parameter for offset calculation
	// 	page := 1
	// 	if pageStr := c.Query("page"); pageStr != "" {
	// 		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
	// 			page = p
	// 		}
	// 	}
	//
	// 	// Calculate offset from page number
	// 	offset := (page - 1) * limit
	//
	// 	malResult, err := s.MALClient.SearchManga(query, limit, offset)
	// 	if err != nil {
	// 		log.Printf("Official MAL API search error: %v, falling back to Jikan", err)
	// 	} else {
	// 		// Convert official MAL data
	// 		mangaList := make([]external.MALMangaNode, 0, len(malResult.Data))
	// 		for _, item := range malResult.Data {
	// 			mangaList = append(mangaList, item.Node)
	// 		}
	// 		manga := external.ConvertMALListToManga(mangaList)
	//
	// 		// MAL API doesn't provide total count in search, but has next/previous
	// 		hasNext := malResult.Paging.Next != ""
	// 		c.JSON(http.StatusOK, gin.H{
	// 			"data":   manga,
	// 			"total":  len(manga),
	// 			"limit":  limit,
	// 			"page":   page,
	// 			"source": "official_mal",
	// 			"pagination": gin.H{
	// 				"has_next_page": hasNext,
	// 				"current_page":  page,
	// 				"items": gin.H{
	// 					"count":    len(manga),
	// 					"per_page": limit,
	// 				},
	// 			},
	// 		})
	// 		return
	// 	}
	// }

	// Use Jikan API for complete pagination support
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limit > 25 {
		limit = 25 // Jikan has lower limit
	}

	// Get sort parameters from query
	orderBy := c.Query("order_by")
	sort := c.Query("sort")

	var jikanManga *external.JikanMangaResponse
	var err error

	if orderBy != "" || sort != "" {
		jikanManga, err = s.JikanClient.SearchMangaWithSort(query, page, limit, orderBy, sort)
	} else {
		jikanManga, err = s.JikanClient.SearchManga(query, page, limit)
	}

	if err != nil {
		log.Printf("Jikan search error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search MyAnimeList"})
		return
	}

	manga := external.ConvertJikanListToManga(jikanManga.Data)

	// Get user ID if authenticated (optional)
	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// Enrich manga with custom ratings instead of MAL ratings
	s.enrichMangaWithRatings(manga, userID)

	c.JSON(http.StatusOK, gin.H{
		"data":   manga,
		"total":  len(manga),
		"page":   page,
		"limit":  limit,
		"source": "jikan",
		"pagination": gin.H{
			"last_visible_page": jikanManga.Pagination.LastVisiblePage,
			"has_next_page":     jikanManga.Pagination.HasNextPage,
			"current_page":      jikanManga.Pagination.CurrentPage,
			"items": gin.H{
				"count":    jikanManga.Pagination.Items.Count,
				"total":    jikanManga.Pagination.Items.Total,
				"per_page": jikanManga.Pagination.Items.PerPage,
			},
		},
	})
}

// getTopMAL gets top manga from MyAnimeList (tries official API first, falls back to Jikan)
func (s *APIServer) getTopMAL(c *gin.Context) {
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// For pagination-heavy operations like browsing, prefer Jikan API as it provides
	// complete pagination info (total count, last_visible_page, etc.)
	// Official MAL API only provides next/previous links without totals

	// Commented out MAL API to prefer Jikan for better pagination
	// if s.MALClient.IsConfigured() {
	// 	// Get page parameter for offset calculation
	// 	page := 1
	// 	if pageStr := c.Query("page"); pageStr != "" {
	// 		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
	// 			page = p
	// 		}
	// 	}
	//
	// 	// Calculate offset from page number
	// 	offset := (page - 1) * limit
	//
	// 	malResult, err := s.MALClient.GetMangaRanking("all", limit, offset)
	// 	if err != nil {
	// 		log.Printf("Official MAL API ranking error: %v, falling back to Jikan", err)
	// 	} else {
	// 		// Convert official MAL data
	// 		mangaList := make([]external.MALMangaNode, 0, len(malResult.Data))
	// 		for _, item := range malResult.Data {
	// 			mangaList = append(mangaList, item.Node)
	// 		}
	// 		manga := external.ConvertMALListToManga(mangaList)
	//
	// 		// MAL API doesn't provide total count in ranking, but has next/previous
	// 		hasNext := malResult.Paging.Next != ""
	// 		c.JSON(http.StatusOK, gin.H{
	// 			"data":   manga,
	// 			"total":  len(manga),
	// 			"limit":  limit,
	// 			"page":   page,
	// 			"source": "official_mal",
	// 			"pagination": gin.H{
	// 				"has_next_page": hasNext,
	// 				"current_page":  page,
	// 				"items": gin.H{
	// 					"count":    len(manga),
	// 					"per_page": limit,
	// 				},
	// 			},
	// 		})
	// 		return
	// 	}
	// }

	// Use Jikan API for complete pagination support
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limit > 25 {
		limit = 25 // Jikan has lower limit
	}

	// Get sort parameters from query
	orderBy := c.Query("order_by")
	sort := c.Query("sort")

	var jikanManga *external.JikanMangaResponse
	var err error

	if orderBy != "" || sort != "" {
		jikanManga, err = s.JikanClient.GetMangaWithSort(page, limit, orderBy, sort)
	} else {
		jikanManga, err = s.JikanClient.GetTopManga(page, limit)
	}

	if err != nil {
		log.Printf("Jikan manga error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get manga from MyAnimeList"})
		return
	}

	manga := external.ConvertJikanListToManga(jikanManga.Data)

	// Get user ID if authenticated (optional)
	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// Enrich manga with custom ratings instead of MAL ratings
	s.enrichMangaWithRatings(manga, userID)

	c.JSON(http.StatusOK, gin.H{
		"data":   manga,
		"total":  len(manga),
		"page":   page,
		"limit":  limit,
		"source": "jikan",
		"pagination": gin.H{
			"last_visible_page": jikanManga.Pagination.LastVisiblePage,
			"has_next_page":     jikanManga.Pagination.HasNextPage,
			"current_page":      jikanManga.Pagination.CurrentPage,
			"items": gin.H{
				"count":    jikanManga.Pagination.Items.Count,
				"total":    jikanManga.Pagination.Items.Total,
				"per_page": jikanManga.Pagination.Items.PerPage,
			},
		},
	})
}

// getMALManga gets a specific manga from MyAnimeList (tries official API first, falls back to Jikan)
func (s *APIServer) getMALManga(c *gin.Context) {
	malIDStr := c.Param("mal_id")
	malID, err := strconv.Atoi(malIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MyAnimeList ID"})
		return
	}

	// Try official MAL API first
	if s.MALClient.IsConfigured() {
		malManga, err := s.MALClient.GetMangaByID(malID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				// Try Jikan as fallback
				log.Printf("Manga %d not found in official MAL API, trying Jikan", malID)
			} else {
				log.Printf("Official MAL API get manga error: %v, falling back to Jikan", err)
			}
		} else {
			manga := external.ConvertMALToManga(malManga)

			// Get user ID if authenticated (optional)
			userID := ""
			if uid, exists := c.Get("user_id"); exists {
				userID = uid.(string)
			}

			// Enrich manga with custom ratings instead of MAL ratings
			s.enrichMangaWithRatings([]*models.Manga{manga}, userID)

			c.JSON(http.StatusOK, manga)
			return
		}
	}

	// Fallback to Jikan API
	jikanManga, err := s.JikanClient.GetMangaByID(malID)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Manga not found on MyAnimeList"})
		} else {
			log.Printf("Jikan get manga error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get manga from MyAnimeList"})
		}
		return
	}

	manga := external.ConvertJikanToManga(jikanManga)

	// Get user ID if authenticated (optional)
	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// Enrich manga with custom ratings instead of MAL ratings
	s.enrichMangaWithRatings([]*models.Manga{manga}, userID)

	c.JSON(http.StatusOK, manga)
}

// Get MAL manga recommendations
func (s *APIServer) getMALRecommendations(c *gin.Context) {
	malIDStr := c.Param("mal_id")
	malID, err := strconv.Atoi(malIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MyAnimeList ID"})
		return
	}

	// Fetch recommendations from Jikan API
	recommendations, err := s.JikanClient.GetMangaRecommendations(malID)
	if err != nil {
		log.Printf("Jikan get recommendations error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recommendations from MyAnimeList"})
		return
	}

	// Convert to our manga format
	mangaList := make([]*models.Manga, 0, len(recommendations))
	for _, rec := range recommendations {
		manga := external.ConvertJikanToManga(&rec)
		mangaList = append(mangaList, manga)
	}

	// Get user ID if authenticated (optional)
	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	// Enrich manga with custom ratings instead of MAL ratings
	s.enrichMangaWithRatings(mangaList, userID)

	c.JSON(http.StatusOK, gin.H{
		"data":  mangaList,
		"count": len(mangaList),
	})
}

// Start starts the HTTP server
func (s *APIServer) Start() error {
	log.Printf("Starting API server on port %s", s.Port)
	return s.Router.Run(":" + s.Port)
}

// Bulk import manga endpoint (admin only)
func (s *APIServer) bulkImportManga(c *gin.Context) {
	var request struct {
		Manga      []models.Manga `json:"manga" binding:"required"`
		SkipExists bool           `json:"skip_exists"`
		Validate   bool           `json:"validate"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := struct {
		Success     int      `json:"success"`
		Failed      int      `json:"failed"`
		Skipped     int      `json:"skipped"`
		Total       int      `json:"total"`
		Errors      []string `json:"errors,omitempty"`
		ImportedIDs []string `json:"imported_ids"`
	}{
		Errors:      []string{},
		ImportedIDs: []string{},
		Total:       len(request.Manga),
	}

	for _, mangaData := range request.Manga {
		// Validate data if requested
		if request.Validate {
			if err := s.validateSingleMangaData(mangaData); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Invalid data for %s: %v", mangaData.Title, err))
				result.Failed++
				continue
			}
		}

		// Check if manga already exists
		if request.SkipExists {
			existing, _ := s.MangaService.GetManga(mangaData.ID)
			if existing != nil {
				result.Skipped++
				continue
			}
		}

		// Create manga
		created, err := s.MangaService.CreateManga(mangaData)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to create %s: %v", mangaData.Title, err))
			result.Failed++
			continue
		}

		result.Success++
		result.ImportedIDs = append(result.ImportedIDs, created.ID)
	}

	c.JSON(http.StatusOK, result)
}

// Validate manga data endpoint (admin only)
func (s *APIServer) validateMangaData(c *gin.Context) {
	var request struct {
		Manga []models.Manga `json:"manga" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := struct {
		Valid     int                      `json:"valid"`
		Invalid   int                      `json:"invalid"`
		Total     int                      `json:"total"`
		Errors    map[string][]string      `json:"errors,omitempty"`
		Validated []map[string]interface{} `json:"validated_data"`
	}{
		Errors:    make(map[string][]string),
		Validated: []map[string]interface{}{},
		Total:     len(request.Manga),
	}

	for _, mangaData := range request.Manga {
		mangaID := mangaData.ID
		if mangaID == "" {
			mangaID = mangaData.Title
		}

		if err := s.validateSingleMangaData(mangaData); err != nil {
			result.Invalid++
			result.Errors[mangaID] = append(result.Errors[mangaID], err.Error())
		} else {
			result.Valid++
			result.Validated = append(result.Validated, map[string]interface{}{
				"id":    mangaData.ID,
				"title": mangaData.Title,
				"valid": true,
			})
		}
	}

	c.JSON(http.StatusOK, result)
}

// Get import statistics endpoint (admin only)
func (s *APIServer) getImportStats(c *gin.Context) {
	stats, err := s.MangaService.GetMangaStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics"})
		return
	}

	// Add import-specific statistics
	importStats := map[string]interface{}{
		"manga_stats": stats,
		"last_import": time.Now().Format("2006-01-02 15:04:05"),
		"import_sources": map[string]interface{}{
			"manual":   stats["total_manga"], // All current manga are manual for now
			"external": 0,
		},
		"data_validation": map[string]interface{}{
			"validation_enabled": true,
			"required_fields":    []string{"id", "title", "genres"},
			"optional_fields":    []string{"description", "author", "artist", "status", "total_chapters", "cover_image_url"},
		},
	}

	c.JSON(http.StatusOK, importStats)
}

// Bulk delete manga endpoint (admin only)
func (s *APIServer) bulkDeleteManga(c *gin.Context) {
	var request struct {
		MangaIDs []string `json:"manga_ids" binding:"required"`
		Confirm  bool     `json:"confirm" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !request.Confirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Confirmation required for bulk delete operation"})
		return
	}

	result := struct {
		Success    int      `json:"success"`
		Failed     int      `json:"failed"`
		Total      int      `json:"total"`
		Errors     []string `json:"errors,omitempty"`
		DeletedIDs []string `json:"deleted_ids"`
	}{
		Errors:     []string{},
		DeletedIDs: []string{},
		Total:      len(request.MangaIDs),
	}

	for _, mangaID := range request.MangaIDs {
		if err := s.MangaService.DeleteManga(mangaID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete %s: %v", mangaID, err))
			result.Failed++
			continue
		}

		result.Success++
		result.DeletedIDs = append(result.DeletedIDs, mangaID)
	}

	c.JSON(http.StatusOK, result)
}

// Helper function to validate individual manga data
func (s *APIServer) validateSingleMangaData(manga models.Manga) error {
	// Check required fields
	if manga.ID == "" {
		return fmt.Errorf("manga ID is required")
	}

	if manga.Title == "" {
		return fmt.Errorf("manga title is required")
	}

	// Validate ID format (no spaces, special characters)
	if strings.Contains(manga.ID, " ") {
		return fmt.Errorf("manga ID cannot contain spaces")
	}

	// Validate status
	if manga.Status != "" {
		validStatuses := []string{"ongoing", "completed", "hiatus", "dropped", "cancelled"}
		isValid := false
		for _, status := range validStatuses {
			if manga.Status == status {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid status: %s (must be one of: %s)", manga.Status, strings.Join(validStatuses, ", "))
		}
	}

	// Validate chapters
	if manga.TotalChapters < 0 {
		return fmt.Errorf("total chapters cannot be negative")
	}

	// Validate genres (at least one genre required)
	if len(manga.Genres) == 0 {
		return fmt.Errorf("at least one genre is required")
	}

	// Validate title length
	if len(manga.Title) > 200 {
		return fmt.Errorf("title too long (max 200 characters)")
	}

	// Validate description length
	if len(manga.Description) > 2000 {
		return fmt.Errorf("description too long (max 2000 characters)")
	}

	// Validate author length
	if len(manga.Author) > 100 {
		return fmt.Errorf("author name too long (max 100 characters)")
	}

	// Validate URL format if provided
	if manga.CoverURL != "" && !strings.HasPrefix(manga.CoverURL, "http") {
		return fmt.Errorf("cover image URL must be a valid HTTP/HTTPS URL")
	}

	return nil
}

// connectToTCPServer connects to the TCP progress sync server as a client
func (s *APIServer) connectToTCPServer() {
	tcpAddr := os.Getenv("TCP_SERVER_ADDR")
	if tcpAddr == "" {
		tcpAddr = "localhost:9000" // Default TCP server address
	}

	maxRetries := 3
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		if i == 0 {
			log.Printf("Attempting to connect to TCP server at %s...", tcpAddr)
		}

		conn, err := net.Dial("tcp", tcpAddr)
		if err != nil {
			if i == maxRetries-1 {
				log.Printf("INFO: TCP server not available. Progress sync features disabled. (This is optional)")
			}
			time.Sleep(retryDelay)
			continue
		}

		s.tcpMu.Lock()
		s.tcpConn = conn
		s.tcpMu.Unlock()

		log.Printf("Successfully connected to TCP server at %s", tcpAddr)

		// Keep connection alive and handle reconnection
		go s.maintainTCPConnection(tcpAddr)
		return
	}

	// TCP server is optional - continue without it
	log.Printf("INFO: Running without TCP server connection. Real-time progress sync disabled.")
}

// maintainTCPConnection monitors the TCP connection and reconnects if needed
func (s *APIServer) maintainTCPConnection(tcpAddr string) {
	reader := bufio.NewReader(s.tcpConn)
	for {
		// Read from connection to detect disconnection
		_, err := reader.ReadByte()
		if err != nil {
			log.Printf("TCP connection lost. Running without TCP server.")
			s.tcpMu.Lock()
			if s.tcpConn != nil {
				s.tcpConn.Close()
				s.tcpConn = nil
			}
			s.tcpMu.Unlock()

			// Don't auto-reconnect to avoid spam
			// Server can function without TCP connection
			return
		}
	}
}

// connectToGRPCServer establishes connection to gRPC server
func (s *APIServer) connectToGRPCServer() {
	grpcAddr := os.Getenv("GRPC_SERVER_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:9001" // Default gRPC server address
	}

	maxRetries := 5
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		if i == 0 {
			log.Printf("Attempting to connect to gRPC server at %s...", grpcAddr)
		}

		client, err := grpcClient.NewClient(grpcAddr)
		if err != nil {
			if i < maxRetries-1 {
				log.Printf("Failed to connect to gRPC server (attempt %d/%d): %v", i+1, maxRetries, err)
				time.Sleep(retryDelay)
				continue
			} else {
				log.Printf("WARNING: gRPC server not available. Some features may be limited. Error: %v", err)
				return
			}
		}

		s.GRPCClient = client
		log.Printf("Successfully connected to gRPC server at %s", grpcAddr)
		return
	}

	// gRPC server is optional - continue without it
	log.Printf("INFO: Running without gRPC server connection.")
}

// broadcastProgressUpdate sends progress update to TCP server for broadcasting
func (s *APIServer) broadcastProgressUpdate(userID, mangaID string, chapter int) {
	s.tcpMu.Lock()
	conn := s.tcpConn
	s.tcpMu.Unlock()

	if conn == nil {
		// TCP connection not available - this is fine, server works without it
		return
	}

	// Create progress update message
	type ProgressUpdate struct {
		UserID    string `json:"user_id"`
		MangaID   string `json:"manga_id"`
		Chapter   int    `json:"chapter"`
		Timestamp int64  `json:"timestamp"`
	}

	update := ProgressUpdate{
		UserID:    userID,
		MangaID:   mangaID,
		Chapter:   chapter,
		Timestamp: time.Now().Unix(),
	}

	// Marshal to JSON
	data, err := json.Marshal(update)
	if err != nil {
		log.Printf("Failed to marshal progress update: %v", err)
		return
	}

	// Send to TCP server (with newline delimiter)
	data = append(data, '\n')

	s.tcpMu.Lock()
	defer s.tcpMu.Unlock()

	if s.tcpConn != nil {
		s.tcpConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_, err = s.tcpConn.Write(data)
		if err != nil {
			log.Printf("Failed to send progress update to TCP server: %v", err)
			s.tcpConn.Close()
			s.tcpConn = nil
		} else {
			log.Printf("Broadcasted progress update to TCP server: User=%s, Manga=%s, Chapter=%d", userID, mangaID, chapter)
		}
	}
}

// handleWebSocketChat handles WebSocket chat connections
func (s *APIServer) handleWebSocketChat(c *gin.Context) {
	// Get user info from auth middleware
	userID := c.GetString("user_id")
	username := c.GetString("username")

	if userID == "" || username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Get room ID from query parameter (manga ID)
	roomID := c.Query("room")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room ID required"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Get or create room
	room := s.RoomHub.GetOrCreateRoom(roomID)

	// Create client connection
	client := &internalWebsocket.ClientConnection{
		Conn:     conn,
		UserID:   userID,
		Username: username,
		Room:     roomID,
	}

	// Register client with room
	room.Register <- client

	// Start reading messages from this client
	go s.readWebSocketMessages(client, room)
}

// readWebSocketMessages reads messages from a WebSocket client
func (s *APIServer) readWebSocketMessages(client *internalWebsocket.ClientConnection, room *internalWebsocket.ChatRoom) {
	defer func() {
		room.Unregister <- client.Conn
	}()

	// Ping interval and timeouts
	const (
		pongWait   = 120 * time.Second   // Time allowed to read the next pong message
		pingPeriod = (pongWait * 9) / 10 // Send pings with this period (must be less than pongWait)
		writeWait  = 10 * time.Second    // Time allowed to write a message
	)

	// Configure read settings with extended deadline
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Start ping ticker to keep connection alive
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	// Start goroutine to send periodic pings
	go func() {
		for range pingTicker.C {
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	// Read messages loop
	for {
		var msg internalWebsocket.ChatMessage
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Reset read deadline on every message
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))

		// Set message metadata
		msg.UserID = client.UserID
		msg.Username = client.Username
		msg.Timestamp = time.Now().Unix()
		msg.Type = "message"
		msg.Room = client.Room

		// Validate message
		if msg.Message == "" || len(msg.Message) > 1000 {
			log.Printf("Invalid message from %s: empty or too long", client.Username)
			continue
		}

		// Broadcast message to all clients in room
		room.Broadcast <- msg
	}
}

// getWebSocketStats returns WebSocket connection statistics
func (s *APIServer) getWebSocketStats(c *gin.Context) {
	roomID := c.Query("room")

	if roomID == "" {
		// Return global stats
		totalRooms, totalClients := s.RoomHub.GetGlobalStats()

		stats := gin.H{
			"total_rooms":   totalRooms,
			"total_clients": totalClients,
			"status":        "running",
		}
		c.JSON(http.StatusOK, stats)
		return
	}

	// Return specific room stats
	room := s.RoomHub.GetRoom(roomID)
	if room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	stats := gin.H{
		"room_id":           roomID,
		"connected_clients": room.GetClientCount(),
		"connected_users":   room.GetConnectedUsers(),
		"status":            "running",
	}

	c.JSON(http.StatusOK, stats)
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

// getChapterList handles GET /api/v1/manga/:id/chapters
func (s *APIServer) getChapterList(c *gin.Context) {
	mangaID := c.Param("id")

	// Get query parameters
	languages := c.QueryArray("language")
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	// Get chapters from service
	chapters, err := s.ChapterService.GetChapterList(mangaID, languages, limit, offset)
	if err != nil {
		log.Printf("Error getting chapter list for manga %s: %v", mangaID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve chapter list",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, chapters)
}

// getChapterPages handles GET /api/v1/manga/chapters/:chapter_id/pages
func (s *APIServer) getChapterPages(c *gin.Context) {
	chapterID := c.Param("chapter_id")
	source := c.DefaultQuery("source", "mangadex")

	// Get chapter pages from service
	pages, err := s.ChapterService.GetChapterPages(chapterID, source)
	if err != nil {
		log.Printf("Error getting chapter pages for chapter %s: %v", chapterID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve chapter pages",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, pages)
}

// searchMangaDex handles GET /api/v1/manga/mangadex/search
func (s *APIServer) searchMangaDex(c *gin.Context) {
	title := c.Query("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "title query parameter is required",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	// Search MangaDex
	client := external.NewMangaDexClient()
	results, err := client.SearchManga(title, limit)
	if err != nil {
		log.Printf("Error searching MangaDex for '%s': %v", title, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search MangaDex",
			"details": err.Error(),
		})
		return
	}

	// Convert to simplified response
	response := make([]gin.H, 0, len(results.Data))
	for _, manga := range results.Data {
		response = append(response, gin.H{
			"id":          manga.ID,
			"title":       manga.GetTitle(),
			"description": manga.GetDescription(),
			"status":      manga.Attributes.Status,
			"year":        manga.Attributes.Year,
			"genres":      manga.GetGenres(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"results": response,
		"total":   results.Total,
		"limit":   results.Limit,
		"offset":  results.Offset,
	})
}

// rateManga handles POST /api/v1/manga/:manga_id/ratings
func (s *APIServer) rateManga(c *gin.Context) {
	mangaID := c.Param("manga_id")
	userID := c.GetString("user_id")

	var req struct {
		Rating int `json:"rating" binding:"required,min=1,max=5"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rating must be between 1 and 5"})
		return
	}

	if err := s.RatingService.RateManga(userID, mangaID, req.Rating); err != nil {
		log.Printf("Error rating manga: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save rating"})
		return
	}

	// Get updated stats
	stats, err := s.RatingService.GetMangaRatingStats(mangaID, userID)
	if err != nil {
		log.Printf("Error getting rating stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rating stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// deleteRating handles DELETE /api/v1/manga/:manga_id/ratings
func (s *APIServer) deleteRating(c *gin.Context) {
	mangaID := c.Param("manga_id")
	userID := c.GetString("user_id")

	if err := s.RatingService.DeleteRating(userID, mangaID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rating not found"})
		} else {
			log.Printf("Error deleting rating: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rating"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rating deleted successfully"})
}

// getMangaRatings handles GET /api/v1/manga/:id/ratings
func (s *APIServer) getMangaRatings(c *gin.Context) {
	mangaID := c.Param("id")

	// Get user ID if authenticated (optional)
	userID := ""
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	stats, err := s.RatingService.GetMangaRatingStats(mangaID, userID)
	if err != nil {
		log.Printf("Error getting rating stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rating stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// enrichMangaWithRatings adds custom user ratings to manga list, replacing MAL ratings
func (s *APIServer) enrichMangaWithRatings(mangaList []*models.Manga, userID string) {
	for _, manga := range mangaList {
		// Get custom rating stats for this manga
		stats, err := s.RatingService.GetMangaRatingStats(manga.ID, userID)
		if err != nil {
			// If there's an error, set rating to 0 (no rating)
			manga.Rating = 0
			manga.RatingCount = 0
			manga.UserRating = nil
			continue
		}

		// Replace MAL rating with our custom rating
		if stats.AverageRating > 0 {
			manga.Rating = stats.AverageRating
		} else {
			manga.Rating = 0
		}
		manga.RatingCount = stats.TotalRatings
		manga.UserRating = stats.UserRating
	}
}

// ============================================================================
// gRPC-backed HTTP Handlers (UC-014, UC-015, UC-016)
// ============================================================================

// getMangaViaGRPC retrieves manga via gRPC service (UC-014)
func (s *APIServer) getMangaViaGRPC(c *gin.Context) {
	mangaID := c.Param("id")

	// Check if gRPC client is available
	if s.GRPCClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "gRPC service unavailable",
		})
		return
	}

	// Call gRPC GetManga method
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.GRPCClient.GetManga(ctx, mangaID)
	if err != nil {
		log.Printf("gRPC GetManga error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve manga via gRPC",
		})
		return
	}

	if resp.Error != "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": resp.Error,
		})
		return
	}

	// Convert protobuf manga to JSON response
	manga := resp.Manga
	c.JSON(http.StatusOK, gin.H{
		"id":               manga.Id,
		"title":            manga.Title,
		"author":           manga.Author,
		"genres":           manga.Genres,
		"status":           manga.Status,
		"total_chapters":   manga.TotalChapters,
		"description":      manga.Description,
		"cover_url":        manga.CoverUrl,
		"publication_year": manga.PublicationYear,
		"rating":           manga.Rating,
		"created_at":       manga.CreatedAt,
		"source":           "grpc",
	})
}

// searchMangaViaGRPC searches manga via gRPC service (UC-015)
func (s *APIServer) searchMangaViaGRPC(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		query = c.Query("query")
	}

	// Check if gRPC client is available
	if s.GRPCClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "gRPC service unavailable",
		})
		return
	}

	// Parse limit parameter
	limit := int32(20)
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = int32(l)
		}
	}

	// Call gRPC SearchManga method
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := s.GRPCClient.SearchManga(ctx, query, limit)
	if err != nil {
		log.Printf("gRPC SearchManga error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search manga via gRPC",
		})
		return
	}

	if resp.Error != "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": resp.Error,
		})
		return
	}

	// Convert protobuf manga list to JSON response
	results := make([]gin.H, len(resp.Manga))
	for i, manga := range resp.Manga {
		results[i] = gin.H{
			"id":               manga.Id,
			"title":            manga.Title,
			"author":           manga.Author,
			"genres":           manga.Genres,
			"status":           manga.Status,
			"total_chapters":   manga.TotalChapters,
			"description":      manga.Description,
			"cover_url":        manga.CoverUrl,
			"publication_year": manga.PublicationYear,
			"rating":           manga.Rating,
			"created_at":       manga.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"manga":  results,
		"total":  resp.Total,
		"query":  query,
		"source": "grpc",
	})
}

// updateProgressViaGRPC updates reading progress via gRPC service (UC-016)
func (s *APIServer) updateProgressViaGRPC(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		MangaID        string `json:"manga_id" binding:"required"`
		CurrentChapter int    `json:"current_chapter" binding:"min=0"`
		Status         string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if gRPC client is available
	if s.GRPCClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "gRPC service unavailable",
		})
		return
	}

	// Call gRPC UpdateProgress method
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.GRPCClient.UpdateProgress(ctx, userID, req.MangaID, int32(req.CurrentChapter), req.Status)
	if err != nil {
		log.Printf("gRPC UpdateProgress error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update progress via gRPC",
		})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": resp.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": resp.Message,
		"source":  "grpc",
		"success": true,
	})
}
