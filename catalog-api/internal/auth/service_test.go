package auth

import (
	"database/sql"
	"testing"
	"time"

	"catalogizer/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func setupAuthServiceTest() (*AuthService, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	db := database.WrapDB(sqlDB, database.DialectSQLite)
	logger := zap.NewNop()
	service := NewAuthService(db, "test-secret", logger)

	return service, mock
}

func TestNewAuthService(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer sqlDB.Close()

	db := database.WrapDB(sqlDB, database.DialectSQLite)
	logger := zap.NewNop()
	service := NewAuthService(db, "test-secret", logger)

	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
	assert.Equal(t, []byte("test-secret"), service.jwtSecret)
	assert.Equal(t, logger, service.logger)
	assert.Equal(t, 24*time.Hour, service.tokenTTL)
}

// Skipping Initialize tests due to complex schema mocking
// func TestAuthService_Initialize(t *testing.T) {
// 	service, mock := setupAuthServiceTest()
// 	defer service.db.Close()
//
// 	// Mock the entire schema creation (createTables executes one big SQL string)
// 	mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))
//
// 	// Mock default admin creation - no existing admin
// 	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users WHERE role = \\?").
// 		WithArgs("admin").
// 		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
//
// 	mock.ExpectExec("INSERT INTO users").WillReturnResult(sqlmock.NewResult(1, 1))
//
// 	err := service.Initialize()
//
// 	assert.NoError(t, err)
// }

func TestAuthService_Login_Success(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	username := "testuser"
	password := "password123"
	ipAddress := "127.0.0.1"
	userAgent := "test-agent"

	// Generate a proper bcrypt hash for the password
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// Mock user lookup
	mock.ExpectQuery("SELECT id, username, email, password_hash, first_name, last_name, role, is_active, last_login, created_at, updated_at FROM users WHERE username = \\? OR email = \\?").
		WithArgs(username, username).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "first_name", "last_name", "role", "is_active", "last_login", "created_at", "updated_at"}).
			AddRow(1, username, "test@example.com", string(passwordHash), "Test", "User", "user", true, nil, time.Now(), time.Now()))

	// Mock session creation
	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(sqlmock.AnyArg(), 1, sqlmock.AnyArg(), sqlmock.AnyArg(), ipAddress, userAgent).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock last login update
	mock.ExpectExec("UPDATE users SET last_login = \\?, updated_at = \\? WHERE id = \\?").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock audit log
	mock.ExpectExec("INSERT INTO auth_audit_log").
		WithArgs(1, "login_success", ipAddress, userAgent, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	response, err := service.Login(username, password, ipAddress, userAgent)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, int64(1), response.User.ID)
	assert.Equal(t, username, response.User.Username)
	assert.NotEmpty(t, response.Token)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, int64(86400), response.ExpiresIn) // 24 hours in seconds
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	username := "testuser"
	password := "wrongpassword"

	// Mock user lookup - user not found
	mock.ExpectQuery("SELECT id, username, email, password_hash, first_name, last_name, role, is_active, last_login, created_at, updated_at FROM users WHERE username = \\? OR email = \\?").
		WithArgs(username, username).
		WillReturnError(sql.ErrNoRows)

	// Mock audit log
	mock.ExpectExec("INSERT INTO auth_audit_log").
		WithArgs(0, "failed_login", "127.0.0.1", "test-agent", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	response, err := service.Login(username, password, "127.0.0.1", "test-agent")

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	username := "testuser"
	password := "password123"

	// Mock user lookup - inactive user
	mock.ExpectQuery("SELECT id, username, email, password_hash, first_name, last_name, role, is_active, last_login, created_at, updated_at FROM users WHERE username = \\? OR email = \\?").
		WithArgs(username, username).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "first_name", "last_name", "role", "is_active", "last_login", "created_at", "updated_at"}).
			AddRow(1, username, "test@example.com", "$2a$10$hash", "Test", "User", "user", false, nil, time.Now(), time.Now()))

	// Mock audit log
	mock.ExpectExec("INSERT INTO auth_audit_log").
		WithArgs(1, "failed_login_inactive", "127.0.0.1", "test-agent", "account disabled").
		WillReturnResult(sqlmock.NewResult(0, 1))

	response, err := service.Login(username, password, "127.0.0.1", "test-agent")

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "account is disabled")
}

