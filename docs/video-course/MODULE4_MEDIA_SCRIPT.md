# Module 4: Media Detection and Processing - Script

**Duration**: 75 minutes
**Module**: 4 - Media Detection and Processing

---

## Scene 1: Universal Scanner Design (0:00 - 20:00)

**[Visual: Architecture diagram showing UniversalScanner -> ProtocolScanner -> FileSystemClient -> Storage]**

**Narrator**: Welcome to Module 4 -- the heart of Catalogizer. The media detection pipeline transforms raw files on remote storage into structured, searchable media entities. It starts with the Universal Scanner.

**[Visual: Open `catalog-api/internal/services/universal_scanner.go`]**

**Narrator**: The `UniversalScanner` is protocol-agnostic. It manages scan jobs across SMB, FTP, NFS, WebDAV, and local filesystems using a single, unified pipeline. Let us examine its design.

```go
// catalog-api/internal/services/universal_scanner.go
type UniversalScanner struct {
    db                 *database.DB
    logger             *zap.Logger
    renameTracker      *UniversalRenameTracker
    clientFactory      filesystem.ClientFactory
    aggregationService *AggregationService
    scanQueue          chan ScanJob
    workers            int
    maxConcurrentScans int
    scanSem            *semaphore.Weighted
    stopCh             chan struct{}
    wg                 sync.WaitGroup
    protocolScanners   map[string]ProtocolScanner
    activeScans        map[string]*ScanStatus
}
```

**[Visual: Highlight key fields]**

**Narrator**: Notice the concurrency controls. A `semaphore.Weighted` limits concurrent scans. A `scanQueue` channel serializes job submission. A `sync.WaitGroup` tracks active workers for clean shutdown. This design prevents resource exhaustion when scanning large NAS devices.

**[Visual: Show `ScanJob` struct]**

**Narrator**: Each scan job specifies the storage root, path, scan type (full, incremental, or verify), include/exclude patterns, and a maximum depth. This flexibility lets users scan entire libraries or target specific subdirectories.

```go
// catalog-api/internal/services/universal_scanner.go
type ScanJob struct {
    ID              string
    StorageRoot     *models.StorageRoot
    Path            string
    Priority        int
    ScanType        string // full, incremental, verify
    MaxDepth        int
    IncludePatterns []string
    ExcludePatterns []string
    Context         context.Context
}
```

**[Visual: Show `ProtocolScanner` interface]**

**Narrator**: The `ProtocolScanner` interface defines protocol-specific behavior. Each protocol implementation knows its optimal scanning strategy, whether it supports incremental scans, and its ideal batch size for database operations.

```go
// catalog-api/internal/services/universal_scanner.go
type ProtocolScanner interface {
    ScanPath(ctx context.Context, client filesystem.FileSystemClient, job ScanJob, status *ScanStatus) error
    GetScanStrategy() ScanStrategy
    SupportsIncrementalScan() bool
    GetOptimalBatchSize() int
}
```

**[Visual: Show `ScanStatus` struct]**

**Narrator**: Every active scan has a `ScanStatus` that tracks files processed, files found, error count, and the current path. This status is exposed via the API and streamed to the frontend through WebSocket for real-time progress updates.

```go
// catalog-api/internal/services/universal_scanner.go
type ScanStatus struct {
    JobID           string
    StorageRootName string
    Protocol        string
    StartTime       time.Time
    CurrentPath     string
    FilesProcessed  int64
    FilesFound      int64
    FilesUpdated    int64
    FilesDeleted    int64
    ErrorCount      int64
    Status          string // running, completed, failed, cancelled
    mu              sync.RWMutex
}
```

**[Visual: Show the client factory creating protocol-specific clients]**

**Narrator**: The scanner uses the `ClientFactory` (from `filesystem/factory.go`) to create protocol-specific clients. The factory reads the storage root configuration and returns the appropriate client -- SMB, FTP, NFS, WebDAV, or local -- all implementing the same `FileSystemClient` interface.

**[Visual: Show error recovery pattern]**

**Narrator**: Error recovery is built into the scanning loop. Network errors during scanning do not abort the entire job. Instead, the scanner logs the error, increments the error counter, and continues with the next file or directory. This is critical for NAS scanning over unreliable networks.

