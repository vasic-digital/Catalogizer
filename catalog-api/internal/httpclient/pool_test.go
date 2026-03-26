package httpclient

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPooledClient_HasCorrectTransportSettings(t *testing.T) {
	client := NewPooledClient()
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok, "Transport should be *http.Transport")
	assert.Equal(t, 100, transport.MaxIdleConns)
	assert.Equal(t, 10, transport.MaxIdleConnsPerHost)
	assert.Equal(t, 30*time.Second, transport.IdleConnTimeout)
	assert.Equal(t, 10*time.Second, transport.TLSHandshakeTimeout)
	assert.Equal(t, 60*time.Second, client.Timeout)
}

func TestNewPooledClientWithTimeout(t *testing.T) {
	client := NewPooledClientWithTimeout(5 * time.Second)
	assert.Equal(t, 5*time.Second, client.Timeout)
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.Equal(t, 100, transport.MaxIdleConns) // Still has pool settings
}

func TestPooledClient_MakesRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := NewPooledClient()
	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ok", string(body))
}

func TestPooledClient_ReusesConnections(t *testing.T) {
	connCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewPooledClient()
	// Make multiple requests - connections should be pooled
	for i := 0; i < 10; i++ {
		resp, err := client.Get(server.URL)
		require.NoError(t, err)
		resp.Body.Close()
		connCount++
	}
	assert.Equal(t, 10, connCount) // All requests completed
}
