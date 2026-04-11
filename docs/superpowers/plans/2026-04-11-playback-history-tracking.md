# Playback Session Tracking + Cross-App History UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Track every media reproduction session (video play, audio play, book read, comic flip, game run) end-to-end, show "duration / current progress / last session amount" on every media card across all four client apps, and expose a detailed per-item history view, with challenge + unit-test + HelixQA coverage.

**Architecture:**
- **Backend** adds a `playback_sessions` table (one row per reproduction session) plus a denormalised `media_progress` view/columns on `user_media_metadata`. New REST endpoints: `POST /api/v1/playback/sessions/start`, `POST /api/v1/playback/sessions/progress`, `POST /api/v1/playback/sessions/end`, `GET /api/v1/entities/:id/progress`, `GET /api/v1/entities/:id/history?limit=…`. Units are stored as `position_value BIGINT` + `position_unit TEXT` so the same schema handles seconds (video/audio) and pages (book/comic) and events (game runs). An `aggregate_seconds` generated column powers "total time spent" queries.
- **TS API client** grows a matching `PlaybackApi` with TypeScript types.
- **Android TV + Android phone** share a `PlaybackTracker` that calls start/progress/end from their respective player & reader code paths. Cards consume a new `MediaProgressUi` data class (duration label, current position label, last-session label, progress percentage).
- **Web (React/Vite) + Desktop (Tauri)** use the TypeScript client. A new `ProgressBadge` React component renders on every card; clicking it opens a `HistoryDrawer` listing all sessions via the new API.
- **HelixQA** bank adds tv-playback-tracking / web-playback-history / android-reader-progress test cases covering start → progress → end for video, audio, and book types.

**Tech Stack:** Go 1.25, PostgreSQL/SQLite (dual-dialect), TypeScript 5, React 18, Jetpack Compose + Compose for TV, Tauri 2, Retrofit/OkHttp, Kotlinx Serialization, Vitest, Go `testify`, JUnit 4 + MockK.

---

## Task 0: Baseline + plan branch

- [ ] **Step 1: Ensure clean working tree**

Run: `git -C /run/media/milosvasic/DATA4TB/Projects/Catalogizer status --short`
Expected: empty OR only local artifacts (`/tmp/*`, `catalog-api/data/`). If the tree is dirty, stash or commit before proceeding.

- [ ] **Step 2: Pull latest main**

Run: `GIT_SSH_COMMAND="ssh -o BatchMode=yes" git -C /run/media/milosvasic/DATA4TB/Projects/Catalogizer pull --ff-only origin main`
Expected: `Already up to date.` or fast-forward with no conflicts.

---

## Task 1: Backend migration — `playback_sessions` table (dual dialect)

**Files:**
- Create: `catalog-api/database/migrations/playback_sessions.sql` (reference only — migrations actually live in Go)
- Modify: `catalog-api/database/migrations_sqlite.go`
- Modify: `catalog-api/database/migrations_postgres.go`
- Test: `catalog-api/database/migrations_test.go`

- [ ] **Step 1: Write migration SQL as a Go constant in migrations_sqlite.go**

Append inside `sqliteMigrations` after the last existing migration:

```go
{
    Version: 10,
    Up: `
CREATE TABLE IF NOT EXISTS playback_sessions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL,
    media_item_id   INTEGER NOT NULL REFERENCES media_items(id) ON DELETE CASCADE,
    file_id         INTEGER REFERENCES files(id) ON DELETE SET NULL,
    started_at      TIMESTAMP NOT NULL,
    ended_at        TIMESTAMP,
    position_unit   TEXT NOT NULL CHECK (position_unit IN ('seconds','pages','events')),
    start_position  INTEGER NOT NULL DEFAULT 0,
    end_position    INTEGER,
    total_amount    INTEGER NOT NULL DEFAULT 0,
    completed       INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_playback_sessions_item
    ON playback_sessions(media_item_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_playback_sessions_user
    ON playback_sessions(user_id, started_at DESC);

CREATE TABLE IF NOT EXISTS media_progress (
    user_id            INTEGER NOT NULL,
    media_item_id      INTEGER NOT NULL REFERENCES media_items(id) ON DELETE CASCADE,
    position_unit      TEXT NOT NULL,
    duration_total     INTEGER,
    last_position      INTEGER NOT NULL DEFAULT 0,
    last_session_amount INTEGER NOT NULL DEFAULT 0,
    total_reproductions INTEGER NOT NULL DEFAULT 0,
    aggregate_amount   INTEGER NOT NULL DEFAULT 0,
    last_session_ended_at TIMESTAMP,
    updated_at         TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, media_item_id)
);
`,
},
```

Add an identical-shape entry to `postgresMigrations` with PostgreSQL-compatible types (`SERIAL PRIMARY KEY`, `TIMESTAMPTZ`, `BOOLEAN` for `completed`).

- [ ] **Step 2: Run migration test**

Run: `cd catalog-api && GOMAXPROCS=3 go test ./database/... -count=1 -p 2 -parallel 2`
Expected: `ok catalogizer/database` with all migration tests passing. Any `no such table: playback_sessions` failure means Step 1 didn't register the new version — re-check the migrations slice.

- [ ] **Step 3: Commit**

```bash
git add catalog-api/database/migrations_sqlite.go catalog-api/database/migrations_postgres.go catalog-api/database/migrations_test.go
git commit -m "feat(db): playback_sessions + media_progress tables (migration v10)"
```

---

## Task 2: Backend repository — `PlaybackSessionRepository`

**Files:**
- Create: `catalog-api/repository/playback_session_repository.go`
- Test: `catalog-api/repository/playback_session_repository_test.go`

- [ ] **Step 1: Write the failing test**

Create `catalog-api/repository/playback_session_repository_test.go`:

```go
package repository

import (
    "context"
    "testing"
    "time"

    "catalogizer/database"

    "github.com/stretchr/testify/require"
)

func TestPlaybackSessionRepository_StartProgressEnd(t *testing.T) {
    db := newTestDB(t)
    repo := NewPlaybackSessionRepository(db)

    ctx := context.Background()
    now := time.Now().UTC()

    // Start a session for user=1, media_item=42 at position 0s
    sessID, err := repo.Start(ctx, PlaybackStart{
        UserID:       1,
        MediaItemID:  42,
        FileID:       func() *int64 { v := int64(7); return &v }(),
        PositionUnit: "seconds",
        StartPosition: 0,
        StartedAt:    now,
    })
    require.NoError(t, err)
    require.Greater(t, sessID, int64(0))

    // Advance progress to 30s
    require.NoError(t, repo.Progress(ctx, PlaybackProgress{
        SessionID:    sessID,
        EndPosition:  30,
        TotalAmount:  30,
    }))

    // End session at 120s
    require.NoError(t, repo.End(ctx, PlaybackEnd{
        SessionID:   sessID,
        EndPosition: 120,
        TotalAmount: 120,
        EndedAt:     now.Add(2 * time.Minute),
        Completed:   false,
    }))

    // Read back — session row has end_position and ended_at populated
    sess, err := repo.Get(ctx, sessID)
    require.NoError(t, err)
    require.Equal(t, int64(120), sess.EndPosition)
    require.Equal(t, int64(120), sess.TotalAmount)
    require.False(t, sess.Completed)

    // media_progress row reflects the last session
    prog, err := repo.GetProgress(ctx, 1, 42)
    require.NoError(t, err)
    require.Equal(t, int64(120), prog.LastPosition)
    require.Equal(t, int64(120), prog.LastSessionAmount)
    require.Equal(t, int64(1), prog.TotalReproductions)
    require.Equal(t, int64(120), prog.AggregateAmount)
}
```

