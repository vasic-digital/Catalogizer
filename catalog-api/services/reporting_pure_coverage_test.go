package services

import (
	"testing"
	"time"

	"catalogizer/models"

	"github.com/stretchr/testify/require"
)

// Pure-helper coverage tests for ReportingService. These take in-memory
// slices and don't touch the DB or analytics repo, so they exercise
// pure-Go paths that previously had ~0% coverage.

func newReportingPureTestService() *ReportingService {
	return &ReportingService{}
}

func atHour(h int) time.Time {
	return time.Date(2026, 4, 11, h, 0, 0, 0, time.UTC)
}

func TestReportingService_AnalyzeTimeDistribution_AllBuckets(t *testing.T) {
	s := newReportingPureTestService()
	logs := []models.MediaAccessLog{
		{AccessTime: atHour(8)},  // morning
		{AccessTime: atHour(11)}, // morning
		{AccessTime: atHour(13)}, // afternoon
		{AccessTime: atHour(17)}, // afternoon
		{AccessTime: atHour(19)}, // evening
		{AccessTime: atHour(21)}, // evening
		{AccessTime: atHour(23)}, // night
		{AccessTime: atHour(2)},  // night
		{AccessTime: atHour(5)},  // night (boundary: 5 < 6 → night)
	}
	dist := s.analyzeTimeDistribution(logs)
	require.Equal(t, 2, dist["morning"])
	require.Equal(t, 2, dist["afternoon"])
	require.Equal(t, 2, dist["evening"])
	require.Equal(t, 3, dist["night"])
}

func TestReportingService_AnalyzeTimeDistribution_Empty(t *testing.T) {
	s := newReportingPureTestService()
	dist := s.analyzeTimeDistribution(nil)
	require.NotNil(t, dist)
	require.Empty(t, dist)
}

func TestReportingService_AnalyzeTimeDistribution_ExactBoundaries(t *testing.T) {
	s := newReportingPureTestService()
	// Hour 6 = morning, 12 = afternoon, 18 = evening, 22 = night, 0 = night
	logs := []models.MediaAccessLog{
		{AccessTime: atHour(6)},  // morning
		{AccessTime: atHour(12)}, // afternoon
		{AccessTime: atHour(18)}, // evening
		{AccessTime: atHour(22)}, // night
		{AccessTime: atHour(0)},  // night
	}
	dist := s.analyzeTimeDistribution(logs)
	require.Equal(t, 1, dist["morning"])
	require.Equal(t, 1, dist["afternoon"])
	require.Equal(t, 1, dist["evening"])
	require.Equal(t, 2, dist["night"])
}

func TestReportingService_GetLastActivityTime_Empty(t *testing.T) {
	s := newReportingPureTestService()
	got := s.getLastActivityTime(nil)
	require.True(t, got.IsZero())
}

func TestReportingService_GetLastActivityTime_PicksLatest(t *testing.T) {
	s := newReportingPureTestService()
	t1 := atHour(8)
	t2 := atHour(15)
	t3 := atHour(21)
	logs := []models.MediaAccessLog{
		{AccessTime: t1},
		{AccessTime: t3}, // latest — out of order
		{AccessTime: t2},
	}
	got := s.getLastActivityTime(logs)
	require.Equal(t, t3, got)
}

func TestReportingService_GetMostActiveHour_Empty(t *testing.T) {
	s := newReportingPureTestService()
	require.Equal(t, 0, s.getMostActiveHour(nil))
}

func TestReportingService_GetMostActiveHour_PicksMode(t *testing.T) {
	s := newReportingPureTestService()
	logs := []models.MediaAccessLog{
		{AccessTime: atHour(9)},
		{AccessTime: atHour(15)},
		{AccessTime: atHour(15)},
		{AccessTime: atHour(15)}, // hour 15 has 3 hits, the most
		{AccessTime: atHour(20)},
	}
	require.Equal(t, 15, s.getMostActiveHour(logs))
}

func TestReportingService_CalculateAverageSessionDuration_Empty(t *testing.T) {
	s := newReportingPureTestService()
	require.Equal(t, time.Duration(0), s.calculateAverageSessionDuration(nil))
}

func TestReportingService_CalculateAverageSessionDuration_Single(t *testing.T) {
	s := newReportingPureTestService()
	sessions := []models.SessionData{
		{Duration: 5 * time.Minute},
	}
	require.Equal(t, 5*time.Minute, s.calculateAverageSessionDuration(sessions))
}

func TestReportingService_CalculateAverageSessionDuration_Multiple(t *testing.T) {
	s := newReportingPureTestService()
	sessions := []models.SessionData{
		{Duration: 10 * time.Minute},
		{Duration: 20 * time.Minute},
		{Duration: 30 * time.Minute},
	}
	require.Equal(t, 20*time.Minute, s.calculateAverageSessionDuration(sessions))
}
