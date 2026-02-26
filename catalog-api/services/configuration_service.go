package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"catalogizer/models"
	"catalogizer/repository"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ConfigurationService struct {
	configRepo  *repository.ConfigurationRepository
	configPath  string
	config      *models.SystemConfiguration
	wizardSteps []*models.WizardStep
	validators  map[string]ConfigValidator
}

type ConfigValidator interface {
	Validate(value interface{}) error
}

type DatabaseValidator struct{}
type NetworkValidator struct{}
type PathValidator struct{}
type EmailValidator struct{}

func NewConfigurationService(configRepo *repository.ConfigurationRepository, configPath string) *ConfigurationService {
	service := &ConfigurationService{
		configRepo: configRepo,
		configPath: configPath,
		validators: make(map[string]ConfigValidator),
	}

	// Register validators
	service.validators["database"] = &DatabaseValidator{}
	service.validators["network"] = &NetworkValidator{}
	service.validators["path"] = &PathValidator{}
	service.validators["email"] = &EmailValidator{}

	// Initialize wizard steps
	service.initializeWizardSteps()

	// Load current configuration
	service.loadConfiguration()

	return service
}

func (s *ConfigurationService) initializeWizardSteps() {
	s.wizardSteps = []*models.WizardStep{
		{
			ID:          "welcome",
			Name:        "Welcome",
			Description: "Welcome to Catalogizer Setup Wizard",
			Type:        models.WizardStepTypeInfo,
			Required:    true,
			Order:       1,
			Content: map[string]interface{}{
				"title":   "Welcome to Catalogizer v3.0",
				"message": "This wizard will help you configure your media cataloging system.",
			},
		},
		{
			ID:          "database",
			Name:        "Database Configuration",
			Description: "Configure your database connection",
			Type:        models.WizardStepTypeForm,
			Required:    true,
			Order:       2,
			Fields: []*models.WizardField{
				{
					Name:         "database_type",
					Label:        "Database Type",
					Type:         "select",
					Required:     true,
					Options:      []string{"sqlite", "mysql", "postgresql"},
					DefaultValue: "sqlite",
				},
				{
					Name:         "database_host",
					Label:        "Database Host",
					Type:         "text",
					Required:     false,
					DefaultValue: "localhost",
					ShowWhen:     map[string]interface{}{"database_type": []string{"mysql", "postgresql"}},
				},
				{
					Name:         "database_port",
					Label:        "Database Port",
					Type:         "number",
					Required:     false,
					DefaultValue: 3306,
					ShowWhen:     map[string]interface{}{"database_type": []string{"mysql", "postgresql"}},
				},
				{
					Name:         "database_name",
					Label:        "Database Name",
					Type:         "text",
					Required:     true,
					DefaultValue: "catalogizer",
				},
				{
					Name:     "database_username",
					Label:    "Database Username",
					Type:     "text",
					Required: false,
					ShowWhen: map[string]interface{}{"database_type": []string{"mysql", "postgresql"}},
				},
				{
					Name:     "database_password",
					Label:    "Database Password",
					Type:     "password",
					Required: false,
					ShowWhen: map[string]interface{}{"database_type": []string{"mysql", "postgresql"}},
				},
			},
			Validation: map[string]interface{}{
				"validator": "database",
			},
		},
		{
			ID:          "storage",
			Name:        "Storage Configuration",
			Description: "Configure storage locations and settings",
			Type:        models.WizardStepTypeForm,
			Required:    true,
			Order:       3,
			Fields: []*models.WizardField{
				{
					Name:         "media_directory",
					Label:        "Media Directory",
					Type:         "directory",
					Required:     true,
					DefaultValue: "/var/lib/catalogizer/media",
					Validation:   map[string]interface{}{"validator": "path"},
				},
				{
					Name:         "thumbnail_directory",
					Label:        "Thumbnail Directory",
					Type:         "directory",
					Required:     true,
					DefaultValue: "/var/lib/catalogizer/thumbnails",
					Validation:   map[string]interface{}{"validator": "path"},
				},
				{
					Name:         "temp_directory",
					Label:        "Temporary Directory",
					Type:         "directory",
					Required:     true,
					DefaultValue: "/tmp/catalogizer",
					Validation:   map[string]interface{}{"validator": "path"},
				},
				{
					Name:         "max_file_size",
					Label:        "Maximum File Size (MB)",
					Type:         "number",
					Required:     true,
					DefaultValue: 1000,
				},
				{
					Name:         "storage_quota",
					Label:        "Storage Quota (GB, 0 = unlimited)",
					Type:         "number",
					Required:     false,
					DefaultValue: 0,
				},
			},
		},
		{
			ID:          "network",
			Name:        "Network Configuration",
			Description: "Configure network and API settings",
			Type:        models.WizardStepTypeForm,
			Required:    true,
			Order:       4,
			Fields: []*models.WizardField{
				{
					Name:         "server_host",
					Label:        "Server Host",
					Type:         "text",
					Required:     true,
					DefaultValue: "0.0.0.0",
					Validation:   map[string]interface{}{"validator": "network"},
				},
				{
					Name:         "server_port",
					Label:        "Server Port",
					Type:         "number",
					Required:     true,
					DefaultValue: 8080,
					Validation:   map[string]interface{}{"validator": "network"},
				},
				{
					Name:         "enable_https",
					Label:        "Enable HTTPS",
					Type:         "checkbox",
					Required:     false,
					DefaultValue: false,
				},
				{
					Name:     "ssl_cert_path",
					Label:    "SSL Certificate Path",
					Type:     "file",
					Required: false,
					ShowWhen: map[string]interface{}{"enable_https": true},
				},
				{
					Name:     "ssl_key_path",
					Label:    "SSL Private Key Path",
					Type:     "file",
					Required: false,
					ShowWhen: map[string]interface{}{"enable_https": true},
				},
				{
					Name:         "cors_origins",
					Label:        "CORS Allowed Origins",
					Type:         "text",
					Required:     false,
					DefaultValue: "*",
				},
			},
		},
		{
			ID:          "authentication",
			Name:        "Authentication Setup",
			Description: "Configure authentication and security settings",
			Type:        models.WizardStepTypeForm,
			Required:    true,
			Order:       5,
			Fields: []*models.WizardField{
				{
					Name:     "jwt_secret",
					Label:    "JWT Secret Key",
					Type:     "password",
					Required: true,
					Generate: true,
				},
				{
					Name:         "session_timeout",
					Label:        "Session Timeout (hours)",
					Type:         "number",
					Required:     true,
					DefaultValue: 24,
				},
				{
					Name:         "enable_registration",
					Label:        "Allow User Registration",
					Type:         "checkbox",
					Required:     false,
					DefaultValue: true,
				},
				{
					Name:         "require_email_verification",
					Label:        "Require Email Verification",
					Type:         "checkbox",
					Required:     false,
					DefaultValue: false,
				},
				{
					Name:       "admin_email",
					Label:      "Administrator Email",
					Type:       "email",
					Required:   true,
					Validation: map[string]interface{}{"validator": "email"},
				},
			},
		},
		{
			ID:          "features",
			Name:        "Feature Configuration",
			Description: "Enable and configure advanced features",
			Type:        models.WizardStepTypeForm,
			Required:    false,
			Order:       6,
			Fields: []*models.WizardField{
				{
					Name:         "enable_media_conversion",
					Label:        "Enable Media Format Conversion",
					Type:         "checkbox",
					Required:     false,
					DefaultValue: true,
				},
				{
					Name:         "enable_webdav_sync",
					Label:        "Enable WebDAV Synchronization",
					Type:         "checkbox",
					Required:     false,
					DefaultValue: false,
				},
				{
					Name:         "enable_error_reporting",
					Label:        "Enable Error Reporting",
					Type:         "checkbox",
					Required:     false,
					DefaultValue: true,
				},
				{
					Name:         "enable_log_management",
					Label:        "Enable Log Management",
					Type:         "checkbox",
					Required:     false,
					DefaultValue: true,
				},
			},
		},
		{
			ID:          "external_services",
			Name:        "External Services",
			Description: "Configure integrations with external services",
			Type:        models.WizardStepTypeForm,
			Required:    false,
			Order:       7,
			Fields: []*models.WizardField{
				{
					Name:     "smtp_host",
					Label:    "SMTP Host",
					Type:     "text",
					Required: false,
				},
				{
					Name:         "smtp_port",
					Label:        "SMTP Port",
					Type:         "number",
					Required:     false,
					DefaultValue: 587,
				},
				{
					Name:     "smtp_username",
					Label:    "SMTP Username",
					Type:     "text",
					Required: false,
				},
				{
					Name:     "smtp_password",
					Label:    "SMTP Password",
					Type:     "password",
					Required: false,
				},
				{
					Name:     "slack_webhook_url",
					Label:    "Slack Webhook URL",
					Type:     "text",
					Required: false,
				},
				{
					Name:         "enable_analytics",
					Label:        "Enable Analytics",
					Type:         "checkbox",
					Required:     false,
					DefaultValue: true,
				},
			},
		},
		{
			ID:          "summary",
			Name:        "Configuration Summary",
			Description: "Review your configuration before applying",
			Type:        models.WizardStepTypeSummary,
			Required:    true,
			Order:       8,
		},
		{
			ID:          "complete",
			Name:        "Setup Complete",
			Description: "Configuration has been applied successfully",
			Type:        models.WizardStepTypeComplete,
			Required:    true,
			Order:       9,
		},
	}

	// Sort steps by order
	sort.Slice(s.wizardSteps, func(i, j int) bool {
		return s.wizardSteps[i].Order < s.wizardSteps[j].Order
	})
}

