package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"catalogizer/models"
	"catalogizer/repository"
)

type ErrorReportingService struct {
	errorRepo        *repository.ErrorReportingRepository
	crashRepo        *repository.CrashReportingRepository
	config           *ErrorReportingConfig
	httpClient       *http.Client
	enabled          bool
	crashlyticAPIKey string
}

type ErrorReportingConfig struct {
	CrashlyticsEnabled  bool   `json:"crashlytics_enabled"`
	CrashlyticsAPIKey   string `json:"crashlytics_api_key"`
	SlackWebhookURL     string `json:"slack_webhook_url"`
	EmailNotifications  bool   `json:"email_notifications"`
	SentryDSN           string `json:"sentry_dsn"`
	AutoReporting       bool   `json:"auto_reporting"`
	MaxErrorsPerHour    int    `json:"max_errors_per_hour"`
	RetentionDays       int    `json:"retention_days"`
	IncludeStackTrace   bool   `json:"include_stack_trace"`
	IncludeSystemInfo   bool   `json:"include_system_info"`
	FilterSensitiveData bool   `json:"filter_sensitive_data"`
}

func NewErrorReportingService(errorRepo *repository.ErrorReportingRepository, crashRepo *repository.CrashReportingRepository) *ErrorReportingService {
	config := &ErrorReportingConfig{
		CrashlyticsEnabled:  false,
		EmailNotifications:  true,
		AutoReporting:       true,
		MaxErrorsPerHour:    100,
		RetentionDays:       30,
		IncludeStackTrace:   true,
		IncludeSystemInfo:   true,
		FilterSensitiveData: true,
	}

	return &ErrorReportingService{
		errorRepo:  errorRepo,
		crashRepo:  crashRepo,
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		enabled:    true,
	}
}

func (s *ErrorReportingService) ReportError(userID int, errorReport *models.ErrorReportRequest) (*models.ErrorReport, error) {
	if !s.enabled {
		return nil, fmt.Errorf("error reporting is disabled")
	}

	// Check rate limiting
	if s.config.MaxErrorsPerHour > 0 {
		count, err := s.errorRepo.GetErrorCountInLastHour(userID)
		if err == nil && count >= s.config.MaxErrorsPerHour {
			return nil, fmt.Errorf("error reporting rate limit exceeded")
		}
	}

	// Create error report
	report := &models.ErrorReport{
		UserID:     userID,
		Level:      errorReport.Level,
		Message:    errorReport.Message,
		ErrorCode:  errorReport.ErrorCode,
		Component:  errorReport.Component,
		StackTrace: errorReport.StackTrace,
		Context:    errorReport.Context,
		UserAgent:  errorReport.UserAgent,
		URL:        errorReport.URL,
		ReportedAt: time.Now(),
		Status:     models.ErrorStatusNew,
	}

	// Filter sensitive data
	if s.config.FilterSensitiveData {
		report = s.filterSensitiveData(report)
	}

	// Add system information
	if s.config.IncludeSystemInfo {
		report.SystemInfo = s.collectSystemInfo()
	}

	// Generate fingerprint for deduplication
	report.Fingerprint = s.generateFingerprint(report)

	// Save to database
	if err := s.errorRepo.CreateErrorReport(report); err != nil {
		return nil, fmt.Errorf("failed to save error report: %w", err)
	}

	// Send notifications asynchronously
	if s.config.AutoReporting {
		go s.sendNotifications(report)
	}

	// Send to external services
	go s.sendToExternalServices(report)

	return report, nil
}

