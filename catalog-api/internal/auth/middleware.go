package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	authService *AuthService
	logger      *zap.Logger
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService *AuthService, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid authorization header"})
			c.Abort()
			return
		}

		user, err := m.authService.ValidateToken(token)
		if err != nil {
			m.logger.Debug("Token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("role", user.Role)
		c.Set("permissions", user.Permissions)

		c.Next()
	}
}

// RequirePermission middleware that requires specific permission
func (m *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userObj, ok := user.(*User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
			c.Abort()
			return
		}

		if !userObj.HasPermission(permission) && !userObj.IsAdmin() {
			m.logger.Debug("Permission denied",
				zap.String("username", userObj.Username),
				zap.String("required_permission", permission))
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole middleware that requires specific role
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userObj, ok := user.(*User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, role := range roles {
			if userObj.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole && !userObj.IsAdmin() {
			m.logger.Debug("Role access denied",
				zap.String("username", userObj.Username),
				zap.String("user_role", userObj.Role),
				zap.Strings("required_roles", roles))
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient role privileges"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireResourceAccess middleware for resource-action based permissions
func (m *AuthMiddleware) RequireResourceAccess(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userObj, ok := user.(*User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
			c.Abort()
			return
		}

		if !userObj.CanAccess(resource, action) {
			m.logger.Debug("Resource access denied",
				zap.String("username", userObj.Username),
				zap.String("resource", resource),
				zap.String("action", action))
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "Access denied to resource",
				"resource": resource,
				"action":   action,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware that extracts user if token is present but doesn't require it
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token != "" {
			user, err := m.authService.ValidateToken(token)
			if err == nil {
				// Set user in context if token is valid
				c.Set("user", user)
				c.Set("user_id", user.ID)
				c.Set("username", user.Username)
				c.Set("role", user.Role)
				c.Set("permissions", user.Permissions)
			}
		}
		c.Next()
	}
}

// AdminOnly middleware that requires admin role
func (m *AuthMiddleware) AdminOnly() gin.HandlerFunc {
	return m.RequireRole(RoleAdmin)
}

// ModeratorOrAdmin middleware that requires moderator or admin role
func (m *AuthMiddleware) ModeratorOrAdmin() gin.HandlerFunc {
	return m.RequireRole(RoleModerator, RoleAdmin)
}

// extractToken extracts JWT token from Authorization header
func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

// GetCurrentUser helper function to get current user from context
func GetCurrentUser(c *gin.Context) (*User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	userObj, ok := user.(*User)
	return userObj, ok
}

// GetCurrentUserID helper function to get current user ID from context
func GetCurrentUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(int64)
	return id, ok
}

// HasPermission helper function to check if current user has permission
func HasPermission(c *gin.Context, permission string) bool {
	user, exists := GetCurrentUser(c)
	if !exists {
		return false
	}

	return user.HasPermission(permission) || user.IsAdmin()
}

// IsAdmin helper function to check if current user is admin
func IsAdmin(c *gin.Context) bool {
	user, exists := GetCurrentUser(c)
	if !exists {
		return false
	}

	return user.IsAdmin()
}

// CanAccessResource helper function to check resource access
func CanAccessResource(c *gin.Context, resource, action string) bool {
	user, exists := GetCurrentUser(c)
	if !exists {
		return false
	}

	return user.CanAccess(resource, action)
}

// LogUserActivity logs user activity for audit purposes
func (m *AuthMiddleware) LogUserActivity() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Log after request completion
		user, exists := GetCurrentUser(c)
		if exists {
			m.logger.Info("User activity",
				zap.String("username", user.Username),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.String("ip", c.ClientIP()),
				zap.String("user_agent", c.Request.UserAgent()),
			)
		}
	}
}

// RateLimitByUser implements per-user rate limiting
func (m *AuthMiddleware) RateLimitByUser(requests int, window string) gin.HandlerFunc {
	// This is a placeholder - implement with Redis or in-memory store
	return func(c *gin.Context) {
		user, exists := GetCurrentUser(c)
		if !exists {
			c.Next()
			return
		}

		// TODO: Implement rate limiting logic
		// For now, just pass through
		_ = user
		_ = requests
		_ = window

		c.Next()
	}
}
