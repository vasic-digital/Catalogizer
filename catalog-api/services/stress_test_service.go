package services

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"catalogizer/models"
	"catalogizer/repository"
)

type StressTestService struct {
	stressRepo  *repository.StressTestRepository
	authService *AuthService
	activeTests map[int]*TestExecution
	testMutex   sync.RWMutex
}

type TestExecution struct {
	Test      *models.StressTest
	Context   context.Context
	Cancel    context.CancelFunc
	Workers   []*TestWorker
	Metrics   *TestMetrics
	StartTime time.Time
	IsRunning bool
	Results   *models.StressTestResult
}

type TestWorker struct {
	ID      int
	Context context.Context
	Metrics *WorkerMetrics
}

type TestMetrics struct {
	TotalRequests     int64
	SuccessfulReqs    int64
	FailedRequests    int64
	TotalResponseTime time.Duration
	MinResponseTime   time.Duration
	MaxResponseTime   time.Duration
	ErrorCounts       map[string]int64
	StatusCounts      map[int]int64
	mutex             sync.RWMutex
}

type WorkerMetrics struct {
	RequestCount    int64
	SuccessCount    int64
	ErrorCount      int64
	TotalRespTime   time.Duration
	LastRequestTime time.Time
}

func NewStressTestService(stressRepo *repository.StressTestRepository, authService *AuthService) *StressTestService {
	return &StressTestService{
		stressRepo:  stressRepo,
		authService: authService,
		activeTests: make(map[int]*TestExecution),
	}
}

func (s *StressTestService) CreateStressTest(userID int, test *models.StressTest) (*models.StressTest, error) {
	hasPermission, err := s.authService.CheckPermission(userID, models.PermissionSystemAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions")
	}

	if err := s.validateTestConfiguration(test); err != nil {
		return nil, fmt.Errorf("invalid test configuration: %w", err)
	}

	test.UserID = userID
	test.CreatedAt = time.Now()
	test.Status = models.StressTestStatusPending

	id, err := s.stressRepo.CreateTest(test)
	if err != nil {
		return nil, fmt.Errorf("failed to create stress test: %w", err)
	}

	test.ID = id
	return test, nil
}