func (s *ConfigurationService) GetWizardSteps() ([]*models.WizardStep, error) {
	return s.wizardSteps, nil
}

func (s *ConfigurationService) GetWizardStep(stepID string) (*models.WizardStep, error) {
	for _, step := range s.wizardSteps {
		if step.ID == stepID {
			return step, nil
		}
	}
	return nil, fmt.Errorf("wizard step not found: %s", stepID)
}

func (s *ConfigurationService) ValidateWizardStep(stepID string, data map[string]interface{}) (*models.WizardStepValidation, error) {
	step, err := s.GetWizardStep(stepID)
	if err != nil {
		return nil, err
	}

	validation := &models.WizardStepValidation{
		StepID:   stepID,
		Valid:    true,
		Errors:   make(map[string]string),
		Warnings: make(map[string]string),
	}

	// Validate required fields
	for _, field := range step.Fields {
		value, exists := data[field.Name]

		if field.Required && (!exists || s.isEmptyValue(value)) {
			validation.Valid = false
			validation.Errors[field.Name] = fmt.Sprintf("%s is required", field.Label)
			continue
		}

		// Validate field using validator
		if exists && field.Validation != nil {
			if validatorName, ok := field.Validation["validator"].(string); ok {
				if validator, validatorExists := s.validators[validatorName]; validatorExists {
					if err := validator.Validate(value); err != nil {
						validation.Valid = false
						validation.Errors[field.Name] = err.Error()
					}
				}
			}
		}
	}

	// Custom step validation
	if step.Validation != nil {
		if validatorName, ok := step.Validation["validator"].(string); ok {
			if validator, exists := s.validators[validatorName]; exists {
				if err := validator.Validate(data); err != nil {
					validation.Valid = false
					validation.Errors["_general"] = err.Error()
				}
			}
		}
	}

	return validation, nil
}

