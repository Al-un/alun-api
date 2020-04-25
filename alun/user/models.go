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
	// Password reset request
	userPwdRequestPwdReset = 0
	// New user creation
	userPwdRequestNewUser = 1
)

// BaseUser has the single Email field to strip out any other field
// sent during a user registration request or password reset request
//
// An BaseUser does not need an ID as it must be transform into an
// User for being created in the database
//
// BaseUser must be an exportable struct so that `bson` tag works
type BaseUser struct {
	Email string `json:"email" bson:"email"`
}

// PasswordRequest involves an email and a request:
//	- new user creation
//	- password reset request
// The redirectURL will tell the server which link has to be added in the email
type PasswordRequest struct {
	RedirectURL string `json:"redirectUrl"`
	// see
	//	userPwdRequestPwdReset => default value: "0"
	//	userPwdRequestNewUser
	RequestType int8 `json:"requestType"`
	BaseUser
}

// CreatePwdResetToken builds a password reset token for an UserPasswordRequest
func (upr *PasswordRequest) createPwdResetToken() (*User, error) {
	// Create random value
	token, err := crypto.GenerateRandomString(32)
	if err != nil {
		return nil, err
	}

	// build token
	resetToken := pwdResetToken{
		Token:       token,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(pwdResetNewAccountTTL),
		RequestType: upr.RequestType,
	}

	newUser := &User{
		PwdResetToken: resetToken,
		BaseUser:      upr.BaseUser,
	}

	return newUser, nil
}

// User represents an entity which can login
//
// db.al_users.insertOne({username:"pouet", password:"plop"})
// curl http://localhost:8000/users/register --data '{"username": "plop", "password": "plop"}'
type User struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	BaseUser      `bson:",inline"`
	Username      string        `json:"username,omitempty" bson:"username,omitempty"`
	IsAdmin       bool          `json:"isAdmin" bson:"isAdmin"`
	PwdResetToken pwdResetToken `json:"-" bson:"pwdResetToken,omitempty"` // not present in JSON: https://golang.org/pkg/encoding/json/
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
	Token       string    `json:"resetToken" bson:"token"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt" bson:"expiresAt"`
	RequestType int8      `json:"requestType" bson:"requestType"`
}

// pwdChangeRequest defines how a client changes an user password
type pwdChangeRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
	Username string `json:"username,omitempty"` // only for new account
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
