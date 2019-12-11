package core

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
