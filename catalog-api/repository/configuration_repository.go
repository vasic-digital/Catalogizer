package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"catalog-api/models"
)

type ConfigurationRepository struct {
	db *sql.DB
}

func NewConfigurationRepository(db *sql.DB) *ConfigurationRepository {
	return &ConfigurationRepository{db: db}
}

func (r *ConfigurationRepository) SaveConfiguration(config *models.SystemConfiguration) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	query := `
		INSERT OR REPLACE INTO system_configuration (
			id, version, configuration, created_at, updated_at
		) VALUES (1, ?, ?, ?, ?)`

	_, err = r.db.Exec(query, config.Version, string(configJSON), config.CreatedAt, config.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

func (r *ConfigurationRepository) GetConfiguration() (*models.SystemConfiguration, error) {
	query := `
		SELECT version, configuration, created_at, updated_at
		FROM system_configuration
		WHERE id = 1`

	var version, configJSON string
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(query).Scan(&version, &configJSON, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	var config models.SystemConfiguration
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	config.Version = version
	config.CreatedAt = createdAt
	config.UpdatedAt = updatedAt

	return &config, nil
}

func (r *ConfigurationRepository) SaveWizardProgress(progress *models.WizardProgress) error {
	stepDataJSON, err := json.Marshal(progress.StepData)
	if err != nil {
		return fmt.Errorf("failed to marshal step data: %w", err)
	}

	allDataJSON, err := json.Marshal(progress.AllData)
	if err != nil {
		return fmt.Errorf("failed to marshal all data: %w", err)
	}

	query := `
		INSERT OR REPLACE INTO wizard_progress (
			user_id, current_step, step_data, all_data, updated_at
		) VALUES (?, ?, ?, ?, ?)`

	_, err = r.db.Exec(query,
		progress.UserID, progress.CurrentStep, string(stepDataJSON),
		string(allDataJSON), progress.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to save wizard progress: %w", err)
	}

	return nil
}

func (r *ConfigurationRepository) GetWizardProgress(userID int) (*models.WizardProgress, error) {
	query := `
		SELECT user_id, current_step, step_data, all_data, updated_at
		FROM wizard_progress
		WHERE user_id = ?`

	var progress models.WizardProgress
	var stepDataJSON, allDataJSON string

	err := r.db.QueryRow(query, userID).Scan(
		&progress.UserID, &progress.CurrentStep, &stepDataJSON,
		&allDataJSON, &progress.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get wizard progress: %w", err)
	}

	if err := json.Unmarshal([]byte(stepDataJSON), &progress.StepData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal step data: %w", err)
	}

	if err := json.Unmarshal([]byte(allDataJSON), &progress.AllData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal all data: %w", err)
	}

	return &progress, nil
}

func (r *ConfigurationRepository) DeleteWizardProgress(userID int) error {
	query := "DELETE FROM wizard_progress WHERE user_id = ?"
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete wizard progress: %w", err)
	}
	return nil
}

func (r *ConfigurationRepository) MarkWizardCompleted(userID int) error {
	query := `
		INSERT OR REPLACE INTO wizard_completion (
			user_id, completed_at
		) VALUES (?, ?)`

	_, err := r.db.Exec(query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to mark wizard as completed: %w", err)
	}

	return nil
}

func (r *ConfigurationRepository) IsWizardCompleted(userID int) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM wizard_completion
		WHERE user_id = ?`

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check wizard completion: %w", err)
	}

	return count > 0, nil
}

func (r *ConfigurationRepository) GetConfigurationHistory(limit int) ([]*models.ConfigurationHistory, error) {
	query := `
		SELECT id, version, created_at, updated_at
		FROM system_configuration_history
		ORDER BY created_at DESC
		LIMIT ?`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration history: %w", err)
	}
	defer rows.Close()

	var history []*models.ConfigurationHistory
	for rows.Next() {
		var entry models.ConfigurationHistory
		err := rows.Scan(&entry.ID, &entry.Version, &entry.CreatedAt, &entry.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan configuration history: %w", err)
		}
		history = append(history, &entry)
	}

	return history, nil
}

func (r *ConfigurationRepository) SaveConfigurationHistory(config *models.SystemConfiguration) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	query := `
		INSERT INTO system_configuration_history (
			version, configuration, created_at, updated_at
		) VALUES (?, ?, ?, ?)`

	_, err = r.db.Exec(query, config.Version, string(configJSON), config.CreatedAt, config.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save configuration history: %w", err)
	}

	return nil
}

func (r *ConfigurationRepository) CreateConfigurationBackup(name string, config *models.SystemConfiguration) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	query := `
		INSERT INTO configuration_backups (
			name, version, configuration, created_at
		) VALUES (?, ?, ?, ?)`

	_, err = r.db.Exec(query, name, config.Version, string(configJSON), time.Now())
	if err != nil {
		return fmt.Errorf("failed to create configuration backup: %w", err)
	}

	return nil
}

func (r *ConfigurationRepository) GetConfigurationBackups() ([]*models.ConfigurationBackup, error) {
	query := `
		SELECT id, name, version, created_at
		FROM configuration_backups
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration backups: %w", err)
	}
	defer rows.Close()

	var backups []*models.ConfigurationBackup
	for rows.Next() {
		var backup models.ConfigurationBackup
		err := rows.Scan(&backup.ID, &backup.Name, &backup.Version, &backup.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan configuration backup: %w", err)
		}
		backups = append(backups, &backup)
	}

	return backups, nil
}

