package api

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// validateJWT checks if a token is a valid, unexpired JWT signed with the given secret.
func validateJWT(tokenString, secret, expectedAudience, expectedIssuer string) (userID string, ok bool) {
	if secret == "" || tokenString == "" {
		return "", false
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return "", false
		}
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", false
	}

	if expectedAudience != "" {
		switch aud := claims["aud"].(type) {
		case string:
			if aud != expectedAudience {
				return "", false
			}
		case []interface{}:
			found := false
			for _, a := range aud {
				if s, ok := a.(string); ok && s == expectedAudience {
					found = true
					break
				}
			}
			if !found {
				return "", false
			}
		default:
			return "", false
		}
	}

	if expectedIssuer != "" {
		iss, _ := claims["iss"].(string)
		if iss != expectedIssuer {
			return "", false
		}
	}

	return sub, true
}