func (s *StressTestService) StartStressTest(testID int, userID int) (*models.StressTestResult, error) {
	hasPermission, err := s.authService.CheckPermission(userID, models.PermissionSystemAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions")
	}

	test, err := s.stressRepo.GetTest(testID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test: %w", err)
	}

	if test.Status != models.StressTestStatusPending {
		return nil, fmt.Errorf("test is not in pending status")
	}

	s.testMutex.Lock()
	if _, exists := s.activeTests[testID]; exists {
		s.testMutex.Unlock()
		return nil, fmt.Errorf("test is already running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(test.Duration)*time.Second)
	execution := &TestExecution{
		Test:      test,
		Context:   ctx,
		Cancel:    cancel,
		Workers:   make([]*TestWorker, test.ConcurrentUsers),
		Metrics:   s.createTestMetrics(),
		StartTime: time.Now(),
		IsRunning: true,
	}

	s.activeTests[testID] = execution
	s.testMutex.Unlock()

	test.Status = models.StressTestStatusRunning
	test.StartedAt = &execution.StartTime
	s.stressRepo.UpdateTest(test)

	go s.executeStressTest(execution)

	result := &models.StressTestResult{
		TestID:    int64(testID),
		Status:    models.StressTestStatusRunning,
		StartTime: execution.StartTime,
	}

	return result, nil
}

func (s *StressTestService) executeStressTest(execution *TestExecution) {
	defer func() {
		s.testMutex.Lock()
		delete(s.activeTests, int(execution.Test.ID))
		s.testMutex.Unlock()

		if r := recover(); r != nil {
			s.handleTestError(execution, fmt.Errorf("test panic: %v", r))
		}
	}()

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < execution.Test.ConcurrentUsers; i++ {
		worker := &TestWorker{
			ID:      i,
			Context: execution.Context,
			Metrics: &WorkerMetrics{},
		}
		execution.Workers[i] = worker

		wg.Add(1)
		go func(w *TestWorker) {
			defer wg.Done()
			s.runWorker(execution, w)
		}(worker)
	}

	// Wait for completion or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.handleTestCompletion(execution)
	case <-execution.Context.Done():
		s.handleTestTimeout(execution)
	}
}

func (s *StressTestService) runWorker(execution *TestExecution, worker *TestWorker) {
	client := &http.Client{
		Timeout: time.Duration(execution.Test.RequestTimeout) * time.Second,
	}

	for {
		select {
		case <-worker.Context.Done():
			return
		default:
			s.executeRequest(execution, worker, client)

			if execution.Test.RequestDelay > 0 {
				time.Sleep(time.Duration(execution.Test.RequestDelay) * time.Millisecond)
			}
		}
	}
}

func (s *StressTestService) executeRequest(execution *TestExecution, worker *TestWorker, client *http.Client) {
	startTime := time.Now()
	atomic.AddInt64(&worker.Metrics.RequestCount, 1)
	atomic.AddInt64(&execution.Metrics.TotalRequests, 1)

	scenario := s.selectRandomScenario(execution.Test.Scenarios)
	if scenario == nil {
		return
	}

	req, err := s.buildRequest(scenario)
	if err != nil {
		s.recordError(execution, worker, err, 0)
		return
	}

	resp, err := client.Do(req)
	responseTime := time.Since(startTime)

	if err != nil {
		s.recordError(execution, worker, err, 0)
		return
	}
	defer resp.Body.Close()

	s.recordResponse(execution, worker, resp, responseTime)
}

func (s *StressTestService) selectRandomScenario(scenarios []models.StressTestScenario) *models.StressTestScenario {
	if len(scenarios) == 0 {
		return nil
	}

	totalWeight := 0
	for _, scenario := range scenarios {
		totalWeight += scenario.Weight
	}

	if totalWeight == 0 {
		// #nosec G404 - math/rand is appropriate for test scenario selection (non-cryptographic)
		return &scenarios[rand.Intn(len(scenarios))]
	}

	// #nosec G404 - math/rand is appropriate for test scenario selection (non-cryptographic)
	randomWeight := rand.Intn(totalWeight)
	currentWeight := 0

	for _, scenario := range scenarios {
		currentWeight += scenario.Weight
		if randomWeight < currentWeight {
			return &scenario
		}
	}

	return &scenarios[0]
}

func (s *StressTestService) buildRequest(scenario *models.StressTestScenario) (*http.Request, error) {
	var body io.Reader
	if scenario.RequestBody != nil {
		body = strings.NewReader(*scenario.RequestBody)
	}

	req, err := http.NewRequest(scenario.Method, scenario.URL, body)
	if err != nil {
		return nil, err
	}

	for key, value := range scenario.Headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

func (s *StressTestService) recordResponse(execution *TestExecution, worker *TestWorker, resp *http.Response, responseTime time.Duration) {
	execution.Metrics.mutex.Lock()
	defer execution.Metrics.mutex.Unlock()

	execution.Metrics.TotalResponseTime += responseTime

	if execution.Metrics.MinResponseTime == 0 || responseTime < execution.Metrics.MinResponseTime {
		execution.Metrics.MinResponseTime = responseTime
	}

	if responseTime > execution.Metrics.MaxResponseTime {
		execution.Metrics.MaxResponseTime = responseTime
	}

	if execution.Metrics.StatusCounts == nil {
		execution.Metrics.StatusCounts = make(map[int]int64)
	}

	execution.Metrics.StatusCounts[resp.StatusCode]++

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		atomic.AddInt64(&execution.Metrics.SuccessfulReqs, 1)
		atomic.AddInt64(&worker.Metrics.SuccessCount, 1)
	} else {
		atomic.AddInt64(&execution.Metrics.FailedRequests, 1)
		atomic.AddInt64(&worker.Metrics.ErrorCount, 1)
	}

	worker.Metrics.TotalRespTime += responseTime
	worker.Metrics.LastRequestTime = time.Now()
}

func (s *StressTestService) recordError(execution *TestExecution, worker *TestWorker, err error, statusCode int) {
	execution.Metrics.mutex.Lock()
	defer execution.Metrics.mutex.Unlock()

	atomic.AddInt64(&execution.Metrics.FailedRequests, 1)
	atomic.AddInt64(&worker.Metrics.ErrorCount, 1)

	if execution.Metrics.ErrorCounts == nil {
		execution.Metrics.ErrorCounts = make(map[string]int64)
	}

	execution.Metrics.ErrorCounts[err.Error()]++

	if statusCode > 0 {
		if execution.Metrics.StatusCounts == nil {
			execution.Metrics.StatusCounts = make(map[int]int64)
		}
		execution.Metrics.StatusCounts[statusCode]++
	}
}

func (s *StressTestService) handleTestCompletion(execution *TestExecution) {
	execution.IsRunning = false
	result := s.generateTestResult(execution, models.StressTestStatusCompleted)

	execution.Test.Status = models.StressTestStatusCompleted
	execution.Test.CompletedAt = result.CompletedAt
	s.stressRepo.UpdateTest(execution.Test)

	s.saveTestResult(result)
}

func (s *StressTestService) handleTestTimeout(execution *TestExecution) {
	execution.IsRunning = false
	result := s.generateTestResult(execution, models.StressTestStatusTimeout)

	execution.Test.Status = models.StressTestStatusTimeout
	execution.Test.CompletedAt = result.CompletedAt
	s.stressRepo.UpdateTest(execution.Test)

	s.saveTestResult(result)
}

func (s *StressTestService) handleTestError(execution *TestExecution, testError error) {
	execution.IsRunning = false
	result := s.generateTestResult(execution, models.StressTestStatusFailed)
	errorMsg := testError.Error()
	result.ErrorMessage = &errorMsg

	execution.Test.Status = models.StressTestStatusFailed
	execution.Test.CompletedAt = result.CompletedAt
	s.stressRepo.UpdateTest(execution.Test)

	s.saveTestResult(result)
}

func (s *StressTestService) generateTestResult(execution *TestExecution, status string) *models.StressTestResult {
	completedAt := time.Now()
	duration := completedAt.Sub(execution.StartTime)

	avgResponseTime := time.Duration(0)
	if execution.Metrics.TotalRequests > 0 {
		avgResponseTime = execution.Metrics.TotalResponseTime / time.Duration(execution.Metrics.TotalRequests)
	}

	requestsPerSecond := float64(0)
	if duration.Seconds() > 0 {
		requestsPerSecond = float64(execution.Metrics.TotalRequests) / duration.Seconds()
	}

	errorRate := float64(0)
	if execution.Metrics.TotalRequests > 0 {
		errorRate = float64(execution.Metrics.FailedRequests) / float64(execution.Metrics.TotalRequests) * 100
	}

	return &models.StressTestResult{
		TestID:            execution.Test.ID,
		Status:            status,
		StartTime:         execution.StartTime,
		EndTime:           &completedAt,
		CompletedAt:       &completedAt,
		Duration:          duration,
		TotalRequests:     execution.Metrics.TotalRequests,
		SuccessfulReqs:    execution.Metrics.SuccessfulReqs,
		FailedRequests:    execution.Metrics.FailedRequests,
		RequestsPerSecond: requestsPerSecond,
		AvgResponseTime:   float64(avgResponseTime),
		MinResponseTime:   float64(execution.Metrics.MinResponseTime),
		MaxResponseTime:   float64(execution.Metrics.MaxResponseTime),
		ErrorRate:         errorRate,
		StatusCodeDist:    convertIntMapToStringInt(execution.Metrics.StatusCounts),
		ErrorDistribution: convertInt64MapToInt(execution.Metrics.ErrorCounts),
	}
}

func convertIntMapToStringInt(m map[int]int64) map[string]int {
	result := make(map[string]int)
	for k, v := range m {
		result[fmt.Sprintf("%d", k)] = int(v)
	}
	return result
}

func convertInt64MapToInt(m map[string]int64) map[string]int {
	result := make(map[string]int)
	for k, v := range m {
		result[k] = int(v)
	}
	return result
}

func (s *StressTestService) saveTestResult(result *models.StressTestResult) {
	err := s.stressRepo.SaveResult(result)
	if err != nil {
		fmt.Printf("Failed to save test result for test %d: %v\n", result.TestID, err)
	}
}

func (s *StressTestService) StopStressTest(testID int, userID int) error {
	hasPermission, err := s.authService.CheckPermission(userID, models.PermissionSystemAdmin)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}

	if !hasPermission {
		return fmt.Errorf("insufficient permissions")
	}

	s.testMutex.RLock()
	execution, exists := s.activeTests[testID]
	s.testMutex.RUnlock()

	if !exists {
		return fmt.Errorf("test is not running")
	}

	execution.Cancel()

	execution.Test.Status = models.StressTestStatusCancelled
	completedAt := time.Now()
	execution.Test.CompletedAt = &completedAt
	s.stressRepo.UpdateTest(execution.Test)

	return nil
}

