package services

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/hirochachacha/go-smb2"
	"go.uber.org/zap"
)

// SMBShareInfo represents an SMB share
type SMBShareInfo struct {
	Host        string  `json:"host"`
	ShareName   string  `json:"share_name"`
	Path        string  `json:"path"`
	Writable    bool    `json:"writable"`
	Description *string `json:"description"`
}

// SMBFileEntry represents a file or directory in an SMB share
type SMBFileEntry struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	IsDirectory bool      `json:"is_directory"`
	Size        *int64    `json:"size"`
	Modified    *string   `json:"modified"`
}

// SMBConnectionConfig represents SMB connection parameters
type SMBConnectionConfig struct {
	Host     string  `json:"host"`
	Port     int     `json:"port"`
	Share    string  `json:"share"`
	Username string  `json:"username"`
	Password string  `json:"password"`
	Domain   *string `json:"domain"`
}

// SMBDiscoveryService provides SMB share discovery and testing
type SMBDiscoveryService struct {
	logger *zap.Logger
	timeout time.Duration
}

// NewSMBDiscoveryService creates a new SMB discovery service
func NewSMBDiscoveryService(logger *zap.Logger) *SMBDiscoveryService {
	return &SMBDiscoveryService{
		logger:  logger,
		timeout: 10 * time.Second,
	}
}

// DiscoverShares discovers available SMB shares on a host
func (s *SMBDiscoveryService) DiscoverShares(ctx context.Context, host string, username, password string, domain *string) ([]SMBShareInfo, error) {
	s.logger.Info("Discovering SMB shares", zap.String("host", host), zap.String("username", username))

	// Establish connection
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:445", host), s.timeout)
	if err != nil {
		s.logger.Error("Failed to connect to SMB host", zap.String("host", host), zap.Error(err))
		return nil, fmt.Errorf("failed to connect to SMB host %s: %w", host, err)
	}
	defer conn.Close()

	// Create SMB session
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     username,
			Password: password,
			Domain:   getStringValue(domain),
		},
	}

	session, err := d.Dial(conn)
	if err != nil {
		s.logger.Error("Failed to create SMB session", zap.String("host", host), zap.Error(err))
		return nil, fmt.Errorf("failed to create SMB session: %w", err)
	}
	defer session.Logoff()

	// Try to enumerate shares using IPC$ administrative share
	shares, err := s.enumerateShares(session, host)
	if err != nil {
		s.logger.Warn("Failed to enumerate shares via IPC$, falling back to common shares", zap.Error(err))
		// Fallback to common share names
		return s.getCommonShares(ctx, host, username, password, domain), nil
	}

	return shares, nil
}

// enumerateShares attempts to enumerate shares using administrative interfaces
func (s *SMBDiscoveryService) enumerateShares(session *smb2.Session, host string) ([]SMBShareInfo, error) {
	// This is a simplified implementation. In practice, you might need to use
	// Windows administrative APIs through SMB to enumerate shares properly.
	// For now, we'll try to mount IPC$ and see if we can get share information.

	// Try common administrative share names to detect existence
	commonShares := []string{
		"C$", "D$", "E$", "F$", "admin$", "print$", "ipc$",
		"shared", "public", "media", "downloads", "documents",
		"music", "videos", "pictures", "backup", "data",
	}

	var availableShares []SMBShareInfo

	for _, shareName := range commonShares {
		if s.testShareAccess(session, shareName) {
			availableShares = append(availableShares, SMBShareInfo{
				Host:        host,
				ShareName:   shareName,
				Path:        fmt.Sprintf("\\\\%s\\%s", host, shareName),
				Writable:    false, // We don't test write access here
				Description: getShareDescription(shareName),
			})
		}
	}

	return availableShares, nil
}

// testShareAccess tests if a share can be accessed
func (s *SMBDiscoveryService) testShareAccess(session *smb2.Session, shareName string) bool {
	share, err := session.Mount(shareName)
	if err != nil {
		return false
	}
	defer share.Umount()

	// Try to list the root directory
	_, err = share.ReadDir(".")
	return err == nil
}

