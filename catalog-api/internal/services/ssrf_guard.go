package services

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// SSRF defence for catalog-api's outbound provider HTTP calls. This
// is the catalog-api mirror of pkg/nexus/ai.SSRFGuardConfig — the
// logic lives here so the internal services package does not take a
// dependency on HelixQA's module graph. Fanart.tv / IGDB / Cover-Art
// Archive endpoints are operator-configured and could be pointed at
// internal hosts in a misconfiguration or compromise, so every
// outbound resolver request runs through this guard by default.
//
// Mirrors the tldrsec/awesome-secure-defaults "SSRF Defense" item.

// SSRFGuardConfig tunes provider SSRF defence. Zero value = safe
// defaults: public hosts only, http + https only.
type SSRFGuardConfig struct {
	// AllowPrivateNetworks lets requests reach RFC1918 / loopback /
	// link-local destinations. Only flip on for a dev provider.
	AllowPrivateNetworks bool

	// AllowedSchemes — empty = http+https only.
	AllowedSchemes []string

	// Resolver overrides the default net.Resolver (tests inject
	// deterministic DNS).
	Resolver SSRFResolver
}

// SSRFResolver is the narrow DNS contract the guard needs.
type SSRFResolver interface {
	LookupIP(network, host string) ([]net.IP, error)
}

type stdlibSSRFResolver struct{}

func (stdlibSSRFResolver) LookupIP(network, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(context.Background(), network, host)
}

// ErrSSRFBlocked is returned when the guard rejects an outbound URL.
var ErrSSRFBlocked = errors.New("services: ssrf blocked")

// GuardProviderURL parses target and runs every SSRF check. Returns
// ErrSSRFBlocked (wrapped) on rejection, nil on pass.
func GuardProviderURL(target string, cfg SSRFGuardConfig) error {
	if target == "" {
		return fmt.Errorf("%w: empty url", ErrSSRFBlocked)
	}
	u, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("%w: parse: %v", ErrSSRFBlocked, err)
	}
	if err := guardScheme(u.Scheme, cfg); err != nil {
		return err
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("%w: empty host", ErrSSRFBlocked)
	}
	if host == "0.0.0.0" || host == "::" {
		return fmt.Errorf("%w: unspecified address %q", ErrSSRFBlocked, host)
	}

	// Direct IP path — no DNS needed.
	if ip := net.ParseIP(host); ip != nil {
		return guardIP(ip, cfg)
	}

	// Hostname path: resolve + reject if any returned IP is private.
	resolver := cfg.Resolver
	if resolver == nil {
		resolver = stdlibSSRFResolver{}
	}
	ips, lookupErr := resolver.LookupIP("ip", host)
	if lookupErr != nil {
		return fmt.Errorf("%w: lookup %s: %v", ErrSSRFBlocked, host, lookupErr)
	}
	if len(ips) == 0 {
		return fmt.Errorf("%w: host %q resolves to zero IPs", ErrSSRFBlocked, host)
	}
	for _, ip := range ips {
		if err := guardIP(ip, cfg); err != nil {
			return err
		}
	}
	return nil
}

func guardScheme(scheme string, cfg SSRFGuardConfig) error {
	scheme = strings.ToLower(scheme)
	allowed := cfg.AllowedSchemes
	if len(allowed) == 0 {
		allowed = []string{"http", "https"}
	}
	for _, s := range allowed {
		if scheme == strings.ToLower(s) {
			return nil
		}
	}
	return fmt.Errorf("%w: scheme %q not in allow list", ErrSSRFBlocked, scheme)
}

func guardIP(ip net.IP, cfg SSRFGuardConfig) error {
	if ip.IsUnspecified() {
		return fmt.Errorf("%w: unspecified address %s", ErrSSRFBlocked, ip)
	}
	if cfg.AllowPrivateNetworks {
		return nil
	}
	if ip.IsLoopback() {
		return fmt.Errorf("%w: loopback %s", ErrSSRFBlocked, ip)
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return fmt.Errorf("%w: link-local %s", ErrSSRFBlocked, ip)
	}
	if ip.IsPrivate() {
		return fmt.Errorf("%w: private address %s", ErrSSRFBlocked, ip)
	}
	if isIPv6UniqueLocal(ip) {
		return fmt.Errorf("%w: ULA fc00::/7 %s", ErrSSRFBlocked, ip)
	}
	if ip.IsInterfaceLocalMulticast() || ip.IsMulticast() {
		return fmt.Errorf("%w: multicast %s", ErrSSRFBlocked, ip)
	}
	return nil
}

func isIPv6UniqueLocal(ip net.IP) bool {
	v6 := ip.To16()
	if v6 == nil || ip.To4() != nil {
		return false
	}
	return v6[0]&0xfe == 0xfc
}