func (s *StressTestService) GetTestStatus(testID int, userID int) (*models.StressTestStatus, error) {
	hasPermission, err := s.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions")
	}

	test, err := s.stressRepo.GetTest(testID)
	if err != nil {
		return nil, err
	}

	status := &models.StressTestStatus{
		TestID:    int64(testID),
		Status:    test.Status,
		CreatedAt: test.CreatedAt,
		StartedAt: *test.StartedAt,
	}

	s.testMutex.RLock()
	execution, isRunning := s.activeTests[testID]
	s.testMutex.RUnlock()

	if isRunning {
		status.IsRunning = true
		status.ElapsedTime = time.Since(execution.StartTime)
		status.TotalRequests = execution.Metrics.TotalRequests
		status.SuccessfulReqs = execution.Metrics.SuccessfulReqs
		status.FailedRequests = execution.Metrics.FailedRequests

		if execution.Metrics.TotalRequests > 0 {
			status.RequestsPerSecond = float64(execution.Metrics.TotalRequests) / status.ElapsedTime.Seconds()
			status.ErrorRate = float64(execution.Metrics.FailedRequests) / float64(execution.Metrics.TotalRequests) * 100
		}
	}

	return status, nil
}

func (s *StressTestService) GetTestResults(testID int, userID int) (*models.StressTestResult, error) {
	hasPermission, err := s.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions")
	}

	return s.stressRepo.GetResult(testID)
}

