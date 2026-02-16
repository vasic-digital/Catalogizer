package services

import (
	"testing"
	"time"

	"catalogizer/models"

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
	assert.Greater(t, len(templates), 0)
}

func TestConfigurationWizardService_ValidateStepData(t *testing.T) {
	service := NewConfigurationWizardService(nil)

	tests := []struct {
		name    string
		step    WizardStep
		data    map[string]interface{}
		wantErr bool
	}{
		{
			name: "empty step with empty data",
			step: WizardStep{
				StepID: "",
				Fields: []FieldDefinition{},
			},
			data:    map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "step with required field missing",
			step: WizardStep{
				StepID: "step_1",
				Fields: []FieldDefinition{
					{
						FieldID:  "name",
						Label:    "Name",
						Type:     "text",
						Required: true,
					},
				},
			},
			data:    nil,
			wantErr: true,
		},
		{
			name: "step with required field present",
			step: WizardStep{
				StepID: "step_1",
				Fields: []FieldDefinition{
					{
						FieldID:  "name",
						Label:    "Name",
						Type:     "text",
						Required: true,
					},
				},
			},
			data:    map[string]interface{}{"name": "test"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateStepData(tt.step, tt.data)
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
		name    string
		field   FieldDefinition
		value   interface{}
		wantErr bool
	}{
		{
			name:    "valid text field",
			field:   FieldDefinition{FieldID: "f1", Type: "text"},
			value:   "hello",
			wantErr: false,
		},
		{
			name:    "valid number field",
			field:   FieldDefinition{FieldID: "f2", Type: "number"},
			value:   float64(42),
			wantErr: false,
		},
		{
			name:    "valid boolean field",
			field:   FieldDefinition{FieldID: "f3", Type: "boolean"},
			value:   true,
			wantErr: false,
		},
		{
			name:    "invalid number field with string",
			field:   FieldDefinition{FieldID: "f4", Type: "number"},
			value:   "not a number",
			wantErr: true,
		},
		{
			name:    "invalid boolean field with string",
			field:   FieldDefinition{FieldID: "f5", Type: "boolean"},
			value:   "not a bool",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateFieldType(tt.field, tt.value)
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

	// When the current session is set in memory, GetWizardProgress works without repo
	service.currentSession = &models.WizardSession{
		SessionID:   "test-session",
		UserID:      1,
		CurrentStep: 2,
		TotalSteps:  5,
		ConfigType:  "basic",
		StartedAt:   time.Now(),
		LastActivity: time.Now(),
	}

	progress, err := service.GetWizardProgress("test-session")
	assert.NoError(t, err)
	assert.NotNil(t, progress)
	assert.Equal(t, "test-session", progress.SessionID)
	assert.Greater(t, progress.Progress, float64(0))
}
