package services

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	_ "github.com/mattn/go-sqlite3"
)

type CatalogServiceTestSuite struct {
	suite.Suite
	db      *sql.DB
	service *CatalogService
	logger  *zap.Logger
}

func (suite *CatalogServiceTestSuite) SetupTest() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	suite.logger = logger

	// Initialize in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	suite.Require().NoError(err)
	suite.db = db

	// Create tables
	suite.setupDatabase()

	// Initialize service
	suite.service = NewCatalogService(nil, logger)
	suite.service.SetDB(db)
}

func (suite *CatalogServiceTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *CatalogServiceTestSuite) setupDatabase() {
	// Create test tables
	_, err := suite.db.Exec(`
		CREATE TABLE IF NOT EXISTS media_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			size INTEGER,
			media_type TEXT,
			file_hash TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS directories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT UNIQUE NOT NULL,
			parent_path TEXT,
			total_size INTEGER DEFAULT 0,
			file_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	suite.Require().NoError(err)

	// Insert test data
	_, err = suite.db.Exec(`
		INSERT INTO media_items (name, path, size, media_type, file_hash) VALUES
		('movie1.mp4', '/media/movies/movie1.mp4', 1000000, 'movie', 'hash1'),
		('movie2.mkv', '/media/movies/movie2.mkv', 2000000, 'movie', 'hash2'),
		('song1.mp3', '/media/music/song1.mp3', 5000000, 'music', 'hash3'),
		('game1.iso', '/media/games/game1.iso', 50000000, 'game', 'hash4');
	`)
	suite.Require().NoError(err)

	_, err = suite.db.Exec(`
		INSERT INTO directories (path, parent_path, total_size, file_count) VALUES
		('/media', NULL, 56500000, 4),
		('/media/movies', '/media', 3000000, 2),
		('/media/music', '/media', 5000000, 1),
		('/media/games', '/media', 50000000, 1);
	`)
	suite.Require().NoError(err)
}

func (suite *CatalogServiceTestSuite) TestNewCatalogService() {
	service := NewCatalogService(nil, suite.logger)
	assert.NotNil(suite.T(), service)
	assert.NotNil(suite.T(), service.logger)
}

func (suite *CatalogServiceTestSuite) TestSetDB() {
	service := NewCatalogService(nil, suite.logger)
	assert.Nil(suite.T(), service.db)

	service.SetDB(suite.db)
	assert.NotNil(suite.T(), service.db)
}

func (suite *CatalogServiceTestSuite) TestListDirectory() {
	// Test root directory
	items, err := suite.service.ListDirectory("/")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), items, 1)
	assert.Equal(suite.T(), "media", items[0].Name)
	assert.Equal(suite.T(), "directory", items[0].Type)

	// Test media directory
	items, err = suite.service.ListDirectory("/media")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), items, 3) // movies, music, games

	// Test movies subdirectory
	items, err = suite.service.ListDirectory("/media/movies")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), items, 2) // movie1.mp4, movie2.mkv
	assert.Equal(suite.T(), "movie1.mp4", items[0].Name)
	assert.Equal(suite.T(), "file", items[0].Type)
}

func (suite *CatalogServiceTestSuite) TestGetFileInfo() {
	// Test existing file
	info, err := suite.service.GetFileInfo("/media/movies/movie1.mp4")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), info)
	assert.Equal(suite.T(), "movie1.mp4", info.Name)
	assert.Equal(suite.T(), int64(1000000), info.Size)
	assert.Equal(suite.T(), "movie", info.MediaType)

	// Test non-existing file
	info, err = suite.service.GetFileInfo("/nonexistent/file.mp4")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), info)
}

func (suite *CatalogServiceTestSuite) TestSearch() {
	// Search for movies
	results, err := suite.service.Search("movie", "", 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 2)
	assert.Equal(suite.T(), "movie1.mp4", results[0].Name)
	assert.Equal(suite.T(), "movie2.mkv", results[1].Name)

	// Search for specific file
	results, err = suite.service.Search("song1", "", 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 1)
	assert.Equal(suite.T(), "song1.mp3", results[0].Name)

	// Search with type filter
	results, err = suite.service.Search("game", "game", 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 1)
	assert.Equal(suite.T(), "game1.iso", results[0].Name)

	// Search non-existing
	results, err = suite.service.Search("nonexistent", "", 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 0)
}

func (suite *CatalogServiceTestSuite) TestSearchDuplicates() {
	// Add duplicate file
	_, err := suite.db.Exec(`
		INSERT INTO media_items (name, path, size, media_type, file_hash)
		VALUES ('duplicate.mp4', '/media/movies/duplicate.mp4', 1000000, 'movie', 'hash1')
	`)
	suite.Require().NoError(err)

	duplicates, err := suite.service.SearchDuplicates()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), duplicates, 1)
	assert.Len(suite.T(), duplicates[0].Files, 2) // original + duplicate
	assert.Equal(suite.T(), "hash1", duplicates[0].Hash)
}

func (suite *CatalogServiceTestSuite) TestGetDirectoriesBySize() {
	dirs, err := suite.service.GetDirectoriesBySize(10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), dirs, 3) // Should return top 3 directories by size

	// Check ordering (largest first)
	assert.Equal(suite.T(), "/media/games", dirs[0].Path)
	assert.Equal(suite.T(), int64(50000000), dirs[0].TotalSize)
	assert.Equal(suite.T(), "/media/music", dirs[1].Path)
	assert.Equal(suite.T(), "/media/movies", dirs[2].Path)
}

func (suite *CatalogServiceTestSuite) TestGetDuplicatesCount() {
	// Initially no duplicates
	count, err := suite.service.GetDuplicatesCount()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)

	// Add duplicate
	_, err = suite.db.Exec(`
		INSERT INTO media_items (name, path, size, media_type, file_hash)
		VALUES ('duplicate.mp4', '/media/movies/duplicate.mp4', 1000000, 'movie', 'hash1')
	`)
	suite.Require().NoError(err)

	count, err = suite.service.GetDuplicatesCount()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, count)
}

func (suite *CatalogServiceTestSuite) TestPagination() {
	// Add more test data for pagination testing
	for i := 3; i <= 15; i++ {
		_, err := suite.db.Exec(`
			INSERT INTO media_items (name, path, size, media_type, file_hash)
			VALUES (?, ?, ?, 'movie', ?)
		`, "movie"+string(rune(i+'0'))+".mp4", "/media/movies/movie"+string(rune(i+'0'))+".mp4", 1000000, "hash"+string(rune(i+'0')))
		suite.Require().NoError(err)
	}

	// Test pagination - page 1 (limit 5)
	results, err := suite.service.Search("movie", "", 5, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 5)

	// Test pagination - page 2 (limit 5, offset 5)
	results, err = suite.service.Search("movie", "", 5, 5)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 5)

	// Test pagination - page 3 (limit 5, offset 10)
	results, err = suite.service.Search("movie", "", 5, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 3) // Should have remaining 3 items
}

func (suite *CatalogServiceTestSuite) TestErrorHandling() {
	// Test with closed database
	suite.db.Close()
	suite.service.SetDB(suite.db)

	_, err := suite.service.ListDirectory("/")
	assert.Error(suite.T(), err)

	_, err = suite.service.GetFileInfo("/test")
	assert.Error(suite.T(), err)

	_, err = suite.service.Search("test", "", 10, 0)
	assert.Error(suite.T(), err)
}

func TestCatalogServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CatalogServiceTestSuite))
}