package providers

import (
	"catalogizer/internal/media/models"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"go.uber.org/zap"
)

// LLMProvider queries LLMs via OpenAI-compatible APIs as a fallback when
// traditional metadata providers (TMDB, IMDB, etc.) are unreachable.
type LLMProvider struct {
	name       string
	client     *http.Client
	logger     *zap.Logger
	apiKey     string
	baseURL    string
	model      string
	enabled    bool
}

// NewLLMProvider creates a new LLM metadata provider.
// It auto-detects the first available LLM API from the environment.
func NewLLMProvider(client *http.Client, logger *zap.Logger) *LLMProvider {
	// Try providers in order of preference
	candidates := []struct {
		apiKeyEnv string
		baseURL   string
		model     string
		name      string
	}{
		{"DEEPSEEK_API_KEY", "https://api.deepseek.com/v1", "deepseek-chat", "deepseek"},
		{"OPENROUTER_API_KEY", "https://openrouter.ai/api/v1", "openrouter/auto", "openrouter"},
		{"GROQ_API_KEY", "https://api.groq.com/openai/v1", "llama-3.3-70b-versatile", "groq"},
		{"KIMI_API_KEY", "https://api.moonshot.cn/v1", "kimi-k2.5", "kimi"},
		{"GEMINI_API_KEY", "https://generativelanguage.googleapis.com/v1beta/openai", "gemini-2.0-flash", "gemini"},
	}

	for _, c := range candidates {
		apiKey := os.Getenv(c.apiKeyEnv)
		if apiKey != "" {
			if logger != nil {
				logger.Info("LLM metadata provider enabled",
					zap.String("provider", c.name),
					zap.String("model", c.model))
			}
			return &LLMProvider{
				name:    "llm-" + c.name,
				client:  client,
				logger:  logger,
				apiKey:  apiKey,
				baseURL: c.baseURL,
				model:   c.model,
				enabled: true,
			}
		}
	}

	if logger != nil {
		logger.Warn("No LLM API key found; LLM metadata provider disabled")
	}
	return &LLMProvider{
		name:    "llm-fallback",
		client:  client,
		logger:  logger,
		enabled: false,
	}
}

func (l *LLMProvider) GetName() string {
	return l.name
}

func (l *LLMProvider) IsEnabled() bool {
	return l.enabled
}

func (l *LLMProvider) Search(ctx context.Context, query string, mediaType string, year *int) ([]SearchResult, error) {
	if !l.enabled {
		return nil, fmt.Errorf("LLM provider disabled")
	}

	metadata, err := l.queryLLM(ctx, query, mediaType, year)
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

func (l *LLMProvider) queryLLM(ctx context.Context, query string, mediaType string, year *int) (*llmMetadataResponse, error) {
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
		"model":       l.model,
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

	req, err := http.NewRequestWithContext(ctx, "POST", l.baseURL+"/chat/completions", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+l.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM API returned HTTP %d", resp.StatusCode)
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
		l.logger.Warn("Failed to parse LLM metadata JSON",
			zap.String("content", content),
			zap.Error(err))
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
