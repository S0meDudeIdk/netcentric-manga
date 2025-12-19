package api

import (
	"context"
	"log"
	pb "mangahub/proto"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

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

	// Parse offset parameter
	offset := int32(0)
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = int32(o)
		}
	}

	// Get sort parameter
	sort := c.Query("sort")
	if sort == "" {
		sort = "title" // Default sort
	}

	// Call gRPC SearchManga method with all parameters
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := s.GRPCClient.SearchManga(ctx, query, limit, offset, sort)
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

// getLibraryViaGRPC retrieves user's library via gRPC service
func (s *APIServer) getLibraryViaGRPC(c *gin.Context) {
	userID := c.GetString("user_id")

	if s.GRPCClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "gRPC service unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := s.GRPCClient.GetLibrary(ctx, userID)
	if err != nil {
		log.Printf("gRPC GetLibrary error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get library via gRPC"})
		return
	}

	if resp.Error != "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": resp.Error})
		return
	}

	// Convert protobuf to JSON
	convertProgress := func(pbProgress []*pb.UserProgress) []gin.H {
		result := make([]gin.H, len(pbProgress))
		for i, p := range pbProgress {
			result[i] = gin.H{
				"manga_id":        p.MangaId,
				"current_chapter": p.CurrentChapter,
				"status":          p.Status,
				"last_updated":    p.LastUpdated,
				"title":           p.Title,
				"author":          p.Author,
				"cover_url":       p.CoverUrl,
			}
		}
		return result
	}

	c.JSON(http.StatusOK, gin.H{
		"reading":      convertProgress(resp.Reading),
		"completed":    convertProgress(resp.Completed),
		"plan_to_read": convertProgress(resp.PlanToRead),
		"dropped":      convertProgress(resp.Dropped),
		"on_hold":      convertProgress(resp.OnHold),
		"re_reading":   convertProgress(resp.ReReading),
		"source":       "grpc",
	})
}

// addToLibraryViaGRPC adds manga to user's library via gRPC service
func (s *APIServer) addToLibraryViaGRPC(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		MangaID string `json:"manga_id" binding:"required"`
		Status  string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if s.GRPCClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "gRPC service unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.GRPCClient.AddToLibrary(ctx, userID, req.MangaID, req.Status)
	if err != nil {
		log.Printf("gRPC AddToLibrary error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to library via gRPC"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": resp.Message,
		"source":  "grpc",
		"success": true,
	})
}

// removeFromLibraryViaGRPC removes manga from user's library via gRPC service
func (s *APIServer) removeFromLibraryViaGRPC(c *gin.Context) {
	userID := c.GetString("user_id")
	mangaID := c.Param("manga_id")

	if s.GRPCClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "gRPC service unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.GRPCClient.RemoveFromLibrary(ctx, userID, mangaID)
	if err != nil {
		log.Printf("gRPC RemoveFromLibrary error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from library via gRPC"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": resp.Message,
		"source":  "grpc",
		"success": true,
	})
}

// getLibraryStatsViaGRPC retrieves user's library statistics via gRPC service
func (s *APIServer) getLibraryStatsViaGRPC(c *gin.Context) {
	userID := c.GetString("user_id")

	if s.GRPCClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "gRPC service unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.GRPCClient.GetLibraryStats(ctx, userID)
	if err != nil {
		log.Printf("gRPC GetLibraryStats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get library stats via gRPC"})
		return
	}

	if resp.Error != "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": resp.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_manga":         resp.TotalManga,
		"reading":             resp.Reading,
		"completed":           resp.Completed,
		"plan_to_read":        resp.PlanToRead,
		"dropped":             resp.Dropped,
		"on_hold":             resp.OnHold,
		"re_reading":          resp.ReReading,
		"total_chapters_read": resp.TotalChaptersRead,
		"source":              "grpc",
	})
}

// rateMangaViaGRPC submits a manga rating via gRPC service
func (s *APIServer) rateMangaViaGRPC(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		MangaID string `json:"manga_id" binding:"required"`
		Rating  int    `json:"rating" binding:"required,min=1,max=10"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if s.GRPCClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "gRPC service unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.GRPCClient.RateManga(ctx, userID, req.MangaID, int32(req.Rating))
	if err != nil {
		log.Printf("gRPC RateManga error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rate manga via gRPC"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        resp.Message,
		"average_rating": resp.AverageRating,
		"total_ratings":  resp.TotalRatings,
		"user_rating":    req.Rating,
		"source":         "grpc",
		"success":        true,
	})
}

// getMangaRatingsViaGRPC retrieves manga ratings via gRPC service
func (s *APIServer) getMangaRatingsViaGRPC(c *gin.Context) {
	mangaID := c.Param("manga_id")
	userID := c.GetString("user_id")

	if s.GRPCClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "gRPC service unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.GRPCClient.GetMangaRatings(ctx, mangaID, userID)
	if err != nil {
		log.Printf("gRPC GetMangaRatings error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ratings via gRPC"})
		return
	}

	if resp.Error != "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": resp.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"average_rating":      resp.AverageRating,
		"total_ratings":       resp.TotalRatings,
		"user_rating":         resp.UserRating,
		"rating_distribution": resp.RatingDistribution,
		"source":              "grpc",
	})
}

// deleteRatingViaGRPC deletes user's manga rating via gRPC service
func (s *APIServer) deleteRatingViaGRPC(c *gin.Context) {
	userID := c.GetString("user_id")
	mangaID := c.Param("manga_id")

	if s.GRPCClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "gRPC service unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.GRPCClient.DeleteRating(ctx, userID, mangaID)
	if err != nil {
		log.Printf("gRPC DeleteRating error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rating via gRPC"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": resp.Message,
		"source":  "grpc",
		"success": true,
	})
}
