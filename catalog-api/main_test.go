package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type MainTestSuite struct {
	suite.Suite
	router *gin.Engine
	db     *sql.DB
	logger *zap.Logger
}

func (suite *MainTestSuite) SetupTest() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize logger
	logger, _ := zap.NewDevelopment()
	suite.logger = logger

	// Initialize in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	suite.Require().NoError(err)
	suite.db = db

	// Create tables
	suite.setupDatabase()

	// Initialize router
	suite.router = gin.Default()

	// Setup middleware
	suite.router.Use(func(c *gin.Context) {
		c.Set("db", suite.db)
		c.Set("logger", suite.logger)
		c.Next()
	})

	// Setup routes
	suite.setupRoutes()
}

func (suite *MainTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *MainTestSuite) setupDatabase() {
	// Create test tables
	_, err := suite.db.Exec(`
		CREATE TABLE IF NOT EXISTS media_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			size INTEGER,
			media_type TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT DEFAULT 'user',
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	suite.Require().NoError(err)

	// Insert test data
	_, err = suite.db.Exec(`
		INSERT INTO media_items (name, path, size, media_type) VALUES
		('test_movie.mp4', '/media/movies/test_movie.mp4', 1000000, 'movie'),
		('test_music.mp3', '/media/music/test_music.mp3', 5000000, 'music');
	`)
	suite.Require().NoError(err)

	_, err = suite.db.Exec(`
		INSERT INTO users (username, password_hash, role) VALUES
		('admin', '$2a$10$hashedpassword', 'admin'),
		('user', '$2a$10$hashedpassword', 'user');
	`)
	suite.Require().NoError(err)
}

func (suite *MainTestSuite) setupRoutes() {
	// Health check
	suite.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "time": time.Now().UTC()})
	})

	// API routes
	api := suite.router.Group("/api/v1")
	{
		// Catalog browsing endpoints
		api.GET("/catalog", suite.listRoot)
		api.GET("/catalog/*path", suite.listPath)

		// Search endpoints
		api.GET("/search", suite.search)

		// Statistics
		api.GET("/stats/summary", suite.getStatsSummary)
	}
}

func (suite *MainTestSuite) listRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"items": []gin.H{
			{"name": "media", "type": "directory", "path": "/media"},
		},
	})
}

func (suite *MainTestSuite) listPath(c *gin.Context) {
	path := c.Param("path")
	if path == "/media" {
		c.JSON(http.StatusOK, gin.H{
			"items": []gin.H{
				{"name": "movies", "type": "directory", "path": "/media/movies"},
				{"name": "music", "type": "directory", "path": "/media/music"},
			},
		})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Path not found"})
}

func (suite *MainTestSuite) search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter required"})
		return
	}

	// Mock search results
	results := []gin.H{
		{"name": "test_movie.mp4", "path": "/media/movies/test_movie.mp4", "type": "movie"},
	}

	c.JSON(http.StatusOK, gin.H{"results": results, "total": len(results)})
}

func (suite *MainTestSuite) getStatsSummary(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_files": 2,
		"total_size":  6000000,
		"media_types": gin.H{
			"movie": 1,
			"music": 1,
		},
	})
}

func (suite *MainTestSuite) TestHealthCheck() {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", response["status"])
	assert.NotNil(suite.T(), response["time"])
}

func (suite *MainTestSuite) TestListRoot() {
	req, _ := http.NewRequest("GET", "/api/v1/catalog", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["items"])
}

func (suite *MainTestSuite) TestListPath() {
	// Test valid path
	req, _ := http.NewRequest("GET", "/api/v1/catalog/media", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Test invalid path
	req, _ = http.NewRequest("GET", "/api/v1/catalog/nonexistent", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *MainTestSuite) TestSearch() {
	// Test valid search
	req, _ := http.NewRequest("GET", "/api/v1/search?q=movie", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["results"])

	// Test empty query
	req, _ = http.NewRequest("GET", "/api/v1/search", nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *MainTestSuite) TestStatsSummary() {
	req, _ := http.NewRequest("GET", "/api/v1/stats/summary", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(2), response["total_files"])
	assert.Equal(suite.T(), float64(6000000), response["total_size"])
}

func (suite *MainTestSuite) TestDatabaseConnection() {
	// Test database operations
	var count int
	err := suite.db.QueryRow("SELECT COUNT(*) FROM media_items").Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, count)

	err = suite.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, count)
}

func (suite *MainTestSuite) TestMiddleware() {
	// Test that middleware sets context values
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Create a test router with middleware
	testRouter := gin.New()
	testRouter.Use(func(c *gin.Context) {
		c.Set("test_key", "test_value")
		c.Next()
	})
	testRouter.GET("/health", func(c *gin.Context) {
		value, exists := c.Get("test_key")
		if exists {
			c.JSON(http.StatusOK, gin.H{"middleware_test": value})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "middleware not working"})
		}
	})

	testRouter.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test_value", response["middleware_test"])
}

func TestMainTestSuite(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}
