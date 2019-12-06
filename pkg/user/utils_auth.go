package user

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Al-un/alun-api/pkg/core"
	"github.com/dgrijalva/jwt-go"
)

// ----------------------------------------------------------------------------
// Various utilities for authentication and authorization
// ----------------------------------------------------------------------------

// ---------- Variables -------------------------------------------------------

// pwdSecretSalt is used ONLY as a salt for hashing password
var pwdSecretSalt string

// jwtSecretKey is used ONLY for signing JWT
var jwtSecretKey string

func init() {
	pwdSecretSalt = "6acaa86d5e15e3df48b4eeb11dcd5c07aab709b2124424ff790304fe94b0cb2f"
	jwtSecretKey = "1f6797e3545d8d4d4b3ddd8792224e85344be25bd7aa5b8ab63ea72a4186b03f"
}

// ---------- Utilities (public) ----------------------------------------------

// IsAuthorized checks if the request has the appropriate JWT
func IsAuthorized(endpoint func(http.ResponseWriter, *http.Request), needAdmin bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if needAdmin {
			if isJWTAdmin(r) {
				endpoint(w, r)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}

			return
		}

		if isJWTLogged(r) {
			endpoint(w, r)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
	})
}

// ---------- Utilities (local) -----------------------------------------------

func rejectAuthentication(reason string) func(http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		core.AddCommonHeaders(w, "POST")
		w.WriteHeader(http.StatusForbidden)

		json.NewEncoder(w).Encode(reason)
	}
}

// hashPassword hashes a password with the "pwdSecretSalt" which is appended to
// the password as a salt
func hashPassword(clearPassword string) string {
	h := sha512.New()
	h.Write([]byte(clearPassword))
	h.Write([]byte(pwdSecretSalt))
	hashedPassword := string(h.Sum(nil))

	return hashedPassword
}

// generateJWT generate a JWT for a specific user with claims basically representing
// the user properties. List of claims is based on https://tools.ietf.org/html/rfc7519
// found through https://auth0.com/docs/tokens/jwt-claims. Tokens are valid 60 days
//
// HMAC is chosen over RSA to protect against manipulation:
// https://security.stackexchange.com/a/220190
//
// Generate Token	: https://godoc.org/github.com/dgrijalva/jwt-go#example-New--Hmac
// Custom claims	: https://godoc.org/github.com/dgrijalva/jwt-go#NewWithClaims
func generateJWT(user User) (string, error) {
	userClaims := JwtClaims{
		IsAdmin: user.IsAdmin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 60).Unix(),
			Issuer:    "api.al-un.fr",
			IssuedAt:  time.Now().Unix(),
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)

	tokenString, err := token.SignedString([]byte(jwtSecretKey))

	if err != nil {
		fmt.Printf("[JWT generation] error: %s\n", err.Error())
		return "", err
	}

	return tokenString, nil
}

// decodeJWT extracts the claims from a JWT if it is valid.
// Parse token		: https://godoc.org/github.com/dgrijalva/jwt-go#example-Parse--Hmac
// Custom claims	: https://godoc.org/github.com/dgrijalva/jwt-go#ParseWithClaims
func decodeJWT(r *http.Request) (JwtClaims, error) {
	authHeaders := r.Header["Authorization"]
	if len(authHeaders) == 0 {
		return JwtClaims{}, errors.New("Missing Authorization header")
	}
	authHeader := authHeaders[0]
	if len(authHeader) == 0 {
		return JwtClaims{}, errors.New("Missing Authorization header")
	}
	if authHeader[:6] != "Bearer" {
		return JwtClaims{}, errors.New("Invalid Authorization header")
	}

	tokenString := authHeader[7:]

	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("[JWT decode] Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecretKey), nil
	})

	if claims, ok := token.Claims.(JwtClaims); ok && token.Valid {
		fmt.Println("Claims: ", claims)
		return claims, nil
	}
	return JwtClaims{}, err
}

// isJWTExpired checks whether the claims provide an expired expiration date
//
// To extract an int64 from claims, checking https://stackoverflow.com/a/58441957/4906586
func isJWTExpired(claims JwtClaims) bool {
	// no ExpiresAt claim found
	if claims.ExpiresAt == 0 {
		return false
	}

	expTime := time.Unix(claims.ExpiresAt, 0)
	return expTime.After(time.Now())
}

// isLogged only checks if user is logged by checking a valid JWT
func isJWTLogged(r *http.Request) bool {
	claims, err := decodeJWT(r)
	// Invalid JWT
	if err != nil {
		return false
	}

	return !isJWTExpired(claims)
}

// isAdmin only checks if user is an admin by having a valid JWT with the appropriate claim
func isJWTAdmin(r *http.Request) bool {
	claims, err := decodeJWT(r)
	if err != nil {
		return false
	}

	// expired?
	if isJWTExpired(claims) {
		return false
	}

	return claims.IsAdmin
}

func authenticateCredentials(username string, clearPassword string) func(http.ResponseWriter) {
	user, err := findUserByUsernamePassword(username, clearPassword)

	if err != nil {
		rejectAuthentication("Invalid credentials")
	}

	jwt, err := generateJWT(user)
	if err != nil {
		rejectAuthentication("Error when generating JWT")
	}

	return func(w http.ResponseWriter) {
		core.AddCommonHeaders(w, "POST")
		json.NewEncoder(w).Encode(jwt)
	}
}

func authenticateBasic(authHeader string) func(http.ResponseWriter) {
	basicAuth := authHeader[6:]

	// https://golang.org/pkg/encoding/base64/#pkg-variables
	decodedBasicCredentials, err := base64.StdEncoding.DecodeString(basicAuth)
	if err != nil {
		return func(w http.ResponseWriter) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Base64 decoding error: %v\n", err)))
		}
	}

	basicCredentials := strings.Split(string(decodedBasicCredentials), ":")

	if len(basicCredentials) != 2 {
		return rejectAuthentication("Invalid authorization header")
	}

	username, password := basicCredentials[0], basicCredentials[1]
	return authenticateCredentials(username, password)
}
