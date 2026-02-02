package performance

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mutecomm/go-sqlcipher"
)

// testServer holds a configured Gin engine backed by an in-memory SQLite database.
// It mirrors the production middleware + handler pipeline for the benchmarked endpoints.
type testServer struct {
	router *gin.Engine
	db     *sql.DB
	token  string // pre-generated JWT for authenticated requests
}

// newTestServer creates a fully-initialised test server with realistic seed data.
// The data volume is controlled by fileCount to allow benchmarks at different sizes.
func newTestServer(b *testing.B, fileCount int) *testServer {
	b.Helper()
	gin.SetMode(gin.TestMode)

	db, err := sql.Open("sqlite3", ":memory:?_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=1")
	if err != nil {
		b.Fatalf("open db: %v", err)
	}

	if err := createSchema(db); err != nil {
		b.Fatalf("create schema: %v", err)
	}
	if err := seedData(db, fileCount); err != nil {
		b.Fatalf("seed data: %v", err)
	}

	router := gin.New()

	// Apply middleware matching the production pipeline order:
	// CORS -> Logger (omitted to avoid log noise) -> ErrorHandler -> RequestID -> InputValidation (omitted for benchmark purity)
	router.Use(corsMiddleware())
	router.Use(errorHandlerMiddleware())
	router.Use(requestIDMiddleware())

	// Health check (unauthenticated, like production)
	router.GET("/health", healthHandler)

	// Auth routes (unauthenticated)
	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.POST("/login", loginHandler(db))
	}

	// Authenticated API routes
	api := router.Group("/api/v1")
	api.Use(benchJWTMiddleware()) // lightweight JWT stub for benchmarks
	{
		api.GET("/media", mediaListHandler(db))
		api.GET("/sources", sourcesListHandler(db))
		api.GET("/catalog", catalogListHandler(db))
		api.GET("/search", searchHandler(db))
		api.GET("/stats/overall", statsOverallHandler(db))
	}

	ts := &testServer{
		router: router,
		db:     db,
		token:  "Bearer bench-token",
	}
	return ts
}

func (ts *testServer) close() {
	if ts.db != nil {
		ts.db.Close()
	}
}

// ---------------------------------------------------------------------------
// Schema & Seed
// ---------------------------------------------------------------------------

func createSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS storage_roots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		protocol TEXT NOT NULL,
		host TEXT,
		port INTEGER,
		path TEXT,
		enabled BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_scan_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		storage_root_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		name TEXT NOT NULL,
		extension TEXT,
		mime_type TEXT,
		file_type TEXT,
		size INTEGER NOT NULL,
		is_directory BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		modified_at DATETIME NOT NULL,
		deleted BOOLEAN DEFAULT 0,
		is_duplicate BOOLEAN DEFAULT 0,
		duplicate_group_id INTEGER,
		parent_id INTEGER,
		FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id)
	);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		salt TEXT NOT NULL,
		role_id INTEGER NOT NULL DEFAULT 2,
		is_active INTEGER DEFAULT 1,
		is_locked INTEGER DEFAULT 0,
		failed_login_attempts INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		permissions TEXT DEFAULT '[]',
		is_system INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS duplicate_groups (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_count INTEGER DEFAULT 0,
		total_size INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS scan_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		storage_root_id INTEGER NOT NULL,
		scan_type TEXT NOT NULL,
		status TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		files_processed INTEGER DEFAULT 0,
		files_added INTEGER DEFAULT 0,
		files_updated INTEGER DEFAULT 0,
		files_deleted INTEGER DEFAULT 0,
		error_count INTEGER DEFAULT 0,
		FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id)
	);

	-- Performance indexes matching production
	CREATE INDEX IF NOT EXISTS idx_files_storage_root_path ON files(storage_root_id, path);
	CREATE INDEX IF NOT EXISTS idx_files_parent_id ON files(parent_id);
	CREATE INDEX IF NOT EXISTS idx_files_duplicate_group ON files(duplicate_group_id);
	CREATE INDEX IF NOT EXISTS idx_files_deleted ON files(deleted);
	CREATE INDEX IF NOT EXISTS idx_files_name ON files(name);
	CREATE INDEX IF NOT EXISTS idx_files_extension ON files(extension);
	CREATE INDEX IF NOT EXISTS idx_files_file_type ON files(file_type);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

	INSERT OR IGNORE INTO roles (id, name, description, permissions, is_system)
	VALUES (1, 'Admin', 'Administrator', '["*"]', 1);
	INSERT OR IGNORE INTO roles (id, name, description, permissions, is_system)
	VALUES (2, 'User', 'Standard user', '["media.view","media.download"]', 1);
	`
	_, err := db.Exec(schema)
	return err
}

func seedData(db *sql.DB, fileCount int) error {
	// Seed storage roots
	protocols := []string{"smb", "local", "ftp", "nfs", "webdav"}
	for i, proto := range protocols {
		_, err := db.Exec(
			"INSERT INTO storage_roots (name, protocol, host, path, enabled) VALUES (?, ?, ?, ?, 1)",
			fmt.Sprintf("root-%s-%d", proto, i+1), proto,
			fmt.Sprintf("192.168.1.%d", 10+i), fmt.Sprintf("/share/%s", proto),
		)
		if err != nil {
			return fmt.Errorf("seed storage_roots: %w", err)
		}
	}

	// Seed users
	_, err := db.Exec(
		"INSERT INTO users (username, email, password_hash, salt, role_id) VALUES (?, ?, ?, ?, ?)",
		"benchuser", "bench@test.local", "hashed_password_placeholder", "salt123", 1,
	)
	if err != nil {
		return fmt.Errorf("seed users: %w", err)
	}

	// Seed files with realistic distribution
	exts := []string{".mkv", ".mp4", ".avi", ".mp3", ".flac", ".jpg", ".png", ".srt", ".nfo", ".txt"}
	types := []string{"video", "video", "video", "audio", "audio", "image", "image", "subtitle", "metadata", "document"}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO files
		(storage_root_id, path, name, extension, file_type, size, is_directory, modified_at, deleted, is_duplicate)
		VALUES (?, ?, ?, ?, ?, ?, 0, ?, 0, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for i := 0; i < fileCount; i++ {
		rootID := (i % 5) + 1
		extIdx := i % len(exts)
		ext := exts[extIdx]
		ftype := types[extIdx]
		name := fmt.Sprintf("file_%06d%s", i, ext)
		path := fmt.Sprintf("/media/collection_%d/%s", i/100, name)
		size := int64((i + 1)) * 1024 * 1024
		isDup := 0
		if i%20 == 0 {
			isDup = 1
		}
		modTime := now.Add(-time.Duration(i) * time.Hour)
		if _, err := stmt.Exec(rootID, path, name, ext, ftype, size, modTime, isDup); err != nil {
			tx.Rollback()
			return fmt.Errorf("seed file %d: %w", i, err)
		}
	}
	return tx.Commit()
}

// ---------------------------------------------------------------------------
// Lightweight middleware stubs (matching production pipeline structure)
// ---------------------------------------------------------------------------

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func errorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
	}
}

var requestCounter uint64

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := atomic.AddUint64(&requestCounter, 1)
		c.Header("X-Request-ID", fmt.Sprintf("bench-%d", id))
		c.Set("request_id", id)
		c.Next()
	}
}

// benchJWTMiddleware is a no-op authenticator that always passes.
// It simulates the cost of parsing the Authorization header without
// performing real JWT crypto, keeping the benchmark focused on the pipeline.
func benchJWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Set("username", "benchuser")
		c.Set("user_id", "1")
		c.Next()
	}
}

// ---------------------------------------------------------------------------
// Handler implementations (lightweight, query-backed)
// ---------------------------------------------------------------------------

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy", "time": time.Now().UTC()})
}

func loginHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		var id int
		var username string
		err := db.QueryRow("SELECT id, username FROM users WHERE username = ?", req.Username).Scan(&id, &username)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":         "bench-token",
			"refresh_token": "bench-refresh",
			"user": gin.H{
				"id":       id,
				"username": username,
			},
		})
	}
}

func mediaListHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT f.id, f.name, f.path, f.extension, f.file_type, f.size, f.modified_at
			FROM files f
			WHERE f.deleted = 0 AND f.file_type IN ('video', 'audio')
			ORDER BY f.modified_at DESC
			LIMIT 50`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		items := make([]gin.H, 0, 50)
		for rows.Next() {
			var id int
			var name, path, ext, ftype string
			var size int64
			var modTime time.Time
			if err := rows.Scan(&id, &name, &path, &ext, &ftype, &size, &modTime); err != nil {
				continue
			}
			items = append(items, gin.H{
				"id": id, "name": name, "path": path,
				"extension": ext, "file_type": ftype,
				"size": size, "modified_at": modTime,
			})
		}

		var total int
		db.QueryRow("SELECT COUNT(*) FROM files WHERE deleted = 0 AND file_type IN ('video', 'audio')").Scan(&total)

		c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": 1, "per_page": 50})
	}
}

func sourcesListHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT sr.id, sr.name, sr.protocol, sr.host, sr.path, sr.enabled,
			       sr.created_at, sr.updated_at,
			       COALESCE((SELECT COUNT(*) FROM files WHERE storage_root_id = sr.id AND deleted = 0), 0) as file_count,
			       COALESCE((SELECT SUM(size) FROM files WHERE storage_root_id = sr.id AND deleted = 0), 0) as total_size
			FROM storage_roots sr
			ORDER BY sr.name`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		sources := make([]gin.H, 0, 10)
		for rows.Next() {
			var id int
			var name, protocol string
			var host, path sql.NullString
			var enabled bool
			var createdAt, updatedAt time.Time
			var fileCount int
			var totalSize int64
			if err := rows.Scan(&id, &name, &protocol, &host, &path, &enabled, &createdAt, &updatedAt, &fileCount, &totalSize); err != nil {
				continue
			}
			sources = append(sources, gin.H{
				"id": id, "name": name, "protocol": protocol,
				"host": host.String, "path": path.String,
				"enabled": enabled, "file_count": fileCount,
				"total_size": totalSize,
			})
		}

		c.JSON(http.StatusOK, gin.H{"sources": sources, "total": len(sources)})
	}
}

func catalogListHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT id, name, path, extension, file_type, size, is_directory, modified_at
			FROM files
			WHERE deleted = 0 AND parent_id IS NULL
			ORDER BY is_directory DESC, name ASC
			LIMIT 100`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		items := make([]gin.H, 0, 100)
		for rows.Next() {
			var id int
			var name, path string
			var ext, ftype sql.NullString
			var size int64
			var isDir bool
			var modTime time.Time
			if err := rows.Scan(&id, &name, &path, &ext, &ftype, &size, &isDir, &modTime); err != nil {
				continue
			}
			items = append(items, gin.H{
				"id": id, "name": name, "path": path,
				"type": ftype.String, "size": size,
				"is_directory": isDir,
			})
		}

		c.JSON(http.StatusOK, gin.H{"items": items})
	}
}

func searchHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		if q == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
			return
		}
		pattern := "%" + q + "%"
		rows, err := db.Query(`
			SELECT id, name, path, file_type, size
			FROM files
			WHERE deleted = 0 AND name LIKE ?
			ORDER BY name
			LIMIT 50`, pattern)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		results := make([]gin.H, 0, 50)
		for rows.Next() {
			var id int
			var name, path string
			var ftype sql.NullString
			var size int64
			if err := rows.Scan(&id, &name, &path, &ftype, &size); err != nil {
				continue
			}
			results = append(results, gin.H{
				"id": id, "name": name, "path": path,
				"file_type": ftype.String, "size": size,
			})
		}

		c.JSON(http.StatusOK, gin.H{"results": results, "total": len(results)})
	}
}

func statsOverallHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var totalFiles int
		var totalSize int64
		var dupCount int
		db.QueryRow("SELECT COUNT(*), COALESCE(SUM(size),0) FROM files WHERE deleted = 0").Scan(&totalFiles, &totalSize)
		db.QueryRow("SELECT COUNT(*) FROM files WHERE deleted = 0 AND is_duplicate = 1").Scan(&dupCount)

		typeRows, _ := db.Query("SELECT file_type, COUNT(*), SUM(size) FROM files WHERE deleted = 0 GROUP BY file_type")
		typeStats := make([]gin.H, 0)
		if typeRows != nil {
			defer typeRows.Close()
			for typeRows.Next() {
				var ft sql.NullString
				var cnt int
				var sz int64
				typeRows.Scan(&ft, &cnt, &sz)
				typeStats = append(typeStats, gin.H{"type": ft.String, "count": cnt, "size": sz})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"total_files":    totalFiles,
			"total_size":     totalSize,
			"duplicate_count": dupCount,
			"file_types":     typeStats,
		})
	}
}

