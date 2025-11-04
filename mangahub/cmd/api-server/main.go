package main

import (
	"log"
	"mangahub/internal/auth"
	"mangahub/internal/manga"
	"mangahub/internal/user"
	"mangahub/pkg/database"
	"mangahub/pkg/models"
	"mangahub/pkg/utils"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIServer represents the HTTP API server
type APIServer struct {
	Router       *gin.Engine
	UserService  *user.Service
	MangaService *manga.Service
	Port         string
}

// NewAPIServer creates a new API server instance
func NewAPIServer() *APIServer {
	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add CORS middleware
	router.Use(corsMiddleware())

	// Add request logging middleware
	router.Use(gin.Logger())

	// Add recovery middleware
	router.Use(gin.Recovery())

	server := &APIServer{
		Router:       router,
		UserService:  user.NewService(),
		MangaService: manga.NewService(),
		Port:         getPort(),
	}

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
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// authMiddleware validates JWT tokens
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := auth.ValidateToken(tokenParts[1])
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

		// Protected routes
		protected := v1.Group("/")
		protected.Use(authMiddleware())
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/profile", s.getProfile)
				users.GET("/library", s.getLibrary)
				users.POST("/library", s.addToLibrary)
				users.PUT("/progress", s.updateProgress)
			}

			// Manga routes
			manga := protected.Group("/manga")
			{
				manga.GET("/", s.searchManga)
				manga.GET("/:id", s.getManga)
				manga.GET("/genres", s.getGenres)
				manga.GET("/popular", s.getPopularManga)
			}
		}
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

	c.JSON(http.StatusOK, gin.H{"message": "Progress updated successfully"})
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

// Start starts the HTTP server
func (s *APIServer) Start() error {
	log.Printf("Starting API server on port %s", s.Port)
	return s.Router.Run(":" + s.Port)
}

func main() {
	// Initialize database
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Load manga data if not already loaded
	log.Println("Loading manga data...")
	if err := utils.LoadMangaData("data/manga.json"); err != nil {
		log.Printf("Warning: Failed to load manga data: %v", err)
	}

	// Create and start server
	server := NewAPIServer()

	log.Println("MangaHub API Server starting...")
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
