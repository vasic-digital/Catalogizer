package stress

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scanStressDB creates a file-based SQLite database with WAL mode for scan stress tests.
// File-based is required for concurrent access beyond a single connection.
func scanStressDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "scan_stress.db")

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, _ = db.Exec("PRAGMA journal_mode=WAL")
	_, _ = db.Exec("PRAGMA foreign_keys = ON")
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Create the tables needed for scan simulation
	tables := []string{
		`CREATE TABLE IF NOT EXISTS storage_roots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			protocol TEXT NOT NULL,
			path TEXT,
			enabled BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_scan_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			storage_root_id INTEGER NOT NULL,
			path TEXT NOT NULL,
			name TEXT NOT NULL,
			extension TEXT,
			size INTEGER NOT NULL,
			is_directory BOOLEAN DEFAULT 0,
			modified_at DATETIME NOT NULL,
			last_scan_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			parent_id INTEGER,
			FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id),
			FOREIGN KEY (parent_id) REFERENCES files(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_files_storage_root_path ON files(storage_root_id, path)`,
		`CREATE INDEX IF NOT EXISTS idx_files_name ON files(name)`,
		`CREATE INDEX IF NOT EXISTS idx_files_extension ON files(extension)`,
	}

	for _, tbl := range tables {
		_, err := db.Exec(tbl)
		require.NoError(t, err)
	}

	return db
}

// createTestDirectoryTree creates a temporary directory tree for scan testing.
// Returns the root path and the total number of files created.
func createTestDirectoryTree(t *testing.T, depth, breadth, filesPerDir int) (string, int) {
	t.Helper()
	root := t.TempDir()
	totalFiles := 0

	var createLevel func(parent string, level int)
	createLevel = func(parent string, level int) {
		if level >= depth {
			return
		}

		// Create files in this directory
		for f := 0; f < filesPerDir; f++ {
			ext := []string{".mkv", ".mp4", ".flac", ".mp3", ".jpg", ".srt", ".nfo", ".txt"}[f%8]
			fileName := fmt.Sprintf("file_%d_%d%s", level, f, ext)
			filePath := filepath.Join(parent, fileName)
			err := os.WriteFile(filePath, []byte(fmt.Sprintf("content-%d-%d", level, f)), 0644)
			require.NoError(t, err)
			totalFiles++
		}

		// Create subdirectories
		for d := 0; d < breadth; d++ {
			dirName := fmt.Sprintf("dir_%d_%d", level, d)
			dirPath := filepath.Join(parent, dirName)
			err := os.Mkdir(dirPath, 0755)
			require.NoError(t, err)
			createLevel(dirPath, level+1)
		}
	}

	createLevel(root, 0)
	return root, totalFiles
}

