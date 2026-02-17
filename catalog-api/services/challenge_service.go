package services

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/registry"
	"digital.vasic.challenges/pkg/runner"
)

// ChallengeService manages challenge registration, execution,
// and result retrieval.
type ChallengeService struct {
	mu         sync.RWMutex
	registry   *registry.DefaultRegistry
	runner     *runner.DefaultRunner
	resultsDir string
	results    []*challenge.Result
}

// NewChallengeService creates a new challenge service with
// the given results directory.
func NewChallengeService(resultsDir string) *ChallengeService {
	reg := registry.NewRegistry()
	r := runner.NewRunner(
		runner.WithRegistry(reg),
		runner.WithTimeout(10*time.Minute),
		runner.WithResultsDir(resultsDir),
	)
	return &ChallengeService{
		registry:   reg,
		runner:     r,
		resultsDir: resultsDir,
	}
}

// Registry returns the underlying challenge registry for
// external registration.
func (s *ChallengeService) Registry() *registry.DefaultRegistry {
	return s.registry
}

// Register adds a challenge to the registry.
func (s *ChallengeService) Register(c challenge.Challenge) error {
	return s.registry.Register(c)
}

// ListChallenges returns all registered challenges as
// summary items.
func (s *ChallengeService) ListChallenges() []ChallengeSummary {
	challenges := s.registry.List()
	summaries := make([]ChallengeSummary, len(challenges))
	for i, c := range challenges {
		deps := c.Dependencies()
		depStrings := make([]string, len(deps))
		for j, d := range deps {
			depStrings[j] = string(d)
		}
		summaries[i] = ChallengeSummary{
			ID:           string(c.ID()),
			Name:         c.Name(),
			Description:  c.Description(),
			Category:     c.Category(),
			Dependencies: depStrings,
		}
	}
	return summaries
}

// RunChallenge executes a single challenge by ID.
func (s *ChallengeService) RunChallenge(
	ctx context.Context, id string,
) (*challenge.Result, error) {
	cfg := &challenge.Config{
		ResultsDir: filepath.Join(
			s.resultsDir, id,
			time.Now().Format("20060102_150405"),
		),
		Timeout: 10 * time.Minute,
		Verbose: true,
	}
	result, err := s.runner.Run(
		ctx, challenge.ID(id), cfg,
	)
	if err != nil {
		return nil, fmt.Errorf("run challenge %s: %w", id, err)
	}

	s.mu.Lock()
	s.results = append(s.results, result)
	s.mu.Unlock()

	return result, nil
}

// RunAll executes all challenges in dependency order.
func (s *ChallengeService) RunAll(
	ctx context.Context,
) ([]*challenge.Result, error) {
	cfg := &challenge.Config{
		ResultsDir: filepath.Join(
			s.resultsDir, "all",
			time.Now().Format("20060102_150405"),
		),
		Timeout: 10 * time.Minute,
		Verbose: true,
	}
	results, err := s.runner.RunAll(ctx, cfg)
	if err != nil {
		return results, fmt.Errorf("run all challenges: %w", err)
	}

	s.mu.Lock()
	s.results = append(s.results, results...)
	s.mu.Unlock()

	return results, nil
}

// RunByCategory executes all challenges in a category.
func (s *ChallengeService) RunByCategory(
	ctx context.Context, category string,
) ([]*challenge.Result, error) {
	challenges := s.registry.ListByCategory(category)
	if len(challenges) == 0 {
		return nil, fmt.Errorf(
			"no challenges found for category: %s", category,
		)
	}

	ids := make([]challenge.ID, len(challenges))
	for i, c := range challenges {
		ids[i] = c.ID()
	}

	cfg := &challenge.Config{
		ResultsDir: filepath.Join(
			s.resultsDir, "category", category,
			time.Now().Format("20060102_150405"),
		),
		Timeout: 10 * time.Minute,
		Verbose: true,
	}
	results, err := s.runner.RunSequence(ctx, ids, cfg)
	if err != nil {
		return results, fmt.Errorf(
			"run category %s: %w", category, err,
		)
	}

	s.mu.Lock()
	s.results = append(s.results, results...)
	s.mu.Unlock()

	return results, nil
}

// GetResults returns all stored challenge results.
func (s *ChallengeService) GetResults() []*challenge.Result {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*challenge.Result, len(s.results))
	copy(out, s.results)
	return out
}

// ChallengeSummary is a lightweight representation of a
// registered challenge.
type ChallengeSummary struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Dependencies []string `json:"dependencies"`
}
