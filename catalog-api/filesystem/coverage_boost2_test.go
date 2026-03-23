//go:build linux
// +build linux

package filesystem

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// WebDAV Client — ReadFile paths via httptest (Boost2)
// =============================================================================

func TestWebDAVClient_ReadFile_SuccessB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "GET":
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("file content here"))
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	reader, err := client.ReadFile(ctx, "test.txt")
	require.NoError(t, err)
	defer reader.Close()

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "file content here", string(data))
}

func TestWebDAVClient_ReadFile_NotFoundB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "GET":
			w.WriteHeader(http.StatusNotFound)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	_, err := client.ReadFile(ctx, "missing.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "returned status 404")
}

func TestWebDAVClient_ReadFile_AuthVerificationB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "GET":
			user, pass, ok := r.BasicAuth()
			if !ok || user != "testuser" || pass != "testpass" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated content"))
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	reader, err := client.ReadFile(ctx, "secure.txt")
	require.NoError(t, err)
	defer reader.Close()

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "authenticated content", string(data))
}

// =============================================================================
// WebDAV Client — WriteFile paths (Boost2)
// =============================================================================

func TestWebDAVClient_WriteFile_SuccessB2(t *testing.T) {
	var receivedBody string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "PUT":
			body, _ := io.ReadAll(r.Body)
			receivedBody = string(body)
			w.WriteHeader(http.StatusCreated)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	data := strings.NewReader("upload data")
	err := client.WriteFile(ctx, "upload.txt", data)
	require.NoError(t, err)
	assert.Equal(t, "upload data", receivedBody)
}

func TestWebDAVClient_WriteFile_ServerErrorB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "PUT":
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	data := strings.NewReader("some data")
	err := client.WriteFile(ctx, "fail.txt", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "returned status 500")
}

func TestWebDAVClient_WriteFile_NotConnectedB2(t *testing.T) {
	config := &WebDAVConfig{URL: "http://localhost:0"}
	client := NewWebDAVClient(config)
	ctx := context.Background()

	data := strings.NewReader("some data")
	err := client.WriteFile(ctx, "file.txt", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebDAVClient_WriteFile_AuthB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "PUT":
			user, pass, ok := r.BasicAuth()
			if !ok || user != "testuser" || pass != "testpass" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusCreated)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	err := client.WriteFile(ctx, "auth.txt", strings.NewReader("secure data"))
	assert.NoError(t, err)
}

// =============================================================================
// WebDAV Client — GetFileInfo paths (Boost2)
// =============================================================================

func TestWebDAVClient_GetFileInfo_SuccessB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "HEAD":
			w.Header().Set("Content-Length", "12345")
			w.Header().Set("Last-Modified", time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC).Format(time.RFC1123))
			w.Header().Set("Content-Type", "video/mp4")
			w.WriteHeader(http.StatusOK)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	info, err := client.GetFileInfo(ctx, "video.mp4")
	require.NoError(t, err)
	assert.Equal(t, "video.mp4", info.Name)
	assert.Equal(t, int64(12345), info.Size)
	assert.False(t, info.IsDir)
}

func TestWebDAVClient_GetFileInfo_DirectoryB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "HEAD":
			w.Header().Set("Content-Type", "httpd/unix-directory")
			w.WriteHeader(http.StatusOK)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	info, err := client.GetFileInfo(ctx, "somedir")
	require.NoError(t, err)
	assert.True(t, info.IsDir)
}

func TestWebDAVClient_GetFileInfo_DirectorySlashB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "HEAD":
			w.WriteHeader(http.StatusOK)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	info, err := client.GetFileInfo(ctx, "somedir/")
	require.NoError(t, err)
	assert.True(t, info.IsDir)
}

func TestWebDAVClient_GetFileInfo_NotFoundB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "HEAD":
			w.WriteHeader(http.StatusNotFound)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	_, err := client.GetFileInfo(ctx, "missing.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "returned status 404")
}

func TestWebDAVClient_GetFileInfo_NoContentLengthB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "HEAD":
			w.WriteHeader(http.StatusOK)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	info, err := client.GetFileInfo(ctx, "nosize.txt")
	require.NoError(t, err)
	assert.Equal(t, int64(0), info.Size)
}

