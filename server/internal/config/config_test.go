package config

import "testing"

func TestValidateRestoresRequiredJWTClaims(t *testing.T) {
	cfg := Default()
	cfg.JWTSecret = "0123456789abcdef0123456789abcdef"
	cfg.JWTExpectedAudience = ""
	cfg.JWTExpectedIssuer = ""
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.JWTExpectedAudience != "bitrok-cli" || cfg.JWTExpectedIssuer != "bitrok" {
		t.Fatalf("claims = %q/%q", cfg.JWTExpectedAudience, cfg.JWTExpectedIssuer)
	}
}

func TestValidateRejectsUnsafeConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"invalid domain", func(c *Config) { c.Domain = "bad domain" }},
		{"invalid redis url", func(c *Config) { c.RedisURL = "https://redis.example.com" }},
		{"disabled rate limit", func(c *Config) { c.RateLimitCapacity = 0 }},
		{"disabled body limit", func(c *Config) { c.MaxRequestBodyBytes = 0 }},
		{"disabled tunnel quota", func(c *Config) { c.MaxTunnelsPerUser = 0 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			cfg.JWTSecret = "0123456789abcdef0123456789abcdef"
			tt.mutate(cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("Validate unexpectedly succeeded")
			}
		})
	}
}