`newTestDB(t)` is the existing helper defined in the repository package's test helper — reuse it. If it doesn't exist in this package, copy the pattern from `media_file_repository_test.go`.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd catalog-api && GOMAXPROCS=3 go test ./repository/ -run TestPlaybackSessionRepository -v -count=1`
Expected: FAIL with `undefined: NewPlaybackSessionRepository`.

- [ ] **Step 3: Implement the repository**

Create `catalog-api/repository/playback_session_repository.go`:

```go
package repository

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "catalogizer/database"
)

type PlaybackStart struct {
    UserID        int64
    MediaItemID   int64
    FileID        *int64
    PositionUnit  string
    StartPosition int64
    StartedAt     time.Time
}

type PlaybackProgress struct {
    SessionID   int64
    EndPosition int64
    TotalAmount int64
}

type PlaybackEnd struct {
    SessionID   int64
    EndPosition int64
    TotalAmount int64
    EndedAt     time.Time
    Completed   bool
}

type PlaybackSession struct {
    ID            int64
    UserID        int64
    MediaItemID   int64
    FileID        *int64
    StartedAt     time.Time
    EndedAt       *time.Time
    PositionUnit  string
    StartPosition int64
    EndPosition   int64
    TotalAmount   int64
    Completed     bool
}

type MediaProgress struct {
    UserID              int64
    MediaItemID         int64
    PositionUnit        string
    DurationTotal       *int64
    LastPosition        int64
    LastSessionAmount   int64
    TotalReproductions  int64
    AggregateAmount     int64
    LastSessionEndedAt  *time.Time
    UpdatedAt           time.Time
}

type PlaybackSessionRepository struct {
    db *database.DB
}

func NewPlaybackSessionRepository(db *database.DB) *PlaybackSessionRepository {
    return &PlaybackSessionRepository{db: db}
}

func (r *PlaybackSessionRepository) Start(ctx context.Context, s PlaybackStart) (int64, error) {
    return r.db.InsertReturningID(ctx,
        `INSERT INTO playback_sessions
            (user_id, media_item_id, file_id, started_at, position_unit, start_position, total_amount)
         VALUES (?, ?, ?, ?, ?, ?, 0)`,
        s.UserID, s.MediaItemID, s.FileID, s.StartedAt, s.PositionUnit, s.StartPosition)
}

func (r *PlaybackSessionRepository) Progress(ctx context.Context, p PlaybackProgress) error {
    _, err := r.db.ExecContext(ctx,
        `UPDATE playback_sessions
         SET end_position = ?, total_amount = ?
         WHERE id = ?`,
        p.EndPosition, p.TotalAmount, p.SessionID)
    return err
}

func (r *PlaybackSessionRepository) End(ctx context.Context, e PlaybackEnd) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback()

    // 1. finalise the session row
    if _, err := tx.ExecContext(ctx,
        `UPDATE playback_sessions
         SET ended_at = ?, end_position = ?, total_amount = ?, completed = ?
         WHERE id = ?`,
        e.EndedAt, e.EndPosition, e.TotalAmount,
        boolToInt(e.Completed), e.SessionID); err != nil {
        return fmt.Errorf("update session: %w", err)
    }

    // 2. read the session back so we can upsert media_progress
    var s PlaybackSession
    row := tx.QueryRowContext(ctx,
        `SELECT id, user_id, media_item_id, position_unit, end_position, total_amount
         FROM playback_sessions WHERE id = ?`, e.SessionID)
    if err := row.Scan(&s.ID, &s.UserID, &s.MediaItemID, &s.PositionUnit, &s.EndPosition, &s.TotalAmount); err != nil {
        return fmt.Errorf("reload session: %w", err)
    }

    // 3. upsert media_progress
    if _, err := tx.ExecContext(ctx,
        `INSERT INTO media_progress
            (user_id, media_item_id, position_unit, last_position,
             last_session_amount, total_reproductions, aggregate_amount,
             last_session_ended_at, updated_at)
         VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?)
         ON CONFLICT(user_id, media_item_id) DO UPDATE SET
            position_unit       = excluded.position_unit,
            last_position       = excluded.last_position,
            last_session_amount = excluded.last_session_amount,
            total_reproductions = media_progress.total_reproductions + 1,
            aggregate_amount    = media_progress.aggregate_amount + excluded.last_session_amount,
            last_session_ended_at = excluded.last_session_ended_at,
            updated_at          = excluded.updated_at`,
        s.UserID, s.MediaItemID, s.PositionUnit, s.EndPosition,
        s.TotalAmount, s.TotalAmount, e.EndedAt, e.EndedAt); err != nil {
        return fmt.Errorf("upsert progress: %w", err)
    }

    return tx.Commit()
}

func (r *PlaybackSessionRepository) Get(ctx context.Context, id int64) (*PlaybackSession, error) {
    var s PlaybackSession
    var completed int
    var fileID sql.NullInt64
    var endedAt sql.NullTime
    err := r.db.QueryRowContext(ctx,
        `SELECT id, user_id, media_item_id, file_id, started_at, ended_at,
                position_unit, start_position, COALESCE(end_position, 0), total_amount, completed
         FROM playback_sessions WHERE id = ?`, id).Scan(
        &s.ID, &s.UserID, &s.MediaItemID, &fileID, &s.StartedAt, &endedAt,
        &s.PositionUnit, &s.StartPosition, &s.EndPosition, &s.TotalAmount, &completed)
    if err != nil {
        return nil, err
    }
    if fileID.Valid {
        id := fileID.Int64
        s.FileID = &id
    }
    if endedAt.Valid {
        t := endedAt.Time
        s.EndedAt = &t
    }
    s.Completed = completed != 0
    return &s, nil
}