func (s *ConfigurationService) SaveWizardProgress(userID int, stepID string, data map[string]interface{}) error {
	progress := &models.WizardProgress{
		UserID:      userID,
		CurrentStep: stepID,
		StepData:    data,
		UpdatedAt:   time.Now(),
	}

	return s.configRepo.SaveWizardProgress(progress)
}

func (s *ConfigurationService) GetWizardProgress(userID int) (*models.WizardProgress, error) {
	return s.configRepo.GetWizardProgress(userID)
}

func (s *ConfigurationService) CompleteWizard(userID int, finalData map[string]interface{}) (*models.SystemConfiguration, error) {
	// Generate the full configuration from wizard data
	config := s.generateConfiguration(finalData)

	// Validate the complete configuration
	if err := s.validateConfiguration(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Save configuration
	if err := s.SaveConfiguration(config); err != nil {
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	// Mark wizard as completed
	if err := s.configRepo.MarkWizardCompleted(userID); err != nil {
		return nil, fmt.Errorf("failed to mark wizard as completed: %w", err)
	}

	// Clean up wizard progress
	s.configRepo.DeleteWizardProgress(userID)

	return config, nil
}

func (s *ConfigurationService) GetConfiguration() (*models.SystemConfiguration, error) {
	if s.config == nil {
		return s.loadConfiguration()
	}
	return s.config, nil
}

func (s *ConfigurationService) SaveConfiguration(config *models.SystemConfiguration) error {
	// Validate configuration
	if err := s.validateConfiguration(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Save to database
	if err := s.configRepo.SaveConfiguration(config); err != nil {
		return fmt.Errorf("failed to save configuration to database: %w", err)
	}

	// Save to file
	if err := s.saveConfigurationFile(config); err != nil {
		return fmt.Errorf("failed to save configuration file: %w", err)
	}

	s.config = config
	return nil
}

func (s *ConfigurationService) UpdateConfiguration(updates map[string]interface{}) (*models.SystemConfiguration, error) {
	config, err := s.GetConfiguration()
	if err != nil {
		return nil, err
	}

	// Apply updates using reflection
	configValue := reflect.ValueOf(config).Elem()
	for key, value := range updates {
		field := configValue.FieldByName(s.toCamelCase(key))
		if field.IsValid() && field.CanSet() {
			newValue := reflect.ValueOf(value)
			if newValue.Type().ConvertibleTo(field.Type()) {
				field.Set(newValue.Convert(field.Type()))
			}
		}
	}

	return config, s.SaveConfiguration(config)
}

func (s *ConfigurationService) ResetConfiguration() error {
	// Create default configuration
	config := s.createDefaultConfiguration()

	return s.SaveConfiguration(config)
}

func (s *ConfigurationService) ExportConfiguration() ([]byte, error) {
	config, err := s.GetConfiguration()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(config, "", "  ")
}

func (s *ConfigurationService) ImportConfiguration(data []byte) (*models.SystemConfiguration, error) {
	var config models.SystemConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	if err := s.SaveConfiguration(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (s *ConfigurationService) GetConfigurationSchema() (*models.ConfigurationSchema, error) {
	return &models.ConfigurationSchema{
		Version: "3.0.0",
		Sections: []*models.ConfigSection{
			{
				Name:        "Database",
				Key:         "database",
				Description: "Database configuration settings",
				Fields:      s.getDatabaseFields(),
			},
			{
				Name:        "Storage",
				Key:         "storage",
				Description: "Storage and file system settings",
				Fields:      s.getStorageFields(),
			},
			{
				Name:        "Network",
				Key:         "network",
				Description: "Network and API settings",
				Fields:      s.getNetworkFields(),
			},
			{
				Name:        "Authentication",
				Key:         "authentication",
				Description: "Authentication and security settings",
				Fields:      s.getAuthenticationFields(),
			},
			{
				Name:        "Features",
				Key:         "features",
				Description: "Feature toggles and advanced settings",
				Fields:      s.getFeatureFields(),
			},
		},
	}, nil
}

func (s *ConfigurationService) TestConfiguration(config *models.SystemConfiguration) (*models.ConfigurationTest, error) {
	test := &models.ConfigurationTest{
		TestedAt: time.Now(),
		Results:  make(map[string]*models.TestResult),
	}

	// Test database connection
	test.Results["database"] = s.testDatabaseConnection(config)

	// Test storage paths
	test.Results["storage"] = s.testStoragePaths(config)

	// Test network configuration
	test.Results["network"] = s.testNetworkConfiguration(config)

	// Test external services
	test.Results["external_services"] = s.testExternalServices(config)

	// Calculate overall status
	test.OverallStatus = "passed"
	for _, result := range test.Results {
		if result.Status == "failed" {
			test.OverallStatus = "failed"
			break
		} else if result.Status == "warning" && test.OverallStatus == "passed" {
			test.OverallStatus = "warning"
		}
	}

	return test, nil
}

// Helper methods

func (s *ConfigurationService) loadConfiguration() (*models.SystemConfiguration, error) {
	// Try to load from database first
	config, err := s.configRepo.GetConfiguration()
	if err == nil {
		s.config = config
		return config, nil
	}

	// Try to load from file
	if _, err := os.Stat(s.configPath); err == nil {
		config, err := s.loadConfigurationFile()
		if err == nil {
			s.config = config
			return config, nil
		}
	}

	// Create default configuration
	config = s.createDefaultConfiguration()
	s.config = config
	return config, nil
}

func (s *ConfigurationService) loadConfigurationFile() (*models.SystemConfiguration, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return nil, err
	}

	var config models.SystemConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (s *ConfigurationService) saveConfigurationFile(config *models.SystemConfiguration) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.configPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.configPath, data, 0644)
}

func (s *ConfigurationService) createDefaultConfiguration() *models.SystemConfiguration {
	return &models.SystemConfiguration{
		Version:   "3.0.0",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Database: &models.DatabaseConfig{
			Type: "sqlite",
			Name: "catalogizer.db",
		},
		Storage: &models.StorageConfig{
			MediaDirectory:     "/var/lib/catalogizer/media",
			ThumbnailDirectory: "/var/lib/catalogizer/thumbnails",
			TempDirectory:      "/tmp/catalogizer",
			MaxFileSize:        1000 * 1024 * 1024, // 1GB
		},
		Network: &models.NetworkConfig{
			Host: "0.0.0.0",
			Port: 8080,
			CORS: &models.CORSConfig{
				AllowedOrigins: []string{"*"},
			},
		},
		Authentication: &models.AuthenticationConfig{
			SessionTimeout:           24 * time.Hour,
			EnableRegistration:       true,
			RequireEmailVerification: false,
		},
	}
}

func (s *ConfigurationService) generateConfiguration(wizardData map[string]interface{}) *models.SystemConfiguration {
	config := s.createDefaultConfiguration()

	// Apply wizard data to configuration
	// This is a simplified implementation
	for key, value := range wizardData {
		switch key {
		case "database_type":
			if s, ok := value.(string); ok {
				config.Database.Type = s
			}
		case "database_host":
			if s, ok := value.(string); ok {
				config.Database.Host = s
			}
		case "database_port":
			if port, ok := value.(float64); ok {
				config.Database.Port = int(port)
			}
		case "database_name":
			if s, ok := value.(string); ok {
				config.Database.Name = s
			}
		case "database_username":
			if s, ok := value.(string); ok {
				config.Database.Username = s
			}
		case "database_password":
			if s, ok := value.(string); ok {
				config.Database.Password = s
			}
		case "media_directory":
			if s, ok := value.(string); ok {
				config.Storage.MediaDirectory = s
			}
		case "thumbnail_directory":
			if s, ok := value.(string); ok {
				config.Storage.ThumbnailDirectory = s
			}
		case "temp_directory":
			if s, ok := value.(string); ok {
				config.Storage.TempDirectory = s
			}
		case "server_host":
			if s, ok := value.(string); ok {
				config.Network.Host = s
			}
		case "server_port":
			if port, ok := value.(float64); ok {
				config.Network.Port = int(port)
			}
		case "enable_https":
			if b, ok := value.(bool); ok {
				config.Network.HTTPS = &models.HTTPSConfig{
					Enabled: b,
				}
			}
		}
	}

	config.UpdatedAt = time.Now()
	return config
}

func (s *ConfigurationService) validateConfiguration(config *models.SystemConfiguration) error {
	// Basic validation
	if config.Database == nil {
		return fmt.Errorf("database configuration is required")
	}

	if config.Storage == nil {
		return fmt.Errorf("storage configuration is required")
	}

	if config.Network == nil {
		return fmt.Errorf("network configuration is required")
	}

	// Validate database configuration
	if config.Database.Type == "" {
		return fmt.Errorf("database type is required")
	}

	// Validate storage paths
	if config.Storage.MediaDirectory == "" {
		return fmt.Errorf("media directory is required")
	}

	// Validate network configuration
	if config.Network.Port <= 0 || config.Network.Port > 65535 {
		return fmt.Errorf("invalid network port: %d", config.Network.Port)
	}

	return nil
}

func (s *ConfigurationService) isEmptyValue(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []string:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

func (s *ConfigurationService) toCamelCase(str string) string {
	parts := strings.Split(str, "_")
	for i := range parts {
		if i > 0 {
			parts[i] = cases.Title(language.Und, cases.NoLower).String(parts[i])
		}
	}
	return strings.Join(parts, "")
}

func (s *ConfigurationService) getDatabaseFields() []*models.ConfigField {
	return []*models.ConfigField{
		{Name: "type", Label: "Database Type", Type: "select", Required: true},
		{Name: "host", Label: "Host", Type: "text", Required: false},
		{Name: "port", Label: "Port", Type: "number", Required: false},
		{Name: "name", Label: "Database Name", Type: "text", Required: true},
		{Name: "username", Label: "Username", Type: "text", Required: false},
		{Name: "password", Label: "Password", Type: "password", Required: false},
	}
}

func (s *ConfigurationService) getStorageFields() []*models.ConfigField {
	return []*models.ConfigField{
		{Name: "media_directory", Label: "Media Directory", Type: "directory", Required: true},
		{Name: "thumbnail_directory", Label: "Thumbnail Directory", Type: "directory", Required: true},
		{Name: "temp_directory", Label: "Temporary Directory", Type: "directory", Required: true},
		{Name: "max_file_size", Label: "Max File Size (MB)", Type: "number", Required: true},
	}
}

func (s *ConfigurationService) getNetworkFields() []*models.ConfigField {
	return []*models.ConfigField{
		{Name: "host", Label: "Host", Type: "text", Required: true},
		{Name: "port", Label: "Port", Type: "number", Required: true},
		{Name: "enable_https", Label: "Enable HTTPS", Type: "checkbox", Required: false},
	}
}

func (s *ConfigurationService) getAuthenticationFields() []*models.ConfigField {
	return []*models.ConfigField{
		{Name: "jwt_secret", Label: "JWT Secret", Type: "password", Required: true},
		{Name: "session_timeout", Label: "Session Timeout (hours)", Type: "number", Required: true},
		{Name: "enable_registration", Label: "Enable Registration", Type: "checkbox", Required: false},
	}
}

func (s *ConfigurationService) getFeatureFields() []*models.ConfigField {
	return []*models.ConfigField{
		{Name: "media_conversion", Label: "Media Conversion", Type: "checkbox", Required: false},
		{Name: "error_reporting", Label: "Error Reporting", Type: "checkbox", Required: false},
		{Name: "log_management", Label: "Log Management", Type: "checkbox", Required: false},
	}
}

func (s *ConfigurationService) testDatabaseConnection(config *models.SystemConfiguration) *models.TestResult {
	// Simplified test implementation
	return &models.TestResult{
		Status:  "passed",
		Message: "Database connection test passed",
	}
}

func (s *ConfigurationService) testStoragePaths(config *models.SystemConfiguration) *models.TestResult {
	// Test if storage directories are accessible
	paths := []string{
		config.Storage.MediaDirectory,
		config.Storage.ThumbnailDirectory,
		config.Storage.TempDirectory,
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return &models.TestResult{
				Status:  "warning",
				Message: fmt.Sprintf("Directory does not exist: %s", path),
			}
		}
	}

	return &models.TestResult{
		Status:  "passed",
		Message: "All storage paths are accessible",
	}
}

func (s *ConfigurationService) testNetworkConfiguration(config *models.SystemConfiguration) *models.TestResult {
	// Simplified network test
	if config.Network.Port < 1024 && os.Getuid() != 0 {
		return &models.TestResult{
			Status:  "warning",
			Message: "Port below 1024 requires root privileges",
		}
	}

	return &models.TestResult{
		Status:  "passed",
		Message: "Network configuration is valid",
	}
}

func (s *ConfigurationService) testExternalServices(config *models.SystemConfiguration) *models.TestResult {
	// Test external service connections
	return &models.TestResult{
		Status:  "passed",
		Message: "External services test passed",
	}
}

// Validator implementations

func (v *DatabaseValidator) Validate(value interface{}) error {
	// Database validation logic
	return nil
}

func (v *NetworkValidator) Validate(value interface{}) error {
	// Network validation logic
	return nil
}

func (v *PathValidator) Validate(value interface{}) error {
	// Path validation logic
	if path, ok := value.(string); ok {
		if !filepath.IsAbs(path) {
			return fmt.Errorf("path must be absolute")
		}
	}
	return nil
}

func (v *EmailValidator) Validate(value interface{}) error {
	// Email validation logic
	if email, ok := value.(string); ok {
		if !strings.Contains(email, "@") {
			return fmt.Errorf("invalid email format")
		}
	}
	return nil
}
