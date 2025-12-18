package user

import (
	"database/sql"
	"fmt"
	"log"
	"mangahub/internal/auth"
	"mangahub/internal/external"
	"mangahub/pkg/database"
	"mangahub/pkg/models"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Service handles user-related operations
type Service struct {
	db        *sql.DB
	malClient *external.MALClient
}

// NewService creates a new user service
func NewService() *Service {
	return &Service{
		db:        database.GetDB(),
		malClient: external.NewMALClient(),
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

	// Get user from database by email or username
	err := s.db.QueryRow(`
		SELECT id, username, email, password_hash, created_at 
		FROM users WHERE email = ? OR username = ?`, req.Email, req.Email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid email/username or password")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if err := auth.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, fmt.Errorf("invalid email/username or password")
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
	// Use LEFT JOIN to include external manga that aren't in local manga table
	rows, err := s.db.Query(`
		SELECT up.manga_id, up.current_chapter, up.status, up.last_updated,
			   m.title, m.author, m.cover_url
		FROM user_progress up
		LEFT JOIN manga m ON up.manga_id = m.id
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
		OnHold:     []models.UserProgress{},
		ReReading:  []models.UserProgress{},
	}

	for rows.Next() {
		var progress models.UserProgress
		var title, author, coverURL sql.NullString

		err := rows.Scan(&progress.MangaID, &progress.CurrentChapter, &progress.Status,
			&progress.LastUpdated, &title, &author, &coverURL)
		if err != nil {
			log.Printf("Error scanning progress row: %v", err)
			continue
		}

		progress.UserID = userID

		// If manga details are null (external manga), fetch from MAL API
		if !title.Valid && strings.HasPrefix(progress.MangaID, "mal-") {
			// Extract MAL ID and fetch details
			malIDStr := strings.TrimPrefix(progress.MangaID, "mal-")
			if malID, err := strconv.Atoi(malIDStr); err == nil {
				if malManga, err := s.malClient.GetMangaByID(malID); err == nil {
					progress.Title = malManga.Title
					// Get author from authors array
					if len(malManga.Authors) > 0 {
						progress.Author = malManga.Authors[0].Node.FirstName + " " + malManga.Authors[0].Node.LastName
					}
					if malManga.MainPicture.Large != "" {
						progress.CoverURL = malManga.MainPicture.Large
					} else {
						progress.CoverURL = malManga.MainPicture.Medium
					}
				} else {
					log.Printf("Failed to fetch MAL manga %s: %v", malIDStr, err)
					progress.Title = "Unknown Manga"
				}
			}
		} else if !title.Valid && strings.HasPrefix(progress.MangaID, "mangadex-") {
			// For MangaDex, use a placeholder (could implement MangaDex API later)
			progress.Title = "MangaDex Manga"
			progress.Author = "Unknown"
		} else {
			// Use local manga data
			if title.Valid {
				progress.Title = title.String
			}
			if author.Valid {
				progress.Author = author.String
			}
			if coverURL.Valid {
				progress.CoverURL = coverURL.String
			}
		}

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
		case "on_hold":
			library.OnHold = append(library.OnHold, progress)
		case "re_reading":
			library.ReReading = append(library.ReReading, progress)
		}
	}

	return library, nil
}

// AddToLibrary adds a manga to user's library
func (s *Service) AddToLibrary(userID string, req models.AddToLibraryRequest) error {
	// Only check if manga exists for local manga (not external like MAL or MangaDex)
	// External manga IDs start with "mal-" or "mangadex-"
	isExternalManga := strings.HasPrefix(req.MangaID, "mal-") || strings.HasPrefix(req.MangaID, "mangadex-")

	if !isExternalManga {
		// Check if local manga exists
		var exists bool
		err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM manga WHERE id = ?)", req.MangaID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check manga existence: %w", err)
		}
		if !exists {
			return fmt.Errorf("manga not found")
		}
	}

	// Insert or update user progress
	_, err := s.db.Exec(`
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

// GetUserProgress retrieves user's reading progress for a specific manga (for TCP endpoint)
func (s *Service) GetUserProgress(userID, mangaID string) (*models.UserProgress, error) {
	query := `
		SELECT user_id, manga_id, current_chapter, status, last_updated
		FROM user_progress
		WHERE user_id = ? AND manga_id = ?
	`

	var progress models.UserProgress
	err := s.db.QueryRow(query, userID, mangaID).Scan(
		&progress.UserID,
		&progress.MangaID,
		&progress.CurrentChapter,
		&progress.Status,
		&progress.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, nil // User hasn't started reading yet
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user progress: %w", err)
	}

	return &progress, nil
}

// UpdateProfile updates a user's profile information
func (s *Service) UpdateProfile(userID string, username, email string) (*models.UserResponse, error) {
	// Check if user exists
	var existingID string
	err := s.db.QueryRow("SELECT id FROM users WHERE id = ?", userID).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	// Check if username is already taken by another user
	if username != "" {
		var existingUserID string
		err = s.db.QueryRow("SELECT id FROM users WHERE username = ? AND id != ?", username, userID).Scan(&existingUserID)
		if err == nil {
			return nil, fmt.Errorf("username already taken")
		} else if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to check username availability: %w", err)
		}
	}

	// Build update query dynamically based on provided fields
	updates := []string{}
	args := []interface{}{}

	if username != "" {
		updates = append(updates, "username = ?")
		args = append(args, username)
	}

	if email != "" {
		updates = append(updates, "email = ?")
		args = append(args, email)
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Add userID as last argument
	args = append(args, userID)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(updates, ", "))
	_, err = s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Get updated profile
	return s.GetProfile(userID)
}

// ChangePassword changes a user's password
func (s *Service) ChangePassword(userID, oldPassword, newPassword string) error {
	// Get current user
	var hashedPassword string
	err := s.db.QueryRow("SELECT password_hash FROM users WHERE id = ?", userID).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(oldPassword)); err != nil {
		return fmt.Errorf("incorrect old password")
	}

	// Hash new password
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	_, err = s.db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(newHashedPassword), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
