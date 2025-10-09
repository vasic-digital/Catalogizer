package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type MockServer struct {
	Server        *httptest.Server
	RequestLog    []MockRequest
	ResponseDelay time.Duration
}

type MockRequest struct {
	Method    string                 `json:"method"`
	URL       string                 `json:"url"`
	Headers   map[string]string      `json:"headers"`
	Body      string                 `json:"body"`
	Timestamp time.Time              `json:"timestamp"`
	Query     map[string][]string    `json:"query"`
}

type MockResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       interface{}            `json:"body"`
	Delay      time.Duration          `json:"delay"`
}

func NewMockServer() *MockServer {
	mock := &MockServer{
		RequestLog: make([]MockRequest, 0),
	}

	router := mux.NewRouter()
	mock.setupRoutes(router)

	mock.Server = httptest.NewServer(router)
	return mock
}

func (m *MockServer) Close() {
	m.Server.Close()
}

func (m *MockServer) URL() string {
	return m.Server.URL
}

func (m *MockServer) GetRequestLog() []MockRequest {
	return m.RequestLog
}

func (m *MockServer) ClearRequestLog() {
	m.RequestLog = make([]MockRequest, 0)
}

func (m *MockServer) SetResponseDelay(delay time.Duration) {
	m.ResponseDelay = delay
}

func (m *MockServer) logRequest(r *http.Request, body string) {
	headers := make(map[string]string)
	for key, values := range r.Header {
		headers[key] = strings.Join(values, ", ")
	}

	request := MockRequest{
		Method:    r.Method,
		URL:       r.URL.String(),
		Headers:   headers,
		Body:      body,
		Timestamp: time.Now(),
		Query:     r.URL.Query(),
	}

	m.RequestLog = append(m.RequestLog, request)
}

func (m *MockServer) respondWithDelay(w http.ResponseWriter, response MockResponse) {
	if m.ResponseDelay > 0 {
		time.Sleep(m.ResponseDelay)
	}
	if response.Delay > 0 {
		time.Sleep(response.Delay)
	}

	for key, value := range response.Headers {
		w.Header().Set(key, value)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)
	json.NewEncoder(w).Encode(response.Body)
}

func (m *MockServer) setupRoutes(router *mux.Router) {
	// OpenSubtitles API Mock
	router.HandleFunc("/opensubtitles/api/v1/login", m.mockOpenSubtitlesLogin).Methods("POST")
	router.HandleFunc("/opensubtitles/api/v1/subtitles", m.mockOpenSubtitlesSearch).Methods("GET")
	router.HandleFunc("/opensubtitles/api/v1/download", m.mockOpenSubtitlesDownload).Methods("POST")

	// SubDB API Mock
	router.HandleFunc("/subdb", m.mockSubDBSearch).Methods("GET")

	// YifySubtitles API Mock
	router.HandleFunc("/yifysubtitles/api/v1/subtitles", m.mockYifySubtitlesSearch).Methods("GET")

	// Genius API Mock
	router.HandleFunc("/genius/api/search", m.mockGeniusSearch).Methods("GET")
	router.HandleFunc("/genius/api/songs/{id}/lyrics", m.mockGeniusLyrics).Methods("GET")

	// Musixmatch API Mock
	router.HandleFunc("/musixmatch/ws/1.1/track.search", m.mockMusixmatchSearch).Methods("GET")
	router.HandleFunc("/musixmatch/ws/1.1/track.lyrics.get", m.mockMusixmatchLyrics).Methods("GET")

	// AZLyrics Mock
	router.HandleFunc("/azlyrics/{artist}/{song}", m.mockAZLyrics).Methods("GET")

	// MusicBrainz API Mock
	router.HandleFunc("/musicbrainz/ws/2/recording", m.mockMusicBrainzSearch).Methods("GET")
	router.HandleFunc("/musicbrainz/ws/2/release", m.mockMusicBrainzRelease).Methods("GET")

	// Last.FM API Mock
	router.HandleFunc("/lastfm/2.0", m.mockLastFMAPI).Methods("GET")

	// iTunes API Mock
	router.HandleFunc("/itunes/search", m.mockiTunesSearch).Methods("GET")

	// Spotify API Mock
	router.HandleFunc("/spotify/v1/search", m.mockSpotifySearch).Methods("GET")
	router.HandleFunc("/spotify/api/token", m.mockSpotifyToken).Methods("POST")

	// Discogs API Mock
	router.HandleFunc("/discogs/database/search", m.mockDiscogsSearch).Methods("GET")

	// Google Translate Mock
	router.HandleFunc("/google/translate/v2", m.mockGoogleTranslate).Methods("POST")
	router.HandleFunc("/google/translate/v2/detect", m.mockGoogleDetect).Methods("POST")

	// LibreTranslate Mock
	router.HandleFunc("/libretranslate/translate", m.mockLibreTranslate).Methods("POST")
	router.HandleFunc("/libretranslate/detect", m.mockLibreDetect).Methods("POST")

	// MyMemory Translation Mock
	router.HandleFunc("/mymemory/get", m.mockMyMemoryTranslate).Methods("GET")

	// Setlist.fm API Mock
	router.HandleFunc("/setlistfm/rest/1.0/search/setlists", m.mockSetlistFMSearch).Methods("GET")
}

