package services

import (
	"errors"
	"os"
	"testing"

	"catalogizer/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type errorValidator struct{}

func (e *errorValidator) Validate(value interface{}) error {
	return errors.New("step validation failed")
}

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
		{
			name: "database username and password",
			wizardData: map[string]interface{}{
				"database_type":     "mysql",
				"database_username": "myuser",
				"database_password": "mypassword",
			},
			checks: func(t *testing.T, config *models.SystemConfiguration) {
				assert.Equal(t, "mysql", config.Database.Type)
				assert.Equal(t, "myuser", config.Database.Username)
				assert.Equal(t, "mypassword", config.Database.Password)
			},
		},
		{
			name: "wrong type for string field - silently ignored",
			wizardData: map[string]interface{}{
				"database_type": 42, // int instead of string
				"database_name": "test.db",
			},
			checks: func(t *testing.T, config *models.SystemConfiguration) {
				// database_type should remain default (sqlite) because 42 is not a string
				assert.Equal(t, "sqlite", config.Database.Type)
				// database_name should be set because it's a string
				assert.Equal(t, "test.db", config.Database.Name)
			},
		},
		{
			name: "wrong type for bool field - silently ignored",
			wizardData: map[string]interface{}{
				"enable_https": "yes", // string instead of bool
			},
			checks: func(t *testing.T, config *models.SystemConfiguration) {
				// HTTPS should remain nil because "yes" is not a bool
				assert.Nil(t, config.Network.HTTPS)
			},
		},
		{
			name: "wrong type for port field - silently ignored",
			wizardData: map[string]interface{}{
				"database_port": "3306", // string instead of float64
			},
			checks: func(t *testing.T, config *models.SystemConfiguration) {
				// Port should remain default (0 for database port)
				assert.Equal(t, 0, config.Database.Port)
			},
		},
		{
			name: "wrong type for server port - silently ignored",
			wizardData: map[string]interface{}{
				"server_port": 8080, // int instead of float64 (JSON numbers are float64)
			},
			checks: func(t *testing.T, config *models.SystemConfiguration) {
				// Server port should be set because int can be converted to float64 in Go?
				// Actually, int value will be float64(8080) when passed as interface{} from JSON unmarshal
				// But in test, int is int, not float64, so type assertion will fail
				// So port should remain default (8080)
				assert.Equal(t, 8080, config.Network.Port) // default is 8080
			},
		},
		{
			name: "mixed correct and wrong types",
			wizardData: map[string]interface{}{
				"database_type":       "postgresql", // correct string
				"database_host":       12345,        // wrong type (int)
				"media_directory":     "/media",     // correct string
				"thumbnail_directory": 999,          // wrong type (int)
				"enable_https":        false,        // correct bool
			},
			checks: func(t *testing.T, config *models.SystemConfiguration) {
				assert.Equal(t, "postgresql", config.Database.Type)
				assert.Equal(t, "", config.Database.Host) // default (empty string), wrong type ignored
				assert.Equal(t, "/media", config.Storage.MediaDirectory)
				assert.Equal(t, "/var/lib/catalogizer/thumbnails", config.Storage.ThumbnailDirectory) // default, wrong type ignored
				require.NotNil(t, config.Network.HTTPS)
				assert.False(t, config.Network.HTTPS.Enabled)
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
			name:    "non-existent step",
			stepID:  "nonexistent",
			data:    map[string]interface{}{},
			wantErr: true,
		},
		{
			name:      "welcome step - always valid (no fields)",
			stepID:    "welcome",
			data:      map[string]interface{}{},
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
		// New test cases for uncovered branches
		{
			name:   "storage step - field validation with non-existent validator",
			stepID: "storage",
			data: map[string]interface{}{
				"media_directory":     "/valid/path",
				"thumbnail_directory": "/valid/path",
				"temp_directory":      "/valid/path",
				"max_file_size":       float64(1000),
			},
			wantValid: true, // validator exists (path validator is registered)
		},
		{
			name:   "authentication step - email validation passes",
			stepID: "authentication",
			data: map[string]interface{}{
				"jwt_secret":      "secret",
				"session_timeout": float64(24),
				"admin_email":     "admin@example.com",
			},
			wantValid: true,
		},
		{
			name:   "authentication step - email validation fails",
			stepID: "authentication",
			data: map[string]interface{}{
				"jwt_secret":      "secret",
				"session_timeout": float64(24),
				"admin_email":     "not-an-email",
			},
			wantValid: false,
		},
		{
			name:   "network step - field with validator but validator not string",
			stepID: "network",
			data: map[string]interface{}{
				"server_host": "0.0.0.0",
				"server_port": float64(8080),
			},
			wantValid: true, // network validator exists and passes
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

func TestConfigurationService_ValidateWizardStep_EdgeCases(t *testing.T) {
	// Test with empty validators map to trigger validator not found
	svc := &ConfigurationService{
		validators: make(map[string]ConfigValidator), // no validators registered
	}
	svc.initializeWizardSteps()

	// This should still pass because validator lookup fails but validatorExists is false
	// and the code continues without validation error (validator not found is ignored)
	validation, err := svc.ValidateWizardStep("storage", map[string]interface{}{
		"media_directory":     "/valid/path",
		"thumbnail_directory": "/valid/path",
		"temp_directory":      "/valid/path",
		"max_file_size":       float64(1000),
	})
	require.NoError(t, err)
	require.NotNil(t, validation)
	assert.True(t, validation.Valid, "missing validator should not cause validation failure")

	// Test with validator name not a string
	svc2 := &ConfigurationService{
		validators: make(map[string]ConfigValidator),
	}
	svc2.validators["path"] = &PathValidator{}
	svc2.initializeWizardSteps()
	// Manually modify a field's validation to have non-string validator name
	// This is a bit hacky but we need to test the type assertion branch
	for _, step := range svc2.wizardSteps {
		if step.ID == "storage" {
			for _, field := range step.Fields {
				if field.Name == "media_directory" {
					field.Validation = map[string]interface{}{
						"validator": 42, // non-string
					}
					break
				}
			}
			break
		}
	}
	validation, err = svc2.ValidateWizardStep("storage", map[string]interface{}{
		"media_directory":     "/valid/path",
		"thumbnail_directory": "/valid/path",
		"temp_directory":      "/valid/path",
		"max_file_size":       float64(1000),
	})
	require.NoError(t, err)
	require.NotNil(t, validation)
	// Should still be valid because validator name is not a string, so validation is skipped
	assert.True(t, validation.Valid)

	// Test custom step validation with non-existent validator
	svc3 := &ConfigurationService{
		validators: make(map[string]ConfigValidator),
	}
	svc3.validators["database"] = &DatabaseValidator{}
	svc3.initializeWizardSteps()
	// Database step has custom validation with "database" validator which exists
	validation, err = svc3.ValidateWizardStep("database", map[string]interface{}{
		"database_type": "sqlite",
		"database_name": "test.db",
	})
	require.NoError(t, err)
	require.NotNil(t, validation)
	assert.True(t, validation.Valid)
}

func TestConfigurationService_ValidateWizardStep_StepValidation(t *testing.T) {
	// Test step validation edge cases
	svc := &ConfigurationService{
		validators: make(map[string]ConfigValidator),
	}
	svc.validators["test"] = &PathValidator{} // any validator
	svc.initializeWizardSteps()

	// Find a step and modify its Validation for testing
	for _, step := range svc.wizardSteps {
		if step.ID == "database" {
			// Test 1: step.Validation exists but validator is not a string
			originalValidation := step.Validation
			step.Validation = map[string]interface{}{
				"validator": 42, // non-string
			}

			validation, err := svc.ValidateWizardStep("database", map[string]interface{}{
				"database_type": "sqlite",
				"database_name": "test.db",
			})
			require.NoError(t, err)
			require.NotNil(t, validation)
			assert.True(t, validation.Valid, "non-string validator should be skipped")

			// Test 2: step.Validation exists, validator is string but not in validators map
			step.Validation = map[string]interface{}{
				"validator": "nonexistent",
			}

			validation, err = svc.ValidateWizardStep("database", map[string]interface{}{
				"database_type": "sqlite",
				"database_name": "test.db",
			})
			require.NoError(t, err)
			require.NotNil(t, validation)
			assert.True(t, validation.Valid, "non-existent validator should be skipped")

			// Test 3: step.Validation exists, validator exists but validator.Validate returns error
			svc.validators["error"] = &errorValidator{}
			step.Validation = map[string]interface{}{
				"validator": "error",
			}
			validation, err = svc.ValidateWizardStep("database", map[string]interface{}{
				"database_type": "sqlite",
				"database_name": "test.db",
			})
			require.NoError(t, err)
			require.NotNil(t, validation)
			assert.False(t, validation.Valid, "validator error should make step invalid")
			assert.Contains(t, validation.Errors["_general"], "step validation failed")

			// Restore original validation
			step.Validation = originalValidation
			break
		}
	}
}

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

// Note: NewConfigurationService requires a valid repository to load configuration.
// Testing with nil repository would cause a panic. This test verifies the method exists.

func TestConfigurationService_GetDatabaseFields(t *testing.T) {
	svc := &ConfigurationService{}

	fields := svc.getDatabaseFields()
	assert.NotNil(t, fields)
	assert.NotEmpty(t, fields)
}

// ---------------------------------------------------------------------------
// GetStorageFields tests
// ---------------------------------------------------------------------------

func TestConfigurationService_GetStorageFields(t *testing.T) {
	svc := &ConfigurationService{}

	fields := svc.getStorageFields()
	assert.NotNil(t, fields)
	assert.NotEmpty(t, fields)
}

// ---------------------------------------------------------------------------
// GetNetworkFields tests
// ---------------------------------------------------------------------------

func TestConfigurationService_GetNetworkFields(t *testing.T) {
	svc := &ConfigurationService{}

	fields := svc.getNetworkFields()
	assert.NotNil(t, fields)
	assert.NotEmpty(t, fields)
}

// ---------------------------------------------------------------------------
// GetAuthenticationFields tests
// ---------------------------------------------------------------------------

func TestConfigurationService_GetAuthenticationFields(t *testing.T) {
	svc := &ConfigurationService{}

	fields := svc.getAuthenticationFields()
	assert.NotNil(t, fields)
	assert.NotEmpty(t, fields)
}

// ---------------------------------------------------------------------------
// GetFeatureFields tests
// ---------------------------------------------------------------------------

func TestConfigurationService_GetFeatureFields(t *testing.T) {
	svc := &ConfigurationService{}

	fields := svc.getFeatureFields()
	assert.NotNil(t, fields)
	assert.NotEmpty(t, fields)
}

func TestConfigurationService_TestStoragePaths(t *testing.T) {
	svc := &ConfigurationService{}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		mediaDir       string
		thumbnailDir   string
		tempDir        string
		expectedStatus string
		expectedMsg    string
	}{
		{
			name:           "all directories exist",
			mediaDir:       tempDir + "/media",
			thumbnailDir:   tempDir + "/thumbnails",
			tempDir:        tempDir + "/tmp",
			expectedStatus: "passed",
			expectedMsg:    "All storage paths are accessible",
		},
		{
			name:           "media directory missing",
			mediaDir:       tempDir + "/nonexistent/media",
			thumbnailDir:   tempDir + "/thumbnails",
			tempDir:        tempDir + "/tmp",
			expectedStatus: "warning",
			expectedMsg:    "Directory does not exist: " + tempDir + "/nonexistent/media",
		},
		{
			name:           "thumbnail directory missing",
			mediaDir:       tempDir + "/media",
			thumbnailDir:   tempDir + "/nonexistent/thumbnails",
			tempDir:        tempDir + "/tmp",
			expectedStatus: "warning",
			expectedMsg:    "Directory does not exist: " + tempDir + "/nonexistent/thumbnails",
		},
		{
			name:           "temp directory missing",
			mediaDir:       tempDir + "/media",
			thumbnailDir:   tempDir + "/thumbnails",
			tempDir:        tempDir + "/nonexistent/tmp",
			expectedStatus: "warning",
			expectedMsg:    "Directory does not exist: " + tempDir + "/nonexistent/tmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create directories that should exist
			if tt.expectedStatus == "passed" {
				os.MkdirAll(tt.mediaDir, 0755)
				os.MkdirAll(tt.thumbnailDir, 0755)
				os.MkdirAll(tt.tempDir, 0755)
			} else {
				// Create only the directories that exist in the test case
				// For simplicity, create parent tempDir
				os.MkdirAll(tempDir, 0755)
			}

			config := &models.SystemConfiguration{
				Storage: &models.StorageConfig{
					MediaDirectory:     tt.mediaDir,
					ThumbnailDirectory: tt.thumbnailDir,
					TempDirectory:      tt.tempDir,
				},
			}

			result := svc.testStoragePaths(config)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Contains(t, result.Message, tt.expectedMsg)
		})
	}
}

func TestConfigurationService_TestNetworkConfiguration(t *testing.T) {
	svc := &ConfigurationService{}

	tests := []struct {
		name           string
		port           int
		expectedStatus string
		expectedMsg    string
	}{
		{
			name:           "port above 1023 - normal user",
			port:           1024,
			expectedStatus: "passed",
			expectedMsg:    "Network configuration is valid",
		},
		{
			name:           "port 80 - normal user (warning)",
			port:           80,
			expectedStatus: "warning",
			expectedMsg:    "Port below 1024 requires root privileges",
		},
		{
			name:           "port 443 - normal user (warning)",
			port:           443,
			expectedStatus: "warning",
			expectedMsg:    "Port below 1024 requires root privileges",
		},
		{
			name:           "port 0 - warning (port < 1024, not root)",
			port:           0,
			expectedStatus: "warning", // Port 0 is < 1024 and we're not root
			expectedMsg:    "Port below 1024 requires root privileges",
		},
		{
			name:           "port 1023 - warning (port < 1024, not root)",
			port:           1023,
			expectedStatus: "warning",
			expectedMsg:    "Port below 1024 requires root privileges",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &models.SystemConfiguration{
				Network: &models.NetworkConfig{
					Port: tt.port,
				},
			}

			result := svc.testNetworkConfiguration(config)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Contains(t, result.Message, tt.expectedMsg)
		})
	}
}
