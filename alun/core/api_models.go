package core

import (
	"fmt"
	"net/http"

	"github.com/Al-un/alun-api/pkg/logger"
	"github.com/gorilla/mux"
)

// ----------------------------------------------------------------------------
// 	Types
// ----------------------------------------------------------------------------

// EndpointAdapter (or Decorator design pattern) wrapper consecutive middlewares.
//
// EndpointAdapter must use the standard http.Handler method and, when authentication is
// required, another type of Adapter might be required
type EndpointAdapter func(http.Handler) http.Handler

// API exposes the list of handler with a specific URL root
//
// Each APIHandler will be linked to the route URL. URL is defined by an urlBuilder
type API struct {
	root        string            // root for all handlers URL
	corsHosts   string            // authorized hosts
	corsHeaders string            // authorized headers
	middlewares []EndpointAdapter // list of standard HTTP Handler
	endpoints   []APIEndpoint
	urlBuilder  URLBuilder // how endpoints URL are built
	logger      logger.Logger
}

// APIEndpoint maps a handler (authenticated or standard) with an URL pattern
// and a HTTP method
//
// If both "publicHandler" and "protectedHandler" are defined, the public
// version takes precedence
type APIEndpoint struct {
	url              string // url to access the handler
	httpMethod       string // HTTP method, all capitlized for the provided URL
	version          string // Arbitrary version text
	accessChecker    AccessChecker
	handler          http.HandlerFunc     // final handler
	publicHandler    http.HandlerFunc     // public handler input
	protectedHandler AuthenticatedHandler // protected handler input
}

// CorsConfig allows a flexible way to handle CORS stuff
type CorsConfig struct {
	Hosts   string
	Methods string
	Headers string
}

// URLBuilder defines how an API generates a final endpoint URL given an (optional)
// root url, a version and the endpoint url
//
// Root URL is optional for microservices-ready structure
type URLBuilder func(root string, version string, url string) string

// ----------------------------------------------------------------------------
// 	Constructors / Constant
// ----------------------------------------------------------------------------

// NewAPI is the API constructor
//
// - Allowed CORS headers and hosts are, for the moment, "*"
// - By default, the JSON middleware is always added
func NewAPI(root string, logger logger.Logger) *API {
	// init
	api := &API{
		root:        root,
		corsHosts:   "*",
		corsHeaders: "*",
		endpoints:   make([]APIEndpoint, 0), // explicitly define an empty array
		urlBuilder:  URLDefaultBuilder,
		logger:      logger,
	}

	return api
}

// URLDefaultBuilder concatenates "/{version}/{root}/{url}""
var URLDefaultBuilder URLBuilder = func(root string, version string, url string) string {
	if root == "" {
		return fmt.Sprintf("/%s/%s", version, url)
	}

	return fmt.Sprintf("/%s/%s/%s", root, version, url)
}

// ----------------------------------------------------------------------------
// 	Methods
// ----------------------------------------------------------------------------

// AddMiddleware appends a middleware to a specific API.
//
// There is no check about duplicates
func (api *API) AddMiddleware(mw EndpointAdapter) {
	api.middlewares = append(api.middlewares, mw)
}

