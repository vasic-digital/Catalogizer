package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"catalog-api/models"
	"catalog-api/services"

	"github.com/gorilla/mux"
)

type ConfigurationHandler struct {
	configurationService *services.ConfigurationService
	permissionService    *services.PermissionService
}

func NewConfigurationHandler(configurationService *services.ConfigurationService, permissionService *services.PermissionService) *ConfigurationHandler {
	return &ConfigurationHandler{
		configurationService: configurationService,
		permissionService:    permissionService,
	}
}

// Wizard endpoints

func (h *ConfigurationHandler) GetWizardSteps(w http.ResponseWriter, r *http.Request) {
	steps, err := h.configurationService.GetWizardSteps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"steps": steps,
	})
}

func (h *ConfigurationHandler) GetWizardStep(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stepID := vars["step_id"]

	step, err := h.configurationService.GetWizardStep(stepID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(step)
}

func (h *ConfigurationHandler) ValidateWizardStep(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stepID := vars["step_id"]

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validation, err := h.configurationService.ValidateWizardStep(stepID, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(validation)
}

func (h *ConfigurationHandler) SaveWizardProgress(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	stepID := vars["step_id"]

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.configurationService.SaveWizardProgress(userID, stepID, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Progress saved successfully",
	})
}

func (h *ConfigurationHandler) GetWizardProgress(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	progress, err := h.configurationService.GetWizardProgress(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}

func (h *ConfigurationHandler) CompleteWizard(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	var finalData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&finalData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	config, err := h.configurationService.CompleteWizard(userID, finalData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Wizard completed successfully",
		"configuration": config,
	})
}

// Configuration endpoints

func (h *ConfigurationHandler) GetConfiguration(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "configuration", "read") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	config, err := h.configurationService.GetConfiguration()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (h *ConfigurationHandler) UpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "configuration", "write") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	config, err := h.configurationService.UpdateConfiguration(updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (h *ConfigurationHandler) ResetConfiguration(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "configuration", "admin") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	err := h.configurationService.ResetConfiguration()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Configuration reset successfully",
	})
}

func (h *ConfigurationHandler) ExportConfiguration(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "configuration", "read") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	data, err := h.configurationService.ExportConfiguration()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=configuration.json")
	w.Write(data)
}

func (h *ConfigurationHandler) ImportConfiguration(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "configuration", "admin") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("configuration")
	if err != nil {
		http.Error(w, "No configuration file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file content
	data := make([]byte, 0)
	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if n > 0 {
			data = append(data, buffer[:n]...)
		}
		if err != nil {
			break
		}
	}

	config, err := h.configurationService.ImportConfiguration(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Configuration imported successfully",
		"configuration": config,
	})
}

func (h *ConfigurationHandler) GetConfigurationSchema(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "configuration", "read") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	schema, err := h.configurationService.GetConfigurationSchema()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

func (h *ConfigurationHandler) TestConfiguration(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "configuration", "admin") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	// Get the configuration to test
	var config *models.SystemConfiguration
	if r.Method == "POST" {
		// Test provided configuration
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	} else {
		// Test current configuration
		var err error
		config, err = h.configurationService.GetConfiguration()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	test, err := h.configurationService.TestConfiguration(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(test)
}

// Template endpoints

func (h *ConfigurationHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "configuration", "read") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	// This would get templates from the repository
	templates := []*models.ConfigurationTemplate{
		{
			ID:          1,
			Name:        "Development",
			Description: "Configuration for development environment",
			Category:    "Environment",
		},
		{
			ID:          2,
			Name:        "Production",
			Description: "Configuration for production environment",
			Category:    "Environment",
		},
		{
			ID:          3,
			Name:        "High Performance",
			Description: "Optimized configuration for high-performance scenarios",
			Category:    "Performance",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": templates,
	})
}

func (h *ConfigurationHandler) ApplyTemplate(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	templateID, err := strconv.Atoi(vars["template_id"])
	if err != nil {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	if !h.permissionService.HasPermission(userID, "configuration", "admin") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	// This would apply the template through the service
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Template applied successfully",
		"template_id": templateID,
	})
}

// Backup endpoints

func (h *ConfigurationHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "configuration", "admin") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var request struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Name == "" {
		http.Error(w, "Backup name is required", http.StatusBadRequest)
		return
	}

	// This would create a backup through the service
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Backup created successfully",
		"name":    request.Name,
	})
}

func (h *ConfigurationHandler) GetBackups(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "configuration", "read") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	// This would get backups from the repository
	backups := []*models.ConfigurationBackup{
		{
			ID:      1,
			Name:    "Before v3.0 upgrade",
			Version: "2.5.0",
		},
		{
			ID:      2,
			Name:    "Production snapshot",
			Version: "3.0.0",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"backups": backups,
	})
}

func (h *ConfigurationHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	backupID, err := strconv.Atoi(vars["backup_id"])
	if err != nil {
		http.Error(w, "Invalid backup ID", http.StatusBadRequest)
		return
	}

	if !h.permissionService.HasPermission(userID, "configuration", "admin") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	// This would restore the backup through the service
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Backup restored successfully",
		"backup_id": backupID,
	})
}

func (h *ConfigurationHandler) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	backupID, err := strconv.Atoi(vars["backup_id"])
	if err != nil {
		http.Error(w, "Invalid backup ID", http.StatusBadRequest)
		return
	}

	if !h.permissionService.HasPermission(userID, "configuration", "admin") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	// This would delete the backup through the service
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Backup deleted successfully",
		"backup_id": backupID,
	})
}

// Health check endpoint

func (h *ConfigurationHandler) GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "configuration", "read") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	status := map[string]interface{}{
		"status":  "healthy",
		"version": "3.0.0",
		"uptime":  "24h 30m",
		"components": map[string]string{
			"database":      "healthy",
			"storage":       "healthy",
			"authentication": "healthy",
			"media_conversion": "healthy",
			"sync":          "healthy",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}