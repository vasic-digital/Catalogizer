package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
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
	rardecode "github.com/nwaples/rardecode/v2"
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
		"001.png":            red,
		"002.png":            green,
		"003.png":            blue,
		"ComicInfo.xml":      []byte("<ComicInfo/>"), // non-image -> filtered out
		"__MACOSX/._001.png": []byte("junk"),         // resource fork -> filtered out
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

// TestComicPages_CBRMissingFileErrors is the §11.4.120 reconciliation of the
// former TestComicPages_CBRNotSupported gate. .cbr is now a SUPPORTED format
// (rardecode streaming decoder), so the old "return 501 before opening" path
// is gone. A .cbr pointing at a file that does not exist must therefore fail
// with an honest 500 (failed to open the archive) — NOT a 501, NOT a panic,
// NOT a silent success (§11.4.1).
func TestComicPages_CBRMissingFileErrors(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	h := newComicHandler(t, db)
	r := newComicRouter(h)

	// .cbr stored path that does not exist on disk: resolveArchive accepts
	// the .cbr extension, then opening the archive over the (local) storage
	// protocol fails -> honest 500.
	itemID := insertComicItem(t, db, itemRepo, "/comics/does-not-exist.cbr", "issue.cbr")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages", itemID), nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
	// It must NOT be the retired 501 "not supported" response.
	assert.NotEqual(t, http.StatusNotImplemented, w.Code)
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
func (c *streamingOnlyClient) GetProtocol() string                           { return "webdav" }
func (c *streamingOnlyClient) GetConfig() interface{}                        { return nil }

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

// ---------------------------------------------------------------------------
// .cbr (RAR) hermetic fixture support (§11.4.98)
//
// No RAR-creation tool exists on the host (only `unrar`, which is
// extract-only), so the .cbr test fixture is generated byte-by-byte in pure
// Go as a minimal single-volume RAR 4.x (archive format 1.5) archive with all
// entries STORE-mode (uncompressed, METHOD 0x30). The byte layout is taken
// directly from the decoder the handler uses (github.com/nwaples/rardecode/v2,
// archive15.go) so the fixture is exactly what that decoder parses, and
// makeCBR self-validates it by decoding every entry back (CRC-checked) BEFORE
// the test relies on it — a malformed fixture fails loudly and distinctly from
// a handler bug (§11.4.6, no guessing).
// ---------------------------------------------------------------------------

// rarFile is one entry to place in the hand-rolled RAR fixture.
type rarFile struct {
	name  string
	data  []byte // file contents (ignored when isDir)
	isDir bool
}

func le16(buf *bytes.Buffer, v uint16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	buf.Write(b[:])
}

func le32(buf *bytes.Buffer, v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	buf.Write(b[:])
}

// buildRARBlock assembles one RAR 1.5 block header: a 2-byte little-endian
// header CRC (low 16 bits of CRC32-IEEE over the rest of the header) followed
// by htype(1) + flags(2 LE) + headSize(2 LE) + extra. headSize counts the
// whole header (the 7 fixed bytes + extra), NOT any file data that follows.
func buildRARBlock(htype byte, flags uint16, extra []byte) []byte {
	headSize := 7 + len(extra)
	var body bytes.Buffer
	body.WriteByte(htype)
	le16(&body, flags)
	le16(&body, uint16(headSize))
	body.Write(extra)
	crc := uint16(crc32.ChecksumIEEE(body.Bytes()))
	var out bytes.Buffer
	le16(&out, crc)
	out.Write(body.Bytes())
	return out.Bytes()
}

// buildRARFileBlock builds a FILE block (type 0x74) header + its stored file
// bytes. The block-has-data flag (0x8000) is always set so PACK_SIZE is the
// first extra field, matching the decoder's parse order. Directory entries set
// the window-mask flags (0x00e0) and carry no data.
func buildRARFileBlock(f rarFile) []byte {
	content := f.data
	var flags uint16 = 0x8000 // block-has-data (PACK_SIZE present)
	if f.isDir {
		flags |= 0x00e0 // window-mask all set => directory entry
		content = nil
	}
	name := []byte(f.name)

	var extra bytes.Buffer
	le32(&extra, uint32(len(content)))        // PACK_SIZE (store => == UNP_SIZE)
	le32(&extra, uint32(len(content)))        // UNP_SIZE
	extra.WriteByte(0x00)                     // HOST_OS
	le32(&extra, crc32.ChecksumIEEE(content)) // FILE_CRC (LE, matches leHash32)
	le32(&extra, 0x00000000)                  // FTIME (DOS time; unused on read)
	extra.WriteByte(20)                       // UNP_VER (irrelevant for store mode)
	extra.WriteByte(0x30)                     // METHOD 0x30 => store (no decoder)
	le16(&extra, uint16(len(name)))           // NAME_SIZE
	le32(&extra, 0x00000020)                  // ATTR
	extra.Write(name)                         // filename (no unicode flag => raw)

	head := buildRARBlock(0x74, flags, extra.Bytes())
	return append(head, content...)
}

// buildStoreModeRAR assembles the complete archive: marker + MAIN block +
// FILE blocks + END block.
func buildStoreModeRAR(files []rarFile) []byte {
	var out bytes.Buffer
	out.Write([]byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x00}) // "Rar!\x1A\x07\x00" marker (v1.5)
	out.Write(buildRARBlock(0x73, 0x0000, make([]byte, 6)))     // MAIN archive block (6 reserved bytes)
	for _, f := range files {
		out.Write(buildRARFileBlock(f))
	}
	out.Write(buildRARBlock(0x7b, 0x0000, nil)) // END block
	return out.Bytes()
}

