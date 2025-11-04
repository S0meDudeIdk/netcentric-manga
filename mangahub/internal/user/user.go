package user

import (
	"database/sql"
	"fmt"
	"log"
	"mangahub/internal/auth"
	"mangahub/pkg/database"
	"mangahub/pkg/models"
	"strings"
	"time"
)

// Service handles user-related operations
type Service struct {
	db *sql.DB
}

// NewService creates a new user service
func NewService() *Service {
	return &Service{
		db: database.GetDB(),
	}
}

// Register creates a new user account
func (s *Service) Register(req models.UserRegistration) (*models.AuthResponse, error) {
	// Check if username or email already exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ? OR email = ?)",
		req.Username, req.Email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("username or email already exists")
	}

	// Generate user ID
	userID, err := auth.GenerateUserID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate user ID: %w", err)
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user in database
	_, err = s.db.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at) 
		VALUES (?, ?, ?, ?, ?)`,
		userID, req.Username, req.Email, hashedPassword, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := auth.GenerateToken(userID, req.Username, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Return response
	userResponse := models.UserResponse{
		ID:        userID,
		Username:  req.Username,
		Email:     req.Email,
		CreatedAt: time.Now(),
	}

	return &models.AuthResponse{
		User:  userResponse,
		Token: token,
	}, nil
}

// Login authenticates a user and returns a token
func (s *Service) Login(req models.UserLogin) (*models.LoginResponse, error) {
	var user models.User

	// Get user from database
	err := s.db.QueryRow(`
		SELECT id, username, email, password_hash, created_at 
		FROM users WHERE email = ?`, req.Email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid email or password")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if err := auth.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Return response
	userResponse := models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}

	return &models.LoginResponse{
		User:  userResponse,
		Token: token,
	}, nil
}

// GetProfile returns user profile information
func (s *Service) GetProfile(userID string) (*models.UserResponse, error) {
	var user models.User

	err := s.db.QueryRow(`
		SELECT id, username, email, created_at 
		FROM users WHERE id = ?`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return &models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

// GetLibrary returns user's manga library organized by status
func (s *Service) GetLibrary(userID string) (*models.UserLibrary, error) {
	rows, err := s.db.Query(`
		SELECT up.manga_id, up.current_chapter, up.status, up.last_updated,
			   m.title, m.author, m.cover_url
		FROM user_progress up
		JOIN manga m ON up.manga_id = m.id
		WHERE up.user_id = ?
		ORDER BY up.last_updated DESC`, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to get user library: %w", err)
	}
	defer rows.Close()

	library := &models.UserLibrary{
		Reading:    []models.UserProgress{},
		Completed:  []models.UserProgress{},
		PlanToRead: []models.UserProgress{},
		Dropped:    []models.UserProgress{},
	}

	for rows.Next() {
		var progress models.UserProgress
		var title, author, coverURL string

		err := rows.Scan(&progress.MangaID, &progress.CurrentChapter, &progress.Status,
			&progress.LastUpdated, &title, &author, &coverURL)
		if err != nil {
			log.Printf("Error scanning progress row: %v", err)
			continue
		}

		progress.UserID = userID

		// Organize by status
		switch progress.Status {
		case "reading":
			library.Reading = append(library.Reading, progress)
		case "completed":
			library.Completed = append(library.Completed, progress)
		case "plan_to_read":
			library.PlanToRead = append(library.PlanToRead, progress)
		case "dropped":
			library.Dropped = append(library.Dropped, progress)
		}
	}

	return library, nil
}

// AddToLibrary adds a manga to user's library
func (s *Service) AddToLibrary(userID string, req models.AddToLibraryRequest) error {
	// Check if manga exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM manga WHERE id = ?)", req.MangaID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check manga existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("manga not found")
	}

	// Insert or update user progress
	_, err = s.db.Exec(`
		INSERT INTO user_progress (user_id, manga_id, status, last_updated)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, manga_id) DO UPDATE SET
			status = excluded.status,
			last_updated = excluded.last_updated`,
		userID, req.MangaID, req.Status, time.Now())

	if err != nil {
		return fmt.Errorf("failed to add manga to library: %w", err)
	}

	return nil
}

// UpdateProgress updates user's reading progress for a manga
func (s *Service) UpdateProgress(userID string, req models.UpdateProgressRequest) error {
	// Check if user has this manga in their library
	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM user_progress WHERE user_id = ? AND manga_id = ?)`,
		userID, req.MangaID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check progress existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("manga not found in user's library")
	}

	// Update progress
	_, err = s.db.Exec(`
		UPDATE user_progress 
		SET current_chapter = ?, status = ?, last_updated = ?
		WHERE user_id = ? AND manga_id = ?`,
		req.CurrentChapter, req.Status, time.Now(), userID, req.MangaID)

	if err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}

	return nil
}

