// Package utils gathers various useful stuff
//
// All listing or list-like information, usually cross-packages, are all gathered
// in this package in order to have a centralized location for information such as
// - Error codes list
// - Email account configuration and templates
// - Environment variables name
package utils

import (
	"github.com/Al-un/alun-api/pkg/logger"
)

var utilsLogger = logger.NewConsoleLogger(logger.LogLevelVerbose)
