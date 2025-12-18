package manga

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"mangahub/pkg/database"
	"mangahub/pkg/models"
	"strings"
	"time"
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
		SELECT id, title, author, genres, status, total_chapters, description, cover_url, publication_year, created_at 
		FROM manga WHERE id = ?`, id).Scan(
		&manga.ID, &manga.Title, &manga.Author, &manga.GenresJSON,
		&manga.Status, &manga.TotalChapters, &manga.Description,
		&manga.CoverURL, &manga.PublicationYear, &manga.CreatedAt)

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
	query := `SELECT id, title, author, genres, status, total_chapters, description, cover_url, publication_year, created_at FROM manga WHERE 1=1`
	args := []interface{}{}

	// Add search conditions
	// * Keep it in case of future need for case-insensitive search
	// if req.Query != "" {
	// 	query += ` AND (title LIKE ? OR author LIKE ? OR description LIKE ?)`
	// 	searchTerm := "%" + strings.ToLower(req.Query) + "%"
	// 	args = append(args, searchTerm, searchTerm, searchTerm)
	// }
	if req.Query != "" {
		query += ` AND (title LIKE ? COLLATE NOCASE OR author LIKE ? COLLATE NOCASE OR description LIKE ? COLLATE NOCASE)`
		searchTerm := "%" + req.Query + "%"
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

	// Add ordering based on Sort parameter
	orderBy := "title" // Default sort
	switch req.Sort {
	case "title":
		orderBy = "title ASC"
	case "relevant":
		// Relevance sorting: prioritize title matches over author/description
		if req.Query != "" {
			orderBy = "CASE WHEN title LIKE ? COLLATE NOCASE THEN 1 WHEN author LIKE ? COLLATE NOCASE THEN 2 ELSE 3 END, title ASC"
			// Note: The LIKE parameters are already added to args earlier in the query
		} else {
			orderBy = "title ASC"
		}
	case "newest":
		orderBy = "publication_year DESC, created_at DESC"
	case "popular":
		// Could be based on ratings, followers, etc. For now use total_chapters as proxy
		orderBy = "total_chapters DESC"
	default:
		orderBy = "title ASC"
	}
	query += ` ORDER BY ` + orderBy

	// Debug logging for sort issues
	log.Printf("DEBUG SearchManga - Sort: %s, OrderBy: %s", req.Sort, orderBy)
	log.Printf("DEBUG SearchManga - Final Query: %s", query)

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
			&manga.CoverURL, &manga.PublicationYear, &manga.CreatedAt)
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

	// Debug: Log first 5 results to verify sort order
	if len(mangaList) > 0 {
		log.Printf("DEBUG SearchManga - First 5 results:")
		for i := 0; i < len(mangaList) && i < 5; i++ {
			log.Printf("  %d. %s (Year: %d)", i+1, mangaList[i].Title, mangaList[i].PublicationYear)
		}
	}

	return mangaList, nil
}

// GetMangaCount returns the total count of manga matching the search criteria
func (s *Service) GetMangaCount(req models.MangaSearchRequest) (int, error) {
	// Build count query (same conditions as SearchManga but without pagination)
	query := `SELECT COUNT(*) FROM manga WHERE 1=1`
	args := []interface{}{}

	// Add same search conditions as SearchManga
	if req.Query != "" {
		query += ` AND (title LIKE ? COLLATE NOCASE OR author LIKE ? COLLATE NOCASE OR description LIKE ? COLLATE NOCASE)`
		searchTerm := "%" + req.Query + "%"
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

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count manga: %w", err)
	}

	return count, nil
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
		SELECT id, title, author, genres, status, total_chapters, description, cover_url, publication_year, created_at 
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
			&manga.CoverURL, &manga.PublicationYear, &manga.CreatedAt)
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
		SELECT id, title, author, genres, status, total_chapters, description, cover_url, publication_year, created_at 
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
			&manga.CoverURL, &manga.PublicationYear, &manga.CreatedAt)
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
		SELECT id, title, author, genres, status, total_chapters, description, cover_url, publication_year, created_at 
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
			&manga.CoverURL, &manga.PublicationYear, &manga.CreatedAt)
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

// CreateManga creates a new manga entry
func (s *Service) CreateManga(manga models.Manga) (*models.Manga, error) {
	// Set genres JSON
	if err := manga.SetGenres(manga.Genres); err != nil {
		return nil, fmt.Errorf("failed to set genres: %w", err)
	}

	// Set created time
	manga.CreatedAt = time.Now()

	// Insert into database
	_, err := s.db.Exec(`
		INSERT INTO manga 
		(id, title, author, genres, status, total_chapters, description, cover_url, publication_year, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		manga.ID, manga.Title, manga.Author, manga.GenresJSON,
		manga.Status, manga.TotalChapters, manga.Description,
		manga.CoverURL, manga.PublicationYear, manga.CreatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, fmt.Errorf("manga with ID '%s' already exists", manga.ID)
		}
		return nil, fmt.Errorf("failed to create manga: %w", err)
	}

	return &manga, nil
}

