package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"catalogizer/database"
	"catalogizer/models"
)

type StressTestRepository struct {
	db *database.DB
}

func NewStressTestRepository(db *database.DB) *StressTestRepository {
	return &StressTestRepository{db: db}
}

func (r *StressTestRepository) CreateStressTest(test *models.StressTest) error {
	scenariosJSON, err := json.Marshal(test.Scenarios)
	if err != nil {
		return fmt.Errorf("failed to marshal scenarios: %w", err)
	}

	configJSON, err := json.Marshal(test.Configuration)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	query := `
		INSERT INTO stress_tests (
			name, description, type, status, scenarios, configuration,
			concurrent_users, duration_seconds, ramp_up_time, created_by, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	id, err := r.db.InsertReturningID(context.Background(), query,
		test.Name, test.Description, test.Type, test.Status,
		string(scenariosJSON), string(configJSON),
		test.ConcurrentUsers, test.DurationSeconds, test.RampUpTime,
		test.CreatedBy, test.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create stress test: %w", err)
	}

	test.ID = id
	return nil
}

func (r *StressTestRepository) GetStressTest(id int) (*models.StressTest, error) {
	query := `
		SELECT id, name, description, type, status, scenarios, configuration,
			   concurrent_users, duration_seconds, ramp_up_time, created_by,
			   created_at, started_at, completed_at
		FROM stress_tests WHERE id = ?`

	var test models.StressTest
	var scenariosJSON, configJSON string
	var startedAt, completedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&test.ID, &test.Name, &test.Description, &test.Type, &test.Status,
		&scenariosJSON, &configJSON, &test.ConcurrentUsers,
		&test.DurationSeconds, &test.RampUpTime, &test.CreatedBy,
		&test.CreatedAt, &startedAt, &completedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get stress test: %w", err)
	}

	if startedAt.Valid {
		test.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		test.CompletedAt = &completedAt.Time
	}

	if err := json.Unmarshal([]byte(scenariosJSON), &test.Scenarios); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scenarios: %w", err)
	}

	if err := json.Unmarshal([]byte(configJSON), &test.Configuration); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &test, nil
}

func (r *StressTestRepository) UpdateStressTest(test *models.StressTest) error {
	scenariosJSON, err := json.Marshal(test.Scenarios)
	if err != nil {
		return fmt.Errorf("failed to marshal scenarios: %w", err)
	}

	configJSON, err := json.Marshal(test.Configuration)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	query := `
		UPDATE stress_tests SET
			name = ?, description = ?, type = ?, status = ?,
			scenarios = ?, configuration = ?, concurrent_users = ?,
			duration_seconds = ?, ramp_up_time = ?, started_at = ?, completed_at = ?
		WHERE id = ?`

	_, err = r.db.Exec(query,
		test.Name, test.Description, test.Type, test.Status,
		string(scenariosJSON), string(configJSON),
		test.ConcurrentUsers, test.DurationSeconds, test.RampUpTime,
		test.StartedAt, test.CompletedAt, test.ID)

	if err != nil {
		return fmt.Errorf("failed to update stress test: %w", err)
	}

	return nil
}

func (r *StressTestRepository) GetStressTestsByUser(userID int, limit, offset int) ([]*models.StressTest, error) {
	query := `
		SELECT id, name, description, type, status, scenarios, configuration,
			   concurrent_users, duration_seconds, ramp_up_time, created_by,
			   created_at, started_at, completed_at
		FROM stress_tests
		WHERE created_by = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get stress tests: %w", err)
	}
	defer rows.Close()

	var tests []*models.StressTest
	for rows.Next() {
		var test models.StressTest
		var scenariosJSON, configJSON string
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&test.ID, &test.Name, &test.Description, &test.Type, &test.Status,
			&scenariosJSON, &configJSON, &test.ConcurrentUsers,
			&test.DurationSeconds, &test.RampUpTime, &test.CreatedBy,
			&test.CreatedAt, &startedAt, &completedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan stress test: %w", err)
		}

		if startedAt.Valid {
			test.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			test.CompletedAt = &completedAt.Time
		}

		if err := json.Unmarshal([]byte(scenariosJSON), &test.Scenarios); err != nil {
			return nil, fmt.Errorf("failed to unmarshal scenarios: %w", err)
		}

		if err := json.Unmarshal([]byte(configJSON), &test.Configuration); err != nil {
			return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
		}

		tests = append(tests, &test)
	}

	return tests, nil
}

