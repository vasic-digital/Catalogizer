package handlers

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"catalogizer/database"
	"catalogizer/filesystem"
	"catalogizer/internal/services"
	root_models "catalogizer/models"
	"catalogizer/repository"
	"catalogizer/utils"

	"github.com/gin-gonic/gin"
)

// ComicPagesHandler serves individual image pages out of comic-book
// archive files so a reader client (e.g. the Android TV comic reader)
// can page through a comic without downloading the whole archive.
//
// Multi-protocol storage: a comic's file path stored in the DB is
// RELATIVE to its storage root (an SMB share path, an FTP/WebDAV/NFS
// path, or a local base-path subtree) — it is NOT an absolute local
// filesystem path. The handler therefore resolves the archive through
// the SAME filesystem.ClientFactory the StreamHandler uses, opening the
// archive over whatever protocol the storage root declares (SMB, FTP,
// NFS, WebDAV, local). Opening it with a bare local archive/zip reader
// (as an earlier version did) only worked for local-absolute paths and
// failed for every comic on an SMB/FTP/NFS/WebDAV share.
//
// Supported archive formats:
//   - .cbz  (ZIP container)  — fully supported via the stdlib
//     archive/zip reader.
//   - .cbr  (RAR container)  — NOT yet supported. The project has no
//     RAR decoder dependency, so rather than silently mis-serve or add
//     a dependency without authorisation, the handler returns an honest
//     HTTP 501 (see ListComicPages / GetComicPage). TODO: wire a rar
//     decoder (e.g. github.com/nwaples/rardecode) once approved.
//
// Page indexing is ZERO-BASED: the `index` field returned by
// ListComicPages starts at 0, and GET .../pages/:n expects the same
// 0-based n. n in the closed range [0, total_pages-1].
type ComicPagesHandler struct {
	fileRepo      *repository.MediaFileRepository
	db            *database.DB
	clientFactory filesystem.ClientFactory
}

// NewComicPagesHandler constructs the handler. It depends on:
//   - the MediaFileRepository (resolves an entity id to its concrete
//     file via GetPrimaryStreamableFile — the same path StreamEntity
//     uses);
//   - the *database.DB (looks up the file's storage root: protocol,
//     host, share, credentials — exactly as StreamHandler does);
//   - the filesystem.ClientFactory (opens the archive bytes over the
//     storage root's protocol — SMB/FTP/NFS/WebDAV/local).
func NewComicPagesHandler(fileRepo *repository.MediaFileRepository, db *database.DB, clientFactory filesystem.ClientFactory) *ComicPagesHandler {
	return &ComicPagesHandler{fileRepo: fileRepo, db: db, clientFactory: clientFactory}
}

// comicImageExts are the in-archive entry extensions treated as comic
// pages (lower-cased, leading dot included).
var comicImageExts = map[string]struct{}{
	".jpg": {}, ".jpeg": {}, ".png": {}, ".webp": {}, ".gif": {},
}

// comicContentType maps a page entry name to its image MIME type.
func comicContentType(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

// resolveArchive parses the :id param, resolves the entity's primary
// file through the same GetPrimaryStreamableFile path StreamEntity
// uses, and verifies the resolved file is a comic archive (.cbz/.cbr).
// On any failure it writes the HTTP error response and returns ok=false.
//
// The archive path is the DB-resolved item path, never derived from
// user input, so there is no path-traversal surface from the :id.
func (h *ComicPagesHandler) resolveArchive(c *gin.Context) (sf *repository.StreamableFile, ok bool) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return nil, false
	}

	primary, err := h.fileRepo.GetPrimaryStreamableFile(ctx, id)
	if err != nil {
		if err == repository.ErrNoStreamableFile {
			utils.SendErrorResponse(c, http.StatusNotFound, "No file available for this entity", err)
			return nil, false
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to resolve comic file", err)
		return nil, false
	}

	switch strings.ToLower(filepath.Ext(primary.Path)) {
	case ".cbz", ".cbr":
		return primary, true
	default:
		utils.SendErrorResponse(c, http.StatusBadRequest, "Entity is not a comic archive (.cbz/.cbr)", nil)
		return nil, false
	}
}