func TestWebDAVClient_GetFileInfo_InvalidLastModifiedB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "HEAD":
			w.Header().Set("Last-Modified", "not-a-date")
			w.WriteHeader(http.StatusOK)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	info, err := client.GetFileInfo(ctx, "baddate.txt")
	require.NoError(t, err)
	assert.True(t, time.Since(info.ModTime) < 5*time.Second)
}

func TestWebDAVClient_GetFileInfo_AuthB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "HEAD":
			user, _, ok := r.BasicAuth()
			if !ok || user != "testuser" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Length", "5000")
			w.WriteHeader(http.StatusOK)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	info, err := client.GetFileInfo(ctx, "secure.txt")
	require.NoError(t, err)
	assert.Equal(t, int64(5000), info.Size)
}

// =============================================================================
// WebDAV Client — FileExists paths (Boost2)
// =============================================================================

func TestWebDAVClient_FileExists_TrueB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "HEAD":
			w.WriteHeader(http.StatusOK)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	exists, err := client.FileExists(ctx, "existing.txt")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestWebDAVClient_FileExists_FalseB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "HEAD":
			w.WriteHeader(http.StatusNotFound)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	exists, err := client.FileExists(ctx, "missing.txt")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestWebDAVClient_FileExists_AuthB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "HEAD":
			user, _, ok := r.BasicAuth()
			if !ok || user != "testuser" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	exists, err := client.FileExists(ctx, "secure.txt")
	require.NoError(t, err)
	assert.True(t, exists)
}

// =============================================================================
// WebDAV Client — CreateDirectory, DeleteDirectory, DeleteFile (Boost2)
// =============================================================================

func TestWebDAVClient_CreateDirectory_SuccessB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "MKCOL":
			w.WriteHeader(http.StatusCreated)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	err := client.CreateDirectory(ctx, "newdir")
	assert.NoError(t, err)
}

func TestWebDAVClient_CreateDirectory_FailureB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "MKCOL":
			w.WriteHeader(http.StatusConflict)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	err := client.CreateDirectory(ctx, "existing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "returned status 409")
}

func TestWebDAVClient_CreateDirectory_AuthB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "MKCOL":
			user, _, ok := r.BasicAuth()
			if !ok || user != "testuser" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusCreated)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	err := client.CreateDirectory(ctx, "securedir")
	assert.NoError(t, err)
}

func TestWebDAVClient_DeleteDirectory_SuccessB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "DELETE":
			w.WriteHeader(http.StatusNoContent)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	err := client.DeleteDirectory(ctx, "olddir")
	assert.NoError(t, err)
}

func TestWebDAVClient_DeleteDirectory_FailureB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "DELETE":
			w.WriteHeader(http.StatusForbidden)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	err := client.DeleteDirectory(ctx, "protected")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "returned status 403")
}

func TestWebDAVClient_DeleteDirectory_AuthB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "DELETE":
			user, _, ok := r.BasicAuth()
			if !ok || user != "testuser" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	err := client.DeleteDirectory(ctx, "securedir")
	assert.NoError(t, err)
}

func TestWebDAVClient_DeleteFile_SuccessB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "DELETE":
			w.WriteHeader(http.StatusNoContent)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	err := client.DeleteFile(ctx, "remove.txt")
	assert.NoError(t, err)
}

func TestWebDAVClient_DeleteFile_FailureB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "DELETE":
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	err := client.DeleteFile(ctx, "fail.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "returned status 500")
}

func TestWebDAVClient_DeleteFile_AuthB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "DELETE":
			user, _, ok := r.BasicAuth()
			if !ok || user != "testuser" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	err := client.DeleteFile(ctx, "securefile.txt")
	assert.NoError(t, err)
}

// =============================================================================
// WebDAV Client — CopyFile (Boost2)
// =============================================================================

