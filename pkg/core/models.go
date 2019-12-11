package core

// CorsConfig allows a flexible way to handle CORS stuff
type CorsConfig struct {
	Hosts   string
	Methods string
	Headers string
}

// ErrorMsg is a generic error message ready to be json-ed
type ErrorMsg struct {
	Error string `json:"error"`
}

// CheckStatus encapsulates a status code as well as an Error message
// to send as a default response
type CheckStatus struct {
	Code     int
	ErrorMsg *ErrorMsg
}
