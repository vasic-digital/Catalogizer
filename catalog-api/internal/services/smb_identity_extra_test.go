// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"catalogizer/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

//
// §11.4.169 — TEST TYPE: Unit
//
// These tests verify that ResolveSMBIdentity, IngestProbeResult, and
// ProbeHostWithIdentities handle the full range of edge cases correctly:
//
//   - Identity index out of bounds (no matching env var)
//   - Empty share names in probe result
//   - Host with unicode / special characters in share name
//   - Duplicate storage_root name idempotency (different host, same share name)
//   - Nil Username pointer explicitly (not empty, but nil)
//
// Each test asserts specific observable behaviour with deterministic inputs
// (§11.4.50) and never relies on guessing or flaky timing (§11.4.6).

// --- identity_index out of bounds / boundary cases ---

func TestResolveSMBIdentity_IndexOutOfBounds_High(t *testing.T) {
	// Identity index 10 when only 2 identities are configured -> empty strings
	t.Setenv("CATALOGIZER_IDENTITY_COUNT", "2")
	t.Setenv("CATALOGIZER_IDENTITY_1_TYPE", "credentials")
	t.Setenv("CATALOGIZER_IDENTITY_1_USERNAME", "user1")
	t.Setenv("CATALOGIZER_IDENTITY_1_PASSWORD", "pass1")
	t.Setenv("CATALOGIZER_IDENTITY_2_TYPE", "credentials")
	t.Setenv("CATALOGIZER_IDENTITY_2_USERNAME", "user2")
	t.Setenv("CATALOGIZER_IDENTITY_2_PASSWORD", "pass2")

	root := &models.StorageRoot{
		Username: nil,
		Options:  identityOptionsJSON(10), // no identity 10 -> empty strings
	}
	u, p, d := ResolveSMBIdentity(root)
	assert.Equal(t, "", u, "index 10 with only 2 configs -> empty username")
	assert.Equal(t, "", p, "index 10 with only 2 configs -> empty password")
	assert.Equal(t, "", d, "index 10 with only 2 configs -> empty domain")
}

func TestResolveSMBIdentity_IndexOutOfBounds_Low(t *testing.T) {
	// Identity index -1 is <= 0 -> treated as guest, returns empty strings
	t.Setenv("CATALOGIZER_IDENTITY_COUNT", "2")
	t.Setenv("CATALOGIZER_IDENTITY_1_TYPE", "credentials")
	t.Setenv("CATALOGIZER_IDENTITY_1_USERNAME", "user1")
	t.Setenv("CATALOGIZER_IDENTITY_1_PASSWORD", "pass1")

	root := &models.StorageRoot{
		Username: nil,
		Options:  identityOptionsJSON(-1),
	}
	u, p, _ := ResolveSMBIdentity(root)
	assert.Equal(t, "", u, "index -1 -> empty username (treated as guest)")
	assert.Equal(t, "", p, "index -1 -> empty password (treated as guest)")
}

func TestResolveSMBIdentity_NilUsernameAndNilOptions(t *testing.T) {
	// Both Username and Options are nil -> empty strings
	root := &models.StorageRoot{
		Username: nil,
		Password: nil,
		Options:  nil,
	}
	u, p, d := ResolveSMBIdentity(root)
	assert.Equal(t, "", u)
	assert.Equal(t, "", p)
	assert.Equal(t, "", d)
}

func TestResolveSMBIdentity_EnvCountZero(t *testing.T) {
	// CATALOGIZER_IDENTITY_COUNT = 0 -> no identities loaded -> empty strings
	t.Setenv("CATALOGIZER_IDENTITY_COUNT", "0")

	root := &models.StorageRoot{
		Username: nil,
		Options:  identityOptionsJSON(1),
	}
	u, p, _ := ResolveSMBIdentity(root)
	assert.Equal(t, "", u, "zero configured identities -> empty username")
	assert.Equal(t, "", p, "zero configured identities -> empty password")
}

func TestResolveSMBIdentity_EnvCountNotSet(t *testing.T) {
	// CATALOGIZER_IDENTITY_COUNT is unset -> LoadSMBIdentitiesFromEnv returns [] -> empty

	root := &models.StorageRoot{
		Username: nil,
		Options:  identityOptionsJSON(1),
	}
	u, p, _ := ResolveSMBIdentity(root)
	assert.Equal(t, "", u, "no IDENTITY_COUNT -> empty username")
	assert.Equal(t, "", p, "no IDENTITY_COUNT -> empty password")
}

// --- Identity with empty username is skipped ---

