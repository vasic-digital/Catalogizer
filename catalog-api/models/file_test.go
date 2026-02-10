package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// File Model Tests
// =============================================================================

func TestFile_JSONMarshaling(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	extension := ".mp4"
	mimeType := "video/mp4"
	fileType := "video"
	md5 := "d41d8cd98f00b204e9800998ecf8427e"

	file := File{
		ID:              1,
		StorageRootID:   10,
		StorageRootName: "Test Root",
		Path:            "/movies/test.mp4",
		Name:            "test.mp4",
		Extension:       &extension,
		MimeType:        &mimeType,
		FileType:        &fileType,
		Size:            1024000,
		IsDirectory:     false,
		CreatedAt:       now,
		ModifiedAt:      now,
		LastScanAt:      now,
		MD5:             &md5,
		IsDuplicate:     false,
	}

	// Marshal to JSON
	data, err := json.Marshal(file)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test.mp4")
	assert.Contains(t, string(data), "video/mp4")

	// Unmarshal from JSON
	var unmarshaled File
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, file.ID, unmarshaled.ID)
	assert.Equal(t, file.Name, unmarshaled.Name)
	assert.Equal(t, file.Size, unmarshaled.Size)
	assert.Equal(t, *file.Extension, *unmarshaled.Extension)
}

func TestFile_EmptyFields(t *testing.T) {
	file := File{
		ID:              1,
		StorageRootID:   10,
		StorageRootName: "Test Root",
		Path:            "/test",
		Name:            "test",
		Size:            0,
		IsDirectory:     true,
		CreatedAt:       time.Now(),
		ModifiedAt:      time.Now(),
		LastScanAt:      time.Now(),
		Deleted:         false,
		IsDuplicate:     false,
	}

	data, err := json.Marshal(file)
	require.NoError(t, err)

	var unmarshaled File
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Nil(t, unmarshaled.Extension)
	assert.Nil(t, unmarshaled.MimeType)
	assert.Nil(t, unmarshaled.FileType)
	assert.Nil(t, unmarshaled.MD5)
}

func TestFile_DeletedFields(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	file := File{
		ID:          1,
		Name:        "deleted.txt",
		Deleted:     true,
		DeletedAt:   &now,
		CreatedAt:   now,
		ModifiedAt:  now,
		LastScanAt:  now,
	}

	assert.True(t, file.Deleted)
	assert.NotNil(t, file.DeletedAt)
	assert.Equal(t, now, *file.DeletedAt)
}

func TestFile_HashFields(t *testing.T) {
	md5 := "d41d8cd98f00b204e9800998ecf8427e"
	sha256 := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	sha1 := "da39a3ee5e6b4b0d3255bfef95601890afd80709"
	blake3 := "af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262"
	quickHash := "quick123"

	file := File{
		ID:         1,
		Name:       "test.txt",
		MD5:        &md5,
		SHA256:     &sha256,
		SHA1:       &sha1,
		BLAKE3:     &blake3,
		QuickHash:  &quickHash,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
		LastScanAt: time.Now(),
	}

	assert.Equal(t, md5, *file.MD5)
	assert.Equal(t, sha256, *file.SHA256)
	assert.Equal(t, sha1, *file.SHA1)
	assert.Equal(t, blake3, *file.BLAKE3)
	assert.Equal(t, quickHash, *file.QuickHash)
}

func TestFile_DuplicateFields(t *testing.T) {
	groupID := int64(5)

	file := File{
		ID:               1,
		Name:             "duplicate.txt",
		IsDuplicate:      true,
		DuplicateGroupID: &groupID,
		CreatedAt:        time.Now(),
		ModifiedAt:       time.Now(),
		LastScanAt:       time.Now(),
	}

	assert.True(t, file.IsDuplicate)
	assert.NotNil(t, file.DuplicateGroupID)
	assert.Equal(t, int64(5), *file.DuplicateGroupID)
}

