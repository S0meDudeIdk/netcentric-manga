package api

import (
	"fmt"
	"mangahub/pkg/models"
	"strings"
)

// enrichMangaWithRatings adds custom user ratings to manga list, replacing MAL ratings
func (s *APIServer) enrichMangaWithRatings(mangaList []*models.Manga, userID string) {
	for _, manga := range mangaList {
		// Get custom rating stats for this manga
		stats, err := s.RatingService.GetMangaRatingStats(manga.ID, userID)
		if err != nil {
			// If there's an error, set rating to 0 (no rating)
			manga.Rating = 0
			manga.RatingCount = 0
			manga.UserRating = nil
			continue
		}

		// Replace MAL rating with our custom rating
		if stats.AverageRating > 0 {
			manga.Rating = stats.AverageRating
		} else {
			manga.Rating = 0
		}
		manga.RatingCount = stats.TotalRatings
		manga.UserRating = stats.UserRating
	}
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
