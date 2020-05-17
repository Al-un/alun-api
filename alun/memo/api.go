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

func initAPI() {
	apiRoot := "memos"
	MemoAPI = core.NewAPI(apiRoot, memoLogger)
	MemoAPI.AddMiddleware(core.AddJSONHeaders)

	MemoAPI.AddProtectedEndpoint("/boards", http.MethodGet, core.APIv1, core.CheckIfLogged, handleListBoards)
	MemoAPI.AddProtectedEndpoint("/boards", http.MethodPost, core.APIv1, core.CheckIfLogged, handleCreateBoard)
	MemoAPI.AddProtectedEndpoint("/boards/{boardId}", http.MethodGet, core.APIv1, core.CheckIfLogged, handleGetBoard)
	MemoAPI.AddProtectedEndpoint("/boards/{boardId}", http.MethodPut, core.APIv1, core.CheckIfLogged, handleUpdateBoard)
	MemoAPI.AddProtectedEndpoint("/boards/{boardId}", http.MethodDelete, core.APIv1, core.CheckIfLogged, handleDeleteBoard)
	MemoAPI.AddProtectedEndpoint("/boards/{boardId}/memos", http.MethodPost, core.APIv1, core.CheckIfLogged, handleCreateMemo)
	MemoAPI.AddProtectedEndpoint("/boards/{boardId}/memos", http.MethodDelete, core.APIv1, core.CheckIfLogged, handleDeleteMemo)
	MemoAPI.AddProtectedEndpoint("/memos/{memoId}", http.MethodPut, core.APIv1, core.CheckIfLogged, handleUpdateMemo)
}
