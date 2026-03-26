package providers

import (
	"catalogizer/internal/media/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// MetadataProvider interface for external metadata sources
type MetadataProvider interface {
	GetName() string
	Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error)
	GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error)
	IsEnabled() bool
}

// SearchResult represents search results from providers
type SearchResult struct {
	ExternalID  string   `json:"external_id"`
	Title       string   `json:"title"`
	Year        *int     `json:"year,omitempty"`
	Rating      *float64 `json:"rating,omitempty"`
	Description *string  `json:"description,omitempty"`
	CoverURL    *string  `json:"cover_url,omitempty"`
	Relevance   float64  `json:"relevance"`
}

// ProviderManager manages all metadata providers
type ProviderManager struct {
	providers map[string]MetadataProvider
	logger    *zap.Logger
	client    *http.Client
}

// NewProviderManager creates a new provider manager
func NewProviderManager(logger *zap.Logger) *ProviderManager {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	pm := &ProviderManager{
		providers: make(map[string]MetadataProvider),
		logger:    logger,
		client:    client,
	}

	// Initialize all providers
	pm.registerProviders()

	return pm
}

// registerProviders registers all available providers
func (pm *ProviderManager) registerProviders() {
	// Movie/TV providers
	pm.providers["tmdb"] = NewTMDBProvider(pm.client, pm.logger)
	pm.providers["imdb"] = NewIMDBProvider(pm.client, pm.logger)
	pm.providers["tvdb"] = NewTVDBProvider(pm.client, pm.logger)

	// Music providers
	pm.providers["musicbrainz"] = NewMusicBrainzProvider(pm.client, pm.logger)
	pm.providers["spotify"] = NewSpotifyProvider(pm.client, pm.logger)
	pm.providers["lastfm"] = NewLastFMProvider(pm.client, pm.logger)

	// Gaming providers
	pm.providers["igdb"] = NewIGDBProvider(pm.client, pm.logger)
	pm.providers["steam"] = NewSteamProvider(pm.client, pm.logger)

	// Books providers
	pm.providers["goodreads"] = NewGoodreadsProvider(pm.client, pm.logger)
	pm.providers["openlibrary"] = NewOpenLibraryProvider(pm.client, pm.logger)

	// Anime providers
	pm.providers["anidb"] = NewAniDBProvider(pm.client, pm.logger)
	pm.providers["myanimelist"] = NewMyAnimeListProvider(pm.client, pm.logger)

	// YouTube/Social providers
	pm.providers["youtube"] = NewYouTubeProvider(pm.client, pm.logger)

	// Software providers
	pm.providers["github"] = NewGitHubProvider(pm.client, pm.logger)

	pm.logger.Info("Metadata providers registered", zap.Int("count", len(pm.providers)))
}

// RegisterProvider registers or replaces a named metadata provider.
// This is primarily useful for testing with mock providers.
func (pm *ProviderManager) RegisterProvider(name string, provider MetadataProvider) {
	pm.providers[name] = provider
}

// SearchAll searches across all relevant providers
func (pm *ProviderManager) SearchAll(ctx context.Context, query string, mediaType string, year *int, providers []string) (map[string][]SearchResult, error) {
	results := make(map[string][]SearchResult)

	// If no specific providers requested, use all relevant ones
	if len(providers) == 0 {
		providers = pm.getProvidersForMediaType(mediaType)
	}

	for _, providerName := range providers {
		provider, exists := pm.providers[providerName]
		if !exists || !provider.IsEnabled() {
			continue
		}

		providerResults, err := provider.Search(ctx, query, mediaType, year)
		if err != nil {
			pm.logger.Error("Provider search failed",
				zap.String("provider", providerName),
				zap.String("query", query),
				zap.Error(err))
			continue
		}

		if len(providerResults) > 0 {
			results[providerName] = providerResults
		}
	}

	return results, nil
}

// GetBestMatch finds the best matching result across all providers
func (pm *ProviderManager) GetBestMatch(ctx context.Context, query string, mediaType string, year *int) (*SearchResult, string, error) {
	allResults, err := pm.SearchAll(ctx, query, mediaType, year, nil)
	if err != nil {
		return nil, "", err
	}

	var bestResult *SearchResult
	var bestProvider string
	var bestScore float64

	for providerName, results := range allResults {
		for _, result := range results {
			score := pm.calculateRelevanceScore(result, query, year)
			if score > bestScore {
				bestScore = score
				bestResult = &result
				bestProvider = providerName
			}
		}
	}

	return bestResult, bestProvider, nil
}

