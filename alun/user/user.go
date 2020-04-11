// Package user handles the user management
//
// Scope includes:
//	- User creation and edition
//	- Authentication
package user

import (
	"github.com/Al-un/alun-api/pkg/logger"
)

var (
	userLogger = logger.NewConsoleLogger(logger.LogLevelVerbose)
	// pwdSecretSalt is used ONLY as a salt for hashing password
	pwdSecretSalt string
)

const (
	defaultPwdSalt string = "6acaa86d5e15e3df48b4eeb11dcd5c07aab709b2124424ff790304fe94b0cb2f"
)

func init() {
	pwdSecretSalt = defaultPwdSalt
}
