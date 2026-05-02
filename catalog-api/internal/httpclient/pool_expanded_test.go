package httpclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// NewPooledClient — transport details
// ---------------------------------------------------------------------------

func TestNewPooledClient_TransportResponseHeaderTimeout(t *testing.T) {
	client := NewPooledClient()
	transport := client.Transport.(*http.Transport)
	assert.Equal(t, 30*time.Second, transport.ResponseHeaderTimeout)
}

func TestNewPooledClient_TransportExpectContinueTimeout(t *testing.T) {
	client := NewPooledClient()
	transport := client.Transport.(*http.Transport)
	assert.Equal(t, 1*time.Second, transport.ExpectContinueTimeout)
}

// ---------------------------------------------------------------------------
// NewPooledClientWithTimeout — various durations
// ---------------------------------------------------------------------------

func TestNewPooledClientWithTimeout_VariousDurations(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{"1 second", 1 * time.Second},
		{"30 seconds", 30 * time.Second},
		{"5 minutes", 5 * time.Minute},
		{"zero timeout", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewPooledClientWithTimeout(tt.timeout)
			assert.Equal(t, tt.timeout, client.Timeout)

			// Transport should still have pool settings
			transport, ok := client.Transport.(*http.Transport)
			require.True(t, ok)
			assert.Equal(t, 100, transport.MaxIdleConns)
			assert.Equal(t, 10, transport.MaxIdleConnsPerHost)
		})
	}
}

// ---------------------------------------------------------------------------
// Concurrent request handling
// ---------------------------------------------------------------------------

func TestPooledClient_ConcurrentRequests(t *testing.T) {
	var requestCount int32
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := NewPooledClient()

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			resp, err := client.Get(server.URL)
			if err != nil {
				errors <- err
				return
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("unexpected error: %v", err)
	}

	mu.Lock()
	assert.Equal(t, int32(goroutines), requestCount)
	mu.Unlock()
}

// ---------------------------------------------------------------------------
// POST/PUT/DELETE methods
// ---------------------------------------------------------------------------

func TestPooledClient_PostRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		w.Write(body)
	}))
	defer server.Close()

	client := NewPooledClient()
	resp, err := client.Post(server.URL, "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Context cancellation
// ---------------------------------------------------------------------------

func TestPooledClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewPooledClient()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	_, err = client.Do(req)
	assert.Error(t, err, "request should fail due to context timeout")
}

// ---------------------------------------------------------------------------
// Response body reading
// ---------------------------------------------------------------------------

func TestPooledClient_LargeResponse(t *testing.T) {
	// 1MB response
	largeBody := make([]byte, 1024*1024)
	for i := range largeBody {
		largeBody[i] = 'A'
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(largeBody)
	}))
	defer server.Close()

	client := NewPooledClient()
	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, len(largeBody), len(body))
}

// ---------------------------------------------------------------------------
// HTTP status codes
// ---------------------------------------------------------------------------

func TestPooledClient_HTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"200 OK", http.StatusOK},
		{"201 Created", http.StatusCreated},
		{"204 No Content", http.StatusNoContent},
		{"400 Bad Request", http.StatusBadRequest},
		{"401 Unauthorized", http.StatusUnauthorized},
		{"403 Forbidden", http.StatusForbidden},
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := NewPooledClient()
			resp, err := client.Get(server.URL)
			require.NoError(t, err)
			resp.Body.Close()
			assert.Equal(t, tt.statusCode, resp.StatusCode)
		})
	}
}
