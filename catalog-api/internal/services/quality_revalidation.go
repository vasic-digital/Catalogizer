package services

import (
	"context"
	"sync"
	"time"

	"catalogizer/repository"

	"go.uber.org/zap"
)

// QualityRevalidator periodically re-checks stored image-quality assessments
// so that entries whose source thresholds have since tightened (or whose
// underlying bytes have been replaced) eventually get re-resolved.
//
// The worker is intentionally minimal: it is not responsible for re-fetching
// bytes; it simply touches last_checked_at on rows older than the configured
// freshness window, which signals the cover-art service to re-resolve on
// next access. A more active re-download can be layered on later without
// changing the worker's contract.
type QualityRevalidator struct {
	repo     *repository.ImageQualityRepository
	logger   *zap.Logger
	interval time.Duration
	staleAge time.Duration
	batch    int

	stopOnce sync.Once
	stop     chan struct{}
	wg       sync.WaitGroup
}

// QualityRevalidatorOption configures the revalidator.
type QualityRevalidatorOption func(*QualityRevalidator)

// WithRevalidationInterval overrides the default 24h tick.
func WithRevalidationInterval(d time.Duration) QualityRevalidatorOption {
	return func(r *QualityRevalidator) {
		if d > 0 {
			r.interval = d
		}
	}
}

// WithRevalidationStaleAge overrides the default 7d stale threshold.
func WithRevalidationStaleAge(d time.Duration) QualityRevalidatorOption {
	return func(r *QualityRevalidator) {
		if d > 0 {
			r.staleAge = d
		}
	}
}

// WithRevalidationBatch caps rows touched per tick.
func WithRevalidationBatch(n int) QualityRevalidatorOption {
	return func(r *QualityRevalidator) {
		if n > 0 {
			r.batch = n
		}
	}
}

// NewQualityRevalidator returns a worker wired to the supplied repository.
func NewQualityRevalidator(repo *repository.ImageQualityRepository, logger *zap.Logger, opts ...QualityRevalidatorOption) *QualityRevalidator {
	if logger == nil {
		logger = zap.NewNop()
	}
	r := &QualityRevalidator{
		repo:     repo,
		logger:   logger,
		interval: 24 * time.Hour,
		staleAge: 7 * 24 * time.Hour,
		batch:    256,
		stop:     make(chan struct{}),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Start launches the revalidation loop. Start is safe to call once; repeated
// calls have no effect.
func (r *QualityRevalidator) Start(ctx context.Context) {
	if r == nil || r.repo == nil {
		return
	}
	r.wg.Add(1)
	go r.loop(ctx)
}

// Stop signals the loop to exit and waits for it to finish.
func (r *QualityRevalidator) Stop() {
	if r == nil {
		return
	}
	r.stopOnce.Do(func() { close(r.stop) })
	r.wg.Wait()
}

func (r *QualityRevalidator) loop(ctx context.Context) {
	defer r.wg.Done()
	// Tick on first iteration so a fresh startup touches stale rows promptly.
	r.runOnce(ctx)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-r.stop:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.runOnce(ctx)
		}
	}
}

// runOnce performs one sweep; exported only through tests in the same package
// via an alias in the test file.
func (r *QualityRevalidator) runOnce(ctx context.Context) {
	cutoff := time.Now().Add(-r.staleAge)
	rows, err := r.repo.SampleForRevalidation(ctx, cutoff, r.batch)
	if err != nil {
		r.logger.Warn("quality revalidation: sample failed", zap.Error(err))
		return
	}
	for _, row := range rows {
		if err := r.repo.TouchLastChecked(ctx, row.ID); err != nil {
			r.logger.Warn("quality revalidation: touch failed",
				zap.Int64("id", row.ID), zap.Error(err))
		}
	}
	r.logger.Debug("quality revalidation: swept", zap.Int("touched", len(rows)))
}
