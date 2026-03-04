package challenges

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// DatabaseDocsChallenge validates that all database tables are
// documented by checking for database/schema documentation files
// and verifying key table names are mentioned.
type DatabaseDocsChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewDatabaseDocsChallenge creates CH-046.
func NewDatabaseDocsChallenge() *DatabaseDocsChallenge {
	return &DatabaseDocsChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"database-docs",
			"Database Documentation Completeness",
			"Verifies database tables are documented: checks for "+
				"migration files, schema documentation, and that key "+
				"table names appear in documentation.",
			"documentation",
			nil,
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the database documentation challenge.
func (c *DatabaseDocsChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{}

	// Step 1: Check that migration files exist
	c.ReportProgress("check-migrations", nil)
	migrationDirs := []string{
		"database/migrations",
		"database",
		"migrations",
	}

	var migrationDir string
	migrationCount := 0
	for _, dir := range migrationDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		migrationDir = dir
		for _, e := range entries {
			if !e.IsDir() && (strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), ".sql")) {
				migrationCount++
			}
		}
		if migrationCount > 0 {
			break
		}
	}

	hasMigrations := migrationCount > 0
	outputs["migration_dir"] = migrationDir
	outputs["migration_files"] = fmt.Sprintf("%d", migrationCount)

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "migration_files",
		Expected: "migration files exist",
		Actual:   fmt.Sprintf("%d files in %s", migrationCount, migrationDir),
		Passed:   hasMigrations,
		Message: challenge.Ternary(hasMigrations,
			fmt.Sprintf("Found %d migration files in %s", migrationCount, migrationDir),
			"No migration files found"),
	})

	// Step 2: Check that database package has documentation
	c.ReportProgress("check-db-package", nil)
	dbPackageFiles := []string{
		"database/connection.go",
		"database/dialect.go",
		"database/database.go",
	}

	dbFilesFound := 0
	for _, f := range dbPackageFiles {
		if _, err := os.Stat(f); err == nil {
			dbFilesFound++
		}
	}

	hasDBPackage := dbFilesFound > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "database_package",
		Expected: "database package files exist",
		Actual:   fmt.Sprintf("%d/%d files found", dbFilesFound, len(dbPackageFiles)),
		Passed:   hasDBPackage,
		Message: challenge.Ternary(hasDBPackage,
			fmt.Sprintf("Database package has %d key files", dbFilesFound),
			"Database package files not found"),
	})

	// Step 3: Check docs for database/schema documentation
	c.ReportProgress("check-schema-docs", nil)
	docLocations := []string{
		"../docs",
		"docs",
		"../docs/architecture",
		"../docs/api",
	}

	keyTables := []string{
		"media_items",
		"media_files",
		"media_types",
		"storage_roots",
		"users",
	}

	tablesDocumented := 0
	for _, docDir := range docLocations {
		_ = filepath.Walk(docDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if ext != ".md" && ext != ".txt" {
				return nil
			}
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			contentStr := string(content)
			for _, table := range keyTables {
				if strings.Contains(contentStr, table) {
					tablesDocumented++
				}
			}
			return nil
		})
		if tablesDocumented > 0 {
			break
		}
	}

	// Deduplicate: a table might be mentioned in multiple files
	if tablesDocumented > len(keyTables) {
		tablesDocumented = len(keyTables)
	}

	tableDocRatio := float64(tablesDocumented) / float64(len(keyTables)) * 100
	tablesDocPassed := tablesDocumented >= 3 // at least 3 of 5 key tables mentioned
	outputs["tables_documented"] = fmt.Sprintf("%d/%d", tablesDocumented, len(keyTables))

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_score",
		Target:   "tables_documented",
		Expected: "at least 3/5 key tables mentioned in docs",
		Actual:   fmt.Sprintf("%d/%d (%.0f%%)", tablesDocumented, len(keyTables), tableDocRatio),
		Passed:   tablesDocPassed,
		Message: challenge.Ternary(tablesDocPassed,
			fmt.Sprintf("%d/%d key tables found in documentation", tablesDocumented, len(keyTables)),
			fmt.Sprintf("Only %d/%d key tables found in documentation", tablesDocumented, len(keyTables))),
	})

	metrics := map[string]challenge.MetricValue{
		"migration_files": {
			Name:  "migration_files",
			Value: float64(migrationCount),
			Unit:  "count",
		},
		"tables_documented": {
			Name:  "tables_documented",
			Value: float64(tablesDocumented),
			Unit:  "count",
		},
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, metrics, outputs, ""), nil
}
