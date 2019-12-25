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

	AuthAPI.AddPublicEndpoint("login", "POST", "v1", authUser)
	AuthAPI.AddProtectedEndpoint("logout", "POST", "v1", CheckIfLogged, logoutUser)
	AuthAPI.AddPublicEndpoint("register", "POST", "v1", registerUser)
	AuthAPI.AddProtectedEndpoint("detail/{userId}", "GET", "v1", isAdminOrOwnUser, handleGetUser)
	AuthAPI.AddProtectedEndpoint("detail/{userId}", "PUT", "v1", isAdminOrOwnUser, handleUpdateUser)
	AuthAPI.AddProtectedEndpoint("detail/{userId}", "DELETE", "v1", isAdminOrOwnUser, handleDeleteUser)
}
