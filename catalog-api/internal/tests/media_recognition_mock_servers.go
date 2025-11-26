package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Comprehensive mock servers for media recognition APIs
type MediaRecognitionMockServers struct {
	logger            *zap.Logger
	tmdbServer        *httptest.Server
	omdbServer        *httptest.Server
	lastfmServer      *httptest.Server
	musicbrainzServer *httptest.Server
	acoustidServer    *httptest.Server
	igdbServer        *httptest.Server
	steamServer       *httptest.Server
	githubServer      *httptest.Server
	googlebooksServer *httptest.Server
	openlibraryServer *httptest.Server
	crossrefServer    *httptest.Server
	ocrServer         *httptest.Server
	wingetServer      *httptest.Server
	flatpakServer     *httptest.Server
	snapcraftServer   *httptest.Server
	homebrewServer    *httptest.Server

	// Request logging
	requestLogs []RequestLog
}

type RequestLog struct {
	Timestamp  time.Time         `json:"timestamp"`
	Method     string            `json:"method"`
	URL        string            `json:"url"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Response   string            `json:"response"`
	StatusCode int               `json:"status_code"`
	ServerType string            `json:"server_type"`
}

func NewMediaRecognitionMockServers(logger *zap.Logger) *MediaRecognitionMockServers {
	m := &MediaRecognitionMockServers{
		logger:      logger,
		requestLogs: make([]RequestLog, 0),
	}

	m.setupTMDbServer()
	m.setupOMDbServer()
	m.setupLastFMServer()
	m.setupMusicBrainzServer()
	m.setupAcoustIDServer()
	m.setupIGDBServer()
	m.setupSteamServer()
	m.setupGitHubServer()
	m.setupGoogleBooksServer()
	m.setupOpenLibraryServer()
	m.setupCrossrefServer()
	m.setupOCRServer()
	m.setupWingetServer()
	m.setupFlatpakServer()
	m.setupSnapcraftServer()
	m.setupHomebrewServer()

	return m
}

// TMDb Mock Server
func (m *MediaRecognitionMockServers) setupTMDbServer() {
	m.tmdbServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "tmdb")

		path := r.URL.Path
		query := r.URL.Query().Get("query")

		if strings.Contains(path, "/search/multi") {
			m.handleTMDbSearch(w, r, query)
		} else if strings.Contains(path, "/movie/") {
			m.handleTMDbMovieDetails(w, r)
		} else if strings.Contains(path, "/tv/") {
			m.handleTMDbTVDetails(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))
}

func (m *MediaRecognitionMockServers) handleTMDbSearch(w http.ResponseWriter, r *http.Request, query string) {
	response := map[string]interface{}{
		"page":          1,
		"total_pages":   1,
		"total_results": 1,
		"results": []map[string]interface{}{
			{
				"id":                12345,
				"title":             fmt.Sprintf("Mock Movie: %s", query),
				"name":              fmt.Sprintf("Mock TV Show: %s", query),
				"media_type":        "movie",
				"overview":          "This is a mock movie/TV show for testing purposes.",
				"release_date":      "2023-01-01",
				"first_air_date":    "2023-01-01",
				"genre_ids":         []int{28, 12, 878},
				"vote_average":      8.5,
				"vote_count":        1500,
				"poster_path":       "/mock_poster.jpg",
				"backdrop_path":     "/mock_backdrop.jpg",
				"popularity":        95.5,
				"adult":             false,
				"original_language": "en",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (m *MediaRecognitionMockServers) handleTMDbMovieDetails(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":             12345,
		"title":          "Mock Movie Details",
		"original_title": "Mock Movie Details",
		"overview":       "Detailed mock movie information for testing.",
		"release_date":   "2023-01-01",
		"runtime":        125,
		"vote_average":   8.5,
		"vote_count":     1500,
		"poster_path":    "/mock_poster.jpg",
		"backdrop_path":  "/mock_backdrop.jpg",
		"imdb_id":        "tt1234567",
		"budget":         150000000,
		"revenue":        500000000,
		"status":         "Released",
		"tagline":        "The ultimate mock movie experience",
		"genres": []map[string]interface{}{
			{"id": 28, "name": "Action"},
			{"id": 12, "name": "Adventure"},
			{"id": 878, "name": "Science Fiction"},
		},
		"production_companies": []map[string]interface{}{
			{
				"id":             1,
				"name":           "Mock Studios",
				"logo_path":      "/mock_logo.png",
				"origin_country": "US",
			},
		},
		"production_countries": []map[string]interface{}{
			{"iso_3166_1": "US", "name": "United States of America"},
		},
		"spoken_languages": []map[string]interface{}{
			{"iso_639_1": "en", "name": "English", "english_name": "English"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (m *MediaRecognitionMockServers) handleTMDbTVDetails(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"id":                 12345,
		"name":               "Mock TV Series",
		"original_name":      "Mock TV Series",
		"overview":           "Detailed mock TV series information for testing.",
		"first_air_date":     "2023-01-01",
		"last_air_date":      "2023-12-31",
		"vote_average":       8.7,
		"vote_count":         2000,
		"poster_path":        "/mock_tv_poster.jpg",
		"backdrop_path":      "/mock_tv_backdrop.jpg",
		"number_of_episodes": 24,
		"number_of_seasons":  2,
		"status":             "Ended",
		"type":               "Scripted",
		"external_ids": map[string]interface{}{
			"imdb_id": "tt7654321",
			"tvdb_id": 98765,
		},
		"genres": []map[string]interface{}{
			{"id": 18, "name": "Drama"},
			{"id": 80, "name": "Crime"},
		},
		"created_by": []map[string]interface{}{
			{
				"id":           1,
				"name":         "Mock Creator",
				"profile_path": "/mock_creator.jpg",
			},
		},
		"networks": []map[string]interface{}{
			{
				"id":             1,
				"name":           "Mock Network",
				"logo_path":      "/mock_network.png",
				"origin_country": "US",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// OMDb Mock Server
func (m *MediaRecognitionMockServers) setupOMDbServer() {
	m.omdbServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "omdb")

		title := r.URL.Query().Get("t")
		if title == "" {
			title = "Mock Title"
		}

		response := map[string]interface{}{
			"Title":    title,
			"Year":     "2023",
			"Rated":    "PG-13",
			"Released": "01 Jan 2023",
			"Runtime":  "125 min",
			"Genre":    "Action, Adventure, Sci-Fi",
			"Director": "Mock Director",
			"Writer":   "Mock Writer",
			"Actors":   "Mock Actor 1, Mock Actor 2, Mock Actor 3",
			"Plot":     "A comprehensive mock plot for testing purposes.",
			"Language": "English",
			"Country":  "USA",
			"Awards":   "Won 2 Oscars. Another 15 wins & 30 nominations.",
			"Poster":   "https://example.com/mock_poster.jpg",
			"Ratings": []map[string]string{
				{"Source": "Internet Movie Database", "Value": "8.5/10"},
				{"Source": "Rotten Tomatoes", "Value": "85%"},
				{"Source": "Metacritic", "Value": "78/100"},
			},
			"Metascore":  "78",
			"imdbRating": "8.5",
			"imdbVotes":  "150,000",
			"imdbID":     "tt1234567",
			"Type":       "movie",
			"DVD":        "01 Jun 2023",
			"BoxOffice":  "$500,000,000",
			"Production": "Mock Studios",
			"Website":    "https://example.com/mock-movie",
			"Response":   "True",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

// Last.fm Mock Server
func (m *MediaRecognitionMockServers) setupLastFMServer() {
	m.lastfmServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "lastfm")

		method := r.URL.Query().Get("method")
		track := r.URL.Query().Get("track")
		artist := r.URL.Query().Get("artist")

		if method == "track.search" {
			m.handleLastFMTrackSearch(w, r, track)
		} else if method == "track.getInfo" {
			m.handleLastFMTrackInfo(w, r, track, artist)
		} else {
			http.Error(w, "Method not supported", http.StatusBadRequest)
		}
	}))
}

func (m *MediaRecognitionMockServers) handleLastFMTrackSearch(w http.ResponseWriter, r *http.Request, track string) {
	response := map[string]interface{}{
		"results": map[string]interface{}{
			"trackmatches": map[string]interface{}{
				"track": []map[string]interface{}{
					{
						"name":       fmt.Sprintf("Mock Track: %s", track),
						"artist":     "Mock Artist",
						"url":        "https://last.fm/music/mock-artist/mock-track",
						"streamable": "1",
						"listeners":  "50000",
						"mbid":       "mock-mbid-12345",
						"image": []map[string]string{
							{"#text": "https://example.com/mock_small.jpg", "size": "small"},
							{"#text": "https://example.com/mock_medium.jpg", "size": "medium"},
							{"#text": "https://example.com/mock_large.jpg", "size": "large"},
						},
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (m *MediaRecognitionMockServers) handleLastFMTrackInfo(w http.ResponseWriter, r *http.Request, track, artist string) {
	response := map[string]interface{}{
		"track": map[string]interface{}{
			"name":      track,
			"mbid":      "mock-mbid-12345",
			"url":       "https://last.fm/music/mock-artist/mock-track",
			"duration":  "240000",
			"listeners": "50000",
			"playcount": "500000",
			"streamable": map[string]string{
				"#text":     "1",
				"fulltrack": "0",
			},
			"artist": map[string]string{
				"name": artist,
				"mbid": "mock-artist-mbid",
				"url":  "https://last.fm/music/mock-artist",
			},
			"album": map[string]interface{}{
				"artist": artist,
				"title":  "Mock Album",
				"mbid":   "mock-album-mbid",
				"url":    "https://last.fm/music/mock-artist/mock-album",
				"image": []map[string]string{
					{"#text": "https://example.com/mock_album_small.jpg", "size": "small"},
					{"#text": "https://example.com/mock_album_medium.jpg", "size": "medium"},
					{"#text": "https://example.com/mock_album_large.jpg", "size": "large"},
				},
			},
			"toptags": map[string]interface{}{
				"tag": []map[string]string{
					{"name": "rock", "url": "https://last.fm/tag/rock"},
					{"name": "alternative", "url": "https://last.fm/tag/alternative"},
					{"name": "indie", "url": "https://last.fm/tag/indie"},
				},
			},
			"wiki": map[string]string{
				"published": "01 Jan 2023, 12:00",
				"summary":   "This is a mock track summary for testing purposes.",
				"content":   "Extended mock content about this track with detailed information for testing.",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MusicBrainz Mock Server
func (m *MediaRecognitionMockServers) setupMusicBrainzServer() {
	m.musicbrainzServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "musicbrainz")

		path := r.URL.Path
		if strings.Contains(path, "/recording") {
			m.handleMusicBrainzRecordingSearch(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))
}

func (m *MediaRecognitionMockServers) handleMusicBrainzRecordingSearch(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"recordings": []map[string]interface{}{
			{
				"id":             "mock-recording-id-12345",
				"score":          95,
				"title":          "Mock Recording Title",
				"length":         240000,
				"disambiguation": "",
				"artist-credit": []map[string]interface{}{
					{
						"name": "Mock Artist",
						"artist": map[string]interface{}{
							"id":        "mock-artist-id",
							"name":      "Mock Artist",
							"sort-name": "Artist, Mock",
							"type":      "Person",
						},
					},
				},
				"releases": []map[string]interface{}{
					{
						"id":      "mock-release-id",
						"title":   "Mock Album",
						"status":  "Official",
						"date":    "2023",
						"country": "US",
					},
				},
				"tags": []map[string]interface{}{
					{"count": 5, "name": "rock"},
					{"count": 3, "name": "alternative"},
				},
				"genres": []map[string]interface{}{
					{"count": 8, "name": "rock"},
					{"count": 4, "name": "pop"},
				},
				"isrcs": []string{"USMOCK2300001"},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AcoustID Mock Server
func (m *MediaRecognitionMockServers) setupAcoustIDServer() {
	m.acoustidServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "acoustid")

		r.ParseForm()
		fingerprint := r.Form.Get("fingerprint")

		if fingerprint == "" {
			http.Error(w, "Fingerprint required", http.StatusBadRequest)
			return
		}

		response := map[string]interface{}{
			"status": "ok",
			"results": []map[string]interface{}{
				{
					"id":    "mock-acoustid-" + fingerprint[:8],
					"score": 0.95,
					"recordings": []map[string]interface{}{
						{
							"id":       "mock-recording-id-12345",
							"title":    "Mock Track from Fingerprint",
							"duration": 240.5,
							"artists": []map[string]interface{}{
								{
									"id":   "mock-artist-id",
									"name": "Mock Artist",
								},
							},
							"releases": []map[string]interface{}{
								{
									"id":      "mock-release-id",
									"title":   "Mock Album",
									"date":    "2023",
									"country": "US",
								},
							},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

// IGDB Mock Server
func (m *MediaRecognitionMockServers) setupIGDBServer() {
	m.igdbServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "igdb")

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := []map[string]interface{}{
			{
				"id":                 12345,
				"name":               "Mock Game Title",
				"summary":            "This is a comprehensive mock game for testing purposes with detailed gameplay mechanics.",
				"storyline":          "An epic storyline that spans across multiple dimensions and realities.",
				"first_release_date": 1672531200, // Unix timestamp for 2023-01-01
				"category":           0,
				"status":             6,
				"rating":             85.5,
				"rating_count":       1500,
				"aggregated_rating":  88.2,
				"total_rating":       86.8,
				"popularity":         95.7,
				"cover": map[string]interface{}{
					"id":       1,
					"url":      "//images.igdb.com/igdb/image/upload/t_cover_big/mock_cover.jpg",
					"image_id": "mock_cover",
					"width":    264,
					"height":   374,
				},
				"screenshots": []map[string]interface{}{
					{
						"id":       1,
						"url":      "//images.igdb.com/igdb/image/upload/t_screenshot_med/mock_screenshot1.jpg",
						"image_id": "mock_screenshot1",
						"width":    1920,
						"height":   1080,
					},
				},
				"genres": []map[string]interface{}{
					{"id": 12, "name": "Role-playing (RPG)", "slug": "role-playing-rpg"},
					{"id": 31, "name": "Adventure", "slug": "adventure"},
				},
				"themes": []map[string]interface{}{
					{"id": 1, "name": "Action", "slug": "action"},
					{"id": 17, "name": "Fantasy", "slug": "fantasy"},
				},
				"platforms": []map[string]interface{}{
					{
						"id":           6,
						"name":         "PC (Microsoft Windows)",
						"abbreviation": "PC",
						"category":     4,
						"generation":   8,
					},
					{
						"id":           48,
						"name":         "PlayStation 4",
						"abbreviation": "PS4",
						"category":     1,
						"generation":   8,
					},
				},
				"involved_companies": []map[string]interface{}{
					{
						"id":        1,
						"developer": true,
						"publisher": false,
						"company": map[string]interface{}{
							"id":   1,
							"name": "Mock Game Studios",
							"slug": "mock-game-studios",
						},
					},
					{
						"id":        2,
						"developer": false,
						"publisher": true,
						"company": map[string]interface{}{
							"id":   2,
							"name": "Mock Publishers",
							"slug": "mock-publishers",
						},
					},
				},
				"external_games": []map[string]interface{}{
					{
						"id":       1,
						"category": 1, // Steam
						"uid":      "12345",
						"url":      "https://store.steampowered.com/app/12345/",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

// GitHub Mock Server
func (m *MediaRecognitionMockServers) setupGitHubServer() {
	m.githubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "github")

		path := r.URL.Path
		if strings.Contains(path, "/search/repositories") {
			m.handleGitHubRepositorySearch(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))
}

func (m *MediaRecognitionMockServers) handleGitHubRepositorySearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	response := map[string]interface{}{
		"total_count":        1,
		"incomplete_results": false,
		"items": []map[string]interface{}{
			{
				"id":               12345,
				"name":             fmt.Sprintf("mock-%s", strings.ReplaceAll(query, " ", "-")),
				"full_name":        fmt.Sprintf("mockuser/mock-%s", strings.ReplaceAll(query, " ", "-")),
				"description":      fmt.Sprintf("Mock repository for %s - comprehensive testing implementation", query),
				"html_url":         fmt.Sprintf("https://github.com/mockuser/mock-%s", strings.ReplaceAll(query, " ", "-")),
				"stargazers_count": 1250,
				"forks_count":      234,
				"language":         "Go",
				"license": map[string]interface{}{
					"key":     "mit",
					"name":    "MIT License",
					"spdx_id": "MIT",
				},
				"owner": map[string]interface{}{
					"login":      "mockuser",
					"id":         67890,
					"avatar_url": "https://avatars.githubusercontent.com/u/67890?v=4",
					"type":       "User",
				},
				"created_at":     "2020-01-01T00:00:00Z",
				"updated_at":     "2023-12-01T00:00:00Z",
				"pushed_at":      "2023-12-01T00:00:00Z",
				"default_branch": "main",
				"topics":         []string{"mock", "testing", "api", "software"},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Google Books Mock Server
func (m *MediaRecognitionMockServers) setupGoogleBooksServer() {
	m.googlebooksServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "googlebooks")

		query := r.URL.Query().Get("q")

		response := map[string]interface{}{
			"kind":       "books#volumes",
			"totalItems": 1,
			"items": []map[string]interface{}{
				{
					"kind":     "books#volume",
					"id":       "mock-book-id-12345",
					"etag":     "mock-etag",
					"selfLink": "https://www.googleapis.com/books/v1/volumes/mock-book-id-12345",
					"volumeInfo": map[string]interface{}{
						"title":         fmt.Sprintf("Mock Book: %s", query),
						"subtitle":      "A Comprehensive Guide to Testing",
						"authors":       []string{"Mock Author", "Test Writer"},
						"publisher":     "Mock Publishing House",
						"publishedDate": "2023-01-01",
						"description":   "This is a comprehensive mock book designed for testing purposes. It contains detailed information about various testing methodologies and best practices.",
						"industryIdentifiers": []map[string]interface{}{
							{
								"type":       "ISBN_13",
								"identifier": "9781234567890",
							},
							{
								"type":       "ISBN_10",
								"identifier": "1234567890",
							},
						},
						"readingModes": map[string]bool{
							"text":  true,
							"image": true,
						},
						"pageCount":      456,
						"printType":      "BOOK",
						"categories":     []string{"Computers", "Testing", "Software Engineering"},
						"averageRating":  4.5,
						"ratingsCount":   125,
						"maturityRating": "NOT_MATURE",
						"imageLinks": map[string]string{
							"smallThumbnail": "https://example.com/mock_book_small.jpg",
							"thumbnail":      "https://example.com/mock_book_medium.jpg",
							"small":          "https://example.com/mock_book_small.jpg",
							"medium":         "https://example.com/mock_book_medium.jpg",
							"large":          "https://example.com/mock_book_large.jpg",
						},
						"language":            "en",
						"previewLink":         "https://books.google.com/books?id=mock-book-id-12345",
						"infoLink":            "https://books.google.com/books?id=mock-book-id-12345",
						"canonicalVolumeLink": "https://books.google.com/books/about/Mock_Book.html?id=mock-book-id-12345",
					},
					"saleInfo": map[string]interface{}{
						"country":     "US",
						"saleability": "FOR_SALE",
						"isEbook":     true,
						"listPrice": map[string]interface{}{
							"amount":       29.99,
							"currencyCode": "USD",
						},
						"retailPrice": map[string]interface{}{
							"amount":       24.99,
							"currencyCode": "USD",
						},
						"buyLink": "https://books.google.com/books?id=mock-book-id-12345&buy",
					},
					"accessInfo": map[string]interface{}{
						"country":                "US",
						"viewability":            "PARTIAL",
						"embeddable":             true,
						"publicDomain":           false,
						"textToSpeechPermission": "ALLOWED",
						"epub": map[string]interface{}{
							"isAvailable": true,
						},
						"pdf": map[string]interface{}{
							"isAvailable": true,
						},
						"webReaderLink":       "https://books.google.com/books/reader?id=mock-book-id-12345",
						"accessViewStatus":    "SAMPLE",
						"quoteSharingAllowed": true,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

// Open Library Mock Server
func (m *MediaRecognitionMockServers) setupOpenLibraryServer() {
	m.openlibraryServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "openlibrary")

		query := r.URL.Query().Get("q")

		response := map[string]interface{}{
			"numFound":      1,
			"start":         0,
			"numFoundExact": true,
			"docs": []map[string]interface{}{
				{
					"key":                    "/works/OL12345W",
					"type":                   "work",
					"title":                  fmt.Sprintf("Mock Open Library Book: %s", query),
					"title_suggest":          fmt.Sprintf("Mock Open Library Book: %s", query),
					"subtitle":               "An Open Source Testing Guide",
					"author_name":            []string{"Mock Author", "Open Contributor"},
					"author_key":             []string{"/authors/OL123A", "/authors/OL456A"},
					"publisher":              []string{"Open Source Press", "Community Publishers"},
					"publish_date":           []string{"2023", "January 2023"},
					"publish_year":           []int{2023},
					"first_publish_year":     2023,
					"number_of_pages_median": 340,
					"edition_count":          3,
					"edition_key":            []string{"/books/OL123M", "/books/OL456M", "/books/OL789M"},
					"isbn":                   []string{"9780987654321", "0987654321"},
					"lccn":                   []string{"2023123456"},
					"oclc":                   []string{"1234567890"},
					"subject":                []string{"Testing", "Software Development", "Open Source", "Programming"},
					"place":                  []string{"San Francisco", "California"},
					"person":                 []string{"Linus Torvalds", "Richard Stallman"},
					"language":               []string{"eng"},
					"id_goodreads":           []string{"12345678"},
					"id_librarything":        []string{"987654"},
					"cover_i":                98765,
					"cover_edition_key":      "/books/OL123M",
					"first_sentence":         []string{"This comprehensive guide introduces the fundamentals of testing in open source environments."},
					"ebook_count_i":          2,
					"ebook_access":           "borrowable",
					"has_fulltext":           true,
					"public_scan_b":          true,
					"last_modified_i":        time.Now().Unix(),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

// Crossref Mock Server
func (m *MediaRecognitionMockServers) setupCrossrefServer() {
	m.crossrefServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "crossref")

		query := r.URL.Query().Get("query")

		response := map[string]interface{}{
			"status": "ok",
			"message": map[string]interface{}{
				"total-results":  1,
				"items-per-page": 20,
				"items": []map[string]interface{}{
					{
						"indexed": map[string]interface{}{
							"date-parts": [][]int{{2023, 1, 15}},
							"date-time":  "2023-01-15T10:30:00Z",
							"timestamp":  1673779800000,
						},
						"reference-count": 25,
						"publisher":       "Mock Academic Press",
						"issue":           "1",
						"content-domain": map[string]interface{}{
							"domain":                []string{"mockacademic.org"},
							"crossmark-restriction": false,
						},
						"published": map[string]interface{}{
							"date-parts": [][]int{{2023, 1, 1}},
						},
						"abstract": fmt.Sprintf("This is a comprehensive academic paper about %s, providing detailed analysis and research findings.", query),
						"DOI":      "10.1000/mock-doi-12345",
						"type":     "journal-article",
						"created": map[string]interface{}{
							"date-parts": [][]int{{2023, 1, 1}},
							"date-time":  "2023-01-01T00:00:00Z",
							"timestamp":  1672531200000,
						},
						"page":                   "1-25",
						"source":                 "Crossref",
						"is-referenced-by-count": 15,
						"title":                  []string{fmt.Sprintf("Mock Academic Paper: %s", query)},
						"prefix":                 "10.1000",
						"volume":                 "45",
						"author": []map[string]interface{}{
							{
								"given":    "John",
								"family":   "MockResearcher",
								"sequence": "first",
								"affiliation": []map[string]interface{}{
									{"name": "Mock University, Department of Computer Science"},
								},
							},
							{
								"given":    "Jane",
								"family":   "TestScientist",
								"sequence": "additional",
								"affiliation": []map[string]interface{}{
									{"name": "Research Institute of Technology"},
								},
							},
						},
						"member":          "1000",
						"container-title": []string{"Journal of Mock Research"},
						"language":        "en",
						"deposited": map[string]interface{}{
							"date-parts": [][]int{{2023, 1, 15}},
							"date-time":  "2023-01-15T10:30:00Z",
							"timestamp":  1673779800000,
						},
						"score": 95.5,
						"issued": map[string]interface{}{
							"date-parts": [][]int{{2023, 1, 1}},
						},
						"references-count": 25,
						"URL":              "https://mockacademic.org/articles/mock-doi-12345",
						"ISSN":             []string{"1234-5678", "9876-5432"},
						"subject":          []string{"Computer Science", "Software Engineering", "Testing"},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

// OCR Mock Server
func (m *MediaRecognitionMockServers) setupOCRServer() {
	m.ocrServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "ocr")

		response := map[string]interface{}{
			"ParsedResults": []map[string]interface{}{
				{
					"TextOverlay": map[string]interface{}{
						"Lines": []map[string]interface{}{
							{
								"LineText": "Mock Book Title",
								"Words": []map[string]interface{}{
									{
										"WordText": "Mock",
										"Left":     50,
										"Top":      100,
										"Height":   30,
										"Width":    80,
									},
									{
										"WordText": "Book",
										"Left":     140,
										"Top":      100,
										"Height":   30,
										"Width":    80,
									},
									{
										"WordText": "Title",
										"Left":     230,
										"Top":      100,
										"Height":   30,
										"Width":    80,
									},
								},
							},
							{
								"LineText": "By Mock Author",
								"Words": []map[string]interface{}{
									{
										"WordText": "By",
										"Left":     50,
										"Top":      150,
										"Height":   25,
										"Width":    30,
									},
									{
										"WordText": "Mock",
										"Left":     90,
										"Top":      150,
										"Height":   25,
										"Width":    60,
									},
									{
										"WordText": "Author",
										"Left":     160,
										"Top":      150,
										"Height":   25,
										"Width":    80,
									},
								},
							},
							{
								"LineText": "ISBN: 978-1234567890",
								"Words": []map[string]interface{}{
									{
										"WordText": "ISBN:",
										"Left":     50,
										"Top":      200,
										"Height":   20,
										"Width":    50,
									},
									{
										"WordText": "978-1234567890",
										"Left":     110,
										"Top":      200,
										"Height":   20,
										"Width":    150,
									},
								},
							},
						},
						"HasOverlay": true,
						"Message":    "Total lines: 3",
					},
					"FileParseExitCode": 1,
					"ParsedText":        "Mock Book Title\nBy Mock Author\nISBN: 978-1234567890\n\nThis is sample text content from a book page that has been processed through OCR. The text includes chapter headings, author information, and ISBN details that can be extracted for book recognition purposes.",
					"ErrorMessage":      "",
					"ErrorDetails":      "",
				},
			},
			"OCRExitCode":                  1,
			"IsErroredOnProcessing":        false,
			"ProcessingTimeInMilliseconds": "1250",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

// Package manager mock servers
func (m *MediaRecognitionMockServers) setupWingetServer() {
	m.wingetServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "winget")

		query := r.URL.Query().Get("query")

		response := []map[string]interface{}{
			{
				"PackageIdentifier": "MockSoftware.TestApp",
				"PackageName":       fmt.Sprintf("Mock %s", query),
				"PackageVersion":    "1.2.3",
				"Publisher":         "Mock Software Inc.",
				"Description":       fmt.Sprintf("A comprehensive mock application for %s testing purposes.", query),
				"License":           "MIT",
				"Tags":              []string{"mock", "testing", "software"},
				"Homepage":          "https://mocksoft.com/testapp",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

func (m *MediaRecognitionMockServers) setupFlatpakServer() {
	m.flatpakServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "flatpak")

		query := r.URL.Query().Get("q")

		response := []map[string]interface{}{
			{
				"flatpakAppId":   "org.mocksoft.TestApp",
				"name":           fmt.Sprintf("Mock %s", query),
				"summary":        fmt.Sprintf("Mock application for %s testing", query),
				"description":    fmt.Sprintf("A comprehensive Flatpak application designed for testing %s functionality.", query),
				"developerName":  "Mock Software Foundation",
				"projectLicense": "GPL-3.0+",
				"categories":     []string{"Development", "Education"},
				"screenshots": []string{
					"https://example.com/mock_screenshot1.png",
					"https://example.com/mock_screenshot2.png",
				},
				"iconDesktopUrl":        "https://example.com/mock_icon.png",
				"downloadFlatpakRefUrl": "https://dl.flathub.org/repo/appstream/org.mocksoft.TestApp.flatpakref",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

func (m *MediaRecognitionMockServers) setupSnapcraftServer() {
	m.snapcraftServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "snapcraft")

		query := r.URL.Query().Get("q")

		response := map[string]interface{}{
			"_embedded": []map[string]interface{}{
				{
					"name":        fmt.Sprintf("mock-%s", strings.ReplaceAll(strings.ToLower(query), " ", "-")),
					"title":       fmt.Sprintf("Mock %s", query),
					"summary":     fmt.Sprintf("A mock snap package for %s testing", query),
					"description": fmt.Sprintf("Comprehensive snap package designed for testing %s functionality with all necessary components.", query),
					"publisher": map[string]interface{}{
						"display-name": "Mock Software Publishers",
						"username":     "mockpublisher",
						"validation":   "verified",
					},
					"license":     "MIT",
					"version":     "1.2.3",
					"revision":    42,
					"confinement": "strict",
					"grade":       "stable",
					"categories": []map[string]interface{}{
						{"name": "development"},
						{"name": "education"},
					},
					"screenshots": []string{
						"https://example.com/snap_screenshot1.png",
						"https://example.com/snap_screenshot2.png",
					},
					"media": []map[string]interface{}{
						{
							"type": "icon",
							"url":  "https://example.com/snap_icon.png",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

func (m *MediaRecognitionMockServers) setupHomebrewServer() {
	m.homebrewServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "homebrew")

		// Extract formula name from path
		pathParts := strings.Split(r.URL.Path, "/")
		formulaName := "mock-formula"
		if len(pathParts) > 3 {
			formulaName = strings.TrimSuffix(pathParts[len(pathParts)-1], ".json")
		}

		response := map[string]interface{}{
			"name":               formulaName,
			"full_name":          fmt.Sprintf("mock/%s", formulaName),
			"tap":                "homebrew/core",
			"oldname":            nil,
			"aliases":            []string{},
			"versioned_formulae": []string{},
			"desc":               fmt.Sprintf("Mock Homebrew formula for %s testing purposes", formulaName),
			"license":            "MIT",
			"homepage":           "https://mocksoft.com/homebrew-formula",
			"versions": map[string]interface{}{
				"stable": "1.2.3",
				"head":   "HEAD",
				"bottle": true,
			},
			"urls": map[string]interface{}{
				"stable": map[string]interface{}{
					"url":      fmt.Sprintf("https://github.com/mocksoft/%s/archive/v1.2.3.tar.gz", formulaName),
					"tag":      "v1.2.3",
					"revision": "abc123def456",
				},
			},
			"revision":       0,
			"version_scheme": 0,
			"bottle": map[string]interface{}{
				"stable": map[string]interface{}{
					"rebuild":  0,
					"root_url": "https://homebrew.bintray.com/bottles",
					"files": map[string]interface{}{
						"monterey": map[string]interface{}{
							"cellar": "/usr/local/Cellar",
							"url":    fmt.Sprintf("https://homebrew.bintray.com/bottles/%s-1.2.3.monterey.bottle.tar.gz", formulaName),
							"sha256": "mock_sha256_hash_for_monterey_bottle",
						},
						"big_sur": map[string]interface{}{
							"cellar": "/usr/local/Cellar",
							"url":    fmt.Sprintf("https://homebrew.bintray.com/bottles/%s-1.2.3.big_sur.bottle.tar.gz", formulaName),
							"sha256": "mock_sha256_hash_for_big_sur_bottle",
						},
					},
				},
			},
			"dependencies":             []string{"mock-dependency-1", "mock-dependency-2"},
			"test_dependencies":        []string{"mock-test-dep"},
			"recommended_dependencies": []string{},
			"optional_dependencies":    []string{},
			"build_dependencies":       []string{"cmake", "pkg-config"},
			"conflicts_with":           []string{},
			"caveats":                  nil,
			"installed":                []string{},
			"linked_keg":               nil,
			"pinned":                   false,
			"outdated":                 false,
			"deprecated":               false,
			"deprecation_date":         nil,
			"deprecation_reason":       nil,
			"disabled":                 false,
			"disable_date":             nil,
			"disable_reason":           nil,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

// Steam Mock Server
func (m *MediaRecognitionMockServers) setupSteamServer() {
	m.steamServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logRequest(r, "steam")

		// Mock Steam app details
		response := map[string]interface{}{
			"12345": map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"type":                 "game",
					"name":                 "Mock Steam Game",
					"steam_appid":          12345,
					"required_age":         0,
					"is_free":              false,
					"detailed_description": "A comprehensive mock Steam game for testing purposes with detailed gameplay mechanics and storyline.",
					"about_the_game":       "Experience the ultimate mock gaming adventure with cutting-edge graphics and immersive gameplay.",
					"short_description":    "The definitive mock game experience.",
					"supported_languages":  "English<strong>*</strong>, French, German, Spanish<br><strong>*</strong>languages with full audio support",
					"header_image":         "https://cdn.akamai.steamstatic.com/steam/apps/12345/header.jpg",
					"website":              "https://mockgame.com",
					"pc_requirements": map[string]interface{}{
						"minimum":     "<strong>Minimum:</strong><br><ul class=\"bb_ul\"><li><strong>OS:</strong> Windows 10 64-bit<li><strong>Processor:</strong> Intel Core i5-8400 / AMD Ryzen 5 2600<li><strong>Memory:</strong> 8 GB RAM<li><strong>Graphics:</strong> NVIDIA GTX 1060 / AMD RX 580<li><strong>DirectX:</strong> Version 12<li><strong>Storage:</strong> 50 GB available space</ul>",
						"recommended": "<strong>Recommended:</strong><br><ul class=\"bb_ul\"><li><strong>OS:</strong> Windows 11 64-bit<li><strong>Processor:</strong> Intel Core i7-10700K / AMD Ryzen 7 3700X<li><strong>Memory:</strong> 16 GB RAM<li><strong>Graphics:</strong> NVIDIA RTX 3070 / AMD RX 6700 XT<li><strong>DirectX:</strong> Version 12<li><strong>Storage:</strong> 50 GB available space (SSD recommended)</ul>",
					},
					"developers": []string{"Mock Game Studios"},
					"publishers": []string{"Mock Publishers"},
					"platforms": map[string]bool{
						"windows": true,
						"mac":     false,
						"linux":   true,
					},
					"metacritic": map[string]interface{}{
						"score": 85,
						"url":   "https://www.metacritic.com/game/pc/mock-steam-game",
					},
					"categories": []map[string]interface{}{
						{"id": 2, "description": "Single-player"},
						{"id": 1, "description": "Multi-player"},
						{"id": 22, "description": "Steam Achievements"},
						{"id": 29, "description": "Steam Trading Cards"},
					},
					"genres": []map[string]interface{}{
						{"id": "1", "description": "Action"},
						{"id": "25", "description": "Adventure"},
						{"id": "23", "description": "Indie"},
					},
					"screenshots": []map[string]interface{}{
						{
							"id":             1,
							"path_thumbnail": "https://cdn.akamai.steamstatic.com/steam/apps/12345/ss_1_thumbnail.jpg",
							"path_full":      "https://cdn.akamai.steamstatic.com/steam/apps/12345/ss_1.jpg",
						},
					},
					"release_date": map[string]interface{}{
						"coming_soon": false,
						"date":        "Jan 1, 2023",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

// Request logging functionality
func (m *MediaRecognitionMockServers) logRequest(r *http.Request, serverType string) {
	headers := make(map[string]string)
	for name, values := range r.Header {
		headers[name] = strings.Join(values, ", ")
	}

	body := ""
	if r.Body != nil {
		bodyBytes := make([]byte, r.ContentLength)
		r.Body.Read(bodyBytes)
		body = string(bodyBytes)
	}

	log := RequestLog{
		Timestamp:  time.Now(),
		Method:     r.Method,
		URL:        r.URL.String(),
		Headers:    headers,
		Body:       body,
		ServerType: serverType,
	}

	m.requestLogs = append(m.requestLogs, log)
}

// Get request logs for analysis
func (m *MediaRecognitionMockServers) GetRequestLogs() []RequestLog {
	return m.requestLogs
}

// Clear request logs
func (m *MediaRecognitionMockServers) ClearRequestLogs() {
	m.requestLogs = make([]RequestLog, 0)
}

// Get server URLs for configuration
func (m *MediaRecognitionMockServers) GetURLs() map[string]string {
	return map[string]string{
		"tmdb":        m.tmdbServer.URL,
		"omdb":        m.omdbServer.URL,
		"lastfm":      m.lastfmServer.URL,
		"musicbrainz": m.musicbrainzServer.URL,
		"acoustid":    m.acoustidServer.URL,
		"igdb":        m.igdbServer.URL,
		"steam":       m.steamServer.URL,
		"github":      m.githubServer.URL,
		"googlebooks": m.googlebooksServer.URL,
		"openlibrary": m.openlibraryServer.URL,
		"crossref":    m.crossrefServer.URL,
		"ocr":         m.ocrServer.URL,
		"winget":      m.wingetServer.URL,
		"flatpak":     m.flatpakServer.URL,
		"snapcraft":   m.snapcraftServer.URL,
		"homebrew":    m.homebrewServer.URL,
	}
}

// Close all mock servers
func (m *MediaRecognitionMockServers) Close() {
	if m.tmdbServer != nil {
		m.tmdbServer.Close()
	}
	if m.omdbServer != nil {
		m.omdbServer.Close()
	}
	if m.lastfmServer != nil {
		m.lastfmServer.Close()
	}
	if m.musicbrainzServer != nil {
		m.musicbrainzServer.Close()
	}
	if m.acoustidServer != nil {
		m.acoustidServer.Close()
	}
	if m.igdbServer != nil {
		m.igdbServer.Close()
	}
	if m.steamServer != nil {
		m.steamServer.Close()
	}
	if m.githubServer != nil {
		m.githubServer.Close()
	}
	if m.googlebooksServer != nil {
		m.googlebooksServer.Close()
	}
	if m.openlibraryServer != nil {
		m.openlibraryServer.Close()
	}
	if m.crossrefServer != nil {
		m.crossrefServer.Close()
	}
	if m.ocrServer != nil {
		m.ocrServer.Close()
	}
	if m.wingetServer != nil {
		m.wingetServer.Close()
	}
	if m.flatpakServer != nil {
		m.flatpakServer.Close()
	}
	if m.snapcraftServer != nil {
		m.snapcraftServer.Close()
	}
	if m.homebrewServer != nil {
		m.homebrewServer.Close()
	}
}

// Helper function to simulate network delays for realistic testing
func (m *MediaRecognitionMockServers) simulateNetworkDelay() {
	// Simulate 50-200ms network delay
	delay := time.Duration(50+time.Now().UnixNano()%150) * time.Millisecond
	time.Sleep(delay)
}
