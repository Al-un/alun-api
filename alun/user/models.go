package user

import (
	"time"

	"github.com/Al-un/alun-api/pkg/crypto"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ----------------------------------------------------------------------------
//	Types: User
// ----------------------------------------------------------------------------

const (
	// Password reset validity time when registering a new user
	pwdResetNewAccountTTL = 10 * time.Minute
	// Password reset validity time when changing password
	pwdResetChgPasswordTTL = 10 * time.Minute
	// New user flag for password reset token
	resetTypeNewAccount = "newAccount"
	// Net password flag for password reset token
	resetTypeResetPwd = "resetPwd"
)

// RegisteringUser has the single Email field to strip out any other field
// sent during a user registration request
//
// A registeringUser does not need an ID as it must be transform into an
// User for being created in the database
//
// RegisteringUser must be an exportable struct so that `bson` tag works
type RegisteringUser struct {
	Email string `json:"email" bson:"email"`
}

// prepareForCreationg takes a registeringUser and build an User from it by
// assigning its first ResetToken with the "new user" flag.
func (rg *RegisteringUser) prepareForCreation() (User, error) {
	token, err := crypto.GenerateRandomString(32)

	if err != nil {
		return User{}, nil
	}

	resetToken := pwdResetToken{
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(pwdResetNewAccountTTL),
		ResetType: resetTypeNewAccount,
	}

	newUser := User{
		RegisteringUser: *rg,
		PwdResetToken:   resetToken,
	}

	return newUser, nil
}

// createPwdResetToken
func (rg *RegisteringUser) createPwdResetToken(killPassword bool) (User, error) {
	token, err := crypto.GenerateRandomString(32)

	if err != nil {
		return User{}, nil
	}

	resetToken := pwdResetToken{
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(pwdResetNewAccountTTL),
		ResetType: resetTypeNewAccount,
	}

	newUser := User{
		RegisteringUser: *rg,
		PwdResetToken:   resetToken,
	}

	return newUser, nil
}

// User represents an entity which can login
//
// db.al_users.insertOne({username:"pouet", password:"plop"})
// curl http://localhost:8000/users/register --data '{"username": "plop", "password": "plop"}'
type User struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	RegisteringUser `bson:",inline"`
	Username        string        `json:"username,omitempty" bson:"username,omitempty"`
	IsAdmin         bool          `json:"isAdmin" bson:"isAdmin"`
	PwdResetToken   pwdResetToken `json:"-" bson:"pwdResetToken,omitempty"` // not present in JSON: https://golang.org/pkg/encoding/json/
}

// AuthenticatedUser has the password field so that when the server sends
// an User back to the client, the Password is not sent.
//
// However, the  JSON field is required otherwise the decoding will ignore
// the password field
type authenticatedUser struct {
	User     `bson:",inline"`
	Password string `json:"password" bson:"password"`
}

// ----------------------------------------------------------------------------
//	Types: Password setup / reset
// ----------------------------------------------------------------------------

// PwdResetToken is the token to define a password reset request. An user can have only one
// password reset request at a time. Such token is also used on user account generation
type pwdResetToken struct {
	Token     string    `json:"resetToken" bson:"token"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt" bson:"expiresAt"`
	ResetType string    `json:"resetType" bson:"resetType"`
}

// pwdChangeRequest defines how a client changes an user password
type pwdChangeRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
	Username string `json:"username"` // only for new account
}

// ----------------------------------------------------------------------------
//	Types: Authorisation
// ----------------------------------------------------------------------------

// authToken saves the generated JWT in the database for re-usability or other
// features such as token invalidation
type authToken struct {
	Jwt       string    `json:"jwt" bson:"jwt,omitempty"`                       // Stringified JWT
	ExpiresOn time.Time `json:"expiresOn,omitempty" bson:"expiresOn,omitempty"` // Convenience for checking token expiration
	Status    int       `json:"status" bson:"status"`                           // Token status [11-Apr-2020] Obsolete?
}

const tokenStatusActive = 0       // Token is still active
const tokenStatusLogout = 1       // Token has been disabled by an user logout
const tokenStatusExpired = 2      // Token has expired
const tokenStatusInvalidated = 10 // Token has been manually disabled by user

// ----------------------------------------------------------------------------
//	Types: Login
// ----------------------------------------------------------------------------

// Login tracks user login and associated generated token.
//
// If a token was re-generated, it should create another Login as there is
// no automatic regeneration upon token expiration.
// UserID field is required to avoid decoding the JWT from `Token`
type Login struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"` // Uniquely identify to invalidate it
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`        // Login timestamp
	UserID    primitive.ObjectID `json:"userId" bson:"userId"`              // Logged-in user, should match the token of the token :)
	Token     authToken          `json:"token" bson:"token"`                // Token generated during login
}

type successfulLogin struct {
	Token string `json:"token"`
}
