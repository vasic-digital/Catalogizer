package models

import (
	"time"
)

// ShareIdentityBinding records a remembered working combination of a storage
// share and the identity that successfully authenticated against it. It is the
// persistence core of the identity-share-discovery epic: "we remember the
// combinations of shares and identity which work together".
//
// SECURITY (§11.4.10): this record stores only the identity *index* (the
// CATALOGIZER_IDENTITY_<N> slot that authenticated) and a human-readable
// *label* (a username or the literal "guest"). It MUST NEVER carry a password,
// token, or any other secret — the index is the lookup key back into the
// gitignored credential configuration, not the credential itself.
type ShareIdentityBinding struct {
	ID int64 `json:"id" db:"id"`
	// Host is the storage host (server name or address) the share lives on.
	Host string `json:"host" db:"host"`
	// ShareName is the share/export name on that host.
	ShareName string `json:"share_name" db:"share_name"`
	// Protocol is the access protocol (smb, ftp, nfs, webdav, local); defaults
	// to "smb".
	Protocol string `json:"protocol" db:"protocol"`
	// IdentityIndex is the CATALOGIZER_IDENTITY_<N> index that authenticated.
	// 0 means guest / anonymous access.
	IdentityIndex int `json:"identity_index" db:"identity_index"`
	// IdentityLabel is the username that authenticated, or the literal "guest".
	// NEVER a password or secret (§11.4.10).
	IdentityLabel string `json:"identity_label" db:"identity_label"`
	// LastOKAt is the most recent time this (host, share, identity) combination
	// was confirmed working.
	LastOKAt time.Time `json:"last_ok_at" db:"last_ok_at"`
	// CreatedAt is when the binding was first remembered.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	// UpdatedAt is when the binding was last updated.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
