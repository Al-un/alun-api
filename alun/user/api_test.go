package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Al-un/alun-api/alun/testutils"
	"go.mongodb.org/mongo-driver/bson"
)

func TestE2ERegister(t *testing.T) {
	t.Parallel()

	const userEmail = userRegisterEmail
	const userName = "Pouet"
	const userPassword = "PlopPouetProut"

	var userID string
	var resetToken string
	var testInfo testutils.APITestInfo

	t.Cleanup(func() {
		if userID != "" {
			deleteUser(userID)
			deleteLoginByUserID(userID)
		} else {
			toBeDeletedEmails := []string{userEmail}
			filter := bson.M{
				"email": bson.M{"$in": toBeDeletedEmails},
			}
			_, err := dbUserCollection.DeleteMany(context.TODO(), filter, nil)
			if err != nil {
				userLogger.Info("[User] error in user deletion: ", err)
			}
		}
	})

	t.Run("NewUser", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:   "register",
			Method: http.MethodPost,
			Payload: PasswordRequest{
				RedirectURL: "http://whatever-url.com?t=",
				BaseUser:    BaseUser{Email: userEmail},
				RequestType: userPwdRequestNewUser,
			},
			ExpectedHTTPStatus: http.StatusNoContent,
		}
		apiTester.TestPath(t, testInfo)

		var newUser User
		filter := bson.M{"email": userEmail}
		if err := dbUserCollection.FindOne(context.TODO(), filter).Decode(&newUser); err != nil {
			t.Errorf("Register failed: email %s of new user is not found in the database",
				userEmail)
		}

		reset := newUser.PwdResetToken
		testutils.Assert(t, testutils.CallFromTestFile, reset.Token != "",
			"Email %s has no reset token in the database",
			userEmail)

		userID = newUser.ID.Hex()
		resetToken = reset.Token
	})

	t.Run("ExistingEmail", func(t *testing.T) {
		// Re-running the same registration must lead to 400 error
		testInfo.ExpectedHTTPStatus = http.StatusBadRequest
		apiTester.TestPath(t, testInfo)
	})

	t.Run("SetupUsernamePassword", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:   "password/update",
			Method: http.MethodPost,
			Payload: pwdChangeRequest{
				Token:    resetToken,
				Username: userName,
				Password: userPassword,
			},
			ExpectedHTTPStatus: http.StatusNoContent,
		}
		apiTester.TestPath(t, testInfo)

		updatedUser, _ := findUserByID(userID)
		testutils.Assert(t, testutils.CallFromTestFile, updatedUser.Username == userName,
			"%s got Username %s want %s in the database",
			t.Name(), updatedUser.Username, userName)
	})

	t.Run("Login", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/v1/login", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(userEmail, userPassword)

		rr := apiTester.ServeReq(req)
		testutils.CheckHTTPStatus(t, testutils.CallFromTestFile, rr, http.StatusOK)

		var login successfulLogin
		json.NewDecoder(rr.Body).Decode(&login)

		testutils.Assert(t, testutils.CallFromTestFile, login.Token != "",
			"Login failed: no token found in response %v",
			rr.Body.String())
	})
}

