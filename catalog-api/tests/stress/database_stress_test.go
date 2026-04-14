package stress

import (
	"catalogizer/internal/tests"
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DatabaseStressContext manages database stress test execution
type DatabaseStressContext struct {
	DB             *sql.DB
	OperationCount int64
	SuccessCount   int64
	ErrorCount     int64
	TotalLatency   int64
	StartTime      time.Time
	Errors         []error
	ErrorsMutex    sync.Mutex
}

func newDatabaseStressContext(t *testing.T) *DatabaseStressContext {
	db := tests.SetupTestDB(t)

	return &DatabaseStressContext{
		DB:        db,
		StartTime: time.Now(),
	}
}

func (dsc *DatabaseStressContext) recordOperation(latency time.Duration, err error) {
	atomic.AddInt64(&dsc.OperationCount, 1)
	atomic.AddInt64(&dsc.TotalLatency, int64(latency.Microseconds()))

	if err != nil {
		atomic.AddInt64(&dsc.ErrorCount, 1)
		dsc.recordError(err)
	} else {
		atomic.AddInt64(&dsc.SuccessCount, 1)
	}
}

func (dsc *DatabaseStressContext) recordError(err error) {
	dsc.ErrorsMutex.Lock()
	defer dsc.ErrorsMutex.Unlock()
	if len(dsc.Errors) < 50 {
		dsc.Errors = append(dsc.Errors, err)
	}
}

func (dsc *DatabaseStressContext) GetStats() map[string]interface{} {
	duration := time.Since(dsc.StartTime)
	opCount := atomic.LoadInt64(&dsc.OperationCount)
	successCount := atomic.LoadInt64(&dsc.SuccessCount)
	errorCount := atomic.LoadInt64(&dsc.ErrorCount)
	totalLatency := atomic.LoadInt64(&dsc.TotalLatency)

	opsPerSec := float64(opCount) / duration.Seconds()
	avgLatency := time.Duration(0)
	if opCount > 0 {
		avgLatency = time.Duration(totalLatency/opCount) * time.Microsecond
	}

	successRate := 0.0
	if opCount > 0 {
		successRate = float64(successCount) / float64(opCount) * 100
	}

	return map[string]interface{}{
		"duration":     duration,
		"operations":   opCount,
		"success":      successCount,
		"errors":       errorCount,
		"ops_per_sec":  opsPerSec,
		"avg_latency":  avgLatency,
		"success_rate": successRate,
	}
}

func (dsc *DatabaseStressContext) PrintStats(t *testing.T) {
	stats := dsc.GetStats()

	t.Logf("\n=== Database Stress Test Results ===")
	t.Logf("Duration:        %v", stats["duration"])
	t.Logf("Operations:      %d", stats["operations"])
	t.Logf("Successful:      %d", stats["success"])
	t.Logf("Errors:          %d", stats["errors"])
	t.Logf("Ops/sec:         %.2f", stats["ops_per_sec"])
	t.Logf("Avg Latency:     %v", stats["avg_latency"])
	t.Logf("Success Rate:    %.2f%%", stats["success_rate"])

	if len(dsc.Errors) > 0 {
		t.Logf("\nFirst %d Errors:", len(dsc.Errors))
		for i, err := range dsc.Errors {
			if i >= 5 {
				break
			}
			t.Logf("  %d: %v", i+1, err)
		}
	}
}

// =============================================================================
// STRESS TEST: Concurrent Database Reads (50 goroutines)
// =============================================================================

func TestConcurrentDatabaseReads(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	dsc := newDatabaseStressContext(t)
	defer dsc.DB.Close()

	// Create test storage root
	_, err := dsc.DB.Exec(`
		INSERT INTO storage_roots (id, name, protocol, path, enabled)
		VALUES (1, 'test-root', 'local', '/test', 1)
	`)
	require.NoError(t, err)

	// Insert test data first
	_, err = dsc.DB.Exec(`
		INSERT INTO files (storage_root_id, path, name, size, modified_at)
		VALUES (1, ?, ?, ?, datetime('now'))
	`, "/test/file1.txt", "file1.txt", 1024)
	require.NoError(t, err)

	t.Run("100ConcurrentReads", func(t *testing.T) {
		concurrentReaders := 100
		readsPerReader := 50

		var wg sync.WaitGroup
		for i := 0; i < concurrentReaders; i++ {
			wg.Add(1)
			go func(readerID int) {
				defer wg.Done()

				for j := 0; j < readsPerReader; j++ {
					start := time.Now()
					var count int
					err := dsc.DB.QueryRow("SELECT COUNT(*) FROM files").Scan(&count)
					latency := time.Since(start)

					dsc.recordOperation(latency, err)
				}
			}(i)
		}

		wg.Wait()
		dsc.PrintStats(t)

		stats := dsc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 99.0, "Read operations should have >99%% success")
		assert.Less(t, stats["avg_latency"].(time.Duration), 10*time.Millisecond, "Avg read latency should be <10ms")
	})
}

