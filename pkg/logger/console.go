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
		cl.printf(format, args...)
	}
}

// Debug prints logs in the console
func (cl *ConsoleLogger) Debug(format string, args ...interface{}) {
	if cl.level.WillLog(LogLevelDebug) {
		cl.printf(format, args...)
	}
}

// Info prints logs in the console
func (cl *ConsoleLogger) Info(format string, args ...interface{}) {
	if cl.level.WillLog(LogLevelInfo) {
		cl.printf(format, args...)
	}
}

// Warn prints logs in the console
func (cl *ConsoleLogger) Warn(format string, args ...interface{}) {
	if cl.level.WillLog(LogLevelWarn) {
		cl.printf(format, args...)
	}
}

// Error prints logs in the console
func (cl *ConsoleLogger) Error(format string, args ...interface{}) {
	if cl.level.WillLog(LogLevelError) {
		cl.printf(format, args...)
	}
}

// Fatal is MAYDAY MAYDAY
func (cl *ConsoleLogger) Fatal(exitCode int, format string, args ...interface{}) {
	cl.printf(format, args...)

	os.Exit(exitCode)
}

func (cl *ConsoleLogger) printf(format string, args ...interface{}) {
	text := fmt.Sprintf("[%s] %s\n", cl.level.name, format)

	if len(args) > 0 {
		log.Printf(text, args...)
	} else {
		log.Printf(text)
	}
}
