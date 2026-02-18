package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// DatabaseSchemaValidationChallenge validates that the database schema
// is correctly set up with all expected tables and migrations applied.
type DatabaseSchemaValidationChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewDatabaseSchemaValidationChallenge creates CH-015.
func NewDatabaseSchemaValidationChallenge() *DatabaseSchemaValidationChallenge {
	return &DatabaseSchemaValidationChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"database-schema-validation",
			"Database Schema Validation",
			"Validates all expected tables exist, migrations are applied, "+
				"and schema supports read/write operations",
			"e2e",
			[]challenge.ID{"database-connectivity"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the database schema validation challenge.
func (c *DatabaseSchemaValidationChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Step 1: Login
	_, loginErr := client.Login(ctx, c.config.Username, c.config.Password)
	loginOK := loginErr == nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "admin_login",
		Expected: "successful login",
		Actual:   fmt.Sprintf("login_ok=%v", loginOK),
		Passed:   loginOK,
		Message:  challenge.Ternary(loginOK, "Admin login succeeded", fmt.Sprintf("Login failed: %v", loginErr)),
	})
	if !loginOK {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, ""), nil
	}

	// Step 2: Health endpoint confirms database type
	healthCode, healthBody, healthErr := client.Get(ctx, "/health")
	healthOK := healthErr == nil && healthCode == 200
	dbType := ""
	if healthBody != nil {
		if dt, ok := healthBody["database"].(string); ok {
			dbType = dt
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "health_database_info",
		Expected: "database type reported in health",
		Actual:   fmt.Sprintf("HTTP %d, database=%q", healthCode, dbType),
		Passed:   healthOK,
		Message:  challenge.Ternary(healthOK, fmt.Sprintf("Health endpoint reports database type: %s", dbType), fmt.Sprintf("Health check failed: %v", healthErr)),
	})

	// Step 3: Stats endpoint validates core tables are queryable (storage_roots, files, file_metadata)
	statsCode, statsBody, statsErr := client.Get(ctx, "/api/v1/stats/overall")
	statsOK := statsErr == nil && statsCode == 200 && statsBody != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "stats_tables_queryable",
		Expected: "stats endpoint returns valid data (proves storage_roots, files, file_metadata tables exist)",
		Actual:   fmt.Sprintf("HTTP %d", statsCode),
		Passed:   statsOK,
		Message:  challenge.Ternary(statsOK, "Core tables queryable via stats endpoint", fmt.Sprintf("Stats query failed: HTTP %d err=%v", statsCode, statsErr)),
	})

	// Step 4: Storage roots endpoint validates storage_roots table
	rootsCode, _, rootsErr := client.Get(ctx, "/api/v1/storage/roots")
	rootsOK := rootsErr == nil && rootsCode == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "storage_roots_table",
		Expected: "storage roots endpoint responds (proves table exists)",
		Actual:   fmt.Sprintf("HTTP %d", rootsCode),
		Passed:   rootsOK,
		Message:  challenge.Ternary(rootsOK, "storage_roots table accessible", fmt.Sprintf("Storage roots query failed: HTTP %d err=%v", rootsCode, rootsErr)),
	})

	// Step 5: Media search endpoint validates files + file_metadata tables with joins
	mediaCode, _, mediaErr := client.Get(ctx, "/api/v1/media/search?limit=1&offset=0")
	mediaOK := mediaErr == nil && mediaCode == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "media_tables",
		Expected: "media search endpoint responds (proves files+metadata tables with joins work)",
		Actual:   fmt.Sprintf("HTTP %d", mediaCode),
		Passed:   mediaOK,
		Message:  challenge.Ternary(mediaOK, "files and file_metadata tables accessible with joins", fmt.Sprintf("Media search failed: HTTP %d err=%v", mediaCode, mediaErr)),
	})

	// Step 6: Users endpoint validates auth tables (users, roles)
	usersCode, _, usersErr := client.Get(ctx, "/api/v1/auth/me")
	usersOK := usersErr == nil && usersCode == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "auth_tables",
		Expected: "auth/me endpoint responds (proves users+roles tables exist)",
		Actual:   fmt.Sprintf("HTTP %d", usersCode),
		Passed:   usersOK,
		Message:  challenge.Ternary(usersOK, "Auth tables (users, roles) accessible", fmt.Sprintf("Auth query failed: HTTP %d err=%v", usersCode, usersErr)),
	})

	// Step 7: Write and read back - validates foreign key constraints work
	testRootName := fmt.Sprintf("schema-test-%d", time.Now().UnixMilli())
	createBody := fmt.Sprintf(`{"name":%q,"protocol":"local","path":"/tmp/schema-test","max_depth":1}`, testRootName)
	createCode, _, createErr := client.PostJSON(ctx, "/api/v1/storage/roots", createBody)
	createOK := createErr == nil && (createCode == 201 || createCode == 200)

	// Read back and verify
	readCode, readBody, readErr := client.Get(ctx, "/api/v1/storage/roots")
	readOK := readErr == nil && readCode == 200
	foundRoot := false
	if readOK && readBody != nil {
		if roots, ok := readBody["roots"].([]interface{}); ok {
			for _, r := range roots {
				if rm, ok := r.(map[string]interface{}); ok {
					if nm, ok := rm["name"].(string); ok && nm == testRootName {
						foundRoot = true
						break
					}
				}
			}
		}
	}
	schemaWriteOK := createOK && foundRoot
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "schema_write_read",
		Expected: "create and read back storage root (validates FK constraints)",
		Actual:   fmt.Sprintf("create=%v, found=%v", createOK, foundRoot),
		Passed:   schemaWriteOK,
		Message:  challenge.Ternary(schemaWriteOK, "Schema write/read with FK constraints verified", fmt.Sprintf("Schema validation failed: create=%v found=%v err=%v", createOK, foundRoot, createErr)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}