func TestWebDAVClient_CopyFile_SuccessB2(t *testing.T) {
	var receivedDest string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "COPY":
			receivedDest = r.Header.Get("Destination")
			w.WriteHeader(http.StatusCreated)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	err := client.CopyFile(ctx, "src.txt", "dst.txt")
	assert.NoError(t, err)
	assert.NotEmpty(t, receivedDest)
}

func TestWebDAVClient_CopyFile_FailureB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "COPY":
			w.WriteHeader(http.StatusForbidden)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	err := client.CopyFile(ctx, "src.txt", "dst.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "returned status 403")
}

func TestWebDAVClient_CopyFile_AuthB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.WriteHeader(http.StatusMultiStatus)
		case "COPY":
			user, _, ok := r.BasicAuth()
			if !ok || user != "testuser" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusCreated)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	err := client.CopyFile(ctx, "src.txt", "dst.txt")
	assert.NoError(t, err)
}

// =============================================================================
// WebDAV Client — ListDirectory XML parsing (Boost2)
// =============================================================================

func TestWebDAVClient_ListDirectory_SuccessB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			if r.Header.Get("Depth") == "0" {
				w.WriteHeader(http.StatusMultiStatus)
				return
			}
			w.WriteHeader(http.StatusMultiStatus)
			body := `<?xml version="1.0" encoding="utf-8"?>
<D:multistatus xmlns:D="DAV:">
<D:response>
<D:href>/</D:href>
<D:propstat><D:prop>
<D:displayname>root</D:displayname>
<D:resourcetype><D:collection/></D:resourcetype>
</D:prop></D:propstat>
</D:response>
<D:response>
<D:href>/file1.txt</D:href>
<D:propstat><D:prop>
<D:displayname>file1.txt</D:displayname>
<D:getcontentlength>1024</D:getcontentlength>
<D:getlastmodified>Sun, 15 Jun 2025 10:30:00 GMT</D:getlastmodified>
<D:resourcetype/>
</D:prop></D:propstat>
</D:response>
<D:response>
<D:href>/subdir/</D:href>
<D:propstat><D:prop>
<D:displayname>subdir</D:displayname>
<D:resourcetype><D:collection/></D:resourcetype>
</D:prop></D:propstat>
</D:response>
</D:multistatus>`
			w.Write([]byte(body))
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	files, err := client.ListDirectory(ctx, "/")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), 1)
}

func TestWebDAVClient_ListDirectory_BadStatusB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			if r.Header.Get("Depth") == "0" {
				w.WriteHeader(http.StatusMultiStatus)
				return
			}
			w.WriteHeader(http.StatusForbidden)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	_, err := client.ListDirectory(ctx, "/")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "returned status 403")
}

func TestWebDAVClient_ListDirectory_NotConnectedB2(t *testing.T) {
	config := &WebDAVConfig{URL: "http://localhost:0"}
	client := NewWebDAVClient(config)
	ctx := context.Background()

	_, err := client.ListDirectory(ctx, "/")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestWebDAVClient_ListDirectory_DirResourceTypeB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			if r.Header.Get("Depth") == "0" {
				w.WriteHeader(http.StatusMultiStatus)
				return
			}
			w.WriteHeader(http.StatusMultiStatus)
			body := `<?xml version="1.0" encoding="utf-8"?>
<D:multistatus xmlns:D="DAV:">
<D:response>
<D:href>/mydir/</D:href>
<D:propstat><D:prop>
<D:displayname>mydir</D:displayname>
<D:resourcetype><D:directory/></D:resourcetype>
</D:prop></D:propstat>
</D:response>
</D:multistatus>`
			w.Write([]byte(body))
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	files, err := client.ListDirectory(ctx, "/")
	require.NoError(t, err)
	for _, f := range files {
		if f.Name == "mydir" {
			assert.True(t, f.IsDir)
		}
	}
}

