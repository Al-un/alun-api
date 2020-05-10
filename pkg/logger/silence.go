package logger

import (
	"os"
)

// SilenceLogger is the standard STDOUT logger
type SilenceLogger struct {
}

// NewSilenceLogger is the SilenceLogger constructor
func NewSilenceLogger() *SilenceLogger {
	return &SilenceLogger{}
}

// LogLevel returns logger current logging level
func (cl *SilenceLogger) LogLevel() LogLevel {
	return LogLevelFatal
}

// Verbose prints logs in the console
func (cl *SilenceLogger) Verbose(format string, args ...interface{}) {

}

// Debug prints logs in the console
func (cl *SilenceLogger) Debug(format string, args ...interface{}) {

}

// Info prints logs in the console
func (cl *SilenceLogger) Info(format string, args ...interface{}) {

}

// Warn prints logs in the console
func (cl *SilenceLogger) Warn(format string, args ...interface{}) {

}

// Error prints logs in the console
func (cl *SilenceLogger) Error(format string, args ...interface{}) {
}

// Fatal is MAYDAY MAYDAY
//
// Fatal mimics the behaviour of `log.Fatalf` with a custom exit code
func (cl *SilenceLogger) Fatal(exitCode int, format string, args ...interface{}) {
	os.Exit(exitCode)
}
