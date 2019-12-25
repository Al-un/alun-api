package core

import (
	"net/http"
)

// ---------- Variables -------------------------------------------------------

// pwdSecretSalt is used ONLY as a salt for hashing password
var pwdSecretSalt string

// jwtSecretKey is used ONLY for signing JWT
var jwtSecretKey string

func init() {
	pwdSecretSalt = "6acaa86d5e15e3df48b4eeb11dcd5c07aab709b2124424ff790304fe94b0cb2f"
	jwtSecretKey = "1f6797e3545d8d4d4b3ddd8792224e85344be25bd7aa5b8ab63ea72a4186b03f"
}

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
// Otherwise the request is rejected with the appropriate error code with error
// message
func DoIfAccess(canAccess AccessChecker, authenticatedHandler AuthenticatedHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var claims JwtClaims
		claimsRef, checkServMsg := decodeJWT(r)
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
