package core

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// ----------------------------------------------------------------------------
// 	Constants
// ----------------------------------------------------------------------------

// APIv1 is the standardisation for first version of an API endpoint
const APIv1 string = "v1"

// APIv2 is the standardisation for first version of an API endpoint
const APIv2 string = "v2"

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

// URLBuilder defines how an API generates a final endpoint URL given a root url,
// a version and the endpoint url
type URLBuilder func(root string, version string, url string) string

// ----------------------------------------------------------------------------
// 	Constructors / Constant
// ----------------------------------------------------------------------------

// NewAPI is the API constructor
//
// - Allowed CORS headers and hosts are, for the moment, "*"
// - By default, the JSON middleware is always added
func NewAPI(root string) *API {

	firstChar := root[:1]
	if firstChar != "/" {
		coreLogger.Warn("[API] Root %s does not start with \"/\".", root)
	}

	// init
	api := &API{
		root:        root,
		corsHosts:   "*",
		corsHeaders: "*",
		endpoints:   make([]APIEndpoint, 0), // explicitely define an empty array
		urlBuilder:  URLDefaultBuilder,
	}

	// default middleware
	api.AddMiddleware(AddJSONHeaders)

	return api
}

// URLDefaultBuilder concatenates "/{version}/{root}/{url}""
var URLDefaultBuilder URLBuilder = func(root string, version string, url string) string {
	return fmt.Sprintf("/%s/%s/%s", version, root, url)
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
	if endpoint.httpMethod != "GET" &&
		endpoint.httpMethod != "POST" &&
		endpoint.httpMethod != "PATCH" &&
		endpoint.httpMethod != "PUT" &&
		endpoint.httpMethod != "DELETE" {
		log.Fatalf("Cannot call 'AddHandler' an invalid method \"%s\" for URL %s/%s/%s",
			endpoint.httpMethod, api.root, endpoint.version, endpoint.url)
	}

	// CORS config is the same for both public and protected
	corsConfig := CorsConfig{
		Hosts:   api.corsHosts,
		Headers: api.corsHeaders,
		Methods: endpoint.httpMethod,
	}

	// Public handler: leverage ServeHTTP method
	if endpoint.publicHandler != nil {
		endpoint.handler = AddCorsHeaders(
			endpoint.publicHandler,
			corsConfig,
		).ServeHTTP
	}

	// Protected handler
	if endpoint.protectedHandler != nil {
		endpoint.handler = AddCorsHeaders(
			DoIfAccess(endpoint.accessChecker, endpoint.protectedHandler),
			corsConfig,
		).ServeHTTP
	}

	// All good
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

// LoadInRouter load all API handlers into the provided routing system.
//
// This method aims at making the whole project framework-agonstic: if
// the routing framework change, only this method should change
//
// The Middleware could have been added AFTER the different endpoints
// definition. Consequently, it is better to merge the middleware at the
// last minute, when loading into the router
func (api *API) LoadInRouter(router *mux.Router) {

	// Start with the Identify middleware to avoid multiple IF cases
	var finalMw = func(next http.Handler) http.Handler {
		return next
	}

	// Add all middlewares, worst case, the finalMw remains as the initial
	// Identify middleware
	for _, mw := range api.middlewares {
		finalMw = func(next http.Handler) http.Handler {
			return finalMw(mw(next))
		}
	}

	// Now apply those middleware to all endpoints
	for _, endpoint := range api.endpoints {
		routeURL := api.urlBuilder(api.root, endpoint.version, endpoint.url)
		coreLogger.Debug("[API] \"%s\": URL <%s>", endpoint.httpMethod, routeURL)

		(*router).Handle(routeURL, endpoint.handler).Methods(endpoint.httpMethod)
	}
}
