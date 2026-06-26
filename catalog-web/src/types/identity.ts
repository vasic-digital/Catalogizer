/**
 * Identity + Network-Share discovery type definitions.
 *
 * Domain model per docs/design/identity_share_discovery/DESIGN.md §2.1–§2.3:
 * identities (multi-scheme credentials), discovered hosts/shares, and
 * share-identity bindings (the remembered working pairing).
 *
 * §11.4.10 — NO credential values appear in any type; secret fields carry
 * a *_ref handle string; the UI NEVER renders a secret value into the DOM
 * after entry. The typed meta IS rendered (name, scheme type, username,
 * domain, priority, enabled status) — never the secret itself.
 * §11.4.162 — all UI components consuming these types ship light+dark.
 */

/** Identity authentication scheme. */
export type IdentityType =
  | 'credentials'   // username + password
  | 'api_token'     // bearer / static token
  | 'ssh_key'       // key file + optional passphrase
  | 'webdav_basic'  // WebDAV Basic auth (username + password)
  | 'oauth2'        // OAuth2 token (refresh-token flow)
  | 'anonymous'     // synthetic — always present, tried first

/** Identity — the canonical credential abstraction, decoupled from any share. */
export interface Identity {
  /** Server-assigned id (0 for new local-only). */
  id: number
  /** Operator-facing label, e.g. "nas-admin". */
  name: string
  /** Scheme discriminator. */
  type: IdentityType
  /** Username (nullable — N/A for api_token / ssh_key / oauth2). */
  username: string | null
  /**
   * Opaque handle into the secret store — NEVER the raw secret value.
   * The UI renders a "••••••••" placeholder instead of this ref.
   */
  secret_ref: string | null
  /** SMB/NFS domain (nullable). */
  domain: string | null
  /** SSH key path (ssh_key type only). */
  key_path: string | null
  /** Whether this identity is active for probing. */
  enabled: boolean
  /** Probe priority (0 = anonymous synthetic, lower = tried first). */
  priority: number
  created_at?: string
  updated_at?: string
}

/** Subset used for creating or updating an identity (secrets sent to the API). */
export interface IdentityRequest {
  name: string
  type: IdentityType
  username?: string | null
  /** Raw secret value — sent to API ONCE on create/update, NEVER stored client-side. */
  secret?: string | null
  domain?: string | null
  key_path?: string | null
  enabled?: boolean
  priority?: number
}

/** A discovered host on the LAN. */
export interface DiscoveredHost {
  id: number
  ip: string
  hostname: string | null
  first_seen: string
  last_seen: string
  reachable: boolean
  /** JSON-encoded list of interface names / addresses. */
  ifaces_json: string | null
  /** MAC OUI vendor (heuristic). */
  oui_vendor?: string | null
}

/** A discovered network share on a host. */
export interface DiscoveredShare {
  id: number
  host_id: number
  host_ip: string
  host_hostname: string | null
  protocol: string
  share_name: string
  port: number
  first_seen: string
  last_seen: string
  enumeration_evidence_path: string | null
}

/** Status of a share-identity binding (remembered working pairing). */
export type BindingStatus =
  | 'ok'
  | 'unauthenticated'
  | 'failed'
  | 'revalidating'

/** A share-identity binding — the remembered (host, share, identity, protocol) key. */
export interface ShareIdentityBinding {
  id: number
  share_id: number
  identity_id: number | null
  identity_name: string | null
  status: BindingStatus
  anonymous_ok: boolean
  last_ok_at: string | null
  last_attempt_at: string | null
  probe_evidence_path: string | null
}

/** Grouped view: one host → its shares → each share's current binding. */
export interface HostWithShares {
  host: DiscoveredHost
  shares: Array<{
    share: DiscoveredShare
    binding: ShareIdentityBinding | null
  }>
}

/** Result of a probe action (anon-first → identity-fallback). */
export interface ProbeResult {
  share_id: number
  share_name: string
  protocol: string
  host: string
  anonymous_ok: boolean
  bound_identity_id: number | null
  bound_identity_name: string | null
  status: BindingStatus
  evidence_path: string | null
}

/** Result of a full network scan. */
export interface ScanResult {
  hosts_found: number
  shares_found: number
  bindings_created: number
  scan_duration_ms: number
}