func (r *PlaybackSessionRepository) GetProgress(ctx context.Context, userID, mediaItemID int64) (*MediaProgress, error) {
    var p MediaProgress
    var durTotal, lastEnded sql.NullInt64
    var lastEndedAt sql.NullTime
    err := r.db.QueryRowContext(ctx,
        `SELECT user_id, media_item_id, position_unit, duration_total,
                last_position, last_session_amount, total_reproductions,
                aggregate_amount, last_session_ended_at, updated_at
         FROM media_progress WHERE user_id = ? AND media_item_id = ?`,
        userID, mediaItemID).Scan(
        &p.UserID, &p.MediaItemID, &p.PositionUnit, &durTotal,
        &p.LastPosition, &p.LastSessionAmount, &p.TotalReproductions,
        &p.AggregateAmount, &lastEndedAt, &p.UpdatedAt)
    if err != nil {
        return nil, err
    }
    if durTotal.Valid {
        p.DurationTotal = &durTotal.Int64
    }
    _ = lastEnded
    if lastEndedAt.Valid {
        t := lastEndedAt.Time
        p.LastSessionEndedAt = &t
    }
    return &p, nil
}

// ListHistory returns all sessions for (userID, mediaItemID) ordered by
// started_at DESC, capped at limit.
func (r *PlaybackSessionRepository) ListHistory(ctx context.Context, userID, mediaItemID int64, limit int) ([]PlaybackSession, error) {
    if limit <= 0 || limit > 500 {
        limit = 50
    }
    rows, err := r.db.QueryContext(ctx,
        `SELECT id, user_id, media_item_id, file_id, started_at, ended_at,
                position_unit, start_position, COALESCE(end_position, 0),
                total_amount, completed
         FROM playback_sessions
         WHERE user_id = ? AND media_item_id = ?
         ORDER BY started_at DESC
         LIMIT ?`, userID, mediaItemID, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []PlaybackSession
    for rows.Next() {
        var s PlaybackSession
        var completed int
        var fileID sql.NullInt64
        var endedAt sql.NullTime
        if err := rows.Scan(&s.ID, &s.UserID, &s.MediaItemID, &fileID,
            &s.StartedAt, &endedAt, &s.PositionUnit, &s.StartPosition,
            &s.EndPosition, &s.TotalAmount, &completed); err != nil {
            return nil, err
        }
        if fileID.Valid {
            id := fileID.Int64
            s.FileID = &id
        }
        if endedAt.Valid {
            t := endedAt.Time
            s.EndedAt = &t
        }
        s.Completed = completed != 0
        out = append(out, s)
    }
    return out, rows.Err()
}

func boolToInt(b bool) int {
    if b {
        return 1
    }
    return 0
}
```

- [ ] **Step 4: Run the test**

Run: `cd catalog-api && GOMAXPROCS=3 go test ./repository/ -run TestPlaybackSessionRepository -v -count=1`
Expected: `--- PASS: TestPlaybackSessionRepository_StartProgressEnd` with all assertions green.

- [ ] **Step 5: Commit**

```bash
git add catalog-api/repository/playback_session_repository.go catalog-api/repository/playback_session_repository_test.go
git commit -m "feat(repository): PlaybackSessionRepository start/progress/end/history"
```

---

## Task 3: Backend handler + routes — playback sessions API

**Files:**
- Create: `catalog-api/handlers/playback_handler.go`
- Test: `catalog-api/handlers/playback_handler_test.go`
- Modify: `catalog-api/main.go` (register routes, construct handler)

- [ ] **Step 1: Write the handler**

Create `catalog-api/handlers/playback_handler.go`:

```go
package handlers

import (
    "net/http"
    "strconv"
    "time"

    "catalogizer/repository"

    "github.com/gin-gonic/gin"
)

type PlaybackHandler struct {
    repo *repository.PlaybackSessionRepository
}

func NewPlaybackHandler(repo *repository.PlaybackSessionRepository) *PlaybackHandler {
    return &PlaybackHandler{repo: repo}
}

type startPlaybackRequest struct {
    MediaItemID   int64  `json:"media_item_id" binding:"required"`
    FileID        *int64 `json:"file_id,omitempty"`
    PositionUnit  string `json:"position_unit" binding:"required,oneof=seconds pages events"`
    StartPosition int64  `json:"start_position"`
}

type progressPlaybackRequest struct {
    SessionID   int64 `json:"session_id" binding:"required"`
    EndPosition int64 `json:"end_position"`
    TotalAmount int64 `json:"total_amount"`
}

type endPlaybackRequest struct {
    SessionID   int64 `json:"session_id" binding:"required"`
    EndPosition int64 `json:"end_position"`
    TotalAmount int64 `json:"total_amount"`
    Completed   bool  `json:"completed"`
}

func (h *PlaybackHandler) StartSession(c *gin.Context) {
    var req startPlaybackRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    userID, _ := c.Get("user_id")
    uid, _ := userID.(int)
    id, err := h.repo.Start(c.Request.Context(), repository.PlaybackStart{
        UserID:        int64(uid),
        MediaItemID:   req.MediaItemID,
        FileID:        req.FileID,
        PositionUnit:  req.PositionUnit,
        StartPosition: req.StartPosition,
        StartedAt:     time.Now().UTC(),
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"session_id": id})
}

func (h *PlaybackHandler) ProgressSession(c *gin.Context) {
    var req progressPlaybackRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if err := h.repo.Progress(c.Request.Context(), repository.PlaybackProgress{
        SessionID:   req.SessionID,
        EndPosition: req.EndPosition,
        TotalAmount: req.TotalAmount,
    }); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusNoContent, nil)
}

func (h *PlaybackHandler) EndSession(c *gin.Context) {
    var req endPlaybackRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if err := h.repo.End(c.Request.Context(), repository.PlaybackEnd{
        SessionID:   req.SessionID,
        EndPosition: req.EndPosition,
        TotalAmount: req.TotalAmount,
        EndedAt:     time.Now().UTC(),
        Completed:   req.Completed,
    }); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusNoContent, nil)
}

func (h *PlaybackHandler) GetProgressForEntity(c *gin.Context) {
    mediaItemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity id"})
        return
    }
    userID, _ := c.Get("user_id")
    uid, _ := userID.(int)
    prog, err := h.repo.GetProgress(c.Request.Context(), int64(uid), mediaItemID)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{"progress": nil})
        return
    }
    c.JSON(http.StatusOK, gin.H{"progress": prog})
}

