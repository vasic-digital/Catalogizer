package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"catalogizer/database"
	"catalogizer/filesystem"
	"catalogizer/internal/media/models"
	"catalogizer/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makePNG encodes a guaranteed-valid 1x1 PNG of the given colour so
// each comic page has distinct, verifiable bytes.
func makePNG(t *testing.T, r, g, b uint8) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: r, G: g, B: b, A: 255})
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

// makeCBZ writes a .cbz (zip) archive on disk. entries are written in
// the supplied (deliberately scrambled) order so the test proves the
// handler re-sorts pages into natural order rather than echoing the
// archive's table-of-contents order.
func makeCBZ(t *testing.T, dir string, ordered []string, data map[string][]byte) string {
	t.Helper()
	path := filepath.Join(dir, "comic.cbz")
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	zw := zip.NewWriter(f)
	for _, name := range ordered {
		w, err := zw.Create(name)
		require.NoError(t, err)
		_, err = w.Write(data[name])
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	return path
}

// insertComicItem creates a media item linked to a single file at
// filePath on a LOCAL storage root whose base_path is "/" (so filePath is
// an absolute local path). Mirrors how the scanner links files ->
// media_files -> media_items, the same join StreamEntity resolves through
// GetPrimaryStreamableFile.
func insertComicItem(t *testing.T, db *database.DB, itemRepo *repository.MediaItemRepository, filePath, fileName string) int64 {
	t.Helper()
	return insertComicItemOnRoot(t, db, itemRepo, "local", "/", filePath, fileName)
}

// insertComicItemOnRoot is the parametrized core: it inserts a storage root
// with the given protocol + base path, a file whose stored path is filePath
// (which may be RELATIVE to the storage root's base — exactly how an SMB
// share / FTP / WebDAV / NFS path is stored), and links it to a new media
// item. Returns the media item id.
func insertComicItemOnRoot(t *testing.T, db *database.DB, itemRepo *repository.MediaItemRepository, protocol, rootPath, filePath, fileName string) int64 {
	t.Helper()
	ctx := context.Background()

	_, typeID, err := itemRepo.GetMediaTypeByName(ctx, "movie")
	require.NoError(t, err)

	itemID, err := itemRepo.Create(ctx, &models.MediaItem{
		MediaTypeID: typeID, Title: fileName, Status: "detected",
	})
	require.NoError(t, err)

	// Unique root name per (fileName, rootPath) so multiple roots can
	// coexist in one test DB.
	rootName := fmt.Sprintf("comic-root-%s-%d", fileName, itemID)
	_, err = db.ExecContext(ctx,
		`INSERT INTO storage_roots (name, protocol, path) VALUES (?, ?, ?)`,
		rootName, protocol, rootPath)
	require.NoError(t, err)
	var rootID int64
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT id FROM storage_roots WHERE name = ?`, rootName).Scan(&rootID))

	_, err = db.ExecContext(ctx,
		`INSERT INTO files (storage_root_id, path, name, size, mime_type, modified_at) VALUES (?, ?, ?, ?, ?, datetime('now'))`,
		rootID, filePath, fileName, 1234, "application/octet-stream")
	require.NoError(t, err)
	var fileID int64
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT id FROM files WHERE storage_root_id = ? AND path = ?`, rootID, filePath).Scan(&fileID))

	_, err = db.ExecContext(ctx,
		`INSERT INTO media_files (media_item_id, file_id) VALUES (?, ?)`, itemID, fileID)
	require.NoError(t, err)

	return itemID
}

func newComicRouter(h *ComicPagesHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/v1/entities")
	g.GET("/:id/pages", h.ListComicPages)
	g.GET("/:id/pages/:n", h.GetComicPage)
	return r
}

func newComicHandler(t *testing.T, db *database.DB) *ComicPagesHandler {
	t.Helper()
	return NewComicPagesHandler(repository.NewMediaFileRepository(db), db, filesystem.NewDefaultClientFactory())
}

