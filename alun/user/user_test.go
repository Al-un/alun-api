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
	// Prefix email to avoid conflict between parallel tests
	localUser := user
	localUser.Email = fmt.Sprintf("%s%s", t.Name(), localUser.Email)

	// Create user
	createdUser, err := createUser(localUser)
	if err != nil {
		t.Errorf("Error when creating user %v: %v", user, err)
	}

	// Setup password
	_, err = updatePassword(pwdChangeRequest{
		Token:    localUser.PwdResetToken.Token,
		Password: password,
		Username: user.Username,
	})
	if err != nil {
		t.Errorf("Error when setup password for user %v: %v", user, err)
	}

	jwt, _ := generateJWT(createdUser)

	return createdUser, jwt.Jwt
}

// Returns adminID, adminJwt, basicID, basicJwt
func setupUserBasicAndAdmin(t *testing.T) (User, string, User, string) {
	admin, adminJwt := setupUsers(t, userAdmin, userAdminPassword)
	basic, basicJwt := setupUsers(t, userBasic, userBasicPassword)

	return admin, adminJwt, basic, basicJwt
}

func tearDownBasicAndAdmin(t *testing.T) {
	// db.al_users.deleteMany({ email: {"$in": ["alun.sng+1@gmail.com", "alun.sng+2@gmail.com"]} })
	// { "acknowledged" : true, "deletedCount" : 2 }

	toBeDeletedEmails := []string{
		fmt.Sprintf("%s%s", t.Name(), userAdminEmail),
		fmt.Sprintf("%s%s", t.Name(), userBasicEmail),
	}
	filter := bson.M{
		"email": bson.M{"$in": toBeDeletedEmails},
	}
	d, err := dbUserCollection.DeleteMany(context.TODO(), filter, nil)
	if err != nil {
		userLogger.Info("[User] error in user deletion: ", err)
	}

	fmt.Printf("Deleted %d users with emails %+v\n", d.DeletedCount, toBeDeletedEmails)
}

func tearDownLogins(t *testing.T, userID primitive.ObjectID) {
	filter := bson.M{"userId": userID}
	_, err := dbUserLoginCollection.DeleteMany(context.TODO(), filter)
	if err != nil {
		t.Errorf("Error when cleaning login for userID %s\n", userID)
	}
}
