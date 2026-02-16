package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupUserManagementServer creates a test server with comprehensive user management endpoints
func setupUserManagementServer(t *testing.T) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.HandleMethodNotAllowed = true

	var mu sync.Mutex
	nextID := 3
	users := map[int]map[string]interface{}{
		1: {"id": 1, "username": "admin", "password": "admin123", "email": "admin@test.com", "role": "admin", "is_active": true},
		2: {"id": 2, "username": "user1", "password": "user123", "email": "user1@test.com", "role": "user", "is_active": true},
	}
	tokens := map[string]int{} // token -> userID

	checkAuth := func(c *gin.Context) (int, string, bool) {
		auth := c.GetHeader("Authorization")
		if auth == "" || len(auth) < 8 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return 0, "", false
		}
		token := auth[7:]
		mu.Lock()
		uid, valid := tokens[token]
		var role string
		if valid {
			role = users[uid]["role"].(string)
		}
		mu.Unlock()
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return 0, "", false
		}
		return uid, role, true
	}

	checkAdmin := func(c *gin.Context) (int, bool) {
		uid, role, ok := checkAuth(c)
		if !ok {
			return 0, false
		}
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return 0, false
		}
		return uid, true
	}

	api := router.Group("/api/v1")
	{
		// Auth
		api.POST("/auth/login", func(c *gin.Context) {
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}
			username, _ := data["username"].(string)
			password, _ := data["password"].(string)

			mu.Lock()
			var foundUser map[string]interface{}
			for _, u := range users {
				if u["username"] == username && u["password"] == password {
					if active, ok := u["is_active"].(bool); ok && !active {
						mu.Unlock()
						c.JSON(http.StatusForbidden, gin.H{"error": "Account is deactivated"})
						return
					}
					foundUser = u
					break
				}
			}
			if foundUser == nil {
				mu.Unlock()
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				return
			}
			token := fmt.Sprintf("mgmt-token-%d-%d", foundUser["id"], time.Now().UnixNano())
			tokens[token] = foundUser["id"].(int)
			mu.Unlock()

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"token":   token,
				"user":    gin.H{"id": foundUser["id"], "username": foundUser["username"], "role": foundUser["role"]},
			})
		})

		// User CRUD (admin only for create/update/delete of others)
		api.GET("/users", func(c *gin.Context) {
			if _, ok := checkAdmin(c); !ok {
				return
			}
			mu.Lock()
			result := make([]gin.H, 0, len(users))
			for _, u := range users {
				result = append(result, gin.H{
					"id":        u["id"],
					"username":  u["username"],
					"email":     u["email"],
					"role":      u["role"],
					"is_active": u["is_active"],
				})
			}
			mu.Unlock()
			c.JSON(http.StatusOK, gin.H{"success": true, "data": result, "total": len(result)})
		})

		api.POST("/users", func(c *gin.Context) {
			if _, ok := checkAdmin(c); !ok {
				return
			}
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}
			username, _ := data["username"].(string)
			email, _ := data["email"].(string)
			password, _ := data["password"].(string)
			role, _ := data["role"].(string)
			if username == "" || email == "" || password == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Username, email, and password are required"})
				return
			}
			if role == "" {
				role = "user"
			}

			mu.Lock()
			// Check for duplicate username or email
			for _, u := range users {
				if u["username"] == username {
					mu.Unlock()
					c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
					return
				}
				if u["email"] == email {
					mu.Unlock()
					c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
					return
				}
			}
			id := nextID
			nextID++
			users[id] = map[string]interface{}{
				"id": id, "username": username, "email": email,
				"password": password, "role": role, "is_active": true,
			}
			mu.Unlock()

			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"data":    gin.H{"id": id, "username": username, "email": email, "role": role, "is_active": true},
			})
		})

		api.GET("/users/:id", func(c *gin.Context) {
			_, _, ok := checkAuth(c)
			if !ok {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			mu.Lock()
			u, exists := users[id]
			mu.Unlock()
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    gin.H{"id": u["id"], "username": u["username"], "email": u["email"], "role": u["role"], "is_active": u["is_active"]},
			})
		})

		api.PUT("/users/:id", func(c *gin.Context) {
			if _, ok := checkAdmin(c); !ok {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}

			mu.Lock()
			u, exists := users[id]
			if !exists {
				mu.Unlock()
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			if email, ok := data["email"].(string); ok && email != "" {
				u["email"] = email
			}
			if role, ok := data["role"].(string); ok && role != "" {
				u["role"] = role
			}
			if isActive, ok := data["is_active"].(bool); ok {
				u["is_active"] = isActive
			}
			users[id] = u
			mu.Unlock()

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    gin.H{"id": u["id"], "username": u["username"], "email": u["email"], "role": u["role"], "is_active": u["is_active"]},
			})
		})

		api.DELETE("/users/:id", func(c *gin.Context) {
			callerID, ok := checkAdmin(c)
			if !ok {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			if callerID == id {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
				return
			}

			mu.Lock()
			_, exists := users[id]
			if !exists {
				mu.Unlock()
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			delete(users, id)
			mu.Unlock()

			c.JSON(http.StatusOK, gin.H{"success": true, "message": "User deleted"})
		})

		// Role assignment
		api.PUT("/users/:id/role", func(c *gin.Context) {
			if _, ok := checkAdmin(c); !ok {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}

			role, _ := data["role"].(string)
			validRoles := map[string]bool{"admin": true, "user": true, "viewer": true}
			if !validRoles[role] {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be admin, user, or viewer"})
				return
			}

			mu.Lock()
			u, exists := users[id]
			if !exists {
				mu.Unlock()
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			u["role"] = role
			users[id] = u
			mu.Unlock()

			c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id, "role": role}})
		})

		// Deactivate/Activate user
		api.POST("/users/:id/deactivate", func(c *gin.Context) {
			callerID, ok := checkAdmin(c)
			if !ok {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			if callerID == id {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot deactivate your own account"})
				return
			}

			mu.Lock()
			u, exists := users[id]
			if !exists {
				mu.Unlock()
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			u["is_active"] = false
			users[id] = u
			mu.Unlock()

			c.JSON(http.StatusOK, gin.H{"success": true, "message": "User deactivated"})
		})

		api.POST("/users/:id/activate", func(c *gin.Context) {
			if _, ok := checkAdmin(c); !ok {
				return
			}
			var id int
			fmt.Sscanf(c.Param("id"), "%d", &id)

			mu.Lock()
			u, exists := users[id]
			if !exists {
				mu.Unlock()
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			u["is_active"] = true
			users[id] = u
			mu.Unlock()

			c.JSON(http.StatusOK, gin.H{"success": true, "message": "User activated"})
		})

		// Protected resource for permission testing
		api.GET("/admin/stats", func(c *gin.Context) {
			if _, ok := checkAdmin(c); !ok {
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"total_users": 2}})
		})
	}

	ts := httptest.NewServer(router)
	t.Cleanup(func() { ts.Close() })
	return ts
}

