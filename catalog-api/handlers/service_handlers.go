package handlers

import (
	"net/http"
	"strconv"
	"time"

	"catalogizer/models"
	"catalogizer/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AnalyticsHandler struct {
	service *services.AnalyticsService
	logger  *zap.Logger
}

func NewAnalyticsHandler(service *services.AnalyticsService, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AnalyticsHandler) LogMediaAccess(c *gin.Context) {
	var access models.MediaAccessLog
	if err := c.ShouldBindJSON(&access); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if access.AccessTime.IsZero() {
		access.AccessTime = time.Now()
	}

	if err := h.service.LogMediaAccess(&access); err != nil {
		h.logger.Error("Failed to log media access", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log access"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "logged"})
}

func (h *AnalyticsHandler) LogEvent(c *gin.Context) {
	var event models.AnalyticsEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	if err := h.service.LogEvent(&event); err != nil {
		h.logger.Error("Failed to log analytics event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "logged"})
}

func (h *AnalyticsHandler) GetUserAnalytics(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format"})
		return
	}

	analytics, err := h.service.GetUserAnalytics(userID, startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get user analytics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get analytics"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

func (h *AnalyticsHandler) GetSystemAnalytics(c *gin.Context) {
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format"})
		return
	}

	analytics, err := h.service.GetSystemAnalytics(startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get system analytics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get analytics"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

func (h *AnalyticsHandler) GetMediaAnalytics(c *gin.Context) {
	mediaID, err := strconv.Atoi(c.Param("media_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media ID"})
		return
	}

	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format"})
		return
	}

	analytics, err := h.service.GetMediaAnalytics(mediaID, startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get media analytics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get analytics"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

func (h *AnalyticsHandler) CreateReport(c *gin.Context) {
	var req struct {
		ReportType string                 `json:"report_type" binding:"required"`
		Params     map[string]interface{} `json:"params"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	report, err := h.service.CreateReport(req.ReportType, req.Params)
	if err != nil {
		h.logger.Error("Failed to create report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

type ReportingHandler struct {
	service *services.ReportingService
	logger  *zap.Logger
}

func NewReportingHandler(service *services.ReportingService, logger *zap.Logger) *ReportingHandler {
	return &ReportingHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ReportingHandler) GetUsageReport(c *gin.Context) {
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format"})
		return
	}

	params := map[string]interface{}{
		"start_date": startDate,
		"end_date":   endDate,
	}

	report, err := h.service.GenerateReport("user_analytics", "json", params)
	if err != nil {
		h.logger.Error("Failed to generate usage report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *ReportingHandler) GetPerformanceReport(c *gin.Context) {
	params := map[string]interface{}{}

	report, err := h.service.GenerateReport("system_overview", "json", params)
	if err != nil {
		h.logger.Error("Failed to generate performance report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

type FavoritesHandler struct {
	service *services.FavoritesService
	logger  *zap.Logger
}

func NewFavoritesHandler(service *services.FavoritesService, logger *zap.Logger) *FavoritesHandler {
	return &FavoritesHandler{
		service: service,
		logger:  logger,
	}
}

func (h *FavoritesHandler) ListFavorites(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	mediaType := c.Query("media_type")
	category := c.Query("category")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var entityType *string
	if mediaType != "" {
		entityType = &mediaType
	}
	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	favorites, err := h.service.GetUserFavorites(uid, entityType, categoryPtr, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get favorites", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get favorites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"favorites": favorites,
		"count":     len(favorites),
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *FavoritesHandler) AddFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	var req struct {
		EntityID   int    `json:"entity_id" binding:"required"`
		EntityType string `json:"entity_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	favorite := &models.Favorite{
		UserID:     uid,
		EntityID:   req.EntityID,
		EntityType: req.EntityType,
	}

	_, err := h.service.AddFavorite(uid, favorite)
	if err != nil {
		h.logger.Error("Failed to add favorite", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add favorite"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "added"})
}

func (h *FavoritesHandler) RemoveFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	entityID, err := strconv.Atoi(c.Param("entity_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity ID"})
		return
	}

	entityType := c.Param("entity_type")
	if entityType == "" {
		entityType = c.Query("entity_type")
	}

	if err := h.service.RemoveFavorite(uid, entityType, entityID); err != nil {
		h.logger.Error("Failed to remove favorite", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove favorite"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

func (h *FavoritesHandler) CheckFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	entityID, err := strconv.Atoi(c.Param("entity_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity ID"})
		return
	}

	entityType := c.Param("entity_type")
	if entityType == "" {
		entityType = c.Query("entity_type")
	}

	isFavorite, err := h.service.IsFavorite(uid, entityType, entityID)
	if err != nil {
		h.logger.Error("Failed to check favorite", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check favorite"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_favorite": isFavorite})
}