---

## Scene 2: Media Detection Pipeline (20:00 - 45:00)

**[Visual: Pipeline diagram: Detector -> Analyzer -> Providers]**

**Narrator**: Once files are scanned, the detection pipeline classifies them. This is a three-stage process: detection, analysis, and provider enrichment.

**[Visual: Open `catalog-api/internal/media/detector/engine.go`]**

**Narrator**: The `DetectionEngine` is the first stage. It loads detection rules from configuration and applies them against file paths and metadata. Rules are priority-sorted -- higher priority rules are evaluated first.

```go
// catalog-api/internal/media/detector/engine.go
type DetectionEngine struct {
    logger     *zap.Logger
    rules      []models.DetectionRule
    mediaTypes map[int64]*models.MediaType
}

func (e *DetectionEngine) LoadRules(rules []models.DetectionRule, mediaTypes []models.MediaType) {
    e.rules = rules
    sort.Slice(e.rules, func(i, j int) bool {
        return e.rules[i].Priority > e.rules[j].Priority
    })
    for _, mt := range mediaTypes {
        e.mediaTypes[mt.ID] = &mt
    }
}
```

**[Visual: Show `DetectionResult` struct]**

**Narrator**: Each detection produces a result with a media type, confidence score, the detection method used, matched patterns, a suggested title, year, and quality hints. Multiple rules may match a single directory -- the engine selects the highest-confidence result.

```go
// catalog-api/internal/media/detector/engine.go
type DetectionResult struct {
    MediaTypeID     int64
    MediaType       *models.MediaType
    Confidence      float64
    Method          string
    MatchedPatterns []string
    AnalysisData    *models.AnalysisData
    SuggestedTitle  string
    SuggestedYear   *int
    QualityHints    []string
}
```

**[Visual: Open `catalog-api/internal/media/analyzer/analyzer.go`]**

**Narrator**: The `MediaAnalyzer` is the second stage. It takes detection results and performs deeper analysis -- examining file sizes, codec information, directory structure, and content patterns. The analyzer runs asynchronously with a configurable worker pool.

```go
// catalog-api/internal/media/analyzer/analyzer.go
type MediaAnalyzer struct {
    db              *database.DB
    detector        *detector.DetectionEngine
    providerManager *providers.ProviderManager
    logger          *zap.Logger
    analysisQueue   chan AnalysisRequest
    workers         int
    stopCh          chan struct{}
    wg              sync.WaitGroup
    pendingAnalysis map[string]*AnalysisRequest
}
```

**[Visual: Show `AnalysisResult` struct]**

**Narrator**: Analysis results include the directory analysis, a created or matched `MediaItem`, external metadata from providers, quality analysis (available qualities, duplicates), and updated file records.

```go
// catalog-api/internal/media/analyzer/analyzer.go
type AnalysisResult struct {
    DirectoryAnalysis *mediamodels.DirectoryAnalysis
    MediaItem         *mediamodels.MediaItem
    ExternalMetadata  []mediamodels.ExternalMetadata
    QualityAnalysis   *QualityAnalysis
    UpdatedFiles      []mediamodels.MediaFile
}
```

**[Visual: Open `catalog-api/internal/media/providers/providers.go`]**

**Narrator**: The third stage is provider enrichment. The `ProviderManager` queries external metadata sources -- TMDB, IMDB, and others -- to fill in titles, descriptions, ratings, and cover art.

```go
// catalog-api/internal/media/providers/providers.go
type MetadataProvider interface {
    GetName() string
    Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error)
    GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error)
    IsEnabled() bool
}

type ProviderManager struct {
    providers map[string]MetadataProvider
    logger    *zap.Logger
    client    *http.Client
}
```

**[Visual: Show provider search results]**

**Narrator**: Each provider returns search results ranked by relevance. The manager queries all enabled providers in parallel and merges results. The best match -- based on title similarity and year proximity -- is selected for entity enrichment.

**[Visual: Show recognition providers for specific media types]**

