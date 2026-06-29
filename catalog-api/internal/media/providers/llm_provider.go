package providers

import (
	"catalogizer/internal/media/models"
	"catalogizer/internal/services"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"go.uber.org/zap"
)

// llmCandidate is one configured OpenAI-compatible LLM backend.
type llmCandidate struct {
	name    string
	apiKey  string
	baseURL string
	model   string
}

// LLMProvider queries LLMs via OpenAI-compatible APIs as a fallback when
// traditional metadata providers (TMDB, IMDB, etc.) are unreachable.
//
// It holds every configured backend (one per non-empty API key in the
// environment) in priority order and fails over between them at request
// time on payment/quota/server/transport errors, so a single out-of-credit
// or rate-limited backend (e.g. DeepSeek HTTP 402) no longer disables
// enrichment for the whole process while another funded backend
// (e.g. Groq HTTP 200) works.
type LLMProvider struct {
	client     *http.Client
	logger     *zap.Logger
	candidates []llmCandidate
	// active is the index of the candidate to try first. It advances to
	// whichever candidate last served a successful request so a dead
	// backend is not re-probed first on every call.
	active  atomic.Int32
	enabled bool
}

// llmEnvCandidates lists the supported LLM backends in priority order.
var llmEnvCandidates = []llmCandidate{
	{name: "deepseek", baseURL: "https://api.deepseek.com/v1", model: "deepseek-chat"},
	{name: "openrouter", baseURL: "https://openrouter.ai/api/v1", model: "openrouter/auto"},
	{name: "groq", baseURL: "https://api.groq.com/openai/v1", model: "llama-3.3-70b-versatile"},
	{name: "kimi", baseURL: "https://api.moonshot.cn/v1", model: "kimi-k2.5"},
	{name: "gemini", baseURL: "https://generativelanguage.googleapis.com/v1beta/openai", model: "gemini-2.0-flash"},
}

// llmEnvKeys maps a candidate name to its API-key environment variable.
var llmEnvKeys = map[string]string{
	"deepseek":   "DEEPSEEK_API_KEY",
	"openrouter": "OPENROUTER_API_KEY",
	"groq":       "GROQ_API_KEY",
	"kimi":       "KIMI_API_KEY",
	"gemini":     "GEMINI_API_KEY",
}

// buildLLMCandidates returns every backend with a non-empty API key in the
// environment, preserving the priority order of llmEnvCandidates.
func buildLLMCandidates() []llmCandidate {
	configured := make([]llmCandidate, 0, len(llmEnvCandidates))
	for _, c := range llmEnvCandidates {
		apiKey := os.Getenv(llmEnvKeys[c.name])
		if apiKey == "" {
			continue
		}
		cand := c
		cand.apiKey = apiKey
		configured = append(configured, cand)
	}
	return configured
}

// NewLLMProvider creates a new LLM metadata provider. It auto-detects every
// configured LLM API from the environment and fails over between them at
// request time.
func NewLLMProvider(client *http.Client, logger *zap.Logger) *LLMProvider {
	return newLLMProviderWithCandidates(client, logger, buildLLMCandidates())
}

// newLLMProviderWithCandidates builds a provider from an explicit candidate
// list. Used by tests that point candidates at in-process httptest servers;
// production constructs the list from the environment via NewLLMProvider.
func newLLMProviderWithCandidates(client *http.Client, logger *zap.Logger, candidates []llmCandidate) *LLMProvider {
	if len(candidates) == 0 {
		if logger != nil {
			logger.Warn("No LLM API key found; LLM metadata provider disabled")
		}
		return &LLMProvider{client: client, logger: logger, enabled: false}
	}
	if logger != nil {
		names := make([]string, len(candidates))
		for i, c := range candidates {
			names[i] = c.name
		}
		logger.Info("LLM metadata provider enabled",
			zap.Strings("providers", names),
			zap.String("primary", candidates[0].name),
			zap.String("primary_model", candidates[0].model))
	}
	return &LLMProvider{
		client:     client,
		logger:     logger,
		candidates: candidates,
		enabled:    true,
	}
}

func (l *LLMProvider) GetName() string {
	if !l.enabled || len(l.candidates) == 0 {
		return "llm-fallback"
	}
	idx := int(l.active.Load())
	if idx < 0 || idx >= len(l.candidates) {
		idx = 0
	}
	return "llm-" + l.candidates[idx].name
}

func (l *LLMProvider) IsEnabled() bool {
	return l.enabled
}

func (l *LLMProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !l.enabled {
		return nil, fmt.Errorf("LLM provider disabled")
	}

	metadata, err := l.queryLLMWithFailover(ctx, query, mediaType, year)
	if err != nil {
		return nil, err
	}
	if metadata == nil {
		return nil, nil
	}

	result := SearchResult{
		ExternalID:  "llm:" + sanitizeExternalID(query),
		Title:       metadata.Title,
		Year:        metadata.Year,
		Description: &metadata.Description,
		Relevance:   0.6, // Lower base relevance than primary providers
	}
	if metadata.CoverURL != "" {
		result.CoverURL = &metadata.CoverURL
	}
	if metadata.Rating > 0 {
		result.Rating = &metadata.Rating
	}

	return []SearchResult{result}, nil
}

func (l *LLMProvider) GetDetails(ctx context.Context, externalID string) (*models.ExternalMetadata, error) {
	if !l.enabled {
		return nil, fmt.Errorf("LLM provider disabled")
	}
	// LLM provider does not support detail lookups by externalID;
	// the search response already contains all available metadata.
	return nil, fmt.Errorf("GetDetails not supported by LLM provider")
}

