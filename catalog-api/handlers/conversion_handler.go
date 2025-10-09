package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"catalogizer/models"
	"catalogizer/services"
)

type ConversionHandler struct {
	conversionService *services.ConversionService
	authService       *services.AuthService
}

func NewConversionHandler(conversionService *services.ConversionService, authService *services.AuthService) *ConversionHandler {
	return &ConversionHandler{
		conversionService: conversionService,
		authService:       authService,
	}
}

func (h *ConversionHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hasPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionUploadMedia)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req models.ConversionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job, err := h.conversionService.CreateConversionJob(currentUser.ID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

func (h *ConversionHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	jobIDStr := strings.TrimPrefix(r.URL.Path, "/api/conversion/jobs/")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	job, err := h.conversionService.GetJob(jobID, currentUser.ID)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get job", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (h *ConversionHandler) ListUserJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	status := r.URL.Query().Get("status")
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	jobs, err := h.conversionService.GetUserJobs(currentUser.ID, statusPtr, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get jobs", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"jobs":   jobs,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ConversionHandler) StartJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hasPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionManageUsers)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	jobIDStr := strings.TrimPrefix(r.URL.Path, "/api/conversion/jobs/")
	jobIDStr = strings.TrimSuffix(jobIDStr, "/start")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	err = h.conversionService.StartConversion(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Job started successfully"})
}

func (h *ConversionHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	jobIDStr := strings.TrimPrefix(r.URL.Path, "/api/conversion/jobs/")
	jobIDStr = strings.TrimSuffix(jobIDStr, "/cancel")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	err = h.conversionService.CancelJob(jobID, currentUser.ID)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Job cancelled successfully"})
}

func (h *ConversionHandler) RetryJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	jobIDStr := strings.TrimPrefix(r.URL.Path, "/api/conversion/jobs/")
	jobIDStr = strings.TrimSuffix(jobIDStr, "/retry")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	err = h.conversionService.RetryJob(jobID, currentUser.ID)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Job restarted successfully"})
}

func (h *ConversionHandler) GetSupportedFormats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	formats := h.conversionService.GetSupportedFormats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(formats)
}

func (h *ConversionHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	userIDStr := r.URL.Query().Get("user_id")

	var startDate, endDate time.Time
	var userID *int

	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		} else {
			http.Error(w, "Invalid start_date format", http.StatusBadRequest)
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0) // Default to last month
	}

	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		} else {
			http.Error(w, "Invalid end_date format", http.StatusBadRequest)
			return
		}
	} else {
		endDate = time.Now()
	}

	if userIDStr != "" {
		if parsed, err := strconv.Atoi(userIDStr); err == nil {
			if parsed != currentUser.ID {
				hasPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionViewAnalytics)
				if err != nil || !hasPermission {
					http.Error(w, "Insufficient permissions", http.StatusForbidden)
					return
				}
			}
			userID = &parsed
		} else {
			http.Error(w, "Invalid user_id format", http.StatusBadRequest)
			return
		}
	} else {
		userID = &currentUser.ID
	}

	stats, err := h.conversionService.GetJobStatistics(userID, startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to get statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *ConversionHandler) ProcessQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hasPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionSystemAdmin)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	err = h.conversionService.ProcessJobQueue()
	if err != nil {
		http.Error(w, "Failed to process queue", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Queue processing started"})
}

func (h *ConversionHandler) GetQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hasPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionViewAnalytics)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	jobs, err := h.conversionService.GetJobQueue()
	if err != nil {
		http.Error(w, "Failed to get queue", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"queue": jobs,
		"count": len(jobs),
	})
}

func (h *ConversionHandler) CleanupJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hasPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionSystemAdmin)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	daysStr := r.URL.Query().Get("days")
	days := 30 // Default to 30 days

	if daysStr != "" {
		if parsed, err := strconv.Atoi(daysStr); err == nil && parsed > 0 {
			days = parsed
		}
	}

	olderThan := time.Now().AddDate(0, 0, -days)

	err = h.conversionService.CleanupCompletedJobs(olderThan)
	if err != nil {
		http.Error(w, "Failed to cleanup jobs", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Cleanup completed successfully"})
}

func (h *ConversionHandler) getCurrentUser(r *http.Request) (*models.User, error) {
	token := extractToken(r)
	if token == "" {
		return nil, models.ErrUnauthorized
	}

	return h.authService.GetCurrentUser(token)
}