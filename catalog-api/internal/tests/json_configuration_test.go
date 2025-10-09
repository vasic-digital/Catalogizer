package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"catalog-api/internal/handlers"
	"catalog-api/internal/services"
)

func TestJSONConfigurationHandlers(t *testing.T) {
	// Setup test environment
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	translationService := services.NewTranslationService(db, logger, "")
	cacheService := services.NewCacheService(db, logger)
	localizationService := services.NewLocalizationService(db, logger, translationService, cacheService)

	handlers := handlers.NewLocalizationHandlers(logger, localizationService)

	// Setup router
	router := mux.NewRouter()
	handlers.RegisterRoutes(router)

	userID := int64(300)

	// Setup initial user localization for testing
	localizationReq := &services.WizardLocalizationStep{
		UserID:                userID,
		PrimaryLanguage:       "en",
		SecondaryLanguages:    []string{"es"},
		SubtitleLanguages:     []string{"en", "es"},
		AutoTranslate:         true,
		AutoDownloadSubtitles: true,
		PreferredRegion:       "US",
		CurrencyCode:          "USD",
		DateFormat:            "MM/DD/YYYY",
		TimeFormat:            "12h",
		Timezone:              "America/New_York",
	}
	_, err := localizationService.SetupUserLocalization(context.Background(), localizationReq)
	require.NoError(t, err)

	t.Run("ExportConfiguration", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"config_type": "full",
			"description": "Test configuration export",
			"tags":        []string{"test", "handler"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/wizard/configuration/export", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "300")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Success bool                            `json:"success"`
			Data    services.ConfigurationExport   `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "1.0", response.Data.Version)
		assert.Equal(t, userID, response.Data.ExportedBy)
		assert.Equal(t, "full", response.Data.ConfigType)
		assert.Equal(t, "Test configuration export", response.Data.Description)
		assert.Contains(t, response.Data.Tags, "test")
		assert.NotNil(t, response.Data.Localization)
	})

	t.Run("ValidateConfiguration", func(t *testing.T) {
		// Test valid configuration
		validConfig := map[string]interface{}{
			"version":      "1.0",
			"config_type":  "localization",
			"localization": map[string]interface{}{
				"user_id":           userID,
				"primary_language":  "de",
				"secondary_languages": []string{"en"},
				"preferred_region":  "DE",
				"currency_code":     "EUR",
			},
		}

		validJSON, _ := json.Marshal(validConfig)
		reqBody := map[string]interface{}{
			"config_json": string(validJSON),
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/wizard/configuration/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Success bool                               `json:"success"`
			Data    services.ConfigurationValidation  `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.True(t, response.Data.Valid)
		assert.Empty(t, response.Data.Errors)

		// Test invalid configuration
		invalidConfig := map[string]interface{}{
			"version": "invalid",
			"localization": map[string]interface{}{
				"primary_language": "",
			},
		}

		invalidJSON, _ := json.Marshal(invalidConfig)
		invalidReqBody := map[string]interface{}{
			"config_json": string(invalidJSON),
		}

		invalidJSONBody, _ := json.Marshal(invalidReqBody)
		invalidReq := httptest.NewRequest("POST", "/api/v1/wizard/configuration/validate", bytes.NewBuffer(invalidJSONBody))
		invalidReq.Header.Set("Content-Type", "application/json")

		invalidW := httptest.NewRecorder()
		router.ServeHTTP(invalidW, invalidReq)

		assert.Equal(t, http.StatusOK, invalidW.Code)

		var invalidResponse struct {
			Success bool                               `json:"success"`
			Data    services.ConfigurationValidation  `json:"data"`
		}
		err = json.Unmarshal(invalidW.Body.Bytes(), &invalidResponse)
		require.NoError(t, err)

		assert.True(t, invalidResponse.Success)
		assert.False(t, invalidResponse.Data.Valid)
		assert.NotEmpty(t, invalidResponse.Data.Errors)
	})

	t.Run("ImportConfiguration", func(t *testing.T) {
		// Create configuration to import
		importConfig := services.ConfigurationExport{
			Version:    "1.0",
			ConfigType: "localization",
			Localization: &services.UserLocalization{
				UserID:            userID + 1,
				PrimaryLanguage:   "fr",
				SecondaryLanguages: []string{"en"},
				SubtitleLanguages: []string{"fr", "en"},
				AutoTranslate:     true,
				PreferredRegion:   "FR",
				CurrencyCode:      "EUR",
				DateFormat:        "DD/MM/YYYY",
				TimeFormat:        "24h",
				Timezone:          "Europe/Paris",
			},
			Description: "French configuration",
			Tags:        []string{"french", "europe"},
		}

		configJSON, _ := json.Marshal(importConfig)
		reqBody := map[string]interface{}{
			"config_json": string(configJSON),
			"options": map[string]bool{
				"overwrite_existing": true,
				"backup_current":     true,
				"validate_only":      false,
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/wizard/configuration/import", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "301")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Success bool                                `json:"success"`
			Data    services.ConfigurationImportResult `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.True(t, response.Data.Success)
		assert.Equal(t, "localization", response.Data.ConfigType)
		assert.NotEmpty(t, response.Data.BackupID)
	})

	t.Run("EditConfiguration", func(t *testing.T) {
		// First export current configuration
		exportReqBody := map[string]interface{}{
			"config_type": "localization",
			"description": "Configuration for editing",
		}

		exportJSONBody, _ := json.Marshal(exportReqBody)
		exportReq := httptest.NewRequest("POST", "/api/v1/wizard/configuration/export", bytes.NewBuffer(exportJSONBody))
		exportReq.Header.Set("Content-Type", "application/json")
		exportReq.Header.Set("X-User-ID", "300")

		exportW := httptest.NewRecorder()
		router.ServeHTTP(exportW, exportReq)

		var exportResponse struct {
			Success bool                            `json:"success"`
			Data    services.ConfigurationExport   `json:"data"`
		}
		err := json.Unmarshal(exportW.Body.Bytes(), &exportResponse)
		require.NoError(t, err)

		// Convert to JSON for editing
		currentJSON, _ := json.Marshal(exportResponse.Data)

		// Edit the configuration
		reqBody := map[string]interface{}{
			"config_json": string(currentJSON),
			"edits": map[string]interface{}{
				"localization.primary_language": "de",
				"localization.currency_code":    "EUR",
				"description":                   "Updated to German",
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/wizard/configuration/edit", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "300")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Success bool `json:"success"`
			Data    struct {
				EditedConfig string `json:"edited_config"`
				Message      string `json:"message"`
			} `json:"data"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "Configuration edited successfully", response.Data.Message)

		// Verify the edits were applied
		var editedConfig services.ConfigurationExport
		err = json.Unmarshal([]byte(response.Data.EditedConfig), &editedConfig)
		require.NoError(t, err)

		assert.Equal(t, "de", editedConfig.Localization.PrimaryLanguage)
		assert.Equal(t, "EUR", editedConfig.Localization.CurrencyCode)
		assert.Equal(t, "Updated to German", editedConfig.Description)
	})

	t.Run("GetConfigurationTemplates", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/wizard/configuration/templates", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Success bool `json:"success"`
			Data    struct {
				Templates []services.ConfigurationTemplate `json:"templates"`
				Count     int                              `json:"count"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.NotEmpty(t, response.Data.Templates)
		assert.Equal(t, len(response.Data.Templates), response.Data.Count)

		// Verify template structure
		template := response.Data.Templates[0]
		assert.NotEmpty(t, template.Name)
		assert.NotEmpty(t, template.Description)
		assert.NotEmpty(t, template.Template)

		// Verify template JSON is valid
		var templateConfig map[string]interface{}
		err = json.Unmarshal([]byte(template.Template), &templateConfig)
		require.NoError(t, err)
	})

	t.Run("ErrorCases", func(t *testing.T) {
		// Test unauthenticated user
		req := httptest.NewRequest("POST", "/api/v1/wizard/configuration/export", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		// Test invalid JSON
		invalidReq := httptest.NewRequest("POST", "/api/v1/wizard/configuration/validate", bytes.NewBuffer([]byte("invalid json")))
		invalidReq.Header.Set("Content-Type", "application/json")

		invalidW := httptest.NewRecorder()
		router.ServeHTTP(invalidW, invalidReq)

		assert.Equal(t, http.StatusBadRequest, invalidW.Code)

		// Test missing configuration JSON
		emptyReq := map[string]interface{}{}
		emptyJSON, _ := json.Marshal(emptyReq)

		missingReq := httptest.NewRequest("POST", "/api/v1/wizard/configuration/validate", bytes.NewBuffer(emptyJSON))
		missingReq.Header.Set("Content-Type", "application/json")

		missingW := httptest.NewRecorder()
		router.ServeHTTP(missingW, missingReq)

		assert.Equal(t, http.StatusBadRequest, missingW.Code)
	})
}

func TestJSONConfigurationPerformance(t *testing.T) {
	// Test performance with large configurations
	db := setupTestDB(t)
	defer db.Close()

	logger := zaptest.NewLogger(t)
	translationService := services.NewTranslationService(db, logger, "")
	cacheService := services.NewCacheService(db, logger)
	localizationService := services.NewLocalizationService(db, logger, translationService, cacheService)

	userID := int64(400)

	// Setup user with extensive configuration
	localizationReq := &services.WizardLocalizationStep{
		UserID:                userID,
		PrimaryLanguage:       "en",
		SecondaryLanguages:    []string{"es", "fr", "de", "it", "pt", "ru", "ja", "ko", "zh"},
		SubtitleLanguages:     []string{"en", "es", "fr", "de", "it", "pt", "ru", "ja", "ko", "zh"},
		AutoTranslate:         true,
		AutoDownloadSubtitles: true,
		PreferredRegion:       "US",
		CurrencyCode:          "USD",
		DateFormat:            "MM/DD/YYYY",
		TimeFormat:            "12h",
		Timezone:              "America/New_York",
	}
	_, err := localizationService.SetupUserLocalization(context.Background(), localizationReq)
	require.NoError(t, err)

	t.Run("LargeConfigurationExport", func(t *testing.T) {
		// Test exporting large configuration
		config, err := localizationService.ExportConfiguration(context.Background(), userID, "full", "Large configuration test", []string{"performance", "large"})
		require.NoError(t, err)
		require.NotNil(t, config)

		// Verify structure is complete
		assert.NotNil(t, config.Localization)
		assert.NotNil(t, config.MediaSettings)
		assert.Len(t, config.Localization.SecondaryLanguages, 9)
		assert.Len(t, config.Localization.SubtitleLanguages, 10) // Primary + secondary
	})

	t.Run("LargeConfigurationValidation", func(t *testing.T) {
		// Create large configuration for validation
		largeConfig := map[string]interface{}{
			"version":     "1.0",
			"config_type": "full",
			"localization": map[string]interface{}{
				"user_id":             userID,
				"primary_language":    "en",
				"secondary_languages": []string{"es", "fr", "de", "it", "pt", "ru", "ja", "ko", "zh"},
				"subtitle_languages":  []string{"en", "es", "fr", "de", "it", "pt", "ru", "ja", "ko", "zh"},
				"auto_translate":      true,
			},
			"media_settings": map[string]interface{}{
				"playback_settings": map[string]interface{}{
					"default_volume":        0.8,
					"enable_crossfade":      true,
					"crossfade_duration":    3.0,
					"enable_replay_gain":    true,
				},
				"video_settings": map[string]interface{}{
					"default_quality":       "1080p",
					"enable_hardware_accel": true,
					"subtitle_font_size":    16,
				},
				"audio_settings": map[string]interface{}{
					"sample_rate":       44100,
					"bit_depth":         16,
					"enable_equalizer":  true,
				},
			},
		}

		largeJSON, _ := json.Marshal(largeConfig)
		validation := localizationService.ValidateConfigurationJSON(context.Background(), string(largeJSON))

		assert.True(t, validation.Valid)
		assert.Empty(t, validation.Errors)
		assert.NotEmpty(t, validation.Summary)
	})

	t.Run("BulkConfigurationOperations", func(t *testing.T) {
		// Test multiple rapid operations
		for i := 0; i < 10; i++ {
			// Export
			config, err := localizationService.ExportConfiguration(context.Background(), userID, "localization", "Bulk test", []string{"bulk"})
			require.NoError(t, err)

			// Validate
			configJSON, _ := json.Marshal(config)
			validation := localizationService.ValidateConfigurationJSON(context.Background(), string(configJSON))
			assert.True(t, validation.Valid)

			// Edit
			edits := map[string]interface{}{
				"description": "Bulk operation test iteration",
			}
			editedJSON, err := localizationService.EditConfiguration(context.Background(), userID, string(configJSON), edits)
			require.NoError(t, err)
			assert.NotEmpty(t, editedJSON)
		}
	})
}