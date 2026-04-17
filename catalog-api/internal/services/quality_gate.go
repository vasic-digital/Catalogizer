package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"catalogizer/repository"

	"digital.vasic.assets/pkg/resolver"
	"digital.vasic.media/pkg/quality"
)

// HintFn maps a ResolveRequest onto a quality.Hint. The default implementation
// (DefaultHintFn) inspects req.EntityType and req.Metadata["variant"] to
// choose a Hint; callers can supply their own to override behavior.
type HintFn func(req *resolver.ResolveRequest) quality.Hint

// DefaultHintFn picks a Hint from the request's entity type. Variants named
// "backdrop" or "fanart" map to HintBackdrop regardless of entity.
func DefaultHintFn(req *resolver.ResolveRequest) quality.Hint {
	if req == nil {
		return quality.HintGeneric
	}
	variant := strings.ToLower(req.Metadata["variant"])
	if variant == "backdrop" || variant == "fanart" {
		return quality.HintBackdrop
	}
	switch strings.ToLower(req.EntityType) {
	case "movie", "movies":
		return quality.HintMoviePoster
	case "tv", "tv_show", "tv_season", "tv_episode":
		return quality.HintTvPoster
	case "music_artist", "music_album", "song":
		return quality.HintMusicAlbum
	case "book", "comic":
		return quality.HintBookCover
	case "game", "software":
		return quality.HintGameCover
	default:
		return quality.HintGeneric
	}
}

// QualityGate wraps another Resolver and applies deterministic quality checks
// to its output. A result that fails the gate is discarded and an error is
// returned so the enclosing ChainResolver advances to the next provider.
//
// When a *ImageQualityRepository is supplied, every assessment (pass or fail)
// is persisted so operators can audit which provider supplied each cover.
type QualityGate struct {
	inner    resolver.Resolver
	assessor *quality.Assessor
	hintFn   HintFn
	repo     *repository.ImageQualityRepository
	source   string

	mu        sync.RWMutex
	lastScore quality.Score
}

// QualityGateOption configures a QualityGate at construction time.
type QualityGateOption func(*QualityGate)

// WithQualityRepository enables persistence of every assessment.
func WithQualityRepository(repo *repository.ImageQualityRepository) QualityGateOption {
	return func(g *QualityGate) { g.repo = repo }
}

// WithQualityAssessor overrides the default pkg/quality assessor.
func WithQualityAssessor(a *quality.Assessor) QualityGateOption {
	return func(g *QualityGate) { g.assessor = a }
}

// WithHintFn overrides the default hint mapping.
func WithHintFn(fn HintFn) QualityGateOption {
	return func(g *QualityGate) { g.hintFn = fn }
}

// WithSourceLabel sets the label written to image_quality_assessments.source.
// When omitted, the gate falls back to the wrapped resolver's Name().
func WithSourceLabel(src string) QualityGateOption {
	return func(g *QualityGate) { g.source = src }
}

// NewQualityGate wraps inner with the quality gate.
func NewQualityGate(inner resolver.Resolver, opts ...QualityGateOption) *QualityGate {
	g := &QualityGate{
		inner:    inner,
		assessor: quality.NewAssessor(),
		hintFn:   DefaultHintFn,
	}
	for _, opt := range opts {
		opt(g)
	}
	if g.source == "" && inner != nil {
		g.source = inner.Name()
	}
	return g
}

// Name returns a composed name so logs distinguish gated from ungated resolvers.
func (g *QualityGate) Name() string {
	if g.inner == nil {
		return "quality_gate"
	}
	return "quality_gate(" + g.inner.Name() + ")"
}

// Priority defers to the wrapped resolver.
func (g *QualityGate) Priority() int {
	if g.inner == nil {
		return 0
	}
	return g.inner.Priority()
}

// CanResolve defers to the wrapped resolver.
func (g *QualityGate) CanResolve(ctx context.Context, req *resolver.ResolveRequest) bool {
	if g.inner == nil {
		return false
	}
	return g.inner.CanResolve(ctx, req)
}

// ErrQualityGateFailed is returned when a result fails the gate. It wraps the
// verdict string so callers can log or convert to headers.
var ErrQualityGateFailed = errors.New("quality gate failed")