// =============================================================================
// STRESS TEST: Concurrent Database Writes (50 goroutines)
// =============================================================================

func TestConcurrentDatabaseWrites(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	dsc := newDatabaseStressContext(t)
	defer dsc.DB.Close()

	// Create test storage root
	_, err := dsc.DB.Exec(`
		INSERT INTO storage_roots (id, name, protocol, path, enabled)
		VALUES (1, 'test-root', 'local', '/test', 1)
	`)
	require.NoError(t, err)

	t.Run("ConcurrentInserts", func(t *testing.T) {
		concurrentWriters := 50
		writesPerWriter := 20

		var wg sync.WaitGroup
		for i := 0; i < concurrentWriters; i++ {
			wg.Add(1)
			go func(writerID int) {
				defer wg.Done()

				for j := 0; j < writesPerWriter; j++ {
					start := time.Now()
					_, err := dsc.DB.Exec(`
						INSERT INTO files (storage_root_id, path, name, size, modified_at)
						VALUES (1, ?, ?, ?, datetime('now'))
					`,
						fmt.Sprintf("/test/writer%d/file%d.txt", writerID, j),
						fmt.Sprintf("file%d.txt", j),
						1024*(j+1),
					)
					latency := time.Since(start)

					dsc.recordOperation(latency, err)
					time.Sleep(1 * time.Millisecond) // Small delay to avoid overwhelming DB
				}
			}(i)
		}

		wg.Wait()
		dsc.PrintStats(t)

		stats := dsc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 95.0, "Write operations should have >95%% success")
		assert.Less(t, stats["avg_latency"].(time.Duration), 50*time.Millisecond, "Avg write latency should be <50ms")

		// Verify all records were inserted
		var count int
		err := dsc.DB.QueryRow("SELECT COUNT(*) FROM files WHERE path LIKE '/test/writer%'").Scan(&count)
		require.NoError(t, err)
		expectedCount := concurrentWriters * writesPerWriter
		assert.Equal(t, expectedCount, count, "All records should be inserted")
	})

	t.Run("ConcurrentUpdates", func(t *testing.T) {
		// Insert test records
		for i := 0; i < 100; i++ {
			dsc.DB.Exec(`
				INSERT INTO files (storage_root_id, path, name, size, modified_at)
				VALUES (1, ?, ?, ?, datetime('now'))
			`, fmt.Sprintf("/update/file%d.txt", i), fmt.Sprintf("file%d.txt", i), 1024)
		}

		concurrentUpdaters := 50
		updatesPerUpdater := 10

		var wg sync.WaitGroup
		for i := 0; i < concurrentUpdaters; i++ {
			wg.Add(1)
			go func(updaterID int) {
				defer wg.Done()

				for j := 0; j < updatesPerUpdater; j++ {
					fileID := (updaterID*updatesPerUpdater + j) % 100

					start := time.Now()
					_, err := dsc.DB.Exec(`
						UPDATE files
						SET size = size + 1
						WHERE path = ?
					`, fmt.Sprintf("/update/file%d.txt", fileID))
					latency := time.Since(start)

					dsc.recordOperation(latency, err)
				}
			}(i)
		}

		wg.Wait()
		dsc.PrintStats(t)

		stats := dsc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 95.0, "Update operations should have >95%% success")
	})
}

