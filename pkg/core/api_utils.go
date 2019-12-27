package core

import (
	"net/http"

	"github.com/gorilla/mux"
)

// GetVar fetch the variable defined in the route.
//
// Such method can be framework-dependent.
func GetVar(r *http.Request, varName string) string {
	return mux.Vars(r)[varName]
}

// HandleServerError is the generic way to handle server error: just send
// a 500 and the message with it
func HandleServerError(w http.ResponseWriter, r *http.Request, err error) {
	servMsg := ServiceMessage{
		HTTPStatus: 500,
		Error:      err,
	}

	servMsg.Write(w, r)
}
