package user

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Al-un/alun-api/alun/utils"
	"github.com/Al-un/alun-api/pkg/logger"
	"github.com/Al-un/alun-api/pkg/test"
	"go.mongodb.org/mongo-driver/bson"
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
	apiTester *test.APITester
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
	apiTester = test.NewAPITester(UserAPI)
}

func tearDownGlobal() {
}

// Returns userId / User JWT token
func setupUsers(t *testing.T, user User) (string, string) {
	createdUser, err := createUser(user)
	if err != nil {
		t.Errorf("Error when setup user %v: %v", user, err)
	}

	jwt, _ := generateJWT(createdUser)

	return createdUser.ID.Hex(), jwt.Jwt
}

// Returns adminID, adminJwt, basicID, basicJwt
func setupUserBasicAndAdmin(t *testing.T) (string, string, string, string) {
	adminID, adminJwt := setupUsers(t, userAdmin)
	basicID, basicJwt := setupUsers(t, userBasic)

	return adminID, adminJwt, basicID, basicJwt
}

func tearDownBasicAndAdmin(t *testing.T) {
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
