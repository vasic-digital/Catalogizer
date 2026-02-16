package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaimsValid(t *testing.T) {
	tests := []struct {
		name      string
		claims    Claims
		expectErr bool
	}{
		{
			name: "valid token not expired",
			claims: Claims{
				UserID:    1,
				Username:  "testuser",
				Role:      RoleAdmin,
				ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
				IssuedAt:  time.Now().Unix(),
			},
			expectErr: false,
		},
		{
			name: "expired token",
			claims: Claims{
				UserID:    1,
				Username:  "testuser",
				Role:      RoleUser,
				ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
				IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
			},
			expectErr: true,
		},
		{
			name: "token expires exactly now",
			claims: Claims{
				UserID:    1,
				Username:  "testuser",
				Role:      RoleUser,
				ExpiresAt: time.Now().Unix(),
				IssuedAt:  time.Now().Add(-1 * time.Hour).Unix(),
			},
			// time.Now().Unix() > c.ExpiresAt will be false if same second
			expectErr: false,
		},
		{
			name: "token expired one second ago",
			claims: Claims{
				UserID:    1,
				Username:  "testuser",
				Role:      RoleViewer,
				ExpiresAt: time.Now().Add(-1 * time.Second).Unix(),
				IssuedAt:  time.Now().Add(-1 * time.Hour).Unix(),
			},
			expectErr: true,
		},
		{
			name: "far future expiration",
			claims: Claims{
				UserID:    999,
				Username:  "farfuture",
				Role:      RoleAdmin,
				ExpiresAt: time.Now().Add(365 * 24 * time.Hour).Unix(),
				IssuedAt:  time.Now().Unix(),
			},
			expectErr: false,
		},
		{
			name: "zero expiration (epoch)",
			claims: Claims{
				UserID:    1,
				Username:  "testuser",
				Role:      RoleUser,
				ExpiresAt: 0,
				IssuedAt:  0,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.claims.Valid()
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "token expired")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClaimsGetExpirationTime(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt int64
	}{
		{"future time", time.Now().Add(1 * time.Hour).Unix()},
		{"past time", time.Now().Add(-1 * time.Hour).Unix()},
		{"zero time", 0},
		{"large timestamp", 2000000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := Claims{ExpiresAt: tt.expiresAt}
			result, err := claims.GetExpirationTime()
			assert.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, time.Unix(tt.expiresAt, 0), result.Time)
		})
	}
}

func TestClaimsGetIssuedAt(t *testing.T) {
	tests := []struct {
		name     string
		issuedAt int64
	}{
		{"current time", time.Now().Unix()},
		{"past time", time.Now().Add(-24 * time.Hour).Unix()},
		{"zero time", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := Claims{IssuedAt: tt.issuedAt}
			result, err := claims.GetIssuedAt()
			assert.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, time.Unix(tt.issuedAt, 0), result.Time)
		})
	}
}

func TestClaimsGetNotBefore(t *testing.T) {
	claims := Claims{
		UserID:   1,
		Username: "testuser",
	}
	result, err := claims.GetNotBefore()
	assert.NoError(t, err)
	assert.Nil(t, result, "GetNotBefore should return nil")
}

func TestClaimsGetIssuer(t *testing.T) {
	claims := Claims{
		UserID:   1,
		Username: "testuser",
	}
	result, err := claims.GetIssuer()
	assert.NoError(t, err)
	assert.Equal(t, "", result, "GetIssuer should return empty string")
}

func TestClaimsGetSubject(t *testing.T) {
	tests := []struct {
		name     string
		username string
	}{
		{"regular username", "testuser"},
		{"email username", "user@example.com"},
		{"empty username", ""},
		{"username with spaces", "test user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := Claims{Username: tt.username}
			result, err := claims.GetSubject()
			assert.NoError(t, err)
			assert.Equal(t, tt.username, result)
		})
	}
}

func TestClaimsGetAudience(t *testing.T) {
	claims := Claims{
		UserID:   1,
		Username: "testuser",
	}
	result, err := claims.GetAudience()
	assert.NoError(t, err)
	assert.Nil(t, result, "GetAudience should return nil")
}

func TestClaimsImplementsInterface(t *testing.T) {
	// Verify Claims implements jwt.Claims interface
	var _ jwt.Claims = Claims{}
}