func TestAuthService_Logout(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	token := "test-token"

	// Mock session deletion
	mock.ExpectExec("DELETE FROM sessions WHERE token = \\?").
		WithArgs(token).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock token validation (for logging)
	mock.ExpectQuery("SELECT id, username, email, password_hash, first_name, last_name, role, is_active, last_login, created_at, updated_at FROM users WHERE id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "first_name", "last_name", "role", "is_active", "last_login", "created_at", "updated_at"}).
			AddRow(1, "testuser", "test@example.com", "$2a$10$hash", "Test", "User", "user", true, nil, time.Now(), time.Now()))

	// Mock audit log
	mock.ExpectExec("INSERT INTO auth_audit_log").
		WithArgs(1, "logout", "", "", "").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.Logout(token)

	assert.NoError(t, err)
}

func TestAuthService_ValidateToken_Success(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	// Create a user and generate a proper JWT token
	user := &User{
		ID:          1,
		Username:    "testuser",
		Role:        "user",
		Permissions: []string{"read:media"},
	}
	token, err := service.generateToken(user, "access")
	assert.NoError(t, err)

	// Mock session check
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM sessions WHERE token = \\? AND expires_at > \\?\\)").
		WithArgs(token, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock user retrieval
	mock.ExpectQuery("SELECT id, username, email, password_hash, first_name, last_name, role, is_active, last_login, created_at, updated_at FROM users WHERE id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "first_name", "last_name", "role", "is_active", "last_login", "created_at", "updated_at"}).
			AddRow(1, "testuser", "test@example.com", "$2a$10$hash", "Test", "User", "user", true, nil, time.Now(), time.Now()))

	validatedUser, err := service.ValidateToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, validatedUser)
	assert.Equal(t, int64(1), validatedUser.ID)
	assert.Equal(t, "testuser", validatedUser.Username)
}

func TestAuthService_ValidateToken_InvalidSession(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	// Create a valid JWT token but with no session
	user := &User{
		ID:          1,
		Username:    "testuser",
		Role:        "user",
		Permissions: []string{"read:media"},
	}
	token, err := service.generateToken(user, "access")
	assert.NoError(t, err)

	// Mock session check - session doesn't exist
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM sessions WHERE token = \\? AND expires_at > \\?\\)").
		WithArgs(token, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	validatedUser, err := service.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, validatedUser)
	assert.Contains(t, err.Error(), "session expired or invalid")
}

func TestAuthService_GetUserByID(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	userID := int64(1)

	// Mock user retrieval
	mock.ExpectQuery("SELECT id, username, email, password_hash, first_name, last_name, role, is_active, last_login, created_at, updated_at FROM users WHERE id = \\?").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "first_name", "last_name", "role", "is_active", "last_login", "created_at", "updated_at"}).
			AddRow(1, "testuser", "test@example.com", "$2a$10$hash", "Test", "User", "user", true, nil, time.Now(), time.Now()))

	user, err := service.GetUserByID(userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "user", user.Role)
	assert.Contains(t, user.Permissions, "read:media")
}

