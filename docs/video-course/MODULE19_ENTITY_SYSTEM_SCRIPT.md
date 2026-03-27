# Module 19: Entity System Deep Dive -- Video Script

**Duration**: 60 minutes
**Prerequisites**: Module 4 (Media Detection and Processing), Module 2 (Backend Development)

---

## Video 19.1: Entity Creation and the 11 Media Types (15 min)

### Opening

Welcome to Module 19, where we take a deep dive into the entity system -- the structured data layer that transforms raw scanned files into browsable, searchable media. In Module 4 we covered the detection pipeline at a high level. Here we go deeper into every aspect of entity creation, hierarchy building, metadata enrichment, duplicate detection, and the title parsing pipeline.

### The 11 Media Types

**[Visual: Table showing all 11 media types from the `media_types` table]**

**Narrator**: Catalogizer supports 11 distinct media types, seeded into the `media_types` table at database initialization. Each type drives detection rules, hierarchy behavior, and metadata provider selection.

```sql
-- catalog-api/database/migrations/ (seed data)
INSERT INTO media_types (name, description) VALUES
    ('movie',        'Feature films and documentaries'),
    ('tv_show',      'Television series'),
    ('tv_season',    'Season of a TV series'),
    ('tv_episode',   'Individual TV episode'),
    ('music_artist', 'Musical artist or band'),
    ('music_album',  'Music album or compilation'),
    ('song',         'Individual music track'),
    ('game',         'Video game'),
    ('software',     'Software application or tool'),
    ('book',         'Book or e-book'),
    ('comic',        'Comic book or graphic novel');
```

**[Visual: Highlight hierarchical types vs. flat types]**

**Narrator**: Six of these types form two hierarchies: TV (`tv_show` -> `tv_season` -> `tv_episode`) and Music (`music_artist` -> `music_album` -> `song`). The remaining five -- movie, game, software, book, and comic -- are flat entities without parent-child relationships.

### Entity Creation Flow

**[Visual: Sequence diagram: Scan completes -> AggregationService -> Title Parser -> MediaItemRepository -> MediaFileRepository]**

**Narrator**: Entity creation is triggered by the `AggregateAfterScan` method. After a scan completes, the aggregation service iterates over top-level directories, runs the title parser to extract structured metadata, then creates or updates entities in the `media_items` table.

```go
// catalog-api/internal/services/aggregation_service.go
func (s *AggregationService) processDirectory(
    ctx context.Context,
    dir directoryInfo,
    storageRootID int64,
) (bool, error) {
    // 1. Parse the directory name to extract title, year, type hints
    parsed := s.parseTitle(dir.path)

    // 2. Detect media type from directory structure and file contents
    mediaType := s.detectMediaType(ctx, dir, parsed)

    // 3. Check for existing entity (same title + type + year)
    existing, err := s.itemRepo.FindByTitleAndType(
        ctx, parsed.Title, mediaType.ID, parsed.Year,
    )
    if err != nil {
        return false, fmt.Errorf("find existing entity: %w", err)
    }

    // 4. Create or update the entity
    if existing == nil {
        return true, s.createEntity(ctx, dir, parsed, mediaType, storageRootID)
    }
    return false, s.updateEntity(ctx, existing, dir, storageRootID)
}
```

**[Visual: Show the `media_items` table schema]**

**Narrator**: The `media_items` table is the central entity store. Each row represents one media entity with a title, media type, optional parent reference, year, rating, description, cover URL, and status field.

