# Module 21: AI-Powered Features -- Video Script

**Duration**: 55 minutes
**Prerequisites**: Module 4 (Media Detection and Processing), Module 19 (Entity System Deep Dive)

---

## Video 21.1: AI Metadata Extraction (12 min)

### Opening

Welcome to Module 21. Artificial intelligence augments every stage of the Catalogizer pipeline -- from metadata extraction and content analysis to recommendations and autonomous quality assurance. This module explores how AI is integrated, which providers are supported, and how the AI dashboard gives users visibility into model-driven decisions.

### Where AI Fits in the Pipeline

**[Visual: Pipeline diagram highlighting AI-augmented stages: Detection -> Analysis -> Enrichment -> Recommendation]**

**Narrator**: AI is not a separate subsystem. It enhances existing stages. During detection, language models disambiguate titles that regex parsers cannot handle. During analysis, content analysis classifies ambiguous directories. During enrichment, AI summarizes metadata from multiple providers into coherent descriptions. And in the recommendation layer, models suggest related content based on user behavior and entity attributes.

### AI-Assisted Title Resolution

**[Visual: Show an ambiguous directory name that the regex parser cannot handle]**

**Narrator**: The title parser handles well-structured names like "The.Matrix.1999.1080p.BluRay" with regex. But what about "matrix trilogy complete box set rip" or "breaking bad all seasons dvdrip"? These lack year, quality, and season markers. The AI metadata extractor handles these edge cases.

```go
// catalog-api/internal/services/ai_metadata_service.go
type AIMetadataService struct {
    provider    AIProvider
    titleParser *TitleParser
    logger      *zap.Logger
    cache       *CacheService
}

type AIExtractionResult struct {
    Title         string   `json:"title"`
    Year          *int     `json:"year"`
    MediaType     string   `json:"media_type"`
    Season        *int     `json:"season,omitempty"`
    Episode       *int     `json:"episode,omitempty"`
    Confidence    float64  `json:"confidence"`
    AlternateNames []string `json:"alternate_names,omitempty"`
}
```

**[Visual: Show the fallback chain: regex parser -> AI extractor -> manual entry]**

**Narrator**: The extraction pipeline tries the regex parser first. If the parser returns a low confidence score -- below a configurable threshold -- the AI extractor is invoked. If AI extraction also fails or is unavailable, the entity is flagged for manual review.

```go
// catalog-api/internal/services/ai_metadata_service.go
func (s *AIMetadataService) ExtractMetadata(
    ctx context.Context,
    dirname string,
    fileList []string,
) (*AIExtractionResult, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("ai_extract:%s", dirname)
    if cached, ok := s.cache.Get(cacheKey); ok {
        return cached.(*AIExtractionResult), nil
    }

    prompt := s.buildExtractionPrompt(dirname, fileList)
    response, err := s.provider.Complete(ctx, prompt)
    if err != nil {
        return nil, fmt.Errorf("AI extraction failed: %w", err)
    }

    result, err := s.parseExtractionResponse(response)
    if err != nil {
        return nil, fmt.Errorf("parse AI response: %w", err)
    }

    s.cache.Set(cacheKey, result, 24*time.Hour)
    return result, nil
}
```

**Narrator**: Results are cached for 24 hours to avoid redundant API calls. The prompt includes the directory name and a sample of contained file names to give the model maximum context.

---

## Video 21.2: Content Analysis Pipeline (12 min)

### Multi-Stage Analysis

**[Visual: Diagram showing content flowing through classification, summarization, and tagging stages]**

**Narrator**: The content analysis pipeline runs after entity creation. It operates in three stages: classification, summarization, and tagging. Each stage can be powered by a different model depending on cost and latency requirements.

### Classification

**[Visual: Show an entity with ambiguous type being classified by AI]**

**Narrator**: Classification resolves ambiguous media types. A directory named "Avatar" could be a movie, a TV show, or a game. The classifier examines file extensions, sizes, directory structure, and file count to make a determination.

```go
// catalog-api/internal/services/content_analysis_service.go
type ContentClassification struct {
    PredictedType   string  `json:"predicted_type"`
    Confidence      float64 `json:"confidence"`
    SecondaryType   string  `json:"secondary_type,omitempty"`
    Reasoning       string  `json:"reasoning"`
}

func (s *ContentAnalysisService) Classify(
    ctx context.Context,
    entityID int64,
    files []models.File,
) (*ContentClassification, error) {
    features := s.extractFeatures(files)
    prompt := s.buildClassificationPrompt(features)
    return s.runClassification(ctx, prompt)
}
```

**Narrator**: The feature extraction step distills file metadata into a compact representation: total size, file count by extension, average file size, directory depth, and presence of specific file types like .nfo, .srt, or .cue.

### Summarization

**[Visual: Show an entity with a long description being summarized into a concise paragraph]**

**Narrator**: When metadata providers return verbose descriptions or when multiple providers contribute conflicting summaries, the AI summarizer produces a single, coherent description. It preserves factual content while reducing length.

### Auto-Tagging

**[Visual: Show an entity receiving auto-generated genre and keyword tags]**

