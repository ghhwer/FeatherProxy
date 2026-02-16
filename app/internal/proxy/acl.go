package proxy

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"

	"FeatherProxy/app/internal/database/schema"
)

// clientIPFromRequest returns the client IP used for ACL checks.
// Order of precedence:
// 1. opts.ClientIPHeader header
// 2. X-Forwarded-For header
// 3. RemoteAddr

func clientIPFromRequest(r *http.Request, opts *schema.ACLOptions) net.IP {
	if opts != nil && opts.ClientIPHeader != "" {
		log.Println("clientIPFromRequest", "ClientIPHeader", opts.ClientIPHeader)
		return net.ParseIP(r.Header.Get(opts.ClientIPHeader))
	}
	if r.Header.Get("X-Forwarded-For") != "" {
		log.Println("clientIPFromRequest", "X-Forwarded-For", r.Header.Get("X-Forwarded-For"))
		log.Println("clientIPFromRequest", "X-Forwarded-For value", net.ParseIP(r.Header.Get("X-Forwarded-For")))
		return net.ParseIP(r.Header.Get("X-Forwarded-For"))
	}
	log.Println("clientIPFromRequest", "RemoteAddr", r.RemoteAddr)
	log.Println("clientIPFromRequest", "RemoteAddr value", net.ParseIP(r.RemoteAddr))

	return net.ParseIP(r.RemoteAddr)
}

// clientMatchesACL returns true if the client with the given IP matches any
// entry in the list. Entries may be:
//   - exact IP (IPv4/IPv6)
//   - CIDR
//   - exact hostname
//   - wildcard hostname (e.g. "*.internal.com.br")
//
// Matching is performed in two phases:
//  1. IP/CIDR entries are evaluated exactly as before (no DNS).
//  2. If there are hostname entries and resolver is non-nil, reverse DNS
//     is performed for the IP and the resulting hostnames are matched
//     against hostname and wildcard entries.
func clientMatchesACL(ctx context.Context, ip net.IP, list []string, resolver HostnameResolver) bool {
	if ip == nil {
		return false
	}

	// First pass: IP and CIDR entries (backwards compatible behavior).
	hasHostnameEntries := false
	for _, raw := range list {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			_, cidr, err := net.ParseCIDR(entry)
			if err != nil {
				continue
			}
			if cidr.Contains(ip) {
				return true
			}
			continue
		}
		if other := net.ParseIP(entry); other != nil {
			if ip.Equal(other) {
				return true
			}
			continue
		}
		// Not CIDR and not IP -> treat as hostname pattern.
		hasHostnameEntries = true
	}

	// No hostname entries or no resolver configured: no further matches.
	if !hasHostnameEntries || resolver == nil {
		return false
	}

	if ctx == nil {
		ctx = context.Background()
	}

	hostnames, err := resolver.ReverseLookup(ctx, ip)
	if err != nil || len(hostnames) == 0 {
		// On lookup failure or no names, hostname entries do not match.
		return false
	}

	// Normalize ACL hostname entries once for comparison.
	normalizedEntries := make([]string, 0, len(list))
	for _, raw := range list {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			continue
		}
		if other := net.ParseIP(entry); other != nil {
			continue
		}
		entry = strings.ToLower(entry)
		normalizedEntries = append(normalizedEntries, entry)
	}
	if len(normalizedEntries) == 0 {
		return false
	}

	for _, h := range hostnames {
		host := strings.ToLower(strings.TrimSpace(h))
		if host == "" {
			continue
		}
		for _, entry := range normalizedEntries {
			if strings.HasPrefix(entry, "*.") && len(entry) > 2 {
				// Wildcard suffix match on label boundary, e.g. "*.internal.com.br".
				suffix := entry[2:]
				if len(host) <= len(suffix) {
					continue
				}
				if strings.HasSuffix(host, "."+suffix) {
					return true
				}
				continue
			}
			// Exact hostname match.
			if host == entry {
				return true
			}
		}
	}

	return false
}

// aclDeny returns true if the request should be denied based on ACL options.
// If opts is nil or opts.Mode is "off", returns false (allow).
// allow_only: deny if client IP is not in AllowList.
// deny_only: deny if client IP is in DenyList.
func aclDeny(ctx context.Context, r *http.Request, opts *schema.ACLOptions, resolver HostnameResolver) bool {
	if opts == nil || opts.Mode == "off" {
		return false
	}
	clientIP := clientIPFromRequest(r, opts)
	switch opts.Mode {
	case "allow_only":
		if len(opts.AllowList) == 0 {
			return true // allow nobody -> deny all
		}
		return !clientMatchesACL(ctx, clientIP, opts.AllowList, resolver)
	case "deny_only":
		return clientMatchesACL(ctx, clientIP, opts.DenyList, resolver)
	default:
		return false
	}
}
