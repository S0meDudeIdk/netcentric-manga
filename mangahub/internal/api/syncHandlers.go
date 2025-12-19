package api

import (
	"fmt"
	"log"
	"mangahub/pkg/database"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// syncMangaFromMAL syncs manga from MAL to local database
// Only stores manga that have chapters available on MangaDex/MangaPlus
func (s *APIServer) syncMangaFromMAL(c *gin.Context) {
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
func (s *APIServer) syncMangaChapters(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"total_fetched": result.TotalFetched,
		"synced":        result.Synced,
		"skipped":       result.Skipped,
		"failed":        result.Failed,
		"message":       fmt.Sprintf("Synced chapters for %d manga", result.Synced),
	})
}

// autoSyncManga runs on server startup to populate database with manga
func (s *APIServer) autoSyncManga() {
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

	// Skip sync if database already has manga
	if count > 0 {
		log.Println("Database already contains manga. Skipping auto-sync.")
		log.Println("Use the /api/v1/manga/sync endpoint manually if you want to sync more manga.")
		log.Println("=================================================")
		return
	}

	log.Println("Database is empty. Starting auto-sync...")
	log.Println("This will fetch ALL manga with readable chapters from MangaDex")
	log.Println("Note: This will take a while. Use 0 for unlimited sync.")

	// Sync directly from MangaDex (0 = unlimited, will fetch all available manga)
	result, err := s.SyncService.SyncFromMangaDex(1000)
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
}
