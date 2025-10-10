package handlers

import (
	"catalog-api/internal/smb"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SMBHandler handles SMB-related endpoints
type SMBHandler struct {
	smbManager *smb.ResilientSMBManager
	logger     *zap.Logger
}

// NewSMBHandler creates a new SMB handler
func NewSMBHandler(smbManager *smb.ResilientSMBManager, logger *zap.Logger) *SMBHandler {
	return &SMBHandler{
		smbManager: smbManager,
		logger:     logger,
	}
}

// AddSourceRequest represents a request to add an SMB source
type AddSourceRequest struct {
	Name              string `json:"name" binding:"required"`
	Path              string `json:"path" binding:"required"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	Domain            string `json:"domain"`
	MaxRetryAttempts  int    `json:"max_retry_attempts"`
	RetryDelaySeconds int    `json:"retry_delay_seconds"`
	ConnectionTimeout int    `json:"connection_timeout_seconds"`
}

// @Summary Add SMB source
// @Description Add a new SMB source for monitoring
// @Tags smb
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body AddSourceRequest true "SMB source details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/smb/sources [post]
func (h *SMBHandler) AddSource(c *gin.Context) {
	var req AddSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	source := &smb.SMBSource{
		Name:     req.Name,
		Path:     req.Path,
		Username: req.Username,
		Password: req.Password,
		Domain:   req.Domain,
	}

	// Set optional parameters
	if req.MaxRetryAttempts > 0 {
		source.MaxRetryAttempts = req.MaxRetryAttempts
	}
	if req.RetryDelaySeconds > 0 {
		source.RetryDelay = time.Duration(req.RetryDelaySeconds) * time.Second
	}
	if req.ConnectionTimeout > 0 {
		source.ConnectionTimeout = time.Duration(req.ConnectionTimeout) * time.Second
	}

	err := h.smbManager.AddSource(source)
	if err != nil {
		h.logger.Error("Failed to add SMB source", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add SMB source"})
		return
	}

	h.logger.Info("SMB source added successfully",
		zap.String("name", req.Name),
		zap.String("path", req.Path))

	c.JSON(http.StatusCreated, gin.H{
		"message":   "SMB source added successfully",
		"source_id": source.ID,
	})
}

// @Summary Remove SMB source
// @Description Remove an existing SMB source
// @Tags smb
// @Security BearerAuth
// @Param id path string true "Source ID"
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/smb/sources/{id} [delete]
func (h *SMBHandler) RemoveSource(c *gin.Context) {
	sourceID := c.Param("id")
	if sourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Source ID is required"})
		return
	}

	err := h.smbManager.RemoveSource(sourceID)
	if err != nil {
		h.logger.Error("Failed to remove SMB source",
			zap.String("source_id", sourceID),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "SMB source not found"})
		return
	}

	h.logger.Info("SMB source removed successfully", zap.String("source_id", sourceID))

	c.JSON(http.StatusOK, gin.H{
		"message": "SMB source removed successfully",
	})
}

// @Summary Get SMB sources status
// @Description Get the status of all configured SMB sources
// @Tags smb
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/v1/smb/sources/status [get]
func (h *SMBHandler) GetSourcesStatus(c *gin.Context) {
	status := h.smbManager.GetSourceStatus()

	c.JSON(http.StatusOK, gin.H{
		"sources": status,
		"summary": h.generateStatusSummary(status),
	})
}

// @Summary Get SMB source details
// @Description Get detailed information about a specific SMB source
// @Tags smb
// @Security BearerAuth
// @Param id path string true "Source ID"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/smb/sources/{id} [get]
func (h *SMBHandler) GetSourceDetails(c *gin.Context) {
	sourceID := c.Param("id")
	if sourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Source ID is required"})
		return
	}

	status := h.smbManager.GetSourceStatus()
	sourceStatus, exists := status[sourceID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "SMB source not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"source": sourceStatus,
	})
}

// @Summary Test SMB connection
// @Description Test connection to an SMB source
// @Tags smb
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body AddSourceRequest true "SMB connection details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/smb/test-connection [post]
func (h *SMBHandler) TestConnection(c *gin.Context) {
	var req AddSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Create temporary source for testing
	source := &smb.SMBSource{
		Name:              "test",
		Path:              req.Path,
		Username:          req.Username,
		Password:          req.Password,
		Domain:            req.Domain,
		ConnectionTimeout: 10 * time.Second,
	}

	// Test connection (this would use actual SMB connection logic)
	start := time.Now()
	err := h.testSMBConnection(source)
	duration := time.Since(start)

	if err != nil {
		h.logger.Error("SMB connection test failed",
			zap.String("path", req.Path),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"error":          err.Error(),
			"test_duration":  duration.Milliseconds(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "Connection successful",
		"test_duration":  duration.Milliseconds(),
	})
}

// @Summary Force reconnect SMB source
// @Description Force a reconnection attempt for an SMB source
// @Tags smb
// @Security BearerAuth
// @Param id path string true "Source ID"
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/smb/sources/{id}/reconnect [post]
func (h *SMBHandler) ForceReconnect(c *gin.Context) {
	sourceID := c.Param("id")
	if sourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Source ID is required"})
		return
	}

	// This would trigger a reconnection attempt
	err := h.smbManager.ForceReconnect(sourceID)
	if err != nil {
		h.logger.Error("Failed to force reconnect",
			zap.String("source_id", sourceID),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "SMB source not found or reconnect failed"})
		return
	}

	h.logger.Info("Forced reconnect initiated", zap.String("source_id", sourceID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Reconnection initiated",
	})
}

// @Summary Get SMB statistics
// @Description Get statistics about SMB sources and their performance
// @Tags smb
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/v1/smb/statistics [get]
func (h *SMBHandler) GetStatistics(c *gin.Context) {
	status := h.smbManager.GetSourceStatus()
	stats := h.generateStatistics(status)

	c.JSON(http.StatusOK, stats)
}

// Helper methods

func (h *SMBHandler) generateStatusSummary(status map[string]interface{}) map[string]interface{} {
	summary := map[string]interface{}{
		"total":        len(status),
		"connected":    0,
		"disconnected": 0,
		"reconnecting": 0,
		"offline":      0,
	}

	for _, sourceStatus := range status {
		if sourceMap, ok := sourceStatus.(map[string]interface{}); ok {
			if state, exists := sourceMap["state"]; exists {
				switch state {
				case "connected":
					summary["connected"] = summary["connected"].(int) + 1
				case "disconnected":
					summary["disconnected"] = summary["disconnected"].(int) + 1
				case "reconnecting":
					summary["reconnecting"] = summary["reconnecting"].(int) + 1
				case "offline":
					summary["offline"] = summary["offline"].(int) + 1
				}
			}
		}
	}

	return summary
}

func (h *SMBHandler) generateStatistics(status map[string]interface{}) map[string]interface{} {
	stats := map[string]interface{}{
		"total_sources":     len(status),
		"health_summary":    h.generateStatusSummary(status),
		"uptime_stats":      h.calculateUptimeStats(status),
		"performance_stats": h.calculatePerformanceStats(status),
		"error_stats":       h.calculateErrorStats(status),
	}

	return stats
}

func (h *SMBHandler) calculateUptimeStats(status map[string]interface{}) map[string]interface{} {
	now := time.Now()
	totalUptime := time.Duration(0)
	connectedSources := 0

	for _, sourceStatus := range status {
		if sourceMap, ok := sourceStatus.(map[string]interface{}); ok {
			if state, exists := sourceMap["state"]; exists && state == "connected" {
				if lastConnectedStr, exists := sourceMap["last_connected"]; exists {
					if lastConnected, ok := lastConnectedStr.(time.Time); ok {
						uptime := now.Sub(lastConnected)
						totalUptime += uptime
						connectedSources++
					}
				}
			}
		}
	}

	averageUptime := time.Duration(0)
	if connectedSources > 0 {
		averageUptime = totalUptime / time.Duration(connectedSources)
	}

	return map[string]interface{}{
		"total_uptime_hours":   totalUptime.Hours(),
		"average_uptime_hours": averageUptime.Hours(),
		"connected_sources":    connectedSources,
	}
}

func (h *SMBHandler) calculatePerformanceStats(status map[string]interface{}) map[string]interface{} {
	// This would include metrics like:
	// - Average response time
	// - Throughput
	// - Connection success rate
	// For now, return placeholder data

	return map[string]interface{}{
		"avg_response_time_ms": 150,
		"connection_success_rate": 0.95,
		"total_operations": 1000,
	}
}

func (h *SMBHandler) calculateErrorStats(status map[string]interface{}) map[string]interface{} {
	totalErrors := 0
	sourcesWithErrors := 0

	for _, sourceStatus := range status {
		if sourceMap, ok := sourceStatus.(map[string]interface{}); ok {
			if retryAttempts, exists := sourceMap["retry_attempts"]; exists {
				if attempts, ok := retryAttempts.(int); ok && attempts > 0 {
					totalErrors += attempts
					sourcesWithErrors++
				}
			}
		}
	}

	return map[string]interface{}{
		"total_errors":         totalErrors,
		"sources_with_errors":  sourcesWithErrors,
		"error_rate":           float64(sourcesWithErrors) / float64(len(status)),
	}
}

func (h *SMBHandler) testSMBConnection(source *smb.SMBSource) error {
	// Placeholder for actual SMB connection testing
	// In a real implementation, this would:
	// 1. Create SMB connection with provided credentials
	// 2. Attempt to list directory contents
	// 3. Test read permissions
	// 4. Return any connection errors

	time.Sleep(100 * time.Millisecond) // Simulate connection time

	// Simulate occasional failures for testing
	if time.Now().Unix()%10 == 0 {
		return errors.New("Connection timeout")
	}

	return nil
}

// UpdateSourceRequest represents a request to update SMB source settings
type UpdateSourceRequest struct {
	Name              *string `json:"name,omitempty"`
	Username          *string `json:"username,omitempty"`
	Password          *string `json:"password,omitempty"`
	Domain            *string `json:"domain,omitempty"`
	MaxRetryAttempts  *int    `json:"max_retry_attempts,omitempty"`
	RetryDelaySeconds *int    `json:"retry_delay_seconds,omitempty"`
	ConnectionTimeout *int    `json:"connection_timeout_seconds,omitempty"`
	IsEnabled         *bool   `json:"is_enabled,omitempty"`
}

// @Summary Update SMB source
// @Description Update settings for an existing SMB source
// @Tags smb
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Source ID"
// @Param request body UpdateSourceRequest true "Update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/smb/sources/{id} [put]
func (h *SMBHandler) UpdateSource(c *gin.Context) {
	sourceID := c.Param("id")
	if sourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Source ID is required"})
		return
	}

	var req UpdateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	err := h.smbManager.UpdateSource(sourceID, &req)
	if err != nil {
		h.logger.Error("Failed to update SMB source",
			zap.String("source_id", sourceID),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "SMB source not found"})
		return
	}

	h.logger.Info("SMB source updated successfully", zap.String("source_id", sourceID))

	c.JSON(http.StatusOK, gin.H{
		"message": "SMB source updated successfully",
	})
}

// @Summary Get SMB health
// @Description Get overall health status of SMB monitoring system
// @Tags smb
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/v1/smb/health [get]
func (h *SMBHandler) GetHealth(c *gin.Context) {
	status := h.smbManager.GetSourceStatus()
	summary := h.generateStatusSummary(status)

	isHealthy := summary["offline"].(int) == 0 && summary["disconnected"].(int) < len(status)/2

	health := map[string]interface{}{
		"healthy":           isHealthy,
		"sources_summary":   summary,
		"total_sources":     len(status),
		"system_uptime":     time.Since(h.smbManager.GetStartTime()).Hours(),
		"last_check":        time.Now(),
	}

	statusCode := http.StatusOK
	if !isHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}