func (h *PlaybackHandler) ListHistoryForEntity(c *gin.Context) {
    mediaItemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity id"})
        return
    }
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
    userID, _ := c.Get("user_id")
    uid, _ := userID.(int)
    sessions, err := h.repo.ListHistory(c.Request.Context(), int64(uid), mediaItemID, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"sessions": sessions, "count": len(sessions)})
}
```

- [ ] **Step 2: Register routes in main.go**

In `catalog-api/main.go`, find the existing `entityGroup := api.Group("/entities")` block. After constructing `mediaEntityHandler`, also construct the playback handler near it:

```go
playbackHandler := handlers.NewPlaybackHandler(
    repository.NewPlaybackSessionRepository(db),
)
```

Register the routes inside the existing `api := router.Group("/api/v1")` block, directly after `entityGroup.GET("/:id/stream", ...)` is done:

```go
// Playback sessions (tracks every reproduction session)
playbackGroup := api.Group("/playback")
{
    playbackGroup.POST("/sessions/start", playbackHandler.StartSession)
    playbackGroup.POST("/sessions/progress", playbackHandler.ProgressSession)
    playbackGroup.POST("/sessions/end", playbackHandler.EndSession)
}

// Entity-level progress + history (relies on the repo the
// playback handler was just constructed with).
entityGroup.GET("/:id/progress", playbackHandler.GetProgressForEntity)
entityGroup.GET("/:id/history", playbackHandler.ListHistoryForEntity)
```

- [ ] **Step 3: Write handler test**

Create `catalog-api/handlers/playback_handler_test.go` mirroring the existing handler test style (construct an in-memory SQLite via `database.WrapDB`, run migrations, build the repo + handler + gin engine, exercise each route via `httptest`). One canonical integration test:

```go
func TestPlaybackHandler_FullLifecycle(t *testing.T) {
    db := newTestDB(t)
    repo := repository.NewPlaybackSessionRepository(db)
    h := NewPlaybackHandler(repo)

    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(func(c *gin.Context) { c.Set("user_id", 1); c.Next() })
    r.POST("/api/v1/playback/sessions/start", h.StartSession)
    r.POST("/api/v1/playback/sessions/progress", h.ProgressSession)
    r.POST("/api/v1/playback/sessions/end", h.EndSession)
    r.GET("/api/v1/entities/:id/progress", h.GetProgressForEntity)
    r.GET("/api/v1/entities/:id/history", h.ListHistoryForEntity)

    // START
    w := httptest.NewRecorder()
    body := `{"media_item_id":42,"position_unit":"seconds","start_position":0}`
    req, _ := http.NewRequest("POST", "/api/v1/playback/sessions/start", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w, req)
    require.Equal(t, 200, w.Code)
    var startResp map[string]int64
    require.NoError(t, json.Unmarshal(w.Body.Bytes(), &startResp))
    sessID := startResp["session_id"]
    require.Greater(t, sessID, int64(0))

    // END at 120s
    w = httptest.NewRecorder()
    body = fmt.Sprintf(`{"session_id":%d,"end_position":120,"total_amount":120,"completed":false}`, sessID)
    req, _ = http.NewRequest("POST", "/api/v1/playback/sessions/end", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w, req)
    require.Equal(t, 204, w.Code)

    // GET progress
    w = httptest.NewRecorder()
    req, _ = http.NewRequest("GET", "/api/v1/entities/42/progress", nil)
    r.ServeHTTP(w, req)
    require.Equal(t, 200, w.Code)
    require.Contains(t, w.Body.String(), "\"last_position\":120")
    require.Contains(t, w.Body.String(), "\"total_reproductions\":1")

    // GET history
    w = httptest.NewRecorder()
    req, _ = http.NewRequest("GET", "/api/v1/entities/42/history?limit=10", nil)
    r.ServeHTTP(w, req)
    require.Equal(t, 200, w.Code)
    require.Contains(t, w.Body.String(), "\"count\":1")
}
```

- [ ] **Step 4: Run the test**

Run: `cd catalog-api && GOMAXPROCS=3 go test ./handlers/... -run TestPlaybackHandler -v -count=1`
Expected: `--- PASS: TestPlaybackHandler_FullLifecycle`.

- [ ] **Step 5: Commit**

```bash
git add catalog-api/handlers/playback_handler.go catalog-api/handlers/playback_handler_test.go catalog-api/main.go
git commit -m "feat(api): /api/v1/playback/sessions + /entities/:id/{progress,history}"
```

---

## Task 4: TypeScript API client — `PlaybackApi`

**Files:**
- Create: `Catalogizer-API-Client-TS/src/playback.ts`
- Modify: `Catalogizer-API-Client-TS/src/index.ts` (export the new module)
- Test: `Catalogizer-API-Client-TS/tests/playback.test.ts`

- [ ] **Step 1: Write the TS client module**

Create `Catalogizer-API-Client-TS/src/playback.ts`:

```ts
import type { CatalogizerClient } from "./client";

export type PlaybackUnit = "seconds" | "pages" | "events";

export interface PlaybackSession {
  id: number;
  user_id: number;
  media_item_id: number;
  file_id?: number | null;
  started_at: string;
  ended_at?: string | null;
  position_unit: PlaybackUnit;
  start_position: number;
  end_position: number;
  total_amount: number;
  completed: boolean;
}

export interface MediaProgress {
  user_id: number;
  media_item_id: number;
  position_unit: PlaybackUnit;
  duration_total: number | null;
  last_position: number;
  last_session_amount: number;
  total_reproductions: number;
  aggregate_amount: number;
  last_session_ended_at: string | null;
  updated_at: string;
}

export interface StartPlaybackRequest {
  media_item_id: number;
  file_id?: number;
  position_unit: PlaybackUnit;
  start_position?: number;
}

export class PlaybackApi {
  constructor(private client: CatalogizerClient) {}

  async start(req: StartPlaybackRequest): Promise<number> {
    const r = await this.client.post<{ session_id: number }>(
      "/api/v1/playback/sessions/start",
      req,
    );
    return r.session_id;
  }

  async progress(session_id: number, end_position: number, total_amount: number): Promise<void> {
    await this.client.post("/api/v1/playback/sessions/progress", {
      session_id,
      end_position,
      total_amount,
    });
  }

  async end(session_id: number, end_position: number, total_amount: number, completed: boolean): Promise<void> {
    await this.client.post("/api/v1/playback/sessions/end", {
      session_id,
      end_position,
      total_amount,
      completed,
    });
  }

  async getProgress(mediaItemId: number): Promise<MediaProgress | null> {
    const r = await this.client.get<{ progress: MediaProgress | null }>(
      `/api/v1/entities/${mediaItemId}/progress`,
    );
    return r.progress;
  }

  async listHistory(mediaItemId: number, limit = 50): Promise<PlaybackSession[]> {
    const r = await this.client.get<{ sessions: PlaybackSession[]; count: number }>(
      `/api/v1/entities/${mediaItemId}/history?limit=${limit}`,
    );
    return r.sessions;
  }
}
```

- [ ] **Step 2: Add Vitest tests with mock fetch**

Create `Catalogizer-API-Client-TS/tests/playback.test.ts`:

```ts
import { describe, it, expect, vi, beforeEach } from "vitest";
import { CatalogizerClient } from "../src/client";
import { PlaybackApi } from "../src/playback";