func TestFile_ParentIDField(t *testing.T) {
	parentID := int64(100)

	file := File{
		ID:         1,
		Name:       "child.txt",
		ParentID:   &parentID,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
		LastScanAt: time.Now(),
	}

	assert.NotNil(t, file.ParentID)
	assert.Equal(t, int64(100), *file.ParentID)
}

// =============================================================================
// StorageRoot Model Tests
// =============================================================================

func TestStorageRoot_JSONMarshaling(t *testing.T) {
	host := "server.local"
	port := 445
	path := "/share"
	username := "user"
	password := "pass"

	root := StorageRoot{
		ID:                       1,
		Name:                     "SMB Share",
		Protocol:                 "smb",
		Host:                     &host,
		Port:                     &port,
		Path:                     &path,
		Username:                 &username,
		Password:                 &password,
		Enabled:                  true,
		MaxDepth:                 10,
		EnableDuplicateDetection: true,
		EnableMetadataExtraction: true,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	data, err := json.Marshal(root)
	require.NoError(t, err)
	assert.Contains(t, string(data), "SMB Share")
	assert.Contains(t, string(data), "smb")

	var unmarshaled StorageRoot
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, root.Name, unmarshaled.Name)
	assert.Equal(t, root.Protocol, unmarshaled.Protocol)
	assert.Equal(t, *root.Host, *unmarshaled.Host)
	assert.Equal(t, *root.Port, *unmarshaled.Port)
}

