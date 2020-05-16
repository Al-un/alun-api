package testutils

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

const (
	// CallFromTestFile to use when Assert() is used in a test file
	CallFromTestFile = 2
	// CallFromHelperMethod to user when Asset() is used from a helper method
	CallFromHelperMethod = 3
)

// copied from https://github.com/benbjohnson/testing

// Assert fails the test if the condition is false.
func Assert(tb testing.TB, callerSkip int, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(callerSkip)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// Ok fails the test if an err is not nil.
func Ok(tb testing.TB, callerSkip int, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(callerSkip)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// Equals fails the test if exp is not equal to act.
func Equals(tb testing.TB, callerSkip int, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(callerSkip - 1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
