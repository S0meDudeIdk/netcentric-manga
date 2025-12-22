package manga

import (
	"database/sql"
	"fmt"
	"log"
	"mangahub/internal/external"
	"mangahub/pkg/database"
	"mangahub/pkg/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SyncService handles syncing manga from external sources to local database
type SyncService struct {
	db              *sql.DB
	jikanClient     *external.JikanClient
	mangaDexClient  *external.MangaDexClient
	mangaPlusClient *external.MangaPlusClient
}

// NewSyncService creates a new sync service
func NewSyncService(jikan *external.JikanClient) *SyncService {
	return &SyncService{
		db:              database.GetDB(),
		jikanClient:     jikan,
		mangaDexClient:  external.NewMangaDexClient(),
		mangaPlusClient: external.NewMangaPlusClient(),
	}
}

// SyncFromMAL fetches manga from MAL and stores only those with available chapters
func (s *SyncService) SyncFromMAL(query string, limit int) (*SyncResult, error) {
	log.Printf("Starting sync from MAL: query=%s, limit=%d", query, limit)

	// Test database connection first
	db := database.GetDB()
	if db == nil {
		log.Printf("ERROR: Database connection is nil!")
		return nil, fmt.Errorf("database connection is nil")
	}
	log.Printf("Database connection OK")

	// Test database with a simple query
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM manga").Scan(&count)
	if err != nil {
		log.Printf("ERROR: Database query test failed: %v", err)
		return nil, fmt.Errorf("database query test failed: %v", err)
	}
	log.Printf("Database test OK, current manga count: %d", count)

	// Fetch manga from MAL/Jikan
	log.Printf("Searching MAL via Jikan...")
	malManga, err := s.jikanClient.SearchManga(query, 1, limit)
	if err != nil {
		log.Printf("ERROR: Jikan search failed: %v", err)
		return nil, fmt.Errorf("failed to search MAL: %w", err)
	}

	if malManga == nil {
		log.Printf("ERROR: Jikan returned nil results")
		return nil, fmt.Errorf("jikan returned nil results")
	}

	if len(malManga.Data) == 0 {
		log.Printf("WARNING: No manga found for query: %s", query)
		return &SyncResult{
			TotalFetched: 0,
			Synced:       0,
			Skipped:      0,
			Failed:       0,
			Details:      []string{},
		}, nil
	}

	result := &SyncResult{
		TotalFetched: len(malManga.Data),
		Synced:       0,
		Skipped:      0,
		Failed:       0,
		Details:      []string{},
	}

	log.Printf("Fetched %d manga from MAL", result.TotalFetched)

	// Process each manga
	for i, malData := range malManga.Data {
		log.Printf("Processing manga %d/%d: %s (MAL ID: %d)", i+1, len(malManga.Data), malData.Title, malData.MalID)

		// Try to find chapters on MangaDex
		chapters, mangaDexID, source := s.findChapters(malData.Title)

		if len(chapters) == 0 {
			result.Skipped++
			result.Details = append(result.Details, fmt.Sprintf("❌ Skipped '%s': No chapters found", malData.Title))
			log.Printf("  No chapters found, skipping")
			continue
		}

		log.Printf("  Found %d chapters from %s", len(chapters), source)

		// Convert MAL manga to local manga model
		manga := s.convertMALToManga(malData, mangaDexID, source)

		// Store manga in database
		malIDStr := fmt.Sprintf("%d", malData.MalID)
		log.Printf("  Storing manga in database...")
		if err := s.storeManga(manga, malIDStr, mangaDexID, source); err != nil {
			result.Failed++
			result.Details = append(result.Details, fmt.Sprintf("❌ Failed '%s': %v", malData.Title, err))
			log.Printf("  ERROR: Failed to store manga: %v", err)
			continue
		}
		log.Printf("  Manga stored successfully")

		// Store chapters
		log.Printf("  Storing %d chapters...", len(chapters))
		stored := 0
		for j, chapter := range chapters {
			if err := s.storeChapter(manga.ID, chapter); err != nil {
				log.Printf("  WARNING: Failed to store chapter %d/%d (%s): %v", j+1, len(chapters), chapter.ChapterNumber, err)
			} else {
				stored++
			}
		}

		// Update total_chapters in database with actual stored count
		if stored > 0 {
			log.Printf("  Updating total_chapters to %d...", stored)
			_, err := s.db.Exec(`UPDATE manga SET total_chapters = ? WHERE id = ?`, stored, manga.ID)
			if err != nil {
				log.Printf("  WARNING: Failed to update total_chapters: %v", err)
			} else {
				log.Printf("  Total chapters updated successfully")
			}
		}

		result.Synced++
		result.Details = append(result.Details, fmt.Sprintf("✅ Synced '%s': %d chapters from %s", malData.Title, stored, source))
		log.Printf("  Successfully synced with %d/%d chapters stored", stored, len(chapters))
	}

	log.Printf("Sync completed: %d synced, %d skipped, %d failed", result.Synced, result.Skipped, result.Failed)
	return result, nil
}

// SyncTopManga fetches top manga from MAL and stores those with chapters
func (s *SyncService) SyncTopManga(limit int) (*SyncResult, error) {
	log.Printf("Starting auto-sync of top manga from MAL (limit: %d)", limit)

	// Fetch top manga from Jikan
	topManga, err := s.jikanClient.GetTopManga(1, limit)
	if err != nil {
		log.Printf("ERROR: Failed to fetch top manga: %v", err)
		return nil, fmt.Errorf("failed to fetch top manga: %w", err)
	}

	if topManga == nil || len(topManga.Data) == 0 {
		log.Printf("WARNING: No top manga returned")
		return &SyncResult{
			TotalFetched: 0,
			Synced:       0,
			Skipped:      0,
			Failed:       0,
			Details:      []string{},
		}, nil
	}

	result := &SyncResult{
		TotalFetched: len(topManga.Data),
		Synced:       0,
		Skipped:      0,
		Failed:       0,
		Details:      []string{},
	}

	log.Printf("Fetched %d top manga from MAL", result.TotalFetched)

	// Process each manga
	for i, malData := range topManga.Data {
		log.Printf("Processing manga %d/%d: %s (MAL ID: %d)", i+1, len(topManga.Data), malData.Title, malData.MalID)

		// Check if already in database
		var exists bool
		mangaID := fmt.Sprintf("mal-%d", malData.MalID)
		err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM manga WHERE id = ?)", mangaID).Scan(&exists)
		if err == nil && exists {
			log.Printf("  Already in database, skipping")
			continue
		}

		// Try to find chapters on MangaDex
		chapters, mangaDexID, source := s.findChapters(malData.Title)

		if len(chapters) == 0 {
			result.Skipped++
			log.Printf("  No chapters found, skipping")
			continue
		}

		log.Printf("  Found %d chapters from %s", len(chapters), source)

		// Convert MAL manga to local manga model
		manga := s.convertMALToManga(malData, mangaDexID, source)

		// Store manga in database
		malIDStr := fmt.Sprintf("%d", malData.MalID)
		if err := s.storeManga(manga, malIDStr, mangaDexID, source); err != nil {
			result.Failed++
			log.Printf("  ERROR: Failed to store manga: %v", err)
			continue
		}

		// Store chapters
		stored := 0
		for _, chapter := range chapters {
			if err := s.storeChapter(manga.ID, chapter); err != nil {
				log.Printf("  WARNING: Failed to store chapter %s: %v", chapter.ChapterNumber, err)
			} else {
				stored++
			}
		}

		// Update total_chapters in database with actual stored count
		if stored > 0 {
			log.Printf("  Updating total_chapters to %d...", stored)
			_, err := s.db.Exec(`UPDATE manga SET total_chapters = ? WHERE id = ?`, stored, manga.ID)
			if err != nil {
				log.Printf("  WARNING: Failed to update total_chapters: %v", err)
			} else {
				log.Printf("  Total chapters updated successfully")
			}
		}

		result.Synced++
		log.Printf("  Successfully synced with %d chapters", stored)
	}

	log.Printf("Auto-sync complete: fetched=%d, synced=%d, skipped=%d, failed=%d",
		result.TotalFetched, result.Synced, result.Skipped, result.Failed)

	return result, nil
}

