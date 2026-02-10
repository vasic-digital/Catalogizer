package performance

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"catalogizer/filesystem"
)

// ---------------------------------------------------------------------------
// Local Client Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkLocalClient_ListDirectory(b *testing.B) {
	tmpDir := setupTempDirectory(b, 100) // 100 files
	defer os.RemoveAll(tmpDir)

	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.ListDirectory(ctx, ".")
		if err != nil {
			b.Fatalf("list directory: %v", err)
		}
	}
}

func BenchmarkLocalClient_ListDirectory_Large(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000}
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("files=%d", sz), func(b *testing.B) {
			tmpDir := setupTempDirectory(b, sz)
			defer os.RemoveAll(tmpDir)

			client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
			ctx := context.Background()

			if err := client.Connect(ctx); err != nil {
				b.Fatalf("connect: %v", err)
			}
			defer client.Disconnect(ctx)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := client.ListDirectory(ctx, ".")
				if err != nil {
					b.Fatalf("list directory: %v", err)
				}
			}
		})
	}
}

func BenchmarkLocalClient_ReadFile(b *testing.B) {
	tmpDir := setupTempDirectory(b, 10)
	defer os.RemoveAll(tmpDir)

	// Create test file with known content
	testFilePath := filepath.Join(tmpDir, "test_read.txt")
	testContent := strings.Repeat("benchmark test content\n", 100)
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		b.Fatalf("create test file: %v", err)
	}

	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rc, err := client.ReadFile(ctx, "test_read.txt")
		if err != nil {
			b.Fatalf("read file: %v", err)
		}
		io.Copy(io.Discard, rc)
		rc.Close()
	}
}

func BenchmarkLocalClient_ReadFile_Sizes(b *testing.B) {
	sizes := []struct {
		name  string
		bytes int
	}{
		{"1KB", 1024},
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
		{"1MB", 1024 * 1024},
		{"10MB", 10 * 1024 * 1024},
	}

	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			tmpDir := setupTempDirectory(b, 1)
			defer os.RemoveAll(tmpDir)

			testFilePath := filepath.Join(tmpDir, "test_file.bin")
			testData := make([]byte, sz.bytes)
			if err := os.WriteFile(testFilePath, testData, 0644); err != nil {
				b.Fatalf("create test file: %v", err)
			}

			client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
			ctx := context.Background()

			if err := client.Connect(ctx); err != nil {
				b.Fatalf("connect: %v", err)
			}
			defer client.Disconnect(ctx)

			b.SetBytes(int64(sz.bytes))
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				rc, err := client.ReadFile(ctx, "test_file.bin")
				if err != nil {
					b.Fatalf("read file: %v", err)
				}
				io.Copy(io.Discard, rc)
				rc.Close()
			}
		})
	}
}

func BenchmarkLocalClient_WriteFile(b *testing.B) {
	tmpDir := setupTempDirectory(b, 0)
	defer os.RemoveAll(tmpDir)

	testContent := strings.Repeat("benchmark test content\n", 100)

	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		filename := fmt.Sprintf("test_write_%d.txt", i)
		err := client.WriteFile(ctx, filename, strings.NewReader(testContent))
		if err != nil {
			b.Fatalf("write file: %v", err)
		}
	}
}

func BenchmarkLocalClient_WriteFile_Sizes(b *testing.B) {
	sizes := []struct {
		name  string
		bytes int
	}{
		{"1KB", 1024},
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
		{"1MB", 1024 * 1024},
	}

	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			tmpDir := setupTempDirectory(b, 0)
			defer os.RemoveAll(tmpDir)

			testData := make([]byte, sz.bytes)

			client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
			ctx := context.Background()

			if err := client.Connect(ctx); err != nil {
				b.Fatalf("connect: %v", err)
			}
			defer client.Disconnect(ctx)

			b.SetBytes(int64(sz.bytes))
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				filename := fmt.Sprintf("test_file_%d.bin", i)
				err := client.WriteFile(ctx, filename, strings.NewReader(string(testData)))
				if err != nil {
					b.Fatalf("write file: %v", err)
				}
			}
		})
	}
}

func BenchmarkLocalClient_GetFileInfo(b *testing.B) {
	tmpDir := setupTempDirectory(b, 10)
	defer os.RemoveAll(tmpDir)

	testFilePath := filepath.Join(tmpDir, "test_info.txt")
	if err := os.WriteFile(testFilePath, []byte("test"), 0644); err != nil {
		b.Fatalf("create test file: %v", err)
	}

	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.GetFileInfo(ctx, "test_info.txt")
		if err != nil {
			b.Fatalf("get file info: %v", err)
		}
	}
}

func BenchmarkLocalClient_FileExists(b *testing.B) {
	tmpDir := setupTempDirectory(b, 10)
	defer os.RemoveAll(tmpDir)

	testFilePath := filepath.Join(tmpDir, "test_exists.txt")
	if err := os.WriteFile(testFilePath, []byte("test"), 0644); err != nil {
		b.Fatalf("create test file: %v", err)
	}

	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.FileExists(ctx, "test_exists.txt")
		if err != nil {
			b.Fatalf("file exists: %v", err)
		}
	}
}

