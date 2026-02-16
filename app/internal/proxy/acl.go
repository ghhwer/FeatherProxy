package proxy

import (
	"net"
	"net/http"
	"strings"

	"FeatherProxy/app/internal/database/schema"
)

// clientIPFromRequest returns the client IP used for ACL checks.
// If opts.ClientIPHeader is non-empty, that header's value is used (first comma-separated segment);
// otherwise r.RemoteAddr is used (host part only). Returns nil if the value cannot be parsed as IP.
func clientIPFromRequest(r *http.Request, opts *schema.ACLOptions) net.IP {
	var raw string
	if opts != nil && strings.TrimSpace(opts.ClientIPHeader) != "" {
		raw = strings.TrimSpace(r.Header.Get(opts.ClientIPHeader))
		if idx := strings.Index(raw, ","); idx >= 0 {
			raw = strings.TrimSpace(raw[:idx])
		}
	}
	if raw == "" {
		raw = r.RemoteAddr
	}
	host, _, err := net.SplitHostPort(raw)
	if err != nil {
		// No port (e.g. single IP)
		host = raw
	}
	host = strings.TrimSpace(host)
	if host == "" {
		return nil
	}
	ip := net.ParseIP(host)
	return ip
}

// ipInList returns true if ip is in the list (exact match or CIDR).
func ipInList(ip net.IP, list []string) bool {
	if ip == nil {
		return false
	}
	for _, entry := range list {
		entry = strings.TrimSpace(entry)
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
		} else {
			other := net.ParseIP(entry)
			if other != nil && ip.Equal(other) {
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
func aclDeny(r *http.Request, opts *schema.ACLOptions) bool {
	if opts == nil || opts.Mode == "off" {
		return false
	}
	clientIP := clientIPFromRequest(r, opts)
	switch opts.Mode {
	case "allow_only":
		if len(opts.AllowList) == 0 {
			return true // allow nobody -> deny all
		}
		return !ipInList(clientIP, opts.AllowList)
	case "deny_only":
		return ipInList(clientIP, opts.DenyList)
	default:
		return false
	}
}