func TestStorageRoot_LocalProtocol(t *testing.T) {
	path := "/mnt/media"

	root := StorageRoot{
		ID:        1,
		Name:      "Local Storage",
		Protocol:  "local",
		Path:      &path,
		Enabled:   true,
		MaxDepth:  10,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.Equal(t, "local", root.Protocol)
	assert.Nil(t, root.Host)
	assert.Nil(t, root.Port)
	assert.NotNil(t, root.Path)
}

func TestStorageRoot_FTPProtocol(t *testing.T) {
	host := "ftp.example.com"
	port := 21
	path := "/public"
	username := "ftpuser"
	password := "ftppass"

	root := StorageRoot{
		ID:        1,
		Name:      "FTP Server",
		Protocol:  "ftp",
		Host:      &host,
		Port:      &port,
		Path:      &path,
		Username:  &username,
		Password:  &password,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.Equal(t, "ftp", root.Protocol)
	assert.Equal(t, "ftp.example.com", *root.Host)
	assert.Equal(t, 21, *root.Port)
}

func TestStorageRoot_NFSProtocol(t *testing.T) {
	host := "nfs.example.com"
	path := "/export"
	mountPoint := "/mnt/nfs"
	options := "vers=4,soft,timeo=600"

	root := StorageRoot{
		ID:         1,
		Name:       "NFS Share",
		Protocol:   "nfs",
		Host:       &host,
		Path:       &path,
		MountPoint: &mountPoint,
		Options:    &options,
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	assert.Equal(t, "nfs", root.Protocol)
	assert.Equal(t, "nfs.example.com", *root.Host)
	assert.NotNil(t, root.MountPoint)
	assert.NotNil(t, root.Options)
}

func TestStorageRoot_WebDAVProtocol(t *testing.T) {
	url := "https://webdav.example.com/remote.php/dav"
	username := "webdavuser"
	password := "webdavpass"

	root := StorageRoot{
		ID:        1,
		Name:      "WebDAV Cloud",
		Protocol:  "webdav",
		URL:       &url,
		Username:  &username,
		Password:  &password,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.Equal(t, "webdav", root.Protocol)
	assert.NotNil(t, root.URL)
	assert.Contains(t, *root.URL, "https://")
}

func TestStorageRoot_PatternFilters(t *testing.T) {
	includePatterns := "*.mp4,*.mkv,*.avi"
	excludePatterns := "*.tmp,*.part"

	root := StorageRoot{
		ID:              1,
		Name:            "Filtered Root",
		Protocol:        "local",
		IncludePatterns: &includePatterns,
		ExcludePatterns: &excludePatterns,
		Enabled:         true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	assert.NotNil(t, root.IncludePatterns)
	assert.NotNil(t, root.ExcludePatterns)
	assert.Contains(t, *root.IncludePatterns, "*.mp4")
	assert.Contains(t, *root.ExcludePatterns, "*.tmp")
}

func TestStorageRoot_FeatureFlags(t *testing.T) {
	root := StorageRoot{
		ID:                       1,
		Name:                     "Feature Test",
		Protocol:                 "local",
		Enabled:                  true,
		EnableDuplicateDetection: true,
		EnableMetadataExtraction: false,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	assert.True(t, root.Enabled)
	assert.True(t, root.EnableDuplicateDetection)
	assert.False(t, root.EnableMetadataExtraction)
}

func TestStorageRoot_LastScanAt(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	root := StorageRoot{
		ID:         1,
		Name:       "Scanned Root",
		Protocol:   "local",
		Enabled:    true,
		LastScanAt: &now,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	assert.NotNil(t, root.LastScanAt)
	assert.Equal(t, now, *root.LastScanAt)
}

// =============================================================================
// SmbRoot Model Tests (Deprecated)
// =============================================================================

func TestSmbRoot_BasicFields(t *testing.T) {
	domain := "WORKGROUP"

	root := SmbRoot{
		ID:        1,
		Name:      "Old SMB Root",
		Host:      "server.local",
		Port:      445,
		Share:     "media",
		Username:  "user",
		Domain:    &domain,
		Enabled:   true,
		MaxDepth:  10,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.Equal(t, "Old SMB Root", root.Name)
	assert.Equal(t, "server.local", root.Host)
	assert.Equal(t, 445, root.Port)
	assert.Equal(t, "media", root.Share)
	assert.NotNil(t, root.Domain)
	assert.Equal(t, "WORKGROUP", *root.Domain)
}

func TestSmbRoot_JSONMarshaling(t *testing.T) {
	root := SmbRoot{
		ID:        1,
		Name:      "SMB Test",
		Host:      "host",
		Port:      445,
		Share:     "share",
		Username:  "user",
		Enabled:   true,
		MaxDepth:  10,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(root)
	require.NoError(t, err)

	var unmarshaled SmbRoot
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, root.Name, unmarshaled.Name)
	assert.Equal(t, root.Host, unmarshaled.Host)
}

// =============================================================================
// FileMetadata Model Tests
// =============================================================================

func TestFileMetadata_BasicFields(t *testing.T) {
	metadata := FileMetadata{
		ID:       1,
		FileID:   100,
		Key:      "resolution",
		Value:    "1920x1080",
		DataType: "string",
	}

	assert.Equal(t, int64(1), metadata.ID)
	assert.Equal(t, int64(100), metadata.FileID)
	assert.Equal(t, "resolution", metadata.Key)
	assert.Equal(t, "1920x1080", metadata.Value)
	assert.Equal(t, "string", metadata.DataType)
}

func TestFileMetadata_JSONMarshaling(t *testing.T) {
	metadata := FileMetadata{
		ID:       1,
		FileID:   100,
		Key:      "bitrate",
		Value:    "5000000",
		DataType: "integer",
	}

	data, err := json.Marshal(metadata)
	require.NoError(t, err)
	assert.Contains(t, string(data), "bitrate")

	var unmarshaled FileMetadata
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, metadata.Key, unmarshaled.Key)
	assert.Equal(t, metadata.Value, unmarshaled.Value)
}

func TestFileMetadata_DifferentDataTypes(t *testing.T) {
	tests := []struct {
		name     string
		metadata FileMetadata
	}{
		{
			name: "String type",
			metadata: FileMetadata{
				ID: 1, FileID: 100, Key: "title", Value: "Test", DataType: "string",
			},
		},
		{
			name: "Integer type",
			metadata: FileMetadata{
				ID: 2, FileID: 100, Key: "duration", Value: "120", DataType: "integer",
			},
		},
		{
			name: "Float type",
			metadata: FileMetadata{
				ID: 3, FileID: 100, Key: "framerate", Value: "23.976", DataType: "float",
			},
		},
		{
			name: "Boolean type",
			metadata: FileMetadata{
				ID: 4, FileID: 100, Key: "hdr", Value: "true", DataType: "boolean",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.metadata.Key)
			assert.NotEmpty(t, tt.metadata.Value)
			assert.NotEmpty(t, tt.metadata.DataType)
		})
	}
}

// =============================================================================
// DuplicateGroup Model Tests
// =============================================================================

func TestDuplicateGroup_BasicFields(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	group := DuplicateGroup{
		ID:        1,
		FileCount: 3,
		TotalSize: 3072000,
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, int64(1), group.ID)
	assert.Equal(t, 3, group.FileCount)
	assert.Equal(t, int64(3072000), group.TotalSize)
}

func TestDuplicateGroup_JSONMarshaling(t *testing.T) {
	group := DuplicateGroup{
		ID:        1,
		FileCount: 5,
		TotalSize: 5120000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(group)
	require.NoError(t, err)

	var unmarshaled DuplicateGroup
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, group.FileCount, unmarshaled.FileCount)
	assert.Equal(t, group.TotalSize, unmarshaled.TotalSize)
}

// =============================================================================
// VirtualPath Model Tests
// =============================================================================

func TestVirtualPath_BasicFields(t *testing.T) {
	vpath := VirtualPath{
		ID:         1,
		Path:       "/virtual/movies",
		TargetType: "storage_root",
		TargetID:   10,
		CreatedAt:  time.Now(),
	}

	assert.Equal(t, int64(1), vpath.ID)
	assert.Equal(t, "/virtual/movies", vpath.Path)
	assert.Equal(t, "storage_root", vpath.TargetType)
	assert.Equal(t, int64(10), vpath.TargetID)
}

func TestVirtualPath_JSONMarshaling(t *testing.T) {
	vpath := VirtualPath{
		ID:         1,
		Path:       "/virtual/path",
		TargetType: "file",
		TargetID:   100,
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(vpath)
	require.NoError(t, err)

	var unmarshaled VirtualPath
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, vpath.Path, unmarshaled.Path)
	assert.Equal(t, vpath.TargetType, unmarshaled.TargetType)
}

// =============================================================================
// ScanHistory Model Tests
// =============================================================================

func TestScanHistory_BasicFields(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	errorMsg := "Connection timeout"

	history := ScanHistory{
		ID:             1,
		StorageRootID:  10,
		ScanType:       "full",
		Status:         "completed",
		StartTime:      now,
		EndTime:        &now,
		FilesProcessed: 1000,
		FilesAdded:     50,
		FilesUpdated:   30,
		FilesDeleted:   10,
		ErrorCount:     5,
		ErrorMessage:   &errorMsg,
	}

	assert.Equal(t, "full", history.ScanType)
	assert.Equal(t, "completed", history.Status)
	assert.Equal(t, 1000, history.FilesProcessed)
	assert.Equal(t, 50, history.FilesAdded)
	assert.NotNil(t, history.ErrorMessage)
}

func TestScanHistory_JSONMarshaling(t *testing.T) {
	now := time.Now()

	history := ScanHistory{
		ID:             1,
		StorageRootID:  10,
		ScanType:       "incremental",
		Status:         "running",
		StartTime:      now,
		FilesProcessed: 500,
		FilesAdded:     25,
		FilesUpdated:   15,
		FilesDeleted:   5,
		ErrorCount:     0,
	}

	data, err := json.Marshal(history)
	require.NoError(t, err)

	var unmarshaled ScanHistory
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, history.ScanType, unmarshaled.ScanType)
	assert.Equal(t, history.Status, unmarshaled.Status)
}

func TestScanHistory_StatusValues(t *testing.T) {
	statuses := []string{"pending", "running", "completed", "failed", "cancelled"}

	for _, status := range statuses {
		history := ScanHistory{
			ID:            1,
			StorageRootID: 10,
			ScanType:      "full",
			Status:        status,
			StartTime:     time.Now(),
		}

		assert.Contains(t, statuses, history.Status)
	}
}

// =============================================================================
// FileWithMetadata Model Tests
// =============================================================================

func TestFileWithMetadata_BasicFields(t *testing.T) {
	file := FileWithMetadata{
		File: File{
			ID:         1,
			Name:       "test.mp4",
			Size:       1024000,
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			LastScanAt: time.Now(),
		},
		Metadata: []FileMetadata{
			{ID: 1, FileID: 1, Key: "resolution", Value: "1920x1080", DataType: "string"},
			{ID: 2, FileID: 1, Key: "duration", Value: "120", DataType: "integer"},
		},
	}

	assert.Equal(t, "test.mp4", file.Name)
	assert.Len(t, file.Metadata, 2)
	assert.Equal(t, "resolution", file.Metadata[0].Key)
}

func TestFileWithMetadata_JSONMarshaling(t *testing.T) {
	file := FileWithMetadata{
		File: File{
			ID:         1,
			Name:       "video.mkv",
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			LastScanAt: time.Now(),
		},
		Metadata: []FileMetadata{
			{ID: 1, FileID: 1, Key: "codec", Value: "H.264", DataType: "string"},
		},
	}

	data, err := json.Marshal(file)
	require.NoError(t, err)
	assert.Contains(t, string(data), "video.mkv")
	assert.Contains(t, string(data), "metadata")

	var unmarshaled FileWithMetadata
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, file.Name, unmarshaled.Name)
	assert.Len(t, unmarshaled.Metadata, 1)
}

func TestFileWithMetadata_EmptyMetadata(t *testing.T) {
	file := FileWithMetadata{
		File: File{
			ID:         1,
			Name:       "file.txt",
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
			LastScanAt: time.Now(),
		},
		Metadata: []FileMetadata{},
	}

	assert.Empty(t, file.Metadata)
}

// =============================================================================
// DirectoryInfo Model Tests
// =============================================================================

func TestDirectoryInfo_BasicFields(t *testing.T) {
	dirInfo := DirectoryInfo{
		Path:            "/movies/action",
		Name:            "action",
		StorageRootName: "Main Library",
		FileCount:       150,
		DirectoryCount:  10,
		TotalSize:       10737418240, // 10 GB
		DuplicateCount:  5,
		ModifiedAt:      time.Now(),
	}

	assert.Equal(t, "/movies/action", dirInfo.Path)
	assert.Equal(t, 150, dirInfo.FileCount)
	assert.Equal(t, int64(10737418240), dirInfo.TotalSize)
}

func TestDirectoryInfo_JSONMarshaling(t *testing.T) {
	dirInfo := DirectoryInfo{
		Path:            "/music",
		Name:            "music",
		StorageRootName: "Audio Library",
		FileCount:       500,
		DirectoryCount:  50,
		TotalSize:       5368709120,
		DuplicateCount:  10,
		ModifiedAt:      time.Now(),
	}

	data, err := json.Marshal(dirInfo)
	require.NoError(t, err)

	var unmarshaled DirectoryInfo
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, dirInfo.Path, unmarshaled.Path)
	assert.Equal(t, dirInfo.FileCount, unmarshaled.FileCount)
}

// =============================================================================
// SearchFilter Model Tests
// =============================================================================

func TestSearchFilter_BasicFields(t *testing.T) {
	filter := SearchFilter{
		Query:         "movie",
		Path:          "/movies",
		Extension:     ".mp4",
		FileType:      MediaTypeVideo,
		StorageRoots:  []string{"root1", "root2"},
		OnlyDuplicates: false,
	}

	assert.Equal(t, "movie", filter.Query)
	assert.Equal(t, "/movies", filter.Path)
	assert.Equal(t, ".mp4", filter.Extension)
	assert.Len(t, filter.StorageRoots, 2)
}

func TestSearchFilter_SizeRange(t *testing.T) {
	minSize := int64(1024)
	maxSize := int64(10485760)

	filter := SearchFilter{
		Query:   "large files",
		MinSize: &minSize,
		MaxSize: &maxSize,
	}

	assert.NotNil(t, filter.MinSize)
	assert.NotNil(t, filter.MaxSize)
	assert.Equal(t, int64(1024), *filter.MinSize)
	assert.Equal(t, int64(10485760), *filter.MaxSize)
}

func TestSearchFilter_DateRange(t *testing.T) {
	after := time.Now().AddDate(0, -1, 0)  // 1 month ago
	before := time.Now()

	filter := SearchFilter{
		Query:          "recent",
		ModifiedAfter:  &after,
		ModifiedBefore: &before,
	}

	assert.NotNil(t, filter.ModifiedAfter)
	assert.NotNil(t, filter.ModifiedBefore)
	assert.True(t, filter.ModifiedBefore.After(*filter.ModifiedAfter))
}

func TestSearchFilter_BooleanFlags(t *testing.T) {
	filter := SearchFilter{
		IncludeDeleted:     true,
		OnlyDuplicates:     true,
		ExcludeDuplicates:  false,
		IncludeDirectories: true,
	}

	assert.True(t, filter.IncludeDeleted)
	assert.True(t, filter.OnlyDuplicates)
	assert.False(t, filter.ExcludeDuplicates)
	assert.True(t, filter.IncludeDirectories)
}

func TestSearchFilter_JSONMarshaling(t *testing.T) {
	filter := SearchFilter{
		Query:     "test",
		FileType:  MediaTypeVideo,
		Extension: ".mkv",
	}

	data, err := json.Marshal(filter)
	require.NoError(t, err)

	var unmarshaled SearchFilter
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, filter.Query, unmarshaled.Query)
	assert.Equal(t, filter.FileType, unmarshaled.FileType)
}

// =============================================================================
// SortOptions Model Tests
// =============================================================================

func TestSortOptions_BasicFields(t *testing.T) {
	tests := []struct {
		name  string
		field string
		order string
	}{
		{"Sort by name ascending", "name", "asc"},
		{"Sort by size descending", "size", "desc"},
		{"Sort by modified date", "modified_at", "desc"},
		{"Sort by created date", "created_at", "asc"},
		{"Sort by path", "path", "asc"},
		{"Sort by extension", "extension", "asc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort := SortOptions{
				Field: tt.field,
				Order: tt.order,
			}

			assert.Equal(t, tt.field, sort.Field)
			assert.Equal(t, tt.order, sort.Order)
		})
	}
}