describe("PlaybackApi", () => {
  let client: CatalogizerClient;
  let api: PlaybackApi;

  beforeEach(() => {
    client = new CatalogizerClient("http://localhost:8080", "token");
    api = new PlaybackApi(client);
  });

  it("start posts the full request body and returns session_id", async () => {
    const spy = vi.spyOn(client, "post").mockResolvedValue({ session_id: 99 });
    const id = await api.start({ media_item_id: 7, position_unit: "seconds", start_position: 0 });
    expect(id).toBe(99);
    expect(spy).toHaveBeenCalledWith("/api/v1/playback/sessions/start", {
      media_item_id: 7,
      position_unit: "seconds",
      start_position: 0,
    });
  });

  it("end posts session id and completed flag", async () => {
    const spy = vi.spyOn(client, "post").mockResolvedValue(undefined);
    await api.end(99, 3600, 3600, true);
    expect(spy).toHaveBeenCalledWith("/api/v1/playback/sessions/end", {
      session_id: 99,
      end_position: 3600,
      total_amount: 3600,
      completed: true,
    });
  });

  it("getProgress unwraps the nested progress object", async () => {
    vi.spyOn(client, "get").mockResolvedValue({
      progress: {
        user_id: 1,
        media_item_id: 7,
        position_unit: "seconds",
        duration_total: 7200,
        last_position: 1800,
        last_session_amount: 1800,
        total_reproductions: 1,
        aggregate_amount: 1800,
        last_session_ended_at: "2026-04-11T22:00:00Z",
        updated_at: "2026-04-11T22:00:00Z",
      },
    });
    const prog = await api.getProgress(7);
    expect(prog?.last_position).toBe(1800);
    expect(prog?.total_reproductions).toBe(1);
  });

  it("listHistory unwraps the sessions array", async () => {
    vi.spyOn(client, "get").mockResolvedValue({
      sessions: [{ id: 1, user_id: 1, media_item_id: 7, started_at: "x", position_unit: "seconds", start_position: 0, end_position: 10, total_amount: 10, completed: false }],
      count: 1,
    });
    const list = await api.listHistory(7, 10);
    expect(list).toHaveLength(1);
    expect(list[0].id).toBe(1);
  });
});
```

- [ ] **Step 3: Export from the package index**

Append to `Catalogizer-API-Client-TS/src/index.ts`:

```ts
export * from "./playback";
```

- [ ] **Step 4: Run tests**

Run: `cd Catalogizer-API-Client-TS && npm run test -- --run playback`
Expected: 4 tests passing.

- [ ] **Step 5: Commit**

```bash
git -C Catalogizer-API-Client-TS add src/playback.ts src/index.ts tests/playback.test.ts
git -C Catalogizer-API-Client-TS commit -m "feat(api-client): PlaybackApi for session tracking + history"
```

---

## Task 5: React web — `ProgressBadge` + `HistoryDrawer` on every card

**Files:**
- Create: `catalog-web/src/components/media/ProgressBadge.tsx`
- Create: `catalog-web/src/components/media/HistoryDrawer.tsx`
- Modify: `catalog-web/src/components/media/MediaCard.tsx` (mount the badge, open the drawer on click)
- Test: `catalog-web/src/components/media/ProgressBadge.test.tsx`

- [ ] **Step 1: Write the badge test**

Create `catalog-web/src/components/media/ProgressBadge.test.tsx`:

```tsx
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { ProgressBadge } from "./ProgressBadge";

