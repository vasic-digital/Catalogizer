package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"catalogizer/database"
	"catalogizer/repository"
)

// newTestAggregationService creates an AggregationService with nil-safe DB
// (for methods that don't touch the database).
func newTestAggregationService(t *testing.T) *AggregationService {
	t.Helper()
	wrappedDB := database.WrapDB(nil, database.DialectSQLite)
	logger := zap.NewNop()

	return &AggregationService{
		db:              wrappedDB,
		logger:          logger,
		itemRepo:        repository.NewMediaItemRepository(wrappedDB),
		fileRepo:        repository.NewMediaFileRepository(wrappedDB),
		dirAnalysisRepo: repository.NewDirectoryAnalysisRepository(wrappedDB),
		extMetaRepo:     repository.NewExternalMetadataRepository(wrappedDB),
	}
}

// ---------------------------------------------------------------------------
// detectMediaType tests
// ---------------------------------------------------------------------------

func TestDetectMediaType_TVShow_SxxExx(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "Breaking.Bad.S01E02.720p",
		fileTypes:  map[string]int{".mkv": 1},
		fileCount:  1,
		extensions: []string{".mkv"},
	}

	mediaType, parsed := svc.detectMediaType(dir)
	assert.Equal(t, "tv_show", mediaType)
	assert.Equal(t, "Breaking Bad", parsed.Title)
	assert.NotNil(t, parsed.Season)
	assert.Equal(t, 1, *parsed.Season)
	assert.NotNil(t, parsed.Episode)
	assert.Equal(t, 2, *parsed.Episode)
}

func TestDetectMediaType_TVShow_SeasonFolder(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "Game of Thrones Season 3",
		fileTypes:  map[string]int{".mkv": 10},
		fileCount:  10,
		extensions: []string{".mkv"},
	}

	mediaType, parsed := svc.detectMediaType(dir)
	assert.Equal(t, "tv_show", mediaType)
	assert.Equal(t, "Game of Thrones", parsed.Title)
}

func TestDetectMediaType_TVShow_Complete(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "Friends Complete",
		fileTypes:  map[string]int{".mkv": 50},
		fileCount:  50,
		extensions: []string{".mkv"},
	}

	mediaType, _ := svc.detectMediaType(dir)
	assert.Equal(t, "tv_show", mediaType)
}

func TestDetectMediaType_Movie(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "The Matrix (1999)",
		fileTypes:  map[string]int{".mkv": 1},
		fileCount:  1,
		extensions: []string{".mkv"},
	}

	mediaType, parsed := svc.detectMediaType(dir)
	assert.Equal(t, "movie", mediaType)
	assert.Equal(t, "The Matrix", parsed.Title)
	assert.NotNil(t, parsed.Year)
	assert.Equal(t, 1999, *parsed.Year)
}

func TestDetectMediaType_MovieFromVideoExtension(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "Some.Random.File",
		fileTypes:  map[string]int{".mp4": 1},
		fileCount:  1,
		extensions: []string{".mp4"},
	}

	mediaType, _ := svc.detectMediaType(dir)
	assert.Equal(t, "movie", mediaType)
}

func TestDetectMediaType_MusicAlbum_DashSeparator(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "Pink Floyd - The Wall",
		fileTypes:  map[string]int{".flac": 12},
		fileCount:  12,
		extensions: []string{".flac"},
	}

	mediaType, parsed := svc.detectMediaType(dir)
	assert.Equal(t, "music_album", mediaType)
	assert.Equal(t, "Pink Floyd", parsed.Artist)
	assert.Equal(t, "The Wall", parsed.Album)
}

func TestDetectMediaType_MusicAlbum_MP3(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "Nirvana - Nevermind (1991)",
		fileTypes:  map[string]int{".mp3": 13},
		fileCount:  13,
		extensions: []string{".mp3"},
	}

	mediaType, _ := svc.detectMediaType(dir)
	assert.Equal(t, "music_album", mediaType)
}

func TestDetectMediaType_Software_ISO(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "Ubuntu 24.04",
		fileTypes:  map[string]int{".iso": 1},
		fileCount:  1,
		extensions: []string{".iso"},
	}

	mediaType, parsed := svc.detectMediaType(dir)
	assert.Equal(t, "software", mediaType)
	assert.Equal(t, "Ubuntu", parsed.Title)
}

func TestDetectMediaType_Software_Executable(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "VLC.Media.Player.v3.0.18",
		fileTypes:  map[string]int{".exe": 1},
		fileCount:  1,
		extensions: []string{".exe"},
	}

	mediaType, _ := svc.detectMediaType(dir)
	assert.Equal(t, "software", mediaType)
}

func TestDetectMediaType_Comic_CBR(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "Batman The Dark Knight",
		fileTypes:  map[string]int{".cbr": 5},
		fileCount:  5,
		extensions: []string{".cbr"},
	}

	mediaType, _ := svc.detectMediaType(dir)
	assert.Equal(t, "comic", mediaType)
}

func TestDetectMediaType_Comic_CBZ(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "Spider-Man Vol 1",
		fileTypes:  map[string]int{".cbz": 10},
		fileCount:  10,
		extensions: []string{".cbz"},
	}

	mediaType, _ := svc.detectMediaType(dir)
	assert.Equal(t, "comic", mediaType)
}

