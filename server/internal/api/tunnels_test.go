package api

import (
	"strings"
	"testing"
)

func TestValidateTunnelHostEnforcesPlatformOwnership(t *testing.T) {
	valid := []string{"app-alice.bitrok.tech", "my-app-alice.bitrok.tech"}
	for _, host := range valid {
		if err := validateTunnelHost(host, "bitrok.tech", "alice"); err != nil {
			t.Errorf("validateTunnelHost(%q) returned %v", host, err)
		}
	}

	invalid := []string{"bitrok.tech", "api.bitrok.tech", "app-bob.bitrok.tech", "app-alice.bitrok.tech.evil.com", "demo.example.com", "not a host"}
	for _, host := range invalid {
		if err := validateTunnelHost(host, "bitrok.tech", "alice"); err == nil {
			t.Errorf("validateTunnelHost(%q) unexpectedly succeeded", host)
		}
	}

	if err := validateTunnelHost("app-alice.bitrok.tech", "bitrok.tech", ""); err == nil {
		t.Fatal("platform host succeeded without a username claim")
	}
}

func TestDecodeSingleJSONRejectsUnknownAndTrailingValues(t *testing.T) {
	var dst struct {
		Name string `json:"name"`
	}
	if err := decodeSingleJSON(strings.NewReader(`{"name":"ok","admin":true}`), &dst); err == nil {
		t.Fatal("unknown field was accepted")
	}
	if err := decodeSingleJSON(strings.NewReader(`{"name":"ok"} {"name":"second"}`), &dst); err == nil {
		t.Fatal("multiple JSON values were accepted")
	}
}
