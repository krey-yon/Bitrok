package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// UsernameFromToken extracts a URL slug from a JWT payload without verifying
// the signature. Prefer `username`, then fall back to other claims so older
// tokens still work until the user re-logins.
//
// Returns ("", nil) only when no usable claim exists.
func UsernameFromToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("malformed token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid token payload: %w", err)
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("invalid token claims: %w", err)
	}

	// Preferred explicit claim (dashboard mints this).
	if u := claimString(claims, "username"); u != "" {
		if s := Slugify(u); s != "" {
			return s, nil
		}
	}
	// Other possible claim names.
	for _, key := range []string{"preferred_username", "login", "name", "nickname"} {
		if u := claimString(claims, key); u != "" {
			if s := Slugify(u); s != "" {
				return s, nil
			}
		}
	}
	// Email local-part (e.g. vikas@… → vikas). Last resort for legacy tokens.
	if email := claimString(claims, "email"); email != "" {
		local, _, _ := strings.Cut(email, "@")
		if s := Slugify(local); s != "" {
			return s, nil
		}
	}

	return "", nil
}

func claimString(claims map[string]any, key string) string {
	v, _ := claims[key].(string)
	return strings.TrimSpace(v)
}