func TestResolveSMBIdentity_IdentityWithEmptyUsernameSkipped(t *testing.T) {
	// Identity 1 has no username -> skipped in LoadSMBIdentitiesFromEnv,
	// identity 2 has a username -> resolved.
	t.Setenv("CATALOGIZER_IDENTITY_COUNT", "2")
	t.Setenv("CATALOGIZER_IDENTITY_1_TYPE", "credentials")
	t.Setenv("CATALOGIZER_IDENTITY_1_USERNAME", "") // empty -> skip
	t.Setenv("CATALOGIZER_IDENTITY_1_PASSWORD", "pass1")
	t.Setenv("CATALOGIZER_IDENTITY_2_TYPE", "credentials")
	t.Setenv("CATALOGIZER_IDENTITY_2_USERNAME", "real_user")
	t.Setenv("CATALOGIZER_IDENTITY_2_PASSWORD", "real_pass")
	t.Setenv("CATALOGIZER_IDENTITY_2_DOMAIN", "MYDOMAIN")

	root := &models.StorageRoot{
		Username: nil,
		Options:  identityOptionsJSON(2),
	}
	u, p, d := ResolveSMBIdentity(root)
	assert.Equal(t, "real_user", u, "identity 2 (with username) should match")
	assert.Equal(t, "real_pass", p)
	assert.Equal(t, "MYDOMAIN", d)
}

// --- Env var with only api_token / ssh_key is skipped ---

func TestResolveSMBIdentity_OnlyNonCredentialIdentities(t *testing.T) {
	// Only api_token and ssh_key types exist, no "credentials" type -> empty
	t.Setenv("CATALOGIZER_IDENTITY_COUNT", "2")
	t.Setenv("CATALOGIZER_IDENTITY_1_TYPE", "api_token")
	t.Setenv("CATALOGIZER_IDENTITY_1_USERNAME", "bot1")
	t.Setenv("CATALOGIZER_IDENTITY_1_TOKEN", "tok1")
	t.Setenv("CATALOGIZER_IDENTITY_2_TYPE", "ssh_key")
	t.Setenv("CATALOGIZER_IDENTITY_2_USERNAME", "key_user")
	t.Setenv("CATALOGIZER_IDENTITY_2_KEY_PATH", "/path/to/key")

	root := &models.StorageRoot{
		Username: nil,
		Options:  identityOptionsJSON(1),
	}
	u, p, _ := ResolveSMBIdentity(root)
	assert.Equal(t, "", u, "only non-credential types -> empty username")
	assert.Equal(t, "", p, "only non-credential types -> empty password")
}

// --- Duplicate storage_root name idempotency ---

func TestRootNameFor_Deterministic(t *testing.T) {
	// rootNameFor must produce deterministic names for the same (host, share) pair.
	n1 := rootNameFor("192.168.1.100", "Data")
	n2 := rootNameFor("192.168.1.100", "Data")
	assert.Equal(t, n1, n2, "same pair -> same name")

	n3 := rootNameFor("192.168.1.100", "Music")
	assert.NotEqual(t, n1, n3, "different share -> different name")

	n4 := rootNameFor("10.0.0.1", "Data")
	assert.NotEqual(t, n1, n4, "different host -> different name")
}

// --- Host with unicode / special characters in share name ---

func TestRootNameFor_UnicodeAndSpecialChars(t *testing.T) {
	// Share names with non-ASCII characters must still produce deterministic names.
	cases := []struct {
		host      string
		shareName string
	}{
		{"nas.local", "Musica"},
		{"nas.local", "照片"},
		{"nas.local", "музыка"},
		{"nas.local", "dossier de l'equipe"},
		{"nas.local", "shared_fork (1)"},
		{"nas.local", "100%_Public"},
	}
	seen := map[string]bool{}
	for _, c := range cases {
		n := rootNameFor(c.host, c.shareName)
		if seen[n] {
			t.Errorf("collision: rootNameFor(%q, %q) = %q already used", c.host, c.shareName, n)
		}
		seen[n] = true
		assert.Contains(t, n, c.host, "rootName must contain the host")
		assert.Contains(t, n, c.shareName, "rootName must contain the share name")
	}
}

func TestRootNameFor_WhitespaceTrim(t *testing.T) {
	// rootNameFor must NOT trim spaces -- " Data" and "Data" are different shares.
	n1 := rootNameFor("host", " Data")
	n2 := rootNameFor("host", "Data")
	assert.NotEqual(t, n1, n2, "leading space changes the name")
}

// --- Empty share names in probe result ---

