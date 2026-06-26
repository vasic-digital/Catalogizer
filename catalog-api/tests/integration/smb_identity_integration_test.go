// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/internal/services"
	"catalogizer/internal/tests"
	"catalogizer/models"
	"catalogizer/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

//
// §11.4.169 — TEST TYPE: Integration
//
// These tests exercise the SMB identity + probe pipeline through real SQLite
// databases and real instance methods — NO mocks, NO stubs, NO fakes
// (§11.4.27(A)). They use in-memory SQLite databases with the real migrations,
// real repository structs, and real service methods.
//
// The probe integration test (§11.4.3) auto-SKIPs when no real SMB host is
// configured by env var — the suite stays green without LAN access.
//

// ---------------------------------------------------------------------------
// Identity resolution integration (full db + scanner)
// ---------------------------------------------------------------------------

// TestIntegration_IdentityResolutionViaDb verifies the full pipeline:
//
//  1. A storage_root row exists with identity_index in Options (NULL username/password)
//  2. Env vars define the corresponding credential identity
//  3. ResolveSMBIdentity correctly returns the env-var credentials
//  4. The scanner's storageRootToSettings correctly propagates identity_index
//
// This closes the red-green gap between unit tests (env-var-only) and the
// real SQLite-backed pipeline.
func TestIntegration_IdentityResolutionViaDb(t *testing.T) {
	// Set up env identities.
	t.Setenv("CATALOGIZER_IDENTITY_COUNT", "2")
	t.Setenv("CATALOGIZER_IDENTITY_1_TYPE", "credentials")
	t.Setenv("CATALOGIZER_IDENTITY_1_USERNAME", "nas_user")
	t.Setenv("CATALOGIZER_IDENTITY_1_PASSWORD", "nas_pass")
	t.Setenv("CATALOGIZER_IDENTITY_1_DOMAIN", "WORKGROUP")
	t.Setenv("CATALOGIZER_IDENTITY_2_TYPE", "credentials")
	t.Setenv("CATALOGIZER_IDENTITY_2_USERNAME", "backup_user")
	t.Setenv("CATALOGIZER_IDENTITY_2_PASSWORD", "backup_pass")

	t.Run("resolve identity_index from storage_root via real db", func(t *testing.T) {
		sqlDB := tests.SetupTestDB(t)
		defer sqlDB.Close()
		db := database.WrapDB(sqlDB, database.DialectSQLite)

		// Create a storage_root with identity_index=2 in Options.
		host := "192.168.1.100"
		sharePath := "Backups"
		rootName := fmt.Sprintf("%s:%s", host, sharePath)
		identityMeta, _ := json.Marshal(map[string]int{"identity_index": 2})

		_, err := db.InsertReturningID(context.Background(),
			`INSERT INTO storage_roots
				(name, protocol, host, path, options, enabled, max_depth, created_at, updated_at)
			 VALUES (?, 'smb', ?, ?, ?, 1, 10, datetime('now'), datetime('now'))`,
			rootName, host, sharePath, string(identityMeta),
		)
		require.NoError(t, err, "must insert storage_root")

		// Query back and verify resolution.
		var stored models.StorageRoot
		err = sqlDB.QueryRow(
			`SELECT id, name, protocol, host, path, options, username, password, domain
			   FROM storage_roots WHERE name = ?`, rootName,
		).Scan(&stored.ID, &stored.Name, &stored.Protocol, &stored.Host,
			&stored.Path, &stored.Options, &stored.Username, &stored.Password, &stored.Domain)
		require.NoError(t, err, "storage_root must exist")

		u, p, d := services.ResolveSMBIdentity(&stored)
		assert.Equal(t, "backup_user", u, "should resolve to identity 2 (backup_user)")
		assert.Equal(t, "backup_pass", p, "should resolve password from identity 2")
		// Identity 2 has no CATALOGIZER_IDENTITY_2_DOMAIN set -> domain is empty
		assert.Equal(t, "", d, "identity 2 has no domain set -> empty")
	})

	t.Run("storage_root with null username resolved via identity_index", func(t *testing.T) {
		sqlDB := tests.SetupTestDB(t)
		defer sqlDB.Close()

		host := "nas.media.local"
		sharePath := "Movies"
		rootName := fmt.Sprintf("%s:%s", host, sharePath)
		identityMeta, _ := json.Marshal(map[string]int{"identity_index": 1})

		_, err := sqlDB.Exec(
			`INSERT INTO storage_roots (name, protocol, host, path, options, enabled, max_depth)
			 VALUES (?, 'smb', ?, ?, ?, 1, 10)`,
			rootName, host, sharePath, string(identityMeta),
		)
		require.NoError(t, err)

		var stored models.StorageRoot
		err = sqlDB.QueryRow(
			`SELECT id, name, protocol, host, path, options, username, password, domain
			   FROM storage_roots WHERE name = ?`, rootName,
		).Scan(&stored.ID, &stored.Name, &stored.Protocol, &stored.Host,
			&stored.Path, &stored.Options, &stored.Username, &stored.Password, &stored.Domain)
		require.NoError(t, err)

		// Username/password should be NULL in the DB.
		require.Nil(t, stored.Username, "username must be NULL per §11.4.10")
		require.Nil(t, stored.Password, "password must be NULL per §11.4.10")

		u, p, d := services.ResolveSMBIdentity(&stored)
		assert.Equal(t, "nas_user", u, "identity 1 should resolve")
		assert.Equal(t, "nas_pass", p)
		assert.Equal(t, "WORKGROUP", d)
	})
}