// isCBR reports whether the resolved archive is a RAR-backed comic.
func isCBR(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".cbr")
}

// openComicArchive resolves the comic file's storage backend (SMB / FTP /
// NFS / WebDAV / local) via the same ClientFactory the StreamHandler uses,
// obtains a random-access view of the archive bytes, and returns a
// *zip.Reader plus a cleanup func that releases every underlying resource
// (file handle / temp file / client connection). The caller MUST defer the
// returned cleanup.
//
// Two access strategies, chosen from the real client capabilities (§11.4.6),
// never guessed:
//
//   - Random-access (local + SMB): the protocol client implements
//     filesystem.SeekableClient and the handle it returns is an io.ReaderAt
//     (os.File and smb2.File both implement ReadAt). archive/zip.NewReader
//     reads only the central directory + the bytes of the page actually
//     requested — the whole archive is NEVER downloaded. Bounded memory and
//     bounded network (§12.6), which matters because comics are browsed
//     page-by-page.
//   - Buffered fallback (FTP / WebDAV / NFS, or any seekable handle that is
//     not an io.ReaderAt): the only view is a streaming io.Reader, so the
//     archive is copied to a temp file (os.CreateTemp), which IS an
//     io.ReaderAt. Bounded to disk rather than RAM (§12.6); the temp file is
//     removed on cleanup (§11.4.14).
func (h *ComicPagesHandler) openComicArchive(ctx context.Context, sf *repository.StreamableFile) (*zip.Reader, func(), error) {
	root, err := h.getStorageRootForFile(ctx, sf.FileID)
	if err != nil {
		return nil, nil, err
	}

	fsClient, err := h.clientFactory.CreateClient(&filesystem.StorageConfig{
		ID:       root.Name,
		Name:     root.Name,
		Protocol: root.Protocol,
		Settings: comicStorageRootSettings(root),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create %s filesystem client: %w", root.Protocol, err)
	}
	if err := fsClient.Connect(ctx); err != nil {
		return nil, nil, fmt.Errorf("connect to %s storage: %w", root.Protocol, err)
	}
	// The base cleanup disconnects the client; later strategies prepend their
	// own resource releases so cleanup always tears everything down in order.
	disconnect := func() { _ = fsClient.Disconnect(ctx) }

	// Strategy 1 — random-access, no full download.
	if sc, ok := fsClient.(filesystem.SeekableClient); ok {
		if rs, oerr := sc.OpenSeekable(ctx, sf.Path); oerr == nil {
			size, serr := rs.Seek(0, io.SeekEnd)
			if ra, isReaderAt := rs.(io.ReaderAt); isReaderAt && serr == nil && size > 0 {
				zr, zerr := zip.NewReader(ra, size)
				if zerr != nil {
					_ = rs.Close()
					disconnect()
					return nil, nil, fmt.Errorf("read comic zip directory: %w", zerr)
				}
				return zr, func() { _ = rs.Close(); disconnect() }, nil
			}
			// Seekable handle is not a usable io.ReaderAt (or size unknown):
			// release it and fall through to the buffered path.
			_ = rs.Close()
		}
		// OpenSeekable failed: fall through to ReadFile + buffering.
	}

	// Strategy 2 — buffer the streaming reader into a temp file.
	reader, err := fsClient.ReadFile(ctx, sf.Path)
	if err != nil {
		disconnect()
		return nil, nil, fmt.Errorf("open comic file on %s storage: %w", root.Protocol, err)
	}
	tmp, err := os.CreateTemp("", "catalogizer-comic-*.cbz")
	if err != nil {
		_ = reader.Close()
		disconnect()
		return nil, nil, fmt.Errorf("create temp comic buffer: %w", err)
	}
	// Removes the temp file AND disconnects. os.Remove after Close so the fd
	// is released first (§11.4.14 — leave no orphan temp files).
	removeAndDisconnect := func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		disconnect()
	}
	n, err := io.Copy(tmp, reader)
	_ = reader.Close()
	if err != nil {
		removeAndDisconnect()
		return nil, nil, fmt.Errorf("buffer comic file from %s storage: %w", root.Protocol, err)
	}
	// *os.File is an io.ReaderAt; ReadAt is offset-independent, so the
	// end-of-file write offset left by io.Copy does not matter here.
	zr, err := zip.NewReader(tmp, n)
	if err != nil {
		removeAndDisconnect()
		return nil, nil, fmt.Errorf("read comic zip directory: %w", err)
	}
	return zr, removeAndDisconnect, nil
}

