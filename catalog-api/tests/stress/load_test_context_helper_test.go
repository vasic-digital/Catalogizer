package stress

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"catalogizer/internal/tests"
	"catalogizer/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// HTTP STRESS TEST HARNESS
//
// setupStressTestServer stands up a real, in-process HTTP server (via
// httptest.NewServer) backed by an in-memory SQLite database. It wires the same
// production middleware stack (request ID, security headers, timeout,
// concurrency limiter) used by the application and registers the endpoints that
// the concurrent-handler stress tests exercise so they hit a genuinely working
// server rather than a stub.
//
// The companion loadTestContext drives that server: it logs in to obtain a
// bearer token, issues authenticated requests, and records per-request latency
// and success/failure counts so the stress tests can assert on p95 latency and
// success rate.
// =============================================================================

// stressTestServer bundles the running HTTP test server with its backing
// database so tests can both make HTTP calls and (if needed) inspect state.
type stressTestServer struct {
	*httptest.Server
	DB *sql.DB
}

// setupStressTestServer builds and starts an HTTP test server with the full
// middleware chain and the endpoints exercised by the concurrent handler stress
// tests. Cleanup (server shutdown and DB close) is registered with t.
func setupStressTestServer(t *testing.T) *stressTestServer {
	t.Helper()

	db := tests.SetupTestDB(t)
	t.Cleanup(func() { db.Close() })

	seedStressTestData(t, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Production middleware stack — the same one main.go installs.
	router.Use(middleware.RequestID())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RequestTimeout(30 * time.Second))
	router.Use(middleware.ConcurrencyLimiter(200))

	// In-memory token store standing in for the JWT auth layer. A successful
	// login mints a token; protected reads/writes require it.
	validTokens := &sync.Map{}

	requireAuth := func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "missing or invalid authorization"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if _, ok := validTokens.Load(token); !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": "invalid token"})
			return
		}
		c.Next()
	}

	// Health check — unauthenticated, used by the health-check load test.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Login — issues a bearer token for the seeded user.
	router.POST("/auth/login", func(c *gin.Context) {
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if body.Username == stressTestUser && body.Password == stressTestPass {
			token := fmt.Sprintf("stress-token-%d", time.Now().UnixNano())
			validTokens.Store(token, true)
			c.JSON(http.StatusOK, gin.H{
				"token": token,
				"user":  gin.H{"id": 1, "username": stressTestUser},
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	})

	// Paginated media listing — reads real rows from the database.
	router.GET("/media", requireAuth, func(c *gin.Context) {
		page := parsePositiveInt(c.Query("page"), 1)
		limit := parsePositiveInt(c.Query("limit"), 10)
		offset := (page - 1) * limit

		rows, err := db.Query(
			"SELECT id, title, type FROM media_items ORDER BY id LIMIT ? OFFSET ?",
			limit, offset,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
			return
		}
		defer rows.Close()

		items := make([]gin.H, 0, limit)
		for rows.Next() {
			var id int
			var title, mediaType string
			if scanErr := rows.Scan(&id, &title, &mediaType); scanErr != nil {
				continue
			}
			items = append(items, gin.H{"id": id, "title": title, "type": mediaType})
		}
		if rows.Err() != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"items": items,
			"page":  page,
			"limit": limit,
		})
	})

	// Storage roots listing.
	router.GET("/storage/roots", requireAuth, func(c *gin.Context) {
		rows, err := db.Query(
			"SELECT id, name, protocol FROM storage_roots ORDER BY id",
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
			return
		}
		defer rows.Close()

		roots := make([]gin.H, 0)
		for rows.Next() {
			var id int
			var name, protocol string
			if scanErr := rows.Scan(&id, &name, &protocol); scanErr != nil {
				continue
			}
			roots = append(roots, gin.H{"id": id, "name": name, "protocol": protocol})
		}
		if rows.Err() != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"roots": roots})
	})

	// Analytics dashboard — aggregates real counts from the database.
	router.GET("/analytics/dashboard", requireAuth, func(c *gin.Context) {
		var mediaCount, eventCount int
		if err := db.QueryRow("SELECT COUNT(*) FROM media_items").Scan(&mediaCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
			return
		}
		if err := db.QueryRow("SELECT COUNT(*) FROM analytics_events").Scan(&eventCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"media_count": mediaCount,
			"event_count": eventCount,
		})
	})

	// Collections listing.
	router.GET("/collections", requireAuth, func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name FROM collections ORDER BY id")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
			return
		}
		defer rows.Close()

		collections := make([]gin.H, 0)
		for rows.Next() {
			var id int
			var name string
			if scanErr := rows.Scan(&id, &name); scanErr != nil {
				continue
			}
			collections = append(collections, gin.H{"id": id, "name": name})
		}
		if rows.Err() != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"collections": collections})
	})

	// Analytics tracking — a write endpoint that persists an event row.
	router.POST("/analytics/track", requireAuth, func(c *gin.Context) {
		var body struct {
			EventType  string `json:"event_type"`
			EntityType string `json:"entity_type"`
			EntityID   int    `json:"entity_id"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		if _, err := db.Exec(
			"INSERT INTO analytics_events (event_type, entity_type, entity_id) VALUES (?, ?, ?)",
			body.EventType, body.EntityType, body.EntityID,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "tracked"})
	})

	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	return &stressTestServer{Server: ts, DB: db}
}

const (
	stressTestUser = "stress-admin"
	stressTestPass = "stress-pass-123"
)

// seedStressTestData creates the auxiliary tables the HTTP endpoints need
// (collections, analytics_events) and inserts a baseline of rows so list and
// dashboard endpoints return meaningful data under load.
func seedStressTestData(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS collections (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS analytics_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_type TEXT NOT NULL,
		entity_type TEXT NOT NULL,
		entity_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)

	// A storage root for the storage/roots endpoint.
	_, err = db.Exec(
		`INSERT INTO storage_roots (name, protocol, path, enabled)
		 VALUES ('stress-root', 'local', '/stress', 1)`,
	)
	require.NoError(t, err)

	// Media items so /media pagination returns rows.
	for i := 0; i < 50; i++ {
		_, err = db.Exec(
			`INSERT INTO media_items (path, title, type, media_type, storage_root_id)
			 VALUES (?, ?, 'movie', 'movie', 1)`,
			fmt.Sprintf("/stress/item_%d.mkv", i),
			fmt.Sprintf("Stress Item %d", i),
		)
		require.NoError(t, err)
	}

	// A few collections.
	for i := 0; i < 5; i++ {
		_, err = db.Exec(
			"INSERT INTO collections (name) VALUES (?)",
			fmt.Sprintf("Collection %d", i),
		)
		require.NoError(t, err)
	}

	// The user the load test authenticates as.
	_, err = db.Exec(
		`INSERT INTO users (username, email, role_id, is_active)
		 VALUES (?, ?, 1, 1)`,
		stressTestUser, stressTestUser+"@example.com",
	)
	require.NoError(t, err)
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	n := 0
	for _, r := range raw {
		if r < '0' || r > '9' {
			return fallback
		}
		n = n*10 + int(r-'0')
	}
	if n <= 0 {
		return fallback
	}
	return n
}

// =============================================================================
// loadTestContext — authenticated load driver with latency/success metrics
// =============================================================================

type loadTestContext struct {
	baseURL string
	client  *http.Client
	token   string

	requests  int64
	successes int64
	failures  int64
}

// newLoadTestContext creates a load driver targeting the given server base URL.
//
// The HTTP client is given a transport sized for high concurrency. Go's default
// transport caps idle keep-alive connections per host at 2
// (http.DefaultMaxIdleConnsPerHost), so dozens of concurrent goroutines would
// otherwise serialize onto a tiny connection pool and measure connection-wait
// time as if it were server latency. Raising the per-host limits lets the load
// test measure the server's behaviour rather than client-side connection
// starvation.
func newLoadTestContext(baseURL string) *loadTestContext {
	transport := &http.Transport{
		MaxIdleConns:        512,
		MaxIdleConnsPerHost: 512,
		MaxConnsPerHost:     512,
		IdleConnTimeout:     30 * time.Second,
	}
	return &loadTestContext{
		baseURL: baseURL,
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
	}
}

// authenticate logs in against the test server and stores the bearer token used
// for subsequent authenticated requests.
func (ltc *loadTestContext) authenticate(t *testing.T) {
	t.Helper()

	payload := map[string]string{
		"username": stressTestUser,
		"password": stressTestPass,
	}
	bodyBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	resp, err := ltc.client.Post(
		ltc.baseURL+"/auth/login",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "login should succeed")

	var result struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.NotEmpty(t, result.Token, "login should return a token")

	ltc.token = result.Token
}

// makeRequest issues a single authenticated request, records its latency and
// outcome, and returns the response, the measured latency, and any error. A 2xx
// response counts as a success; anything else (or a transport error) counts as a
// failure. The caller is responsible for closing resp.Body.
func (ltc *loadTestContext) makeRequest(
	method, path string, body interface{},
) (*http.Response, time.Duration, error) {
	var reqBody *bytes.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			atomic.AddInt64(&ltc.requests, 1)
			atomic.AddInt64(&ltc.failures, 1)
			return nil, 0, err
		}
		reqBody = bytes.NewReader(bodyBytes)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, ltc.baseURL+path, reqBody)
	if err != nil {
		atomic.AddInt64(&ltc.requests, 1)
		atomic.AddInt64(&ltc.failures, 1)
		return nil, 0, err
	}
	if ltc.token != "" {
		req.Header.Set("Authorization", "Bearer "+ltc.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	resp, err := ltc.client.Do(req)
	latency := time.Since(start)

	atomic.AddInt64(&ltc.requests, 1)
	if err != nil || resp == nil {
		atomic.AddInt64(&ltc.failures, 1)
		return resp, latency, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		atomic.AddInt64(&ltc.successes, 1)
	} else {
		atomic.AddInt64(&ltc.failures, 1)
	}

	return resp, latency, nil
}

// GetStats returns a snapshot of the accumulated request metrics. success_rate
// is a percentage (0-100); requests is the total request count as int64.
func (ltc *loadTestContext) GetStats() map[string]interface{} {
	requests := atomic.LoadInt64(&ltc.requests)
	successes := atomic.LoadInt64(&ltc.successes)
	failures := atomic.LoadInt64(&ltc.failures)

	successRate := 0.0
	if requests > 0 {
		successRate = float64(successes) / float64(requests) * 100
	}

	return map[string]interface{}{
		"requests":     requests,
		"successes":    successes,
		"failures":     failures,
		"success_rate": successRate,
	}
}

// PrintStats logs the current metrics snapshot through the test logger.
func (ltc *loadTestContext) PrintStats(t *testing.T) {
	t.Helper()
	stats := ltc.GetStats()
	t.Logf("Load stats: requests=%d successes=%d failures=%d success_rate=%.2f%%",
		stats["requests"].(int64),
		stats["successes"].(int64),
		stats["failures"].(int64),
		stats["success_rate"].(float64),
	)
}
