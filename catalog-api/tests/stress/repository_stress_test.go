package stress

import (
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"catalogizer/internal/tests"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// STRESS TEST: Repository layer (concurrent inserts, reads during writes)
// =============================================================================

func TestRepositoryStress_ConcurrentInserts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	db := tests.SetupTestDB(t)
	defer db.Close()

	// Create a storage root for file inserts
	_, err := db.Exec(
		`INSERT INTO storage_roots (name, protocol, path, enabled)
		 VALUES ('repo-stress', 'local', '/repo-stress', 1)`,
	)
	require.NoError(t, err)

	var rootID int64
	err = db.QueryRow(
		`SELECT id FROM storage_roots WHERE name = 'repo-stress'`,
	).Scan(&rootID)
	require.NoError(t, err)

	t.Run("50ConcurrentFileInserts", func(t *testing.T) {
		writers := 50
		insertsPerWriter := 20
		var successCount, errorCount int64

		var wg sync.WaitGroup
		start := time.Now()

		for w := 0; w < writers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < insertsPerWriter; j++ {
					path := fmt.Sprintf(
						"/repo-stress/w%d/file%d.txt",
						workerID, j)
					name := fmt.Sprintf("file%d.txt", j)

					_, err := db.Exec(
						`INSERT INTO files
						 (storage_root_id, path, name, size,
						  modified_at)
						 VALUES (?, ?, ?, ?, datetime('now'))`,
						rootID, path, name,
						int64(1024*(j+1)),
					)
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
					}
					// Small delay to reduce contention
					time.Sleep(1 * time.Millisecond)
				}
			}(w)
		}

		wg.Wait()
		elapsed := time.Since(start)

		success := atomic.LoadInt64(&successCount)
		errors := atomic.LoadInt64(&errorCount)
		total := int64(writers * insertsPerWriter)

		t.Logf("Repository insert stress:")
		t.Logf("  Duration:  %v", elapsed)
		t.Logf("  Success:   %d/%d", success, total)
		t.Logf("  Errors:    %d", errors)
		t.Logf("  Rate:      %.0f inserts/sec",
			float64(success)/elapsed.Seconds())

		successRate := float64(success) / float64(total) * 100
		assert.Greater(t, successRate, 95.0,
			"Insert success rate should be >95%%")

		// Verify actual count in DB
		var dbCount int
		err := db.QueryRow(
			`SELECT COUNT(*) FROM files
			 WHERE storage_root_id = ?`, rootID,
		).Scan(&dbCount)
		require.NoError(t, err)
		assert.Equal(t, int(success), dbCount,
			"DB count should match successful inserts")
	})
}

func TestRepositoryStress_ConcurrentReadsAndWrites(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Use file-based SQLite with WAL mode for concurrent read/write testing.
	// In-memory SQLite with MaxOpenConns(1) causes goroutines to block waiting
	// for the single connection, leading to hangs.
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/rw_stress_test.db"
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, _ = db.Exec("PRAGMA journal_mode=WAL")
	_, _ = db.Exec("PRAGMA foreign_keys = ON")
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

	_, err = db.Exec(
		`INSERT INTO storage_roots (name, protocol, path, enabled)
		 VALUES ('rw-stress', 'local', '/rw-stress', 1)`,
	)
	require.NoError(t, err)

	var rootID int64
	err = db.QueryRow(
		`SELECT id FROM storage_roots WHERE name = 'rw-stress'`,
	).Scan(&rootID)
	require.NoError(t, err)

	// Pre-populate with some data
	for i := 0; i < 100; i++ {
		db.Exec(
			`INSERT INTO files
			 (storage_root_id, path, name, size, modified_at)
			 VALUES (?, ?, ?, ?, datetime('now'))`,
			rootID,
			fmt.Sprintf("/rw-stress/seed/file%d.txt", i),
			fmt.Sprintf("file%d.txt", i),
			int64(1024*(i+1)),
		)
	}

	t.Run("20ReadersAndWriters", func(t *testing.T) {
		readers := 20
		writers := 20
		duration := 3 * time.Second

		var readOps, writeOps int64
		var readErrors, writeErrors int64
		done := make(chan struct{})
		var wg sync.WaitGroup

		// Reader goroutines
		for r := 0; r < readers; r++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-done:
						return
					default:
						var count int
						err := db.QueryRow(
							`SELECT COUNT(*) FROM files
							 WHERE storage_root_id = ?`,
							rootID,
						).Scan(&count)
						if err != nil {
							atomic.AddInt64(&readErrors, 1)
						} else {
							atomic.AddInt64(&readOps, 1)
						}
						time.Sleep(2 * time.Millisecond)
					}
				}
			}()
		}

		// Writer goroutines
		for w := 0; w < writers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				counter := 0
				for {
					select {
					case <-done:
						return
					default:
						path := fmt.Sprintf(
							"/rw-stress/live/w%d_%d.txt",
							workerID, counter)
						_, err := db.Exec(
							`INSERT INTO files
							 (storage_root_id, path, name,
							  size, modified_at)
							 VALUES (?, ?, 'file.txt',
							  1024, datetime('now'))`,
							rootID, path,
						)
						if err != nil {
							atomic.AddInt64(&writeErrors, 1)
						} else {
							atomic.AddInt64(&writeOps, 1)
						}
						counter++
						time.Sleep(5 * time.Millisecond)
					}
				}
			}(w)
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
			t.Fatal("ConcurrentReadsAndWrites: goroutines did not finish within deadline")
		}

		rOps := atomic.LoadInt64(&readOps)
		wOps := atomic.LoadInt64(&writeOps)
		rErr := atomic.LoadInt64(&readErrors)
		wErr := atomic.LoadInt64(&writeErrors)

		t.Logf("Concurrent read/write results:")
		t.Logf("  Read ops:    %d (errors: %d)", rOps, rErr)
		t.Logf("  Write ops:   %d (errors: %d)", wOps, wErr)
		t.Logf("  Read rate:   %.0f ops/sec",
			float64(rOps)/duration.Seconds())
		t.Logf("  Write rate:  %.0f ops/sec",
			float64(wOps)/duration.Seconds())

		// Reads should mostly succeed
		if rOps+rErr > 0 {
			readSuccessRate := float64(rOps) /
				float64(rOps+rErr) * 100
			assert.Greater(t, readSuccessRate, 90.0,
				"Read success rate should be >90%%")
		}

		// Writes should mostly succeed
		if wOps+wErr > 0 {
			writeSuccessRate := float64(wOps) /
				float64(wOps+wErr) * 100
			assert.Greater(t, writeSuccessRate, 90.0,
				"Write success rate should be >90%%")
		}
	})
}