// TestComicPages_ListAndExtract is the primary RED->GREEN test: a real
// .cbz with 3 distinct PNG pages (plus junk entries) must list in
// natural order and stream the correct page bytes.
func TestComicPages_ListAndExtract(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	h := newComicHandler(t, db)
	r := newComicRouter(h)

	dir := t.TempDir()
	red := makePNG(t, 255, 0, 0)
	green := makePNG(t, 0, 255, 0)
	blue := makePNG(t, 0, 0, 255)
	pageData := map[string][]byte{
		"001.png":           red,
		"002.png":           green,
		"003.png":           blue,
		"ComicInfo.xml":     []byte("<ComicInfo/>"), // non-image -> filtered out
		"__MACOSX/._001.png": []byte("junk"),        // resource fork -> filtered out
	}
	// Deliberately scrambled TOC order to prove natural sort.
	cbzPath := makeCBZ(t, dir, []string{"002.png", "ComicInfo.xml", "003.png", "__MACOSX/._001.png", "001.png"}, pageData)

	itemID := insertComicItem(t, db, itemRepo, cbzPath, "comic.cbz")

	t.Run("list returns 3 pages in natural order", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages", itemID), nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var resp struct {
			TotalPages int `json:"total_pages"`
			Pages      []struct {
				Index int    `json:"index"`
				Name  string `json:"name"`
			} `json:"pages"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, 3, resp.TotalPages)
		require.Len(t, resp.Pages, 3)
		assert.Equal(t, []string{"001.png", "002.png", "003.png"},
			[]string{resp.Pages[0].Name, resp.Pages[1].Name, resp.Pages[2].Name})
		assert.Equal(t, 0, resp.Pages[0].Index)
		assert.Equal(t, 1, resp.Pages[1].Index)
		assert.Equal(t, 2, resp.Pages[2].Index)
	})

	t.Run("page 0 streams first image bytes (0-indexed)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages/0", itemID), nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
		assert.Equal(t, red, w.Body.Bytes())
	})

	t.Run("page 1 streams second image bytes", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages/1", itemID), nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
		assert.Equal(t, green, w.Body.Bytes())
	})

	t.Run("out-of-range page -> 404", func(t *testing.T) {
		for _, n := range []string{"3", "99"} {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages/%s", itemID, n), nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code, "n=%s", n)
		}
	})

	t.Run("invalid page index -> 400", func(t *testing.T) {
		for _, n := range []string{"abc", "-1"} {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages/%s", itemID, n), nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code, "n=%s", n)
		}
	})
}

// TestComicPages_NonComicItem proves a non-archive item is rejected
// rather than mis-served.
func TestComicPages_NonComicItem(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	h := newComicHandler(t, db)
	r := newComicRouter(h)

	itemID := insertComicItem(t, db, itemRepo, "/quality/movie.mkv", "movie.mkv")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages", itemID), nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}

// TestComicPages_NoFiles proves an entity with no linked files yields 404.
func TestComicPages_NoFiles(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	h := newComicHandler(t, db)
	r := newComicRouter(h)

	ctx := context.Background()
	_, typeID, err := itemRepo.GetMediaTypeByName(ctx, "movie")
	require.NoError(t, err)
	itemID, err := itemRepo.Create(ctx, &models.MediaItem{
		MediaTypeID: typeID, Title: "orphan", Status: "detected",
	})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages", itemID), nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
}

// TestComicPages_CBRNotSupported proves the honest 501 for .cbr until a
// rar decoder is wired in.
func TestComicPages_CBRNotSupported(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	h := newComicHandler(t, db)
	r := newComicRouter(h)

	// The .cbr file need not exist on disk: the handler returns 501
	// before attempting to open it.
	itemID := insertComicItem(t, db, itemRepo, "/comics/issue.cbr", "issue.cbr")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages", itemID), nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotImplemented, w.Code, w.Body.String())
	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "cbr not yet supported", resp["error"])
}

// TestComicPages_StorageRootRelativePath is the RED->GREEN test for the
// multi-protocol fix. The comic's stored file path is RELATIVE to its
// storage root's base — exactly how an SMB share path (the real-world
// 2192 cbr + 1230 cbz on the 192.168.0.241:DATA8 share) and FTP/WebDAV/NFS
// paths are stored. The previous handler called archive/zip.OpenReader on
// the bare relative path and failed ("no such file or directory"); the
// factory-routed handler resolves it against the storage root and opens it.
//
// Local protocol stands in for SMB here (both are SeekableClient -> the
// io.ReaderAt random-access path) because a real SMB server is unavailable
// in a unit test (honest §11.4.3 gap — real-SMB extraction is verified
// on-device / in integration, not here).
func TestComicPages_StorageRootRelativePath(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	h := newComicHandler(t, db)
	r := newComicRouter(h)

	base := t.TempDir()
	red := makePNG(t, 255, 0, 0)
	green := makePNG(t, 0, 255, 0)
	pageData := map[string][]byte{"001.png": red, "002.png": green}
	// makeCBZ writes <base>/comic.cbz; the STORED path is the bare relative
	// "comic.cbz", resolved against the storage root's base_path (= base).
	makeCBZ(t, base, []string{"002.png", "001.png"}, pageData)

	itemID := insertComicItemOnRoot(t, db, itemRepo, "local", base, "comic.cbz", "comic.cbz")

	t.Run("list resolves relative path via factory", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages", itemID), nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var resp struct {
			TotalPages int `json:"total_pages"`
			Pages      []struct {
				Index int    `json:"index"`
				Name  string `json:"name"`
			} `json:"pages"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, 2, resp.TotalPages)
		require.Len(t, resp.Pages, 2)
		assert.Equal(t, []string{"001.png", "002.png"},
			[]string{resp.Pages[0].Name, resp.Pages[1].Name})
	})

	t.Run("extract page 0 over factory client", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages/0", itemID), nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
		assert.Equal(t, red, w.Body.Bytes())
	})
}

