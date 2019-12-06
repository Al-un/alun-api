package core

import "net/http"

// AddCommonHeaders set the CORS header and JSON headers
func AddCommonHeaders(w http.ResponseWriter, methods string) {
	corsAllowedHosts := "*"
	corsAllowedHeaders := "*"

	// CORS
	w.Header().Set("Access-Control-Allow-Origin", corsAllowedHosts)
	w.Header().Set("Access-Control-Allow-Methods", methods)
	w.Header().Set("Access-Control-Allow-Headers", corsAllowedHeaders)

	// JSON
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Accept", "application/json")
}
