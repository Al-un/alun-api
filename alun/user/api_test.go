package user

import (
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
	apiTester.TestPath(t, test.APITestInfo{
		Path:   "register",
		Method: "POST",
		Payload: PasswordRequest{
			RedirectURL: "http://whatever-url.com?t=",
			BaseUser:    BaseUser{Email: userNewEmail},
			RequestType: 1,
		},
		ExpectedHTTPStatus: http.StatusNoContent,
	})

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
	apiTester.TestPath(t, test.APITestInfo{
		Path:   "register",
		Method: "POST",
		Payload: PasswordRequest{
			RedirectURL: "http://whatever-url.com?t=",
			BaseUser:    BaseUser{Email: userNewEmail},
			RequestType: userPwdRequestNewUser,
		},
		ExpectedHTTPStatus: http.StatusBadRequest,
	})
}

func TestPasswordUpdateNewUser(t *testing.T) {
	apiTester.TestPath(t, test.APITestInfo{
		Path:   "password/update",
		Method: "POST",
		Payload: pwdChangeRequest{
			Token:    userNewResetToken,
			Username: userNewUsername,
			Password: userNewPassword,
		},
		ExpectedHTTPStatus: http.StatusNoContent,
	})
}

func TestPasswordUpdateWrongToken(t *testing.T) {
	apiTester.TestPath(t, test.APITestInfo{
		Path:   "password/update",
		Method: "POST",
		Payload: pwdChangeRequest{
			Token:    "Pouet1234",
			Username: userNewUsername,
			Password: userNewPassword,
		},
		ExpectedHTTPStatus: http.StatusBadRequest,
	})
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
	rr := apiTester.TestPath(t, test.APITestInfo{
		Path:   "login",
		Method: "POST",
		Payload: authenticatedUser{
			User:     User{BaseUser: BaseUser{Email: userNewEmail}},
			Password: userNewPassword,
		},
		ExpectedHTTPStatus: http.StatusOK,
	})

	var login successfulLogin
	json.NewDecoder(rr.Body).Decode(&login)
	if login.Token == "" {
		t.Errorf("Login failed: no token found in response %v",
			rr.Body.String())
	}

	testAuthToken = login.Token
}

func TestLoginJSONBadCredentials(t *testing.T) {
	apiTester.TestPath(t, test.APITestInfo{
		Path:   "login",
		Method: "POST",
		Payload: authenticatedUser{
			User:     User{BaseUser: BaseUser{Email: userNewEmail}},
			Password: "Prout password",
		},
		ExpectedHTTPStatus: http.StatusForbidden,
	})
}

func TestPasswordResetRequest(t *testing.T) {
	apiTester.TestPath(t, test.APITestInfo{
		Path:   "password/request",
		Method: "POST",
		Payload: PasswordRequest{
			RedirectURL: "http://whatever-url.com?t=",
			BaseUser:    BaseUser{Email: userNewEmail},
			RequestType: userPwdRequestPwdReset,
		},
		ExpectedHTTPStatus: http.StatusNoContent,
	})
}

// func TestPasswordResetRequestInvalidEmail(t *testing.T) {
// 	apiTester.TestPath(t, "/v1/password/request", "POST", PasswordRequest{
// 		RedirectURL: "http://whatever-url.com?t=",
// 		BaseUser:    BaseUser{Email: "animpossibleemail"},
// 		RequestType: 0,
// 	}, http.StatusNotFound)
// }

func TestGetUser(t *testing.T) {
	rr := apiTester.TestPath(t, test.APITestInfo{
		Path:               fmt.Sprintf("detail/%s", userNewID),
		Method:             "GET",
		Payload:            nil,
		ExpectedHTTPStatus: http.StatusOK,
		AuthToken:          testAuthToken,
	})

	var user User
	json.NewDecoder(rr.Body).Decode(&user)
	if user.Email != userNewEmail {
		t.Errorf("GetUser got email %s want %s in response %v",
			user.Email, userNewEmail, rr.Body.String())
	}
}

func TestGetUserNotAuthorized(t *testing.T) {
	apiTester.TestPath(t, test.APITestInfo{
		Path:               fmt.Sprintf("detail/%s", "someInvalidID"),
		Method:             "GET",
		Payload:            nil,
		ExpectedHTTPStatus: http.StatusUnauthorized,
		AuthToken:          testAuthToken,
	})
}

func TestUpdateUser(t *testing.T) {
	newUsername := "Prout prout"

	rr := apiTester.TestPath(t, test.APITestInfo{
		Path:   fmt.Sprintf("detail/%s", userNewID),
		Method: "PUT",
		Payload: User{
			BaseUser: BaseUser{Email: userNewEmail},
			Username: newUsername,
			IsAdmin:  true, // should not be true in DB
		},
		ExpectedHTTPStatus: http.StatusOK,
		AuthToken:          testAuthToken,
	})

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
	apiTester.TestPath(t, test.APITestInfo{
		Path:               fmt.Sprintf("detail/%s", userNewID),
		Method:             "DELETE",
		Payload:            nil,
		ExpectedHTTPStatus: http.StatusNoContent,
		AuthToken:          testAuthToken,
	})

	_, err := findUserByID(userNewID)
	if err == nil {
		t.Errorf("DeleteUser did not delete the user in the database")
	}
}