// OpenSubtitles API Mocks

func (m *MockServer) mockOpenSubtitlesLogin(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"token": "mock_opensubtitles_token_12345",
			"user": map[string]interface{}{
				"allowed_downloads": 200,
				"level":            "VIP",
				"user_id":          12345,
			},
		},
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) mockOpenSubtitlesSearch(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	imdbID := r.URL.Query().Get("imdb_id")
	languages := r.URL.Query().Get("languages")

	if imdbID == "" {
		response := MockResponse{
			StatusCode: 400,
			Body: map[string]interface{}{
				"message": "imdb_id parameter is required",
			},
		}
		m.respondWithDelay(w, response)
		return
	}

	subtitles := []map[string]interface{}{
		{
			"id":           "subtitle_123",
			"type":         "subtitle",
			"language":     strings.Split(languages, ",")[0],
			"filename":     "movie.srt",
			"url":          fmt.Sprintf("%s/opensubtitles/files/subtitle_123.srt", m.URL()),
			"download_url": fmt.Sprintf("%s/opensubtitles/api/v1/download", m.URL()),
			"fps":          23.976,
			"file_id":      123456,
			"rating":       8.5,
			"downloads":    15420,
		},
		{
			"id":           "subtitle_124",
			"type":         "subtitle",
			"language":     "en",
			"filename":     "movie_eng.srt",
			"url":          fmt.Sprintf("%s/opensubtitles/files/subtitle_124.srt", m.URL()),
			"download_url": fmt.Sprintf("%s/opensubtitles/api/v1/download", m.URL()),
			"fps":          23.976,
			"file_id":      123457,
			"rating":       9.0,
			"downloads":    25830,
		},
	}

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"total_pages": 1,
			"total_count": len(subtitles),
			"per_page":    60,
			"page":        1,
			"data":        subtitles,
		},
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) mockOpenSubtitlesDownload(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"link":          fmt.Sprintf("%s/opensubtitles/files/subtitle.srt", m.URL()),
			"file_name":     "subtitle.srt",
			"requests":      199,
			"remaining":     199,
			"message":       "Download successful",
			"reset_time":    "24:00:00",
			"reset_time_utc": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		},
	}
	m.respondWithDelay(w, response)
}

// SubDB API Mock

func (m *MockServer) mockSubDBSearch(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	action := r.URL.Query().Get("action")
	hash := r.URL.Query().Get("hash")

	if action != "search" || hash == "" {
		w.WriteHeader(404)
		return
	}

	response := MockResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
		Body: "en,es,fr,de",
	}
	m.respondWithDelay(w, response)
}

// Genius API Mocks

