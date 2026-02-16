package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigurationWizardService(t *testing.T) {
	service := NewConfigurationWizardService(nil)

	assert.NotNil(t, service)
}

func TestConfigurationWizardService_GetAvailableTemplates(t *testing.T) {
	service := NewConfigurationWizardService(nil)

	templates := service.GetAvailableTemplates()

	// Should return built-in templates even with nil repo
	assert.NotNil(t, templates)
}

func TestConfigurationWizardService_GetWizardSteps(t *testing.T) {
	service := NewConfigurationWizardService(nil)

	steps := service.GetWizardSteps()

	assert.NotNil(t, steps)
	assert.Greater(t, len(steps), 0)
}

func TestConfigurationWizardService_ValidateStepData(t *testing.T) {
	service := NewConfigurationWizardService(nil)

	tests := []struct {
		name    string
		stepID  string
		data    map[string]interface{}
		wantErr bool
	}{
		{
			name:    "empty step ID",
			stepID:  "",
			data:    map[string]interface{}{},
			wantErr: true,
		},
		{
			name:    "nil data",
			stepID:  "step_1",
			data:    nil,
			wantErr: true,
		},
		{
			name:    "unknown step ID",
			stepID:  "nonexistent_step",
			data:    map[string]interface{}{"key": "value"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateStepData(tt.stepID, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigurationWizardService_ValidateFieldType(t *testing.T) {
	service := NewConfigurationWizardService(nil)

	tests := []struct {
		name      string
		fieldType string
		value     interface{}
		wantErr   bool
	}{
		{
			name:      "valid text field",
			fieldType: "text",
			value:     "hello",
			wantErr:   false,
		},
		{
			name:      "valid number field",
			fieldType: "number",
			value:     float64(42),
			wantErr:   false,
		},
		{
			name:      "valid boolean field",
			fieldType: "boolean",
			value:     true,
			wantErr:   false,
		},
		{
			name:      "invalid number field with string",
			fieldType: "number",
			value:     "not a number",
			wantErr:   true,
		},
		{
			name:      "invalid boolean field with string",
			fieldType: "boolean",
			value:     "not a bool",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateFieldType(tt.fieldType, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigurationWizardService_CollectSystemInfo(t *testing.T) {
	service := NewConfigurationWizardService(nil)

	info := service.collectSystemInfo()

	assert.NotNil(t, info)
	assert.NotEmpty(t, info.OS)
	assert.NotEmpty(t, info.Architecture)
	assert.NotEmpty(t, info.GoVersion)
	assert.Greater(t, info.CPUCores, 0)
}

func TestConfigurationWizardService_GetWizardProgress(t *testing.T) {
	service := NewConfigurationWizardService(nil)

	// Without a session, should return sensible default
	progress := service.GetWizardProgress()

	assert.NotNil(t, progress)
}
