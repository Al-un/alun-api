package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Al-un/alun-api/pkg/test"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	userAdminEmail    = "admin@test.com"
	userAdminUsername = "Admin user"
	userAdminPassword = "adminPassword"
	userBasicEmail    = "basic@test.com"
	userBasicUsername = "Basic user"
	userBasicPassword = "basicPassword"
)

var (
	userAdmin = User{
		BaseUser: BaseUser{Email: userAdminEmail},
		IsAdmin:  true,
		PwdResetToken: pwdResetToken{
			Token:     "adminToken",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}
	userBasic = User{
		BaseUser: BaseUser{Email: userBasicEmail},
		IsAdmin:  true,
		PwdResetToken: pwdResetToken{
			Token:     "basicToken",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}
)

func setupUsers(t *testing.T) (string, string, string, string) {
	adminUser, err := createUser(userAdmin)
	if err != nil {
		t.Errorf("Error when setup Admin user %v", err)
	}
	basicUser, err := createUser(userBasic)
	if err != nil {
		t.Errorf("Error when setup Basic user %v", err)
	}

	// adminAuthUser, _ := updatePassword(pwdChangeRequest{
	// 	Token:    "adminToken",
	// 	Password: userAdminPassword,
	// 	Username: userAdminUsername,
	// })
	// basicAuthUser, _ := updatePassword(pwdChangeRequest{
	// 	Token:    "userToken",
	// 	Password: userAdminPassword,
	// 	Username: userAdminUsername,
	// })

	adminJwt, _ := generateJWT(adminUser)
	basicJwt, _ := generateJWT(basicUser)

	return adminUser.ID.Hex(), basicUser.ID.Hex(), adminJwt.Jwt, basicJwt.Jwt
}

func teardownUsers(t *testing.T) {
	// db.al_users.deleteMany({ email: {"$in": ["alun.sng+1@gmail.com", "alun.sng+2@gmail.com"]} })
	// { "acknowledged" : true, "deletedCount" : 2 }

	toBeDeletedEmails := []string{userAdminEmail, userBasicEmail}
	filter := bson.M{
		"email": bson.M{"$in": toBeDeletedEmails},
	}
	d, err := dbUserCollection.DeleteMany(context.TODO(), filter, nil)
	if err != nil {
		userLogger.Info("[User] error in user deletion: ", err)
	}

	fmt.Printf("Deleted %d users with emails %+v\n", d.DeletedCount, toBeDeletedEmails)
}

func TestE2ERegister(t *testing.T) {
	t.Parallel()

	const userEmail = "registering@test.com"
	const userName = "Pouet"
	const userPassword = "PlopPouetProut"

	var userID string
	var resetToken string
	var testInfo test.APITestInfo

	t.Run("NewUser", func(t *testing.T) {
		testInfo = test.APITestInfo{
			Path:   "register",
			Method: "POST",
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
		if reset.Token == "" {
			t.Errorf("Register failed: email %s has no reset token in the database",
				userEmail)
		}

		userID = newUser.ID.Hex()
		resetToken = reset.Token
	})

	t.Run("ExistingEmail", func(t *testing.T) {
		// Re-running the same registration must lead to 400 error
		testInfo.ExpectedHTTPStatus = http.StatusBadRequest
		apiTester.TestPath(t, testInfo)
	})

	t.Run("SetupUsernamePasswordWithWrongToken", func(t *testing.T) {
		testInfo = test.APITestInfo{
			Path:   "password/update",
			Method: "POST",
			Payload: pwdChangeRequest{
				Token:    "A wrong token here",
				Username: userName,
				Password: userPassword,
			},
			ExpectedHTTPStatus: http.StatusBadRequest,
		}
		apiTester.TestPath(t, testInfo)
	})

	t.Run("SetupUsernamePassword", func(t *testing.T) {
		testInfo.Payload = pwdChangeRequest{
			Token:    resetToken,
			Username: userName,
			Password: userPassword,
		}
		testInfo.ExpectedHTTPStatus = http.StatusNoContent
		apiTester.TestPath(t, testInfo)

		updatedUser, _ := findUserByID(userID)
		if updatedUser.Username != userName {
			t.Errorf("%s got Username %s want %s in the database",
				t.Name(), updatedUser.Username, userName)
		}
	})

	t.Run("Login", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/v1/login", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(userEmail, userPassword)

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		apiTester.ServeHTTP(rr, req)

		test.CheckHTTPStatus(t, 2, rr, http.StatusOK)

		var login successfulLogin
		json.NewDecoder(rr.Body).Decode(&login)
		if login.Token == "" {
			t.Errorf("Login failed: no token found in response %v",
				rr.Body.String())
		}
	})

	deleteUser(userID)
	deleteLoginByUserID(userID)
}

// func TestLoginBasic(t *testing.T) {
// 	req, err := http.NewRequest("POST", "/v1/login", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	req.SetBasicAuth(userNewEmail, userNewPassword)

// 	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
// 	rr := httptest.NewRecorder()
// 	apiTester.ServeHTTP(rr, req)

// 	test.CheckHTTPStatus(t, rr, http.StatusOK)

// 	var login successfulLogin
// 	json.NewDecoder(rr.Body).Decode(&login)
// 	if login.Token == "" {
// 		t.Errorf("Login failed: no token found in response %v",
// 			rr.Body.String())
// 	}

// 	testAuthToken = login.Token
// }

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

// func TestLoginJSON(t *testing.T) {
// 	rr := apiTester.TestPath(t, test.APITestInfo{
// 		Path:   "login",
// 		Method: "POST",
// 		Payload: authenticatedUser{
// 			User:     User{BaseUser: BaseUser{Email: userNewEmail}},
// 			Password: userNewPassword,
// 		},
// 		ExpectedHTTPStatus: http.StatusOK,
// 	})

// 	var login successfulLogin
// 	json.NewDecoder(rr.Body).Decode(&login)
// 	if login.Token == "" {
// 		t.Errorf("Login failed: no token found in response %v",
// 			rr.Body.String())
// 	}

// 	testAuthToken = login.Token
// }

// func TestLoginJSONBadCredentials(t *testing.T) {
// 	apiTester.TestPath(t, test.APITestInfo{
// 		Path:   "login",
// 		Method: "POST",
// 		Payload: authenticatedUser{
// 			User:     User{BaseUser: BaseUser{Email: userNewEmail}},
// 			Password: "Prout password",
// 		},
// 		ExpectedHTTPStatus: http.StatusForbidden,
// 	})
// }

// func TestPasswordResetRequest(t *testing.T) {
// 	apiTester.TestPath(t, test.APITestInfo{
// 		Path:   "password/request",
// 		Method: "POST",
// 		Payload: PasswordRequest{
// 			RedirectURL: "http://whatever-url.com?t=",
// 			BaseUser:    BaseUser{Email: userNewEmail},
// 			RequestType: userPwdRequestPwdReset,
// 		},
// 		ExpectedHTTPStatus: http.StatusNoContent,
// 	})
// }

// // func TestPasswordResetRequestInvalidEmail(t *testing.T) {
// // 	apiTester.TestPath(t, "/v1/password/request", "POST", PasswordRequest{
// // 		RedirectURL: "http://whatever-url.com?t=",
// // 		BaseUser:    BaseUser{Email: "animpossibleemail"},
// // 		RequestType: 0,
// // 	}, http.StatusNotFound)
// // }

// func TestGetUser(t *testing.T) {
// 	rr := apiTester.TestPath(t, test.APITestInfo{
// 		Path:               fmt.Sprintf("detail/%s", userNewID),
// 		Method:             "GET",
// 		Payload:            nil,
// 		ExpectedHTTPStatus: http.StatusOK,
// 		AuthToken:          testAuthToken,
// 	})

// 	var user User
// 	json.NewDecoder(rr.Body).Decode(&user)
// 	if user.Email != userNewEmail {
// 		t.Errorf("GetUser got email %s want %s in response %v",
// 			user.Email, userNewEmail, rr.Body.String())
// 	}
// }

// func TestGetUserNotAuthorized(t *testing.T) {
// 	apiTester.TestPath(t, test.APITestInfo{
// 		Path:               fmt.Sprintf("detail/%s", "someInvalidID"),
// 		Method:             "GET",
// 		Payload:            nil,
// 		ExpectedHTTPStatus: http.StatusUnauthorized,
// 		AuthToken:          testAuthToken,
// 	})
// }

// func TestUpdateUser(t *testing.T) {
// 	newUsername := "Prout prout"

// 	rr := apiTester.TestPath(t, test.APITestInfo{
// 		Path:   fmt.Sprintf("detail/%s", userNewID),
// 		Method: "PUT",
// 		Payload: User{
// 			BaseUser: BaseUser{Email: userNewEmail},
// 			Username: newUsername,
// 			IsAdmin:  true, // should not be true in DB
// 		},
// 		ExpectedHTTPStatus: http.StatusOK,
// 		AuthToken:          testAuthToken,
// 	})

// 	updatedUser, _ := findUserByID(userNewID)
// 	if updatedUser.Username != newUsername {
// 		t.Errorf("UpdateUser got Username %s want %s in the database",
// 			updatedUser.Username, newUsername)
// 	}
// 	if updatedUser.IsAdmin {
// 		t.Errorf("UpdateUser allows admin in the database!")
// 	}

// 	var user User
// 	json.NewDecoder(rr.Body).Decode(&user)
// 	if user.Username != newUsername {
// 		t.Errorf("UpdateUser got Username %s want %s in response %v",
// 			user.Username, newUsername, rr.Body.String())
// 	}
// }

func TestEndpointDeleteUser(t *testing.T) {
	t.Parallel()

	_, basicID, _, basicToken := setupUsers(t)

	var testInfo test.APITestInfo

	t.Run("DeleteExistingUser", func(t *testing.T) {
		testInfo = test.APITestInfo{
			Path:               fmt.Sprintf("detail/%s", basicID),
			Method:             "DELETE",
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusNoContent,
			AuthToken:          basicToken,
		}
		apiTester.TestPath(t, testInfo)

		isUserIDExist, err := isUserIDExist(basicID)
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

	teardownUsers(t)
}
