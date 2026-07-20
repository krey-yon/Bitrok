package api

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "0123456789abcdef0123456789abcdef"

func signTestJWT(t *testing.T, method jwt.SigningMethod, claims jwt.Claims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(method, claims).SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestValidateJWTAcceptsRequiredClaims(t *testing.T) {
	token := signTestJWT(t, jwt.SigningMethodHS256, relayClaims{
		Username: "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			Issuer:    "bitrok",
			Audience:  jwt.ClaimStrings{"bitrok-cli"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})

	userID, username, ok := validateJWT(token, testJWTSecret, "bitrok-cli", "bitrok")
	if !ok || userID != "user-1" || username != "alice" {
		t.Fatalf("validateJWT() = %q, %q, %v", userID, username, ok)
	}
}

func TestValidateJWTRejectsUnsafeClaims(t *testing.T) {
	tests := map[string]struct {
		method jwt.SigningMethod
		claims relayClaims
	}{
		"missing expiration": {
			method: jwt.SigningMethodHS256,
			claims: relayClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", Issuer: "bitrok", Audience: jwt.ClaimStrings{"bitrok-cli"}}},
		},
		"expired": {
			method: jwt.SigningMethodHS256,
			claims: relayClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", Issuer: "bitrok", Audience: jwt.ClaimStrings{"bitrok-cli"}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute))}},
		},
		"wrong issuer": {
			method: jwt.SigningMethodHS256,
			claims: relayClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", Issuer: "other", Audience: jwt.ClaimStrings{"bitrok-cli"}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}},
		},
		"wrong audience": {
			method: jwt.SigningMethodHS256,
			claims: relayClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", Issuer: "bitrok", Audience: jwt.ClaimStrings{"other"}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}},
		},
		"wrong algorithm": {
			method: jwt.SigningMethodHS384,
			claims: relayClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", Issuer: "bitrok", Audience: jwt.ClaimStrings{"bitrok-cli"}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			token := signTestJWT(t, tt.method, tt.claims)
			if _, _, ok := validateJWT(token, testJWTSecret, "bitrok-cli", "bitrok"); ok {
				t.Fatal("validateJWT accepted an unsafe token")
			}
		})
	}
}