func (s *ErrorReportingService) ReportCrash(userID int, crashReport *models.CrashReportRequest) (*models.CrashReport, error) {
	if !s.enabled {
		return nil, fmt.Errorf("crash reporting is disabled")
	}

	// Create crash report
	report := &models.CrashReport{
		UserID:     userID,
		Signal:     crashReport.Signal,
		Message:    crashReport.Message,
		StackTrace: crashReport.StackTrace,
		Context:    crashReport.Context,
		ReportedAt: time.Now(),
		Status:     models.CrashStatusNew,
	}

	// Add system information
	report.SystemInfo = s.collectSystemInfo()

	// Generate fingerprint for deduplication
	report.Fingerprint = s.generateCrashFingerprint(report)

	// Save to database
	if err := s.crashRepo.CreateCrashReport(report); err != nil {
		return nil, fmt.Errorf("failed to save crash report: %w", err)
	}

	// Send critical notifications immediately
	go s.sendCrashNotifications(report)

	// Send to Crashlytics
	if s.config.CrashlyticsEnabled {
		go s.sendToCrashlytics(report)
	}

	return report, nil
}

func (s *ErrorReportingService) GetErrorReport(id int, userID int) (*models.ErrorReport, error) {
	report, err := s.errorRepo.GetErrorReport(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get error report: %w", err)
	}

	// Check if user has access
	if report.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	return report, nil
}

func (s *ErrorReportingService) GetCrashReport(id int, userID int) (*models.CrashReport, error) {
	report, err := s.crashRepo.GetCrashReport(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get crash report: %w", err)
	}

	// Check if user has access
	if report.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	return report, nil
}

func (s *ErrorReportingService) UpdateErrorStatus(id int, userID int, status string) error {
	report, err := s.errorRepo.GetErrorReport(id)
	if err != nil {
		return fmt.Errorf("failed to get error report: %w", err)
	}

	if report.UserID != userID {
		return fmt.Errorf("access denied")
	}

	report.Status = status
	if status == models.ErrorStatusResolved {
		now := time.Now()
		report.ResolvedAt = &now
	}

	return s.errorRepo.UpdateErrorReport(report)
}

func (s *ErrorReportingService) UpdateCrashStatus(id int, userID int, status string) error {
	report, err := s.crashRepo.GetCrashReport(id)
	if err != nil {
		return fmt.Errorf("failed to get crash report: %w", err)
	}

	if report.UserID != userID {
		return fmt.Errorf("access denied")
	}

	report.Status = status
	if status == models.CrashStatusResolved {
		now := time.Now()
		report.ResolvedAt = &now
	}

	return s.crashRepo.UpdateCrashReport(report)
}

func (s *ErrorReportingService) GetErrorReportsByUser(userID int, filters *models.ErrorReportFilters) ([]*models.ErrorReport, error) {
	return s.errorRepo.GetErrorReportsByUser(userID, filters)
}

func (s *ErrorReportingService) GetCrashReportsByUser(userID int, filters *models.CrashReportFilters) ([]*models.CrashReport, error) {
	return s.crashRepo.GetCrashReportsByUser(userID, filters)
}

func (s *ErrorReportingService) GetErrorStatistics(userID int) (*models.ErrorStatistics, error) {
	return s.errorRepo.GetErrorStatistics(userID)
}

func (s *ErrorReportingService) GetCrashStatistics(userID int) (*models.CrashStatistics, error) {
	return s.crashRepo.GetCrashStatistics(userID)
}

func (s *ErrorReportingService) GetSystemHealth() (*models.SystemHealth, error) {
	health := &models.SystemHealth{
		CheckedAt: time.Now(),
		Status:    "healthy",
		Metrics:   make(map[string]interface{}),
	}

	// Check recent error rates
	recentErrors, err := s.errorRepo.GetRecentErrorCount(1 * time.Hour)
	if err == nil {
		health.Metrics["recent_errors"] = recentErrors
		if recentErrors > 100 {
			health.Status = "degraded"
		}
	}

	// Check recent crash rates
	recentCrashes, err := s.crashRepo.GetRecentCrashCount(1 * time.Hour)
	if err == nil {
		health.Metrics["recent_crashes"] = recentCrashes
		if recentCrashes > 5 {
			health.Status = "critical"
		}
	}

	// Check system resources
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	health.Metrics["memory_used"] = m.Alloc
	health.Metrics["memory_total"] = m.Sys
	health.Metrics["goroutines"] = runtime.NumGoroutine()

	return health, nil
}

