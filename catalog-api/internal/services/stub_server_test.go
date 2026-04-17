package services

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

// newStubServer returns an httptest.Server that impersonates the
// IGDB v4 endpoints used by our B2 regression test. Every incoming
// request has its Client-ID + Authorization headers captured via
// the supplied callback so tests can assert IGDB auth wiring.
func newStubServer(capture func(clientID, auth, body string)) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		capture(r.Header.Get("Client-ID"), r.Header.Get("Authorization"), string(body))
		_, _ = w.Write([]byte(`[{"id": 999, "name": "Portal"}]`))
	})
	mux.HandleFunc("/covers", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		capture(r.Header.Get("Client-ID"), r.Header.Get("Authorization"), string(body))
		_, _ = w.Write([]byte(`[{"id": 1, "url": "//x.test/t_thumb/abc.jpg", "image_id": "abc"}]`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/igdb/image/upload/") {
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte("JPEGBYTES"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	return httptest.NewServer(mux)
}
