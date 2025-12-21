package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"mangahub/internal/external"
	"mangahub/internal/manga"
	"mangahub/internal/udp"
	"mangahub/pkg/database"
	"mangahub/pkg/models"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// FetchMangaServer represents the manga fetching service
type FetchMangaServer struct {
	Router        *gin.Engine
	MangaService  *manga.Service
	SyncService   *manga.SyncService
	RatingService *manga.RatingService
	MALClient     *external.MALClient
	JikanClient   *external.JikanClient
	Port          string
	udpServerURL  string
	httpClient    *http.Client
}

// NewFetchMangaServer creates a new fetch manga server instance
func NewFetchMangaServer() *FetchMangaServer {
	// Set Gin mode from environment
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(corsMiddleware())

	jikanClient := external.NewJikanClient()

	server := &FetchMangaServer{
		Router:        router,
		MangaService:  manga.NewService(),
		SyncService:   manga.NewSyncService(jikanClient),
		RatingService: manga.NewRatingService(),
		MALClient:     external.NewMALClient(),
		JikanClient:   jikanClient,
		Port:          getPort(),
		httpClient:    &http.Client{Timeout: 5 * time.Second},
	}

	// Initialize UDP notification support
	server.initializeUDP()

	// Setup routes
	server.setupRoutes()

	// Auto-sync manga from MAL on startup (in background)
	go server.autoSyncManga()

	// Start periodic sync every 15 minutes
	go server.startPeriodicSync()

	return server
}

// getPort returns the port from environment or default
func getPort() string {
	port := os.Getenv("FETCH_MANGA_PORT")
	if port == "" {
		port = "8082" // Different port from main API server
	}
	return port
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// setupRoutes configures all API routes
func (s *FetchMangaServer) setupRoutes() {
	// Health check
	s.Router.GET("/health", s.healthCheck)

	// API version 1
	v1 := s.Router.Group("/api/v1")
	{
		manga := v1.Group("/manga")
		{
			// MAL/Jikan API endpoints
			manga.GET("/mal/search", s.searchMAL)
			manga.GET("/mal/top", s.getTopMAL)
			manga.GET("/mal/:mal_id", s.getMALManga)
			manga.GET("/mal/:mal_id/recommendations", s.getMALRecommendations)

			// Sync endpoints
			manga.POST("/sync", s.syncMangaFromMAL)
			manga.POST("/sync-chapters", s.syncMangaChapters)

			// Admin endpoints (in production, add authentication middleware)
			manga.POST("/bulk-import", s.bulkImportManga)
			manga.POST("/validate", s.validateMangaData)
			manga.GET("/import-stats", s.getImportStats)
		}
	}
}

// Health check endpoint
func (s *FetchMangaServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "fetch-manga-server",
	})
}