func TestRepositoryStress_MediaItemCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	db := tests.SetupTestDB(t)
	defer db.Close()

	_, err := db.Exec(
		`INSERT INTO storage_roots (name, protocol, path, enabled)
		 VALUES ('crud-stress', 'local', '/crud', 1)`,
	)
	require.NoError(t, err)

	var rootID int64
	err = db.QueryRow(
		`SELECT id FROM storage_roots WHERE name = 'crud-stress'`,
	).Scan(&rootID)
	require.NoError(t, err)

	t.Run("ConcurrentMediaItemCRUD", func(t *testing.T) {
		workers := 20
		opsPerWorker := 10
		var createOps, readOps, updateOps, deleteOps int64
		var errors int64

		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < opsPerWorker; j++ {
					// Create
					result, err := db.Exec(
						`INSERT INTO media_items
						 (path, title, type, media_type,
						  storage_root_id)
						 VALUES (?, ?, 'movie', 'movie', ?)`,
						fmt.Sprintf("/crud/w%d/m%d.mkv",
							workerID, j),
						fmt.Sprintf("Movie W%d-%d",
							workerID, j),
						rootID,
					)
					if err != nil {
						atomic.AddInt64(&errors, 1)
						continue
					}
					atomic.AddInt64(&createOps, 1)

					id, _ := result.LastInsertId()

					// Read
					var title string
					err = db.QueryRow(
						`SELECT title FROM media_items
						 WHERE id = ?`, id,
					).Scan(&title)
					if err != nil {
						atomic.AddInt64(&errors, 1)
						continue
					}
					atomic.AddInt64(&readOps, 1)

					// Update
					_, err = db.Exec(
						`UPDATE media_items
						 SET title = ?, year = ?
						 WHERE id = ?`,
						title+" (Updated)",
						2020+j%5, id,
					)
					if err != nil {
						atomic.AddInt64(&errors, 1)
						continue
					}
					atomic.AddInt64(&updateOps, 1)

					// Delete (every other item)
					if j%2 == 0 {
						_, err = db.Exec(
							`DELETE FROM media_items
							 WHERE id = ?`, id)
						if err != nil {
							atomic.AddInt64(&errors, 1)
						} else {
							atomic.AddInt64(&deleteOps, 1)
						}
					}
				}
			}(w)
		}

		wg.Wait()

		c := atomic.LoadInt64(&createOps)
		r := atomic.LoadInt64(&readOps)
		u := atomic.LoadInt64(&updateOps)
		d := atomic.LoadInt64(&deleteOps)
		e := atomic.LoadInt64(&errors)

		t.Logf("CRUD stress results:")
		t.Logf("  Creates: %d", c)
		t.Logf("  Reads:   %d", r)
		t.Logf("  Updates: %d", u)
		t.Logf("  Deletes: %d", d)
		t.Logf("  Errors:  %d", e)

		total := c + r + u + d + e
		successRate := float64(c+r+u+d) / float64(total) * 100

		assert.Greater(t, successRate, 90.0,
			"CRUD success rate should be >90%%")
		assert.Greater(t, c, int64(0),
			"Should have successful creates")
		assert.Greater(t, r, int64(0),
			"Should have successful reads")
		assert.Greater(t, u, int64(0),
			"Should have successful updates")
		assert.Greater(t, d, int64(0),
			"Should have successful deletes")
	})
}