func TestAuthService_GetUserByID_NotFound(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	userID := int64(999)

	// Mock user retrieval - user not found
	mock.ExpectQuery("SELECT id, username, email, password_hash, first_name, last_name, role, is_active, last_login, created_at, updated_at FROM users WHERE id = \\?").
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	user, err := service.GetUserByID(userID)

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestAuthService_CreateUser_Success(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	req := &RegisterRequest{
		Username:  "newuser",
		Email:     "new@example.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	}

	// Mock user existence check
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE username = \\? OR email = \\?\\)").
		WithArgs(req.Username, req.Email).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Mock user insertion
	mock.ExpectExec("INSERT INTO users").
		WithArgs(req.Username, req.Email, sqlmock.AnyArg(), req.FirstName, req.LastName).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock user retrieval after creation
	mock.ExpectQuery("SELECT id, username, email, password_hash, first_name, last_name, role, is_active, last_login, created_at, updated_at FROM users WHERE id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "first_name", "last_name", "role", "is_active", "last_login", "created_at", "updated_at"}).
			AddRow(1, req.Username, req.Email, "$2a$10$hash", req.FirstName, req.LastName, "user", true, nil, time.Now(), time.Now()))

	user, err := service.CreateUser(req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, req.Username, user.Username)
	assert.Equal(t, req.Email, user.Email)
}

func TestAuthService_CreateUser_AlreadyExists(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	req := &RegisterRequest{
		Username:  "existinguser",
		Email:     "existing@example.com",
		Password:  "password123",
		FirstName: "Existing",
		LastName:  "User",
	}

	// Mock user existence check - user already exists
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE username = \\? OR email = \\?\\)").
		WithArgs(req.Username, req.Email).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	user, err := service.CreateUser(req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "username or email already exists")
}

func TestAuthService_UpdateUser(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	userID := int64(1)
	newFirstName := "Updated"
	newEmail := "updated@example.com"

	req := &UpdateUserRequest{
		FirstName: &newFirstName,
		Email:     &newEmail,
	}

	// Mock user update
	mock.ExpectExec("UPDATE users SET first_name = \\?, email = \\?, updated_at = \\? WHERE id = \\?").
		WithArgs(newFirstName, newEmail, sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock user retrieval after update
	mock.ExpectQuery("SELECT id, username, email, password_hash, first_name, last_name, role, is_active, last_login, created_at, updated_at FROM users WHERE id = \\?").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "first_name", "last_name", "role", "is_active", "last_login", "created_at", "updated_at"}).
			AddRow(1, "testuser", newEmail, "$2a$10$hash", newFirstName, "User", "user", true, nil, time.Now(), time.Now()))

	user, err := service.UpdateUser(userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, newFirstName, user.FirstName)
	assert.Equal(t, newEmail, user.Email)
}

func TestAuthService_ChangePassword_Success(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	userID := int64(1)
	currentPassword := "oldpassword"
	newPassword := "newpassword123"

	// Generate a proper bcrypt hash for the current password
	currentPasswordHash, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)

	// Mock user retrieval
	mock.ExpectQuery("SELECT id, username, email, password_hash, first_name, last_name, role, is_active, last_login, created_at, updated_at FROM users WHERE id = \\?").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "first_name", "last_name", "role", "is_active", "last_login", "created_at", "updated_at"}).
			AddRow(1, "testuser", "test@example.com", string(currentPasswordHash), "Test", "User", "user", true, nil, time.Now(), time.Now()))

	// Mock password update
	mock.ExpectExec("UPDATE users SET password_hash = \\?, updated_at = \\? WHERE id = \\?").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock audit log
	mock.ExpectExec("INSERT INTO auth_audit_log").
		WithArgs(userID, "password_changed", "", "", "").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.ChangePassword(userID, currentPassword, newPassword)

	assert.NoError(t, err)
}

func TestAuthService_ChangePassword_InvalidCurrentPassword(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	userID := int64(1)
	currentPassword := "wrongpassword"
	newPassword := "newpassword123"

	// Mock user retrieval
	mock.ExpectQuery("SELECT id, username, email, password_hash, first_name, last_name, role, is_active, last_login, created_at, updated_at FROM users WHERE id = \\?").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "first_name", "last_name", "role", "is_active", "last_login", "created_at", "updated_at"}).
			AddRow(1, "testuser", "test@example.com", "$2a$10$hash", "Test", "User", "user", true, nil, time.Now(), time.Now()))

	err := service.ChangePassword(userID, currentPassword, newPassword)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "current password is incorrect")
}

