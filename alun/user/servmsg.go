package user

import (
	"net/http"

	"github.com/Al-un/alun-api/alun/core"
)

// ----------------------------------------------------------------------------
//	User management: Code 102xx
// ----------------------------------------------------------------------------

var hasNoValidEmail = &core.ServiceMessage{
	Code:       10200,
	HTTPStatus: http.StatusBadRequest,
	Message:    "Email is not valid",
}

var hasEmailNotAvailable = &core.ServiceMessage{
	Code:       10201,
	HTTPStatus: http.StatusBadRequest,
	Message:    "Email is already taken",
}

var hasNoEmail = &core.ServiceMessage{
	Code:       10202,
	HTTPStatus: http.StatusBadRequest,
	Message:    "Email is missing",
}

var pwdResetTokenNotFound = &core.ServiceMessage{
	Code:       10203,
	HTTPStatus: http.StatusNotFound,
	Message:    "Password reset token not found",
}

var pwdResetTokenExpired = &core.ServiceMessage{
	Code:       10204,
	HTTPStatus: http.StatusBadRequest,
	Message:    "Password reset token is expired",
}

var isEmailNotFound = &core.ServiceMessage{
	Code:       10205,
	HTTPStatus: http.StatusNotFound,
	Message:    "Email is not found",
}
