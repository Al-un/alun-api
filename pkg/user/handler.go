package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Al-un/alun-api/pkg/core"
	"github.com/gorilla/mux"
)

// authUser authenticates user with BASIC methods or other
func authUser(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received headers %v\n", r.Header)

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
	core.AddCommonHeaders(w, "POST")

	var creatingUser User
	json.NewDecoder(r.Body).Decode(&creatingUser)

	createdUser := createUser(creatingUser)
	http.Redirect(w, r, fmt.Sprintf("/users/details/%v", createdUser.InsertedID), http.StatusCreated)
}

// GetUser fetch some user info. Password should be omitted
func handleGetUser(w http.ResponseWriter, r *http.Request) {
	core.AddCommonHeaders(w, "GET")

	userID := mux.Vars(r)["id"]
	user, err := findUserByID(userID)
	if err != nil {
		fmt.Println("[User] findByID error: ", err)
	}

	json.NewEncoder(w).Encode(user)
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	core.AddCommonHeaders(w, "PUT")

	var updatingUser User
	json.NewDecoder(r.Body).Decode(&updatingUser)

	userID := mux.Vars(r)["id"]
	result := updateUser(userID, updatingUser)

	json.NewEncoder(w).Encode(result)
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	core.AddCommonHeaders(w, "DELETE")

	userID := mux.Vars(r)["id"]
	deleteUser(userID)

	w.WriteHeader(http.StatusNoContent)
}