func (s *ErrorReportingService) UpdateConfiguration(config *ErrorReportingConfig) error {
	s.config = config
	return nil
}

func (s *ErrorReportingService) GetConfiguration() *ErrorReportingConfig {
	return s.config
}

func (s *ErrorReportingService) CleanupOldReports(olderThan time.Time) error {
	if err := s.errorRepo.CleanupOldReports(olderThan); err != nil {
		return fmt.Errorf("failed to cleanup old error reports: %w", err)
	}

	if err := s.crashRepo.CleanupOldReports(olderThan); err != nil {
		return fmt.Errorf("failed to cleanup old crash reports: %w", err)
	}

	return nil
}

func (s *ErrorReportingService) ExportReports(userID int, filters *models.ExportFilters) ([]byte, error) {
	var reports []interface{}

	// Get error reports
	if filters.IncludeErrors {
		errorReports, err := s.errorRepo.GetErrorReportsByUser(userID, &models.ErrorReportFilters{
			StartDate: filters.StartDate,
			EndDate:   filters.EndDate,
			Level:     filters.Level,
			Component: filters.Component,
			Limit:     filters.Limit,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get error reports: %w", err)
		}
		for _, report := range errorReports {
			reports = append(reports, report)
		}
	}

	// Get crash reports
	if filters.IncludeCrashes {
		crashReports, err := s.crashRepo.GetCrashReportsByUser(userID, &models.CrashReportFilters{
			StartDate: filters.StartDate,
			EndDate:   filters.EndDate,
			Signal:    filters.Signal,
			Limit:     filters.Limit,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get crash reports: %w", err)
		}
		for _, report := range crashReports {
			reports = append(reports, report)
		}
	}

	// Export based on format
	switch filters.Format {
	case "json":
		return json.MarshalIndent(reports, "", "  ")
	case "csv":
		return s.exportToCSV(reports)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", filters.Format)
	}
}

// Helper methods

func (s *ErrorReportingService) filterSensitiveData(report *models.ErrorReport) *models.ErrorReport {
	// Remove sensitive patterns from message and stack trace
	sensitivePatterns := []string{
		"password", "token", "key", "secret", "auth",
		"email", "phone", "ssn", "credit",
	}

	for _, pattern := range sensitivePatterns {
		report.Message = strings.ReplaceAll(strings.ToLower(report.Message), pattern, "[REDACTED]")
		report.StackTrace = strings.ReplaceAll(strings.ToLower(report.StackTrace), pattern, "[REDACTED]")
	}

	// Filter context data
	if report.Context != nil {
		filteredContext := make(map[string]interface{})
		for key, value := range report.Context {
			keyLower := strings.ToLower(key)
			isSensitive := false
			for _, pattern := range sensitivePatterns {
				if strings.Contains(keyLower, pattern) {
					isSensitive = true
					break
				}
			}
			if !isSensitive {
				filteredContext[key] = value
			} else {
				filteredContext[key] = "[REDACTED]"
			}
		}
		report.Context = filteredContext
	}

	return report
}

func (s *ErrorReportingService) collectSystemInfo() map[string]interface{} {
	info := make(map[string]interface{})

	info["os"] = runtime.GOOS
	info["arch"] = runtime.GOARCH
	info["go_version"] = runtime.Version()
	info["num_cpu"] = runtime.NumCPU()
	info["num_goroutine"] = runtime.NumGoroutine()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	info["memory_alloc"] = m.Alloc
	info["memory_total_alloc"] = m.TotalAlloc
	info["memory_sys"] = m.Sys

	if hostname, err := os.Hostname(); err == nil {
		info["hostname"] = hostname
	}

	if wd, err := os.Getwd(); err == nil {
		info["working_directory"] = wd
	}

	return info
}

func (s *ErrorReportingService) generateFingerprint(report *models.ErrorReport) string {
	// Create a unique fingerprint based on error characteristics
	data := fmt.Sprintf("%s:%s:%s", report.Level, report.Component, report.ErrorCode)
	return fmt.Sprintf("%x", data)[:16]
}

func (s *ErrorReportingService) generateCrashFingerprint(report *models.CrashReport) string {
	// Create a unique fingerprint based on crash characteristics
	data := fmt.Sprintf("%s:%s", report.Signal, report.Message)
	return fmt.Sprintf("%x", data)[:16]
}

func (s *ErrorReportingService) sendNotifications(report *models.ErrorReport) {
	// Send Slack notification
	if s.config.SlackWebhookURL != "" {
		s.sendSlackNotification(report)
	}

	// Send email notification
	if s.config.EmailNotifications {
		s.sendEmailNotification(report)
	}
}

func (s *ErrorReportingService) sendCrashNotifications(report *models.CrashReport) {
	// Send critical notifications for crashes
	if s.config.SlackWebhookURL != "" {
		s.sendSlackCrashNotification(report)
	}

	if s.config.EmailNotifications {
		s.sendEmailCrashNotification(report)
	}
}

func (s *ErrorReportingService) sendSlackNotification(report *models.ErrorReport) error {
	if s.config.SlackWebhookURL == "" {
		return nil
	}

	message := map[string]interface{}{
		"text": fmt.Sprintf("🚨 Error Report: %s", report.Message),
		"attachments": []map[string]interface{}{
			{
				"color": s.getColorForLevel(report.Level),
				"fields": []map[string]interface{}{
					{"title": "Level", "value": report.Level, "short": true},
					{"title": "Component", "value": report.Component, "short": true},
					{"title": "Error Code", "value": report.ErrorCode, "short": true},
					{"title": "Time", "value": report.ReportedAt.Format(time.RFC3339), "short": true},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(message)
	_, err := s.httpClient.Post(s.config.SlackWebhookURL, "application/json", bytes.NewBuffer(jsonData))
	return err
}

func (s *ErrorReportingService) sendSlackCrashNotification(report *models.CrashReport) error {
	if s.config.SlackWebhookURL == "" {
		return nil
	}

	message := map[string]interface{}{
		"text": fmt.Sprintf("💥 CRASH REPORT: %s", report.Message),
		"attachments": []map[string]interface{}{
			{
				"color": "danger",
				"fields": []map[string]interface{}{
					{"title": "Signal", "value": report.Signal, "short": true},
					{"title": "Time", "value": report.ReportedAt.Format(time.RFC3339), "short": true},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(message)
	_, err := s.httpClient.Post(s.config.SlackWebhookURL, "application/json", bytes.NewBuffer(jsonData))
	return err
}

func (s *ErrorReportingService) sendEmailNotification(report *models.ErrorReport) error {
	// Email implementation would go here
	// This is a placeholder for email notification logic
	return nil
}

func (s *ErrorReportingService) sendEmailCrashNotification(report *models.CrashReport) error {
	// Email implementation would go here
	// This is a placeholder for email notification logic
	return nil
}

func (s *ErrorReportingService) sendToExternalServices(report *models.ErrorReport) {
	// Send to Sentry
	if s.config.SentryDSN != "" {
		s.sendToSentry(report)
	}
}

func (s *ErrorReportingService) sendToSentry(report *models.ErrorReport) error {
	// Sentry integration would go here
	// This is a placeholder for Sentry integration
	return nil
}

func (s *ErrorReportingService) sendToCrashlytics(report *models.CrashReport) error {
	if s.config.CrashlyticsAPIKey == "" {
		return nil
	}

	// Crashlytics integration would go here
	// This is a placeholder for Firebase Crashlytics integration
	return nil
}

func (s *ErrorReportingService) getColorForLevel(level string) string {
	switch strings.ToLower(level) {
	case "error", "fatal":
		return "danger"
	case "warning", "warn":
		return "warning"
	case "info":
		return "good"
	default:
		return "#36a64f"
	}
}

func (s *ErrorReportingService) exportToCSV(reports []interface{}) ([]byte, error) {
	// CSV export implementation would go here
	// This is a placeholder for CSV export logic
	return []byte("CSV export not implemented"), nil
}