// searchMAL searches manga from MyAnimeList (uses Jikan API for better pagination)
func (s *FetchMangaServer) searchMAL(c *gin.Context) {
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

// getTopMAL gets top manga from MyAnimeList (uses Jikan API for better pagination)
func (s *FetchMangaServer) getTopMAL(c *gin.Context) {
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

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
func (s *FetchMangaServer) getMALManga(c *gin.Context) {
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

// getMALRecommendations gets manga recommendations from MyAnimeList
func (s *FetchMangaServer) getMALRecommendations(c *gin.Context) {
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

// syncMangaFromMAL syncs manga from MAL to local database
// Only stores manga that have chapters available on MangaDex/MangaPlus
func (s *FetchMangaServer) syncMangaFromMAL(c *gin.Context) {
	// Get query parameters
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}

	limit := 20 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	// Validate limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	log.Printf("Starting manga sync: query='%s', limit=%d", query, limit)

	// Run sync
	result, err := s.SyncService.SyncFromMAL(query, limit)
	if err != nil {
		log.Printf("Sync error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to sync manga from MAL",
		})
		return
	}

	// Trigger UDP notification if manga were synced
	if result.Synced > 0 {
		go s.triggerUDPNotification(
			"new_comics",
			fmt.Sprintf("🆕 %d new comics synced from MyAnimeList", result.Synced),
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"total_fetched": result.TotalFetched,
		"synced":        result.Synced,
		"skipped":       result.Skipped,
		"failed":        result.Failed,
		"details":       result.Details,
		"message":       fmt.Sprintf("Synced %d out of %d manga", result.Synced, result.TotalFetched),
	})
}

// syncMangaChapters forces a re-sync of chapters for manga that don't have chapters
func (s *FetchMangaServer) syncMangaChapters(c *gin.Context) {
	log.Println("Starting chapter sync for manga without chapters")

	// Run the MangaDex sync which now checks for missing chapters
	result, err := s.SyncService.SyncFromMangaDex(0) // 0 = unlimited
	if err != nil {
		log.Printf("Chapter sync error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to sync chapters",
		})
		return
	}

	// Trigger UDP notification if chapters were synced
	if result.Synced > 0 {
		go s.triggerUDPNotification(
			"new_chapters",
			fmt.Sprintf("📖 New chapters available! %d manga updated with fresh chapters", result.Synced),
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"total_fetched": result.TotalFetched,
		"synced":        result.Synced,
		"skipped":       result.Skipped,
		"failed":        result.Failed,
		"message":       fmt.Sprintf("Synced chapters for %d manga", result.Synced),
	})
}

// bulkImportManga imports multiple manga at once
func (s *FetchMangaServer) bulkImportManga(c *gin.Context) {
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

// validateMangaData validates manga data without importing
func (s *FetchMangaServer) validateMangaData(c *gin.Context) {
	var request struct {
		Manga []models.Manga `json:"manga" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var validationResults []struct {
		Index  int      `json:"index"`
		Title  string   `json:"title"`
		Valid  bool     `json:"valid"`
		Errors []string `json:"errors,omitempty"`
	}

	for i, mangaData := range request.Manga {
		result := struct {
			Index  int      `json:"index"`
			Title  string   `json:"title"`
			Valid  bool     `json:"valid"`
			Errors []string `json:"errors,omitempty"`
		}{
			Index:  i,
			Title:  mangaData.Title,
			Valid:  true,
			Errors: []string{},
		}

		if err := s.validateSingleMangaData(mangaData); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err.Error())
		}

		validationResults = append(validationResults, result)
	}

	c.JSON(http.StatusOK, gin.H{
		"results": validationResults,
		"total":   len(request.Manga),
	})
}

// getImportStats returns statistics about imported manga
func (s *FetchMangaServer) getImportStats(c *gin.Context) {
	stats, err := s.MangaService.GetMangaStats()
	if err != nil {
		log.Printf("Error getting manga stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get manga statistics"})
		return
	}

	// Add import-specific statistics
	db := database.GetDB()
	var importStats struct {
		TotalManga           int        `json:"total_manga"`
		MangaWithChapters    int        `json:"manga_with_chapters"`
		MangaWithoutChapters int        `json:"manga_without_chapters"`
		LastSyncTime         *time.Time `json:"last_sync_time,omitempty"`
	}

	db.QueryRow("SELECT COUNT(*) FROM manga").Scan(&importStats.TotalManga)
	db.QueryRow("SELECT COUNT(DISTINCT manga_id) FROM chapters").Scan(&importStats.MangaWithChapters)
	importStats.MangaWithoutChapters = importStats.TotalManga - importStats.MangaWithChapters

	c.JSON(http.StatusOK, gin.H{
		"stats":        stats,
		"import_stats": importStats,
	})
}

// validateSingleMangaData validates individual manga data
func (s *FetchMangaServer) validateSingleMangaData(manga models.Manga) error {
	// Required fields
	if manga.Title == "" {
		return fmt.Errorf("title is required")
	}
	if manga.ID == "" {
		return fmt.Errorf("ID is required")
	}

	// Validate status
	validStatuses := map[string]bool{
		"ongoing":   true,
		"completed": true,
		"hiatus":    true,
		"cancelled": true,
		"":          true, // Empty is allowed
	}
	if manga.Status != "" && !validStatuses[manga.Status] {
		return fmt.Errorf("invalid status: %s", manga.Status)
	}

	// Validate rating (0-10 range)
	if manga.Rating < 0 || manga.Rating > 10 {
		return fmt.Errorf("rating must be between 0 and 10")
	}

	// Validate publication year
	if manga.PublicationYear < 0 || manga.PublicationYear > time.Now().Year()+1 {
		return fmt.Errorf("invalid publication year: %d", manga.PublicationYear)
	}

	// Validate total chapters count
	if manga.TotalChapters < 0 {
		return fmt.Errorf("total chapters count cannot be negative")
	}

	return nil
}

// enrichMangaWithRatings adds custom user ratings to manga list, replacing MAL ratings
func (s *FetchMangaServer) enrichMangaWithRatings(mangaList []*models.Manga, userID string) {
	for _, manga := range mangaList {
		// Get custom rating statistics
		stats, err := s.RatingService.GetMangaRatingStats(manga.ID, userID)
		if err == nil && stats.TotalRatings > 0 {
			// Replace MAL rating with our custom rating
			manga.Rating = stats.AverageRating
			manga.RatingCount = stats.TotalRatings
		}

		// User's rating is already included in stats.UserRating
		if stats != nil && stats.UserRating != nil {
			manga.UserRating = stats.UserRating
		}
	}
}

// initializeUDP configures UDP notification support
func (s *FetchMangaServer) initializeUDP() {
	udpServerHost := os.Getenv("UDP_SERVER_HTTP_ADDR")
	if udpServerHost == "" {
		udpServerHost = "http://udp-server:9020" // Default to Docker service name
		log.Printf("⚠️  UDP_SERVER_HTTP_ADDR not set, using default: %s", udpServerHost)
	}
	s.udpServerURL = udpServerHost
	log.Printf("✅ UDP Server HTTP API configured at %s", s.udpServerURL)
	log.Printf("   HTTP Client timeout: %v", s.httpClient.Timeout)
}

// triggerUDPNotification sends a notification to the UDP server
func (s *FetchMangaServer) triggerUDPNotification(notifType, message string) {
	log.Printf("🔔 Attempting to send UDP notification: type=%s, message=%s", notifType, message)

	if s.udpServerURL == "" {
		log.Printf("❌ Cannot send UDP notification: udpServerURL is empty")
		return
	}

	if s.httpClient == nil {
		log.Printf("❌ Cannot send UDP notification: httpClient is nil")
		return
	}

	notification := udp.Notification{
		Type:      notifType,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(notification)
	if err != nil {
		log.Printf("❌ Failed to marshal UDP notification: %v", err)
		return
	}

	targetURL := s.udpServerURL + "/trigger"
	log.Printf("📡 Sending POST request to %s", targetURL)

	resp, err := s.httpClient.Post(
		targetURL,
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		log.Printf("❌ Failed to trigger UDP notification: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("📥 UDP server response: status=%d", resp.StatusCode)

	if resp.StatusCode == http.StatusOK {
		log.Printf("✅ UDP notification sent successfully: %s - %s", notifType, message)
	} else {
		log.Printf("⚠️  UDP notification sent but got non-OK status: %d", resp.StatusCode)
	}
}

// autoSyncManga runs on server startup to populate database with manga
func (s *FetchMangaServer) autoSyncManga() {
	log.Println("=================================================")
	log.Println("Starting automatic manga sync on server startup")
	log.Println("=================================================")

	// Wait a bit for server to fully start
	time.Sleep(2 * time.Second)

	// Check how many manga we already have
	var count int
	db := database.GetDB()
	err := db.QueryRow("SELECT COUNT(*) FROM manga").Scan(&count)
	if err != nil {
		log.Printf("ERROR: Failed to check manga count: %v", err)
		return
	}

	log.Printf("Current manga in database: %d", count)

	// Always run sync to update/add new manga
	log.Println("Starting sync to update database with latest manga...")
	log.Println("This will fetch manga with readable chapters from MangaDex")
	log.Println("Note: Existing manga will be skipped automatically.")

	// Sync directly from MangaDex (small limit to respect API rate limits)
	result, err := s.SyncService.SyncFromMangaDex(500)
	if err != nil {
		log.Printf("ERROR: Auto-sync failed: %v", err)
		return
	}

	log.Println("=================================================")
	log.Printf("Auto-sync completed!")
	log.Printf("  Total fetched: %d", result.TotalFetched)
	log.Printf("  Synced: %d", result.Synced)
	log.Printf("  Skipped: %d (includes already existing)", result.Skipped)
	log.Printf("  Failed: %d", result.Failed)
	log.Println("=================================================")

	// Send UDP notification about sync completion
	if result.Synced > 0 {
		log.Printf("🔔 Triggering UDP notification: %d new manga synced", result.Synced)
		s.triggerUDPNotification(
			"new_comics",
			fmt.Sprintf("🆕 %d new comics added to the library! Browse now to discover fresh content.", result.Synced),
		)
	} else {
		log.Printf("ℹ️  No new manga to sync (all %d already exist)", result.Skipped)
		s.triggerUDPNotification(
			"sync_complete",
			fmt.Sprintf("✅ Database check complete: All %d comics are up to date", result.TotalFetched),
		)
	}
}

// startPeriodicSync runs manga sync every 15 minutes
func (s *FetchMangaServer) startPeriodicSync() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	log.Println("⏰ Periodic sync enabled: Will sync every 15 minutes")

	for range ticker.C {
		log.Println("")
		log.Println("=================================================")
		log.Println("⏰ Starting periodic manga sync (15-minute interval)")
		log.Println("=================================================")

		// Check current manga count
		var count int
		db := database.GetDB()
		err := db.QueryRow("SELECT COUNT(*) FROM manga").Scan(&count)
		if err != nil {
			log.Printf("ERROR: Failed to check manga count: %v", err)
			continue
		}

		log.Printf("Current manga in database: %d", count)

		// Sync from MangaDex (small limit to respect API rate limits)
		// MangaDex allows ~5 requests/sec, so 20 manga = ~21 requests (1 batch + 20 chapter fetches)
		result, err := s.SyncService.SyncFromMangaDex(20)
		if err != nil {
			log.Printf("ERROR: Periodic sync failed: %v", err)
			continue
		}

		log.Println("=================================================")
		log.Printf("⏰ Periodic sync completed!")
		log.Printf("  Total fetched: %d", result.TotalFetched)
		log.Printf("  Synced: %d", result.Synced)
		log.Printf("  Skipped: %d (already existing)", result.Skipped)
		log.Printf("  Failed: %d", result.Failed)
		log.Println("=================================================")

		// Send UDP notification about sync results
		if result.Synced > 0 {
			log.Printf("🔔 Triggering UDP notification: %d new manga synced", result.Synced)
			s.triggerUDPNotification(
				"new_comics",
				fmt.Sprintf("🆕 %d new comics added to the library! Browse now to discover fresh content.", result.Synced),
			)
		} else {
			log.Printf("ℹ️  No new manga to sync (all %d already exist)", result.Skipped)
			// Don't send notification if nothing new - reduces noise
		}

		log.Printf("⏰ Next sync will run in 15 minutes")
	}
}

// Start starts the HTTP server
func (s *FetchMangaServer) Start() error {
	log.Printf("Fetch Manga Server starting on port %s", s.Port)
	return s.Router.Run(":" + s.Port)
}

func main() {
	// Load .env file - try multiple locations
	envLocations := []string{
		".env",       // Current directory
		"../.env",    // Parent directory
		"../../.env", // Grandparent directory (for cmd/fetch-manga-server)
	}

	envLoaded := false
	for _, envPath := range envLocations {
		if err := godotenv.Load(envPath); err == nil {
			log.Printf("✅ Loaded environment from %s", envPath)
			envLoaded = true
			break
		}
	}

	if !envLoaded {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	// Initialize database
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create and start server
	server := NewFetchMangaServer()

	log.Printf("=================================================")
	log.Printf("Fetch Manga Server running on port %s", server.Port)
	log.Printf("=================================================")

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
