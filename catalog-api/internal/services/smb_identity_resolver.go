package services

import (
	"encoding/json"

	"catalogizer/models"
)

// identityOptions is the JSON structure optionally stored in
// storage_roots.options (e.g. {"identity_index": 1}) by the probe-and-ingest
// pipeline. When Username is nil, the scanner reads identity_index from this
// struct and resolves it to the env-var credential set.
type identityOptions struct {
	IdentityIndex int `json:"identity_index"`
}

// ResolveSMBIdentity returns (username, password, domain) for an SMB storage root.
//
// Priority order (§11.4.6 — deterministic, never guesses):
//  1. Direct credentials — if StorageRoot.Username is non-nil and non-empty,
//     those are returned as-is (backward compatibility for storage roots with
//     hardcoded credentials).
//  2. Identity-index from options — if Username is nil/empty AND
//     StorageRoot.Options is a JSON object with "identity_index": N,
//     LoadSMBIdentitiesFromEnv() is called and the Nth credential identity is
//     returned.
//  3. Empty strings — when neither path resolves, the caller will attempt guest
//     or fail at auth time (the SMB client does not need non-empty credentials).
//
// SECURITY (§11.4.10): the returned password is read from env vars only
// (never from disk, never logged). The caller MUST NOT log the password.
func ResolveSMBIdentity(root *models.StorageRoot) (username, password, domain string) {
	// 1. Direct credentials take priority
	if root.Username != nil && *root.Username != "" {
		return *root.Username, getStringOrEmpty(root.Password), getStringOrEmpty(root.Domain)
	}

	// 2. No direct credentials — try identity_index from options JSON
	if root.Options == nil || *root.Options == "" {
		return "", "", ""
	}

	var opts identityOptions
	if err := json.Unmarshal([]byte(*root.Options), &opts); err != nil {
		return "", "", ""
	}
	if opts.IdentityIndex <= 0 {
		return "", "", ""
	}

	// Look up the identity from environment variables
	identities := LoadSMBIdentitiesFromEnv()
	for _, id := range identities {
		if id.Index == opts.IdentityIndex {
			username = id.Username
			password = id.Password
			if id.Domain != nil {
				domain = *id.Domain
			}
			return
		}
	}

	return "", "", ""
}

// getStringOrEmpty returns the value of a *string pointer, or "" when nil.
func getStringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
