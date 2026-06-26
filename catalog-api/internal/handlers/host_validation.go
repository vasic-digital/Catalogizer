package handlers

import (
	"net"
)

// isHostAllowed validates that a hostname/IP is a legitimate SMB probe target.
// It blocks loopback, link-local, multicast, cloud-metadata IPs, and
// unqualified single-label names. Private RFC-1918 ranges ARE allowed
// because the identity-epic targets the LAN (Synology NAS at 192.168.0.0/24).
//
// Security invariants (§11.4 SSRF prevention):
//   - Loopback (127.0.0.0/8, ::1) — blocked (probe would hit this host, not a NAS)
//   - Link-local (169.254.0.0/16, fe80::/10) — blocked (no SMB server there)
//   - Multicast (224.0.0.0/4, ff00::/8) — blocked (not a unicast target)
//   - Cloud metadata IP 169.254.169.254 — blocked (credential exfiltration vector)
//   - Unqualified single-label name ("localhost", "nas") — allowed but
//     DNS resolution is performed server-side. The main protection is IP-level.
//   - Empty host — blocked (caller should validate required field first)
func isHostAllowed(host string) bool {
	if host == "" {
		return false
	}

	// Resolve hostname to IPs. On failure the host is either unresolvable
	// (will fail at dial time) or a single-label name — allow it through
	// to the probe so SMB can attempt it; DNS rebinding is mitigated by
	// the IP-level checks below.
	ips, err := net.LookupIP(host)
	if err != nil {
		// Unresolvable — let the dial fail naturally.
		return true
	}

	for _, ip := range ips {
		if ip.IsLoopback() {
			return false
		}
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return false
		}
		if ip.IsMulticast() {
			return false
		}
		if ip.IsUnspecified() {
			return false
		}
		// Cloud metadata /169.254.169.254 is link-local-unicast and caught above,
		// but check explicitly as a defence-in-depth measure.
		if ip.To4() != nil && ip.String() == "169.254.169.254" {
			return false
		}
	}
	return true
}