func (m *MockServer) mockGeniusSearch(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	query := r.URL.Query().Get("q")
	if query == "" {
		response := MockResponse{
			StatusCode: 400,
			Body: map[string]interface{}{
				"error": "Missing required parameter: q",
			},
		}
		m.respondWithDelay(w, response)
		return
	}

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"meta": map[string]interface{}{
				"status": 200,
			},
			"response": map[string]interface{}{
				"hits": []map[string]interface{}{
					{
						"type":   "song",
						"index":  "song",
						"result": map[string]interface{}{
							"id":                    123456,
							"title":                 "Test Song",
							"title_with_featured":   "Test Song (feat. Test Artist)",
							"full_title":            "Test Song by Test Artist",
							"artist_names":          "Test Artist",
							"primary_artist":        map[string]interface{}{
								"id":   98765,
								"name": "Test Artist",
								"url":  fmt.Sprintf("%s/genius/artists/98765", m.URL()),
							},
							"url": fmt.Sprintf("%s/genius/songs/123456", m.URL()),
							"song_art_image_thumbnail_url": fmt.Sprintf("%s/images/song_art_123456_thumb.jpg", m.URL()),
							"song_art_image_url":           fmt.Sprintf("%s/images/song_art_123456.jpg", m.URL()),
						},
					},
				},
			},
		},
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) mockGeniusLyrics(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	songID := mux.Vars(r)["id"]

	mockLyrics := `[Verse 1]
This is a test song
With some mock lyrics
For testing purposes only

[Chorus]
La la la la la
Test test test
Mock lyrics here

[Verse 2]
More test lyrics
In a structured format
With timestamps if needed

[Outro]
End of mock song`

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"meta": map[string]interface{}{
				"status": 200,
			},
			"response": map[string]interface{}{
				"song": map[string]interface{}{
					"id":     songID,
					"lyrics": mockLyrics,
					"title":  "Test Song",
					"artist": "Test Artist",
				},
			},
		},
	}
	m.respondWithDelay(w, response)
}

// Google Translate Mocks

func (m *MockServer) mockGoogleTranslate(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	var reqData map[string]interface{}
	json.Unmarshal([]byte(body), &reqData)

	text := "Mock translated text"
	if q, ok := reqData["q"].(string); ok {
		text = fmt.Sprintf("Translated: %s", q)
	}

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"data": map[string]interface{}{
				"translations": []map[string]interface{}{
					{
						"translatedText":   text,
						"detectedSourceLanguage": "en",
					},
				},
			},
		},
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) mockGoogleDetect(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"data": map[string]interface{}{
				"detections": [][]map[string]interface{}{
					{
						{
							"language":   "en",
							"isReliable": true,
							"confidence": 0.95,
						},
					},
				},
			},
		},
	}
	m.respondWithDelay(w, response)
}

// LibreTranslate Mocks

func (m *MockServer) mockLibreTranslate(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	var reqData map[string]interface{}
	json.Unmarshal([]byte(body), &reqData)

	text := "Mock LibreTranslate result"
	if q, ok := reqData["q"].(string); ok {
		text = fmt.Sprintf("LibreTranslated: %s", q)
	}

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"translatedText": text,
		},
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) mockLibreDetect(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	response := MockResponse{
		StatusCode: 200,
		Body: []map[string]interface{}{
			{
				"confidence": 0.92,
				"language":   "en",
			},
		},
	}
	m.respondWithDelay(w, response)
}

// MyMemory Translation Mock

func (m *MockServer) mockMyMemoryTranslate(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	q := r.URL.Query().Get("q")

	text := "Mock MyMemory translation"
	if q != "" {
		text = fmt.Sprintf("MyMemory: %s", q)
	}

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"responseData": map[string]interface{}{
				"translatedText": text,
				"match":          0.85,
			},
			"quotaFinished": false,
			"mtLangSupported": true,
			"responseDetails": "",
			"responseStatus":  200,
			"responderId":     "MyMemory",
		},
	}
	m.respondWithDelay(w, response)
}

// Music API Mocks

