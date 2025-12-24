package models

import (
	"encoding/json"
	"time"
)

// Manga represents a manga series in the system
type Manga struct {
	ID              string    `json:"id" db:"id"`
	Title           string    `json:"title" db:"title"`
	Author          string    `json:"author" db:"author"`
	Genres          []string  `json:"genres" db:"-"` // Will be handled separately for DB
	GenresJSON      string    `json:"-" db:"genres"` // JSON string for database storage
	Status          string    `json:"status" db:"status"`
	TotalChapters   int       `json:"total_chapters" db:"total_chapters"`
	Description     string    `json:"description" db:"description"`
	CoverURL        string    `json:"cover_url" db:"cover_url"`
	PublicationYear int       `json:"publication_year" db:"publication_year"`
	Rating          float64   `json:"rating" db:"rating"`           // Average rating from users
	RatingCount     int       `json:"rating_count" db:"-"`          // Number of ratings
	UserRating      *int      `json:"user_rating,omitempty" db:"-"` // Current user's rating (0-10)
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// MarshalJSON handles the conversion of genres for JSON output
func (m *Manga) MarshalJSON() ([]byte, error) {
	type Alias Manga
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	return json.Marshal(aux)
}

// UnmarshalJSON handles the conversion of genres from JSON input
func (m *Manga) UnmarshalJSON(data []byte) error {
	type Alias Manga
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}

// SetGenres sets the genres and updates the JSON representation
func (m *Manga) SetGenres(genres []string) error {
	m.Genres = genres
	genresJSON, err := json.Marshal(genres)
	if err != nil {
		return err
	}
	m.GenresJSON = string(genresJSON)
	return nil
}

// GetGenres parses the JSON genres and sets the Genres field
func (m *Manga) GetGenres() error {
	if m.GenresJSON == "" {
		m.Genres = []string{}
		return nil
	}
	return json.Unmarshal([]byte(m.GenresJSON), &m.Genres)
}

// UserProgress represents a user's reading progress for a manga
type UserProgress struct {
	UserID         string    `json:"user_id" db:"user_id"`
	MangaID        string    `json:"manga_id" db:"manga_id"`
	CurrentChapter int       `json:"current_chapter" db:"current_chapter"`
	Status         string    `json:"status" db:"status"` // reading, completed, plan_to_read, dropped
	LastUpdated    time.Time `json:"last_updated" db:"last_updated"`
	// Manga details (populated from local DB or external API)
	Title    string `json:"title,omitempty"`
	Author   string `json:"author,omitempty"`
	CoverURL string `json:"cover_url,omitempty"`
}

// UserLibrary represents a user's manga library organized by status
type UserLibrary struct {
	Reading    []UserProgress `json:"reading"`
	Completed  []UserProgress `json:"completed"`
	PlanToRead []UserProgress `json:"plan_to_read"`
	Dropped    []UserProgress `json:"dropped"`
	OnHold     []UserProgress `json:"on_hold"`
	ReReading  []UserProgress `json:"re_reading"`
}

// MangaSearchRequest represents search parameters for manga
type MangaSearchRequest struct {
	Query  string   `json:"query" form:"query"`
	Genres []string `json:"genres" form:"genres"`
	Status string   `json:"status" form:"status"`
	Author string   `json:"author" form:"author"`
	Sort   string   `json:"sort" form:"sort"`
	Limit  int      `json:"limit" form:"limit"`
	Offset int      `json:"offset" form:"offset"`
}

// UpdateProgressRequest represents a request to update reading progress
type UpdateProgressRequest struct {
	MangaID        string `json:"manga_id" binding:"required"`
	CurrentChapter int    `json:"current_chapter" binding:"min=0"`
	Status         string `json:"status" binding:"required,oneof=reading completed plan_to_read dropped on_hold re_reading"`
}

// AddToLibraryRequest represents a request to add manga to user's library
type AddToLibraryRequest struct {
	MangaID string `json:"manga_id" binding:"required"`
	Status  string `json:"status" binding:"required,oneof=reading completed plan_to_read dropped on_hold re_reading"`
}

// CreateMangaRequest represents a request to create a new manga
type CreateMangaRequest struct {
	ID            string   `json:"id" binding:"required,min=1,max=100"`
	Title         string   `json:"title" binding:"required,min=1,max=200"`
	Author        string   `json:"author" binding:"required,min=1,max=100"`
	Genres        []string `json:"genres" binding:"required,min=1"`
	Status        string   `json:"status" binding:"required,oneof=ongoing completed hiatus dropped cancelled"`
	TotalChapters int      `json:"total_chapters" binding:"min=0"`
	Description   string   `json:"description" binding:"max=2000"`
	CoverURL      string   `json:"cover_url" binding:"omitempty,url"`
}

// UpdateMangaRequest represents a request to update manga information
type UpdateMangaRequest struct {
	Title         string   `json:"title" binding:"omitempty,min=1,max=200"`
	Author        string   `json:"author" binding:"omitempty,min=1,max=100"`
	Genres        []string `json:"genres" binding:"omitempty,min=1"`
	Status        string   `json:"status" binding:"omitempty,oneof=ongoing completed hiatus dropped cancelled"`
	TotalChapters int      `json:"total_chapters" binding:"omitempty,min=0"`
	Description   string   `json:"description" binding:"omitempty,max=2000"`
	CoverURL      string   `json:"cover_url" binding:"omitempty,url"`
}

// MangaListResponse represents a paginated list of manga
type MangaListResponse struct {
	Manga      []Manga `json:"manga"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	PerPage    int     `json:"per_page"`
	TotalPages int     `json:"total_pages"`
}

// LibraryStatsResponse represents user library statistics
type LibraryStatsResponse struct {
	TotalManga    int `json:"total_manga"`
	Reading       int `json:"reading"`
	Completed     int `json:"completed"`
	PlanToRead    int `json:"plan_to_read"`
	Dropped       int `json:"dropped"`
	TotalChapters int `json:"total_chapters_read"`
}

// ChapterInfo represents chapter metadata
type ChapterInfo struct {
	ID              string  `json:"id"`
	MangaID         string  `json:"manga_id"`
	ChapterNumber   string  `json:"chapter_number"`
	VolumeNumber    string  `json:"volume_number,omitempty"`
	Title           string  `json:"title,omitempty"`
	Language        string  `json:"language"`
	Pages           int     `json:"pages"`
	PublishedAt     string  `json:"published_at,omitempty"`
	Source          string  `json:"source"`           // "mangadex" or "mangaplus"
	ScanlationGroup string  `json:"scanlation_group"` // Scanlation group name
	ExternalUrl     *string `json:"external_url"`     // URL to external site for licensed manga
	IsExternal      bool    `json:"is_external"`      // true if chapter is only available externally
}

// ChapterPages represents the pages/images of a chapter
type ChapterPages struct {
	ChapterID  string   `json:"chapter_id"`
	MangaID    string   `json:"manga_id"`
	ChapterNum string   `json:"chapter_number"`
	Pages      []string `json:"pages"`  // Array of image URLs
	Source     string   `json:"source"` // "mangadex" or "mangaplus"
	BaseURL    string   `json:"base_url,omitempty"`
	Hash       string   `json:"hash,omitempty"`
}

// ChapterListRequest represents a request for chapter list
type ChapterListRequest struct {
	MangaID  string   `json:"manga_id" form:"manga_id" binding:"required"`
	Language []string `json:"language" form:"language"`
	Limit    int      `json:"limit" form:"limit"`
	Offset   int      `json:"offset" form:"offset"`
}

// ChapterListResponse represents chapter list response
type ChapterListResponse struct {
	Chapters []ChapterInfo `json:"chapters"`
	Total    int           `json:"total"`
	Limit    int           `json:"limit"`
	Offset   int           `json:"offset"`
}

// BatchUpdateRequest represents a request to update multiple manga progress
type BatchUpdateRequest struct {
	Updates []UpdateProgressRequest `json:"updates" binding:"required,dive"`
}

// LibraryFilterRequest represents filtering parameters for user library
type LibraryFilterRequest struct {
	Status string `json:"status" form:"status"`
	SortBy string `json:"sort_by" form:"sort_by"` // title, author, progress, updated
	Limit  int    `json:"limit" form:"limit"`
	Offset int    `json:"offset" form:"offset"`
}

// MangaRating represents a user's rating for a manga
type MangaRating struct {
	ID        int       `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	MangaID   string    `json:"manga_id" db:"manga_id"`
	Rating    int       `json:"rating" db:"rating"` // 0-10
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// RateMangaRequest represents a request to rate a manga
type RateMangaRequest struct {
	MangaID string `json:"manga_id" binding:"required"`
	Rating  int    `json:"rating" binding:"required,min=0,max=10"`
}

// MangaRatingStats represents rating statistics for a manga
type MangaRatingStats struct {
	MangaID            string      `json:"manga_id"`
	AverageRating      float64     `json:"average_rating"`
	TotalRatings       int         `json:"total_ratings"`
	UserRating         *int        `json:"user_rating,omitempty"` // Current user's rating if authenticated
	RatingDistribution map[int]int `json:"rating_distribution"`   // Distribution of ratings 1-5
}
