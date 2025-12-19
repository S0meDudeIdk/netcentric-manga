package api

import (
	"fmt"
	"mangahub/pkg/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

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
