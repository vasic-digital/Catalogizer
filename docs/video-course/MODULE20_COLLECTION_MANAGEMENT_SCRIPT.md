# Module 20: Collection Management -- Video Script

**Duration**: 50 minutes
**Prerequisites**: Module 19 (Entity System Deep Dive), Module 5 (Frontend Development)

---

## Video 20.1: Smart Collections with Rules (12 min)

### Opening

Welcome to Module 20. Collections are how users organize their media beyond the automatic hierarchy. While the entity system creates structure from scanned files, collections let users curate, group, and share media according to their own criteria. This module covers smart collections, templates, sharing, real-time sync, analytics, bulk operations, and export/import.

### What Is a Smart Collection?

**[Visual: Browser showing a collection titled "80s Sci-Fi Classics" with automatically matched movies]**

**Narrator**: A smart collection is a saved query with rules. Instead of manually adding items, the user defines criteria -- media type, year range, rating threshold, genre, keyword -- and the system automatically populates the collection. When new media is scanned that matches the rules, it appears in the collection without user intervention.

**[Visual: Open `catalog-api/repository/media_collection_repository.go`]**

**Narrator**: The collection system is backed by two tables: `media_collections` stores the collection metadata and rules, and `media_collection_items` is the junction table linking entities to collections.

```sql
CREATE TABLE media_collections (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    description TEXT,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    is_smart    BOOLEAN DEFAULT 0,
    rules       TEXT,  -- JSON-encoded smart collection rules
    is_public   BOOLEAN DEFAULT 0,
    sort_order  TEXT DEFAULT 'title_asc',
    cover_url   TEXT,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE media_collection_items (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    collection_id INTEGER NOT NULL REFERENCES media_collections(id),
    media_item_id INTEGER NOT NULL REFERENCES media_items(id),
    position      INTEGER DEFAULT 0,
    added_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(collection_id, media_item_id)
);
```

### Rule Engine

**[Visual: Show JSON rule structure for a smart collection]**

**Narrator**: Smart collection rules are stored as JSON. Each rule has a field, operator, and value. Rules can be combined with AND/OR logic.

```json
{
  "operator": "AND",
  "rules": [
    { "field": "media_type", "op": "eq", "value": "movie" },
    { "field": "year", "op": "between", "value": [1980, 1989] },
    { "field": "rating", "op": "gte", "value": 7.0 },
    {
      "operator": "OR",
      "rules": [
        { "field": "genre", "op": "contains", "value": "sci-fi" },
        { "field": "genre", "op": "contains", "value": "science fiction" }
      ]
    }
  ]
}
```

**Narrator**: The rule engine translates this JSON into SQL WHERE clauses at query time. The `between` operator maps to `year >= ? AND year <= ?`. The `contains` operator maps to `LIKE '%?%'`. Nested groups become parenthesized sub-expressions.

```go
// catalog-api/internal/services/collection_service.go
func (s *CollectionService) EvaluateSmartCollection(
    ctx context.Context,
    collectionID int64,
) ([]models.MediaItem, error) {
    collection, err := s.collectionRepo.GetByID(ctx, collectionID)
    if err != nil {
        return nil, fmt.Errorf("get collection: %w", err)
    }

    if !collection.IsSmart || collection.Rules == "" {
        return nil, fmt.Errorf("collection %d is not a smart collection", collectionID)
    }

    var rules RuleGroup
    if err := json.Unmarshal([]byte(collection.Rules), &rules); err != nil {
        return nil, fmt.Errorf("parse rules: %w", err)
    }

    whereClause, args := s.buildWhereClause(rules)
    return s.itemRepo.QueryWithFilter(ctx, whereClause, args)
}
```

**[Visual: Show a smart collection updating after a new scan]**

**Narrator**: After every scan, the aggregation service triggers a re-evaluation of all smart collections. Newly created entities that match the rules are automatically added. Entities that no longer match -- for example, if their rating changes -- are removed.

---

## Video 20.2: Collection Templates (8 min)

### Pre-Built Templates

**[Visual: Collection template gallery showing "Top Rated Movies", "Recent Additions", "TV Series In Progress"]**

**Narrator**: Collection templates are pre-configured smart collection definitions that users can apply with one click. They provide a starting point that users can customize.

**Narrator**: Built-in templates include:

| Template | Rules |
|----------|-------|
| Top Rated Movies | movie, rating >= 8.0 |
| Recent Additions | any type, added in last 30 days |
| TV Series In Progress | tv_show, status = "watching" |
| Unrated Media | any type, rating is null |
| 4K Content | any type, quality contains "2160p" |
| Music by Decade | music_album, year between X and X+9 |

**[Visual: Show template customization UI]**

**Narrator**: Applying a template creates a new collection with the template's rules pre-filled. Users can then modify the name, adjust rules, change the sort order, or add additional criteria. The template itself is never modified.