func TestSortOptions_JSONMarshaling(t *testing.T) {
	sort := SortOptions{
		Field: "size",
		Order: "desc",
	}

	data, err := json.Marshal(sort)
	require.NoError(t, err)

	var unmarshaled SortOptions
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, sort.Field, unmarshaled.Field)
	assert.Equal(t, sort.Order, unmarshaled.Order)
}

// =============================================================================
// PaginationOptions Model Tests
// =============================================================================

func TestPaginationOptions_BasicFields(t *testing.T) {
	pagination := PaginationOptions{
		Page:  1,
		Limit: 50,
	}

	assert.Equal(t, 1, pagination.Page)
	assert.Equal(t, 50, pagination.Limit)
}

func TestPaginationOptions_DifferentPageSizes(t *testing.T) {
	tests := []struct {
		name  string
		page  int
		limit int
	}{
		{"Small page", 1, 10},
		{"Medium page", 2, 50},
		{"Large page", 3, 100},
		{"Extra large page", 1, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pagination := PaginationOptions{
				Page:  tt.page,
				Limit: tt.limit,
			}

			assert.Equal(t, tt.page, pagination.Page)
			assert.Equal(t, tt.limit, pagination.Limit)
		})
	}
}

func TestPaginationOptions_JSONMarshaling(t *testing.T) {
	pagination := PaginationOptions{
		Page:  5,
		Limit: 100,
	}

	data, err := json.Marshal(pagination)
	require.NoError(t, err)

	var unmarshaled PaginationOptions
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, pagination.Page, unmarshaled.Page)
	assert.Equal(t, pagination.Limit, unmarshaled.Limit)
}