// SearchUsers searches for users by username (for admin or social features)
func (s *Service) SearchUsers(query string, limit int) ([]models.UserResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	rows, err := s.db.Query(`
		SELECT id, username, email, created_at 
		FROM users 
		WHERE username LIKE ? 
		ORDER BY username 
		LIMIT ?`, "%"+strings.ToLower(query)+"%", limit)

	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	var users []models.UserResponse
	for rows.Next() {
		var user models.UserResponse
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
		if err != nil {
			log.Printf("Error scanning user row: %v", err)
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// GetLibraryStats returns statistics about user's library
func (s *Service) GetLibraryStats(userID string) (*models.LibraryStatsResponse, error) {
	var stats models.LibraryStatsResponse

	// Get counts by status
	rows, err := s.db.Query(`
		SELECT status, COUNT(*), COALESCE(SUM(current_chapter), 0)
		FROM user_progress 
		WHERE user_id = ? 
		GROUP BY status`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get library stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count, chapters int
		if err := rows.Scan(&status, &count, &chapters); err != nil {
			log.Printf("Error scanning stats row: %v", err)
			continue
		}

		switch status {
		case "reading":
			stats.Reading = count
		case "completed":
			stats.Completed = count
		case "plan_to_read":
			stats.PlanToRead = count
		case "dropped":
			stats.Dropped = count
		}
		stats.TotalChapters += chapters
	}

	stats.TotalManga = stats.Reading + stats.Completed + stats.PlanToRead + stats.Dropped
	return &stats, nil
}

// GetFilteredLibrary returns user's library with filtering and sorting
func (s *Service) GetFilteredLibrary(userID string, status string, sortBy string, limit, offset int) ([]models.UserProgress, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT up.manga_id, up.current_chapter, up.status, up.last_updated,
			   m.title, m.author, m.cover_url, m.total_chapters
		FROM user_progress up
		JOIN manga m ON up.manga_id = m.id
		WHERE up.user_id = ?`
	args := []interface{}{userID}

	// Add status filter if specified
	if status != "" {
		query += ` AND up.status = ?`
		args = append(args, status)
	}

	// Add sorting
	switch sortBy {
	case "title":
		query += ` ORDER BY m.title`
	case "author":
		query += ` ORDER BY m.author`
	case "progress":
		query += ` ORDER BY up.current_chapter DESC`
	case "updated":
		query += ` ORDER BY up.last_updated DESC`
	default:
		query += ` ORDER BY up.last_updated DESC`
	}

	query += ` LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered library: %w", err)
	}
	defer rows.Close()

	var progressList []models.UserProgress
	for rows.Next() {
		var progress models.UserProgress
		var title, author, coverURL string
		var totalChapters int

		err := rows.Scan(&progress.MangaID, &progress.CurrentChapter, &progress.Status,
			&progress.LastUpdated, &title, &author, &coverURL, &totalChapters)
		if err != nil {
			log.Printf("Error scanning progress row: %v", err)
			continue
		}

		progress.UserID = userID
		progressList = append(progressList, progress)
	}

	return progressList, nil
}

// BatchUpdateProgress updates progress for multiple manga
func (s *Service) BatchUpdateProgress(userID string, updates []models.UpdateProgressRequest) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO user_progress (user_id, manga_id, current_chapter, status, last_updated)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(user_id, manga_id) DO UPDATE SET
			current_chapter = excluded.current_chapter,
			status = excluded.status,
			last_updated = excluded.last_updated`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, update := range updates {
		// Check if manga exists
		var exists bool
		err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM manga WHERE id = ?)", update.MangaID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check manga existence: %w", err)
		}
		if !exists {
			return fmt.Errorf("manga with ID '%s' not found", update.MangaID)
		}

		// Execute update
		_, err = stmt.Exec(userID, update.MangaID, update.CurrentChapter, update.Status, time.Now())
		if err != nil {
			return fmt.Errorf("failed to update progress for manga %s: %w", update.MangaID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RemoveFromLibrary removes manga from user's library
func (s *Service) RemoveFromLibrary(userID, mangaID string) error {
	result, err := s.db.Exec("DELETE FROM user_progress WHERE user_id = ? AND manga_id = ?", userID, mangaID)
	if err != nil {
		return fmt.Errorf("failed to remove manga from library: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("manga not found in user's library")
	}

	return nil
}

// GetReadingRecommendations returns manga recommendations based on user's library
func (s *Service) GetReadingRecommendations(userID string, limit int) ([]models.Manga, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	// Simple recommendation: suggest manga with similar genres to user's completed/reading manga
	query := `
		SELECT DISTINCT m.id, m.title, m.author, m.genres, m.status, 
			   m.total_chapters, m.description, m.cover_url, m.created_at
		FROM manga m
		WHERE m.id NOT IN (
			SELECT manga_id FROM user_progress WHERE user_id = ?
		)
		AND EXISTS (
			SELECT 1 FROM user_progress up 
			JOIN manga um ON up.manga_id = um.id 
			WHERE up.user_id = ? 
			AND up.status IN ('reading', 'completed')
			AND (
				m.genres LIKE '%' || JSON_EXTRACT(um.genres, '$[0]') || '%' OR
				m.genres LIKE '%' || JSON_EXTRACT(um.genres, '$[1]') || '%'
			)
		)
		ORDER BY m.title
		LIMIT ?`

	rows, err := s.db.Query(query, userID, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}
	defer rows.Close()

	var recommendations []models.Manga
	for rows.Next() {
		var manga models.Manga
		err := rows.Scan(&manga.ID, &manga.Title, &manga.Author, &manga.GenresJSON,
			&manga.Status, &manga.TotalChapters, &manga.Description,
			&manga.CoverURL, &manga.CreatedAt)
		if err != nil {
			log.Printf("Error scanning recommendation row: %v", err)
			continue
		}

		// Parse genres
		if err := manga.GetGenres(); err != nil {
			log.Printf("Error parsing genres for manga %s: %v", manga.ID, err)
			manga.Genres = []string{}
		}

		recommendations = append(recommendations, manga)
	}

	return recommendations, nil
}
