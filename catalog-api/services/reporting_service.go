package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"

	"catalogizer/models"
	"catalogizer/repository"

	"github.com/jung-kurt/gofpdf"
)

type ReportingService struct {
	analyticsRepo *repository.AnalyticsRepository
	userRepo      *repository.UserRepository
}

func NewReportingService(analyticsRepo *repository.AnalyticsRepository, userRepo *repository.UserRepository) *ReportingService {
	return &ReportingService{
		analyticsRepo: analyticsRepo,
		userRepo:      userRepo,
	}
}

func (s *ReportingService) GenerateReport(reportType string, format string, params map[string]interface{}) (*models.GeneratedReport, error) {
	var data interface{}
	var err error

	switch reportType {
	case "user_analytics":
		data, err = s.generateUserAnalyticsData(params)
	case "system_overview":
		data, err = s.generateSystemOverviewData(params)
	case "media_analytics":
		data, err = s.generateMediaAnalyticsData(params)
	case "user_activity":
		data, err = s.generateUserActivityData(params)
	case "security_audit":
		data, err = s.generateSecurityAuditData(params)
	case "performance_metrics":
		data, err = s.generatePerformanceMetricsData(params)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", reportType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate report data: %w", err)
	}

	content, err := s.formatReport(data, format, reportType)
	if err != nil {
		return nil, fmt.Errorf("failed to format report: %w", err)
	}

	report := &models.GeneratedReport{
		Type:        reportType,
		Format:      format,
		Content:     content,
		GeneratedAt: time.Now(),
		Parameters:  params,
	}

	return report, nil
}

func (s *ReportingService) generateUserAnalyticsData(params map[string]interface{}) (interface{}, error) {
	userID, ok := params["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("user_id parameter required")
	}

	startDate, endDate, err := s.extractDateRange(params)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	mediaAccessLogs, err := s.analyticsRepo.GetUserMediaAccessLogs(userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get media access logs: %w", err)
	}

	events, err := s.analyticsRepo.GetUserEvents(userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get user events: %w", err)
	}

	analytics := &models.UserAnalyticsReport{
		User:               user,
		StartDate:          startDate,
		EndDate:            endDate,
		TotalMediaAccesses: len(mediaAccessLogs),
		TotalEvents:        len(events),
		MediaAccessLogs:    mediaAccessLogs,
		Events:             events,
		AccessPatterns:     s.analyzeUserAccessPatterns(mediaAccessLogs),
		DeviceUsage:        s.analyzeUserDeviceUsage(mediaAccessLogs),
		LocationAnalysis:   s.analyzeUserLocations(mediaAccessLogs),
		TimePatterns:       s.analyzeUserTimePatterns(mediaAccessLogs),
		PopularContent:     s.analyzeUserPopularContent(mediaAccessLogs),
	}

	return analytics, nil
}

func (s *ReportingService) generateSystemOverviewData(params map[string]interface{}) (interface{}, error) {
	startDate, endDate, err := s.extractDateRange(params)
	if err != nil {
		return nil, err
	}

	totalUsers, err := s.analyticsRepo.GetTotalUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get total users: %w", err)
	}

	activeUsers, err := s.analyticsRepo.GetActiveUsers(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	totalMediaAccesses, err := s.analyticsRepo.GetTotalMediaAccesses(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get total media accesses: %w", err)
	}

	totalEvents, err := s.analyticsRepo.GetTotalEvents(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get total events: %w", err)
	}

	topMedia, err := s.analyticsRepo.GetTopAccessedMedia(startDate, endDate, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to get top media: %w", err)
	}

	userGrowth, err := s.analyticsRepo.GetUserGrowthData(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get user growth data: %w", err)
	}

	overview := &models.SystemOverviewReport{
		StartDate:          startDate,
		EndDate:            endDate,
		TotalUsers:         totalUsers,
		ActiveUsers:        activeUsers,
		TotalMediaAccesses: totalMediaAccesses,
		TotalEvents:        totalEvents,
		TopAccessedMedia:   topMedia,
		UserGrowthData:     userGrowth,
		SystemHealth:       s.calculateSystemHealth(totalUsers, activeUsers, totalMediaAccesses),
		UsageStatistics:    s.calculateUsageStatistics(startDate, endDate),
		PerformanceMetrics: s.calculatePerformanceMetrics(startDate, endDate),
	}

	return overview, nil
}

