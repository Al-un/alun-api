package core

import "net/http"

import "time"

// AddCorsHeaders set the usual CORS headers.
//
// It is a special middleware requiring parameters so it cannot use the
// standard adapter pattern
func AddCorsHeaders(next http.Handler, corsConfig CorsConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		// Next
		next.ServeHTTP(w, r)
	})
}

// AddJSONHeaders add the required header for accepting and providing JSON
func AddJSONHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// JSON
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Accept", "application/json")

		// Next
		next.ServeHTTP(w, r)
	})
}

// LoggerInOutRequest displays information for inbound request and outbound result
func LoggerInOutRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		elapsed := time.Since(start)
		coreLogger.Info("Request to %s handled in %v", r.URL.Path, elapsed)
	})
}