```go
// catalog-api/internal/services/collection_service.go
func (s *CollectionService) CreateFromTemplate(
    ctx context.Context,
    userID int64,
    templateName string,
    customName string,
) (*models.MediaCollection, error) {
    template, err := s.getTemplate(templateName)
    if err != nil {
        return nil, fmt.Errorf("get template: %w", err)
    }

    collection := &models.MediaCollection{
        Name:    customName,
        UserID:  userID,
        IsSmart: true,
        Rules:   template.Rules,
    }
    return s.collectionRepo.Create(ctx, collection)
}
```

---

## Video 20.3: Sharing and Permissions (8 min)

### Collection Visibility

**[Visual: Collection settings panel showing visibility options]**

**Narrator**: Collections support three visibility levels: private (visible only to the owner), shared (visible to specific users), and public (visible to all authenticated users).

**Narrator**: The `is_public` flag on the collection controls public visibility. For fine-grained sharing, a `collection_shares` table maps collections to specific users with a permission level.

```sql
CREATE TABLE collection_shares (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    collection_id INTEGER NOT NULL REFERENCES media_collections(id),
    user_id       INTEGER NOT NULL REFERENCES users(id),
    permission    TEXT NOT NULL DEFAULT 'read', -- read, write, admin
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(collection_id, user_id)
);
```

**[Visual: Show permission levels]**

**Narrator**: Three permission levels exist. `read` allows viewing the collection and its items. `write` allows adding or removing items from non-smart collections. `admin` allows modifying collection settings, rules, and sharing with additional users.

### Permission Enforcement

**[Visual: Show middleware checking collection access]**

**Narrator**: Permission enforcement happens at the service layer. Every collection operation checks the requesting user's access level before proceeding.

```go
// catalog-api/internal/services/collection_service.go
func (s *CollectionService) checkAccess(
    ctx context.Context,
    collectionID int64,
    userID int64,
    requiredLevel string,
) error {
    collection, err := s.collectionRepo.GetByID(ctx, collectionID)
    if err != nil {
        return fmt.Errorf("get collection: %w", err)
    }

    // Owner has full access
    if collection.UserID == userID {
        return nil
    }

    // Public collections grant read access
    if collection.IsPublic && requiredLevel == "read" {
        return nil
    }

    // Check explicit share
    share, err := s.shareRepo.GetShare(ctx, collectionID, userID)
    if err != nil || !hasPermission(share.Permission, requiredLevel) {
        return fmt.Errorf("access denied: requires %s permission", requiredLevel)
    }
    return nil
}
```

---

## Video 20.4: Real-Time Sync (7 min)

### WebSocket-Driven Updates

**[Visual: Two browser windows side by side, one adding an item to a collection, the other updating in real time]**

**Narrator**: Collection changes propagate in real time through the existing WebSocket infrastructure. When a user adds or removes an item, or when a smart collection re-evaluates after a scan, connected clients receive update events.

**[Visual: Show the event bus publishing a collection update]**

**Narrator**: The event flow is: service method -> event bus -> WebSocket handler -> connected clients. The event payload includes the collection ID, action type (item_added, item_removed, rules_changed), and the affected item IDs.

```go
// catalog-api/internal/media/realtime/ (event bus integration)
type CollectionEvent struct {
    CollectionID int64    `json:"collection_id"`
    Action       string   `json:"action"`
    ItemIDs      []int64  `json:"item_ids,omitempty"`
    Timestamp    int64    `json:"timestamp"`
}
```

**[Visual: Show React Query invalidation on WebSocket message]**

**Narrator**: On the frontend, the WebSocket message triggers a React Query cache invalidation for the affected collection. The UI re-fetches the collection data and updates without a full page reload.

```typescript
// catalog-web/src/hooks/useCollectionSync.ts
useWebSocket('collection_update', (event: CollectionEvent) => {
  queryClient.invalidateQueries({
    queryKey: ['collection', event.collection_id],
  });
});
```

---

## Video 20.5: Collection Analytics and Bulk Operations (8 min)

### Analytics Dashboard

**[Visual: Collection analytics panel showing media type distribution, storage usage, and growth over time]**

**Narrator**: Each collection has an analytics view that summarizes its contents: total items by media type, total storage size, average rating, year distribution, and growth over time.

**Narrator**: Analytics are computed on demand and cached for performance. The cache is invalidated when items are added or removed.

```go
// catalog-api/internal/services/collection_service.go
type CollectionAnalytics struct {
    TotalItems     int                `json:"total_items"`
    TypeBreakdown  map[string]int     `json:"type_breakdown"`
    TotalSizeBytes int64              `json:"total_size_bytes"`
    AverageRating  float64            `json:"average_rating"`
    YearRange      [2]int             `json:"year_range"`
    GrowthHistory  []GrowthDataPoint  `json:"growth_history"`
}

type GrowthDataPoint struct {
    Date  string `json:"date"`
    Count int    `json:"count"`
}
```

### Bulk Operations

**[Visual: Multi-select interface with bulk action toolbar]**

**Narrator**: Bulk operations let users act on multiple items at once. The supported operations are: add to collection, remove from collection, move between collections, change status, and re-run metadata enrichment.

