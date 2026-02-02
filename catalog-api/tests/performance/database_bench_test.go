package performance

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/mutecomm/go-sqlcipher"
)

// benchDB creates a fresh in-memory SQLite database with schema and seed data.
func benchDB(b *testing.B, fileCount int) *sql.DB {
	b.Helper()

	db, err := sql.Open("sqlite3", ":memory:?_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=1")
	if err != nil {
		b.Fatalf("open db: %v", err)
	}

	if err := createSchema(db); err != nil {
		b.Fatalf("create schema: %v", err)
	}
	if err := seedData(db, fileCount); err != nil {
		b.Fatalf("seed data: %v", err)
	}

	return db
}

// ---------------------------------------------------------------------------
// Benchmark: single-row lookup by primary key
// ---------------------------------------------------------------------------

func BenchmarkDB_SelectByID(b *testing.B) {
	sizes := []int{100, 1000, 5000}
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("rows=%d", sz), func(b *testing.B) {
			db := benchDB(b, sz)
			defer db.Close()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				id := (i % sz) + 1
				var name, path string
				var size int64
				err := db.QueryRow("SELECT name, path, size FROM files WHERE id = ?", id).Scan(&name, &path, &size)
				if err != nil {
					b.Fatalf("query by id: %v", err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Benchmark: indexed range scan (storage_root_id + path prefix)
// ---------------------------------------------------------------------------

func BenchmarkDB_IndexedRangeScan(b *testing.B) {
	db := benchDB(b, 5000)
	defer db.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rootID := (i % 5) + 1
		prefix := fmt.Sprintf("/media/collection_%d/%%", i%50)
		rows, err := db.Query(
			"SELECT id, name, size FROM files WHERE storage_root_id = ? AND path LIKE ? AND deleted = 0 LIMIT 50",
			rootID, prefix,
		)
		if err != nil {
			b.Fatalf("range scan: %v", err)
		}
		count := 0
		for rows.Next() {
			var id int
			var name string
			var size int64
			rows.Scan(&id, &name, &size)
			count++
		}
		rows.Close()
	}
}

// ---------------------------------------------------------------------------
// Benchmark: LIKE search (simulates user text search)
// ---------------------------------------------------------------------------

func BenchmarkDB_LikeSearch(b *testing.B) {
	sizes := []int{500, 2000, 5000}
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("rows=%d", sz), func(b *testing.B) {
			db := benchDB(b, sz)
			defer db.Close()

			patterns := []string{"%file_000%", "%file_001%", "%.mkv", "%.mp4", "%collection_1%"}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				pat := patterns[i%len(patterns)]
				rows, err := db.Query(
					"SELECT id, name, path, size FROM files WHERE deleted = 0 AND name LIKE ? LIMIT 50", pat)
				if err != nil {
					b.Fatalf("like search: %v", err)
				}
				for rows.Next() {
					var id int
					var name, path string
					var size int64
					rows.Scan(&id, &name, &path, &size)
				}
				rows.Close()
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Benchmark: aggregate queries (COUNT, SUM, GROUP BY)
// ---------------------------------------------------------------------------

func BenchmarkDB_AggregateCount(b *testing.B) {
	db := benchDB(b, 5000)
	defer db.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var count int
		db.QueryRow("SELECT COUNT(*) FROM files WHERE deleted = 0").Scan(&count)
	}
}

func BenchmarkDB_AggregateSum(b *testing.B) {
	db := benchDB(b, 5000)
	defer db.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var totalSize int64
		db.QueryRow("SELECT COALESCE(SUM(size), 0) FROM files WHERE deleted = 0").Scan(&totalSize)
	}
}

func BenchmarkDB_GroupByFileType(b *testing.B) {
	db := benchDB(b, 5000)
	defer db.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rows, err := db.Query(
			"SELECT file_type, COUNT(*), SUM(size) FROM files WHERE deleted = 0 GROUP BY file_type")
		if err != nil {
			b.Fatalf("group by: %v", err)
		}
		for rows.Next() {
			var ft string
			var cnt int
			var sz int64
			rows.Scan(&ft, &cnt, &sz)
		}
		rows.Close()
	}
}

func BenchmarkDB_GroupByStorageRoot(b *testing.B) {
	db := benchDB(b, 5000)
	defer db.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rows, err := db.Query(`
			SELECT sr.name, sr.protocol, COUNT(f.id), COALESCE(SUM(f.size), 0)
			FROM storage_roots sr
			LEFT JOIN files f ON f.storage_root_id = sr.id AND f.deleted = 0
			GROUP BY sr.id`)
		if err != nil {
			b.Fatalf("group by root: %v", err)
		}
		for rows.Next() {
			var name, proto string
			var cnt int
			var sz int64
			rows.Scan(&name, &proto, &cnt, &sz)
		}
		rows.Close()
	}
}

// ---------------------------------------------------------------------------
// Benchmark: INSERT single row
// ---------------------------------------------------------------------------

func BenchmarkDB_InsertSingle(b *testing.B) {
	db := benchDB(b, 0)
	defer db.Close()

	now := time.Now()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := db.Exec(
			`INSERT INTO files (storage_root_id, path, name, extension, file_type, size, modified_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			1,
			fmt.Sprintf("/bench/file_%d.mkv", i),
			fmt.Sprintf("file_%d.mkv", i),
			".mkv", "video", int64(i)*1024*1024, now,
		)
		if err != nil {
			b.Fatalf("insert: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// Benchmark: batch INSERT in a transaction
// ---------------------------------------------------------------------------

func BenchmarkDB_BatchInsert(b *testing.B) {
	batchSizes := []int{10, 50, 200}
	for _, bsz := range batchSizes {
		b.Run(fmt.Sprintf("batch=%d", bsz), func(b *testing.B) {
			db := benchDB(b, 0)
			defer db.Close()

			now := time.Now()
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				tx, _ := db.Begin()
				stmt, _ := tx.Prepare(
					`INSERT INTO files (storage_root_id, path, name, extension, file_type, size, modified_at)
					 VALUES (?, ?, ?, ?, ?, ?, ?)`)
				for j := 0; j < bsz; j++ {
					idx := i*bsz + j
					stmt.Exec(
						(idx%5)+1,
						fmt.Sprintf("/batch/file_%d.mkv", idx),
						fmt.Sprintf("file_%d.mkv", idx),
						".mkv", "video", int64(idx)*512*1024, now,
					)
				}
				stmt.Close()
				tx.Commit()
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Benchmark: UPDATE with indexed lookup
// ---------------------------------------------------------------------------

func BenchmarkDB_UpdateByID(b *testing.B) {
	db := benchDB(b, 1000)
	defer db.Close()

	now := time.Now()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		id := (i % 1000) + 1
		_, err := db.Exec(
			"UPDATE files SET size = ?, modified_at = ? WHERE id = ?",
			int64(i)*2048, now, id,
		)
		if err != nil {
			b.Fatalf("update: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// Benchmark: DELETE with indexed lookup
// ---------------------------------------------------------------------------

func BenchmarkDB_SoftDeleteByID(b *testing.B) {
	db := benchDB(b, 5000)
	defer db.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		id := (i % 5000) + 1
		db.Exec("UPDATE files SET deleted = 1 WHERE id = ? AND deleted = 0", id)
	}
}

// ---------------------------------------------------------------------------
// Benchmark: index effectiveness -- compare with/without index
// ---------------------------------------------------------------------------

func BenchmarkDB_IndexEffectiveness_FileType(b *testing.B) {
	// With index (created by default schema)
	b.Run("with_index", func(b *testing.B) {
		db := benchDB(b, 5000)
		defer db.Close()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var count int
			db.QueryRow("SELECT COUNT(*) FROM files WHERE file_type = 'video' AND deleted = 0").Scan(&count)
		}
	})

	// Without index (drop the file_type index)
	b.Run("without_index", func(b *testing.B) {
		db := benchDB(b, 5000)
		defer db.Close()

		db.Exec("DROP INDEX IF EXISTS idx_files_file_type")

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var count int
			db.QueryRow("SELECT COUNT(*) FROM files WHERE file_type = 'video' AND deleted = 0").Scan(&count)
		}
	})
}

func BenchmarkDB_IndexEffectiveness_StorageRootPath(b *testing.B) {
	b.Run("with_composite_index", func(b *testing.B) {
		db := benchDB(b, 5000)
		defer db.Close()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			rootID := (i % 5) + 1
			rows, _ := db.Query(
				"SELECT id, name FROM files WHERE storage_root_id = ? AND path LIKE '/media/collection_1/%' AND deleted = 0 LIMIT 20",
				rootID)
			if rows != nil {
				for rows.Next() {
					var id int
					var name string
					rows.Scan(&id, &name)
				}
				rows.Close()
			}
		}
	})

	b.Run("without_composite_index", func(b *testing.B) {
		db := benchDB(b, 5000)
		defer db.Close()

		db.Exec("DROP INDEX IF EXISTS idx_files_storage_root_path")

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			rootID := (i % 5) + 1
			rows, _ := db.Query(
				"SELECT id, name FROM files WHERE storage_root_id = ? AND path LIKE '/media/collection_1/%' AND deleted = 0 LIMIT 20",
				rootID)
			if rows != nil {
				for rows.Next() {
					var id int
					var name string
					rows.Scan(&id, &name)
				}
				rows.Close()
			}
		}
	})
}

func BenchmarkDB_IndexEffectiveness_UsernameLookup(b *testing.B) {
	b.Run("with_index", func(b *testing.B) {
		db := benchDB(b, 100)
		defer db.Close()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var id int
			db.QueryRow("SELECT id FROM users WHERE username = ?", "benchuser").Scan(&id)
		}
	})

	b.Run("without_index", func(b *testing.B) {
		db := benchDB(b, 100)
		defer db.Close()

		db.Exec("DROP INDEX IF EXISTS idx_users_username")

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var id int
			db.QueryRow("SELECT id FROM users WHERE username = ?", "benchuser").Scan(&id)
		}
	})
}

// ---------------------------------------------------------------------------
// Benchmark: concurrent reads
// ---------------------------------------------------------------------------

func BenchmarkDB_ConcurrentReads(b *testing.B) {
	db := benchDB(b, 2000)
	defer db.Close()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		localRand := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			id := localRand.Intn(2000) + 1
			var name string
			var size int64
			db.QueryRow("SELECT name, size FROM files WHERE id = ?", id).Scan(&name, &size)
		}
	})
}

// ---------------------------------------------------------------------------
// Benchmark: concurrent reads + writes (realistic mixed load)
// ---------------------------------------------------------------------------

func BenchmarkDB_ConcurrentReadWrite(b *testing.B) {
	db := benchDB(b, 2000)
	defer db.Close()

	now := time.Now()
	var writeCounter uint64

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		localRand := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			// 80% reads, 20% writes
			if localRand.Intn(5) == 0 {
				// Write: insert a new file
				idx := atomic.AddUint64(&writeCounter, 1)
				db.Exec(
					`INSERT INTO files (storage_root_id, path, name, extension, file_type, size, modified_at)
					 VALUES (?, ?, ?, ?, ?, ?, ?)`,
					(idx%5)+1,
					fmt.Sprintf("/concurrent/file_%d.mkv", idx),
					fmt.Sprintf("file_%d.mkv", idx),
					".mkv", "video", int64(idx)*1024, now,
				)
			} else {
				// Read: random lookup
				id := localRand.Intn(2000) + 1
				var name string
				var size int64
				db.QueryRow("SELECT name, size FROM files WHERE id = ?", id).Scan(&name, &size)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// Benchmark: prepared statement reuse vs ad-hoc
// ---------------------------------------------------------------------------

func BenchmarkDB_PreparedVsAdHoc(b *testing.B) {
	db := benchDB(b, 1000)
	defer db.Close()

	b.Run("prepared", func(b *testing.B) {
		stmt, err := db.Prepare("SELECT name, path, size FROM files WHERE id = ?")
		if err != nil {
			b.Fatal(err)
		}
		defer stmt.Close()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var name, path string
			var size int64
			stmt.QueryRow((i%1000)+1).Scan(&name, &path, &size)
		}
	})

	b.Run("ad_hoc", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var name, path string
			var size int64
			db.QueryRow("SELECT name, path, size FROM files WHERE id = ?", (i%1000)+1).Scan(&name, &path, &size)
		}
	})
}

// ---------------------------------------------------------------------------
// Benchmark: complex join query (sources with aggregated file stats)
// ---------------------------------------------------------------------------

func BenchmarkDB_JoinQuery(b *testing.B) {
	sizes := []int{500, 2000, 5000}
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("rows=%d", sz), func(b *testing.B) {
			db := benchDB(b, sz)
			defer db.Close()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				rows, err := db.Query(`
					SELECT sr.id, sr.name, sr.protocol,
					       COUNT(f.id) as file_count,
					       COALESCE(SUM(f.size), 0) as total_size,
					       COUNT(CASE WHEN f.is_duplicate = 1 THEN 1 END) as dup_count
					FROM storage_roots sr
					LEFT JOIN files f ON f.storage_root_id = sr.id AND f.deleted = 0
					GROUP BY sr.id
					ORDER BY total_size DESC`)
				if err != nil {
					b.Fatalf("join: %v", err)
				}
				for rows.Next() {
					var id int
					var name, proto string
					var fc, dc int
					var ts int64
					rows.Scan(&id, &name, &proto, &fc, &ts, &dc)
				}
				rows.Close()
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Benchmark: duplicate detection query
// ---------------------------------------------------------------------------

func BenchmarkDB_DuplicateDetectionQuery(b *testing.B) {
	db := benchDB(b, 5000)
	defer db.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rows, err := db.Query(`
			SELECT name, size, COUNT(*) as cnt
			FROM files
			WHERE deleted = 0 AND is_directory = 0
			GROUP BY name, size
			HAVING cnt > 1
			ORDER BY cnt DESC, size DESC
			LIMIT 20`)
		if err != nil {
			b.Fatalf("dup query: %v", err)
		}
		for rows.Next() {
			var name string
			var size int64
			var cnt int
			rows.Scan(&name, &size, &cnt)
		}
		rows.Close()
	}
}

// ---------------------------------------------------------------------------
// Benchmark: transaction throughput
// ---------------------------------------------------------------------------

func BenchmarkDB_TransactionThroughput(b *testing.B) {
	db := benchDB(b, 0)
	defer db.Close()

	now := time.Now()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tx, err := db.Begin()
		if err != nil {
			b.Fatal(err)
		}

		// Simulate a scan operation: insert 5 files then update the storage root
		for j := 0; j < 5; j++ {
			idx := i*5 + j
			tx.Exec(
				`INSERT INTO files (storage_root_id, path, name, extension, file_type, size, modified_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?)`,
				1, fmt.Sprintf("/tx/file_%d.mp4", idx),
				fmt.Sprintf("file_%d.mp4", idx),
				".mp4", "video", int64(idx)*2048, now,
			)
		}

		tx.Exec("UPDATE storage_roots SET last_scan_at = ?, updated_at = ? WHERE id = ?", now, now, 1)

		if err := tx.Commit(); err != nil {
			b.Fatalf("commit: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// Benchmark: concurrent transaction contention
// ---------------------------------------------------------------------------

func BenchmarkDB_ConcurrentTransactions(b *testing.B) {
	db := benchDB(b, 500)
	defer db.Close()

	now := time.Now()
	var counter uint64

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx := atomic.AddUint64(&counter, 1)
			tx, err := db.Begin()
			if err != nil {
				// SQLite may return busy; retry is acceptable in benchmarks
				continue
			}

			tx.Exec(
				`INSERT INTO files (storage_root_id, path, name, extension, file_type, size, modified_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?)`,
				(idx%5)+1,
				fmt.Sprintf("/ctx/file_%d.mkv", idx),
				fmt.Sprintf("file_%d.mkv", idx),
				".mkv", "video", int64(idx)*1024, now,
			)

			tx.Commit()
		}
	})
}

// ---------------------------------------------------------------------------
// Benchmark: WAL mode vs journal mode comparison
// ---------------------------------------------------------------------------

func BenchmarkDB_WALvsJournal(b *testing.B) {
	modes := []struct {
		name string
		dsn  string
	}{
		{"WAL", ":memory:?_journal_mode=WAL&_synchronous=NORMAL"},
		{"DELETE", ":memory:?_journal_mode=DELETE&_synchronous=FULL"},
	}

	for _, mode := range modes {
		b.Run(mode.name, func(b *testing.B) {
			db, err := sql.Open("sqlite3", mode.dsn)
			if err != nil {
				b.Fatal(err)
			}
			defer db.Close()

			if err := createSchema(db); err != nil {
				b.Fatal(err)
			}

			now := time.Now()
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				tx, _ := db.Begin()
				for j := 0; j < 10; j++ {
					idx := i*10 + j
					tx.Exec(
						`INSERT INTO files (storage_root_id, path, name, extension, file_type, size, modified_at)
						 VALUES (?, ?, ?, ?, ?, ?, ?)`,
						1, fmt.Sprintf("/wal/file_%d.mkv", idx),
						fmt.Sprintf("file_%d.mkv", idx),
						".mkv", "video", int64(idx)*1024, now,
					)
				}
				tx.Commit()
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Benchmark: pagination patterns
// ---------------------------------------------------------------------------

func BenchmarkDB_Pagination(b *testing.B) {
	db := benchDB(b, 5000)
	defer db.Close()

	pages := []struct {
		name   string
		offset int
	}{
		{"page_1", 0},
		{"page_10", 450},
		{"page_50", 2450},
		{"page_100", 4950},
	}

	for _, p := range pages {
		b.Run(p.name, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				rows, err := db.Query(
					"SELECT id, name, path, size FROM files WHERE deleted = 0 ORDER BY id LIMIT 50 OFFSET ?",
					p.offset)
				if err != nil {
					b.Fatal(err)
				}
				for rows.Next() {
					var id int
					var name, path string
					var size int64
					rows.Scan(&id, &name, &path, &size)
				}
				rows.Close()
			}
		})
	}
}