func (r *ConfigurationRepository) RestoreConfigurationBackup(backupID int) (*models.SystemConfiguration, error) {
	query := `
		SELECT configuration
		FROM configuration_backups
		WHERE id = ?`

	var configJSON string
	err := r.db.QueryRow(query, backupID).Scan(&configJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration backup: %w", err)
	}

	var config models.SystemConfiguration
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Update timestamps
	config.UpdatedAt = time.Now()

	// Save as current configuration
	if err := r.SaveConfiguration(&config); err != nil {
		return nil, fmt.Errorf("failed to restore configuration: %w", err)
	}

	return &config, nil
}

func (r *ConfigurationRepository) DeleteConfigurationBackup(backupID int) error {
	query := "DELETE FROM configuration_backups WHERE id = ?"
	_, err := r.db.Exec(query, backupID)
	if err != nil {
		return fmt.Errorf("failed to delete configuration backup: %w", err)
	}
	return nil
}

func (r *ConfigurationRepository) GetConfigurationTemplates() ([]*models.ConfigurationTemplate, error) {
	query := `
		SELECT id, name, description, category, configuration, created_at
		FROM configuration_templates
		ORDER BY category, name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.ConfigurationTemplate
	for rows.Next() {
		var template models.ConfigurationTemplate
		var configJSON string

		err := rows.Scan(
			&template.ID, &template.Name, &template.Description,
			&template.Category, &configJSON, &template.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan configuration template: %w", err)
		}

		if err := json.Unmarshal([]byte(configJSON), &template.Configuration); err != nil {
			return nil, fmt.Errorf("failed to unmarshal template configuration: %w", err)
		}

		templates = append(templates, &template)
	}

	return templates, nil
}

func (r *ConfigurationRepository) CreateConfigurationTemplate(template *models.ConfigurationTemplate) error {
	configJSON, err := json.Marshal(template.Configuration)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	query := `
		INSERT INTO configuration_templates (
			name, description, category, configuration, created_at
		) VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		template.Name, template.Description, template.Category,
		string(configJSON), template.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create configuration template: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	template.ID = int(id)
	return nil
}

func (r *ConfigurationRepository) DeleteConfigurationTemplate(templateID int) error {
	query := "DELETE FROM configuration_templates WHERE id = ?"
	_, err := r.db.Exec(query, templateID)
	if err != nil {
		return fmt.Errorf("failed to delete configuration template: %w", err)
	}
	return nil
}

func (r *ConfigurationRepository) ApplyConfigurationTemplate(templateID int) (*models.SystemConfiguration, error) {
	query := `
		SELECT configuration
		FROM configuration_templates
		WHERE id = ?`

	var configJSON string
	err := r.db.QueryRow(query, templateID).Scan(&configJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration template: %w", err)
	}

	var config models.SystemConfiguration
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Update timestamps
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	// Save as current configuration
	if err := r.SaveConfiguration(&config); err != nil {
		return nil, fmt.Errorf("failed to apply template: %w", err)
	}

	return &config, nil
}

func (r *ConfigurationRepository) CleanupOldHistory(olderThan time.Time) error {
	query := "DELETE FROM system_configuration_history WHERE created_at < ?"
	_, err := r.db.Exec(query, olderThan)
	if err != nil {
		return fmt.Errorf("failed to cleanup old configuration history: %w", err)
	}
	return nil
}

func (r *ConfigurationRepository) GetConfigurationStatistics() (*models.ConfigurationStatistics, error) {
	stats := &models.ConfigurationStatistics{}

	// Total configurations in history
	err := r.db.QueryRow("SELECT COUNT(*) FROM system_configuration_history").Scan(&stats.TotalConfigurations)
	if err != nil {
		return nil, fmt.Errorf("failed to get total configurations: %w", err)
	}

	// Total backups
	err = r.db.QueryRow("SELECT COUNT(*) FROM configuration_backups").Scan(&stats.TotalBackups)
	if err != nil {
		return nil, fmt.Errorf("failed to get total backups: %w", err)
	}

	// Total templates
	err = r.db.QueryRow("SELECT COUNT(*) FROM configuration_templates").Scan(&stats.TotalTemplates)
	if err != nil {
		return nil, fmt.Errorf("failed to get total templates: %w", err)
	}

	// Wizard completions
	err = r.db.QueryRow("SELECT COUNT(*) FROM wizard_completion").Scan(&stats.WizardCompletions)
	if err != nil {
		return nil, fmt.Errorf("failed to get wizard completions: %w", err)
	}

	// Last configuration update
	err = r.db.QueryRow("SELECT updated_at FROM system_configuration WHERE id = 1").Scan(&stats.LastUpdate)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get last update: %w", err)
	}

	return stats, nil
}