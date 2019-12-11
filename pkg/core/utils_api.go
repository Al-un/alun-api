package core

import "net/http"

// AddCommonHeaders set the CORS header and JSON headers. If "methods" argument is an empty string,
// it will default to "*"
func AddCommonHeaders(w http.ResponseWriter, corsConfig CorsConfig) {
	corsAllowedHosts := "*"
	corsAllowedHeaders := "*"
	corsAllowedMethods := "*"
	if len(corsConfig.Methods) != 0 {
		corsAllowedMethods = corsConfig.Methods
	}
	if len(corsConfig.Hosts) != 0 {
		corsAllowedHosts = corsConfig.Hosts
	}
	if len(corsConfig.Headers) != 0 {
		corsAllowedHeaders = corsConfig.Headers
	}

	// CORS
	w.Header().Set("Access-Control-Allow-Origin", corsAllowedHosts)
	w.Header().Set("Access-Control-Allow-Methods", corsAllowedMethods)
	w.Header().Set("Access-Control-Allow-Headers", corsAllowedHeaders)

	// JSON
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Accept", "application/json")
}
