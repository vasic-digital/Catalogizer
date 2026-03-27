# Module 23: Subtitle & Conversion Tools -- Video Script

**Duration**: 45 minutes
**Prerequisites**: Module 4 (Media Detection and Processing), Module 7 (Storage Protocols)

---

## Video 23.1: Subtitle Search and Download (12 min)

### Opening

Welcome to Module 23. Subtitles and format conversion are essential features for any media collection manager. Catalogizer provides a unified subtitle pipeline -- search, download, upload, translate -- and a job-based conversion system with real-time progress tracking. This module covers both subsystems end to end.

### Subtitle Provider Architecture

**[Visual: Architecture diagram: SubtitleService -> ProviderManager -> OpenSubtitles / Subscene / Addic7ed / Local Files]**

**Narrator**: The subtitle system follows the same provider pattern as metadata enrichment. A `SubtitleProvider` interface abstracts the differences between subtitle sources. Catalogizer ships with integrations for OpenSubtitles, Subscene, and Addic7ed, plus a local file scanner that finds subtitle files already present alongside media files.

```go
// catalog-api/internal/services/subtitle_service.go
type SubtitleProvider interface {
    GetName() string
    Search(ctx context.Context, query SubtitleQuery) ([]SubtitleResult, error)
    Download(ctx context.Context, subtitleID string) ([]byte, error)
    IsEnabled() bool
    GetRateLimit() time.Duration
}

type SubtitleQuery struct {
    Title      string  `json:"title"`
    Year       *int    `json:"year,omitempty"`
    Season     *int    `json:"season,omitempty"`
    Episode    *int    `json:"episode,omitempty"`
    Language   string  `json:"language"`
    FileHash   string  `json:"file_hash,omitempty"`
    FileSize   int64   `json:"file_size,omitempty"`
    MediaType  string  `json:"media_type"`
}
```

**[Visual: Show `SubtitleResult` struct]**

**Narrator**: Each search result includes the provider name, subtitle language, format, download count (as a quality signal), and a match confidence score. Results from all enabled providers are merged and ranked.

```go
// catalog-api/internal/services/subtitle_service.go
type SubtitleResult struct {
    ID            string  `json:"id"`
    Provider      string  `json:"provider"`
    Title         string  `json:"title"`
    Language      string  `json:"language"`
    Format        string  `json:"format"` // srt, ass, vtt, sub
    DownloadCount int     `json:"download_count"`
    Confidence    float64 `json:"confidence"`
    MatchMethod   string  `json:"match_method"` // hash, title, imdb_id
}
```

### Hash-Based Matching

**[Visual: Show the hash calculation flow for a media file]**

**Narrator**: The most accurate subtitle matching uses file hashes. OpenSubtitles uses a custom 64-bit hash based on file size and the first and last 64KB of the file. This hash uniquely identifies a specific release of a movie or episode, matching subtitles that are frame-synchronized with that exact file.

```go
// catalog-api/internal/services/subtitle_hash.go
func ComputeOpenSubtitlesHash(filePath string) (string, int64, error) {
    f, err := os.Open(filePath)
    if err != nil {
        return "", 0, fmt.Errorf("open file: %w", err)
    }
    defer f.Close()

    fi, err := f.Stat()
    if err != nil {
        return "", 0, fmt.Errorf("stat file: %w", err)
    }
    fileSize := fi.Size()

    hash := uint64(fileSize)

    // Read first 64KB
    buf := make([]byte, 65536)
    if _, err := io.ReadFull(f, buf); err != nil {
        return "", 0, fmt.Errorf("read head: %w", err)
    }
    for i := 0; i < len(buf); i += 8 {
        hash += binary.LittleEndian.Uint64(buf[i : i+8])
    }

    // Read last 64KB
    if _, err := f.Seek(-65536, io.SeekEnd); err != nil {
        return "", 0, fmt.Errorf("seek tail: %w", err)
    }
    if _, err := io.ReadFull(f, buf); err != nil {
        return "", 0, fmt.Errorf("read tail: %w", err)
    }
    for i := 0; i < len(buf); i += 8 {
        hash += binary.LittleEndian.Uint64(buf[i : i+8])
    }

    return fmt.Sprintf("%016x", hash), fileSize, nil
}
```

