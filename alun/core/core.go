// Package core is the essential package required by all other packages. While
// all packages are, as much as possible, independent from each other, all of
// them depend on the core package.
//
// Loggers are not in the core package only for "Separation of concerns" +
// "Code like a library". Exposing the logging library should not bring all
// the API related code with it.
package core

import (
	"os"

	"github.com/Al-un/alun-api/alun/utils"
	"github.com/Al-un/alun-api/pkg/logger"
)

var coreLogger = logger.NewConsoleLogger(logger.LogLevelVerbose)

var (
	clientDomain string
)

func init() {
	clientDomain = os.Getenv(utils.EnvVarClientDomain)
}