func TestGetRolePermissionsComprehensive(t *testing.T) {
	tests := []struct {
		name                string
		role                string
		expectedPermissions []string
		minPermissions      int
	}{
		{
			name: "admin has all permissions",
			role: RoleAdmin,
			expectedPermissions: []string{
				PermissionReadMedia, PermissionWriteMedia, PermissionDeleteMedia,
				PermissionReadCatalog, PermissionWriteCatalog, PermissionDeleteCatalog,
				PermissionTriggerAnalysis, PermissionViewAnalysis,
				PermissionManageUsers, PermissionManageRoles, PermissionViewLogs,
				PermissionSystemAdmin, PermissionAPIAccess, PermissionAPIWrite,
			},
			minPermissions: 14,
		},
		{
			name: "moderator has moderate permissions",
			role: RoleModerator,
			expectedPermissions: []string{
				PermissionReadMedia, PermissionWriteMedia,
				PermissionReadCatalog, PermissionWriteCatalog,
				PermissionTriggerAnalysis, PermissionViewAnalysis,
				PermissionAPIAccess, PermissionAPIWrite,
			},
			minPermissions: 8,
		},
		{
			name: "user has basic permissions",
			role: RoleUser,
			expectedPermissions: []string{
				PermissionReadMedia, PermissionWriteMedia,
				PermissionReadCatalog, PermissionWriteCatalog,
				PermissionViewAnalysis, PermissionAPIAccess,
			},
			minPermissions: 6,
		},
		{
			name: "viewer has read-only permissions",
			role: RoleViewer,
			expectedPermissions: []string{
				PermissionReadMedia, PermissionReadCatalog,
				PermissionViewAnalysis, PermissionAPIAccess,
			},
			minPermissions: 4,
		},
		{
			name:                "unknown role gets minimal permissions",
			role:                "unknown",
			expectedPermissions: []string{PermissionAPIAccess},
			minPermissions:      1,
		},
		{
			name:                "empty role gets minimal permissions",
			role:                "",
			expectedPermissions: []string{PermissionAPIAccess},
			minPermissions:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms := GetRolePermissions(tt.role)
			assert.GreaterOrEqual(t, len(perms), tt.minPermissions)
			assert.Equal(t, tt.expectedPermissions, perms)
		})
	}
}

func TestGetRolePermissionsHierarchy(t *testing.T) {
	// Verify admin has more permissions than moderator
	adminPerms := GetRolePermissions(RoleAdmin)
	modPerms := GetRolePermissions(RoleModerator)
	userPerms := GetRolePermissions(RoleUser)
	viewerPerms := GetRolePermissions(RoleViewer)

	assert.Greater(t, len(adminPerms), len(modPerms), "admin should have more permissions than moderator")
	assert.Greater(t, len(modPerms), len(userPerms), "moderator should have more permissions than user")
	assert.Greater(t, len(userPerms), len(viewerPerms), "user should have more permissions than viewer")
}

func TestGetRolePermissionsAdminExclusivePermissions(t *testing.T) {
	// Verify only admin has management permissions
	adminPerms := GetRolePermissions(RoleAdmin)
	modPerms := GetRolePermissions(RoleModerator)
	userPerms := GetRolePermissions(RoleUser)
	viewerPerms := GetRolePermissions(RoleViewer)

	adminExclusive := []string{
		PermissionManageUsers, PermissionManageRoles,
		PermissionViewLogs, PermissionSystemAdmin,
		PermissionDeleteMedia, PermissionDeleteCatalog,
	}

	for _, perm := range adminExclusive {
		assert.Contains(t, adminPerms, perm, "admin should have %s", perm)
		assert.NotContains(t, modPerms, perm, "moderator should not have %s", perm)
		assert.NotContains(t, userPerms, perm, "user should not have %s", perm)
		assert.NotContains(t, viewerPerms, perm, "viewer should not have %s", perm)
	}
}