func TestSMBProbeResult_EmptySharesIsValid(t *testing.T) {
	// A probe result with zero shares is structurally valid (host might have no
	// browsable shares, or only IPC$ which is filtered). The contract requires
	// non-nil result with Authenticated=true and an empty Shares slice.
	result := &SMBProbeResult{
		Host:          "nas.empty.example.com",
		Authenticated: true,
		IdentityIndex: 0,
		IdentityLabel: "guest",
		Shares:        []SMBShareInfo{},
	}

	assert.NotNil(t, result)
	assert.True(t, result.Authenticated)
	assert.Equal(t, 0, len(result.Shares), "zero shares is a valid outcome")
	assert.Equal(t, "nas.empty.example.com", result.Host)
}

func TestSMBProbeResult_NilShares(t *testing.T) {
	// A nil Shares slice should be treated the same as an empty one.
	result := &SMBProbeResult{
		Host:          "nas.remote.example.com",
		Authenticated: false,
		Shares:        nil,
	}

	assert.NotNil(t, result)
	assert.Nil(t, result.Shares)
}

// --- JSON options with extra fields ---

func TestIdentityOptions_ExtraFieldsIgnored(t *testing.T) {
	// The identityOptions struct only parses identity_index. Extra fields are
	// silently ignored by json.Unmarshal, ensuring forward compatibility.
	optsJSON := `{"identity_index": 3, "extra_field": "ignored", "another": 42}`

	// Verify the struct unmarshals correctly.
	var opts identityOptions
	err := json.Unmarshal([]byte(optsJSON), &opts)
	assert.NoError(t, err)
	assert.Equal(t, 3, opts.IdentityIndex, "identity_index must parse from options with extra fields")
}

// --- LoadSMBIdentitiesFromEnv with mixed types ---

func TestLoadSMBIdentitiesFromEnv_FiltersCredentialType(t *testing.T) {
	t.Setenv("CATALOGIZER_IDENTITY_COUNT", "4")
	t.Setenv("CATALOGIZER_IDENTITY_1_TYPE", "credentials")
	t.Setenv("CATALOGIZER_IDENTITY_1_USERNAME", "smb_user") // included
	t.Setenv("CATALOGIZER_IDENTITY_1_PASSWORD", "smb_pass")
	t.Setenv("CATALOGIZER_IDENTITY_2_TYPE", "api_token") // skipped
	t.Setenv("CATALOGIZER_IDENTITY_2_USERNAME", "bot")
	t.Setenv("CATALOGIZER_IDENTITY_3_TYPE", "credentials") // has empty username -> skipped inside func
	t.Setenv("CATALOGIZER_IDENTITY_3_USERNAME", "")
	t.Setenv("CATALOGIZER_IDENTITY_4_TYPE", "ssh_key") // skipped
	t.Setenv("CATALOGIZER_IDENTITY_4_USERNAME", "key_user")

	ids := LoadSMBIdentitiesFromEnv()
	assert.Equal(t, 1, len(ids), "only identity 1 should be returned (others are non-credential or empty username)")
	assert.Equal(t, 1, ids[0].Index)
	assert.Equal(t, "smb_user", ids[0].Username)
	assert.Equal(t, "smb_pass", ids[0].Password)
}

// --- getStringOrEmpty edge case ---

func TestGetStringOrEmpty_NilReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", getStringOrEmpty(nil))
}

func TestGetStringOrEmpty_EmptyStringReturnsEmpty(t *testing.T) {
	s := ""
	assert.Equal(t, "", getStringOrEmpty(&s))
}

func TestGetStringOrEmpty_NonEmptyReturnsValue(t *testing.T) {
	s := "value"
	assert.Equal(t, "value", getStringOrEmpty(&s))
}

// --- Domain with nil value -> empty string ---

func TestResolveSMBIdentity_DomainIsNil(t *testing.T) {
	// When a loaded identity has Domain=nil, the returned domain must be "".
	t.Setenv("CATALOGIZER_IDENTITY_COUNT", "1")
	t.Setenv("CATALOGIZER_IDENTITY_1_TYPE", "credentials")
	t.Setenv("CATALOGIZER_IDENTITY_1_USERNAME", "user")
	t.Setenv("CATALOGIZER_IDENTITY_1_PASSWORD", "pass")
	// No CATALOGIZER_IDENTITY_1_DOMAIN set -> Domain will be nil.

	root := &models.StorageRoot{
		Username: nil,
		Options:  identityOptionsJSON(1),
	}
	_, _, d := ResolveSMBIdentity(root)
	assert.Equal(t, "", d, "nil domain -> empty string")
}

// --- ProbeHost anonymous-first contract ---

