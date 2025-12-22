package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DB holds the database connection
var DB *sql.DB

// findProjectRoot finds the project root by looking for go.mod
// In Docker, falls back to /app or current directory
func findProjectRoot() (string, error) {
	// Check if we're in Docker (look for /app directory)
	if _, err := os.Stat("/app"); err == nil {
		return "/app", nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding go.mod, use current directory
			log.Println("Warning: go.mod not found, using current directory")
			return dir, nil
		}
		dir = parent
	}
}

// InitDatabase initializes the SQLite database connection and creates tables
func InitDatabase() error {
	// Find project root (where go.mod is located)
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Ensure data directory exists at project root
	dataDir := filepath.Join(projectRoot, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Database file path - always at project root
	dbPath := filepath.Join(dataDir, "mangahub.db")
	log.Printf("Using database at: %s", dbPath)

	// Open database connection
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables
	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// createTables creates the required database tables
func createTables() error {
	queries := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Manga table
		`CREATE TABLE IF NOT EXISTS manga (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			author TEXT,
			genres TEXT, -- JSON array as text
			status TEXT,
			total_chapters INTEGER,
			description TEXT,
			cover_url TEXT,
			publication_year INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// User progress table
		`CREATE TABLE IF NOT EXISTS user_progress (
			user_id TEXT,
			manga_id TEXT,
			current_chapter INTEGER DEFAULT 0,
			status TEXT DEFAULT 'plan_to_read', -- reading, completed, plan_to_read, dropped
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, manga_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
		)`,

		// Manga ratings table
		`CREATE TABLE IF NOT EXISTS manga_ratings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			manga_id TEXT NOT NULL,
			rating INTEGER NOT NULL CHECK(rating >= 1 AND rating <= 5),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, manga_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
		)`,

		// Manga chapters table - stores chapter metadata from MangaDex/MangaPlus
		`CREATE TABLE IF NOT EXISTS manga_chapters (
			id TEXT PRIMARY KEY,
			manga_id TEXT NOT NULL,
			chapter_number TEXT NOT NULL,
			title TEXT,
			volume TEXT,
			language TEXT DEFAULT 'en',
			pages INTEGER DEFAULT 0,
			source TEXT NOT NULL, -- 'mangadex' or 'mangaplus'
			source_chapter_id TEXT NOT NULL, -- ID in the external source
			scanlation_group TEXT, -- Name of scanlation group (for multiple versions)
			external_url TEXT, -- URL to external site (e.g., MangaPlus) for licensed manga
			is_external INTEGER DEFAULT 0, -- 1 if chapter is only available externally
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
		)`,

		// Manga source mapping table - links manga to external sources
		`CREATE TABLE IF NOT EXISTS manga_sources (
			manga_id TEXT NOT NULL,
			source TEXT NOT NULL, -- 'mangadex', 'mangaplus', 'mal'
			source_id TEXT NOT NULL, -- ID in the external source
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (manga_id, source),
			FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
		)`,

		// Create indexes for better performance
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_manga_title ON manga(title)`,
		`CREATE INDEX IF NOT EXISTS idx_manga_author ON manga(author)`,
		`CREATE INDEX IF NOT EXISTS idx_progress_user ON user_progress(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_progress_manga ON user_progress(manga_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ratings_manga ON manga_ratings(manga_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ratings_user ON manga_ratings(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chapters_manga ON manga_chapters(manga_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chapters_number ON manga_chapters(chapter_number)`,
		`CREATE INDEX IF NOT EXISTS idx_sources_manga ON manga_sources(manga_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sources_source ON manga_sources(source)`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return DB
}
