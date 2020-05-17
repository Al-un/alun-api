package user

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Al-un/alun-api/alun/testutils"
	"github.com/Al-un/alun-api/alun/utils"
	"github.com/Al-un/alun-api/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	userRegisterEmail = "register@test.com"
	userAdminEmail    = "admin@test.com"
	userAdminUsername = "Admin user"
	userAdminPassword = "adminPassword"
	userBasicEmail    = "basic@test.com"
	userBasicUsername = "Basic user"
	userBasicPassword = "basicPassword"
)

var (
	apiTester *testutils.APITester
	userAdmin = User{
		BaseUser: BaseUser{Email: userAdminEmail},
		IsAdmin:  true,
		Username: userAdminUsername,
		PwdResetToken: pwdResetToken{
			Token:     "adminToken",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}
	userBasic = User{
		BaseUser: BaseUser{Email: userBasicEmail},
		IsAdmin:  false,
		Username: userBasicUsername,
		PwdResetToken: pwdResetToken{
			Token:     "basicToken",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}
)

func TestMain(m *testing.M) {
	setupGlobal()

	code := m.Run()

	tearDownGlobal()

	os.Exit(code)
}

func setupGlobal() {
	// Dummy implementation
	userLogger = logger.NewSilenceLogger()
	alunEmail = utils.GetDummyEmail()

	// Setup router
	apiTester = testutils.NewAPITester(UserAPI)
}

func tearDownGlobal() {
}

// Returns userId / User JWT token
func setupUsers(t *testing.T, user User, password string) (User, string) {
	createdUser, jwt, err := SetupUser(user, t.Name(), password)
	testutils.Assert(t, testutils.CallFromHelperMethod, err == nil, "Error when setting up users: %v", err)

	return createdUser, jwt
}

// Returns adminID, adminJwt, basicID, basicJwt
func setupUserBasicAndAdmin(t *testing.T) (User, string, User, string) {
	admin, adminJwt := setupUsers(t, userAdmin, userAdminPassword)
	basic, basicJwt := setupUsers(t, userBasic, userBasicPassword)

	return admin, adminJwt, basic, basicJwt
}

func tearDownBasicAndAdmin(t *testing.T) {
	d, err := TearDownUsers([]User{userAdmin, userBasic}, t.Name())
	if err != nil {
		userLogger.Info("[User] error in user deletion: ", err)
	}

	fmt.Printf("%s > deleted %d users\n", t.Name(), d)
}

func tearDownLogins(t *testing.T, userID primitive.ObjectID) {
	filter := bson.M{"userId": userID}
	_, err := dbUserLoginCollection.DeleteMany(context.TODO(), filter)
	if err != nil {
		t.Errorf("Error when cleaning login for userID %s\n", userID)
	}
}
