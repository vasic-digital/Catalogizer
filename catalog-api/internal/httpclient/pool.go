package httpclient

import (
	"net"
	"net/http"
	"time"
)

// NewPooledClient creates an HTTP client with connection pooling optimized for
// making many concurrent requests to a small number of hosts.
func NewPooledClient() *http.Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:  10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}
}

// NewPooledClientWithTimeout creates a pooled HTTP client with a custom request timeout.
func NewPooledClientWithTimeout(timeout time.Duration) *http.Client {
	client := NewPooledClient()
	client.Timeout = timeout
	return client
}