**Narrator**: Catalogizer also has specialized recognition providers in `internal/services/`: `movie_recognition_provider.go`, `music_recognition_provider.go`, `book_recognition_provider.go`, and `game_software_recognition_provider.go`. Each one understands the conventions and metadata patterns for its media type.

**[Visual: Show duplicate detection]**

**Narrator**: The `duplicate_detection_service.go` in `internal/services/` uses file hashes (MD5, SHA256, SHA1, BLAKE3, quick hash) and title matching to identify duplicates. Same title, type, and year across different storage roots is flagged as a potential duplicate.

---

## Scene 3: Entity Aggregation (45:00 - 65:00)

**[Visual: Entity hierarchy diagram: TV Show -> Seasons -> Episodes, Music Artist -> Albums -> Songs]**

**Narrator**: The aggregation service is the bridge between raw scanned files and structured media entities. After a scan completes, it runs as a post-scan hook.

**[Visual: Open `catalog-api/internal/services/aggregation_service.go`]**

**Narrator**: `AggregationService` receives the database, logger, and four repositories: media items, media files, directory analysis, and external metadata.

```go
// catalog-api/internal/services/aggregation_service.go
type AggregationService struct {
    db              *database.DB
    logger          *zap.Logger
    itemRepo        *repository.MediaItemRepository
    fileRepo        *repository.MediaFileRepository
    dirAnalysisRepo *repository.DirectoryAnalysisRepository
    extMetaRepo     *repository.ExternalMetadataRepository
}
```

**[Visual: Show `AggregateAfterScan` method]**

**Narrator**: `AggregateAfterScan` is called after every scan completes. It gets top-level directories from the storage root, processes each one to detect content type and create entities, and reports creation and update counts.

```go
// catalog-api/internal/services/aggregation_service.go
func (s *AggregationService) AggregateAfterScan(ctx context.Context, storageRootID int64) error {
    s.logger.Info("Starting post-scan aggregation", zap.Int64("storage_root_id", storageRootID))

    dirs, err := s.getTopLevelDirectories(ctx, storageRootID)
    if err != nil {
        return fmt.Errorf("get top-level directories: %w", err)
    }

    created, updated := 0, 0
    for _, dir := range dirs {
        isNew, err := s.processDirectory(ctx, dir, storageRootID)
        if err != nil {
            s.logger.Warn("Failed to process directory", zap.String("path", dir.path), zap.Error(err))
            continue
        }
        if isNew { created++ } else { updated++ }
    }
    // ...
}
```

**[Visual: Open `catalog-api/internal/services/title_parser.go`]**

**Narrator**: Title parsing is delegated to the `digital.vasic.entities` submodule. Catalogizer provides thin wrappers that call the parser for each media type: movies, TV shows, music albums, games, and software.

```go
// catalog-api/internal/services/title_parser.go
func ParseMovieTitle(dirname string) ParsedTitle {
    return vasicparser.ParseMovieTitle(dirname)
}

func ParseTVShow(dirname string) ParsedTitle {
    return vasicparser.ParseTVShow(dirname)
}

func ParseMusicAlbum(dirname string) ParsedTitle {
    return vasicparser.ParseMusicAlbum(dirname)
}
```

**[Visual: Show examples of title parsing]**

**Narrator**: The movie parser extracts "The Matrix" and year "1999" from "The.Matrix.1999.1080p.BluRay". The TV parser extracts show name, season number, and episode number from "Breaking.Bad.S01E01.720p". The music parser splits "Pink Floyd - The Wall (1979)" into artist, album, and year.

**[Visual: Show the 11 media types]**

**Narrator**: The system supports 11 media types, seeded in the `media_types` table: movie, tv_show, tv_season, tv_episode, music_artist, music_album, song, game, software, book, and comic.

**[Visual: Show parent_id hierarchy in media_items table]**

**Narrator**: Entity hierarchy uses a self-referencing `parent_id` column. A TV show is the root entity. Seasons have `parent_id` pointing to the show. Episodes point to their season. The same pattern works for music: artist, then album, then song. This recursive structure supports arbitrary depth.

**[Visual: Show media_files junction table]**

**Narrator**: The `media_files` junction table links file records to media entities. A single entity may have multiple files (a movie in multiple formats), and detection rules control which entity a file belongs to.

---

