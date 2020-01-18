package core

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// Various utilities for authentication and authorization
// ----------------------------------------------------------------------------

// ---------- Utilities (local) -----------------------------------------------

func rejectAuthentication(reason string) func(http.ResponseWriter) {
	return func(w http.ResponseWriter) {
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

	coreLogger.Verbose("HashPassword: <%s> + <%s> = <%s>", clearPassword, pwdSecretSalt, hashedPassword)

	return hashedPassword
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

		login = Login{UserID: user.ID, Token: jwt}
		createLogin(login)
	}

	return func(w http.ResponseWriter) {
		json.NewEncoder(w).Encode(successfulLogin{Token: login.Token.Jwt})
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
	coreLogger.Verbose("Basic authentication with <%s/%s>", username, password)

	return authenticateCredentials(username, password)
}
