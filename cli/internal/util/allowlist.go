package util

import (
	"fmt"
	"net"
	"strings"
)

// AllowList is a set of CIDR networks used to filter visitor IPs.
type AllowList struct {
	nets []*net.IPNet
	raw  []string
}

// ParseAllowList parses CIDR strings (e.g. "10.0.0.0/8", "192.168.1.1/32").
// Bare IPs are treated as /32 (IPv4) or /128 (IPv6).
func ParseAllowList(cidrs []string) (*AllowList, error) {
	if len(cidrs) == 0 {
		return nil, nil
	}
	al := &AllowList{raw: append([]string{}, cidrs...)}
	for _, c := range cidrs {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if !strings.Contains(c, "/") {
			if ip := net.ParseIP(c); ip != nil {
				if ip.To4() != nil {
					c = c + "/32"
				} else {
					c = c + "/128"
				}
			}
		}
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			return nil, fmt.Errorf("invalid allow-ip %q: %w", c, err)
		}
		al.nets = append(al.nets, n)
	}
	if len(al.nets) == 0 {
		return nil, nil
	}
	return al, nil
}

// Empty reports whether no networks were configured.
func (a *AllowList) Empty() bool {
	return a == nil || len(a.nets) == 0
}

// Contains reports whether ip is inside any allowed network.
func (a *AllowList) Contains(ip net.IP) bool {
	if a.Empty() {
		return true
	}
	if ip == nil {
		return false
	}
	for _, n := range a.nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// Strings returns the original CIDR strings.
func (a *AllowList) Strings() []string {
	if a == nil {
		return nil
	}
	return a.raw
}

// ClientIPFromHeaders extracts the visitor IP from proxy headers.
// Preference: first hop of X-Forwarded-For, then X-Real-IP, then empty.
func ClientIPFromHeaders(headers map[string]string) net.IP {
	if headers == nil {
		return nil
	}
	// Header keys may arrive with various casings; scan case-insensitively.
	var xff, xri string
	for k, v := range headers {
		switch strings.ToLower(k) {
		case "x-forwarded-for":
			xff = v
		case "x-real-ip":
			xri = v
		}
	}
	if xff != "" {
		// "client, proxy1, proxy2" — leftmost is original client.
		parts := strings.Split(xff, ",")
		ipStr := strings.TrimSpace(parts[0])
		// RemoteAddr form host:port
		if host, _, err := net.SplitHostPort(ipStr); err == nil {
			ipStr = host
		}
		return net.ParseIP(ipStr)
	}
	if xri != "" {
		ipStr := strings.TrimSpace(xri)
		if host, _, err := net.SplitHostPort(ipStr); err == nil {
			ipStr = host
		}
		return net.ParseIP(ipStr)
	}
	return nil
}
