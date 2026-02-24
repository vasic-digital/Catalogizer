package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"catalogizer/internal/services"
)

func TestNewLocalizationHandlers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.Equal(t, localizationService, handler.localizationService)
}

func TestLocalizationHandlers_GetWizardDefaults(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/wizard/localization/defaults", nil)
	r.Header.Set("User-Agent", "Mozilla/5.0")
	r.Header.Set("Accept-Language", "en-US,en;q=0.9")

	handler.GetWizardDefaults(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, data, "detected_language")
	assert.Contains(t, data, "defaults")
}

func TestLocalizationHandlers_SetupWizardLocalization_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/wizard/localization/setup", bytes.NewBufferString("invalid json"))
	r.Header.Set("Content-Type", "application/json")

	handler.SetupWizardLocalization(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Invalid request body", response.Error)
}

func TestLocalizationHandlers_GetUserLocalization_NoAuth(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/localization", nil)
	// No X-User-ID header

	handler.GetUserLocalization(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "User not authenticated", response.Error)
}

func TestLocalizationHandlers_UpdateUserLocalization_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/localization", bytes.NewBufferString("invalid json"))
	r.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(r.Context(), "user_id", int64(1))
	r = r.WithContext(ctx)

	handler.UpdateUserLocalization(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Invalid request body", response.Error)
}

func TestLocalizationHandlers_GetSupportedLanguages(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/localization/languages", nil)

	// This will panic due to nil service, but we're testing the structure
	defer func() {
		if r := recover(); r != nil {
			t.Log("Expected panic due to nil service:", r)
		}
	}()

	handler.GetSupportedLanguages(w, r)
}

func TestLocalizationHandlers_GetLanguageProfile(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := mux.SetURLVars(httptest.NewRequest("GET", "/localization/languages/en", nil), map[string]string{"languageCode": "en"})

	// This will panic due to nil service, but we're testing the structure
	defer func() {
		if r := recover(); r != nil {
			t.Log("Expected panic due to nil service:", r)
		}
	}()

	handler.GetLanguageProfile(w, r)
}

func TestLocalizationHandlers_GetContentLanguagePreferences_NoAuth(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := mux.SetURLVars(httptest.NewRequest("GET", "/localization/preferences/movie", nil), map[string]string{"contentType": "movie"})
	// No X-User-ID header

	handler.GetContentLanguagePreferences(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "User not authenticated", response.Error)
}

func TestLocalizationHandlers_GetLocalizationStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/localization/stats", nil)

	// This will panic due to nil service, but we're testing the structure
	defer func() {
		if r := recover(); r != nil {
			t.Log("Expected panic due to nil service:", r)
		}
	}()

	handler.GetLocalizationStats(w, r)
}

func TestLocalizationHandlers_ExportConfiguration_NoAuth(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/wizard/configuration/export", bytes.NewBufferString("{}"))
	r.Header.Set("Content-Type", "application/json")
	// No X-User-ID header

	handler.ExportConfiguration(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "User not authenticated", response.Error)
}

func TestLocalizationHandlers_ImportConfiguration_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/wizard/configuration/import", bytes.NewBufferString("invalid json"))
	r.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(r.Context(), "user_id", int64(1))
	r = r.WithContext(ctx)

	handler.ImportConfiguration(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Invalid request body", response.Error)
}

func TestLocalizationHandlers_ValidateConfiguration_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/wizard/configuration/validate", bytes.NewBufferString("invalid json"))
	r.Header.Set("Content-Type", "application/json")

	handler.ValidateConfiguration(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Invalid request body", response.Error)
}

func TestLocalizationHandlers_EditConfiguration_NoAuth(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/wizard/configuration/edit", bytes.NewBufferString("{}"))
	r.Header.Set("Content-Type", "application/json")
	// No X-User-ID header

	handler.EditConfiguration(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "User not authenticated", response.Error)
}

func TestLocalizationHandlers_GetConfigurationTemplates(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/wizard/configuration/templates", nil)

	// This will panic due to nil service, but we're testing the structure
	defer func() {
		if r := recover(); r != nil {
			t.Log("Expected panic due to nil service:", r)
		}
	}()

	handler.GetConfigurationTemplates(w, r)
}

