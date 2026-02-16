package filesystem

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestFileInfoCreation(t *testing.T) {
	now := time.Now()
	fi := FileInfo{
		Name:    "test.mp4",
		Size:    1024000,
		ModTime: now,
		IsDir:   false,
		Mode:    os.FileMode(0644),
		Path:    "/media/movies/test.mp4",
	}

	if fi.Name != "test.mp4" {
		t.Errorf("expected Name 'test.mp4', got '%s'", fi.Name)
	}
	if fi.Size != 1024000 {
		t.Errorf("expected Size 1024000, got %d", fi.Size)
	}
	if !fi.ModTime.Equal(now) {
		t.Errorf("expected ModTime %v, got %v", now, fi.ModTime)
	}
	if fi.IsDir {
		t.Error("expected IsDir false, got true")
	}
	if fi.Mode != os.FileMode(0644) {
		t.Errorf("expected Mode 0644, got %v", fi.Mode)
	}
	if fi.Path != "/media/movies/test.mp4" {
		t.Errorf("expected Path '/media/movies/test.mp4', got '%s'", fi.Path)
	}
}

func TestFileInfoDirectory(t *testing.T) {
	fi := FileInfo{
		Name:  "movies",
		Size:  0,
		IsDir: true,
		Mode:  os.ModeDir | os.FileMode(0755),
		Path:  "/media/movies",
	}

	if !fi.IsDir {
		t.Error("expected IsDir true for directory")
	}
	if fi.Mode&os.ModeDir == 0 {
		t.Error("expected ModeDir bit to be set")
	}
}

func TestFileInfoZeroValue(t *testing.T) {
	var fi FileInfo

	if fi.Name != "" {
		t.Errorf("expected zero-value Name to be empty, got '%s'", fi.Name)
	}
	if fi.Size != 0 {
		t.Errorf("expected zero-value Size to be 0, got %d", fi.Size)
	}
	if fi.IsDir {
		t.Error("expected zero-value IsDir to be false")
	}
	if fi.Path != "" {
		t.Errorf("expected zero-value Path to be empty, got '%s'", fi.Path)
	}
}

func TestStorageConfigJSON(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	config := StorageConfig{
		ID:       "store-001",
		Name:     "Local Media",
		Protocol: "local",
		Enabled:  true,
		MaxDepth: 5,
		Settings: map[string]interface{}{
			"root_path": "/media/library",
			"recursive": true,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal StorageConfig: %v", err)
	}

	var decoded StorageConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal StorageConfig: %v", err)
	}

	if decoded.ID != config.ID {
		t.Errorf("expected ID '%s', got '%s'", config.ID, decoded.ID)
	}
	if decoded.Name != config.Name {
		t.Errorf("expected Name '%s', got '%s'", config.Name, decoded.Name)
	}
	if decoded.Protocol != config.Protocol {
		t.Errorf("expected Protocol '%s', got '%s'", config.Protocol, decoded.Protocol)
	}
	if decoded.Enabled != config.Enabled {
		t.Errorf("expected Enabled %v, got %v", config.Enabled, decoded.Enabled)
	}
	if decoded.MaxDepth != config.MaxDepth {
		t.Errorf("expected MaxDepth %d, got %d", config.MaxDepth, decoded.MaxDepth)
	}
	if decoded.Settings["root_path"] != "/media/library" {
		t.Errorf("expected Settings root_path '/media/library', got '%v'", decoded.Settings["root_path"])
	}
	if !decoded.CreatedAt.Equal(config.CreatedAt) {
		t.Errorf("expected CreatedAt %v, got %v", config.CreatedAt, decoded.CreatedAt)
	}
	if !decoded.UpdatedAt.Equal(config.UpdatedAt) {
		t.Errorf("expected UpdatedAt %v, got %v", config.UpdatedAt, decoded.UpdatedAt)
	}
}