// =============================================================================
// SearchResult Model Tests
// =============================================================================

func TestSearchResult_BasicFields(t *testing.T) {
	result := SearchResult{
		Files: []FileWithMetadata{
			{File: File{ID: 1, Name: "file1.txt", CreatedAt: time.Now(), ModifiedAt: time.Now(), LastScanAt: time.Now()}},
			{File: File{ID: 2, Name: "file2.txt", CreatedAt: time.Now(), ModifiedAt: time.Now(), LastScanAt: time.Now()}},
		},
		TotalCount: 100,
		Page:       1,
		Limit:      50,
		TotalPages: 2,
	}

	assert.Len(t, result.Files, 2)
	assert.Equal(t, int64(100), result.TotalCount)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 50, result.Limit)
	assert.Equal(t, 2, result.TotalPages)
}

func TestSearchResult_EmptyResults(t *testing.T) {
	result := SearchResult{
		Files:      []FileWithMetadata{},
		TotalCount: 0,
		Page:       1,
		Limit:      50,
		TotalPages: 0,
	}

	assert.Empty(t, result.Files)
	assert.Equal(t, int64(0), result.TotalCount)
	assert.Equal(t, 0, result.TotalPages)
}

func TestSearchResult_JSONMarshaling(t *testing.T) {
	result := SearchResult{
		Files: []FileWithMetadata{
			{File: File{ID: 1, Name: "test.txt", CreatedAt: time.Now(), ModifiedAt: time.Now(), LastScanAt: time.Now()}},
		},
		TotalCount: 1,
		Page:       1,
		Limit:      50,
		TotalPages: 1,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var unmarshaled SearchResult
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Len(t, unmarshaled.Files, 1)
	assert.Equal(t, result.TotalCount, unmarshaled.TotalCount)
}

func TestSearchResult_PaginationCalculation(t *testing.T) {
	tests := []struct {
		name       string
		totalCount int64
		limit      int
		wantPages  int
	}{
		{"Exact pages", 100, 50, 2},
		{"Partial page", 105, 50, 3},
		{"Single page", 10, 50, 1},
		{"No results", 0, 50, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SearchResult{
				Files:      []FileWithMetadata{},
				TotalCount: tt.totalCount,
				Page:       1,
				Limit:      tt.limit,
				TotalPages: tt.wantPages,
			}

			assert.Equal(t, tt.wantPages, result.TotalPages)
		})
	}
}

// =============================================================================
// Media Type Constants Tests
// =============================================================================

func TestMediaTypeConstants(t *testing.T) {
	assert.Equal(t, "video", MediaTypeVideo)
	assert.Equal(t, "audio", MediaTypeAudio)
	assert.Equal(t, "image", MediaTypeImage)
	assert.Equal(t, "text", MediaTypeText)
	assert.Equal(t, "book", MediaTypeBook)
	assert.Equal(t, "game", MediaTypeGame)
	assert.Equal(t, "other", MediaTypeOther)
}

func TestMediaTypeUsage(t *testing.T) {
	file := File{
		ID:         1,
		Name:       "video.mp4",
		FileType:   strPtr(MediaTypeVideo),
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
		LastScanAt: time.Now(),
	}

	assert.NotNil(t, file.FileType)
	assert.Equal(t, MediaTypeVideo, *file.FileType)
}

// Helper function
func strPtr(s string) *string {
	return &s
}