### Search Flow

**[Visual: Show the search UI with results from multiple providers]**

**Narrator**: When a user searches for subtitles, the service queries all enabled providers in parallel, merges results, deduplicates by language and format, and returns a ranked list. Hash-matched results rank highest, followed by IMDB ID matches, then title matches.

```go
// catalog-api/internal/services/subtitle_service.go
func (s *SubtitleService) Search(
    ctx context.Context,
    entityID int64,
    language string,
) ([]SubtitleResult, error) {
    entity, err := s.entityRepo.GetByID(ctx, entityID)
    if err != nil {
        return nil, fmt.Errorf("get entity: %w", err)
    }

    query := s.buildQuery(entity, language)

    var allResults []SubtitleResult
    var mu sync.Mutex
    var wg sync.WaitGroup

    for _, provider := range s.providers {
        if !provider.IsEnabled() {
            continue
        }
        wg.Add(1)
        go func(p SubtitleProvider) {
            defer wg.Done()
            results, err := p.Search(ctx, query)
            if err != nil {
                s.logger.Warn("Subtitle search failed",
                    zap.String("provider", p.GetName()),
                    zap.Error(err))
                return
            }
            mu.Lock()
            allResults = append(allResults, results...)
            mu.Unlock()
        }(provider)
    }
    wg.Wait()

    return s.rankAndDeduplicate(allResults), nil
}
```

---

## Video 23.2: Subtitle Upload and Translation (10 min)

### Uploading Custom Subtitles

**[Visual: Show the subtitle upload dialog in the entity detail page]**

**Narrator**: Users can upload their own subtitle files for any entity. The upload endpoint accepts SRT, ASS, VTT, and SUB formats. Uploaded subtitles are stored alongside the media file and linked to the entity in the database.

```
POST /api/v1/entities/:id/subtitles/upload
     Content-Type: multipart/form-data
     Fields: file, language, format (auto-detected if omitted)
```

**[Visual: Show the subtitle file being stored]**

**Narrator**: The uploaded file is validated for format correctness -- timestamp parsing, encoding detection (UTF-8, UTF-16, Latin-1), and line count sanity checks. Invalid files are rejected with a descriptive error.

```go
// catalog-api/internal/services/subtitle_service.go
func (s *SubtitleService) Upload(
    ctx context.Context,
    entityID int64,
    file io.Reader,
    filename string,
    language string,
) (*SubtitleRecord, error) {
    data, err := io.ReadAll(io.LimitReader(file, s.maxUploadSize))
    if err != nil {
        return nil, fmt.Errorf("read upload: %w", err)
    }

    format := s.detectFormat(filename, data)
    if format == "" {
        return nil, fmt.Errorf("unsupported subtitle format")
    }

    encoding := s.detectEncoding(data)
    if encoding != "utf-8" {
        data, err = s.convertToUTF8(data, encoding)
        if err != nil {
            return nil, fmt.Errorf("encoding conversion: %w", err)
        }
    }

    if err := s.validateSubtitle(data, format); err != nil {
        return nil, fmt.Errorf("invalid subtitle: %w", err)
    }

    return s.store(ctx, entityID, data, format, language)
}
```

### Translation Pipeline

**[Visual: Diagram showing subtitle text extraction -> translation API -> subtitle reconstruction]**

**Narrator**: The translation pipeline converts subtitles from one language to another while preserving timing, formatting, and style tags. It extracts plain text from each cue, sends batches to a translation service, and reassembles the subtitle file with translated text and original timestamps.

