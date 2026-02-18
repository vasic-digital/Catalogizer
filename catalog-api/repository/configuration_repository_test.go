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

func newMockConfigRepo(t *testing.T) (*ConfigurationRepository, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	return NewConfigurationRepository(db), mock
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestConfigurationRepository_Constructor(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	repo := NewConfigurationRepository(db)
	assert.NotNil(t, repo)
}

// ---------------------------------------------------------------------------
// SaveConfiguration
// ---------------------------------------------------------------------------

func TestConfigurationRepository_SaveConfiguration(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		config  *models.SystemConfiguration
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			config: &models.SystemConfiguration{
				Version:   "1.0.0",
				CreatedAt: now,
				UpdatedAt: now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT OR REPLACE INTO system_configuration").
					WithArgs("1.0.0", sqlmock.AnyArg(), now, now).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "database error",
			config: &models.SystemConfiguration{
				Version:   "1.0.0",
				CreatedAt: now,
				UpdatedAt: now,
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT OR REPLACE INTO system_configuration").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockConfigRepo(t)
			tt.setup(mock)

			err := repo.SaveConfiguration(tt.config)
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
// GetConfiguration
// ---------------------------------------------------------------------------

func TestConfigurationRepository_GetConfiguration(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
		check   func(t *testing.T, config *models.SystemConfiguration)
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				configJSON := `{"version":"1.0.0"}`
				mock.ExpectQuery("SELECT version, configuration, created_at, updated_at").
					WillReturnRows(sqlmock.NewRows([]string{"version", "configuration", "created_at", "updated_at"}).
						AddRow("1.0.0", configJSON, now, now))
			},
			check: func(t *testing.T, config *models.SystemConfiguration) {
				assert.Equal(t, "1.0.0", config.Version)
			},
		},
		{
			name: "not found",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT version, configuration, created_at, updated_at").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockConfigRepo(t)
			tt.setup(mock)

			config, err := repo.GetConfiguration()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, config)
			tt.check(t, config)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// SaveWizardProgress / GetWizardProgress
// ---------------------------------------------------------------------------

func TestConfigurationRepository_WizardProgress(t *testing.T) {
	now := time.Now()

	t.Run("save wizard progress", func(t *testing.T) {
		repo, mock := newMockConfigRepo(t)
		progress := &models.WizardProgress{
			UserID:      1,
			CurrentStep: "database",
			StepData:    map[string]interface{}{"host": "localhost"},
			AllData:     map[string]interface{}{"step1": "done"},
			UpdatedAt:   now,
		}

		mock.ExpectExec("INSERT OR REPLACE INTO wizard_progress").
			WithArgs(1, "database", sqlmock.AnyArg(), sqlmock.AnyArg(), now).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.SaveWizardProgress(progress)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get wizard progress", func(t *testing.T) {
		repo, mock := newMockConfigRepo(t)

		mock.ExpectQuery("SELECT user_id, current_step, step_data, all_data, updated_at").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_step", "step_data", "all_data", "updated_at"}).
				AddRow(1, "database", `{"host":"localhost"}`, `{"step1":"done"}`, now))

		progress, err := repo.GetWizardProgress(1)
		require.NoError(t, err)
		assert.Equal(t, 1, progress.UserID)
		assert.Equal(t, "database", progress.CurrentStep)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get wizard progress not found", func(t *testing.T) {
		repo, mock := newMockConfigRepo(t)

		mock.ExpectQuery("SELECT user_id, current_step, step_data, all_data, updated_at").
			WithArgs(999).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetWizardProgress(999)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ---------------------------------------------------------------------------
// DeleteWizardProgress
// ---------------------------------------------------------------------------

func TestConfigurationRepository_DeleteWizardProgress(t *testing.T) {
	tests := []struct {
		name    string
		userID  int
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:   "success",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM wizard_progress WHERE user_id").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:   "database error",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM wizard_progress WHERE user_id").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockConfigRepo(t)
			tt.setup(mock)

			err := repo.DeleteWizardProgress(tt.userID)
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
// MarkWizardCompleted / IsWizardCompleted
// ---------------------------------------------------------------------------

func TestConfigurationRepository_WizardCompletion(t *testing.T) {
	t.Run("mark wizard completed", func(t *testing.T) {
		repo, mock := newMockConfigRepo(t)
		mock.ExpectExec("INSERT OR REPLACE INTO wizard_completion").
			WithArgs(1, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.MarkWizardCompleted(1)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("is wizard completed - true", func(t *testing.T) {
		repo, mock := newMockConfigRepo(t)
		mock.ExpectQuery("SELECT COUNT").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		completed, err := repo.IsWizardCompleted(1)
		require.NoError(t, err)
		assert.True(t, completed)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("is wizard completed - false", func(t *testing.T) {
		repo, mock := newMockConfigRepo(t)
		mock.ExpectQuery("SELECT COUNT").
			WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		completed, err := repo.IsWizardCompleted(2)
		require.NoError(t, err)
		assert.False(t, completed)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ---------------------------------------------------------------------------
// GetConfigurationHistory
// ---------------------------------------------------------------------------

func TestConfigurationRepository_GetConfigurationHistory(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		limit   int
		setup   func(mock sqlmock.Sqlmock)
		want    int
		wantErr bool
	}{
		{
			name:  "returns history",
			limit: 10,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "version", "created_at", "updated_at"}).
					AddRow(1, "1.0.0", now, now).
					AddRow(2, "1.1.0", now, now)
				mock.ExpectQuery("SELECT id, version, created_at, updated_at").
					WithArgs(10).
					WillReturnRows(rows)
			},
			want: 2,
		},
		{
			name:  "empty history",
			limit: 10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, version, created_at, updated_at").
					WithArgs(10).
					WillReturnRows(sqlmock.NewRows([]string{"id", "version", "created_at", "updated_at"}))
			},
			want: 0,
		},
		{
			name:  "database error",
			limit: 10,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, version, created_at, updated_at").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockConfigRepo(t)
			tt.setup(mock)

			history, err := repo.GetConfigurationHistory(tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, history, tt.want)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// ---------------------------------------------------------------------------
// GetConfigurationBackups
// ---------------------------------------------------------------------------

func TestConfigurationRepository_GetConfigurationBackups(t *testing.T) {
	now := time.Now()

	repo, mock := newMockConfigRepo(t)
	rows := sqlmock.NewRows([]string{"id", "name", "version", "created_at"}).
		AddRow(1, "backup-1", "1.0.0", now).
		AddRow(2, "backup-2", "1.1.0", now)
	mock.ExpectQuery("SELECT id, name, version, created_at").
		WillReturnRows(rows)

	backups, err := repo.GetConfigurationBackups()
	require.NoError(t, err)
	assert.Len(t, backups, 2)
	assert.Equal(t, "backup-1", backups[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// DeleteConfigurationBackup
// ---------------------------------------------------------------------------

func TestConfigurationRepository_DeleteConfigurationBackup(t *testing.T) {
	tests := []struct {
		name     string
		backupID int
		setup    func(mock sqlmock.Sqlmock)
		wantErr  bool
	}{
		{
			name:     "success",
			backupID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM configuration_backups WHERE id").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:     "database error",
			backupID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM configuration_backups WHERE id").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockConfigRepo(t)
			tt.setup(mock)

			err := repo.DeleteConfigurationBackup(tt.backupID)
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
// CleanupOldHistory
// ---------------------------------------------------------------------------

func TestConfigurationRepository_CleanupOldHistory(t *testing.T) {
	olderThan := time.Now().Add(-30 * 24 * time.Hour)

	repo, mock := newMockConfigRepo(t)
	mock.ExpectExec("DELETE FROM system_configuration_history WHERE created_at").
		WithArgs(olderThan).
		WillReturnResult(sqlmock.NewResult(0, 5))

	err := repo.CleanupOldHistory(olderThan)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// GetConfigurationStatistics
// ---------------------------------------------------------------------------

func TestConfigurationRepository_GetConfigurationStatistics(t *testing.T) {
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		repo, mock := newMockConfigRepo(t)

		mock.ExpectQuery("SELECT COUNT.+ FROM system_configuration_history").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
		mock.ExpectQuery("SELECT COUNT.+ FROM configuration_backups").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
		mock.ExpectQuery("SELECT COUNT.+ FROM configuration_templates").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
		mock.ExpectQuery("SELECT COUNT.+ FROM wizard_completion").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
		mock.ExpectQuery("SELECT updated_at FROM system_configuration").
			WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(now))

		stats, err := repo.GetConfigurationStatistics()
		require.NoError(t, err)
		assert.Equal(t, 10, stats.TotalConfigurations)
		assert.Equal(t, 3, stats.TotalBackups)
		assert.Equal(t, 5, stats.TotalTemplates)
		assert.Equal(t, 2, stats.WizardCompletions)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on total configurations", func(t *testing.T) {
		repo, mock := newMockConfigRepo(t)

		mock.ExpectQuery("SELECT COUNT.+ FROM system_configuration_history").
			WillReturnError(sql.ErrConnDone)

		_, err := repo.GetConfigurationStatistics()
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ---------------------------------------------------------------------------
// DeleteConfigurationTemplate
// ---------------------------------------------------------------------------

func TestConfigurationRepository_DeleteConfigurationTemplate(t *testing.T) {
	repo, mock := newMockConfigRepo(t)
	mock.ExpectExec("DELETE FROM configuration_templates WHERE id").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteConfigurationTemplate(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