// =============================================================================
// E2E TEST: User CRUD Operations
// =============================================================================

func TestUserManagement_CRUD(t *testing.T) {
	ts := setupUserManagementServer(t)
	ec := newE2EContext(ts.URL)

	// Admin login
	resp := ec.doRequest(t, "POST", "/auth/login", map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	})
	result := ec.parseJSON(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	ec.AuthToken = result["token"].(string)

	var createdUserID int

	t.Run("ListUsers", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/users", nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
		total := int(result["total"].(float64))
		assert.GreaterOrEqual(t, total, 2)
	})

	t.Run("CreateUser", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", "/users", map[string]interface{}{
			"username": "newuser",
			"email":    "newuser@test.com",
			"password": "NewPass123!",
			"role":     "user",
		})
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.True(t, result["success"].(bool))

		data := result["data"].(map[string]interface{})
		createdUserID = int(data["id"].(float64))
		assert.Equal(t, "newuser", data["username"])
		assert.Equal(t, "user", data["role"])
		assert.Equal(t, true, data["is_active"])
	})

	t.Run("CreateDuplicateUser", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", "/users", map[string]interface{}{
			"username": "newuser",
			"email":    "another@test.com",
			"password": "Pass123!",
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("CreateUserMissingFields", func(t *testing.T) {
		resp := ec.doRequest(t, "POST", "/users", map[string]interface{}{
			"username": "incomplete",
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("GetUser", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", fmt.Sprintf("/users/%d", createdUserID), nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data := result["data"].(map[string]interface{})
		assert.Equal(t, "newuser", data["username"])
	})

	t.Run("GetNonExistentUser", func(t *testing.T) {
		resp := ec.doRequest(t, "GET", "/users/9999", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		resp := ec.doRequest(t, "PUT", fmt.Sprintf("/users/%d", createdUserID), map[string]interface{}{
			"email": "updated@test.com",
			"role":  "viewer",
		})
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data := result["data"].(map[string]interface{})
		assert.Equal(t, "updated@test.com", data["email"])
		assert.Equal(t, "viewer", data["role"])
	})

	t.Run("DeleteUser", func(t *testing.T) {
		resp := ec.doRequest(t, "DELETE", fmt.Sprintf("/users/%d", createdUserID), nil)
		result := ec.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
	})

	t.Run("DeleteNonExistentUser", func(t *testing.T) {
		resp := ec.doRequest(t, "DELETE", "/users/9999", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("DeleteSelf", func(t *testing.T) {
		resp := ec.doRequest(t, "DELETE", "/users/1", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// E2E TEST: Role Assignment and Permission Checks
// =============================================================================

func TestUserManagement_RolesAndPermissions(t *testing.T) {
	ts := setupUserManagementServer(t)
	adminCtx := newE2EContext(ts.URL)
	userCtx := newE2EContext(ts.URL)

	// Admin login
	resp := adminCtx.doRequest(t, "POST", "/auth/login", map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	})
	result := adminCtx.parseJSON(t, resp)
	adminCtx.AuthToken = result["token"].(string)

	// Regular user login
	resp = userCtx.doRequest(t, "POST", "/auth/login", map[string]interface{}{
		"username": "user1",
		"password": "user123",
	})
	result = userCtx.parseJSON(t, resp)
	userCtx.AuthToken = result["token"].(string)

	t.Run("AdminCanListUsers", func(t *testing.T) {
		resp := adminCtx.doRequest(t, "GET", "/users", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("UserCannotListUsers", func(t *testing.T) {
		resp := userCtx.doRequest(t, "GET", "/users", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("AdminCanAccessAdminStats", func(t *testing.T) {
		resp := adminCtx.doRequest(t, "GET", "/admin/stats", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("UserCannotAccessAdminStats", func(t *testing.T) {
		resp := userCtx.doRequest(t, "GET", "/admin/stats", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("AdminCanAssignRole", func(t *testing.T) {
		resp := adminCtx.doRequest(t, "PUT", "/users/2/role", map[string]interface{}{
			"role": "viewer",
		})
		result := adminCtx.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
		data := result["data"].(map[string]interface{})
		assert.Equal(t, "viewer", data["role"])
	})

	t.Run("UserCannotAssignRole", func(t *testing.T) {
		resp := userCtx.doRequest(t, "PUT", "/users/2/role", map[string]interface{}{
			"role": "admin",
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("InvalidRoleRejected", func(t *testing.T) {
		resp := adminCtx.doRequest(t, "PUT", "/users/2/role", map[string]interface{}{
			"role": "superadmin",
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("UserCanViewOwnProfile", func(t *testing.T) {
		resp := userCtx.doRequest(t, "GET", "/users/2", nil)
		result := userCtx.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
	})
}

// =============================================================================
// E2E TEST: Account Deactivation and Reactivation
// =============================================================================

func TestUserManagement_AccountActivation(t *testing.T) {
	ts := setupUserManagementServer(t)
	adminCtx := newE2EContext(ts.URL)

	// Admin login
	resp := adminCtx.doRequest(t, "POST", "/auth/login", map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	})
	result := adminCtx.parseJSON(t, resp)
	adminCtx.AuthToken = result["token"].(string)

	t.Run("DeactivateUser", func(t *testing.T) {
		resp := adminCtx.doRequest(t, "POST", "/users/2/deactivate", nil)
		result := adminCtx.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
	})

	t.Run("DeactivatedUserCannotLogin", func(t *testing.T) {
		userCtx := newE2EContext(ts.URL)
		resp := userCtx.doRequest(t, "POST", "/auth/login", map[string]interface{}{
			"username": "user1",
			"password": "user123",
		})
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("CannotDeactivateSelf", func(t *testing.T) {
		resp := adminCtx.doRequest(t, "POST", "/users/1/deactivate", nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("ReactivateUser", func(t *testing.T) {
		resp := adminCtx.doRequest(t, "POST", "/users/2/activate", nil)
		result := adminCtx.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
	})

	t.Run("ReactivatedUserCanLogin", func(t *testing.T) {
		userCtx := newE2EContext(ts.URL)
		resp := userCtx.doRequest(t, "POST", "/auth/login", map[string]interface{}{
			"username": "user1",
			"password": "user123",
		})
		result := userCtx.parseJSON(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, result["success"].(bool))
	})
}