```go
// catalog-api/internal/services/subtitle_translation_service.go
type TranslationRequest struct {
    EntityID       int64  `json:"entity_id"`
    SubtitleID     int64  `json:"subtitle_id"`
    SourceLanguage string `json:"source_language"`
    TargetLanguage string `json:"target_language"`
}

type TranslationProgress struct {
    RequestID    string  `json:"request_id"`
    Status       string  `json:"status"` // pending, translating, complete, failed
    TotalCues    int     `json:"total_cues"`
    TranslatedCues int  `json:"translated_cues"`
    Percent      float64 `json:"percent"`
}
```

**Narrator**: Translation is a long-running operation. It runs as a background job with progress tracking. The frontend polls the progress endpoint or receives WebSocket updates as each batch completes.

---

## Video 23.3: Format Conversion Jobs (10 min)

### Conversion Job System

**[Visual: Architecture diagram: ConversionService -> JobQueue -> Worker Pool -> FFmpeg / HandBrake CLI]**

**Narrator**: The conversion system handles media format transformations: video transcoding, audio conversion, container remuxing, and subtitle format conversion. All conversions run as background jobs with progress tracking and cancellation support.

```go
// catalog-api/internal/services/conversion_service.go
type ConversionJob struct {
    ID             string    `json:"id"`
    EntityID       int64     `json:"entity_id"`
    SourceFileID   int64     `json:"source_file_id"`
    SourcePath     string    `json:"source_path"`
    TargetFormat   string    `json:"target_format"`
    TargetCodec    string    `json:"target_codec,omitempty"`
    TargetQuality  string    `json:"target_quality,omitempty"`
    OutputPath     string    `json:"output_path"`
    Status         string    `json:"status"` // queued, running, complete, failed, cancelled
    Progress       float64   `json:"progress"`
    StartedAt      *time.Time `json:"started_at,omitempty"`
    CompletedAt    *time.Time `json:"completed_at,omitempty"`
    ErrorMessage   string    `json:"error_message,omitempty"`
    CreatedAt      time.Time `json:"created_at"`
}
```

### Worker Pool

**[Visual: Show the worker pool processing conversion jobs sequentially]**

**Narrator**: Conversion jobs are resource-intensive. The worker pool limits concurrency to prevent overwhelming the host. By default, one conversion runs at a time -- this respects the 30-40% host resource limit. Each worker wraps FFmpeg or HandBrake CLI with progress parsing.

```go
// catalog-api/internal/services/conversion_service.go
type ConversionService struct {
    db        *database.DB
    jobQueue  chan *ConversionJob
    workers   int
    stopCh    chan struct{}
    wg        sync.WaitGroup
    logger    *zap.Logger
    activeJob *ConversionJob
    mu        sync.RWMutex
}

func (s *ConversionService) worker(ctx context.Context) {
    defer s.wg.Done()
    for {
        select {
        case <-ctx.Done():
            return
        case <-s.stopCh:
            return
        case job := <-s.jobQueue:
            s.processJob(ctx, job)
        }
    }
}
```

### FFmpeg Integration

**[Visual: Terminal showing FFmpeg output with progress parsing]**

**Narrator**: FFmpeg is invoked as a subprocess. The service constructs the command line from the job specification and parses stderr for progress information. FFmpeg reports progress as frame count, time position, speed, and bitrate. The service converts the time position into a percentage based on the source file's duration.

```go
// catalog-api/internal/services/conversion_service.go
func (s *ConversionService) runFFmpeg(
    ctx context.Context,
    job *ConversionJob,
) error {
    args := s.buildFFmpegArgs(job)
    cmd := exec.CommandContext(ctx, "ffmpeg", args...)

    stderr, err := cmd.StderrPipe()
    if err != nil {
        return fmt.Errorf("stderr pipe: %w", err)
    }

    if err := cmd.Start(); err != nil {
        return fmt.Errorf("start ffmpeg: %w", err)
    }

    // Parse progress from stderr
    scanner := bufio.NewScanner(stderr)
    scanner.Split(scanFFmpegProgress)
    for scanner.Scan() {
        line := scanner.Text()
        if progress, ok := parseFFmpegProgress(line, job.sourceDuration); ok {
            s.updateProgress(job, progress)
        }
    }

    return cmd.Wait()
}
```