func (m *MockServer) mockMusicBrainzSearch(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	format := r.URL.Query().Get("fmt")

	if format != "json" {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(200)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><metadata/>`))
		return
	}

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"recordings": []map[string]interface{}{
				{
					"id":    "test-recording-id-123",
					"title": "Test Recording",
					"artist-credit": []map[string]interface{}{
						{
							"name": "Test Artist",
							"artist": map[string]interface{}{
								"id":   "test-artist-id-456",
								"name": "Test Artist",
							},
						},
					},
					"releases": []map[string]interface{}{
						{
							"id":    "test-release-id-789",
							"title": "Test Album",
							"date":  "2023-01-01",
						},
					},
				},
			},
			"count":  1,
			"offset": 0,
		},
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) mockLastFMAPI(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	method := r.URL.Query().Get("method")
	format := r.URL.Query().Get("format")

	if format != "json" {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(200)
		w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?><lfm status="ok"></lfm>`))
		return
	}

	var responseBody interface{}

	switch method {
	case "album.getinfo":
		responseBody = map[string]interface{}{
			"album": map[string]interface{}{
				"name":   "Test Album",
				"artist": "Test Artist",
				"image": []map[string]interface{}{
					{
						"#text": fmt.Sprintf("%s/images/album_small.jpg", m.URL()),
						"size":  "small",
					},
					{
						"#text": fmt.Sprintf("%s/images/album_large.jpg", m.URL()),
						"size":  "large",
					},
				},
			},
		}
	case "track.getinfo":
		responseBody = map[string]interface{}{
			"track": map[string]interface{}{
				"name":   "Test Track",
				"artist": map[string]interface{}{
					"name": "Test Artist",
				},
				"album": map[string]interface{}{
					"title": "Test Album",
					"image": []map[string]interface{}{
						{
							"#text": fmt.Sprintf("%s/images/track_large.jpg", m.URL()),
							"size":  "large",
						},
					},
				},
			},
		}
	default:
		responseBody = map[string]interface{}{
			"error": 6,
			"message": "Invalid method",
		}
	}

	response := MockResponse{
		StatusCode: 200,
		Body:       responseBody,
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) mockiTunesSearch(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	results := []map[string]interface{}{
		{
			"trackId":              123456789,
			"trackName":            "Test Song",
			"artistName":           "Test Artist",
			"collectionName":       "Test Album",
			"artworkUrl30":         fmt.Sprintf("%s/images/itunes_30.jpg", m.URL()),
			"artworkUrl60":         fmt.Sprintf("%s/images/itunes_60.jpg", m.URL()),
			"artworkUrl100":        fmt.Sprintf("%s/images/itunes_100.jpg", m.URL()),
			"artworkUrl500":        fmt.Sprintf("%s/images/itunes_500.jpg", m.URL()),
			"releaseDate":          "2023-01-01T00:00:00Z",
			"kind":                 "song",
			"trackPrice":           0.99,
			"currency":             "USD",
		},
	}

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"resultCount": len(results),
			"results":     results,
		},
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) mockSpotifyToken(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"access_token": "mock_spotify_access_token_12345",
			"token_type":   "Bearer",
			"expires_in":   3600,
		},
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) mockSpotifySearch(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	searchType := r.URL.Query().Get("type")

	var results interface{}

	if strings.Contains(searchType, "track") {
		results = map[string]interface{}{
			"tracks": map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"id":   "test_track_id_123",
						"name": "Test Track",
						"artists": []map[string]interface{}{
							{
								"id":   "test_artist_id_456",
								"name": "Test Artist",
							},
						},
						"album": map[string]interface{}{
							"id":   "test_album_id_789",
							"name": "Test Album",
							"images": []map[string]interface{}{
								{
									"url":    fmt.Sprintf("%s/images/spotify_640.jpg", m.URL()),
									"height": 640,
									"width":  640,
								},
								{
									"url":    fmt.Sprintf("%s/images/spotify_300.jpg", m.URL()),
									"height": 300,
									"width":  300,
								},
							},
						},
					},
				},
			},
		}
	}

	response := MockResponse{
		StatusCode: 200,
		Body:       results,
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) mockDiscogsSearch(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"results": []map[string]interface{}{
				{
					"id":    123456,
					"title": "Test Artist - Test Album",
					"type":  "release",
					"thumb": fmt.Sprintf("%s/images/discogs_thumb.jpg", m.URL()),
					"cover_image": fmt.Sprintf("%s/images/discogs_cover.jpg", m.URL()),
					"year": 2023,
				},
			},
			"pagination": map[string]interface{}{
				"page":     1,
				"pages":    1,
				"per_page": 50,
				"items":    1,
			},
		},
	}
	m.respondWithDelay(w, response)
}