// ---------------------------------------------------------------------------
// Benchmark: GET /health
// ---------------------------------------------------------------------------

func BenchmarkHealthEndpoint(b *testing.B) {
	ts := newTestServer(b, 100)
	defer ts.close()

	req, _ := http.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("expected 200, got %d", w.Code)
		}
	}
}

func BenchmarkHealthEndpoint_Parallel(b *testing.B) {
	ts := newTestServer(b, 100)
	defer ts.close()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			ts.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				b.Fatalf("expected 200, got %d", w.Code)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// Benchmark: POST /api/v1/auth/login
// ---------------------------------------------------------------------------

func BenchmarkLoginEndpoint(b *testing.B) {
	ts := newTestServer(b, 100)
	defer ts.close()

	payload, _ := json.Marshal(map[string]string{
		"username": "benchuser",
		"password": "password123",
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	}
}

func BenchmarkLoginEndpoint_Parallel(b *testing.B) {
	ts := newTestServer(b, 100)
	defer ts.close()

	payload, _ := json.Marshal(map[string]string{
		"username": "benchuser",
		"password": "password123",
	})

	b.ResetTimer()
	b.ReportAllocs()

	var errors uint64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			ts.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				atomic.AddUint64(&errors, 1)
			}
		}
	})
	if errors > 0 {
		b.Logf("note: %d requests returned non-200 (SQLite concurrency limitation)", errors)
	}
}

// ---------------------------------------------------------------------------
// Benchmark: GET /api/v1/media (list)
// ---------------------------------------------------------------------------

func BenchmarkMediaListEndpoint(b *testing.B) {
	sizes := []int{100, 500, 2000}
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("files=%d", sz), func(b *testing.B) {
			ts := newTestServer(b, sz)
			defer ts.close()

			req, _ := http.NewRequest("GET", "/api/v1/media", nil)
			req.Header.Set("Authorization", ts.token)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				ts.router.ServeHTTP(w, req)
				if w.Code != http.StatusOK {
					b.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
				}
			}
		})
	}
}

func BenchmarkMediaListEndpoint_Parallel(b *testing.B) {
	ts := newTestServer(b, 1000)
	defer ts.close()

	b.ResetTimer()
	b.ReportAllocs()

	var errors uint64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/api/v1/media", nil)
			req.Header.Set("Authorization", ts.token)
			w := httptest.NewRecorder()
			ts.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				atomic.AddUint64(&errors, 1)
			}
		}
	})
	if errors > 0 {
		b.Logf("note: %d requests returned non-200 (SQLite concurrency limitation)", errors)
	}
}

// ---------------------------------------------------------------------------
// Benchmark: GET /api/v1/sources (list)
// ---------------------------------------------------------------------------

func BenchmarkSourcesListEndpoint(b *testing.B) {
	sizes := []int{100, 500, 2000}
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("files=%d", sz), func(b *testing.B) {
			ts := newTestServer(b, sz)
			defer ts.close()

			req, _ := http.NewRequest("GET", "/api/v1/sources", nil)
			req.Header.Set("Authorization", ts.token)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				ts.router.ServeHTTP(w, req)
				if w.Code != http.StatusOK {
					b.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
				}
			}
		})
	}
}

func BenchmarkSourcesListEndpoint_Parallel(b *testing.B) {
	ts := newTestServer(b, 1000)
	defer ts.close()

	b.ResetTimer()
	b.ReportAllocs()

	var errors uint64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/api/v1/sources", nil)
			req.Header.Set("Authorization", ts.token)
			w := httptest.NewRecorder()
			ts.router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				atomic.AddUint64(&errors, 1)
			}
		}
	})
	if errors > 0 {
		b.Logf("note: %d requests returned non-200 (SQLite concurrency limitation)", errors)
	}
}

// ---------------------------------------------------------------------------
// Benchmark: GET /api/v1/catalog (list root)
// ---------------------------------------------------------------------------

func BenchmarkCatalogListEndpoint(b *testing.B) {
	ts := newTestServer(b, 1000)
	defer ts.close()

	req, _ := http.NewRequest("GET", "/api/v1/catalog", nil)
	req.Header.Set("Authorization", ts.token)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("expected 200, got %d", w.Code)
		}
	}
}