```sql
CREATE TABLE media_items (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    media_type_id   INTEGER NOT NULL REFERENCES media_types(id),
    parent_id       INTEGER REFERENCES media_items(id),
    title           TEXT NOT NULL,
    original_title  TEXT,
    year            INTEGER,
    rating          REAL,
    description     TEXT,
    cover_url       TEXT,
    status          TEXT DEFAULT 'active',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## Video 19.2: Hierarchy Building (15 min)

### TV Show Hierarchy: Show -> Season -> Episode

**[Visual: Tree diagram showing a TV show entity with 3 seasons, each with multiple episodes]**

**Narrator**: Television content uses a three-level hierarchy. The show is the root entity with `parent_id` set to NULL. Each season points to the show via `parent_id`. Each episode points to its season. This self-referencing design supports arbitrary nesting without additional tables.

```go
// catalog-api/repository/media_item_repository.go
func (r *MediaItemRepository) CreateHierarchy(
    ctx context.Context,
    showTitle string,
    seasonNum int,
    episodeNum int,
    episodeTitle string,
) (*models.MediaItem, error) {
    // Find or create the show
    show, err := r.FindOrCreate(ctx, showTitle, mediaTypeTV, nil)
    if err != nil {
        return nil, fmt.Errorf("find/create show: %w", err)
    }

    // Find or create the season under the show
    seasonTitle := fmt.Sprintf("Season %d", seasonNum)
    season, err := r.FindOrCreate(ctx, seasonTitle, mediaTypeSeason, &show.ID)
    if err != nil {
        return nil, fmt.Errorf("find/create season: %w", err)
    }

    // Create the episode under the season
    episode, err := r.FindOrCreate(ctx, episodeTitle, mediaTypeEpisode, &season.ID)
    if err != nil {
        return nil, fmt.Errorf("find/create episode: %w", err)
    }

    return episode, nil
}
```

**[Visual: SQL query showing hierarchical retrieval]**

**Narrator**: Retrieving an entire show with all seasons and episodes uses recursive queries. The repository provides `GetChildren` and `GetDescendants` methods. `GetChildren` returns direct children (one level deep), while `GetDescendants` returns the entire subtree.

```go
// catalog-api/repository/media_item_repository.go
func (r *MediaItemRepository) GetChildren(
    ctx context.Context,
    parentID int64,
) ([]models.MediaItem, error) {
    query := `SELECT id, media_type_id, parent_id, title, year,
              rating, description, cover_url, status
              FROM media_items WHERE parent_id = ?
              ORDER BY title`
    rows, err := r.db.QueryContext(ctx, query, parentID)
    // ...
}
```

### Music Hierarchy: Artist -> Album -> Song

**[Visual: Tree diagram showing an artist with albums and songs]**

**Narrator**: Music follows the same pattern. The artist is the root, albums are children of the artist, and songs are children of their album. The title parser extracts the artist name, album title, and track information from directory names and file metadata.

```
/music/Pink Floyd/The Wall (1979)/01 - In The Flesh.flac
       ^^^^^^^^^  ^^^^^^^^ ^^^^  ^^ - ^^^^^^^^^^^^
       Artist     Album    Year  Track  Song Title
