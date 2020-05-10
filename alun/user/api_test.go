package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
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
		t.Errorf("Register failed: wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}

	var newUser User
	filter := bson.M{"email": userNewEmail}
	if err := dbUserCollection.FindOne(context.TODO(), filter).Decode(&newUser); err != nil {
		t.Errorf("Register failed: email %s of new user is not found in the database",
			userNewEmail)
	}

	reset := newUser.PwdResetToken
	if reset.Token == "" {
		t.Errorf("Register failed: email %s has no reset token in the database",
			userNewEmail)
	}

	userNewResetToken = reset.Token
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
		t.Errorf("Register error: wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestPasswordUpdateNewUser(t *testing.T) {
	reqBody, _ := json.Marshal(pwdChangeRequest{
		Token:    userNewResetToken,
		Username: userNewUsername,
		Password: userNewPassword,
	})
	req, err := http.NewRequest("POST", "/v1/password/update", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("PasswordUpdate (new user) failed: wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}
}

func TestPasswordUpdateWrongToken(t *testing.T) {
	reqBody, _ := json.Marshal(pwdChangeRequest{
		Token:    "Pouet1234",
		Password: userNewPassword,
	})
	req, err := http.NewRequest("POST", "/v1/password/update", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("PasswordUpdate failed: wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}
}

func TestLogin(t *testing.T) {
	reqBody, _ := json.Marshal(authenticatedUser{
		User:     User{BaseUser: BaseUser{Email: userNewEmail}},
		Password: userNewPassword,
	})
	req, err := http.NewRequest("POST", "/v1/login", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Login failed: wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}

	var login successfulLogin
	json.NewDecoder(rr.Body).Decode(&login)
	if login.Token == "" {
		t.Errorf("Login failed: no token found in response %v",
			rr.Body.String())
	}
}
