package services

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"catalogizer/models"
	"catalogizer/repository"
)

type SyncService struct {
	syncRepo      *repository.SyncRepository
	userRepo      *repository.UserRepository
	authService   *AuthService
	webdavClients map[int]*WebDAVClient
}

func NewSyncService(syncRepo *repository.SyncRepository, userRepo *repository.UserRepository, authService *AuthService) *SyncService {
	return &SyncService{
		syncRepo:      syncRepo,
		userRepo:      userRepo,
		authService:   authService,
		webdavClients: make(map[int]*WebDAVClient),
	}
}

func (s *SyncService) CreateSyncEndpoint(userID int, endpoint *models.SyncEndpoint) (*models.SyncEndpoint, error) {
	endpoint.UserID = userID
	endpoint.CreatedAt = time.Now()
	endpoint.UpdatedAt = time.Now()
	endpoint.Status = models.SyncStatusActive

	if err := s.validateSyncEndpoint(endpoint); err != nil {
		return nil, fmt.Errorf("invalid sync endpoint: %w", err)
	}

	if err := s.testConnection(endpoint); err != nil {
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	id, err := s.syncRepo.CreateEndpoint(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync endpoint: %w", err)
	}

	endpoint.ID = id
	return endpoint, nil
}

func (s *SyncService) GetUserEndpoints(userID int) ([]models.SyncEndpoint, error) {
	return s.syncRepo.GetUserEndpoints(userID)
}

func (s *SyncService) GetEndpoint(endpointID int, userID int) (*models.SyncEndpoint, error) {
	endpoint, err := s.syncRepo.GetEndpoint(endpointID)
	if err != nil {
		return nil, err
	}

	if endpoint.UserID != userID {
		hasPermission, err := s.authService.CheckPermission(userID, models.PermissionViewShares)
		if err != nil || !hasPermission {
			return nil, fmt.Errorf("unauthorized to view this endpoint")
		}
	}

	return endpoint, nil
}

func (s *SyncService) UpdateEndpoint(endpointID int, userID int, updates *models.UpdateSyncEndpointRequest) (*models.SyncEndpoint, error) {
	endpoint, err := s.syncRepo.GetEndpoint(endpointID)
	if err != nil {
		return nil, err
	}

	if endpoint.UserID != userID {
		hasPermission, err := s.authService.CheckPermission(userID, models.PermissionEditShares)
		if err != nil || !hasPermission {
			return nil, fmt.Errorf("unauthorized to update this endpoint")
		}
	}

	if updates.Name != "" {
		endpoint.Name = updates.Name
	}

	if updates.URL != "" {
		endpoint.URL = updates.URL
	}

	if updates.Username != "" {
		endpoint.Username = updates.Username
	}

	if updates.Password != "" {
		endpoint.Password = updates.Password
	}

	if updates.SyncDirection != "" {
		endpoint.SyncDirection = updates.SyncDirection
	}

	if updates.LocalPath != "" {
		endpoint.LocalPath = updates.LocalPath
	}

	if updates.RemotePath != "" {
		endpoint.RemotePath = updates.RemotePath
	}

	if updates.SyncSettings != nil {
		endpoint.SyncSettings = updates.SyncSettings
	}

	if updates.IsActive != nil {
		if *updates.IsActive {
			endpoint.Status = models.SyncStatusActive
		} else {
			endpoint.Status = models.SyncStatusInactive
		}
	}

	endpoint.UpdatedAt = time.Now()

	if err := s.validateSyncEndpoint(endpoint); err != nil {
		return nil, fmt.Errorf("invalid endpoint update: %w", err)
	}

	if err := s.testConnection(endpoint); err != nil {
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	err = s.syncRepo.UpdateEndpoint(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to update endpoint: %w", err)
	}

	return endpoint, nil
}

func (s *SyncService) DeleteEndpoint(endpointID int, userID int) error {
	endpoint, err := s.syncRepo.GetEndpoint(endpointID)
	if err != nil {
		return err
	}

	if endpoint.UserID != userID {
		hasPermission, err := s.authService.CheckPermission(userID, models.PermissionDeleteShares)
		if err != nil || !hasPermission {
			return fmt.Errorf("unauthorized to delete this endpoint")
		}
	}

	return s.syncRepo.DeleteEndpoint(endpointID)
}

func (s *SyncService) StartSync(endpointID int, userID int) (*models.SyncSession, error) {
	endpoint, err := s.syncRepo.GetEndpoint(endpointID)
	if err != nil {
		return nil, err
	}

	if endpoint.UserID != userID {
		hasPermission, err := s.authService.CheckPermission(userID, models.PermissionEditShares)
		if err != nil || !hasPermission {
			return nil, fmt.Errorf("unauthorized to sync this endpoint")
		}
	}

	if endpoint.Status != models.SyncStatusActive {
		return nil, fmt.Errorf("endpoint is not active")
	}

	session := &models.SyncSession{
		EndpointID:  endpointID,
		UserID:      userID,
		Status:      models.SyncSessionStatusRunning,
		StartedAt:   time.Now(),
		SyncType:    models.SyncTypeManual,
	}

	sessionID, err := s.syncRepo.CreateSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync session: %w", err)
	}

	session.ID = sessionID

	go s.performSync(session, endpoint)

	return session, nil
}

func (s *SyncService) performSync(session *models.SyncSession, endpoint *models.SyncEndpoint) {
	defer func() {
		if r := recover(); r != nil {
			s.handleSyncError(session, fmt.Errorf("sync panic: %v", r))
		}
	}()

	var err error

	switch endpoint.Type {
	case models.SyncTypeWebDAV:
		err = s.performWebDAVSync(session, endpoint)
	case models.SyncTypeCloudStorage:
		err = s.performCloudSync(session, endpoint)
	case models.SyncTypeLocal:
		err = s.performLocalSync(session, endpoint)
	default:
		err = fmt.Errorf("unsupported sync type: %s", endpoint.Type)
	}

	if err != nil {
		s.handleSyncError(session, err)
		return
	}

	s.handleSyncSuccess(session)
}

func (s *SyncService) performWebDAVSync(session *models.SyncSession, endpoint *models.SyncEndpoint) error {
	client, err := s.getWebDAVClient(endpoint)
	if err != nil {
		return fmt.Errorf("failed to get WebDAV client: %w", err)
	}

	switch endpoint.SyncDirection {
	case models.SyncDirectionUpload:
		return s.uploadToWebDAV(session, endpoint, client)
	case models.SyncDirectionDownload:
		return s.downloadFromWebDAV(session, endpoint, client)
	case models.SyncDirectionBidirectional:
		if err := s.uploadToWebDAV(session, endpoint, client); err != nil {
			return err
		}
		return s.downloadFromWebDAV(session, endpoint, client)
	default:
		return fmt.Errorf("unsupported sync direction: %s", endpoint.SyncDirection)
	}
}

func (s *SyncService) uploadToWebDAV(session *models.SyncSession, endpoint *models.SyncEndpoint, client *WebDAVClient) error {
	localFiles, err := s.scanLocalFiles(endpoint.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to scan local files: %w", err)
	}

	session.TotalFiles = len(localFiles)
	s.syncRepo.UpdateSession(session)

	for _, localFile := range localFiles {
		relativePath, err := filepath.Rel(endpoint.LocalPath, localFile)
		if err != nil {
			continue
		}

		remotePath := filepath.Join(endpoint.RemotePath, relativePath)
		remotePath = filepath.ToSlash(remotePath)

		if s.shouldSkipFile(localFile, endpoint) {
			continue
		}

		remoteModTime, err := client.GetModTime(remotePath)
		if err == nil {
			localInfo, err := os.Stat(localFile)
			if err != nil {
				continue
			}

			if !localInfo.ModTime().After(remoteModTime) {
				session.SkippedFiles++
				continue
			}
		}

		err = client.UploadFile(localFile, remotePath)
		if err != nil {
			session.FailedFiles++
			s.logSyncError(session, fmt.Sprintf("Failed to upload %s: %v", localFile, err))
		} else {
			session.SyncedFiles++
		}

		s.syncRepo.UpdateSession(session)
	}

	return nil
}

func (s *SyncService) downloadFromWebDAV(session *models.SyncSession, endpoint *models.SyncEndpoint, client *WebDAVClient) error {
	remoteFiles, err := client.ListFiles(endpoint.RemotePath)
	if err != nil {
		return fmt.Errorf("failed to list remote files: %w", err)
	}

	session.TotalFiles += len(remoteFiles)
	s.syncRepo.UpdateSession(session)

	for _, remoteFile := range remoteFiles {
		relativePath, err := filepath.Rel(endpoint.RemotePath, remoteFile.Path)
		if err != nil {
			continue
		}

		localPath := filepath.Join(endpoint.LocalPath, relativePath)

		if s.shouldSkipRemoteFile(remoteFile, endpoint) {
			continue
		}

		localInfo, err := os.Stat(localPath)
		if err == nil {
			if !remoteFile.ModTime.After(localInfo.ModTime()) {
				session.SkippedFiles++
				continue
			}
		}

		err = os.MkdirAll(filepath.Dir(localPath), 0755)
		if err != nil {
			session.FailedFiles++
			continue
		}

		err = client.DownloadFile(remoteFile.Path, localPath)
		if err != nil {
			session.FailedFiles++
			s.logSyncError(session, fmt.Sprintf("Failed to download %s: %v", remoteFile.Path, err))
		} else {
			session.SyncedFiles++
		}

		s.syncRepo.UpdateSession(session)
	}

	return nil
}

func (s *SyncService) performCloudSync(session *models.SyncSession, endpoint *models.SyncEndpoint) error {
	// Placeholder for cloud storage sync (Google Drive, Dropbox, OneDrive)
	return fmt.Errorf("cloud storage sync not yet implemented")
}

func (s *SyncService) performLocalSync(session *models.SyncSession, endpoint *models.SyncEndpoint) error {
	// Placeholder for local folder sync
	return fmt.Errorf("local folder sync not yet implemented")
}

func (s *SyncService) handleSyncSuccess(session *models.SyncSession) {
	session.Status = models.SyncSessionStatusCompleted
	session.CompletedAt = &time.Time{}
	*session.CompletedAt = time.Now()

	if session.StartedAt != (time.Time{}) {
		duration := session.CompletedAt.Sub(session.StartedAt)
		session.Duration = &duration
	}

	s.syncRepo.UpdateSession(session)
	s.notifyUser(session, "Sync completed successfully")
}

func (s *SyncService) handleSyncError(session *models.SyncSession, syncError error) {
	session.Status = models.SyncSessionStatusFailed
	session.CompletedAt = &time.Time{}
	*session.CompletedAt = time.Now()
	errorMsg := syncError.Error()
	session.ErrorMessage = &errorMsg

	if session.StartedAt != (time.Time{}) {
		duration := session.CompletedAt.Sub(session.StartedAt)
		session.Duration = &duration
	}

	s.syncRepo.UpdateSession(session)
	s.notifyUser(session, fmt.Sprintf("Sync failed: %s", syncError.Error()))
}

func (s *SyncService) logSyncError(session *models.SyncSession, message string) {
	// In a full implementation, this would log to a sync error log
	fmt.Printf("Sync error for session %d: %s\n", session.ID, message)
}

func (s *SyncService) notifyUser(session *models.SyncSession, message string) {
	// In a full implementation, this would send notifications
	fmt.Printf("Notification for user %d: %s (Session %d)\n", session.UserID, message, session.ID)
}

func (s *SyncService) GetUserSessions(userID int, limit, offset int) ([]models.SyncSession, error) {
	return s.syncRepo.GetUserSessions(userID, limit, offset)
}

func (s *SyncService) GetSession(sessionID int, userID int) (*models.SyncSession, error) {
	session, err := s.syncRepo.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.UserID != userID {
		hasPermission, err := s.authService.CheckPermission(userID, models.PermissionViewShares)
		if err != nil || !hasPermission {
			return nil, fmt.Errorf("unauthorized to view this session")
		}
	}

	return session, nil
}

func (s *SyncService) ScheduleSync(endpointID int, userID int, schedule *models.SyncSchedule) (*models.SyncSchedule, error) {
	endpoint, err := s.syncRepo.GetEndpoint(endpointID)
	if err != nil {
		return nil, err
	}

	if endpoint.UserID != userID {
		hasPermission, err := s.authService.CheckPermission(userID, models.PermissionEditShares)
		if err != nil || !hasPermission {
			return nil, fmt.Errorf("unauthorized to schedule sync for this endpoint")
		}
	}

	schedule.EndpointID = endpointID
	schedule.UserID = userID
	schedule.CreatedAt = time.Now()
	schedule.IsActive = true

	id, err := s.syncRepo.CreateSchedule(schedule)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync schedule: %w", err)
	}

	schedule.ID = id
	return schedule, nil
}