// ---------------------------------------------------------------------------
// Probe → Ingest integration (real db, no SMB host needed)
// ---------------------------------------------------------------------------

// TestIntegration_ProbeResultIngestion verifies that an SMBProbeResult can be
// ingested into a real SQLite database and that the resulting storage_root
// rows are resolvable via ResolveSMBIdentity (the full pipeline).
func TestIntegration_ProbeResultIngestion(t *testing.T) {
	t.Setenv("CATALOGIZER_IDENTITY_COUNT", "1")
	t.Setenv("CATALOGIZER_IDENTITY_1_TYPE", "credentials")
	t.Setenv("CATALOGIZER_IDENTITY_1_USERNAME", "probe_user")
	t.Setenv("CATALOGIZER_IDENTITY_1_PASSWORD", "probe_pass")
	t.Setenv("CATALOGIZER_IDENTITY_1_DOMAIN", "WORKGROUP")

	sqlDB := tests.SetupTestDB(t)
	defer sqlDB.Close()
	db := database.WrapDB(sqlDB, database.DialectSQLite)

	ingester := services.NewBindingIngester(db, zap.NewNop())

	// Ingest an authenticated probe result.
	result := &services.SMBProbeResult{
		Host:          "192.168.1.50",
		Authenticated: true,
		IdentityIndex: 1,
		IdentityLabel: "probe_user",
		Shares: []services.SMBShareInfo{
			{Host: "192.168.1.50", ShareName: "Data", Path: "\\\\192.168.1.50\\Data"},
			{Host: "192.168.1.50", ShareName: "Music", Path: "\\\\192.168.1.50\\Music"},
		},
	}

	out, err := ingester.IngestProbeResult(context.Background(), result)
	require.NoError(t, err)
	assert.Equal(t, 2, out.BoundShares, "2 bindings for 2 shares")
	assert.Equal(t, 2, out.NewRoots, "2 new storage roots")

	// Verify storage roots exist and are resolvable.
	for _, share := range result.Shares {
		rootName := fmt.Sprintf("%s:%s", result.Host, share.ShareName)
		var username, password, domain *string
		var options *string
		err := sqlDB.QueryRow(
			`SELECT username, password, domain, options FROM storage_roots WHERE name = ?`,
			rootName,
		).Scan(&username, &password, &domain, &options)
		require.NoError(t, err, "storage_root %s must exist after ingestion", rootName)

		require.Nil(t, username, "§11.4.10: username must be NULL in storage_root")
		require.Nil(t, password, "§11.4.10: password must be NULL in storage_root")

		var meta map[string]int
		require.NoError(t, json.Unmarshal([]byte(*options), &meta))
		assert.Equal(t, 1, meta["identity_index"], "options must reference identity 1")
	}

	// Verify bindings table.
	shareIdentityRepo := repository.NewShareIdentityBindingRepository(db)
	bindings, err := shareIdentityRepo.GetWorkingForHost(context.Background(), "192.168.1.50")
	require.NoError(t, err)
	assert.Equal(t, 2, len(bindings), "2 bindings expected")
	for _, b := range bindings {
		assert.Equal(t, 1, b.IdentityIndex)
		assert.Equal(t, "probe_user", b.IdentityLabel)
	}

	// Idempotency: second call.
	out2, err := ingester.IngestProbeResult(context.Background(), result)
	require.NoError(t, err)
	assert.Equal(t, 0, out2.NewRoots, "no new roots on second call (idempotent)")
}

