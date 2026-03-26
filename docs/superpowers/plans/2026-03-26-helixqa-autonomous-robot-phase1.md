# HelixQA Autonomous Robot — Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the two foundational packages (`pkg/llm/` and `pkg/memory/`) that every subsequent phase depends on, then wire the autonomous CLI subcommand.

**Architecture:** Adaptive LLM provider with pluggable backends (Anthropic, OpenAI, Google, Ollama, UI-TARS) behind a unified `Provider` interface. SQLite-backed photographic memory store for session history, findings, coverage, and project knowledge. CLI wiring connects `SessionCoordinator` to real dependencies.

**Tech Stack:** Go 1.25, SQLite (via existing `digital.vasic.database` patterns), `net/http` for LLM API clients, `encoding/json` for response parsing, `testify` for testing.

**Spec:** `docs/superpowers/specs/2026-03-26-helixqa-autonomous-robot-design.md`

---

## File Structure

### New Files

| File | Responsibility |
|------|---------------|
| `pkg/llm/provider.go` | `Provider` interface + `Message`/`Response` types |
| `pkg/llm/anthropic.go` | Claude API implementation |
| `pkg/llm/openai.go` | GPT-4o API implementation |
| `pkg/llm/ollama.go` | Local Ollama HTTP API implementation |
| `pkg/llm/adaptive.go` | Multi-provider wrapper with auto-selection + fallback |
| `pkg/llm/prompt.go` | Versioned prompt templates for each analysis type |
| `pkg/llm/provider_test.go` | Unit tests for Provider interface compliance |
| `pkg/llm/anthropic_test.go` | Anthropic client tests with HTTP mocks |
| `pkg/llm/openai_test.go` | OpenAI client tests with HTTP mocks |
| `pkg/llm/ollama_test.go` | Ollama client tests with HTTP mocks |
| `pkg/llm/adaptive_test.go` | Adaptive selection + fallback tests |
| `pkg/memory/store.go` | SQLite wrapper with schema migrations |
| `pkg/memory/sessions.go` | Session CRUD operations |
| `pkg/memory/findings.go` | Finding CRUD + `docs/issues/` markdown sync |
| `pkg/memory/coverage.go` | Coverage tracking with persistence |
| `pkg/memory/knowledge.go` | Key-value project knowledge store |
| `pkg/memory/store_test.go` | Store creation, migration, close tests |
| `pkg/memory/sessions_test.go` | Session CRUD tests |
| `pkg/memory/findings_test.go` | Finding CRUD + markdown generation tests |
| `pkg/memory/coverage_test.go` | Coverage tracking tests |
| `pkg/memory/knowledge_test.go` | Knowledge store tests |

### Modified Files

| File | Change |
|------|--------|
| `cmd/helixqa/main.go:400-416` | Wire autonomous subcommand with real dependencies |
| `pkg/config/config.go` | Add LLM provider config fields to `AutonomousConfig` |

---

## Task 1: LLM Provider Interface & Types

**Files:**
- Create: `pkg/llm/provider.go`
- Test: `pkg/llm/provider_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/llm/provider_test.go
package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		msg     Message
		wantErr bool
	}{
		{"valid user message", Message{Role: RoleUser, Content: "hello"}, false},
		{"valid system message", Message{Role: RoleSystem, Content: "you are a tester"}, false},
		{"empty role", Message{Role: "", Content: "hello"}, true},
		{"empty content", Message{Role: RoleUser, Content: ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResponse_HasContent(t *testing.T) {
	assert.True(t, Response{Content: "hello"}.HasContent())
	assert.False(t, Response{Content: ""}.HasContent())
	assert.False(t, Response{Content: "  "}.HasContent())
}

func TestProviderConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ProviderConfig
		wantErr bool
	}{
		{"valid anthropic", ProviderConfig{Name: ProviderAnthropic, APIKey: "sk-test"}, false},
		{"valid ollama", ProviderConfig{Name: ProviderOllama, BaseURL: "http://localhost:11434"}, false},
		{"missing name", ProviderConfig{Name: ""}, true},
		{"anthropic missing key", ProviderConfig{Name: ProviderAnthropic}, true},
		{"ollama missing url", ProviderConfig{Name: ProviderOllama}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/llm/ -v -run TestMessage`
Expected: FAIL — package does not exist

- [ ] **Step 3: Write the implementation**

```go
// pkg/llm/provider.go
package llm

import (
	"context"
	"fmt"
	"strings"
)

// Provider names
const (
	ProviderAnthropic = "anthropic"
	ProviderOpenAI    = "openai"
	ProviderGoogle    = "google"
	ProviderOllama    = "ollama"
	ProviderUITars    = "ui-tars"
)

// Message roles
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Provider is the unified interface for all LLM backends.
type Provider interface {
	// Chat sends a text conversation and returns a text response.
	Chat(ctx context.Context, messages []Message) (*Response, error)

	// Vision analyzes an image with a text prompt.
	Vision(ctx context.Context, imageData []byte, prompt string) (*Response, error)

	// Name returns the provider identifier.
	Name() string

	// SupportsVision reports whether this provider can analyze images.
	SupportsVision() bool
}

// Message represents a single message in a conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Validate checks that the message has required fields.
func (m Message) Validate() error {
	if m.Role == "" {
		return fmt.Errorf("llm: message role is required")
	}
	if m.Content == "" {
		return fmt.Errorf("llm: message content is required")
	}
	return nil
}

// Response represents an LLM response.
type Response struct {
	Content      string `json:"content"`
	Model        string `json:"model"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

// HasContent checks if the response has non-whitespace content.
func (r Response) HasContent() bool {
	return strings.TrimSpace(r.Content) != ""
}

// ProviderConfig holds configuration for a single LLM provider.
type ProviderConfig struct {
	Name    string `json:"name"`
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
}