// SyncFromMangaDex fetches manga directly from MangaDex with chapters
func (s *SyncService) SyncFromMangaDex(maxManga int) (*SyncResult, error) {
	log.Printf("Starting MangaDex sync (max: %d manga, 0 = unlimited)", maxManga)

	result := &SyncResult{
		TotalFetched: 0,
		Synced:       0,
		Skipped:      0,
		Failed:       0,
		Details:      []string{},
	}

	limit := 100 // MangaDex API limit per request
	offset := 0
	totalSynced := 0
	unlimited := maxManga == 0

	for unlimited || totalSynced < maxManga {
		// Rate limiting - wait between requests
		if offset > 0 {
			time.Sleep(300 * time.Millisecond) // Slightly longer delay to avoid rate limits
		}

		log.Printf("Fetching manga batch: offset=%d, limit=%d", offset, limit)

		// Fetch manga list from MangaDex
		mangaList, err := s.mangaDexClient.GetMangaList(limit, offset)
		if err != nil {
			log.Printf("ERROR: Failed to fetch manga list: %v", err)
			// Don't fail completely, just log and continue
			break
		}

		if len(mangaList.Data) == 0 {
			log.Printf("No more manga to fetch")
			break
		}

		log.Printf("Processing %d manga from MangaDex (total available: %d)", len(mangaList.Data), mangaList.Total)
		result.TotalFetched += len(mangaList.Data)

		// Process each manga
		for _, mdManga := range mangaList.Data {
			if !unlimited && totalSynced >= maxManga {
				break
			}

			mangaID := "md-" + mdManga.ID
			title := s.getMangaDexTitle(mdManga)

			log.Printf("  Processing: %s (ID: %s)", title, mdManga.ID)

			// Check if already in database
			var mangaExists bool
			err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM manga WHERE id = ?)", mangaID).Scan(&mangaExists)
			if err != nil {
				log.Printf("    ERROR: Failed to check manga existence: %v", err)
				result.Failed++
				continue
			}

			// Check if chapters exist
			var chapterCount int
			err = s.db.QueryRow("SELECT COUNT(*) FROM manga_chapters WHERE manga_id = ?", mangaID).Scan(&chapterCount)
			if err != nil {
				log.Printf("    ERROR: Failed to check chapter count: %v", err)
				result.Failed++
				continue
			}

			// If manga exists and has chapters, skip
			if mangaExists && chapterCount > 0 {
				log.Printf("    Already exists with %d chapters, skipping", chapterCount)
				result.Skipped++
				continue
			}

			// Get chapters for this manga
			chapters, err := s.mangaDexClient.GetMangaChapterFeed(mdManga.ID, 500, 0, []string{"en"})
			if err != nil {
				log.Printf("    ERROR: Failed to get chapters: %v", err)
				result.Failed++
				continue
			}

			if len(chapters.Data) == 0 {
				log.Printf("    No chapters found, skipping")
				result.Skipped++
				continue
			}

			log.Printf("    Found %d chapters", len(chapters.Data))

			// Convert to local manga model
			manga := s.convertMangaDexToManga(mdManga)

			// Store manga only if it doesn't exist
			if !mangaExists {
				if err := s.storeMangaDirect(manga, mdManga.ID); err != nil {
					log.Printf("    ERROR: Failed to store manga: %v", err)
					result.Failed++
					continue
				}
				log.Printf("    Manga stored successfully")
			} else {
				log.Printf("    Manga already exists, updating chapters only")
			}

			// Store chapters (whether manga is new or existing)
			stored := 0
			for _, ch := range chapters.Data {
				chNum := ""
				if ch.Attributes.Chapter != nil {
					chNum = *ch.Attributes.Chapter
				}
				vol := ""
				if ch.Attributes.Volume != nil {
					vol = *ch.Attributes.Volume
				}

				// Extract scanlation group from relationships
				scanlationGroup := "Unknown"
				for _, rel := range ch.Relationships {
					if rel.Type == "scanlation_group" && rel.Attributes != nil {
						if name, ok := rel.Attributes["name"].(string); ok && name != "" {
							scanlationGroup = name
							break
						}
					}
				}

				chapterInfo := ChapterInfo{
					ChapterNumber:   chNum,
					Title:           ch.Attributes.Title,
					Volume:          vol,
					Language:        ch.Attributes.TranslatedLanguage,
					Pages:           ch.Attributes.Pages,
					SourceChapterID: ch.ID,
					ScanlationGroup: scanlationGroup,
					ExternalUrl:     ch.Attributes.ExternalUrl,
				}

				if err := s.storeChapter(manga.ID, chapterInfo); err != nil {
					log.Printf("    WARNING: Failed to store chapter %s: %v", chNum, err)
				} else {
					stored++
				}
			}

			// Update total_chapters in database with actual stored count
			if stored > 0 {
				log.Printf("    Updating total_chapters to %d...", stored)
				_, err := s.db.Exec(`UPDATE manga SET total_chapters = ? WHERE id = ?`, stored, manga.ID)
				if err != nil {
					log.Printf("    WARNING: Failed to update total_chapters: %v", err)
				} else {
					log.Printf("    Total chapters updated successfully")
				}
			}

			result.Synced++
			totalSynced++
			log.Printf("    Successfully synced with %d chapters", stored)
		}

		offset += limit

		// Check if we've reached the end
		if offset >= mangaList.Total {
			break
		}
	}

	log.Printf("MangaDex sync complete: fetched=%d, synced=%d, skipped=%d, failed=%d",
		result.TotalFetched, result.Synced, result.Skipped, result.Failed)

	return result, nil
}