func TestWebDAVClient_ListDirectory_WithSizeAndDateB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			if r.Header.Get("Depth") == "0" {
				w.WriteHeader(http.StatusMultiStatus)
				return
			}
			w.WriteHeader(http.StatusMultiStatus)
			body := `<?xml version="1.0" encoding="utf-8"?>
<D:multistatus xmlns:D="DAV:">
<D:response>
<D:href>/data/</D:href>
<D:propstat><D:prop>
<D:displayname>data</D:displayname>
<D:resourcetype><D:collection/></D:resourcetype>
</D:prop></D:propstat>
</D:response>
<D:response>
<D:href>/data/movie.mkv</D:href>
<D:propstat><D:prop>
<D:displayname>movie.mkv</D:displayname>
<D:getcontentlength>1073741824</D:getcontentlength>
<D:getlastmodified>Sun, 15 Jun 2025 10:30:00 UTC</D:getlastmodified>
<D:resourcetype/>
</D:prop></D:propstat>
</D:response>
<D:response>
<D:href>/data/photos/</D:href>
<D:propstat><D:prop>
<D:displayname>photos</D:displayname>
<D:resourcetype><D:collection/></D:resourcetype>
</D:prop></D:propstat>
</D:response>
</D:multistatus>`
			w.Write([]byte(body))
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	files, err := client.ListDirectory(ctx, "/data/")
	require.NoError(t, err)

	for _, f := range files {
		if f.Name == "movie.mkv" {
			assert.Equal(t, int64(1073741824), f.Size)
			assert.False(t, f.IsDir)
		}
		if f.Name == "photos" {
			assert.True(t, f.IsDir)
		}
	}
}

func TestWebDAVClient_ListDirectory_EmptyResponseB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			if r.Header.Get("Depth") == "0" {
				w.WriteHeader(http.StatusMultiStatus)
				return
			}
			w.WriteHeader(http.StatusMultiStatus)
			body := `<?xml version="1.0" encoding="utf-8"?>
<D:multistatus xmlns:D="DAV:">
</D:multistatus>`
			w.Write([]byte(body))
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	files, err := client.ListDirectory(ctx, "/empty/")
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestWebDAVClient_ListDirectory_AuthB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			user, _, ok := r.BasicAuth()
			if !ok || user != "testuser" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if r.Header.Get("Depth") == "0" {
				w.WriteHeader(http.StatusMultiStatus)
				return
			}
			w.WriteHeader(http.StatusMultiStatus)
			body := `<?xml version="1.0" encoding="utf-8"?>
<D:multistatus xmlns:D="DAV:">
<D:response>
<D:href>/authfile.txt</D:href>
<D:propstat><D:prop>
<D:displayname>authfile.txt</D:displayname>
<D:getcontentlength>42</D:getcontentlength>
<D:resourcetype/>
</D:prop></D:propstat>
</D:response>
</D:multistatus>`
			w.Write([]byte(body))
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	files, err := client.ListDirectory(ctx, "/")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), 1)
}

// =============================================================================
// WebDAV Client — resolveURL edge cases (Boost2)
// =============================================================================

func TestWebDAVClient_ResolveURL_TraversalPreventionB2(t *testing.T) {
	config := &WebDAVConfig{
		URL:  "http://localhost:8080",
		Path: "/webdav",
	}
	client := NewWebDAVClient(config)

	result := client.resolveURL("../../../etc/passwd")
	assert.NotContains(t, result, "..")
}

func TestWebDAVClient_ResolveURL_CleanPathB2(t *testing.T) {
	config := &WebDAVConfig{
		URL: "http://localhost:8080",
	}
	client := NewWebDAVClient(config)

	result := client.resolveURL("dir/./subdir/../file.txt")
	assert.Contains(t, result, "dir")
}

// =============================================================================
// WebDAV Client — TestConnection and constructors (Boost2)
// =============================================================================

func TestWebDAVClient_TestConnection_SuccessB2(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PROPFIND" {
			w.WriteHeader(http.StatusMultiStatus)
		}
	})
	_, client := newTestWebDAVServer(t, handler)
	ctx := context.Background()

	err := client.TestConnection(ctx)
	assert.NoError(t, err)
}

func TestNewWebDAVClient_WithPathB2(t *testing.T) {
	config := &WebDAVConfig{
		URL:      "http://dav.example.com",
		Username: "admin",
		Password: "pass",
		Path:     "/files/media",
	}

	client := NewWebDAVClient(config)
	assert.NotNil(t, client)
	assert.Equal(t, "webdav", client.GetProtocol())
	assert.Equal(t, config, client.GetConfig().(*WebDAVConfig))
	assert.NotNil(t, client.baseURL)
	assert.Equal(t, "/files/media", client.baseURL.Path)
}

