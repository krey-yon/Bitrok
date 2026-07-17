package util

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
)

var hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)

// ValidateHostname checks if a string is a valid FQDN.
func ValidateHostname(h string) error {
	if h == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	if len(h) > 253 {
		return fmt.Errorf("hostname too long")
	}
	if !hostnameRegex.MatchString(h) {
		return fmt.Errorf("invalid hostname format")
	}
	return nil
}

// ValidatePort checks if a port is in the valid range.
func ValidatePort(p int) error {
	if p < 1 || p > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}

// ValidatePortString parses and validates a port string.
func ValidatePortString(s string) (int, error) {
	p, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid port number")
	}
	if err := ValidatePort(p); err != nil {
		return 0, err
	}
	return p, nil
}

// ResolveLocalAddr checks if localhost:port is reachable.
func ResolveLocalAddr(port int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("cannot connect to %s: %w", addr, err)
	}
	conn.Close()
	return nil
}