func TestStorageConfigJSONAllProtocols(t *testing.T) {
	protocols := []string{"smb", "ftp", "nfs", "webdav", "local"}
	for _, proto := range protocols {
		t.Run(proto, func(t *testing.T) {
			config := StorageConfig{
				ID:       "id-" + proto,
				Name:     proto + " storage",
				Protocol: proto,
				Enabled:  true,
				MaxDepth: 3,
				Settings: map[string]interface{}{},
			}

			data, err := json.Marshal(config)
			if err != nil {
				t.Fatalf("failed to marshal config for protocol %s: %v", proto, err)
			}

			var decoded StorageConfig
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("failed to unmarshal config for protocol %s: %v", proto, err)
			}

			if decoded.Protocol != proto {
				t.Errorf("expected Protocol '%s', got '%s'", proto, decoded.Protocol)
			}
		})
	}
}

func TestStorageConfigJSONFromString(t *testing.T) {
	raw := `{
		"id": "smb-nas",
		"name": "NAS Share",
		"protocol": "smb",
		"enabled": true,
		"max_depth": 10,
		"settings": {
			"host": "192.168.1.100",
			"share": "media",
			"port": 445
		},
		"created_at": "2025-01-01T00:00:00Z",
		"updated_at": "2025-06-15T12:00:00Z"
	}`

	var config StorageConfig
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		t.Fatalf("failed to unmarshal JSON string: %v", err)
	}

	if config.ID != "smb-nas" {
		t.Errorf("expected ID 'smb-nas', got '%s'", config.ID)
	}
	if config.Protocol != "smb" {
		t.Errorf("expected Protocol 'smb', got '%s'", config.Protocol)
	}
	if config.MaxDepth != 10 {
		t.Errorf("expected MaxDepth 10, got %d", config.MaxDepth)
	}
	if config.Settings["host"] != "192.168.1.100" {
		t.Errorf("expected Settings host '192.168.1.100', got '%v'", config.Settings["host"])
	}
	// JSON numbers decode as float64
	if port, ok := config.Settings["port"].(float64); !ok || port != 445 {
		t.Errorf("expected Settings port 445, got '%v'", config.Settings["port"])
	}
}

