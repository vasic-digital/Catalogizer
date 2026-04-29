package stress

import (
	"database/sql"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"catalogizer/database"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseResponseTime(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")  // SKIP-OK: #legacy-skip-untriaged-2026-04-29
	}

	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, val TEXT)")
	require.NoError(t, err)

	// Insert 1000 rows
	for i := 0; i < 1000; i++ {
		_, err = db.Exec("INSERT INTO test (val) VALUES (?)", "test")
		require.NoError(t, err)
	}

	// Measure query response time using Query + Scan pattern
	// (avoids context lifecycle issue with QueryRow's deferred cancel)
	start := time.Now()
	iterations := 100
	for i := 0; i < iterations; i++ {
		rows, qErr := db.Query("SELECT COUNT(*) FROM test")
		require.NoError(t, qErr)
		require.True(t, rows.Next())
		var count int
		err = rows.Scan(&count)
		rows.Close()
		require.NoError(t, err)
	}
	elapsed := time.Since(start)
	avgMs := float64(elapsed.Milliseconds()) / float64(iterations)

	t.Logf("Average query time: %.2fms over %d iterations", avgMs, iterations)
	assert.Less(t, avgMs, 50.0, "Average query should be under 50ms")
}

func TestConcurrentQueryThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")  // SKIP-OK: #legacy-skip-untriaged-2026-04-29
	}

	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	defer db.Close()

	_, _ = db.Exec("PRAGMA journal_mode=WAL")
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, val TEXT)")
	require.NoError(t, err)

	// Pre-populate
	for i := 0; i < 100; i++ {
		_, _ = db.Exec("INSERT INTO test (val) VALUES (?)", "data")
	}

	var ops int64
	var wg sync.WaitGroup
	workers := 10
	duration := 3 * time.Second

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			deadline := time.After(duration)
			for {
				select {
				case <-deadline:
					return
				default:
					rows, err := db.Query("SELECT COUNT(*) FROM test")
					if err == nil {
						if rows.Next() {
							var count int
							_ = rows.Scan(&count)
						}
						rows.Close()
					}
					atomic.AddInt64(&ops, 1)
				}
			}
		}()
	}
	wg.Wait()

	opsPerSec := float64(ops) / duration.Seconds()
	t.Logf("Throughput: %.0f ops/sec with %d workers", opsPerSec, workers)
	assert.Greater(t, opsPerSec, 100.0, "Should handle >100 ops/sec")
}
