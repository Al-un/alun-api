package memo

import (
	"os"
	"testing"

	"github.com/Al-un/alun-api/alun/testutils"
	"github.com/Al-un/alun-api/alun/user"
	"github.com/Al-un/alun-api/pkg/logger"
)

// ---------- Variables ------------------------------------------------------

const (
	userTestEmail    = "pouet@test.com"
	userTestUsername = "pouet"
	userTestPassword = "pouet"
)

var (
	apiTester *testutils.APITester
	userTest  = user.User{
		BaseUser: user.BaseUser{Email: userTestEmail},
		IsAdmin:  false,
		Username: userTestUsername,
	}
)

// ---------- Main ------------------------------------------------------------
func TestMain(m *testing.M) {
	setupGlobal()

	code := m.Run()

	os.Exit(code)
}

func setupGlobal() {
	// Dummy implementation
	memoLogger = logger.NewSilenceLogger()

	// Setup router
	apiTester = testutils.NewAPITester(MemoAPI)
}

// ---------- Helpers ---------------------------------------------------------
func setupUser(t *testing.T) (user.User, string) {
	createdUser, jwt, err := user.SetupUser(userTest, t.Name(), userTestPassword)
	testutils.Assert(t, testutils.CallFromHelperMethod, err == nil, "Error when setting up users: %v", err)

	return createdUser, jwt
}

func tearDownUser(t *testing.T) {
	_, err := user.TearDownUsers([]user.User{userTest}, t.Name())
	if err != nil {
		t.Logf("Error when CleaningUp users: %+v\n", err)
	}
}
