// Package core is the essential package required by all other packages. While
// all packages are, as much as possible, independent from each other, all of
// them depend on the core package.
//
// Loggers are not in the core package only for "Separation of concerns" +
// "Code like a library". Exposing the logging library should not bring all
// the API related code with it.
package core

import (
	"github.com/Al-un/alun-api/pkg/logger"
)

var (
	coreLogger logger.Logger = logger.NewConsoleLogger(logger.LogLevelVerbose)

	// ClientDomain refers to the expected domain of the client application
	ClientDomain string
	// JwtSecretKey is used ONLY for signing JWT
	jwtSecretKey string
)

const (
	defaultJwtSecret string = "1f6797e3545d8d4d4b3ddd8792224e85344be25bd7aa5b8ab63ea72a4186b03f"

	// APIv1 is the standardisation for first version of an API endpoint
	APIv1 string = "v1"
	// APIv2 is the standardisation for second version of an API endpoint
	APIv2 string = "v2"
	// APIMonolithic to enable monolithic mode
	APIMonolithic = true
	// APIMicroservice to enable microservice mode
	APIMicroservice = false
)

func init() {
	jwtSecretKey = defaultJwtSecret
}