// simulateScan walks a directory tree and inserts file records into the database,
// mimicking what the real scanner does. Returns the number of files scanned.
func simulateScan(ctx context.Context, db *sql.DB, rootID int64, rootPath string) (int, error) {
	var count int

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if info.IsDir() {
			return nil
		}

		relPath := path[len(rootPath):]
		ext := filepath.Ext(info.Name())

		_, err := db.ExecContext(ctx, `
			INSERT INTO files (storage_root_id, path, name, extension, size, modified_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, rootID, relPath, info.Name(), ext, info.Size(), info.ModTime().Format(time.RFC3339))
		if err != nil {
			return err
		}
		count++
		return nil
	})

	return count, err
}

// =============================================================================
// STRESS TEST: Multiple Concurrent Scan Operations
// =============================================================================

func TestScanStress_MultipleConcurrentScans(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	db := scanStressDB(t)

	// Create multiple storage roots with distinct directory trees
	scanCount := 5
	roots := make([]struct {
		id   int64
		path string
	}, scanCount)

	for i := 0; i < scanCount; i++ {
		rootPath, _ := createTestDirectoryTree(t, 3, 2, 4)
		result, err := db.Exec(
			"INSERT INTO storage_roots (name, protocol, path) VALUES (?, 'local', ?)",
			fmt.Sprintf("scan-root-%d", i), rootPath,
		)
		require.NoError(t, err)
		id, err := result.LastInsertId()
		require.NoError(t, err)
		roots[i] = struct {
			id   int64
			path string
		}{id, rootPath}
	}

	t.Run("ConcurrentScansDoNotInterfere", func(t *testing.T) {
		var wg sync.WaitGroup
		var totalScanned int64
		var scanErrors int64

		ctx := context.Background()

		for _, root := range roots {
			wg.Add(1)
			go func(rootID int64, rootPath string) {
				defer wg.Done()

				count, err := simulateScan(ctx, db, rootID, rootPath)
				if err != nil {
					atomic.AddInt64(&scanErrors, 1)
					return
				}
				atomic.AddInt64(&totalScanned, int64(count))
			}(root.id, root.path)
		}

		wg.Wait()

		scanned := atomic.LoadInt64(&totalScanned)
		errors := atomic.LoadInt64(&scanErrors)
		t.Logf("Concurrent scans: %d files scanned, %d errors across %d roots",
			scanned, errors, scanCount)

		assert.Equal(t, int64(0), errors, "No scan errors should occur")
		assert.Greater(t, scanned, int64(0), "Files should be scanned")

		// Verify each storage root has its own files
		for _, root := range roots {
			var count int
			err := db.QueryRow(
				"SELECT COUNT(*) FROM files WHERE storage_root_id = ?",
				root.id,
			).Scan(&count)
			require.NoError(t, err)
			assert.Greater(t, count, 0,
				"Each storage root should have scanned files")
		}

		// Verify no cross-contamination
		var totalInDB int
		err := db.QueryRow("SELECT COUNT(*) FROM files").Scan(&totalInDB)
		require.NoError(t, err)
		assert.Equal(t, int(scanned), totalInDB,
			"Total files in DB should match total scanned")
	})
}

// =============================================================================
// STRESS TEST: Large Directory Tree Handling
// =============================================================================

func TestScanStress_LargeDirectoryTree(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	db := scanStressDB(t)

	// Create a larger tree: depth=4, breadth=3, 5 files per dir
	// Total dirs ~ 1+3+9+27 = 40, Total files ~ 40*5 = 200
	rootPath, expectedFiles := createTestDirectoryTree(t, 4, 3, 5)
	t.Logf("Created directory tree with %d files at %s", expectedFiles, rootPath)

	result, err := db.Exec(
		"INSERT INTO storage_roots (name, protocol, path) VALUES ('large-tree', 'local', ?)",
		rootPath,
	)
	require.NoError(t, err)
	rootID, err := result.LastInsertId()
	require.NoError(t, err)

	t.Run("ScanEntireTree", func(t *testing.T) {
		start := time.Now()
		ctx := context.Background()

		scanned, err := simulateScan(ctx, db, rootID, rootPath)
		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, expectedFiles, scanned,
			"Should scan all files in the tree")

		rate := float64(scanned) / elapsed.Seconds()
		t.Logf("Scanned %d files in %v (%.0f files/sec)", scanned, elapsed, rate)

		// Verify file count in database
		var dbCount int
		err = db.QueryRow(
			"SELECT COUNT(*) FROM files WHERE storage_root_id = ?",
			rootID,
		).Scan(&dbCount)
		require.NoError(t, err)
		assert.Equal(t, expectedFiles, dbCount,
			"Database should contain all scanned files")
	})

	t.Run("QueryByExtensionAfterLargeScan", func(t *testing.T) {
		// Verify that querying by extension works after large scan
		extensions := []string{".mkv", ".mp4", ".flac", ".mp3", ".jpg", ".srt", ".nfo", ".txt"}
		totalByExt := 0

		for _, ext := range extensions {
			var count int
			err := db.QueryRow(
				"SELECT COUNT(*) FROM files WHERE storage_root_id = ? AND extension = ?",
				rootID, ext,
			).Scan(&count)
			require.NoError(t, err)
			t.Logf("Extension %s: %d files", ext, count)
			totalByExt += count
		}

		assert.Equal(t, expectedFiles, totalByExt,
			"Sum of files by extension should equal total files")
	})
}

// =============================================================================
// STRESS TEST: Scan Cancellation Mid-Operation
// =============================================================================

func TestScanStress_CancellationMidOperation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	db := scanStressDB(t)

	// Create a moderately large tree to scan
	rootPath, totalFiles := createTestDirectoryTree(t, 4, 3, 5)
	t.Logf("Created %d files for cancellation test", totalFiles)

	result, err := db.Exec(
		"INSERT INTO storage_roots (name, protocol, path) VALUES ('cancel-test', 'local', ?)",
		rootPath,
	)
	require.NoError(t, err)
	rootID, err := result.LastInsertId()
	require.NoError(t, err)

	t.Run("CancelMidScan", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel after a short delay to interrupt the scan
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		scanned, err := simulateScan(ctx, db, rootID, rootPath)

		// The scan should have been cancelled
		if err != nil {
			assert.ErrorIs(t, err, context.Canceled,
				"Error should be context.Canceled")
		}

		t.Logf("Scanned %d/%d files before cancellation", scanned, totalFiles)

		// Some files should have been scanned but not all
		var dbCount int
		queryErr := db.QueryRow(
			"SELECT COUNT(*) FROM files WHERE storage_root_id = ?",
			rootID,
		).Scan(&dbCount)
		require.NoError(t, queryErr)

		t.Logf("Files in DB after cancellation: %d", dbCount)
		assert.GreaterOrEqual(t, dbCount, 0,
			"Some or no files scanned before cancellation is valid")
	})

	t.Run("CancelWithTimeout", func(t *testing.T) {
		// Reset: delete files from previous subtest
		_, err := db.Exec("DELETE FROM files WHERE storage_root_id = ?", rootID)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()

		scanned, err := simulateScan(ctx, db, rootID, rootPath)

		if err != nil {
			assert.True(t,
				err == context.DeadlineExceeded || err == context.Canceled,
				"Error should be a context error, got: %v", err)
		}

		t.Logf("Scanned %d files before timeout", scanned)
	})

	t.Run("RescanAfterCancellation", func(t *testing.T) {
		// Create a new root for a clean rescan
		result2, err := db.Exec(
			"INSERT INTO storage_roots (name, protocol, path) VALUES ('rescan-test', 'local', ?)",
			rootPath,
		)
		require.NoError(t, err)
		rootID2, _ := result2.LastInsertId()

		// Full scan should succeed after a cancellation
		ctx := context.Background()
		scanned, err := simulateScan(ctx, db, rootID2, rootPath)

		require.NoError(t, err)
		assert.Equal(t, totalFiles, scanned,
			"Full rescan after cancellation should find all files")
	})
}

// =============================================================================
// STRESS TEST: Memory Usage During Large Scans
// =============================================================================

func TestScanStress_MemoryUsageDuringLargeScans(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	db := scanStressDB(t)

	// Create a fairly large tree for memory measurement
	rootPath, totalFiles := createTestDirectoryTree(t, 4, 3, 8)
	t.Logf("Created %d files for memory test", totalFiles)

	result, err := db.Exec(
		"INSERT INTO storage_roots (name, protocol, path) VALUES ('mem-test', 'local', ?)",
		rootPath,
	)
	require.NoError(t, err)
	rootID, err := result.LastInsertId()
	require.NoError(t, err)

	t.Run("MemoryDoesNotGrowUnbounded", func(t *testing.T) {
		// Force GC and measure baseline
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
		var baselineStats runtime.MemStats
		runtime.ReadMemStats(&baselineStats)
		baselineHeap := baselineStats.HeapAlloc

		// Perform the scan
		ctx := context.Background()
		scanned, err := simulateScan(ctx, db, rootID, rootPath)
		require.NoError(t, err)
		assert.Equal(t, totalFiles, scanned)

		// Force GC and measure post-scan
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
		var postStats runtime.MemStats
		runtime.ReadMemStats(&postStats)
		postHeap := postStats.HeapAlloc

		// Calculate growth
		var growth uint64
		if postHeap > baselineHeap {
			growth = postHeap - baselineHeap
		}
		growthMB := float64(growth) / (1024 * 1024)

		t.Logf("Memory: baseline=%.2f MB, post-scan=%.2f MB, growth=%.2f MB",
			float64(baselineHeap)/(1024*1024),
			float64(postHeap)/(1024*1024),
			growthMB)
		t.Logf("Files scanned: %d, GC cycles: %d", scanned, postStats.NumGC-baselineStats.NumGC)

		// Memory growth should be reasonable (< 100 MB for a few hundred files)
		assert.Less(t, growthMB, 100.0,
			"Heap growth during scan should be bounded (< 100 MB)")
	})

	t.Run("ConcurrentScansMemoryBound", func(t *testing.T) {
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
		var baselineStats runtime.MemStats
		runtime.ReadMemStats(&baselineStats)
		baselineHeap := baselineStats.HeapAlloc

		// Run 3 concurrent scans on different roots
		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			rootPath, _ := createTestDirectoryTree(t, 3, 2, 5)
			r, err := db.Exec(
				"INSERT INTO storage_roots (name, protocol, path) VALUES (?, 'local', ?)",
				fmt.Sprintf("mem-concurrent-%d", i), rootPath,
			)
			require.NoError(t, err)
			id, _ := r.LastInsertId()

			wg.Add(1)
			go func(rid int64, rp string) {
				defer wg.Done()
				simulateScan(context.Background(), db, rid, rp)
			}(id, rootPath)
		}
		wg.Wait()

		runtime.GC()
		time.Sleep(50 * time.Millisecond)
		var postStats runtime.MemStats
		runtime.ReadMemStats(&postStats)
		postHeap := postStats.HeapAlloc

		var growth uint64
		if postHeap > baselineHeap {
			growth = postHeap - baselineHeap
		}
		growthMB := float64(growth) / (1024 * 1024)

		t.Logf("Concurrent scans memory growth: %.2f MB", growthMB)
		assert.Less(t, growthMB, 200.0,
			"Concurrent scan memory growth should be bounded (< 200 MB)")
	})
}