// addEndpoint appends the provided endpoint in the endpoints list, regardless
// public or protected handlers, if some checks are OK:
// 	- httpMethod is valid
func (api *API) addEndpoint(endpoint APIEndpoint) {
	// httpMethod check
	if endpoint.httpMethod != http.MethodGet &&
		endpoint.httpMethod != http.MethodPost &&
		endpoint.httpMethod != http.MethodPatch &&
		endpoint.httpMethod != http.MethodPut &&
		endpoint.httpMethod != http.MethodDelete {
		api.logger.Fatal(1, "Cannot call 'AddHandler' an invalid method \"%s\" for URL %s/%s/%s",
			endpoint.httpMethod, api.root, endpoint.version, endpoint.url)
	}

	// endpoint handler check
	var handler http.HandlerFunc

	if endpoint.publicHandler != nil {
		// Public handler: leverage ServeHTTP method
		handler = endpoint.publicHandler
	} else if endpoint.protectedHandler != nil {
		// Protected handler
		handler = DoIfAccess(endpoint.accessChecker, endpoint.protectedHandler).ServeHTTP
	} else {
		// Error: missing handler
		api.logger.Fatal(1, "[API] Endpoint %s:%s does not have any handler", endpoint.httpMethod, endpoint.url)
		return
	}

	// CORS config is the same for both public and protected
	corsConfig := CorsConfig{
		Hosts:   api.corsHosts,
		Headers: api.corsHeaders,
		Methods: endpoint.httpMethod,
	}

	// Apply CORS handers
	endpoint.handler = AddCorsHeaders(handler, corsConfig).ServeHTTP

	// Add new endpoints to the list
	api.endpoints = append(api.endpoints, endpoint)
}

// AddProtectedEndpoint appends a handler which requires an AccessControl
//
// AddHandler also checks if the provided httpMethod is valid.
func (api *API) AddProtectedEndpoint(url string, httpMethod string, version string,
	accessChecker AccessChecker, handler AuthenticatedHandler) {

	api.addEndpoint(APIEndpoint{
		url:              url,
		httpMethod:       httpMethod,
		version:          version,
		accessChecker:    accessChecker,
		protectedHandler: handler,
	})
}

// AddPublicEndpoint adds a traditional HTTP handler without access check
func (api *API) AddPublicEndpoint(url string, httpMethod string, version string,
	publicHandler http.HandlerFunc) {

	api.addEndpoint(APIEndpoint{
		url:           url,
		httpMethod:    httpMethod,
		version:       version,
		publicHandler: publicHandler,
	})
}

func (api *API) applyMergedMiddlewares(h http.Handler) http.Handler {
	// If there is no registered middleware, return a dummy Adapter
	if len(api.middlewares) == 0 {
		return h
	}

	// concatenated
	for _, mw := range api.middlewares {
		h = mw(h)
	}

	return h
}

// LoadInRouter load all API handlers into the provided routing system.
//
// This method aims at making the whole project framework-agonstic: if
// the routing framework change, only this method should change
//
// The Middleware could have been added AFTER the different endpoints
// definition. Consequently, it is better to merge the middleware at the
// last minute, when loading into the router
//
// This method also list all endpoints and the associated HTTP methods for
// CORS handling. Doing this way allows a single endpoint to support multiple
// HTTP method. However, each endpoint, given a method, still need CORS headers
func (api *API) LoadInRouter(router *mux.Router) {

	// to store all endpoints URLs with the associated HTTP methods
	corsOptionsMap := make(map[string]string)

	for _, endpoint := range api.endpoints {
		routeURL := api.urlBuilder(api.root, endpoint.version, endpoint.url)
		api.logger.Debug("[API] \"%s\": URL <%s>", endpoint.httpMethod, routeURL)

		(*router).Handle(
			routeURL,
			api.applyMergedMiddlewares(endpoint.handler),
		).Methods(endpoint.httpMethod)

		// And update CORS definitions
		// TODO: check duplicates of HTTP methods
		endpointDefinition, exist := corsOptionsMap[routeURL]
		if !exist {
			endpointDefinition = endpoint.httpMethod
		} else {
			// It is assumed that there is no duplicate for the moment
			endpointDefinition = endpointDefinition + ", " + endpoint.httpMethod
		}
		corsOptionsMap[routeURL] = endpointDefinition
	}

	api.logger.Verbose("%v", corsOptionsMap)

	// Add CORS handlers
	for endpoint, methods := range corsOptionsMap {
		corsHandler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", api.corsHosts)
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", api.corsHeaders)

			w.WriteHeader(http.StatusOK)
		}

		(*router).HandleFunc(endpoint, corsHandler).Methods(http.MethodOptions)
	}

}
