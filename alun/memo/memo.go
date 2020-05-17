// Package memo is about doing a simple TODO list-like task
package memo

import (
	"github.com/Al-un/alun-api/alun/utils"
	"github.com/Al-un/alun-api/pkg/logger"
)

var (
	memoLogger logger.Logger
)

func init() {
	if utils.IsTest() {
		memoLogger = logger.NewSilenceLogger()
	}

	// --- Init logger
	if memoLogger == nil {
		memoLogger = logger.NewConsoleLogger(logger.LogLevelVerbose)
	}

	// --- Init DAO
	initDao()

	// ---- Init API
	initAPI()
}