// Additional mocks for other services

func (m *MockServer) mockMusixmatchSearch(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"message": map[string]interface{}{
				"header": map[string]interface{}{
					"status_code": 200,
				},
				"body": map[string]interface{}{
					"track_list": []map[string]interface{}{
						{
							"track": map[string]interface{}{
								"track_id":          123456,
								"track_name":        "Test Song",
								"artist_name":       "Test Artist",
								"album_name":        "Test Album",
								"has_lyrics":        1,
								"has_subtitles":     1,
								"has_richsync":      1,
							},
						},
					},
				},
			},
		},
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) mockMusixmatchLyrics(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	mockLyrics := `This is a test song
With mock lyrics from Musixmatch
Line by line format
For testing purposes`

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"message": map[string]interface{}{
				"header": map[string]interface{}{
					"status_code": 200,
				},
				"body": map[string]interface{}{
					"lyrics": map[string]interface{}{
						"lyrics_id":   123456,
						"lyrics_body": mockLyrics,
						"script_tracking_url": "",
						"pixel_tracking_url":  "",
						"lyrics_copyright":    "Mock Copyright",
					},
				},
			},
		},
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) mockAZLyrics(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	artist := mux.Vars(r)["artist"]
	song := mux.Vars(r)["song"]

	mockHTML := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><title>%s - %s Lyrics</title></head>
<body>
<div class="lyrics">
This is a test song<br>
With mock lyrics from AZLyrics<br>
Artist: %s<br>
Song: %s<br>
For testing purposes only
</div>
</body>
</html>`, artist, song, artist, song)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	w.Write([]byte(mockHTML))
}

func (m *MockServer) mockYifySubtitlesSearch(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	subtitles := []map[string]interface{}{
		{
			"id":        "yify_123",
			"language":  "English",
			"lang_code": "en",
			"url":       fmt.Sprintf("%s/yifysubtitles/files/subtitle_en.srt", m.URL()),
			"rating":    "good",
			"downloads": 1250,
		},
		{
			"id":        "yify_124",
			"language":  "Spanish",
			"lang_code": "es",
			"url":       fmt.Sprintf("%s/yifysubtitles/files/subtitle_es.srt", m.URL()),
			"rating":    "good",
			"downloads": 890,
		},
	}

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"success": true,
			"data":    subtitles,
		},
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) mockMusicBrainzRelease(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"id":    "test-release-id-789",
			"title": "Test Album",
			"date":  "2023-01-01",
			"cover-art-archive": map[string]interface{}{
				"artwork": true,
				"count":   1,
				"front":   true,
				"back":    false,
			},
		},
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) mockSetlistFMSearch(w http.ResponseWriter, r *http.Request) {
	body := m.readBody(r)
	m.logRequest(r, body)

	artistName := r.URL.Query().Get("artistName")

	response := MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"type":         "setlists",
			"itemsPerPage": 20,
			"page":         1,
			"total":        1,
			"setlist": []map[string]interface{}{
				{
					"id":           "test-setlist-123",
					"versionId":    "test-version-456",
					"eventDate":    "01-01-2023",
					"artist": map[string]interface{}{
						"mbid": "test-artist-mbid-789",
						"name": artistName,
					},
					"venue": map[string]interface{}{
						"id":   "test-venue-123",
						"name": "Test Venue",
						"city": map[string]interface{}{
							"id":      "test-city-456",
							"name":    "Test City",
							"country": map[string]interface{}{
								"code": "US",
								"name": "United States",
							},
						},
					},
					"sets": map[string]interface{}{
						"set": []map[string]interface{}{
							{
								"song": []map[string]interface{}{
									{
										"name": "Test Song 1",
									},
									{
										"name": "Test Song 2",
									},
									{
										"name": "Test Song 3",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	m.respondWithDelay(w, response)
}

func (m *MockServer) readBody(r *http.Request) string {
	if r.Body == nil {
		return ""
	}

	buf := make([]byte, r.ContentLength)
	r.Body.Read(buf)
	return string(buf)
}