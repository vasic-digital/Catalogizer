package stress

import (
	"catalogizer/internal/tests"
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DatabaseStressContext manages database stress test execution
type DatabaseStressContext struct {
	DB              *sql.DB
	OperationCount  int64
	SuccessCount    int64
	ErrorCount      int64
	TotalLatency    int64
	StartTime       time.Time
	Errors          []error
	ErrorsMutex     sync.Mutex
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
// STRESS TEST: Concurrent Database Reads
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
		assert.Greater(t, stats["success_rate"].(float64), 99.0, "Read operations should have >99% success")
		assert.Less(t, stats["avg_latency"].(time.Duration), 10*time.Millisecond, "Avg read latency should be <10ms")
	})
}

// =============================================================================
// STRESS TEST: Concurrent Database Writes
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
		assert.Greater(t, stats["success_rate"].(float64), 95.0, "Write operations should have >95% success")
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
		assert.Greater(t, stats["success_rate"].(float64), 95.0, "Update operations should have >95% success")
	})
}

// =============================================================================
// STRESS TEST: Mixed Read/Write Workload
// =============================================================================

func TestMixedReadWriteWorkload(t *testing.T) {
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

	// Pre-populate with some data
	for i := 0; i < 100; i++ {
		dsc.DB.Exec(`
			INSERT INTO files (storage_root_id, path, name, size, modified_at)
			VALUES (1, ?, ?, ?, datetime('now'))
		`, fmt.Sprintf("/mixed/file%d.txt", i), fmt.Sprintf("file%d.txt", i), 1024*(i+1))
	}

	t.Run("70PercentReads30PercentWrites", func(t *testing.T) {
		duration := 15 * time.Second
		concurrentWorkers := 50

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
		wg.Wait()

		dsc.PrintStats(t)

		stats := dsc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 90.0, "Mixed workload should have >90% success")
		assert.Greater(t, stats["ops_per_sec"].(float64), 100.0, "Should handle >100 ops/sec")
	})
}

// =============================================================================
// STRESS TEST: Transaction Stress
// =============================================================================

func TestTransactionStress(t *testing.T) {
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
		assert.Greater(t, stats["success_rate"].(float64), 95.0, "Transactions should have >95% success")

		// Verify all transactions completed
		var count int
		err := dsc.DB.QueryRow("SELECT COUNT(*) FROM files WHERE path LIKE '/tx/%'").Scan(&count)
		require.NoError(t, err)
		expectedCount := concurrentTxs * operationsPerTx
		assert.Equal(t, expectedCount, count, "All transaction operations should complete")
	})
}

// =============================================================================
// STRESS TEST: Connection Pool Stress
// =============================================================================

func TestConnectionPoolStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	t.Skip("Connection pool stress test incompatible with in-memory SQLite - requires production database")
	// NOTE: This test requires a real database connection pool (PostgreSQL, MySQL).
	// SQLite :memory: databases MUST use MaxOpenConns=1 (each connection creates a separate DB).
	// This test is designed for production database validation where connection pooling matters.

	dsc := newDatabaseStressContext(t)
	defer dsc.DB.Close()

	// Create test storage root
	_, err := dsc.DB.Exec(`
		INSERT INTO storage_roots (id, name, protocol, path, enabled)
		VALUES (1, 'test-root', 'local', '/test', 1)
	`)
	require.NoError(t, err)

	// Configure connection pool
	dsc.DB.SetMaxOpenConns(25)
	dsc.DB.SetMaxIdleConns(10)
	dsc.DB.SetConnMaxLifetime(5 * time.Minute)

	t.Run("ExceedConnectionPool", func(t *testing.T) {
		// Try to create more concurrent operations than pool size
		concurrentOps := 100 // More than max open connections
		duration := 10 * time.Second

		done := make(chan bool)
		var wg sync.WaitGroup

		for i := 0; i < concurrentOps; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for {
					select {
					case <-done:
						return
					default:
						start := time.Now()
						var count int
						err := dsc.DB.QueryRow("SELECT COUNT(*) FROM files").Scan(&count)
						latency := time.Since(start)
						dsc.recordOperation(latency, err)

						time.Sleep(20 * time.Millisecond)
					}
				}
			}()
		}

		time.Sleep(duration)
		close(done)
		wg.Wait()

		dsc.PrintStats(t)

		stats := dsc.GetStats()
		assert.Greater(t, stats["success_rate"].(float64), 95.0, "Should handle connection pool saturation gracefully")

		// Check pool stats
		dbStats := dsc.DB.Stats()
		t.Logf("DB Pool Stats:")
		t.Logf("  Max Open Connections: %d", dbStats.MaxOpenConnections)
		t.Logf("  Open Connections: %d", dbStats.OpenConnections)
		t.Logf("  In Use: %d", dbStats.InUse)
		t.Logf("  Idle: %d", dbStats.Idle)
		t.Logf("  Wait Count: %d", dbStats.WaitCount)
		t.Logf("  Wait Duration: %v", dbStats.WaitDuration)
	})
}

// =============================================================================
// STRESS TEST: Large Query Results
// =============================================================================

func TestLargeQueryResults(t *testing.T) {
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
	tx, _ := dsc.DB.Begin()
	for i := 0; i < 10000; i++ {
		tx.Exec(`
			INSERT INTO files (storage_root_id, path, name, size, modified_at)
			VALUES (1, ?, ?, ?, datetime('now'))
		`, fmt.Sprintf("/large/file%d.txt", i), fmt.Sprintf("file%d.txt", i), 1024*(i+1))
	}
	tx.Commit()

	t.Run("ConcurrentLargeQueries", func(t *testing.T) {
		concurrentQueries := 10

		var wg sync.WaitGroup
		for i := 0; i < concurrentQueries; i++ {
			wg.Add(1)
			go func(queryID int) {
				defer wg.Done()

				start := time.Now()
				rows, err := dsc.DB.Query("SELECT * FROM files WHERE path LIKE '/large/%' LIMIT 1000")
				if err != nil {
					dsc.recordOperation(time.Since(start), err)
					return
				}
				defer rows.Close()

				count := 0
				for rows.Next() {
					var id int
					var path, name string
					var size int64
					var modTime string
					rows.Scan(&id, &path, &name, &size, &modTime)
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
}