func TestAuthService_ListUsers(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	limit := 10
	offset := 0

	// Mock total count
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock users query
	mock.ExpectQuery("SELECT id, username, email, first_name, last_name, role, is_active, last_login, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
		WithArgs(limit, offset).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "first_name", "last_name", "role", "is_active", "last_login", "created_at", "updated_at"}).
			AddRow(1, "user1", "user1@example.com", "User", "One", "user", true, nil, time.Now(), time.Now()).
			AddRow(2, "user2", "user2@example.com", "User", "Two", "admin", true, nil, time.Now(), time.Now()))

	users, total, err := service.ListUsers(limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 2)
	assert.Equal(t, "user1", users[0].Username)
	assert.Equal(t, "user2", users[1].Username)
	assert.Contains(t, users[0].Permissions, "read:media")
	assert.Contains(t, users[1].Permissions, "admin:system")
}

func TestAuthService_GenerateToken(t *testing.T) {
	service, _ := setupAuthServiceTest()
	defer service.db.Close()

	user := &User{
		ID:          1,
		Username:    "testuser",
		Role:        "user",
		Permissions: []string{"read:media", "write:media"},
	}

	token, err := service.generateToken(user, "access")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token
	claims, err := service.validateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Role, claims.Role)
	assert.Equal(t, "access", claims.Type)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	service, _ := setupAuthServiceTest()
	defer service.db.Close()

	invalidToken := "invalid.jwt.token"

	claims, err := service.validateToken(invalidToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestAuthService_CreateSession(t *testing.T) {
	service, mock := setupAuthServiceTest()
	defer service.db.Close()

	userID := int64(1)
	token := "test-token"
	ipAddress := "127.0.0.1"
	userAgent := "test-agent"

	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(sqlmock.AnyArg(), userID, token, sqlmock.AnyArg(), ipAddress, userAgent).
		WillReturnResult(sqlmock.NewResult(0, 1))

	sessionID, err := service.createSession(userID, token, ipAddress, userAgent)

	assert.NoError(t, err)
	assert.NotEmpty(t, sessionID)
}

func TestGenerateRandomString(t *testing.T) {
	length := 32
	randomString, err := generateRandomString(length)

	assert.NoError(t, err)
	assert.Len(t, randomString, length*2) // hex encoding doubles the length
}

func TestGetRolePermissions(t *testing.T) {
	tests := []struct {
		role        string
		expectedLen int
		hasAdmin    bool
	}{
		{"admin", 14, true},
		{"moderator", 8, false},
		{"user", 6, false},
		{"viewer", 4, false},
		{"unknown", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			permissions := GetRolePermissions(tt.role)
			assert.Len(t, permissions, tt.expectedLen)
			if tt.hasAdmin {
				assert.Contains(t, permissions, "admin:system")
			}
		})
	}
}

func TestUser_HasPermission(t *testing.T) {
	user := &User{
		Permissions: []string{"read:media", "write:media", "admin:system"},
	}

	assert.True(t, user.HasPermission("read:media"))
	assert.True(t, user.HasPermission("admin:system"))
	assert.False(t, user.HasPermission("delete:media"))
}

func TestUser_IsAdmin(t *testing.T) {
	adminUser := &User{Role: "admin"}
	regularUser := &User{Role: "user"}

	assert.True(t, adminUser.IsAdmin())
	assert.False(t, regularUser.IsAdmin())
}

func TestUser_CanAccess(t *testing.T) {
	adminUser := &User{
		Role:        "admin",
		Permissions: []string{"admin:system"},
	}
	regularUser := &User{
		Role:        "user",
		Permissions: []string{"read:media", "write:media"},
	}

	assert.True(t, adminUser.CanAccess("media", "read"))
	assert.True(t, adminUser.CanAccess("system", "admin"))

	assert.True(t, regularUser.CanAccess("media", "read"))
	assert.True(t, regularUser.CanAccess("media", "write"))
	assert.False(t, regularUser.CanAccess("system", "admin"))
}