func BenchmarkLocalClient_DeleteFile(b *testing.B) {
	tmpDir := setupTempDirectory(b, 0)
	defer os.RemoveAll(tmpDir)

	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		filename := fmt.Sprintf("test_delete_%d.txt", i)
		testFilePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(testFilePath, []byte("test"), 0644); err != nil {
			b.Fatalf("create test file: %v", err)
		}
		b.StartTimer()

		err := client.DeleteFile(ctx, filename)
		if err != nil {
			b.Fatalf("delete file: %v", err)
		}
	}
}

func BenchmarkLocalClient_CopyFile(b *testing.B) {
	tmpDir := setupTempDirectory(b, 0)
	defer os.RemoveAll(tmpDir)

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.txt")
	testContent := strings.Repeat("test content\n", 100)
	if err := os.WriteFile(srcFile, []byte(testContent), 0644); err != nil {
		b.Fatalf("create source file: %v", err)
	}

	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		dstFile := fmt.Sprintf("dest_%d.txt", i)
		err := client.CopyFile(ctx, "source.txt", dstFile)
		if err != nil {
			b.Fatalf("copy file: %v", err)
		}
	}
}

func BenchmarkLocalClient_CreateDirectory(b *testing.B) {
	tmpDir := setupTempDirectory(b, 0)
	defer os.RemoveAll(tmpDir)

	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		dirname := fmt.Sprintf("test_dir_%d", i)
		err := client.CreateDirectory(ctx, dirname)
		if err != nil {
			b.Fatalf("create directory: %v", err)
		}
	}
}

func BenchmarkLocalClient_DeleteDirectory(b *testing.B) {
	tmpDir := setupTempDirectory(b, 0)
	defer os.RemoveAll(tmpDir)

	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		dirname := fmt.Sprintf("test_dir_%d", i)
		dirPath := filepath.Join(tmpDir, dirname)
		if err := os.Mkdir(dirPath, 0755); err != nil {
			b.Fatalf("create test directory: %v", err)
		}
		b.StartTimer()

		err := client.DeleteDirectory(ctx, dirname)
		if err != nil {
			b.Fatalf("delete directory: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// Concurrent Operations Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkLocalClient_ConcurrentReads(b *testing.B) {
	tmpDir := setupTempDirectory(b, 10)
	defer os.RemoveAll(tmpDir)

	// Create test file
	testFilePath := filepath.Join(tmpDir, "concurrent_read.txt")
	testContent := strings.Repeat("concurrent test content\n", 1000)
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		b.Fatalf("create test file: %v", err)
	}

	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rc, err := client.ReadFile(ctx, "concurrent_read.txt")
			if err != nil {
				b.Fatalf("read file: %v", err)
			}
			io.Copy(io.Discard, rc)
			rc.Close()
		}
	})
}

func BenchmarkLocalClient_ConcurrentWrites(b *testing.B) {
	tmpDir := setupTempDirectory(b, 0)
	defer os.RemoveAll(tmpDir)

	testContent := strings.Repeat("concurrent write content\n", 100)

	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	counter := int64(0)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Use atomic counter to ensure unique filenames
			filename := fmt.Sprintf("concurrent_write_%d.txt", counter)
			counter++
			err := client.WriteFile(ctx, filename, strings.NewReader(testContent))
			if err != nil {
				b.Fatalf("write file: %v", err)
			}
		}
	})
}

func BenchmarkLocalClient_ConcurrentListDirectory(b *testing.B) {
	tmpDir := setupTempDirectory(b, 100)
	defer os.RemoveAll(tmpDir)

	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := client.ListDirectory(ctx, ".")
			if err != nil {
				b.Fatalf("list directory: %v", err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// Connection Management Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkLocalClient_ConnectDisconnect(b *testing.B) {
	tmpDir := setupTempDirectory(b, 0)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
		if err := client.Connect(ctx); err != nil {
			b.Fatalf("connect: %v", err)
		}
		if err := client.Disconnect(ctx); err != nil {
			b.Fatalf("disconnect: %v", err)
		}
	}
}

func BenchmarkLocalClient_TestConnection(b *testing.B) {
	tmpDir := setupTempDirectory(b, 0)
	defer os.RemoveAll(tmpDir)

	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := client.TestConnection(ctx); err != nil {
			b.Fatalf("test connection: %v", err)
		}
	}
}

func BenchmarkLocalClient_IsConnected(b *testing.B) {
	tmpDir := setupTempDirectory(b, 0)
	defer os.RemoveAll(tmpDir)

	client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		b.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = client.IsConnected()
	}
}

// ---------------------------------------------------------------------------
// Helper Functions
// ---------------------------------------------------------------------------

