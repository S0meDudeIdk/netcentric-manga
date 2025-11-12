package external

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// MALClient handles requests to the official MyAnimeList API
type MALClient struct {
	ClientID     string
	ClientSecret string
	BaseURL      string
	HTTPClient   *http.Client
}

// NewMALClient creates a new official MAL API client
func NewMALClient() *MALClient {
	clientID := os.Getenv("MAL_CLIENT_ID")
	clientSecret := os.Getenv("MAL_CLIENT_SECRET")

	return &MALClient{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		BaseURL:      "https://api.myanimelist.net/v2",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsConfigured checks if MAL client is properly configured
func (c *MALClient) IsConfigured() bool {
	return c.ClientID != "" && c.ClientID != "your_mal_client_id_here"
}

// MALMangaNode represents a manga entry from MAL API
type MALMangaNode struct {
	ID                int             `json:"id"`
	Title             string          `json:"title"`
	MainPicture       MALPicture      `json:"main_picture"`
	AlternativeTitles MALAltTitles    `json:"alternative_titles"`
	StartDate         string          `json:"start_date"`
	Synopsis          string          `json:"synopsis"`
	Mean              float64         `json:"mean"`
	Rank              int             `json:"rank"`
	Popularity        int             `json:"popularity"`
	NumListUsers      int             `json:"num_list_users"`
	NumScoringUsers   int             `json:"num_scoring_users"`
	Status            string          `json:"status"`
	Genres            []MALGenre      `json:"genres"`
	MediaType         string          `json:"media_type"`
	NumChapters       int             `json:"num_chapters"`
	NumVolumes        int             `json:"num_volumes"`
	Authors           []MALAuthorNode `json:"authors"`
}

// MALPicture represents manga cover images
type MALPicture struct {
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

// MALAltTitles represents alternative titles
type MALAltTitles struct {
	Synonyms []string `json:"synonyms"`
	En       string   `json:"en"`
	Ja       string   `json:"ja"`
}

// MALGenre represents a genre
type MALGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// MALAuthorNode represents author information
type MALAuthorNode struct {
	Node MALAuthor `json:"node"`
	Role string    `json:"role"`
}

// MALAuthor represents author details
type MALAuthor struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// MALMangaListResponse represents the response from manga list endpoint
type MALMangaListResponse struct {
	Data []struct {
		Node MALMangaNode `json:"node"`
	} `json:"data"`
	Paging MALPaging `json:"paging"`
}

// MALPaging represents pagination info
type MALPaging struct {
	Previous string `json:"previous"`
	Next     string `json:"next"`
}

// SearchManga searches for manga using the official MAL API
func (c *MALClient) SearchManga(query string, limit int) (*MALMangaListResponse, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("MAL API not configured: please set MAL_CLIENT_ID in .env")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	params := url.Values{}
	params.Add("q", query)
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("fields", "id,title,main_picture,alternative_titles,start_date,synopsis,mean,rank,popularity,num_list_users,num_scoring_users,status,genres,media_type,num_chapters,num_volumes,authors")

	fullURL := fmt.Sprintf("%s/manga?%s", c.BaseURL, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MAL-Client-ID", c.ClientID)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MAL API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result MALMangaListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetMangaByID gets manga details by ID using the official MAL API
func (c *MALClient) GetMangaByID(id int) (*MALMangaNode, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("MAL API not configured: please set MAL_CLIENT_ID in .env")
	}

	params := url.Values{}
	params.Add("fields", "id,title,main_picture,alternative_titles,start_date,synopsis,mean,rank,popularity,num_list_users,num_scoring_users,status,genres,media_type,num_chapters,num_volumes,authors")

	fullURL := fmt.Sprintf("%s/manga/%d?%s", c.BaseURL, id, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MAL-Client-ID", c.ClientID)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("manga not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MAL API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result MALMangaNode
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetMangaRanking gets manga ranking (top manga)
func (c *MALClient) GetMangaRanking(rankingType string, limit int) (*MALMangaListResponse, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("MAL API not configured: please set MAL_CLIENT_ID in .env")
	}

	if limit <= 0 || limit > 500 {
		limit = 20
	}

	if rankingType == "" {
		rankingType = "all" // Options: all, manga, novels, oneshots, doujin, manhwa, manhua, bypopularity, favorite
	}

	params := url.Values{}
	params.Add("ranking_type", rankingType)
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("fields", "id,title,main_picture,alternative_titles,start_date,synopsis,mean,rank,popularity,num_list_users,num_scoring_users,status,genres,media_type,num_chapters,num_volumes,authors")

	fullURL := fmt.Sprintf("%s/manga/ranking?%s", c.BaseURL, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MAL-Client-ID", c.ClientID)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MAL API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result MALMangaListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