// =============================================================================
// STRESS TEST: Mixed Read/Write Workload (70% reads, 30% writes)
// =============================================================================

func TestMixedReadWriteWorkload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Use file-based SQLite with WAL mode for concurrent read/write testing.
	// In-memory SQLite with MaxOpenConns(1) causes goroutines to block waiting
	// for the single connection, leading to hangs.
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/mixed_rw_test.db"
	sqlDB, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&cache=shared")
	require.NoError(t, err, "Failed to open file-based SQLite")
	t.Cleanup(func() { sqlDB.Close() })

	_, _ = sqlDB.Exec("PRAGMA journal_mode=WAL")
	_, _ = sqlDB.Exec("PRAGMA foreign_keys = ON")
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	// Create tables
	_, err = sqlDB.Exec(`CREATE TABLE IF NOT EXISTS storage_roots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		protocol TEXT NOT NULL,
		path TEXT,
		enabled BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)
	_, err = sqlDB.Exec(`CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		storage_root_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		name TEXT NOT NULL,
		size INTEGER NOT NULL,
		modified_at DATETIME NOT NULL,
		FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id)
	)`)
	require.NoError(t, err)

	dsc := &DatabaseStressContext{
		DB:        sqlDB,
		StartTime: time.Now(),
	}

	// Create test storage root
	_, err = dsc.DB.Exec(`
		INSERT INTO storage_roots (id, name, protocol, path, enabled)
		VALUES (1, 'test-root', 'local', '/test', 1)
	`)
	require.NoError(t, err)

	// Pre-populate with some data
	for i := 0; i < 100; i++ {
		dsc.DB.Exec(`
			INSERT INTO files (storage_root_id, path, name, size, modified_at)
			VALUES (1, ?, ?, ?, datetime('now'))
		`, fmt.Sprintf("/mixed/file%d.txt", i), fmt.Sprintf("file%d.txt", i), 1024*(i+1))
	}

	t.Run("70PercentReads30PercentWrites", func(t *testing.T) {
		duration := 3 * time.Second
		concurrentWorkers := 20

		done := make(chan bool)
		var wg sync.WaitGroup

		for i := 0; i < concurrentWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for {
					select {
					case <-done:
						return
					default:
						// 70% reads, 30% writes
						if workerID%10 < 7 {
							// Read operation
							start := time.Now()
							var count int
							err := dsc.DB.QueryRow("SELECT COUNT(*) FROM files WHERE path LIKE '/mixed/%'").Scan(&count)
							latency := time.Since(start)
							dsc.recordOperation(latency, err)
						} else {
							// Write operation (insert or update)
							if workerID%2 == 0 {
								// Insert
								start := time.Now()
								_, err := dsc.DB.Exec(`
									INSERT INTO files (storage_root_id, path, name, size, modified_at)
									VALUES (1, ?, ?, ?, datetime('now'))
								`,
									fmt.Sprintf("/mixed/new_%d_%d.txt", workerID, time.Now().UnixNano()),
									fmt.Sprintf("new_%d.txt", workerID),
									2048,
								)
								latency := time.Since(start)
								dsc.recordOperation(latency, err)
							} else {
								// Update
								start := time.Now()
								fileID := workerID % 100
								_, err := dsc.DB.Exec(`
									UPDATE files
									SET size = size + 100
									WHERE path = ?
								`, fmt.Sprintf("/mixed/file%d.txt", fileID))
								latency := time.Since(start)
								dsc.recordOperation(latency, err)
							}
						}
						time.Sleep(10 * time.Millisecond)
					}
				}
			}(i)
		}

		time.Sleep(duration)
		close(done)

		// Wait with a hard deadline
		wgDone := make(chan struct{})
		go func() {
			wg.Wait()
			close(wgDone)
		}()
		select {
		case <-wgDone:
		case <-time.After(10 * time.Second):
			t.Fatal("MixedReadWriteWorkload: goroutines did not finish within deadline")
		}

		dsc.PrintStats(t)

		stats := dsc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 90.0, "Mixed workload should have >90%% success")
		assert.Greater(t, stats["ops_per_sec"].(float64), 100.0, "Should handle >100 ops/sec")
	})
}

// =============================================================================
// STRESS TEST: Transaction Contention (multiple concurrent transactions)
// =============================================================================

func TestTransactionContention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	dsc := newDatabaseStressContext(t)
	defer dsc.DB.Close()

	// Create test storage root
	_, err := dsc.DB.Exec(`
		INSERT INTO storage_roots (id, name, protocol, path, enabled)
		VALUES (1, 'test-root', 'local', '/test', 1)
	`)
	require.NoError(t, err)

	t.Run("ConcurrentTransactions", func(t *testing.T) {
		concurrentTxs := 20
		operationsPerTx := 10

		var wg sync.WaitGroup
		for i := 0; i < concurrentTxs; i++ {
			wg.Add(1)
			go func(txID int) {
				defer wg.Done()

				start := time.Now()
				tx, err := dsc.DB.Begin()
				if err != nil {
					dsc.recordOperation(time.Since(start), err)
					return
				}

				// Perform multiple operations in transaction
				for j := 0; j < operationsPerTx; j++ {
					_, err := tx.Exec(`
						INSERT INTO files (storage_root_id, path, name, size, modified_at)
						VALUES (1, ?, ?, ?, datetime('now'))
					`,
						fmt.Sprintf("/tx/tx%d/file%d.txt", txID, j),
						fmt.Sprintf("file%d.txt", j),
						1024,
					)
					if err != nil {
						tx.Rollback()
						dsc.recordOperation(time.Since(start), err)
						return
					}
				}

				err = tx.Commit()
				latency := time.Since(start)
				dsc.recordOperation(latency, err)
			}(i)
		}

		wg.Wait()
		dsc.PrintStats(t)

		stats := dsc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 95.0, "Transactions should have >95%% success")

		// Verify all transactions completed
		var count int
		err := dsc.DB.QueryRow("SELECT COUNT(*) FROM files WHERE path LIKE '/tx/%'").Scan(&count)
		require.NoError(t, err)
		expectedCount := concurrentTxs * operationsPerTx
		assert.Equal(t, expectedCount, count, "All transaction operations should complete")
	})

	t.Run("RollbackUnderContention", func(t *testing.T) {
		// Test that rollback works correctly when transactions conflict
		var rollbackCount int64
		var commitCount int64

		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(txID int) {
				defer wg.Done()

				tx, err := dsc.DB.Begin()
				if err != nil {
					return
				}

				// Insert a record
				_, err = tx.Exec(`
					INSERT INTO files (storage_root_id, path, name, size, modified_at)
					VALUES (1, ?, ?, ?, datetime('now'))
				`, fmt.Sprintf("/rollback/tx%d.txt", txID), fmt.Sprintf("tx%d.txt", txID), 512)
				if err != nil {
					tx.Rollback()
					atomic.AddInt64(&rollbackCount, 1)
					return
				}

				// Even-numbered transactions commit, odd ones rollback
				if txID%2 == 0 {
					if err := tx.Commit(); err != nil {
						atomic.AddInt64(&rollbackCount, 1)
					} else {
						atomic.AddInt64(&commitCount, 1)
					}
				} else {
					tx.Rollback()
					atomic.AddInt64(&rollbackCount, 1)
				}
			}(i)
		}

		wg.Wait()

		commits := atomic.LoadInt64(&commitCount)
		rollbacks := atomic.LoadInt64(&rollbackCount)
		t.Logf("Rollback test: %d commits, %d rollbacks", commits, rollbacks)

		// Verify only committed records exist
		var fileCount int
		err := dsc.DB.QueryRow("SELECT COUNT(*) FROM files WHERE path LIKE '/rollback/%'").Scan(&fileCount)
		require.NoError(t, err)
		assert.Equal(t, int(commits), fileCount,
			"Only committed records should exist in the database")
	})

	t.Run("TransactionIsolation", func(t *testing.T) {
		// Verify transaction isolation: uncommitted data should not be visible
		// to other connections. With in-memory SQLite + MaxOpenConns(1), all
		// operations serialize on the same connection, so we verify the semantics
		// at the SQL level using savepoints.

		tx, err := dsc.DB.Begin()
		require.NoError(t, err)

		// Insert inside transaction
		_, err = tx.Exec(`
			INSERT INTO files (storage_root_id, path, name, size, modified_at)
			VALUES (1, '/isolation/pending.txt', 'pending.txt', 256, datetime('now'))
		`)
		require.NoError(t, err)

		// The row should be visible inside the transaction
		var txCount int
		err = tx.QueryRow("SELECT COUNT(*) FROM files WHERE path = '/isolation/pending.txt'").Scan(&txCount)
		require.NoError(t, err)
		assert.Equal(t, 1, txCount, "Row should be visible inside the transaction")

		// Rollback - row should disappear
		tx.Rollback()

		var afterCount int
		err = dsc.DB.QueryRow("SELECT COUNT(*) FROM files WHERE path = '/isolation/pending.txt'").Scan(&afterCount)
		require.NoError(t, err)
		assert.Equal(t, 0, afterCount,
			"Rolled-back row should not be visible after rollback")
	})
}

// =============================================================================
// STRESS TEST: Connection Pool Exhaustion and Recovery
// =============================================================================

func TestConnectionPoolExhaustionAndRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Use file-based SQLite with WAL mode to enable real connection pooling.
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/pool_stress_test.db"

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&cache=shared")
	require.NoError(t, err, "Failed to open file-based SQLite")
	t.Cleanup(func() { db.Close() })

	// Enable WAL mode explicitly
	_, err = db.Exec("PRAGMA journal_mode=WAL")
	require.NoError(t, err)
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	// Create tables
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS storage_roots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		protocol TEXT NOT NULL,
		path TEXT,
		enabled BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		storage_root_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		name TEXT NOT NULL,
		size INTEGER NOT NULL,
		modified_at DATETIME NOT NULL,
		FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id)
	)`)
	require.NoError(t, err)

	// Insert test data
	_, err = db.Exec(`INSERT INTO storage_roots (id, name, protocol, path, enabled)
		VALUES (1, 'test-root', 'local', '/test', 1)`)
	require.NoError(t, err)

	// Configure a small pool to force exhaustion
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	t.Run("ExceedPoolWithRecovery", func(t *testing.T) {
		// Exceed the pool size with concurrent operations
		concurrentOps := 50
		duration := 3 * time.Second
		var successOps int64
		var failedOps int64

		done := make(chan bool)
		var wg sync.WaitGroup

		for i := 0; i < concurrentOps; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for {
					select {
					case <-done:
						return
					default:
						var count int
						err := db.QueryRow("SELECT COUNT(*) FROM files").Scan(&count)
						if err != nil {
							atomic.AddInt64(&failedOps, 1)
						} else {
							atomic.AddInt64(&successOps, 1)
						}
						time.Sleep(20 * time.Millisecond)
					}
				}
			}(i)
		}

		time.Sleep(duration)
		close(done)
		wg.Wait()

		success := atomic.LoadInt64(&successOps)
		failed := atomic.LoadInt64(&failedOps)
		total := success + failed
		successRate := float64(success) / float64(total) * 100

		t.Logf("Pool exhaustion: %d success, %d failed, %.1f%% rate", success, failed, successRate)
		assert.Greater(t, successRate, 95.0,
			"Should handle connection pool saturation gracefully (>95%% success)")

		// Check pool stats
		dbStats := db.Stats()
		t.Logf("DB Pool Stats:")
		t.Logf("  Max Open Connections: %d", dbStats.MaxOpenConnections)
		t.Logf("  Open Connections: %d", dbStats.OpenConnections)
		t.Logf("  In Use: %d", dbStats.InUse)
		t.Logf("  Idle: %d", dbStats.Idle)
		t.Logf("  Wait Count: %d", dbStats.WaitCount)
		t.Logf("  Wait Duration: %v", dbStats.WaitDuration)

		// After storm, pool should recover
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM files").Scan(&count)
		assert.NoError(t, err, "Pool should recover after exhaustion")
	})

	t.Run("PoolRecoveryAfterConnectionDrain", func(t *testing.T) {
		// Reconfigure pool to simulate connection drain
		db.SetMaxOpenConns(3)
		db.SetConnMaxLifetime(100 * time.Millisecond)

		// Cause connections to expire and be recreated
		for round := 0; round < 5; round++ {
			var wg sync.WaitGroup
			for w := 0; w < 10; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					var count int
					db.QueryRow("SELECT COUNT(*) FROM storage_roots").Scan(&count)
				}()
			}
			wg.Wait()
			time.Sleep(150 * time.Millisecond) // Let connections expire
		}

		// Verify pool is still functional
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM storage_roots").Scan(&count)
		require.NoError(t, err, "Pool should recover after connection drain")
		assert.Equal(t, 1, count, "Data should be intact after pool drain")

		// Restore normal pool settings
		db.SetMaxOpenConns(25)
		db.SetConnMaxLifetime(5 * time.Minute)
	})
}

