package util

import (
	"fmt"
	"regexp"
	"strings"
)

// FormatBytes returns a human-readable byte count.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

var (
	slugNonAlnum = regexp.MustCompile(`[^a-z0-9-]+`)
	slugDashRun  = regexp.MustCompile(`-+`)
)

// Slugify normalizes an app name into a DNS-label-safe slug: lowercase,
// spaces/underscores → dashes, strip everything else, collapse dash runs,
// trim leading/trailing dashes, cap at 63 chars (DNS label limit).
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	s = slugNonAlnum.ReplaceAllString(s, "")
	s = slugDashRun.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 63 {
		s = s[:63]
		s = strings.Trim(s, "-")
	}
	return s
}

// BuildTunnelHost joins an app and username into one DNS label, shortening the
// app portion when necessary so the label never exceeds 63 bytes.
func BuildTunnelHost(app, username, domain string) (string, string, error) {
	app = Slugify(app)
	username = Slugify(username)
	domain = strings.ToLower(strings.Trim(strings.TrimSpace(domain), "."))
	if app == "" || username == "" || domain == "" {
		return "", "", fmt.Errorf("app, username, and domain are required")
	}
	maxAppLength := 63 - len(username) - 1
	if maxAppLength < 1 {
		return "", "", fmt.Errorf("username is too long for a tunnel hostname")
	}
	if len(app) > maxAppLength {
		app = strings.Trim(app[:maxAppLength], "-")
	}
	if app == "" {
		return "", "", fmt.Errorf("app name is too long for a tunnel hostname")
	}
	return app, app + "-" + username + "." + domain, nil
}
