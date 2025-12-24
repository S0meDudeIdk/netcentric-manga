package api

import (
	"mangahub/pkg/models"
	"mangahub/pkg/utils"
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
	if err := utils.ValidateNonEmpty(manga.ID, "manga ID"); err != nil {
		return err
	}

	if err := utils.ValidateNonEmpty(manga.Title, "manga title"); err != nil {
		return err
	}

	// Validate ID format (no spaces, special characters)
	if err := utils.ValidateMangaID(manga.ID); err != nil {
		return err
	}

	// Validate status
	if err := utils.ValidateStatus(manga.Status); err != nil {
		return err
	}

	// Validate chapters
	if err := utils.ValidateNonNegative(manga.TotalChapters, "total chapters"); err != nil {
		return err
	}

	// Validate genres (at least one genre required)
	if err := utils.ValidateSliceNotEmpty(manga.Genres, "genre"); err != nil {
		return err
	}

	// Validate title length
	if err := utils.ValidateStringLength(manga.Title, "title", 200); err != nil {
		return err
	}

	// Validate description length
	if err := utils.ValidateStringLength(manga.Description, "description", 2000); err != nil {
		return err
	}

	// Validate author length
	if err := utils.ValidateStringLength(manga.Author, "author name", 100); err != nil {
		return err
	}

	// Validate URL format if provided
	if err := utils.ValidateURL(manga.CoverURL); err != nil {
		return err
	}

	return nil
}
