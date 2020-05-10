// Package user handles the user management
//
// Scope includes:
//	- User creation and edition
//	- Authentication
package user

import (
	"os"

	"github.com/Al-un/alun-api/alun/utils"
	"github.com/Al-un/alun-api/pkg/logger"
)

var (
	userLogger    logger.Logger
	pwdSecretSalt string // pwdSecretSalt is used ONLY as a salt for hashing password
	alunEmail     utils.AlunEmailSender
)

const (
	defaultPwdSalt string = "6acaa86d5e15e3df48b4eeb11dcd5c07aab709b2124424ff790304fe94b0cb2f"
)

func init() {
	if utils.IsTest() {
		userLogger = logger.NewSilenceLogger()
		alunEmail = utils.GetDummyEmail()
	}

	// --- Init logger
	if userLogger == nil {
		userLogger = logger.NewConsoleLogger(logger.LogLevelVerbose)
	}

	// --- Init salts
	pwdSecretSalt = os.Getenv(utils.EnvVarUserSaltPwd)
	if pwdSecretSalt == "" {
		pwdSecretSalt = defaultPwdSalt
	}

	// --- Init Email
	if alunEmail == nil {
		utils.GetAlunEmail()
	}

	// --- Init DAO
	initDao()

	// ---- Init API
	initAPI()
}
