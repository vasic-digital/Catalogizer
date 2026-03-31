package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"catalogizer/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// =====================================================================
// RecommendationHandler tests (54.5% / 55.6% coverage)
// =====================================================================

func TestRecommendationHandler_GetTrendingItems_WithQueryParams(t *testing.T) {
	handler := NewRecommendationHandler(nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/recommendations/trending?media_type=movie&limit=3&time_range=month", nil)

	handler.GetTrendingItems(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp TrendingResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "movie", resp.MediaType)
	assert.Equal(t, "month", resp.TimeRange)
	assert.GreaterOrEqual(t, len(resp.Items), 1)
}

func TestRecommendationHandler_GetTrendingItems_Defaults(t *testing.T) {
	handler := NewRecommendationHandler(nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/recommendations/trending", nil)

	handler.GetTrendingItems(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp TrendingResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "week", resp.TimeRange)
}

func TestRecommendationHandler_GetPersonalizedRecommendations(t *testing.T) {
	handler := NewRecommendationHandler(nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/recommendations/personalized/42?limit=5", nil)
	c.Params = gin.Params{{Key: "user_id", Value: "42"}}

	handler.GetPersonalizedRecommendations(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp PersonalizedResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, int64(42), resp.UserID)
	assert.GreaterOrEqual(t, len(resp.Items), 1)
}

func TestRecommendationHandler_GetPersonalizedRecommendations_BadUserID(t *testing.T) {
	handler := NewRecommendationHandler(nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/recommendations/personalized/abc", nil)
	c.Params = gin.Params{{Key: "user_id", Value: "abc"}}

	handler.GetPersonalizedRecommendations(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRecommendationHandler_GetSimilarItems_InvalidMediaID(t *testing.T) {
	handler := NewRecommendationHandler(nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/recommendations/similar/abc", nil)
	c.Params = gin.Params{{Key: "media_id", Value: "abc"}}

	handler.GetSimilarItems(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =====================================================================
// RoleHandler — UpdateRole tests (52.5% coverage)
// =====================================================================

// Mock role user service
type mockRoleUserService struct {
	createRoleFn func(role *models.Role) (int, error)
	getRoleFn    func(roleID int) (*models.Role, error)
	updateRoleFn func(role *models.Role) error
	deleteRoleFn func(roleID int) error
	listRolesFn  func() ([]models.Role, error)
}

func (m *mockRoleUserService) CreateRole(role *models.Role) (int, error) {
	if m.createRoleFn != nil {
		return m.createRoleFn(role)
	}
	return 1, nil
}

func (m *mockRoleUserService) GetRole(roleID int) (*models.Role, error) {
	if m.getRoleFn != nil {
		return m.getRoleFn(roleID)
	}
	return &models.Role{ID: roleID, Name: "test"}, nil
}

func (m *mockRoleUserService) UpdateRole(role *models.Role) error {
	if m.updateRoleFn != nil {
		return m.updateRoleFn(role)
	}
	return nil
}

func (m *mockRoleUserService) DeleteRole(roleID int) error {
	if m.deleteRoleFn != nil {
		return m.deleteRoleFn(roleID)
	}
	return nil
}

func (m *mockRoleUserService) ListRoles() ([]models.Role, error) {
	if m.listRolesFn != nil {
		return m.listRolesFn()
	}
	return []models.Role{{ID: 1, Name: "admin"}}, nil
}

// Mock role auth service
type mockRoleAuthService struct {
	checkPermissionFn func(userID int, permission string) (bool, error)
	getCurrentUserFn  func(token string) (*models.User, error)
}

func (m *mockRoleAuthService) CheckPermission(userID int, permission string) (bool, error) {
	if m.checkPermissionFn != nil {
		return m.checkPermissionFn(userID, permission)
	}
	return true, nil
}

func (m *mockRoleAuthService) GetCurrentUser(token string) (*models.User, error) {
	if m.getCurrentUserFn != nil {
		return m.getCurrentUserFn(token)
	}
	return &models.User{ID: 1, Username: "admin"}, nil
}

func newTestRoleHandler() (*RoleHandler, *mockRoleUserService, *mockRoleAuthService) {
	userRepo := &mockRoleUserService{}
	authService := &mockRoleAuthService{}
	return NewRoleHandler(userRepo, authService), userRepo, authService
}

func TestRoleHandler_UpdateRole_Success(t *testing.T) {
	handler, userRepo, _ := newTestRoleHandler()
	userRepo.getRoleFn = func(roleID int) (*models.Role, error) {
		desc := "desc"
		return &models.Role{ID: roleID, Name: "updated", Description: &desc}, nil
	}

	name := "updated"
	desc := "desc"
	body, _ := json.Marshal(models.UpdateRoleRequest{Name: &name, Description: &desc})

	req := httptest.NewRequest("PUT", "/api/roles/5", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.UpdateRole(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRoleHandler_UpdateRole_WrongMethod(t *testing.T) {
	handler, _, _ := newTestRoleHandler()

	req := httptest.NewRequest("POST", "/api/roles/5", nil)
	rr := httptest.NewRecorder()

	handler.UpdateRole(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestRoleHandler_UpdateRole_NoAuth(t *testing.T) {
	handler, _, authSvc := newTestRoleHandler()
	authSvc.getCurrentUserFn = func(token string) (*models.User, error) {
		return nil, errors.New("invalid token")
	}

	req := httptest.NewRequest("PUT", "/api/roles/5", nil)
	rr := httptest.NewRecorder()

	handler.UpdateRole(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRoleHandler_UpdateRole_Forbidden(t *testing.T) {
	handler, _, authSvc := newTestRoleHandler()
	authSvc.checkPermissionFn = func(userID int, permission string) (bool, error) {
		return false, nil
	}

	req := httptest.NewRequest("PUT", "/api/roles/5", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.UpdateRole(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRoleHandler_UpdateRole_InvalidID(t *testing.T) {
	handler, _, _ := newTestRoleHandler()

	req := httptest.NewRequest("PUT", "/api/roles/abc", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.UpdateRole(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRoleHandler_UpdateRole_BadJSON(t *testing.T) {
	handler, _, _ := newTestRoleHandler()

	req := httptest.NewRequest("PUT", "/api/roles/5", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.UpdateRole(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRoleHandler_UpdateRole_NotFound(t *testing.T) {
	handler, userRepo, _ := newTestRoleHandler()
	userRepo.updateRoleFn = func(role *models.Role) error {
		return errors.New("not found")
	}

	name := "test"
	body, _ := json.Marshal(models.UpdateRoleRequest{Name: &name})

	req := httptest.NewRequest("PUT", "/api/roles/5", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.UpdateRole(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestRoleHandler_UpdateRole_UniqueConflict(t *testing.T) {
	handler, userRepo, _ := newTestRoleHandler()
	userRepo.updateRoleFn = func(role *models.Role) error {
		return errors.New("UNIQUE constraint failed")
	}

	name := "existing"
	body, _ := json.Marshal(models.UpdateRoleRequest{Name: &name})

	req := httptest.NewRequest("PUT", "/api/roles/5", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.UpdateRole(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestRoleHandler_UpdateRole_InternalError(t *testing.T) {
	handler, userRepo, _ := newTestRoleHandler()
	userRepo.updateRoleFn = func(role *models.Role) error {
		return errors.New("database error")
	}

	name := "test"
	body, _ := json.Marshal(models.UpdateRoleRequest{Name: &name})

	req := httptest.NewRequest("PUT", "/api/roles/5", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.UpdateRole(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestRoleHandler_UpdateRole_GetUpdatedRoleFails(t *testing.T) {
	handler, userRepo, _ := newTestRoleHandler()
	callCount := 0
	userRepo.getRoleFn = func(roleID int) (*models.Role, error) {
		callCount++
		if callCount > 0 {
			return nil, errors.New("get failed")
		}
		return &models.Role{ID: roleID}, nil
	}

	name := "test"
	body, _ := json.Marshal(models.UpdateRoleRequest{Name: &name})

	req := httptest.NewRequest("PUT", "/api/roles/5", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.UpdateRole(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// =====================================================================
// RoleHandler — DeleteRole tests
// =====================================================================

func TestRoleHandler_DeleteRole_SuccessPath(t *testing.T) {
	handler, _, _ := newTestRoleHandler()

	req := httptest.NewRequest("DELETE", "/api/roles/5", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.DeleteRole(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestRoleHandler_DeleteRole_MethodNotAllowed(t *testing.T) {
	handler, _, _ := newTestRoleHandler()

	req := httptest.NewRequest("GET", "/api/roles/5", nil)
	rr := httptest.NewRecorder()

	handler.DeleteRole(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestRoleHandler_DeleteRole_Unauthorized(t *testing.T) {
	handler, _, authSvc := newTestRoleHandler()
	authSvc.getCurrentUserFn = func(token string) (*models.User, error) {
		return nil, errors.New("invalid")
	}

	req := httptest.NewRequest("DELETE", "/api/roles/5", nil)
	rr := httptest.NewRecorder()

	handler.DeleteRole(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRoleHandler_DeleteRole_Forbidden(t *testing.T) {
	handler, _, authSvc := newTestRoleHandler()
	authSvc.checkPermissionFn = func(userID int, permission string) (bool, error) {
		return false, nil
	}

	req := httptest.NewRequest("DELETE", "/api/roles/5", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.DeleteRole(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestRoleHandler_DeleteRole_InvalidID(t *testing.T) {
	handler, _, _ := newTestRoleHandler()

	req := httptest.NewRequest("DELETE", "/api/roles/abc", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.DeleteRole(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRoleHandler_DeleteRole_NotFound(t *testing.T) {
	handler, userRepo, _ := newTestRoleHandler()
	userRepo.deleteRoleFn = func(roleID int) error {
		return errors.New("not found")
	}

	req := httptest.NewRequest("DELETE", "/api/roles/5", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.DeleteRole(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestRoleHandler_DeleteRole_InUseByUsers(t *testing.T) {
	handler, userRepo, _ := newTestRoleHandler()
	userRepo.deleteRoleFn = func(roleID int) error {
		return errors.New("assigned to users")
	}

	req := httptest.NewRequest("DELETE", "/api/roles/5", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.DeleteRole(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestRoleHandler_DeleteRole_SystemRole(t *testing.T) {
	handler, userRepo, _ := newTestRoleHandler()
	userRepo.deleteRoleFn = func(roleID int) error {
		return errors.New("system role cannot be deleted")
	}

	req := httptest.NewRequest("DELETE", "/api/roles/5", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.DeleteRole(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestRoleHandler_DeleteRole_InternalError(t *testing.T) {
	handler, userRepo, _ := newTestRoleHandler()
	userRepo.deleteRoleFn = func(roleID int) error {
		return errors.New("database error")
	}

	req := httptest.NewRequest("DELETE", "/api/roles/5", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.DeleteRole(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// =====================================================================
// RoleHandler — GetPermissions
// =====================================================================

func TestRoleHandler_GetPermissions_Success(t *testing.T) {
	handler, _, _ := newTestRoleHandler()

	req := httptest.NewRequest("GET", "/api/roles/permissions", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.GetPermissions(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp, "user_management")
	assert.Contains(t, resp, "role_management")
	assert.Contains(t, resp, "media_management")
	assert.Contains(t, resp, "share_management")
	assert.Contains(t, resp, "system")
}

func TestRoleHandler_GetPermissions_MethodNotAllowed(t *testing.T) {
	handler, _, _ := newTestRoleHandler()

	req := httptest.NewRequest("POST", "/api/roles/permissions", nil)
	rr := httptest.NewRecorder()

	handler.GetPermissions(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestRoleHandler_GetPermissions_Unauthorized(t *testing.T) {
	handler, _, authSvc := newTestRoleHandler()
	authSvc.getCurrentUserFn = func(token string) (*models.User, error) {
		return nil, errors.New("invalid")
	}

	req := httptest.NewRequest("GET", "/api/roles/permissions", nil)
	rr := httptest.NewRecorder()

	handler.GetPermissions(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRoleHandler_GetPermissions_PermissionError(t *testing.T) {
	handler, _, authSvc := newTestRoleHandler()
	authSvc.checkPermissionFn = func(userID int, permission string) (bool, error) {
		return false, errors.New("check failed")
	}

	req := httptest.NewRequest("GET", "/api/roles/permissions", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.GetPermissions(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// =====================================================================
// RoleHandler — ListRoles
// =====================================================================

func TestRoleHandler_ListRoles_Success(t *testing.T) {
	handler, _, _ := newTestRoleHandler()

	req := httptest.NewRequest("GET", "/api/roles", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.ListRoles(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRoleHandler_ListRoles_Error(t *testing.T) {
	handler, userRepo, _ := newTestRoleHandler()
	userRepo.listRolesFn = func() ([]models.Role, error) {
		return nil, errors.New("db error")
	}

	req := httptest.NewRequest("GET", "/api/roles", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.ListRoles(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestRoleHandler_ListRoles_MethodNotAllowed(t *testing.T) {
	handler, _, _ := newTestRoleHandler()

	req := httptest.NewRequest("POST", "/api/roles", nil)
	rr := httptest.NewRecorder()

	handler.ListRoles(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

// =====================================================================
// RoleHandler — CreateRole error paths
// =====================================================================

func TestRoleHandler_CreateRole_UniqueConflict(t *testing.T) {
	handler, userRepo, _ := newTestRoleHandler()
	userRepo.createRoleFn = func(role *models.Role) (int, error) {
		return 0, errors.New("UNIQUE constraint failed")
	}

	body, _ := json.Marshal(models.CreateRoleRequest{Name: "existing"})

	req := httptest.NewRequest("POST", "/api/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.CreateRole(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestRoleHandler_CreateRole_InternalError(t *testing.T) {
	handler, userRepo, _ := newTestRoleHandler()
	userRepo.createRoleFn = func(role *models.Role) (int, error) {
		return 0, errors.New("database error")
	}

	body, _ := json.Marshal(models.CreateRoleRequest{Name: "test"})

	req := httptest.NewRequest("POST", "/api/roles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.CreateRole(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestRoleHandler_CreateRole_PermCheckError(t *testing.T) {
	handler, _, authSvc := newTestRoleHandler()
	authSvc.checkPermissionFn = func(userID int, permission string) (bool, error) {
		return false, errors.New("check failed")
	}

	req := httptest.NewRequest("POST", "/api/roles", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.CreateRole(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// =====================================================================
// RoleHandler — GetRole error paths
// =====================================================================

func TestRoleHandler_GetRole_NotFoundError(t *testing.T) {
	handler, userRepo, _ := newTestRoleHandler()
	userRepo.getRoleFn = func(roleID int) (*models.Role, error) {
		return nil, errors.New("not found")
	}

	req := httptest.NewRequest("GET", "/api/roles/999", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.GetRole(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestRoleHandler_GetRole_InternalError(t *testing.T) {
	handler, userRepo, _ := newTestRoleHandler()
	userRepo.getRoleFn = func(roleID int) (*models.Role, error) {
		return nil, errors.New("database error")
	}

	req := httptest.NewRequest("GET", "/api/roles/5", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.GetRole(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestRoleHandler_GetRole_PermissionCheckError(t *testing.T) {
	handler, _, authSvc := newTestRoleHandler()
	authSvc.checkPermissionFn = func(userID int, permission string) (bool, error) {
		return false, errors.New("check failed")
	}

	req := httptest.NewRequest("GET", "/api/roles/5", nil)
	req.Header.Set("Authorization", "Bearer token")
	rr := httptest.NewRecorder()

	handler.GetRole(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// =====================================================================
// LogManagementHandler — StreamLogs with origin header
// =====================================================================

// flushableRecorder wraps httptest.ResponseRecorder to implement http.Flusher
type flushableRecorder struct {
	*httptest.ResponseRecorder
}

func (f *flushableRecorder) Flush() {
	// no-op for testing
}

func TestLogManagementHandler_StreamLogs_WithOriginHeader(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)

	ch := make(chan models.LogEntry, 1)
	ch <- models.LogEntry{ID: 1, Level: "info", Message: "test log"}
	close(ch)

	// Mock only returns (channel, error) — the done channel is created internally by the mock
	mockLogService.On("StreamLogs", 1, mock.AnythingOfType("*models.LogStreamFilters")).Return((<-chan models.LogEntry)(ch), nil)

	req := httptest.NewRequest("GET", "/stream-logs", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := &flushableRecorder{httptest.NewRecorder()}

	handler.StreamLogs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/event-stream", rr.Header().Get("Content-Type"))
	assert.Contains(t, rr.Body.String(), "data:")
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_StreamLogs_WithDisallowedOrigin(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)

	ch := make(chan models.LogEntry)
	close(ch)

	mockLogService.On("StreamLogs", 1, mock.AnythingOfType("*models.LogStreamFilters")).Return((<-chan models.LogEntry)(ch), nil)

	req := httptest.NewRequest("GET", "/stream-logs", nil)
	req.Header.Set("Origin", "http://evil.com")
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := &flushableRecorder{httptest.NewRecorder()}

	handler.StreamLogs(rr, req)

	// Origin header not matched, Access-Control-Allow-Origin should NOT be set
	assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
	mockLogService.AssertExpectations(t)
}

// =====================================================================
// LogManagementHandler — additional error paths
// =====================================================================

func TestLogManagementHandler_AnalyzeLogs_ServiceError(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("AnalyzeLogs", 1, 1).Return(nil, errors.New("analysis failed"))

	req := httptest.NewRequest("GET", "/log-collections/1/analyze", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.AnalyzeLogs(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_GetLogStatistics_ServiceError(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("GetLogStatistics", 1).Return(nil, errors.New("stats failed"))

	req := httptest.NewRequest("GET", "/log-statistics", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.GetLogStatistics(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_UpdateConfiguration_ServiceError(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("UpdateConfiguration", mock.AnythingOfType("*services.LogManagementConfig")).Return(errors.New("update failed"))

	body := `{"max_log_size": 1000}`
	req := httptest.NewRequest("PUT", "/log-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.UpdateConfiguration(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_ExportLogs_ServiceError(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("ExportLogs", 1, 1, "json").Return(nil, errors.New("export failed"))

	req := httptest.NewRequest("GET", "/log-collections/1/export?format=json", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.ExportLogs(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_CreateLogShare_ServiceError(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)

	body, _ := json.Marshal(models.LogShareRequest{CollectionID: 1})
	var shareReq models.LogShareRequest
	_ = json.Unmarshal(body, &shareReq)
	mockLogService.On("CreateLogShare", 1, &shareReq).Return(nil, errors.New("share failed"))

	httpReq := httptest.NewRequest("POST", "/log-shares", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq = httpReq.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.CreateLogShare(rr, httpReq)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_RevokeLogShare_ServiceError(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("RevokeLogShare", 5, 1).Return(errors.New("revoke failed"))

	req := httptest.NewRequest("DELETE", "/log-shares/5", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	req = mux.SetURLVars(req, map[string]string{"id": "5"})
	rr := httptest.NewRecorder()

	handler.RevokeLogShare(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_GetLogCollection_ServiceError(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("GetLogCollection", 1, 1).Return(nil, errors.New("not found"))

	req := httptest.NewRequest("GET", "/log-collections/1", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.GetLogCollection(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_ListLogCollections_ServiceError(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("GetLogCollectionsByUser", 1, 20, 0).Return(nil, errors.New("list failed"))

	req := httptest.NewRequest("GET", "/log-collections", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.ListLogCollections(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_GetLogEntries_ServiceError(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	mockLogService.On("GetLogEntries", 1, 1, mock.AnythingOfType("*models.LogEntryFilters")).Return(nil, errors.New("entries failed"))

	req := httptest.NewRequest("GET", "/log-collections/1/entries", nil)
	req = req.WithContext(context.WithValue(context.Background(), "user_id", 1))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.GetLogEntries(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockLogService.AssertExpectations(t)
}

func TestLogManagementHandler_CreateLogCollection_ServiceError(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	mockAuthService.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)

	body, _ := json.Marshal(models.LogCollectionRequest{Name: "Test"})
	var collReq models.LogCollectionRequest
	_ = json.Unmarshal(body, &collReq)
	mockLogService.On("CollectLogs", 1, &collReq).Return(nil, errors.New("collect failed"))

	httpReq := httptest.NewRequest("POST", "/log-collections", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq = httpReq.WithContext(context.WithValue(context.Background(), "user_id", 1))
	rr := httptest.NewRecorder()

	handler.CreateLogCollection(rr, httpReq)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockLogService.AssertExpectations(t)
}

// =====================================================================
// LogManagementHandler — GetLogShare collection error
// =====================================================================

func TestLogManagementHandler_GetLogShare_CollectionError(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := &LogManagementHandler{
		logManagementService: mockLogService,
		authService:          mockAuthService,
	}

	share := &models.LogShare{
		ID:           1,
		ShareToken:   "abc123",
		CollectionID: 1,
		UserID:       1,
		Permissions:  []string{"read"},
	}

	mockLogService.On("GetLogShare", "abc123").Return(share, nil)
	mockLogService.On("GetLogCollection", 1, 1).Return(nil, errors.New("collection not found"))

	req := httptest.NewRequest("GET", "/log-shares/abc123", nil)
	req = mux.SetURLVars(req, map[string]string{"token": "abc123"})
	rr := httptest.NewRecorder()

	handler.GetLogShare(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockLogService.AssertExpectations(t)
}
