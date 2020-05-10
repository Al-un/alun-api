package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterNewUser(t *testing.T) {
	reqBody, _ := json.Marshal(PasswordRequest{
		RedirectURL: "http://whatever-url.com?t=",
		BaseUser:    BaseUser{Email: userNewEmail},
		RequestType: 1,
	})
	req, err := http.NewRequest("POST", "/v1/register", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}
}

func TestRegisterExistingEmail(t *testing.T) {
	reqBody, _ := json.Marshal(PasswordRequest{
		RedirectURL: "http://whatever-url.com?t=",
		BaseUser:    BaseUser{Email: userNewEmail},
		RequestType: 1,
	})
	req, err := http.NewRequest("POST", "/v1/register", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}
