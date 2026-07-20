package cli

import "testing"

func TestAuthAliasUsesBrowserLogin(t *testing.T) {
	if !loginCmd.HasAlias("auth") {
		t.Fatal("login command must expose auth as a browser-login alias")
	}

	resolved, _, err := rootCmd.Find([]string{"auth"})
	if err != nil {
		t.Fatalf("resolve auth alias: %v", err)
	}
	if resolved != loginCmd {
		t.Fatalf("auth resolved to %q, want login", resolved.Name())
	}
}
