// Package logger gathers all possible type of logging, from console logging
// to remote log management
package logger

// Logger defines logging behaviour.
//
// A logger is expected to:
//	- have multiple levels of logging depending on the severity of the logged
//	  message
//	- has a stand-alone configuration, not bound to some API definition or
//	  other loggers
//	- Automatically handle new line characters whenever required
type Logger interface {
	// Verbose is mainly used during develop, well, flooding !
	Verbose(format string, args ...interface{})

	// Debug is expected to be relevant only for development environment
	Debug(format string, args ...interface{})

	// Standard logging level
	Info(format string, args ...interface{})

	// Something went wrong but the involved process could finish
	Warn(format string, args ...interface{})

	// Something went wrong but the involved process could NOT finish
	Error(format string, args ...interface{})
	// The app crashed
	Fatal(exitCode int, format string, args ...interface{})

	// Logger logging level is a read-only property
	LogLevel() LogLevel
}

// LogLevel defines the minimum logging for a specific logger
type LogLevel struct {
	// User-friendly logging level
	name string
	// Logging level to
	order int
}

// WillLog checks the current logger has a sufficient logging level to continue
func (lvl *LogLevel) WillLog(targetLevel LogLevel) bool {
	return lvl.order <= targetLevel.order
}

// LogLevelVerbose to say everything
var LogLevelVerbose = LogLevel{
	name:  "VERBOSE",
	order: 1,
}

// LogLevelDebug for development
var LogLevelDebug = LogLevel{
	name:  "DEBUG",
	order: 3,
}

// LogLevelInfo for standard logging
var LogLevelInfo = LogLevel{
	name:  "INFO",
	order: 6,
}

// LogLevelWarn is...starting to get problems
var LogLevelWarn = LogLevel{
	name:  "WARN",
	order: 10,
}

// LogLevelError now that's a problem
var LogLevelError = LogLevel{
	name:  "ERROR",
	order: 15,
}
