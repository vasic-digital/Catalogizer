package handlers

import (
	"catalogizer/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SMBDiscoveryHandler handles SMB discovery API requests
type SMBDiscoveryHandler struct {
	service *services.SMBDiscoveryService
	logger  *zap.Logger
}

// NewSMBDiscoveryHandler creates a new SMB discovery handler
func NewSMBDiscoveryHandler(service *services.SMBDiscoveryService, logger *zap.Logger) *SMBDiscoveryHandler {
	return &SMBDiscoveryHandler{
		service: service,
		logger:  logger,
	}
}

// DiscoverSharesRequest represents the request to discover SMB shares
type DiscoverSharesRequest struct {
	Host     string  `json:"host" binding:"required"`
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required"`
	Domain   *string `json:"domain"`
}

// TestConnectionRequest represents the request to test SMB connection
type TestConnectionRequest struct {
	Host     string  `json:"host" binding:"required"`
	Port     int     `json:"port"`
	Share    string  `json:"share" binding:"required"`
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required"`
	Domain   *string `json:"domain"`
}

// BrowseShareRequest represents the request to browse SMB share
type BrowseShareRequest struct {
	Host     string  `json:"host" binding:"required"`
	Port     int     `json:"port"`
	Share    string  `json:"share" binding:"required"`
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required"`
	Domain   *string `json:"domain"`
	Path     string  `json:"path"`
}

// DiscoverShares discovers available SMB shares on a host
// @Summary Discover SMB shares
// @Description Discovers available SMB shares on the specified host
// @Tags SMB
// @Accept json
// @Produce json
// @Param request body DiscoverSharesRequest true "Discovery request"
// @Success 200 {array} services.SMBShareInfo
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/smb/discover [post]
func (h *SMBDiscoveryHandler) DiscoverShares(c *gin.Context) {
	var req DiscoverSharesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	h.logger.Info("Discovering SMB shares", zap.String("host", req.Host), zap.String("username", req.Username))

	shares, err := h.service.DiscoverShares(c.Request.Context(), req.Host, req.Username, req.Password, req.Domain)
	if err != nil {
		h.logger.Error("Failed to discover SMB shares", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to discover shares: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, shares)
}

// TestConnection tests an SMB connection
// @Summary Test SMB connection
// @Description Tests connectivity to an SMB share with the provided credentials
// @Tags SMB
// @Accept json
// @Produce json
// @Param request body TestConnectionRequest true "Connection test request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/smb/test [post]
func (h *SMBDiscoveryHandler) TestConnection(c *gin.Context) {
	var req TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Set default port if not provided
	if req.Port == 0 {
		req.Port = 445
	}

	h.logger.Info("Testing SMB connection", zap.String("host", req.Host), zap.String("share", req.Share))

	config := services.SMBConnectionConfig{
		Host:     req.Host,
		Port:     req.Port,
		Share:    req.Share,
		Username: req.Username,
		Password: req.Password,
		Domain:   req.Domain,
	}

	success := h.service.TestConnection(c.Request.Context(), config)

	c.JSON(http.StatusOK, gin.H{
		"success":    success,
		"host":       req.Host,
		"share":      req.Share,
		"username":   req.Username,
		"connection": success,
	})
}

// BrowseShare browses files and directories in an SMB share
// @Summary Browse SMB share
// @Description Lists files and directories in the specified SMB share path
// @Tags SMB
// @Accept json
// @Produce json
// @Param request body BrowseShareRequest true "Browse request"
// @Success 200 {array} services.SMBFileEntry
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/smb/browse [post]
func (h *SMBDiscoveryHandler) BrowseShare(c *gin.Context) {
	var req BrowseShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Set default port if not provided
	if req.Port == 0 {
		req.Port = 445
	}

	// Set default path if not provided
	if req.Path == "" {
		req.Path = "."
	}

	h.logger.Info("Browsing SMB share", zap.String("host", req.Host), zap.String("share", req.Share), zap.String("path", req.Path))

	config := services.SMBConnectionConfig{
		Host:     req.Host,
		Port:     req.Port,
		Share:    req.Share,
		Username: req.Username,
		Password: req.Password,
		Domain:   req.Domain,
	}

	entries, err := h.service.BrowseShare(c.Request.Context(), config, req.Path)
	if err != nil {
		h.logger.Error("Failed to browse SMB share", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to browse share: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, entries)
}

// DiscoverSharesGET discovers SMB shares using GET parameters (for simple testing)
// @Summary Discover SMB shares (GET)
// @Description Discovers available SMB shares using GET parameters
// @Tags SMB
// @Produce json
// @Param host query string true "SMB host"
// @Param username query string true "Username"
// @Param password query string true "Password"
// @Param domain query string false "Domain"
// @Success 200 {array} services.SMBShareInfo
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/smb/discover [get]
func (h *SMBDiscoveryHandler) DiscoverSharesGET(c *gin.Context) {
	host := c.Query("host")
	username := c.Query("username")
	password := c.Query("password")
	domain := c.Query("domain")

	if host == "" || username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "host, username, and password are required"})
		return
	}

	var domainPtr *string
	if domain != "" {
		domainPtr = &domain
	}

	shares, err := h.service.DiscoverShares(c.Request.Context(), host, username, password, domainPtr)
	if err != nil {
		h.logger.Error("Failed to discover SMB shares", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to discover shares: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, shares)
}

// TestConnectionGET tests SMB connection using GET parameters (for simple testing)
// @Summary Test SMB connection (GET)
// @Description Tests SMB connection using GET parameters
// @Tags SMB
// @Produce json
// @Param host query string true "SMB host"
// @Param share query string true "Share name"
// @Param username query string true "Username"
// @Param password query string true "Password"
// @Param domain query string false "Domain"
// @Param port query int false "Port (default 445)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/smb/test [get]
func (h *SMBDiscoveryHandler) TestConnectionGET(c *gin.Context) {
	host := c.Query("host")
	share := c.Query("share")
	username := c.Query("username")
	password := c.Query("password")
	domain := c.Query("domain")
	portStr := c.Query("port")

	if host == "" || share == "" || username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "host, share, username, and password are required"})
		return
	}

	port := 445
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	var domainPtr *string
	if domain != "" {
		domainPtr = &domain
	}

	config := services.SMBConnectionConfig{
		Host:     host,
		Port:     port,
		Share:    share,
		Username: username,
		Password: password,
		Domain:   domainPtr,
	}

	success := h.service.TestConnection(c.Request.Context(), config)

	c.JSON(http.StatusOK, gin.H{
		"success":    success,
		"host":       host,
		"share":      share,
		"username":   username,
		"connection": success,
	})
}