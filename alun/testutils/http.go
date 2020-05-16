package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Al-un/alun-api/alun/core"
	"github.com/gorilla/mux"
)

// APITestInfo gathers info for a single test run
type APITestInfo struct {
	Path               string
	Method             string
	Version            string // default to core.APIv1
	Payload            interface{}
	AuthToken          string
	Headers            map[string]string
	ExpectedHTTPStatus int
}

// APITester saves a router and run tests against the saved router
type APITester struct {
	router *mux.Router
}

// NewAPITester build an APITester from a core.API reference. router is built in
// microservice mode
func NewAPITester(api *core.API) *APITester {
	testedRouter := core.SetupRouter(
		core.APIMicroservice,
		api,
	)

	return &APITester{router: testedRouter}
}

// TestPath tests a single API call and returns the generated ResponseRecorder
func (at *APITester) TestPath(t *testing.T, apiTest APITestInfo) *httptest.ResponseRecorder {
	// Build api path
	apiPathVersion := apiTest.Version
	if apiPathVersion == "" {
		apiPathVersion = core.APIv1
	}
	apiPath := fmt.Sprintf("/%s/%s", apiPathVersion, apiTest.Path)

	// Build request
	var req *http.Request
	var err error
	if apiTest.Payload != nil {
		reqBody, _ := json.Marshal(apiTest.Payload)
		req, err = http.NewRequest(apiTest.Method, apiPath, bytes.NewBuffer(reqBody))
	} else {
		req, err = http.NewRequest(apiTest.Method, apiPath, nil)
	}
	if err != nil {
		t.Fatal(err)
	}

	// Request headers
	if apiTest.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiTest.AuthToken))
	}
	for headerKey, headerVal := range apiTest.Headers {
		req.Header.Set(headerKey, headerVal)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	at.router.ServeHTTP(rr, req)

	CheckHTTPStatus(t, CallLvlHelperMethod, rr, apiTest.ExpectedHTTPStatus)

	return rr
}

func (at *APITester) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	at.router.ServeHTTP(w, req)
}

// CheckHTTPStatus makes a test fails if the response recorder status code
// is not the expected code
func CheckHTTPStatus(t *testing.T, callDepth int, rr *httptest.ResponseRecorder, expectedStatus int) {
	status := rr.Code
	Assert(t, callDepth, status == expectedStatus,
		"%s got an incorrect HTTP status code: got %v want %v",
		t.Name(), status, expectedStatus)
}
