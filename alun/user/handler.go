package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Al-un/alun-api/alun/core"
	"github.com/Al-un/alun-api/alun/utils"
)

// AuthUser authenticates user.
//
// Accepted authentication methods are:
//	- JSON based
//	- BASIC authentication
func authUser(w http.ResponseWriter, r *http.Request) {

	// --- JSON-based authentication
	var user authenticatedUser
	json.NewDecoder(r.Body).Decode(&user)
	if user.Email != "" && user.Password != "" {
		userLogger.Verbose("JSON authentication: %s/%s", user.Email, user.Password)
		authenticateCredentials(user.Email, user.Password)(w)
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

// LogoutUser invalidates the provided token. Request is assumed to be properly formed
func logoutUser(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
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
		hasNoEmail.Write(w, r)
		return
	}

	// Prepare a proper User
	toCreateUser, err := registeringUser.prepareForCreation()
	if err != nil {
		core.HandleServerError(w, r, err)
		return
	}

	// Email unicity is checked by DAO
	createdUser, errMsg := createUser(toCreateUser)
	if errMsg != nil {
		errMsg.Write(w, r)
		return
	}

	// Email: Password setup url
	subject := fmt.Sprintf("Welcome to %s", core.ClientDomain)
	destURL := fmt.Sprintf("%s/user/password/?t=%s",
		core.ClientDomain, createdUser.PwdResetToken.Token)

	// No go-routine: wait for email being sent before answering the client
	utils.SendNoReplyEmail(
		[]string{toCreateUser.Email},
		subject,
		utils.EmailTemplateUserRegistration,
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

	userLogger.Verbose("Password updated for %+v", authUser)

	w.WriteHeader(http.StatusNoContent)
}

func handleRequestPasswordChange(w http.ResponseWriter, r *http.Request) {
	var pwdChgRequest pwdChangeRequest
	json.NewDecoder(r.Body).Decode(&pwdChgRequest)

	w.WriteHeader(http.StatusNoContent)
}

// GetUser fetch some user info. Password should be omitted
func handleGetUser(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	userID := core.GetVar(r, "userId")
	user, err := findUserByID(userID)
	if err != nil {
		userLogger.Debug("[User] findByID error: ", err)
	}

	json.NewEncoder(w).Encode(user)
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	var updatingUser User
	json.NewDecoder(r.Body).Decode(&updatingUser)

	userID := core.GetVar(r, "userId")
	result, err := updateUser(userID, updatingUser)
	if err != nil {
		core.HandleServerError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	userID := core.GetVar(r, "userId")
	deleteUser(userID)

	w.WriteHeader(http.StatusNoContent)
}