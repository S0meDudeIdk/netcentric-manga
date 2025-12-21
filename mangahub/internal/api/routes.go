package api

import (
	"mangahub/internal/auth"
	internalWebsocket "mangahub/internal/websocket"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

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
			// Logout requires authentication
			auth.POST("/logout", authMiddleware(), s.logout)
		}

		// Public manga browsing routes (no auth required)
		publicManga := v1.Group("/manga")
		{
			publicManga.GET("/", s.searchManga)
			publicManga.GET("/genres", s.getGenres)
			publicManga.GET("/popular", s.getPopularManga)
			publicManga.GET("/stats", s.getMangaStats)

			// Sync endpoint - fetch from MAL and store manga with chapters
			publicManga.POST("/sync", s.syncMangaFromMAL)
			// Force sync chapters for manga without chapters
			publicManga.POST("/sync-chapters", s.syncMangaChapters)

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
				users.PUT("/profile", s.updateProfile)
				users.PUT("/password", s.changePassword)
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
			protected.GET("/ws/chat", internalWebsocket.HandleWebSocketChat(s.ChatHub, s.upgrader))

			// SSE endpoints for real-time updates (protected)
			sseRoutes := protected.Group("/sse")
			{
				sseRoutes.GET("/progress", s.streamProgressUpdates)    // TCP progress sync via SSE
				sseRoutes.GET("/notifications", s.streamNotifications) // UDP notifications via SSE
			}

			// gRPC-backed endpoints (protected)
			grpcRoutes := protected.Group("/grpc")
			{
				grpcRoutes.GET("/manga/:id", s.getMangaViaGRPC)
				grpcRoutes.GET("/manga/search", s.searchMangaViaGRPC)
				grpcRoutes.PUT("/progress/update", s.updateProgressViaGRPC)

				// Library management via gRPC
				grpcRoutes.GET("/library", s.getLibraryViaGRPC)
				grpcRoutes.POST("/library", s.addToLibraryViaGRPC)
				grpcRoutes.DELETE("/library/:manga_id", s.removeFromLibraryViaGRPC)
				grpcRoutes.GET("/library/stats", s.getLibraryStatsViaGRPC)

				// Rating system via gRPC
				grpcRoutes.POST("/rating", s.rateMangaViaGRPC)
				grpcRoutes.GET("/rating/:manga_id", s.getMangaRatingsViaGRPC)
				grpcRoutes.DELETE("/rating/:manga_id", s.deleteRatingViaGRPC)
			}
		}
		// WebSocket stats endpoint (public for monitoring)
		v1.GET("/ws/stats", internalWebsocket.GetWebSocketStats(s.ChatHub))
	}
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
