package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// authUser authenticates user with BASIC methods or other
func authUser(w http.ResponseWriter, r *http.Request) {
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

	if authHeader[:5] == "Basic" {
		authenticateBasic(authHeader)(w)
	}
}

// RegisterUser create a new user. Username does not have to be unique
func registerUser(w http.ResponseWriter, r *http.Request) {
	var creatingUser authenticatedUser
	json.NewDecoder(r.Body).Decode(&creatingUser)

	coreLogger.Verbose("Registering user %v", creatingUser)

	createdUser := createUser(creatingUser)
	http.Redirect(w, r, fmt.Sprintf("/users/details/%v", createdUser.InsertedID), http.StatusCreated)
}

// GetUser fetch some user info. Password should be omitted
func handleGetUser(w http.ResponseWriter, r *http.Request, claims JwtClaims) {
	userID := mux.Vars(r)["userId"]
	user, err := findUserByID(userID)
	if err != nil {
		coreLogger.Debug("[User] findByID error: ", err)
	}

	json.NewEncoder(w).Encode(user)
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request, claims JwtClaims) {
	var updatingUser User
	json.NewDecoder(r.Body).Decode(&updatingUser)

	userID := mux.Vars(r)["userId"]
	result := updateUser(userID, updatingUser)

	json.NewEncoder(w).Encode(result)
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request, claims JwtClaims) {
	userID := mux.Vars(r)["userId"]
	deleteUser(userID)

	w.WriteHeader(http.StatusNoContent)
}
