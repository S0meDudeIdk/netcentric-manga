package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := "./data/mangahub.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	log.Printf("Migrating database: %s", dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	log.Println("Step 1: Creating new library table...")
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS library (
			user_id TEXT NOT NULL,
			manga_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'plan_to_read', -- reading, completed, plan_to_read, dropped, on_hold, re_reading
			added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, manga_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
		)`)
	if err != nil {
		log.Fatalf("Failed to create library table: %v", err)
	}
	log.Println("✅ Library table created")

	log.Println("Step 2: Migrating existing data from user_progress to library...")
	// Copy status information to new library table
	_, err = tx.Exec(`
		INSERT OR IGNORE INTO library (user_id, manga_id, status, last_updated)
		SELECT user_id, manga_id, COALESCE(status, 'reading'), last_updated
		FROM user_progress
	`)
	if err != nil {
		log.Fatalf("Failed to migrate data to library: %v", err)
	}
	log.Println("✅ Data migrated to library table")

	log.Println("Step 3: Creating backup of user_progress...")
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS user_progress_backup AS
		SELECT * FROM user_progress
	`)
	if err != nil {
		log.Fatalf("Failed to create backup: %v", err)
	}
	log.Println("✅ Backup created: user_progress_backup")

	log.Println("Step 4: Dropping old user_progress table...")
	_, err = tx.Exec(`DROP TABLE user_progress`)
	if err != nil {
		log.Fatalf("Failed to drop old user_progress table: %v", err)
	}
	log.Println("✅ Old user_progress table dropped")

	log.Println("Step 5: Creating new user_progress table (without status)...")
	_, err = tx.Exec(`
		CREATE TABLE user_progress (
			user_id TEXT NOT NULL,
			manga_id TEXT NOT NULL,
			current_chapter INTEGER DEFAULT 0,
			last_read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, manga_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`)
	if err != nil {
		log.Fatalf("Failed to create new user_progress table: %v", err)
	}
	log.Println("✅ New user_progress table created")

	log.Println("Step 6: Migrating reading progress data...")
	_, err = tx.Exec(`
		INSERT INTO user_progress (user_id, manga_id, current_chapter, last_read_at)
		SELECT user_id, manga_id, current_chapter, last_updated
		FROM user_progress_backup
	`)
	if err != nil {
		log.Fatalf("Failed to migrate progress data: %v", err)
	}
	log.Println("✅ Progress data migrated")

	log.Println("Step 7: Creating indexes...")
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_library_user ON library(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_library_manga ON library(manga_id)`,
		`CREATE INDEX IF NOT EXISTS idx_library_status ON library(status)`,
		`CREATE INDEX IF NOT EXISTS idx_progress_user ON user_progress(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_progress_manga ON user_progress(manga_id)`,
	}
	for _, idx := range indexes {
		if _, err := tx.Exec(idx); err != nil {
			log.Fatalf("Failed to create index: %v", err)
		}
	}
	log.Println("✅ Indexes created")

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Println("\n🎉 Migration completed successfully!")
	log.Println("\nSummary:")
	log.Println("- ✅ New 'library' table created (tracks which manga are in user's library with status)")
	log.Println("- ✅ New 'user_progress' table created (tracks reading progress for ANY manga)")
	log.Println("- ✅ All existing data preserved and migrated")
	log.Println("- ✅ Backup saved as 'user_progress_backup' table")
	log.Println("\nNew schema:")
	log.Println("  library: user_id, manga_id, status, added_at, last_updated")
	log.Println("  user_progress: user_id, manga_id, current_chapter, last_read_at")

	// Print statistics
	var libraryCount, progressCount int
	db.QueryRow("SELECT COUNT(*) FROM library").Scan(&libraryCount)
	db.QueryRow("SELECT COUNT(*) FROM user_progress").Scan(&progressCount)

	fmt.Printf("\nDatabase statistics:\n")
	fmt.Printf("- Library entries: %d\n", libraryCount)
	fmt.Printf("- Progress records: %d\n", progressCount)
}