func TestE2EPasswordReset(t *testing.T) {
	t.Parallel()

	var testInfo testutils.APITestInfo
	var passwordResetToken string
	_, _, basicUser, _ := setupUserBasicAndAdmin(t)

	t.Cleanup(func() {
		tearDownBasicAndAdmin(t)
	})

	t.Run("RequestPasswordReset", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:   "password/request",
			Method: http.MethodPost,
			Payload: PasswordRequest{
				RedirectURL: "http://whatever-url.com?t=",
				BaseUser:    BaseUser{Email: basicUser.Email},
				RequestType: userPwdRequestPwdReset,
			},
			ExpectedHTTPStatus: http.StatusNoContent,
		}
		apiTester.TestPath(t, testInfo)

		usrWithPwdResetToken, err := findUserByID(basicUser.ID.Hex())
		if err != nil {
			t.Errorf("Error when loading user %+v\n", basicUser)
		}
		passwordResetToken = usrWithPwdResetToken.PwdResetToken.Token
	})

	t.Run("RequestPasswordResetInvalidEmail", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:   "password/request",
			Method: http.MethodPost,
			Payload: PasswordRequest{
				RedirectURL: "http://whatever-url.com?t=",
				BaseUser:    BaseUser{Email: "whateveremail"},
				RequestType: userPwdRequestPwdReset,
			},
			ExpectedHTTPStatus: http.StatusNotFound,
		}
		apiTester.TestPath(t, testInfo)
	})

	t.Run("UpdatePassword", func(t *testing.T) {
		newPassword := "SomeFreshlyNewPassword"
		newUsername := "It should not be changed"

		testInfo = testutils.APITestInfo{
			Path:   "password/update",
			Method: http.MethodPost,
			Payload: pwdChangeRequest{
				Token:    passwordResetToken,
				Password: newPassword,
				Username: newUsername,
			},
			ExpectedHTTPStatus: http.StatusNoContent,
		}
		apiTester.TestPath(t, testInfo)

		// Password is updated
		_, err := findUserByEmailPassword(basicUser.Email, newPassword)
		testutils.Ok(t, testutils.CallFromTestFile, err)

		// Username is unchanged
		updatedUser, _ := findUserByID(basicUser.ID.Hex())
		testutils.Assert(t, testutils.CallFromTestFile, updatedUser.Username == userBasicUsername,
			"Username of Email %s is %s. Expected unchanged: %s",
			updatedUser.Email, updatedUser.Username, userBasicUsername)
	})

	t.Run("UpdatePasswordInvalidToken", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:   "password/update",
			Method: http.MethodPost,
			Payload: pwdChangeRequest{
				Token:    "osef",
				Password: "osef",
			},
			ExpectedHTTPStatus: http.StatusNotFound,
		}
		apiTester.TestPath(t, testInfo)
	})
}

func TestEndpointLogin(t *testing.T) {
	t.Parallel()

	var testInfo testutils.APITestInfo
	_, _, basicUser, _ := setupUserBasicAndAdmin(t)

	t.Cleanup(func() {
		tearDownBasicAndAdmin(t)
		tearDownLogins(t, basicUser.ID)
	})

	testLoginHasToken := func(rr *httptest.ResponseRecorder) {
		var login successfulLogin
		json.NewDecoder(rr.Body).Decode(&login)

		testutils.Assert(t, testutils.CallFromTestFile, login.Token != "",
			"Login failed: no token found in response %v",
			rr.Body.String())
	}

	t.Run("BasicAuthSuccessful", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/v1/login", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(basicUser.Email, userBasicPassword)

		rr := apiTester.ServeReq(req)

		testutils.CheckHTTPStatus(t, testutils.CallFromTestFile, rr, http.StatusOK)

		testLoginHasToken(rr)
	})

	t.Run("BasicAuthIncorrect", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/v1/login", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(basicUser.Email, "SomeInvalidPassword")

		rr := apiTester.ServeReq(req)

		testutils.CheckHTTPStatus(t, testutils.CallFromTestFile, rr, http.StatusForbidden)
	})

	t.Run("JSONAuthSuccessful", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:   "login",
			Method: http.MethodPost,
			Payload: authenticatedUser{
				User:     User{BaseUser: BaseUser{Email: basicUser.Email}},
				Password: userBasicPassword,
			},
			ExpectedHTTPStatus: http.StatusOK,
		}
		rr := apiTester.TestPath(t, testInfo)

		testLoginHasToken(rr)
	})

	t.Run("JSONAuthIncorrect", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:   "login",
			Method: http.MethodPost,
			Payload: authenticatedUser{
				User:     User{BaseUser: BaseUser{Email: basicUser.Email}},
				Password: "ProutPassword",
			},
			ExpectedHTTPStatus: http.StatusForbidden,
		}
		apiTester.TestPath(t, testInfo)
	})

	t.Run("HandleMissingAuthInfo", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/v1/login", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := apiTester.ServeReq(req)

		testutils.CheckHTTPStatus(t, testutils.CallFromTestFile, rr, http.StatusBadRequest)
	})

	t.Run("HandleMissingAuthInfo", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/v1/login", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "")
		rr := apiTester.ServeReq(req)

		testutils.CheckHTTPStatus(t, testutils.CallFromTestFile, rr, http.StatusBadRequest)
	})
}

