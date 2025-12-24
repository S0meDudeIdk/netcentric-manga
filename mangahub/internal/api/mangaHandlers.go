package api

import (
	"fmt"
	"log"
	"mangahub/internal/external"
	"mangahub/internal/udp"
	"mangahub/pkg/models"
	"mangahub/pkg/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Search manga endpoint
func (s *APIServer) searchManga(c *gin.Context) {
	// Parse query parameters
	req := models.MangaSearchRequest{
		Query:  c.Query("query"),
		Author: c.Query("author"),
		Status: c.Query("status"),
		Sort:   c.Query("sort"),
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
			req.Genres[i] = utils.SanitizeString(req.Genres[i])
		}
	}

	// Search manga
	mangaList, err := s.MangaService.SearchManga(req)
	if err != nil {
		log.Printf("Search manga error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Get total count for pagination
	totalCount, err := s.MangaService.GetMangaCount(req)
	if err != nil {
		log.Printf("Get manga count error: %v", err)
		totalCount = len(mangaList) // Fallback to current count
	}

	c.JSON(http.StatusOK, gin.H{
		"manga":  mangaList,
		"count":  totalCount,
		"limit":  req.Limit,
		"offset": req.Offset,
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

	// Send UDP notification for new manga via HTTP trigger
	go s.triggerUDPNotification(udp.Notification{
		Type:      "manga_update",
		MangaID:   createdManga.ID,
		Message:   fmt.Sprintf("New manga added: %s by %s", createdManga.Title, createdManga.Author),
		Timestamp: time.Now().Unix(),
	})

	// Also broadcast to WebSocket clients
	s.ChatHub.BroadcastNotification(
		createdManga.ID,
		"manga_update",
		fmt.Sprintf("📚 New manga added: %s by %s", createdManga.Title, createdManga.Author),
	)

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

	// Send UDP notification for manga update via HTTP trigger
	var notification udp.Notification
	var wsMessage string
	if req.TotalChapters > 0 {
		notification = udp.Notification{
			Type:      "chapter_release",
			MangaID:   updatedManga.ID,
			Message:   fmt.Sprintf("New chapter %d released for %s", req.TotalChapters, updatedManga.Title),
			Timestamp: time.Now().Unix(),
		}
		wsMessage = fmt.Sprintf("📖 New chapter %d released for %s", req.TotalChapters, updatedManga.Title)
	} else {
		notification = udp.Notification{
			Type:      "manga_update",
			MangaID:   updatedManga.ID,
			Message:   fmt.Sprintf("Manga %s updated: %s", updatedManga.Title, updatedManga.Status),
			Timestamp: time.Now().Unix(),
		}
		wsMessage = fmt.Sprintf("🔄 Manga %s updated: %s", updatedManga.Title, updatedManga.Status)
	}

	go s.triggerUDPNotification(notification)

	// Also broadcast to WebSocket clients
	s.ChatHub.BroadcastNotification(updatedManga.ID, notification.Type, wsMessage)

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