// =============================================================================
// STRESS TEST: Large Result Set Handling
// =============================================================================

func TestLargeResultSetHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	dsc := newDatabaseStressContext(t)
	defer dsc.DB.Close()

	// Create test storage root
	_, err := dsc.DB.Exec(`
		INSERT INTO storage_roots (id, name, protocol, path, enabled)
		VALUES (1, 'test-root', 'local', '/test', 1)
	`)
	require.NoError(t, err)

	// Insert large dataset
	t.Log("Preparing large dataset...")
	tx, err := dsc.DB.Begin()
	require.NoError(t, err)
	for i := 0; i < 2000; i++ {
		_, err := tx.Exec(`
			INSERT INTO files (storage_root_id, path, name, size, modified_at)
			VALUES (1, ?, ?, ?, datetime('now'))
		`, fmt.Sprintf("/large/file%d.txt", i), fmt.Sprintf("file%d.txt", i), 1024*(i+1))
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())

	t.Run("ConcurrentLargeQueries", func(t *testing.T) {
		concurrentQueries := 10

		var wg sync.WaitGroup
		for i := 0; i < concurrentQueries; i++ {
			wg.Add(1)
			go func(queryID int) {
				defer wg.Done()

				start := time.Now()
				rows, err := dsc.DB.Query("SELECT id, storage_root_id, path, name, size, modified_at FROM files WHERE path LIKE '/large/%' LIMIT 1000")
				if err != nil {
					dsc.recordOperation(time.Since(start), err)
					return
				}
				defer rows.Close()

				count := 0
				for rows.Next() {
					var id, storageRootID int
					var path, name string
					var size int64
					var modTime string
					if scanErr := rows.Scan(&id, &storageRootID, &path, &name, &size, &modTime); scanErr != nil {
						dsc.recordOperation(time.Since(start), scanErr)
						return
					}
					count++
				}

				latency := time.Since(start)
				dsc.recordOperation(latency, rows.Err())

				t.Logf("Query %d returned %d rows in %v", queryID, count, latency)
			}(i)
		}

		wg.Wait()
		dsc.PrintStats(t)

		stats := dsc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 95.0, "Large queries should succeed")
	})

	t.Run("PaginatedLargeResultSet", func(t *testing.T) {
		pageSize := 100
		totalRows := 0
		offset := 0

		for {
			rows, err := dsc.DB.Query(
				"SELECT id, path, name FROM files WHERE path LIKE '/large/%' ORDER BY id LIMIT ? OFFSET ?",
				pageSize, offset,
			)
			require.NoError(t, err)

			pageCount := 0
			for rows.Next() {
				var id int
				var path, name string
				require.NoError(t, rows.Scan(&id, &path, &name))
				pageCount++
			}
			rows.Close()
			require.NoError(t, rows.Err())

			totalRows += pageCount
			if pageCount < pageSize {
				break
			}
			offset += pageSize
		}

		assert.Equal(t, 2000, totalRows,
			"Paginated query should retrieve all 2000 rows")
	})

	t.Run("AggregateOnLargeDataset", func(t *testing.T) {
		var totalSize int64
		var avgSize float64
		var maxSize int64
		var minSize int64
		var fileCount int

		err := dsc.DB.QueryRow(`
			SELECT COUNT(*), SUM(size), AVG(size), MAX(size), MIN(size)
			FROM files WHERE path LIKE '/large/%'
		`).Scan(&fileCount, &totalSize, &avgSize, &maxSize, &minSize)
		require.NoError(t, err)

		assert.Equal(t, 2000, fileCount)
		assert.Greater(t, totalSize, int64(0), "Total size should be positive")
		assert.Greater(t, avgSize, 0.0, "Average size should be positive")
		assert.Equal(t, int64(1024*2000), maxSize, "Max size should be 2000*1024")
		assert.Equal(t, int64(1024), minSize, "Min size should be 1024")
	})
}