**Narrator**: The tagging stage generates genre tags, mood keywords, and thematic labels from the entity's title, description, and metadata. These tags power smart collections and search. For example, a movie might receive tags like "neo-noir", "dystopian", "philosophical" beyond its provider-supplied genres.

---

## Video 21.3: LLM Provider Integration (12 min)

### 40+ Provider Support via HelixQA

**[Visual: Architecture diagram showing HelixQA's provider abstraction layer]**

**Narrator**: Catalogizer integrates with over 40 large language model providers through the HelixQA framework. Rather than coupling to a single vendor, the system uses a provider abstraction that supports OpenAI, Anthropic, Google, Mistral, Cohere, local models via Ollama, and many more.

**[Visual: Show the AIProvider interface]**

**Narrator**: Every LLM integration implements the same `AIProvider` interface. This allows transparent provider switching based on cost, latency, availability, or user preference.

```go
// HelixQA/pkg/ai/provider.go
type AIProvider interface {
    Complete(ctx context.Context, prompt string) (string, error)
    CompleteWithSchema(ctx context.Context, prompt string, schema interface{}) (string, error)
    StreamComplete(ctx context.Context, prompt string, callback StreamCallback) error
    GetModelName() string
    GetProviderName() string
    IsAvailable() bool
    EstimateCost(promptTokens, completionTokens int) float64
}
```

### Provider Configuration

**[Visual: Show configuration for multiple providers]**

**Narrator**: Providers are configured in the application config with API keys, model names, and resource limits. The system supports primary and fallback providers.

```json
{
  "ai": {
    "primary_provider": "anthropic",
    "fallback_provider": "ollama",
    "providers": {
      "anthropic": {
        "api_key": "${ANTHROPIC_API_KEY}",
        "model": "claude-sonnet-4-20250514",
        "max_tokens": 4096,
        "rate_limit_rpm": 60
      },
      "openai": {
        "api_key": "${OPENAI_API_KEY}",
        "model": "gpt-4o",
        "max_tokens": 4096,
        "rate_limit_rpm": 60
      },
      "ollama": {
        "base_url": "http://localhost:11434",
        "model": "llama3",
        "max_tokens": 4096
      }
    },
    "cache_ttl_hours": 24,
    "max_concurrent_requests": 4
  }
}
```

**[Visual: Show provider fallback in action]**

**Narrator**: If the primary provider returns an error or times out, the system automatically retries with the fallback provider. This ensures AI features remain available even when a cloud provider has an outage. Local models via Ollama serve as a reliable fallback that does not depend on internet connectivity.

### Cost Management

**[Visual: Show cost tracking dashboard]**

**Narrator**: Every AI request is tracked with token counts and estimated cost. The `EstimateCost` method on each provider calculates the cost based on the provider's pricing model. Daily and monthly spending limits prevent runaway costs.

```go
// catalog-api/internal/services/ai_metadata_service.go
type AIUsageRecord struct {
    Provider         string    `json:"provider"`
    Model            string    `json:"model"`
    PromptTokens     int       `json:"prompt_tokens"`
    CompletionTokens int       `json:"completion_tokens"`
    EstimatedCost    float64   `json:"estimated_cost"`
    Operation        string    `json:"operation"`
    Timestamp        time.Time `json:"timestamp"`
}
```

---

## Video 21.4: AI Dashboard Walkthrough (10 min)

### Dashboard Overview

**[Visual: Browser showing the AI dashboard at `/dashboard/ai`]**

**Narrator**: The AI dashboard provides visibility into every AI-driven decision. It shows recent extractions, classification results, provider usage, cost breakdown, and cache hit rates.

**[Visual: Show the extraction log]**

**Narrator**: The extraction log lists every AI metadata extraction with the input directory name, the model's output, confidence score, and whether the result was used or overridden by the user. This audit trail is critical for understanding why an entity was classified a certain way.

**[Visual: Show provider usage chart]**

**Narrator**: The provider usage chart shows request counts by provider over time. It highlights fallback events, rate limit hits, and error rates. This helps operators decide when to switch providers or adjust rate limits.

### Confidence Visualization

**[Visual: Show entities color-coded by AI confidence: green (high), yellow (medium), red (low)]**

**Narrator**: The entity browser can be filtered by AI confidence. High-confidence extractions (above 0.9) are shown in green. Medium confidence (0.7 to 0.9) in yellow. Low confidence (below 0.7) in red. Users can focus on low-confidence entities for manual review.

### Manual Override

**[Visual: Show a user correcting an AI classification]**

**Narrator**: Every AI decision can be overridden. When a user corrects a classification or title, the override is recorded. These overrides serve as training signal -- they improve the prompt engineering and rule definitions over time.

---

## Video 21.5: Recommendation Engine Concepts (9 min)

### Recommendation Architecture

**[Visual: Diagram showing user behavior -> feature extraction -> similarity scoring -> ranked suggestions]**

**Narrator**: The recommendation engine suggests media that users are likely to enjoy based on their collection contents, browsing history, and rating patterns. It uses content-based filtering rather than collaborative filtering, since Catalogizer may serve a single household.

### Feature Vectors

**[Visual: Show an entity's feature vector with genre, year, rating, and tag dimensions]**

**Narrator**: Each entity is represented as a feature vector. Dimensions include genre tags, decade, rating bucket, media type, and auto-generated keywords from the AI tagging stage.

```go
// catalog-api/internal/services/recommendation_service.go
type EntityFeatures struct {
    EntityID  int64              `json:"entity_id"`
    Genres    []string           `json:"genres"`
    Decade    int                `json:"decade"`
    Rating    float64            `json:"rating"`
    Keywords  []string           `json:"keywords"`
    MediaType string             `json:"media_type"`
    Vector    []float64          `json:"vector"`
}

func (s *RecommendationService) GetRecommendations(
    ctx context.Context,
    userID int64,
    limit int,
) ([]RecommendedItem, error) {
    // Build user profile from collections and ratings
    profile, err := s.buildUserProfile(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("build user profile: %w", err)
    }

    // Score all entities against the profile
    candidates, err := s.scoreCandidates(ctx, profile)
    if err != nil {
        return nil, fmt.Errorf("score candidates: %w", err)
    }

    // Return top N by score, excluding already-seen items
    return s.rankAndFilter(candidates, limit), nil
}
```

### Similarity Scoring

**[Visual: Show cosine similarity calculation between two entity vectors]**

**Narrator**: Similarity between entities uses cosine similarity on their feature vectors. The user profile is a weighted average of feature vectors from their highest-rated and most-viewed entities. Recommendations are entities with the highest cosine similarity to the user profile that the user has not already interacted with.

**[Visual: Show recommendation results on the home page]**

**Narrator**: Recommendations appear on the home dashboard as a "Suggested For You" row. Each recommendation includes a relevance score and a brief explanation: "Because you liked The Matrix and Blade Runner" or "Similar to albums in your Jazz collection."

---

## Key Code Examples

### AI Pipeline Flow
```
1. File scanned -> regex title parser attempts extraction
2. Low confidence? -> AI metadata extractor invoked
3. Entity created with extracted metadata
4. Content analysis: classification, summarization, tagging
5. Results cached (24h TTL) to avoid repeat API calls
6. AI dashboard logs every decision with confidence scores
7. User overrides recorded for continuous improvement
```

### AI API Endpoints
```
GET  /api/v1/ai/dashboard          -- AI usage summary and stats
GET  /api/v1/ai/extractions        -- Recent AI extraction log
GET  /api/v1/ai/providers          -- Provider status and usage
POST /api/v1/ai/classify/:id       -- Trigger AI classification for entity
POST /api/v1/ai/summarize/:id      -- Trigger AI summarization for entity
GET  /api/v1/ai/recommendations    -- Get personalized recommendations
POST /api/v1/ai/override/:id       -- Record manual override
GET  /api/v1/ai/cost               -- Cost breakdown by provider/period
```

---

## Key Files Referenced

- `catalog-api/internal/services/ai_metadata_service.go` -- AI extraction and caching
- `catalog-api/internal/services/content_analysis_service.go` -- Classification, summarization, tagging
- `catalog-api/internal/services/recommendation_service.go` -- Content-based recommendations
- `HelixQA/pkg/ai/provider.go` -- AIProvider interface for 40+ LLM providers
- `catalog-api/internal/services/title_parser.go` -- Regex parser (AI fallback trigger)
- `catalog-api/internal/services/cache_service.go` -- AI result caching

---

## Exercises

1. Write a prompt template for the AI metadata extractor that handles multi-season TV show box sets (e.g., "breaking bad complete series bluray remux").
2. Implement a simple content-based recommendation engine using cosine similarity on genre vectors. Write a table-driven test with at least 5 entities.
3. Add a new AI provider adapter for a local Ollama instance running Mistral. Implement the `AIProvider` interface with streaming support.
4. Create a cost alerting function that sends a notification when daily AI spending exceeds a configurable threshold.
5. Write a classification prompt that distinguishes between a "game" and "software" entity based on file extensions and directory structure.

---

## Quiz Questions

1. When does the AI metadata extractor run, and what triggers it?
   **Answer**: It runs when the regex title parser returns a confidence score below the configured threshold. The aggregation service calls it as a fallback during the entity creation pipeline. Results are cached for 24 hours.

2. How does the system handle LLM provider outages?
   **Answer**: The configuration defines a primary and fallback provider. If the primary returns an error or times out, the request is automatically retried with the fallback. Local models via Ollama serve as a reliable fallback independent of internet connectivity.

3. What is the recommendation strategy, and why was content-based filtering chosen?
   **Answer**: The engine uses content-based filtering with cosine similarity on entity feature vectors. This was chosen because Catalogizer may serve a single household, making collaborative filtering (which requires many users) impractical.

4. How are AI decisions audited and corrected?
   **Answer**: Every AI extraction, classification, and tagging decision is logged with the input, output, confidence score, and model used. The AI dashboard displays these logs. Users can override any decision, and overrides are recorded to improve prompt engineering over time.
