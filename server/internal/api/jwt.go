package api

import (
	"github.com/golang-jwt/jwt/v5"
)

type relayClaims struct {
	Username string `json:"username,omitempty"`
	jwt.RegisteredClaims
}

// validateJWT checks if a token is a valid, unexpired JWT signed with the given secret.
func validateJWT(tokenString, secret, expectedAudience, expectedIssuer string) (userID, username string, ok bool) {
	if secret == "" || tokenString == "" {
		return "", "", false
	}

	claims := &relayClaims{}
	options := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	}
	if expectedAudience != "" {
		options = append(options, jwt.WithAudience(expectedAudience))
	}
	if expectedIssuer != "" {
		options = append(options, jwt.WithIssuer(expectedIssuer))
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, options...)
	if err != nil || !token.Valid {
		return "", "", false
	}

	if claims.Subject == "" {
		return "", "", false
	}

	return claims.Subject, claims.Username, true
}
