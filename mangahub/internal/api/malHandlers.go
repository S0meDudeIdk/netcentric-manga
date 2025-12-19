package api

import (
	"log"
	"mangahub/internal/external"
	"mangahub/pkg/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

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