func (s *SyncService) GetSyncStatistics(userID *int, startDate, endDate time.Time) (*models.SyncStatistics, error) {
	return s.syncRepo.GetStatistics(userID, startDate, endDate)
}

func (s *SyncService) ProcessScheduledSyncs() error {
	schedules, err := s.syncRepo.GetActiveSchedules()
	if err != nil {
		return err
	}

	for _, schedule := range schedules {
		if s.shouldRunSchedule(&schedule) {
			_, err := s.StartSync(schedule.EndpointID, schedule.UserID)
			if err != nil {
				fmt.Printf("Failed to start scheduled sync for endpoint %d: %v\n", schedule.EndpointID, err)
			}
		}
	}

	return nil
}

func (s *SyncService) shouldRunSchedule(schedule *models.SyncSchedule) bool {
	now := time.Now()

	switch schedule.Frequency {
	case models.SyncFrequencyHourly:
		return schedule.LastRun == nil || schedule.LastRun.Add(time.Hour).Before(now)
	case models.SyncFrequencyDaily:
		return schedule.LastRun == nil || schedule.LastRun.Add(24*time.Hour).Before(now)
	case models.SyncFrequencyWeekly:
		return schedule.LastRun == nil || schedule.LastRun.Add(7*24*time.Hour).Before(now)
	case models.SyncFrequencyMonthly:
		return schedule.LastRun == nil || schedule.LastRun.AddDate(0, 1, 0).Before(now)
	default:
		return false
	}
}

