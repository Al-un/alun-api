package core

import (
	"net/http"

	"github.com/gorilla/mux"
)

// AuthAPI is the user and authentication related API
var AuthAPI *API

// isAdminOrOwnUser checks access for user modification: only self data
// are allowed
var isAdminOrOwnUser = func(r *http.Request, jwtClaims JwtClaims) bool {
	if jwtClaims.IsAdmin {
		return true
	}

	userID := mux.Vars(r)["userId"]
	return userID == jwtClaims.UserID
}

func init() {
	AuthAPI = NewAPI("users")

	AuthAPI.AddPublicEndpoint("login", "POST", APIv1, authUser)
	AuthAPI.AddProtectedEndpoint("logout", "POST", APIv1, CheckIfLogged, logoutUser)
	AuthAPI.AddPublicEndpoint("register", "POST", APIv1, registerUser)
	AuthAPI.AddProtectedEndpoint("detail/{userId}", "GET", APIv1, isAdminOrOwnUser, handleGetUser)
	AuthAPI.AddProtectedEndpoint("detail/{userId}", "PUT", APIv1, isAdminOrOwnUser, handleUpdateUser)
	AuthAPI.AddProtectedEndpoint("detail/{userId}", "DELETE", APIv1, isAdminOrOwnUser, handleDeleteUser)
}