// GetDetails gets detailed metadata from a specific provider
func (pm *ProviderManager) GetDetails(ctx context.Context, providerName, externalID string) (*models.ExternalMetadata, error) {
	provider, exists := pm.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", providerName)
	}

	if !provider.IsEnabled() {
		return nil, fmt.Errorf("provider disabled: %s", providerName)
	}

	return provider.GetDetails(ctx, externalID)
}

// calculateRelevanceScore calculates how relevant a search result is
func (pm *ProviderManager) calculateRelevanceScore(result SearchResult, query string, year *int) float64 {
	score := result.Relevance

	// Bonus for exact title match
	if strings.EqualFold(result.Title, query) {
		score += 0.3
	} else if strings.Contains(strings.ToLower(result.Title), strings.ToLower(query)) {
		score += 0.2
	}

	// Bonus for year match
	if year != nil && result.Year != nil && *result.Year == *year {
		score += 0.2
	}

	// Bonus for having rating
	if result.Rating != nil && *result.Rating > 0 {
		score += 0.1
	}

	return score
}

// getProvidersForMediaType returns relevant providers for a media type
func (pm *ProviderManager) getProvidersForMediaType(mediaType string) []string {
	providerMap := map[string][]string{
		"movie":         {"tmdb", "imdb"},
		"tv_show":       {"tmdb", "imdb", "tvdb"},
		"anime":         {"anidb", "myanimelist", "tmdb"},
		"music":         {"musicbrainz", "spotify", "lastfm"},
		"pc_game":       {"igdb", "steam"},
		"console_game":  {"igdb"},
		"mobile_game":   {"igdb"},
		"ebook":         {"goodreads", "openlibrary"},
		"audiobook":     {"goodreads"},
		"documentary":   {"tmdb", "imdb"},
		"youtube_video": {"youtube"},
		"software":      {"github"},
		"comic":         {"comicvine", "openlibrary"},
		"podcast":       {"spotify"},
	}

	if providers, exists := providerMap[mediaType]; exists {
		return providers
	}

	// Default providers for unknown types
	return []string{"tmdb", "imdb"}
}

// Base provider struct with common functionality
type BaseProvider struct {
	name    string
	client  *http.Client
	logger  *zap.Logger
	enabled bool
	apiKey  string
	baseURL string
}

func NewBaseProvider(name, baseURL, apiKey string, client *http.Client, logger *zap.Logger) *BaseProvider {
	return &BaseProvider{
		name:    name,
		client:  client,
		logger:  logger,
		enabled: apiKey != "",
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

func (bp *BaseProvider) GetName() string {
	return bp.name
}

func (bp *BaseProvider) IsEnabled() bool {
	return bp.enabled
}

// GetBaseURL returns the provider's base URL. Useful for testing with mock servers.
func (bp *BaseProvider) GetBaseURL() string {
	return bp.baseURL
}

// MakeRequest is the exported version of makeRequest for use by external test providers.
func (bp *BaseProvider) MakeRequest(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	return bp.makeRequest(ctx, url, headers)
}

// makeRequest makes an HTTP request with error handling
func (bp *BaseProvider) makeRequest(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add API key if available
	if bp.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+bp.apiKey)
	}

	resp, err := bp.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

// TMDB Provider Implementation
type TMDBProvider struct {
	*BaseProvider
}

func NewTMDBProvider(client *http.Client, logger *zap.Logger) *TMDBProvider {
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		if logger != nil {
			logger.Warn("TMDB_API_KEY not set, TMDB provider disabled")
		}
	}
	return &TMDBProvider{
		BaseProvider: NewBaseProvider("tmdb", "https://api.themoviedb.org/3", apiKey, client, logger),
	}
}