// makeCBR writes a hand-rolled .cbr to <dir>/comic.cbr, after self-validating
// it: the bundled rardecode decoder must parse every entry and its CRC-checked
// contents. Entries are written in the supplied (deliberately scrambled) order
// so the test proves the handler re-sorts into natural order.
func makeCBR(t *testing.T, dir string, files []rarFile) string {
	t.Helper()
	raw := buildStoreModeRAR(files)

	// Self-validate the fixture with the SAME decoder the handler uses, so a
	// fixture defect is caught here (distinct from a handler defect).
	rr, err := rardecode.NewReader(bytes.NewReader(raw))
	require.NoError(t, err, "hand-rolled .cbr must be a valid RAR")
	seen := 0
	for {
		hdr, nerr := rr.Next()
		if nerr == io.EOF {
			break
		}
		require.NoError(t, nerr, "fixture entry header must decode")
		body, rerr := io.ReadAll(rr)
		require.NoErrorf(t, rerr, "fixture entry %q contents must decode (CRC ok)", hdr.Name)
		if !hdr.IsDir {
			require.Lenf(t, body, len(fixtureContent(files, hdr.Name)),
				"fixture entry %q must round-trip its bytes", hdr.Name)
		}
		seen++
	}
	require.Equal(t, len(files), seen, "decoder must see every fixture entry")

	path := filepath.Join(dir, "comic.cbr")
	require.NoError(t, os.WriteFile(path, raw, 0o644))
	return path
}

func fixtureContent(files []rarFile, name string) []byte {
	for _, f := range files {
		if f.name == name {
			return f.data
		}
	}
	return nil
}

// TestComicPages_CBR_ListAndExtract is the RED->GREEN test for .cbr support.
// A real, decoder-validated .cbr with 3 distinct PNG pages (plus a non-image,
// a directory, and macOS resource-fork junk) must list in natural reading
// order (page1 < page2 < page10, proving the SAME natural sort + junk filter
// as .cbz) and stream the correct, decodable page bytes.
//
// Local protocol stands in for SMB (both are the io.ReaderAt random-access
// path). Real-SMB .cbr extraction is the honest §11.4.3 on-device gap covered
// by integration, not this unit; here the bytes are resolved via the real
// local-protocol ClientFactory exactly as cbz is.
func TestComicPages_CBR_ListAndExtract(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	h := newComicHandler(t, db)
	r := newComicRouter(h)

	dir := t.TempDir()
	page1 := makePNG(t, 255, 0, 0)
	page2 := makePNG(t, 0, 255, 0)
	page10 := makePNG(t, 0, 0, 255)

	// Deliberately scrambled order + junk entries, mirroring makeCBZ's intent.
	files := []rarFile{
		{name: "page2.png", data: page2},
		{name: "info.txt", data: []byte("not an image -> filtered")},
		{name: "page10.png", data: page10},
		{name: "__MACOSX/page1.png", data: []byte("resource fork -> filtered")},
		{name: "._page1.png", data: []byte("dotfile -> filtered")},
		{name: "extras", isDir: true}, // directory -> filtered
		{name: "page1.png", data: page1},
	}
	cbrPath := makeCBR(t, dir, files)

	itemID := insertComicItem(t, db, itemRepo, cbrPath, "comic.cbr")

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
		assert.Equal(t, []string{"page1.png", "page2.png", "page10.png"},
			[]string{resp.Pages[0].Name, resp.Pages[1].Name, resp.Pages[2].Name})
		assert.Equal(t, 0, resp.Pages[0].Index)
		assert.Equal(t, 1, resp.Pages[1].Index)
		assert.Equal(t, 2, resp.Pages[2].Index)
	})

	t.Run("page 0 streams first image bytes (decodable PNG, 0-indexed)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages/0", itemID), nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
		assert.Equal(t, page1, w.Body.Bytes())

		img, format, derr := image.Decode(bytes.NewReader(w.Body.Bytes()))
		require.NoError(t, derr, "served page 0 must be a decodable image")
		assert.Equal(t, "png", format)
		assert.Greater(t, img.Bounds().Dx(), 0)
		assert.Greater(t, img.Bounds().Dy(), 0)
	})

	t.Run("page 2 (last) streams correct bytes", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/pages/2", itemID), nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
		assert.Equal(t, page10, w.Body.Bytes())
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
