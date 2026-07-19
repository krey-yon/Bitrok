package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// UsernameFromToken extracts the `username` claim from a JWT's payload without
// verifying the signature. The token was obtained from the dashboard (trusted
// source), and the relay server validates the signature on every call — here
// we only read the claim the dashboard minted so we can build a deterministic
// host without an extra API round-trip.
//
// Returns ("", nil) when the token is well-formed but has no username claim
// (e.g. a token minted before the username field existed); the caller prompts
// or asks for a re-login.
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
	username, _ := claims["username"].(string)
	return username, nil
}
