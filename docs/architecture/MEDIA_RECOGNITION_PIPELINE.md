# Media Recognition Pipeline Architecture

## Overview

The media recognition pipeline is responsible for automatically identifying media files (movies, TV shows, music, podcasts) and enriching them with metadata from external providers like TMDB, IMDB, and OMDB. This document describes the complete architecture, data flow, and integration patterns.

## Table of Contents

1. [Pipeline Architecture](#pipeline-architecture)
2. [Detection Flow](#detection-flow)
3. [Analyzer Components](#analyzer-components)
4. [External Provider Integration](#external-provider-integration)
5. [Caching Strategy](#caching-strategy)
6. [WebSocket Updates](#websocket-updates)
7. [Error Handling & Retry](#error-handling--retry)
8. [Performance Optimization](#performance-optimization)

## Pipeline Architecture

### High-Level Overview

```
┌─────────────────┐
│   File Scanner  │
│  (Filesystem)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Detector     │  ◄── Pattern matching, extension filtering
│  (file_detector)│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Analyzer     │  ◄── Extract metadata, parse filenames
│ (media_analyzer)│
└────────┬────────┘
         │
         ├──────────────────┬──────────────────┬───────────────┐
         ▼                  ▼                  ▼               ▼
    ┌────────┐        ┌────────┐        ┌────────┐      ┌────────┐
    │  TMDB  │        │  IMDB  │        │  OMDB  │      │  Other │
    │Provider│        │Provider│        │Provider│      │Provider│
    └───┬────┘        └───┬────┘        └───┬────┘      └───┬────┘
        └────────────────┬┴──────────────────┴──────────────┘
                         ▼
                   ┌──────────┐
                   │   Cache  │
                   │  (Redis) │
                   └─────┬────┘
                         ▼
                   ┌──────────┐
                   │ Database │
                   │(Metadata)│
                   └─────┬────┘
                         ▼
                   ┌──────────┐
                   │ WebSocket│
                   │  Update  │
                   └──────────┘
```

### Component Locations

| Component | Location | Purpose |
|-----------|----------|---------|
| File Detector | `catalog-api/internal/media/detector/` | Identify media files by pattern |
| Media Analyzer | `catalog-api/internal/media/analyzer/` | Extract metadata, parse filenames |
| TMDB Provider | `catalog-api/internal/media/providers/tmdb/` | TMDB API integration |
| IMDB Provider | `catalog-api/internal/media/providers/imdb/` | IMDB scraping/API |
| OMDB Provider | `catalog-api/internal/media/providers/omdb/` | OMDB API integration |
| Event Bus | `catalog-api/internal/media/realtime/event_bus.go` | WebSocket event distribution |
| Cache Layer | `catalog-api/internal/cache/` | Redis caching |

## Detection Flow

### Step 1: File Detection

**Detector**: `detector/file_detector.go`

Scans filesystem and identifies potential media files based on:

1. **File Extensions**:
   ```go
   var videoExtensions = map[string]bool{
       ".mp4":  true,
       ".mkv":  true,
       ".avi":  true,
       ".mov":  true,
       ".wmv":  true,
       ".flv":  true,
       ".webm": true,
       ".m4v":  true,
   }

   var audioExtensions = map[string]bool{
       ".mp3":  true,
       ".flac": true,
       ".wav":  true,
       ".aac":  true,
       ".m4a":  true,
       ".ogg":  true,
       ".wma":  true,
   }
   ```

2. **File Size** (configurable minimum):
   ```go
   const minVideoFileSize = 10 * 1024 * 1024  // 10 MB
   const minAudioFileSize = 1 * 1024 * 1024   // 1 MB
   ```

3. **Path Patterns**:
   - Exclude system directories: `.git/`, `node_modules/`, `.cache/`
   - Exclude hidden files: files starting with `.`
   - Prioritize media directories: `Movies/`, `TV Shows/`, `Music/`

### Step 2: Type Classification

**Classifier**: `detector/type_classifier.go`

Determines media type based on naming patterns:

```go
type MediaType int

const (
    MediaTypeUnknown MediaType = iota
    MediaTypeMovie              // Single file movies
    MediaTypeTVShow             // Episodic content
    MediaTypeMusic              // Audio albums/songs
    MediaTypePodcast            // Podcast episodes
)

func ClassifyMedia(path string) MediaType {
    filename := filepath.Base(path)

    // Check for TV show patterns
    if isTVShowPattern(filename) {
        return MediaTypeTVShow
    }

    // Check for movie patterns
    if isMoviePattern(filename) {
        return MediaTypeMovie
    }

    // Check for music patterns (albums, artists)
    if isMusicPattern(path) {
        return MediaTypeMusic
    }

    return MediaTypeUnknown
}

func isTVShowPattern(filename string) bool {
    patterns := []string{
        `S\d{2}E\d{2}`,      // S01E01
        `\d{1}x\d{2}`,       // 1x01
        `Episode\s*\d+`,     // Episode 1
        `[Ee]p\s*\d+`,       // Ep 1
    }

    for _, pattern := range patterns {
        matched, _ := regexp.MatchString(pattern, filename)
        if matched {
            return true
        }
    }
    return false
}
```

## Analyzer Components

### Filename Parser

**Parser**: `analyzer/filename_parser.go`

Extracts structured data from filenames using regex patterns:

#### Movie Parsing

```go
type MovieInfo struct {
    Title    string
    Year     int
    Quality  string  // 1080p, 720p, 4K, etc.
    Source   string  // BluRay, WEB-DL, HDTV, etc.
    Codec    string  // x264, x265, HEVC, etc.
    Release  string  // Release group
}

func ParseMovieFilename(filename string) *MovieInfo {
    // Example: "The.Matrix.1999.1080p.BluRay.x264-GROUP.mkv"

    // Remove extension
    name := strings.TrimSuffix(filename, filepath.Ext(filename))

    info := &MovieInfo{}

    // Extract year (1900-2099)
    yearRegex := regexp.MustCompile(`(19|20)\d{2}`)
    if match := yearRegex.FindString(name); match != "" {
        info.Year, _ = strconv.Atoi(match)
        // Remove year from title
        name = strings.Replace(name, match, "", 1)
    }

    // Extract quality
    qualityRegex := regexp.MustCompile(`(?i)(2160p|1080p|720p|480p|4K|UHD)`)
    if match := qualityRegex.FindString(name); match != "" {
        info.Quality = strings.ToLower(match)
    }

    // Extract source
    sourceRegex := regexp.MustCompile(`(?i)(BluRay|BRRip|WEB-DL|WEBRip|HDTV|DVDRip)`)
    if match := sourceRegex.FindString(name); match != "" {
        info.Source = match
    }

    // Extract codec
    codecRegex := regexp.MustCompile(`(?i)(x264|x265|H\.264|H\.265|HEVC|XviD)`)
    if match := codecRegex.FindString(name); match != "" {
        info.Codec = match
    }

    // Extract release group
    releaseRegex := regexp.MustCompile(`-([A-Za-z0-9]+)$`)
    if match := releaseRegex.FindStringSubmatch(name); len(match) > 1 {
        info.Release = match[1]
    }

    // Clean title
    info.Title = cleanTitle(name)

    return info
}

func cleanTitle(title string) string {
    // Remove common patterns
    patterns := []string{
        `(?i)(2160p|1080p|720p|480p|4K)`,
        `(?i)(BluRay|BRRip|WEB-DL|WEBRip)`,
        `(?i)(x264|x265|H\.264|HEVC)`,
        `-[A-Za-z0-9]+$`, // Release group
    }

    for _, pattern := range patterns {
        re := regexp.MustCompile(pattern)
        title = re.ReplaceAllString(title, "")
    }

    // Replace dots/underscores with spaces
    title = strings.ReplaceAll(title, ".", " ")
    title = strings.ReplaceAll(title, "_", " ")

    // Trim and normalize spaces
    title = strings.TrimSpace(title)
    title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")

    return title
}
```

#### TV Show Parsing

```go
type TVShowInfo struct {
    Title   string
    Season  int
    Episode int
    Year    int
    Quality string
}

func ParseTVShowFilename(filename string) *TVShowInfo {
    // Examples:
    // "Game.of.Thrones.S08E06.1080p.WEB.H264-MEMENTO.mkv"
    // "Breaking.Bad.1x01.Pilot.720p.BluRay.x264.mkv"

    info := &TVShowInfo{}

    // S##E## pattern
    sePattern := regexp.MustCompile(`[Ss](\d{1,2})[Ee](\d{1,2})`)
    if match := sePattern.FindStringSubmatch(filename); len(match) > 2 {
        info.Season, _ = strconv.Atoi(match[1])
        info.Episode, _ = strconv.Atoi(match[2])
    }

    // #x## pattern
    xPattern := regexp.MustCompile(`(\d{1,2})x(\d{1,2})`)
    if match := xPattern.FindStringSubmatch(filename); len(match) > 2 {
        info.Season, _ = strconv.Atoi(match[1])
        info.Episode, _ = strconv.Atoi(match[2])
    }

    // Extract title (everything before season/episode marker)
    titleEnd := sePattern.FindStringIndex(filename)
    if titleEnd == nil {
        titleEnd = xPattern.FindStringIndex(filename)
    }

    if titleEnd != nil {
        title := filename[:titleEnd[0]]
        info.Title = cleanTitle(title)
    }

    // Extract quality
    qualityRegex := regexp.MustCompile(`(?i)(2160p|1080p|720p|480p|4K)`)
    if match := qualityRegex.FindString(filename); match != "" {
        info.Quality = strings.ToLower(match)
    }

    return info
}
```

## External Provider Integration

### Provider Interface

```go
type MetadataProvider interface {
    // Search for media by title and year
    Search(ctx context.Context, query SearchQuery) ([]SearchResult, error)

    // Get detailed metadata by external ID
    GetDetails(ctx context.Context, externalID string) (*MediaDetails, error)

    // Get provider name
    GetProviderName() string
}

type SearchQuery struct {
    Title      string
    Year       int
    MediaType  string // "movie", "tv", "music"
}

type SearchResult struct {
    ExternalID  string
    Title       string
    Year        int
    Overview    string
    PosterURL   string
    BackdropURL string
    Rating      float64
}

type MediaDetails struct {
    ExternalID   string
    Title        string
    OriginalTitle string
    Year         int
    Overview     string
    Genres       []string
    Cast         []Person
    Crew         []Person
    Runtime      int // minutes
    Rating       float64
    PosterURL    string
    BackdropURL  string
    TrailerURL   string
    ReleaseDate  string
    Metadata     map[string]interface{}
}
```

### TMDB Provider

**Implementation**: `providers/tmdb/client.go`

The Movie Database (TMDB) is the primary metadata source:

```go
type TMDBClient struct {
    apiKey     string
    baseURL    string
    httpClient *http.Client
    cache      *cache.Cache
}

func (c *TMDBClient) Search(ctx context.Context, query SearchQuery) ([]SearchResult, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("tmdb:search:%s:%d:%s", query.Title, query.Year, query.MediaType)
    if cached, found := c.cache.Get(cacheKey); found {
        return cached.([]SearchResult), nil
    }

    // Build API URL
    url := fmt.Sprintf("%s/search/%s", c.baseURL, query.MediaType)
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    q := req.URL.Query()
    q.Add("api_key", c.apiKey)
    q.Add("query", query.Title)
    if query.Year > 0 {
        q.Add("year", strconv.Itoa(query.Year))
    }
    req.URL.RawQuery = q.Encode()

    // Execute request
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("TMDB API request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
    }

    // Parse response
    var apiResp struct {
        Results []struct {
            ID           int     `json:"id"`
            Title        string  `json:"title"`
            Name         string  `json:"name"` // For TV shows
            ReleaseDate  string  `json:"release_date"`
            FirstAirDate string  `json:"first_air_date"` // For TV shows
            Overview     string  `json:"overview"`
            PosterPath   string  `json:"poster_path"`
            BackdropPath string  `json:"backdrop_path"`
            VoteAverage  float64 `json:"vote_average"`
        } `json:"results"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, err
    }

    // Convert to SearchResult
    results := make([]SearchResult, 0, len(apiResp.Results))
    for _, r := range apiResp.Results {
        result := SearchResult{
            ExternalID: strconv.Itoa(r.ID),
            Overview:   r.Overview,
            Rating:     r.VoteAverage,
        }

        // Handle movie vs TV show differences
        if query.MediaType == "movie" {
            result.Title = r.Title
            if r.ReleaseDate != "" {
                if t, err := time.Parse("2006-01-02", r.ReleaseDate); err == nil {
                    result.Year = t.Year()
                }
            }
        } else {
            result.Title = r.Name
            if r.FirstAirDate != "" {
                if t, err := time.Parse("2006-01-02", r.FirstAirDate); err == nil {
                    result.Year = t.Year()
                }
            }
        }

        // Build image URLs
        if r.PosterPath != "" {
            result.PosterURL = "https://image.tmdb.org/t/p/w500" + r.PosterPath
        }
        if r.BackdropPath != "" {
            result.BackdropURL = "https://image.tmdb.org/t/p/w1280" + r.BackdropPath
        }

        results = append(results, result)
    }

    // Cache results for 24 hours
    c.cache.Set(cacheKey, results, 24*time.Hour)

    return results, nil
}

func (c *TMDBClient) GetDetails(ctx context.Context, externalID string) (*MediaDetails, error) {
    // Check cache
    cacheKey := fmt.Sprintf("tmdb:details:%s", externalID)
    if cached, found := c.cache.Get(cacheKey); found {
        return cached.(*MediaDetails), nil
    }

    // Fetch from API...
    // (Similar implementation with caching)

    return details, nil
}
```

### OMDB Provider

**Implementation**: `providers/omdb/client.go`

Open Movie Database for additional metadata:

```go
type OMDBClient struct {
    apiKey     string
    baseURL    string
    httpClient *http.Client
}

func (c *OMDBClient) Search(ctx context.Context, query SearchQuery) ([]SearchResult, error) {
    url := fmt.Sprintf("%s/?apikey=%s&s=%s&type=%s",
        c.baseURL,
        c.apiKey,
        url.QueryEscape(query.Title),
        query.MediaType,
    )

    if query.Year > 0 {
        url += fmt.Sprintf("&y=%d", query.Year)
    }

    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Parse and return results...
}
```

### Provider Fallback Strategy

```go
type ProviderChain struct {
    providers []MetadataProvider
    logger    *zap.Logger
}

func (pc *ProviderChain) Search(ctx context.Context, query SearchQuery) ([]SearchResult, error) {
    var lastErr error

    for _, provider := range pc.providers {
        results, err := provider.Search(ctx, query)
        if err != nil {
            pc.logger.Warn("Provider search failed",
                zap.String("provider", provider.GetProviderName()),
                zap.Error(err),
            )
            lastErr = err
            continue
        }

        if len(results) > 0 {
            pc.logger.Info("Provider search succeeded",
                zap.String("provider", provider.GetProviderName()),
                zap.Int("results", len(results)),
            )
            return results, nil
        }
    }

    return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
}
```

## Caching Strategy

### Multi-Level Cache

1. **In-Memory Cache** (Go map with TTL):
   - Hot data (< 1 minute old)
   - Prevents duplicate API calls in rapid succession
   - Per-goroutine safety with sync.RWMutex

2. **Redis Cache**:
   - Warm data (1 hour - 24 hours)
   - Shared across all API instances
   - Invalidation on manual metadata refresh

3. **Database Cache**:
   - Cold data (permanent)
   - Stored in `external_metadata` table
   - Only refreshed on explicit user request

### Cache Keys

```go
// Search results
"tmdb:search:{title}:{year}:{media_type}" -> []SearchResult

// Media details
"tmdb:details:{external_id}" -> MediaDetails
"omdb:details:{imdb_id}" -> MediaDetails

// Images (URLs cached, not binary data)
"tmdb:images:{id}:{type}" -> []ImageURL
```

### Cache Invalidation

```go
func (s *Service) RefreshMetadata(ctx context.Context, mediaID int64) error {
    // Get media item
    media, err := s.repo.GetMediaItem(ctx, mediaID)
    if err != nil {
        return err
    }

    // Invalidate all caches
    cacheKeys := []string{
        fmt.Sprintf("tmdb:search:%s:%d:movie", media.Title, media.Year),
        fmt.Sprintf("tmdb:details:%s", media.ExternalID),
    }

    for _, key := range cacheKeys {
        s.cache.Delete(key)
    }

    // Re-fetch from providers
    return s.enrichMetadata(ctx, media)
}
```

## WebSocket Updates

### Real-Time Event Flow

When metadata is enriched, clients receive real-time updates:

```go
type MetadataEvent struct {
    Type      string      `json:"type"`      // "metadata_updated"
    MediaID   int64       `json:"media_id"`
    Timestamp time.Time   `json:"timestamp"`
    Data      interface{} `json:"data"`
}

func (s *Service) enrichMetadata(ctx context.Context, media *Media) error {
    // Fetch metadata from providers
    metadata, err := s.providers.Search(ctx, SearchQuery{
        Title:     media.Title,
        Year:      media.Year,
        MediaType: media.MediaType,
    })
    if err != nil {
        return err
    }

    // Update database
    if err := s.repo.UpdateMetadata(ctx, media.ID, metadata); err != nil {
        return err
    }

    // Broadcast WebSocket event
    event := MetadataEvent{
        Type:      "metadata_updated",
        MediaID:   media.ID,
        Timestamp: time.Now(),
        Data:      metadata,
    }

    s.eventBus.Publish("media:metadata", event)

    return nil
}
```

### Event Bus Implementation

**Location**: `internal/media/realtime/event_bus.go`

```go
type EventBus struct {
    subscribers map[string][]chan interface{}
    mu          sync.RWMutex
}

func (eb *EventBus) Subscribe(topic string) <-chan interface{} {
    eb.mu.Lock()
    defer eb.mu.Unlock()

    ch := make(chan interface{}, 100) // Buffered
    eb.subscribers[topic] = append(eb.subscribers[topic], ch)
    return ch
}

func (eb *EventBus) Publish(topic string, event interface{}) {
    eb.mu.RLock()
    defer eb.mu.RUnlock()

    for _, ch := range eb.subscribers[topic] {
        select {
        case ch <- event:
        default:
            // Channel full, skip (prevents blocking)
        }
    }
}
```

## Error Handling & Retry

### Retry with Exponential Backoff

```go
func (c *TMDBClient) searchWithRetry(ctx context.Context, query SearchQuery) ([]SearchResult, error) {
    maxRetries := 3
    baseDelay := 1 * time.Second

    for attempt := 0; attempt < maxRetries; attempt++ {
        results, err := c.search(ctx, query)
        if err == nil {
            return results, nil
        }

        // Check if error is retryable
        if !isRetryableError(err) {
            return nil, err
        }

        // Calculate delay with exponential backoff
        delay := baseDelay * time.Duration(1<<uint(attempt))

        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(delay):
            // Retry
        }
    }

    return nil, errors.New("max retries exceeded")
}

func isRetryableError(err error) bool {
    // Retry on network errors and 5xx status codes
    if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
        return true
    }

    if httpErr, ok := err.(*HTTPError); ok {
        return httpErr.StatusCode >= 500
    }

    return false
}
```

### Rate Limiting

```go
type RateLimiter struct {
    requests chan struct{}
    interval time.Duration
}

func NewRateLimiter(requestsPerSecond int) *RateLimiter {
    rl := &RateLimiter{
        requests: make(chan struct{}, requestsPerSecond),
        interval: time.Second / time.Duration(requestsPerSecond),
    }

    // Refill token bucket
    go func() {
        ticker := time.NewTicker(rl.interval)
        defer ticker.Stop()

        for range ticker.C {
            select {
            case rl.requests <- struct{}{}:
            default:
            }
        }
    }()

    return rl
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-rl.requests:
        return nil
    }
}

// Usage
rateLimiter := NewRateLimiter(4) // 4 requests per second (TMDB limit)

func (c *TMDBClient) search(ctx context.Context, query SearchQuery) ([]SearchResult, error) {
    if err := c.rateLimiter.Wait(ctx); err != nil {
        return nil, err
    }

    // Make API request...
}
```

## Performance Optimization

### Batch Processing

```go
func (s *Service) EnrichMediaBatch(ctx context.Context, mediaIDs []int64) error {
    // Process in batches of 100
    batchSize := 100
    sem := make(chan struct{}, 10) // Max 10 concurrent

    for i := 0; i < len(mediaIDs); i += batchSize {
        end := i + batchSize
        if end > len(mediaIDs) {
            end = len(mediaIDs)
        }

        batch := mediaIDs[i:end]

        sem <- struct{}{}
        go func(ids []int64) {
            defer func() { <-sem }()

            for _, id := range ids {
                if err := s.enrichMetadata(ctx, id); err != nil {
                    s.logger.Error("Failed to enrich metadata",
                        zap.Int64("media_id", id),
                        zap.Error(err),
                    )
                }
            }
        }(batch)
    }

    // Wait for all batches
    for i := 0; i < cap(sem); i++ {
        sem <- struct{}{}
    }

    return nil
}
```

### Lazy Loading

Only fetch metadata when actually needed:

```go
func (h *Handler) GetMediaDetails(c *gin.Context) {
    mediaID := c.Param("id")

    // Get basic media info from database
    media, err := h.service.GetMedia(c.Request.Context(), mediaID)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // Check if metadata exists
    if media.ExternalMetadata == nil || len(media.ExternalMetadata) == 0 {
        // Fetch metadata in background
        go func() {
            ctx := context.Background()
            h.service.EnrichMetadata(ctx, media.ID)
        }()
    }

    c.JSON(200, media)
}
```

## Summary

The media recognition pipeline:

1. **Detects** media files using pattern matching and file extensions
2. **Parses** filenames to extract title, year, season/episode, quality
3. **Searches** external providers (TMDB, OMDB) for metadata
4. **Enriches** media items with detailed information (cast, crew, ratings)
5. **Caches** metadata at multiple levels for performance
6. **Broadcasts** updates via WebSocket for real-time UI updates
7. **Handles** errors gracefully with retry and fallback strategies
8. **Optimizes** performance with batching, rate limiting, and lazy loading

This architecture ensures accurate media identification, rich metadata, and responsive user experience.
