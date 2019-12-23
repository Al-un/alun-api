package logger

import (
	"fmt"
	"log"
	"os"
)

// ConsoleLogger is the standard STDOUT logger
type ConsoleLogger struct {
	level LogLevel
}

// NewConsoleLogger is the ConsoleLogger constructor
func NewConsoleLogger(level LogLevel) *ConsoleLogger {
	newLogger := &ConsoleLogger{
		level: level,
	}

	return newLogger
}

// LogLevel returns logger current logging level
func (cl *ConsoleLogger) LogLevel() LogLevel {
	return cl.level
}

// Verbose prints logs in the console
func (cl *ConsoleLogger) Verbose(format string, args ...interface{}) {
	if cl.level.WillLog(LogLevelVerbose) {
		cl.printf(LogLevelVerbose, format, args...)
	}
}

// Debug prints logs in the console
func (cl *ConsoleLogger) Debug(format string, args ...interface{}) {
	if cl.level.WillLog(LogLevelDebug) {
		cl.printf(LogLevelDebug, format, args...)
	}
}

// Info prints logs in the console
func (cl *ConsoleLogger) Info(format string, args ...interface{}) {
	if cl.level.WillLog(LogLevelInfo) {
		cl.printf(LogLevelInfo, format, args...)
	}
}

// Warn prints logs in the console
func (cl *ConsoleLogger) Warn(format string, args ...interface{}) {
	if cl.level.WillLog(LogLevelWarn) {
		cl.printf(LogLevelWarn, format, args...)
	}
}

// Error prints logs in the console
func (cl *ConsoleLogger) Error(format string, args ...interface{}) {
	if cl.level.WillLog(LogLevelError) {
		cl.printf(LogLevelError, format, args...)
	}
}

// Fatal is MAYDAY MAYDAY
//
// Fatal mimics the behaviour of `log.Fatalf` with a custom exit code
func (cl *ConsoleLogger) Fatal(exitCode int, format string, args ...interface{}) {
	cl.printf(LogLevelFatal, format, args...)

	os.Exit(exitCode)
}

// Local helper specific to console logging
func (cl *ConsoleLogger) printf(level LogLevel, format string, args ...interface{}) {
	text := fmt.Sprintf("[%s] %s\n", level.name, format)

	if len(args) > 0 {
		log.Printf(text, args...)
	} else {
		log.Printf(text)
	}
}