// Validate checks provider config for required fields.
func (c ProviderConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("llm: provider name is required")
	}
	switch c.Name {
	case ProviderAnthropic, ProviderOpenAI, ProviderGoogle:
		if c.APIKey == "" {
			return fmt.Errorf("llm: %s requires an API key", c.Name)
		}
	case ProviderOllama, ProviderUITars:
		if c.BaseURL == "" {
			return fmt.Errorf("llm: %s requires a base URL", c.Name)
		}
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/llm/ -v`
Expected: PASS (all 3 test functions)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/llm/provider.go HelixQA/pkg/llm/provider_test.go
git commit -m "feat(helixqa): add LLM Provider interface and types"
```

---

## Task 2: Anthropic Provider Implementation

**Files:**
- Create: `pkg/llm/anthropic.go`
- Test: `pkg/llm/anthropic_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/llm/anthropic_test.go
package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnthropicProvider_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/messages", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

		resp := anthropicResponse{
			Content: []anthropicContent{{Type: "text", Text: "Hello from Claude"}},
			Model:   "claude-sonnet-4-20250514",
			Usage:   anthropicUsage{InputTokens: 10, OutputTokens: 5},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropicProvider(ProviderConfig{
		Name:    ProviderAnthropic,
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "claude-sonnet-4-20250514",
	})

	resp, err := p.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hello"},
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello from Claude", resp.Content)
	assert.Equal(t, 10, resp.InputTokens)
	assert.Equal(t, 5, resp.OutputTokens)
}

func TestAnthropicProvider_Name(t *testing.T) {
	p := NewAnthropicProvider(ProviderConfig{Name: ProviderAnthropic, APIKey: "k"})
	assert.Equal(t, ProviderAnthropic, p.Name())
}

func TestAnthropicProvider_SupportsVision(t *testing.T) {
	p := NewAnthropicProvider(ProviderConfig{Name: ProviderAnthropic, APIKey: "k"})
	assert.True(t, p.SupportsVision())
}

func TestAnthropicProvider_Chat_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer server.Close()

	p := NewAnthropicProvider(ProviderConfig{
		Name: ProviderAnthropic, APIKey: "k", BaseURL: server.URL,
	})
	_, err := p.Chat(context.Background(), []Message{{Role: RoleUser, Content: "hi"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/llm/ -v -run TestAnthropic`
Expected: FAIL — `NewAnthropicProvider` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/llm/anthropic.go
package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultAnthropicURL = "https://api.anthropic.com"
const defaultAnthropicModel = "claude-sonnet-4-20250514"

type anthropicProvider struct {
	config ProviderConfig
	client *http.Client
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMsg     `json:"messages"`
	System    string             `json:"system,omitempty"`
}

type anthropicMsg struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type anthropicContent struct {
	Type   string `json:"type"`
	Text   string `json:"text,omitempty"`
	Source *anthropicSource `json:"source,omitempty"`
}

type anthropicSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
	Model   string             `json:"model"`
	Usage   anthropicUsage     `json:"usage"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// NewAnthropicProvider creates a Claude API provider.
func NewAnthropicProvider(cfg ProviderConfig) Provider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultAnthropicURL
	}
	if cfg.Model == "" {
		cfg.Model = defaultAnthropicModel
	}
	return &anthropicProvider{
		config: cfg,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (p *anthropicProvider) Name() string         { return ProviderAnthropic }
func (p *anthropicProvider) SupportsVision() bool  { return true }

func (p *anthropicProvider) Chat(ctx context.Context, messages []Message) (*Response, error) {
	var system string
	var msgs []anthropicMsg
	for _, m := range messages {
		if m.Role == RoleSystem {
			system = m.Content
			continue
		}
		msgs = append(msgs, anthropicMsg{Role: m.Role, Content: m.Content})
	}

	reqBody := anthropicRequest{
		Model:     p.config.Model,
		MaxTokens: 4096,
		Messages:  msgs,
		System:    system,
	}

	return p.doRequest(ctx, reqBody)
}

func (p *anthropicProvider) Vision(ctx context.Context, imageData []byte, prompt string) (*Response, error) {
	content := []anthropicContent{
		{
			Type: "image",
			Source: &anthropicSource{
				Type:      "base64",
				MediaType: "image/png",
				Data:      base64.StdEncoding.EncodeToString(imageData),
			},
		},
		{Type: "text", Text: prompt},
	}

	reqBody := anthropicRequest{
		Model:     p.config.Model,
		MaxTokens: 4096,
		Messages:  []anthropicMsg{{Role: RoleUser, Content: content}},
	}

	return p.doRequest(ctx, reqBody)
}

func (p *anthropicProvider) doRequest(ctx context.Context, reqBody anthropicRequest) (*Response, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("anthropic: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("anthropic: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("anthropic: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("anthropic: parse response: %w", err)
	}

	var text string
	for _, c := range apiResp.Content {
		if c.Type == "text" {
			text += c.Text
		}
	}

	return &Response{
		Content:      text,
		Model:        apiResp.Model,
		InputTokens:  apiResp.Usage.InputTokens,
		OutputTokens: apiResp.Usage.OutputTokens,
	}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/llm/ -v -run TestAnthropic`
Expected: PASS (all 4 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/llm/anthropic.go HelixQA/pkg/llm/anthropic_test.go
git commit -m "feat(helixqa): add Anthropic Claude LLM provider"
```

---

## Task 3: OpenAI Provider Implementation

**Files:**
- Create: `pkg/llm/openai.go`
- Test: `pkg/llm/openai_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/llm/openai_test.go
package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIProvider_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer test-key")

		resp := openaiResponse{
			Choices: []openaiChoice{{Message: openaiMsg{Content: "Hello from GPT"}}},
			Model:   "gpt-4o",
			Usage:   openaiUsage{PromptTokens: 8, CompletionTokens: 4},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAIProvider(ProviderConfig{
		Name: ProviderOpenAI, APIKey: "test-key", BaseURL: server.URL, Model: "gpt-4o",
	})

	resp, err := p.Chat(context.Background(), []Message{{Role: RoleUser, Content: "hello"}})
	require.NoError(t, err)
	assert.Equal(t, "Hello from GPT", resp.Content)
}

func TestOpenAIProvider_Name(t *testing.T) {
	p := NewOpenAIProvider(ProviderConfig{Name: ProviderOpenAI, APIKey: "k"})
	assert.Equal(t, ProviderOpenAI, p.Name())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/llm/ -v -run TestOpenAI`
Expected: FAIL — `NewOpenAIProvider` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/llm/openai.go
package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultOpenAIURL = "https://api.openai.com"
const defaultOpenAIModel = "gpt-4o"

type openaiProvider struct {
	config ProviderConfig
	client *http.Client
}

type openaiRequest struct {
	Model    string      `json:"model"`
	Messages []openaiMsg `json:"messages"`
}

type openaiMsg struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type openaiContentPart struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	ImageURL *openaiImageURL `json:"image_url,omitempty"`
}

type openaiImageURL struct {
	URL string `json:"url"`
}

type openaiResponse struct {
	Choices []openaiChoice `json:"choices"`
	Model   string         `json:"model"`
	Usage   openaiUsage    `json:"usage"`
}

type openaiChoice struct {
	Message openaiMsg `json:"message"`
}

type openaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

func NewOpenAIProvider(cfg ProviderConfig) Provider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultOpenAIURL
	}
	if cfg.Model == "" {
		cfg.Model = defaultOpenAIModel
	}
	return &openaiProvider{
		config: cfg,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (p *openaiProvider) Name() string        { return ProviderOpenAI }
func (p *openaiProvider) SupportsVision() bool { return true }

func (p *openaiProvider) Chat(ctx context.Context, messages []Message) (*Response, error) {
	var msgs []openaiMsg
	for _, m := range messages {
		msgs = append(msgs, openaiMsg{Role: m.Role, Content: m.Content})
	}
	return p.doRequest(ctx, openaiRequest{Model: p.config.Model, Messages: msgs})
}

func (p *openaiProvider) Vision(ctx context.Context, imageData []byte, prompt string) (*Response, error) {
	content := []openaiContentPart{
		{Type: "text", Text: prompt},
		{Type: "image_url", ImageURL: &openaiImageURL{
			URL: "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageData),
		}},
	}
	msgs := []openaiMsg{{Role: RoleUser, Content: content}}
	return p.doRequest(ctx, openaiRequest{Model: p.config.Model, Messages: msgs})
}

func (p *openaiProvider) doRequest(ctx context.Context, reqBody openaiRequest) (*Response, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("openai: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openai: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openai: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp openaiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("openai: parse response: %w", err)
	}

	var text string
	if len(apiResp.Choices) > 0 {
		if s, ok := apiResp.Choices[0].Message.Content.(string); ok {
			text = s
		}
	}

	return &Response{
		Content:      text,
		Model:        apiResp.Model,
		InputTokens:  apiResp.Usage.PromptTokens,
		OutputTokens: apiResp.Usage.CompletionTokens,
	}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/llm/ -v -run TestOpenAI`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/llm/openai.go HelixQA/pkg/llm/openai_test.go
git commit -m "feat(helixqa): add OpenAI GPT LLM provider"
```

---

## Task 4: Ollama Provider Implementation

**Files:**
- Create: `pkg/llm/ollama.go`
- Test: `pkg/llm/ollama_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/llm/ollama_test.go
package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOllamaProvider_Chat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/chat", r.URL.Path)

		resp := ollamaChatResponse{
			Message: ollamaMsg{Role: "assistant", Content: "Hello from Ollama"},
			Model:   "qwen2.5",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOllamaProvider(ProviderConfig{
		Name: ProviderOllama, BaseURL: server.URL, Model: "qwen2.5",
	})

	resp, err := p.Chat(context.Background(), []Message{{Role: RoleUser, Content: "hello"}})
	require.NoError(t, err)
	assert.Equal(t, "Hello from Ollama", resp.Content)
}

func TestOllamaProvider_Name(t *testing.T) {
	p := NewOllamaProvider(ProviderConfig{Name: ProviderOllama, BaseURL: "http://localhost:11434"})
	assert.Equal(t, ProviderOllama, p.Name())
}

func TestOllamaProvider_SupportsVision(t *testing.T) {
	p := NewOllamaProvider(ProviderConfig{Name: ProviderOllama, BaseURL: "http://localhost:11434"})
	assert.True(t, p.SupportsVision())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/llm/ -v -run TestOllama`
Expected: FAIL — `NewOllamaProvider` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/llm/ollama.go
package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultOllamaModel = "qwen2.5"

type ollamaProvider struct {
	config ProviderConfig
	client *http.Client
}

type ollamaChatRequest struct {
	Model    string      `json:"model"`
	Messages []ollamaMsg `json:"messages"`
	Stream   bool        `json:"stream"`
}

type ollamaMsg struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

type ollamaChatResponse struct {
	Message ollamaMsg `json:"message"`
	Model   string    `json:"model"`
}

func NewOllamaProvider(cfg ProviderConfig) Provider {
	if cfg.Model == "" {
		cfg.Model = defaultOllamaModel
	}
	return &ollamaProvider{
		config: cfg,
		client: &http.Client{Timeout: 300 * time.Second},
	}
}

func (p *ollamaProvider) Name() string        { return ProviderOllama }
func (p *ollamaProvider) SupportsVision() bool { return true }

func (p *ollamaProvider) Chat(ctx context.Context, messages []Message) (*Response, error) {
	var msgs []ollamaMsg
	for _, m := range messages {
		msgs = append(msgs, ollamaMsg{Role: m.Role, Content: m.Content})
	}

	reqBody := ollamaChatRequest{Model: p.config.Model, Messages: msgs, Stream: false}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: send: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama: read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama: error %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp ollamaChatResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("ollama: parse: %w", err)
	}

	return &Response{Content: apiResp.Message.Content, Model: apiResp.Model}, nil
}

func (p *ollamaProvider) Vision(ctx context.Context, imageData []byte, prompt string) (*Response, error) {
	msgs := []ollamaMsg{{
		Role:    RoleUser,
		Content: prompt,
		Images:  []string{base64.StdEncoding.EncodeToString(imageData)},
	}}

	reqBody := ollamaChatRequest{Model: p.config.Model, Messages: msgs, Stream: false}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: send: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama: read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama: error %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp ollamaChatResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("ollama: parse: %w", err)
	}

	return &Response{Content: apiResp.Message.Content, Model: apiResp.Model}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/llm/ -v -run TestOllama`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/llm/ollama.go HelixQA/pkg/llm/ollama_test.go
git commit -m "feat(helixqa): add Ollama self-hosted LLM provider"
```

---

## Task 5: Adaptive Provider (auto-selection + fallback)

**Files:**
- Create: `pkg/llm/adaptive.go`
- Test: `pkg/llm/adaptive_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/llm/adaptive_test.go
package llm

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	name       string
	vision     bool
	chatResp   *Response
	chatErr    error
	visionResp *Response
	visionErr  error
}

func (m *mockProvider) Name() string        { return m.name }
func (m *mockProvider) SupportsVision() bool { return m.vision }
func (m *mockProvider) Chat(ctx context.Context, msgs []Message) (*Response, error) {
	return m.chatResp, m.chatErr
}
func (m *mockProvider) Vision(ctx context.Context, img []byte, prompt string) (*Response, error) {
	return m.visionResp, m.visionErr
}

func TestAdaptiveProvider_SelectsFirst(t *testing.T) {
	p1 := &mockProvider{name: "first", chatResp: &Response{Content: "from first"}}
	p2 := &mockProvider{name: "second", chatResp: &Response{Content: "from second"}}

	adaptive := NewAdaptiveProvider(p1, p2)
	resp, err := adaptive.Chat(context.Background(), []Message{{Role: RoleUser, Content: "hi"}})
	require.NoError(t, err)
	assert.Equal(t, "from first", resp.Content)
}

func TestAdaptiveProvider_FallsBack(t *testing.T) {
	p1 := &mockProvider{name: "broken", chatErr: fmt.Errorf("connection refused")}
	p2 := &mockProvider{name: "backup", chatResp: &Response{Content: "from backup"}}

	adaptive := NewAdaptiveProvider(p1, p2)
	resp, err := adaptive.Chat(context.Background(), []Message{{Role: RoleUser, Content: "hi"}})
	require.NoError(t, err)
	assert.Equal(t, "from backup", resp.Content)
}

func TestAdaptiveProvider_AllFail(t *testing.T) {
	p1 := &mockProvider{name: "a", chatErr: fmt.Errorf("fail a")}
	p2 := &mockProvider{name: "b", chatErr: fmt.Errorf("fail b")}

	adaptive := NewAdaptiveProvider(p1, p2)
	_, err := adaptive.Chat(context.Background(), []Message{{Role: RoleUser, Content: "hi"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "all providers failed")
}

func TestAdaptiveProvider_VisionSelectsCapable(t *testing.T) {
	p1 := &mockProvider{name: "text-only", vision: false}
	p2 := &mockProvider{name: "multimodal", vision: true, visionResp: &Response{Content: "I see an image"}}

	adaptive := NewAdaptiveProvider(p1, p2)
	resp, err := adaptive.Vision(context.Background(), []byte("img"), "describe")
	require.NoError(t, err)
	assert.Equal(t, "I see an image", resp.Content)
}

func TestAdaptiveProvider_NoProviders(t *testing.T) {
	adaptive := NewAdaptiveProvider()
	_, err := adaptive.Chat(context.Background(), []Message{{Role: RoleUser, Content: "hi"}})
	require.Error(t, err)
}

func TestNewAdaptiveFromEnv(t *testing.T) {
	// Test that factory builds providers from env-style config
	configs := []ProviderConfig{
		{Name: ProviderOllama, BaseURL: "http://localhost:11434"},
	}
	adaptive, err := NewAdaptiveFromConfigs(configs)
	require.NoError(t, err)
	assert.Equal(t, "adaptive", adaptive.Name())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/llm/ -v -run TestAdaptive`
Expected: FAIL — `NewAdaptiveProvider` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/llm/adaptive.go
package llm

import (
	"context"
	"fmt"
	"strings"
)

// AdaptiveProvider wraps multiple providers and auto-selects/falls back.
type AdaptiveProvider struct {
	providers []Provider
}

// NewAdaptiveProvider creates a provider that tries each provider in order.
func NewAdaptiveProvider(providers ...Provider) *AdaptiveProvider {
	return &AdaptiveProvider{providers: providers}
}

// NewAdaptiveFromConfigs creates providers from config structs and wraps them.
func NewAdaptiveFromConfigs(configs []ProviderConfig) (*AdaptiveProvider, error) {
	var providers []Provider
	for _, cfg := range configs {
		if err := cfg.Validate(); err != nil {
			continue // skip invalid configs silently
		}
		switch cfg.Name {
		case ProviderAnthropic:
			providers = append(providers, NewAnthropicProvider(cfg))
		case ProviderOpenAI:
			providers = append(providers, NewOpenAIProvider(cfg))
		case ProviderOllama, ProviderUITars:
			providers = append(providers, NewOllamaProvider(cfg))
		}
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("llm: no valid provider configurations found")
	}
	return NewAdaptiveProvider(providers...), nil
}

func (a *AdaptiveProvider) Name() string        { return "adaptive" }
func (a *AdaptiveProvider) SupportsVision() bool {
	for _, p := range a.providers {
		if p.SupportsVision() {
			return true
		}
	}
	return false
}

func (a *AdaptiveProvider) Chat(ctx context.Context, messages []Message) (*Response, error) {
	if len(a.providers) == 0 {
		return nil, fmt.Errorf("llm: no providers configured")
	}

	var errors []string
	for _, p := range a.providers {
		resp, err := p.Chat(ctx, messages)
		if err == nil {
			return resp, nil
		}
		errors = append(errors, fmt.Sprintf("%s: %v", p.Name(), err))
	}
	return nil, fmt.Errorf("llm: all providers failed: %s", strings.Join(errors, "; "))
}

func (a *AdaptiveProvider) Vision(ctx context.Context, imageData []byte, prompt string) (*Response, error) {
	if len(a.providers) == 0 {
		return nil, fmt.Errorf("llm: no providers configured")
	}

	var errors []string
	for _, p := range a.providers {
		if !p.SupportsVision() {
			continue
		}
		resp, err := p.Vision(ctx, imageData, prompt)
		if err == nil {
			return resp, nil
		}
		errors = append(errors, fmt.Sprintf("%s: %v", p.Name(), err))
	}
	if len(errors) == 0 {
		return nil, fmt.Errorf("llm: no vision-capable providers available")
	}
	return nil, fmt.Errorf("llm: all providers failed: %s", strings.Join(errors, "; "))
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/llm/ -v -run TestAdaptive`
Expected: PASS (all 6 tests)

- [ ] **Step 5: Run all LLM package tests**

Run: `cd HelixQA && go test ./pkg/llm/ -v -race`
Expected: All tests PASS, no race conditions

- [ ] **Step 6: Commit**

```bash
git add HelixQA/pkg/llm/adaptive.go HelixQA/pkg/llm/adaptive_test.go
git commit -m "feat(helixqa): add adaptive multi-provider LLM with fallback"
```

---

## Task 6: Memory Store — SQLite Schema & Migrations

**Files:**
- Create: `pkg/memory/store.go`
- Test: `pkg/memory/store_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/memory/store_test.go
package memory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStore_CreatesDatabase(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	assert.FileExists(t, dbPath)
}

func TestNewStore_RunsMigrations(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(filepath.Join(dir, "test.db"))
	require.NoError(t, err)
	defer store.Close()

	// Verify tables exist
	tables := []string{"sessions", "test_results", "findings", "screenshots", "metrics", "knowledge", "coverage"}
	for _, table := range tables {
		var count int
		err := store.db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
		assert.NoError(t, err, "table %s should exist", table)
	}
}

func TestNewStore_IdempotentMigrations(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store1, err := NewStore(dbPath)
	require.NoError(t, err)
	store1.Close()

	// Opening again should not fail (migrations are idempotent)
	store2, err := NewStore(dbPath)
	require.NoError(t, err)
	defer store2.Close()
}

func TestStore_Close(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(filepath.Join(dir, "test.db"))
	require.NoError(t, err)

	err = store.Close()
	assert.NoError(t, err)

	// Double close should not panic
	err = store.Close()
	assert.NoError(t, err)
}

func TestNewStore_InvalidPath(t *testing.T) {
	_, err := NewStore("/nonexistent/deeply/nested/path/db.sqlite")
	// Should create parent dirs or fail gracefully
	if err != nil {
		assert.Contains(t, err.Error(), "memory:")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/memory/ -v -run TestNewStore`
Expected: FAIL — package does not exist

- [ ] **Step 3: Write the implementation**

```go
// pkg/memory/store.go
package memory

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// Store provides persistent photographic memory across QA sessions.
type Store struct {
	db     *sql.DB
	closed bool
	mu     sync.Mutex
}

// NewStore opens or creates a SQLite database at the given path.
func NewStore(dbPath string) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("memory: create directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("memory: open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("memory: ping database: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("memory: run migrations: %w", err)
	}

	return store, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true
	return s.db.Close()
}

// DB returns the underlying sql.DB for direct queries.
func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			started_at DATETIME NOT NULL,
			ended_at DATETIME,
			duration_seconds INTEGER,
			platforms TEXT,
			coverage_pct REAL DEFAULT 0,
			total_tests INTEGER DEFAULT 0,
			passed INTEGER DEFAULT 0,
			failed INTEGER DEFAULT 0,
			findings_count INTEGER DEFAULT 0,
			pass_number INTEGER DEFAULT 1,
			notes TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS test_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL REFERENCES sessions(id),
			test_case_id TEXT NOT NULL,
			platform TEXT NOT NULL,
			status TEXT NOT NULL,
			duration_ms INTEGER,
			evidence_paths TEXT,
			error_message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS findings (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL REFERENCES sessions(id),
			severity TEXT NOT NULL,
			category TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			repro_steps TEXT,
			evidence_paths TEXT,
			platform TEXT,
			screen TEXT,
			status TEXT DEFAULT 'open',
			found_date DATE,
			fixed_date DATE,
			verified_date DATE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS screenshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL REFERENCES sessions(id),
			screen_name TEXT NOT NULL,
			platform TEXT NOT NULL,
			file_path TEXT NOT NULL,
			width INTEGER,
			height INTEGER,
			hash TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL REFERENCES sessions(id),
			platform TEXT NOT NULL,
			metric_type TEXT NOT NULL,
			value REAL NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS knowledge (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			source TEXT,
			last_verified DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS coverage (
			screen_name TEXT NOT NULL,
			platform TEXT NOT NULL,
			last_tested DATETIME,
			times_tested INTEGER DEFAULT 0,
			last_status TEXT,
			PRIMARY KEY (screen_name, platform)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_test_results_session ON test_results(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_findings_session ON findings(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_findings_status ON findings(status)`,
		`CREATE INDEX IF NOT EXISTS idx_screenshots_session ON screenshots(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_session ON metrics(session_id)`,
	}

	for _, m := range migrations {
		if _, err := s.db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/memory/ -v`
Expected: PASS (all 5 tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/memory/store.go HelixQA/pkg/memory/store_test.go
git commit -m "feat(helixqa): add photographic memory SQLite store with schema"
```

---

## Task 7: Memory Sessions CRUD

**Files:**
- Create: `pkg/memory/sessions.go`
- Test: `pkg/memory/sessions_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/memory/sessions_test.go
package memory

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

func TestStore_CreateSession(t *testing.T) {
	s := newTestStore(t)

	sess := Session{
		ID:        "session-001",
		StartedAt: time.Now(),
		Platforms: "web,android",
		PassNumber: 1,
	}
	err := s.CreateSession(sess)
	require.NoError(t, err)
}

func TestStore_GetSession(t *testing.T) {
	s := newTestStore(t)
	now := time.Now().Truncate(time.Second)

	s.CreateSession(Session{ID: "s1", StartedAt: now, Platforms: "web", PassNumber: 1})

	sess, err := s.GetSession("s1")
	require.NoError(t, err)
	assert.Equal(t, "s1", sess.ID)
	assert.Equal(t, "web", sess.Platforms)
	assert.Equal(t, 1, sess.PassNumber)
}

func TestStore_GetSession_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetSession("nonexistent")
	require.Error(t, err)
}

func TestStore_UpdateSession(t *testing.T) {
	s := newTestStore(t)
	now := time.Now().Truncate(time.Second)

	s.CreateSession(Session{ID: "s1", StartedAt: now, Platforms: "web"})

	ended := now.Add(30 * time.Minute)
	err := s.UpdateSession("s1", SessionUpdate{
		EndedAt:       &ended,
		Duration:      1800,
		TotalTests:    50,
		Passed:        48,
		Failed:        2,
		FindingsCount: 2,
		CoveragePct:   0.85,
	})
	require.NoError(t, err)

	sess, _ := s.GetSession("s1")
	assert.Equal(t, 50, sess.TotalTests)
	assert.Equal(t, 48, sess.Passed)
	assert.InDelta(t, 0.85, sess.CoveragePct, 0.001)
}

func TestStore_ListSessions(t *testing.T) {
	s := newTestStore(t)
	now := time.Now()

	s.CreateSession(Session{ID: "s1", StartedAt: now.Add(-2 * time.Hour), Platforms: "web"})
	s.CreateSession(Session{ID: "s2", StartedAt: now.Add(-1 * time.Hour), Platforms: "android"})
	s.CreateSession(Session{ID: "s3", StartedAt: now, Platforms: "all"})

	sessions, err := s.ListSessions(10)
	require.NoError(t, err)
	assert.Len(t, sessions, 3)
	// Most recent first
	assert.Equal(t, "s3", sessions[0].ID)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/memory/ -v -run TestStore_Create`
Expected: FAIL — `CreateSession` method undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/memory/sessions.go
package memory

import (
	"database/sql"
	"fmt"
	"time"
)

// Session represents a QA session record.
type Session struct {
	ID            string    `json:"id"`
	StartedAt     time.Time `json:"started_at"`
	EndedAt       time.Time `json:"ended_at,omitempty"`
	Duration      int       `json:"duration_seconds"`
	Platforms     string    `json:"platforms"`
	CoveragePct   float64   `json:"coverage_pct"`
	TotalTests    int       `json:"total_tests"`
	Passed        int       `json:"passed"`
	Failed        int       `json:"failed"`
	FindingsCount int       `json:"findings_count"`
	PassNumber    int       `json:"pass_number"`
	Notes         string    `json:"notes"`
}

// SessionUpdate holds optional fields for updating a session.
type SessionUpdate struct {
	EndedAt       *time.Time
	Duration      int
	TotalTests    int
	Passed        int
	Failed        int
	FindingsCount int
	CoveragePct   float64
	Notes         string
}

// CreateSession inserts a new session record.
func (s *Store) CreateSession(sess Session) error {
	_, err := s.db.Exec(
		`INSERT INTO sessions (id, started_at, platforms, pass_number, notes)
		 VALUES (?, ?, ?, ?, ?)`,
		sess.ID, sess.StartedAt, sess.Platforms, sess.PassNumber, sess.Notes,
	)
	if err != nil {
		return fmt.Errorf("memory: create session: %w", err)
	}
	return nil
}

// GetSession retrieves a session by ID.
func (s *Store) GetSession(id string) (*Session, error) {
	var sess Session
	var endedAt sql.NullTime
	err := s.db.QueryRow(
		`SELECT id, started_at, ended_at, duration_seconds, platforms,
		        coverage_pct, total_tests, passed, failed, findings_count, pass_number, notes
		 FROM sessions WHERE id = ?`, id,
	).Scan(
		&sess.ID, &sess.StartedAt, &endedAt, &sess.Duration, &sess.Platforms,
		&sess.CoveragePct, &sess.TotalTests, &sess.Passed, &sess.Failed,
		&sess.FindingsCount, &sess.PassNumber, &sess.Notes,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("memory: session %s not found", id)
		}
		return nil, fmt.Errorf("memory: get session: %w", err)
	}
	if endedAt.Valid {
		sess.EndedAt = endedAt.Time
	}
	return &sess, nil
}

// UpdateSession updates a session with completion data.
func (s *Store) UpdateSession(id string, update SessionUpdate) error {
	_, err := s.db.Exec(
		`UPDATE sessions SET
			ended_at = ?, duration_seconds = ?, total_tests = ?,
			passed = ?, failed = ?, findings_count = ?, coverage_pct = ?,
			notes = COALESCE(NULLIF(?, ''), notes)
		 WHERE id = ?`,
		update.EndedAt, update.Duration, update.TotalTests,
		update.Passed, update.Failed, update.FindingsCount, update.CoveragePct,
		update.Notes, id,
	)
	if err != nil {
		return fmt.Errorf("memory: update session: %w", err)
	}
	return nil
}

// ListSessions returns the most recent sessions.
func (s *Store) ListSessions(limit int) ([]Session, error) {
	rows, err := s.db.Query(
		`SELECT id, started_at, ended_at, duration_seconds, platforms,
		        coverage_pct, total_tests, passed, failed, findings_count, pass_number, notes
		 FROM sessions ORDER BY started_at DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("memory: list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var sess Session
		var endedAt sql.NullTime
		if err := rows.Scan(
			&sess.ID, &sess.StartedAt, &endedAt, &sess.Duration, &sess.Platforms,
			&sess.CoveragePct, &sess.TotalTests, &sess.Passed, &sess.Failed,
			&sess.FindingsCount, &sess.PassNumber, &sess.Notes,
		); err != nil {
			return nil, fmt.Errorf("memory: scan session: %w", err)
		}
		if endedAt.Valid {
			sess.EndedAt = endedAt.Time
		}
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

// LatestPassNumber returns the highest pass number across all sessions.
func (s *Store) LatestPassNumber() (int, error) {
	var pass sql.NullInt64
	err := s.db.QueryRow("SELECT MAX(pass_number) FROM sessions").Scan(&pass)
	if err != nil || !pass.Valid {
		return 0, nil
	}
	return int(pass.Int64), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/memory/ -v -run TestStore_`
Expected: PASS (all 5 session tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/memory/sessions.go HelixQA/pkg/memory/sessions_test.go
git commit -m "feat(helixqa): add session CRUD to memory store"
```

---

## Task 8: Memory Findings CRUD + Markdown Issue Generation

**Files:**
- Create: `pkg/memory/findings.go`
- Test: `pkg/memory/findings_test.go`

- [ ] **Step 1: Write the failing test**

```go
// pkg/memory/findings_test.go
package memory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_CreateFinding(t *testing.T) {
	s := newTestStore(t)

	f := Finding{
		ID:          "HELIX-001",
		SessionID:   "s1",
		Severity:    "high",
		Category:    "visual",
		Title:       "Cover art clipped on rotation",
		Description: "Cover art overflows container by 20px in landscape",
		ReproSteps:  "1. Go to detail\n2. Rotate\n3. Observe",
		Platform:    "android",
		Screen:      "media-detail",
		Status:      "open",
	}
	err := s.CreateFinding(f)
	require.NoError(t, err)
}

func TestStore_GetFinding(t *testing.T) {
	s := newTestStore(t)
	s.CreateFinding(Finding{ID: "HELIX-001", SessionID: "s1", Severity: "high", Category: "visual", Title: "Test bug", Status: "open"})

	f, err := s.GetFinding("HELIX-001")
	require.NoError(t, err)
	assert.Equal(t, "Test bug", f.Title)
	assert.Equal(t, "open", f.Status)
}

func TestStore_UpdateFindingStatus(t *testing.T) {
	s := newTestStore(t)
	s.CreateFinding(Finding{ID: "HELIX-001", SessionID: "s1", Severity: "high", Category: "visual", Title: "Bug", Status: "open"})

	err := s.UpdateFindingStatus("HELIX-001", "fixed")
	require.NoError(t, err)

	f, _ := s.GetFinding("HELIX-001")
	assert.Equal(t, "fixed", f.Status)
}

func TestStore_ListOpenFindings(t *testing.T) {
	s := newTestStore(t)
	s.CreateFinding(Finding{ID: "H-1", SessionID: "s1", Severity: "high", Category: "visual", Title: "Open bug", Status: "open"})
	s.CreateFinding(Finding{ID: "H-2", SessionID: "s1", Severity: "low", Category: "ux", Title: "Fixed bug", Status: "fixed"})
	s.CreateFinding(Finding{ID: "H-3", SessionID: "s1", Severity: "critical", Category: "crash", Title: "Another open", Status: "open"})

	findings, err := s.ListFindingsByStatus("open")
	require.NoError(t, err)
	assert.Len(t, findings, 2)
}

func TestStore_NextFindingID(t *testing.T) {
	s := newTestStore(t)
	id, err := s.NextFindingID()
	require.NoError(t, err)
	assert.Equal(t, "HELIX-001", id)

	s.CreateFinding(Finding{ID: "HELIX-001", SessionID: "s1", Severity: "low", Category: "ux", Title: "Bug", Status: "open"})
	id2, _ := s.NextFindingID()
	assert.Equal(t, "HELIX-002", id2)
}

func TestFinding_ToMarkdown(t *testing.T) {
	f := Finding{
		ID:          "HELIX-042",
		Severity:    "high",
		Category:    "visual",
		Platform:    "android",
		Screen:      "media-detail",
		Title:       "Cover art clipped on rotation",
		Description: "Cover art overflows container by 20px in landscape.",
		ReproSteps:  "1. Navigate to detail\n2. Rotate\n3. Observe",
		Status:      "open",
		SessionID:   "session-20260326",
		FoundDate:   "2026-03-26",
	}

	md := f.ToMarkdown()
	assert.Contains(t, md, "id: HELIX-042")
	assert.Contains(t, md, "severity: high")
	assert.Contains(t, md, "# HELIX-042: Cover art clipped on rotation")
	assert.Contains(t, md, "## Steps to Reproduce")
	assert.Contains(t, md, "1. Navigate to detail")
}

func TestFinding_WriteToDir(t *testing.T) {
	dir := t.TempDir()
	f := Finding{
		ID:       "HELIX-001",
		Severity: "high",
		Category: "visual",
		Title:    "Test issue",
		Status:   "open",
	}

	path, err := f.WriteToDir(dir)
	require.NoError(t, err)
	assert.FileExists(t, path)

	content, _ := os.ReadFile(path)
	assert.Contains(t, string(content), "HELIX-001")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd HelixQA && go test ./pkg/memory/ -v -run TestStore_CreateFinding`
Expected: FAIL — `CreateFinding` undefined

- [ ] **Step 3: Write the implementation**

```go
// pkg/memory/findings.go
package memory

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Finding represents a discovered issue.
type Finding struct {
	ID            string `json:"id"`
	SessionID     string `json:"session_id"`
	Severity      string `json:"severity"`
	Category      string `json:"category"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	ReproSteps    string `json:"repro_steps"`
	EvidencePaths string `json:"evidence_paths"`
	Platform      string `json:"platform"`
	Screen        string `json:"screen"`
	Status        string `json:"status"`
	FoundDate     string `json:"found_date"`
	FixedDate     string `json:"fixed_date"`
	VerifiedDate  string `json:"verified_date"`
}

// CreateFinding inserts a new finding.
func (s *Store) CreateFinding(f Finding) error {
	_, err := s.db.Exec(
		`INSERT INTO findings (id, session_id, severity, category, title, description,
		 repro_steps, evidence_paths, platform, screen, status, found_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		f.ID, f.SessionID, f.Severity, f.Category, f.Title, f.Description,
		f.ReproSteps, f.EvidencePaths, f.Platform, f.Screen, f.Status, f.FoundDate,
	)
	if err != nil {
		return fmt.Errorf("memory: create finding: %w", err)
	}
	return nil
}

// GetFinding retrieves a finding by ID.
func (s *Store) GetFinding(id string) (*Finding, error) {
	var f Finding
	var fixedDate, verifiedDate sql.NullString
	err := s.db.QueryRow(
		`SELECT id, session_id, severity, category, title, description,
		        repro_steps, evidence_paths, platform, screen, status,
		        COALESCE(found_date, ''), fixed_date, verified_date
		 FROM findings WHERE id = ?`, id,
	).Scan(
		&f.ID, &f.SessionID, &f.Severity, &f.Category, &f.Title, &f.Description,
		&f.ReproSteps, &f.EvidencePaths, &f.Platform, &f.Screen, &f.Status,
		&f.FoundDate, &fixedDate, &verifiedDate,
	)
	if err != nil {
		return nil, fmt.Errorf("memory: get finding %s: %w", id, err)
	}
	if fixedDate.Valid {
		f.FixedDate = fixedDate.String
	}
	if verifiedDate.Valid {
		f.VerifiedDate = verifiedDate.String
	}
	return &f, nil
}

// UpdateFindingStatus changes a finding's status.
func (s *Store) UpdateFindingStatus(id, status string) error {
	_, err := s.db.Exec(
		"UPDATE findings SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("memory: update finding status: %w", err)
	}
	return nil
}

// ListFindingsByStatus returns all findings with the given status.
func (s *Store) ListFindingsByStatus(status string) ([]Finding, error) {
	rows, err := s.db.Query(
		`SELECT id, session_id, severity, category, title, description,
		        platform, screen, status, COALESCE(found_date, '')
		 FROM findings WHERE status = ? ORDER BY created_at DESC`, status,
	)
	if err != nil {
		return nil, fmt.Errorf("memory: list findings: %w", err)
	}
	defer rows.Close()

	var findings []Finding
	for rows.Next() {
		var f Finding
		if err := rows.Scan(
			&f.ID, &f.SessionID, &f.Severity, &f.Category, &f.Title,
			&f.Description, &f.Platform, &f.Screen, &f.Status, &f.FoundDate,
		); err != nil {
			return nil, fmt.Errorf("memory: scan finding: %w", err)
		}
		findings = append(findings, f)
	}
	return findings, nil
}

// NextFindingID returns the next available HELIX-NNN ID.
func (s *Store) NextFindingID() (string, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM findings").Scan(&count)
	if err != nil {
		return "", fmt.Errorf("memory: count findings: %w", err)
	}
	return fmt.Sprintf("HELIX-%03d", count+1), nil
}

// ToMarkdown renders the finding as a markdown issue file.
func (f Finding) ToMarkdown() string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("id: %s\n", f.ID))
	b.WriteString(fmt.Sprintf("status: %s\n", f.Status))
	b.WriteString(fmt.Sprintf("severity: %s\n", f.Severity))
	b.WriteString(fmt.Sprintf("category: %s\n", f.Category))
	if f.Platform != "" {
		b.WriteString(fmt.Sprintf("platform: %s\n", f.Platform))
	}
	if f.Screen != "" {
		b.WriteString(fmt.Sprintf("screen: %s\n", f.Screen))
	}
	b.WriteString(fmt.Sprintf("found_session: %s\n", f.SessionID))
	if f.FoundDate != "" {
		b.WriteString(fmt.Sprintf("found_date: %s\n", f.FoundDate))
	}
	b.WriteString("---\n\n")
	b.WriteString(fmt.Sprintf("# %s: %s\n\n", f.ID, f.Title))
	if f.Description != "" {
		b.WriteString("## Description\n\n")
		b.WriteString(f.Description + "\n\n")
	}
	if f.ReproSteps != "" {
		b.WriteString("## Steps to Reproduce\n\n")
		b.WriteString(f.ReproSteps + "\n\n")
	}
	if f.EvidencePaths != "" {
		b.WriteString("## Evidence\n\n")
		b.WriteString(f.EvidencePaths + "\n\n")
	}
	return b.String()
}

// WriteToDir writes the finding as a markdown file in the given directory.
func (f Finding) WriteToDir(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("memory: create issues dir: %w", err)
	}

	slug := strings.ToLower(strings.ReplaceAll(f.Title, " ", "-"))
	if len(slug) > 50 {
		slug = slug[:50]
	}
	filename := fmt.Sprintf("%s-%s.md", f.ID, slug)
	path := filepath.Join(dir, filename)

	if err := os.WriteFile(path, []byte(f.ToMarkdown()), 0644); err != nil {
		return "", fmt.Errorf("memory: write finding: %w", err)
	}
	return path, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd HelixQA && go test ./pkg/memory/ -v`
Expected: PASS (all store + finding tests)

- [ ] **Step 5: Commit**

```bash
git add HelixQA/pkg/memory/findings.go HelixQA/pkg/memory/findings_test.go
git commit -m "feat(helixqa): add findings CRUD + markdown issue generation"
```

---

## Task 9: Memory Coverage & Knowledge Stores

**Files:**
- Create: `pkg/memory/coverage.go`, `pkg/memory/knowledge.go`
- Test: `pkg/memory/coverage_test.go`, `pkg/memory/knowledge_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// pkg/memory/coverage_test.go
package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_RecordCoverage(t *testing.T) {
	s := newTestStore(t)

	err := s.RecordCoverage("login-screen", "android", "passed")
	require.NoError(t, err)

	// Second time increments count
	err = s.RecordCoverage("login-screen", "android", "passed")
	require.NoError(t, err)

	cov, err := s.GetCoverage("login-screen", "android")
	require.NoError(t, err)
	assert.Equal(t, 2, cov.TimesTested)
	assert.Equal(t, "passed", cov.LastStatus)
}

func TestStore_ListUncoveredScreens(t *testing.T) {
	s := newTestStore(t)

	s.RecordCoverage("login", "web", "passed")
	s.RecordCoverage("dashboard", "web", "passed")

	// All known screens
	allScreens := []string{"login", "dashboard", "media", "browse", "settings"}
	uncovered := s.ListUncoveredScreens(allScreens, "web")
	assert.Len(t, uncovered, 3)
	assert.Contains(t, uncovered, "media")
}
```

```go
// pkg/memory/knowledge_test.go
package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_SetKnowledge(t *testing.T) {
	s := newTestStore(t)

	err := s.SetKnowledge("total_screens", "15", "codebase_scan")
	require.NoError(t, err)

	val, err := s.GetKnowledge("total_screens")
	require.NoError(t, err)
	assert.Equal(t, "15", val)
}

func TestStore_SetKnowledge_Upsert(t *testing.T) {
	s := newTestStore(t)

	s.SetKnowledge("api_count", "100", "scan_v1")
	s.SetKnowledge("api_count", "132", "scan_v2")

	val, _ := s.GetKnowledge("api_count")
	assert.Equal(t, "132", val)
}

func TestStore_GetKnowledge_NotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetKnowledge("nonexistent")
	require.Error(t, err)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd HelixQA && go test ./pkg/memory/ -v -run "TestStore_Record|TestStore_List|TestStore_Set|TestStore_Get"`
Expected: FAIL

- [ ] **Step 3: Write coverage.go**

```go
// pkg/memory/coverage.go
package memory

import (
	"fmt"
	"time"
)

// CoverageEntry represents coverage data for a screen/platform pair.
type CoverageEntry struct {
	ScreenName  string    `json:"screen_name"`
	Platform    string    `json:"platform"`
	LastTested  time.Time `json:"last_tested"`
	TimesTested int       `json:"times_tested"`
	LastStatus  string    `json:"last_status"`
}

// RecordCoverage upserts a coverage entry.
func (s *Store) RecordCoverage(screen, platform, status string) error {
	_, err := s.db.Exec(
		`INSERT INTO coverage (screen_name, platform, last_tested, times_tested, last_status)
		 VALUES (?, ?, CURRENT_TIMESTAMP, 1, ?)
		 ON CONFLICT(screen_name, platform) DO UPDATE SET
		   last_tested = CURRENT_TIMESTAMP,
		   times_tested = times_tested + 1,
		   last_status = ?`,
		screen, platform, status, status,
	)
	if err != nil {
		return fmt.Errorf("memory: record coverage: %w", err)
	}
	return nil
}

// GetCoverage retrieves coverage for a specific screen/platform.
func (s *Store) GetCoverage(screen, platform string) (*CoverageEntry, error) {
	var c CoverageEntry
	err := s.db.QueryRow(
		`SELECT screen_name, platform, last_tested, times_tested, last_status
		 FROM coverage WHERE screen_name = ? AND platform = ?`,
		screen, platform,
	).Scan(&c.ScreenName, &c.Platform, &c.LastTested, &c.TimesTested, &c.LastStatus)
	if err != nil {
		return nil, fmt.Errorf("memory: get coverage: %w", err)
	}
	return &c, nil
}

// ListUncoveredScreens returns screens from allScreens not in coverage for the platform.
func (s *Store) ListUncoveredScreens(allScreens []string, platform string) []string {
	covered := make(map[string]bool)
	rows, err := s.db.Query(
		"SELECT screen_name FROM coverage WHERE platform = ?", platform,
	)
	if err != nil {
		return allScreens
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		rows.Scan(&name)
		covered[name] = true
	}

	var uncovered []string
	for _, screen := range allScreens {
		if !covered[screen] {
			uncovered = append(uncovered, screen)
		}
	}
	return uncovered
}
```

- [ ] **Step 4: Write knowledge.go**

```go
// pkg/memory/knowledge.go
package memory

import (
	"database/sql"
	"fmt"
)

// SetKnowledge upserts a key-value knowledge entry.
func (s *Store) SetKnowledge(key, value, source string) error {
	_, err := s.db.Exec(
		`INSERT INTO knowledge (key, value, source, last_verified)
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(key) DO UPDATE SET
		   value = ?, source = ?, last_verified = CURRENT_TIMESTAMP`,
		key, value, source, value, source,
	)
	if err != nil {
		return fmt.Errorf("memory: set knowledge: %w", err)
	}
	return nil
}

// GetKnowledge retrieves a value by key.
func (s *Store) GetKnowledge(key string) (string, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM knowledge WHERE key = ?", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("memory: knowledge key %q not found", key)
		}
		return "", fmt.Errorf("memory: get knowledge: %w", err)
	}
	return value, nil
}

// AllKnowledge returns all knowledge entries as a map.
func (s *Store) AllKnowledge() (map[string]string, error) {
	rows, err := s.db.Query("SELECT key, value FROM knowledge")
	if err != nil {
		return nil, fmt.Errorf("memory: list knowledge: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		result[k] = v
	}
	return result, nil
}
```

- [ ] **Step 5: Run all memory tests**

Run: `cd HelixQA && go test ./pkg/memory/ -v -race`
Expected: PASS (all tests, no races)

- [ ] **Step 6: Commit**

```bash
git add HelixQA/pkg/memory/coverage.go HelixQA/pkg/memory/coverage_test.go \
       HelixQA/pkg/memory/knowledge.go HelixQA/pkg/memory/knowledge_test.go
git commit -m "feat(helixqa): add coverage tracking and knowledge store to memory"
```

---

## Task 10: Wire Autonomous CLI Subcommand

**Files:**
- Modify: `cmd/helixqa/main.go:400-416`
- Modify: `pkg/config/config.go` (add LLM fields)

- [ ] **Step 1: Add LLM config fields to AutonomousConfig**

In `pkg/config/config.go`, add after the existing `AutonomousConfig` fields:

```go
// Add to AutonomousConfig struct:
LLMProvider       string   // anthropic, openai, ollama, adaptive
LLMAPIKey         string
LLMBaseURL        string
LLMModel          string
VisionAPIKey      string
VisionBaseURL     string
VisionModel       string
MemoryDBPath      string   // path to memory.db
IssuesDir         string   // path to docs/issues/
```

- [ ] **Step 2: Replace the TODO in main.go**

Replace lines 400-416 in `cmd/helixqa/main.go`:

```go
// Build LLM provider from environment
var llmConfigs []llm.ProviderConfig
if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
    llmConfigs = append(llmConfigs, llm.ProviderConfig{
        Name: llm.ProviderAnthropic, APIKey: apiKey,
        Model: os.Getenv("HELIX_LLM_MODEL"),
    })
}
if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
    llmConfigs = append(llmConfigs, llm.ProviderConfig{
        Name: llm.ProviderOpenAI, APIKey: apiKey,
    })
}
if ollamaURL := os.Getenv("HELIX_OLLAMA_URL"); ollamaURL != "" {
    llmConfigs = append(llmConfigs, llm.ProviderConfig{
        Name: llm.ProviderOllama, BaseURL: ollamaURL,
        Model: os.Getenv("HELIX_OLLAMA_MODEL"),
    })
}

if len(llmConfigs) == 0 {
    fmt.Println("ERROR: No LLM provider configured.")
    fmt.Println("Set ANTHROPIC_API_KEY, OPENAI_API_KEY, or HELIX_OLLAMA_URL")
    os.Exit(1)
}

provider, err := llm.NewAdaptiveFromConfigs(llmConfigs)
if err != nil {
    fmt.Printf("ERROR: Failed to create LLM provider: %v\n", err)
    os.Exit(1)
}

// Open memory store
memDBPath := os.Getenv("HELIX_MEMORY_DB")
if memDBPath == "" {
    memDBPath = filepath.Join(projectRoot, "HelixQA", "data", "memory.db")
}
memStore, err := memory.NewStore(memDBPath)
if err != nil {
    fmt.Printf("ERROR: Failed to open memory store: %v\n", err)
    os.Exit(1)
}
defer memStore.Close()

// Determine pass number
passNum, _ := memStore.LatestPassNumber()
passNum++

fmt.Printf("HelixQA Autonomous Robot — Pass #%d\n", passNum)
fmt.Printf("LLM Provider: %s\n", provider.Name())
fmt.Printf("Platforms: %v\n", platformStrs)
fmt.Printf("Memory DB: %s\n", memDBPath)
fmt.Println()
fmt.Println("Phase 1 foundation ready. Phases 2-4 (learning, planning, execution, analysis) coming in next implementation phase.")
```

- [ ] **Step 3: Add imports to main.go**

Add to import block:
```go
"digital.vasic.helixqa/pkg/llm"
"digital.vasic.helixqa/pkg/memory"
```

- [ ] **Step 4: Run the autonomous subcommand to verify wiring**

Run: `cd HelixQA && ANTHROPIC_API_KEY=test go run ./cmd/helixqa autonomous --project /tmp --platforms web`
Expected: Prints "HelixQA Autonomous Robot — Pass #1" and provider info (no crash)

- [ ] **Step 5: Run all HelixQA tests to verify nothing broke**

Run: `cd HelixQA && go test ./... -race -count=1`
Expected: All existing 235+ tests still pass

- [ ] **Step 6: Commit**

```bash
git add HelixQA/cmd/helixqa/main.go HelixQA/pkg/config/config.go
git commit -m "feat(helixqa): wire autonomous CLI with adaptive LLM provider and memory store"
```

---

## Task 11: Final Integration Test

- [ ] **Step 1: Run the complete test suite**

Run: `cd HelixQA && go test ./... -v -race -count=1`
Expected: All tests pass including new pkg/llm/ and pkg/memory/ tests

- [ ] **Step 2: Verify test count increased**

Run: `cd HelixQA && go test ./... -v 2>&1 | grep -c "--- PASS"`
Expected: 235 + ~30 new tests = ~265+ passing tests

- [ ] **Step 3: Run go vet**

Run: `cd HelixQA && go vet ./...`
Expected: No issues

- [ ] **Step 4: Commit and push**

```bash
git add -A HelixQA/
git commit -m "test(helixqa): verify Phase 1 integration — LLM providers + memory store"
git push origin main
```