func (r *StressTestRepository) DeleteStressTest(id int) error {
	// First delete associated executions
	if err := r.DeleteExecutionsByTestID(id); err != nil {
		return fmt.Errorf("failed to delete test executions: %w", err)
	}

	query := "DELETE FROM stress_tests WHERE id = ?"
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete stress test: %w", err)
	}

	return nil
}

func (r *StressTestRepository) CreateExecution(execution *models.StressTestExecution) error {
	metricsJSON, err := json.Marshal(execution.Metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	resultsJSON, err := json.Marshal(execution.Results)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	query := `
		INSERT INTO stress_test_executions (
			stress_test_id, status, started_at, metrics, results, error_message
		) VALUES (?, ?, ?, ?, ?, ?)`

	id, err := r.db.InsertReturningID(context.Background(), query,
		execution.StressTestID, execution.Status, execution.StartedAt,
		string(metricsJSON), string(resultsJSON), execution.ErrorMessage)

	if err != nil {
		return fmt.Errorf("failed to create execution: %w", err)
	}

	execution.ID = id
	return nil
}

func (r *StressTestRepository) UpdateExecution(execution *models.StressTestExecution) error {
	metricsJSON, err := json.Marshal(execution.Metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	resultsJSON, err := json.Marshal(execution.Results)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	query := `
		UPDATE stress_test_executions SET
			status = ?, completed_at = ?, metrics = ?, results = ?, error_message = ?
		WHERE id = ?`

	_, err = r.db.Exec(query,
		execution.Status, execution.CompletedAt,
		string(metricsJSON), string(resultsJSON),
		execution.ErrorMessage, execution.ID)

	if err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	return nil
}

func (r *StressTestRepository) GetExecution(id int) (*models.StressTestExecution, error) {
	query := `
		SELECT id, stress_test_id, status, started_at, completed_at,
			   metrics, results, error_message
		FROM stress_test_executions WHERE id = ?`

	var execution models.StressTestExecution
	var metricsJSON, resultsJSON string
	var completedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&execution.ID, &execution.StressTestID, &execution.Status,
		&execution.StartedAt, &completedAt,
		&metricsJSON, &resultsJSON, &execution.ErrorMessage)

	if err != nil {
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	if completedAt.Valid {
		execution.CompletedAt = &completedAt.Time
	}

	if err := json.Unmarshal([]byte(metricsJSON), &execution.Metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	if err := json.Unmarshal([]byte(resultsJSON), &execution.Results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal results: %w", err)
	}

	return &execution, nil
}

func (r *StressTestRepository) GetExecutionsByTestID(testID int) ([]*models.StressTestExecution, error) {
	query := `
		SELECT id, stress_test_id, status, started_at, completed_at,
			   metrics, results, error_message
		FROM stress_test_executions
		WHERE stress_test_id = ?
		ORDER BY started_at DESC`

	rows, err := r.db.Query(query, testID)
	if err != nil {
		return nil, fmt.Errorf("failed to get executions: %w", err)
	}
	defer rows.Close()

	var executions []*models.StressTestExecution
	for rows.Next() {
		var execution models.StressTestExecution
		var metricsJSON, resultsJSON string
		var completedAt sql.NullTime

		err := rows.Scan(
			&execution.ID, &execution.StressTestID, &execution.Status,
			&execution.StartedAt, &completedAt,
			&metricsJSON, &resultsJSON, &execution.ErrorMessage)

		if err != nil {
			return nil, fmt.Errorf("failed to scan execution: %w", err)
		}

		if completedAt.Valid {
			execution.CompletedAt = &completedAt.Time
		}

		if err := json.Unmarshal([]byte(metricsJSON), &execution.Metrics); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
		}

		if err := json.Unmarshal([]byte(resultsJSON), &execution.Results); err != nil {
			return nil, fmt.Errorf("failed to unmarshal results: %w", err)
		}

		executions = append(executions, &execution)
	}

	return executions, nil
}

func (r *StressTestRepository) DeleteExecutionsByTestID(testID int) error {
	query := "DELETE FROM stress_test_executions WHERE stress_test_id = ?"
	_, err := r.db.Exec(query, testID)
	if err != nil {
		return fmt.Errorf("failed to delete executions: %w", err)
	}
	return nil
}

func (r *StressTestRepository) GetStatistics(userID int) (*models.StressTestStatistics, error) {
	stats := &models.StressTestStatistics{}

	// Total tests
	err := r.db.QueryRow("SELECT COUNT(*) FROM stress_tests WHERE created_by = ?", userID).Scan(&stats.TotalTests)
	if err != nil {
		return nil, fmt.Errorf("failed to get total tests: %w", err)
	}

	// Tests by status
	statusQuery := `
		SELECT status, COUNT(*)
		FROM stress_tests
		WHERE created_by = ?
		GROUP BY status`

	rows, err := r.db.Query(statusQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}
	defer rows.Close()

	stats.TestsByStatus = make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status count: %w", err)
		}
		stats.TestsByStatus[status] = count
	}

	// Total executions
	err = r.db.QueryRow(`
		SELECT COUNT(*)
		FROM stress_test_executions ste
		JOIN stress_tests st ON ste.stress_test_id = st.id
		WHERE st.created_by = ?`, userID).Scan(&stats.TotalExecutions)
	if err != nil {
		return nil, fmt.Errorf("failed to get total executions: %w", err)
	}

	// Average execution duration
	err = r.db.QueryRow(`
		SELECT AVG(
			CASE
				WHEN ste.completed_at IS NOT NULL
				THEN (julianday(ste.completed_at) - julianday(ste.started_at)) * 24 * 60 * 60
				ELSE NULL
			END
		)
		FROM stress_test_executions ste
		JOIN stress_tests st ON ste.stress_test_id = st.id
		WHERE st.created_by = ? AND ste.completed_at IS NOT NULL`, userID).Scan(&stats.AvgExecutionDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to get average duration: %w", err)
	}

	return stats, nil
}