// streamingOnlyClient is a fake FileSystemClient — permitted in unit tests
// per §11.4.27 — that exposes ONLY a streaming io.Reader via ReadFile and
// deliberately does NOT implement filesystem.SeekableClient. It models the
// FTP / WebDAV / NFS reality (no random access), driving the comic handler's
// temp-file buffering fallback (§12.6 bounded memory, §11.4.14 cleanup).
type streamingOnlyClient struct{ data []byte }

func (c *streamingOnlyClient) Connect(context.Context) error        { return nil }
func (c *streamingOnlyClient) Disconnect(context.Context) error     { return nil }
func (c *streamingOnlyClient) IsConnected() bool                    { return true }
func (c *streamingOnlyClient) TestConnection(context.Context) error { return nil }
func (c *streamingOnlyClient) ReadFile(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(c.data)), nil
}
func (c *streamingOnlyClient) WriteFile(context.Context, string, io.Reader) error { return nil }
func (c *streamingOnlyClient) GetFileInfo(context.Context, string) (*filesystem.FileInfo, error) {
	return nil, nil
}
func (c *streamingOnlyClient) FileExists(context.Context, string) (bool, error) { return true, nil }
func (c *streamingOnlyClient) DeleteFile(context.Context, string) error         { return nil }
func (c *streamingOnlyClient) CopyFile(context.Context, string, string) error   { return nil }
func (c *streamingOnlyClient) ListDirectory(context.Context, string) ([]*filesystem.FileInfo, error) {
	return nil, nil
}
func (c *streamingOnlyClient) CreateDirectory(context.Context, string) error { return nil }
func (c *streamingOnlyClient) DeleteDirectory(context.Context, string) error { return nil }
func (c *streamingOnlyClient) GetProtocol() string                          { return "webdav" }
func (c *streamingOnlyClient) GetConfig() interface{}                       { return nil }

// streamingOnlyFactory returns streamingOnlyClient for any protocol, so the
// handler's CreateClient call yields the streaming-only (non-seekable) client.
type streamingOnlyFactory struct{ data []byte }

func (f *streamingOnlyFactory) CreateClient(*filesystem.StorageConfig) (filesystem.FileSystemClient, error) {
	return &streamingOnlyClient{data: f.data}, nil
}
func (f *streamingOnlyFactory) SupportedProtocols() []string { return []string{"webdav"} }

// TestComicPages_BufferedFallback exercises the temp-file buffering path for
// streaming-only protocols (FTP/WebDAV/NFS): a client that gives only an
// io.Reader. Proves the handler buffers the archive to a temp file, reads
// pages out of it, and serves the correct bytes.
func TestComicPages_BufferedFallback(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)

	// Build a real .cbz on disk, then hand its bytes to the streaming-only
	// fake so ReadFile streams them (no seek, no random access).
	dir := t.TempDir()
	red := makePNG(t, 255, 0, 0)
	green := makePNG(t, 0, 255, 0)
	cbzPath := makeCBZ(t, dir, []string{"002.png", "001.png"}, map[string][]byte{"001.png": red, "002.png": green})
	cbzBytes, err := os.ReadFile(cbzPath)
	require.NoError(t, err)

	h := NewComicPagesHandler(repository.NewMediaFileRepository(db), db, &streamingOnlyFactory{data: cbzBytes})
	r := newComicRouter(h)

	// A webdav storage root with a relative path; the fake factory ignores
	// the config and always returns the streaming-only client.
	itemID := insertComicItemOnRoot(t, db, itemRepo, "webdav", "/dav", "comic.cbz", "comic.cbz")

	t.Run("list via buffered temp file", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages", itemID), nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var resp struct {
			TotalPages int `json:"total_pages"`
			Pages      []struct {
				Index int    `json:"index"`
				Name  string `json:"name"`
			} `json:"pages"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, 2, resp.TotalPages)
		require.Len(t, resp.Pages, 2)
		assert.Equal(t, []string{"001.png", "002.png"},
			[]string{resp.Pages[0].Name, resp.Pages[1].Name})
	})

	t.Run("extract page 1 via buffered temp file", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages/1", itemID), nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
		assert.Equal(t, green, w.Body.Bytes())
	})
}