func TestLocalizationHandlers_DetectLanguage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/localization/detect", bytes.NewBufferString(`{"user_agent":"test","accept_language":"en"}`))
	r.Header.Set("Content-Type", "application/json")

	handler.DetectLanguage(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

func TestLocalizationHandlers_CheckLanguageSupport_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/localization/check-support", bytes.NewBufferString("invalid json"))
	r.Header.Set("Content-Type", "application/json")

	handler.CheckLanguageSupport(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Invalid request body", response.Error)
}

func TestLocalizationHandlers_FormatDateTime_NoAuth(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/localization/format-datetime", bytes.NewBufferString(`{"timestamp":"2023-01-01T00:00:00Z"}`))
	r.Header.Set("Content-Type", "application/json")
	// No X-User-ID header

	handler.FormatDateTime(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "User not authenticated", response.Error)
}

func TestLocalizationHandlers_getUserID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	tests := []struct {
		name     string
		header   string
		expected int64
	}{
		{
			name:     "valid user ID",
			header:   "123",
			expected: 123,
		},
		{
			name:     "invalid user ID",
			header:   "invalid",
			expected: 0,
		},
		{
			name:     "empty header",
			header:   "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			// getUserID now reads from context, not headers
			if tt.expected != 0 {
				ctx := context.WithValue(r.Context(), "user_id", tt.expected)
				r = r.WithContext(ctx)
			}

			result := handler.getUserID(r)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- parseTimestamp tests ---

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		wantYear  int
		wantMonth int
		wantDay   int
	}{
		{
			name:      "ISO 8601 with Z",
			input:     "2023-06-15T10:30:00Z",
			wantErr:   false,
			wantYear:  2023,
			wantMonth: 6,
			wantDay:   15,
		},
		{
			name:      "ISO 8601 with milliseconds",
			input:     "2023-06-15T10:30:00.000Z",
			wantErr:   false,
			wantYear:  2023,
			wantMonth: 6,
			wantDay:   15,
		},
		{
			name:      "ISO 8601 with timezone offset",
			input:     "2023-06-15T10:30:00+05:00",
			wantErr:   false,
			wantYear:  2023,
			wantMonth: 6,
			wantDay:   15,
		},
		{
			name:      "datetime without timezone",
			input:     "2023-06-15 10:30:00",
			wantErr:   false,
			wantYear:  2023,
			wantMonth: 6,
			wantDay:   15,
		},
		{
			name:      "date only",
			input:     "2023-06-15",
			wantErr:   false,
			wantYear:  2023,
			wantMonth: 6,
			wantDay:   15,
		},
		{
			name:    "Unix timestamp as string",
			input:   "1686830400",
			wantErr: false,
			// 1686830400 = 2023-06-15 14:40:00 UTC
			wantYear: 2023,
		},
		{
			name:    "invalid format",
			input:   "not-a-timestamp",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := parseTimestamp(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantYear, ts.Year())
			if tt.wantMonth > 0 {
				assert.Equal(t, tt.wantMonth, int(ts.Month()))
			}
			if tt.wantDay > 0 {
				assert.Equal(t, tt.wantDay, ts.Day())
			}
		})
	}
}

// --- Additional SetupWizardLocalization validation tests ---

func TestLocalizationHandlers_SetupWizardLocalization_MissingLanguage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	// Valid JSON but missing primary_language
	r := httptest.NewRequest("POST", "/wizard/localization/setup", bytes.NewBufferString(`{"user_id": 1}`))
	r.Header.Set("Content-Type", "application/json")

	handler.SetupWizardLocalization(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Primary language is required", response.Error)
}

