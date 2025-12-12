package manga

import (
	"fmt"
	"mangahub/internal/external"
	"mangahub/pkg/models"
	"strconv"
	"strings"
)

// ChapterService handles chapter-related operations
type ChapterService struct {
	mangaDexClient  *external.MangaDexClient
	mangaPlusClient *external.MangaPlusClient
	mangaService    *Service
	jikanClient     *external.JikanClient
	malClient       *external.MALClient
}

// NewChapterService creates a new chapter service
func NewChapterService() *ChapterService {
	return &ChapterService{
		mangaDexClient:  external.NewMangaDexClient(),
		mangaPlusClient: external.NewMangaPlusClient(),
		jikanClient:     external.NewJikanClient(),
		malClient:       external.NewMALClient(),
	}
}

// SetMangaService sets the manga service for looking up manga metadata
func (s *ChapterService) SetMangaService(mangaService *Service) {
	s.mangaService = mangaService
}

// GetChapterList retrieves the chapter list for a manga
func (s *ChapterService) GetChapterList(mangaID string, languages []string, limit, offset int) (*models.ChapterListResponse, error) {
	// Determine the source based on manga ID prefix
	if strings.HasPrefix(mangaID, "mangadex-") || isMangaDexUUID(mangaID) {
		return s.getMangaDexChapters(mangaID, languages, limit, offset)
	} else if strings.HasPrefix(mangaID, "mangaplus-") {
		return s.getMangaPlusChapters(mangaID)
	}

	// Check if this is a MAL ID (format: mal-123 or just a number)
	var mangaTitle string
	if strings.HasPrefix(mangaID, "mal-") {
		// Extract MAL ID and fetch from MAL API
		malIDStr := strings.TrimPrefix(mangaID, "mal-")
		malID, err := strconv.Atoi(malIDStr)
		if err == nil {
			// Try to get manga from MAL/Jikan to get the title
			var malManga *models.Manga

			// Try official MAL API first
			if s.malClient != nil && s.malClient.IsConfigured() {
				malMangaData, err := s.malClient.GetMangaByID(malID)
				if err == nil {
					malManga = external.ConvertMALToManga(malMangaData)
				}
			}

			// Fallback to Jikan if MAL API failed or not configured
			if malManga == nil && s.jikanClient != nil {
				jikanManga, err := s.jikanClient.GetMangaByID(malID)
				if err == nil {
					malManga = external.ConvertJikanToManga(jikanManga)
				}
			}

			if malManga != nil {
				mangaTitle = malManga.Title
			}
		}
	}

	// Try to get manga title from local database if not already fetched
	if mangaTitle == "" && s.mangaService != nil {
		manga, err := s.mangaService.GetManga(mangaID)
		if err == nil && manga != nil {
			mangaTitle = manga.Title
		}
	}

	// If we have a title, try to find chapters on MangaDex
	if mangaTitle != "" {
		// Try multiple search strategies
		mdMangaID := s.searchMangaDexByTitle(mangaTitle)
		if mdMangaID != "" {
			chapters, err := s.getMangaDexChapters(mdMangaID, languages, limit, offset)
			if err == nil && len(chapters.Chapters) > 0 {
				return chapters, nil
			}
		}
	}

	// If MangaDex didn't work, return empty list instead of error
	// This allows the frontend to show "No chapters available" instead of an error
	return &models.ChapterListResponse{
		Chapters: []models.ChapterInfo{},
		Total:    0,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

// searchMangaDexByTitle attempts to find a manga on MangaDex using multiple search strategies
func (s *ChapterService) searchMangaDexByTitle(title string) string {
	// Strategy 1: Exact title search
	searchResults, err := s.mangaDexClient.SearchManga(title, 10)
	if err == nil && searchResults != nil && len(searchResults.Data) > 0 {
		// Try to find exact match first
		for _, manga := range searchResults.Data {
			mdTitle := manga.GetTitle()
			if strings.EqualFold(mdTitle, title) {
				return manga.ID
			}
		}

		// If no exact match, try case-insensitive contains
		for _, manga := range searchResults.Data {
			mdTitle := manga.GetTitle()
			if strings.Contains(strings.ToLower(mdTitle), strings.ToLower(title)) ||
				strings.Contains(strings.ToLower(title), strings.ToLower(mdTitle)) {
				return manga.ID
			}
		}

		// If still no match, return the first result (most relevant by MangaDex ranking)
		return searchResults.Data[0].ID
	}

	// Strategy 2: Try removing common suffixes/prefixes
	cleanTitle := strings.TrimSpace(title)
	cleanTitle = strings.TrimSuffix(cleanTitle, " (TV)")
	cleanTitle = strings.TrimSuffix(cleanTitle, " (Dub)")
	cleanTitle = strings.TrimSuffix(cleanTitle, " (Sub)")

	if cleanTitle != title {
		searchResults, err = s.mangaDexClient.SearchManga(cleanTitle, 5)
		if err == nil && searchResults != nil && len(searchResults.Data) > 0 {
			return searchResults.Data[0].ID
		}
	}

	return ""
}

// GetChapterPages retrieves the pages for a specific chapter
func (s *ChapterService) GetChapterPages(chapterID, source string) (*models.ChapterPages, error) {
	if source == "mangadex" || source == "" {
		return s.getMangaDexPages(chapterID)
	} else if source == "mangaplus" {
		return s.getMangaPlusPages(chapterID)
	}

	return nil, fmt.Errorf("unsupported source: %s", source)
}

// getMangaDexChapters retrieves chapters from MangaDex
func (s *ChapterService) getMangaDexChapters(mangaID string, languages []string, limit, offset int) (*models.ChapterListResponse, error) {
	// Extract MangaDex ID
	mdID := strings.TrimPrefix(mangaID, "mangadex-")

	// Set defaults
	if limit <= 0 {
		limit = 100
	}
	if len(languages) == 0 {
		languages = []string{"en"}
	}

	// Get chapters from MangaDex
	feedResp, err := s.mangaDexClient.GetMangaChapterFeed(mdID, limit, offset, languages)
	if err != nil {
		return nil, fmt.Errorf("failed to get MangaDex chapters: %w", err)
	}

	// Convert to our model
	chapters := make([]models.ChapterInfo, 0, len(feedResp.Data))
	for _, mdChapter := range feedResp.Data {
		chapterInfo := models.ChapterInfo{
			ID:            mdChapter.ID,
			MangaID:       mangaID,
			ChapterNumber: mdChapter.GetChapterNumber(),
			VolumeNumber:  mdChapter.GetVolumeNumber(),
			Title:         mdChapter.Attributes.Title,
			Language:      mdChapter.Attributes.TranslatedLanguage,
			Pages:         mdChapter.Attributes.Pages,
			PublishedAt:   mdChapter.Attributes.PublishAt.Format("2006-01-02"),
			Source:        "mangadex",
		}

		// Check if this is a licensed chapter with external URL
		if mdChapter.Attributes.ExternalUrl != nil && *mdChapter.Attributes.ExternalUrl != "" {
			chapterInfo.ExternalUrl = mdChapter.Attributes.ExternalUrl
			chapterInfo.IsExternal = true
			// If it's a MangaPlus URL, mark the source
			if strings.Contains(*mdChapter.Attributes.ExternalUrl, "mangaplus.shueisha.co.jp") {
				chapterInfo.Source = "mangaplus"
			}
		}

		chapters = append(chapters, chapterInfo)
	}

	return &models.ChapterListResponse{
		Chapters: chapters,
		Total:    feedResp.Total,
		Limit:    feedResp.Limit,
		Offset:   feedResp.Offset,
	}, nil
}

// getMangaPlusChapters retrieves chapters from MangaPlus
func (s *ChapterService) getMangaPlusChapters(mangaID string) (*models.ChapterListResponse, error) {
	// Extract MangaPlus title ID
	titleIDStr := strings.TrimPrefix(mangaID, "mangaplus-")
	titleID, err := strconv.Atoi(titleIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid MangaPlus title ID: %w", err)
	}

	// Get title details from MangaPlus
	titleResp, err := s.mangaPlusClient.GetTitleDetail(titleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get MangaPlus title: %w", err)
	}

	if titleResp.TitleDetailView == nil {
		return nil, fmt.Errorf("no title detail view available")
	}

	// Get all chapters
	mpChapters := titleResp.TitleDetailView.GetAllChapters()

	// Convert to our model
	chapters := make([]models.ChapterInfo, 0, len(mpChapters))
	for _, mpChapter := range mpChapters {
		// Extract chapter number from name (e.g., "#123")
		chapterNum := strings.TrimPrefix(mpChapter.Name, "#")

		chapterInfo := models.ChapterInfo{
			ID:            fmt.Sprintf("%d", mpChapter.ChapterID),
			MangaID:       mangaID,
			ChapterNumber: chapterNum,
			Title:         mpChapter.SubTitle,
			Language:      "en", // MangaPlus is primarily English
			Pages:         0,    // Not available in list view
			Source:        "mangaplus",
		}
		chapters = append(chapters, chapterInfo)
	}

	return &models.ChapterListResponse{
		Chapters: chapters,
		Total:    len(chapters),
		Limit:    len(chapters),
		Offset:   0,
	}, nil
}

// getMangaDexPages retrieves chapter pages from MangaDex
func (s *ChapterService) getMangaDexPages(chapterID string) (*models.ChapterPages, error) {
	// Get at-home server response
	atHome, err := s.mangaDexClient.GetChapterPages(chapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get MangaDex pages: %w", err)
	}

	// Build page URLs
	pages := make([]string, 0, len(atHome.Chapter.Data))
	for _, filename := range atHome.Chapter.Data {
		pageURL := external.BuildMangaDexPageURL(atHome.BaseUrl, atHome.Chapter.Hash, filename, false)
		pages = append(pages, pageURL)
	}

	return &models.ChapterPages{
		ChapterID: chapterID,
		Pages:     pages,
		Source:    "mangadex",
		BaseURL:   atHome.BaseUrl,
		Hash:      atHome.Chapter.Hash,
	}, nil
}

// getMangaPlusPages retrieves chapter pages from MangaPlus
func (s *ChapterService) getMangaPlusPages(chapterID string) (*models.ChapterPages, error) {
	// Parse chapter ID
	chapterIDInt, err := strconv.Atoi(chapterID)
	if err != nil {
		return nil, fmt.Errorf("invalid MangaPlus chapter ID: %w", err)
	}

	// Get chapter detail
	chapterResp, err := s.mangaPlusClient.GetChapterDetail(chapterIDInt)
	if err != nil {
		return nil, fmt.Errorf("failed to get MangaPlus chapter: %w", err)
	}

	if chapterResp.MangaViewer == nil {
		return nil, fmt.Errorf("no manga viewer available")
	}

	// Extract page URLs
	pages := s.mangaPlusClient.GetPageImages(chapterResp.MangaViewer.Pages)

	return &models.ChapterPages{
		ChapterID:  chapterID,
		MangaID:    fmt.Sprintf("mangaplus-%d", chapterResp.MangaViewer.TitleID),
		ChapterNum: chapterResp.MangaViewer.ChapterName,
		Pages:      pages,
		Source:     "mangaplus",
	}, nil
}

// isMangaDexUUID checks if a string is a valid MangaDex UUID
func isMangaDexUUID(s string) bool {
	return len(s) == 36 && strings.Count(s, "-") == 4
}