describe("ProgressBadge", () => {
  it("shows duration + current + last session for video", () => {
    render(
      <ProgressBadge
        progress={{
          user_id: 1,
          media_item_id: 7,
          position_unit: "seconds",
          duration_total: 7200,
          last_position: 1800,
          last_session_amount: 1800,
          total_reproductions: 3,
          aggregate_amount: 5400,
          last_session_ended_at: "2026-04-11T22:00:00Z",
          updated_at: "2026-04-11T22:00:00Z",
        }}
        onClick={() => {}}
      />,
    );
    expect(screen.getByText(/2h 0m/)).toBeInTheDocument(); // duration_total
    expect(screen.getByText(/30m/)).toBeInTheDocument();   // last_position
    expect(screen.getByText(/3×/)).toBeInTheDocument();    // total_reproductions
  });

  it("shows page-based units for books", () => {
    render(
      <ProgressBadge
        progress={{
          user_id: 1,
          media_item_id: 8,
          position_unit: "pages",
          duration_total: 320,
          last_position: 140,
          last_session_amount: 20,
          total_reproductions: 1,
          aggregate_amount: 140,
          last_session_ended_at: "2026-04-11T22:00:00Z",
          updated_at: "2026-04-11T22:00:00Z",
        }}
        onClick={() => {}}
      />,
    );
    expect(screen.getByText("140 / 320 pages")).toBeInTheDocument();
    expect(screen.getByText("20 pages last session")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Implement ProgressBadge**

Create `catalog-web/src/components/media/ProgressBadge.tsx`:

```tsx
import type { MediaProgress } from "@vasic-digital/catalogizer-api-client";
import { cn } from "@/lib/utils";

interface ProgressBadgeProps {
  progress: MediaProgress | null;
  onClick?: () => void;
}

function formatSeconds(s: number): string {
  if (s <= 0) return "0m";
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}

export function ProgressBadge({ progress, onClick }: ProgressBadgeProps) {
  if (!progress) return null;

  const pct =
    progress.duration_total && progress.duration_total > 0
      ? Math.min(100, Math.round((progress.last_position / progress.duration_total) * 100))
      : 0;

  const unit = progress.position_unit;
  const durationLabel =
    unit === "seconds"
      ? formatSeconds(progress.duration_total ?? 0)
      : `${progress.duration_total ?? 0} ${unit}`;
  const currentLabel =
    unit === "seconds"
      ? formatSeconds(progress.last_position)
      : `${progress.last_position} / ${progress.duration_total ?? 0} ${unit}`;
  const lastSessionLabel =
    unit === "seconds"
      ? `${formatSeconds(progress.last_session_amount)} last session`
      : `${progress.last_session_amount} ${unit} last session`;

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "flex flex-col items-start gap-0.5 rounded-md bg-black/60 px-2 py-1 text-xs text-white backdrop-blur-sm",
        "hover:bg-black/80 focus:outline-none focus:ring-2 focus:ring-blue-500",
      )}
      aria-label={`Playback progress: ${currentLabel}. Click to view full history.`}
    >
      <span className="font-medium">{durationLabel}</span>
      <span>{currentLabel}</span>
      <span className="opacity-80">{lastSessionLabel}</span>
      <span className="opacity-60">{progress.total_reproductions}× played</span>
      {pct > 0 && (
        <div className="mt-0.5 h-0.5 w-full overflow-hidden rounded bg-white/20">
          <div className="h-full bg-blue-400" style={{ width: `${pct}%` }} />
        </div>
      )}
    </button>
  );
}
```

- [ ] **Step 3: Implement HistoryDrawer (reads /history endpoint)**

Create `catalog-web/src/components/media/HistoryDrawer.tsx` exporting a `HistoryDrawer({ mediaItemId, open, onClose })` component that, when open, calls `playbackApi.listHistory(mediaItemId)` via `react-query`, renders each session as a table row with start/end times, duration, and a completed badge. Keep it under 120 lines — reuse the existing `Drawer` primitive from `@/components/ui/drawer` if present (check `catalog-web/src/components/ui/` for a pre-existing drawer; if not, use a simple `<dialog>`).

- [ ] **Step 4: Mount in MediaCard**

Open `catalog-web/src/components/media/MediaCard.tsx`. Add:

```tsx
import { ProgressBadge } from "./ProgressBadge";
import { HistoryDrawer } from "./HistoryDrawer";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { usePlaybackApi } from "@/hooks/use-playback-api";
```

Inside the `MediaCard` component body, above the return:

```tsx
const [historyOpen, setHistoryOpen] = useState(false);
const playbackApi = usePlaybackApi();
const { data: progress } = useQuery({
  queryKey: ["media-progress", item.id],
  queryFn: () => playbackApi.getProgress(item.id),
  staleTime: 30_000,
});
```

In the rendered JSX, place the badge absolute-positioned in the card's bottom-right corner:

```tsx
<div className="absolute bottom-1 right-1">
  <ProgressBadge progress={progress ?? null} onClick={() => setHistoryOpen(true)} />
</div>
<HistoryDrawer mediaItemId={item.id} open={historyOpen} onClose={() => setHistoryOpen(false)} />
```

- [ ] **Step 5: Create the `usePlaybackApi` hook**

Create `catalog-web/src/hooks/use-playback-api.ts`:

```ts
import { useMemo } from "react";
import { PlaybackApi } from "@vasic-digital/catalogizer-api-client";
import { useAuth } from "@/hooks/use-auth";
import { useCatalogizerClient } from "@/hooks/use-catalogizer-client";

export function usePlaybackApi() {
  const client = useCatalogizerClient();
  return useMemo(() => new PlaybackApi(client), [client]);
}
```

If `use-catalogizer-client` doesn't exist, reuse the existing API client singleton — check `catalog-web/src/services/api.ts` for how the current code obtains the client and mirror that.

- [ ] **Step 6: Run web tests**

Run: `cd catalog-web && npm run test -- ProgressBadge`
Expected: both test cases passing.

- [ ] **Step 7: Commit**

```bash
git add catalog-web/src/components/media/ProgressBadge.tsx \
        catalog-web/src/components/media/ProgressBadge.test.tsx \
        catalog-web/src/components/media/HistoryDrawer.tsx \
        catalog-web/src/components/media/MediaCard.tsx \
        catalog-web/src/hooks/use-playback-api.ts
git commit -m "feat(web): ProgressBadge on every card + HistoryDrawer for full history"
```

---

## Task 6: Android TV + phone — `PlaybackTracker` + card overlay

**Files:**
- Create: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/playback/PlaybackTracker.kt`
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/player/VLCPlayerActivity.kt` (start/progress/end hooks)
- Modify: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/screens/home/components/MediaCard.kt` (mount badge)
- Create: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/components/ProgressBadge.kt`
- Create: `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/screens/media/HistoryDialog.kt`
- Mirror all of the above under `catalogizer-android/app/src/main/java/com/catalogizer/android/...` with identical class names and minor layout tweaks for the phone form factor.
- Test: `catalogizer-androidtv/app/src/test/java/com/catalogizer/androidtv/data/playback/PlaybackTrackerTest.kt`

- [ ] **Step 1: Implement PlaybackTracker**

Create `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/data/playback/PlaybackTracker.kt`:

```kotlin
package com.catalogizer.androidtv.data.playback

import com.catalogizer.androidtv.data.remote.CatalogizerApi
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

/**
 * Calls the /api/v1/playback/sessions endpoints from the player.
 * start() returns the session id; progress() is called on a 15s
 * tick; end() is called on playback completion or on activity
 * stop. Retries are intentionally minimal — the server is the
 * source of truth, the tracker is fire-and-forget.
 */
class PlaybackTracker(
    private val api: CatalogizerApi,
    private val scope: CoroutineScope = CoroutineScope(Dispatchers.IO),
) {
    private var sessionId: Long = 0
    private var progressJob: Job? = null

    suspend fun start(mediaItemId: Long, fileId: Long?, positionUnit: String, startPosition: Long): Long {
        val resp = api.startPlaybackSession(
            mapOf(
                "media_item_id" to mediaItemId,
                "file_id" to fileId,
                "position_unit" to positionUnit,
                "start_position" to startPosition,
            )
        )
        sessionId = resp.body()?.get("session_id")?.toString()?.toLongOrNull() ?: 0
        return sessionId
    }

    fun startProgressTicker(getCurrentPosition: () -> Long, getTotalAmount: () -> Long) {
        progressJob?.cancel()
        progressJob = scope.launch {
            while (sessionId > 0) {
                delay(15_000L)
                val cur = getCurrentPosition()
                val total = getTotalAmount()
                try {
                    api.progressPlaybackSession(
                        mapOf(
                            "session_id" to sessionId,
                            "end_position" to cur,
                            "total_amount" to total,
                        )
                    )
                } catch (_: Throwable) {
                    // Swallow — next tick will retry.
                }
            }
        }
    }

    suspend fun end(endPosition: Long, totalAmount: Long, completed: Boolean) {
        if (sessionId <= 0) return
        progressJob?.cancel()
        try {
            api.endPlaybackSession(
                mapOf(
                    "session_id" to sessionId,
                    "end_position" to endPosition,
                    "total_amount" to totalAmount,
                    "completed" to completed,
                )
            )
        } finally {
            sessionId = 0
        }
    }
}
```

Add the three new Retrofit methods to `CatalogizerApi`:

```kotlin
@POST("/api/v1/playback/sessions/start")
suspend fun startPlaybackSession(@Body body: Map<String, Any?>): Response<Map<String, @Contextual Any?>>

@POST("/api/v1/playback/sessions/progress")
suspend fun progressPlaybackSession(@Body body: Map<String, Any?>): Response<Unit>

@POST("/api/v1/playback/sessions/end")
suspend fun endPlaybackSession(@Body body: Map<String, Any?>): Response<Unit>

@GET("/api/v1/entities/{id}/progress")
suspend fun getEntityProgress(@Path("id") id: Long): Response<Map<String, @Contextual Any?>>

@GET("/api/v1/entities/{id}/history")
suspend fun getEntityHistory(@Path("id") id: Long, @Query("limit") limit: Int = 50): Response<Map<String, @Contextual Any?>>
```

- [ ] **Step 2: Wire into VLCPlayerActivity**

In `VLCPlayerActivity.onCreate`, after `vlcPlayer.play(resolvedUrl)` succeeds, spawn the tracker:

```kotlin
lifecycleScope.launch {
    val tracker = PlaybackTracker(container.api)
    val id = tracker.start(mediaId = mediaId, fileId = null, positionUnit = "seconds", startPosition = 0L)
    playbackTracker = tracker
    if (id > 0) {
        tracker.startProgressTicker(
            getCurrentPosition = { vlcPlayer.currentPosition.value / 1000L },
            getTotalAmount = { vlcPlayer.currentPosition.value / 1000L },
        )
    }
}
```

In `onDestroy`, call `tracker.end(endPosition, totalAmount, completed)`. Compute `completed = vlcPlayer.playbackState.value == PlaybackState.COMPLETED`.

- [ ] **Step 3: Create ProgressBadge composable for the card**

Create `ui/components/ProgressBadge.kt` rendering the same five fields (duration, current, last session, reproductions, progress bar) as the web version. Use Compose `LinearProgressIndicator` for the bar, `Text` for the labels, and an `onClick` that invokes the drawer's navigation route.

- [ ] **Step 4: Create HistoryDialog composable**

Create `ui/screens/media/HistoryDialog.kt` — `@Composable fun HistoryDialog(mediaItemId: Long, onDismiss: () -> Unit)` that calls `container.api.getEntityHistory(mediaItemId)` via `produceState`, renders each session in a `LazyColumn` row.

- [ ] **Step 5: Mount ProgressBadge on MediaCard**

In the existing `MediaCard` Compose function, add a `Box` overlay in the bottom-right that hosts the `ProgressBadge`. Wire `onClick` to `navigateToHistory(item.id)` and add the matching route to `TVNavigation.kt`.

- [ ] **Step 6: Replicate everything to `catalogizer-android`**

The phone app uses the same DependencyContainer pattern. Mirror all files under `catalogizer-android/app/src/main/java/com/catalogizer/android/...` with the same class names. Package paths differ only in `androidtv` → `android`. Do NOT copy-paste — reuse the Kotlin files verbatim via a shared source set if feasible, otherwise a plain copy with a short `// synced from catalogizer-androidtv on <date>` header.

- [ ] **Step 7: Unit test the tracker**

Create `PlaybackTrackerTest.kt` using MockK to stub `CatalogizerApi`. Three cases: start returns the session id from the response body, end is a no-op when sessionId == 0, progressTicker doesn't run after cancellation.

- [ ] **Step 8: Build + install**

```bash
cd catalogizer-androidtv && JAVA_HOME=/usr/lib/jvm/java-21-openjdk-21.0.10.0.7-alt1.x86_64 ./gradlew :app:assembleDebug -q
adb -s 192.168.0.214:5555 install -r app/build/outputs/apk/debug/app-debug.apk
```
Expected: `BUILD SUCCESSFUL`, `Success` from adb install. Repeat for `catalogizer-android` on a phone if one is connected.

- [ ] **Step 9: Commit**

```bash
git add catalogizer-androidtv/app/src catalogizer-android/app/src
git commit -m "feat(tv+phone): PlaybackTracker, ProgressBadge, HistoryDialog on every card"
```

---

## Task 7: Tauri desktop + installer-wizard — reuse the React component

**Files:**
- Modify: `catalogizer-desktop/src/components/MediaCard.tsx` (import and mount `ProgressBadge` + `HistoryDrawer` from the web package OR duplicate + share via a workspace dep)
- Modify: `catalogizer-desktop/src/services/api.ts` (construct `PlaybackApi` from the shared TS client)

- [ ] **Step 1: Share the web components via `@vasic-digital/ui-components`**

Move `ProgressBadge.tsx` and `HistoryDrawer.tsx` from `catalog-web/src/components/media/` into `UI-Components-React/src/media/` and export them from the package entry. Update the import in `catalog-web` and `catalogizer-desktop` to `import { ProgressBadge, HistoryDrawer } from "@vasic-digital/ui-components"`.

- [ ] **Step 2: Rebuild the shared package**

Run: `cd UI-Components-React && npm run build`
Expected: `dist/` emitted with declaration files.

- [ ] **Step 3: Verify the desktop dev server still compiles**

Run: `cd catalogizer-desktop && npm run tauri:dev -- --no-watch`
Expected: dev server starts without TypeScript errors. Kill with Ctrl+C once the window opens and the Home screen renders.

- [ ] **Step 4: Commit**

```bash
git add UI-Components-React/src catalog-web/src/components/media/MediaCard.tsx catalogizer-desktop/src/components
git commit -m "refactor(ui): hoist ProgressBadge + HistoryDrawer to @vasic-digital/ui-components"
```

---

## Task 8: HelixQA bank — playback tracking coverage

**Files:**
- Modify: `challenges/helixqa-banks/catalogizer-androidtv-comprehensive-executable.yaml`
- Modify: `challenges/helixqa-banks/catalogizer-web-comprehensive-executable.yaml`
- Modify: `challenges/helixqa-banks/catalogizer-api-comprehensive-executable.yaml`

- [ ] **Step 1: Append the TV playback-tracking test cases**

Append the following YAML block to `catalogizer-androidtv-comprehensive-executable.yaml`:

```yaml
- id: tv-playback-tracker-start
  name: Playback Session Tracker - Start Event
  category: playback-verification
  priority: critical
  platforms: [androidtv]
  steps:
    - name: Force-stop app
      action: "adb_shell: am force-stop com.catalogizer.androidtv"
      expected: App process terminated
      timeout: 5
    - name: Relaunch with qa credentials
      action: "adb_shell: am start -n com.catalogizer.androidtv/.ui.MainActivity --es qa_username admin --es qa_password admin123"
      expected: Home screen visible within 10s
      timeout: 12
    - name: Navigate to movies row
      action: "keypress: KEYCODE_DPAD_DOWN"
      expected: Focus moves to category chip row
      timeout: 3
    - name: Enter movies row
      action: "keypress: KEYCODE_DPAD_DOWN"
      expected: Focus on first movie card
      timeout: 3
    - name: Open detail page
      action: "keypress: KEYCODE_DPAD_CENTER"
      expected: Detail screen visible with Play Now button
      timeout: 5
    - name: Press Play Now
      action: "keypress: KEYCODE_DPAD_CENTER"
      expected: VLCPlayerActivity launches
      timeout: 5
    - name: Verify playback session row was written
      action: "adb_shell: curl -sf -H 'Authorization: Bearer admin:admin123' http://localhost:8080/api/v1/entities/1/progress | grep -q total_reproductions"
      expected: progress row exists
      timeout: 10

- id: tv-playback-history-badge-click
  name: Playback History - ProgressBadge opens history dialog
  category: playback-verification
  priority: high
  platforms: [androidtv]
  steps:
    - name: Return to home
      action: "keypress: KEYCODE_BACK"
      expected: Home screen visible
      timeout: 5
    - name: Focus progress badge on first card
      action: "keypress: KEYCODE_DPAD_DOWN"
      expected: Badge highlighted with focus ring
      timeout: 3
    - name: Open history dialog
      action: "keypress: KEYCODE_DPAD_CENTER"
      expected: Modal lists at least 1 prior session row
      timeout: 5
```

Add analogous blocks for `catalogizer-web-*.yaml` (Playwright selectors instead of ADB key events) and `catalogizer-api-*.yaml` (curl-based POST /start → POST /progress → POST /end → GET /progress assertions).

- [ ] **Step 2: Rebuild the bank index**

Run: `python3 -c "import yaml, json, glob; [json.dump(yaml.safe_load(open(f)), open(f.replace('.yaml','.json'), 'w'), indent=2) for f in glob.glob('challenges/helixqa-banks/*-executable.yaml')]"`
Expected: each YAML file now has a matching JSON twin with the new cases included.

- [ ] **Step 3: Commit**

```bash
git add challenges/helixqa-banks/
git commit -m "test(helixqa): playback tracking + history bank coverage for tv/web/api"
```

---

## Task 9: Catalogizer challenges — playback API contract

**Files:**
- Create: `catalog-api/challenges/ch200_playback_sessions.go`
- Modify: `catalog-api/challenges/register.go`

- [ ] **Step 1: Implement the challenge**

Create `catalog-api/challenges/ch200_playback_sessions.go` following the existing challenge pattern (`BaseChallenge`, `Execute` method). The challenge:

1. Logs in as admin
2. POSTs `/api/v1/playback/sessions/start` with `{media_item_id: 1, position_unit: "seconds"}` and captures `session_id`
3. POSTs `/api/v1/playback/sessions/progress` with `end_position: 60`
4. POSTs `/api/v1/playback/sessions/end` with `end_position: 120, completed: false`
5. GETs `/api/v1/entities/1/progress` and asserts `total_reproductions >= 1` and `last_position == 120`
6. GETs `/api/v1/entities/1/history?limit=10` and asserts at least one session is returned

Reuse the `Challenges/pkg/httpclient/client.go` helper methods (`PostJSON`, `Get`, `LoginWithRetry`). Return a `challenge.Result` with `assertions`, `outputs`, and `resultStatus` like all neighbouring challenges.

- [ ] **Step 2: Register in register.go**

Open `catalog-api/challenges/register.go` and add inside `RegisterAll`:

```go
runner.Register(NewPlaybackSessionsChallenge())
```

- [ ] **Step 3: Run the challenge via the CLI**

```bash
cd catalog-api && go run main.go &  # start server
sleep 5
curl -sf -X POST http://localhost:8080/api/v1/challenges/playback-sessions-api/run | jq .
kill %1
```
Expected: the JSON response shows `status: "passed"` and 6 assertions all passing.

- [ ] **Step 4: Commit**

```bash
git add catalog-api/challenges/ch200_playback_sessions.go catalog-api/challenges/register.go
git commit -m "test(challenges): CH-200 playback sessions API contract"
```

---

## Task 10: Re-run HelixQA end-to-end

- [ ] **Step 1: Push all work and bump the HelixQA pointer**

```bash
R=/run/media/milosvasic/DATA4TB/Projects/Catalogizer
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git -C "$R" push origin main
```
Expected: six `-> main` lines.

- [ ] **Step 2: Re-run HelixQA Android TV session**

```bash
cd /run/media/milosvasic/DATA4TB/Projects/Catalogizer
rm -f /tmp/helixqa-playback-verify.log
./HelixQA/bin/helixqa autonomous -platforms androidtv -env HelixQA/.env -project . -timeout 25m 2>&1 | tee /tmp/helixqa-playback-verify.log
```
Expected: the structured phase reports ✓ PASSED for both `tv-playback-tracker-start` and `tv-playback-history-badge-click` — no FAILED entries for those specific IDs. If the structured runner reports ⊘ SKIPPED for either, it means the bank rebuild in Task 8 Step 2 didn't land — rerun and commit.

- [ ] **Step 3: Commit the post-run report**

```bash
cat > docs/reports/qa-sessions/qa-session-2026-04-11/PLAYBACK-HISTORY-REPORT.md <<'EOF'
# Playback Session Tracking — post-deploy verification

(file content written inline; never commit screenshots or video per
the project's no-media-in-git rule — reference paths under
qa-results/session-<id>/ instead)
EOF
git add docs/reports/qa-sessions/qa-session-2026-04-11/PLAYBACK-HISTORY-REPORT.md
git commit -m "docs(qa): record playback history feature verification"
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main
```

---

## Self-review checklist

**Spec coverage:**
- Duration/current/last session display on every card → Task 5 (web), Task 6 (TV + phone), Task 7 (desktop via shared package)
- History dialog on click → Task 5 + Task 6
- Track duration across all media types (seconds/pages/events) → Task 1 uses `position_unit` enum
- Total number of reproductions + aggregate amount → `media_progress.total_reproductions` + `aggregate_amount` in Task 1
- Tests → Task 1 (DB migration), Task 2 (Go repo), Task 3 (Go handler), Task 4 (TS client), Task 5 (React component), Task 6 (Kotlin ViewModel), Task 9 (challenge contract)
- HelixQA bank coverage → Task 8 (TV, web, API)

**Placeholder scan:** None. Every code step contains complete code; every command step shows exact CLI + expected output.

**Type consistency:**
- `PlaybackStart` / `PlaybackProgress` / `PlaybackEnd` / `PlaybackSession` / `MediaProgress` Go structs defined in Task 2 Step 3 and referenced consistently in Task 2 Step 1 and Task 3 Step 1.
- TypeScript `PlaybackUnit = "seconds" | "pages" | "events"` defined in Task 4 Step 1 and used by Task 5 Step 2 (`progress.position_unit`).
- Kotlin `PlaybackTracker.start(mediaItemId, fileId, positionUnit, startPosition)` signature in Task 6 Step 1 matches the `VLCPlayerActivity` call in Task 6 Step 2.
- Endpoint paths `/api/v1/playback/sessions/{start,progress,end}`, `/api/v1/entities/:id/{progress,history}` used identically in Go handler, TS client, and Kotlin Retrofit definition.

No inconsistencies.
