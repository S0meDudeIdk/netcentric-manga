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