func TestUserHasPermission(t *testing.T) {
	tests := []struct {
		name       string
		user       User
		permission string
		expected   bool
	}{
		{
			name: "user has the permission",
			user: User{
				Permissions: []string{PermissionReadMedia, PermissionWriteMedia},
			},
			permission: PermissionReadMedia,
			expected:   true,
		},
		{
			name: "user does not have the permission",
			user: User{
				Permissions: []string{PermissionReadMedia},
			},
			permission: PermissionWriteMedia,
			expected:   false,
		},
		{
			name: "user has no permissions",
			user: User{
				Permissions: []string{},
			},
			permission: PermissionReadMedia,
			expected:   false,
		},
		{
			name: "user has nil permissions",
			user: User{
				Permissions: nil,
			},
			permission: PermissionReadMedia,
			expected:   false,
		},
		{
			name: "check for system admin permission",
			user: User{
				Permissions: []string{PermissionSystemAdmin},
			},
			permission: PermissionSystemAdmin,
			expected:   true,
		},
		{
			name: "empty permission string",
			user: User{
				Permissions: []string{PermissionReadMedia},
			},
			permission: "",
			expected:   false,
		},
		{
			name: "user has many permissions",
			user: User{
				Permissions: GetRolePermissions(RoleAdmin),
			},
			permission: PermissionManageUsers,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.HasPermission(tt.permission)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserIsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		expected bool
	}{
		{
			name:     "admin role",
			user:     User{Role: RoleAdmin},
			expected: true,
		},
		{
			name:     "moderator role",
			user:     User{Role: RoleModerator},
			expected: false,
		},
		{
			name:     "user role",
			user:     User{Role: RoleUser},
			expected: false,
		},
		{
			name:     "viewer role",
			user:     User{Role: RoleViewer},
			expected: false,
		},
		{
			name:     "empty role",
			user:     User{Role: ""},
			expected: false,
		},
		{
			name:     "unknown role",
			user:     User{Role: "superuser"},
			expected: false,
		},
		{
			name:     "case sensitive admin",
			user:     User{Role: "Admin"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.IsAdmin()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserCanAccess(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		resource string
		action   string
		expected bool
	}{
		{
			name: "user has exact permission",
			user: User{
				Permissions: []string{"read:media"},
			},
			resource: "media",
			action:   "read",
			expected: true,
		},
		{
			name: "user does not have permission",
			user: User{
				Permissions: []string{"read:media"},
			},
			resource: "media",
			action:   "write",
			expected: false,
		},
		{
			name: "system admin can access anything",
			user: User{
				Permissions: []string{PermissionSystemAdmin},
			},
			resource: "media",
			action:   "delete",
			expected: true,
		},
		{
			name: "system admin can access users",
			user: User{
				Permissions: []string{PermissionSystemAdmin},
			},
			resource: "users",
			action:   "manage",
			expected: true,
		},
		{
			name: "user with no permissions cannot access",
			user: User{
				Permissions: []string{},
			},
			resource: "media",
			action:   "read",
			expected: false,
		},
		{
			name: "user with nil permissions cannot access",
			user: User{
				Permissions: nil,
			},
			resource: "catalog",
			action:   "write",
			expected: false,
		},
		{
			name: "permission format is action:resource",
			user: User{
				Permissions: []string{"trigger:analysis"},
			},
			resource: "analysis",
			action:   "trigger",
			expected: true,
		},
		{
			name: "catalog write permission",
			user: User{
				Permissions: []string{PermissionWriteCatalog},
			},
			resource: "catalog",
			action:   "write",
			expected: true,
		},
		{
			name: "has specific permission plus system admin",
			user: User{
				Permissions: []string{PermissionReadMedia, PermissionSystemAdmin},
			},
			resource: "anything",
			action:   "any_action",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.CanAccess(tt.resource, tt.action)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoleConstants(t *testing.T) {
	assert.Equal(t, "admin", RoleAdmin)
	assert.Equal(t, "moderator", RoleModerator)
	assert.Equal(t, "user", RoleUser)
	assert.Equal(t, "viewer", RoleViewer)
}

func TestPermissionConstants(t *testing.T) {
	// Media permissions
	assert.Equal(t, "read:media", PermissionReadMedia)
	assert.Equal(t, "write:media", PermissionWriteMedia)
	assert.Equal(t, "delete:media", PermissionDeleteMedia)
	assert.Equal(t, "view:media", PermissionViewMedia)

	// Catalog permissions
	assert.Equal(t, "read:catalog", PermissionReadCatalog)
	assert.Equal(t, "write:catalog", PermissionWriteCatalog)
	assert.Equal(t, "delete:catalog", PermissionDeleteCatalog)

	// Analysis permissions
	assert.Equal(t, "trigger:analysis", PermissionTriggerAnalysis)
	assert.Equal(t, "view:analysis", PermissionViewAnalysis)

	// Admin permissions
	assert.Equal(t, "manage:users", PermissionManageUsers)
	assert.Equal(t, "manage:roles", PermissionManageRoles)
	assert.Equal(t, "view:logs", PermissionViewLogs)
	assert.Equal(t, "admin:system", PermissionSystemAdmin)

	// API permissions
	assert.Equal(t, "access:api", PermissionAPIAccess)
	assert.Equal(t, "write:api", PermissionAPIWrite)
}

func TestUserStruct(t *testing.T) {
	now := time.Now()
	lastLogin := now.Add(-1 * time.Hour)

	user := User{
		ID:           42,
		Username:     "johndoe",
		Email:        "john@example.com",
		PasswordHash: "hashed_password",
		FirstName:    "John",
		LastName:     "Doe",
		Role:         RoleAdmin,
		IsActive:     true,
		LastLogin:    &lastLogin,
		CreatedAt:    now,
		UpdatedAt:    now,
		Permissions:  []string{PermissionReadMedia, PermissionWriteMedia},
	}

	assert.Equal(t, int64(42), user.ID)
	assert.Equal(t, "johndoe", user.Username)
	assert.Equal(t, "john@example.com", user.Email)
	assert.Equal(t, "hashed_password", user.PasswordHash)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
	assert.Equal(t, RoleAdmin, user.Role)
	assert.True(t, user.IsActive)
	assert.NotNil(t, user.LastLogin)
	assert.Equal(t, lastLogin, *user.LastLogin)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
	assert.Len(t, user.Permissions, 2)
}

func TestUserStructNilLastLogin(t *testing.T) {
	user := User{
		ID:       1,
		Username: "testuser",
	}
	assert.Nil(t, user.LastLogin)
}

func TestRoleStruct(t *testing.T) {
	now := time.Now()
	role := Role{
		ID:          1,
		Name:        "admin",
		Description: "Administrator with full access",
		Permissions: []string{PermissionSystemAdmin},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assert.Equal(t, int64(1), role.ID)
	assert.Equal(t, "admin", role.Name)
	assert.Equal(t, "Administrator with full access", role.Description)
	assert.Len(t, role.Permissions, 1)
	assert.Equal(t, now, role.CreatedAt)
}

func TestSessionStruct(t *testing.T) {
	now := time.Now()
	session := Session{
		ID:        "session-uuid-123",
		UserID:    42,
		Token:     "jwt-token-here",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		IPAddress: "192.168.1.100",
		UserAgent: "Mozilla/5.0",
	}

	assert.Equal(t, "session-uuid-123", session.ID)
	assert.Equal(t, int64(42), session.UserID)
	assert.Equal(t, "jwt-token-here", session.Token)
	assert.True(t, session.ExpiresAt.After(now))
	assert.Equal(t, "192.168.1.100", session.IPAddress)
	assert.Equal(t, "Mozilla/5.0", session.UserAgent)
}

func TestLoginRequest(t *testing.T) {
	req := LoginRequest{
		Username: "testuser",
		Password: "testpass",
	}
	assert.Equal(t, "testuser", req.Username)
	assert.Equal(t, "testpass", req.Password)
}

func TestLoginResponse(t *testing.T) {
	resp := LoginResponse{
		User:         User{ID: 1, Username: "test"},
		Token:        "access-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    3600,
	}
	assert.Equal(t, int64(1), resp.User.ID)
	assert.Equal(t, "access-token", resp.Token)
	assert.Equal(t, "refresh-token", resp.RefreshToken)
	assert.Equal(t, int64(3600), resp.ExpiresIn)
}

func TestRegisterRequest(t *testing.T) {
	req := RegisterRequest{
		Username:  "newuser",
		Email:     "new@example.com",
		Password:  "securepassword",
		FirstName: "New",
		LastName:  "User",
	}
	assert.Equal(t, "newuser", req.Username)
	assert.Equal(t, "new@example.com", req.Email)
	assert.Equal(t, "securepassword", req.Password)
	assert.Equal(t, "New", req.FirstName)
	assert.Equal(t, "User", req.LastName)
}

func TestUpdateUserRequest(t *testing.T) {
	firstName := "Updated"
	email := "updated@example.com"
	isActive := false

	req := UpdateUserRequest{
		FirstName: &firstName,
		LastName:  nil,
		Email:     &email,
		Role:      nil,
		IsActive:  &isActive,
	}

	require.NotNil(t, req.FirstName)
	assert.Equal(t, "Updated", *req.FirstName)
	assert.Nil(t, req.LastName)
	require.NotNil(t, req.Email)
	assert.Equal(t, "updated@example.com", *req.Email)
	assert.Nil(t, req.Role)
	require.NotNil(t, req.IsActive)
	assert.False(t, *req.IsActive)
}

func TestChangePasswordRequest(t *testing.T) {
	req := ChangePasswordRequest{
		CurrentPassword: "oldpass",
		NewPassword:     "newpass123",
	}
	assert.Equal(t, "oldpass", req.CurrentPassword)
	assert.Equal(t, "newpass123", req.NewPassword)
}

func TestPermissionStruct(t *testing.T) {
	perm := Permission{
		ID:          1,
		Name:        "read:media",
		Resource:    "media",
		Action:      "read",
		Description: "Read media files",
	}
	assert.Equal(t, int64(1), perm.ID)
	assert.Equal(t, "read:media", perm.Name)
	assert.Equal(t, "media", perm.Resource)
	assert.Equal(t, "read", perm.Action)
	assert.Equal(t, "Read media files", perm.Description)
}

func TestClaimsStruct(t *testing.T) {
	claims := Claims{
		UserID:      42,
		Username:    "testuser",
		Role:        RoleAdmin,
		Permissions: []string{PermissionReadMedia, PermissionWriteMedia},
		Type:        "access",
		IssuedAt:    time.Now().Unix(),
		ExpiresAt:   time.Now().Add(1 * time.Hour).Unix(),
	}

	assert.Equal(t, int64(42), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, RoleAdmin, claims.Role)
	assert.Len(t, claims.Permissions, 2)
	assert.Equal(t, "access", claims.Type)
	assert.NotZero(t, claims.IssuedAt)
	assert.NotZero(t, claims.ExpiresAt)
}

func TestClaimsRefreshType(t *testing.T) {
	claims := Claims{
		UserID:    1,
		Username:  "testuser",
		Role:      RoleUser,
		Type:      "refresh",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	assert.Equal(t, "refresh", claims.Type)
	assert.NoError(t, claims.Valid())
}

func TestCanAccessPermissionFormat(t *testing.T) {
	// Verify CanAccess constructs permission as "action:resource"
	user := User{
		Permissions: []string{"read:media"},
	}

	// This should match because CanAccess builds "read" + ":" + "media" = "read:media"
	assert.True(t, user.CanAccess("media", "read"))

	// Reversed order should not match
	assert.False(t, user.CanAccess("read", "media"))
}

func TestUserMethodsOnZeroValue(t *testing.T) {
	var user User

	assert.False(t, user.HasPermission(PermissionReadMedia))
	assert.False(t, user.IsAdmin())
	assert.False(t, user.CanAccess("media", "read"))
}

func TestAllRolesHaveAPIAccess(t *testing.T) {
	roles := []string{RoleAdmin, RoleModerator, RoleUser, RoleViewer}

	for _, role := range roles {
		perms := GetRolePermissions(role)
		assert.Contains(t, perms, PermissionAPIAccess,
			"role %s should have API access", role)
	}
}

func TestViewerCannotWrite(t *testing.T) {
	perms := GetRolePermissions(RoleViewer)

	writePerms := []string{
		PermissionWriteMedia, PermissionDeleteMedia,
		PermissionWriteCatalog, PermissionDeleteCatalog,
		PermissionTriggerAnalysis, PermissionAPIWrite,
	}

	for _, wp := range writePerms {
		assert.NotContains(t, perms, wp,
			"viewer should not have write permission: %s", wp)
	}
}

// Ensure Claims satisfies the jwt.Claims interface at compile time
var _ jwt.Claims = (*Claims)(nil)
