package memo

import (
	"net/http"

	"github.com/Al-un/alun-api/alun/core"
)

// MemoAPI for Memo
var MemoAPI *core.API

// isAdminOrOwnUser checks access for user modification: only self data
// are allowed
var isAdminOrOwnUser = func(r *http.Request, jwtClaims core.JwtClaims) bool {
	if jwtClaims.IsAdmin {
		return true
	}

	userID := core.GetVar(r, "userId")
	return userID == jwtClaims.UserID
}

func init() {
	apiRoot := "memos"
	MemoAPI = core.NewAPI(apiRoot)
	MemoAPI.AddMiddleware(core.AddJSONHeaders)

	MemoAPI.AddProtectedEndpoint("", "GET", core.APIv1, core.CheckIfLogged, handleListMemo)
	MemoAPI.AddProtectedEndpoint("", "POST", core.APIv1, core.CheckIfLogged, handleCreateMemo)
	MemoAPI.AddProtectedEndpoint("{memoId}", "GET", core.APIv1, core.CheckIfLogged, handleGetMemo)
	MemoAPI.AddProtectedEndpoint("{memoId}", "PUT", core.APIv1, core.CheckIfLogged, handleUpdateMemo)
	MemoAPI.AddProtectedEndpoint("{memoId}", "DELETE", core.APIv1, core.CheckIfLogged, handleDeleteMemo)
}
