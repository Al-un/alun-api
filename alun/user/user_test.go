package user

import (
	"context"
	"os"
	"testing"

	"github.com/Al-un/alun-api/alun/core"
	"github.com/Al-un/alun-api/alun/utils"
	"github.com/Al-un/alun-api/pkg/logger"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	testRouter      *mux.Router
	userNewEmail    string
	userNewPassword string
	userNewUsername string
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()

	tearDown()
	os.Exit(code)
}

func setup() {
	// Dummy implementation
	userLogger = logger.NewSilenceLogger()
	alunEmail = utils.GetDummyEmail()

	// Setup router
	testRouter = core.SetupRouter(
		core.APIMicroservice,
		UserAPI,
	)

	// New user
	userNewEmail = "test@test.com"
	userNewUsername = "Testing account"
	userNewPassword = "Tester"
}

func tearDown() {
	filter := bson.M{"email": userNewEmail}
	dbUserCollection.DeleteMany(context.TODO(), filter, nil)
}