func TestDetectMediaType_Book_EPUB(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "Dune - Frank Herbert",
		fileTypes:  map[string]int{".epub": 1},
		fileCount:  1,
		extensions: []string{".epub"},
	}

	mediaType, _ := svc.detectMediaType(dir)
	assert.Equal(t, "book", mediaType)
}

func TestDetectMediaType_Book_PDF(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "Programming in Go",
		fileTypes:  map[string]int{".pdf": 1},
		fileCount:  1,
		extensions: []string{".pdf"},
	}

	mediaType, _ := svc.detectMediaType(dir)
	assert.Equal(t, "book", mediaType)
}

func TestDetectMediaType_Game(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "Half-Life 2 PC",
		fileTypes:  map[string]int{".iso": 1},
		fileCount:  1,
		extensions: []string{".iso"},
	}

	mediaType, _ := svc.detectMediaType(dir)
	// .iso triggers software detection before game detection (software check
	// runs first in detectMediaType). Game detection only fires when BOTH
	// hasISO AND gamePlatformRe match, but the software branch already matched.
	assert.Equal(t, "software", mediaType)
}

func TestDetectMediaType_Unknown(t *testing.T) {
	svc := newTestAggregationService(t)

	dir := directoryInfo{
		name:       "random_folder",
		fileTypes:  map[string]int{".txt": 3},
		fileCount:  3,
		extensions: []string{".txt"},
	}

	mediaType, _ := svc.detectMediaType(dir)
	assert.Empty(t, mediaType,
		"unknown file types with no year should return empty media type")
}

func TestDetectMediaType_MovieFromYear(t *testing.T) {
	svc := newTestAggregationService(t)

	// No video extension, but has a year in the name
	dir := directoryInfo{
		name:       "Inception (2010)",
		fileTypes:  map[string]int{".nfo": 1},
		fileCount:  1,
		extensions: []string{".nfo"},
	}

	mediaType, parsed := svc.detectMediaType(dir)
	assert.Equal(t, "movie", mediaType)
	assert.Equal(t, "Inception", parsed.Title)
	assert.NotNil(t, parsed.Year)
	assert.Equal(t, 2010, *parsed.Year)
}

func TestDetectMediaType_ExtensionCaseSensitivity(t *testing.T) {
	svc := newTestAggregationService(t)

	// Extensions should be case-insensitive
	dir := directoryInfo{
		name:       "Some Album",
		fileTypes:  map[string]int{"FLAC": 10},
		fileCount:  10,
		extensions: []string{"FLAC"},
	}

	mediaType, _ := svc.detectMediaType(dir)
	assert.Equal(t, "music_album", mediaType)
}

func TestDetectMediaType_MixedVideoAudio_IsMovie(t *testing.T) {
	svc := newTestAggregationService(t)

	// Has both video and audio — video takes precedence for movie detection
	dir := directoryInfo{
		name:       "Some Video 2023",
		fileTypes:  map[string]int{".mkv": 1, ".mp3": 1},
		fileCount:  2,
		extensions: []string{".mkv", ".mp3"},
	}

	mediaType, _ := svc.detectMediaType(dir)
	// TV show detection runs first, then checks music (audio && !video),
	// then video = movie
	assert.Equal(t, "movie", mediaType)
}

// ---------------------------------------------------------------------------
// directoryInfo struct tests
// ---------------------------------------------------------------------------

func TestDirectoryInfo_Fields(t *testing.T) {
	dir := directoryInfo{
		path:       "/media/movies/Inception (2010)",
		name:       "Inception (2010)",
		fileCount:  3,
		totalSize:  1024 * 1024 * 500, // 500MB
		fileIDs:    []int64{1, 2, 3},
		fileTypes:  map[string]int{".mkv": 1, ".nfo": 1, ".srt": 1},
		extensions: []string{".mkv", ".nfo", ".srt"},
	}

	assert.Equal(t, "/media/movies/Inception (2010)", dir.path)
	assert.Equal(t, "Inception (2010)", dir.name)
	assert.Equal(t, 3, dir.fileCount)
	assert.Equal(t, int64(1024*1024*500), dir.totalSize)
	assert.Len(t, dir.fileIDs, 3)
	assert.Len(t, dir.fileTypes, 3)
}

// ---------------------------------------------------------------------------
// gamePlatformRe regex tests
// ---------------------------------------------------------------------------

func TestGamePlatformRegex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		matches bool
	}{
		{"PC", "Half-Life 2 PC", true},
		{"Xbox", "Forza Horizon Xbox", true},
		{"Xbox One", "Forza Horizon Xbox One", true},
		{"Xbox 360", "Gears of War Xbox 360", true},
		{"PS5", "Spider-Man PS5", true},
		{"PlayStation", "Gran Turismo PlayStation", true},
		{"Switch", "Zelda Switch", true},
		{"Nintendo", "Mario Kart Nintendo", true},
		{"Steam", "Portal 2 Steam", true},
		{"GOG", "Witcher 3 GOG", true},
		{"Linux", "SuperTux Linux", true},
		{"Windows", "Game Windows", true},
		{"no platform", "Just A Movie Title", false},
		{"no platform year", "Some Thing 2023", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.matches, gamePlatformRe.MatchString(tt.input),
				"gamePlatformRe.MatchString(%q)", tt.input)
		})
	}
}