func TestEndpointGetUser(t *testing.T) {
	t.Parallel()

	adminUser, adminToken, basicUser, basicToken := setupUserBasicAndAdmin(t)
	var testInfo testutils.APITestInfo

	t.Cleanup(func() {
		tearDownBasicAndAdmin(t)
	})

	t.Run("GetOwnUserWithBasicToken", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:               fmt.Sprintf("detail/%s", basicUser.ID.Hex()),
			Method:             http.MethodGet,
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusOK,
			AuthToken:          basicToken,
		}
		rr := apiTester.TestPath(t, testInfo)

		var user User
		json.NewDecoder(rr.Body).Decode(&user)
		testutils.Equals(t, testutils.CallFromTestFile, basicUser.BaseUser, user.BaseUser)
	})

	t.Run("GetUserWithAdminToken", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:               fmt.Sprintf("detail/%s", basicUser.ID.Hex()),
			Method:             http.MethodGet,
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusOK,
			AuthToken:          adminToken,
		}
		rr := apiTester.TestPath(t, testInfo)

		var user User
		json.NewDecoder(rr.Body).Decode(&user)
		testutils.Equals(t, testutils.CallFromTestFile, basicUser.BaseUser, user.BaseUser)
	})

	t.Run("GetAdminUserWithBasicToken", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:               fmt.Sprintf("detail/%s", adminUser.ID.Hex()),
			Method:             http.MethodGet,
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusUnauthorized,
			AuthToken:          basicToken,
		}
		apiTester.TestPath(t, testInfo)
	})
}

func TestEndpointUpdateUser(t *testing.T) {
	t.Parallel()

	adminUser, adminToken, basicUser, basicToken := setupUserBasicAndAdmin(t)
	var testInfo testutils.APITestInfo

	t.Cleanup(func() {
		tearDownBasicAndAdmin(t)
	})

	checkUserUpdate := func(rr *httptest.ResponseRecorder, newUsername, newEmail string) {
		var userResp User
		json.NewDecoder(rr.Body).Decode(&userResp)
		testutils.Equals(t, testutils.CallFromHelperMethod, newEmail, userResp.Email)
		testutils.Equals(t, testutils.CallFromHelperMethod, newUsername, userResp.Username)

		updatedUser, _ := findUserByID(basicUser.ID.Hex())
		testutils.Equals(t, testutils.CallFromHelperMethod, newEmail, updatedUser.Email)
		testutils.Equals(t, testutils.CallFromHelperMethod, newUsername, updatedUser.Username)
		testutils.Equals(t, testutils.CallFromHelperMethod, false, updatedUser.IsAdmin)
	}

	t.Run("UpdateOwnUserWithBasicToken", func(t *testing.T) {
		var basicUserNewEmail = "PlopMachine@test.com"
		var basicUserNewUsername = "PlopMachine"

		testInfo = testutils.APITestInfo{
			Path:      fmt.Sprintf("detail/%s", basicUser.ID.Hex()),
			Method:    http.MethodPut,
			AuthToken: basicToken,
			Payload: User{
				BaseUser: BaseUser{Email: basicUserNewEmail},
				Username: basicUserNewUsername,
				IsAdmin:  true, // should not be true in DB
			},
			ExpectedHTTPStatus: http.StatusOK,
		}
		rr := apiTester.TestPath(t, testInfo)

		checkUserUpdate(rr, basicUserNewUsername, basicUserNewEmail)
	})

	t.Run("UpdateUserWithAdminToken", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:      fmt.Sprintf("detail/%s", basicUser.ID.Hex()),
			Method:    http.MethodPut,
			AuthToken: adminToken,
			Payload: User{
				BaseUser: BaseUser{Email: userBasicEmail},
				Username: userBasicUsername,
				IsAdmin:  true, // should not be true in DB
			},
			ExpectedHTTPStatus: http.StatusOK,
		}
		rr := apiTester.TestPath(t, testInfo)

		checkUserUpdate(rr, userBasicUsername, userBasicEmail)
	})

	t.Run("UpdateAdminUserWithBasicToken", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:      fmt.Sprintf("detail/%s", adminUser.ID.Hex()),
			Method:    http.MethodPut,
			AuthToken: basicToken,
			Payload: User{
				BaseUser: BaseUser{Email: userBasicEmail},
				Username: userBasicUsername,
				IsAdmin:  true, // should not be true in DB
			},
			ExpectedHTTPStatus: http.StatusUnauthorized,
		}
		apiTester.TestPath(t, testInfo)
	})
}