func (r *StressTestRepository) CleanupOldExecutions(olderThan time.Time) error {
	query := "DELETE FROM stress_test_executions WHERE started_at < ?"
	_, err := r.db.Exec(query, olderThan)
	if err != nil {
		return fmt.Errorf("failed to cleanup old executions: %w", err)
	}
	return nil
}

// Wrapper methods to match service expectations
func (r *StressTestRepository) CreateTest(test *models.StressTest) (int64, error) {
	err := r.CreateStressTest(test)
	if err != nil {
		return 0, err
	}
	return test.ID, nil
}

func (r *StressTestRepository) GetTest(id int) (*models.StressTest, error) {
	return r.GetStressTest(id)
}

func (r *StressTestRepository) UpdateTest(test *models.StressTest) error {
	return r.UpdateStressTest(test)
}

func (r *StressTestRepository) GetUserTests(userID int, limit, offset int) ([]*models.StressTest, error) {
	return r.GetStressTestsByUser(userID, limit, offset)
}

func (r *StressTestRepository) DeleteTest(id int) error {
	return r.DeleteStressTest(id)
}

func (r *StressTestRepository) SaveResult(result *models.StressTestResult) error {
	// Results are stored in executions, so we create/update an execution entry
	execution := &models.StressTestExecution{
		StressTestID: result.TestID,
		Status:       result.Status,
		StartedAt:    &result.StartTime,
		CompletedAt:  result.EndTime,
		Metrics: map[string]interface{}{
			"total_requests":       result.Metrics.TotalRequests,
			"successful_requests":  result.Metrics.SuccessfulRequests,
			"failed_requests":      result.Metrics.FailedRequests,
			"avg_response_time_ns": int64(result.Metrics.AverageResponseTime),
			"min_response_time_ns": int64(result.Metrics.MinResponseTime),
			"max_response_time_ns": int64(result.Metrics.MaxResponseTime),
			"p95_response_time_ns": int64(result.Metrics.P95ResponseTime),
			"p99_response_time_ns": int64(result.Metrics.P99ResponseTime),
			"requests_per_second":  result.Metrics.RequestsPerSecond,
			"error_rate":           result.Metrics.ErrorRate,
			"throughput":           result.Metrics.Throughput,
		},
		Results: models.StressTestExecutionResults{
			TotalRequests:       result.TotalRequests,
			SuccessfulRequests:  result.SuccessfulReqs,
			FailedRequests:      result.FailedRequests,
			AverageResponseTime: result.AvgResponseTime,
			MinResponseTime:     result.MinResponseTime,
			MaxResponseTime:     result.MaxResponseTime,
			ErrorRate:           result.ErrorRate,
			Throughput:          result.RequestsPerSecond,
		},
		ErrorMessage: result.ErrorMessage,
	}
	return r.CreateExecution(execution)
}

