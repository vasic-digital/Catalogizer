package services

import (
	"context"
	"testing"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testChallenge is a minimal challenge for testing.
type testChallenge struct {
	challenge.BaseChallenge
	shouldFail bool
}

func newTestChallenge(
	id, name, category string, shouldFail bool,
) *testChallenge {
	return &testChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			challenge.ID(id), name,
			"Test challenge: "+name,
			category, nil,
		),
		shouldFail: shouldFail,
	}
}

func (c *testChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {
	start := time.Now()
	status := challenge.StatusPassed
	assertions := []challenge.AssertionResult{
		{
			Type:    "not_empty",
			Target:  "test_output",
			Passed:  !c.shouldFail,
			Message: "test assertion",
		},
	}
	if c.shouldFail {
		status = challenge.StatusFailed
	}
	return c.CreateResult(
		status, start, assertions, nil,
		map[string]string{"output": "test"},
		"",
	), nil
}

func TestNewChallengeService(t *testing.T) {
	svc := NewChallengeService(t.TempDir())
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.Registry())
}

func TestChallengeService_Register(t *testing.T) {
	svc := NewChallengeService(t.TempDir())

	err := svc.Register(
		newTestChallenge("test_1", "Test One", "unit", false),
	)
	require.NoError(t, err)

	challenges := svc.ListChallenges()
	assert.Len(t, challenges, 1)
	assert.Equal(t, "test_1", challenges[0].ID)
	assert.Equal(t, "Test One", challenges[0].Name)
	assert.Equal(t, "unit", challenges[0].Category)
}

func TestChallengeService_ListChallenges_Empty(t *testing.T) {
	svc := NewChallengeService(t.TempDir())
	challenges := svc.ListChallenges()
	assert.Empty(t, challenges)
}

func TestChallengeService_ListChallenges_Multiple(t *testing.T) {
	svc := NewChallengeService(t.TempDir())

	svc.Register(
		newTestChallenge("a", "Alpha", "integration", false),
	)
	svc.Register(
		newTestChallenge("b", "Beta", "unit", false),
	)
	svc.Register(
		newTestChallenge("c", "Charlie", "integration", false),
	)

	challenges := svc.ListChallenges()
	assert.Len(t, challenges, 3)
	// Registry sorts by ID
	assert.Equal(t, "a", challenges[0].ID)
	assert.Equal(t, "b", challenges[1].ID)
	assert.Equal(t, "c", challenges[2].ID)
}

func TestChallengeService_RunChallenge(t *testing.T) {
	svc := NewChallengeService(t.TempDir())

	svc.Register(
		newTestChallenge("run_test", "Run Test", "unit", false),
	)

	result, err := svc.RunChallenge(
		context.Background(), "run_test",
	)
	require.NoError(t, err)
	assert.Equal(t, challenge.StatusPassed, result.Status)
	assert.Equal(t, "Run Test", result.ChallengeName)
	assert.True(t, result.AllPassed())
}

func TestChallengeService_RunChallenge_Failing(t *testing.T) {
	svc := NewChallengeService(t.TempDir())

	svc.Register(
		newTestChallenge("fail_test", "Fail Test", "unit", true),
	)

	result, err := svc.RunChallenge(
		context.Background(), "fail_test",
	)
	require.NoError(t, err)
	assert.Equal(t, challenge.StatusFailed, result.Status)
	assert.False(t, result.AllPassed())
}

func TestChallengeService_RunChallenge_NotFound(t *testing.T) {
	svc := NewChallengeService(t.TempDir())

	_, err := svc.RunChallenge(
		context.Background(), "nonexistent",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestChallengeService_RunAll(t *testing.T) {
	svc := NewChallengeService(t.TempDir())

	svc.Register(
		newTestChallenge("all_1", "All One", "unit", false),
	)
	svc.Register(
		newTestChallenge("all_2", "All Two", "unit", false),
	)

	results, err := svc.RunAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, results, 2)
	for _, r := range results {
		assert.Equal(t, challenge.StatusPassed, r.Status)
	}
}

func TestChallengeService_GetResults(t *testing.T) {
	svc := NewChallengeService(t.TempDir())

	svc.Register(
		newTestChallenge("res_1", "Res One", "unit", false),
	)

	// No results initially
	assert.Empty(t, svc.GetResults())

	// Run a challenge
	svc.RunChallenge(context.Background(), "res_1")

	// Should have 1 result
	results := svc.GetResults()
	assert.Len(t, results, 1)
	assert.Equal(t, challenge.ID("res_1"), results[0].ChallengeID)
}

func TestChallengeService_DuplicateRegister(t *testing.T) {
	svc := NewChallengeService(t.TempDir())

	err := svc.Register(
		newTestChallenge("dup", "Dup", "unit", false),
	)
	require.NoError(t, err)

	err = svc.Register(
		newTestChallenge("dup", "Dup Again", "unit", false),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}
