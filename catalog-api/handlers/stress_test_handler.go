package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"catalog-api/models"
	"catalog-api/services"

	"github.com/gorilla/mux"
)

type StressTestHandler struct {
	stressTestService *services.StressTestService
	permissionService *services.PermissionService
}

func NewStressTestHandler(stressTestService *services.StressTestService, permissionService *services.PermissionService) *StressTestHandler {
	return &StressTestHandler{
		stressTestService: stressTestService,
		permissionService: permissionService,
	}
}

func (h *StressTestHandler) CreateStressTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "stress_testing", "write") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var request models.CreateStressTestRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	test, err := h.stressTestService.CreateStressTest(userID, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(test)
}

func (h *StressTestHandler) GetStressTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	testID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	if !h.permissionService.HasPermission(userID, "stress_testing", "read") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	test, err := h.stressTestService.GetStressTest(testID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check if user owns the test or has admin permissions
	if test.CreatedBy != userID && !h.permissionService.HasPermission(userID, "stress_testing", "admin") {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(test)
}

func (h *StressTestHandler) UpdateStressTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	testID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	if !h.permissionService.HasPermission(userID, "stress_testing", "write") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var request models.UpdateStressTestRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	test, err := h.stressTestService.UpdateStressTest(testID, userID, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(test)
}

func (h *StressTestHandler) DeleteStressTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	testID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	if !h.permissionService.HasPermission(userID, "stress_testing", "write") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	err = h.stressTestService.DeleteStressTest(testID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *StressTestHandler) ListStressTests(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "stress_testing", "read") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // Default limit
	offset := 0 // Default offset

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	tests, err := h.stressTestService.GetStressTestsByUser(userID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tests":  tests,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *StressTestHandler) StartStressTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	testID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	if !h.permissionService.HasPermission(userID, "stress_testing", "execute") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	execution, err := h.stressTestService.StartStressTest(testID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(execution)
}

func (h *StressTestHandler) StopStressTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	executionID, err := strconv.Atoi(vars["execution_id"])
	if err != nil {
		http.Error(w, "Invalid execution ID", http.StatusBadRequest)
		return
	}

	if !h.permissionService.HasPermission(userID, "stress_testing", "execute") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	err = h.stressTestService.StopStressTest(executionID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *StressTestHandler) GetExecution(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	executionID, err := strconv.Atoi(vars["execution_id"])
	if err != nil {
		http.Error(w, "Invalid execution ID", http.StatusBadRequest)
		return
	}

	if !h.permissionService.HasPermission(userID, "stress_testing", "read") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	execution, err := h.stressTestService.GetExecution(executionID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(execution)
}

func (h *StressTestHandler) GetExecutions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)
	testID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	if !h.permissionService.HasPermission(userID, "stress_testing", "read") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	executions, err := h.stressTestService.GetExecutionsByTestID(testID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(executions)
}

func (h *StressTestHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "stress_testing", "read") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	stats, err := h.stressTestService.GetStatistics(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *StressTestHandler) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "stress_testing", "admin") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	metrics, err := h.stressTestService.GetSystemMetrics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *StressTestHandler) CleanupOldExecutions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "stress_testing", "admin") {
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

	err := h.stressTestService.CleanupOldExecutions(request.DaysOld)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Cleanup completed successfully",
	})
}

func (h *StressTestHandler) GetPresets(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "stress_testing", "read") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	presets := h.stressTestService.GetTestPresets()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(presets)
}

func (h *StressTestHandler) ValidateScenario(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	if !h.permissionService.HasPermission(userID, "stress_testing", "read") {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var scenario models.TestScenario
	if err := json.NewDecoder(r.Body).Decode(&scenario); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := h.stressTestService.ValidateScenario(&scenario)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}