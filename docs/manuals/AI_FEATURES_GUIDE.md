# Catalogizer -- AI Features Guide

## Table of Contents

1. [What AI Features Are Available](#what-ai-features-are-available)
2. [AI Metadata Extraction](#ai-metadata-extraction)
3. [Content Analysis](#content-analysis)
4. [AI Dashboard Overview](#ai-dashboard-overview)
5. [Confidence Scores and Manual Overrides](#confidence-scores-and-manual-overrides)
6. [LLM Provider Configuration](#llm-provider-configuration)
7. [Cost Management](#cost-management)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

---

## What AI Features Are Available

Catalogizer integrates AI capabilities to enhance media library management beyond what traditional metadata providers and pattern-matching rules can achieve. AI features are optional -- the core cataloging, scanning, and playback functionality works fully without them. When enabled, AI provides the following capabilities:

- **Metadata extraction** -- Identifies and enriches entity metadata when traditional providers return incomplete results or when file naming conventions are ambiguous.
- **Content analysis** -- Analyzes media content to extract descriptive tags, detect genres, identify themes, and generate summaries.
- **Duplicate resolution** -- Uses semantic similarity to detect duplicates that differ in naming but represent the same content.
- **Title disambiguation** -- Resolves ambiguous titles by considering file context, directory structure, and surrounding files.
- **Smart collection suggestions** -- Recommends collection groupings based on patterns in your library.

AI features communicate with external LLM providers via API. No AI inference runs locally on the Catalogizer server -- all processing is offloaded to the configured provider.

---

## AI Metadata Extraction

When the standard metadata pipeline (title parser, TMDB, OMDB, OpenLibrary, MusicBrainz) cannot fully identify or enrich an entity, AI metadata extraction serves as a fallback enrichment layer.

### How It Works

1. After the standard aggregation pipeline completes, entities with missing or incomplete metadata are flagged as candidates for AI enrichment.
2. For each candidate, Catalogizer constructs a prompt containing the file path, directory structure, parsed title fragments, and any partial metadata already gathered.
3. The prompt is sent to the configured LLM provider.
4. The LLM response is parsed and validated. Extracted fields are written to the entity's metadata with an AI source marker.

### What AI Can Extract

| Field | Example |
|-------|---------|
| **Corrected title** | "Blade Runner 2049" from a file named `BR2049.2017.UHD.mkv` |
| **Year** | Release year when not present in the filename |
| **Genre** | Genre classification from plot context or title recognition |
| **Description** | A brief plot summary or content description |
| **Language** | Primary language of the content |
| **Cast and crew** | Key actors and directors for well-known titles |
| **Series information** | Show name, season, and episode number from non-standard naming |

### When AI Extraction Runs

AI metadata extraction is triggered in the following scenarios:

- **Post-scan enrichment** -- Automatically for entities where the title parser produced low-confidence results and traditional providers returned no matches.
- **Manual trigger** -- From the entity detail page, click **AI Enrich** to request AI-assisted metadata for a specific entity.
- **Batch enrichment** -- From the admin panel, run AI enrichment across all entities with missing metadata.

AI extraction does not overwrite metadata already confirmed by traditional providers. It fills gaps and adds supplementary information.

---

## Content Analysis

Content analysis goes beyond metadata lookup by examining the media content itself (or its associated metadata) to produce descriptive tags, thematic classifications, and summaries.

### Descriptive Tagging

AI generates descriptive tags that capture the themes, mood, and characteristics of media items. Unlike genre tags from metadata providers (which are broad categories like "Action" or "Drama"), AI-generated tags are more specific:

- "time travel", "dystopian future", "unreliable narrator"
- "acoustic", "melancholic", "live recording"
- "pixel art", "open world", "roguelike"

Tags are displayed on the entity detail page and are searchable from the entity browser. They can be used in smart collection rules.

### Genre Refinement

Traditional metadata providers assign genres from a fixed list. AI can refine these by adding sub-genres and cross-genre classifications:

- A movie tagged as "Horror" by TMDB might be further classified as "psychological horror" or "folk horror" by AI.
- An album tagged as "Rock" by MusicBrainz might be refined to "progressive rock" or "post-rock".

### Summary Generation

For entities without a plot summary from providers (common for older, obscure, or non-English content), AI generates a brief summary based on the title, year, genre, and any available contextual information. Summaries are marked with an AI attribution badge so users know the source.

### Similarity Mapping

AI analyzes the tags, genres, and descriptions of all entities in your library to build a similarity index. This powers the "Similar Items" section on the entity detail page, showing entities that are thematically or stylistically related even if they are of different media types.

---

## AI Dashboard Overview

The AI dashboard is accessible from the web application at `/admin/ai` (admin users) or from the main navigation under **AI Insights** (all users, with limited scope).

### Dashboard Sections

#### Enrichment Status

A summary of AI enrichment activity across your library:

- **Total entities** -- Number of entities in the library.
- **AI-enriched** -- Number of entities that received AI metadata.
- **Pending enrichment** -- Entities flagged for AI processing but not yet processed.
- **Enrichment rate** -- Percentage of the library with AI metadata.

#### Recent AI Activity

A chronological log of recent AI operations:

- Entity name, operation type (metadata extraction, content analysis, duplicate detection).
- Timestamp and processing duration.
- Result status (success, partial, failed).
- Token usage and estimated cost for each operation.

#### Provider Health

Status of the configured LLM provider:

- **Connection status** -- Whether the provider API is reachable.
- **Average latency** -- Mean response time for recent requests.
- **Error rate** -- Percentage of failed requests in the last 24 hours.
- **Rate limit status** -- Current usage against the provider's rate limits.

#### Cost Summary

A breakdown of AI costs for the current billing period:

- Total tokens consumed (input and output).
- Estimated cost in the provider's pricing currency.
- Cost per entity (average).
- Daily and weekly cost trends as a line chart.

---

## Confidence Scores and Manual Overrides

Every piece of AI-generated metadata includes a confidence score indicating how certain the model is about the result.

### Confidence Levels

| Score Range | Level | Display | Meaning |
|-------------|-------|---------|---------|
| 0.90 -- 1.00 | High | Green badge | The model is highly confident. The result is almost certainly correct. |
| 0.70 -- 0.89 | Medium | Yellow badge | The model is moderately confident. Review recommended. |
| 0.50 -- 0.69 | Low | Orange badge | The model is uncertain. Manual verification is advised. |
| Below 0.50 | Very Low | Red badge | The model is guessing. Treat as a suggestion only. |

Confidence scores are displayed next to each AI-generated field on the entity detail page. Hover over the badge to see the exact numeric score.

### Automatic Thresholds

In **Settings > AI > Confidence Threshold**, you can set the minimum confidence score for AI metadata to be applied automatically. The default is 0.70 (medium confidence). Results below this threshold are stored but not displayed by default -- they appear in the AI review queue for manual approval.

### Manual Overrides

You can override any AI-generated metadata:

1. Open the entity detail page.
2. Click the edit icon next to an AI-generated field.
3. Modify the value and click **Save**.
4. The field is marked as "manually verified" and will not be overwritten by future AI enrichment runs.

### AI Review Queue

The review queue (accessible from the AI dashboard or **Admin > AI Review**) lists all AI results that fell below the confidence threshold. For each item, you can:

- **Accept** -- Apply the AI suggestion to the entity.
- **Edit and accept** -- Modify the suggestion before applying.
- **Reject** -- Discard the AI suggestion permanently.
- **Re-run** -- Request a new AI analysis for the entity.

---

## LLM Provider Configuration

Catalogizer supports multiple LLM providers for AI features. Configuration is managed from the admin panel at **Settings > AI > Providers** or via environment variables.

### Supported Providers

| Provider | Environment Variable | Models |
|----------|---------------------|--------|
| **OpenAI** | `OPENAI_API_KEY` | GPT-4o, GPT-4o mini |
| **Anthropic** | `ANTHROPIC_API_KEY` | Claude Sonnet, Claude Haiku |
| **Local (Ollama)** | `OLLAMA_URL` | Any model served by Ollama |

### Configuration via Environment Variables

Set the API key for your chosen provider in the `catalog-api/.env` file or as container environment variables:

```env
# OpenAI
OPENAI_API_KEY=sk-your-key-here
AI_PROVIDER=openai
AI_MODEL=gpt-4o-mini

# Anthropic
ANTHROPIC_API_KEY=sk-ant-your-key-here
AI_PROVIDER=anthropic
AI_MODEL=claude-sonnet-4-20250514

# Local Ollama
OLLAMA_URL=http://localhost:11434
AI_PROVIDER=ollama
AI_MODEL=llama3
```

### Configuration via Admin Panel

1. Navigate to **Admin > Settings > AI**.
2. Select the **Provider** from the dropdown.
3. Enter the **API Key** (not required for Ollama).
4. Select the **Model** from the list of available models for that provider.
5. Click **Test Connection** to verify the configuration.
6. Click **Save**.

### Provider Fallback

You can configure a primary and a fallback provider. If the primary provider is unreachable or returns errors, Catalogizer automatically routes requests to the fallback. Configure this in **Settings > AI > Fallback Provider**.

### Local Inference with Ollama

For users who prefer not to send data to external APIs, Catalogizer supports local inference via Ollama. Install Ollama on the Catalogizer server (or a machine on the same network), pull a model, and configure the URL. Local inference has no API costs but requires sufficient hardware (a GPU is recommended for reasonable performance).

---

## Cost Management

AI features consume API tokens, which incur costs with commercial LLM providers. Catalogizer provides tools to monitor and control spending.

### Token Usage Tracking

Every AI request logs the number of input and output tokens consumed. This data is visible in the AI dashboard's cost summary and in the recent activity log.

### Budget Limits

Set a monthly budget limit in **Settings > AI > Budget**:

- **Monthly limit** -- Maximum estimated cost per month (in USD). When the limit is reached, AI features are paused until the next billing cycle.
- **Per-entity limit** -- Maximum token usage for a single entity enrichment request. Prevents runaway costs from unusually complex prompts.
- **Daily limit** -- Optional daily spending cap for finer control.

When a limit is reached, the system logs a warning and queues remaining enrichment tasks for later processing. Manual AI operations from the entity detail page display a warning but can still be executed if the user confirms.

### Cost Optimization Tips

- **Use smaller models for bulk operations.** GPT-4o mini or Claude Haiku are significantly cheaper than their larger counterparts and produce adequate results for metadata extraction.
- **Set a high confidence threshold.** A threshold of 0.80 or above reduces the volume of AI requests by skipping entities where pattern matching and traditional providers already produced good results.
- **Run enrichment in batches.** Batch enrichment from the admin panel groups multiple entities into fewer API calls, reducing per-request overhead.
- **Use local inference for experimentation.** Test AI features with Ollama before committing to a commercial provider.

### Cost Estimation

Before running a batch enrichment, the admin panel shows an estimated cost based on the number of entities, the selected model's pricing, and the average prompt size. Review this estimate before confirming the operation.

---

## Best Practices

### Start with Traditional Providers

Enable TMDB, OMDB, OpenLibrary, and MusicBrainz before relying on AI. These providers are free (or very low cost) and cover the vast majority of well-known media. Use AI for the long tail of content that traditional providers cannot identify.

### Use AI for Cleanup, Not as Primary

AI works best as a cleanup and enrichment layer applied after the standard pipeline. Running AI on your entire library from scratch is expensive and unnecessary. Let the title parser and metadata providers do the bulk of the work, then use AI to fill remaining gaps.

### Review Low-Confidence Results

Do not blindly accept all AI output. Regularly check the AI review queue and verify low-confidence suggestions. AI models can hallucinate metadata -- a movie title might be matched to the wrong film, or a generated description might contain inaccuracies.

### Standardize File Naming First

AI metadata extraction produces better results when file names and directory structures follow common conventions. Before running AI enrichment, consider renaming poorly named files using the patterns described in the Entity User Guide (e.g., `Movie Title (Year).mkv`, `Show/Season XX/SXXEXX.mkv`).

### Monitor Costs Weekly

Check the AI dashboard cost summary at least weekly during initial setup. Costs can spike during the first enrichment pass across a large library. Once the initial pass is complete, ongoing costs are typically minimal (only new items trigger enrichment).

### Keep Models Updated

LLM providers periodically release improved models. Newer models often produce better results at lower cost. Check the provider configuration periodically and update to the recommended model version.

---

## Troubleshooting

### AI Features Not Available

- Verify that an LLM provider is configured in **Settings > AI**. AI features are disabled entirely when no provider is set.
- Check that the API key is valid. Use the **Test Connection** button in the provider settings.
- Ensure the Catalogizer API container has outbound internet access (required for OpenAI and Anthropic; not required for Ollama).

### AI Enrichment Produces No Results

- Check the AI dashboard's recent activity log for error messages. Common causes: rate limiting by the provider, invalid API key, or network timeout.
- If using Ollama, verify that the Ollama service is running and the model is downloaded (`ollama list`).
- Ensure the entity has enough contextual information (file path, partial title) for the AI to work with. A file named `video.mkv` in a flat directory provides almost no context.

### High Error Rate

- A spike in errors is often caused by provider-side issues (outages, rate limiting). Check the provider's status page.
- If errors are "rate limit exceeded", reduce the batch size for enrichment operations or add a delay between requests in **Settings > AI > Rate Limit Delay**.
- For timeout errors, increase the AI request timeout in **Settings > AI > Request Timeout** (default: 30 seconds).

### AI Metadata Overwrites Manual Edits

- By design, manually edited fields are marked as "verified" and should not be overwritten by subsequent AI runs. If this is happening, ensure you clicked **Save** after editing. Unsaved changes are not marked as verified.
- Check that the entity was not deleted and re-created by a re-scan (which would lose the manual edit marker). This can happen if storage roots are removed and re-added.

### Unexpected Costs

- Review the AI dashboard cost summary and recent activity log to identify which operations consumed the most tokens.
- Check whether batch enrichment was triggered inadvertently (e.g., by a scheduled task or an automated scan hook).
- Lower the budget limits in **Settings > AI > Budget** to prevent future overruns.
- Switch to a cheaper model for routine enrichment tasks.
