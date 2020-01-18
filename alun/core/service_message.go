package core

import "net/http"

// ----------------------------------------------------------------------------
//	Authentication check: 101xx code
// ----------------------------------------------------------------------------

var isAuthorized = &ServiceMessage{Code: 0, Message: ""}

// Token is all good but nope
var isNotAuthorized = &ServiceMessage{
	Code:       10101,
	HTTPStatus: http.StatusUnauthorized,
	Message:    "Not authorized",
}

// Timing is everything
var isTokenExpired = &ServiceMessage{
	Code:       10102,
	HTTPStatus: http.StatusForbidden,
	Message:    "Token is expired",
}

// What the hell was sent?
var isTokenMalformed = &ServiceMessage{
	Code:       10103,
	HTTPStatus: http.StatusForbidden,
	Message:    "Token is malformed",
}

// Unknown Token parsing error
var isTokenInvalid = &ServiceMessage{
	Code:       10104,
	HTTPStatus: http.StatusForbidden,
	Message:    "Token is just invalid",
}

// Token has been manually invalidated
var isTokenInvalidated = &ServiceMessage{
	Code:       10105,
	HTTPStatus: http.StatusForbidden,
	Message:    "Token has been invalidated",
}

// Token is already logged out
var isTokenLogout = &ServiceMessage{
	Code:       10106,
	HTTPStatus: http.StatusForbidden,
	Message:    "Token is already logged out",
}

// Hey, you need to send something!
var isAuthorizationMissing = &ServiceMessage{
	Code:       10107,
	HTTPStatus: http.StatusForbidden,
	Message:    "Authorization header is missing",
}

// You need to send something correct
var isAuthorizationInvalid = &ServiceMessage{
	Code:       10108,
	HTTPStatus: http.StatusForbidden,
	Message:    "Authorization header must start with \"Bearer \"",
}

// I just don't know
var isUnknownError = &ServiceMessage{
	Code:       10109,
	HTTPStatus: http.StatusForbidden,
	Message:    "Unknown error during Authorization check",
}

// ----------------------------------------------------------------------------
//	User management: Code 102xx
// ----------------------------------------------------------------------------

var hasUsernameAlreadyTaken = &ServiceMessage{
	Code:       10201,
	HTTPStatus: http.StatusBadRequest,
	Message:    "Username is already taken",
}
