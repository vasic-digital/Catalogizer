package services

import (
	"testing"

	"catalogizer/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurationService_IsEmptyValue(t *testing.T) {
	svc := &ConfigurationService{}

	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{
			name:     "nil value",
			value:    nil,
			expected: true,
		},
		{
			name:     "empty string",
			value:    "",
			expected: true,
		},
		{
			name:     "whitespace-only string",
			value:    "   ",
			expected: true,
		},
		{
			name:     "non-empty string",
			value:    "hello",
			expected: false,
		},
		{
			name:     "empty string slice",
			value:    []string{},
			expected: true,
		},
		{
			name:     "non-empty string slice",
			value:    []string{"a"},
			expected: false,
		},
		{
			name:     "empty map",
			value:    map[string]interface{}{},
			expected: true,
		},
		{
			name:     "non-empty map",
			value:    map[string]interface{}{"key": "val"},
			expected: false,
		},
		{
			name:     "integer value (not empty)",
			value:    42,
			expected: false,
		},
		{
			name:     "boolean false (not empty)",
			value:    false,
			expected: false,
		},
		{
			name:     "boolean true (not empty)",
			value:    true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.isEmptyValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigurationService_ToCamelCase(t *testing.T) {
	svc := &ConfigurationService{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single word",
			input:    "database",
			expected: "database",
		},
		{
			name:     "two words",
			input:    "database_type",
			expected: "databaseType",
		},
		{
			name:     "three words",
			input:    "max_file_size",
			expected: "maxFileSize",
		},
		{
			name:     "already camelCase",
			input:    "myField",
			expected: "myField",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.toCamelCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigurationService_CreateDefaultConfiguration(t *testing.T) {
	svc := &ConfigurationService{}

	config := svc.createDefaultConfiguration()

	require.NotNil(t, config)
	assert.Equal(t, "3.0.0", config.Version)
	assert.NotZero(t, config.CreatedAt)
	assert.NotZero(t, config.UpdatedAt)

	// Database defaults
	require.NotNil(t, config.Database)
	assert.Equal(t, "sqlite", config.Database.Type)
	assert.Equal(t, "catalogizer.db", config.Database.Name)

	// Storage defaults
	require.NotNil(t, config.Storage)
	assert.Equal(t, "/var/lib/catalogizer/media", config.Storage.MediaDirectory)
	assert.Equal(t, "/var/lib/catalogizer/thumbnails", config.Storage.ThumbnailDirectory)
	assert.Equal(t, "/tmp/catalogizer", config.Storage.TempDirectory)
	assert.Equal(t, int64(1000*1024*1024), config.Storage.MaxFileSize)

	// Network defaults
	require.NotNil(t, config.Network)
	assert.Equal(t, "0.0.0.0", config.Network.Host)
	assert.Equal(t, 8080, config.Network.Port)
	require.NotNil(t, config.Network.CORS)
	assert.Equal(t, []string{"*"}, config.Network.CORS.AllowedOrigins)

	// Authentication defaults
	require.NotNil(t, config.Authentication)
	assert.True(t, config.Authentication.EnableRegistration)
	assert.False(t, config.Authentication.RequireEmailVerification)

	// Feature defaults
	require.NotNil(t, config.Features)
	assert.True(t, config.Features.MediaConversion)
	assert.True(t, config.Features.ErrorReporting)
	assert.True(t, config.Features.LogManagement)
}

func TestConfigurationService_ValidateConfiguration(t *testing.T) {
	svc := &ConfigurationService{}

	tests := []struct {
		name    string
		config  *models.SystemConfiguration
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid configuration",
			config:  svc.createDefaultConfiguration(),
			wantErr: false,
		},
		{
			name: "nil database",
			config: &models.SystemConfiguration{
				Database: nil,
				Storage:  &models.StorageConfig{MediaDirectory: "/media"},
				Network:  &models.NetworkConfig{Port: 8080},
			},
			wantErr: true,
			errMsg:  "database configuration is required",
		},
		{
			name: "nil storage",
			config: &models.SystemConfiguration{
				Database: &models.DatabaseConfig{Type: "sqlite"},
				Storage:  nil,
				Network:  &models.NetworkConfig{Port: 8080},
			},
			wantErr: true,
			errMsg:  "storage configuration is required",
		},
		{
			name: "nil network",
			config: &models.SystemConfiguration{
				Database: &models.DatabaseConfig{Type: "sqlite"},
				Storage:  &models.StorageConfig{MediaDirectory: "/media"},
				Network:  nil,
			},
			wantErr: true,
			errMsg:  "network configuration is required",
		},
		{
			name: "empty database type",
			config: &models.SystemConfiguration{
				Database: &models.DatabaseConfig{Type: ""},
				Storage:  &models.StorageConfig{MediaDirectory: "/media"},
				Network:  &models.NetworkConfig{Port: 8080},
			},
			wantErr: true,
			errMsg:  "database type is required",
		},
		{
			name: "empty media directory",
			config: &models.SystemConfiguration{
				Database: &models.DatabaseConfig{Type: "sqlite"},
				Storage:  &models.StorageConfig{MediaDirectory: ""},
				Network:  &models.NetworkConfig{Port: 8080},
			},
			wantErr: true,
			errMsg:  "media directory is required",
		},
		{
			name: "invalid port - zero",
			config: &models.SystemConfiguration{
				Database: &models.DatabaseConfig{Type: "sqlite"},
				Storage:  &models.StorageConfig{MediaDirectory: "/media"},
				Network:  &models.NetworkConfig{Port: 0},
			},
			wantErr: true,
			errMsg:  "invalid network port: 0",
		},
		{
			name: "invalid port - negative",
			config: &models.SystemConfiguration{
				Database: &models.DatabaseConfig{Type: "sqlite"},
				Storage:  &models.StorageConfig{MediaDirectory: "/media"},
				Network:  &models.NetworkConfig{Port: -1},
			},
			wantErr: true,
			errMsg:  "invalid network port: -1",
		},
		{
			name: "invalid port - too high",
			config: &models.SystemConfiguration{
				Database: &models.DatabaseConfig{Type: "sqlite"},
				Storage:  &models.StorageConfig{MediaDirectory: "/media"},
				Network:  &models.NetworkConfig{Port: 70000},
			},
			wantErr: true,
			errMsg:  "invalid network port: 70000",
		},
		{
			name: "valid port - boundary 1",
			config: &models.SystemConfiguration{
				Database: &models.DatabaseConfig{Type: "sqlite"},
				Storage:  &models.StorageConfig{MediaDirectory: "/media"},
				Network:  &models.NetworkConfig{Port: 1},
			},
			wantErr: false,
		},
		{
			name: "valid port - boundary 65535",
			config: &models.SystemConfiguration{
				Database: &models.DatabaseConfig{Type: "sqlite"},
				Storage:  &models.StorageConfig{MediaDirectory: "/media"},
				Network:  &models.NetworkConfig{Port: 65535},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateConfiguration(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigurationService_GenerateConfiguration(t *testing.T) {
	svc := &ConfigurationService{}

	tests := []struct {
		name       string
		wizardData map[string]interface{}
		checks     func(t *testing.T, config *models.SystemConfiguration)
	}{
		{
			name:       "empty wizard data returns defaults",
			wizardData: map[string]interface{}{},
			checks: func(t *testing.T, config *models.SystemConfiguration) {
				assert.Equal(t, "sqlite", config.Database.Type)
				assert.Equal(t, "catalogizer.db", config.Database.Name)
			},
		},
		{
			name: "database type override",
			wizardData: map[string]interface{}{
				"database_type": "postgresql",
				"database_name": "mydb",
			},
			checks: func(t *testing.T, config *models.SystemConfiguration) {
				assert.Equal(t, "postgresql", config.Database.Type)
				assert.Equal(t, "mydb", config.Database.Name)
			},
		},
		{
			name: "database host and port",
			wizardData: map[string]interface{}{
				"database_type": "mysql",
				"database_host": "db.example.com",
				"database_port": float64(3306),
				"database_name": "catalogizer",
			},
			checks: func(t *testing.T, config *models.SystemConfiguration) {
				assert.Equal(t, "mysql", config.Database.Type)
				assert.Equal(t, "db.example.com", config.Database.Host)
				assert.Equal(t, 3306, config.Database.Port)
			},
		},
		{
			name: "storage directories",
			wizardData: map[string]interface{}{
				"media_directory":     "/custom/media",
				"thumbnail_directory": "/custom/thumbnails",
				"temp_directory":      "/custom/tmp",
			},
			checks: func(t *testing.T, config *models.SystemConfiguration) {
				assert.Equal(t, "/custom/media", config.Storage.MediaDirectory)
				assert.Equal(t, "/custom/thumbnails", config.Storage.ThumbnailDirectory)
				assert.Equal(t, "/custom/tmp", config.Storage.TempDirectory)
			},
		},
		{
			name: "network configuration",
			wizardData: map[string]interface{}{
				"server_host": "192.168.1.100",
				"server_port": float64(9090),
			},
			checks: func(t *testing.T, config *models.SystemConfiguration) {
				assert.Equal(t, "192.168.1.100", config.Network.Host)
				assert.Equal(t, 9090, config.Network.Port)
			},
		},
		{
			name: "enable HTTPS",
			wizardData: map[string]interface{}{
				"enable_https": true,
			},
			checks: func(t *testing.T, config *models.SystemConfiguration) {
				require.NotNil(t, config.Network.HTTPS)
				assert.True(t, config.Network.HTTPS.Enabled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := svc.generateConfiguration(tt.wizardData)
			require.NotNil(t, config)
			assert.Equal(t, "3.0.0", config.Version)
			tt.checks(t, config)
		})
	}
}

func TestConfigurationService_GetWizardStep(t *testing.T) {
	svc := &ConfigurationService{
		validators: make(map[string]ConfigValidator),
	}
	svc.initializeWizardSteps()

	tests := []struct {
		name    string
		stepID  string
		wantErr bool
	}{
		{
			name:    "welcome step exists",
			stepID:  "welcome",
			wantErr: false,
		},
		{
			name:    "database step exists",
			stepID:  "database",
			wantErr: false,
		},
		{
			name:    "storage step exists",
			stepID:  "storage",
			wantErr: false,
		},
		{
			name:    "summary step exists",
			stepID:  "summary",
			wantErr: false,
		},
		{
			name:    "complete step exists",
			stepID:  "complete",
			wantErr: false,
		},
		{
			name:    "non-existent step",
			stepID:  "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step, err := svc.GetWizardStep(tt.stepID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, step)
				assert.Contains(t, err.Error(), "wizard step not found")
			} else {
				require.NoError(t, err)
				require.NotNil(t, step)
				assert.Equal(t, tt.stepID, step.ID)
			}
		})
	}
}

func TestConfigurationService_GetWizardSteps(t *testing.T) {
	svc := &ConfigurationService{
		validators: make(map[string]ConfigValidator),
	}
	svc.initializeWizardSteps()

	steps, err := svc.GetWizardSteps()
	require.NoError(t, err)
	require.NotNil(t, steps)
	assert.Greater(t, len(steps), 0)

	// Verify steps are sorted by order
	for i := 1; i < len(steps); i++ {
		assert.LessOrEqual(t, steps[i-1].Order, steps[i].Order,
			"steps should be sorted by order: step %s (order %d) should come before %s (order %d)",
			steps[i-1].ID, steps[i-1].Order, steps[i].ID, steps[i].Order)
	}

	// Verify required steps exist
	stepIDs := make(map[string]bool)
	for _, step := range steps {
		stepIDs[step.ID] = true
	}
	assert.True(t, stepIDs["welcome"])
	assert.True(t, stepIDs["database"])
	assert.True(t, stepIDs["storage"])
	assert.True(t, stepIDs["network"])
	assert.True(t, stepIDs["authentication"])
	assert.True(t, stepIDs["summary"])
	assert.True(t, stepIDs["complete"])
}

func TestConfigurationService_GetConfigurationSchema(t *testing.T) {
	svc := &ConfigurationService{}

	schema, err := svc.GetConfigurationSchema()
	require.NoError(t, err)
	require.NotNil(t, schema)
	assert.Equal(t, "3.0.0", schema.Version)

	// Verify sections exist
	require.GreaterOrEqual(t, len(schema.Sections), 5)

	sectionKeys := make(map[string]bool)
	for _, section := range schema.Sections {
		sectionKeys[section.Key] = true
		assert.NotEmpty(t, section.Name)
		assert.NotEmpty(t, section.Description)
		assert.NotNil(t, section.Fields)
	}

	assert.True(t, sectionKeys["database"])
	assert.True(t, sectionKeys["storage"])
	assert.True(t, sectionKeys["network"])
	assert.True(t, sectionKeys["authentication"])
	assert.True(t, sectionKeys["features"])
}

func TestConfigurationService_ExportConfiguration(t *testing.T) {
	svc := &ConfigurationService{
		config: nil,
	}

	// Set a default config so GetConfiguration works
	defaultConfig := svc.createDefaultConfiguration()
	svc.config = defaultConfig

	data, err := svc.ExportConfiguration()
	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Greater(t, len(data), 0)

	// Should be valid JSON
	assert.Contains(t, string(data), `"version"`)
	assert.Contains(t, string(data), `"database"`)
	assert.Contains(t, string(data), `"storage"`)
	assert.Contains(t, string(data), `"network"`)
}

func TestConfigurationService_ImportConfiguration(t *testing.T) {
	// ImportConfiguration calls SaveConfiguration which needs configRepo and file system.
	// We test only the parse failure case here.
	svc := &ConfigurationService{}

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "invalid JSON",
			data:    []byte("not json"),
			wantErr: true,
		},
		{
			name:    "empty JSON",
			data:    []byte("{}"),
			wantErr: true, // will fail validation since Database is nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ImportConfiguration(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

func TestPathValidator_Validate(t *testing.T) {
	validator := &PathValidator{}

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "absolute path",
			value:   "/var/lib/data",
			wantErr: false,
		},
		{
			name:    "relative path",
			value:   "relative/path",
			wantErr: true,
		},
		{
			name:    "non-string value",
			value:   42,
			wantErr: false, // PathValidator only checks strings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "path must be absolute")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailValidator_Validate(t *testing.T) {
	validator := &EmailValidator{}

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "valid email",
			value:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "invalid email - no at sign",
			value:   "userexample.com",
			wantErr: true,
		},
		{
			name:    "non-string value",
			value:   42,
			wantErr: false, // EmailValidator only checks strings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid email format")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDatabaseValidator_Validate(t *testing.T) {
	validator := &DatabaseValidator{}

	// DatabaseValidator currently always returns nil
	err := validator.Validate("anything")
	assert.NoError(t, err)

	err = validator.Validate(nil)
	assert.NoError(t, err)
}

func TestNetworkValidator_Validate(t *testing.T) {
	validator := &NetworkValidator{}

	// NetworkValidator currently always returns nil
	err := validator.Validate("anything")
	assert.NoError(t, err)

	err = validator.Validate(nil)
	assert.NoError(t, err)
}

func TestConfigurationService_TestConfiguration(t *testing.T) {
	svc := &ConfigurationService{}

	config := svc.createDefaultConfiguration()
	testResult, err := svc.TestConfiguration(config)

	require.NoError(t, err)
	require.NotNil(t, testResult)
	assert.NotZero(t, testResult.TestedAt)
	assert.Contains(t, testResult.Results, "database")
	assert.Contains(t, testResult.Results, "storage")
	assert.Contains(t, testResult.Results, "network")
	assert.Contains(t, testResult.Results, "external_services")

	// Database test should pass (simplified implementation)
	assert.Equal(t, "passed", testResult.Results["database"].Status)

	// External services test should pass (simplified implementation)
	assert.Equal(t, "passed", testResult.Results["external_services"].Status)
}

func TestConfigurationService_ValidateWizardStep(t *testing.T) {
	svc := &ConfigurationService{
		validators: make(map[string]ConfigValidator),
	}
	svc.validators["path"] = &PathValidator{}
	svc.validators["email"] = &EmailValidator{}
	svc.validators["database"] = &DatabaseValidator{}
	svc.validators["network"] = &NetworkValidator{}
	svc.initializeWizardSteps()

	tests := []struct {
		name      string
		stepID    string
		data      map[string]interface{}
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "non-existent step",
			stepID:    "nonexistent",
			data:      map[string]interface{}{},
			wantErr:   true,
		},
		{
			name:   "welcome step - always valid (no fields)",
			stepID: "welcome",
			data:   map[string]interface{}{},
			wantValid: true,
		},
		{
			name:   "database step - valid data",
			stepID: "database",
			data: map[string]interface{}{
				"database_type": "sqlite",
				"database_name": "test.db",
			},
			wantValid: true,
		},
		{
			name:   "database step - missing required field",
			stepID: "database",
			data: map[string]interface{}{
				"database_name": "test.db",
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validation, err := svc.ValidateWizardStep(tt.stepID, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, validation)
			assert.Equal(t, tt.wantValid, validation.Valid)
		})
	}
}