// getStorageRootForFile looks up the storage root (protocol, host, share,
// credentials) that backs a given file id — the comic-handler analogue of
// StreamHandler.getStorageRootByName, resolved directly from the file id the
// MediaFileRepository already gave us.
func (h *ComicPagesHandler) getStorageRootForFile(ctx context.Context, fileID int64) (*root_models.StorageRoot, error) {
	query := `
		SELECT sr.id, sr.name, sr.protocol, sr.host, sr.port, sr.path, sr.username,
		       sr.password, sr.domain, sr.mount_point, sr.options, sr.url, sr.enabled,
		       sr.max_depth, sr.enable_duplicate_detection, sr.enable_metadata_extraction,
		       sr.include_patterns, sr.exclude_patterns, sr.created_at, sr.updated_at, sr.last_scan_at
		FROM files f
		JOIN storage_roots sr ON sr.id = f.storage_root_id
		WHERE f.id = ?`

	var root root_models.StorageRoot
	err := h.db.QueryRowContext(ctx, query, fileID).Scan(
		&root.ID, &root.Name, &root.Protocol, &root.Host, &root.Port, &root.Path,
		&root.Username, &root.Password, &root.Domain, &root.MountPoint, &root.Options,
		&root.URL, &root.Enabled, &root.MaxDepth, &root.EnableDuplicateDetection,
		&root.EnableMetadataExtraction, &root.IncludePatterns, &root.ExcludePatterns,
		&root.CreatedAt, &root.UpdatedAt, &root.LastScanAt,
	)
	if err != nil {
		return nil, fmt.Errorf("storage root for comic file %d not found: %w", fileID, err)
	}
	return &root, nil
}

// comicStorageRootSettings converts a StorageRoot into the settings map the
// filesystem ClientFactory expects, per protocol. It mirrors the mapping in
// StreamHandler (internal/handlers) — SMB credentials are resolved via
// services.ResolveSMBIdentity (env-var identities; §11.4.10 — the password is
// never logged here).
func comicStorageRootSettings(root *root_models.StorageRoot) map[string]interface{} {
	settings := make(map[string]interface{})

	switch root.Protocol {
	case "local":
		if root.Path != nil {
			settings["base_path"] = *root.Path
		}

	case "smb":
		if root.Host != nil {
			settings["host"] = *root.Host
		}
		if root.Port != nil {
			settings["port"] = *root.Port
		}
		if root.Path != nil {
			settings["share"] = *root.Path
		}
		user, pass, dom := services.ResolveSMBIdentity(root)
		if user != "" {
			settings["username"] = user
		}
		if pass != "" {
			settings["password"] = pass
		}
		if dom != "" {
			settings["domain"] = dom
		}
		if root.Domain != nil {
			settings["domain"] = *root.Domain
		}

	case "ftp":
		if root.Host != nil {
			settings["host"] = *root.Host
		}
		if root.Port != nil {
			settings["port"] = *root.Port
		}
		if root.Username != nil {
			settings["username"] = *root.Username
		}
		if root.Password != nil {
			settings["password"] = *root.Password
		}

	case "nfs":
		if root.Host != nil {
			settings["host"] = *root.Host
		}
		if root.Path != nil {
			settings["export_path"] = *root.Path
		}
		if root.MountPoint != nil {
			settings["mount_point"] = *root.MountPoint
		}
		if root.Options != nil {
			settings["options"] = *root.Options
		}

	case "webdav":
		if root.URL != nil {
			settings["url"] = *root.URL
		}
		if root.Username != nil {
			settings["username"] = *root.Username
		}
		if root.Password != nil {
			settings["password"] = *root.Password
		}
	}

	return settings
}

