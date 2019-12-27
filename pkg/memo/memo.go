// Package memo is about doing a simple TODO list-like task
package memo

import (
	"github.com/Al-un/alun-api/pkg/core"
	"github.com/Al-un/alun-api/pkg/logger"
)

var memoLogger = logger.NewConsoleLogger(logger.LogLevelVerbose)

// MemoAPI for Memo
var MemoAPI *core.API

func init() {
	MemoAPI = core.NewAPI("memos")
	MemoAPI.AddMiddleware(core.AddJSONHeaders)

	MemoAPI.AddProtectedEndpoint("", "GET", core.APIv1, core.CheckIfLogged, handleListMemo)
	MemoAPI.AddProtectedEndpoint("", "POST", core.APIv1, core.CheckIfLogged, handleCreateMemo)
	MemoAPI.AddProtectedEndpoint("{memoId}", "GET", core.APIv1, core.CheckIfLogged, handleGetMemo)
	MemoAPI.AddProtectedEndpoint("{memoId}", "PUT", core.APIv1, core.CheckIfLogged, handleUpdateMemo)
	MemoAPI.AddProtectedEndpoint("{memoId}", "DELETE", core.APIv1, core.CheckIfLogged, handleDeleteMemo)
}
