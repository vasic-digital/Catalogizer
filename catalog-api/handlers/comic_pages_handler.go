package handlers

import (
	"archive/zip"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"catalogizer/repository"
	"catalogizer/utils"

	"github.com/gin-gonic/gin"
)

// ComicPagesHandler serves individual image pages out of comic-book
// archive files so a reader client (e.g. the Android TV comic reader)
// can page through a comic without downloading the whole archive.
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
	fileRepo *repository.MediaFileRepository
}

// NewComicPagesHandler constructs the handler. It depends only on the
// MediaFileRepository — the same repository StreamEntity uses to
// resolve an entity id to its concrete on-disk file via
// GetPrimaryStreamableFile.
func NewComicPagesHandler(fileRepo *repository.MediaFileRepository) *ComicPagesHandler {
	return &ComicPagesHandler{fileRepo: fileRepo}
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
// file path through the same GetPrimaryStreamableFile path StreamEntity
// uses, and verifies the resolved file is a comic archive (.cbz/.cbr).
// On any failure it writes the HTTP error response and returns ok=false.
//
// The archive path is the DB-resolved item path, never derived from
// user input, so there is no path-traversal surface from the :id.
func (h *ComicPagesHandler) resolveArchive(c *gin.Context) (path string, ok bool) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return "", false
	}

	primary, err := h.fileRepo.GetPrimaryStreamableFile(ctx, id)
	if err != nil {
		if err == repository.ErrNoStreamableFile {
			utils.SendErrorResponse(c, http.StatusNotFound, "No file available for this entity", err)
			return "", false
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to resolve comic file", err)
		return "", false
	}

	switch strings.ToLower(filepath.Ext(primary.Path)) {
	case ".cbz", ".cbr":
		return primary.Path, true
	default:
		utils.SendErrorResponse(c, http.StatusBadRequest, "Entity is not a comic archive (.cbz/.cbr)", nil)
		return "", false
	}
}

// isCBR reports whether the resolved archive is a RAR-backed comic.
func isCBR(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".cbr")
}

// comicPageEntries returns the image entries of a .cbz, filtered to
// page images and sorted into natural reading order. Directory entries,
// non-image entries, and macOS resource-fork junk (__MACOSX/, ._*) are
// excluded.
func comicPageEntries(zr *zip.ReadCloser) []*zip.File {
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
// It opens the entity's resolved .cbz archive, lists its image page
// entries in natural order, and returns:
//
//	{"total_pages": N, "pages": [{"index": 0, "name": "001.png"}, ...]}
//
// For .cbr archives it returns 501 with {"error":"cbr not yet supported"}.
func (h *ComicPagesHandler) ListComicPages(c *gin.Context) {
	path, ok := h.resolveArchive(c)
	if !ok {
		return
	}
	if isCBR(path) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "cbr not yet supported"})
		return
	}

	zr, err := zip.OpenReader(path)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to open comic archive", err)
		return
	}
	defer zr.Close()

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
// and streams its image bytes with the correct Content-Type. Only the
// requested entry is opened and streamed — the archive is never fully
// extracted, so memory stays bounded to one page.
//
// For .cbr archives it returns 501 with {"error":"cbr not yet supported"}.
func (h *ComicPagesHandler) GetComicPage(c *gin.Context) {
	path, ok := h.resolveArchive(c)
	if !ok {
		return
	}

	n, err := strconv.Atoi(c.Param("n"))
	if err != nil || n < 0 {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid page index", err)
		return
	}

	if isCBR(path) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "cbr not yet supported"})
		return
	}

	zr, err := zip.OpenReader(path)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to open comic archive", err)
		return
	}
	defer zr.Close()

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