func TestRepositoryStress_UserOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	db := tests.SetupTestDB(t)
	defer db.Close()

	t.Run("ConcurrentUserInserts", func(t *testing.T) {
		workers := 20
		usersPerWorker := 10
		var success, errors int64

		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < usersPerWorker; j++ {
					username := fmt.Sprintf(
						"stress_user_w%d_u%d", workerID, j)
					email := fmt.Sprintf(
						"%s@test.com", username)

					_, err := db.Exec(
						`INSERT INTO users
						 (username, email, password_hash,
						  is_active, role_id)
						 VALUES (?, ?, 'hash', 1, 1)`,
						username, email,
					)
					if err != nil {
						atomic.AddInt64(&errors, 1)
					} else {
						atomic.AddInt64(&success, 1)
					}
				}
			}(w)
		}

		wg.Wait()

		s := atomic.LoadInt64(&success)
		e := atomic.LoadInt64(&errors)
		expected := int64(workers * usersPerWorker)

		t.Logf("User inserts: %d success, %d errors "+
			"(expected %d)", s, e, expected)

		successRate := float64(s) / float64(expected) * 100
		assert.Greater(t, successRate, 95.0,
			"User insert success rate should be >95%%")

		// Verify count
		var count int
		err := db.QueryRow(
			`SELECT COUNT(*) FROM users
			 WHERE username LIKE 'stress_user_%'`,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, int(s), count,
			"DB count should match successful inserts")
	})
}

func TestRepositoryStress_PlaylistOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	db := tests.SetupTestDB(t)
	defer db.Close()

	userID := 1 // Default test user

	t.Run("ConcurrentPlaylistCreation", func(t *testing.T) {
		workers := 20
		playlistsPerWorker := 5
		var success int64

		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < playlistsPerWorker; j++ {
					name := fmt.Sprintf(
						"Stress Playlist W%d-%d", workerID, j)
					_, err := db.Exec(
						`INSERT INTO playlists
						 (user_id, name, description)
						 VALUES (?, ?, 'stress test')`,
						userID, name,
					)
					if err == nil {
						atomic.AddInt64(&success, 1)
					}
				}
			}(w)
		}

		wg.Wait()

		s := atomic.LoadInt64(&success)
		expected := int64(workers * playlistsPerWorker)

		t.Logf("Playlist creation: %d/%d succeeded", s, expected)

		successRate := float64(s) / float64(expected) * 100
		assert.Greater(t, successRate, 95.0,
			"Playlist creation success rate should be >95%%")
	})
}

func TestRepositoryStress_TransactionConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	db := tests.SetupTestDB(t)
	defer db.Close()

	_, err := db.Exec(
		`INSERT INTO storage_roots (name, protocol, path, enabled)
		 VALUES ('tx-stress', 'local', '/tx', 1)`,
	)
	require.NoError(t, err)

	var rootID int64
	err = db.QueryRow(
		`SELECT id FROM storage_roots WHERE name = 'tx-stress'`,
	).Scan(&rootID)
	require.NoError(t, err)

	t.Run("ConcurrentTransactionsWithRollback", func(t *testing.T) {
		workers := 10
		txPerWorker := 5
		var committed, rolledBack int64

		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < txPerWorker; j++ {
					tx, err := db.Begin()
					if err != nil {
						continue
					}

					// Insert a media item
					_, err = tx.Exec(
						`INSERT INTO media_items
						 (path, title, type, media_type,
						  storage_root_id)
						 VALUES (?, ?, 'movie', 'movie', ?)`,
						fmt.Sprintf("/tx/w%d/m%d.mkv",
							workerID, j),
						fmt.Sprintf("TX Movie %d-%d",
							workerID, j),
						rootID,
					)
					if err != nil {
						tx.Rollback()
						atomic.AddInt64(&rolledBack, 1)
						continue
					}

					// Simulate: rollback every 3rd transaction
					if j%3 == 2 {
						tx.Rollback()
						atomic.AddInt64(&rolledBack, 1)
					} else {
						if err := tx.Commit(); err != nil {
							atomic.AddInt64(&rolledBack, 1)
						} else {
							atomic.AddInt64(&committed, 1)
						}
					}
				}
			}(w)
		}

		wg.Wait()

		c := atomic.LoadInt64(&committed)
		r := atomic.LoadInt64(&rolledBack)

		t.Logf("Transaction results: %d committed, %d rolled back",
			c, r)

		// Verify only committed items exist
		var count int
		err := db.QueryRow(
			`SELECT COUNT(*) FROM media_items
			 WHERE path LIKE '/tx/%'`,
		).Scan(&count)
		require.NoError(t, err)

		assert.Equal(t, int(c), count,
			"DB should only contain committed records")
		assert.Greater(t, c, int64(0),
			"Some transactions should commit")
		assert.Greater(t, r, int64(0),
			"Some transactions should roll back")
	})
}
