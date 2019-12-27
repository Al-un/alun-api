package core

import (
	"encoding/json"
	"net/http"
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

// RegisterUser create a new user. Username does not have to be unique
func registerUser(w http.ResponseWriter, r *http.Request) {
	var creatingUser authenticatedUser
	json.NewDecoder(r.Body).Decode(&creatingUser)

	coreLogger.Verbose("Registering user %v", creatingUser)

	usernameAlreadyTaken, err := isUsernameAlreadyRegistered(creatingUser.Username)
	if usernameAlreadyTaken {
		hasUsernameAlreadyTaken.Write(w, r)
		return
	}

	createdUser, err := createUser(creatingUser)
	if err != nil {
		HandleServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
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
