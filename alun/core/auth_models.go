package core

import (
	"net/http"
	"time"

	"github.com/Al-un/alun-api/pkg/crypto"
	"github.com/dgrijalva/jwt-go"
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

// User represents a loggable entity
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

// authenticatedUser has the password field so that when the server sends
// an User back to the client, the Password is not sent.

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

// ----------------------------------------------------------------------------
//	Types: Authorisation
// ----------------------------------------------------------------------------

// JwtClaims extends standard claims for our User model.
//
// By including the IsAdmin and UserID fields, authorization check can be
// based on those values
type JwtClaims struct {
	IsAdmin bool   `json:"isAdmin"`
	UserID  string `json:"userId"`
	jwt.StandardClaims
}

// authToken saves the generated JWT in the database for re-usability or other
// features such as token invalidation
type authToken struct {
	Jwt       string    `json:"jwt" bson:"jwt,omitempty"`                       // Stringified JWT
	ExpiresOn time.Time `json:"expiresOn,omitempty" bson:"expiresOn,omitempty"` // Convenience for checking token expiration
	Status    int       `json:"status" bson:"status"`                           // Token status
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

// AuthenticatedHandler is meant to be the core logic of the handler with the user
// informations already extracted from the request.
//
// claims are assumed to be always non-nil and always validated beforehand
type AuthenticatedHandler func(w http.ResponseWriter, r *http.Request, claims JwtClaims)

// AccessChecker ensures that the provided request is allowed to proceed or not.
//
// Most of the checks are based on the token header. An AccessChecker always
// assumes that the JWT is properly formed, hence the jwtClaims argument.
// An AccessChecker's check success often leads to some AuthenticatedHandler to
// proceed.
type AccessChecker func(r *http.Request, jwtClaims JwtClaims) bool

type successfulLogin struct {
	Token string `json:"token"`
}
