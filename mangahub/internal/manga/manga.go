package manga

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"mangahub/pkg/database"
	"mangahub/pkg/models"
	"strings"
)

// Service handles manga-related operations
type Service struct {
	db *sql.DB
}

// NewService creates a new manga service
func NewService() *Service {
	return &Service{
		db: database.GetDB(),
	}
}

// GetManga retrieves a single manga by ID
func (s *Service) GetManga(id string) (*models.Manga, error) {
	var manga models.Manga

	err := s.db.QueryRow(`
		SELECT id, title, author, genres, status, total_chapters, description, cover_url, created_at 
		FROM manga WHERE id = ?`, id).Scan(
		&manga.ID, &manga.Title, &manga.Author, &manga.GenresJSON,
		&manga.Status, &manga.TotalChapters, &manga.Description,
		&manga.CoverURL, &manga.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("manga not found")
		}
		return nil, fmt.Errorf("failed to get manga: %w", err)
	}

	// Parse genres
	if err := manga.GetGenres(); err != nil {
		log.Printf("Error parsing genres for manga %s: %v", manga.ID, err)
		manga.Genres = []string{}
	}

	return &manga, nil
}

// SearchManga searches for manga based on various criteria
func (s *Service) SearchManga(req models.MangaSearchRequest) ([]models.Manga, error) {
	// Build query
	query := `SELECT id, title, author, genres, status, total_chapters, description, cover_url, created_at FROM manga WHERE 1=1`
	args := []interface{}{}

	// Add search conditions
	if req.Query != "" {
		query += ` AND (title LIKE ? OR author LIKE ? OR description LIKE ?)`
		searchTerm := "%" + strings.ToLower(req.Query) + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	if req.Author != "" {
		query += ` AND author LIKE ?`
		args = append(args, "%"+req.Author+"%")
	}

	if req.Status != "" {
		query += ` AND status = ?`
		args = append(args, req.Status)
	}

	// Add genre filter if specified
	if len(req.Genres) > 0 {
		for _, genre := range req.Genres {
			query += ` AND genres LIKE ?`
			args = append(args, "%\""+genre+"\"%")
		}
	}

	// Add ordering
	query += ` ORDER BY title`

	// Add pagination
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	query += ` LIMIT ? OFFSET ?`
	args = append(args, req.Limit, req.Offset)

	// Execute query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search manga: %w", err)
	}
	defer rows.Close()

	var mangaList []models.Manga
	for rows.Next() {
		var manga models.Manga
		err := rows.Scan(&manga.ID, &manga.Title, &manga.Author, &manga.GenresJSON,
			&manga.Status, &manga.TotalChapters, &manga.Description,
			&manga.CoverURL, &manga.CreatedAt)
		if err != nil {
			log.Printf("Error scanning manga row: %v", err)
			continue
		}

		// Parse genres
		if err := manga.GetGenres(); err != nil {
			log.Printf("Error parsing genres for manga %s: %v", manga.ID, err)
			manga.Genres = []string{}
		}

		mangaList = append(mangaList, manga)
	}

	return mangaList, nil
}

// GetAllManga retrieves all manga with pagination
func (s *Service) GetAllManga(limit, offset int) ([]models.Manga, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.db.Query(`
		SELECT id, title, author, genres, status, total_chapters, description, cover_url, created_at 
		FROM manga 
		ORDER BY title 
		LIMIT ? OFFSET ?`, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to get all manga: %w", err)
	}
	defer rows.Close()

	var mangaList []models.Manga
	for rows.Next() {
		var manga models.Manga
		err := rows.Scan(&manga.ID, &manga.Title, &manga.Author, &manga.GenresJSON,
			&manga.Status, &manga.TotalChapters, &manga.Description,
			&manga.CoverURL, &manga.CreatedAt)
		if err != nil {
			log.Printf("Error scanning manga row: %v", err)
			continue
		}

		// Parse genres
		if err := manga.GetGenres(); err != nil {
			log.Printf("Error parsing genres for manga %s: %v", manga.ID, err)
			manga.Genres = []string{}
		}

		mangaList = append(mangaList, manga)
	}

	return mangaList, nil
}

// GetPopularManga retrieves popular manga (for now, just returns all manga)
func (s *Service) GetPopularManga(limit int) ([]models.Manga, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	// For now, just return the first N manga ordered by title
	// In a real application, you'd have a popularity metric
	rows, err := s.db.Query(`
		SELECT id, title, author, genres, status, total_chapters, description, cover_url, created_at 
		FROM manga 
		ORDER BY title 
		LIMIT ?`, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to get popular manga: %w", err)
	}
	defer rows.Close()

	var mangaList []models.Manga
	for rows.Next() {
		var manga models.Manga
		err := rows.Scan(&manga.ID, &manga.Title, &manga.Author, &manga.GenresJSON,
			&manga.Status, &manga.TotalChapters, &manga.Description,
			&manga.CoverURL, &manga.CreatedAt)
		if err != nil {
			log.Printf("Error scanning manga row: %v", err)
			continue
		}

		// Parse genres
		if err := manga.GetGenres(); err != nil {
			log.Printf("Error parsing genres for manga %s: %v", manga.ID, err)
			manga.Genres = []string{}
		}

		mangaList = append(mangaList, manga)
	}

	return mangaList, nil
}

// GetMangaByGenre retrieves manga filtered by genre
func (s *Service) GetMangaByGenre(genre string, limit, offset int) ([]models.Manga, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.db.Query(`
		SELECT id, title, author, genres, status, total_chapters, description, cover_url, created_at 
		FROM manga 
		WHERE genres LIKE ?
		ORDER BY title 
		LIMIT ? OFFSET ?`, "%\""+genre+"\"%", limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to get manga by genre: %w", err)
	}
	defer rows.Close()

	var mangaList []models.Manga
	for rows.Next() {
		var manga models.Manga
		err := rows.Scan(&manga.ID, &manga.Title, &manga.Author, &manga.GenresJSON,
			&manga.Status, &manga.TotalChapters, &manga.Description,
			&manga.CoverURL, &manga.CreatedAt)
		if err != nil {
			log.Printf("Error scanning manga row: %v", err)
			continue
		}

		// Parse genres
		if err := manga.GetGenres(); err != nil {
			log.Printf("Error parsing genres for manga %s: %v", manga.ID, err)
			manga.Genres = []string{}
		}

		mangaList = append(mangaList, manga)
	}

	return mangaList, nil
}

// GetAllGenres retrieves all unique genres from the database
func (s *Service) GetAllGenres() ([]string, error) {
	rows, err := s.db.Query(`SELECT DISTINCT genres FROM manga WHERE genres IS NOT NULL AND genres != ''`)
	if err != nil {
		return nil, fmt.Errorf("failed to get genres: %w", err)
	}
	defer rows.Close()

	genreSet := make(map[string]bool)
	for rows.Next() {
		var genresJSON string
		if err := rows.Scan(&genresJSON); err != nil {
			log.Printf("Error scanning genres row: %v", err)
			continue
		}

		var genres []string
		if err := json.Unmarshal([]byte(genresJSON), &genres); err != nil {
			log.Printf("Error parsing genres JSON: %v", err)
			continue
		}

		for _, genre := range genres {
			genreSet[genre] = true
		}
	}

	// Convert map to slice
	var allGenres []string
	for genre := range genreSet {
		allGenres = append(allGenres, genre)
	}

	return allGenres, nil
}