---

## Video 23.4: Progress Tracking and Supported Formats (13 min)

### Real-Time Progress

**[Visual: Browser showing a conversion job with a progress bar updating in real time]**

**Narrator**: Conversion progress is tracked at two levels. The service writes progress to the database every 5 seconds for persistence across restarts. Simultaneously, progress events are published to the WebSocket event bus for real-time UI updates.

```go
// catalog-api/internal/services/conversion_service.go
func (s *ConversionService) updateProgress(
    job *ConversionJob,
    percent float64,
) {
    s.mu.Lock()
    job.Progress = percent
    s.mu.Unlock()

    // Persist to database (throttled to every 5 seconds)
    if time.Since(job.lastDBUpdate) > 5*time.Second {
        s.db.ExecContext(context.Background(),
            "UPDATE conversion_jobs SET progress = ? WHERE id = ?",
            percent, job.ID)
        job.lastDBUpdate = time.Now()
    }

    // Broadcast via WebSocket
    s.eventBus.Publish("conversion_progress", ConversionProgressEvent{
        JobID:    job.ID,
        Progress: percent,
        Status:   job.Status,
    })
}
```

**[Visual: Show React component consuming progress updates]**

**Narrator**: The frontend subscribes to conversion progress events through the WebSocket hook. The progress bar component renders the percentage and estimated time remaining, calculated from the rate of progress change.

```typescript
// catalog-web/src/hooks/useConversionProgress.ts
export function useConversionProgress(jobId: string) {
  const [progress, setProgress] = useState<ConversionProgress | null>(null);

  useWebSocket('conversion_progress', (event: ConversionProgressEvent) => {
    if (event.job_id === jobId) {
      setProgress({
        percent: event.progress,
        status: event.status,
        estimatedRemaining: calculateETA(event),
      });
    }
  });

  return progress;
}
```

### Job Management API

**[Visual: Show the conversion API endpoints]**

**Narrator**: The conversion API provides endpoints for creating, monitoring, and managing jobs.

```
POST   /api/v1/conversions            -- Create conversion job
GET    /api/v1/conversions             -- List all jobs (with filters)
GET    /api/v1/conversions/:id         -- Get job status and progress
DELETE /api/v1/conversions/:id         -- Cancel a running job
GET    /api/v1/conversions/:id/log     -- Get FFmpeg output log
POST   /api/v1/conversions/batch       -- Create multiple jobs at once
```

### Supported Formats

**[Visual: Table showing supported input and output formats]**

**Narrator**: Catalogizer supports a wide range of media formats through FFmpeg.

**Video Containers and Codecs:**

| Category | Formats |
|----------|---------|
| Containers | MP4, MKV, AVI, MOV, WebM, FLV, TS, M2TS |
| Video Codecs | H.264, H.265 (HEVC), VP9, AV1, MPEG-2, ProRes |
| Audio Codecs | AAC, AC3, DTS, FLAC, Opus, MP3, Vorbis, PCM |

**Subtitle Formats:**

| Format | Extension | Description |
|--------|-----------|-------------|
| SubRip | .srt | Plain text with timestamps |
| Advanced SubStation | .ass | Styled text with positioning |
| WebVTT | .vtt | Web-native subtitle format |
| MicroDVD | .sub | Frame-based subtitle format |
| SubStation Alpha | .ssa | Legacy styled subtitle format |

**Subtitle Conversion:**

**Narrator**: Subtitle format conversion is handled in-process without FFmpeg. The service parses the source format into an intermediate representation (list of timed cues with text and style), then serializes to the target format.