```
POST /api/v1/collections/:id/bulk-add
     Body: { "item_ids": [1, 2, 3, 4, 5] }

POST /api/v1/collections/:id/bulk-remove
     Body: { "item_ids": [3, 5] }

POST /api/v1/collections/bulk-move
     Body: { "from_id": 1, "to_id": 2, "item_ids": [3, 5] }
```

**Narrator**: Bulk operations run in a single database transaction. If any item fails -- for example, it no longer exists -- the entire operation rolls back and returns a detailed error.

---

## Video 20.6: Export and Import (7 min)

### Export Formats

**[Visual: Export dialog with format options: JSON, CSV, M3U (for music)]**

**Narrator**: Collections can be exported for backup, migration, or sharing outside Catalogizer. The export includes collection metadata, rules (for smart collections), and the item list with titles, years, and file paths.

```json
{
  "name": "80s Sci-Fi Classics",
  "description": "Science fiction movies from the 1980s",
  "is_smart": true,
  "rules": { "operator": "AND", "rules": [...] },
  "items": [
    { "title": "Blade Runner", "year": 1982, "media_type": "movie" },
    { "title": "The Terminator", "year": 1984, "media_type": "movie" },
    { "title": "Aliens", "year": 1986, "media_type": "movie" }
  ],
  "exported_at": "2026-03-27T10:00:00Z",
  "version": "1.0"
}
```

### Import and Reconciliation

**[Visual: Import dialog showing matched and unmatched items]**

**Narrator**: Importing a collection performs reconciliation. Each item in the import is matched against existing entities by title, type, and year. Matched items are linked to the new collection. Unmatched items are reported so the user can scan the missing media or manually resolve them.

```go
// catalog-api/internal/services/collection_service.go
type ImportResult struct {
    CollectionID   int64           `json:"collection_id"`
    TotalItems     int             `json:"total_items"`
    MatchedItems   int             `json:"matched_items"`
    UnmatchedItems []UnmatchedItem `json:"unmatched_items"`
}

type UnmatchedItem struct {
    Title     string `json:"title"`
    Year      *int   `json:"year"`
    MediaType string `json:"media_type"`
    Reason    string `json:"reason"`
}
```

**Narrator**: The import endpoint accepts the same JSON format produced by export, ensuring round-trip compatibility.

```
POST /api/v1/collections/import
     Content-Type: application/json
     Body: { ... exported JSON ... }
```

---

## Key Code Examples

### Collection API Endpoints
```
GET    /api/v1/collections                -- List user's collections
POST   /api/v1/collections                -- Create collection
GET    /api/v1/collections/:id            -- Get collection with items
PUT    /api/v1/collections/:id            -- Update collection
DELETE /api/v1/collections/:id            -- Delete collection
POST   /api/v1/collections/:id/items      -- Add items
DELETE /api/v1/collections/:id/items      -- Remove items
POST   /api/v1/collections/:id/bulk-add   -- Bulk add
POST   /api/v1/collections/:id/evaluate   -- Re-evaluate smart rules
GET    /api/v1/collections/:id/analytics  -- Collection analytics
POST   /api/v1/collections/import         -- Import collection
GET    /api/v1/collections/:id/export     -- Export collection
POST   /api/v1/collections/from-template  -- Create from template
POST   /api/v1/collections/:id/share      -- Share with user
```

---

## Key Files Referenced

- `catalog-api/repository/media_collection_repository.go` -- Collection CRUD
- `catalog-api/internal/services/collection_service.go` -- Smart rules, templates, sharing, analytics
- `catalog-api/handlers/media_entity_handler.go` -- Collection REST endpoints
- `catalog-api/internal/media/realtime/` -- WebSocket event broadcasting
- `catalog-web/src/hooks/useCollectionSync.ts` -- Frontend real-time sync

---

## Exercises

1. Define a smart collection rule set that matches all TV episodes from shows with a rating above 8.5, added in the last 90 days.
2. Write a Go function that translates a nested rule group (AND/OR with sub-groups) into a parameterized SQL WHERE clause.
3. Implement a "Duplicate Finder" collection template that automatically groups entities flagged as duplicates.
4. Add CSV export support alongside the existing JSON format, including column headers for title, year, type, and file path.

---

## Quiz Questions

1. What is the difference between a regular collection and a smart collection?
   **Answer**: A regular collection has manually added items. A smart collection has JSON-encoded rules that automatically match entities. Smart collections re-evaluate after scans, adding new matches and removing entities that no longer qualify.

2. How are collection permissions enforced for shared collections?
   **Answer**: The `collection_shares` table maps collections to users with a permission level (read, write, admin). The service layer checks the user's permission before every operation. Owners have implicit full access. Public collections grant read access to all authenticated users.

3. How does real-time collection sync work across multiple clients?
   **Answer**: Collection changes publish events through the event bus to the WebSocket handler. Connected clients receive the event payload (collection ID, action, affected item IDs) and invalidate the relevant React Query cache, triggering a UI refresh without a page reload.
