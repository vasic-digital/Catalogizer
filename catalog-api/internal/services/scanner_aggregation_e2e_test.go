package services

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/filesystem"
	"catalogizer/models"
	"catalogizer/repository"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// scannerE2EDBCounter gives each test a uniquely-named shared-cache in-memory DB.
var scannerE2EDBCounter int

// setupScannerE2EDB builds a real in-memory SQLite DB with the full migration
// chain applied. Shared-cache mode is required because insertFileRecord opens
// nested BeginTx/QueryRow operations that need concurrent connections.
func setupScannerE2EDB(t *testing.T) *database.DB {
	t.Helper()
	scannerE2EDBCounter++
	dsn := fmt.Sprintf("file:scanagge2e%d?mode=memory&cache=shared&_foreign_keys=1", scannerE2EDBCounter)
	sqlDB, err := sql.Open("sqlite3", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	db := database.WrapDB(sqlDB, database.DialectSQLite)
	require.NoError(t, db.RunMigrations(context.Background()))
	return db
}

// fakeScanNode is one node in a fake SMB tree for the e2e reproduction.
type fakeScanNode struct {
	name     string
	isDir    bool
	size     int64
	children []fakeScanNode
}

// driveSMBScan replays exactly what SMBScanner.scanDirectory does — it walks the
// fake tree, computes fullPath := filepath.Join(path, name) for each node, and
// calls the REAL insertFileRecord (the shared scanner persistence). This exercises
// the real ensureDirectoryPathExists + parent_id arithmetic without needing a live
// SMB server, so it faithfully reproduces what a real SMB share scan writes to the DB.
func driveSMBScan(t *testing.T, db *database.DB, job ScanJob, status *ScanStatus, path string, nodes []fakeScanNode) {
	t.Helper()
	for _, n := range nodes {
		fullPath := filepath.Join(path, n.name)
		fi := &filesystem.FileInfo{
			Name:    n.name,
			Size:    n.size,
			IsDir:   n.isDir,
			ModTime: time.Now(),
		}
		require.NoError(t, insertFileRecord(context.Background(), db, fullPath, fi, job, status, zap.NewNop()))
		if n.isDir {
			driveSMBScan(t, db, job, status, fullPath, n.children)
		}
	}
}

// TestSMBScanThenAggregate_CreatesEntities is the FINDING-3 reproduction:
// a realistic SMB share scan (job.Path = ".") indexes titled directories
// directly under the share root (e.g. "The Matrix (1999)/<file>"), then
// AggregateAfterScan MUST create > 0 browse entities.
//
// On the pre-fix code this is RED (0 entities): ensureDirectoryPathExists
// early-returned nil for any parent path WITHOUT a "/" separator
// (universal_scanner.go), so a file living directly inside a single-component
// top-level directory got parent_id = NULL instead of pointing at that
// directory. getTopLevelDirectories then found the titled directory at the top
// level but its child-files lookup (parent_id = dirID) returned nothing, so the
// directory — and all its media — was dropped (entities_created = 0).
func TestSMBScanThenAggregate_CreatesEntities(t *testing.T) {
	db := setupScannerE2EDB(t)
	ctx := context.Background()
	logger := zap.NewNop()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	svc := NewAggregationService(db, logger, itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	// Storage root inserted by the scanner on first file (mirrors production):
	// we pre-create it so we know its ID for AggregateAfterScan.
	rootID, err := db.InsertReturningID(ctx,
		`INSERT INTO storage_roots (name, protocol, path, enabled, max_depth) VALUES (?, 'smb', '/', 1, 10)`,
		"testshare")
	require.NoError(t, err)

	job := ScanJob{
		ID:          "job-finding3",
		StorageRoot: &models.StorageRoot{ID: rootID, Name: "testshare", Protocol: "smb"},
		Path:        ".", // common default: scan the share root
		MaxDepth:    10,
		Context:     ctx,
	}
	status := &ScanStatus{JobID: job.ID, StartTime: time.Now(), Status: "running"}

	// Realistic SMB share layout: per-title folders directly under the share
	// root, each holding the media file. This is the FINDING-3 topology.
	tree := []fakeScanNode{
		{name: "The Matrix (1999)", isDir: true, children: []fakeScanNode{
			{name: "The.Matrix.1999.1080p.mkv", size: 4_000_000_000},
			{name: "The.Matrix.1999.srt", size: 50_000},
		}},
		{name: "Inception (2010)", isDir: true, children: []fakeScanNode{
			{name: "Inception.2010.mkv", size: 3_000_000_000},
		}},
	}
	driveSMBScan(t, db, job, status, job.Path, tree)

	// Sanity: the scanner actually wrote file rows (not asserting the bug yet).
	var fileCount int
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM files WHERE storage_root_id = ? AND deleted = 0`, rootID).Scan(&fileCount))
	require.Greater(t, fileCount, 0, "scanner must have indexed files")

	// Run post-scan aggregation.
	require.NoError(t, svc.AggregateAfterScan(ctx, rootID))

	// The browse layer reads media_items entities. There MUST be at least one.
	count, err := itemRepo.Count(ctx)
	require.NoError(t, err)
	require.Greater(t, count, int64(0),
		"AggregateAfterScan must create browse entities from a nested SMB scan (FINDING-3)")

	// And specifically the two movies must be browsable.
	_, movieTypeID, err := itemRepo.GetMediaTypeByName(ctx, "movie")
	require.NoError(t, err)
	matrix, err := itemRepo.GetByTitle(ctx, "The Matrix", movieTypeID)
	require.NoError(t, err)
	require.NotNil(t, matrix, "The Matrix entity must exist for browse")
	inception, err := itemRepo.GetByTitle(ctx, "Inception", movieTypeID)
	require.NoError(t, err)
	require.NotNil(t, inception, "Inception entity must exist for browse")
}