func TestEndpointDeleteUser(t *testing.T) {
	t.Parallel()

	_, _, basicUser, basicToken := setupUserBasicAndAdmin(t)
	var testInfo testutils.APITestInfo

	t.Cleanup(func() {
		tearDownBasicAndAdmin(t)
	})

	t.Run("DeleteExistingUser", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:               fmt.Sprintf("detail/%s", basicUser.ID.Hex()),
			Method:             "DELETE",
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusNoContent,
			AuthToken:          basicToken,
		}
		apiTester.TestPath(t, testInfo)

		isUserIDExist, err := isUserIDExist(basicUser.ID.Hex())
		if err != nil {
			t.Errorf("DeleteUser isUserIDExist check trigger error %v\n", err)
		}
		if isUserIDExist {
			t.Errorf("DeleteUser did not delete the user in the database")
		}
	})

	t.Run("DeleteUnexistentUser", func(t *testing.T) {
		testInfo.ExpectedHTTPStatus = http.StatusNotFound
		apiTester.TestPath(t, testInfo)
	})
}

func TestEndpointLogout(t *testing.T) {
	t.Parallel()

	_, _, basicUser, _ := setupUserBasicAndAdmin(t)
	var authToken string
	var testInfo testutils.APITestInfo

	t.Cleanup(func() {
		tearDownBasicAndAdmin(t)
		tearDownLogins(t, basicUser.ID)
	})

	t.Run("Logout", func(t *testing.T) {
		// Login first
		testInfo = testutils.APITestInfo{
			Path:   "login",
			Method: http.MethodPost,
			Payload: authenticatedUser{
				User:     User{BaseUser: BaseUser{Email: basicUser.Email}},
				Password: userBasicPassword,
			},
			ExpectedHTTPStatus: http.StatusOK,
		}
		rr := apiTester.TestPath(t, testInfo)

		var successfulLogin successfulLogin
		json.NewDecoder(rr.Body).Decode(&successfulLogin)
		authToken = successfulLogin.Token

		testInfo = testutils.APITestInfo{
			Path:               "logout",
			Method:             http.MethodPost,
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusNoContent,
			AuthToken:          authToken,
		}
		apiTester.TestPath(t, testInfo)
	})

	t.Run("LogoutWithInvalidtoken", func(t *testing.T) {
		t.Skip("Invalid/Malformed JWT Token to be handled")
		testInfo = testutils.APITestInfo{
			Path:               "logout",
			Method:             http.MethodPost,
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusNoContent,
			AuthToken:          "someInvalidToken",
		}
		apiTester.TestPath(t, testInfo)
	})
}

// func TestLogout(t *testing.T) {
// 	req, err := http.NewRequest(http.MethodPost, "/v1/logout", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testAuthToken))

// 	rr := httptest.NewRecorder()
// 	apiTester.ServeHTTP(rr, req)

// 	testutils.CheckHTTPStatus(t, rr, http.StatusNoContent)
// }
