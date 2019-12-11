package user

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Al-un/alun-api/pkg/core"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
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
func IsAuthorized(endpoint func(http.ResponseWriter, *http.Request), authCheckConfig AuthCheckConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var claims JwtClaims
		claimsRef, authCheckStatusCode := decodeJWT(r)
		if claimsRef != nil {
			claims = *claimsRef
		}

		var log = func(msg string) {
			log.Printf("%v: %v\n", r.URL.Path, msg)
		}

		// fmt.Printf("%v: %v\n", authCheckStatusCode, claims)

		// For the moment, only use user ":userId" in the route
		authCheckConfig.UserID = mux.Vars(r)["userId"]

		// Get the proper check
		switch authCheckConfig.Mode {
		case AuthCheckMode.isLogged:
			// do nothing, already handled by decodeJWT

		case AuthCheckMode.isAdmin:
			// ignore if previous check already fail
			if authCheckStatusCode == authStatus.isAuthorized {
				// isAdmin?
				if !claims.IsAdmin {
					authCheckStatusCode = authStatus.isNotAuthorized
				}
			}

		case AuthCheckMode.isAdminOrUser:
			// isAdmin or OwnUser?
			if authCheckStatusCode == authStatus.isAuthorized {
				if !(claims.IsAdmin || claims.UserID == authCheckConfig.UserID) {
					authCheckStatusCode = authStatus.isNotAuthorized
				}
			}

		case AuthCheckMode.isUser:
			// isOwnUser?
			if authCheckStatusCode == authStatus.isAuthorized {
				if claims.UserID != authCheckConfig.UserID {
					authCheckStatusCode = authStatus.isNotAuthorized
				}
			}
		}

		// fmt.Printf("AuthConfig: %v vs %v is %v\n", authCheckConfig, claims, claims.UserID == authCheckConfig.UserID)

		// Here were go
		switch authCheckStatusCode {

		case authStatus.isAuthorized:
			endpoint(w, r)

		case authStatus.isNotAuthorized:
			log("Not authorized")
			w.WriteHeader(http.StatusForbidden)

		case authStatus.isTokenExpired:
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(core.ErrorMsg{Error: authErrorMsg.isTokenExpired})
			log(authErrorMsg.isTokenExpired)
		case authStatus.isTokenMalformed:
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(core.ErrorMsg{Error: authErrorMsg.isTokenMalformed})
			log(authErrorMsg.isTokenMalformed)
		case authStatus.isTokenInvalid:
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(core.ErrorMsg{Error: authErrorMsg.isTokenInvalid})
			log(authErrorMsg.isTokenInvalid)

		case authStatus.isAuthorizationMissing:
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(core.ErrorMsg{Error: authErrorMsg.isAuthorizationMissing})
			log(authErrorMsg.isAuthorizationMissing)
		case authStatus.isAuthorizationInvalid:
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(core.ErrorMsg{Error: authErrorMsg.isAuthorizationInvalid})
			log(authErrorMsg.isAuthorizationInvalid)

		case authStatus.isUnknownError:
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(core.ErrorMsg{Error: authErrorMsg.isUnknownError})
			log(authErrorMsg.isUnknownError)
		}
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
func generateJWT(user User) (Token, error) {
	tokenExpiration := time.Now().Add(time.Hour * 24 * 60)

	userClaims := JwtClaims{
		IsAdmin: user.IsAdmin,
		UserID:  user.ID.Hex(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: tokenExpiration.Unix(),
			Issuer:    "api.al-un.fr",
			IssuedAt:  time.Now().Unix(),
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)

	// fmt.Println("Generating token ", token)

	tokenString, err := token.SignedString([]byte(jwtSecretKey))

	if err != nil {
		fmt.Printf("[JWT generation] error: %s\n", err.Error())
		return Token{}, err
	}

	return Token{Jwt: tokenString, ExpiresOn: tokenExpiration, IsInvalid: false}, nil
}

// decodeJWT extracts the claims from a JWT if it is valid.
// Parse token		: https://godoc.org/github.com/dgrijalva/jwt-go#example-Parse--Hmac
// Custom claims	: https://godoc.org/github.com/dgrijalva/jwt-go#ParseWithClaims
//
// Returns:
// - JwtClaims 	: if token is present and valid
// - int			: an `authStatus` code if some check already fails
func decodeJWT(r *http.Request) (*JwtClaims, int) {
	// Fetch the Authorization header
	authHeaders := r.Header["Authorization"]
	if len(authHeaders) == 0 {
		return nil, authStatus.isAuthorizationMissing
	}
	authHeader := authHeaders[0]
	if len(authHeader) == 0 {
		return nil, authStatus.isAuthorizationMissing
	}
	if authHeader[:6] != "Bearer" {
		return nil, authStatus.isAuthorizationInvalid
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

			// // check if token has been invalidated
			// login, err := findLoginByToken(tokenString)
			// if err != nil {
			// 	fmt.Println("[User] error when getting login by token: ", err)
			// }

			// if login.Token.IsInvalid {
			// 	return &claims, authStatus.isTokenInvalidated
			// }

			return claims, authStatus.isAuthorized
		}

		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, authStatus.isTokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, authStatus.isTokenExpired
			}
		}

		return claims, authStatus.isTokenInvalid
	}

	return nil, authStatus.isUnknownError
}

func authenticateCredentials(username string, clearPassword string) func(http.ResponseWriter) {
	user, err := findUserByUsernamePassword(username, clearPassword)

	if err != nil {
		return rejectAuthentication("Invalid credentials")
	}

	login, err := findLoginWithValidToken(user)
	// TODO: assuming error is when no login is found
	// Also handled expired token
	if err != nil || login.Token.ExpiresOn.After(time.Now()) {
		// fmt.Printf(">>>>>>>>>>>> %v <<<<<<<<, \n", err)

		jwt, err := generateJWT(user)
		if err != nil {
			return rejectAuthentication("Error when generating JWT")
		}

		login = Login{
			UserID: user.ID,
			Token:  jwt,
		}
		createLogin(login)
	}

	return func(w http.ResponseWriter) {
		core.AddCommonHeaders(w, "POST")
		json.NewEncoder(w).Encode(login.Token.Jwt)
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
