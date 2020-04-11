package core

import (
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

// ----------------------------------------------------------------------------
//	Authorization: Access checking
// ----------------------------------------------------------------------------

// CheckPublic always returns true to provided public access
var CheckPublic AccessChecker = func(r *http.Request, claims JwtClaims) bool {
	return true
}

// CheckIfLogged is the AuthChecher ensuring the request has a properly
// logged-in user
var CheckIfLogged AccessChecker = func(r *http.Request, claims JwtClaims) bool {
	return claims.UserID != ""
}

// CheckIfAdmin simply checks if the JWT has admin privilege
var CheckIfAdmin AccessChecker = func(r *http.Request, claims JwtClaims) bool {
	return claims.IsAdmin
}

// DoIfAccess ensures that the provided accessChecker passes before proceeding
// to the authenticatedHandler.
//
// Otherwise the request is rejected with the appropriate error code with error
// message
func DoIfAccess(canAccess AccessChecker, authenticatedHandler AuthenticatedHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var claims JwtClaims
		claimsRef, checkServMsg := DecodeJWT(r)
		if claimsRef != nil {
			claims = *claimsRef
		}

		if checkServMsg.HTTPStatus != 0 {
			checkServMsg.Write(w, r)
			return
		}

		if canAccess(r, claims) {
			authenticatedHandler(w, r, claims)
		} else {
			isNotAuthorized.Write(w, r)
		}

	})
}

// ----------------------------------------------------------------------------
//	JWT
// ----------------------------------------------------------------------------

// BuildJWT generate a JWT from a specific list of claims.
// List of claims is based on https://tools.ietf.org/html/rfc7519 found through
// https://auth0.com/docs/tokens/jwt-claims.
//
// HMAC is chosen over RSA to protect against manipulation:
// https://security.stackexchange.com/a/220190
//
// Generate Token	: https://godoc.org/github.com/dgrijalva/jwt-go#example-New--Hmac
// Custom claims	: https://godoc.org/github.com/dgrijalva/jwt-go#NewWithClaims
func BuildJWT(claims JwtClaims) (string, error) {
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return newToken.SignedString([]byte(jwtSecretKey))
}

// DecodeJWT extracts the claims from a JWT if it is valid.
// Parse token		: https://godoc.org/github.com/dgrijalva/jwt-go#example-Parse--Hmac
// Custom claims	: https://godoc.org/github.com/dgrijalva/jwt-go#ParseWithClaims
//
// 11-Apr-2020: Make authentication check 100% stateless for a microservice-ready
// architecture by removing the check in the database of the JWT status: a pure JWT
// is 100% stateless.
//
// Returns:
// - JwtClaims 	: if token is present and valid
// - int			: an `authStatus` code if some check already fails
func DecodeJWT(r *http.Request) (*JwtClaims, *ServiceMessage) {
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
			return claims, isAuthorized
		}

		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, isTokenMalformed
			}
		}

		return claims, isTokenInvalid
	}

	return nil, isUnknownError
}
