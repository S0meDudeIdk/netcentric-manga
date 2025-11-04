package main

import (
	"fmt"
	"log"
	"mangahub/internal/auth"
	"mangahub/internal/manga"
	"mangahub/internal/user"
	"mangahub/pkg/database"
	"mangahub/pkg/middleware"
	"mangahub/pkg/models"
	"mangahub/pkg/utils"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

	// Add security headers
	router.Use(middleware.SecurityHeaders())

	// Add rate limiting (100 requests per minute)
	router.Use(middleware.CreateRateLimiter(100))

	// Add request size limit (10MB)
	router.Use(middleware.RequestSizeLimit(10 * 1024 * 1024))

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
			}

			// Manga routes
			manga := protected.Group("/manga")
			{
				manga.GET("/", s.searchManga)
				manga.GET("/:id", s.getManga)
				manga.GET("/genres", s.getGenres)
				manga.GET("/popular", s.getPopularManga)
				manga.GET("/stats", s.getMangaStats)

				// Admin routes for manga management
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
