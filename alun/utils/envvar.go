package utils

import (
	"os"
	"strings"
)

// List all environment variable names here to have a centralized list
const (
	// === Server
	EnvVarServerPort         = "ALUN_SERVER_MONOLITHIC_PORT"
	EnvVarServerIsMonolithic = "ALUN_SERVER_IS_MONOLITHIC"
	EnvVarMode               = "ALUN_MODE"
	// === Application: User
	EnvVarUserPort    = "ALUN_USER_PORT"
	EnvVarUserDbURL   = "ALUN_USER_DATABASE_URL"
	EnvVarUserSaltPwd = "ALUN_SECRET_PWD"
	EnvVarUserSaltJwt = "ALUN_SECRET_JWT"
	// === Application: Memo
	EnvVarMemoPort  = "ALUN_MEMO_PORT"
	EnvVarMemoDbURL = "ALUN_MEMO_DATABASE_URL"
	// === Email
	EnvVarEmailUsername = "ALUN_EMAIL_USERNAME"
	EnvVarEmailPassword = "ALUN_EMAIL_PASSWORD"
	EnvVarEmailHost     = "ALUN_EMAIL_HOST"
	EnvVarEmailPort     = "ALUN_EMAIL_PORT"
	EnvVarEmailSender   = "ALUN_EMAIL_SENDER"
)

// List all running mode of the applications
const (
	AlunModeDev  = "Development"
	AlunModeTest = "Test"
	AlunModeProd = "Production"
)

// IsDev returns true if the environment variables define a Development mode
func IsDev() bool {
	currMode := os.Getenv(EnvVarMode)
	return strings.ToLower(currMode) == strings.ToLower(AlunModeDev)
}

// IsTest returns true if the environment variables define a test mode
func IsTest() bool {
	currMode := os.Getenv(EnvVarMode)
	return strings.ToLower(currMode) == strings.ToLower(AlunModeTest)
}

// IsProd returns true if the environment variables define a production mode
func IsProd() bool {
	currMode := os.Getenv(EnvVarMode)
	return strings.ToLower(currMode) == strings.ToLower(AlunModeProd)
}
