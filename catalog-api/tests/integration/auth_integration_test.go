package integration

import (
	"fmt"
	"testing"
	"time"

	"catalogizer/internal/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// INTEGRATION TEST: Auth flow with in-memory SQLite
// =============================================================================

func TestAuthIntegration_UserCreationAndLookup(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer db.Close()

	t.Run("InsertUserAndQueryBack", func(t *testing.T) {
		username := fmt.Sprintf("authtest_%d", time.Now().UnixNano())
		email := fmt.Sprintf("%s@example.com", username)
		passwordHash := "hashed_password_placeholder"

		result, err := db.Exec(
			`INSERT INTO users (username, email, password_hash, is_active, role_id)
			 VALUES (?, ?, ?, 1, 1)`,
			username, email, passwordHash,
		)
		require.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		require.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)

		// Verify user exists
		var storedUsername, storedEmail string
		var isActive bool
		err = db.QueryRow(
			`SELECT username, email, is_active FROM users WHERE username = ?`,
			username,
		).Scan(&storedUsername, &storedEmail, &isActive)
		require.NoError(t, err)

		assert.Equal(t, username, storedUsername)
		assert.Equal(t, email, storedEmail)
		assert.True(t, isActive)
	})

	t.Run("DuplicateUsernameRejected", func(t *testing.T) {
		username := fmt.Sprintf("dupuser_%d", time.Now().UnixNano())
		email1 := fmt.Sprintf("%s_1@example.com", username)
		email2 := fmt.Sprintf("%s_2@example.com", username)

		_, err := db.Exec(
			`INSERT INTO users (username, email, password_hash, is_active, role_id)
			 VALUES (?, ?, 'hash1', 1, 1)`,
			username, email1,
		)
		require.NoError(t, err)

		// Second insert with same username should fail (UNIQUE constraint)
		_, err = db.Exec(
			`INSERT INTO users (username, email, password_hash, is_active, role_id)
			 VALUES (?, ?, 'hash2', 1, 1)`,
			username, email2,
		)
		assert.Error(t, err, "Duplicate username should be rejected")
		assert.Contains(t, err.Error(), "UNIQUE",
			"Error should be a unique constraint violation")
	})

	t.Run("DuplicateEmailRejected", func(t *testing.T) {
		email := fmt.Sprintf("dupemail_%d@example.com", time.Now().UnixNano())
		user1 := fmt.Sprintf("user1_%d", time.Now().UnixNano())
		user2 := fmt.Sprintf("user2_%d", time.Now().UnixNano())

		_, err := db.Exec(
			`INSERT INTO users (username, email, password_hash, is_active, role_id)
			 VALUES (?, ?, 'hash1', 1, 1)`,
			user1, email,
		)
		require.NoError(t, err)

		_, err = db.Exec(
			`INSERT INTO users (username, email, password_hash, is_active, role_id)
			 VALUES (?, ?, 'hash2', 1, 1)`,
			user2, email,
		)
		assert.Error(t, err, "Duplicate email should be rejected")
	})
}

// =============================================================================
// Auth integration test for failed_login_attempts tracking
// =============================================================================

func TestAuthIntegration_FailedLoginAttemptsTracking(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer db.Close()

	username := fmt.Sprintf("locktest_%d", time.Now().UnixNano())
	_, err := db.Exec(
		`INSERT INTO users (username, email, password_hash, is_active,
		 role_id, failed_login_attempts)
		 VALUES (?, ?, 'bcrypt_hash', 1, 1, 0)`,
		username, fmt.Sprintf("%s@example.com", username),
	)
	require.NoError(t, err)

	// Simulate incrementing failed_login_attempts
	for i := 1; i <= 5; i++ {
		_, err := db.Exec(
			`UPDATE users SET failed_login_attempts = ? WHERE username = ?`,
			i, username,
		)
		require.NoError(t, err)
	}

	var attempts int
	err = db.QueryRow(
		`SELECT failed_login_attempts FROM users WHERE username = ?`,
		username,
	).Scan(&attempts)
	require.NoError(t, err)
	assert.Equal(t, 5, attempts,
		"Failed login attempts should be tracked")

	// Simulate lock
	_, err = db.Exec(
		`UPDATE users SET is_locked = 1,
		 locked_until = datetime('now', '+30 minutes')
		 WHERE username = ?`,
		username,
	)
	require.NoError(t, err)

	var isLocked bool
	err = db.QueryRow(
		`SELECT is_locked FROM users WHERE username = ?`,
		username,
	).Scan(&isLocked)
	require.NoError(t, err)
	assert.True(t, isLocked,
		"Account should be locked after max failed attempts")

	// Simulate successful login resetting attempts
	_, err = db.Exec(
		`UPDATE users SET failed_login_attempts = 0, is_locked = 0,
		 locked_until = NULL WHERE username = ?`,
		username,
	)
	require.NoError(t, err)

	err = db.QueryRow(
		`SELECT failed_login_attempts FROM users WHERE username = ?`,
		username,
	).Scan(&attempts)
	require.NoError(t, err)
	assert.Equal(t, 0, attempts,
		"Failed attempts should reset on successful login")
}