// getCommonShares returns common share names to try
func (s *SMBDiscoveryService) getCommonShares(ctx context.Context, host, username, password string, domain *string) []SMBShareInfo {
	commonShares := []SMBShareInfo{
		{Host: host, ShareName: "shared", Path: fmt.Sprintf("\\\\%s\\shared", host), Description: smbStringPtr("Shared folder")},
		{Host: host, ShareName: "public", Path: fmt.Sprintf("\\\\%s\\public", host), Description: smbStringPtr("Public folder")},
		{Host: host, ShareName: "media", Path: fmt.Sprintf("\\\\%s\\media", host), Description: smbStringPtr("Media files")},
		{Host: host, ShareName: "downloads", Path: fmt.Sprintf("\\\\%s\\downloads", host), Description: smbStringPtr("Downloads")},
		{Host: host, ShareName: "documents", Path: fmt.Sprintf("\\\\%s\\documents", host), Description: smbStringPtr("Documents")},
		{Host: host, ShareName: "music", Path: fmt.Sprintf("\\\\%s\\music", host), Description: smbStringPtr("Music files")},
		{Host: host, ShareName: "videos", Path: fmt.Sprintf("\\\\%s\\videos", host), Description: smbStringPtr("Video files")},
		{Host: host, ShareName: "pictures", Path: fmt.Sprintf("\\\\%s\\pictures", host), Description: smbStringPtr("Pictures")},
		{Host: host, ShareName: "backup", Path: fmt.Sprintf("\\\\%s\\backup", host), Description: smbStringPtr("Backup files")},
	}

	// Test which ones are actually accessible
	var accessibleShares []SMBShareInfo
	for _, share := range commonShares {
		if s.TestConnection(ctx, SMBConnectionConfig{
			Host:     host,
			Port:     445,
			Share:    share.ShareName,
			Username: username,
			Password: password,
			Domain:   domain,
		}) {
			accessibleShares = append(accessibleShares, share)
		}
	}

	return accessibleShares
}

// TestConnection tests an SMB connection with the provided credentials
func (s *SMBDiscoveryService) TestConnection(ctx context.Context, config SMBConnectionConfig) bool {
	s.logger.Info("Testing SMB connection",
		zap.String("host", config.Host),
		zap.String("share", config.Share),
		zap.String("username", config.Username))

	// Establish connection
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), s.timeout)
	if err != nil {
		s.logger.Debug("Failed to connect to SMB host", zap.String("host", config.Host), zap.Error(err))
		return false
	}
	defer conn.Close()

	// Create SMB session
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     config.Username,
			Password: config.Password,
			Domain:   getStringValue(config.Domain),
		},
	}

	session, err := d.Dial(conn)
	if err != nil {
		s.logger.Debug("Failed to create SMB session", zap.String("host", config.Host), zap.Error(err))
		return false
	}
	defer session.Logoff()

	// Try to mount the share
	share, err := session.Mount(config.Share)
	if err != nil {
		s.logger.Debug("Failed to mount SMB share", zap.String("share", config.Share), zap.Error(err))
		return false
	}
	defer share.Umount()

	// Try to list the root directory
	_, err = share.ReadDir(".")
	if err != nil {
		s.logger.Debug("Failed to read SMB share directory", zap.String("share", config.Share), zap.Error(err))
		return false
	}

	s.logger.Info("SMB connection test successful", zap.String("host", config.Host), zap.String("share", config.Share))
	return true
}

// BrowseShare browses files and directories in an SMB share
func (s *SMBDiscoveryService) BrowseShare(ctx context.Context, config SMBConnectionConfig, path string) ([]SMBFileEntry, error) {
	s.logger.Info("Browsing SMB share",
		zap.String("host", config.Host),
		zap.String("share", config.Share),
		zap.String("path", path))

	// Establish connection
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), s.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SMB host %s: %w", config.Host, err)
	}
	defer conn.Close()

	// Create SMB session
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     config.Username,
			Password: config.Password,
			Domain:   getStringValue(config.Domain),
		},
	}

	session, err := d.Dial(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMB session: %w", err)
	}
	defer session.Logoff()

	// Mount the share
	share, err := session.Mount(config.Share)
	if err != nil {
		return nil, fmt.Errorf("failed to mount SMB share: %w", err)
	}
	defer share.Umount()

	// List directory contents
	entries, err := share.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	// Convert to our format
	var fileEntries []SMBFileEntry
	for _, entry := range entries {
		var size *int64
		if !entry.IsDir() {
			entrySize := entry.Size()
			size = &entrySize
		}

		modTime := entry.ModTime().Format("2006-01-02 15:04:05")

		fileEntries = append(fileEntries, SMBFileEntry{
			Name:        entry.Name(),
			Path:        path + "/" + entry.Name(),
			IsDirectory: entry.IsDir(),
			Size:        size,
			Modified:    &modTime,
		})
	}

	return fileEntries, nil
}

// Helper functions
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func smbStringPtr(s string) *string {
	return &s
}

func getShareDescription(shareName string) *string {
	descriptions := map[string]string{
		"C$":        "System drive (administrative)",
		"D$":        "Data drive (administrative)",
		"E$":        "Additional drive (administrative)",
		"F$":        "Additional drive (administrative)",
		"admin$":    "Administrative share",
		"print$":    "Printer drivers",
		"ipc$":      "Inter-process communication",
		"shared":    "Shared folder",
		"public":    "Public folder",
		"media":     "Media files",
		"downloads": "Downloads",
		"documents": "Documents",
		"music":     "Music files",
		"videos":    "Video files",
		"pictures":  "Pictures",
		"backup":    "Backup files",
		"data":      "Data files",
	}

	if desc, exists := descriptions[strings.ToLower(shareName)]; exists {
		return &desc
	}
	return nil
}