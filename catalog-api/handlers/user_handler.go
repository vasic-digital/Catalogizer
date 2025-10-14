package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"catalogizer/models"
	"catalogizer/repository"
	"catalogizer/services"
)

type UserHandler struct {
	userRepo    *repository.UserRepository
	authService *services.AuthService
}

func NewUserHandler(userRepo *repository.UserRepository, authService *services.AuthService) *UserHandler {
	return &UserHandler{
		userRepo:    userRepo,
		authService: authService,
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hasPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionUserCreate)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.authService.ValidatePassword(req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	salt, err := h.authService.GenerateSecureToken(16)
	if err != nil {
		http.Error(w, "Failed to generate salt", http.StatusInternalServerError)
		return
	}

	passwordHash, err := h.hashPassword(req.Password, salt)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Salt:         salt,
		RoleID:       req.RoleID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		DisplayName:  req.DisplayName,
		TimeZone:     req.TimeZone,
		Language:     req.Language,
		IsActive:     isActive,
	}

	id, err := h.userRepo.Create(user)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "Username or email already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	user.ID = id
	user.PasswordHash = ""
	user.Salt = ""

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userIDStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if currentUser.ID != userID {
		hasPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionUserView)
		if err != nil {
			http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
			return
		}

		if !hasPermission {
			http.Error(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	role, err := h.userRepo.GetRole(user.RoleID)
	if err != nil {
		http.Error(w, "Failed to get user role", http.StatusInternalServerError)
		return
	}
	user.Role = role

	user.PasswordHash = ""
	user.Salt = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userIDStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if currentUser.ID != userID {
		hasPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionUserUpdate)
		if err != nil {
			http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
			return
		}

		if !hasPermission {
			http.Error(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.FirstName != nil {
		user.FirstName = req.FirstName
	}
	if req.LastName != nil {
		user.LastName = req.LastName
	}
	if req.DisplayName != nil {
		user.DisplayName = req.DisplayName
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.TimeZone != nil {
		user.TimeZone = req.TimeZone
	}
	if req.Language != nil {
		user.Language = req.Language
	}

	if currentUser.ID != userID {
		hasAdminPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionUserManage)
		if err != nil {
			http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
			return
		}

		if hasAdminPermission {
			if req.RoleID != nil {
				user.RoleID = *req.RoleID
			}
			if req.IsActive != nil {
				user.IsActive = *req.IsActive
			}
		}
	}

	// TODO: handle Settings update
	// if req.Settings != nil {
	// 	// marshal to string
	// 	user.Settings = req.Settings
	// }

	err = h.userRepo.Update(user)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "Username or email already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	role, err := h.userRepo.GetRole(user.RoleID)
	if err != nil {
		http.Error(w, "Failed to get user role", http.StatusInternalServerError)
		return
	}
	user.Role = role

	user.PasswordHash = ""
	user.Salt = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hasPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionUserDelete)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	userIDStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if currentUser.ID == userID {
		http.Error(w, "Cannot delete your own account", http.StatusBadRequest)
		return
	}

	err = h.userRepo.Delete(userID)
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hasPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionUserView)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	users, err := h.userRepo.List(limit, offset)
	if err != nil {
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	for i := range users {
		users[i].PasswordHash = ""
		users[i].Salt = ""

		role, err := h.userRepo.GetRole(users[i].RoleID)
		if err == nil {
			users[i].Role = role
		}
	}

	totalCount, err := h.userRepo.Count()
	if err != nil {
		http.Error(w, "Failed to get user count", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"users":       users,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hasPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionUserManage)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	userIDStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	userIDStr = strings.TrimSuffix(userIDStr, "/reset-password")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.authService.ValidatePassword(req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.authService.ResetPassword(userID, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Password reset successfully"})
}

func (h *UserHandler) LockAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hasPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionUserManage)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	userIDStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	userIDStr = strings.TrimSuffix(userIDStr, "/lock")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if currentUser.ID == userID {
		http.Error(w, "Cannot lock your own account", http.StatusBadRequest)
		return
	}

	var req struct {
		LockUntil string `json:"lock_until"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	lockUntil, err := parseTime(req.LockUntil)
	if err != nil {
		http.Error(w, "Invalid lock_until format", http.StatusBadRequest)
		return
	}

	err = h.authService.LockAccount(userID, lockUntil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Account locked successfully"})
}

func (h *UserHandler) UnlockAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser, err := h.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hasPermission, err := h.authService.CheckPermission(currentUser.ID, models.PermissionUserManage)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !hasPermission {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	userIDStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	userIDStr = strings.TrimSuffix(userIDStr, "/unlock")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.authService.UnlockAccount(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Account unlocked successfully"})
}

func (h *UserHandler) getCurrentUser(r *http.Request) (*models.User, error) {
	token := extractToken(r)
	if token == "" {
		return nil, models.ErrUnauthorized
	}

	return h.authService.GetCurrentUser(token)
}

func (h *UserHandler) hashPassword(password, salt string) (string, error) {
	combined := password + salt
	hash := h.authService.HashData(combined)
	return hash, nil
}

func parseTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}