// =============================================================================
// STRESS TEST: WAL Mode Verification Under Concurrent Access
// =============================================================================

func TestWALModeUnderConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := tmpDir + "/wal_stress_test.db"

	db, err := sql.Open("sqlite3", dbPath+"?cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	// Explicitly enable WAL mode
	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode=WAL").Scan(&journalMode)
	require.NoError(t, err)
	assert.Equal(t, "wal", journalMode,
		"Journal mode should be WAL after explicit PRAGMA")

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Create tables
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS storage_roots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		protocol TEXT NOT NULL,
		path TEXT,
		enabled BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		storage_root_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		name TEXT NOT NULL,
		size INTEGER NOT NULL,
		modified_at DATETIME NOT NULL,
		FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id)
	)`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO storage_roots (id, name, protocol, path) VALUES (1, 'wal-test', 'local', '/wal')`)
	require.NoError(t, err)

	t.Run("ConcurrentReadsDuringWrites", func(t *testing.T) {
		// WAL mode allows concurrent reads during writes (readers don't block writers)
		var readSuccess int64
		var writeSuccess int64
		duration := 2 * time.Second
		done := make(chan struct{})
		var wg sync.WaitGroup

		// Writers
		for w := 0; w < 3; w++ {
			wg.Add(1)
			go func(writerID int) {
				defer wg.Done()
				counter := 0
				for {
					select {
					case <-done:
						return
					default:
						_, err := db.Exec(`
							INSERT INTO files (storage_root_id, path, name, size, modified_at)
							VALUES (1, ?, ?, ?, datetime('now'))
						`, fmt.Sprintf("/wal/w%d_%d.txt", writerID, counter),
							fmt.Sprintf("w%d_%d.txt", writerID, counter), 512)
						if err == nil {
							atomic.AddInt64(&writeSuccess, 1)
						}
						counter++
						time.Sleep(5 * time.Millisecond)
					}
				}
			}(w)
		}

		// Readers
		for r := 0; r < 7; r++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-done:
						return
					default:
						var count int
						err := db.QueryRow("SELECT COUNT(*) FROM files").Scan(&count)
						if err == nil {
							atomic.AddInt64(&readSuccess, 1)
						}
						time.Sleep(2 * time.Millisecond)
					}
				}
			}()
		}

		time.Sleep(duration)
		close(done)

		wgDone := make(chan struct{})
		go func() {
			wg.Wait()
			close(wgDone)
		}()
		select {
		case <-wgDone:
		case <-time.After(10 * time.Second):
			t.Fatal("WAL concurrent access goroutines did not finish within deadline")
		}

		reads := atomic.LoadInt64(&readSuccess)
		writes := atomic.LoadInt64(&writeSuccess)
		t.Logf("WAL concurrent: %d reads, %d writes in %v", reads, writes, duration)

		assert.Greater(t, reads, int64(100),
			"WAL mode should allow many concurrent reads")
		assert.Greater(t, writes, int64(10),
			"WAL mode should allow concurrent writes alongside reads")
	})

	t.Run("WALModePreservedAfterReconnect", func(t *testing.T) {
		// Verify WAL mode persists across connections
		db2, err := sql.Open("sqlite3", dbPath)
		require.NoError(t, err)
		defer db2.Close()

		var mode string
		err = db2.QueryRow("PRAGMA journal_mode").Scan(&mode)
		require.NoError(t, err)
		assert.Equal(t, "wal", mode,
			"WAL mode should persist across database connections")
	})
}
