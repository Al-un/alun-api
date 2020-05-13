package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Al-un/alun-api/pkg/test"
	"go.mongodb.org/mongo-driver/bson"
)

func TestRegisterNewUser(t *testing.T) {
	apiTester.TestPath(t, "/v1/register", "POST", PasswordRequest{
		RedirectURL: "http://whatever-url.com?t=",
		BaseUser:    BaseUser{Email: userNewEmail},
		RequestType: 1,
	}, http.StatusNoContent)

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

	userNewID = newUser.ID.Hex()
	userNewResetToken = reset.Token
}

func TestRegisterExistingEmail(t *testing.T) {
	apiTester.TestPath(t, "/v1/register", "POST", PasswordRequest{
		RedirectURL: "http://whatever-url.com?t=",
		BaseUser:    BaseUser{Email: userNewEmail},
		RequestType: 1,
	}, http.StatusBadRequest)
}

func TestPasswordUpdateNewUser(t *testing.T) {
	apiTester.TestPath(t, "/v1/password/update", "POST", pwdChangeRequest{
		Token:    userNewResetToken,
		Username: userNewUsername,
		Password: userNewPassword,
	}, http.StatusNoContent)
}

func TestPasswordUpdateWrongToken(t *testing.T) {
	apiTester.TestPath(t, "/v1/password/update", "POST", pwdChangeRequest{
		Token:    "Pouet1234",
		Username: userNewUsername,
		Password: userNewPassword,
	}, http.StatusBadRequest)
}

func TestLoginBasic(t *testing.T) {
	req, err := http.NewRequest("POST", "/v1/login", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth(userNewEmail, userNewPassword)

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	apiTester.ServeHTTP(rr, req)

	test.CheckHTTPStatus(t, rr, http.StatusOK)

	var login successfulLogin
	json.NewDecoder(rr.Body).Decode(&login)
	if login.Token == "" {
		t.Errorf("Login failed: no token found in response %v",
			rr.Body.String())
	}

	testAuthToken = login.Token
}

// func TestLogout(t *testing.T) {
// 	req, err := http.NewRequest("POST", "/v1/logout", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testAuthToken))

// 	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
// 	rr := httptest.NewRecorder()
// 	apiTester.ServeHTTP(rr, req)

// 	test.CheckHTTPStatus(t, rr, http.StatusNoContent)
// }

func TestLoginJSON(t *testing.T) {
	rr := apiTester.TestPath(t, "/v1/login", "POST", authenticatedUser{
		User:     User{BaseUser: BaseUser{Email: userNewEmail}},
		Password: userNewPassword,
	}, http.StatusOK)

	var login successfulLogin
	json.NewDecoder(rr.Body).Decode(&login)
	if login.Token == "" {
		t.Errorf("Login failed: no token found in response %v",
			rr.Body.String())
	}

	testAuthToken = login.Token
}

func TestLoginJSONBadCredentials(t *testing.T) {
	apiTester.TestPath(t, "/v1/login", "POST", authenticatedUser{
		User:     User{BaseUser: BaseUser{Email: userNewEmail}},
		Password: "wrong password",
	}, http.StatusForbidden)
}

func TestPasswordResetRequest(t *testing.T) {
	apiTester.TestPath(t, "/v1/password/request", "POST", PasswordRequest{
		RedirectURL: "http://whatever-url.com?t=",
		BaseUser:    BaseUser{Email: userNewEmail},
		RequestType: 0,
	}, http.StatusNoContent)
}

// func TestPasswordResetRequestInvalidEmail(t *testing.T) {
// 	apiTester.TestPath(t, "/v1/password/request", "POST", PasswordRequest{
// 		RedirectURL: "http://whatever-url.com?t=",
// 		BaseUser:    BaseUser{Email: "animpossibleemail"},
// 		RequestType: 0,
// 	}, http.StatusNotFound)
// }

func TestGetUser(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/v1/detail/%s", userNewID), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testAuthToken))

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	apiTester.ServeHTTP(rr, req)

	test.CheckHTTPStatus(t, rr, http.StatusOK)

	var user User
	json.NewDecoder(rr.Body).Decode(&user)
	if user.Email != userNewEmail {
		t.Errorf("GetUser got email %s want %s in response %v",
			user.Email, userNewEmail, rr.Body.String())
	}
}

func TestGetUserNotAuthorized(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/v1/detail/%s", "someInvalidID"), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testAuthToken))

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	apiTester.ServeHTTP(rr, req)

	test.CheckHTTPStatus(t, rr, http.StatusUnauthorized)
}

func TestUpdateUser(t *testing.T) {
	newUsername := "Prout prout"
	reqBody, _ := json.Marshal(&User{
		BaseUser: BaseUser{Email: userNewEmail},
		Username: newUsername,
		IsAdmin:  true,
	})

	req, err := http.NewRequest("PUT", fmt.Sprintf("/v1/detail/%s", userNewID), bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testAuthToken))

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	apiTester.ServeHTTP(rr, req)

	test.CheckHTTPStatus(t, rr, http.StatusOK)

	updatedUser, _ := findUserByID(userNewID)
	if updatedUser.Username != newUsername {
		t.Errorf("UpdateUser got Username %s want %s in the database",
			updatedUser.Username, newUsername)
	}
	if updatedUser.IsAdmin {
		t.Errorf("UpdateUser allows admin in the database!")
	}

	var user User
	json.NewDecoder(rr.Body).Decode(&user)
	if user.Username != newUsername {
		t.Errorf("UpdateUser got Username %s want %s in response %v",
			user.Username, newUsername, rr.Body.String())
	}
}

func TestDeleteUser(t *testing.T) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/v1/detail/%s", userNewID), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testAuthToken))

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	apiTester.ServeHTTP(rr, req)

	test.CheckHTTPStatus(t, rr, http.StatusNoContent)

	_, err = findUserByID(userNewID)
	if err == nil {
		t.Errorf("DeleteUser did not delete the user in the database")
	}
}