## Scene 4: Thumbnail and Preview Generation (65:00 - 75:00)

**[Visual: Open `catalog-api/internal/services/cover_art_service.go`]**

**Narrator**: The cover art service handles thumbnail and preview image management. It works with the asset system to store, cache, and serve images per media entity.

**[Visual: Show asset resolution pipeline]**

**Narrator**: When a client requests a thumbnail for a media entity, the system checks the local cache first, then looks for embedded cover art in the media files, then queries external providers for poster or album art. Results are cached through the asset management system.

**[Visual: Show `internal/services/asset_resolvers.go`]**

**Narrator**: Asset resolvers are pluggable strategies for locating assets. The resolver chain tries each strategy in order: local file, database cache, external provider. The first resolver that returns a result wins.

**[Visual: Show caching with the CacheService]**

**Narrator**: The `CacheService` in `internal/services/cache_service.go` manages multiple cache tiers: a general key-value cache, a media metadata cache, an API response cache, and a thumbnail cache. Each has its own TTL and eviction policy.

```go
// catalog-api/internal/services/cache_service.go
type CacheEntry struct {
    ID        int64     `json:"id"`
    CacheKey  string    `json:"cache_key"`
    Value     string    `json:"value"`
    ExpiresAt time.Time `json:"expires_at"`
}

type ThumbnailCache struct {
    ID        int64     `json:"id"`
    VideoID   int64     `json:"video_id"`
    Position  int64     `json:"position"`
    // ...
}
```

**[Visual: Course title card]**

**Narrator**: The media detection pipeline is what makes Catalogizer intelligent. From the universal scanner's protocol-agnostic design, through the three-stage detection pipeline, to the aggregation service that builds structured hierarchies -- every scanned file becomes a categorized, searchable entity. In Module 5, we move to the frontend and see how these entities are displayed to users.

---

## Key Code Examples

### Full Scan Pipeline
```
1. ScanJob submitted to UniversalScanner
2. ClientFactory creates protocol-specific FileSystemClient
3. ProtocolScanner walks the filesystem tree
4. Files stored in database (files table)
5. AggregationService.AggregateAfterScan() runs
6. Title parser extracts structured metadata
7. MediaItem created/updated (media_items table)
8. MediaFile junction records created (media_files table)
9. Hierarchy built (parent_id references)
10. External providers queried (TMDB, IMDB)
11. WebSocket notification sent to connected clients
```

### Media Entity Tables
```sql
-- 11 media types (seeded)
CREATE TABLE media_types (id, name, description);

-- Hierarchical media items (self-referencing parent_id)
CREATE TABLE media_items (
    id, media_type_id, parent_id, title, original_title,
    year, rating, description, cover_url, status
);

-- Junction table linking files to entities
CREATE TABLE media_files (
    id, media_item_id, file_id, role, quality, is_primary
);
```

---

## Quiz Questions

1. How does the UniversalScanner handle network errors during a scan without aborting the entire job?
   **Answer**: The scanner catches errors at the per-file level, logs them, increments the `ErrorCount` in the `ScanStatus`, and continues processing the next file or directory. A weighted semaphore controls concurrency to prevent resource exhaustion.

2. What are the three stages of the media detection pipeline?
   **Answer**: (1) The DetectionEngine applies priority-sorted rules to classify files by media type with a confidence score. (2) The MediaAnalyzer performs deeper content analysis and quality assessment. (3) The ProviderManager queries external metadata sources (TMDB, IMDB) to enrich entities with titles, descriptions, ratings, and cover art.

3. How does the entity hierarchy work for TV shows?
   **Answer**: TV shows use the `parent_id` self-reference in the `media_items` table. A TV show entity is the root (parent_id = NULL). Season entities have parent_id pointing to the show. Episode entities have parent_id pointing to their season. This creates a three-level hierarchy: show -> season -> episode.

4. What triggers the `AggregateAfterScan` method, and what does it produce?
   **Answer**: It is triggered as a post-scan hook when a scan job completes. It processes top-level directories from the scanned storage root, uses the title parser to extract structured metadata, creates or updates MediaItem entities in the database, links files to entities via the media_files junction table, and builds parent-child hierarchies.