// getMangaDexTitle gets the best title from MangaDex manga
func (s *SyncService) getMangaDexTitle(manga external.MangaDexManga) string {
	if title, ok := manga.Attributes.Title["en"]; ok && title != "" {
		return title
	}
	// Try romaji
	if title, ok := manga.Attributes.Title["ja-ro"]; ok && title != "" {
		return title
	}
	// Try any title
	for _, title := range manga.Attributes.Title {
		if title != "" {
			return title
		}
	}
	return "Unknown Title"
}

// convertMangaDexToManga converts MangaDex manga to local model
func (s *SyncService) convertMangaDexToManga(mdManga external.MangaDexManga) *models.Manga {
	mangaID := "md-" + mdManga.ID
	title := s.getMangaDexTitle(mdManga)

	// Get author
	author := "Unknown"
	for _, rel := range mdManga.Relationships {
		if rel.Type == "author" && rel.Attributes != nil {
			if name, ok := rel.Attributes["name"].(string); ok {
				author = name
				break
			}
		}
	}

	// Get genres
	genres := []string{}
	if mdManga.Attributes.Tags != nil {
		for _, tag := range mdManga.Attributes.Tags {
			if name, ok := tag.Attributes.Name["en"]; ok {
				genres = append(genres, name)
			}
		}
	}

	// Get cover URL
	coverURL := ""
	for _, rel := range mdManga.Relationships {
		if rel.Type == "cover_art" && rel.Attributes != nil {
			if fileName, ok := rel.Attributes["fileName"].(string); ok {
				coverURL = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", mdManga.ID, fileName)
				break
			}
		}
	}

	// Get description
	description := ""
	if desc, ok := mdManga.Attributes.Description["en"]; ok {
		description = desc
	}

	// Get publication year
	publicationYear := 0
	if mdManga.Attributes.Year > 0 {
		publicationYear = mdManga.Attributes.Year
		log.Printf("    DEBUG: Found year %d from MangaDex API", publicationYear)
	} else {
		log.Printf("    DEBUG: No year in MangaDex data (Year field = %d)", mdManga.Attributes.Year)
	}

	return &models.Manga{
		ID:              mangaID,
		Title:           title,
		Author:          author,
		Genres:          genres,
		Status:          strings.ToLower(mdManga.Attributes.Status),
		TotalChapters:   0, // Will be updated from database count
		Description:     description,
		CoverURL:        coverURL,
		PublicationYear: publicationYear,
		CreatedAt:       time.Now(),
	}
}

