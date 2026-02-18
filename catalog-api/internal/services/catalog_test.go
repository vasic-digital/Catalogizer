package services

import (
	"catalogizer/database"
	"database/sql"
	"testing"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type CatalogServiceTestSuite struct {
	suite.Suite
	db      *database.DB
	service *CatalogService
	logger  *zap.Logger
}

func (suite *CatalogServiceTestSuite) SetupTest() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	suite.logger = logger

	// Initialize in-memory database with shared cache
	sqlDB, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	suite.Require().NoError(err)
	suite.db = database.WrapDB(sqlDB, database.DialectSQLite)

	// Create tables
	suite.setupDatabase()

	// Initialize service
	suite.service = NewCatalogService(nil, logger)
	suite.service.SetDB(suite.db)
}

func (suite *CatalogServiceTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *CatalogServiceTestSuite) setupDatabase() {
	// Create test tables
	_, err := suite.db.Exec(`
		CREATE TABLE IF NOT EXISTS storage_roots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			protocol TEXT,
			host TEXT,
			port INTEGER,
			path TEXT,
			username TEXT,
			password TEXT,
			domain TEXT,
			enabled BOOLEAN DEFAULT 1,
			max_depth INTEGER DEFAULT 10,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_scan_at DATETIME
		);

		CREATE TABLE IF NOT EXISTS files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			storage_root_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			is_directory BOOLEAN DEFAULT 0,
			size INTEGER,
			modified_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			quick_hash TEXT,
			extension TEXT,
			mime_type TEXT,
			file_type TEXT,
			parent_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_scan_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted BOOLEAN DEFAULT 0,
			is_duplicate BOOLEAN DEFAULT 0,
			duplicate_group_id INTEGER,
			FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id),
			FOREIGN KEY (parent_id) REFERENCES files(id)
		);

		CREATE TABLE IF NOT EXISTS duplicate_groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_hash TEXT NOT NULL,
			file_count INTEGER DEFAULT 0,
			total_size INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	suite.Require().NoError(err)

	// Insert test storage root
	_, err = suite.db.Exec(`INSERT INTO storage_roots (id, name, protocol) VALUES (1, 'test', 'local')`)
	suite.Require().NoError(err)

	// Insert test data with explicit parent_id
	_, err = suite.db.Exec(`
		INSERT INTO files (id, storage_root_id, name, path, is_directory, size, modified_at, parent_id) VALUES
		(1, 1, 'media', '/media', 1, 0, CURRENT_TIMESTAMP, NULL),
		(2, 1, 'movies', '/media/movies', 1, 0, CURRENT_TIMESTAMP, 1),
		(3, 1, 'music', '/media/music', 1, 0, CURRENT_TIMESTAMP, 1),
		(4, 1, 'games', '/media/games', 1, 0, CURRENT_TIMESTAMP, 1),
		(5, 1, 'movie1.mp4', '/media/movies/movie1.mp4', 0, 1000000, CURRENT_TIMESTAMP, 2),
		(6, 1, 'movie2.mkv', '/media/movies/movie2.mkv', 0, 2000000, CURRENT_TIMESTAMP, 2),
		(7, 1, 'song1.mp3', '/media/music/song1.mp3', 0, 5000000, CURRENT_TIMESTAMP, 3),
		(8, 1, 'game1.iso', '/media/games/game1.iso', 0, 50000000, CURRENT_TIMESTAMP, 4);
	`)
	suite.Require().NoError(err)

	// Insert duplicate groups
	_, err = suite.db.Exec(`
		INSERT INTO duplicate_groups (file_hash, file_count, total_size) VALUES
		('dup_hash', 2, 1000000);
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

	// Test non-existing file
	info, err = suite.service.GetFileInfo("/nonexistent/file.mp4")
	assert.NoError(suite.T(), err)
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

	// Search for specific file by name
	results, err = suite.service.Search("game1", "", 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 1)
	assert.Equal(suite.T(), "game1.iso", results[0].Name)

	// Search non-existing
	results, err = suite.service.Search("nonexistent", "", 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 0)
}

func (suite *CatalogServiceTestSuite) TestSearchDuplicates() {
	// Add duplicate files
	_, err := suite.db.Exec(`
		INSERT INTO files (storage_root_id, name, path, is_directory, size, modified_at, quick_hash, parent_id)
		VALUES (1, 'duplicate1.mp4', '/media/movies/duplicate1.mp4', 0, 1000000, CURRENT_TIMESTAMP, 'hash1',
			(SELECT id FROM files WHERE path = '/media/movies' AND is_directory = 1)),
		(1, 'duplicate2.mp4', '/media/movies/duplicate2.mp4', 0, 1000000, CURRENT_TIMESTAMP, 'hash1',
			(SELECT id FROM files WHERE path = '/media/movies' AND is_directory = 1))
	`)
	suite.Require().NoError(err)

	duplicates, err := suite.service.SearchDuplicates()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), duplicates, 1)
	assert.Len(suite.T(), duplicates[0].Files, 2) // two duplicates
	assert.Equal(suite.T(), "hash1", duplicates[0].Hash)
}

func (suite *CatalogServiceTestSuite) TestGetDirectoriesBySize() {
	dirs, err := suite.service.GetDirectoriesBySize("test", 10)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), len(dirs) > 0, "Should return at least one directory")
}

func (suite *CatalogServiceTestSuite) TestGetDuplicatesCount() {
	// Initially no duplicates
	count, err := suite.service.GetDuplicatesCount()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), count)

	// Add duplicates
	_, err = suite.db.Exec(`
		INSERT INTO files (storage_root_id, name, path, is_directory, size, modified_at, quick_hash, parent_id)
		VALUES (1, 'duplicate1.mp4', '/media/movies/duplicate1.mp4', 0, 1000000, CURRENT_TIMESTAMP, 'hash1',
			(SELECT id FROM files WHERE path = '/media/movies' AND is_directory = 1)),
		(1, 'duplicate2.mp4', '/media/movies/duplicate2.mp4', 0, 1000000, CURRENT_TIMESTAMP, 'hash1',
			(SELECT id FROM files WHERE path = '/media/movies' AND is_directory = 1))
	`)
	suite.Require().NoError(err)

	count, err = suite.service.GetDuplicatesCount()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *CatalogServiceTestSuite) TestPagination() {
	// Add more test data for pagination testing
	for i := 3; i <= 13; i++ {
		_, err := suite.db.Exec(`
			INSERT INTO files (storage_root_id, name, path, is_directory, size, modified_at, quick_hash, parent_id)
			VALUES (1, ?, ?, 0, ?, CURRENT_TIMESTAMP, ?,
				(SELECT id FROM files WHERE path = '/media/movies' AND is_directory = 1))
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
