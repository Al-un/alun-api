package utils

// List all environment variable names here to have a centralized list
const (
	// === Server
	EnvVarServerPort         = "ALUN_SERVER_MONOLITHIC_PORT"
	EnvVarServerIsMonolithic = "ALUN_SERVER_IS_MONOLITHIC"
	// === Misc
	EnvVarClientDomain = "ALUN_CLIENT_DOMAIN"
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
