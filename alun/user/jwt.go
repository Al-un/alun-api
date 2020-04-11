package user

import (
	"time"

	"github.com/Al-un/alun-api/alun/core"
	"github.com/dgrijalva/jwt-go"
)

const jwtClaimsIssuer = "api.al-un.fr"

// generateJWT generate a JWT for a specific user with claims basically representing
// the user properties. List of claims is based on https://tools.ietf.org/html/rfc7519
// found through https://auth0.com/docs/tokens/jwt-claims. Tokens are valid 60 days
//
// HMAC is chosen over RSA to protect against manipulation:
// https://security.stackexchange.com/a/220190
//
// Generate Token	: https://godoc.org/github.com/dgrijalva/jwt-go#example-New--Hmac
// Custom claims	: https://godoc.org/github.com/dgrijalva/jwt-go#NewWithClaims
func generateJWT(user User) (authToken, error) {
	tokenExpiration := time.Now().Add(time.Hour * 24 * 60)

	userClaims := core.JwtClaims{
		IsAdmin: user.IsAdmin,
		UserID:  user.ID.Hex(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: tokenExpiration.Unix(),
			Issuer:    jwtClaimsIssuer,
			IssuedAt:  time.Now().Unix(),
			Subject:   user.Username,
		},
	}

	tokenString, err := core.BuildJWT(userClaims)

	if err != nil {
		userLogger.Warn("[JWT generation] error: %s", err.Error())
		return authToken{}, err
	}

	return authToken{Jwt: tokenString, ExpiresOn: tokenExpiration, Status: tokenStatusActive}, nil
}
