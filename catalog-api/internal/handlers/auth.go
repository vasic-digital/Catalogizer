package handlers

import (
	"catalogizer/internal/auth"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *auth.AuthService
	logger      *zap.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.LoginRequest true "Login credentials"
// @Success 200 {object} auth.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	response, err := h.authService.Login(
		req.Username,
		req.Password,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		h.logger.Debug("Login failed",
			zap.String("username", req.Username),
			zap.String("ip", c.ClientIP()),
			zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	h.logger.Info("User logged in successfully",
		zap.String("username", response.User.Username),
		zap.String("ip", c.ClientIP()))

	c.JSON(http.StatusOK, response)
}

// @Summary User logout
// @Description Invalidate user session
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
		return
	}

	token := authHeader[7:] // Remove "Bearer " prefix
	if err := h.authService.Logout(token); err != nil {
		h.logger.Error("Logout failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Logout failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// @Summary User registration
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.RegisterRequest true "Registration details"
// @Success 201 {object} auth.User
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	user, err := h.authService.CreateUser(&req)
	if err != nil {
		h.logger.Debug("Registration failed",
			zap.String("username", req.Username),
			zap.String("email", req.Email),
			zap.Error(err))

		if err.Error() == "username or email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		}
		return
	}

	h.logger.Info("User registered successfully",
		zap.String("username", user.Username),
		zap.String("email", user.Email))

	c.JSON(http.StatusCreated, user)
}

// @Summary Get current user profile
// @Description Get authenticated user's profile information
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} auth.User
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	user, exists := auth.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary Update user profile
// @Description Update authenticated user's profile information
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body auth.UpdateUserRequest true "Profile updates"
// @Success 200 {object} auth.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req auth.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Users can only update their own non-privileged fields
	safeReq := &auth.UpdateUserRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	user, err := h.authService.UpdateUser(userID, safeReq)
	if err != nil {
		h.logger.Error("Profile update failed",
			zap.Int64("user_id", userID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Profile update failed"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary Change password
// @Description Change user's password
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body auth.ChangePasswordRequest true "Password change request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req auth.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	err := h.authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		h.logger.Debug("Password change failed",
			zap.Int64("user_id", userID),
			zap.Error(err))

		if err.Error() == "current password is incorrect" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Password change failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// Admin endpoints

// @Summary List all users
// @Description Get paginated list of all users (admin only)
// @Tags auth
// @Security BearerAuth
// @Param limit query int false "Number of users per page" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/admin/users [get]
func (h *AuthHandler) ListUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100 // Cap at 100
	}

	users, total, err := h.authService.ListUsers(limit, offset)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// @Summary Get user by ID
// @Description Get specific user information (admin only)
// @Tags auth
// @Security BearerAuth
// @Param id path int true "User ID"
// @Produce json
// @Success 200 {object} auth.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/auth/admin/users/{id} [get]
func (h *AuthHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		h.logger.Debug("User not found", zap.Int64("user_id", userID))
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary Update user
// @Description Update user information (admin only)
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Param id path int true "User ID"
// @Param request body auth.UpdateUserRequest true "User updates"
// @Produce json
// @Success 200 {object} auth.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/auth/admin/users/{id} [put]
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req auth.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	user, err := h.authService.UpdateUser(userID, &req)
	if err != nil {
		h.logger.Error("Failed to update user",
			zap.Int64("user_id", userID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Log the admin action
	currentUser, _ := auth.GetCurrentUser(c)
	h.logger.Info("User updated by admin",
		zap.Int64("target_user_id", userID),
		zap.String("admin_username", currentUser.Username))

	c.JSON(http.StatusOK, user)
}

// @Summary Get authentication status
// @Description Check if user is authenticated and return user info
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/status [get]
func (h *AuthHandler) GetAuthStatus(c *gin.Context) {
	user, exists := auth.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"authenticated": false,
			"error":         "Not authenticated",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user":          user,
		"permissions":   user.Permissions,
	})
}

// @Summary Get user permissions
// @Description Get current user's permissions
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/permissions [get]
func (h *AuthHandler) GetPermissions(c *gin.Context) {
	user, exists := auth.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role":        user.Role,
		"permissions": user.Permissions,
		"is_admin":    user.IsAdmin(),
	})
}

// @Summary Get system initialization status
// @Description Check if system has been initialized with admin user
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/auth/init-status [get]
func (h *AuthHandler) GetInitStatus(c *gin.Context) {
	// Check if any admin users exist
	users, _, err := h.authService.ListUsers(1, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check initialization status"})
		return
	}

	hasAdmin := false
	for _, user := range users {
		if user.Role == auth.RoleAdmin {
			hasAdmin = true
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"initialized": hasAdmin,
		"has_admin":   hasAdmin,
		"user_count":  len(users),
	})
}