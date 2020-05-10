package user

import (
	"net/http"

	"github.com/Al-un/alun-api/alun/core"
)

// UserAPI exposes User endpoints
var UserAPI *core.API

// isAdminOrOwnUser checks access for user modification: only self data
// are allowed
var isAdminOrOwnUser = func(r *http.Request, jwtClaims core.JwtClaims) bool {
	if jwtClaims.IsAdmin {
		return true
	}

	userID := core.GetVar(r, "userId")
	return userID == jwtClaims.UserID
}

func initAPI() {
	apiRoot := "users"
	UserAPI = core.NewAPI(apiRoot)
	UserAPI.AddMiddleware(core.AddJSONHeaders)

	UserAPI.AddPublicEndpoint("login", "POST", core.APIv1, authUser)
	UserAPI.AddProtectedEndpoint("logout", "POST", core.APIv1, core.CheckIfLogged, logoutUser)
	UserAPI.AddPublicEndpoint("register", "POST", core.APIv1, handleRequestPassword)
	UserAPI.AddPublicEndpoint("password/update", "POST", core.APIv1, handleUpdatePassword)
	UserAPI.AddPublicEndpoint("password/request", "POST", core.APIv1, handleRequestPassword)
	UserAPI.AddProtectedEndpoint("detail/{userId}", "GET", core.APIv1, isAdminOrOwnUser, handleGetUser)
	UserAPI.AddProtectedEndpoint("detail/{userId}", "PUT", core.APIv1, isAdminOrOwnUser, handleUpdateUser)
	UserAPI.AddProtectedEndpoint("detail/{userId}", "DELETE", core.APIv1, isAdminOrOwnUser, handleDeleteUser)
}
