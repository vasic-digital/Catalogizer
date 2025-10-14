package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"catalogizer/models"
	"catalogizer/services"

	"github.com/gorilla/mux"
)

type ErrorReportingHandler struct {
	errorReportingService *services.ErrorReportingService
	authService           *services.AuthService
}

func NewErrorReportingHandler(errorReportingService *services.ErrorReportingService, authService *services.AuthService) *ErrorReportingHandler {
	return &ErrorReportingHandler{
		errorReportingService: errorReportingService,
		authService:           authService,
	}
}

func (h *ErrorReportingHandler) ReportError(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionReportCreate)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var request models.ErrorReportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	report, err := h.errorReportingService.ReportError(userID, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *ErrorReportingHandler) ReportCrash(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionReportCreate)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var request models.CrashReportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	report, err := h.errorReportingService.ReportCrash(userID, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *ErrorReportingHandler) GetErrorReport(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	reportID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionReportView)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	report, err := h.errorReportingService.GetErrorReport(reportID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *ErrorReportingHandler) GetCrashReport(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	reportID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionReportView)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	report, err := h.errorReportingService.GetCrashReport(reportID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *ErrorReportingHandler) ListErrorReports(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionReportView)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	filters := h.parseErrorReportFilters(r)
	reports, err := h.errorReportingService.GetErrorReportsByUser(userID, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reports": reports,
		"filters": filters,
	})
}

func (h *ErrorReportingHandler) ListCrashReports(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionReportView)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	filters := h.parseCrashReportFilters(r)
	reports, err := h.errorReportingService.GetCrashReportsByUser(userID, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reports": reports,
		"filters": filters,
	})
}

func (h *ErrorReportingHandler) UpdateErrorStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	reportID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionReportCreate)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var request struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.errorReportingService.UpdateErrorStatus(reportID, userID, request.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ErrorReportingHandler) UpdateCrashStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	reportID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionReportCreate)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var request struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.errorReportingService.UpdateCrashStatus(reportID, userID, request.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ErrorReportingHandler) GetErrorStatistics(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionReportView)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	stats, err := h.errorReportingService.GetErrorStatistics(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *ErrorReportingHandler) GetCrashStatistics(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionReportView)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	stats, err := h.errorReportingService.GetCrashStatistics(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *ErrorReportingHandler) GetSystemHealth(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionSystemAdmin)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	health, err := h.errorReportingService.GetSystemHealth()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (h *ErrorReportingHandler) UpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionSystemAdmin)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var config services.ErrorReportingConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.errorReportingService.UpdateConfiguration(&config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Configuration updated successfully",
	})
}

func (h *ErrorReportingHandler) GetConfiguration(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionSystemAdmin)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	config := h.errorReportingService.GetConfiguration()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (h *ErrorReportingHandler) CleanupOldReports(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionSystemAdmin)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var request struct {
		DaysOld int `json:"days_old"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.DaysOld <= 0 {
		request.DaysOld = 30 // Default to 30 days
	}

	olderThan := time.Now().AddDate(0, 0, -request.DaysOld)
	err = h.errorReportingService.CleanupOldReports(olderThan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Cleanup completed successfully",
	})
}

func (h *ErrorReportingHandler) ExportReports(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionReportView)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	filters := h.parseExportFilters(r)
	data, err := h.errorReportingService.ExportReports(userID, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set appropriate content type and headers
	switch filters.Format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=error_reports.json")
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=error_reports.csv")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=error_reports.txt")
	}

	w.Write(data)
}

// Helper methods

func (h *ErrorReportingHandler) parseErrorReportFilters(r *http.Request) *models.ErrorReportFilters {
	filters := &models.ErrorReportFilters{}

	if level := r.URL.Query().Get("level"); level != "" {
		filters.Level = level
	}

	if component := r.URL.Query().Get("component"); component != "" {
		filters.Component = component
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = status
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			filters.StartDate = &date
		}
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			filters.EndDate = &date
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	return filters
}

func (h *ErrorReportingHandler) parseCrashReportFilters(r *http.Request) *models.CrashReportFilters {
	filters := &models.CrashReportFilters{}

	if signal := r.URL.Query().Get("signal"); signal != "" {
		filters.Signal = signal
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = status
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			filters.StartDate = &date
		}
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			filters.EndDate = &date
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	return filters
}

func (h *ErrorReportingHandler) parseExportFilters(r *http.Request) *models.ExportFilters {
	filters := &models.ExportFilters{
		Format:         "json", // Default format
		IncludeErrors:  true,   // Default to include errors
		IncludeCrashes: true,   // Default to include crashes
	}

	if format := r.URL.Query().Get("format"); format != "" {
		filters.Format = format
	}

	if level := r.URL.Query().Get("level"); level != "" {
		filters.Level = level
	}

	if component := r.URL.Query().Get("component"); component != "" {
		filters.Component = component
	}

	if signal := r.URL.Query().Get("signal"); signal != "" {
		filters.Signal = signal
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			filters.StartDate = &date
		}
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			filters.EndDate = &date
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	if includeErrors := r.URL.Query().Get("include_errors"); includeErrors == "false" {
		filters.IncludeErrors = false
	}

	if includeCrashes := r.URL.Query().Get("include_crashes"); includeCrashes == "false" {
		filters.IncludeCrashes = false
	}

	return filters
}