func (s *StressTestService) ListUserTests(userID int, limit, offset int) ([]*models.StressTest, error) {
	hasPermission, err := s.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions")
	}

	return s.stressRepo.GetUserTests(userID, limit, offset)
}

func (s *StressTestService) DeleteTest(testID int, userID int) error {
	hasPermission, err := s.authService.CheckPermission(userID, models.PermissionSystemAdmin)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}

	if !hasPermission {
		return fmt.Errorf("insufficient permissions")
	}

	test, err := s.stressRepo.GetTest(testID)
	if err != nil {
		return err
	}

	if test.Status == models.StressTestStatusRunning {
		return fmt.Errorf("cannot delete running test")
	}

	return s.stressRepo.DeleteTest(testID)
}

func (s *StressTestService) GenerateLoadReport(testID int, userID int) (*models.LoadTestReport, error) {
	hasPermission, err := s.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions")
	}

	test, err := s.stressRepo.GetTest(testID)
	if err != nil {
		return nil, err
	}

	result, err := s.stressRepo.GetResult(testID)
	if err != nil {
		return nil, err
	}

	report := &models.LoadTestReport{
		Test:            test,
		Result:          result,
		GeneratedAt:     time.Now(),
		Summary:         s.generateReportSummary(test, result),
		Recommendations: s.generateRecommendations(result),
	}

	return report, nil
}

func (s *StressTestService) generateReportSummary(test *models.StressTest, result *models.StressTestResult) string {
	return fmt.Sprintf(
		"Load test completed with %d concurrent users over %d seconds. "+
			"Total requests: %d, Success rate: %.2f%%, Average response time: %v",
		test.ConcurrentUsers,
		test.Duration,
		result.TotalRequests,
		100-result.ErrorRate,
		result.AvgResponseTime,
	)
}

func (s *StressTestService) generateRecommendations(result *models.StressTestResult) []string {
	var recommendations []string

	if result.ErrorRate > 5 {
		recommendations = append(recommendations, "High error rate detected. Consider investigating server capacity and error handling.")
	}

	if result.AvgResponseTime > float64(2*time.Second) {
		recommendations = append(recommendations, "Average response time is high. Consider optimizing database queries and caching.")
	}

	if result.RequestsPerSecond < 10 {
		recommendations = append(recommendations, "Low throughput detected. Consider scaling horizontally or optimizing server performance.")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "System performance appears to be within acceptable limits.")
	}

	return recommendations
}

func (s *StressTestService) validateTestConfiguration(test *models.StressTest) error {
	if test.Name == "" {
		return fmt.Errorf("test name is required")
	}

	if test.ConcurrentUsers <= 0 {
		return fmt.Errorf("concurrent users must be greater than 0")
	}

	if test.ConcurrentUsers > 1000 {
		return fmt.Errorf("concurrent users cannot exceed 1000")
	}

	if test.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}

	if test.Duration > 3600 {
		return fmt.Errorf("duration cannot exceed 3600 seconds")
	}

	if len(test.Scenarios) == 0 {
		return fmt.Errorf("at least one test scenario is required")
	}

	for _, scenario := range test.Scenarios {
		if scenario.URL == "" {
			return fmt.Errorf("scenario URL is required")
		}

		if scenario.Method == "" {
			return fmt.Errorf("scenario method is required")
		}

		if scenario.Weight < 0 {
			return fmt.Errorf("scenario weight cannot be negative")
		}
	}

	return nil
}

func (s *StressTestService) createTestMetrics() *TestMetrics {
	return &TestMetrics{
		ErrorCounts:  make(map[string]int64),
		StatusCounts: make(map[int]int64),
	}
}

func (s *StressTestService) GetSystemLoad() (*models.SystemLoadMetrics, error) {
	// This would integrate with system monitoring tools
	// For now, return placeholder data
	return &models.SystemLoadMetrics{
		CPUUsage:    45.2,
		MemoryUsage: 68.5,
		DiskUsage:   32.1,
		NetworkIO:   15.8,
		ActiveTests: len(s.activeTests),
		Timestamp:   time.Now(),
	}, nil
}
