package external

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// MangaPlusClient handles requests to the MangaPlus API
type MangaPlusClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Debug      bool
}

// NewMangaPlusClient creates a new MangaPlus API client
func NewMangaPlusClient() *MangaPlusClient {
	baseURL := os.Getenv("MANGAPLUS_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://jumpg-webapi.tokyo-cdn.com/api"
	}

	timeoutStr := os.Getenv("MANGAPLUS_API_TIMEOUT")
	timeout := 15 * time.Second
	if timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
			timeout = time.Duration(t) * time.Second
		}
	}

	debug := os.Getenv("MANGAPLUS_DEBUG") == "true"

	return &MangaPlusClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Debug: debug,
	}
}

// MangaPlusResponse is the base response wrapper
type MangaPlusResponse struct {
	Success bool            `json:"success"`
	Error   *MangaPlusError `json:"error,omitempty"`
}

// MangaPlusError represents an API error
type MangaPlusError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MangaPlusTitleDetailResponse represents a manga title detail response
type MangaPlusTitleDetailResponse struct {
	MangaPlusResponse
	TitleDetailView *MangaPlusTitleDetailView `json:"titleDetailView,omitempty"`
}

// MangaPlusTitleDetailView contains the manga details
type MangaPlusTitleDetailView struct {
	Title              *MangaPlusTitle    `json:"title"`
	TitleImageUrl      string             `json:"titleImageUrl"`
	Synopsis           string             `json:"synopsis"`
	BackgroundImageUrl string             `json:"backgroundImageUrl"`
	NextTimeStamp      int64              `json:"nextTimeStamp"`
	UpdateTiming       string             `json:"updateTiming"`
	ViewingPeriod      string             `json:"viewingPeriod"`
	NonAppearanceInfo  string             `json:"nonAppearanceInfo"`
	FirstChapterList   []MangaPlusChapter `json:"firstChapterList"`
	LastChapterList    []MangaPlusChapter `json:"lastChapterList"`
	IsSimulReleased    bool               `json:"isSimulReleased"`
	ChaptersDescending bool               `json:"chaptersDescending"`
}

// MangaPlusTitle represents basic title information
type MangaPlusTitle struct {
	TitleID      int    `json:"titleId"`
	Name         string `json:"name"`
	Author       string `json:"author"`
	PortraitUrl  string `json:"portraitImageUrl"`
	LandscapeUrl string `json:"landscapeImageUrl"`
	ViewCount    int64  `json:"viewCount"`
	Language     int    `json:"language"`
}

// MangaPlusChapter represents a chapter entry
type MangaPlusChapter struct {
	ChapterID      int    `json:"chapterId"`
	Name           string `json:"name"`
	SubTitle       string `json:"subTitle"`
	ThumbnailUrl   string `json:"thumbnailUrl"`
	StartTimeStamp int64  `json:"startTimeStamp"`
	EndTimeStamp   int64  `json:"endTimeStamp"`
}

// MangaPlusChapterDetailResponse represents a chapter detail response
type MangaPlusChapterDetailResponse struct {
	MangaPlusResponse
	MangaViewer *MangaPlusMangaViewer `json:"mangaViewer,omitempty"`
}

// MangaPlusMangaViewer contains the chapter pages
type MangaPlusMangaViewer struct {
	TitleID       int             `json:"titleId"`
	ChapterID     int             `json:"chapterId"`
	TitleName     string          `json:"titleName"`
	ChapterName   string          `json:"chapterName"`
	NumberOfPages int             `json:"numberOfPages"`
	Pages         []MangaPlusPage `json:"pages"`
}

// MangaPlusPage represents a single page in the viewer
type MangaPlusPage struct {
	Page *MangaPlusImagePage `json:"mangaPage,omitempty"`
}

// MangaPlusImagePage contains the actual image data
type MangaPlusImagePage struct {
	ImageUrl      string `json:"imageUrl"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	EncryptionKey string `json:"encryptionKey,omitempty"`
}

// GetTitleDetail retrieves manga details by title ID
func (c *MangaPlusClient) GetTitleDetail(titleID int) (*MangaPlusTitleDetailResponse, error) {
	url := fmt.Sprintf("%s/title_detailV3?title_id=%d", c.BaseURL, titleID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add required headers
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("SESSION-TOKEN", "")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result MangaPlusTitleDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		if result.Error != nil {
			return nil, fmt.Errorf("MangaPlus API error: %s (code: %d)", result.Error.Message, result.Error.Code)
		}
		return nil, fmt.Errorf("MangaPlus API request failed")
	}

	return &result, nil
}

// GetChapterDetail retrieves chapter pages by chapter ID
func (c *MangaPlusClient) GetChapterDetail(chapterID int) (*MangaPlusChapterDetailResponse, error) {
	url := fmt.Sprintf("%s/manga_viewer?chapter_id=%d&split=yes&img_quality=super_high", c.BaseURL, chapterID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add required headers
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("SESSION-TOKEN", "")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result MangaPlusChapterDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		if result.Error != nil {
			return nil, fmt.Errorf("MangaPlus API error: %s (code: %d)", result.Error.Message, result.Error.Code)
		}
		return nil, fmt.Errorf("MangaPlus API request failed")
	}

	return &result, nil
}

// GetPageImages extracts image URLs from chapter pages
func (c *MangaPlusClient) GetPageImages(pages []MangaPlusPage) []string {
	var imageUrls []string
	for _, page := range pages {
		if page.Page != nil && page.Page.ImageUrl != "" {
			imageUrls = append(imageUrls, page.Page.ImageUrl)
		}
	}
	return imageUrls
}

// SearchByTitle attempts to find a manga by searching (Note: MangaPlus doesn't have a direct search API)
// This is a placeholder - in reality, you'd need to maintain a mapping or scrape
func (c *MangaPlusClient) SearchByTitle(title string) ([]int, error) {
	// MangaPlus doesn't provide a public search API
	// You would need to either:
	// 1. Maintain a local database of title_id to title mappings
	// 2. Use web scraping (not recommended)
	// 3. Use a third-party API that indexes MangaPlus content

	return nil, fmt.Errorf("MangaPlus does not provide a public search API")
}

// GetAllChapters combines first and last chapter lists
func (v *MangaPlusTitleDetailView) GetAllChapters() []MangaPlusChapter {
	chapters := make([]MangaPlusChapter, 0)

	// Add first chapters
	chapters = append(chapters, v.FirstChapterList...)

	// Add last chapters (avoiding duplicates)
	lastChapterIDs := make(map[int]bool)
	for _, ch := range v.LastChapterList {
		lastChapterIDs[ch.ChapterID] = true
	}

	for _, ch := range v.FirstChapterList {
		delete(lastChapterIDs, ch.ChapterID)
	}

	for _, ch := range v.LastChapterList {
		if lastChapterIDs[ch.ChapterID] {
			chapters = append(chapters, ch)
		}
	}

	return chapters
}