type llmMetadataResponse struct {
	Title       string  `json:"title"`
	Year        *int    `json:"year,omitempty"`
	Description string  `json:"description"`
	CoverURL    string  `json:"cover_url"`
	Rating      float64 `json:"rating"`
	Error       string  `json:"error,omitempty"`
}

// retryableLLMError marks a backend failure that should trigger failover to
// the next configured candidate: a transport error (status 0) or a
// payment/quota/server HTTP status (402, 429, 5xx). Non-retryable failures
// (a 200 with unparseable content, a blocked endpoint, a 4xx other than
// 402/429) are surfaced directly without burning the other candidates.
type retryableLLMError struct {
	provider string
	status   int // 0 == transport error
	err      error
}

func (e *retryableLLMError) Error() string {
	if e.status > 0 {
		return fmt.Sprintf("LLM provider %s returned HTTP %d", e.provider, e.status)
	}
	return fmt.Sprintf("LLM provider %s request failed: %v", e.provider, e.err)
}

func (e *retryableLLMError) Unwrap() error { return e.err }

// isRetryableStatus reports whether an HTTP status warrants failover to the
// next candidate: payment required (402), rate-limited (429), or any 5xx.
func isRetryableStatus(code int) bool {
	return code == http.StatusPaymentRequired ||
		code == http.StatusTooManyRequests ||
		code >= 500
}

// queryLLMWithFailover tries the configured candidates in priority order,
// starting from the active candidate (the one that last succeeded). A
// payment/quota/server/transport failure (retryableLLMError) advances to
// the next candidate; the first success is returned and remembered as the
// new active candidate so dead backends are not re-probed first on every
// call. If every candidate fails, the last retryable error is returned —
// never a false success. A non-retryable error (e.g. a 200 with unparseable
// content or a blocked endpoint) is surfaced immediately.
func (l *LLMProvider) queryLLMWithFailover(ctx context.Context, query string, mediaType string, year *int) (*llmMetadataResponse, error) {
	n := len(l.candidates)
	if n == 0 {
		return nil, fmt.Errorf("no LLM candidates configured")
	}

	start := int(l.active.Load())
	if start < 0 || start >= n {
		start = 0
	}

	var lastErr error
	for i := 0; i < n; i++ {
		idx := (start + i) % n
		cand := l.candidates[idx]

		metadata, err := l.queryLLM(ctx, cand, query, mediaType, year)
		if err == nil {
			// Success (including an authoritative "not found" with nil
			// metadata) — remember this candidate for the next call.
			l.active.Store(int32(idx))
			return metadata, nil
		}

		var retryable *retryableLLMError
		if errors.As(err, &retryable) {
			lastErr = err
			if l.logger != nil {
				l.logger.Warn("LLM candidate failed; failing over to next",
					zap.String("provider", cand.name),
					zap.Error(err))
			}
			continue
		}

		// Non-retryable: the backend responded but the content/endpoint is
		// bad. Surface it without burning the remaining candidates.
		return nil, err
	}

	return nil, fmt.Errorf("all %d LLM candidates failed: %w", n, lastErr)
}

func (l *LLMProvider) queryLLM(ctx context.Context, cand llmCandidate, query string, mediaType string, year *int) (*llmMetadataResponse, error) {
	yearHint := ""
	if year != nil {
		yearHint = fmt.Sprintf(" (%d)", *year)
	}

	systemPrompt := `You are a metadata lookup assistant for a media catalog.
Respond ONLY with a single JSON object using this exact schema:
{"title":"string","year":integer,"description":"string","cover_url":"string","rating":number}
If you cannot find the item, respond with {"error":"not found"}.
Do not include markdown, explanations, or any text outside the JSON object.`

	userPrompt := fmt.Sprintf("Find metadata and a direct cover image URL for the %s '%s%s'.", mediaType, query, yearHint)

	payload := map[string]interface{}{
		"model": cand.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens":  800,
		"temperature": 0.1,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal LLM request: %w", err)
	}

	target := cand.baseURL + "/chat/completions"
	if err := services.GuardProviderURL(target, services.SSRFGuardConfig{}); err != nil {
		return nil, fmt.Errorf("unsafe LLM endpoint: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", target, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cand.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		// Transport error — let the caller fail over to the next candidate.
		return nil, &retryableLLMError{provider: cand.name, status: 0, err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if isRetryableStatus(resp.StatusCode) {
			return nil, &retryableLLMError{provider: cand.name, status: resp.StatusCode}
		}
		return nil, fmt.Errorf("LLM API %s returned HTTP %d", cand.name, resp.StatusCode)
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode LLM response: %w", err)
	}
	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("empty LLM response")
	}

	content := strings.TrimSpace(apiResp.Choices[0].Message.Content)
	// Strip markdown code fences if present
	content = stripMarkdownCodeFences(content)

	var result llmMetadataResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		if l.logger != nil {
			l.logger.Warn("Failed to parse LLM metadata JSON",
				zap.String("content", content),
				zap.Error(err))
		}
		return nil, fmt.Errorf("failed to parse LLM metadata: %w", err)
	}

	if result.Error != "" {
		return nil, nil // not found
	}

	return &result, nil
}

func stripMarkdownCodeFences(s string) string {
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimSuffix(s, "```")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

func sanitizeExternalID(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "-")
}