func (s *SyncService) validateSyncEndpoint(endpoint *models.SyncEndpoint) error {
	if endpoint.Name == "" {
		return fmt.Errorf("name is required")
	}

	if endpoint.URL == "" {
		return fmt.Errorf("URL is required")
	}

	if endpoint.Type == "" {
		return fmt.Errorf("type is required")
	}

	if endpoint.SyncDirection == "" {
		return fmt.Errorf("sync direction is required")
	}

	if endpoint.LocalPath == "" {
		return fmt.Errorf("local path is required")
	}

	validTypes := []string{models.SyncTypeWebDAV, models.SyncTypeCloudStorage, models.SyncTypeLocal}
	if !s.isValidType(endpoint.Type, validTypes) {
		return fmt.Errorf("invalid sync type: %s", endpoint.Type)
	}

	validDirections := []string{models.SyncDirectionUpload, models.SyncDirectionDownload, models.SyncDirectionBidirectional}
	if !s.isValidType(endpoint.SyncDirection, validDirections) {
		return fmt.Errorf("invalid sync direction: %s", endpoint.SyncDirection)
	}

	return nil
}

func (s *SyncService) isValidType(value string, validValues []string) bool {
	for _, valid := range validValues {
		if value == valid {
			return true
		}
	}
	return false
}