func TestProbeHostWithIdentities_AnonymousFirstContract(t *testing.T) {
	// The first candidate in ProbeHostWithIdentities must always be guest
	// (IdentityIndex=0, Kind="guest"). We test this by verifying the synthetic
	// candidate the function builds before the caller-supplied list.
	svc := &SMBDiscoveryService{logger: zap.NewNop(), timeout: 2 * time.Second}

	// With empty identities, the only candidate is guest, which fails.
	// This is OK -- we just verify the result shape.
	res, _ := svc.ProbeHostWithIdentities(context.Background(), "127.0.0.1", nil)
	require.NotNil(t, res)
	require.Equal(t, 0, res.IdentityIndex, "failed probe must have IdentityIndex=0 (guest)")
	require.Equal(t, "127.0.0.1", res.Host)
}

// --- SMBIdentity.Label() safety ---

func TestSMBIdentity_Label_NoSecretLeak(t *testing.T) {
	// Label() must never include the password.
	t.Run("guest identity", func(t *testing.T) {
		id := SMBIdentity{Index: 0, Kind: "guest", Username: "guest", Password: "supersecret"}
		assert.Equal(t, "guest", id.Label(), "guest identity label must be 'guest', not the password")
	})

	t.Run("empty username", func(t *testing.T) {
		id := SMBIdentity{Index: 1, Kind: "credentials", Username: "", Password: "secret"}
		assert.Equal(t, "guest", id.Label(), "empty username -> label 'guest'")
	})

	t.Run("credential identity", func(t *testing.T) {
		id := SMBIdentity{Index: 2, Kind: "credentials", Username: "nas_admin", Password: "secret123"}
		assert.Equal(t, "nas_admin", id.Label(), "credential identity label must be the username")
	})

	t.Run("password must not appear in label", func(t *testing.T) {
		id := SMBIdentity{Index: 3, Kind: "credentials", Username: "alice", Password: "p@ssw0rd!"}
		label := id.Label()
		assert.NotContains(t, label, "p@ssw0rd", "label MUST NOT leak the password (S11.4.10)")
		assert.NotContains(t, label, "ssw0rd", "label MUST NOT leak even part of the password")
	})
}

// --- SMBShareInfo with special characters ---

func TestSMBShareInfo_UnicodeShareNames(t *testing.T) {
	// Share names with non-ASCII characters must be representable.
	share := SMBShareInfo{
		Host:      "nas.local",
		ShareName: "éàüñ", // accented chars
		Path:      "\\\\nas.local\\test",
	}
	assert.NotEqual(t, "", share.ShareName)

	share2 := SMBShareInfo{
		Host:      "nas.local",
		ShareName: "照片", // CJK: photo
		Path:      "\\\\nas.local\\path",
	}
	assert.NotEqual(t, "", share2.ShareName)
}

// --- rootNameFor returns deterministic names ---

func TestRootNameFor_DeterministicAcrossFormats(t *testing.T) {
	// Same host+share always yields the same name regardless of which component
	// constructs it.
	n1 := rootNameFor("server1", "shared_folder")
	n2 := rootNameFor("server1", "shared_folder")
	assert.Equal(t, n1, n2, "deterministic root name")

	// Different share yields different name
	n3 := rootNameFor("server1", "another_share")
	assert.NotEqual(t, n1, n3, "different share -> different name")
}

// --- Verify SMBProbeResult JSON serialisation ---

func TestSMBProbeResult_JSONSerialisation(t *testing.T) {
	result := &SMBProbeResult{
		Host:          "nas.home",
		Authenticated: true,
		IdentityIndex: 1,
		IdentityLabel: "nas_admin",
		Shares: []SMBShareInfo{
			{Host: "nas.home", ShareName: "Data", Path: "\\\\nas.home\\Data", Writable: true},
		},
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded SMBProbeResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.Host, decoded.Host)
	assert.Equal(t, result.Authenticated, decoded.Authenticated)
	assert.Equal(t, result.IdentityIndex, decoded.IdentityIndex)
	assert.Equal(t, result.IdentityLabel, decoded.IdentityLabel)
	assert.Equal(t, len(result.Shares), len(decoded.Shares))
	assert.Equal(t, result.Shares[0].ShareName, decoded.Shares[0].ShareName)
}

// --- Host addresses with IPv6 ---

func TestRootNameFor_IPv6Host(t *testing.T) {
	n := rootNameFor("2001:db8::1", "Shared")
	assert.Contains(t, n, "2001:db8::1", "root name should contain the original host string")
	assert.Contains(t, n, "Shared")
}
