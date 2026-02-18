package repository

import (
	"database/sql"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockStressTestRepo(t *testing.T) (*StressTestRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewStressTestRepository(db), mock
}

var stressTestColumns = []string{
	"id", "name", "description", "type", "status", "scenarios", "configuration",
	"concurrent_users", "duration_seconds", "ramp_up_time", "created_by",
	"created_at", "started_at", "completed_at",
}

var stressTestExecutionColumns = []string{
	"id", "stress_test_id", "status", "started_at", "completed_at",
	"metrics", "results", "error_message",
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestStressTestRepository_Constructor(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	repo := NewStressTestRepository(db)
	assert.NotNil(t, repo)
}

// ---------------------------------------------------------------------------
// CreateStressTest
// ---------------------------------------------------------------------------

func TestStressTestRepository_CreateStressTest(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		test    *models.StressTest
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			test: &models.StressTest{
				Name:            "Load Test",
				Description:     "API load test",
				Type:            "http",
				Status:          "pending",
				Scenarios:       []models.StressTestScenario{{Name: "get-users", URL: "/api/users", Method: "GET", Weight: 1}},
				Configuration:   models.StressTestConfig{Timeout: 30, RetryCount: 3, EnableMetrics: true},
				ConcurrentUsers: 100,
				DurationSeconds: 60,
				RampUpTime:      10,
				CreatedBy:       "admin",
				CreatedAt:       now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO stress_tests").
					WithArgs("Load Test", "API load test", "http", "pending",
						sqlmock.AnyArg(), sqlmock.AnyArg(),
						100, 60, 10, "admin", now).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "database error",
			test: &models.StressTest{
				Name:          "Test",
				Scenarios:     []models.StressTestScenario{},
				Configuration: models.StressTestConfig{},
				CreatedAt:     now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO stress_tests").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStressTestRepo(t)
			tt.setup(mock)

			err := repo.CreateStressTest(tt.test)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, int64(1), tt.test.ID)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetStressTest
// ---------------------------------------------------------------------------

func TestStressTestRepository_GetStressTest(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		id      int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, test *models.StressTest)
	}{
		{
			name: "success",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(stressTestColumns).
					AddRow(1, "Load Test", "API load test", "http", "completed",
						`[{"name":"get-users","url":"/api/users","method":"GET","weight":1}]`,
						`{"timeout":30,"retry_count":3,"enable_metrics":true,"enable_logging":false}`,
						100, 60, 10, "admin", now, nil, nil)
				mock.ExpectQuery("SELECT .+ FROM stress_tests WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			check: func(t *testing.T, test *models.StressTest) {
				assert.Equal(t, int64(1), test.ID)
				assert.Equal(t, "Load Test", test.Name)
				assert.Equal(t, "http", test.Type)
				assert.Equal(t, 100, test.ConcurrentUsers)
				assert.Nil(t, test.StartedAt)
				assert.Nil(t, test.CompletedAt)
				assert.Len(t, test.Scenarios, 1)
				assert.Equal(t, "get-users", test.Scenarios[0].Name)
			},
		},
		{
			name: "not found",
			id:   999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM stress_tests WHERE id").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStressTestRepo(t)
			tt.setup(mock)

			test, err := repo.GetStressTest(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, test)
			tt.check(t, test)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateStressTest
// ---------------------------------------------------------------------------

func TestStressTestRepository_UpdateStressTest(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		test    *models.StressTest
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			test: &models.StressTest{
				ID:              1,
				Name:            "Updated Test",
				Description:     "updated",
				Type:            "http",
				Status:          "completed",
				Scenarios:       []models.StressTestScenario{},
				Configuration:   models.StressTestConfig{Timeout: 60},
				ConcurrentUsers: 200,
				DurationSeconds: 120,
				RampUpTime:      20,
				StartedAt:       &now,
				CompletedAt:     &now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE stress_tests SET").
					WithArgs("Updated Test", "updated", "http", "completed",
						sqlmock.AnyArg(), sqlmock.AnyArg(),
						200, 120, 20, &now, &now, int64(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			test: &models.StressTest{
				ID:            1,
				Scenarios:     []models.StressTestScenario{},
				Configuration: models.StressTestConfig{},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE stress_tests SET").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStressTestRepo(t)
			tt.setup(mock)

			err := repo.UpdateStressTest(tt.test)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteStressTest (cascading delete)
// ---------------------------------------------------------------------------

func TestStressTestRepository_DeleteStressTest(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				// First deletes executions
				mock.ExpectExec("DELETE FROM stress_test_executions WHERE stress_test_id").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 3))
				// Then deletes the stress test
				mock.ExpectExec("DELETE FROM stress_tests WHERE id").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "error deleting executions",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM stress_test_executions WHERE stress_test_id").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
		{
			name: "error deleting test",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM stress_test_executions WHERE stress_test_id").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec("DELETE FROM stress_tests WHERE id").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStressTestRepo(t)
			tt.setup(mock)

			err := repo.DeleteStressTest(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetStressTestsByUser
// ---------------------------------------------------------------------------

func TestStressTestRepository_GetStressTestsByUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		userID  int
		limit   int
		offset  int
		setup   func(mock sqlmock.Sqlmock)
		want    int
		wantErr bool
	}{
		{
			name:   "returns tests",
			userID: 1,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(stressTestColumns).
					AddRow(1, "Test 1", "desc", "http", "completed",
						`[]`, `{"timeout":30,"retry_count":0,"enable_metrics":false,"enable_logging":false}`,
						50, 30, 5, "admin", now, nil, nil)
				mock.ExpectQuery("SELECT .+ FROM stress_tests").
					WithArgs(1, 10, 0).
					WillReturnRows(rows)
			},
			want: 1,
		},
		{
			name:   "empty result",
			userID: 2,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM stress_tests").
					WithArgs(2, 10, 0).
					WillReturnRows(sqlmock.NewRows(stressTestColumns))
			},
			want: 0,
		},
		{
			name:   "database error",
			userID: 1,
			limit:  10,
			offset: 0,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM stress_tests").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStressTestRepo(t)
			tt.setup(mock)

			tests, err := repo.GetStressTestsByUser(tt.userID, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, tests, tt.want)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// CreateExecution
// ---------------------------------------------------------------------------

func TestStressTestRepository_CreateExecution(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		exec    *models.StressTestExecution
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			exec: &models.StressTestExecution{
				StressTestID: 1,
				Status:       "running",
				StartedAt:    &now,
				Metrics:      map[string]interface{}{"total_requests": 100},
				Results:      models.StressTestExecutionResults{TotalRequests: 100},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO stress_test_executions").
					WithArgs(int64(1), "running", &now, sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "database error",
			exec: &models.StressTestExecution{
				StressTestID: 1,
				Status:       "running",
				StartedAt:    &now,
				Metrics:      map[string]interface{}{},
				Results:      models.StressTestExecutionResults{},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO stress_test_executions").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStressTestRepo(t)
			tt.setup(mock)

			err := repo.CreateExecution(tt.exec)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, int64(1), tt.exec.ID)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetExecution
// ---------------------------------------------------------------------------

func TestStressTestRepository_GetExecution(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		id      int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, exec *models.StressTestExecution)
	}{
		{
			name: "success",
			id:   1,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(stressTestExecutionColumns).
					AddRow(1, 1, "completed", now, now,
						`{"total_requests":100}`,
						`{"total_requests":100,"successful_requests":95,"failed_requests":5,"average_response_time":0.1,"min_response_time":0.01,"max_response_time":1.0,"error_rate":0.05,"throughput":50.0}`,
						nil)
				mock.ExpectQuery("SELECT .+ FROM stress_test_executions WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			check: func(t *testing.T, exec *models.StressTestExecution) {
				assert.Equal(t, int64(1), exec.ID)
				assert.Equal(t, "completed", exec.Status)
				assert.NotNil(t, exec.CompletedAt)
			},
		},
		{
			name: "not found",
			id:   999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM stress_test_executions WHERE id").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStressTestRepo(t)
			tt.setup(mock)

			exec, err := repo.GetExecution(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, exec)
			tt.check(t, exec)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetExecutionsByTestID
// ---------------------------------------------------------------------------

func TestStressTestRepository_GetExecutionsByTestID(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		testID  int
		setup   func(mock sqlmock.Sqlmock)
		want    int
		wantErr bool
	}{
		{
			name:   "returns executions",
			testID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(stressTestExecutionColumns).
					AddRow(1, 1, "completed", now, now,
						`{}`, `{"total_requests":0,"successful_requests":0,"failed_requests":0,"average_response_time":0,"min_response_time":0,"max_response_time":0,"error_rate":0,"throughput":0}`, nil).
					AddRow(2, 1, "running", now, nil,
						`{}`, `{"total_requests":0,"successful_requests":0,"failed_requests":0,"average_response_time":0,"min_response_time":0,"max_response_time":0,"error_rate":0,"throughput":0}`, nil)
				mock.ExpectQuery("SELECT .+ FROM stress_test_executions").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: 2,
		},
		{
			name:   "database error",
			testID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT .+ FROM stress_test_executions").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStressTestRepo(t)
			tt.setup(mock)

			execs, err := repo.GetExecutionsByTestID(tt.testID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, execs, tt.want)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// DeleteExecutionsByTestID
// ---------------------------------------------------------------------------

func TestStressTestRepository_DeleteExecutionsByTestID(t *testing.T) {
	tests := []struct {
		name    string
		testID  int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:   "success",
			testID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM stress_test_executions WHERE stress_test_id").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 5))
			},
		},
		{
			name:   "database error",
			testID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM stress_test_executions WHERE stress_test_id").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStressTestRepo(t)
			tt.setup(mock)

			err := repo.DeleteExecutionsByTestID(tt.testID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// CleanupOldExecutions
// ---------------------------------------------------------------------------

func TestStressTestRepository_CleanupOldExecutions(t *testing.T) {
	olderThan := time.Now().Add(-30 * 24 * time.Hour)

	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM stress_test_executions WHERE started_at").
					WithArgs(olderThan).
					WillReturnResult(sqlmock.NewResult(0, 10))
			},
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM stress_test_executions WHERE started_at").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStressTestRepo(t)
			tt.setup(mock)

			err := repo.CleanupOldExecutions(olderThan)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// CreateTest (wrapper)
// ---------------------------------------------------------------------------

func TestStressTestRepository_CreateTest(t *testing.T) {
	now := time.Now()

	repo, mock := newMockStressTestRepo(t)
	mock.ExpectExec("INSERT INTO stress_tests").
		WithArgs("Wrapper Test", "", "http", "pending",
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			50, 30, 5, "admin", now).
		WillReturnResult(sqlmock.NewResult(7, 1))

	id, err := repo.CreateTest(&models.StressTest{
		Name:            "Wrapper Test",
		Type:            "http",
		Status:          "pending",
		Scenarios:       []models.StressTestScenario{},
		Configuration:   models.StressTestConfig{},
		ConcurrentUsers: 50,
		DurationSeconds: 30,
		RampUpTime:      5,
		CreatedBy:       "admin",
		CreatedAt:       now,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(7), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// UpdateExecution
// ---------------------------------------------------------------------------

func TestStressTestRepository_UpdateExecution(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		exec    *models.StressTestExecution
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			exec: &models.StressTestExecution{
				ID:           1,
				Status:       "completed",
				CompletedAt:  &now,
				Metrics:      map[string]interface{}{"total_requests": 500},
				Results:      models.StressTestExecutionResults{TotalRequests: 500, SuccessfulRequests: 480},
				ErrorMessage: nil,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE stress_test_executions SET").
					WithArgs("completed", &now, sqlmock.AnyArg(), sqlmock.AnyArg(), nil, int64(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "database error",
			exec: &models.StressTestExecution{
				ID:      1,
				Metrics: map[string]interface{}{},
				Results: models.StressTestExecutionResults{},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE stress_test_executions SET").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockStressTestRepo(t)
			tt.setup(mock)

			err := repo.UpdateExecution(tt.exec)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