func (s *SyncService) testConnection(endpoint *models.SyncEndpoint) error {
	switch endpoint.Type {
	case models.SyncTypeWebDAV:
		client, err := s.getWebDAVClient(endpoint)
		if err != nil {
			return err
		}
		return client.TestConnection()
	default:
		return nil // Skip test for other types for now
	}
}

func (s *SyncService) getWebDAVClient(endpoint *models.SyncEndpoint) (*WebDAVClient, error) {
	if client, exists := s.webdavClients[endpoint.ID]; exists {
		return client, nil
	}

	client := NewWebDAVClient(endpoint.URL, endpoint.Username, endpoint.Password)
	s.webdavClients[endpoint.ID] = client

	return client, nil
}

func (s *SyncService) scanLocalFiles(path string) ([]string, error) {
	var files []string

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, filePath)
		}

		return nil
	})

	return files, err
}

func (s *SyncService) shouldSkipFile(filePath string, endpoint *models.SyncEndpoint) bool {
	fileName := filepath.Base(filePath)

	// Skip hidden files
	if strings.HasPrefix(fileName, ".") {
		return true
	}

	// Skip temporary files
	if strings.HasSuffix(fileName, ".tmp") || strings.HasSuffix(fileName, ".temp") {
		return true
	}

	// Check file size limits if configured
	if endpoint.SyncSettings != nil {
		// This would parse JSON settings and check file size limits
	}

	return false
}

func (s *SyncService) shouldSkipRemoteFile(file *WebDAVFile, endpoint *models.SyncEndpoint) bool {
	fileName := filepath.Base(file.Path)

	// Skip hidden files
	if strings.HasPrefix(fileName, ".") {
		return true
	}

	// Skip temporary files
	if strings.HasSuffix(fileName, ".tmp") || strings.HasSuffix(fileName, ".temp") {
		return true
	}

	return false
}

func (s *SyncService) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (s *SyncService) CleanupOldSessions(olderThan time.Time) error {
	return s.syncRepo.CleanupSessions(olderThan)
}