```

**Narrator**: When the aggregation service encounters this path, it creates three entities: a `music_artist` for "Pink Floyd", a `music_album` for "The Wall" with year 1979 and `parent_id` pointing to the artist, and a `song` for "In The Flesh" with `parent_id` pointing to the album.

### Junction Table: Linking Files to Entities

**[Visual: Show the `media_files` junction table with arrows to both `media_items` and `files` tables]**

**Narrator**: The `media_files` table is a junction that links file records to entity records. A single entity can have many files -- a movie in multiple resolutions, for example. Each link includes a role (primary, subtitle, trailer), quality tag, and a boolean indicating whether it is the primary file.

```sql
CREATE TABLE media_files (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL REFERENCES media_items(id),
    file_id       INTEGER NOT NULL REFERENCES files(id),
    role          TEXT DEFAULT 'primary',
    quality       TEXT,
    is_primary    BOOLEAN DEFAULT 0,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## Video 19.3: Title Parsing Pipeline (10 min)

### Parser Architecture

**[Visual: Open `catalog-api/internal/services/title_parser.go`]**

**Narrator**: Title parsing is the intelligence behind entity creation. The parser applies media-type-specific regex patterns to directory names, extracting structured fields like title, year, season number, episode number, quality, and codec.

```go
// catalog-api/internal/services/title_parser.go
type ParsedTitle struct {
    Title         string
    OriginalTitle string
    Year          *int
    Season        *int
    Episode       *int
    Quality       string
    Codec         string
    Source        string
    Group         string
    MediaType     string
}
```

**[Visual: Show regex patterns for each media type]**

**Narrator**: The movie parser handles formats like "The.Matrix.1999.1080p.BluRay.x264" and "Inception (2010) [2160p]". The TV parser handles "Breaking.Bad.S01E01.720p" and "Game of Thrones - 3x05 - Kissed by Fire". The music parser handles "Pink Floyd - The Wall (1979)" and "01 - In The Flesh.flac".

### Parser Examples

**[Visual: Terminal showing parser output for various inputs]**

**Narrator**: Let us trace through several examples.

```
Input:  "The.Matrix.1999.1080p.BluRay.x264-GROUP"
Output: Title="The Matrix", Year=1999, Quality="1080p",
        Source="BluRay", Codec="x264", Group="GROUP"

Input:  "Breaking.Bad.S03E07.One.Minute.720p.BluRay"
Output: Title="Breaking Bad", Season=3, Episode=7,
        Quality="720p", Source="BluRay"

Input:  "Pink Floyd - The Dark Side of the Moon (1973)"
Output: Title="The Dark Side of the Moon", Year=1973,
        Artist="Pink Floyd"

Input:  "The Witcher 3 Wild Hunt (2015) [GOG]"
Output: Title="The Witcher 3 Wild Hunt", Year=2015,
        Source="GOG", MediaType="game"
```

**Narrator**: The parser strips dots, underscores, and common separators, normalizing titles for consistent database matching. Quality tags like "1080p", "4K", "BluRay" are extracted but not included in the title.

---

## Video 19.4: Metadata Enrichment (10 min)

### Provider Pipeline

**[Visual: Diagram showing entity -> ProviderManager -> TMDB / OMDB / OpenLibrary / MusicBrainz -> ExternalMetadata table]**

**Narrator**: After an entity is created, metadata enrichment fills in descriptions, ratings, cover art, and additional details from external providers. Each media type has preferred providers.

| Media Type | Primary Provider | Fallback Provider |
|------------|-----------------|-------------------|
| movie      | TMDB            | OMDB              |
| tv_show    | TMDB            | OMDB              |
| music_*    | MusicBrainz     | --                |
| book       | OpenLibrary     | --                |
| game       | IGDB            | GiantBomb         |

**Narrator**: Provider results are stored in the `external_metadata` table, linked to the entity by `media_item_id`. Multiple providers can contribute metadata for the same entity.

```sql
CREATE TABLE external_metadata (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    media_item_id INTEGER NOT NULL REFERENCES media_items(id),
    provider      TEXT NOT NULL,
    external_id   TEXT NOT NULL,
    title         TEXT,
    description   TEXT,
    rating        REAL,
    cover_url     TEXT,
    raw_data      TEXT,
    fetched_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**[Visual: Show graceful degradation when providers are unavailable]**

**Narrator**: Missing API keys or unavailable services do not block the pipeline. Each provider implements the `IsEnabled()` method, and the `ProviderManager` silently skips disabled providers. The entity is created with whatever metadata is available, and enrichment can be retried later.

---

## Video 19.5: Duplicate Detection (5 min)

### Detection Strategy

**[Visual: Open `catalog-api/internal/services/duplicate_detection_service.go`]**

**Narrator**: Duplicate detection identifies the same media across different storage roots or in different formats. The service uses two complementary strategies.

**Narrator**: The first strategy is title-based matching. Two entities with the same title, media type, and year are flagged as potential duplicates. This catches the common case of the same movie stored on two NAS devices.

**Narrator**: The second strategy is hash-based matching. File hashes -- MD5, SHA256, SHA1, BLAKE3, and quick hash -- identify byte-identical files regardless of their file name or location.

```go
// catalog-api/internal/services/duplicate_detection_service.go
type DuplicateGroup struct {
    Title     string
    MediaType string
    Year      *int
    Items     []DuplicateItem
    Strategy  string // "title_match" or "hash_match"
}

type DuplicateItem struct {
    MediaItemID   int64
    StorageRootID int64
    FilePath      string
    FileSize      int64
    Quality       string
    Hash          string
}
```

**Narrator**: Duplicate groups are exposed via the entity API, letting users review and resolve duplicates -- keeping the higher quality version, merging metadata, or marking as intentional copies.

---

## Video 19.6: Entity Browsing UI (5 min)

### Frontend Entity Browser

**[Visual: Browser showing the Entity Browser at `/browse`]**

**Narrator**: The entity browser is the primary way users interact with their media library. It presents entities as a grid of cover art thumbnails with title, year, and type badges.

**[Visual: Navigate from a TV show to its seasons, then to episodes]**

**Narrator**: Clicking a TV show opens its detail page at `/entity/:id`, showing all seasons as cards. Clicking a season shows its episodes. The breadcrumb navigation at the top shows the full hierarchy path: Show > Season 2 > Episode 5.

**[Visual: Show the Entity API endpoints]**

**Narrator**: The frontend communicates with the entity API defined in `handlers/media_entity_handler.go`. Key endpoints include:

```
GET    /api/v1/entities              -- List entities with pagination, filtering
GET    /api/v1/entities/:id          -- Get entity details with children
GET    /api/v1/entities/:id/children -- Get child entities
GET    /api/v1/entities/:id/files    -- Get linked files
GET    /api/v1/entities/search       -- Full-text search
GET    /api/v1/entities/duplicates   -- List duplicate groups
POST   /api/v1/entities/:id/refresh  -- Re-run metadata enrichment
```

**[Visual: Show filtering by media type, year range, and rating]**

**Narrator**: The entity list supports filtering by media type, year range, minimum rating, and status. Pagination uses cursor-based navigation for consistent performance regardless of collection size.

---

## Key Code Examples

### Full Entity Pipeline
```
1. Scan completes -> AggregationService.AggregateAfterScan()
2. Directory name parsed by title_parser.go
3. Media type detected from structure + rules
4. MediaItem created/updated in media_items table
5. Files linked via media_files junction table
6. Hierarchy built with parent_id references
7. Metadata enriched from TMDB/OMDB/MusicBrainz/OpenLibrary
8. Duplicates detected by title + hash matching
9. WebSocket notification sent to connected clients
10. Entity browser UI refreshes in real time
```

### Hierarchy Query Example
```sql
-- Get a TV show with all its children (two levels)
SELECT mi.id, mi.title, mi.media_type_id, mi.parent_id,
       mt.name AS media_type
FROM media_items mi
JOIN media_types mt ON mi.media_type_id = mt.id
WHERE mi.id = ? OR mi.parent_id = ?
   OR mi.parent_id IN (
       SELECT id FROM media_items WHERE parent_id = ?
   )
ORDER BY mi.parent_id NULLS FIRST, mi.title;
```

---

## Key Files Referenced

- `catalog-api/internal/services/aggregation_service.go` -- Post-scan entity creation
- `catalog-api/internal/services/title_parser.go` -- Title parsing with regex patterns
- `catalog-api/internal/services/duplicate_detection_service.go` -- Duplicate detection
- `catalog-api/repository/media_item_repository.go` -- Entity CRUD and hierarchy queries
- `catalog-api/repository/media_file_repository.go` -- File-to-entity linking
- `catalog-api/handlers/media_entity_handler.go` -- Entity REST API endpoints
- `catalog-api/database/migrations/` -- Entity table schemas and seed data

---

## Exercises

1. Write a SQL query that returns all songs by a given artist, traversing the artist -> album -> song hierarchy using the `parent_id` self-reference.
2. Extend the title parser to handle a new format: "Artist - Album [Year] [FLAC]" where the codec tag appears in square brackets after the year.
3. Add a new metadata provider for comics that queries the ComicVine API and stores results in the `external_metadata` table.
4. Write a table-driven test for `FindByTitleAndType` that covers exact matches, case-insensitive matches, and year-range tolerance.

---

## Quiz Questions

1. How many media types does Catalogizer support, and which ones form hierarchies?
   **Answer**: 11 media types. Two hierarchies exist: TV (`tv_show` -> `tv_season` -> `tv_episode`) and Music (`music_artist` -> `music_album` -> `song`). The remaining five (movie, game, software, book, comic) are flat entities.

2. What triggers entity creation, and what is the flow?
   **Answer**: Entity creation is triggered by `AggregateAfterScan()` after a scan completes. The flow is: parse directory name, detect media type, check for existing entity, create or update the entity, link files, build hierarchy, and enrich metadata from external providers.

3. What two strategies does the duplicate detection service use?
   **Answer**: Title-based matching (same title + media type + year) and hash-based matching (MD5, SHA256, SHA1, BLAKE3, quick hash for byte-identical files).

4. How does the system handle unavailable metadata providers?
   **Answer**: Each provider implements `IsEnabled()`. The `ProviderManager` skips disabled providers silently. Entities are created with available metadata, and enrichment can be retried later. Missing API keys or network failures do not block the pipeline.
