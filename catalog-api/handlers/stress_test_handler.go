package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"catalogizer/models"
	"catalogizer/services"
)

type StressTestHandler struct {
	stressTestService *services.StressTestService
	authService       *services.AuthService
}

func NewStressTestHandler(stressTestService *services.StressTestService, authService *services.AuthService) *StressTestHandler {
	return &StressTestHandler{
		stressTestService: stressTestService,
		authService:       authService,
	}
}

// CreateStressTest creates a new stress test
func (h *StressTestHandler) CreateStressTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionSystemAdmin)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var test models.StressTest
	if err := json.NewDecoder(r.Body).Decode(&test); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdTest, err := h.stressTestService.CreateStressTest(userID, &test)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTest)
}

// GetStressTest gets a stress test by ID
func (h *StressTestHandler) GetStressTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	testID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	tests, err := h.stressTestService.ListUserTests(userID, 1000, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, test := range tests {
		if int(test.ID) == testID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(test)
			return
		}
	}

	http.Error(w, "Test not found", http.StatusNotFound)
}

// UpdateStressTest updates a stress test
func (h *StressTestHandler) UpdateStressTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionSystemAdmin)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Update stress test not implemented - tests cannot be modified after creation"})
}

// DeleteStressTest deletes a stress test
func (h *StressTestHandler) DeleteStressTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionSystemAdmin)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	testID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	err = h.stressTestService.DeleteTest(testID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListStressTests lists stress tests for a user
func (h *StressTestHandler) ListStressTests(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	limit := 50
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	tests, err := h.stressTestService.ListUserTests(userID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tests": tests,
		"count": len(tests),
	})
}

// StartStressTest starts a stress test
func (h *StressTestHandler) StartStressTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionSystemAdmin)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	testID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	result, err := h.stressTestService.StartStressTest(testID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// StopStressTest stops a stress test
func (h *StressTestHandler) StopStressTest(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionSystemAdmin)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	testID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	err = h.stressTestService.StopStressTest(testID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Test stopped successfully"})
}

// GetStressTestStatus gets the status of a stress test
func (h *StressTestHandler) GetStressTestStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	testID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	status, err := h.stressTestService.GetTestStatus(testID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetStressTestResults gets the results of a stress test
func (h *StressTestHandler) GetStressTestResults(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	testID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	results, err := h.stressTestService.GetTestResults(testID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GetStressTestExecution gets a specific execution
func (h *StressTestHandler) GetStressTestExecution(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Get execution requires direct repository access - use GetStressTestResults instead"})
}

// GetStressTestExecutions gets executions for a test
func (h *StressTestHandler) GetStressTestExecutions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Get executions requires direct repository access - use GetStressTestResults instead"})
}

// GetStressTestStatistics gets statistics for a test
func (h *StressTestHandler) GetStressTestStatistics(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Statistics requires direct repository access"})
}

// GetSystemMetrics gets system metrics
func (h *StressTestHandler) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	metrics, err := h.stressTestService.GetSystemLoad()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// CleanupOldExecutions cleans up old executions
func (h *StressTestHandler) CleanupOldExecutions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionSystemAdmin)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Cleanup requires direct repository access"})
}

// GetTestPresets gets test presets
func (h *StressTestHandler) GetTestPresets(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	presets := []map[string]interface{}{
		{
			"name":             "Light Load Test",
			"description":      "Basic load test with 10 concurrent users for 60 seconds",
			"concurrent_users": 10,
			"duration":         60,
			"request_delay":    100,
		},
		{
			"name":             "Medium Load Test",
			"description":      "Medium load test with 50 concurrent users for 120 seconds",
			"concurrent_users": 50,
			"duration":         120,
			"request_delay":    50,
		},
		{
			"name":             "Heavy Load Test",
			"description":      "Heavy load test with 200 concurrent users for 300 seconds",
			"concurrent_users": 200,
			"duration":         300,
			"request_delay":    10,
		},
		{
			"name":             "Stress Test",
			"description":      "Maximum stress test with 500 concurrent users for 600 seconds",
			"concurrent_users": 500,
			"duration":         600,
			"request_delay":    0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"presets": presets,
		"count":   len(presets),
	})
}

// ValidateScenario validates a test scenario
func (h *StressTestHandler) ValidateScenario(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var scenario models.StressTestScenario
	if err := json.NewDecoder(r.Body).Decode(&scenario); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errors := []string{}

	if scenario.URL == "" {
		errors = append(errors, "URL is required")
	}

	if scenario.Method == "" {
		errors = append(errors, "Method is required")
	}

	validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true, "HEAD": true, "OPTIONS": true}
	if scenario.Method != "" && !validMethods[scenario.Method] {
		errors = append(errors, "Invalid HTTP method")
	}

	if scenario.Weight < 0 {
		errors = append(errors, "Weight cannot be negative")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":  len(errors) == 0,
		"errors": errors,
	})
}

// GenerateReport generates a load test report
func (h *StressTestHandler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	hasPermission, err := h.authService.CheckPermission(userID, models.PermissionViewAnalytics)
	if err != nil || !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	testID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid test ID", http.StatusBadRequest)
		return
	}

	report, err := h.stressTestService.GenerateLoadReport(testID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
