package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"catalogizer/models"
	"catalogizer/services"
	internalservices "catalogizer/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// =============================================================================
// SubtitleHandler tests
// =============================================================================

func newSubtitleHandler() *SubtitleHandler {
	logger := zap.NewNop()
	svc := internalservices.NewSubtitleService(nil, logger, nil)
	return NewSubtitleHandler(svc, logger)
}

func TestNewSubtitleHandler(t *testing.T) {
	h := newSubtitleHandler()
	assert.NotNil(t, h)
	assert.NotNil(t, h.subtitleService)
	assert.NotNil(t, h.logger)
}

func TestSubtitleHandler_SearchSubtitles_MissingMediaPath(t *testing.T) {
	h := newSubtitleHandler()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/subtitles/search", h.SearchSubtitles)

	req := httptest.NewRequest(http.MethodGet, "/subtitles/search", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "media_path is required")
}

// SearchSubtitles tests with valid media_path that reach service layer are skipped - nil DB panics

func TestSubtitleHandler_GetSubtitles_MissingMediaID(t *testing.T) {
	h := newSubtitleHandler()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/subtitles/media/:media_id", h.GetSubtitles)

	// When the param is an invalid number
	req := httptest.NewRequest(http.MethodGet, "/subtitles/media/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid media_id")
}

// TestSubtitleHandler_GetSubtitles_ValidMediaID skipped - nil DB causes panic in service layer

func TestSubtitleHandler_VerifySubtitleSync_InvalidMediaID(t *testing.T) {
	h := newSubtitleHandler()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/subtitles/:subtitle_id/verify-sync/:media_id", h.VerifySubtitleSync)

	req := httptest.NewRequest(http.MethodGet, "/subtitles/sub-1/verify-sync/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid media_id")
}

// TestSubtitleHandler_VerifySubtitleSync_ValidParams skipped - nil DB causes panic in service layer

