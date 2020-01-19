package utils

// List all environment variable names here to have a centralized list
const (
	// === DB
	EnvVarMongoDbURI = "MONGODB_URI"
	// === Email
	EnvVarEmailUsername = "ALUN_EMAIL_USERNAME"
	EnvVarEmailPassword = "ALUN_EMAIL_PASSWORD"
	EnvVarEmailHost     = "ALUN_EMAIL_HOST"
	EnvVarEmailPort     = "ALUN_EMAIL_PORT"
	EnvVarEmailSender   = "ALUN_EMAIL_SENDER"
	// === Client / Server
	EnvVarServerPort   = "ALUN_SERVER_PORT"
	EnvVarClientDomain = "ALUN_CLIENT_DOMAIN"
)