func TestStorageConfigDisabled(t *testing.T) {
	config := StorageConfig{
		ID:       "disabled-store",
		Name:     "Disabled",
		Protocol: "ftp",
		Enabled:  false,
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded StorageConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Enabled {
		t.Error("expected Enabled to be false after round-trip")
	}
}

func TestCopyOperationFields(t *testing.T) {
	op := CopyOperation{
		SourcePath:        "/media/source/file.mkv",
		DestinationPath:   "/media/dest/file.mkv",
		OverwriteExisting: true,
	}

	if op.SourcePath != "/media/source/file.mkv" {
		t.Errorf("expected SourcePath '/media/source/file.mkv', got '%s'", op.SourcePath)
	}
	if op.DestinationPath != "/media/dest/file.mkv" {
		t.Errorf("expected DestinationPath '/media/dest/file.mkv', got '%s'", op.DestinationPath)
	}
	if !op.OverwriteExisting {
		t.Error("expected OverwriteExisting true, got false")
	}
}

func TestCopyOperationNoOverwrite(t *testing.T) {
	op := CopyOperation{
		SourcePath:        "/a/b.txt",
		DestinationPath:   "/c/d.txt",
		OverwriteExisting: false,
	}

	if op.OverwriteExisting {
		t.Error("expected OverwriteExisting false, got true")
	}
}

func TestCopyOperationZeroValue(t *testing.T) {
	var op CopyOperation

	if op.SourcePath != "" {
		t.Errorf("expected zero-value SourcePath to be empty, got '%s'", op.SourcePath)
	}
	if op.DestinationPath != "" {
		t.Errorf("expected zero-value DestinationPath to be empty, got '%s'", op.DestinationPath)
	}
	if op.OverwriteExisting {
		t.Error("expected zero-value OverwriteExisting to be false")
	}
}

func TestCopyResultSuccess(t *testing.T) {
	result := CopyResult{
		Success:     true,
		BytesCopied: 5242880,
		Error:       nil,
		TimeTaken:   2 * time.Second,
	}

	if !result.Success {
		t.Error("expected Success true")
	}
	if result.BytesCopied != 5242880 {
		t.Errorf("expected BytesCopied 5242880, got %d", result.BytesCopied)
	}
	if result.Error != nil {
		t.Errorf("expected Error nil, got %v", result.Error)
	}
	if result.TimeTaken != 2*time.Second {
		t.Errorf("expected TimeTaken 2s, got %v", result.TimeTaken)
	}
}

func TestCopyResultFailure(t *testing.T) {
	err := os.ErrPermission
	result := CopyResult{
		Success:     false,
		BytesCopied: 0,
		Error:       err,
		TimeTaken:   100 * time.Millisecond,
	}

	if result.Success {
		t.Error("expected Success false for failed copy")
	}
	if result.BytesCopied != 0 {
		t.Errorf("expected BytesCopied 0, got %d", result.BytesCopied)
	}
	if result.Error == nil {
		t.Error("expected Error to be non-nil")
	}
	if result.Error != os.ErrPermission {
		t.Errorf("expected Error to be os.ErrPermission, got %v", result.Error)
	}
}

func TestCopyResultZeroValue(t *testing.T) {
	var result CopyResult

	if result.Success {
		t.Error("expected zero-value Success to be false")
	}
	if result.BytesCopied != 0 {
		t.Errorf("expected zero-value BytesCopied to be 0, got %d", result.BytesCopied)
	}
	if result.Error != nil {
		t.Errorf("expected zero-value Error to be nil, got %v", result.Error)
	}
	if result.TimeTaken != 0 {
		t.Errorf("expected zero-value TimeTaken to be 0, got %v", result.TimeTaken)
	}
}

func TestDirectoryTreeInfo(t *testing.T) {
	file1 := &FileInfo{Name: "movie.mkv", Size: 1000000, Path: "/media/movies/movie.mkv"}
	file2 := &FileInfo{Name: "cover.jpg", Size: 50000, Path: "/media/movies/cover.jpg"}

	subdir := &DirectoryTreeInfo{
		Path:       "/media/movies/extras",
		TotalFiles: 1,
		TotalDirs:  0,
		TotalSize:  200000,
		MaxDepth:   0,
		Files:      []*FileInfo{{Name: "trailer.mp4", Size: 200000, Path: "/media/movies/extras/trailer.mp4"}},
		Subdirs:    nil,
	}

	tree := DirectoryTreeInfo{
		Path:       "/media/movies",
		TotalFiles: 2,
		TotalDirs:  1,
		TotalSize:  1250000,
		MaxDepth:   1,
		Files:      []*FileInfo{file1, file2},
		Subdirs:    []*DirectoryTreeInfo{subdir},
	}

	if tree.Path != "/media/movies" {
		t.Errorf("expected Path '/media/movies', got '%s'", tree.Path)
	}
	if tree.TotalFiles != 2 {
		t.Errorf("expected TotalFiles 2, got %d", tree.TotalFiles)
	}
	if tree.TotalDirs != 1 {
		t.Errorf("expected TotalDirs 1, got %d", tree.TotalDirs)
	}
	if tree.TotalSize != 1250000 {
		t.Errorf("expected TotalSize 1250000, got %d", tree.TotalSize)
	}
	if tree.MaxDepth != 1 {
		t.Errorf("expected MaxDepth 1, got %d", tree.MaxDepth)
	}
	if len(tree.Files) != 2 {
		t.Errorf("expected 2 Files, got %d", len(tree.Files))
	}
	if len(tree.Subdirs) != 1 {
		t.Errorf("expected 1 Subdir, got %d", len(tree.Subdirs))
	}
	if tree.Subdirs[0].Path != "/media/movies/extras" {
		t.Errorf("expected subdir Path '/media/movies/extras', got '%s'", tree.Subdirs[0].Path)
	}
}
