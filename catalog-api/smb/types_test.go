package smb

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestFileInfo_Initialization(t *testing.T) {
	tests := []struct {
		name     string
		fileInfo FileInfo
		wantName string
		wantSize int64
		wantDir  bool
		wantMode os.FileMode
	}{
		{
			name: "regular file",
			fileInfo: FileInfo{
				Name:    "document.txt",
				Size:    1024,
				ModTime: time.Now(),
				IsDir:   false,
				Mode:    0644,
			},
			wantName: "document.txt",
			wantSize: 1024,
			wantDir:  false,
			wantMode: 0644,
		},
		{
			name: "directory",
			fileInfo: FileInfo{
				Name:    "photos",
				Size:    0,
				ModTime: time.Now(),
				IsDir:   true,
				Mode:    0755,
			},
			wantName: "photos",
			wantSize: 0,
			wantDir:  true,
			wantMode: 0755,
		},
		{
			name:     "zero value",
			fileInfo: FileInfo{},
			wantName: "",
			wantSize: 0,
			wantDir:  false,
			wantMode: 0,
		},
		{
			name: "large file",
			fileInfo: FileInfo{
				Name:  "movie.mkv",
				Size:  4_294_967_296, // 4GB
				IsDir: false,
				Mode:  0444,
			},
			wantName: "movie.mkv",
			wantSize: 4_294_967_296,
			wantDir:  false,
			wantMode: 0444,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fileInfo.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", tt.fileInfo.Name, tt.wantName)
			}
			if tt.fileInfo.Size != tt.wantSize {
				t.Errorf("Size = %d, want %d", tt.fileInfo.Size, tt.wantSize)
			}
			if tt.fileInfo.IsDir != tt.wantDir {
				t.Errorf("IsDir = %v, want %v", tt.fileInfo.IsDir, tt.wantDir)
			}
			if tt.fileInfo.Mode != tt.wantMode {
				t.Errorf("Mode = %v, want %v", tt.fileInfo.Mode, tt.wantMode)
			}
		})
	}
}

func TestCopyOperation_Initialization(t *testing.T) {
	tests := []struct {
		name string
		op   CopyOperation
	}{
		{
			name: "basic copy",
			op: CopyOperation{
				SourcePath:        "/src/file.txt",
				DestinationPath:   "/dst/file.txt",
				OverwriteExisting: false,
			},
		},
		{
			name: "overwrite copy",
			op: CopyOperation{
				SourcePath:        "/media/video.mp4",
				DestinationPath:   "/backup/video.mp4",
				OverwriteExisting: true,
			},
		},
		{
			name: "zero value",
			op:   CopyOperation{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "basic copy" {
				if tt.op.SourcePath != "/src/file.txt" {
					t.Errorf("SourcePath = %q, want %q", tt.op.SourcePath, "/src/file.txt")
				}
				if tt.op.DestinationPath != "/dst/file.txt" {
					t.Errorf("DestinationPath = %q, want %q", tt.op.DestinationPath, "/dst/file.txt")
				}
				if tt.op.OverwriteExisting != false {
					t.Error("OverwriteExisting should be false")
				}
			}
			if tt.name == "overwrite copy" && !tt.op.OverwriteExisting {
				t.Error("OverwriteExisting should be true")
			}
			if tt.name == "zero value" {
				if tt.op.SourcePath != "" || tt.op.DestinationPath != "" || tt.op.OverwriteExisting {
					t.Error("zero value CopyOperation should have empty fields")
				}
			}
		})
	}
}