// comicPageEntries returns the image entries of a .cbz, filtered to
// page images and sorted into natural reading order. Directory entries,
// non-image entries, and macOS resource-fork junk (__MACOSX/, ._*) are
// excluded.
func comicPageEntries(zr *zip.Reader) []*zip.File {
	out := make([]*zip.File, 0, len(zr.File))
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := f.Name
		base := filepath.Base(name)
		if strings.HasPrefix(name, "__MACOSX/") || strings.HasPrefix(base, "._") {
			continue
		}
		if _, isImg := comicImageExts[strings.ToLower(filepath.Ext(name))]; !isImg {
			continue
		}
		out = append(out, f)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return naturalLess(out[i].Name, out[j].Name)
	})
	return out
}

// naturalLess compares two strings so that embedded digit runs are
// ordered numerically ("page2.png" < "page10.png"), giving correct
// comic page ordering even when filenames are not zero-padded.
func naturalLess(a, b string) bool {
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		ca, cb := a[i], b[j]
		if isDigit(ca) && isDigit(cb) {
			si, sj := i, j
			for i < len(a) && isDigit(a[i]) {
				i++
			}
			for j < len(b) && isDigit(b[j]) {
				j++
			}
			na := strings.TrimLeft(a[si:i], "0")
			nb := strings.TrimLeft(b[sj:j], "0")
			if len(na) != len(nb) {
				return len(na) < len(nb)
			}
			if na != nb {
				return na < nb
			}
			// Equal numeric value — continue with the trailing text.
			continue
		}
		if ca != cb {
			return ca < cb
		}
		i++
		j++
	}
	return (len(a) - i) < (len(b) - j)
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

// ListComicPages handles GET /api/v1/entities/:id/pages.
//
// It opens the entity's resolved .cbz archive over its storage protocol,
// lists its image page entries in natural order, and returns:
//
//	{"total_pages": N, "pages": [{"index": 0, "name": "001.png"}, ...]}
//
// For .cbr archives it returns 501 with {"error":"cbr not yet supported"}.
func (h *ComicPagesHandler) ListComicPages(c *gin.Context) {
	sf, ok := h.resolveArchive(c)
	if !ok {
		return
	}
	if isCBR(sf.Path) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "cbr not yet supported"})
		return
	}

	zr, cleanup, err := h.openComicArchive(c.Request.Context(), sf)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to open comic archive", err)
		return
	}
	defer cleanup()

	entries := comicPageEntries(zr)
	pages := make([]gin.H, 0, len(entries))
	for i, f := range entries {
		pages = append(pages, gin.H{"index": i, "name": f.Name})
	}
	c.JSON(http.StatusOK, gin.H{
		"total_pages": len(entries),
		"pages":       pages,
	})
}

// GetComicPage handles GET /api/v1/entities/:id/pages/:n.
//
// It extracts the single 0-based page n from the entity's .cbz archive
// (opened over its storage protocol) and streams its image bytes with the
// correct Content-Type. Only the requested entry is opened and streamed —
// the archive is never fully extracted, so memory stays bounded to one
// page (and, on SMB/local, only that page's bytes cross the network).
//
// For .cbr archives it returns 501 with {"error":"cbr not yet supported"}.
func (h *ComicPagesHandler) GetComicPage(c *gin.Context) {
	sf, ok := h.resolveArchive(c)
	if !ok {
		return
	}

	n, err := strconv.Atoi(c.Param("n"))
	if err != nil || n < 0 {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid page index", err)
		return
	}

	if isCBR(sf.Path) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "cbr not yet supported"})
		return
	}

	zr, cleanup, err := h.openComicArchive(c.Request.Context(), sf)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to open comic archive", err)
		return
	}
	defer cleanup()

	entries := comicPageEntries(zr)
	if n >= len(entries) {
		utils.SendErrorResponse(c, http.StatusNotFound, "Page index out of range", nil)
		return
	}

	entry := entries[n]
	rc, err := entry.Open()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to open comic page", err)
		return
	}
	defer rc.Close()

	// Stream the single entry — bounded memory, never unzips the whole
	// archive. DataFromReader sets status, Content-Type and
	// Content-Length and copies the reader to the response.
	c.DataFromReader(
		http.StatusOK,
		int64(entry.UncompressedSize64),
		comicContentType(entry.Name),
		rc,
		nil,
	)
}
