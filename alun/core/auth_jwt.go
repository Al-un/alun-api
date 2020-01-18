package core

import (
	"fmt"
	"net/http"
	"time"

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
func generateJWT(user User) (token, error) {
	tokenExpiration := time.Now().Add(time.Hour * 24 * 60)

	userClaims := JwtClaims{
		IsAdmin: user.IsAdmin,
		UserID:  user.ID.Hex(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: tokenExpiration.Unix(),
			Issuer:    jwtClaimsIssuer,
			IssuedAt:  time.Now().Unix(),
			Subject:   user.Username,
		},
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)

	tokenString, err := newToken.SignedString([]byte(jwtSecretKey))

	if err != nil {
		coreLogger.Warn("[JWT generation] error: %s", err.Error())
		return token{}, err
	}

	return token{Jwt: tokenString, ExpiresOn: tokenExpiration, Status: tokenStatusActive}, nil
}

// decodeJWT extracts the claims from a JWT if it is valid.
// Parse token		: https://godoc.org/github.com/dgrijalva/jwt-go#example-Parse--Hmac
// Custom claims	: https://godoc.org/github.com/dgrijalva/jwt-go#ParseWithClaims
//
// Returns:
// - JwtClaims 	: if token is present and valid
// - int			: an `authStatus` code if some check already fails
func decodeJWT(r *http.Request) (*JwtClaims, *ServiceMessage) {
	// Fetch the Authorization header
	authHeaders := r.Header["Authorization"]
	if len(authHeaders) == 0 {
		return nil, isAuthorized
	}
	authHeader := authHeaders[0]
	if len(authHeader) == 0 {
		return nil, isAuthorizationMissing
	}
	if authHeader[:6] != "Bearer" {
		return nil, isAuthorizationInvalid
	}

	// Get the header value and strip "Bearer " out
	tokenString := authHeader[7:]

	// Parse token. Make sure hashing method is the correct one
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("[JWT decode] Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecretKey), nil
	})

	// Decipher claims
	if claims, ok := token.Claims.(*JwtClaims); ok {

		// Check token validity
		if token.Valid {

			// check if token has been invalidated
			login, err := findLoginByToken(tokenString)
			if err != nil {
				fmt.Println("[User] error when getting login by token: ", err)
			}

			if login.Token.Status != tokenStatusActive {
				switch login.Token.Status {
				case tokenStatusLogout:
					return nil, isTokenLogout
				case tokenStatusExpired:
					return nil, isTokenExpired
				case tokenStatusInvalidated:
				default:
					return nil, isTokenInvalidated
				}
			}

			return claims, isAuthorized
		}

		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				// Invalid format?
				return nil, isTokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Track in DB that token is expired
				invalidateToken(tokenString, tokenStatusExpired)
				return nil, isTokenExpired
			}
		}

		return claims, isTokenInvalid
	}

	return nil, isUnknownError
}
