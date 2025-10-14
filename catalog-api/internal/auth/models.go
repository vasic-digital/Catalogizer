package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User represents a system user
type User struct {
	ID           int64      `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	FirstName    string     `json:"first_name" db:"first_name"`
	LastName     string     `json:"last_name" db:"last_name"`
	Role         string     `json:"role" db:"role"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	LastLogin    *time.Time `json:"last_login,omitempty" db:"last_login"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	Permissions  []string   `json:"permissions,omitempty"`
}

// Role represents user roles
type Role struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Permissions []string  `json:"permissions" db:"permissions"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Session represents user sessions
type Session struct {
	ID        string    `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	User         User   `json:"user"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty"`
	Role      *string `json:"role,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// Permission represents system permissions
type Permission struct {
	ID          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Resource    string `json:"resource" db:"resource"`
	Action      string `json:"action" db:"action"`
	Description string `json:"description" db:"description"`
}

// Claims represents JWT claims
type Claims struct {
	UserID      int64    `json:"user_id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	Type        string   `json:"type"` // access or refresh
	IssuedAt    int64    `json:"iat"`
	ExpiresAt   int64    `json:"exp"`
}

// JWT Claims interface implementation
func (c Claims) Valid() error {
	if time.Now().Unix() > c.ExpiresAt {
		return fmt.Errorf("token expired")
	}
	return nil
}

func (c Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.ExpiresAt, 0)), nil
}

func (c Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.IssuedAt, 0)), nil
}

func (c Claims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (c Claims) GetIssuer() (string, error) {
	return "", nil
}

func (c Claims) GetSubject() (string, error) {
	return c.Username, nil
}

func (c Claims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

// Standard permissions
const (
	// Media permissions
	PermissionReadMedia   = "read:media"
	PermissionWriteMedia  = "write:media"
	PermissionDeleteMedia = "delete:media"
	PermissionViewMedia   = "view:media"

	// Catalog permissions
	PermissionReadCatalog   = "read:catalog"
	PermissionWriteCatalog  = "write:catalog"
	PermissionDeleteCatalog = "delete:catalog"

	// Analysis permissions
	PermissionTriggerAnalysis = "trigger:analysis"
	PermissionViewAnalysis    = "view:analysis"

	// Admin permissions
	PermissionManageUsers = "manage:users"
	PermissionManageRoles = "manage:roles"
	PermissionViewLogs    = "view:logs"
	PermissionSystemAdmin = "admin:system"

	// API permissions
	PermissionAPIAccess = "access:api"
	PermissionAPIWrite  = "write:api"
)

// Standard roles
const (
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
	RoleUser      = "user"
	RoleViewer    = "viewer"
)

// GetRolePermissions returns default permissions for each role
func GetRolePermissions(role string) []string {
	switch role {
	case RoleAdmin:
		return []string{
			PermissionReadMedia, PermissionWriteMedia, PermissionDeleteMedia,
			PermissionReadCatalog, PermissionWriteCatalog, PermissionDeleteCatalog,
			PermissionTriggerAnalysis, PermissionViewAnalysis,
			PermissionManageUsers, PermissionManageRoles, PermissionViewLogs,
			PermissionSystemAdmin, PermissionAPIAccess, PermissionAPIWrite,
		}
	case RoleModerator:
		return []string{
			PermissionReadMedia, PermissionWriteMedia,
			PermissionReadCatalog, PermissionWriteCatalog,
			PermissionTriggerAnalysis, PermissionViewAnalysis,
			PermissionAPIAccess, PermissionAPIWrite,
		}
	case RoleUser:
		return []string{
			PermissionReadMedia, PermissionWriteMedia,
			PermissionReadCatalog, PermissionWriteCatalog,
			PermissionViewAnalysis, PermissionAPIAccess,
		}
	case RoleViewer:
		return []string{
			PermissionReadMedia, PermissionReadCatalog,
			PermissionViewAnalysis, PermissionAPIAccess,
		}
	default:
		return []string{PermissionAPIAccess}
	}
}

// HasPermission checks if user has specific permission
func (u *User) HasPermission(permission string) bool {
	for _, p := range u.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// IsAdmin checks if user is admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanAccess checks if user can access a resource with specific action
func (u *User) CanAccess(resource, action string) bool {
	permission := action + ":" + resource
	return u.HasPermission(permission) || u.HasPermission(PermissionSystemAdmin)
}