func (r *StressTestRepository) GetResult(testID int) (*models.StressTestResult, error) {
	// Get the latest execution for this test
	executions, err := r.GetExecutionsByTestID(testID)
	if err != nil {
		return nil, fmt.Errorf("failed to get executions: %w", err)
	}

	if len(executions) == 0 {
		return nil, fmt.Errorf("no results found for test %d", testID)
	}

	// Get the most recent execution (last in list)
	execution := executions[len(executions)-1]

	// Build StressTestResult from execution
	result := &models.StressTestResult{
		TestID:       execution.StressTestID,
		Status:       execution.Status,
		ErrorMessage: execution.ErrorMessage,
	}

	if execution.StartedAt != nil {
		result.StartTime = *execution.StartedAt
	}
	result.EndTime = execution.CompletedAt
	result.CompletedAt = execution.CompletedAt

	if execution.CompletedAt != nil && execution.StartedAt != nil {
		result.Duration = execution.CompletedAt.Sub(*execution.StartedAt)
	}

	// Extract from Results struct
	result.TotalRequests = execution.Results.TotalRequests
	result.SuccessfulReqs = execution.Results.SuccessfulRequests
	result.FailedRequests = execution.Results.FailedRequests
	result.AvgResponseTime = execution.Results.AverageResponseTime
	result.MinResponseTime = execution.Results.MinResponseTime
	result.MaxResponseTime = execution.Results.MaxResponseTime
	result.ErrorRate = execution.Results.ErrorRate
	result.RequestsPerSecond = execution.Results.Throughput

	// Extract metrics from map
	if execution.Metrics != nil {
		if v, ok := execution.Metrics["total_requests"].(float64); ok {
			result.Metrics.TotalRequests = int64(v)
		}
		if v, ok := execution.Metrics["successful_requests"].(float64); ok {
			result.Metrics.SuccessfulRequests = int64(v)
		}
		if v, ok := execution.Metrics["failed_requests"].(float64); ok {
			result.Metrics.FailedRequests = int64(v)
		}
		if v, ok := execution.Metrics["avg_response_time_ns"].(float64); ok {
			result.Metrics.AverageResponseTime = time.Duration(v)
		}
		if v, ok := execution.Metrics["min_response_time_ns"].(float64); ok {
			result.Metrics.MinResponseTime = time.Duration(v)
		}
		if v, ok := execution.Metrics["max_response_time_ns"].(float64); ok {
			result.Metrics.MaxResponseTime = time.Duration(v)
		}
		if v, ok := execution.Metrics["p95_response_time_ns"].(float64); ok {
			result.Metrics.P95ResponseTime = time.Duration(v)
		}
		if v, ok := execution.Metrics["p99_response_time_ns"].(float64); ok {
			result.Metrics.P99ResponseTime = time.Duration(v)
		}
		if v, ok := execution.Metrics["requests_per_second"].(float64); ok {
			result.Metrics.RequestsPerSecond = v
		}
		if v, ok := execution.Metrics["error_rate"].(float64); ok {
			result.Metrics.ErrorRate = v
		}
		if v, ok := execution.Metrics["throughput"].(float64); ok {
			result.Metrics.Throughput = v
		}
	}

	return result, nil
}