func (s *ReportingService) generateMediaAnalyticsData(params map[string]interface{}) (interface{}, error) {
	mediaIDFloat, ok := params["media_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("media_id parameter required")
	}
	mediaID := int(mediaIDFloat)

	startDate, endDate, err := s.extractDateRange(params)
	if err != nil {
		return nil, err
	}

	accessLogs, err := s.analyticsRepo.GetMediaAccessLogs(0, &mediaID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get media access logs: %w", err)
	}

	filteredLogs := s.filterLogsByDateRange(accessLogs, startDate, endDate)

	analytics := &models.MediaAnalyticsReport{
		MediaID:        mediaID,
		StartDate:      startDate,
		EndDate:        endDate,
		TotalAccesses:  len(filteredLogs),
		UniqueUsers:    s.countUniqueUsers(filteredLogs),
		AccessLogs:     filteredLogs,
		AccessPatterns: s.analyzeAccessPatterns(filteredLogs),
		UserEngagement: s.analyzeUserEngagement(filteredLogs),
		GeographicData: s.analyzeGeographicDistribution(filteredLogs),
		DeviceAnalysis: s.analyzeDeviceDistribution(filteredLogs),
		TimeAnalysis:   s.analyzeTimeDistribution(filteredLogs),
	}

	return analytics, nil
}

func (s *ReportingService) generateUserActivityData(params map[string]interface{}) (interface{}, error) {
	startDate, endDate, err := s.extractDateRange(params)
	if err != nil {
		return nil, err
	}

	allLogs, err := s.analyticsRepo.GetAllMediaAccessLogs(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get all media access logs: %w", err)
	}

	userActivity := make(map[int][]models.MediaAccessLog)
	for _, log := range allLogs {
		userActivity[log.UserID] = append(userActivity[log.UserID], log)
	}

	var userActivities []models.UserActivitySummary
	for userID, logs := range userActivity {
		user, err := s.userRepo.GetByID(userID)
		if err != nil {
			continue
		}

		activity := models.UserActivitySummary{
			User:              user,
			TotalAccesses:     len(logs),
			LastActivity:      s.getLastActivityTime(logs),
			MostActiveHour:    s.getMostActiveHour(logs),
			PreferredDevices:  s.getPreferredDevices(logs),
			AccessedLocations: s.getAccessedLocations(logs),
		}

		userActivities = append(userActivities, activity)
	}

	report := &models.UserActivityReport{
		StartDate:      startDate,
		EndDate:        endDate,
		UserActivities: userActivities,
		TotalUsers:     len(userActivities),
		TotalAccesses:  len(allLogs),
		Summary:        s.generateActivitySummary(userActivities),
	}

	return report, nil
}

func (s *ReportingService) generateSecurityAuditData(params map[string]interface{}) (interface{}, error) {
	startDate, endDate, err := s.extractDateRange(params)
	if err != nil {
		return nil, err
	}

	// For now, return basic security metrics
	// In a full implementation, this would analyze login attempts, failed authentications, etc.
	audit := &models.SecurityAuditReport{
		StartDate:           startDate,
		EndDate:             endDate,
		FailedLoginAttempts: 0, // Would be calculated from actual data
		SuccessfulLogins:    0, // Would be calculated from actual data
		SuspiciousActivity:  []models.SecurityIncident{},
		SecurityMetrics:     s.calculateSecurityMetrics(startDate, endDate),
	}

	return audit, nil
}

func (s *ReportingService) generatePerformanceMetricsData(params map[string]interface{}) (interface{}, error) {
	startDate, endDate, err := s.extractDateRange(params)
	if err != nil {
		return nil, err
	}

	sessionData, err := s.analyticsRepo.GetSessionData(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get session data: %w", err)
	}

	metrics := &models.PerformanceMetricsReport{
		StartDate:              startDate,
		EndDate:                endDate,
		AverageSessionDuration: s.calculateAverageSessionDuration(sessionData),
		TotalSessions:          len(sessionData),
		ResponseTimes:          s.calculateResponseTimes(startDate, endDate),
		SystemLoad:             s.calculateSystemLoad(startDate, endDate),
		ErrorRates:             s.calculateErrorRates(startDate, endDate),
	}

	return metrics, nil
}