// QualityGateError carries the Verdict and FailReason alongside the sentinel.
type QualityGateError struct {
	Verdict    string
	FailReason string
	Source     string
}

// Error satisfies the error interface.
func (e *QualityGateError) Error() string {
	return fmt.Sprintf("quality gate failed: source=%s verdict=%s reason=%s", e.Source, e.Verdict, e.FailReason)
}

// Unwrap exposes the sentinel for errors.Is.
func (e *QualityGateError) Unwrap() error { return ErrQualityGateFailed }

// Resolve asks the inner resolver for content and assesses it. On pass the
// bytes are returned unchanged; on fail a QualityGateError is returned so the
// outer chain advances.
func (g *QualityGate) Resolve(ctx context.Context, req *resolver.ResolveRequest) (*resolver.ResolveResult, error) {
	if g.inner == nil {
		return nil, fmt.Errorf("quality gate: no inner resolver configured")
	}

	result, err := g.inner.Resolve(ctx, req)
	if err != nil {
		return nil, err
	}
	if result == nil || result.Content == nil {
		return nil, fmt.Errorf("quality gate: inner returned nil result")
	}

	// Read all bytes so we can assess and also rewind for downstream reads.
	buf, readErr := io.ReadAll(result.Content)
	_ = result.Content.Close()
	if readErr != nil {
		return nil, fmt.Errorf("quality gate: read body: %w", readErr)
	}

	hint := g.hintFn(req)
	score, _ := g.assessor.Assess(buf, hint)

	g.mu.Lock()
	g.lastScore = score
	g.mu.Unlock()

	if g.repo != nil {
		_ = g.persist(ctx, req, score, buf)
	}

	if score.Verdict != quality.Pass {
		return nil, &QualityGateError{
			Verdict:    score.Verdict.String(),
			FailReason: score.FailReason,
			Source:     g.source,
		}
	}

	return &resolver.ResolveResult{
		Content:     io.NopCloser(bytes.NewReader(buf)),
		ContentType: result.ContentType,
		Size:        int64(len(buf)),
	}, nil
}

// LastScore returns a snapshot of the most recent assessment for observability.
// Concurrent callers receive a copy so there is no shared mutation.
func (g *QualityGate) LastScore() quality.Score {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.lastScore
}

func (g *QualityGate) persist(ctx context.Context, req *resolver.ResolveRequest, score quality.Score, _ []byte) error {
	if g.repo == nil || req == nil {
		return nil
	}
	entityID, err := strconv.ParseInt(req.EntityID, 10, 64)
	if err != nil {
		// Non-numeric entity IDs are valid in the asset system (UUIDs for
		// example). Skip persistence rather than fail the gate.
		return nil
	}
	variant := req.Metadata["variant"]
	if variant == "" {
		variant = "primary"
	}
	failReason := score.FailReason
	if failReason == "" && score.Verdict == quality.Pass {
		failReason = ""
	}
	// Closes W3 from docs/nexus/remaining-work.md: every string that
	// may carry provider-originated content is passed through the
	// project's bluemonday.StrictPolicy before persistence so a
	// compromised upstream cannot inject HTML/JS into the admin
	// tables.
	return g.repo.Upsert(ctx, &repository.ImageQualityAssessment{
		EntityType:    SanitizeMetadataString(req.EntityType),
		EntityID:      entityID,
		Variant:       SanitizeMetadataString(variant),
		Source:        SanitizeMetadataString(g.source),
		ProviderRef:   SanitizeMetadataString(req.SourceHint),
		Width:         score.Width,
		Height:        score.Height,
		BlurVar:       score.BlurVariance,
		BPP:           score.BytesPerPixel,
		AspectRatio:   score.AspectRatio,
		Verdict:       score.Verdict.String(),
		FailReason:    SanitizeMetadataString(failReason),
		Format:        SanitizeMetadataString(score.Format),
		AttemptCount:  1,
		AssessedAt:    time.Now(),
		LastCheckedAt: time.Now(),
	})
}

// Verify interface compliance at compile time.
var _ resolver.Resolver = (*QualityGate)(nil)
