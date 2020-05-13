package test

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

// TestPath tests a single API calls and returns the generated ResponseRecorder
func (at *APITester) TestPath(t *testing.T, path string, method string, payload interface{}, expectedHTTPStatus int) *httptest.ResponseRecorder {
	reqBody, _ := json.Marshal(payload)
	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	at.router.ServeHTTP(rr, req)

	CheckHTTPStatus(t, rr, expectedHTTPStatus)

	return rr
}

func (at *APITester) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	at.router.ServeHTTP(w, req)
}

// CheckHTTPStatus makes a test fails if the response recorder status code
// is not the expected code
func CheckHTTPStatus(t *testing.T, rr *httptest.ResponseRecorder, expectedStatus int) {
	if status := rr.Code; status != expectedStatus {
		failMsg := fmt.Sprintf("%s got an incorrect HTTP status code: got %v want %v",
			t.Name(), status, expectedStatus)
		t.Errorf(failMsg)
	}
}