func (s *ReportingService) formatReport(data interface{}, format string, reportType string) ([]byte, error) {
	switch format {
	case "json":
		return json.MarshalIndent(data, "", "  ")
	case "markdown":
		return s.formatAsMarkdown(data, reportType)
	case "html":
		return s.formatAsHTML(data, reportType)
	case "pdf":
		return s.formatAsPDF(data, reportType)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (s *ReportingService) formatAsMarkdown(data interface{}, reportType string) ([]byte, error) {
	var buffer bytes.Buffer

	switch reportType {
	case "user_analytics":
		report := data.(*models.UserAnalyticsReport)
		buffer.WriteString(fmt.Sprintf("# User Analytics Report\n\n"))
		displayName := ""
		if report.User.DisplayName != nil {
			displayName = *report.User.DisplayName
		}
		buffer.WriteString(fmt.Sprintf("**User:** %s (%s)\n", displayName, report.User.Username))
		buffer.WriteString(fmt.Sprintf("**Period:** %s to %s\n\n", report.StartDate.Format("2006-01-02"), report.EndDate.Format("2006-01-02")))
		buffer.WriteString(fmt.Sprintf("## Summary\n\n"))
		buffer.WriteString(fmt.Sprintf("- Total Media Accesses: %d\n", report.TotalMediaAccesses))
		buffer.WriteString(fmt.Sprintf("- Total Events: %d\n", report.TotalEvents))
		buffer.WriteString(fmt.Sprintf("- Account Created: %s\n\n", report.User.CreatedAt.Format("2006-01-02")))

	case "system_overview":
		report := data.(*models.SystemOverviewReport)
		buffer.WriteString(fmt.Sprintf("# System Overview Report\n\n"))
		buffer.WriteString(fmt.Sprintf("**Period:** %s to %s\n\n", report.StartDate.Format("2006-01-02"), report.EndDate.Format("2006-01-02")))
		buffer.WriteString(fmt.Sprintf("## System Statistics\n\n"))
		buffer.WriteString(fmt.Sprintf("- Total Users: %d\n", report.TotalUsers))
		buffer.WriteString(fmt.Sprintf("- Active Users: %d\n", report.ActiveUsers))
		buffer.WriteString(fmt.Sprintf("- Total Media Accesses: %d\n", report.TotalMediaAccesses))
		buffer.WriteString(fmt.Sprintf("- Total Events: %d\n\n", report.TotalEvents))

	default:
		buffer.WriteString(fmt.Sprintf("# %s Report\n\n", reportType))
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		buffer.WriteString(fmt.Sprintf("```json\n%s\n```\n", string(jsonData)))
	}

	return buffer.Bytes(), nil
}

func (s *ReportingService) formatAsHTML(data interface{}, reportType string) ([]byte, error) {
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { border-bottom: 2px solid #333; padding-bottom: 20px; margin-bottom: 30px; }
        .section { margin-bottom: 30px; }
        .metric { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.Title}}</h1>
        <p>Generated on: {{.GeneratedAt}}</p>
        {{if .Period}}<p>Period: {{.Period}}</p>{{end}}
    </div>

    <div class="content">
        {{.Content}}
    </div>
</body>
</html>`

	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var content string
	switch reportType {
	case "user_analytics":
		report := data.(*models.UserAnalyticsReport)
		displayNameHTML := ""
		if report.User.DisplayName != nil {
			displayNameHTML = *report.User.DisplayName
		}
		content = fmt.Sprintf(`
			<div class="section">
				<h2>User Information</h2>
				<div class="metric">Username: %s</div>
				<div class="metric">Display Name: %s</div>
				<div class="metric">Email: %s</div>
			</div>
			<div class="section">
				<h2>Activity Summary</h2>
				<div class="metric">Total Media Accesses: %d</div>
				<div class="metric">Total Events: %d</div>
			</div>`,
			report.User.Username, displayNameHTML, report.User.Email,
			report.TotalMediaAccesses, report.TotalEvents)

	case "system_overview":
		report := data.(*models.SystemOverviewReport)
		content = fmt.Sprintf(`
			<div class="section">
				<h2>System Statistics</h2>
				<div class="metric">Total Users: %d</div>
				<div class="metric">Active Users: %d</div>
				<div class="metric">Total Media Accesses: %d</div>
				<div class="metric">Total Events: %d</div>
			</div>`,
			report.TotalUsers, report.ActiveUsers, report.TotalMediaAccesses, report.TotalEvents)

	default:
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		content = fmt.Sprintf("<pre>%s</pre>", string(jsonData))
	}

	templateData := struct {
		Title       string
		GeneratedAt string
		Period      string
		Content     template.HTML
	}{
		Title:       fmt.Sprintf("%s Report", reportType),
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Content:     template.HTML(content),
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return buffer.Bytes(), nil
}

func (s *ReportingService) formatAsPDF(data interface{}, reportType string) ([]byte, error) {
	// Create new PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set title
	pdf.SetFont("Arial", "B", 16)
	title := fmt.Sprintf("%s Report", strings.Replace(reportType, "_", " ", -1))
	pdf.Cell(40, 10, title)
	pdf.Ln(15)

	// Add generation timestamp
	pdf.SetFont("Arial", "I", 10)
	pdf.Cell(40, 8, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))
	pdf.Ln(12)

	// Format content based on report type
	switch reportType {
	case "user_analytics":
		return s.formatUserAnalyticsPDF(pdf, data)
	case "system_overview":
		return s.formatSystemOverviewPDF(pdf, data)
	case "media_analytics":
		return s.formatMediaAnalyticsPDF(pdf, data)
	case "user_activity":
		return s.formatUserActivityPDF(pdf, data)
	case "security_audit":
		return s.formatSecurityAuditPDF(pdf, data)
	case "performance_metrics":
		return s.formatPerformanceMetricsPDF(pdf, data)
	default:
		// Fallback to JSON representation
		pdf.SetFont("Courier", "", 10)
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data for PDF: %w", err)
		}

		// Split JSON into lines and add to PDF
		lines := strings.Split(string(jsonData), "\n")
		for _, line := range lines {
			// Truncate long lines to fit page
			if len(line) > 80 {
				for i := 0; i < len(line); i += 80 {
					end := i + 80
					if end > len(line) {
						end = len(line)
					}
					pdf.Cell(40, 5, line[i:end])
					pdf.Ln(5)
				}
			} else {
				pdf.Cell(40, 5, line)
				pdf.Ln(5)
			}
		}

		// Output PDF to bytes
		return s.outputPDFToBytes(pdf)
	}
}

// Helper function to output PDF to bytes
func (s *ReportingService) outputPDFToBytes(pdf *gofpdf.Fpdf) ([]byte, error) {
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return buf.Bytes(), nil
}

// Helper methods for specific report types
func (s *ReportingService) formatUserAnalyticsPDF(pdf *gofpdf.Fpdf, data interface{}) ([]byte, error) {
	report := data.(*models.UserAnalyticsReport)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "User Information")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, fmt.Sprintf("User ID: %d", report.User.ID))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Username: %s", report.User.Username))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Email: %s", report.User.Email))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Period: %s to %s",
		report.StartDate.Format("2006-01-02"),
		report.EndDate.Format("2006-01-02")))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "Summary Statistics")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, fmt.Sprintf("Total Media Accesses: %d", report.TotalMediaAccesses))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Total Events: %d", report.TotalEvents))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Access Logs: %d", len(report.MediaAccessLogs)))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("User Events: %d", len(report.Events)))

	return s.outputPDFToBytes(pdf)
}

func (s *ReportingService) formatSystemOverviewPDF(pdf *gofpdf.Fpdf, data interface{}) ([]byte, error) {
	report := data.(*models.SystemOverviewReport)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "System Statistics")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, fmt.Sprintf("Total Users: %d", report.TotalUsers))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Active Users: %d", report.ActiveUsers))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Total Media Accesses: %d", report.TotalMediaAccesses))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Total Events: %d", report.TotalEvents))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Report Period: %s to %s",
		report.StartDate.Format("2006-01-02"),
		report.EndDate.Format("2006-01-02")))

	return s.outputPDFToBytes(pdf)
}

func (s *ReportingService) formatMediaAnalyticsPDF(pdf *gofpdf.Fpdf, data interface{}) ([]byte, error) {
	// Generic media analytics formatting
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "Media Analytics")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, "Media statistics and analytics data")
	pdf.Ln(6)

	// Add JSON data for now (can be enhanced later)
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal media analytics data: %w", err)
	}

	lines := strings.Split(string(jsonData), "\n")
	pdf.SetFont("Courier", "", 8)
	for _, line := range lines {
		if len(line) > 100 {
			line = line[:100] + "..."
		}
		pdf.Cell(40, 3, line)
		pdf.Ln(3)
	}

	return s.outputPDFToBytes(pdf)
}

func (s *ReportingService) formatUserActivityPDF(pdf *gofpdf.Fpdf, data interface{}) ([]byte, error) {
	// Generic user activity formatting
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "User Activity Report")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, "User activity logs and events")
	pdf.Ln(6)

	// Add summary information
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user activity data: %w", err)
	}

	pdf.SetFont("Courier", "", 8)
	text := string(jsonData)
	if len(text) > 3000 { // Limit text to reasonable size
		text = text[:3000] + "...[truncated]"
	}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if len(line) > 100 {
			line = line[:100] + "..."
		}
		pdf.Cell(40, 3, line)
		pdf.Ln(3)
	}

	return s.outputPDFToBytes(pdf)
}

func (s *ReportingService) formatSecurityAuditPDF(pdf *gofpdf.Fpdf, data interface{}) ([]byte, error) {
	// Security audit specific formatting
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "Security Audit Report")
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 6, "Security Events and Audit Information")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, "This report contains security audit information")
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Generated on: %s", time.Now().Format("2006-01-02 15:04:05")))
	pdf.Ln(10)

	// Add audit data as formatted JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal security audit data: %w", err)
	}

	pdf.SetFont("Courier", "", 8)
	lines := strings.Split(string(jsonData), "\n")
	for _, line := range lines {
		if len(line) > 95 {
			line = line[:95] + "..."
		}
		pdf.Cell(40, 3, line)
		pdf.Ln(3)
	}

	return s.outputPDFToBytes(pdf)
}

func (s *ReportingService) formatPerformanceMetricsPDF(pdf *gofpdf.Fpdf, data interface{}) ([]byte, error) {
	// Performance metrics specific formatting
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "Performance Metrics Report")
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 6, "System Performance Metrics")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, "This report contains performance metrics for the system")
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Generated on: %s", time.Now().Format("2006-01-02 15:04:05")))
	pdf.Ln(10)

	// Add performance data as formatted JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal performance metrics data: %w", err)
	}

	pdf.SetFont("Courier", "", 8)
	lines := strings.Split(string(jsonData), "\n")
	for _, line := range lines {
		if len(line) > 95 {
			line = line[:95] + "..."
		}
		pdf.Cell(40, 3, line)
		pdf.Ln(3)
	}

	return s.outputPDFToBytes(pdf)
}

func (s *ReportingService) extractDateRange(params map[string]interface{}) (time.Time, time.Time, error) {
	startDateStr, ok := params["start_date"].(string)
	if !ok {
		return time.Time{}, time.Time{}, fmt.Errorf("start_date parameter required")
	}

	endDateStr, ok := params["end_date"].(string)
	if !ok {
		return time.Time{}, time.Time{}, fmt.Errorf("end_date parameter required")
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date format")
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date format")
	}

	return startDate, endDate, nil
}

// Helper methods for analytics calculations
func (s *ReportingService) analyzeUserAccessPatterns(logs []models.MediaAccessLog) map[string]interface{} {
	patterns := make(map[string]interface{})

	hourlyAccess := make(map[int]int)
	dailyAccess := make(map[string]int)

	for _, log := range logs {
		hour := log.AccessTime.Hour()
		hourlyAccess[hour]++

		day := log.AccessTime.Weekday().String()
		dailyAccess[day]++
	}

	patterns["hourly"] = hourlyAccess
	patterns["daily"] = dailyAccess

	return patterns
}

func (s *ReportingService) analyzeUserDeviceUsage(logs []models.MediaAccessLog) map[string]int {
	deviceUsage := make(map[string]int)

	for _, log := range logs {
		if log.DeviceInfo != nil {
			deviceModel := ""
			if log.DeviceInfo.DeviceModel != nil {
				deviceModel = *log.DeviceInfo.DeviceModel
			}
			device := fmt.Sprintf("%s %s", log.DeviceInfo.Platform, deviceModel)
			deviceUsage[device]++
		}
	}

	return deviceUsage
}

func (s *ReportingService) analyzeUserLocations(logs []models.MediaAccessLog) map[string]int {
	locations := make(map[string]int)

	for _, log := range logs {
		if log.Location != nil {
			location := fmt.Sprintf("%.2f,%.2f", log.Location.Latitude, log.Location.Longitude)
			locations[location]++
		}
	}

	return locations
}

func (s *ReportingService) analyzeUserTimePatterns(logs []models.MediaAccessLog) map[string]interface{} {
	return s.analyzeUserAccessPatterns(logs) // Same as access patterns
}

func (s *ReportingService) analyzeUserPopularContent(logs []models.MediaAccessLog) []models.MediaAccessCount {
	mediaCount := make(map[int]int)

	for _, log := range logs {
		mediaCount[log.MediaID]++
	}

	var results []models.MediaAccessCount
	for mediaID, count := range mediaCount {
		results = append(results, models.MediaAccessCount{
			MediaID:     mediaID,
			AccessCount: count,
		})
	}

	return results
}

func (s *ReportingService) filterLogsByDateRange(logs []models.MediaAccessLog, startDate, endDate time.Time) []models.MediaAccessLog {
	var filtered []models.MediaAccessLog

	for _, log := range logs {
		if log.AccessTime.After(startDate) && log.AccessTime.Before(endDate) {
			filtered = append(filtered, log)
		}
	}

	return filtered
}

func (s *ReportingService) countUniqueUsers(logs []models.MediaAccessLog) int {
	users := make(map[int]bool)

	for _, log := range logs {
		users[log.UserID] = true
	}

	return len(users)
}

func (s *ReportingService) calculateSystemHealth(totalUsers, activeUsers, mediaAccesses int) models.SystemHealth {
	var healthScore float64

	if totalUsers > 0 {
		activeRatio := float64(activeUsers) / float64(totalUsers)
		healthScore += activeRatio * 50
	}

	if mediaAccesses > 0 {
		healthScore += 30
	}

	if activeUsers > 10 {
		healthScore += 20
	}

	var status string
	switch {
	case healthScore >= 80:
		status = "excellent"
	case healthScore >= 60:
		status = "good"
	case healthScore >= 40:
		status = "fair"
	default:
		status = "poor"
	}

	return models.SystemHealth{
		Score:  healthScore,
		Status: status,
	}
}

func (s *ReportingService) calculateUsageStatistics(startDate, endDate time.Time) models.UsageStatistics {
	// Placeholder implementation
	return models.UsageStatistics{
		PeakHours:    []int{14, 15, 16, 20, 21},
		AverageDaily: 150,
		GrowthRate:   5.2,
	}
}

func (s *ReportingService) calculatePerformanceMetrics(startDate, endDate time.Time) models.PerformanceMetrics {
	// Placeholder implementation
	return models.PerformanceMetrics{
		ResponseTime: 250.5,
		Throughput:   1200,
		ErrorRate:    0.02,
	}
}

func (s *ReportingService) analyzeAccessPatterns(logs []models.MediaAccessLog) map[string]interface{} {
	return s.analyzeUserAccessPatterns(logs)
}

func (s *ReportingService) analyzeUserEngagement(logs []models.MediaAccessLog) models.UserEngagement {
	return models.UserEngagement{
		AverageSessionTime: 15.5,
		ReturnRate:         85.2,
		InteractionDepth:   3.4,
	}
}

func (s *ReportingService) analyzeGeographicDistribution(logs []models.MediaAccessLog) map[string]int {
	return s.analyzeUserLocations(logs)
}

func (s *ReportingService) analyzeDeviceDistribution(logs []models.MediaAccessLog) map[string]int {
	return s.analyzeUserDeviceUsage(logs)
}

func (s *ReportingService) analyzeTimeDistribution(logs []models.MediaAccessLog) map[string]int {
	timeDistribution := make(map[string]int)

	for _, log := range logs {
		hour := log.AccessTime.Hour()
		var timeSlot string

		switch {
		case hour >= 6 && hour < 12:
			timeSlot = "morning"
		case hour >= 12 && hour < 18:
			timeSlot = "afternoon"
		case hour >= 18 && hour < 22:
			timeSlot = "evening"
		default:
			timeSlot = "night"
		}

		timeDistribution[timeSlot]++
	}

	return timeDistribution
}

func (s *ReportingService) getLastActivityTime(logs []models.MediaAccessLog) time.Time {
	if len(logs) == 0 {
		return time.Time{}
	}

	latest := logs[0].AccessTime
	for _, log := range logs {
		if log.AccessTime.After(latest) {
			latest = log.AccessTime
		}
	}

	return latest
}

func (s *ReportingService) getMostActiveHour(logs []models.MediaAccessLog) int {
	hourCounts := make(map[int]int)

	for _, log := range logs {
		hour := log.AccessTime.Hour()
		hourCounts[hour]++
	}

	maxCount := 0
	mostActiveHour := 0

	for hour, count := range hourCounts {
		if count > maxCount {
			maxCount = count
			mostActiveHour = hour
		}
	}

	return mostActiveHour
}

func (s *ReportingService) getPreferredDevices(logs []models.MediaAccessLog) []string {
	deviceCounts := s.analyzeUserDeviceUsage(logs)

	var devices []string
	for device := range deviceCounts {
		devices = append(devices, device)
	}

	return devices
}

func (s *ReportingService) getAccessedLocations(logs []models.MediaAccessLog) []string {
	locationCounts := s.analyzeUserLocations(logs)

	var locations []string
	for location := range locationCounts {
		locations = append(locations, location)
	}

	return locations
}

func (s *ReportingService) generateActivitySummary(activities []models.UserActivitySummary) models.ActivitySummary {
	if len(activities) == 0 {
		return models.ActivitySummary{}
	}

	totalAccesses := 0
	for _, activity := range activities {
		totalAccesses += activity.TotalAccesses
	}

	avgAccesses := float64(totalAccesses) / float64(len(activities))

	return models.ActivitySummary{
		TotalUsers:       len(activities),
		TotalAccesses:    totalAccesses,
		AverageAccesses:  avgAccesses,
		MostActiveUsers:  len(activities), // Simplified
		LeastActiveUsers: 0,               // Simplified
	}
}

func (s *ReportingService) calculateSecurityMetrics(startDate, endDate time.Time) models.SecurityMetrics {
	// Placeholder implementation
	return models.SecurityMetrics{
		ThreatLevel:        "low",
		VulnerabilityCount: 0,
		SecurityScore:      95.5,
	}
}

func (s *ReportingService) calculateAverageSessionDuration(sessions []models.SessionData) time.Duration {
	if len(sessions) == 0 {
		return 0
	}

	var total time.Duration
	for _, session := range sessions {
		total += session.Duration
	}

	return total / time.Duration(len(sessions))
}

func (s *ReportingService) calculateResponseTimes(startDate, endDate time.Time) models.ResponseTimes {
	// Placeholder implementation
	return models.ResponseTimes{
		Average: 250.5,
		Min:     50.2,
		Max:     1200.8,
		P95:     480.3,
		P99:     850.7,
	}
}

func (s *ReportingService) calculateSystemLoad(startDate, endDate time.Time) models.SystemLoad {
	// Placeholder implementation
	return models.SystemLoad{
		CPU:     45.2,
		Memory:  68.5,
		Disk:    32.1,
		Network: 15.8,
	}
}

func (s *ReportingService) calculateErrorRates(startDate, endDate time.Time) models.ErrorRates {
	// Placeholder implementation
	return models.ErrorRates{
		HTTP4xx:  2.1,
		HTTP5xx:  0.3,
		Timeouts: 0.1,
		Total:    2.5,
	}
}
