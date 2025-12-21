package api

import (
	"fmt"
	"log"
	"mangahub/pkg/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

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

// Update profile endpoint
func (s *APIServer) updateProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate at least one field is provided
	if req.Username == "" && req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one field (username or email) must be provided"})
		return
	}

	profile, err := s.UserService.UpdateProfile(userID, req.Username, req.Email)
	if err != nil {
		if strings.Contains(err.Error(), "already taken") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("Update profile error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"profile": profile,
	})
}

// Change password endpoint
func (s *APIServer) changePassword(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.UserService.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		if strings.Contains(err.Error(), "incorrect") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		} else if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			log.Printf("Change password error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
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

	// Get manga info for notifications
	manga, err := s.MangaService.GetManga(req.MangaID)
	if err == nil {
		// Broadcast WebSocket notification
		message := fmt.Sprintf("Added '%s' to library with status: %s", manga.Title, req.Status)
		go s.ChatHub.BroadcastNotification(req.MangaID, "library_add", message)

		// Trigger UDP notification broadcast (NEW - for SSE clients)
		go s.triggerUDPNotification(userID, "library_add", fmt.Sprintf("📚 %s added '%s' to library", c.GetString("username"), manga.Title))
	}

	c.JSON(http.StatusOK, gin.H{"message": "Manga added to library successfully"})
}

// Update progress endpoint
func (s *APIServer) updateProgress(c *gin.Context) {
	userID := c.GetString("user_id")
	userName := c.GetString("username")

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

	// Trigger TCP progress update broadcast via HTTP
	go s.triggerTCPBroadcast(userID, userName, req.MangaID, req.CurrentChapter)

	// Broadcast progress update to WebSocket clients in the manga's chat room
	go s.ChatHub.BroadcastProgressUpdate(userID, userName, req.MangaID, req.CurrentChapter)

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

	// Get manga info for notification
	manga, err := s.MangaService.GetManga(mangaID)
	if err == nil {
		// Broadcast notification to WebSocket
		message := fmt.Sprintf("Removed '%s' from library", manga.Title)
		go s.ChatHub.BroadcastNotification(mangaID, "library_remove", message)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Manga removed from library successfully"})
}
