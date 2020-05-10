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

// handleRequestPassword handles a password request which can be for a new user
// (user creation) or an existing user (password reset)
//
// User is then sent the appropriate email
func handleRequestPassword(w http.ResponseWriter, r *http.Request) {
	var pwdReq PasswordRequest
	json.NewDecoder(r.Body).Decode(&pwdReq)

	// TODO better email check
	userLogger.Verbose("Got password reset %+v", pwdReq)
	if pwdReq.Email == "" {
		hasNoEmail.Write(w, r)
		return
	}

	// Generate token
	toHandleUser, err := pwdReq.createPwdResetToken()
	if err != nil {
		core.HandleServerError(w, r, err)
		return
	}
	userLogger.Verbose("Got PwdResetToken %+v", toHandleUser)

	if pwdReq.RequestType == userPwdRequestNewUser {
		doCreateUser(w, r, toHandleUser, &pwdReq)
	} else if pwdReq.RequestType == userPwdRequestPwdReset {
		doResetPassword(w, r, toHandleUser, &pwdReq)
	}
}

func doCreateUser(w http.ResponseWriter, r *http.Request, newUser *User, pwdReq *PasswordRequest) {
	// Email unicity is checked by DAO
	createdUser, errMsg := createUser(*newUser)
	if errMsg != nil {
		errMsg.Write(w, r)
		return
	}

	// Email: Password setup url
	subject := "Welcome to Al-un.fr"
	destURL := fmt.Sprintf("%s%s",
		pwdReq.RedirectURL, createdUser.PwdResetToken.Token)

	// No go-routine: wait for email being sent before answering the client
	alunEmail.SendNoReplyEmail(
		[]string{createdUser.Email},
		subject,
		utils.EmailTemplateUserRegistration,
		struct{ URL string }{URL: destURL},
	)

	w.WriteHeader(http.StatusNoContent)
}

func doResetPassword(w http.ResponseWriter, r *http.Request, user *User, pwdReq *PasswordRequest) {
	// Email unicity is checked by DAO
	errMsg := updatePwdResetToken(user)
	if errMsg != nil {
		errMsg.Write(w, r)
		return
	}

	// Email: Password setup url
	subject := "Password reset Al-un.fr"
	destURL := fmt.Sprintf("%s%s",
		pwdReq.RedirectURL, user.PwdResetToken.Token)

	// TODO: update email
	alunEmail.SendNoReplyEmail(
		[]string{user.Email},
		subject,
		utils.EmailTemplateUserPwdReset,
		struct{ URL string }{URL: destURL},
	)

	w.WriteHeader(http.StatusNoContent)
}

func handleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	var pwdChgRequest pwdChangeRequest
	json.NewDecoder(r.Body).Decode(&pwdChgRequest)

	authUser, err := updatePassword(pwdChgRequest)
	if err != nil {
		err.Write(w, r)
		return
	}

	userLogger.Verbose("Password updated for %+v", authUser)

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