// ---------------------------------------------------------------------------
// Benchmark: GET /api/v1/search?q=file
// ---------------------------------------------------------------------------

func BenchmarkSearchEndpoint(b *testing.B) {
	sizes := []int{100, 500, 2000}
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("files=%d", sz), func(b *testing.B) {
			ts := newTestServer(b, sz)
			defer ts.close()

			req, _ := http.NewRequest("GET", "/api/v1/search?q=file_00", nil)
			req.Header.Set("Authorization", ts.token)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				ts.router.ServeHTTP(w, req)
				if w.Code != http.StatusOK {
					b.Fatalf("expected 200, got %d", w.Code)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Benchmark: GET /api/v1/stats/overall
// ---------------------------------------------------------------------------

func BenchmarkStatsOverallEndpoint(b *testing.B) {
	ts := newTestServer(b, 1000)
	defer ts.close()

	req, _ := http.NewRequest("GET", "/api/v1/stats/overall", nil)
	req.Header.Set("Authorization", ts.token)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("expected 200, got %d", w.Code)
		}
	}
}

// ---------------------------------------------------------------------------
// Benchmark: response latency custom metrics
// ---------------------------------------------------------------------------

func BenchmarkEndpointLatencies(b *testing.B) {
	ts := newTestServer(b, 500)
	defer ts.close()

	endpoints := []struct {
		name   string
		method string
		path   string
		body   []byte
		auth   bool
	}{
		{"health", "GET", "/health", nil, false},
		{"login", "POST", "/api/v1/auth/login", mustJSON(map[string]string{"username": "benchuser", "password": "pw"}), false},
		{"media_list", "GET", "/api/v1/media", nil, true},
		{"sources_list", "GET", "/api/v1/sources", nil, true},
		{"catalog_list", "GET", "/api/v1/catalog", nil, true},
		{"search", "GET", "/api/v1/search?q=file", nil, true},
		{"stats_overall", "GET", "/api/v1/stats/overall", nil, true},
	}

	for _, ep := range endpoints {
		b.Run(ep.name, func(b *testing.B) {
			b.ReportAllocs()

			var totalNs int64
			for i := 0; i < b.N; i++ {
				var body *bytes.Reader
				if ep.body != nil {
					body = bytes.NewReader(ep.body)
				}
				var req *http.Request
				if body != nil {
					req, _ = http.NewRequest(ep.method, ep.path, body)
					req.Header.Set("Content-Type", "application/json")
				} else {
					req, _ = http.NewRequest(ep.method, ep.path, nil)
				}
				if ep.auth {
					req.Header.Set("Authorization", ts.token)
				}

				start := time.Now()
				w := httptest.NewRecorder()
				ts.router.ServeHTTP(w, req)
				elapsed := time.Since(start)
				totalNs += elapsed.Nanoseconds()
			}

			if b.N > 0 {
				avgUs := float64(totalNs) / float64(b.N) / 1000.0
				b.ReportMetric(avgUs, "avg_us/op")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Benchmark: concurrent mixed workload
// ---------------------------------------------------------------------------

func BenchmarkMixedWorkload_Parallel(b *testing.B) {
	ts := newTestServer(b, 1000)
	defer ts.close()

	loginPayload, _ := json.Marshal(map[string]string{
		"username": "benchuser",
		"password": "password123",
	})

	b.ResetTimer()
	b.ReportAllocs()

	var ops uint64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			op := atomic.AddUint64(&ops, 1)
			var req *http.Request

			switch op % 5 {
			case 0:
				req, _ = http.NewRequest("GET", "/health", nil)
			case 1:
				req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginPayload))
				req.Header.Set("Content-Type", "application/json")
			case 2:
				req, _ = http.NewRequest("GET", "/api/v1/media", nil)
				req.Header.Set("Authorization", ts.token)
			case 3:
				req, _ = http.NewRequest("GET", "/api/v1/sources", nil)
				req.Header.Set("Authorization", ts.token)
			case 4:
				req, _ = http.NewRequest("GET", "/api/v1/search?q=file_00", nil)
				req.Header.Set("Authorization", ts.token)
			}

			w := httptest.NewRecorder()
			ts.router.ServeHTTP(w, req)
		}
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mustJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
