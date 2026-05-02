// Catalogizer → HelixCode QA Integration Example
//
// This example demonstrates how Catalogizer (or any external Go application)
// can trigger QA sessions on a HelixCode server and consume the results.
//
// Usage:
//
//	export HELIXCODE_URL=http://localhost:8080
//	export HELIXCODE_TOKEN=<jwt-token>
//	go run ./examples/helixcode-qa-integration
//
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client is a minimal REST client for HelixCode's QA API.
type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewClient creates a client for the HelixCode QA API.
func NewClient(baseURL, token string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return &Client{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// StartSessionRequest matches HelixCode's StartSessionRequest.
type StartSessionRequest struct {
	Platforms        []string `json:"platforms"`
	Banks            []string `json:"banks"`
	Autonomous       bool     `json:"autonomous"`
	CoverageTarget   float64  `json:"coverage_target"`
	CuriosityEnabled bool     `json:"curiosity_enabled"`
}

// SessionState mirrors HelixCode's helixqa.SessionState.
type SessionState struct {
	ID            string   `json:"id"`
	Status        string   `json:"status"`
	Phase         string   `json:"phase"`
	PhaseProgress float64  `json:"phase_progress"`
	Platforms     []string `json:"platforms"`
	Banks         []string `json:"banks"`
	StartTime     time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	ReportPath    string   `json:"report_path,omitempty"`
}

func (c *Client) do(method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + "/api/v1" + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return c.client.Do(req)
}

// StartSession starts a new QA session on HelixCode.
func (c *Client) StartSession(req StartSessionRequest) (*SessionState, error) {
	data, _ := json.Marshal(req)
	resp, err := c.do("POST", "/qa/session", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("start session failed: %d %s", resp.StatusCode, body)
	}
	var state SessionState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, err
	}
	return &state, nil
}

// GetSession retrieves the current state of a QA session.
func (c *Client) GetSession(id string) (*SessionState, error) {
	resp, err := c.do("GET", "/qa/session/"+id+"/status", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get session failed: %d", resp.StatusCode)
	}
	var state SessionState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, err
	}
	return &state, nil
}

// GetReport fetches the report for a completed session.
func (c *Client) GetReport(id, format string) ([]byte, error) {
	resp, err := c.do("GET", "/qa/session/"+id+"/report?format="+format, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get report failed: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// CancelSession cancels a running QA session.
func (c *Client) CancelSession(id string) error {
	resp, err := c.do("DELETE", "/qa/session/"+id, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cancel session failed: %d", resp.StatusCode)
	}
	return nil
}

func main() {
	baseURL := os.Getenv("HELIXCODE_URL")
	token := os.Getenv("HELIXCODE_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "HELIXCODE_TOKEN is required")
		os.Exit(1)
	}

	client := NewClient(baseURL, token)

	fmt.Println("=== Catalogizer → HelixCode QA Integration Demo ===")

	// 1. Start a QA session
	fmt.Println("\n[1] Starting QA session...")
	state, err := client.StartSession(StartSessionRequest{
		Platforms:        []string{"web"},
		Banks:            []string{"./banks/api"},
		Autonomous:       false,
		CoverageTarget:   0.85,
		CuriosityEnabled: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start session: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("    Session ID: %s\n", state.ID)
	fmt.Printf("    Status: %s | Phase: %s | Progress: %.0f%%\n",
		state.Status, state.Phase, state.PhaseProgress*100)

	// 2. Poll for completion (brief demo — in production use SSE or longer polling)
	fmt.Println("\n[2] Polling session status...")
	for i := 0; i < 3; i++ {
		s, err := client.GetSession(state.ID)
		if err != nil {
			fmt.Printf("    Poll error: %v\n", err)
			break
		}
		fmt.Printf("    Poll %d: %s | %s | %.0f%%\n",
			i+1, s.Status, s.Phase, s.PhaseProgress*100)
		if s.Status == "completed" || s.Status == "failed" || s.Status == "cancelled" {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// 3. Cancel the session (demo cleanup)
	fmt.Println("\n[3] Cancelling session...")
	if err := client.CancelSession(state.ID); err != nil {
		fmt.Printf("    Cancel result: %v\n", err)
	} else {
		fmt.Println("    Session cancelled successfully")
	}

	fmt.Println("\n=== Demo Complete ===")
}