```go
// catalog-api/internal/services/subtitle_converter.go
type SubtitleCue struct {
    Index     int           `json:"index"`
    StartTime time.Duration `json:"start_time"`
    EndTime   time.Duration `json:"end_time"`
    Text      string        `json:"text"`
    Style     *CueStyle     `json:"style,omitempty"`
}

func ConvertSubtitle(
    source []byte,
    sourceFormat string,
    targetFormat string,
) ([]byte, error) {
    cues, err := parseSubtitle(source, sourceFormat)
    if err != nil {
        return nil, fmt.Errorf("parse %s: %w", sourceFormat, err)
    }
    return serializeSubtitle(cues, targetFormat)
}
```

---

## Key Code Examples

### Subtitle API Endpoints
```
GET    /api/v1/entities/:id/subtitles          -- List subtitles for entity
POST   /api/v1/entities/:id/subtitles/search   -- Search providers
POST   /api/v1/entities/:id/subtitles/download -- Download from provider
POST   /api/v1/entities/:id/subtitles/upload   -- Upload custom subtitle
POST   /api/v1/entities/:id/subtitles/translate -- Start translation job
GET    /api/v1/subtitles/translation/:id       -- Translation progress
POST   /api/v1/subtitles/convert               -- Convert subtitle format
```

### Full Subtitle Flow
```
1. User opens entity detail page
2. Existing subtitles listed from database
3. User clicks "Search Subtitles"
4. All enabled providers queried in parallel
5. Results merged, deduplicated, and ranked
6. User selects a result and clicks "Download"
7. Subtitle file downloaded and stored alongside media
8. File linked to entity in database
9. Optional: user requests translation to another language
10. Translation runs as background job with progress tracking
```

---

## Key Files Referenced

- `catalog-api/internal/services/subtitle_service.go` -- Search, download, upload
- `catalog-api/internal/services/subtitle_hash.go` -- OpenSubtitles hash computation
- `catalog-api/internal/services/subtitle_translation_service.go` -- Translation pipeline
- `catalog-api/internal/services/subtitle_converter.go` -- Format conversion (SRT/ASS/VTT/SUB)
- `catalog-api/internal/services/conversion_service.go` -- Media conversion job system
- `catalog-api/internal/media/realtime/` -- WebSocket progress broadcasting

---

## Exercises

1. Implement a subtitle format converter that transforms SRT to WebVTT, preserving italic and bold HTML tags while converting timestamps from SRT format (`00:01:23,456`) to VTT format (`00:01:23.456`).
2. Write a rate limiter for subtitle provider requests that respects each provider's `GetRateLimit()` value, using a token bucket per provider.
3. Add a new subtitle provider integration for a public subtitle API of your choice. Implement the `SubtitleProvider` interface with search and download.
4. Extend the conversion service to support hardware-accelerated encoding via VAAPI or NVENC, with automatic fallback to software encoding when hardware is unavailable.

---

## Quiz Questions

1. What is the most accurate method for matching subtitles to a media file, and how does it work?
   **Answer**: Hash-based matching using the OpenSubtitles 64-bit hash. It computes a hash from the file size plus the first and last 64KB of the file. This uniquely identifies a specific release, ensuring frame-synchronized subtitles.

2. How does the conversion service respect host resource limits?
   **Answer**: The worker pool defaults to one concurrent conversion job. This respects the 30-40% host resource limit. Jobs are queued and processed sequentially. The concurrency is configurable but should not exceed the resource budget.

3. How is conversion progress tracked in real time?
   **Answer**: FFmpeg stderr is parsed for time position. The service converts this to a percentage using the source duration. Progress is written to the database every 5 seconds for persistence and published to the WebSocket event bus for real-time UI updates.

4. What happens when a subtitle file is uploaded with non-UTF-8 encoding?
   **Answer**: The service detects the encoding (UTF-16, Latin-1, etc.) and converts the content to UTF-8 before validation and storage. This ensures all stored subtitles are in a consistent encoding.