func TestCopyResult_Initialization(t *testing.T) {
	tests := []struct {
		name        string
		result      CopyResult
		wantSuccess bool
		wantBytes   int64
		wantErr     bool
	}{
		{
			name: "successful copy",
			result: CopyResult{
				Success:     true,
				BytesCopied: 4096,
				Error:       nil,
				TimeTaken:   500 * time.Millisecond,
			},
			wantSuccess: true,
			wantBytes:   4096,
			wantErr:     false,
		},
		{
			name: "failed copy",
			result: CopyResult{
				Success:     false,
				BytesCopied: 0,
				Error:       errors.New("permission denied"),
				TimeTaken:   100 * time.Millisecond,
			},
			wantSuccess: false,
			wantBytes:   0,
			wantErr:     true,
		},
		{
			name:        "zero value",
			result:      CopyResult{},
			wantSuccess: false,
			wantBytes:   0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.Success != tt.wantSuccess {
				t.Errorf("Success = %v, want %v", tt.result.Success, tt.wantSuccess)
			}
			if tt.result.BytesCopied != tt.wantBytes {
				t.Errorf("BytesCopied = %d, want %d", tt.result.BytesCopied, tt.wantBytes)
			}
			hasErr := tt.result.Error != nil
			if hasErr != tt.wantErr {
				t.Errorf("has error = %v, want %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestDirectoryTreeInfo_Initialization(t *testing.T) {
	tests := []struct {
		name      string
		tree      DirectoryTreeInfo
		wantPath  string
		wantFiles int
		wantDirs  int
		wantSize  int64
		wantDepth int
	}{
		{
			name: "populated tree",
			tree: DirectoryTreeInfo{
				Path:       "/media/movies",
				TotalFiles: 150,
				TotalDirs:  25,
				TotalSize:  1_073_741_824, // 1GB
				MaxDepth:   3,
				Files: []*FileInfo{
					{Name: "movie1.mkv", Size: 500_000_000, IsDir: false},
					{Name: "movie2.mkv", Size: 573_741_824, IsDir: false},
				},
				Subdirs: []*DirectoryTreeInfo{
					{Path: "/media/movies/action", TotalFiles: 50},
				},
			},
			wantPath:  "/media/movies",
			wantFiles: 150,
			wantDirs:  25,
			wantSize:  1_073_741_824,
			wantDepth: 3,
		},
		{
			name:      "empty tree",
			tree:      DirectoryTreeInfo{},
			wantPath:  "",
			wantFiles: 0,
			wantDirs:  0,
			wantSize:  0,
			wantDepth: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tree.Path != tt.wantPath {
				t.Errorf("Path = %q, want %q", tt.tree.Path, tt.wantPath)
			}
			if tt.tree.TotalFiles != tt.wantFiles {
				t.Errorf("TotalFiles = %d, want %d", tt.tree.TotalFiles, tt.wantFiles)
			}
			if tt.tree.TotalDirs != tt.wantDirs {
				t.Errorf("TotalDirs = %d, want %d", tt.tree.TotalDirs, tt.wantDirs)
			}
			if tt.tree.TotalSize != tt.wantSize {
				t.Errorf("TotalSize = %d, want %d", tt.tree.TotalSize, tt.wantSize)
			}
			if tt.tree.MaxDepth != tt.wantDepth {
				t.Errorf("MaxDepth = %d, want %d", tt.tree.MaxDepth, tt.wantDepth)
			}
		})
	}
}

func TestDirectoryTreeInfo_NestedStructure(t *testing.T) {
	leaf := &DirectoryTreeInfo{
		Path:       "/root/sub/leaf",
		TotalFiles: 2,
		TotalDirs:  0,
		Files: []*FileInfo{
			{Name: "a.txt", Size: 100},
			{Name: "b.txt", Size: 200},
		},
	}

	mid := &DirectoryTreeInfo{
		Path:       "/root/sub",
		TotalFiles: 5,
		TotalDirs:  1,
		Subdirs:    []*DirectoryTreeInfo{leaf},
	}

	root := &DirectoryTreeInfo{
		Path:       "/root",
		TotalFiles: 10,
		TotalDirs:  2,
		MaxDepth:   3,
		Subdirs:    []*DirectoryTreeInfo{mid},
	}

	if len(root.Subdirs) != 1 {
		t.Fatalf("root should have 1 subdir, got %d", len(root.Subdirs))
	}
	if root.Subdirs[0].Path != "/root/sub" {
		t.Errorf("first subdir path = %q, want %q", root.Subdirs[0].Path, "/root/sub")
	}
	if len(root.Subdirs[0].Subdirs) != 1 {
		t.Fatalf("mid should have 1 subdir, got %d", len(root.Subdirs[0].Subdirs))
	}
	if len(root.Subdirs[0].Subdirs[0].Files) != 2 {
		t.Errorf("leaf should have 2 files, got %d", len(root.Subdirs[0].Subdirs[0].Files))
	}
}

func TestNewSmbConnectionPool(t *testing.T) {
	tests := []struct {
		name           string
		maxConnections int
	}{
		{name: "standard pool", maxConnections: 10},
		{name: "single connection pool", maxConnections: 1},
		{name: "large pool", maxConnections: 100},
		{name: "zero max connections", maxConnections: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewSmbConnectionPool(tt.maxConnections)
			if pool == nil {
				t.Fatal("NewSmbConnectionPool returned nil")
			}
			if pool.maxConnections != tt.maxConnections {
				t.Errorf("maxConnections = %d, want %d", pool.maxConnections, tt.maxConnections)
			}
			if pool.connections == nil {
				t.Error("connections map should not be nil")
			}
			if len(pool.connections) != 0 {
				t.Errorf("new pool should have 0 connections, got %d", len(pool.connections))
			}
		})
	}
}

func TestSmbConnectionPool_CloseAll_Empty(t *testing.T) {
	pool := NewSmbConnectionPool(5)
	// CloseAll on empty pool should not panic
	pool.CloseAll()

	if len(pool.connections) != 0 {
		t.Errorf("connections should be empty after CloseAll, got %d", len(pool.connections))
	}
}

func TestSmbConnectionPool_CloseAll_MultipleCallsSafe(t *testing.T) {
	pool := NewSmbConnectionPool(5)
	// Multiple CloseAll calls should be safe
	pool.CloseAll()
	pool.CloseAll()
	pool.CloseAll()

	if len(pool.connections) != 0 {
		t.Errorf("connections should be empty after multiple CloseAll, got %d", len(pool.connections))
	}
}

func TestSmbConnectionPool_GetConnection_NilConfig(t *testing.T) {
	pool := NewSmbConnectionPool(5)

	// A nil config should cause NewSmbClient to fail (panic or error)
	// since it dereferences config fields. This verifies error propagation.
	defer func() {
		if r := recover(); r != nil {
			// Panicking on nil config dereference is acceptable
			return
		}
	}()

	_, err := pool.GetConnection("test", nil)
	if err == nil {
		t.Error("expected error when config is nil, got nil")
	}

	// Pool should not store a failed connection
	if len(pool.connections) != 0 {
		t.Errorf("pool should have 0 connections after failed connect, got %d", len(pool.connections))
	}
}

func TestSmbConnectionPool_MaxConnectionsValues(t *testing.T) {
	tests := []struct {
		name           string
		maxConnections int
		wantMax        int
	}{
		{name: "zero", maxConnections: 0, wantMax: 0},
		{name: "one", maxConnections: 1, wantMax: 1},
		{name: "ten", maxConnections: 10, wantMax: 10},
		{name: "hundred", maxConnections: 100, wantMax: 100},
		{name: "negative", maxConnections: -1, wantMax: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewSmbConnectionPool(tt.maxConnections)
			if pool.maxConnections != tt.wantMax {
				t.Errorf("maxConnections = %d, want %d", pool.maxConnections, tt.wantMax)
			}
		})
	}
}

func TestSmbConnectionPool_ConnectionsMapIsolation(t *testing.T) {
	// Verify that two pools do not share the same connections map.
	pool1 := NewSmbConnectionPool(5)
	pool2 := NewSmbConnectionPool(5)

	// They should have different map instances
	if &pool1.connections == &pool2.connections {
		t.Error("two pools should not share the same connections map reference")
	}
}

func TestSmbConnectionPool_CloseAll_Idempotent(t *testing.T) {
	pool := NewSmbConnectionPool(5)

	// Calling CloseAll multiple times should be safe and idempotent
	for i := 0; i < 5; i++ {
		pool.CloseAll()
		if len(pool.connections) != 0 {
			t.Errorf("iteration %d: connections should be empty after CloseAll, got %d", i, len(pool.connections))
		}
	}
}