func (t *TMDBProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	endpoint := "/search/multi"
	if mediaType == "movie" {
		endpoint = "/search/movie"
	} else if mediaType == "tv_show" {
		endpoint = "/search/tv"
	}

	params := url.Values{}
	params.Add("api_key", t.apiKey)
	params.Add("query", query)
	if year != nil {
		params.Add("year", fmt.Sprintf("%d", *year))
	}

	requestURL := t.baseURL + endpoint + "?" + params.Encode()

	body, err := t.makeRequest(ctx, requestURL, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Results []struct {
			ID           int     `json:"id"`
			Title        string  `json:"title"`
			Name         string  `json:"name"`
			ReleaseDate  string  `json:"release_date"`
			FirstAirDate string  `json:"first_air_date"`
			Overview     string  `json:"overview"`
			PosterPath   string  `json:"poster_path"`
			VoteAverage  float64 `json:"vote_average"`
			MediaType    string  `json:"media_type"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := make([]SearchResult, 0, len(response.Results))
	for _, item := range response.Results {
		title := item.Title
		if title == "" {
			title = item.Name
		}

		var year *int
		dateStr := item.ReleaseDate
		if dateStr == "" {
			dateStr = item.FirstAirDate
		}
		if len(dateStr) >= 4 {
			if y := parseInt(dateStr[:4]); y > 1900 {
				year = &y
			}
		}

		var coverURL *string
		if item.PosterPath != "" {
			url := fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", item.PosterPath)
			coverURL = &url
		}

		var rating *float64
		if item.VoteAverage > 0 {
			rating = &item.VoteAverage
		}

		result := SearchResult{
			ExternalID:  fmt.Sprintf("%d", item.ID),
			Title:       title,
			Year:        year,
			Rating:      rating,
			Description: &item.Overview,
			CoverURL:    coverURL,
			Relevance:   0.8, // Base relevance for TMDB
		}

		results = append(results, result)
	}

	return results, nil
}

func (t *TMDBProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	// Implementation for getting detailed TMDB data
	endpoint := fmt.Sprintf("/movie/%s", externalID)
	params := url.Values{}
	params.Add("api_key", t.apiKey)
	params.Add("append_to_response", "credits,videos,external_ids")

	requestURL := t.baseURL + endpoint + "?" + params.Encode()

	body, err := t.makeRequest(ctx, requestURL, nil)
	if err != nil {
		return nil, err
	}

	metadata := &models.ExternalMetadata{
		Provider:    t.name,
		ExternalID:  externalID,
		Data:        string(body),
		LastFetched: time.Now(),
	}

	// Parse specific fields
	var details struct {
		Title       string  `json:"title"`
		VoteAverage float64 `json:"vote_average"`
		PosterPath  string  `json:"poster_path"`
		Homepage    string  `json:"homepage"`
		Videos      struct {
			Results []struct {
				Site string `json:"site"`
				Key  string `json:"key"`
				Type string `json:"type"`
			} `json:"results"`
		} `json:"videos"`
	}

	if err := json.Unmarshal(body, &details); err == nil {
		if details.VoteAverage > 0 {
			metadata.Rating = &details.VoteAverage
		}
		if details.PosterPath != "" {
			url := fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", details.PosterPath)
			metadata.CoverURL = &url
		}
		if details.Homepage != "" {
			metadata.ReviewURL = &details.Homepage
		}

		// Find trailer
		for _, video := range details.Videos.Results {
			if video.Site == "YouTube" && video.Type == "Trailer" {
				trailerURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.Key)
				metadata.TrailerURL = &trailerURL
				break
			}
		}
	}

	return metadata, nil
}

// Helper function (duplicate from detector package)
func parseInt(s string) int {
	var result int
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		}
	}
	return result
}

// Placeholder implementations for other providers
// These would be implemented similarly to TMDBProvider

type IMDBProvider struct{ *BaseProvider }
type TVDBProvider struct{ *BaseProvider }
type MusicBrainzProvider struct{ *BaseProvider }
type SpotifyProvider struct{ *BaseProvider }
type LastFMProvider struct{ *BaseProvider }
type IGDBProvider struct{ *BaseProvider }
type SteamProvider struct{ *BaseProvider }
type GoodreadsProvider struct{ *BaseProvider }
type OpenLibraryProvider struct{ *BaseProvider }
type AniDBProvider struct{ *BaseProvider }
type MyAnimeListProvider struct{ *BaseProvider }
type YouTubeProvider struct{ *BaseProvider }
type GitHubProvider struct{ *BaseProvider }

// Placeholder constructor functions
func NewIMDBProvider(client *http.Client, logger *zap.Logger) *IMDBProvider {
	return &IMDBProvider{NewBaseProvider("imdb", "https://imdb-api.com", "", client, logger)}
}

func NewTVDBProvider(client *http.Client, logger *zap.Logger) *TVDBProvider {
	return &TVDBProvider{NewBaseProvider("tvdb", "https://api.thetvdb.com", "", client, logger)}
}

func NewMusicBrainzProvider(client *http.Client, logger *zap.Logger) *MusicBrainzProvider {
	bp := NewBaseProvider("musicbrainz", "https://musicbrainz.org/ws/2", "", client, logger)
	bp.enabled = true // MusicBrainz is free, no API key required
	return &MusicBrainzProvider{bp}
}

func NewSpotifyProvider(client *http.Client, logger *zap.Logger) *SpotifyProvider {
	return &SpotifyProvider{NewBaseProvider("spotify", "https://api.spotify.com/v1", "", client, logger)}
}

func NewLastFMProvider(client *http.Client, logger *zap.Logger) *LastFMProvider {
	return &LastFMProvider{NewBaseProvider("lastfm", "https://ws.audioscrobbler.com/2.0", "", client, logger)}
}

func NewIGDBProvider(client *http.Client, logger *zap.Logger) *IGDBProvider {
	return &IGDBProvider{NewBaseProvider("igdb", "https://api.igdb.com/v4", "", client, logger)}
}

func NewSteamProvider(client *http.Client, logger *zap.Logger) *SteamProvider {
	return &SteamProvider{NewBaseProvider("steam", "https://store.steampowered.com/api", "", client, logger)}
}

func NewGoodreadsProvider(client *http.Client, logger *zap.Logger) *GoodreadsProvider {
	return &GoodreadsProvider{NewBaseProvider("goodreads", "https://www.goodreads.com/book", "", client, logger)}
}

func NewOpenLibraryProvider(client *http.Client, logger *zap.Logger) *OpenLibraryProvider {
	bp := NewBaseProvider("openlibrary", "https://openlibrary.org", "", client, logger)
	bp.enabled = true // OpenLibrary is free, no API key required
	return &OpenLibraryProvider{bp}
}

func NewAniDBProvider(client *http.Client, logger *zap.Logger) *AniDBProvider {
	return &AniDBProvider{NewBaseProvider("anidb", "https://anidb.net/perl-bin/animedb.pl", "", client, logger)}
}

func NewMyAnimeListProvider(client *http.Client, logger *zap.Logger) *MyAnimeListProvider {
	return &MyAnimeListProvider{NewBaseProvider("myanimelist", "https://api.myanimelist.net/v2", "", client, logger)}
}

func NewYouTubeProvider(client *http.Client, logger *zap.Logger) *YouTubeProvider {
	return &YouTubeProvider{NewBaseProvider("youtube", "https://www.googleapis.com/youtube/v3", "", client, logger)}
}

func NewGitHubProvider(client *http.Client, logger *zap.Logger) *GitHubProvider {
	return &GitHubProvider{NewBaseProvider("github", "https://api.github.com", "", client, logger)}
}

// Implement Search and GetDetails methods for each provider
// (Similar pattern to TMDBProvider, customized for each API)

func (p *IMDBProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("query", query))
		return nil, nil
	}
	// API integration placeholder - implement when API key is available
	p.logger.Info("provider search not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

func (p *IMDBProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("externalID", externalID))
		return nil, nil
	}
	// API integration placeholder - implement when API key is available
	p.logger.Info("provider get details not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

// TVDBProvider GetDetails implementation
func (p *TVDBProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("externalID", externalID))
		return nil, nil
	}
	p.logger.Info("provider get details not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

// MusicBrainzProvider GetDetails implementation
func (p *MusicBrainzProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("externalID", externalID))
		return nil, nil
	}

	params := url.Values{}
	params.Add("inc", "artists+releases")
	params.Add("fmt", "json")

	requestURL := p.baseURL + "/recording/" + externalID + "?" + params.Encode()

	headers := map[string]string{
		"User-Agent": "Catalogizer/1.0 (https://github.com/vasic-digital/Catalogizer)",
		"Accept":     "application/json",
	}

	body, err := p.makeRequest(ctx, requestURL, headers)
	if err != nil {
		return nil, err
	}

	metadata := &models.ExternalMetadata{
		Provider:    p.name,
		ExternalID:  externalID,
		Data:        string(body),
		LastFetched: time.Now(),
	}

	// Parse for structured fields
	var details struct {
		Title        string `json:"title"`
		ArtistCredit []struct {
			Name string `json:"name"`
		} `json:"artist-credit"`
		Releases []struct {
			Title string `json:"title"`
			Date  string `json:"date"`
		} `json:"releases"`
	}

	if err := json.Unmarshal(body, &details); err == nil {
		// MusicBrainz doesn't provide cover art directly; CoverURL remains nil
		// Rating is not available from MusicBrainz recording endpoint
		// ReviewURL could link to the MusicBrainz page
		mbURL := fmt.Sprintf("https://musicbrainz.org/recording/%s", externalID)
		metadata.ReviewURL = &mbURL
	}

	return metadata, nil
}

// SpotifyProvider GetDetails implementation
func (p *SpotifyProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("externalID", externalID))
		return nil, nil
	}
	p.logger.Info("provider get details not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

// LastFMProvider GetDetails implementation
func (p *LastFMProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("externalID", externalID))
		return nil, nil
	}
	p.logger.Info("provider get details not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

// IGDBProvider GetDetails implementation
func (p *IGDBProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("externalID", externalID))
		return nil, nil
	}
	p.logger.Info("provider get details not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

// SteamProvider GetDetails implementation
func (p *SteamProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("externalID", externalID))
		return nil, nil
	}
	p.logger.Info("provider get details not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

// GoodreadsProvider GetDetails implementation
func (p *GoodreadsProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("externalID", externalID))
		return nil, nil
	}
	p.logger.Info("provider get details not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

// OpenLibraryProvider GetDetails implementation
func (p *OpenLibraryProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("externalID", externalID))
		return nil, nil
	}

	// externalID is the OpenLibrary key (e.g., "/works/OL12345W")
	requestURL := p.baseURL + externalID + ".json"

	body, err := p.makeRequest(ctx, requestURL, nil)
	if err != nil {
		return nil, err
	}

	metadata := &models.ExternalMetadata{
		Provider:    p.name,
		ExternalID:  externalID,
		Data:        string(body),
		LastFetched: time.Now(),
	}

	// Parse specific fields — description can be a string or {type, value} object
	var details struct {
		Title       string          `json:"title"`
		Description json.RawMessage `json:"description"`
		Covers      []int           `json:"covers"`
	}

	if err := json.Unmarshal(body, &details); err == nil {
		// Parse description: can be a plain string or an object with "value" field
		if len(details.Description) > 0 {
			var descStr string
			if err := json.Unmarshal(details.Description, &descStr); err != nil {
				// Try object form: {"type": "...", "value": "..."}
				var descObj struct {
					Value string `json:"value"`
				}
				if err := json.Unmarshal(details.Description, &descObj); err == nil {
					descStr = descObj.Value
				}
			}
			// Description is stored in Data field (raw JSON), no separate field in ExternalMetadata
			_ = descStr
		}

		if len(details.Covers) > 0 {
			coverURL := fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", details.Covers[0])
			metadata.CoverURL = &coverURL
		}
	}

	return metadata, nil
}

// AniDBProvider GetDetails implementation
func (p *AniDBProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("externalID", externalID))
		return nil, nil
	}
	p.logger.Info("provider get details not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

// MyAnimeListProvider GetDetails implementation
func (p *MyAnimeListProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("externalID", externalID))
		return nil, nil
	}
	p.logger.Info("provider get details not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

// YouTubeProvider GetDetails implementation
func (p *YouTubeProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("externalID", externalID))
		return nil, nil
	}
	p.logger.Info("provider get details not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

// GitHubProvider GetDetails implementation
func (p *GitHubProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("externalID", externalID))
		return nil, nil
	}
	p.logger.Info("provider get details not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

// Search method implementations for all providers
func (p *TVDBProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("query", query))
		return nil, nil
	}
	p.logger.Info("provider search not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

func (p *MusicBrainzProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("query", query))
		return nil, nil
	}

	params := url.Values{}
	params.Add("query", query)
	params.Add("fmt", "json")
	params.Add("limit", "10")

	requestURL := p.baseURL + "/recording?" + params.Encode()

	headers := map[string]string{
		"User-Agent": "Catalogizer/1.0 (https://github.com/vasic-digital/Catalogizer)",
		"Accept":     "application/json",
	}

	body, err := p.makeRequest(ctx, requestURL, headers)
	if err != nil {
		return nil, err
	}

	var response struct {
		Recordings []struct {
			ID           string `json:"id"`
			Title        string `json:"title"`
			ArtistCredit []struct {
				Name string `json:"name"`
			} `json:"artist-credit"`
			Releases []struct {
				Title string `json:"title"`
				Date  string `json:"date"`
			} `json:"releases"`
		} `json:"recordings"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse MusicBrainz response: %w", err)
	}

	results := make([]SearchResult, 0, len(response.Recordings))
	for _, rec := range response.Recordings {
		// Build artist string
		var artists []string
		for _, ac := range rec.ArtistCredit {
			artists = append(artists, ac.Name)
		}

		// Extract year from first release date
		var year *int
		if len(rec.Releases) > 0 && len(rec.Releases[0].Date) >= 4 {
			y := parseInt(rec.Releases[0].Date[:4])
			if y > 1900 {
				year = &y
			}
		}

		// Build description from artist + release info
		var desc string
		if len(artists) > 0 {
			desc = "by " + strings.Join(artists, ", ")
		}
		if len(rec.Releases) > 0 {
			desc += " (from " + rec.Releases[0].Title + ")"
		}

		result := SearchResult{
			ExternalID: rec.ID,
			Title:      rec.Title,
			Year:       year,
			Relevance:  0.7,
		}

		if desc != "" {
			result.Description = &desc
		}

		results = append(results, result)
	}

	return results, nil
}

func (p *SpotifyProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("query", query))
		return nil, nil
	}
	p.logger.Info("provider search not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

func (p *LastFMProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("query", query))
		return nil, nil
	}
	p.logger.Info("provider search not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

func (p *IGDBProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("query", query))
		return nil, nil
	}
	p.logger.Info("provider search not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

func (p *SteamProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("query", query))
		return nil, nil
	}
	p.logger.Info("provider search not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

func (p *GoodreadsProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("query", query))
		return nil, nil
	}
	p.logger.Info("provider search not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

func (p *OpenLibraryProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("query", query))
		return nil, nil
	}

	params := url.Values{}
	params.Add("q", query)
	params.Add("limit", "10")

	requestURL := p.baseURL + "/search.json?" + params.Encode()

	body, err := p.makeRequest(ctx, requestURL, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Docs []struct {
			Key              string   `json:"key"`
			Title            string   `json:"title"`
			FirstPublishYear *int     `json:"first_publish_year"`
			AuthorName       []string `json:"author_name"`
			CoverI           *int     `json:"cover_i"`
		} `json:"docs"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse OpenLibrary response: %w", err)
	}

	results := make([]SearchResult, 0, len(response.Docs))
	for _, doc := range response.Docs {
		title := doc.Title
		if len(doc.AuthorName) > 0 {
			desc := "by " + strings.Join(doc.AuthorName, ", ")
			result := SearchResult{
				ExternalID:  doc.Key,
				Title:       title,
				Year:        doc.FirstPublishYear,
				Description: &desc,
				Relevance:   0.7,
			}

			if doc.CoverI != nil {
				coverURL := fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", *doc.CoverI)
				result.CoverURL = &coverURL
			}

			results = append(results, result)
		} else {
			result := SearchResult{
				ExternalID: doc.Key,
				Title:      title,
				Year:       doc.FirstPublishYear,
				Relevance:  0.7,
			}

			if doc.CoverI != nil {
				coverURL := fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", *doc.CoverI)
				result.CoverURL = &coverURL
			}

			results = append(results, result)
		}
	}

	return results, nil
}

func (p *AniDBProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("query", query))
		return nil, nil
	}
	p.logger.Info("provider search not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

func (p *MyAnimeListProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("query", query))
		return nil, nil
	}
	p.logger.Info("provider search not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

func (p *YouTubeProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("query", query))
		return nil, nil
	}
	p.logger.Info("provider search not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}

func (p *GitHubProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !p.IsEnabled() {
		p.logger.Debug("provider not enabled, skipping", zap.String("provider", p.GetName()), zap.String("query", query))
		return nil, nil
	}
	p.logger.Info("provider search not yet implemented", zap.String("provider", p.GetName()))
	return nil, nil
}