// setupTempDirectory creates a temporary directory with the specified number of test files
func setupTempDirectory(b *testing.B, fileCount int) string {
	b.Helper()

	tmpDir, err := os.MkdirTemp("", "benchmark_*")
	if err != nil {
		b.Fatalf("create temp dir: %v", err)
	}

	// Create test files
	for i := 0; i < fileCount; i++ {
		filename := fmt.Sprintf("file_%04d.txt", i)
		filepath := filepath.Join(tmpDir, filename)
		content := fmt.Sprintf("Test file content %d\n", i)
		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			os.RemoveAll(tmpDir)
			b.Fatalf("create test file: %v", err)
		}
	}

	return tmpDir
}

// ---------------------------------------------------------------------------
// End-to-End Workflow Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkLocalClient_FullWorkflow(b *testing.B) {
	tmpDir := setupTempDirectory(b, 0)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Connect
		client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
		if err := client.Connect(ctx); err != nil {
			b.Fatalf("connect: %v", err)
		}

		// Create directory
		dirname := fmt.Sprintf("workflow_dir_%d", i)
		if err := client.CreateDirectory(ctx, dirname); err != nil {
			b.Fatalf("create directory: %v", err)
		}

		// Write file
		filename := fmt.Sprintf("%s/test.txt", dirname)
		testContent := "workflow test content"
		if err := client.WriteFile(ctx, filename, strings.NewReader(testContent)); err != nil {
			b.Fatalf("write file: %v", err)
		}

		// Read file
		rc, err := client.ReadFile(ctx, filename)
		if err != nil {
			b.Fatalf("read file: %v", err)
		}
		io.Copy(io.Discard, rc)
		rc.Close()

		// Get file info
		if _, err := client.GetFileInfo(ctx, filename); err != nil {
			b.Fatalf("get file info: %v", err)
		}

		// List directory
		if _, err := client.ListDirectory(ctx, dirname); err != nil {
			b.Fatalf("list directory: %v", err)
		}

		// Copy file
		copyFilename := fmt.Sprintf("%s/test_copy.txt", dirname)
		if err := client.CopyFile(ctx, filename, copyFilename); err != nil {
			b.Fatalf("copy file: %v", err)
		}

		// Delete file
		if err := client.DeleteFile(ctx, copyFilename); err != nil {
			b.Fatalf("delete file: %v", err)
		}

		// Delete directory
		if err := client.DeleteDirectory(ctx, dirname); err != nil {
			b.Fatalf("delete directory: %v", err)
		}

		// Disconnect
		if err := client.Disconnect(ctx); err != nil {
			b.Fatalf("disconnect: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// Realistic Workload Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkLocalClient_ScanDirectory(b *testing.B) {
	// Simulate a realistic directory scan operation (list + get info for each file)
	sizes := []int{10, 50, 100, 500}
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("files=%d", sz), func(b *testing.B) {
			tmpDir := setupTempDirectory(b, sz)
			defer os.RemoveAll(tmpDir)

			client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
			ctx := context.Background()

			if err := client.Connect(ctx); err != nil {
				b.Fatalf("connect: %v", err)
			}
			defer client.Disconnect(ctx)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// List directory
				files, err := client.ListDirectory(ctx, ".")
				if err != nil {
					b.Fatalf("list directory: %v", err)
				}

				// Get info for each file (simulating metadata collection)
				for _, file := range files {
					if !file.IsDir {
						_, err := client.GetFileInfo(ctx, file.Name)
						if err != nil {
							b.Fatalf("get file info: %v", err)
						}
					}
				}
			}
		})
	}
}

func BenchmarkLocalClient_BatchOperations(b *testing.B) {
	// Simulate batch file operations (multiple files at once)
	batchSizes := []int{10, 50, 100}
	for _, bsz := range batchSizes {
		b.Run(fmt.Sprintf("batch=%d", bsz), func(b *testing.B) {
			tmpDir := setupTempDirectory(b, 0)
			defer os.RemoveAll(tmpDir)

			client := filesystem.NewLocalClient(&filesystem.LocalConfig{BasePath: tmpDir})
			ctx := context.Background()

			if err := client.Connect(ctx); err != nil {
				b.Fatalf("connect: %v", err)
			}
			defer client.Disconnect(ctx)

			testContent := "batch test content"

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Write batch of files
				for j := 0; j < bsz; j++ {
					filename := fmt.Sprintf("batch_%d_file_%d.txt", i, j)
					if err := client.WriteFile(ctx, filename, strings.NewReader(testContent)); err != nil {
						b.Fatalf("write file: %v", err)
					}
				}

				// Read batch of files
				for j := 0; j < bsz; j++ {
					filename := fmt.Sprintf("batch_%d_file_%d.txt", i, j)
					rc, err := client.ReadFile(ctx, filename)
					if err != nil {
						b.Fatalf("read file: %v", err)
					}
					io.Copy(io.Discard, rc)
					rc.Close()
				}

				// Delete batch of files
				for j := 0; j < bsz; j++ {
					filename := fmt.Sprintf("batch_%d_file_%d.txt", i, j)
					if err := client.DeleteFile(ctx, filename); err != nil {
						b.Fatalf("delete file: %v", err)
					}
				}
			}
		})
	}
}