func TestSubtitleHandler_DownloadSubtitle_InvalidJSON(t *testing.T) {
	h := newSubtitleHandler()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/subtitles/download", h.DownloadSubtitle)

	req := httptest.NewRequest(http.MethodPost, "/subtitles/download", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubtitleHandler_DownloadSubtitle_MissingRequiredFields(t *testing.T) {
	h := newSubtitleHandler()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/subtitles/download", h.DownloadSubtitle)

	body := `{"media_item_id": 0, "result_id": "", "language": ""}`
	req := httptest.NewRequest(http.MethodPost, "/subtitles/download", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "required")
}

// TestSubtitleHandler_DownloadSubtitle_ValidRequest skipped - nil DB causes panic

func TestSubtitleHandler_TranslateSubtitle_InvalidJSON(t *testing.T) {
	h := newSubtitleHandler()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/subtitles/translate", h.TranslateSubtitle)

	req := httptest.NewRequest(http.MethodPost, "/subtitles/translate", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubtitleHandler_TranslateSubtitle_MissingFields(t *testing.T) {
	h := newSubtitleHandler()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/subtitles/translate", h.TranslateSubtitle)

	body := `{"subtitle_id": "", "source_language": "", "target_language": ""}`
	req := httptest.NewRequest(http.MethodPost, "/subtitles/translate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "required")
}

// TestSubtitleHandler_TranslateSubtitle_ValidRequest skipped - nil DB causes panic

func TestSubtitleHandler_UploadSubtitle_MissingFields(t *testing.T) {
	h := newSubtitleHandler()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/subtitles/upload", h.UploadSubtitle)

	req := httptest.NewRequest(http.MethodPost, "/subtitles/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "required")
}

func TestSubtitleHandler_UploadSubtitle_InvalidMediaID(t *testing.T) {
	h := newSubtitleHandler()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/subtitles/upload", h.UploadSubtitle)

	// Use multipart form with invalid media_item_id
	body := &bytes.Buffer{}
	body.WriteString("--boundary\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"media_item_id\"\r\n\r\nabc\r\n")
	body.WriteString("--boundary\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"language\"\r\n\r\nEnglish\r\n")
	body.WriteString("--boundary\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"language_code\"\r\n\r\nen\r\n")
	body.WriteString("--boundary--\r\n")

	req := httptest.NewRequest(http.MethodPost, "/subtitles/upload", body)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid media_item_id")
}

func TestSubtitleHandler_UploadSubtitle_MissingFile(t *testing.T) {
	h := newSubtitleHandler()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/subtitles/upload", h.UploadSubtitle)

	body := &bytes.Buffer{}
	body.WriteString("--boundary\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"media_item_id\"\r\n\r\n123\r\n")
	body.WriteString("--boundary\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"language\"\r\n\r\nEnglish\r\n")
	body.WriteString("--boundary\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"language_code\"\r\n\r\nen\r\n")
	body.WriteString("--boundary--\r\n")

	req := httptest.NewRequest(http.MethodPost, "/subtitles/upload", body)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "file is required")
}

// Upload tests with valid files that reach service layer are skipped - nil DB causes panic

func TestSubtitleHandler_GetSupportedLanguages(t *testing.T) {
	h := newSubtitleHandler()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/subtitles/languages", h.GetSupportedLanguages)

	req := httptest.NewRequest(http.MethodGet, "/subtitles/languages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp LanguageListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Greater(t, resp.Count, 0)
	assert.Equal(t, resp.Count, len(resp.Languages))
}

func TestSubtitleHandler_GetSupportedProviders(t *testing.T) {
	h := newSubtitleHandler()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/subtitles/providers", h.GetSupportedProviders)

	req := httptest.NewRequest(http.MethodGet, "/subtitles/providers", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ProviderListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, 5, resp.Count)
}

func TestGetProviderDisplayName(t *testing.T) {
	tests := []struct {
		provider internalservices.SubtitleProvider
		expected string
	}{
		{internalservices.ProviderOpenSubtitles, "OpenSubtitles"},
		{internalservices.ProviderSubDB, "SubDB"},
		{internalservices.ProviderYifySubtitles, "YIFY Subtitles"},
		{internalservices.ProviderSubscene, "Subscene"},
		{internalservices.ProviderAddic7ed, "Addic7ed"},
		{internalservices.SubtitleProvider("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			result := getProviderDisplayName(tt.provider)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetProviderDescription(t *testing.T) {
	tests := []struct {
		provider internalservices.SubtitleProvider
		hasDesc  bool
	}{
		{internalservices.ProviderOpenSubtitles, true},
		{internalservices.ProviderSubDB, true},
		{internalservices.ProviderYifySubtitles, true},
		{internalservices.ProviderSubscene, true},
		{internalservices.ProviderAddic7ed, true},
		{internalservices.SubtitleProvider("unknown"), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			result := getProviderDescription(tt.provider)
			assert.NotEmpty(t, result)
		})
	}
}

// =============================================================================
// ErrorReportingHandler tests (mock-based, many functions at 0%)
// =============================================================================

// Mock for ErrorReportingServiceInterface
type mockErrorReportingService struct {
	mock.Mock
}

func (m *mockErrorReportingService) ReportError(userID int, request *models.ErrorReportRequest) (*models.ErrorReport, error) {
	args := m.Called(userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ErrorReport), args.Error(1)
}

func (m *mockErrorReportingService) ReportCrash(userID int, request *models.CrashReportRequest) (*models.CrashReport, error) {
	args := m.Called(userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CrashReport), args.Error(1)
}

func (m *mockErrorReportingService) GetErrorReport(reportID int, userID int) (*models.ErrorReport, error) {
	args := m.Called(reportID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ErrorReport), args.Error(1)
}

func (m *mockErrorReportingService) GetCrashReport(reportID int, userID int) (*models.CrashReport, error) {
	args := m.Called(reportID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CrashReport), args.Error(1)
}

func (m *mockErrorReportingService) GetErrorReportsByUser(userID int, filters *models.ErrorReportFilters) ([]models.ErrorReport, error) {
	args := m.Called(userID, filters)
	return args.Get(0).([]models.ErrorReport), args.Error(1)
}

func (m *mockErrorReportingService) GetCrashReportsByUser(userID int, filters *models.CrashReportFilters) ([]models.CrashReport, error) {
	args := m.Called(userID, filters)
	return args.Get(0).([]models.CrashReport), args.Error(1)
}

func (m *mockErrorReportingService) UpdateErrorStatus(reportID int, userID int, status string) error {
	args := m.Called(reportID, userID, status)
	return args.Error(0)
}

func (m *mockErrorReportingService) UpdateCrashStatus(reportID int, userID int, status string) error {
	args := m.Called(reportID, userID, status)
	return args.Error(0)
}

func (m *mockErrorReportingService) GetErrorStatistics(userID int) (*models.ErrorStatistics, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ErrorStatistics), args.Error(1)
}

func (m *mockErrorReportingService) GetCrashStatistics(userID int) (*models.CrashStatistics, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CrashStatistics), args.Error(1)
}

func (m *mockErrorReportingService) GetSystemHealth() (*models.SystemHealth, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SystemHealth), args.Error(1)
}

func (m *mockErrorReportingService) UpdateConfiguration(config *services.ErrorReportingConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *mockErrorReportingService) GetConfiguration() (*services.ErrorReportingConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ErrorReportingConfig), args.Error(1)
}

func (m *mockErrorReportingService) CleanupOldReports(olderThan time.Time) error {
	args := m.Called(olderThan)
	return args.Error(0)
}

func (m *mockErrorReportingService) ExportReports(userID int, filters *models.ExportFilters) ([]byte, error) {
	args := m.Called(userID, filters)
	return args.Get(0).([]byte), args.Error(1)
}

// Mock for ErrorReportingAuthServiceInterface
type mockERAuthService struct {
	mock.Mock
}

func (m *mockERAuthService) CheckPermission(userID int, permission string) (bool, error) {
	args := m.Called(userID, permission)
	return args.Bool(0), args.Error(1)
}

// Helper to create request with user_id context
func reqWithUserID(method, url string, body *bytes.Buffer, userID int) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, body)
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	ctx := context.WithValue(req.Context(), "user_id", userID)
	return req.WithContext(ctx)
}

func TestNewErrorReportingHandler(t *testing.T) {
	h := NewErrorReportingHandler(nil, nil)
	assert.NotNil(t, h)
}

func TestErrorReportingHandler_ReportError_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportCreate).Return(true, nil)
	report := &models.ErrorReport{ID: 1, Message: "test error"}
	svc.On("ReportError", 1, mock.Anything).Return(report, nil)

	body := bytes.NewBufferString(`{"message": "test error", "level": "error"}`)
	req := reqWithUserID(http.MethodPost, "/errors", body, 1)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ReportError(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorReportingHandler_ReportError_Forbidden(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportCreate).Return(false, nil)

	body := bytes.NewBufferString(`{"message": "test"}`)
	req := reqWithUserID(http.MethodPost, "/errors", body, 1)
	w := httptest.NewRecorder()

	h.ReportError(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestErrorReportingHandler_ReportError_InvalidBody(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportCreate).Return(true, nil)

	body := bytes.NewBufferString("invalid json")
	req := reqWithUserID(http.MethodPost, "/errors", body, 1)
	w := httptest.NewRecorder()

	h.ReportError(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestErrorReportingHandler_ReportCrash_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportCreate).Return(true, nil)
	report := &models.CrashReport{ID: 1}
	svc.On("ReportCrash", 1, mock.Anything).Return(report, nil)

	body := bytes.NewBufferString(`{"signal": "SIGSEGV"}`)
	req := reqWithUserID(http.MethodPost, "/crashes", body, 1)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ReportCrash(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorReportingHandler_GetErrorReport_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportView).Return(true, nil)
	report := &models.ErrorReport{ID: 1, Message: "test"}
	svc.On("GetErrorReport", 1, 1).Return(report, nil)

	req := reqWithUserID(http.MethodGet, "/errors/1", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.GetErrorReport(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorReportingHandler_GetErrorReport_InvalidID(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	req := reqWithUserID(http.MethodGet, "/errors/abc", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	h.GetErrorReport(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestErrorReportingHandler_GetCrashReport_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportView).Return(true, nil)
	report := &models.CrashReport{ID: 1}
	svc.On("GetCrashReport", 1, 1).Return(report, nil)

	req := reqWithUserID(http.MethodGet, "/crashes/1", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.GetCrashReport(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorReportingHandler_ListErrorReports_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportView).Return(true, nil)
	svc.On("GetErrorReportsByUser", 1, mock.Anything).Return([]models.ErrorReport{}, nil)

	req := reqWithUserID(http.MethodGet, "/errors?level=error&component=api&status=open&start_date=2024-01-01&end_date=2024-12-31&limit=10&offset=0", nil, 1)
	w := httptest.NewRecorder()

	h.ListErrorReports(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorReportingHandler_ListErrorReports_Forbidden(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportView).Return(false, nil)

	req := reqWithUserID(http.MethodGet, "/errors", nil, 1)
	w := httptest.NewRecorder()

	h.ListErrorReports(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestErrorReportingHandler_ListCrashReports_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportView).Return(true, nil)
	svc.On("GetCrashReportsByUser", 1, mock.Anything).Return([]models.CrashReport{}, nil)

	req := reqWithUserID(http.MethodGet, "/crashes?signal=SIGSEGV&status=open&start_date=2024-01-01&end_date=2024-12-31&limit=10&offset=5", nil, 1)
	w := httptest.NewRecorder()

	h.ListCrashReports(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorReportingHandler_UpdateErrorStatus_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportCreate).Return(true, nil)
	svc.On("UpdateErrorStatus", 1, 1, "resolved").Return(nil)

	body := bytes.NewBufferString(`{"status": "resolved"}`)
	req := reqWithUserID(http.MethodPut, "/errors/1/status", body, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.UpdateErrorStatus(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestErrorReportingHandler_UpdateErrorStatus_InvalidID(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	req := reqWithUserID(http.MethodPut, "/errors/abc/status", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	h.UpdateErrorStatus(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestErrorReportingHandler_UpdateCrashStatus_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportCreate).Return(true, nil)
	svc.On("UpdateCrashStatus", 1, 1, "resolved").Return(nil)

	body := bytes.NewBufferString(`{"status": "resolved"}`)
	req := reqWithUserID(http.MethodPut, "/crashes/1/status", body, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.UpdateCrashStatus(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestErrorReportingHandler_GetErrorStatistics_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportView).Return(true, nil)
	stats := &models.ErrorStatistics{TotalErrors: 5}
	svc.On("GetErrorStatistics", 1).Return(stats, nil)

	req := reqWithUserID(http.MethodGet, "/errors/stats", nil, 1)
	w := httptest.NewRecorder()

	h.GetErrorStatistics(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorReportingHandler_GetCrashStatistics_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportView).Return(true, nil)
	stats := &models.CrashStatistics{TotalCrashes: 3}
	svc.On("GetCrashStatistics", 1).Return(stats, nil)

	req := reqWithUserID(http.MethodGet, "/crashes/stats", nil, 1)
	w := httptest.NewRecorder()

	h.GetCrashStatistics(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorReportingHandler_GetSystemHealth_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	health := &models.SystemHealth{Status: "healthy"}
	svc.On("GetSystemHealth").Return(health, nil)

	req := reqWithUserID(http.MethodGet, "/health", nil, 1)
	w := httptest.NewRecorder()

	h.GetSystemHealth(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorReportingHandler_GetSystemHealth_Forbidden(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(false, nil)

	req := reqWithUserID(http.MethodGet, "/health", nil, 1)
	w := httptest.NewRecorder()

	h.GetSystemHealth(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestErrorReportingHandler_UpdateConfiguration_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("UpdateConfiguration", mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"max_reports": 100}`)
	req := reqWithUserID(http.MethodPut, "/config", body, 1)
	w := httptest.NewRecorder()

	h.UpdateConfiguration(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorReportingHandler_UpdateConfiguration_InvalidBody(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)

	body := bytes.NewBufferString("invalid json")
	req := reqWithUserID(http.MethodPut, "/config", body, 1)
	w := httptest.NewRecorder()

	h.UpdateConfiguration(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestErrorReportingHandler_GetConfiguration_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	config := &services.ErrorReportingConfig{}
	svc.On("GetConfiguration").Return(config, nil)

	req := reqWithUserID(http.MethodGet, "/config", nil, 1)
	w := httptest.NewRecorder()

	h.GetConfiguration(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorReportingHandler_CleanupOldReports_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("CleanupOldReports", mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"days_old": 30}`)
	req := reqWithUserID(http.MethodPost, "/cleanup", body, 1)
	w := httptest.NewRecorder()

	h.CleanupOldReports(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorReportingHandler_CleanupOldReports_DefaultDays(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("CleanupOldReports", mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"days_old": 0}`)
	req := reqWithUserID(http.MethodPost, "/cleanup", body, 1)
	w := httptest.NewRecorder()

	h.CleanupOldReports(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorReportingHandler_ExportReports_Success(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportView).Return(true, nil)
	svc.On("ExportReports", 1, mock.Anything).Return([]byte(`{"reports":[]}`), nil)

	req := reqWithUserID(http.MethodGet, "/export?format=json&level=error&component=api&signal=SIGSEGV&start_date=2024-01-01&end_date=2024-12-31&limit=100&include_errors=true&include_crashes=false", nil, 1)
	w := httptest.NewRecorder()

	h.ExportReports(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestErrorReportingHandler_ExportReports_CSV(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportView).Return(true, nil)
	svc.On("ExportReports", 1, mock.Anything).Return([]byte("col1,col2\n"), nil)

	req := reqWithUserID(http.MethodGet, "/export?format=csv", nil, 1)
	w := httptest.NewRecorder()

	h.ExportReports(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
}

func TestErrorReportingHandler_ExportReports_Unknown(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportView).Return(true, nil)
	svc.On("ExportReports", 1, mock.Anything).Return([]byte("data"), nil)

	req := reqWithUserID(http.MethodGet, "/export?format=xml", nil, 1)
	w := httptest.NewRecorder()

	h.ExportReports(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
}

func TestErrorReportingHandler_ExportReports_Error(t *testing.T) {
	svc := new(mockErrorReportingService)
	auth := new(mockERAuthService)
	h := NewErrorReportingHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionReportView).Return(true, nil)
	svc.On("ExportReports", 1, mock.Anything).Return([]byte{}, errors.New("export failed"))

	req := reqWithUserID(http.MethodGet, "/export", nil, 1)
	w := httptest.NewRecorder()

	h.ExportReports(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// =============================================================================
// LogManagementHandler tests (mock-based, many functions at 0%)
// =============================================================================

// Mock for LogManagementServiceInterface
type mockLogManagementService struct {
	mock.Mock
}

func (m *mockLogManagementService) CollectLogs(userID int, request *models.LogCollectionRequest) (*models.LogCollection, error) {
	args := m.Called(userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LogCollection), args.Error(1)
}

func (m *mockLogManagementService) GetLogCollection(collectionID int, userID int) (*models.LogCollection, error) {
	args := m.Called(collectionID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LogCollection), args.Error(1)
}

func (m *mockLogManagementService) GetLogCollectionsByUser(userID int, limit, offset int) ([]models.LogCollection, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]models.LogCollection), args.Error(1)
}

func (m *mockLogManagementService) GetLogEntries(collectionID int, userID int, filters *models.LogEntryFilters) ([]models.LogEntry, error) {
	args := m.Called(collectionID, userID, filters)
	return args.Get(0).([]models.LogEntry), args.Error(1)
}

func (m *mockLogManagementService) CreateLogShare(userID int, request *models.LogShareRequest) (*models.LogShare, error) {
	args := m.Called(userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LogShare), args.Error(1)
}

func (m *mockLogManagementService) GetLogShare(token string) (*models.LogShare, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LogShare), args.Error(1)
}

func (m *mockLogManagementService) RevokeLogShare(shareID int, userID int) error {
	args := m.Called(shareID, userID)
	return args.Error(0)
}

func (m *mockLogManagementService) ExportLogs(collectionID int, userID int, format string) ([]byte, error) {
	args := m.Called(collectionID, userID, format)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockLogManagementService) StreamLogs(userID int, filters *models.LogStreamFilters) (<-chan models.LogEntry, error) {
	args := m.Called(userID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan models.LogEntry), args.Error(1)
}

func (m *mockLogManagementService) AnalyzeLogs(collectionID int, userID int) (*models.LogAnalysis, error) {
	args := m.Called(collectionID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LogAnalysis), args.Error(1)
}

func (m *mockLogManagementService) GetLogStatistics(userID int) (*models.LogStatistics, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LogStatistics), args.Error(1)
}

func (m *mockLogManagementService) GetConfiguration() *services.LogManagementConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*services.LogManagementConfig)
}

func (m *mockLogManagementService) UpdateConfiguration(config *services.LogManagementConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *mockLogManagementService) CleanupOldLogs() error {
	args := m.Called()
	return args.Error(0)
}

// Mock for LogManagementAuthServiceInterface
type mockLMAuthService struct {
	mock.Mock
}

func (m *mockLMAuthService) CheckPermission(userID int, permission string) (bool, error) {
	args := m.Called(userID, permission)
	return args.Bool(0), args.Error(1)
}

func TestNewLogManagementHandler(t *testing.T) {
	h := NewLogManagementHandler(nil, nil)
	assert.NotNil(t, h)
}

func TestLogManagementHandler_CreateLogCollection_Success(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	collection := &models.LogCollection{ID: 1, Name: "Test Collection"}
	svc.On("CollectLogs", 1, mock.Anything).Return(collection, nil)

	body := bytes.NewBufferString(`{"name": "Test Collection", "sources": ["api"]}`)
	req := reqWithUserID(http.MethodPost, "/logs", body, 1)
	w := httptest.NewRecorder()

	h.CreateLogCollection(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogManagementHandler_CreateLogCollection_Forbidden(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(false, nil)

	body := bytes.NewBufferString(`{"name": "Test"}`)
	req := reqWithUserID(http.MethodPost, "/logs", body, 1)
	w := httptest.NewRecorder()

	h.CreateLogCollection(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLogManagementHandler_CreateLogCollection_InvalidBody(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)

	body := bytes.NewBufferString("invalid json")
	req := reqWithUserID(http.MethodPost, "/logs", body, 1)
	w := httptest.NewRecorder()

	h.CreateLogCollection(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogManagementHandler_GetLogCollection_Success(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	collection := &models.LogCollection{ID: 1, Name: "Test"}
	svc.On("GetLogCollection", 1, 1).Return(collection, nil)

	req := reqWithUserID(http.MethodGet, "/logs/1", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.GetLogCollection(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogManagementHandler_GetLogCollection_InvalidID(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	req := reqWithUserID(http.MethodGet, "/logs/abc", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	h.GetLogCollection(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogManagementHandler_ListLogCollections_Success(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("GetLogCollectionsByUser", 1, 10, 5).Return([]models.LogCollection{}, nil)

	req := reqWithUserID(http.MethodGet, "/logs?limit=10&offset=5", nil, 1)
	w := httptest.NewRecorder()

	h.ListLogCollections(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogManagementHandler_ListLogCollections_DefaultPagination(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("GetLogCollectionsByUser", 1, 20, 0).Return([]models.LogCollection{}, nil)

	req := reqWithUserID(http.MethodGet, "/logs", nil, 1)
	w := httptest.NewRecorder()

	h.ListLogCollections(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogManagementHandler_GetLogEntries_Success(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("GetLogEntries", 1, 1, mock.Anything).Return([]models.LogEntry{}, nil)

	req := reqWithUserID(http.MethodGet, "/logs/1/entries?level=error&component=api&search=test&limit=10&offset=0", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.GetLogEntries(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogManagementHandler_GetLogEntries_InvalidID(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	req := reqWithUserID(http.MethodGet, "/logs/abc/entries", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	h.GetLogEntries(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogManagementHandler_CreateLogShare_Success(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	share := &models.LogShare{ID: 1, ShareToken: "abc123"}
	svc.On("CreateLogShare", 1, mock.Anything).Return(share, nil)

	body := bytes.NewBufferString(`{"collection_id": 1, "permissions": ["read"]}`)
	req := reqWithUserID(http.MethodPost, "/logs/share", body, 1)
	w := httptest.NewRecorder()

	h.CreateLogShare(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogManagementHandler_CreateLogShare_InvalidBody(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)

	body := bytes.NewBufferString("invalid")
	req := reqWithUserID(http.MethodPost, "/logs/share", body, 1)
	w := httptest.NewRecorder()

	h.CreateLogShare(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogManagementHandler_GetLogShare_Success(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	share := &models.LogShare{ID: 1, ShareToken: "abc", CollectionID: 1, UserID: 1, Permissions: []string{"read"}}
	collection := &models.LogCollection{ID: 1, Name: "Test"}
	svc.On("GetLogShare", "abc").Return(share, nil)
	svc.On("GetLogCollection", 1, 1).Return(collection, nil)

	req := httptest.NewRequest(http.MethodGet, "/logs/share/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"token": "abc"})
	w := httptest.NewRecorder()

	h.GetLogShare(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogManagementHandler_GetLogShare_NotFound(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	svc.On("GetLogShare", "invalid").Return(nil, errors.New("not found"))

	req := httptest.NewRequest(http.MethodGet, "/logs/share/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"token": "invalid"})
	w := httptest.NewRecorder()

	h.GetLogShare(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestLogManagementHandler_GetLogShare_NoReadPermission(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	share := &models.LogShare{ID: 1, ShareToken: "abc", CollectionID: 1, UserID: 1, Permissions: []string{"write"}}
	collection := &models.LogCollection{ID: 1, Name: "Test"}
	svc.On("GetLogShare", "abc").Return(share, nil)
	svc.On("GetLogCollection", 1, 1).Return(collection, nil)

	req := httptest.NewRequest(http.MethodGet, "/logs/share/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"token": "abc"})
	w := httptest.NewRecorder()

	h.GetLogShare(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLogManagementHandler_RevokeLogShare_Success(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("RevokeLogShare", 1, 1).Return(nil)

	req := reqWithUserID(http.MethodDelete, "/logs/share/1", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.RevokeLogShare(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestLogManagementHandler_RevokeLogShare_InvalidID(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	req := reqWithUserID(http.MethodDelete, "/logs/share/abc", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	h.RevokeLogShare(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogManagementHandler_ExportLogs_Success(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("ExportLogs", 1, 1, "json").Return([]byte(`{"logs":[]}`), nil)

	req := reqWithUserID(http.MethodGet, "/logs/1/export?format=json", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.ExportLogs(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestLogManagementHandler_ExportLogs_CSV(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("ExportLogs", 1, 1, "csv").Return([]byte("a,b\n"), nil)

	req := reqWithUserID(http.MethodGet, "/logs/1/export?format=csv", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.ExportLogs(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
}

func TestLogManagementHandler_ExportLogs_TXT(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("ExportLogs", 1, 1, "txt").Return([]byte("log line\n"), nil)

	req := reqWithUserID(http.MethodGet, "/logs/1/export?format=txt", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.ExportLogs(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
}

func TestLogManagementHandler_ExportLogs_ZIP(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("ExportLogs", 1, 1, "zip").Return([]byte("zip data"), nil)

	req := reqWithUserID(http.MethodGet, "/logs/1/export?format=zip", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.ExportLogs(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/zip", w.Header().Get("Content-Type"))
}

func TestLogManagementHandler_ExportLogs_DefaultFormat(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("ExportLogs", 1, 1, "json").Return([]byte("data"), nil)

	req := reqWithUserID(http.MethodGet, "/logs/1/export", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.ExportLogs(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogManagementHandler_ExportLogs_UnknownFormat(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("ExportLogs", 1, 1, "xml").Return([]byte("data"), nil)

	req := reqWithUserID(http.MethodGet, "/logs/1/export?format=xml", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.ExportLogs(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
}

func TestLogManagementHandler_AnalyzeLogs_Success(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	analysis := &models.LogAnalysis{}
	svc.On("AnalyzeLogs", 1, 1).Return(analysis, nil)

	req := reqWithUserID(http.MethodGet, "/logs/1/analyze", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.AnalyzeLogs(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogManagementHandler_AnalyzeLogs_InvalidID(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	req := reqWithUserID(http.MethodGet, "/logs/abc/analyze", nil, 1)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	w := httptest.NewRecorder()

	h.AnalyzeLogs(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogManagementHandler_GetLogStatistics_Success(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	stats := &models.LogStatistics{}
	svc.On("GetLogStatistics", 1).Return(stats, nil)

	req := reqWithUserID(http.MethodGet, "/logs/stats", nil, 1)
	w := httptest.NewRecorder()

	h.GetLogStatistics(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogManagementHandler_GetConfiguration_Success(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	config := &services.LogManagementConfig{}
	svc.On("GetConfiguration").Return(config)

	req := reqWithUserID(http.MethodGet, "/logs/config", nil, 1)
	w := httptest.NewRecorder()

	h.GetConfiguration(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogManagementHandler_UpdateConfiguration_Success(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("UpdateConfiguration", mock.Anything).Return(nil)

	body := bytes.NewBufferString(`{"max_size": 1024}`)
	req := reqWithUserID(http.MethodPut, "/logs/config", body, 1)
	w := httptest.NewRecorder()

	h.UpdateConfiguration(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogManagementHandler_UpdateConfiguration_InvalidBody(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)

	body := bytes.NewBufferString("invalid")
	req := reqWithUserID(http.MethodPut, "/logs/config", body, 1)
	w := httptest.NewRecorder()

	h.UpdateConfiguration(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogManagementHandler_CleanupOldLogs_Success(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("CleanupOldLogs").Return(nil)

	req := reqWithUserID(http.MethodPost, "/logs/cleanup", nil, 1)
	w := httptest.NewRecorder()

	h.CleanupOldLogs(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogManagementHandler_CleanupOldLogs_Forbidden(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(false, nil)

	req := reqWithUserID(http.MethodPost, "/logs/cleanup", nil, 1)
	w := httptest.NewRecorder()

	h.CleanupOldLogs(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLogManagementHandler_StreamLogs_Forbidden(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(false, nil)

	req := reqWithUserID(http.MethodGet, "/logs/stream?level=error&component=api&search=test", nil, 1)
	w := httptest.NewRecorder()

	h.StreamLogs(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLogManagementHandler_StreamLogs_ServiceError(t *testing.T) {
	svc := new(mockLogManagementService)
	auth := new(mockLMAuthService)
	h := NewLogManagementHandler(svc, auth)

	auth.On("CheckPermission", 1, models.PermissionSystemAdmin).Return(true, nil)
	svc.On("StreamLogs", 1, mock.Anything).Return(nil, errors.New("stream error"))

	req := reqWithUserID(http.MethodGet, "/logs/stream", nil, 1)
	w := httptest.NewRecorder()

	h.StreamLogs(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// =============================================================================
// AuthHandler Gin method tests
// =============================================================================

func newAuthHandlerForGinTests() *AuthHandler {
	authSvc := services.NewAuthService(nil, "test-secret-key-for-testing-gin-handlers")
	return &AuthHandler{authService: authSvc}
}

func TestAuthHandler_LoginGin_InvalidJSON(t *testing.T) {
	h := newAuthHandlerForGinTests()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", h.LoginGin)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAuthHandler_LoginGin_ValidBody skipped - nil DB in auth service causes panic

func TestAuthHandler_RefreshTokenGin_InvalidJSON(t *testing.T) {
	h := newAuthHandlerForGinTests()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/refresh", h.RefreshTokenGin)

	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString("bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAuthHandler_RefreshTokenGin_ValidBody skipped - nil DB causes panic

func TestAuthHandler_LogoutGin_NoToken(t *testing.T) {
	h := newAuthHandlerForGinTests()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/logout", h.LogoutGin)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_LogoutGin_WithToken(t *testing.T) {
	h := newAuthHandlerForGinTests()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/logout", h.LogoutGin)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Logout with invalid token will error
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestAuthHandler_GetCurrentUserGin_NoToken(t *testing.T) {
	h := newAuthHandlerForGinTests()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/me", h.GetCurrentUserGin)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_GetCurrentUserGin_InvalidToken(t *testing.T) {
	h := newAuthHandlerForGinTests()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/me", h.GetCurrentUserGin)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_GetAuthStatusGin_NoToken(t *testing.T) {
	h := newAuthHandlerForGinTests()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/auth/status", h.GetAuthStatusGin)

	req := httptest.NewRequest(http.MethodGet, "/auth/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, false, resp["authenticated"])
}

func TestAuthHandler_GetAuthStatusGin_InvalidToken(t *testing.T) {
	h := newAuthHandlerForGinTests()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/auth/status", h.GetAuthStatusGin)

	req := httptest.NewRequest(http.MethodGet, "/auth/status", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, false, resp["authenticated"])
}

func TestAuthHandler_GetPermissionsGin_NoToken(t *testing.T) {
	h := newAuthHandlerForGinTests()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/permissions", h.GetPermissionsGin)

	req := httptest.NewRequest(http.MethodGet, "/permissions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_GetPermissionsGin_BadToken(t *testing.T) {
	h := newAuthHandlerForGinTests()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/permissions", h.GetPermissionsGin)

	req := httptest.NewRequest(http.MethodGet, "/permissions", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_RegisterGin_InvalidJSON_Boost3(t *testing.T) {
	h := newAuthHandlerForGinTests()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/register", func(c *gin.Context) {
		h.RegisterGin(c, nil)
	})

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_RegisterGin_PartialFields(t *testing.T) {
	h := newAuthHandlerForGinTests()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/register", func(c *gin.Context) {
		h.RegisterGin(c, nil)
	})

	// Missing email, password, first_name, last_name
	body := `{"username": "testuser"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =============================================================================
// Additional auth handler tests for http.Handler style methods
// =============================================================================

func TestAuthHandler_ChangePassword_BadJSON(t *testing.T) {
	h := newAuthHandlerForGinTests()
	body := bytes.NewBufferString("invalid json")
	req := httptest.NewRequest(http.MethodPost, "/auth/change-password", body)
	req.Header.Set("Authorization", "Bearer some-token")
	w := httptest.NewRecorder()

	h.ChangePassword(w, req)
	// Either 400 for invalid body or 401 for bad token
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized)
}

func TestAuthHandler_Login_BadJSON(t *testing.T) {
	h := newAuthHandlerForGinTests()
	body := bytes.NewBufferString("invalid json")
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	w := httptest.NewRecorder()

	h.Login(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAuthHandler_Login_ValidCredentials skipped - nil DB causes panic

func TestAuthHandler_RefreshToken_BadJSON(t *testing.T) {
	h := newAuthHandlerForGinTests()
	body := bytes.NewBufferString("invalid json")
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", body)
	w := httptest.NewRecorder()

	h.RefreshToken(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAuthHandler_RefreshToken_BadToken skipped - nil DB causes panic

func TestAuthHandler_Logout_MissingToken(t *testing.T) {
	h := newAuthHandlerForGinTests()
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_LogoutAll_MissingToken(t *testing.T) {
	h := newAuthHandlerForGinTests()
	req := httptest.NewRequest(http.MethodPost, "/auth/logout-all", nil)
	w := httptest.NewRecorder()

	h.LogoutAll(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_GetCurrentUser_MissingToken(t *testing.T) {
	h := newAuthHandlerForGinTests()
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()

	h.GetCurrentUser(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_GetActiveSessions_MissingToken(t *testing.T) {
	h := newAuthHandlerForGinTests()
	req := httptest.NewRequest(http.MethodGet, "/auth/sessions", nil)
	w := httptest.NewRecorder()

	h.GetActiveSessions(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_DeactivateSession_MissingToken(t *testing.T) {
	h := newAuthHandlerForGinTests()
	req := httptest.NewRequest(http.MethodPost, "/auth/sessions/deactivate", nil)
	w := httptest.NewRecorder()

	h.DeactivateSession(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_DeactivateSession_WrongMethod(t *testing.T) {
	h := newAuthHandlerForGinTests()
	req := httptest.NewRequest(http.MethodGet, "/auth/sessions/deactivate", nil)
	w := httptest.NewRecorder()

	h.DeactivateSession(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
