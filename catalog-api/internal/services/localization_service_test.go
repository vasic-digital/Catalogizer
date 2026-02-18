package services

import (
	"catalogizer/database"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewLocalizationService(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()

	service := NewLocalizationService(mockDB, mockLogger, nil, nil)

	assert.NotNil(t, service)
}

func TestLocalizationService_IsLanguageSupported(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewLocalizationService(mockDB, mockLogger, nil, nil)

	tests := []struct {
		name        string
		langCode    string
		contentType string
		expected    bool
	}{
		{
			name:        "english subtitles supported",
			langCode:    "en",
			contentType: "subtitles",
			expected:    true,
		},
		{
			name:        "english lyrics supported",
			langCode:    "en",
			contentType: "lyrics",
			expected:    true,
		},
		{
			name:        "english UI supported",
			langCode:    "en",
			contentType: "ui",
			expected:    true,
		},
		{
			name:        "arabic UI not supported",
			langCode:    "ar",
			contentType: "ui",
			expected:    false,
		},
		{
			name:        "arabic subtitles supported",
			langCode:    "ar",
			contentType: "subtitles",
			expected:    true,
		},
		{
			name:        "unsupported language",
			langCode:    "xx",
			contentType: "subtitles",
			expected:    false,
		},
		{
			name:        "empty language code",
			langCode:    "",
			contentType: "subtitles",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.IsLanguageSupported(context.Background(), tt.langCode, tt.contentType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLocalizationService_DetectUserLanguage(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewLocalizationService(mockDB, mockLogger, nil, nil)

	tests := []struct {
		name           string
		userAgent      string
		acceptLanguage string
		expected       string
	}{
		{
			name:           "english accept language",
			userAgent:      "Mozilla/5.0",
			acceptLanguage: "en-US,en;q=0.9",
			expected:       "en",
		},
		{
			name:           "french accept language",
			userAgent:      "Mozilla/5.0",
			acceptLanguage: "fr-FR,fr;q=0.9,en;q=0.8",
			expected:       "fr",
		},
		{
			name:           "japanese accept language",
			userAgent:      "Mozilla/5.0",
			acceptLanguage: "ja,en;q=0.5",
			expected:       "ja",
		},
		{
			name:           "empty accept language defaults to english",
			userAgent:      "Mozilla/5.0",
			acceptLanguage: "",
			expected:       "en",
		},
		{
			name:           "unsupported language defaults to english",
			userAgent:      "Mozilla/5.0",
			acceptLanguage: "xx-XX",
			expected:       "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.DetectUserLanguage(context.Background(), tt.userAgent, tt.acceptLanguage)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLocalizationService_GetWizardDefaults(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewLocalizationService(mockDB, mockLogger, nil, nil)

	tests := []struct {
		name             string
		detectedLanguage string
		expectedLang     string
		expectedRegion   string
	}{
		{
			name:             "english defaults",
			detectedLanguage: "en",
			expectedLang:     "en",
			expectedRegion:   "US",
		},
		{
			name:             "french defaults",
			detectedLanguage: "fr",
			expectedLang:     "fr",
			expectedRegion:   "FR",
		},
		{
			name:             "unsupported language falls back to english",
			detectedLanguage: "xx",
			expectedLang:     "en",
			expectedRegion:   "US",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetWizardDefaults(context.Background(), tt.detectedLanguage)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedLang, result.PrimaryLanguage)
			assert.Equal(t, tt.expectedRegion, result.PreferredRegion)
			assert.NotEmpty(t, result.DateFormat)
			assert.NotEmpty(t, result.TimeFormat)
			assert.NotEmpty(t, result.CurrencyCode)
		})
	}
}

func TestLocalizationService_GetDefaultDateFormat(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewLocalizationService(mockDB, mockLogger, nil, nil)

	tests := []struct {
		name     string
		region   string
		expected string
	}{
		{
			name:     "US format",
			region:   "US",
			expected: "MM/DD/YYYY",
		},
		{
			name:     "GB format",
			region:   "GB",
			expected: "DD/MM/YYYY",
		},
		{
			name:     "default ISO format",
			region:   "DE",
			expected: "YYYY-MM-DD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getDefaultDateFormat(tt.region)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLocalizationService_GetDefaultTimeFormat(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewLocalizationService(mockDB, mockLogger, nil, nil)

	tests := []struct {
		name     string
		region   string
		expected string
	}{
		{
			name:     "US 12h format",
			region:   "US",
			expected: "12h",
		},
		{
			name:     "default 24h format",
			region:   "DE",
			expected: "24h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getDefaultTimeFormat(tt.region)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLocalizationService_GetDefaultCurrency(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewLocalizationService(mockDB, mockLogger, nil, nil)

	tests := []struct {
		name     string
		region   string
		expected string
	}{
		{
			name:     "US dollar",
			region:   "US",
			expected: "USD",
		},
		{
			name:     "British pound",
			region:   "GB",
			expected: "GBP",
		},
		{
			name:     "Japanese yen",
			region:   "JP",
			expected: "JPY",
		},
		{
			name:     "Euro for France",
			region:   "FR",
			expected: "EUR",
		},
		{
			name:     "unknown region defaults to USD",
			region:   "XX",
			expected: "USD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getDefaultCurrency(tt.region)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLocalizationService_ValidateConfiguration(t *testing.T) {
	mockDB := database.WrapDB(nil, database.DialectSQLite)
	mockLogger := zap.NewNop()
	service := NewLocalizationService(mockDB, mockLogger, nil, nil)

	tests := []struct {
		name       string
		config     ConfigurationExport
		wantErrors bool
	}{
		{
			name: "valid full configuration",
			config: ConfigurationExport{
				Version:    "1.0",
				ConfigType: "full",
				Localization: &UserLocalization{
					PrimaryLanguage: "en",
					DateFormat:      "MM/DD/YYYY",
					TimeFormat:      "12h",
				},
				MediaSettings: &MediaPlayerConfig{
					DefaultQuality:    "high",
					VolumeLevel:       0.8,
					CrossfadeDuration: 3000,
				},
			},
			wantErrors: false,
		},
		{
			name: "missing version",
			config: ConfigurationExport{
				Version:    "",
				ConfigType: "full",
			},
			wantErrors: true,
		},
		{
			name: "invalid config type",
			config: ConfigurationExport{
				Version:    "1.0",
				ConfigType: "invalid",
			},
			wantErrors: true,
		},
		{
			name: "unsupported primary language",
			config: ConfigurationExport{
				Version:    "1.0",
				ConfigType: "localization",
				Localization: &UserLocalization{
					PrimaryLanguage: "xx",
					DateFormat:      "MM/DD/YYYY",
					TimeFormat:      "12h",
				},
			},
			wantErrors: true,
		},
		{
			name: "invalid volume level",
			config: ConfigurationExport{
				Version:    "1.0",
				ConfigType: "media",
				MediaSettings: &MediaPlayerConfig{
					DefaultQuality: "high",
					VolumeLevel:    2.0,
				},
			},
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := service.validateConfiguration(&tt.config)
			if tt.wantErrors {
				assert.NotEmpty(t, errors)
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}