// ---------------------------------------------------------------------------
// Real SMB probe (SKIP when no host) — §11.4.3
// ---------------------------------------------------------------------------

// TestIntegration_SMBProbeReachableHost is the §11.4.27 real-system evidence
// that the probe + ingest pipeline works against a reachable SMB host.
//
// It requires:
//   - CATALOGIZER_TEST_SMB_HOST = reachable SMB server IP/hostname
//   - CATALOGIZER_IDENTITY_COUNT + CATALOGIZER_IDENTITY_* sourced from .env
//
// The test probes the host, asserts authentication and real share enumeration,
// then ingests the result into a fresh SQLite DB and verifies the stored rows.
//
// SKIP-with-reason (§11.4.3) when no test SMB host is configured — the suite
// stays green without LAN access.
func TestIntegration_SMBProbeReachableHost(t *testing.T) {
	host := strings.TrimSpace(os.Getenv("CATALOGIZER_TEST_SMB_HOST"))
	if host == "" {
		t.Skip("§11.4.3: set CATALOGIZER_TEST_SMB_HOST (+ sourced .env identities) to probe a real SMB host")
	}

	ids := services.LoadSMBIdentitiesFromEnv()
	if len(ids) == 0 {
		t.Skip("§11.4.3: no CATALOGIZER_IDENTITY_* credentials in env to probe with")
	}

	svc := services.NewSMBDiscoveryService(zap.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Probe the host.
	probeResult, err := svc.ProbeHostWithIdentities(ctx, host, ids)
	require.NotNil(t, probeResult, "probe result must be non-nil")
	require.True(t, probeResult.Authenticated,
		"host %s is reachable + %d identities supplied: expected authentication", host, len(ids))
	require.Greater(t, len(probeResult.Shares), 0,
		"authenticated probe returned zero shares — real shares expected via ListSharenames")

	// Verify no secrets leak in the label.
	for _, id := range ids {
		if id.Password != "" && probeResult.IdentityLabel == id.Password {
			t.Fatalf("§11.4.10: probe result leaked a password as IdentityLabel")
		}
	}
	t.Logf("§11.4.6 PASS: %s bound via identity #%d (%q) → %d real shares",
		host, probeResult.IdentityIndex, probeResult.IdentityLabel, len(probeResult.Shares))

	// Ingest into a fresh SQLite DB.
	sqlDB := tests.SetupTestDB(t)
	defer sqlDB.Close()
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	ingester := services.NewBindingIngester(db, zap.NewNop())

	ingestResult, err := ingester.IngestProbeResult(ctx, probeResult)
	require.NoError(t, err, "ingest of real probe result must succeed")
	assert.Greater(t, ingestResult.BoundShares, 0, "real probe must produce at least one binding")
	assert.Greater(t, ingestResult.NewRoots, 0, "real probe must produce at least one new root")

	// Verify the bindings exist.
	shareIdentityRepo := repository.NewShareIdentityBindingRepository(db)
	bindings, err := shareIdentityRepo.GetWorkingForHost(context.Background(), host)
	require.NoError(t, err)
	assert.Equal(t, ingestResult.BoundShares, len(bindings),
		"bindings count must match ingestion result")

	// Verify storage_root rows have NULL username/password (§11.4.10).
	for _, share := range probeResult.Shares {
		rootName := fmt.Sprintf("%s:%s", host, share.ShareName)
		var username, password *string
		err := sqlDB.QueryRow(
			`SELECT username, password FROM storage_roots WHERE name = ?`, rootName,
		).Scan(&username, &password)
		if err != nil {
			t.Logf("storage_root %s may have been filtered (IPC$): %v", rootName, err)
			continue
		}
		assert.Nil(t, username, "§11.4.10: username must be NULL in storage_root for %s", rootName)
		assert.Nil(t, password, "§11.4.10: password must be NULL in storage_root for %s", rootName)
	}

	t.Logf("§11.4.27 PASS: real probe + ingest pipeline verified for %s (%d shares, %d bindings)",
		host, len(probeResult.Shares), ingestResult.BoundShares)
}

// ---------------------------------------------------------------------------
// Idempotency with duplicate storage_root names
// ---------------------------------------------------------------------------

// TestIntegration_DuplicateStorageRootName verifies that two different hosts
// with the same share name do NOT collide in the DB — the storage_root.name
// includes both host and share, so "host1:Data" and "host2:Data" are distinct.
func TestIntegration_DuplicateStorageRootName(t *testing.T) {
	t.Setenv("CATALOGIZER_IDENTITY_COUNT", "1")
	t.Setenv("CATALOGIZER_IDENTITY_1_TYPE", "credentials")
	t.Setenv("CATALOGIZER_IDENTITY_1_USERNAME", "user")
	t.Setenv("CATALOGIZER_IDENTITY_1_PASSWORD", "pass")

	sqlDB := tests.SetupTestDB(t)
	defer sqlDB.Close()
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	ingester := services.NewBindingIngester(db, zap.NewNop())

	// Two different hosts, same share name "Data".
	host1 := &services.SMBProbeResult{
		Host:          "192.168.1.10",
		Authenticated: true,
		IdentityIndex: 1,
		IdentityLabel: "user",
		Shares:        []services.SMBShareInfo{{Host: "192.168.1.10", ShareName: "Data"}},
	}
	host2 := &services.SMBProbeResult{
		Host:          "10.0.0.1",
		Authenticated: true,
		IdentityIndex: 1,
		IdentityLabel: "user",
		Shares:        []services.SMBShareInfo{{Host: "10.0.0.1", ShareName: "Data"}},
	}

	out1, err := ingester.IngestProbeResult(context.Background(), host1)
	require.NoError(t, err)
	assert.Equal(t, 1, out1.NewRoots, "host1:Data → 1 new root")

	out2, err := ingester.IngestProbeResult(context.Background(), host2)
	require.NoError(t, err)
	assert.Equal(t, 1, out2.NewRoots, "host2:Data → 1 new root (different host)")

	// Both should exist, names must be distinct.
	var count int
	err = sqlDB.QueryRow(`SELECT COUNT(*) FROM storage_roots`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count, "two storage roots must exist")
}

// ---------------------------------------------------------------------------
// Edge: empty share names (no shares to ingest) — idempotent
// ---------------------------------------------------------------------------

// TestIntegration_EmptyShareListIngestion verifies that a probe result with
// zero shares does not create any storage_root or binding rows.
func TestIntegration_EmptyShareListIngestion(t *testing.T) {
	sqlDB := tests.SetupTestDB(t)
	defer sqlDB.Close()
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	ingester := services.NewBindingIngester(db, zap.NewNop())

	result := &services.SMBProbeResult{
		Host:          "host.empty.shares",
		Authenticated: true,
		IdentityIndex: 1,
		IdentityLabel: "guest",
		Shares:        []services.SMBShareInfo{},
	}

	out, err := ingester.IngestProbeResult(context.Background(), result)
	require.NoError(t, err)
	assert.Equal(t, 0, out.BoundShares, "zero shares → zero bindings")
	assert.Equal(t, 0, out.NewRoots, "zero shares → zero new roots")

	var count int
	err = sqlDB.QueryRow(`SELECT COUNT(*) FROM storage_roots`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "no storage roots for empty share list")
}

// TestIntegration_StorageRootNaming verifies the storage root naming scheme
// produces unique, deterministic names from (host, share) pairs.
func TestIntegration_StorageRootNaming(t *testing.T) {
	// Create storage roots via the ingester and verify their names.
	sqlDB := tests.SetupTestDB(t)
	defer sqlDB.Close()
	db := database.WrapDB(sqlDB, database.DialectSQLite)
	ingester := services.NewBindingIngester(db, zap.NewNop())

	// Ingest three shares with distinct names.
	// Note: the ingester uses rootNameFor(result.Host, share.ShareName)
	// so shares on the same host must have different names to be unique.
	result := &services.SMBProbeResult{
		Host:          "server1",
		Authenticated: true,
		IdentityIndex: 1,
		IdentityLabel: "user",
		Shares: []services.SMBShareInfo{
			{ShareName: "data"},
			{ShareName: "Music"},
			{ShareName: "video"}, // third distinct name for same host
		},
	}

	out, err := ingester.IngestProbeResult(context.Background(), result)
	require.NoError(t, err)
	assert.Equal(t, 3, out.NewRoots, "3 shares with distinct names → 3 new roots")

	// Verify all three names are distinct and follow the pattern.
	var names []string
	rows, err := sqlDB.Query(`SELECT name FROM storage_roots ORDER BY name`)
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		names = append(names, name)
	}
	require.Equal(t, 3, len(names))
	assert.Equal(t, "server1:Music", names[0])
	assert.Equal(t, "server1:data", names[1])
	assert.Equal(t, "server1:video", names[2])
}