func TestNewWebDAVClient_WithRootPathB2(t *testing.T) {
	config := &WebDAVConfig{
		URL:  "http://dav.example.com",
		Path: "/",
	}
	client := NewWebDAVClient(config)
	assert.NotNil(t, client)
}

func TestNewWebDAVClient_WithEmptyPathB2(t *testing.T) {
	config := &WebDAVConfig{
		URL: "http://dav.example.com",
	}
	client := NewWebDAVClient(config)
	assert.NotNil(t, client)
}

func TestWebDAVClient_Connect_ServerDownB2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	serverURL := server.URL
	server.Close()

	config := &WebDAVConfig{URL: serverURL}
	client := NewWebDAVClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
}

func TestWebDAVClient_ReadFile_NetworkErrorB2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PROPFIND" {
			w.WriteHeader(http.StatusMultiStatus)
			return
		}
	}))
	config := &WebDAVConfig{URL: server.URL}
	client := NewWebDAVClient(config)
	client.connected = true
	server.Close()

	ctx := context.Background()
	_, err := client.ReadFile(ctx, "file.txt")
	assert.Error(t, err)
}

// =============================================================================
// Local Client — additional edge cases (Boost2)
// =============================================================================

func TestLocalClient_ReadFile_NonexistentB2(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	_, err = client.ReadFile(ctx, "nonexistent.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open local file")
}

func TestLocalClient_GetFileInfo_MissingB2(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	_, err = client.GetFileInfo(ctx, "nonexistent.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to stat local file")
}

func TestLocalClient_ListDirectory_MissingDirB2(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	_, err = client.ListDirectory(ctx, "nonexistent_dir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list local directory")
}

func TestLocalClient_DeleteFile_MissingB2(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	err = client.DeleteFile(ctx, "nonexistent.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete local file")
}

func TestLocalClient_WriteFile_LargeContentB2(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	largeData := bytes.Repeat([]byte("abcdefghij"), 10000)
	err = client.WriteFile(ctx, "large.bin", bytes.NewReader(largeData))
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "large.bin"))
	require.NoError(t, err)
	assert.Equal(t, len(largeData), len(content))
}

func TestLocalClient_Disconnect_IdempotentB2(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)
	assert.True(t, client.IsConnected())

	err = client.Disconnect(ctx)
	require.NoError(t, err)
	assert.False(t, client.IsConnected())

	err = client.Disconnect(ctx)
	require.NoError(t, err)
	assert.False(t, client.IsConnected())
}

// =============================================================================
// NFS Client — edge cases (Boost2)
// =============================================================================

