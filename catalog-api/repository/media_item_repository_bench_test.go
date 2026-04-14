package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/internal/media/models"

	_ "github.com/mutecomm/go-sqlcipher"
)

// setupBenchDB creates an in-memory SQLite database with the media_items
// table and seeds it with test data. Unlike the mock-based unit tests,
// these benchmarks use a real SQLite engine to measure actual query
// execution and scan overhead.
func setupBenchDB(b *testing.B, rowCount int) *MediaItemRepository {
	b.Helper()

	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("failed to open test db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	schema := `
		CREATE TABLE IF NOT EXISTS media_types (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			detection_patterns TEXT,
			metadata_providers TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		INSERT INTO media_types (id, name, description) VALUES
			(1, 'movie', 'Movie'),
			(2, 'tv_show', 'TV Show'),
			(3, 'music_album', 'Music Album'),
			(4, 'game', 'Game');

		CREATE TABLE IF NOT EXISTS media_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_type_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			original_title TEXT,
			year INTEGER,
			description TEXT,
			genre TEXT,
			director TEXT,
			cast_crew TEXT,
			rating REAL,
			runtime INTEGER,
			language TEXT,
			country TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			parent_id INTEGER,
			season_number INTEGER,
			episode_number INTEGER,
			track_number INTEGER,
			first_detected DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (media_type_id) REFERENCES media_types(id)
		);

		CREATE INDEX IF NOT EXISTS idx_media_items_title ON media_items(title);
		CREATE INDEX IF NOT EXISTS idx_media_items_type ON media_items(media_type_id);
	`

	_, err = sqlDB.Exec(schema)
	if err != nil {
		b.Fatalf("failed to create schema: %v", err)
	}

	// Seed data
	now := time.Now()
	tx, err := sqlDB.Begin()
	if err != nil {
		b.Fatalf("failed to begin tx: %v", err)
	}
	stmt, err := tx.Prepare(`INSERT INTO media_items
		(media_type_id, title, original_title, year, description, genre, director,
		 rating, runtime, language, country, status, first_detected, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		b.Fatalf("failed to prepare insert: %v", err)
	}

	titles := []string{
		"The Matrix", "Inception", "Interstellar", "The Dark Knight",
		"Pulp Fiction", "Fight Club", "Forrest Gump", "The Shawshank Redemption",
		"The Godfather", "Blade Runner", "Alien", "Terminator",
		"Jurassic Park", "Back to the Future", "Gladiator", "Avatar",
	}
	genres := []string{
		`["Action","Sci-Fi"]`, `["Action","Thriller"]`, `["Drama","Sci-Fi"]`,
		`["Action","Crime"]`, `["Crime","Drama"]`, `["Drama"]`,
	}
	directors := []string{
		"Wachowskis", "Christopher Nolan", "Ridley Scott", "Steven Spielberg",
	}
	languages := []string{"en", "fr", "de", "ja"}
	countries := []string{"US", "UK", "FR", "JP"}

	for i := 0; i < rowCount; i++ {
		title := fmt.Sprintf("%s %d", titles[i%len(titles)], i)
		year := 1990 + (i % 35)
		typeID := (i % 4) + 1
		desc := fmt.Sprintf("Description for %s", title)
		_, err := stmt.Exec(
			typeID, title, title, year, desc,
			genres[i%len(genres)], directors[i%len(directors)],
			float64(5+(i%5)), 90+(i%60),
			languages[i%len(languages)], countries[i%len(countries)],
			"active", now, now,
		)
		if err != nil {
			b.Fatalf("failed to insert row %d: %v", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		b.Fatalf("failed to commit seed data: %v", err)
	}

	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewMediaItemRepository(db)
}

// BenchmarkSearch measures the cost of a full-text LIKE search across
// title, original_title, and description columns. This is the primary
// user-facing search path and exercises query building, SQL execution,
// and row scanning for a realistic result set.
func BenchmarkSearch(b *testing.B) {
	repo := setupBenchDB(b, 500)
	ctx := context.Background()
	queries := []string{"Matrix", "Dark", "Fiction", "Runner", "Alien"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := repo.Search(ctx, queries[i%len(queries)], nil, 20, 0)
		if err != nil {
			b.Fatalf("Search failed: %v", err)
		}
	}
}

// BenchmarkSearch_WithTypeFilter measures search performance when
// filtering by media type, which adds an IN clause to the query.
func BenchmarkSearch_WithTypeFilter(b *testing.B) {
	repo := setupBenchDB(b, 500)
	ctx := context.Background()
	typeFilter := []int64{1, 2} // movie + tv_show

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := repo.Search(ctx, "Matrix", typeFilter, 20, 0)
		if err != nil {
			b.Fatalf("Search with type filter failed: %v", err)
		}
	}
}

// BenchmarkGetByID measures single-item lookup by primary key, which
// is the most frequent database operation for item detail views.
func BenchmarkGetByID(b *testing.B) {
	repo := setupBenchDB(b, 500)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := int64((i % 500) + 1)
		_, err := repo.GetByID(ctx, id)
		if err != nil {
			b.Fatalf("GetByID(%d) failed: %v", id, err)
		}
	}
}

// BenchmarkGetByType measures paginated retrieval by media type, which
// runs both a COUNT and a SELECT query.
func BenchmarkGetByType(b *testing.B) {
	repo := setupBenchDB(b, 500)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		typeID := int64((i % 4) + 1)
		_, _, err := repo.GetByType(ctx, typeID, 20, 0)
		if err != nil {
			b.Fatalf("GetByType(%d) failed: %v", typeID, err)
		}
	}
}

// BenchmarkCreate measures the cost of inserting a single media item,
// including JSON marshaling of genre and cast_crew fields.
func BenchmarkCreate(b *testing.B) {
	repo := setupBenchDB(b, 0)
	ctx := context.Background()
	year := 2024
	desc := "A benchmark movie"
	rating := 8.5
	runtime := 120
	lang := "en"
	country := "US"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		item := &models.MediaItem{
			MediaTypeID: 1,
			Title:       fmt.Sprintf("Bench Movie %d", i),
			Year:        &year,
			Description: &desc,
			Genre:       []string{"Action", "Sci-Fi"},
			Rating:      &rating,
			Runtime:     &runtime,
			Language:    &lang,
			Country:     &country,
			Status:      "active",
		}
		_, err := repo.Create(ctx, item)
		if err != nil {
			b.Fatalf("Create failed: %v", err)
		}
	}
}

// BenchmarkCount measures the cost of counting all media items, which
// is used for dashboard statistics.
func BenchmarkCount(b *testing.B) {
	repo := setupBenchDB(b, 500)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := repo.Count(ctx)
		if err != nil {
			b.Fatalf("Count failed: %v", err)
		}
	}
}

// BenchmarkGetByTitle measures lookup by exact title match and media
// type, which is used during aggregation to detect existing entities.
func BenchmarkGetByTitle(b *testing.B) {
	repo := setupBenchDB(b, 500)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetByTitle(ctx, fmt.Sprintf("The Matrix %d", i%500), 1)
		if err != nil {
			b.Fatalf("GetByTitle failed: %v", err)
		}
	}
}
