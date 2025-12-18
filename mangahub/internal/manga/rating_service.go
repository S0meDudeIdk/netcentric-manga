package manga

import (
	"database/sql"
	"fmt"
	"mangahub/pkg/database"
	"mangahub/pkg/models"
)

// RatingService handles manga rating operations
type RatingService struct{}

// NewRatingService creates a new rating service
func NewRatingService() *RatingService {
	return &RatingService{}
}

// RateManga adds or updates a user's rating for a manga
func (s *RatingService) RateManga(userID, mangaID string, rating int) error {
	if rating < 0 || rating > 10 {
		return fmt.Errorf("rating must be between 0 and 10")
	}

	db := database.GetDB()

	// Check if rating already exists
	var existingID int
	err := db.QueryRow(`
		SELECT id FROM manga_ratings 
		WHERE user_id = ? AND manga_id = ?
	`, userID, mangaID).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Insert new rating
		_, err = db.Exec(`
			INSERT INTO manga_ratings (user_id, manga_id, rating, created_at, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, userID, mangaID, rating)
		if err != nil {
			return fmt.Errorf("failed to insert rating: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check existing rating: %w", err)
	} else {
		// Update existing rating
		_, err = db.Exec(`
			UPDATE manga_ratings 
			SET rating = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, rating, existingID)
		if err != nil {
			return fmt.Errorf("failed to update rating: %w", err)
		}
	}

	return nil
}

// GetUserRating gets a specific user's rating for a manga
func (s *RatingService) GetUserRating(userID, mangaID string) (*int, error) {
	db := database.GetDB()

	var rating int
	err := db.QueryRow(`
		SELECT rating FROM manga_ratings 
		WHERE user_id = ? AND manga_id = ?
	`, userID, mangaID).Scan(&rating)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user rating: %w", err)
	}

	return &rating, nil
}

// GetMangaRatingStats gets the rating statistics for a manga
func (s *RatingService) GetMangaRatingStats(mangaID string, userID string) (*models.MangaRatingStats, error) {
	db := database.GetDB()

	stats := &models.MangaRatingStats{
		MangaID:            mangaID,
		RatingDistribution: make(map[int]int),
	}

	// Get average rating and count
	err := db.QueryRow(`
		SELECT 
			COALESCE(AVG(rating), 0) as avg_rating,
			COUNT(*) as total_ratings
		FROM manga_ratings
		WHERE manga_id = ?
	`, mangaID).Scan(&stats.AverageRating, &stats.TotalRatings)

	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get rating stats: %w", err)
	}

	// Get rating distribution (count for each rating 1-5)
	rows, err := db.Query(`
		SELECT rating, COUNT(*) as count
		FROM manga_ratings
		WHERE manga_id = ?
		GROUP BY rating
		ORDER BY rating DESC
	`, mangaID)

	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get rating distribution: %w", err)
	}

	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var rating, count int
			if err := rows.Scan(&rating, &count); err != nil {
				return nil, fmt.Errorf("failed to scan rating distribution: %w", err)
			}
			stats.RatingDistribution[rating] = count
		}
	}

	// Initialize all ratings 1-5 with 0 if not present
	for i := 1; i <= 5; i++ {
		if _, exists := stats.RatingDistribution[i]; !exists {
			stats.RatingDistribution[i] = 0
		}
	}

	// Get user's rating if authenticated
	if userID != "" {
		userRating, err := s.GetUserRating(userID, mangaID)
		if err != nil {
			return nil, err
		}
		stats.UserRating = userRating
	}

	return stats, nil
}

// DeleteRating deletes a user's rating for a manga
func (s *RatingService) DeleteRating(userID, mangaID string) error {
	db := database.GetDB()

	result, err := db.Exec(`
		DELETE FROM manga_ratings 
		WHERE user_id = ? AND manga_id = ?
	`, userID, mangaID)

	if err != nil {
		return fmt.Errorf("failed to delete rating: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("rating not found")
	}

	return nil
}

// GetAllRatingsForManga gets all ratings for a specific manga
func (s *RatingService) GetAllRatingsForManga(mangaID string, limit, offset int) ([]models.MangaRating, error) {
	db := database.GetDB()

	if limit <= 0 {
		limit = 20
	}

	rows, err := db.Query(`
		SELECT id, user_id, manga_id, rating, created_at, updated_at
		FROM manga_ratings
		WHERE manga_id = ?
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`, mangaID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to get ratings: %w", err)
	}
	defer rows.Close()

	var ratings []models.MangaRating
	for rows.Next() {
		var rating models.MangaRating
		err := rows.Scan(
			&rating.ID,
			&rating.UserID,
			&rating.MangaID,
			&rating.Rating,
			&rating.CreatedAt,
			&rating.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rating: %w", err)
		}
		ratings = append(ratings, rating)
	}

	return ratings, nil
}