func TestNFSClient_EmptyMountPointB2(t *testing.T) {
	config := NFSConfig{
		Host: "localhost",
		Path: "/export",
	}
	_, err := NewNFSClient(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mount point is required")
}

func TestNFSClient_ReadFile_NotConnectedB2(t *testing.T) {
	config := NFSConfig{Host: "localhost", Path: "/export", MountPoint: "/mnt/nfs"}
	client, err := NewNFSClient(config)
	require.NoError(t, err)
	_, err = client.ReadFile(context.Background(), "file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_WriteFile_NotConnectedB2(t *testing.T) {
	config := NFSConfig{Host: "localhost", Path: "/export", MountPoint: "/mnt/nfs"}
	client, err := NewNFSClient(config)
	require.NoError(t, err)
	err = client.WriteFile(context.Background(), "file.txt", strings.NewReader("data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_GetFileInfo_NotConnectedB2(t *testing.T) {
	config := NFSConfig{Host: "localhost", Path: "/export", MountPoint: "/mnt/nfs"}
	client, err := NewNFSClient(config)
	require.NoError(t, err)
	_, err = client.GetFileInfo(context.Background(), "file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_ListDirectory_NotConnectedB2(t *testing.T) {
	config := NFSConfig{Host: "localhost", Path: "/export", MountPoint: "/mnt/nfs"}
	client, err := NewNFSClient(config)
	require.NoError(t, err)
	_, err = client.ListDirectory(context.Background(), ".")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_FileExists_NotConnectedB2(t *testing.T) {
	config := NFSConfig{Host: "localhost", Path: "/export", MountPoint: "/mnt/nfs"}
	client, err := NewNFSClient(config)
	require.NoError(t, err)
	_, err = client.FileExists(context.Background(), "file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_CreateDirectory_NotConnectedB2(t *testing.T) {
	config := NFSConfig{Host: "localhost", Path: "/export", MountPoint: "/mnt/nfs"}
	client, err := NewNFSClient(config)
	require.NoError(t, err)
	err = client.CreateDirectory(context.Background(), "dir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_DeleteDirectory_NotConnectedB2(t *testing.T) {
	config := NFSConfig{Host: "localhost", Path: "/export", MountPoint: "/mnt/nfs"}
	client, err := NewNFSClient(config)
	require.NoError(t, err)
	err = client.DeleteDirectory(context.Background(), "dir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_DeleteFile_NotConnectedB2(t *testing.T) {
	config := NFSConfig{Host: "localhost", Path: "/export", MountPoint: "/mnt/nfs"}
	client, err := NewNFSClient(config)
	require.NoError(t, err)
	err = client.DeleteFile(context.Background(), "file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_CopyFile_NotConnectedB2(t *testing.T) {
	config := NFSConfig{Host: "localhost", Path: "/export", MountPoint: "/mnt/nfs"}
	client, err := NewNFSClient(config)
	require.NoError(t, err)
	err = client.CopyFile(context.Background(), "src.txt", "dst.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_TestConnection_NotConnectedB2(t *testing.T) {
	config := NFSConfig{Host: "localhost", Path: "/export", MountPoint: "/mnt/nfs"}
	client, err := NewNFSClient(config)
	require.NoError(t, err)
	err = client.TestConnection(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestNFSClient_GetProtocolB2(t *testing.T) {
	config := NFSConfig{Host: "localhost", Path: "/export", MountPoint: "/mnt/nfs"}
	client, err := NewNFSClient(config)
	require.NoError(t, err)
	assert.Equal(t, "nfs", client.GetProtocol())
}

func TestNFSClient_GetConfigB2(t *testing.T) {
	config := NFSConfig{
		Host:       "nfs.example.com",
		Path:       "/exports/media",
		MountPoint: "/mnt/media",
		Options:    "vers=4,rw",
	}
	client, err := NewNFSClient(config)
	require.NoError(t, err)

	retrieved := client.GetConfig().(*NFSConfig)
	assert.Equal(t, "nfs.example.com", retrieved.Host)
	assert.Equal(t, "/exports/media", retrieved.Path)
	assert.Equal(t, "/mnt/media", retrieved.MountPoint)
	assert.Equal(t, "vers=4,rw", retrieved.Options)
}

func TestNFSClient_ResolvePath_TraversalPreventionB2(t *testing.T) {
	client, _ := newConnectedNFSClient(t)
	result := client.resolvePath("../../../etc/passwd")
	assert.NotContains(t, result, "..")
}

func TestNFSClient_Disconnect_NotMountedB2(t *testing.T) {
	config := NFSConfig{Host: "localhost", Path: "/export", MountPoint: "/mnt/nfs"}
	client, err := NewNFSClient(config)
	require.NoError(t, err)

	err = client.Disconnect(context.Background())
	assert.NoError(t, err)
	assert.False(t, client.IsConnected())
}

func TestNFSClient_IsConnected_InitialB2(t *testing.T) {
	config := NFSConfig{Host: "localhost", Path: "/export", MountPoint: "/mnt/nfs"}
	client, err := NewNFSClient(config)
	require.NoError(t, err)
	assert.False(t, client.IsConnected())
}

func TestNFSClient_WriteFile_ErrorOnCreateB2(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	err := os.WriteFile(filepath.Join(tmpDir, "notadir"), []byte("x"), 0644)
	require.NoError(t, err)

	err = client.WriteFile(ctx, "notadir/sub/file.txt", strings.NewReader("data"))
	assert.Error(t, err)
}

// =============================================================================
// FTP Client — additional edge cases (Boost2)
// =============================================================================

func TestFTPClient_IsConnected_WithClientNilB2(t *testing.T) {
	config := &FTPConfig{Host: "localhost", Port: 21}
	client := NewFTPClient(config)

	assert.False(t, client.IsConnected())

	client.connected = true
	assert.False(t, client.IsConnected())
}

func TestFTPClient_Disconnect_WhenNilClientB2(t *testing.T) {
	config := &FTPConfig{Host: "localhost", Port: 21}
	client := NewFTPClient(config)
	client.connected = true

	err := client.Disconnect(context.Background())
	assert.NoError(t, err)
	assert.False(t, client.connected)
}

// =============================================================================
// SMB Client — Disconnect edge cases (Boost2)
// =============================================================================

func TestSmbClient_Disconnect_AllFieldsNilB2(t *testing.T) {
	config := &SmbConfig{Host: "localhost", Port: 445, Share: "test"}
	client := NewSmbClient(config)

	err := client.Disconnect(context.Background())
	assert.NoError(t, err)
}

func TestSmbClient_IsConnected_PartialStateB2(t *testing.T) {
	config := &SmbConfig{Host: "localhost", Port: 445, Share: "test"}
	client := NewSmbClient(config)
	assert.False(t, client.IsConnected())
}

// =============================================================================
// isNotExistError — additional cases (Boost2)
// =============================================================================

func TestIsNotExistError_NilErrorB2(t *testing.T) {
	assert.False(t, isNotExistError(nil))
}

func TestIsNotExistError_VariousB2(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"file does not exist", fmt.Errorf("file does not exist"), true},
		{"no such file or directory", fmt.Errorf("no such file or directory"), true},
		{"permission denied", fmt.Errorf("permission denied"), false},
		{"connection reset", fmt.Errorf("connection reset by peer"), false},
		{"timeout", fmt.Errorf("i/o timeout"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isNotExistError(tt.err))
		})
	}
}

// =============================================================================
// Factory — NFS missing mount_point and int settings (Boost2)
// =============================================================================

func TestDefaultClientFactory_NFS_MissingMountPointB2(t *testing.T) {
	factory := NewDefaultClientFactory()
	config := &StorageConfig{
		Protocol: "nfs",
		Settings: map[string]interface{}{
			"host": "nfs.example.com",
			"path": "/export",
		},
	}

	client, err := factory.CreateClient(config)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to create NFS client")
}

func TestGetIntSetting_IntTypeB2(t *testing.T) {
	settings := map[string]interface{}{
		"port": 8080,
	}
	assert.Equal(t, 8080, getIntSetting(settings, "port", 0))
}

func TestGetIntSetting_NilSettingsB2(t *testing.T) {
	settings := map[string]interface{}{}
	assert.Equal(t, 99, getIntSetting(settings, "missing", 99))
}

func TestGetStringSetting_NilValueB2(t *testing.T) {
	settings := map[string]interface{}{
		"key": nil,
	}
	assert.Equal(t, "default", getStringSetting(settings, "key", "default"))
}

// =============================================================================
// WebDAV Client — network error paths for all operations (Boost2)
// These test the client.Do() error path by closing the server beforehand.
// =============================================================================

func newClosedWebDAVClient(t *testing.T) *WebDAVClient {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMultiStatus)
	}))
	config := &WebDAVConfig{URL: server.URL, Username: "u", Password: "p"}
	client := NewWebDAVClient(config)
	client.connected = true
	server.Close() // Close immediately to trigger network errors
	return client
}

func TestWebDAVClient_WriteFile_NetworkErrorB2(t *testing.T) {
	client := newClosedWebDAVClient(t)
	err := client.WriteFile(context.Background(), "file.txt", strings.NewReader("data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upload WebDAV file")
}

func TestWebDAVClient_GetFileInfo_NetworkErrorB2(t *testing.T) {
	client := newClosedWebDAVClient(t)
	_, err := client.GetFileInfo(context.Background(), "file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get WebDAV file info")
}

func TestWebDAVClient_FileExists_NetworkErrorB2(t *testing.T) {
	client := newClosedWebDAVClient(t)
	_, err := client.FileExists(context.Background(), "file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check WebDAV file existence")
}

func TestWebDAVClient_CreateDirectory_NetworkErrorB2(t *testing.T) {
	client := newClosedWebDAVClient(t)
	err := client.CreateDirectory(context.Background(), "dir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create WebDAV directory")
}

func TestWebDAVClient_DeleteDirectory_NetworkErrorB2(t *testing.T) {
	client := newClosedWebDAVClient(t)
	err := client.DeleteDirectory(context.Background(), "dir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete WebDAV directory")
}

func TestWebDAVClient_DeleteFile_NetworkErrorB2(t *testing.T) {
	client := newClosedWebDAVClient(t)
	err := client.DeleteFile(context.Background(), "file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete WebDAV file")
}

func TestWebDAVClient_CopyFile_NetworkErrorB2(t *testing.T) {
	client := newClosedWebDAVClient(t)
	err := client.CopyFile(context.Background(), "src.txt", "dst.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to copy WebDAV file")
}

func TestWebDAVClient_ListDirectory_NetworkErrorB2(t *testing.T) {
	client := newClosedWebDAVClient(t)
	_, err := client.ListDirectory(context.Background(), "/")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list WebDAV directory")
}

func TestWebDAVClient_Connect_NetworkErrorB2(t *testing.T) {
	client := newClosedWebDAVClient(t)
	client.connected = false
	err := client.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to WebDAV server")
}

// =============================================================================
// Local Client — WriteFile error when directory creation fails (Boost2)
// =============================================================================

func TestLocalClient_WriteFile_ErrorCreatingFileB2(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	// Create a file where we want a directory to be
	err = os.WriteFile(filepath.Join(tmpDir, "blocker"), []byte("x"), 0644)
	require.NoError(t, err)

	// Try to write a file under that path (blocker is a file, not a dir)
	err = client.WriteFile(ctx, "blocker/sub/file.txt", strings.NewReader("data"))
	assert.Error(t, err)
}

func TestLocalClient_CopyFile_DestCreateErrorB2(t *testing.T) {
	tmpDir := t.TempDir()
	config := &LocalConfig{BasePath: tmpDir}
	client := NewLocalClient(config)
	ctx := context.Background()

	err := client.Connect(ctx)
	require.NoError(t, err)

	// Create source file
	err = os.WriteFile(filepath.Join(tmpDir, "src.txt"), []byte("data"), 0644)
	require.NoError(t, err)

	// Create a file where destination directory should be
	err = os.WriteFile(filepath.Join(tmpDir, "blocker"), []byte("x"), 0644)
	require.NoError(t, err)

	// Try to copy to a path under the blocker file
	err = client.CopyFile(ctx, "src.txt", "blocker/sub/dst.txt")
	assert.Error(t, err)
}

// =============================================================================
// NFS Client — WriteFile and CopyFile error paths (Boost2)
// =============================================================================

func TestNFSClient_CopyFile_DestCreateErrorB2(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	// Create source
	err := os.WriteFile(filepath.Join(tmpDir, "src.txt"), []byte("data"), 0644)
	require.NoError(t, err)

	// Create a file where dest dir should be
	err = os.WriteFile(filepath.Join(tmpDir, "blocker"), []byte("x"), 0644)
	require.NoError(t, err)

	err = client.CopyFile(ctx, "src.txt", "blocker/sub/dst.txt")
	assert.Error(t, err)
}

func TestNFSClient_WriteFile_DirCreateErrorB2(t *testing.T) {
	client, tmpDir := newConnectedNFSClient(t)
	ctx := context.Background()

	// Create a file where a directory should be
	err := os.WriteFile(filepath.Join(tmpDir, "blocker"), []byte("x"), 0644)
	require.NoError(t, err)

	err = client.WriteFile(ctx, "blocker/sub/file.txt", strings.NewReader("data"))
	assert.Error(t, err)
}