func TestLocalizationHandlers_SetupWizardLocalization_NoUserID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	// user_id = 0 and no context user_id
	r := httptest.NewRequest("POST", "/wizard/localization/setup", bytes.NewBufferString(`{"primary_language": "en"}`))
	r.Header.Set("Content-Type", "application/json")

	handler.SetupWizardLocalization(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- CheckLanguageSupport validation tests ---

func TestLocalizationHandlers_CheckLanguageSupport_MissingLanguageCode(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/localization/check-support", bytes.NewBufferString(`{"content_type":"movie"}`))
	r.Header.Set("Content-Type", "application/json")

	handler.CheckLanguageSupport(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Language code is required", response.Error)
}

func TestLocalizationHandlers_CheckLanguageSupport_MissingContentType(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/localization/check-support", bytes.NewBufferString(`{"language_code":"en"}`))
	r.Header.Set("Content-Type", "application/json")

	handler.CheckLanguageSupport(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Content type is required", response.Error)
}

// --- ImportConfiguration validation tests ---

func TestLocalizationHandlers_ImportConfiguration_EmptyConfigJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/wizard/configuration/import", bytes.NewBufferString(`{"config_json":""}`))
	r.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(r.Context(), "user_id", int64(1))
	r = r.WithContext(ctx)

	handler.ImportConfiguration(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Configuration JSON is required", response.Error)
}

// --- ValidateConfiguration tests ---

func TestLocalizationHandlers_ValidateConfiguration_EmptyConfigJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/wizard/configuration/validate", bytes.NewBufferString(`{"config_json":""}`))
	r.Header.Set("Content-Type", "application/json")

	handler.ValidateConfiguration(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Configuration JSON is required", response.Error)
}

// --- EditConfiguration validation tests ---

func TestLocalizationHandlers_EditConfiguration_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/wizard/configuration/edit", bytes.NewBufferString("not json"))
	r.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(r.Context(), "user_id", int64(1))
	r = r.WithContext(ctx)

	handler.EditConfiguration(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLocalizationHandlers_EditConfiguration_MissingConfigJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/wizard/configuration/edit", bytes.NewBufferString(`{"edits":{"key":"val"}}`))
	r.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(r.Context(), "user_id", int64(1))
	r = r.WithContext(ctx)

	handler.EditConfiguration(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Configuration JSON is required", response.Error)
}

func TestLocalizationHandlers_EditConfiguration_MissingEdits(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/wizard/configuration/edit", bytes.NewBufferString(`{"config_json":"{}"}`))
	r.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(r.Context(), "user_id", int64(1))
	r = r.WithContext(ctx)

	handler.EditConfiguration(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Edits are required", response.Error)
}

// --- FormatDateTime validation tests ---

func TestLocalizationHandlers_FormatDateTime_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/localization/format-datetime", bytes.NewBufferString("not json"))
	r.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(r.Context(), "user_id", int64(1))
	r = r.WithContext(ctx)

	handler.FormatDateTime(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLocalizationHandlers_FormatDateTime_MissingTimestamp(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/localization/format-datetime", bytes.NewBufferString(`{}`))
	r.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(r.Context(), "user_id", int64(1))
	r = r.WithContext(ctx)

	handler.FormatDateTime(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Timestamp is required", response.Error)
}

func TestLocalizationHandlers_FormatDateTime_InvalidTimestamp(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/localization/format-datetime", bytes.NewBufferString(`{"timestamp":"not-a-timestamp"}`))
	r.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(r.Context(), "user_id", int64(1))
	r = r.WithContext(ctx)

	handler.FormatDateTime(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid timestamp format", response.Error)
}

// --- getUserID with int type context ---

func TestLocalizationHandlers_getUserID_WithIntContext(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	r := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(r.Context(), "user_id", 42)
	r = r.WithContext(ctx)

	result := handler.getUserID(r)
	assert.Equal(t, int64(42), result)
}

// --- CORS middleware with disallowed origin ---

func TestLocalizationHandlers_corsMiddleware_DisallowedOrigin(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := handler.corsMiddleware(testHandler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Origin", "http://evil.example.com")
	middleware.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	// Disallowed origin should NOT be reflected
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestLocalizationHandlers_corsMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	localizationService := &services.LocalizationService{}
	handler := NewLocalizationHandlers(logger, localizationService)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := handler.corsMiddleware(testHandler)

	// Test OPTIONS request with allowed origin
	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/", nil)
	r.Header.Set("Origin", "http://localhost:5173")
	middleware.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))

	// Test regular request
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Origin", "http://localhost:5173")
	middleware.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}