// UpdateManga updates an existing manga entry
func (s *Service) UpdateManga(id string, manga models.Manga) (*models.Manga, error) {
	// Check if manga exists
	existingManga, err := s.GetManga(id)
	if err != nil {
		return nil, err
	}

	// Set ID to ensure consistency
	manga.ID = id
	manga.CreatedAt = existingManga.CreatedAt

	// Set genres JSON
	if err := manga.SetGenres(manga.Genres); err != nil {
		return nil, fmt.Errorf("failed to set genres: %w", err)
	}

	// Update in database
	_, err = s.db.Exec(`
		UPDATE manga SET 
		title = ?, author = ?, genres = ?, status = ?, 
		total_chapters = ?, description = ?, cover_url = ?, publication_year = ?
		WHERE id = ?`,
		manga.Title, manga.Author, manga.GenresJSON, manga.Status,
		manga.TotalChapters, manga.Description, manga.CoverURL, manga.PublicationYear, id)

	if err != nil {
		return nil, fmt.Errorf("failed to update manga: %w", err)
	}

	return &manga, nil
}

// DeleteManga deletes a manga entry
func (s *Service) DeleteManga(id string) error {
	// Check if manga exists
	_, err := s.GetManga(id)
	if err != nil {
		return err
	}

	// Start transaction to ensure data consistency
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete user progress for this manga
	_, err = tx.Exec("DELETE FROM user_progress WHERE manga_id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete user progress: %w", err)
	}

	// Delete manga
	_, err = tx.Exec("DELETE FROM manga WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete manga: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetMangaStats returns statistics about manga in the database
func (s *Service) GetMangaStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total manga count
	var totalManga int
	err := s.db.QueryRow("SELECT COUNT(*) FROM manga").Scan(&totalManga)
	if err != nil {
		return nil, fmt.Errorf("failed to get total manga count: %w", err)
	}
	stats["total_manga"] = totalManga

	// Total chapters count
	var totalChapters int
	err = s.db.QueryRow("SELECT COUNT(*) FROM manga_chapters").Scan(&totalChapters)
	if err != nil {
		return nil, fmt.Errorf("failed to get total chapters count: %w", err)
	}
	stats["total_chapters"] = totalChapters

	// Count by status
	statusRows, err := s.db.Query("SELECT status, COUNT(*) FROM manga GROUP BY status")
	if err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}
	defer statusRows.Close()

	statusCounts := make(map[string]int)
	for statusRows.Next() {
		var status string
		var count int
		if err := statusRows.Scan(&status, &count); err != nil {
			log.Printf("Error scanning status row: %v", err)
			continue
		}
		statusCounts[status] = count
	}
	stats["by_status"] = statusCounts

	// Count by source
	sourceRows, err := s.db.Query("SELECT source, COUNT(DISTINCT manga_id) FROM manga_sources GROUP BY source")
	if err != nil {
		return nil, fmt.Errorf("failed to get source counts: %w", err)
	}
	defer sourceRows.Close()

	sourceCounts := make(map[string]int)
	for sourceRows.Next() {
		var source string
		var count int
		if err := sourceRows.Scan(&source, &count); err != nil {
			log.Printf("Error scanning source row: %v", err)
			continue
		}
		sourceCounts[source] = count
	}
	stats["by_source"] = sourceCounts

	// Chapter sources
	chapterSourceRows, err := s.db.Query("SELECT source, COUNT(*) FROM manga_chapters GROUP BY source")
	if err != nil {
		return nil, fmt.Errorf("failed to get chapter source counts: %w", err)
	}
	defer chapterSourceRows.Close()

	chapterSourceCounts := make(map[string]int)
	for chapterSourceRows.Next() {
		var source string
		var count int
		if err := chapterSourceRows.Scan(&source, &count); err != nil {
			log.Printf("Error scanning chapter source row: %v", err)
			continue
		}
		chapterSourceCounts[source] = count
	}
	stats["chapters_by_source"] = chapterSourceCounts

	// Average chapters per manga
	var avgChapters sql.NullFloat64
	err = s.db.QueryRow(`
		SELECT AVG(chapter_count) 
		FROM (SELECT COUNT(*) as chapter_count FROM manga_chapters GROUP BY manga_id)
	`).Scan(&avgChapters)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get average chapters: %w", err)
	}
	if avgChapters.Valid {
		stats["average_chapters_per_manga"] = avgChapters.Float64
	} else {
		stats["average_chapters_per_manga"] = 0.0
	}

	return stats, nil
}