// storeMangaDirect stores manga directly with MangaDex source
func (s *SyncService) storeMangaDirect(manga *models.Manga, mangaDexID string) error {
	// Serialize genres
	if err := manga.SetGenres(manga.Genres); err != nil {
		return fmt.Errorf("failed to serialize genres: %w", err)
	}

	// Insert manga
	result, err := s.db.Exec(`
		INSERT OR IGNORE INTO manga (id, title, author, genres, status, total_chapters, description, cover_url, publication_year, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, manga.ID, manga.Title, manga.Author, manga.GenresJSON, manga.Status,
		manga.TotalChapters, manga.Description, manga.CoverURL, manga.PublicationYear, manga.CreatedAt.Format(time.RFC3339))

	if err != nil {
		return fmt.Errorf("failed to insert manga: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil // Already exists
	}

	// Store MangaDex source mapping
	_, err = s.db.Exec(`
		INSERT OR IGNORE INTO manga_sources (manga_id, source, source_id)
		VALUES (?, 'mangadex', ?)
	`, manga.ID, mangaDexID)

	if err != nil {
		log.Printf("WARNING: Failed to store MangaDex source mapping: %v", err)
	}

	return nil
}

// findChapters searches for chapters on MangaDex and MangaPlus
func (s *SyncService) findChapters(title string) ([]ChapterInfo, string, string) {
	// Try MangaDex first
	log.Printf("  Searching MangaDex for: %s", title)
	searchResults, err := s.mangaDexClient.SearchManga(title, 1)
	if err != nil {
		log.Printf("  MangaDex search error: %v", err)
	} else if searchResults == nil {
		log.Printf("  MangaDex search returned nil results")
	} else if len(searchResults.Data) == 0 {
		log.Printf("  MangaDex search returned no results")
	} else {
		mangaDexID := searchResults.Data[0].ID
		log.Printf("  Found MangaDex ID: %s", mangaDexID)

		// Get chapter list
		chapters, err := s.mangaDexClient.GetMangaChapterFeed(mangaDexID, 100, 0, []string{"en"})
		if err != nil {
			log.Printf("  MangaDex chapter feed error: %v", err)
		} else if chapters == nil {
			log.Printf("  MangaDex chapter feed returned nil")
		} else if len(chapters.Data) == 0 {
			log.Printf("  MangaDex chapter feed returned no chapters")
		} else {
			log.Printf("  Found %d chapters on MangaDex", len(chapters.Data))
			chapterInfos := make([]ChapterInfo, 0, len(chapters.Data))
			for _, ch := range chapters.Data {
				chNum := ""
				if ch.Attributes.Chapter != nil {
					chNum = *ch.Attributes.Chapter
				}
				vol := ""
				if ch.Attributes.Volume != nil {
					vol = *ch.Attributes.Volume
				}
				chapterInfos = append(chapterInfos, ChapterInfo{
					ChapterNumber:   chNum,
					Title:           ch.Attributes.Title,
					Volume:          vol,
					Language:        ch.Attributes.TranslatedLanguage,
					Pages:           ch.Attributes.Pages,
					SourceChapterID: ch.ID,
				})
			}
			return chapterInfos, mangaDexID, "mangadex"
		}
	}

	// Try MangaPlus if MangaDex fails
	log.Printf("  MangaDex failed, trying MangaPlus")
	// MangaPlus doesn't have a reliable search API, so we skip it for now
	// You can implement MangaPlus search if you have the API access

	return nil, "", ""
}

// convertMALToManga converts MAL/Jikan manga to local manga model
func (s *SyncService) convertMALToManga(malData external.JikanManga, mangaDexID, source string) *models.Manga {
	mangaID := uuid.New().String()

	// Use MangaDex ID if available, otherwise use MAL ID
	if mangaDexID != "" {
		mangaID = "md-" + mangaDexID
	} else {
		mangaID = fmt.Sprintf("mal-%d", malData.MalID)
	}

	// Convert genres
	genres := make([]string, len(malData.Genres))
	for i, g := range malData.Genres {
		genres[i] = g.Name
	}

	// Get publication year from Published date
	publicationYear := 0
	if malData.Published.From != "" {
		if t, err := time.Parse("2006-01-02T15:04:05-07:00", malData.Published.From); err == nil {
			publicationYear = t.Year()
		}
	}

	return &models.Manga{
		ID:              mangaID,
		Title:           malData.Title,
		Author:          s.extractAuthor(malData.Authors),
		Genres:          genres,
		Status:          strings.ToLower(malData.Status),
		TotalChapters:   malData.Chapters,
		Description:     malData.Synopsis,
		CoverURL:        s.getCoverURL(malData.Images),
		PublicationYear: publicationYear,
		CreatedAt:       time.Now(),
	}
}

// extractAuthor gets the first author from the list
func (s *SyncService) extractAuthor(authors []external.JikanAuthor) string {
	if len(authors) > 0 {
		return authors[0].Name
	}
	return "Unknown"
}

// getCoverURL extracts cover URL from images
func (s *SyncService) getCoverURL(images external.JikanImages) string {
	if images.JPG.LargeImageURL != "" {
		return images.JPG.LargeImageURL
	}
	if images.JPG.ImageURL != "" {
		return images.JPG.ImageURL
	}
	return ""
}

// storeManga stores manga and its source mappings in the database
func (s *SyncService) storeManga(manga *models.Manga, malID, mangaDexID, source string) error {
	log.Printf("    Checking if manga exists: %s", manga.ID)
	// Check if manga already exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM manga WHERE id = ?)", manga.ID).Scan(&exists)
	if err != nil {
		log.Printf("    ERROR: Failed to check existence: %v", err)
		return fmt.Errorf("failed to check manga existence: %w", err)
	}

	if exists {
		log.Printf("    Manga already exists, updating publication_year if needed...")
		// Update publication year if we have one and it's currently NULL/0
		if manga.PublicationYear > 0 {
			_, err = s.db.Exec(`
				UPDATE manga 
				SET publication_year = ? 
				WHERE id = ? AND (publication_year IS NULL OR publication_year = 0)
			`, manga.PublicationYear, manga.ID)
			if err != nil {
				log.Printf("    ERROR: Failed to update publication_year: %v", err)
			} else {
				log.Printf("    ✓ Updated publication_year to %d", manga.PublicationYear)
			}
		}
		return nil
	}

	log.Printf("    Manga is new, inserting...")
	// Serialize genres
	if err := manga.SetGenres(manga.Genres); err != nil {
		log.Printf("    ERROR: Failed to serialize genres: %v", err)
		return fmt.Errorf("failed to serialize genres: %w", err)
	}

	// Insert manga
	log.Printf("    Executing INSERT for manga...")
	result, err := s.db.Exec(`
		INSERT INTO manga (id, title, author, genres, status, total_chapters, description, cover_url, publication_year, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, manga.ID, manga.Title, manga.Author, manga.GenresJSON, manga.Status,
		manga.TotalChapters, manga.Description, manga.CoverURL, manga.PublicationYear, manga.CreatedAt.Format(time.RFC3339))

	if err != nil {
		log.Printf("    ERROR: Failed to insert manga: %v", err)
		return fmt.Errorf("failed to insert manga: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("    Manga inserted successfully, rows affected: %d", rowsAffected)

	// Store source mappings
	if malID != "" {
		log.Printf("    Storing MAL source mapping (ID: %s)", malID)
		_, err = s.db.Exec(`
			INSERT OR IGNORE INTO manga_sources (manga_id, source, source_id)
			VALUES (?, 'mal', ?)
		`, manga.ID, malID)
		if err != nil {
			log.Printf("    WARNING: Failed to store MAL source mapping: %v", err)
		} else {
			log.Printf("    MAL source mapping stored")
		}
	}

	if mangaDexID != "" && source == "mangadex" {
		log.Printf("    Storing MangaDex source mapping (ID: %s)", mangaDexID)
		_, err = s.db.Exec(`
			INSERT OR IGNORE INTO manga_sources (manga_id, source, source_id)
			VALUES (?, 'mangadex', ?)
		`, manga.ID, mangaDexID)
		if err != nil {
			log.Printf("    WARNING: Failed to store MangaDex source mapping: %v", err)
		} else {
			log.Printf("    MangaDex source mapping stored")
		}
	}

	return nil
}

// storeChapter stores chapter metadata in the database
func (s *SyncService) storeChapter(mangaID string, chapter ChapterInfo) error {
	// Use source_chapter_id to create unique ID (allows multiple scanlations per chapter)
	chapterID := fmt.Sprintf("%s-ch-%s", mangaID, chapter.SourceChapterID)

	// Determine if external and what source
	isExternal := 0
	source := "mangadex"
	if chapter.ExternalUrl != nil && *chapter.ExternalUrl != "" {
		isExternal = 1
		if strings.Contains(*chapter.ExternalUrl, "mangaplus.shueisha.co.jp") {
			source = "mangaplus"
		}
	}

	// Check if optional columns exist
	hasExternalUrlColumn := s.checkColumnExists("manga_chapters", "external_url")
	hasScanlationGroupColumn := s.checkColumnExists("manga_chapters", "scanlation_group")

	// Build query based on column availability
	var result sql.Result
	var err error

	if hasExternalUrlColumn && hasScanlationGroupColumn {
		result, err = s.db.Exec(`
			INSERT OR REPLACE INTO manga_chapters 
			(id, manga_id, chapter_number, title, volume, language, pages, source, source_chapter_id, scanlation_group, external_url, is_external)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, chapterID, mangaID, chapter.ChapterNumber, chapter.Title, chapter.Volume,
			chapter.Language, chapter.Pages, source, chapter.SourceChapterID, chapter.ScanlationGroup, chapter.ExternalUrl, isExternal)
	} else if hasExternalUrlColumn {
		result, err = s.db.Exec(`
			INSERT OR REPLACE INTO manga_chapters 
			(id, manga_id, chapter_number, title, volume, language, pages, source, source_chapter_id, external_url, is_external)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, chapterID, mangaID, chapter.ChapterNumber, chapter.Title, chapter.Volume,
			chapter.Language, chapter.Pages, source, chapter.SourceChapterID, chapter.ExternalUrl, isExternal)
	} else {
		// Legacy schema without external_url or scanlation_group
		result, err = s.db.Exec(`
			INSERT OR REPLACE INTO manga_chapters 
			(id, manga_id, chapter_number, title, volume, language, pages, source, source_chapter_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, chapterID, mangaID, chapter.ChapterNumber, chapter.Title, chapter.Volume,
			chapter.Language, chapter.Pages, source, chapter.SourceChapterID)
	}

	if err != nil {
		return fmt.Errorf("failed to insert chapter: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("      Chapter %s already exists", chapter.ChapterNumber)
	}

	return nil
}

// checkColumnExists checks if a column exists in a table
func (s *SyncService) checkColumnExists(tableName, columnName string) bool {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("Error checking column existence: %v", err)
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, dfltValue, pk interface{}
		err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk)
		if err != nil {
			continue
		}
		if name == columnName {
			return true
		}
	}
	return false
}

// ChapterInfo holds chapter metadata
type ChapterInfo struct {
	ChapterNumber   string
	Title           string
	Volume          string
	Language        string
	Pages           int
	SourceChapterID string
	ScanlationGroup string
	ExternalUrl     *string
}

// SyncResult holds the result of a sync operation
type SyncResult struct {
	TotalFetched int
	Synced       int
	Skipped      int
	Failed       int
	Details      []string
}
