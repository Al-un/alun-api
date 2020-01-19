package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Al-un/alun-api/alun/utils"
)

// authUser authenticates user with BASIC methods or other
func authUser(w http.ResponseWriter, r *http.Request) {

	// --- JSON-based authentication
	var user authenticatedUser
	json.NewDecoder(r.Body).Decode(&user)
	if user.Username != "" && user.Password != "" {
		coreLogger.Verbose("JSON authentication: %s/%s", user.Username, user.Password)
		authenticateCredentials(user.Username, user.Password)(w)
		return
	}

	// --- Header-based authentication
	authHeaders := r.Header["Authorization"]
	if len(authHeaders) == 0 {
		rejectAuthentication("Missing Authorization")(w)
		return
	}

	authHeader := authHeaders[0]
	if len(authHeader) == 0 {
		rejectAuthentication("Incorrect Authorization header")(w)
		return
	}

	// ------ BASIC Authentication
	if authHeader[:5] == "Basic" {
		authenticateBasic(authHeader)(w)
	}
}

func logoutUser(w http.ResponseWriter, r *http.Request, claims JwtClaims) {
	// Token is supposed to be here and valid
	tokenString := r.Header["Authorization"][0][7:]
	invalidateToken(tokenString, tokenStatusLogout)

	w.WriteHeader(http.StatusNoContent)
}

// RegisterUser create a new user.
//
// Registration is based on a single email address to which a confirmation email
// will be sent to. With the provided token, user can set up a password
func registerUser(w http.ResponseWriter, r *http.Request) {
	var registeringUser RegisteringUser
	json.NewDecoder(r.Body).Decode(&registeringUser)

	// TODO better email check
	if registeringUser.Email == "" {
		hasNoValidEmail.Write(w, r)
		return
	}

	// Prepare a proper User
	toCreateUser, err := registeringUser.prepareForCreation()
	if err != nil {
		HandleServerError(w, r, err)
		return
	}

	// Email unicity is checked by DAO
	createdUser, errMsg := createUser(toCreateUser)
	if errMsg != nil {
		errMsg.Write(w, r)
		return
	}

	// Email: Password setup url
	subject := fmt.Sprintf("Welcome to %s", clientDomain)
	destURL := fmt.Sprintf("%s/user/password/?t=%s",
		clientDomain, createdUser.PwdResetToken.Token)

	// No go-routine: wait for email being sent before answering the client
	utils.SendNoReplyEmail(
		[]string{toCreateUser.Email},
		subject,
		"user_registration",
		struct{ URL string }{URL: destURL},
	)

	w.WriteHeader(http.StatusNoContent)
}

func handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var pwdChgRequest pwdChangeRequest
	json.NewDecoder(r.Body).Decode(&pwdChgRequest)

	authUser, err := changePassword(pwdChgRequest)
	if err != nil {
		err.Write(w, r)
		return
	}

	coreLogger.Verbose("Password updated for %+v", authUser)

	w.WriteHeader(http.StatusNoContent)
}

// GetUser fetch some user info. Password should be omitted
func handleGetUser(w http.ResponseWriter, r *http.Request, claims JwtClaims) {
	userID := GetVar(r, "userId")
	user, err := findUserByID(userID)
	if err != nil {
		coreLogger.Debug("[User] findByID error: ", err)
	}

	json.NewEncoder(w).Encode(user)
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request, claims JwtClaims) {
	var updatingUser User
	json.NewDecoder(r.Body).Decode(&updatingUser)

	userID := GetVar(r, "userId")
	result, err := updateUser(userID, updatingUser)
	if err != nil {
		HandleServerError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request, claims JwtClaims) {
	userID := GetVar(r, "userId")
	deleteUser(userID)

	w.WriteHeader(http.StatusNoContent)
}
